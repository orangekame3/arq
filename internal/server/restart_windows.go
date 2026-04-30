//go:build windows

package server

import "os/exec"

// detachProcess is a no-op on Windows.
func detachProcess(cmd *exec.Cmd) {}
