package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bingoohuang/gou/ran"
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
}

// Accumulate ...
func (p *RowsPumped) Accumulate(r RowsPumped) {
	p.Rows += r.Rows
	p.Cost += r.Cost

	fmt.Printf("%s pumped %d(%.2f%%) rows cost %s/%s\n",
		r.Table, r.Rows, 100.*float32(p.Rows)/float32(p.TotalRows),
		r.Cost.String(), p.Cost.String())
}

// DbSchema ...
type DbSchema interface {
	Tables() ([]Table, error)
	TableColumns(table string) ([]TableColumn, error)
	Pump(table string, rowsPumped chan<- RowsPumped, config PumpConfig) error
}
