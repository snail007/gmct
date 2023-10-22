package main

import (
	"fmt"
	gcore "github.com/snail007/gmc/core"
	glog "github.com/snail007/gmc/module/log"
	_ "github.com/snail007/gmct/module"
	"github.com/snail007/gmct/module/module"
	"github.com/snail007/gmct/tool"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	tool.Version = version
	if len(os.Args) == 2 && (os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Println(version)
		os.Exit(0)
	}
	glog.SetFlag(gcore.LogFlagNormal)
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "toolchain for go web framework gmc, https://github.com/snail007/gmc",
		Version: version,
	}
	module.Init(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		glog.Fatal(err)
	}
}
