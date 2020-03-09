package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gou/str"
	"github.com/bingoohuang/pump/util"
	"github.com/gosuri/uiprogress"
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

	pumpedRows    chan model.RowsPumped
	batchNum      int
	onErr         string
	retryMaxTimes int
}

// MakeApp ...
func MakeApp() *App {
	schema, err := dbi.CreateMySQLSchema(viper.GetString("ds"))
	if err != nil {
		logrus.Panicf("failed to create datasource, error %v", err)
	}

	a := &App{
		pumpTables:    str.SplitTrim(viper.GetString("tables"), ","),
		pumpRoutines:  viper.GetInt("goroutines"),
		schema:        schema,
		totalRows:     viper.GetInt("rows"),
		batchNum:      viper.GetInt("batch"),
		onErr:         viper.GetString("onerr"),
		retryMaxTimes: viper.GetInt("retry"),
	}
	a.pumpedRows = make(chan model.RowsPumped, len(a.pumpTables)*a.pumpRoutines)

	schema.SetVerbose(viper.GetInt("verbose"))

	return a
}

func (a *App) pumpingTables() {
	if len(a.pumpTables) == 0 {
		fmt.Println("no tables to pump")
		os.Exit(0)
	}

	rows := make(map[string]*model.RowsPumped)
	complete := make(map[string]bool)

	ready := make(chan bool)

	for _, pumpTable := range a.pumpTables {
		rows[pumpTable] = model.MakeRowsPumped(pumpTable, a.totalRows)
		complete[pumpTable] = false

		a.pumpTable(pumpTable, ready)
	}

	for range a.pumpTables {
		for i := 0; i < a.pumpRoutines; i++ {
			<-ready
		}
	}

	uiprogress.Start()
	defer uiprogress.Stop()

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

func (a *App) pumpTable(table string, ready chan bool) {
	routineRows0, routineRows := a.routineRows()

	go a.pump(table, routineRows0, ready)

	for i := 1; i < a.pumpRoutines; i++ {
		go a.pump(table, routineRows, ready)
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

func (a *App) pump(pumpTable string, rows int, ready chan bool) {
	c := model.PumpConfig{PumpMinRows: rows, PumpMaxRows: rows, BatchNum: a.batchNum}
	if err := a.schema.Pump(pumpTable, a.pumpedRows, c, ready, a.onErr, a.retryMaxTimes); err != nil {
		logrus.Panicf("pump error %v", err)
	}
}
