// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"go.bug.st/serial.v1/enumerator"
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

	if network {
		netportList, err := GetNetworkList()
		return netportList, err
	}

	// will timeout in 2 seconds
	arrPorts := []OsSerialPort{}
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return arrPorts, err
	}

	for _, element := range ports {
		if element.IsUSB {
			vid := element.VID
			pid := element.PID
			vidString := fmt.Sprintf("0x%s", vid)
			pidString := fmt.Sprintf("0x%s", pid)
			if vid != "0000" && pid != "0000" {
				arrPorts = append(arrPorts, OsSerialPort{Name: element.Name, IdVendor: vidString, IdProduct: pidString, ISerial: element.SerialNumber})
			}
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
}

func findPortByName(portname string) (*serport, bool) {
	portnamel := strings.ToLower(portname)
	for port := range sh.ports {
		if strings.ToLower(port.portConf.Name) == portnamel {
			// we found our port
			//spHandlerClose(port)
			return port, true
		}
	}
	return nil, false
}

func findPortByNameRerun(portname string, network bool) (OsSerialPort, bool) {
	portnamel := strings.ToLower(portname)
	list, _ := GetList(network)
	for _, item := range list {
		if strings.ToLower(item.Name) == portnamel {
			return item, true
		}
	}
	return OsSerialPort{}, false
}
