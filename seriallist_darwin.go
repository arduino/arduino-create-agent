package main

import (
	"strings"
)

func ExtraFilterPorts(ports []OsSerialPort) []OsSerialPort {

	// prefilter ports
	ports = Filter(ports, func(port OsSerialPort) bool {
		return !strings.Contains(port.Name, "Blue") && !strings.Contains(port.Name, "/cu")
	})
	return ports
}
