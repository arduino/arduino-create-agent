// Copyright 2022 Arduino SA
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.bug.st/serial/enumerator"
)

// OsSerialPort is the Os serial port
type OsSerialPort struct {
	Name         string
	DeviceClass  string
	Manufacturer string
	Product      string
	IDProduct    string
	IDVendor     string
	ISerial      string
}

// enumerateSerialPorts will return the OS serial port
func enumerateSerialPorts() ([]OsSerialPort, error) {
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
				arrPorts = append(arrPorts, OsSerialPort{Name: element.Name, IDVendor: vidString, IDProduct: pidString, ISerial: element.SerialNumber})
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
				log.Debugf("serial port did not match. port: %v\n", element)
			}

		}
		arrPorts = newarrPorts
	}

	return arrPorts, err
}

func findPortByName(portname string) (*serport, bool) {
	portnamel := strings.ToLower(portname)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	for port := range sh.ports {
		if strings.ToLower(port.portConf.Name) == portnamel {
			// we found our port
			//spHandlerClose(port)
			return port, true
		}
	}
	return nil, false
}
