package util

import (
	"fmt"
	gcast "github.com/snail007/gmc/util/cast"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type MustValue struct {
	err          error
	value        interface{}
	defaultValue interface{}
}

func (s *MustValue) getVal() interface{} {
	if s.err == nil {
		return s.value
	}
	if s.defaultValue != nil {
		return s.defaultValue
	}
	return nil
}

func (s *MustValue) Default(defaultValue interface{}) *MustValue {
	s.defaultValue = defaultValue
	return s
}

func (s *MustValue) Bool() bool {
	return gcast.ToBool(s.value)
}

func (s *MustValue) String() string {
	return gcast.ToString(s.value)
}

func (s *MustValue) StringSlice() []string {
	return gcast.ToStringSlice(s.value)
}

func (s *MustValue) Int() int {
	return gcast.ToInt(s.value)
}

func (s *MustValue) Int32() int32 {
	return gcast.ToInt32(s.value)
}

func (s *MustValue) Int64() int64 {
	return gcast.ToInt64(s.value)
}

func (s *MustValue) Uint() uint {
	return gcast.ToUint(s.value)
}

func (s *MustValue) Uint8() uint8 {
	return gcast.ToUint8(s.value)
}

func (s *MustValue) Uint32() uint32 {
	return gcast.ToUint32(s.value)
}

func (s *MustValue) Uin64t() uint64 {
	return gcast.ToUint64(s.value)
}

func (s *MustValue) Float32() float32 {
	return gcast.ToFloat32(s.value)
}

func (s *MustValue) Float64() float64 {
	return gcast.ToFloat64(s.value)
}

func (s *MustValue) Value() any {
	return s.getVal()
}

func Must(value interface{}, err error) *MustValue {
	return &MustValue{value: value, err: err}
}

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
