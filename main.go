package main

import (
	"strings"

	"github.com/bingoohuang/pump/util"
	"github.com/spf13/viper"

	"github.com/bingoohuang/pump/dbi"
	_ "github.com/bingoohuang/pump/dbi"
	"github.com/bingoohuang/pump/model"
)

func main() {
	util.InitFlags()

	app := MakeApp()
	app.pumpingTables()
}

type App struct {
	pumpTables   []string
	pumpRoutines int
	totalRows    int

	schema model.DbSchema

	pumpedRows chan model.RowsPumped
	batchNum   int
}

func MakeApp() *App {
	pumpDataSource := viper.GetString("datasource")
	schema, err := dbi.CreateMySQLSchema(pumpDataSource)
	if err != nil {
		panic(err)
	}

	totalRows := viper.GetInt("rows")
	a := &App{
		pumpTables:   strings.Split(viper.GetString("tables"), ","),
		pumpRoutines: viper.GetInt("goroutines"),
		schema:       schema,
		totalRows:    totalRows,
		batchNum:     viper.GetInt("batch"),
	}
	a.pumpedRows = make(chan model.RowsPumped, len(a.pumpTables)*a.pumpRoutines)

	return a
}

func (a *App) pumpingTables() {
	rows := make(map[string]*model.RowsPumped)
	complete := make(map[string]bool)

	for _, pumpTable := range a.pumpTables {
		rows[pumpTable] = &model.RowsPumped{Table: pumpTable, TotalRows: a.totalRows}
		complete[pumpTable] = false

		routineRows0 := 0
		routineRows := 0
		if a.totalRows < a.pumpRoutines {
			routineRows = a.totalRows
			routineRows0 = a.totalRows
			a.pumpRoutines = 1
		} else {
			routineRows = a.totalRows / a.pumpRoutines
			routineRows0 = routineRows + a.totalRows - routineRows*a.pumpRoutines
		}

		for i := 0; i < a.pumpRoutines; i++ {
			rows := routineRows
			if i == 0 {
				rows = routineRows0
			}
			go a.pump(pumpTable, rows)
		}
	}

	for r := range a.pumpedRows {
		pumped := rows[r.Table]
		pumped.Accumulate(r)

		if pumped.Rows == a.totalRows {
			delete(rows, r.Table)
			if len(rows) == 0 {
				break
			}
		}
	}
}

func (a *App) pump(pumpTable string, rows int) {
	config := model.PumpConfig{PumpMinRows: rows, PumpMaxRows: rows, BatchNum: a.batchNum}
	if err := a.schema.Pump(pumpTable, a.pumpedRows, config); err != nil {
		panic(err)
	}
}
