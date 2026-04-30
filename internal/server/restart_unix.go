//go:build !windows

package server

import (
	"os/exec"
	"syscall"
)

// detachProcess sets up the command to run in a new session so it survives parent exit.
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
