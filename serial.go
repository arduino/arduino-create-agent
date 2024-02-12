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
	"strconv"
	"strings"
	"sync"
	"time"

	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
	"github.com/sirupsen/logrus"
)

type writeRequest struct {
	p      *serport
	d      string
	buffer string
}

type serialhub struct {
	// Opened serial ports.
	ports map[*serport]bool

	//write chan *serport, chan []byte
	write chan writeRequest

	// Register requests from the connections.
	register chan *serport

	// Unregister requests from connections.
	unregister chan *serport

	mu sync.Mutex
}

// SerialPortList is the serial port list
type SerialPortList struct {
	Ports     []*SpPortItem
	portsLock sync.Mutex
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

// serialPorts contains the ports attached to the machine
var serialPorts SerialPortList

var sh = serialhub{
	//write:   	make(chan *serport, chan []byte),
	write:      make(chan writeRequest),
	register:   make(chan *serport),
	unregister: make(chan *serport),
	ports:      make(map[*serport]bool),
}

func (sh *serialhub) run() {

	//log.Print("Inside run of serialhub")
	//cmdIdCtr := 0

	for {
		select {
		case p := <-sh.register:
			sh.mu.Lock()
			//log.Print("Registering a port: ", p.portConf.Name)
			h.broadcastSys <- []byte("{\"Cmd\":\"Open\",\"Desc\":\"Got register/open on port.\",\"Port\":\"" + p.portConf.Name + "\",\"Baud\":" + strconv.Itoa(p.portConf.Baud) + ",\"BufferType\":\"" + p.BufferType + "\"}")
			sh.ports[p] = true
			sh.mu.Unlock()
		case p := <-sh.unregister:
			sh.mu.Lock()
			//log.Print("Unregistering a port: ", p.portConf.Name)
			h.broadcastSys <- []byte("{\"Cmd\":\"Close\",\"Desc\":\"Got unregister/close on port.\",\"Port\":\"" + p.portConf.Name + "\",\"Baud\":" + strconv.Itoa(p.portConf.Baud) + "}")
			delete(sh.ports, p)
			close(p.sendBuffered)
			close(p.sendNoBuf)
			sh.mu.Unlock()
		case wr := <-sh.write:
			// if user sent in the commands as one text mode line
			switch wr.buffer {
			case "send":
				wr.p.sendBuffered <- wr.d
			case "sendnobuf":
				wr.p.sendNoBuf <- []byte(wr.d)
			case "sendraw":
				wr.p.sendRaw <- wr.d
			}
			// no default since we alredy verified in spWrite()
		}
	}
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
func (sp *SerialPortList) List() {
	sp.portsLock.Lock()
	ls, err := json.MarshalIndent(sp, "", "\t")
	sp.portsLock.Unlock()

	if err != nil {
		//log.Println(err)
		h.broadcastSys <- []byte("Error creating json on port list " +
			err.Error())
	} else {
		h.broadcastSys <- ls
	}
}

// Run is the main loop for port discovery and management
func (sp *SerialPortList) Run() {
	for retries := 0; retries < 10; retries++ {
		sp.runSerialDiscovery()

		logrus.Errorf("Serial discovery stopped working, restarting it in 10 seconds...")
		time.Sleep(10 * time.Second)
	}
	logrus.Errorf("Failed restarting serial discovery. Giving up...")
}

func (sp *SerialPortList) runSerialDiscovery() {
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
		logrus.Errorf("Error downloading serial-discovery: %s", err)
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

func (sp *SerialPortList) reset() {
	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()
	sp.Ports = []*SpPortItem{}
}

func (sp *SerialPortList) add(addedPort *discovery.Port) {
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

func (sp *SerialPortList) remove(removedPort *discovery.Port) {
	sp.portsLock.Lock()
	defer sp.portsLock.Unlock()

	// Remove the port from the list
	sp.Ports = slices.DeleteFunc(sp.Ports, func(oldPort *SpPortItem) bool {
		return oldPort.Name == removedPort.Address
	})
}

func spErr(err string) {
	//log.Println("Sending err back: ", err)
	//h.broadcastSys <- []byte(err)
	h.broadcastSys <- []byte("{\"Error\" : \"" + err + "\"}")
}

func spClose(portname string) {
	// look up the registered port by name
	// then call the close method inside serialport
	// that should cause an unregister channel call back
	// to myself

	myport, isFound := sh.FindPortByName(portname)

	if isFound {
		// we found our port
		spHandlerClose(myport)
	} else {
		// we couldn't find the port, so send err
		spErr("We could not find the serial port " + portname + " that you were trying to close.")
	}
}

func spWrite(arg string) {
	// we will get a string of comXX asdf asdf asdf
	//log.Println("Inside spWrite arg: " + arg)
	arg = strings.TrimPrefix(arg, " ")
	//log.Println("arg after trim: " + arg)
	args := strings.SplitN(arg, " ", 3)
	if len(args) != 3 {
		errstr := "Could not parse send command: " + arg
		//log.Println(errstr)
		spErr(errstr)
		return
	}
	portname := strings.Trim(args[1], " ")
	//log.Println("The port to write to is:" + portname + "---")
	//log.Println("The data is:" + args[2] + "---")

	// see if we have this port open
	myport, isFound := sh.FindPortByName(portname)

	if !isFound {
		// we couldn't find the port, so send err
		spErr("We could not find the serial port " + portname + " that you were trying to write to.")
		return
	}

	// we found our port
	// create our write request
	var wr writeRequest
	wr.p = myport

	// see if args[0] is send or sendnobuf or sendraw
	switch args[0] {
	case "send", "sendnobuf", "sendraw":
		wr.buffer = args[0]
	default:
		spErr("Unsupported send command:" + args[0] + ". Please specify a valid one")
		return
	}

	// include newline or not in the write? that is the question.
	// for now lets skip the newline
	//wr.d = []byte(args[2] + "\n")
	wr.d = args[2] //[]byte(args[2])

	// send it to the write channel
	sh.write <- wr
}
