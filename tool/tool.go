package tool

var (
	Version string
)

type GMCTool interface {
	Start(args interface{}) (err error)
	Stop()
}
