package tool

type GMCTool interface {
	Start(args interface{}) (err error)
	Stop()
}
