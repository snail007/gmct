package run

import (
	"context"
	"fmt"
	"github.com/snail007/gmc"
	gmchook "github.com/snail007/gmc/util/process/hook"
	"github.com/snail007/gmct/tool"
	"github.com/snail007/gmct/util"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	tplfilename = "gmcrun.toml"
	tpl         = `[build]
# ${DIR} is a placeholder presents current dir absolute path, no slash in the end.
# you can using it in monitor_dirs, include_files, exclude_files, exclude_dirs.
# "go" command will be called by default, but you can set "cmd" to overwrite it.
monitor_dirs=["."]
cmd=""
args=["-ldflags","-s -w"]
env=["CGO_ENABLED=1","GO111MODULE=on"]
include_exts=[".go",".html",".htm",".tpl",".toml",".ini",".conf",".yaml"]
include_files=[]
exclude_files=["gmcrun.toml"]
exclude_dirs=["vendor"]`
)

type RunArgs struct {
	Args    []string
	SubName *string
}

func NewRunArgs() RunArgs {
	return RunArgs{
		SubName: new(string),
	}
}

type Run struct {
	tool.GMCTool
	onlyExts     map[string]bool
	runName      string
	proc         *exec.Cmd
	restartSig   chan bool
	stopCtx      context.Context
	cancle       context.CancelFunc
	cfg          RunArgs
	buildEnv     []string
	buildArgs    []string
	excludeFiles map[string]bool
	excludeDirs  map[string]bool
	includeFiles map[string]bool
	monitorDirs  []string
	workDir      string
	cmd          string
}

func NewRun() *Run {
	ctx, cancle := context.WithCancel(context.Background())
	return &Run{
		cancle:       cancle,
		stopCtx:      ctx,
		restartSig:   make(chan bool),
		onlyExts:     map[string]bool{},
		excludeFiles: map[string]bool{},
		excludeDirs:  map[string]bool{},
		includeFiles: map[string]bool{},
	}
}

func (s *Run) init(args0 interface{}) (err error) {

	s.cfg = args0.(RunArgs)
	if !util.Exists(tplfilename) {
		err = ioutil.WriteFile(tplfilename, []byte(tpl), 0755)
		if err != nil {
			return
		}
	}
	// init cfg
	cfg := gmc.New.Config()
	cfg.SetConfigFile(tplfilename)
	err = cfg.ReadInConfig()
	if err != nil {
		return
	}
	for _, v := range cfg.GetStringSlice("build.include_exts") {
		s.onlyExts[v] = true
	}
	curdir, _ := os.Getwd()
	for _, v := range cfg.GetStringSlice("build.monitor_dirs") {
		v = strings.Replace(v, "${DIR}", curdir, -1)
		v, _ = filepath.Abs(v)
		s.monitorDirs = append(s.monitorDirs, v)
	}

	for arr, cfgarr := range map[*map[string]bool][]string{
		&s.excludeFiles: cfg.GetStringSlice("build.exclude_files"),
		&s.excludeDirs:  cfg.GetStringSlice("build.exclude_dirs"),
		&s.includeFiles: cfg.GetStringSlice("build.include_files"),
	} {
		for _, v := range cfgarr {
			// ${DIR} presents current directory
			v = strings.Replace(v, "${DIR}", curdir, -1)
			if filepath.IsAbs(v) {
				(*arr)[v] = true
			} else {
				for _, d := range s.monitorDirs {
					(*arr)[filepath.Join(d, v)] = true
				}
			}
		}
	}

	s.workDir, _ = filepath.Abs(".")
	s.buildEnv = os.Environ()
	s.buildEnv = append(s.buildEnv, cfg.GetStringSlice("build.env")...)
	s.cmd = cfg.GetString("build.cmd")
	if s.cmd == "" {
		s.buildArgs = []string{"build"}
		s.buildArgs = append(s.buildArgs, cfg.GetStringSlice("build.args")...)
		s.runName = filepath.Join(s.workDir, filepath.Base(s.workDir)+"_gmcrun")
		if os.PathSeparator == '\\' {
			s.runName += ".exe"
		}
		s.buildArgs = append(s.buildArgs, "-o", s.runName)
	} else {
		s.runName = s.cmd
	}

	//fmt.Println(s.runName, s.monitorDirs, "\n", s.onlyExts, "\n", s.includeFiles, "\n", s.excludeFiles, "\n", s.excludeDirs)

	gmchook.RegistShutdown(func() {
		s.cancle()
		s.kill()
		os.Remove(s.runName)
	})
	return
}

func (s *Run) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	go s.scan()
	go s.restartMonitor()
	gmchook.WaitShutdown()
	return
}
func (s *Run) kill() {
	if s.proc != nil && s.proc.Process != nil {
		s.proc.Process.Kill()
	}
}
func (s *Run) build() {
	cmd := exec.CommandContext(s.stopCtx, "go", s.buildArgs...)
	cmd.Env = s.buildEnv
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	e := cmd.Run()
	if e != nil {
		fmt.Printf("rebuild fail, error: %s\n", e)
	}
}
func (s *Run) start() {
	s.proc = exec.CommandContext(s.stopCtx, s.runName, s.cfg.Args...)
	s.proc.Env = os.Environ()
	s.proc.Stderr = os.Stderr
	s.proc.Stdin = os.Stdin
	s.proc.Stdout = os.Stdout
	go s.proc.Run()
}
func (s *Run) restartMonitor() {
	for {
		select {
		case <-s.stopCtx.Done():
		case <-s.restartSig:
			fmt.Print("\n>>> gmct run: file changed found, rebuilding... <<<\n\n")
			s.kill()
			time.Sleep(time.Second)
			os.Remove(s.runName)
			if s.cmd == "" {
				s.build()
			}
			s.start()
			if s.cmd == "" {
				time.Sleep(time.Second)
			}
		}
	}
}
func (s *Run) scan() {
	list := map[string]int64{}
	for {
		select {
		case <-s.stopCtx.Done():
		case <-time.After(time.Second * 2):
			restart := false
			var names []string
			var newList = map[string]bool{}
			for _, d := range s.monitorDirs {
				s.tree(d, &names)
			}
			for _, v := range names {
				newList[v] = true
			}

			// deleted files found?
			for k := range list {
				if !newList[k] {
					restart = true
					delete(list, k)
				}
			}

			// added and changed will be monitor here
			if !restart {
				for _, v := range names {
					st, err := os.Stat(v)
					if err != nil {
						continue
					}
					t0 := st.ModTime().Unix()
					if t, ok := list[v]; ok {
						// found, compare modify time
						if t0 != t {
							//restart
							//fmt.Printf("modify file: %s, %d -> %d\n", v, t, t0)
							restart = true
						}
					} else {
						// not found, restart
						//fmt.Printf("new file: %s\n", v)
						restart = true
					}
					list[v] = t0
				}
			}

			if restart {
				select {
				case s.restartSig <- true:
				default:
					fmt.Println("send restart signal fail")
				}
			}
		}
	}
}
func (s *Run) Stop() {
	s.cancle()
	return
}

func (s *Run) tree(folder string, names *[]string) (err error) {
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
		n := filepath.Base(v)
		if strings.HasPrefix(n, ".") {
			continue
		}
		if fileInfo.IsDir() {
			if s.excludeDirs[v] {
				continue
			}
			err = s.tree(v, names)
			if err != nil {
				return err
			}
		} else {
			if s.excludeFiles[v] {
				continue
			}
			if !s.onlyExts[filepath.Ext(n)] && !s.includeFiles[v] {
				continue
			}
			if strings.HasPrefix(filepath.Base(v), ".") {
				continue
			}
			*names = append(*names, v)
		}
	}
	if file != nil {
		file.Close()
	}
	return
}
