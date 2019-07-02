package util

import (
	"fmt"
	// pprof debug
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/bingoohuang/gou"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func InitFlags() {
	help := pflag.BoolP("help", "h", false, "help")
	pflag.StringP("datasource", "d", "user:pass@tcp(localhost:3306)/db?charset=utf8mb4&parseTime=true&loc=Local", "help")
	pflag.StringP("tables", "t", "", "pump tables")
	pflag.IntP("rows", "r", 1000, "pump rows")
	pflag.IntP("batch", "b", 1000, "batch rows")
	pflag.IntP("goroutines", "g", 3, "go routines to pump for each table")
	pprofAddr := gou.PprofAddrPflag()

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

	gou.StartPprof(*pprofAddr)

	viper.SetEnvPrefix("PUMP")
	viper.AutomaticEnv()

	// 设置一些配置默认值
	// viper.SetDefault("InfluxAddr", "http://127.0.0.1:8086")
	// viper.SetDefault("CheckIntervalSeconds", 60)

	_ = viper.BindPFlags(pflag.CommandLine)
}
