package tools

import (
	"os/exec"
	"syscall"
	"unsafe"
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

func MessageBox(title, text string) int {
	var mod = syscall.NewLazyDLL("user32.dll")
	var proc = mod.NewProc("MessageBoxW")
	var MB_YESNO = 0x00000004

	ret, _, _ := proc.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(MB_YESNO))
	return int(ret)
}
