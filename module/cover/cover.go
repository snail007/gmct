package cover

import (
	"bytes"
	"crypto/rand"
	"fmt"
	glog "github.com/snail007/gmc/module/log"
	gcast "github.com/snail007/gmc/util/cast"
	gcond "github.com/snail007/gmc/util/cond"
	"github.com/snail007/gmc/util/gpool"
	"github.com/snail007/gmct/module/module"
	"github.com/snail007/gmct/util"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use:     "cover",
			Long:    "run go testing and coverage",
			Aliases: nil,
			RunE: func(c *cobra.Command, a []string) error {
				srv := NewCover(Args{
					Race:       util.Must(c.Flags().GetBool("race")).Bool(),
					Verbose:    util.Must(c.Flags().GetBool("verbose")).Bool(),
					Silent:     util.Must(c.Flags().GetBool("silent")).Bool(),
					KeepResult: util.Must(c.Flags().GetBool("keep")).Bool(),
					ForceCheck: util.Must(c.Flags().GetBool("force")).Bool(),
					Ordered:    util.Must(c.Flags().GetBool("order")).Bool(),
					Only:       util.Must(c.Flags().GetBool("only")).Bool(),
					CoverPkg:   util.Must(c.Flags().GetString("coverpkg")).String(),
					Timeout:    util.Must(c.Flags().GetString("timeout")).String(),
					Debug:      util.Must(c.Flags().GetBool("debug")).Bool(),
				})
				err := srv.init()
				if err != nil {
					return err
				}
				defer srv.Stop()
				return srv.Start()
			},
		}
		cmd.Flags().BoolP("race", "r", false, "enable race checking")
		cmd.Flags().BoolP("verbose", "v", false, "verbose testing logging output")
		cmd.Flags().BoolP("silent", "s", false, "silent mode, not to open a browser")
		cmd.Flags().BoolP("keep", "k", false, "kept the coverage result: coverage.txt")
		cmd.Flags().BoolP("force", "f", false, "overwrite controller file, if it exists.")
		cmd.Flags().BoolP("order", "o", false, "disable parallel run")
		cmd.Flags().Bool("only", false, "only testing current directory without sub directory")
		cmd.Flags().String("coverpkg", "", "additional cover packages split by comma")
		cmd.Flags().String("timeout", "15m", `timeout flag accept any input valid for time.ParseDuration.A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
		cmd.Flags().Bool("debug", false, "in debug mode will logging steps of testing")
		root.AddCommand(cmd)
	})
}

type Args struct {
	Race       bool
	Verbose    bool
	Silent     bool
	KeepResult bool
	ForceCheck bool
	Ordered    bool
	Only       bool
	CoverPkg   string
	Timeout    string
	Debug      bool
	Workers    int
}

type Cover struct {
	args Args
}

func NewCover(args Args) *Cover {
	return &Cover{args: args}
}

func (s *Cover) init() (err error) {
	s.args.Workers = gcond.Cond(s.args.Workers <= 0, runtime.NumCPU(), s.args.Workers).Int()
	return
}

func (s *Cover) Start() (err error) {
	os.Setenv("GMCT_COVER", "true")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return fmt.Errorf("environment variables $GOPATH not found")
	}
	gopath, _ = filepath.Abs(gopath)
	pwd, _ := os.Getwd()
	if !strings.HasPrefix(pwd, gopath) {
		return fmt.Errorf("your current path must be in $GOPATH/src")
	}
	s.debugf("$GOPATH is " + gopath)
	race := ""
	if s.args.Race {
		race = "-race"
	}
	verbose := ""
	if s.args.Verbose {
		verbose = "-v"
	}
	dir := "./..."
	if s.args.Only {
		dir = "./"
	}
	var output string
	output, err = s.exec("go list "+dir+" | grep -v /main | grep -v /vendor | grep -v /examples | grep -v grep", "")
	if err != nil {
		return
	}
	packagesAll := strings.Split(output, "\n")
	packages := []string{}
	for _, pkg := range packagesAll {
		if len(pkg) == 0 {
			continue
		}
		if s.args.ForceCheck || s.hasTestFile(pkg) {
			packages = append(packages, pkg)
		}
	}
	if len(packages) == 0 {
		err = fmt.Errorf("no match package found")
		return
	}
	s.debugf("testing packages list, %s", strings.Join(packages, ","))
	coverProfiles := make([]string, len(packages))
	payload := "mode: atomic\n"
	var p *gpool.GPool
	if !s.args.Ordered {
		p = gpool.New(s.args.Workers)
	}
	timeout := "15m"
	if v := os.Getenv("GMCT_TEST_TIMEOUT"); v != "" {
		timeout = v
	} else if s.args.Timeout != "" {
		timeout = s.args.Timeout
	}
	coverPkgs := strings.Join(packages, ",")
	addtionalCoverPkgs := strings.Split(s.args.CoverPkg, ",")
	if len(addtionalCoverPkgs) > 0 {
		for _, v := range addtionalCoverPkgs {
			if len(v) == 0 {
				continue
			}
			coverPkgs += "," + v
		}
	}
	s.debugf("cover packages list, %s", coverPkgs)
	os.Setenv("GMCT_COVER_VERBOSE", fmt.Sprintf("%v", s.args.Verbose))
	os.Setenv("GMCT_COVER_RACE", fmt.Sprintf("%v", s.args.Race))
	os.Setenv("GMCT_COVER_PACKAGES", coverPkgs)
	os.Setenv("GMCT_TEST_TIMEOUT", timeout)

	for k, pkg := range packages {
		b := make([]byte, 16)
		io.ReadFull(rand.Reader, b)
		coverProfiles[k] = filepath.Join(os.TempDir(), fmt.Sprintf("%x", b)) + ".gocc.tmp"
		w := func(coverprofile, pkg string) {
			workDir := filepath.Join(gopath, "src", pkg)
			cmd := `go test -timeout ` + timeout + ` ` + verbose + ` ` + race +
				` -covermode=atomic -coverprofile=` + coverprofile +
				` -coverpkg=` + coverPkgs + ` ` + pkg
			output, err = s.exec(cmd, workDir)
			if err != nil {
				return
			}
			for _, line := range strings.Split(output, "\n") {
				if strings.Contains(line, "warning: no packages being tested depend on matches for pattern") {
					continue
				}
				s := strings.Trim(strings.Split(line, " of statements in")[0], "\n")
				if s != "" {
					fmt.Println(s)
				}
			}
		}
		if !s.args.Ordered {
			coverProfile, packageName := coverProfiles[k], pkg
			p.Submit(func() {
				w(coverProfile, packageName)
			})
		} else {
			w(coverProfiles[k], pkg)
		}
	}
	if !s.args.Ordered {
		p.WaitDone()
		p.Stop()
	}
	//scan additional code coverage coverProfiles by testing
	for _, pkg := range packages {
		path := filepath.Join(gopath, "src", pkg)
		fs, _ := filepath.Glob(path + "/*.gocc.tmp")
		for _, f := range fs {
			coverProfiles = append(coverProfiles, f)
		}
		if len(fs) > 0 && s.args.Verbose {
			fmt.Printf("cover coverProfiles found %v\n", fs)
		}
	}
	for _, file := range coverProfiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}
		p, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		ps := strings.Split(string(p), "\n")
		payload += strings.Join(ps[1:], "\n")
		os.Remove(file)
	}
	err = ioutil.WriteFile("coverage.txt", []byte(payload), 0755)
	if err != nil {
		return
	}
	_, err = s.exec("go tool cover -func=coverage.txt -o total.txt", "")
	if err != nil {
		return
	}
	output, err = s.exec("tail -n 1 total.txt", "")
	if err != nil {
		return
	}
	if output != "" {
		fmt.Println(output)
	}
	if !s.args.Silent {
		_, err = s.exec("go tool cover -html=coverage.txt", "")
		if err != nil {
			return
		}
	}
	err = os.Remove("total.txt")
	if err != nil {
		return
	}
	if !s.args.KeepResult {
		err = os.Remove("coverage.txt")
		if err != nil {
			return
		}
	}
	return
}
func (s *Cover) debugf(formats string, values ...interface{}) {
	if gcast.ToBool(os.Getenv("DEBUG")) || s.args.Debug {
		glog.Debugf(formats, values...)
	}
}
func (s *Cover) exec(cmd, workDir string) (output string, err error) {
	if workDir == "" {
		workDir = "." + string(os.PathSeparator)
	}
	s.debugf("[WORK DIR]: %v, [COMMAND]: %v", workDir, cmd)
	c := exec.Command("bash", "-c", cmd)
	c.Env = append(c.Env, os.Environ()...)
	c.Dir = workDir
	b, err := c.CombinedOutput()
	output = string(b)
	if err != nil {
		var b bytes.Buffer
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "warning: no packages being tested depend on matches for pattern") {
				continue
			}
			b.WriteString(line + "\n")
		}
		fmt.Printf("EXEC FAIL, COMMAND: %s\n"+
			"ERROR: %s\n"+
			"OUTPUT: \n%s\n",
			cmd, err, b.String())
		return
	}
	return
}

func (s *Cover) hasTestFile(pkg string) bool {
	path := filepath.Join(os.Getenv("GOPATH"), "src", pkg)
	files, _ := filepath.Glob(path + "/*_test.go")
	return len(files) > 0
}
func (s *Cover) Stop() {
	return
}
