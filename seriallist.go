// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"log"
	"os"
	"regexp"
)

type OsSerialPort struct {
	Name         string
	FriendlyName string
	RelatedNames []string // for some devices there are 2 or more ports, i.e. TinyG v9 has 2 serial ports
}

func GetList() ([]OsSerialPort, os.SyscallError) {

	//log.Println("Doing GetList()")

	arrPorts, err := getList()

	// see if we should filter the list
	if len(*regExpFilter) > 0 {
		// yes, user asked for a filter
		reFilter := regexp.MustCompile("(?i)" + *regExpFilter)

		newarrPorts := []OsSerialPort{}
		for _, element := range arrPorts {
			// if matches regex, include
			if reFilter.MatchString(element.Name) {
				newarrPorts = append(newarrPorts, element)
			} else if reFilter.MatchString(element.FriendlyName) {
				newarrPorts = append(newarrPorts, element)
			} else {
				log.Printf("serial port did not match. port: %v\n", element)
			}

		}
		arrPorts = newarrPorts
	}

	//log.Printf("Done doing GetList(). arrPorts:%v\n", arrPorts)

	return arrPorts, err
}
