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
	"slices"

	log "github.com/sirupsen/logrus"
	"go.bug.st/serial/enumerator"
)

// OsSerialPort is the Os serial port
type OsSerialPort struct {
	Name         string
	PID          string
	VID          string
	SerialNumber string
}

// enumerateSerialPorts will return the OS serial port
func enumerateSerialPorts() ([]*OsSerialPort, error) {
	// will timeout in 2 seconds
	arrPorts := []*OsSerialPort{}
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return arrPorts, err
	}

	for _, element := range ports {
		if element.IsUSB {
			vid, pid := "0x"+element.VID, "0x"+element.PID
			if vid != "0x0000" && pid != "0x0000" {
				arrPorts = append(arrPorts, &OsSerialPort{
					Name:         element.Name,
					VID:          vid,
					PID:          pid,
					SerialNumber: element.SerialNumber,
				})
			}
		}
	}

	// see if we should filter the list
	if portsFilter != nil {
		arrPorts = slices.DeleteFunc(arrPorts, func(port *OsSerialPort) bool {
			match := portsFilter.MatchString(port.Name)
			if !match {
				log.Debugf("ignoring port not matching filter. port: %v\n", port)
			}
			return match
		})
	}

	return arrPorts, err
}
