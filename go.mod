module github.com/snail007/gmct

go 1.16

require (
	github.com/AlecAivazis/survey/v2 v2.3.5
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/pkg/errors v0.8.1
	github.com/schollz/progressbar/v3 v3.8.1
	github.com/snail007/gmc v0.0.0-20220826032345-5caed47f4d5e
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf
	golang.org/x/net v0.0.0-20210916014120-12bc252f5db8
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/cheggaaa/pb.v1 v1.0.28
)

//replace github.com/snail007/gmc => ../gmc
