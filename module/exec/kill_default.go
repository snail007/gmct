//go:build !windows
// +build !windows

package exec

import (
	"os/exec"
	"syscall"
)

func killCmd(c *exec.Cmd) {
	syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
}
