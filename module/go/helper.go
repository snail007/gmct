package gotool

import (
	"fmt"
	gexec "github.com/snail007/gmc/util/exec"
	grand "github.com/snail007/gmc/util/rand"
	gset "github.com/snail007/gmc/util/set"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type libraryInfoItem struct {
	Path    string
	Version string
}

func (s libraryInfoItem) ModPath() string {
	return s.Path + "@" + s.Version
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
			r1 := regexp.MustCompile(`= +([^ ]+)/.[^ /]+ +in +(/[^ ]+)/src/([^ ]+)\n`)
			m1 := r1.FindStringSubmatch(out)
			if goroot == "" && len(m1) > 0 && isGoSrcPkg(m1[1]) {
				goroot = m1[2]
			}
			r := regexp.MustCompile(`= +[^ ]+ +in +(/[^ /]+)/pkg/mod/([^ ]+)(@[^ /]+)[/.]([^ ]*)\n`)
			m := r.FindAllStringSubmatch(out, -1)
			for _, v := range m {
				pkgPath := v[2] + v[3]
				pkgSet.Add(pkgPath)
				if gopath == "" {
					gopath = v[1]
				}
			}
		}
	}
	for _, v := range pkgSet.ToStringSlice() {
		path := strings.Split(v, "@")
		info.ImportLibraryList = append(info.ImportLibraryList, libraryInfoItem{
			Path:    path[0],
			Version: path[1],
		})
	}
	info.GoRoot = goroot
	info.GoPath = gopath
	if info.GoPath == "" || info.GoRoot == "" {
		return nil, fmt.Errorf("fail to get GOROOT or GOPATH from profile file")
	}
	return
}

func isGoSrcPkg(p string) bool {
	a := strings.SplitN(p, "/", 2)[0]
	return !(strings.Contains(a, ".") && strings.Contains(p, "/"))
}

func goGet(pkg string, env map[string]string, retryCount int) (err error) {
	tmpPath := "/tmp/gogetmod" + grand.String(32)
	defer gexec.NewCommand("rm -rf " + tmpPath).Exec()
	os.MkdirAll(tmpPath, 0755)
RETRY:
	os.Remove(filepath.Join(tmpPath, "go.mod"))
	os.Remove(filepath.Join(tmpPath, "go.sum"))
	cmd := `
cd ` + tmpPath + `
go mod init demo
go get ` + pkg + `
`
	_, err = gexec.NewCommand(cmd).Env(env).Exec()
	if retryCount > 0 {
		retryCount--
		goto RETRY
	}
	return
}
