package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bingoohuang/pump/durafmt"

	"github.com/bingoohuang/gou"
)

type Table interface {
	GetName() string
	GetComment() string
}

type TableColumn interface {
	GetName() string
	GetComment() string
	GetType() string
	GetDataType() string
	GetMaxSize() sql.NullInt64
	IsAllowNull() bool
}

type PumpColumnConfig struct {
}

type PumpConfig struct {
	ColumnsConfig map[string]PumpColumnConfig
	PumpMinRows   int
	PumpMaxRows   int
	BatchNum      int
}

func (c PumpConfig) RandRows() int {
	min := c.PumpMinRows
	if min <= 0 {
		min = gou.RandomInt()
	}

	max := c.PumpMaxRows
	if max <= 0 {
		max = min + gou.RandomInt()
	}

	if min >= max {
		min = max
	}

	return gou.RandomIntN(uint64(max-min)) + min
}

type RowsPumped struct {
	Table     string
	TotalRows int
	Rows      int
	Cost      time.Duration
}

func (p *RowsPumped) Accumulate(r RowsPumped) {
	p.Rows += r.Rows
	p.Cost += r.Cost

	fmt.Printf("%s pumped %d(%.2f%%) rows cost %s/%s\n",
		r.Table, r.Rows, 100.*float32(p.Rows)/float32(p.TotalRows),
		durafmt.Format(r.Cost), durafmt.Format(p.Cost))
}

type DbSchema interface {
	Tables() ([]Table, error)
	TableColumns(table string) ([]TableColumn, error)
	Pump(table string, rowsPumped chan RowsPumped, config PumpConfig) error
}
