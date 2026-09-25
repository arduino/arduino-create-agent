package updater

import (
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func elevate() error {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, err := syscall.UTF16PtrFromString(verb)
	if err != nil {
		return err
	}
	exePtr, err := syscall.UTF16PtrFromString(exe)
	if err != nil {
		return err
	}
	cwdPtr, err := syscall.UTF16PtrFromString(cwd)
	if err != nil {
		return err
	}
	argPtr, _ := syscall.UTF16PtrFromString(args)
	var showCmd int32 = 1
	return windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
}
