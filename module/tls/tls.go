package tlstool

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	glog "github.com/snail007/gmc/module/log"
	gfile "github.com/snail007/gmc/util/file"
	"github.com/snail007/gmct/tool"
)

type TLSArgs struct {
	InfoAddr *string
	SaveAddr *string
	SaveName *string
	File     *string
	TLSName  *string
	SubName  *string
}

func NewTLSArgs() TLSArgs {
	return TLSArgs{
		TLSName: new(string),
		SubName: new(string),
	}
}

type TLS struct {
	tool.GMCTool
	args TLSArgs
}

func NewTLS() *TLS {
	return &TLS{}
}

func (s *TLS) init(args0 interface{}) (err error) {
	s.args = args0.(TLSArgs)

	switch *s.args.SubName {
	case "info":
		if *s.args.InfoAddr != "" && !strings.Contains(*s.args.InfoAddr, ":") {
			*s.args.InfoAddr += ":443"
		}
	case "save":
		if *s.args.SaveAddr != "" && !strings.Contains(*s.args.SaveAddr, ":") {
			*s.args.SaveAddr += ":443"
		}
	}

	return
}

func (s *TLS) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	switch *s.args.SubName {
	case "info":
		s.info()
	case "save":
		s.save()
	}
	return
}

func (s *TLS) Stop() {
	return
}
func (s *TLS) info() {
	if *s.args.InfoAddr != "" {
		info, err := getTLSInfo(s.getConnectionState(*s.args.InfoAddr))
		if err != nil {
			glog.Error(err)
		}
		fmt.Println(info.String())
	} else if *s.args.File != "" {
		bs, err := x509.ParseCertificates(gfile.Bytes(*s.args.File))
		if err != nil {
			glog.Error(errors.Wrap(err, "ParseCertificates"))
			return
		}
		peerCerts, err := parseCerts(bs)
		if err != nil {
			glog.Error(err)
			return
		}
		fmt.Println(peerCerts)
	}

	return
}
func (s *TLS) save() {
	st := s.getConnectionState(*s.args.SaveAddr)
	buf := bytes.NewBuffer(nil)
	cn := ""
	for _, v := range st.PeerCertificates {
		buf.Write(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: v.Raw}))
		if !v.IsCA {
			cn = strings.ReplaceAll(v.Subject.CommonName, "*", "_")
		}
	}

	if *s.args.SaveName == "" {
		*s.args.SaveName = cn + ".crt"
	}
	gfile.Write(*s.args.SaveName, buf.Bytes(), false)
	return
}

func (s *TLS) getConnectionState(addr string) tls.ConnectionState {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		glog.Error(err)
	}
	defer conn.Close()
	return conn.ConnectionState()
}
