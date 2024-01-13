package module

import (
	"github.com/spf13/cobra"
	"sync"
)

type InitFunc func(root *cobra.Command)

var modules []InitFunc

var l sync.Mutex

func AddCommand(f InitFunc) {
	l.Lock()
	defer l.Unlock()
	modules = append(modules, f)
}

func Init(root *cobra.Command) {
	for _, f := range modules {
		f(root)
	}
}
