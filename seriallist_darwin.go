package main

import (
	//"fmt"
	//"github.com/tarm/goserial"
	log "github.com/Sirupsen/logrus"
	"os"
	"strings"
	//"encoding/binary"
	//"strconv"
	//"syscall"
	//"fmt"
	//"bufio"
	"fmt"
	"io/ioutil"
	"os/exec"
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

		usbcmd := exec.Command("system_profiler", "SPUSBDataType")
		grepcmd := exec.Command("grep", "Location ID: 0x"+port_hash[:len(port_hash)-1], "-B6")
		cmdOutput, _ := pipe_commands(usbcmd, grepcmd)

		if len(cmdOutput) == 0 {
			usbcmd = exec.Command("system_profiler", "SPUSBDataType")
			grepcmd = exec.Command("grep" /*"Serial Number: "+*/, strings.Trim(port_hash, "0"), "-B3", "-A3")
			cmdOutput, _ = pipe_commands(usbcmd, grepcmd)

			fmt.Println(string(cmdOutput))
		}

		if len(cmdOutput) == 0 {
			//give up
			continue
		}

		cmdOutSlice := strings.Split(string(cmdOutput), "\n")

		fmt.Println(cmdOutSlice)

		cmdOutMap := make(map[string]string)

		for _, element := range cmdOutSlice {
			if strings.Contains(element, "ID") || strings.Contains(element, "Manufacturer") {
				element = strings.TrimSpace(element)
				arr := strings.Split(element, ": ")
				cmdOutMap[arr[0]] = arr[1]
			}
		}
		ports[index].IdProduct = strings.Split(cmdOutMap["Product ID"], " ")[0]
		ports[index].IdVendor = strings.Split(cmdOutMap["Vendor ID"], " ")[0]
		ports[index].Manufacturer = cmdOutMap["Manufacturer"]
	}
	return ports
}

func getList() ([]OsSerialPort, os.SyscallError) {
	//return getListViaWmiPnpEntity()
	return getListViaTtyList()
}

func getListViaTtyList() ([]OsSerialPort, os.SyscallError) {
	var err os.SyscallError

	log.Println("getting serial list on darwin")

	// make buffer of 100 max serial ports
	// return a slice
	list := make([]OsSerialPort, 100)

	files, _ := ioutil.ReadDir("/dev/")
	ctr := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "tty.") {
			// it is a legitimate serial port
			list[ctr].Name = "/dev/" + f.Name()
			list[ctr].FriendlyName = f.Name()
			log.Println("Added serial port to list: ", list[ctr])
			ctr++
		}
		// stop-gap in case going beyond 100 (which should never happen)
		// i mean, really, who has more than 100 serial ports?
		if ctr > 99 {
			ctr = 99
		}
		//fmt.Println(f.Name())
		//fmt.Println(f.)
	}
	/*
		list := make([]OsSerialPort, 3)
		list[0].Name = "tty.serial1"
		list[0].FriendlyName = "tty.serial1"
		list[1].Name = "tty.serial2"
		list[1].FriendlyName = "tty.serial2"
		list[2].Name = "tty.Bluetooth-Modem"
		list[2].FriendlyName = "tty.Bluetooth-Modem"
	*/

	return list[0:ctr], err
}
