package tool

import (
	"os"
)

var (
	Version string
	Args    []string
)

type GMCTool interface {
	Start(args interface{}) (err error)
	Stop()
}

func init() {
	// catch arguments after --
	if len(os.Args) > 2 {
		for i, v := range os.Args {
			if v == "--" {
				if len(os.Args) > i+1 {
					Args = os.Args[i+1:]
					os.Args = os.Args[:i]
				}
				break
			}
		}
	}
}
