package model

import (
	"database/sql"

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

type DbSchema interface {
	Tables() ([]Table, error)
	TableColumns(table string) ([]TableColumn, error)
	Pump(table string, config PumpConfig) error
}
