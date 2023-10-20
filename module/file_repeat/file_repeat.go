package file_repeat

import (
	gcond "github.com/snail007/gmc/util/cond"
	"github.com/snail007/gmct/tool"
	"github.com/snail007/gmct/util/file_repeat_scanner"
	"runtime"
)

type FileRepeatArgs struct {
	SubName *string
	Delete  *bool
	Dir     *string
	Logfile *string
	Debug   *bool
	Workers *int
}

func NewFileRepeatArgs() FileRepeatArgs {
	return FileRepeatArgs{
		SubName: new(string),
		Delete:  new(bool),
		Dir:     new(string),
		Debug:   new(bool),
		Logfile: new(string),
		Workers: new(int),
	}
}

type FileRepeat struct {
	tool.GMCTool
	args FileRepeatArgs
}

func NewFileRepeat() *FileRepeat {
	return &FileRepeat{}
}

func (s *FileRepeat) init(args0 interface{}) (err error) {
	s.args = args0.(FileRepeatArgs)
	return
}

func (s *FileRepeat) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	workers := *s.args.Workers
	if workers == 0 {
		workers = runtime.NumCPU()
	}
	dir := gcond.Cond(*s.args.Dir != "", *s.args.Dir, "./").(string)
	file_repeat_scanner.NewRepeatFileScanner(dir).
		WorkersCount(workers).
		Delete(*s.args.Delete).
		Logfile(*s.args.Logfile).
		Debug(*s.args.Debug).
		Scan().
		DeleteRepeat()
	return
}

func (s *FileRepeat) Stop() {
	return
}
