package static

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/snail007/gmct/tool"
	"github.com/snail007/gmct/util"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StaticArgs struct {
	Dir              *string
	GoFilenamePrefix string
	GoFilename       string
	Clean            *bool
	NotExtension     *string
	SubName          *string
}

func NewStaticArgs() StaticArgs {
	return StaticArgs{
		Dir:          new(string),
		NotExtension: new(string),
		Clean:        new(bool),
		SubName:      new(string),
	}
}

type Static struct {
	tool.GMCTool
	args    StaticArgs
	notExts map[string]bool
}

func NewStatic() *Static {
	return &Static{
		notExts: map[string]bool{},
	}
}

func (s *Static) init(args0 interface{}) (err error) {
	s.args = args0.(StaticArgs)
	if *s.args.Dir == "" {
		return fmt.Errorf("static directory not exists")
	}
	//convert pack path to absoulte path of *nix style
	*s.args.Dir, err = filepath.Abs(*s.args.Dir)
	if err != nil {
		return
	}
	*s.args.Dir = strings.Replace(*s.args.Dir, "\\", "/", -1)
	*s.args.Dir = filepath.Join(*s.args.Dir, "/")

	for _, ext := range strings.Split(*s.args.NotExtension, ",") {
		s.notExts[ext] = true
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	s.args.GoFilenamePrefix = "gmc_static_bindata_"
	s.args.GoFilename = s.args.GoFilenamePrefix + fmt.Sprintf("%d", r.Int63()) + ".go"
	return
}

func (s *Static) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	if *s.args.Clean {
		s.clean()
	} else {
		s.clean()
		err = s.pack()
	}
	return
}

func (s *Static) Stop() {

	return
}
func (s *Static) pack() (err error) {
	names := []string{}
	err = s.tree(*s.args.Dir, &names)
	if err != nil {
		return
	}
	filenames := []string{}
	for _, v := range names {
		filename := filepath.Base(v)
		ext := filepath.Ext(filename)
		if s.notExts[ext] {
			continue
		}
		filenames = append(filenames, v)
	}
	var buf bytes.Buffer
	for _, v := range filenames {
		b, err := ioutil.ReadFile(filepath.Join(*s.args.Dir, v))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		str := base64.StdEncoding.EncodeToString(b)
		buf.WriteString(fmt.Sprintf("\t\t\"%s\" : \"%s\",\n", v, str))
	}
	currentDir, _ := filepath.Abs(".")
	packageName := util.GetPackageName(currentDir)
	file := fmt.Sprintf(`package %s

import gmchttpserver "github.com/snail007/gmc/http/server"

func init(){
	gmchttpserver.SetBinData(map[string]string{

{{HOLDER}}
	})
}
`, packageName)
	file = strings.Replace(file, "{{HOLDER}}", buf.String(), 1)
	ioutil.WriteFile(s.args.GoFilename, []byte(file), 0755)
	return
}
func (s *Static) clean() {
	files, err := filepath.Glob(s.args.GoFilenamePrefix + "*.go")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, v := range files {
		filename := filepath.Base(v)
		if strings.HasPrefix(filename, s.args.GoFilenamePrefix) &&
			strings.HasSuffix(filename, ".go") {
			err = os.Remove(v)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
	return
}
func (s *Static) tree(folder string, names *[]string) (err error) {
	name := filepath.Base(folder)
	if strings.HasPrefix(name, ".") {
		return
	}
	f, err := os.Open(folder)
	if err != nil {
		return
	}
	defer f.Close()
	finfo, err := f.Stat()
	if err != nil {
		return
	}
	if !finfo.IsDir() {
		return
	}
	files, err := filepath.Glob(folder + "/*")
	if err != nil {
		return
	}
	var file *os.File
	for _, v := range files {
		if file != nil {
			file.Close()
		}
		file, err = os.Open(v)
		if err != nil {
			return
		}
		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			err = s.tree(v, names)
			if err != nil {
				return err
			}
		} else {
			v, err = filepath.Abs(v)
			if err != nil {
				return err
			}
			if strings.HasPrefix(filepath.Base(v), ".") {
				continue
			}
			v0 := strings.Replace(v, "\\", "/", -1)
			v0 = strings.Replace(v0, *s.args.Dir+"/", "", -1)
			*names = append(*names, v0)
		}
	}
	if file != nil {
		file.Close()
	}
	return
}
