// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/facchinm/go-serial-native"
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
	ISerial      string
	NetworkPort  bool
}

func GetList(network bool) ([]OsSerialPort, error) {

	//log.Println("Doing GetList()")

	if network {
		netportList, err := GetNetworkList()
		return netportList, err
	} else {

		// will timeout in 2 seconds
		arrPorts := []OsSerialPort{}
		ports, err := serial.ListPorts()
		if err != nil {
			return arrPorts, err
		}

		for _, element := range ports {
			vid, pid, _ := element.USBVIDPID()
			vidString := fmt.Sprintf("0x%04X", vid)
			pidString := fmt.Sprintf("0x%04X", pid)
			if vid != 0 && pid != 0 {
				arrPorts = append(arrPorts, OsSerialPort{Name: element.Name(), IdVendor: vidString, IdProduct: pidString, ISerial: element.USBSerialNumber()})
			}
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

		return arrPorts, err
		//log.Printf("Done doing GetList(). arrPorts:%v\n", arrPorts)
	}

}
