// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"github.com/facchinm/go-serial"
	"log"
	//"os"
	"regexp"
)

type OsSerialPort struct {
	Name         string
	FriendlyName string
	RelatedNames []string // for some devices there are 2 or more ports, i.e. TinyG v9 has 2 serial ports
	SerialNumber string
	DeviceClass  string
	Manufacturer string
	Product      string
	IdProduct    string
	IdVendor     string
	NetworkPort  bool
}

func GetList() ([]OsSerialPort, error) {

	//log.Println("Doing GetList()")

	// will timeout in 2 seconds
	ports, err := serial.GetPortsList()

	arrPorts := []OsSerialPort{}
	for _, element := range ports {
		arrPorts = append(arrPorts, OsSerialPort{Name: element, FriendlyName: element})
	}

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

	netportList, _ := GetNetworkList()
	arrPorts = append(arrPorts, netportList...)

	//log.Printf("Done doing GetList(). arrPorts:%v\n", arrPorts)

	return arrPorts, err
}
