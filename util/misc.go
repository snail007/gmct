package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func GetPackageName(dir string) (pkgName string) {
	dir, _ = filepath.Abs(dir)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (e error) {
		if err != nil {
			return
		}
		if info.IsDir() {
			return
		}
		ext := filepath.Ext(info.Name())
		name := strings.Trim(info.Name(), ext)
		if ext != ".go" || strings.HasSuffix(name, "_test") {
			return
		}
		contents, _ := ioutil.ReadFile(path)
		exp := regexp.MustCompile(`package +([^ \n]+) *\n`)
		m := exp.FindStringSubmatch(string(contents))
		if len(m) >= 1 {
			pkgName = m[1]
			return fmt.Errorf("")
		}
		return
	})
	if pkgName == "" {
		return filepath.Base(dir)
	}
	return
}

func Exists(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = f.Stat()
	if err != nil {
		return false
	}
	return true
}

func ExistsFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return !st.IsDir()
}

func ExistsDir(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.IsDir()
}

func IsEmptyDir(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true
	}
	return false
}
