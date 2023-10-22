package tlstool

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/snail007/gmct/module/module"
	"github.com/snail007/gmct/util"
	"github.com/spf13/cobra"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	gproxy "github.com/snail007/gmc/util/proxy"

	glog "github.com/snail007/gmc/module/log"
	gfile "github.com/snail007/gmc/util/file"
)

func init() {
	module.AddCommand(func(root *cobra.Command) {
		s := NewTLS()
		cmd := &cobra.Command{
			Use:  "tls",
			Long: "tls certificate toolkit",
			PersistentPreRunE: func(c *cobra.Command, a []string) error {
				s.timeout = time.Second * 15
				proxy := util.Must(c.Flags().GetString("proxy")).String()
				if proxy != "" {
					var err error
					s.jumper, err = gproxy.NewJumper(proxy, s.timeout)
					if err != nil {
						glog.Fatal(err)
					}
				}
				return nil
			},
		}
		infoCMD := &cobra.Command{
			Use:  "info",
			Long: "print cert file or tls target host:port certificate info",
			Run: func(c *cobra.Command, a []string) {
				s.info(InfoArgs{
					Addr:       util.Must(c.Flags().GetString("addr")).String(),
					Proxy:      util.Must(c.Flags().GetString("proxy")).String(),
					ServerName: util.Must(c.Flags().GetString("servername")).String(),
					File:       util.Must(c.Flags().GetString("file")).String(),
				})
			},
		}
		infoCMD.Flags().StringP("addr", "a", "", "address of tls target, ip:port")
		infoCMD.Flags().StringP("proxy", "p", "", "proxy URL connect to address of tls target, example: http://127.0.0.1:8080")
		infoCMD.Flags().StringP("servername", "s", "", "the server name sent to tls server")
		infoCMD.Flags().StringP("file", "f", "", "path of tls certificate file")
		saveCMD := &cobra.Command{
			Use:  "save",
			Long: "save tls target host:port certificate to file",
			Run: func(c *cobra.Command, a []string) {
				s.save(SaveArgs{
					Addr:       util.Must(c.Flags().GetString("addr")).String(),
					Proxy:      util.Must(c.Flags().GetString("proxy")).String(),
					ServerName: util.Must(c.Flags().GetString("servername")).String(),
					FolderName: util.Must(c.Flags().GetString("name")).String(),
				})
			},
		}
		saveCMD.Flags().StringP("addr", "a", "", "address of tls target, ip:port")
		saveCMD.Flags().StringP("proxy", "p", "", "proxy URL connect to address of tls target, example: http://127.0.0.1:8080")
		saveCMD.Flags().StringP("servername", "s", "", "the server name sent to tls server")
		saveCMD.Flags().StringP("name", "n", "", "save certificate folder name")
		cmd.AddCommand(infoCMD)
		cmd.AddCommand(saveCMD)
		root.AddCommand(cmd)
	})
}

type SaveArgs struct {
	Addr       string
	Proxy      string
	ServerName string
	FolderName string
}

type InfoArgs struct {
	Addr       string
	Proxy      string
	ServerName string
	File       string
}

type TLS struct {
	timeout time.Duration
	jumper  *gproxy.Jumper
}

func NewTLS() *TLS {
	return &TLS{}
}

func (s *TLS) getAddr(addr string) string {
	a := addr
	if a != "" && !strings.Contains(a, ":") {
		a += ":443"
	}
	return a
}
func (s *TLS) getServerName(name, addr string) string {
	a := name
	if a != "" && !strings.Contains(a, ":") {
		a += ":443"
	}
	h, _, _ := net.SplitHostPort(addr)
	if a == "" {
		a = h
	}
	return a
}

func (s *TLS) info(args InfoArgs) {
	if args.Addr != "" {
		info, err := getTLSInfo(s.getConnectionState(args.Addr, args.ServerName))
		if err != nil {
			glog.Panic(err)
		}
		fmt.Println(info.String())
	} else if args.File != "" {
		if !gfile.Exists(args.File) {
			glog.Fatalf("file not found: %s", args.File)
		}
		var blocks []byte
		rest := gfile.Bytes(args.File)
		for {
			var block *pem.Block
			block, rest = pem.Decode(rest)
			if block == nil {
				glog.Panic("Error: PEM not parsed")
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
			glog.Panic(err)
		}
		fmt.Println(peerCerts)
	}

	return
}
func (s *TLS) save(args SaveArgs) {
	st := s.getConnectionState(args.Addr, args.ServerName)
	buf := bytes.NewBuffer(nil)
	folderName := args.FolderName
	if folderName == "" {
		folderName = strings.Replace(args.Addr, ":", "_", -1)
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
		glog.Panic(errors.Wrap(err, "dial tcp fail"))
	}
	defer tcpConn.Close()
	tlsConn := tls.Client(tcpConn, cfg)
	err = tlsConn.Handshake()
	if err != nil {
		glog.Panic(errors.Wrap(err, "tls handshake fail, maybe not a tls server?"))
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
