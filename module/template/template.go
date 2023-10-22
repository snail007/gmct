package template

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/snail007/gmct/module/module"
	"github.com/snail007/gmct/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use:     "tpl",
			Long:    "pack or clean templates go file",
			Aliases: nil,
			RunE: func(c *cobra.Command, a []string) error {
				srv := NewTemplate(Args{
					Dir:       util.Must(c.Flags().GetString("dir")).String(),
					Clean:     util.Must(c.Flags().GetBool("clean")).Bool(),
					Extension: util.Must(c.Flags().GetString("ext")).String(),
				})
				err := srv.init()
				if err != nil {
					return err
				}
				defer srv.Stop()
				return srv.Start()
			},
		}
		cmd.Flags().String("dir", "", "template's template directory path, gmct will convert all template files in the folder to one go file")
		cmd.Flags().String("ext", "", "extension of template files")
		cmd.Flags().Bool("clean", false, "clean packed file, if exists")
		root.AddCommand(cmd)

	})
}

type Args struct {
	Dir              string
	Clean            bool
	Extension        string
	GoFilenamePrefix string
	GoFilename       string
}

type Template struct {
	args Args
}

func NewTemplate(args Args) *Template {
	return &Template{args: args}
}

func (s *Template) init() (err error) {
	if s.args.Dir == "" {
		return fmt.Errorf("templates directory not exists")
	}
	//convert pack path to absoulte path of *nix style
	s.args.Dir, err = filepath.Abs(s.args.Dir)
	if err != nil {
		return
	}
	s.args.Dir = strings.Replace(s.args.Dir, "\\", "/", -1)
	s.args.Dir = filepath.Join(s.args.Dir, "/")

	if s.args.Extension == "" {
		return fmt.Errorf("extension of template file required")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	s.args.GoFilenamePrefix = "gmc_templates_bindata_"
	s.args.GoFilename = s.args.GoFilenamePrefix + fmt.Sprintf("%d", r.Int63()) + ".go"
	return
}

func (s *Template) Start() (err error) {
	if s.args.Clean {
		s.clean()
	} else {
		s.clean()
		err = s.pack()
	}
	return
}

func (s *Template) Stop() {

	return
}
func (s *Template) pack() (err error) {
	names := []string{}
	err = s.tree(s.args.Dir, &names)
	if err != nil {
		return
	}
	tplFilenames := []string{}
	for _, v := range names {
		filename := filepath.Base(v)
		if strings.HasSuffix(filename, s.args.Extension) {
			tplFilenames = append(tplFilenames, v)
		}
	}
	var buf bytes.Buffer
	for _, v := range tplFilenames {
		b, err := ioutil.ReadFile(filepath.Join(s.args.Dir, v))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		str := base64.StdEncoding.EncodeToString(b)
		buf.WriteString(fmt.Sprintf("\t\t\"%s\" : \"%s\",\n", strings.TrimSuffix(v, s.args.Extension), str))
	}
	currentDir, _ := filepath.Abs(".")
	packageName := util.GetPackageName(currentDir)
	tpl := fmt.Sprintf(`package %s

import gmctemplate "github.com/snail007/gmc/http/template"

func init(){
	gmctemplate.SetBinBase64(map[string]string{

{{HOLDER}}
	})
}
`, packageName)
	tpl = strings.Replace(tpl, "{{HOLDER}}", buf.String(), 1)
	ioutil.WriteFile(s.args.GoFilename, []byte(tpl), 0755)
	return
}
func (s *Template) clean() {
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
func (s *Template) tree(folder string, names *[]string) (err error) {
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
			v0 := strings.Replace(v, "\\", "/", -1)
			v0 = strings.Replace(v0, s.args.Dir+"/", "", -1)
			*names = append(*names, v0)
		}
	}
	if file != nil {
		file.Close()
	}
	return
}
