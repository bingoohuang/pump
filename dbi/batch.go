package dbi

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Masterminds/goutils"

	"github.com/bingoohuang/gou/str"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"

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

	verbose bool
}

// NewInsertBatch ...
func NewInsertBatch(table string, columnNames []string,
	batchNum int, db *gorm.DB, batchOp func(int), verbose bool) Batcher {
	b := &InsertBatcher{batchNum: batchNum, db: db, columnCount: len(columnNames)}
	b.rows = make([]interface{}, 0, b.batchNum*b.columnCount)

	bind := "(" + str.Repeat("?", ",", b.columnCount) + ")"
	sql := "insert into " + table + "(" + strings.Join(columnNames, ",") + ") values"
	b.batchSQL = sql + str.Repeat(bind, ",", batchNum)
	logrus.Infof("batchSQL:%s", b.batchSQL)
	b.completeSQL = func() string { return sql + str.Repeat(bind, ",", b.rowsCount) }
	b.batchOp = batchOp
	b.verbose = verbose

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
		completeSQL := b.completeSQL()

		b.executeBatch(completeSQL)
	}

	return left
}

func (b *InsertBatcher) executeBatch(sql string) {
	if b.batchExecuted > 0 && b.sleep > 0 {
		time.Sleep(b.sleep)
	}

	if b.verbose {
		s := fmt.Sprintf("values:%v", b.rows)
		abbr, _ := goutils.Abbreviate(s, 500)
		logrus.Info(abbr)
	}

	b.db.Exec(sql, b.rows...)
	b.batchExecuted++
	b.batchOp(b.rowsCount)
	b.rowsCount = 0
	b.rows = b.rows[0:0]
}
