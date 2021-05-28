package goinstall

import (
	"embed"
	gfile "github.com/snail007/gmc/util/file"
)

var (
	Scripts = map[string]string{}
	//go:embed *.sh
	scripts embed.FS
)

func init() {
	fs, err := scripts.ReadDir(".")
	if err != nil {
		panic(err)
	}
	for _, v := range fs {
		var b []byte
		b, err = scripts.ReadFile(v.Name())
		if err != nil {
			panic(err)
		}
		k := gfile.FileName(v.Name())
		Scripts[k] = string(b)
	}
}
