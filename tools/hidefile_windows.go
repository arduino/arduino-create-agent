package tools

import (
	"os/exec"
	"syscall"
)

func hideFile(path string) {
	cpath, cpathErr := syscall.UTF16PtrFromString(path)
	if cpathErr != nil {
	}
	syscall.SetFileAttributes(cpath, syscall.FILE_ATTRIBUTE_HIDDEN)
}

func TellCommandNotToSpawnShell(oscmd *exec.Cmd) {
	oscmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
