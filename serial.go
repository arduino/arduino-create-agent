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
	"encoding/json"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/arduino/arduino-create-agent/tools"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
	"github.com/sirupsen/logrus"
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

func (h *hub) spErr(err string) {
	//log.Println("Sending err back: ", err)
	//sh.hub.broadcastSys <- []byte(err)
	h.broadcastSys <- []byte("{\"Error\" : \"" + err + "\"}")
}

func (h *hub) spClose(portname string) {
	if myport, ok := h.serialHub.FindPortByName(portname); ok {
		h.broadcastSys <- []byte("Closing serial port " + portname)
		myport.Close()
	} else {
		h.spErr("We could not find the serial port " + portname + " that you were trying to close.")
	}
}

func (h *hub) spWrite(arg string) {
	// we will get a string of comXX asdf asdf asdf
	//log.Println("Inside spWrite arg: " + arg)
	arg = strings.TrimPrefix(arg, " ")
	//log.Println("arg after trim: " + arg)
	args := strings.SplitN(arg, " ", 3)
	if len(args) != 3 {
		errstr := "Could not parse send command: " + arg
		//log.Println(errstr)
		h.spErr(errstr)
		return
	}
	bufferingMode := args[0]
	portname := strings.Trim(args[1], " ")
	data := args[2]

	//log.Println("The port to write to is:" + portname + "---")
	//log.Println("The data is:" + data + "---")

	// See if we have this port open
	port, ok := h.serialHub.FindPortByName(portname)
	if !ok {
		// we couldn't find the port, so send err
		h.spErr("We could not find the serial port " + portname + " that you were trying to write to.")
		return
	}

	// see if bufferingMode is valid
	switch bufferingMode {
	case "send", "sendnobuf", "sendraw":
		// valid buffering mode, go ahead
	default:
		h.spErr("Unsupported send command:" + args[0] + ". Please specify a valid one")
		return
	}

	// send it to the write channel
	port.Write(data, bufferingMode)
}
