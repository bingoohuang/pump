package util

import (
	"fmt"
	_ "net/http/pprof" // nolint gosec
	"os"

	"github.com/bingoohuang/gou/cnf"
	"github.com/bingoohuang/gou/htt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// InitFlags ...
func InitFlags() {
	help := pflag.BoolP("help", "h", false, "help")
	pflag.StringP("sqls", "s", "", "execute sqls, separated by ;")
	pflag.BoolP("eval", "e", false, "eval sqls execution in REPL mode")
	pflag.StringP("onerr", "", "retry", "retry on error or not")
	pflag.IntP("retry", "", 0, "retry max times")
	pflag.StringP("fmt", "f", "txt", "query sql execution result printing format(txt/markdown/html/csv)")

	pflag.StringP("ds", "d", "", `eg.
	MYSQL_PWD=8BE4 mysql -h 127.0.0.1 -P 9633 -u root
	mysql -h 127.0.0.1 -P 9633 -u root -p8BE4
	mysql -h 127.0.0.1 -P 9633 -u root -p8BE4 -Dtest
	mysql -h127.0.0.1 -u root -p8BE4 -Dtest
	127.0.0.1:9633 root/8BE4
	127.0.0.1 root/8BE4
	127.0.0.1:9633 root/8BE4 db=test
	root:8BE4@tcp(127.0.0.1:9633)/?charset=utf8mb4&parseTime=true&loc=Local
`)
	pflag.StringP("tables", "t", "", "pump tables, separated by ,")
	pflag.IntP("rows", "r", 1000, "pump total rows")
	pflag.IntP("batch", "b", 1000, "batch rows")
	pflag.StringP("sleep", "", "", "sleep after each batch, eg. 10s (ns/us/µs/ms/s/m/h)")
	pflag.IntP("goroutines", "g", 1, "go routines to pump for each table")
	pflag.IntP("verbose", "V", 0, "verbose details(0 off, 1 abbreviated, 2 full")

	pprofAddr := htt.PprofAddrPflag()

	pflag.Parse()
	cnf.CheckUnknownPFlags()

	if *help {
		fmt.Printf("v1.0.1 Built on %s from sha1 %s\n", Compile, Version)
		pflag.PrintDefaults()
		os.Exit(0)
	}

	htt.StartPprof(*pprofAddr)

	viper.SetEnvPrefix("PUMP")
	viper.AutomaticEnv()
	_ = viper.BindPFlags(pflag.CommandLine)
}
