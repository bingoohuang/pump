package main

import (
	"bytes"
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/alecthomas/chroma/quick"
	"github.com/bingoohuang/sqlx"
	"github.com/gohxs/readline"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/viper"
)

func dief(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}

	args = append(args, err)

	logrus.Fatalf(format+" error %v", args...)
}

func display(input string) string {
	buf := bytes.NewBuffer([]byte{})
	err := quick.Highlight(buf, input, "mysql", "terminal16m", "monokai")
	dief(err, "quick.Highlight")

	return buf.String()
}

func (a *App) executeSqls() {
	subSqls := sqlx.SplitSqls(viper.GetString("sqls"), ';')
	eval := viper.GetBool("eval")

	if len(subSqls) == 0 && !eval {
		return
	}

	defer os.Exit(0)

	db := sqlx.NewSQLMore("mysql", a.schema.CompatibleDs()).MustOpen()
	//db, _ := sql.Open("mysql", a.schema.CompatibleDs())
	defer db.Close()

	if len(subSqls) > 0 {
		a.executeSQLs(db, subSqls)
	}

	if !eval {
		return
	}

	term, err := readline.NewEx(&readline.Config{Prompt: "MySQL> ", Output: display, HistoryFile: ".pump.sql"})
	dief(err, "readline.NewEx")

	for {
		line, err := term.Readline()
		dief(err, "term.Readline")

		text := strings.TrimSpace(line)
		if text != "" {
			subSqls = sqlx.SplitSqls(text, ';')
		}

		a.executeSQLs(db, subSqls)
	}
}

func (a *App) executeSQLs(db *sql.DB, subSqls []string) {
	tx, err := db.Begin()
	if err != nil {
		logrus.Panicf("failed to begin %v", err)
	}

	defer func() { _ = tx.Commit() }() // nolint errcheck

	for _, s := range subSqls {
		r := sqlx.ExecSQL(tx, s, 3000, "NULL")
		a.printResult(s, r)
	}
}

func (a *App) printResult(s string, r sqlx.ExecResult) {
	if r.Error != nil {
		logrus.Panicf("error occurred %v", r.Error)
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
