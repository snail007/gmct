package cover

import (
	"bytes"
	"crypto/rand"
	"fmt"
	glog "github.com/snail007/gmc/module/log"
	gcast "github.com/snail007/gmc/util/cast"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/snail007/gmct/tool"
)

type CoverArgs struct {
	SubName    *string
	Race       *bool
	Verbose    *bool
	Silent     *bool
	KeepResult *bool
	ForceCheck *bool
	Ordered    *bool
	Only       *bool
	CoverPkg   *string
	Timeout    *string
	Debug      *bool
}

func NewCoverArgs() CoverArgs {
	return *new(CoverArgs)
}

type Cover struct {
	tool.GMCTool
	args CoverArgs
}

func NewCover() *Cover {
	return &Cover{}
}

func (s *Cover) init(args0 interface{}) (err error) {
	s.args = args0.(CoverArgs)
	return
}

func (s *Cover) Start(args interface{}) (err error) {
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
	err = s.init(args)
	if err != nil {
		return
	}
	s.debugf("$GOPATH is " + gopath)
	race := ""
	if *s.args.Race {
		race = "-race"
	}
	verbose := ""
	if *s.args.Verbose {
		verbose = "-v"
	}
	dir := "./..."
	if *s.args.Only {
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
		if *s.args.ForceCheck || s.hasTestFile(pkg) {
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
	var g sync.WaitGroup
	var errChn chan error
	var doneChn chan bool
	if !*s.args.Ordered {
		g = sync.WaitGroup{}
		errChn = make(chan error)
		doneChn = make(chan bool)
		g.Add(len(packages))
	}
	timeout := "15m"
	if v := os.Getenv("GMCT_TEST_TIMEOUT"); v != "" {
		timeout = v
	} else if *s.args.Timeout != "" {
		timeout = *s.args.Timeout
	}
	coverPkgs := strings.Join(packages, ",")
	addtionalCoverPkgs := strings.Split(*s.args.CoverPkg, ",")
	if len(addtionalCoverPkgs) > 0 {
		for _, v := range addtionalCoverPkgs {
			if len(v) == 0 {
				continue
			}
			coverPkgs += "," + v
		}
	}
	s.debugf("cover packages list, %s", coverPkgs)
	os.Setenv("GMCT_COVER_VERBOSE", fmt.Sprintf("%v", *s.args.Verbose))
	os.Setenv("GMCT_COVER_RACE", fmt.Sprintf("%v", *s.args.Race))
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
		if !*s.args.Ordered {
			go func(coverprofile, pkg string) {
				defer g.Done()
				w(coverprofile, pkg)
			}(coverProfiles[k], pkg)
		} else {
			w(coverProfiles[k], pkg)
		}
	}
	if !*s.args.Ordered {
		go func() {
			g.Wait()
			doneChn <- true
		}()
		select {
		case <-doneChn:
		case e := <-errChn:
			return e
		}
	}
	//scan additional code coverage coverProfiles by testing
	for _, pkg := range packages {
		path := filepath.Join(gopath, "src", pkg)
		fs, _ := filepath.Glob(path + "/*.gocc.tmp")
		for _, f := range fs {
			coverProfiles = append(coverProfiles, f)
		}
		if len(fs) > 0 && *s.args.Verbose {
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
	if !*s.args.Silent {
		_, err = s.exec("go tool cover -html=coverage.txt", "")
		if err != nil {
			return
		}
	}
	err = os.Remove("total.txt")
	if err != nil {
		return
	}
	if !*s.args.KeepResult {
		err = os.Remove("coverage.txt")
		if err != nil {
			return
		}
	}
	return
}
func (s *Cover) debugf(formats string, values ...interface{}) {
	if gcast.ToBool(os.Getenv("DEBUG")) || *s.args.Debug {
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
