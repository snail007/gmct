package ssht

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	gproxy "github.com/snail007/gmc/util/proxy"
	"github.com/snail007/gmct/tool"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/http/httpproxy"
	"gopkg.in/cheggaaa/pb.v1"
)

type SshArgs struct {
	File    *string
	SSHURL  *string
	Command *string

	SshURL *url.URL
}

func NewSshArgs() SshArgs {
	return SshArgs{
		File:   new(string),
		SSHURL: new(string),
	}
}

type Ssh struct {
	tool.GMCTool
	args SshArgs
}

func NewSsh() *Ssh {
	return &Ssh{}
}

func (s *Ssh) init(args0 interface{}) (err error) {
	s.args = args0.(SshArgs)

	if *s.args.SSHURL == "" {
		return fmt.Errorf("ssh info is required")
	} else if !strings.HasPrefix(*s.args.SSHURL, "ssh://") {
		*s.args.SSHURL = "ssh://" + *s.args.SSHURL
	}
	s.args.SshURL, err = url.Parse(*s.args.SSHURL)
	if err != nil {
		return
	}
	return
}

func (s *Ssh) Start(args interface{}) (err error) {
	err = s.init(args)
	if err != nil {
		return
	}
	if *s.args.File != "" {
		err = s.copy()
	}
	if *s.args.Command != "" {
		err = s.exec()
	}

	return
}

func (s *Ssh) copy() (err error) {
	a := strings.Split(*s.args.File, ":")
	if len(a) != 2 {
		return fmt.Errorf("error file format")
	}
	src := a[0]
	dest := a[1]
	command := fmt.Sprintf("cat - > \"%s.tmp\" && mv \"%s.tmp\" \"%s\"", dest, dest, dest)
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	client, err := s.client()
	if err != nil {
		return
	}
	// create bars
	pbPool := pb.NewPool()
	info, _ := in.Stat()
	bar := pb.New64(info.Size()).Prefix(fmt.Sprintf("%s [ %s ] ", filepath.Base(src), info.Name()))
	bar.ShowTimeLeft = true
	bar.ShowSpeed = true
	bar.SetUnits(pb.U_BYTES)
	reader := bar.NewProxyReader(in)
	pbPool.Add(bar)

	session, err := client.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return
	}
	go func() {
		io.Copy(os.Stderr, stderr)
	}()
	go func() {
		io.Copy(os.Stdout, stdout)
	}()

	stdin, err := session.StdinPipe()
	if err != nil {
		return
	}
	go func() {
		defer pbPool.Stop()
		io.Copy(stdin, reader)
		time.AfterFunc(time.Millisecond*200, func() { session.Close() })
	}()
	pbPool.Start()
	err = session.Run(command)
	if err != nil {
		return
	}
	return
}

func (s *Ssh) exec() (err error) {
	client, err := s.client()
	if err != nil {
		return
	}

	session, err := client.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return
	}
	go func() {
		io.Copy(os.Stderr, stderr)
	}()
	go func() {
		io.Copy(os.Stdout, stdout)
	}()

	cmd := *s.args.Command
	if strings.HasPrefix(cmd, "@") {
		file := cmd[1:]
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		cmd = string(content)
	}
	cmd = "echo -e '" + cmd + "'|bash"

	err = session.Run(*s.args.Command)
	if err != nil {
		return
	}
	return
}

func (s *Ssh) Stop() {
	return
}

func (s *Ssh) client() (client *ssh.Client, err error) {
	u := s.args.SshURL.User.Username()
	p, _ := s.args.SshURL.User.Password()
	if p == "" {
		p = os.Getenv("SSH_PASSWORD")
	}
	cfg := &ssh.ClientConfig{
		Timeout: 30000 * time.Millisecond,
		User:    u,
		Auth:    []ssh.AuthMethod{ssh.Password(p)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	proxy := httpproxy.FromEnvironment().HTTPProxy
	if proxy != "" {
		var j *gproxy.Jumper
		j, err = gproxy.NewJumper(proxy, cfg.Timeout)
		if err != nil {
			return nil, err
		}
		var conn net.Conn
		conn, err = j.Dial(s.args.SshURL.Host)
		if err != nil {
			return nil, err
		}
		var c ssh.Conn
		var chans <-chan ssh.NewChannel
		var reqs <-chan *ssh.Request
		c, chans, reqs, err = ssh.NewClientConn(conn, s.args.SshURL.Host, cfg)
		if err != nil {
			return nil, err
		}
		return ssh.NewClient(c, chans, reqs), nil
	}
	client, err = ssh.Dial("tcp", s.args.SshURL.Host, cfg)
	return
}
