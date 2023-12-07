package gotool

import (
	"fmt"
	gbytes "github.com/snail007/gmc/util/bytes"
	gexec "github.com/snail007/gmc/util/exec"
	gfile "github.com/snail007/gmc/util/file"
	grand "github.com/snail007/gmc/util/rand"
	gset "github.com/snail007/gmc/util/set"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type libraryInfoItem struct {
	importPath string
	modPath    string
	modVersion string
}

func (s libraryInfoItem) ModFullPath() string {
	return s.modPath + "@" + s.modVersion
}

func (s libraryInfoItem) ModPath() string {
	return s.modPath
}
func (s libraryInfoItem) ModVersion() string {
	return s.modVersion
}

func (s libraryInfoItem) ImportPath() string {
	return s.importPath
}

type profileInfo struct {
	ImportLibraryList []libraryInfoItem
	GoRoot            string
	GoPath            string
}

func getProfileInfo(files []string) (info *profileInfo, err error) {
	info = new(profileInfo)
	gopath := ""
	goroot := ""
	pkgSet := gset.New()
	for _, file := range files {
		modImportPkgListCMD := `go tool pprof -alloc_space -list ".*"  ` + file
		out, _ := gexec.NewCommand(modImportPkgListCMD).Exec()
		if out == "" {
			out, _ = gexec.NewCommand(`go tool pprof -list ".*"  ` + file).Exec()
		}
		if len(out) > 0 {
			r1 := regexp.MustCompile(`= +([^ ]+) +in +([^ ]+\.go)\n`)
			m1 := r1.FindAllStringSubmatch(out, -1)
			for _, v := range m1 {
				if goroot != "" && gopath != "" {
					break
				}
				p1 := filepath.Dir(v[1])
				p2 := v[2]
				p1Path := filepath.Join(p1, strings.SplitN(filepath.Base(v[1]), ".", 2)[0])
				if isGoSrcPkg(p1Path) {
					if goroot == "" {
						idx := strings.Index(p2, "/src/")
						if idx >= 0 {
							goroot = p2[:idx]
						}
					}
				} else {
					if gopath == "" {
						idx := strings.Index(p2, "/pkg/mod/")
						if idx >= 0 {
							gopath = p2[:idx]
						}
					}
				}
			}

			if gopath == "" {
				for _, v := range m1 {
					p1 := filepath.Dir(v[1])
					p2 := v[2]
					p1Path := filepath.Join(p1, strings.SplitN(filepath.Base(v[1]), ".", 2)[0])
					if !isGoSrcPkg(p1Path) {
						idx := strings.Index(p2, "/src/")
						if idx >= 0 {
							gopath = p2[:idx]
						}
					}
				}
			}

			r := regexp.MustCompile(`= +([^ ]+/[^ .]+).[^ ]+ +in +(/[^ ]+)/pkg/mod/([^ ]+)@([^ /]+)[/.]([^ ]*)\n`)
			m := r.FindAllStringSubmatch(out, -1)
			for _, v := range m {
				importPkg := v[1]
				//goPath := v[2]
				modePkg := v[3]
				modeVersion := v[4]
				pkgPath := modePkg + modeVersion
				if !pkgSet.Contains(pkgPath) {
					pkgSet.Add(pkgPath)
					info.ImportLibraryList = append(info.ImportLibraryList, libraryInfoItem{
						importPath: importPkg,
						modPath:    modePkg,
						modVersion: modeVersion,
					})
				}
			}
		}
	}
	info.GoRoot = goroot
	info.GoPath = gopath
	if info.GoRoot == "" {
		return nil, fmt.Errorf("fail to get GOROOT from profile file")
	}
	return
}

func isGoSrcPkg(p string) bool {
	a := strings.SplitN(p, "/", 2)[0]
	return !(strings.Contains(a, ".") && strings.Contains(p, "/"))
}

func goGet(pkg []libraryInfoItem, env map[string]string, retryCount int) (err error) {
	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)
	tmpPath := "/tmp/gogetmod" + grand.String(32)
	defer gexec.NewCommand("rm -rf " + tmpPath).Exec()
	os.MkdirAll(tmpPath, 0755)
	os.Chdir(tmpPath)

RETRY:
	os.Remove(filepath.Join(tmpPath, "go.mod"))
	os.Remove(filepath.Join(tmpPath, "go.sum"))
	gexec.NewCommand("go mod init demo").Env(env).Exec()
	bufMain := gbytes.NewBytesBuilder()
	bufMain.WriteStrLn("package main")
	for _, v := range pkg {
		bufMain.WriteStrLn(`import _ "%s"`, v.importPath)
		//if v.ImportPath() != v.ModPath() {
		//	gexec.NewCommand(fmt.Sprintf("go mod edit -replace=%s=%s", v.ImportPath(), v.ModFullPath())).Env(env).Exec()
		//}
	}
	gfile.WriteString("main.go", bufMain.String(), false)
	_, err = gexec.NewCommand("go mod tidy").Env(env).Exec()
	if retryCount > 0 {
		retryCount--
		goto RETRY
	}
	return
}
