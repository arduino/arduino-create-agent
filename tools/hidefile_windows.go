package tools

import "syscall"

func hideFile(path string) {
	cpath, cpathErr := syscall.UTF16PtrFromString(path)
	if cpathErr != nil {
	}
	syscall.SetFileAttributes(cpath, syscall.FILE_ATTRIBUTE_HIDDEN)
}
