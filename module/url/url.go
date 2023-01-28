package gurl

import (
	"fmt"
	"github.com/snail007/gmct/tool"
	"net/url"
)

type Args struct {
	SubName   *string
	EncodeStr *string
	DecodeStr *string
}

func NewArgs() Args {
	return Args{
		SubName:   new(string),
		EncodeStr: new(string),
		DecodeStr: new(string),
	}
}

type URL struct {
	tool.GMCTool
	args Args
}

func New() *URL {
	return &URL{}
}

func (s *URL) init(args0 interface{}) (err error) {
	s.args = args0.(Args)
	return
}

func (s *URL) Start(args interface{}) (err error) {
	s.init(args)
	if *s.args.EncodeStr != "" {
		fmt.Println(url.QueryEscape(*s.args.EncodeStr))
	} else if *s.args.DecodeStr != "" {
		result, e := url.QueryUnescape(*s.args.EncodeStr)
		if e != nil {
			err = e
			return
		}
		fmt.Println(result)
	}
	return
}

func (s *URL) Stop() {
	return
}
