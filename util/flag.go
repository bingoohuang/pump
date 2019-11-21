package util

import (
	"fmt"
	_ "net/http/pprof" // nolint gosec
	"os"
	"strings"

	"github.com/bingoohuang/gou/http"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// InitFlags ...
func InitFlags() {
	help := pflag.BoolP("help", "h", false, "help")
	pflag.StringP("sqls", "s", "", "execute sqls, separated by ;")
	pflag.StringP("fmt", "f", "txt", "query sql execution result printing format(txt/markdown/html/csv)")
	pflag.StringP("ds", "d", "user:pass@tcp(localhost:3306)/db?charset=utf8mb4&parseTime=true&loc=Local", "help")
	pflag.StringP("tables", "t", "", "pump tables, separated by ,")
	pflag.IntP("rows", "r", 1000, "pump rows")
	pflag.IntP("batch", "b", 1000, "batch rows")
	pflag.IntP("sleep", "", 0, "sleep milli-seconds after a batch")
	pflag.IntP("goroutines", "g", 1, "go routines to pump for each table")

	pprofAddr := http.PprofAddrPflag()

	pflag.Parse()

	args := pflag.Args()
	if len(args) > 0 {
		fmt.Printf("Unknown args %s\n", strings.Join(args, " "))
		pflag.PrintDefaults()
		os.Exit(-1)
	}

	if *help {
		fmt.Printf("Built on %s from sha1 %s\n", Compile, Version)
		pflag.PrintDefaults()
		os.Exit(0)
	}

	http.StartPprof(*pprofAddr)

	viper.SetEnvPrefix("PUMP")
	viper.AutomaticEnv()

	// 设置一些配置默认值
	// viper.SetDefault("InfluxAddr", "http://127.0.0.1:8086")
	// viper.SetDefault("CheckIntervalSeconds", 60)

	_ = viper.BindPFlags(pflag.CommandLine)
}
