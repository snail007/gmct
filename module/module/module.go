package module

import (
	"github.com/spf13/cobra"
)

type InitFunc func(root *cobra.Command)

var modules []InitFunc

func AddCommand(f InitFunc) {
	modules = append(modules, f)
}

func Init(root *cobra.Command) {
	for _, f := range modules {
		f(root)
	}
}
