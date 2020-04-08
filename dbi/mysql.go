package dbi

import (
	"database/sql"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	gouio "github.com/bingoohuang/gou/io"
	"github.com/bingoohuang/gou/str"

	"github.com/bingoohuang/sqlx"

	"github.com/bingoohuang/faker"
	"github.com/bingoohuang/pump/model"
	"github.com/bingoohuang/pump/random"
)

// MySQLTable ...
type MySQLTable struct {
	Name    string `name:"TABLE_NAME"`
	Comment string `name:"TABLE_COMMENT"`
}

var _ model.Table = (*MySQLTable)(nil)

// GetName ...
func (m MySQLTable) GetName() string { return m.Name }

// GetComment  ...
func (m MySQLTable) GetComment() string { return m.Comment }

// MyTableColumn ...
type MyTableColumn struct {
	Name         string `name:"COLUMN_NAME"`
	Type         string `name:"COLUMN_TYPE"`
	Extra        string `name:"EXTRA"` // auto_increment
	Comment      string `name:"COLUMN_COMMENT"`
	DataType     string `name:"DATA_TYPE"`
	MaxLength    int    `name:"CHARACTER_MAXIMUM_LENGTH"`
	Nullable     string `name:"IS_NULLABLE"`
	Default      string `name:"COLUMN_DEFAULT"`
	CharacterSet string `name:"CHARACTER_SET_NAME"`

	NumericPrecision int `name:"NUMERIC_PRECISION"`
	NumericScale     int `name:"NUMERIC_SCALE"`

	randomizer model.Randomizer
}

var _ model.TableColumn = (*MyTableColumn)(nil)

// IsNullable ...
func (c MyTableColumn) IsNullable() bool { return c.Nullable == "YES" }

// GetMaxSize ...
func (c MyTableColumn) GetMaxSize() int { return c.MaxLength }

// GetDataType ...
func (c MyTableColumn) GetDataType() string { return c.DataType }

// GetName ...
func (c MyTableColumn) GetName() string { return c.Name }

// GetComment ...
func (c MyTableColumn) GetComment() string { return c.Comment }

// GetRandomizer ...
func (c MyTableColumn) GetRandomizer() model.Randomizer { return c.randomizer }

// GetCharacterSet returns the CharacterSet of the column
func (c MyTableColumn) GetCharacterSet() string { return c.CharacterSet }

// MySQLSchema ...
type MySQLSchema struct {
	dbFn          func() (*sql.DB, error)
	pumpOptionReg *regexp.Regexp
	DS            string

	verbose int
}

var _ model.DBSchema = (*MySQLSchema)(nil)

// CreateMySQLSchema ...
func CreateMySQLSchema(dataSourceName string) (*MySQLSchema, error) {
	ds := sqlx.CompatibleMySQLDs(dataSourceName)
	more := sqlx.NewSQLMore("mysql", ds)

	return &MySQLSchema{
		dbFn:          more.OpenE,
		pumpOptionReg: regexp.MustCompile(`\bpump:"([^"]+)"`),
		DS:            more.EnhancedURI,
	}, nil
}

// CompatibleDs returns the dataSourceName from various the compatible format.
func (m MySQLSchema) CompatibleDs() string { return m.DS }

// Tables ...
func (m MySQLSchema) Tables() ([]model.Table, error) {
	db, err := m.dbFn()
	if err != nil {
		return nil, err
	}

	defer gouio.Close(db)

	sqlx.DB = db

	var dao mysqlSchemaDao
	if err := sqlx.CreateDao(&dao, sqlx.WithSQLStr(mysqlSchemaDaoSQL)); err != nil {
		return nil, err
	}

	tables := dao.GetTables()
	ts := make([]model.Table, len(tables))

	for i, t := range tables {
		t.Comment = strings.TrimSpace(t.Comment)
		ts[i] = t
	}

	return ts, nil
}

const mysqlSchemaDaoSQL = `
-- name: GetTableColumns
SELECT * FROM information_schema.COLUMNS 
WHERE TABLE_SCHEMA = /* if _1 != "" */ :1  /* else */ database() /* end */ AND TABLE_NAME = :2 
ORDER BY ORDINAL_POSITION;

-- name: GetTables
SELECT * FROM information_schema.TABLES 
WHERE TABLE_SCHEMA = database();
`

type mysqlSchemaDao struct {
	GetTables       func() []MySQLTable
	GetTableColumns func(schema, table string) []MyTableColumn
}

// TableColumns ...
func (m MySQLSchema) TableColumns(table string) ([]model.TableColumn, error) {
	db, err := m.dbFn()
	if err != nil {
		return nil, err
	}

	defer gouio.Close(db)

	schema, tableName := ParseTable(table)

	var dao mysqlSchemaDao

	sqlx.DB = db

	if err := sqlx.CreateDao(&dao, sqlx.WithSQLStr(mysqlSchemaDaoSQL)); err != nil {
		return nil, err
	}

	columns := dao.GetTableColumns(schema, tableName)

	ts := make([]model.TableColumn, len(columns))

	for i, t := range columns {
		t.Comment = strings.TrimSpace(t.Comment)
		t.randomizer = m.makeColumnRandomizer(t)
		ts[i] = t
	}

	return ts, nil
}

// ParseTable parses the schema and table name from table which may be like db1.t1
func ParseTable(table string) (schemaName, tableName string) {
	if strings.Contains(table, ".") {
		return str.Split2(table, ".", true, true)
	}

	return "", table
}

// Pump ...
func (m MySQLSchema) Pump(table string, rowsPumped chan<- model.RowsPumped, config model.PumpConfig,
	ready chan bool, onerr string, retryMaxTimes int) error {
	columns, err := m.TableColumns(table)
	if err != nil {
		return err
	}

	randMap := makeRandomizerMap(columns)
	columnNames := makeInsertColumns(randMap, columns)

	db, err := m.dbFn()
	if err != nil {
		return err
	}

	defer gouio.Close(db)

	t := time.Now()
	rows := config.RandRows()

	batch := NewInsertBatch(table, columnNames, config.BatchNum, db, func(rows int) {
		rowsPumped <- model.RowsPumped{Table: table, Rows: rows, Cost: time.Since(t)}
		t = time.Now()
	}, m.verbose, rows)

	ready <- true

	retryState := &retryState{retryMaxTimes: retryMaxTimes, onErr: onerr, verbose: m.verbose}

	for i := 1; i <= rows; i++ {
		colValues := make([]interface{}, len(columnNames))
		for j, col := range columnNames {
			colValues[j] = randMap[col].Value()
		}

		err := batch.AddRow(colValues)

		if retry, err := retryState.retry(err); err != nil {
			return err
		} else if retry {
			i -= batch.GetBatchNum() // revert the whole batch num rows
		}
	}

	retryState.retries = 0

	for {
		_, err := batch.Complete()
		if retry, err := retryState.retry(err); err != nil {
			return err
		} else if !retry {
			return nil
		}
	}
}

type retryState struct {
	retries       int
	retryMaxTimes int
	onErr         string
	verbose       int
}

func (r *retryState) retry(err error) (bool, error) {
	if err == nil || r.onErr != "retry" {
		return false, err
	}

	if r.verbose > 0 {
		logrus.Warnf("retry %d after error %v", r.retries, err)
	}

	if r.retryMaxTimes <= 0 {
		return true, nil
	}

	if r.retries >= r.retryMaxTimes {
		if r.verbose > 0 {
			logrus.Warnf("retry %d reached max %d", r.retries, r.retryMaxTimes)
		}

		return false, err
	}

	r.retries++

	return true, nil
}

func makeInsertColumns(randMap map[string]model.Randomizer, columns []model.TableColumn) []string {
	columnNames := make([]string, len(randMap))

	i := 0

	for _, c := range columns {
		if _, ok := randMap[c.GetName()]; ok {
			columnNames[i] = c.GetName()
			i++
		}
	}

	return columnNames
}

func makeRandomizerMap(columns []model.TableColumn) map[string]model.Randomizer {
	randMap := make(map[string]model.Randomizer)

	for _, col := range columns {
		if r := col.GetRandomizer(); r != nil {
			randMap[col.GetName()] = r
		}
	}

	return randMap
}

// nolint gomnd
func (m MySQLSchema) makeColumnRandomizer(c MyTableColumn) model.Randomizer {
	sub := m.pumpOptionReg.FindStringSubmatch(c.GetComment())
	pumpOption := ""

	if len(sub) == 2 {
		pumpOption = sub[1]
	}

	if pumpOption == "-" { // ignore
		return nil
	}

	// nolint
	// mysql> show create table a.ta \G
	// *************************** 1. row ***************************
	// 	Table: ta
	// Create Table: CREATE TABLE `ta` (
	// 	`id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id',
	// 	`name` varchar(10) DEFAULT NULL,
	// 	`age` int(11) DEFAULT NULL,
	// 	PRIMARY KEY (`id`)
	// ) ENGINE=InnoDB AUTO_INCREMENT=2079863083 DEFAULT CHARSET=utf8
	// 1 row in set (0.00 sec)
	// mysql> show full columns from a.ta;
	// +-------+-------------+-----------------+------+-----+---------+----------------+---------------------------------+---------+
	// | Field | Type        | Collation       | Null | Key | Default | Extra          | Privileges                      | Comment |
	// 	+-------+-------------+-----------------+------+-----+---------+----------------+---------------------------------+---------+
	// | id    | int(11)     | NULL            | NO   | PRI | NULL    | auto_increment | select,insert,update,references | id      |
	// | name  | varchar(10) | utf8_general_ci | YES  |     | NULL    |                | select,insert,update,references |         |
	// | age   | int(11)     | NULL            | YES  |     | NULL    |                | select,insert,update,references |         |
	// +-------+-------------+-----------------+------+-----+---------+----------------+---------------------------------+---------+
	// mysql> ALTER TABLE a.ta  MODIFY column id int auto_increment comment 'pump:"--"';
	// mysql> ALTER TABLE a.ta  MODIFY column id int auto_increment comment 'pump:"-"';
	// mysql> ALTER TABLE a.ta  MODIFY column id int auto_increment comment 'id';
	if c.IsAutoIncrement() {
		if pumpOption == "" { // ignore auto_increment by default
			return nil
		}

		if pumpOption == "--" { // not ignore
			return c.randomColumn()
		}
	}

	if pumpOption == "" {
		return c.randomColumn()
	}

	return random.NewFn(func() interface{} {
		val, _ := faker.FakeColumnWithType(c.zeroType(), pumpOption)
		return val
	})
}

// SetVerbose set verbose mode
func (m *MySQLSchema) SetVerbose(verbose int) {
	m.verbose = verbose

	if verbose > 0 {
		logrus.Infof("dataSourceName:%s", m.DS)
	}
}

// IsAutoIncrement tells if the col is auto_increment or not.
// eg. create table a.ta(id int auto_increment, name varchar(10), age int, primary key(id));
func (c MyTableColumn) IsAutoIncrement() bool {
	return strings.Contains(c.Extra, "auto_increment")
}

func (c MyTableColumn) zeroType() reflect.Type {
	typ := c.GetDataType()
	switch typ {
	case "tinyint", "smallint", "mediumint", "int", "integer", "bigint":
		return random.IntZero()
	case "float", "decimal", "double":
		return random.DecimalZero()
	case "char", "varchar", "tinyblob",
		"tinytext", "blob", "text", "mediumtext",
		"mediumblob", "longblob", "longtext":
		return random.StrZero()
	case "date":
		return random.DateZero()
	case "datetime", "timestamp":
		return random.DateTimeInRangeZero()
	case "time":
		return random.TimeZero()
	default:
		log.Panicf("cannot get field type: %s: %s\n", c.GetName(), c.GetDataType())
	}

	return reflect.TypeOf(nil)
}

// nolint gomnd
func (c MyTableColumn) randomColumn() model.Randomizer {
	typ := c.GetDataType()
	switch typ {
	case "tinyint", "smallint", "mediumint", "int", "integer", "bigint":
		maxValues := map[string]int64{
			"tinyint":   0xF,
			"smallint":  0xFF,
			"mediumint": 0x7FFFF,
			"int":       0x7FFFFFFF,
			"integer":   0x7FFFFFFF,
			"float":     0x7FFFFFFF,
			"decimal":   0x7FFFFFFF,
			"double":    0x7FFFFFFF,
			"bigint":    0x7FFFFFFFFFFFFFFF,
		}
		maxValue := maxValues["bigint"]

		if m, ok := maxValues[typ]; ok {
			maxValue = m
		}

		return random.NewRandomInt(c, maxValue)

	case "float", "decimal", "double":
		return random.NewRandomDecimal(c, c.NumericPrecision-c.NumericScale)

	case "char", "varchar",
		"tinyblob", "blob", "mediumblob", "longblob",
		"tinytext", "text", "mediumtext", "longtext":
		return random.NewRandomStr(c)

	case "date":
		return random.NewRandomDate(c)
	case "datetime", "timestamp":
		return random.NewRandomDateTime()
	case "time":
		return random.NewRandomTime(c)
	default:
		log.Panicf("cannot get field type: %s: %s\n", c.GetName(), c.GetDataType())
	}

	return nil
}
