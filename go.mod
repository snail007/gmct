module github.com/snail007/gmct

go 1.12

require (
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/schollz/progressbar/v3 v3.8.1
	github.com/snail007/gmc v0.0.0-20210520085422-78904d697668
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/cheggaaa/pb.v1 v1.0.28
)

//replace github.com/snail007/gmc => ../gmc
