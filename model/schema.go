package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bingoohuang/gou/ran"
	"github.com/gosuri/uiprogress"
)

// Table abstract a table information.
type Table interface {
	GetName() string
	GetComment() string
}

// TableColumn ...
type TableColumn interface {
	GetName() string
	GetComment() string
	GetType() string
	GetDataType() string
	GetMaxSize() sql.NullInt64
	IsAllowNull() bool
	GetColumnRandomizer() ColumnRandomizer
}

// PumpColumnConfig  ...
type PumpColumnConfig struct {
}

// PumpConfig ...
type PumpConfig struct {
	ColumnsConfig map[string]PumpColumnConfig
	PumpMinRows   int
	PumpMaxRows   int
	BatchNum      int
}

// RandRows ...
func (c PumpConfig) RandRows() int {
	min := c.PumpMinRows
	if min <= 0 {
		min = ran.Int()
	}

	max := c.PumpMaxRows
	if max <= 0 {
		max = min + ran.Int()
	}

	if min >= max {
		min = max
	}

	return ran.IntN(uint64(max-min)) + min
}

// RowsPumped ...
type RowsPumped struct {
	Table     string
	TotalRows int
	Rows      int
	Cost      time.Duration

	bar *uiprogress.Bar
}

// MakeRowsPumped makes a new RowsPumped
func MakeRowsPumped(pumpTable string, totalRows int) *RowsPumped {
	start := time.Now()
	bar := uiprogress.AddBar(totalRows).AppendCompleted().PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("%s %d/%d %v", pumpTable, b.Current(), totalRows, time.Since(start))
	})

	return &RowsPumped{
		Table:     pumpTable,
		TotalRows: totalRows,
		bar:       bar,
	}
}

// Accumulate ...
func (p *RowsPumped) Accumulate(r RowsPumped) {
	p.Rows += r.Rows
	p.Cost += r.Cost

	_ = p.bar.Set(p.Rows)
}

// DbSchema ...
type DbSchema interface {
	Tables() ([]Table, error)
	TableColumns(table string) ([]TableColumn, error)
	// CompatibleDs returns the dataSourceName from various the compatible format.
	CompatibleDs() string
	Pump(table string, rowsPumped chan<- RowsPumped, config PumpConfig, ready chan bool) error
}
