package dbi

import (
	"database/sql"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/bingoohuang/sqlmore"

	"github.com/bingoohuang/faker"
	"github.com/bingoohuang/pump/model"
	"github.com/bingoohuang/pump/random"
	"github.com/bingoohuang/pump/util"
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
	Comment   string         `gorm:"column:COLUMN_COMMENT"`
	DataType  string         `gorm:"column:DATA_TYPE"`
	MaxLength sql.NullInt64  `gorm:"column:CHARACTER_MAXIMUM_LENGTH"`
	Nullable  string         `gorm:"column:IS_NULLABLE"`
	Default   sql.NullString `gorm:"column:COLUMN_DEFAULT"`

	NumericPrecision sql.NullInt64 `gorm:"column:NUMERIC_PRECISION"`
	NumericScale     sql.NullInt64 `gorm:"column:NUMERIC_SCALE"`
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

// MySQLSchema ...
type MySQLSchema struct {
	dbFn          func() (*gorm.DB, error)
	pumpOptionReg *regexp.Regexp
}

var _ model.DbSchema = (*MySQLSchema)(nil)

// CreateMySQLSchema ...
func CreateMySQLSchema(dataSourceName string) (*MySQLSchema, error) {
	dbFn := func() (*gorm.DB, error) {
		return sqlmore.NewSQLMore("mysql", dataSourceName).GormOpen()
	}
	db, err := dbFn()

	if err != nil {
		return nil, err
	}

	defer util.Closeq(db)

	return &MySQLSchema{dbFn: dbFn, pumpOptionReg: regexp.MustCompile(`\bpump:"([^"]+)"`)}, err
}

// Tables ...
func (m MySQLSchema) Tables() ([]model.Table, error) {
	db, err := m.dbFn()
	if err != nil {
		return nil, err
	}

	defer util.Closeq(db)

	var tables []MySQLTable

	db.Raw(`SELECT * FROM information_schema.TABLES WHERE TABLE_SCHEMA = database()`).Find(&tables)

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

	defer util.Closeq(db)

	columns := make([]MyTableColumn, 0)
	s := `SELECT * FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = database() ` +
		`AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION`
	db.Raw(s, table).Find(&columns)

	ts := make([]model.TableColumn, len(columns))

	for i, t := range columns {
		t.Comment = strings.TrimSpace(t.Comment)
		ts[i] = t
	}

	return ts, db.Error
}

// Pump ...
func (m MySQLSchema) Pump(table string, rowsPumped chan<- model.RowsPumped, config model.PumpConfig) error {
	columns, err := m.TableColumns(table)
	if err != nil {
		return err
	}

	columnsValueRand := make(map[string]random.ColumnRandomizer)

	for _, col := range columns {
		columnsValueRand[col.GetName()] = m.makeColumnRandomizer(col.(MyTableColumn))
	}

	columnNames := make([]string, len(columns))

	for i, c := range columns {
		columnNames[i] = c.GetName()
	}

	db, err := m.dbFn()
	if err != nil {
		return err
	}

	defer util.Closeq(db)

	t := time.Now()

	batch := NewInsertBatch(table, columnNames, config.BatchNum, db, func(rows int) {
		rowsPumped <- model.RowsPumped{Table: table, Rows: rows, Cost: time.Since(t)}
		t = time.Now()
	})

	rows := config.RandRows()

	for i := 1; i <= rows; i++ {
		colValues := make([]interface{}, len(columns))

		for j, col := range columns {
			colValues[j] = columnsValueRand[col.GetName()].Value()
		}

		batch.AddRow(colValues)
	}

	batch.Complete()

	return nil
}

func (m MySQLSchema) makeColumnRandomizer(column MyTableColumn) random.ColumnRandomizer {
	sub := m.pumpOptionReg.FindStringSubmatch(column.GetComment())
	pumpOption := ""

	if len(sub) == 2 {
		pumpOption = sub[1]
	}

	if pumpOption != "" {
		return random.NewFn(func() interface{} {
			val, _ := faker.FakeColumnWithType(column.zeroType(), pumpOption)
			return val
		})
	}

	return column.randomColumn()
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

func (c MyTableColumn) randomColumn() random.ColumnRandomizer {
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
