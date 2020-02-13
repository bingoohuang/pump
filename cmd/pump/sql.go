package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bingoohuang/sqlmore"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/viper"
)

func (a *App) executeSqls() {
	subSqls := sqlmore.SplitSqls(viper.GetString("sqls"), ';')
	eval := viper.GetBool("eval")

	if len(subSqls) == 0 && !eval {
		return
	}

	defer os.Exit(0)

	db := sqlmore.NewSQLMore("mysql", a.schema.CompatibleDs()).MustOpen()
	//db, _ := sql.Open("mysql", a.schema.CompatibleDs())
	defer db.Close()

	if len(subSqls) > 0 {
		a.executeSQLs(db, subSqls)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter your sql (empty to re-execute) :")
		scanner.Scan()
		text := strings.TrimSpace(scanner.Text())

		if text != "" {
			subSqls = sqlmore.SplitSqls(text, ';')
		}

		a.executeSQLs(db, subSqls)
	}
}

func (a *App) executeSQLs(db *sql.DB, subSqls []string) {
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return
	}

	defer func() { _ = tx.Commit() }() // nolint errcheck

	for _, s := range subSqls {
		r := sqlmore.ExecSQL(tx, s, 3000, "NULL")
		a.printResult(s, r)
	}
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

	cols := len(r.Headers) + 1 // nolint gomnd
	header := make(table.Row, cols)
	header[0] = "#"

	for i, h := range r.Headers {
		header[i+1] = h
	}

	t.AppendHeader(header)

	for i, r := range r.Rows {
		row := make(table.Row, cols)
		row[0] = i + 1 // nolint gomnd

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
