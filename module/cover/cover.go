package cover

import (
	"crypto/rand"
	"fmt"
	"github.com/snail007/gmct/tool"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type CoverArgs struct {
	SubName    *string
	Race       *bool
	Verbose    *bool
	Silent     *bool
	KeepResult *bool
	ForceCheck *bool
	Ordered    *bool
}

func NewCoverArgs() CoverArgs {
	return *new(CoverArgs)
}

type Cover struct {
	tool.GMCTool
	args CoverArgs
}

func NewCover() *Cover {
	return &Cover{
	}
}

func (s *Cover) init(args0 interface{}) (err error) {
	s.args = args0.(CoverArgs)
	return
}

func (s *Cover) Start(args interface{}) (err error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return fmt.Errorf("GOPATH not found")
	}
	gopath, _ = filepath.Abs(gopath)
	pwd, _ := os.Getwd()
	if !strings.HasPrefix(pwd, gopath) {
		return fmt.Errorf("your current path must be in GOPATH/src")
	}
	err = s.init(args)
	if err != nil {
		return
	}
	race := ""
	if *s.args.Race {
		race = "-race"
	}
	verbose := ""
	if *s.args.Verbose {
		verbose = "-v"
	}
	var output string
	output, err = s.exec("go list ./... | grep -v /main | grep -v /vendor | grep -v /examples | grep -v grep", "")
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

	files := make([]string, len(packages))
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
	coverPkgs := strings.Join(packages, ",")
	os.Setenv("GMCT_COVER_VERBOSE", fmt.Sprintf("%v", *s.args.Verbose))
	os.Setenv("GMCT_COVER_RACE", fmt.Sprintf("%v", *s.args.Race))
	os.Setenv("GMCT_COVER_PACKAGES", coverPkgs)
	for k, pkg := range packages {
		b := make([]byte, 16)
		io.ReadFull(rand.Reader, b)
		files[k] = filepath.Join(os.TempDir(), fmt.Sprintf("%x", b)) + ".gocc.tmp"
		w := func(file, pkg string) {
			workDir := filepath.Join(gopath, "src", pkg)
			cmd := `go test ` + verbose + ` ` + race + ` -covermode=atomic -coverprofile=` + file +
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
			go func(file, pkg string) {
				defer g.Done()
				w(file, pkg)
			}(files[k], pkg)
		} else {
			w(files[k], pkg)
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
	//scan additional code coverage files by testing
	for _, pkg := range packages {
		path := filepath.Join(gopath, "src", pkg)
		fs, _ := filepath.Glob(path + "/*.gocc.tmp")
		for _, f := range fs {
			files = append(files, f)
		}
		if len(fs) > 0 && *s.args.Verbose {
			fmt.Printf("cover files found %v\n", fs)
		}
	}
	for _, file := range files {
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

func (s *Cover) exec(cmd, workDir string) (output string, err error) {
	c := exec.Command("bash", "-c", cmd)
	c.Env = append(c.Env, os.Environ()...)
	c.Dir = workDir
	b, err := c.CombinedOutput()
	output = string(b)
	if err != nil {
		fmt.Printf("exec fail, err: %v, command: %s\noutput: \n%s", err, cmd, output)
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

func newCmdFromEnv(runName string) string {
	if strings.Contains(runName, "/") {
		rs := strings.Split(runName, "/")
		runName = "^" + strings.Join(rs, "$/^") + "$"
	}
	rb := make([]byte, 16)
	io.ReadFull(rand.Reader, rb)
	d, _ := os.Getwd()
	coverfile := filepath.Join(d, fmt.Sprintf("%x", rb)) + ".gocc.tmp"
	race := os.Getenv("GMCT_COVER_RACE")
	packages := os.Getenv("GMCT_COVER_PACKAGES")
	pkg := strings.TrimPrefix(d, filepath.Join(os.Getenv("GOPATH"), "src"))[1:]
	if race == "true" {
		race = "-v"
	} else {
		race = ""
	}
	cover := ""
	if packages != "" {
		cover = fmt.Sprintf("-covermode=atomic -coverprofile=%s -coverpkg=%s", coverfile, packages)
	}
	return fmt.Sprintf(`go test -v -run=%s %s %s %s`,
		runName, cover, race, pkg)
}

func ExecTestFunc(testFuncName string) (out string, err error) {
	isVerbose := os.Getenv("GMCT_COVER_VERBOSE") == "true"
	defer func() {
		if isVerbose {
			fmt.Printf("output: %s", out)
			fmt.Printf(">>> end child testing process %s\n", testFuncName)
		}
	}()
	if isVerbose {
		fmt.Printf(">>> start child testing process %s\n", testFuncName)
	}
	cmdStr := newCmdFromEnv(testFuncName)
	if isVerbose {
		fmt.Println(cmdStr)
	}
	c := exec.Command("bash", "-c", cmdStr)
	c.Env = append(c.Env, os.Environ()...)
	b, err := c.CombinedOutput()
	out = string(b)
	return
}
