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
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

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

type serialPortList struct {
	Ports     []*SpPortItem
	portsLock sync.Mutex

	OnList func([]byte) `json:"-"`
	OnErr  func(string) `json:"-"`
}

func newSerialPortList() *serialPortList {
	return &serialPortList{}
}

// SpPortItem is the serial port item
type SpPortItem struct {
	Name            string
	SerialNumber    string
	DeviceClass     string
	IsOpen          bool
	IsPrimary       bool
	Baud            int
	BufferAlgorithm string
	Ver             string
	VendorID        string
	ProductID       string
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

// List broadcasts a Json representation of the ports found
func (sp *serialPortList) List() {
	sp.portsLock.Lock()
	ls, err := json.MarshalIndent(sp, "", "\t")
	sp.portsLock.Unlock()

	if err != nil {
		sp.OnErr("Error creating json on port list " + err.Error())
	} else {
		sp.OnList(ls)
	}
}

// Run is the main loop for port discovery and management
func (sp *serialPortList) Run() {
	for retries := 0; retries < 10; retries++ {
		sp.runSerialDiscovery()

		logrus.Errorf("Serial discovery stopped working, restarting it in 10 seconds...")
		time.Sleep(10 * time.Second)
	}
	logrus.Errorf("Failed restarting serial discovery. Giving up...")
}

func (sp *serialPortList) runSerialDiscovery() {
	// First ensure that all the discoveries are available
	if err := Tools.Download("builtin", "serial-discovery", "latest", "keep"); err != nil {
		logrus.Errorf("Error downloading serial-discovery: %s", err)
		panic(err)
	}
	sd, err := Tools.GetLocation("serial-discovery")
	if err != nil {
		logrus.Errorf("Error downloading serial-discovery: %s", err)
		panic(err)
	}
	d := discovery.NewClient("serial", sd+"/serial-discovery")
	dLogger := logrus.WithField("discovery", "serial")
	if *verbose {
		d.SetLogger(dLogger)
	}
	d.SetUserAgent("arduino-create-agent/" + version)
	if err := d.Run(); err != nil {
		logrus.Errorf("Error running serial-discovery: %s", err)
		panic(err)
	}
	defer d.Quit()

	events, err := d.StartSync(10)
	if err != nil {
		logrus.Errorf("Error starting event watcher on serial-discovery: %s", err)
		panic(err)
	}

	logrus.Infof("Serial discovery started, watching for events")
	for ev := range events {
		logrus.WithField("event", ev).Debugf("Serial discovery event")
		switch ev.Type {
		case "add":
			sp.add(ev.Port)
		case "remove":
			sp.remove(ev.Port)
		}
	}

	sp.reset()
	logrus.Errorf("Serial discovery stopped.")
}

func (sp *serialPortList) reset() {
	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()
	sp.Ports = []*SpPortItem{}
}

func (sp *serialPortList) add(addedPort *discovery.Port) {
	if addedPort.Protocol != "serial" {
		return
	}
	props := addedPort.Properties
	if !props.ContainsKey("vid") {
		return
	}
	vid, pid := props.Get("vid"), props.Get("pid")
	if vid == "0x0000" || pid == "0x0000" {
		return
	}
	if portsFilter != nil && !portsFilter.MatchString(addedPort.Address) {
		logrus.Debugf("ignoring port not matching filter. port: %v\n", addedPort.Address)
		return
	}

	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()

	// If the port is already in the list, just update the metadata...
	for _, oldPort := range sp.Ports {
		fmt.Println("oldPort.Name: ", oldPort.Name)
		if oldPort.Name == addedPort.Address {
			oldPort.SerialNumber = props.Get("serialNumber")
			oldPort.VendorID = vid
			oldPort.ProductID = pid
			return
		}
	}
	// ...otherwise, add it to the list
	sp.Ports = append(sp.Ports, &SpPortItem{
		Name:            addedPort.Address,
		SerialNumber:    props.Get("serialNumber"),
		VendorID:        vid,
		ProductID:       pid,
		Ver:             version,
		IsOpen:          false,
		IsPrimary:       false,
		Baud:            0,
		BufferAlgorithm: "",
	})
}

func (sp *serialPortList) remove(removedPort *discovery.Port) {
	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()

	// Remove the port from the list
	sp.Ports = slices.DeleteFunc(sp.Ports, func(oldPort *SpPortItem) bool {
		return oldPort.Name == removedPort.Address
	})
}

// MarkPortAsOpened marks a port as opened by the user
func (sp *serialPortList) MarkPortAsOpened(portname string) {
	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()
	port := sp.getPortByName(portname)
	if port != nil {
		port.IsOpen = true
	}
}

// MarkPortAsClosed marks a port as no more opened by the user
func (sp *serialPortList) MarkPortAsClosed(portname string) {
	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()
	port := sp.getPortByName(portname)
	if port != nil {
		port.IsOpen = false
	}
}

func (sp *serialPortList) getPortByName(portname string) *SpPortItem {
	for _, port := range sp.Ports {
		if port.Name == portname {
			return port
		}
	}
	return nil
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
