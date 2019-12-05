package dbi

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/bingoohuang/pump/util"
	"github.com/jinzhu/gorm"
)

// Batcher ...
type Batcher interface {
	AddRow(colValues []interface{})
	Complete() int
}

// InsertBatcher ...
type InsertBatcher struct {
	batchNum    int
	columnCount int

	rowsCount     int
	batchExecuted int
	rows          []interface{}
	db            *gorm.DB
	batchSQL      string
	completeSQL   func() string
	batchOp       func(int)

	sleep time.Duration
}

// NewInsertBatch ...
func NewInsertBatch(table string, columnNames []string,
	batchNum int, db *gorm.DB, batchOp func(int)) Batcher {
	b := &InsertBatcher{batchNum: batchNum, db: db, columnCount: len(columnNames)}
	b.rows = make([]interface{}, 0, b.batchNum*b.columnCount)

	bind := "(" + util.Repeat("?", ",", b.columnCount) + ")"
	sql := "insert into " + table + "(" + strings.Join(columnNames, ",") + ") values"
	b.batchSQL = sql + util.Repeat(bind, ",", batchNum)
	b.completeSQL = func() string { return sql + util.Repeat(bind, ",", b.rowsCount) }
	b.batchOp = batchOp

	b.setSleepDuration()

	return b
}

func (b *InsertBatcher) setSleepDuration() {
	sleepDuration := viper.GetString("sleep")
	if sleepDuration == "" {
		return
	}

	var err error
	b.sleep, err = time.ParseDuration(sleepDuration)

	if err != nil {
		log.Panicf("fail to parse sleep %s, error %v", sleepDuration, err)
	}
}

// AddRow ...
func (b *InsertBatcher) AddRow(colValues []interface{}) {
	b.rowsCount++
	b.rows = append(b.rows, colValues...)

	if b.rowsCount == b.batchNum {
		b.executeBatch(b.batchSQL)
	}
}

// Complete ...
func (b *InsertBatcher) Complete() int {
	left := b.rowsCount
	if left > 0 {
		b.executeBatch(b.completeSQL())
	}

	return left
}

func (b *InsertBatcher) executeBatch(sql string) {
	if b.batchExecuted > 0 && b.sleep > 0 {
		time.Sleep(b.sleep)
	}

	b.db.Exec(sql, b.rows...)
	b.batchExecuted++
	b.batchOp(b.rowsCount)
	b.rowsCount = 0
	b.rows = b.rows[0:0]
}
