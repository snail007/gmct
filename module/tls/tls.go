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
	"time"

	"github.com/pkg/errors"
	gproxy "github.com/snail007/gmc/util/proxy"

	glog "github.com/snail007/gmc/module/log"
	gfile "github.com/snail007/gmc/util/file"
	"github.com/snail007/gmct/tool"
)

type SaveArgs struct {
	Addr       *string
	Proxy      *string
	ServerName *string
	FolderName *string
}

type InfoArgs struct {
	Addr       *string
	Proxy      *string
	ServerName *string
	File       *string
}

type TLSArgs struct {
	Save    *SaveArgs
	Info    *InfoArgs
	SubName *string
}

func NewTLSArgs() TLSArgs {
	return TLSArgs{
		Save:    new(SaveArgs),
		Info:    new(InfoArgs),
		SubName: new(string),
	}
}

type TLS struct {
	tool.GMCTool
	cfg     TLSArgs
	timeout time.Duration
	jumper  *gproxy.Jumper
}

func NewTLS() *TLS {
	return &TLS{}
}

func (s *TLS) init(args0 interface{}) (err error) {
	s.cfg = args0.(TLSArgs)
	s.timeout = time.Second * 15
	proxy := ""
	switch *s.cfg.SubName {
	case "info":
		proxy = *s.cfg.Info.Proxy
		if *s.cfg.Info.Addr != "" && !strings.Contains(*s.cfg.Info.Addr, ":") {
			*s.cfg.Info.Addr += ":443"
		}
		h, _, _ := net.SplitHostPort(*s.cfg.Info.Addr)
		if *s.cfg.Info.ServerName == "" {
			*s.cfg.Info.ServerName = h
		}
	case "save":
		proxy = *s.cfg.Save.Proxy
		if *s.cfg.Save.Addr != "" && !strings.Contains(*s.cfg.Save.Addr, ":") {
			*s.cfg.Save.Addr += ":443"
		}
		h, _, _ := net.SplitHostPort(*s.cfg.Save.Addr)
		if *s.cfg.Save.ServerName == "" {
			*s.cfg.Save.ServerName = h
		}
	}
	if proxy != "" {
		var err error
		s.jumper, err = gproxy.NewJumper(proxy, s.timeout)
		if err != nil {
			glog.Error(err)
		}
	}
	return
}

func (s *TLS) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	switch *s.cfg.SubName {
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
	if *s.cfg.Info.Addr != "" {
		info, err := getTLSInfo(s.getConnectionState(*s.cfg.Info.Addr, *s.cfg.Info.ServerName))
		if err != nil {
			glog.Error(err)
		}
		fmt.Println(info.String())
	} else if *s.cfg.Info.File != "" {
		if !gfile.Exists(*s.cfg.Info.File) {
			glog.Panicf("file not found: %s", *s.cfg.Info.File)
		}
		var blocks []byte
		rest := gfile.Bytes(*s.cfg.Info.File)
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
	st := s.getConnectionState(*s.cfg.Save.Addr, *s.cfg.Save.ServerName)
	buf := bytes.NewBuffer(nil)
	folderName := *s.cfg.Save.FolderName
	if folderName == "" {
		folderName = strings.Replace(*s.cfg.Save.Addr, ":", "_", -1)
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
	glog.Infof("SAVE TO %s SUCCESS!", folderName)
	return
}

func (s *TLS) getConnectionState(addr, serverName string) statusWrapper {
	requireClientCert := false
	cfg := &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: true,
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			requireClientCert = true
			return &tls.Certificate{}, nil
		},
	}
	var tcpConn net.Conn
	var err error
	if s.jumper != nil {
		tcpConn, err = s.jumper.Dial(addr)
	} else {
		tcpConn, err = net.DialTimeout("tcp", addr, s.timeout)
	}
	if err != nil {
		glog.Error(errors.Wrap(err, "dial tcp fail"))
	}
	defer tcpConn.Close()
	tlsConn := tls.Client(tcpConn, cfg)
	err = tlsConn.Handshake()
	if err != nil {
		glog.Error(errors.Wrap(err, "tls handshake fail, maybe not a tls server?"))
	}
	return statusWrapper{
		ConnectionState:   tlsConn.ConnectionState(),
		requireClientCert: requireClientCert,
	}
}

type statusWrapper struct {
	tls.ConnectionState
	requireClientCert bool
}
