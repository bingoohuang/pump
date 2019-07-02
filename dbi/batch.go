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
	batchOp     func(int)
}

func NewInsertBatch(table string, columnNames []string,
	batchNum int, db *gorm.DB, batchOp func(int)) *InsertBatcher {
	b := &InsertBatcher{batchNum: batchNum, db: db, columnCount: len(columnNames)}
	b.rows = make([]interface{}, 0, b.batchNum*b.columnCount)

	bind := "(" + util.Repeat("?", ",", b.columnCount) + ")"
	sql := "insert into " + table + "(" + strings.Join(columnNames, ",") + ") values"
	b.batchSQL = sql + util.Repeat(bind, ",", batchNum)
	b.completeSQL = func() string { return sql + util.Repeat(bind, ",", b.rowsCount) }
	b.batchOp = batchOp
	return b
}

func (b *InsertBatcher) AddRow(colValues []interface{}) {
	b.rowsCount++
	b.rows = append(b.rows, colValues...)

	if b.rowsCount == b.batchNum {
		b.executeBatch(b.batchSQL)
	}
}

func (b *InsertBatcher) Complete() int {
	left := b.rowsCount
	if left > 0 {
		b.executeBatch(b.completeSQL())
	}

	return left
}

func (b *InsertBatcher) executeBatch(sql string) {
	b.db.Exec(sql, b.rows...)
	b.batchOp(b.rowsCount)
	b.rowsCount = 0
	b.rows = b.rows[0:0]
}
