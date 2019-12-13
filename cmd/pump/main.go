package main

import (
	"github.com/bingoohuang/gou/str"
	"github.com/bingoohuang/pump/util"
	"github.com/spf13/viper"

	"github.com/bingoohuang/pump/dbi"
	_ "github.com/bingoohuang/pump/dbi"
	"github.com/bingoohuang/pump/model"
)

func main() {
	util.InitFlags()

	app := MakeApp()
	app.executeSqls()
	app.pumpingTables()
}

// App ...
type App struct {
	pumpTables   []string
	pumpRoutines int
	totalRows    int

	schema model.DbSchema

	pumpedRows chan model.RowsPumped
	batchNum   int
}

// MakeApp ...
func MakeApp() *App {
	schema, err := dbi.CreateMySQLSchema(viper.GetString("ds"))
	if err != nil {
		panic(err)
	}

	totalRows := viper.GetInt("rows")
	a := &App{
		pumpTables:   str.SplitTrim(viper.GetString("tables"), ","),
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

		a.pumpTable(pumpTable)
	}

	for r := range a.pumpedRows {
		pumped := rows[r.Table]
		pumped.Accumulate(r)

		if pumped.Rows == a.totalRows {
			delete(rows, r.Table)
		}

		if len(rows) == 0 {
			break
		}
	}
}

func (a *App) pumpTable(table string) {
	routineRows0, routineRows := a.routineRows()

	go a.pump(table, routineRows0)

	for i := 1; i < a.pumpRoutines; i++ {
		go a.pump(table, routineRows)
	}
}

func (a *App) routineRows() (routineRows0, routineRows int) {
	if a.totalRows < a.pumpRoutines {
		routineRows = a.totalRows
		routineRows0 = a.totalRows
		a.pumpRoutines = 1
	} else {
		routineRows = a.totalRows / a.pumpRoutines
		routineRows0 = routineRows + a.totalRows - routineRows*a.pumpRoutines
	}

	return routineRows0, routineRows
}

func (a *App) pump(pumpTable string, rows int) {
	c := model.PumpConfig{PumpMinRows: rows, PumpMaxRows: rows, BatchNum: a.batchNum}
	if err := a.schema.Pump(pumpTable, a.pumpedRows, c); err != nil {
		panic(err)
	}
}
