package dbi

import (
	"strings"

	"github.com/bingoohuang/pump/util"
	"github.com/jinzhu/gorm"
)

type InsertBatcher struct {
	batchNum    int
	columnCount int

	rowsCount   int
	rows        []interface{}
	db          *gorm.DB
	batchSQL    string
	completeSQL func() string
}

func NewInsertBatch(table string, columnNames []string, batchNum int, db *gorm.DB) *InsertBatcher {
	b := &InsertBatcher{batchNum: batchNum, db: db}

	b.columnCount = len(columnNames)
	mark := make([]string, b.columnCount)
	for i := range columnNames {
		mark[i] = "?"
	}

	bind := "(" + strings.Join(mark, ",") + ")"
	b.clearBatch()

	sql := "insert into " + table + "(" + strings.Join(columnNames, ",") + ") values"
	b.batchSQL = sql + util.Repeat(bind, ",", batchNum)
	b.completeSQL = func() string { return sql + util.Repeat(bind, ",", b.rowsCount) }
	return b
}

func (b *InsertBatcher) Add(colValues []interface{}) {
	b.rowsCount++
	b.rows = append(b.rows, colValues...)

	if b.rowsCount == b.batchNum {
		b.executeBatch(b.batchSQL)
	}
}

func (b *InsertBatcher) Complete() {
	if b.rowsCount > 0 {
		b.executeBatch(b.completeSQL())
	}
}

func (b *InsertBatcher) executeBatch(sql string) {
	b.db.Exec(sql, b.rows...)

	b.clearBatch()
}

func (b *InsertBatcher) clearBatch() {
	b.rowsCount = 0
	b.rows = make([]interface{}, 0, b.batchNum*b.columnCount)
}
