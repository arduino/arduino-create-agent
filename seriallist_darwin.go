package main

import (
	"os/exec"
	"strings"
)

// execute system_profiler SPUSBDataType | grep "Vendor ID: 0x2341" -A5 -B2
// maybe -B2 is not necessary
// trim whitespace and eol
// map everything with :
// get [map][Location ID] first 5 chars
// search all board.txt files for map[Product ID]
// assign it to name

func associateVidPidWithPort(ports []OsSerialPort) []OsSerialPort {

	// prefilter ports
	ports = Filter(ports, func(port OsSerialPort) bool {
		return !strings.Contains(port.Name, "Blue") && !strings.Contains(port.Name, "/cu")
	})

	for index, _ := range ports {
		port_hash := strings.Trim(ports[index].Name, "/dev/tty.usbmodem")
		port_hash = strings.Trim(port_hash, "/dev/tty.usbserial-")

		port_hash = strings.ToLower(port_hash)

		usbcmd := exec.Command("system_profiler", "SPUSBDataType")
		grepcmd := exec.Command("grep", "Location ID: 0x"+port_hash[:len(port_hash)-1], "-B6")
		cmdOutput, _ := utillities.PipeCommands(usbcmd, grepcmd)

		if len(cmdOutput) == 0 {
			usbcmd = exec.Command("system_profiler", "SPUSBDataType")
			grepcmd = exec.Command("grep" /*"Serial Number: "+*/, strings.Trim(port_hash, "0"), "-B3", "-A3")
			cmdOutput, _ = utillities.PipeCommands(usbcmd, grepcmd)
		}

		if len(cmdOutput) == 0 {
			//give up
			continue
		}

		cmdOutSlice := strings.Split(string(cmdOutput), "\n")

		cmdOutMap := make(map[string]string)

		for _, element := range cmdOutSlice {
			if strings.Contains(element, "ID") || strings.Contains(element, "Manufacturer") {
				element = strings.TrimSpace(element)
				arr := strings.Split(element, ": ")
				if len(arr) > 1 {
					cmdOutMap[arr[0]] = arr[1]
				} else {
					cmdOutMap[arr[0]] = ""
				}
			}
		}
		ports[index].IdProduct = strings.Split(cmdOutMap["Product ID"], " ")[0]
		ports[index].IdVendor = strings.Split(cmdOutMap["Vendor ID"], " ")[0]
		ports[index].Manufacturer = cmdOutMap["Manufacturer"]
	}
	return ports
}

func tellCommandNotToSpawnShell(_ *exec.Cmd) {
}
