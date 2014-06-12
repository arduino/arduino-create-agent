// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"os"
)

type OsSerialPort struct {
	Name         string
	FriendlyName string
}

func GetList() ([]OsSerialPort, os.SyscallError) {
	return getList()
}
