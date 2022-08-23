package dbi

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bingoohuang/pump/util"

	"github.com/bingoohuang/gou/str"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

// Batcher ...
type Batcher interface {
	// GetBatchNum returns the batch num
	GetBatchNum() int
	// AddRow adds a row to batcher and execute when rows accumulated to the batch num.
	AddRow(colValues []interface{}) error
	// Complete completes the left rows that less than batch num.
	Complete() (int, error)
}

// InsertBatcher ...
type InsertBatcher struct {
	batchNum    int
	columnCount int

	rowsCount     int
	batchExecuted int
	rows          []interface{}
	db            *sql.DB
	batchSQL      string
	completeSQL   func() string
	batchOp       func(int)

	sleep time.Duration

	verbose int
}

// NewInsertBatch ...
func NewInsertBatch(table string, columnNames []string, batchNum int, db *sql.DB,
	batchOp func(int), verbose, rows int,
) *InsertBatcher {
	b := &InsertBatcher{batchNum: batchNum, db: db, columnCount: len(columnNames)}
	b.rows = make([]interface{}, 0, b.batchNum*b.columnCount)

	bind := "(" + str.Repeat("?", ",", b.columnCount) + ")"
	s := "insert into " + table + "(" + strings.Join(columnNames, ",") + ") values"
	b.batchSQL = s + str.Repeat(bind, ",", batchNum)

	if verbose > 0 && batchNum >= rows {
		logrus.Infof("batchSQL:%s", util.Abbr(b.batchSQL, verbose, 500))
	}

	b.completeSQL = func() string { return s + str.Repeat(bind, ",", b.rowsCount) }
	b.batchOp = batchOp
	b.verbose = verbose

	b.setSleepDuration()

	return b
}

// GetBatchNum returns the batch num.
func (b InsertBatcher) GetBatchNum() int { return b.batchNum }

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

// AddRow adds a row to batcher and execute when rows accumulated to the batch num.
func (b *InsertBatcher) AddRow(colValues []interface{}) error {
	b.rowsCount++
	b.rows = append(b.rows, colValues...)

	if b.rowsCount == b.batchNum {
		if err := b.executeBatch(b.batchSQL); err != nil {
			return err
		}
	}

	return nil
}

// Complete completes the left rows that less than batch num.
func (b *InsertBatcher) Complete() (int, error) {
	if b.rowsCount <= 0 {
		return 0, nil
	}

	if err := b.executeBatch(b.completeSQL()); err != nil {
		return 0, err
	}

	return b.rowsCount, nil
}

func (b *InsertBatcher) executeBatch(sql string) error {
	if b.batchExecuted > 0 && b.sleep > 0 {
		time.Sleep(b.sleep)
	}

	if b.verbose > 0 {
		logrus.Info(util.Abbr(fmt.Sprintf("values:%v", b.rows), b.verbose, 500))
	}

	if _, err := b.db.Exec(sql, b.rows...); err != nil {
		b.resetBatcherRows()

		return err
	}

	b.batchExecuted++
	b.batchOp(b.rowsCount)
	b.resetBatcherRows()

	return nil
}

func (b *InsertBatcher) resetBatcherRows() {
	b.rowsCount = 0
	b.rows = b.rows[0:0]
}
