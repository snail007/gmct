package tlstool

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	glog "github.com/snail007/gmc/module/log"
	gfile "github.com/snail007/gmc/util/file"
	"github.com/snail007/gmct/tool"
)

type TLSArgs struct {
	InfoAddr       *string
	InfoServerName *string
	SaveAddr       *string
	SaveName       *string
	File           *string
	TLSName        *string
	SubName        *string
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
		h, _, _ := net.SplitHostPort(*s.args.InfoAddr)
		if *s.args.InfoServerName == "" {
			*s.args.InfoServerName = h
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
		if !gfile.Exists(*s.args.File) {
			glog.Errorf("file not found: %s", *s.args.File)
		}

		var blocks []byte
		rest := gfile.Bytes(*s.args.File)
		for {
			var block *pem.Block
			block, rest = pem.Decode(rest)
			if block == nil {
				glog.Error("Error: PEM not parsed")
				break
			}
			blocks = append(blocks, block.Bytes...)
			if len(rest) == 0 {
				break
			}
		}
		bs, err := x509.ParseCertificates(blocks)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}

		peerCerts, err := parseCerts(bs)
		if err != nil {
			glog.Error(err)
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
		if !v.IsCA {
			cn = strings.ReplaceAll(v.Subject.CommonName, "*", "_")
			break
		}
	}
	if cn == "" {
		glog.Error("CommonName is null")
	}

	folderName := *s.args.SaveName
	if folderName == "" {
		folderName = cn
	}
	os.Mkdir(folderName, 0755)

	for i, v := range st.PeerCertificates {
		pemTxt := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: v.Raw})
		buf.Write(pemTxt)
		n := ""
		if !v.IsCA {
			n = strings.ReplaceAll(v.Subject.CommonName, "*", "_")
		} else {
			n = "ca-" + fmt.Sprintf("%d", i)
		}
		gfile.Write(filepath.Join(folderName, n+".crt"), pemTxt, false)
	}
	gfile.Write(filepath.Join(folderName, "all.crt"), buf.Bytes(), false)
	glog.Info("SUCCESS!")
	return
}

func (s *TLS) getConnectionState(addr string) statusWrapper {
	requireClientCert := false
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName:         *s.args.InfoServerName,
		InsecureSkipVerify: true,
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			requireClientCert = true
			return &tls.Certificate{}, nil
		},
	})
	if err != nil {
		glog.Error(err)
	}
	defer conn.Close()
	return statusWrapper{
		ConnectionState:   conn.ConnectionState(),
		requireClientCert: requireClientCert,
	}
}

type statusWrapper struct {
	tls.ConnectionState
	requireClientCert bool
}
