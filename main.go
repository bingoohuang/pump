package main

import (
	"os"
	"strconv"

	"github.com/bingoohuang/pump/dbi"
	_ "github.com/bingoohuang/pump/dbi"
	"github.com/bingoohuang/pump/model"
)

func main() {
	// https://github.com/bingoohuang/StarManager/blob/main/db.go
	pumpDataSource := os.Getenv("PUMP_DATA_SOURCE")
	pumpTable := os.Getenv("PUMP_TABLE")
	pumpRows, _ := strconv.Atoi(os.Getenv("PUMP_ROWS"))
	schema, err := dbi.CreateMySQLSchema(pumpDataSource)
	if err != nil {
		panic(err)
	}

	//tables, err := schema.Tables()
	//if err != nil {
	//	panic(err)
	//}
	//
	//for _, table := range tables {
	//	fmt.Printf("table：%+v\n", table)
	//}

	//columns, err := schema.TableColumns("sc_ecdocument")
	//if err != nil {
	//	panic(err)
	//}
	//
	//for _, tableColumn := range columns {
	//	fmt.Printf("co.umn：%+v\n", tableColumn)
	//}

	err = schema.Pump(pumpTable, model.PumpConfig{PumpMinRows: pumpRows, PumpMaxRows: pumpRows})
	if err != nil {
		panic(err)
	}
}
