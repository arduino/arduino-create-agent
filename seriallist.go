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
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}

	var res []*OsSerialPort
	for _, port := range ports {
		if !port.IsUSB {
			continue
		}
		vid, pid := "0x"+port.VID, "0x"+port.PID
		if vid == "0x0000" || pid == "0x0000" {
			continue
		}
		if portsFilter != nil && !portsFilter.MatchString(port.Name) {
			log.Debugf("ignoring port not matching filter. port: %v\n", port)
			continue
		}
		res = append(res, &OsSerialPort{
			Name:         port.Name,
			VID:          vid,
			PID:          pid,
			SerialNumber: port.SerialNumber,
		})
	}
	return res, err
}
