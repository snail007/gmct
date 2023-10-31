package exec

import "os/exec"

func killCmd(c *exec.Cmd) {
	c.Process.Kill()
}
