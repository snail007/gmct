package file_repeat_scanner

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	gbytes "github.com/snail007/gmc/util/bytes"
	gfile "github.com/snail007/gmc/util/file"
	"github.com/snail007/gmc/util/gpool"
	glist "github.com/snail007/gmc/util/list"
	gmap "github.com/snail007/gmc/util/map"
	"github.com/snail007/gmct/util/checksum"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ResultItem struct {
	info     fs.FileInfo
	filepath string
	md5      string
}

type RepeatResultItem struct {
	md5         string
	fileItem    ResultItem
	repeatFiles []ResultItem
}

type RepeatFileScanner struct {
	dir          string
	scannedCount int64
	allFiles     []ResultItem
	resultMap    *gmap.Map
	repeatCount  int64
	repeatSpace  int64
	logfile      string
	debug        bool
	buf          *gbytes.BytesBuilder
	workersCount int
	doDelete     bool
	deleteNew    bool
}

func (s *RepeatFileScanner) DeleteNew(deleteNew bool) *RepeatFileScanner {
	s.deleteNew = deleteNew
	return s
}

func (s *RepeatFileScanner) Delete(doDelete bool) *RepeatFileScanner {
	s.doDelete = doDelete
	return s
}

func (s *RepeatFileScanner) WorkersCount(workersCount int) *RepeatFileScanner {
	s.workersCount = workersCount
	return s
}

func (s *RepeatFileScanner) Debug(debug bool) *RepeatFileScanner {
	s.debug = debug
	return s
}

func (s *RepeatFileScanner) Logfile(logfile string) *RepeatFileScanner {
	s.logfile = logfile
	return s
}

func NewRepeatFileScanner(dir string) *RepeatFileScanner {
	return &RepeatFileScanner{dir: dir, resultMap: gmap.New(), workersCount: 8, buf: gbytes.NewBytesBuilder()}
}
func (s *RepeatFileScanner) DeleteRepeat() {
	logfile := s.logfile
	if logfile == "" {
		logfile = "repeat_result_" + time.Now().Format("2006-01-02-15-04-05") + "_" + fmt.Sprintf("%d", time.Now().UnixMicro()) + ".txt"
	}
	size, _ := gbytes.SizeStr(uint64(s.repeatSpace))
	s.buf.WriteStrLn("total: %d, repeat: %d, size: %s", len(s.allFiles), s.repeatCount, size)
	for _, item := range s.RepeatResult() {
		s.buf.WriteStrLn("%s -> %d", item.md5, len(item.repeatFiles)+1)
		s.buf.WriteStrLn(item.fileItem.filepath)
		for _, f := range item.repeatFiles {
			if s.doDelete {
				os.Remove(f.filepath)
				gfile.WriteString(f.filepath+".txt",
					fmt.Sprintf("重复文件：%s\n原始文件：%s",
						strings.TrimPrefix(f.filepath, s.dir),
						strings.TrimPrefix(item.fileItem.filepath, s.dir)), false)
				s.buf.WriteStrLn("%s deleted", f.filepath)
			} else {
				s.buf.WriteStrLn("%s repeated", f.filepath)
			}
		}
	}
	gfile.Write(logfile, s.buf.Bytes(), false)
	fmt.Println(fmt.Sprintf("\nresult in file: [%s]", logfile))
}
func (s *RepeatFileScanner) Scan() *RepeatFileScanner {
	s.dir = gfile.Abs(s.dir)
	g := sync.WaitGroup{}
	filepath.Walk(s.dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		s.allFiles = append(s.allFiles, ResultItem{
			filepath: path,
			info:     info,
		})
		return nil
	})
	bar := progressbar.NewOptions(len(s.allFiles),
		progressbar.OptionShowBytes(false),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	g.Add(len(s.allFiles))
	pool := gpool.New(s.workersCount)
	start := time.Now()
	for _, v := range s.allFiles {
		item := v
		path := item.filepath
		pool.Submit(func() {
			defer g.Done()
			file, _ := os.Stat(path)
			size, _ := gbytes.SizeStr(uint64(file.Size()))
			start := time.Now()
			hash, e := checksum.MD5sum(path)
			if e != nil {
				fmt.Println(path, e)
				return
			}
			item.md5 = hash
			if s.debug {
				str := fmt.Sprintf("path:%s, hash:%s, size:%s, time:%s", path, hash, size, time.Since(start).Round(time.Second).String())
				fmt.Println(str)
			}
			l, _ := s.resultMap.LoadAndStoreFunc(hash, func(oldValue interface{}, loaded bool) (newValue interface{}) {
				if !loaded {
					return glist.New()
				}
				return oldValue
			})
			list := l.(*glist.List)
			list.Add(item)
			bar.Add64(1)
		})
	}
	g.Wait()
	pool.Stop()
	s.buf.WriteStrLn("scan start at: %s, cost time: %s",
		start.Format("2006-01-02 15:04:05"), time.Since(start).Round(time.Second).String())

	//sum
	for _, v := range s.RepeatResult() {
		s.repeatCount += int64(len(v.repeatFiles))
		for _, v1 := range v.repeatFiles {
			s.repeatSpace += v1.info.Size()
		}
	}
	return s
}

func (s *RepeatFileScanner) RepeatResult() (result []RepeatResultItem) {
	s.resultMap.RangeFast(func(h, v interface{}) bool {
		hash := h.(string)
		item := v.(*glist.List)
		if item.Len() == 1 {
			return true
		}
		var files []ResultItem
		item.RangeFast(func(_ int, v interface{}) bool {
			files = append(files, v.(ResultItem))
			return true
		})
		var latestItem = files[0]
		for _, f := range files[1:] {
			if s.deleteNew {
				if f.info.ModTime().Before(latestItem.info.ModTime()) {
					latestItem = f
				}
			} else {
				if f.info.ModTime().After(latestItem.info.ModTime()) {
					latestItem = f
				}
			}
		}
		var repeatFiles []ResultItem
		for _, f := range files {
			if f.filepath != latestItem.filepath {
				repeatFiles = append(repeatFiles, f)
			}
		}
		result = append(result, RepeatResultItem{
			md5:         hash,
			fileItem:    latestItem,
			repeatFiles: repeatFiles,
		})
		return true
	})
	return
}

func (s *RepeatFileScanner) Result() *gmap.Map {
	return s.resultMap
}
