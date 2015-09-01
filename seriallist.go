// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/facchinm/go-serial"
	"regexp"
)

type OsSerialPort struct {
	Name         string
	SerialNumber string
	DeviceClass  string
	Manufacturer string
	Product      string
	IdProduct    string
	IdVendor     string
	NetworkPort  bool
}

func GetList(network bool) ([]OsSerialPort, error) {

	//log.Println("Doing GetList()")

	if network {
		netportList, err := GetNetworkList()
		return netportList, err
	} else {

		// will timeout in 2 seconds
		ports, err := serial.GetPortsList()

		arrPorts := []OsSerialPort{}
		for _, element := range ports {
			arrPorts = append(arrPorts, OsSerialPort{Name: element})
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
				} else {
					log.Debug("serial port did not match. port: %v\n", element)
				}

			}
			arrPorts = newarrPorts
		}

		arrPorts = associateVidPidWithPort(arrPorts)
		return arrPorts, err
		//log.Printf("Done doing GetList(). arrPorts:%v\n", arrPorts)
	}

}
