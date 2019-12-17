package dbi

import (
	"database/sql"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/pump/ds"

	gouio "github.com/bingoohuang/gou/io"
	"github.com/bingoohuang/gou/str"

	"github.com/bingoohuang/sqlmore"

	"github.com/bingoohuang/faker"
	"github.com/bingoohuang/pump/model"
	"github.com/bingoohuang/pump/random"
	"github.com/jinzhu/gorm"
)

// MySQLTable ...
type MySQLTable struct {
	Name    string `gorm:"column:TABLE_NAME"`
	Comment string `gorm:"column:TABLE_COMMENT"`
}

var _ model.Table = (*MySQLTable)(nil)

// GetName ...
func (m MySQLTable) GetName() string { return m.Name }

// GetComment  ...
func (m MySQLTable) GetComment() string { return m.Comment }

// MyTableColumn ...
type MyTableColumn struct {
	Name      string         `gorm:"column:COLUMN_NAME"`
	Type      string         `gorm:"column:COLUMN_TYPE"`
	Extra     string         `gorm:"column:EXTRA"` // auto_increment
	Comment   string         `gorm:"column:COLUMN_COMMENT"`
	DataType  string         `gorm:"column:DATA_TYPE"`
	MaxLength sql.NullInt64  `gorm:"column:CHARACTER_MAXIMUM_LENGTH"`
	Nullable  string         `gorm:"column:IS_NULLABLE"`
	Default   sql.NullString `gorm:"column:COLUMN_DEFAULT"`

	NumericPrecision sql.NullInt64 `gorm:"column:NUMERIC_PRECISION"`
	NumericScale     sql.NullInt64 `gorm:"column:NUMERIC_SCALE"`

	randomizer model.ColumnRandomizer
}

var _ model.TableColumn = (*MyTableColumn)(nil)

// IsAllowNull ...
func (c MyTableColumn) IsAllowNull() bool { return c.Nullable == "YES" }

// GetType ...
func (c MyTableColumn) GetType() string { return c.Type }

// GetMaxSize ...
func (c MyTableColumn) GetMaxSize() sql.NullInt64 { return c.MaxLength }

// GetDataType ...
func (c MyTableColumn) GetDataType() string { return c.DataType }

// GetName ...
func (c MyTableColumn) GetName() string { return c.Name }

// GetComment ...
func (c MyTableColumn) GetComment() string { return c.Comment }

// GetColumnRandomizer ...
func (c MyTableColumn) GetColumnRandomizer() model.ColumnRandomizer { return c.randomizer }

// MySQLSchema ...
type MySQLSchema struct {
	dbFn          func() (*gorm.DB, error)
	pumpOptionReg *regexp.Regexp
	compatibleDs  string
}

var _ model.DbSchema = (*MySQLSchema)(nil)

// CreateMySQLSchema ...
func CreateMySQLSchema(dataSourceName string) (*MySQLSchema, error) {
	compatibleDs := ds.CompatibleMySQLDs(dataSourceName)
	logrus.Infof("dataSourceName:%s", compatibleDs)

	dbFn := func() (*gorm.DB, error) { return sqlmore.NewSQLMore("mysql", compatibleDs).GormOpen() }

	return &MySQLSchema{
		dbFn:          dbFn,
		pumpOptionReg: regexp.MustCompile(`\bpump:"([^"]+)"`),
		compatibleDs:  compatibleDs,
	}, nil
}

// CompatibleDs returns the dataSourceName from various the compatible format.
func (m MySQLSchema) CompatibleDs() string { return m.compatibleDs }

// Tables ...
func (m MySQLSchema) Tables() ([]model.Table, error) {
	db, err := m.dbFn()
	if err != nil {
		return nil, err
	}

	defer gouio.Close(db)

	var tables []MySQLTable

	const s = `SELECT * FROM information_schema.TABLES WHERE TABLE_SCHEMA = database()`

	db.Raw(s).Find(&tables)

	ts := make([]model.Table, len(tables))

	for i, t := range tables {
		t.Comment = strings.TrimSpace(t.Comment)
		ts[i] = t
	}

	return ts, db.Error
}

// TableColumns ...
func (m MySQLSchema) TableColumns(table string) ([]model.TableColumn, error) {
	db, err := m.dbFn()
	if err != nil {
		return nil, err
	}

	defer gouio.Close(db)

	columns := make([]MyTableColumn, 0)
	schema, tableName := ParseTable(table)

	if schema != "" {
		const s = `SELECT * FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? ` +
			`AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION`

		db.Raw(s, schema, tableName).Find(&columns)
	} else {
		const s = `SELECT * FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = database() ` +
			`AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION`

		db.Raw(s, tableName).Find(&columns)
	}

	ts := make([]model.TableColumn, len(columns))

	for i, t := range columns {
		t.Comment = strings.TrimSpace(t.Comment)
		t.randomizer = m.makeColumnRandomizer(t)
		ts[i] = t
	}

	return ts, db.Error
}

// ParseTable parses the schema and table name from table which may be like db1.t1
func ParseTable(table string) (schemaName, tableName string) {
	if strings.Contains(table, ".") {
		return str.Split2(table, ".", true, true)
	}

	return "", table
}

// Pump ...
func (m MySQLSchema) Pump(table string, rowsPumped chan<- model.RowsPumped, config model.PumpConfig) error {
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

	batch := NewInsertBatch(table, columnNames, config.BatchNum, db, func(rows int) {
		rowsPumped <- model.RowsPumped{Table: table, Rows: rows, Cost: time.Since(t)}
		t = time.Now()
	})

	rows := config.RandRows()

	for i := 1; i <= rows; i++ {
		colValues := make([]interface{}, len(columnNames))

		for j, col := range columnNames {
			colValues[j] = randMap[col].Value()
		}

		batch.AddRow(colValues)
	}

	batch.Complete()

	return nil
}

func makeInsertColumns(randMap map[string]model.ColumnRandomizer, columns []model.TableColumn) []string {
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

func makeRandomizerMap(columns []model.TableColumn) map[string]model.ColumnRandomizer {
	randMap := make(map[string]model.ColumnRandomizer)

	for _, col := range columns {
		if r := col.GetColumnRandomizer(); r != nil {
			randMap[col.GetName()] = r
		}
	}

	return randMap
}

func (m MySQLSchema) makeColumnRandomizer(c MyTableColumn) model.ColumnRandomizer {
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

func (c MyTableColumn) randomColumn() model.ColumnRandomizer {
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
		return random.NewRandomDecimal(c, c.NumericPrecision.Int64-c.NumericScale.Int64)

	case "char", "varchar", "tinyblob",
		"tinytext", "blob", "text", "mediumtext",
		"mediumblob", "longblob", "longtext":
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
