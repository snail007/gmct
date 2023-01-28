package config

import (
	_ "embed"
	gcore "github.com/snail007/gmc/core"
	gconfig "github.com/snail007/gmc/module/config"
	glog "github.com/snail007/gmc/module/log"
	gfile "github.com/snail007/gmc/util/file"
	"os"
	"path/filepath"
)

var (
	Options gcore.Config
	//go:embed config.sample.toml
	tpl []byte
)

func init() {
	file := filepath.Join(gfile.HomeDir(), ".gmct/options")
	if v := os.Getenv("OPTIONS"); v != "" {
		file = v
	}
	if !gfile.Exists(file) {
		os.Mkdir(filepath.Dir(file), 700)
		gfile.Write(file, tpl, false)
	}
	Options = gconfig.New()
	Options.SetConfigFile(file)
	Options.SetConfigType("toml")
	err := Options.ReadInConfig()
	if err != nil {
		glog.Errorf("parse config file [%s] error: %s\nplease remove it or modify it correctly", file, err)
	}
}
