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
	"io/ioutil"
)

func getList() ([]OsSerialPort, os.SyscallError) {
	//return getListViaWmiPnpEntity()
	return getListViaTtyList()
}

func getListViaTtyList() ([]OsSerialPort, os.SyscallError) {
	var err os.SyscallError

	log.Println("getting serial list on darwin")

	// make buffer of 1000 max serial ports
	// return a slice
	list := make([]OsSerialPort, 1000)

	files, _ := ioutil.ReadDir("/dev/")
	ctr := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "tty") {
			// it is a legitimate serial port
			list[ctr].Name = "/dev/" + f.Name()
			list[ctr].FriendlyName = f.Name()
			log.Println("Added serial port to list: ", list[ctr])
			ctr++
		}
		// stop-gap in case going beyond 1000 (which should never happen)
		// i mean, really, who has more than 1000 serial ports?
		if ctr > 999 {
			ctr = 999
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
