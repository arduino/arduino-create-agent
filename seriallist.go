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
	"strings"

	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/arduino/discovery/discoverymanager"
)

var discoveryManager *discoverymanager.DiscoveryManager

// OsSerialPort is the Os serial port
type OsSerialPort struct {
	Name         string
	DeviceClass  string
	Manufacturer string
	Product      string
	IDProduct    string
	IDVendor     string
	ISerial      string
	NetworkPort  bool
}

// GetList will return the OS serial port
func GetList() ([]OsSerialPort, error) {

	if discoveryManager == nil {
		discoveryManager = discoverymanager.New()
		Tools.Download("builtin", "serial-discovery", "latest", "keep")
		Tools.Download("builtin", "mdns-discovery", "latest", "keep")
		sd, err := Tools.GetLocation("serial-discovery")
		if err == nil {
			d := discovery.New("serial", sd+"/serial-discovery")
			discoveryManager.Add(d)
		}
		md, err := Tools.GetLocation("mdns-discovery")
		if err == nil {
			d := discovery.New("mdns", md+"/mdns-discovery")
			discoveryManager.Add(d)
		}
		discoveryManager.Start()
	}

	fmt.Println("calling getList")
	ports := discoveryManager.List()

	fmt.Println(ports)

	arrPorts := []OsSerialPort{}
	for _, port := range ports {
		vid := port.Properties.AsMap()["vid"]
		pid := port.Properties.AsMap()["pid"]
		arrPorts = append(arrPorts, OsSerialPort{Name: port.Address, IDVendor: vid, IDProduct: pid, ISerial: port.HardwareID})
	}
	fmt.Println(arrPorts)

	return arrPorts, nil
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
