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

// Supports Windows, Linux, Mac, BeagleBone Black, and Raspberry Pi

package main

import (
	"strings"
	"sync"
)

type serialhub struct {
	// Opened serial ports.
	ports map[*serport]bool
	mu    sync.Mutex

	OnRegister   func(port *serport)
	OnUnregister func(port *serport)
}

func newSerialHub() *serialhub {
	return &serialhub{
		ports: make(map[*serport]bool),
	}
}

// Register serial ports from the connections.
func (sh *serialhub) Register(port *serport) {
	sh.mu.Lock()
	//log.Print("Registering a port: ", p.portConf.Name)
	sh.OnRegister(port)
	sh.ports[port] = true
	sh.mu.Unlock()
}

// Unregister requests from connections.
func (sh *serialhub) Unregister(port *serport) {
	sh.mu.Lock()
	//log.Print("Unregistering a port: ", p.portConf.Name)
	sh.OnUnregister(port)
	delete(sh.ports, port)
	close(port.sendBuffered)
	close(port.sendNoBuf)
	sh.mu.Unlock()
}

func (sh *serialhub) FindPortByName(portname string) (*serport, bool) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	for port := range sh.ports {
		if strings.EqualFold(port.portConf.Name, portname) {
			// we found our port
			//spHandlerClose(port)
			return port, true
		}
	}
	return nil, false
}
