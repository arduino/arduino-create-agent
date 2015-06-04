package main

import (
	//"fmt"
	//"github.com/tarm/goserial"
	"log"
	"os"
	"strings"
	//"encoding/binary"
	//"strconv"
	//"syscall"
	//"fmt"
	//"bufio"
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

func removeNonArduinoBoards(ports []OsSerialPort) []OsSerialPort {
	usbcmd := exec.Command("system_profiler", "SPUSBDataType")
	grepcmd := exec.Command("grep", "0x2341", "-A5", "-B1")

	cmdOutput, _ := pipe_commands(usbcmd, grepcmd)

	//log.Println(string(cmdOutput))
	cmdOutSlice := strings.Split(string(cmdOutput), "\n")

	var arduino_ports []OsSerialPort
	other_ports := ports

	// how many lines is the output? boards attached = lines/8
	for i := 0; i < len(cmdOutSlice)/8; i++ {

		cmdOutSliceN := cmdOutSlice[i*8 : (i+1)*8]

		cmdOutMap := make(map[string]string)

		for _, element := range cmdOutSliceN {
			if strings.Contains(element, "ID") {
				element = strings.TrimSpace(element)
				arr := strings.Split(element, ": ")
				cmdOutMap[arr[0]] = arr[1]
			}
		}

		archBoardName, boardName, _ := getBoardName(cmdOutMap["Product ID"])

		// remove initial 0x and final zeros
		ttyHeader := strings.Trim((cmdOutMap["Location ID"]), "0x")
		ttyHeader = strings.Split(ttyHeader, " ")[0]
		ttyHeader = strings.Trim(ttyHeader, "0")

		for i, port := range ports {
			if strings.Contains(port.Name, ttyHeader) && !strings.Contains(port.Name, "/cu.") {
				port.RelatedNames = append(port.RelatedNames, archBoardName)
				port.FriendlyName = strings.Trim(boardName, "\n")
				arduino_ports = append(arduino_ports, port)
				other_ports = removePortFromSlice(other_ports, i)
			}
		}
	}

	arduino_ports = append(arduino_ports, other_ports...)

	return arduino_ports
}

func removePortFromSlice(a []OsSerialPort, i int) []OsSerialPort {
	copy(a[i:], a[i+1:])
	a = a[:len(a)-1]
	return a
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
