package main

import (
	"log"
	"os"

	"github.com/bingoohuang/sqlmore"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/viper"
)

func (a *App) executeSqls() {
	subSqls := sqlmore.SplitSqls(viper.GetString("sqls"), ';')

	if len(subSqls) == 0 {
		return
	}

	db := sqlmore.NewSQLMore("mysql", a.schema.CompatibleDs()).MustOpen()

	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatalln(err)
	}

	for _, s := range subSqls {
		r := sqlmore.ExecSQL(tx, s, 3000, "NULL")
		a.printResult(s, r)
	}

	tx.Commit() // nolint errcheck
	os.Exit(0)
}

func (a *App) printResult(s string, r sqlmore.ExecResult) {
	if r.Error != nil {
		log.Println(r.Error)
		return
	}

	log.Printf("SQL: %s\n", s)
	log.Printf("cost: %s\n", r.CostTime.String())

	if !r.IsQuerySQL {
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	cols := len(r.Headers) + 1
	header := make(table.Row, cols)
	header[0] = "#"

	for i, h := range r.Headers {
		header[i+1] = h
	}

	t.AppendHeader(header)

	for i, r := range r.Rows {
		row := make(table.Row, cols)
		row[0] = i + 1

		for j, c := range r {
			row[j+1] = c
		}

		t.AppendRow(row)
	}

	fmt := viper.GetString("fmt")
	switch fmt {
	case "csv":
		t.RenderCSV()
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
}
