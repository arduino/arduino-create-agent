package utilities

import (
	"os/exec"
	"syscall"
)

func TellCommandNotToSpawnShell(oscmd *exec.Cmd) {
	oscmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
