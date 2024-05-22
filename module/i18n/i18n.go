package i18n

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
			Use:     "i18n",
			Long:    "pack or clean i18n go file",
			Aliases: nil,
			RunE: func(c *cobra.Command, a []string) error {
				srv := NewI18n(Args{
					Dir:       util.Must(c.Flags().GetString("dir")).String(),
					Clean:     util.Must(c.Flags().GetBool("clean")).Bool(),
					Extension: ".toml",
				})
				err := srv.init()
				if err != nil {
					return err
				}
				defer srv.Stop()
				return srv.Start()
			},
		}
		cmd.Flags().String("dir", ".", "i18n's template directory path, gmct will convert all i18n files in the folder to one go file")
		cmd.Flags().Bool("clean", false, "clean packed file, if exists")
		root.AddCommand(cmd)
	})
}

type Args struct {
	Dir              string
	GoFilenamePrefix string
	GoFilename       string
	Extension        string
	Clean            bool
}

type I18n struct {
	args Args
}

func NewI18n(args Args) *I18n {
	return &I18n{args: args}
}

func (s *I18n) init() (err error) {
	//convert pack path to absoulte path of *nix style
	s.args.Dir, err = filepath.Abs(s.args.Dir)
	if err != nil {
		return
	}
	s.args.Dir = strings.Replace(s.args.Dir, "\\", "/", -1)
	s.args.Dir = filepath.Join(s.args.Dir, "/")

	if s.args.Extension == "" {
		return fmt.Errorf("extension of i18n file required")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	s.args.GoFilenamePrefix = "gmc_i18n_bindata_"
	s.args.GoFilename = s.args.GoFilenamePrefix + fmt.Sprintf("%d", r.Int63()) + ".go"
	return
}

func (s *I18n) Start() (err error) {
	if s.args.Clean {
		s.clean()
	} else {
		s.clean()
		err = s.pack()
	}
	return
}

func (s *I18n) Stop() {

	return
}
func (s *I18n) pack() (err error) {
	names := []string{}
	err = s.tree(s.args.Dir, &names)
	if err != nil {
		return
	}
	i18nFilenames := []string{}
	for _, v := range names {
		filename := filepath.Base(v)
		if strings.HasSuffix(filename, s.args.Extension) {
			i18nFilenames = append(i18nFilenames, v)
		}
	}
	var buf bytes.Buffer
	for _, v := range i18nFilenames {
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
	i18n := fmt.Sprintf(`package %s

import gi18n "github.com/snail007/gmc/module/i18n"

func init(){
	gi18n.SetBinData(map[string]string{

{{HOLDER}}
	})
}
`, packageName)
	i18n = strings.Replace(i18n, "{{HOLDER}}", buf.String(), 1)
	ioutil.WriteFile(s.args.GoFilename, []byte(i18n), 0755)
	return
}
func (s *I18n) clean() {
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
func (s *I18n) tree(folder string, names *[]string) (err error) {
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
