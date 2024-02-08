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
	"strconv"
	"strings"
	"sync"
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
	Ports           []SpPortItem
	portsLock       sync.Mutex
	enumerationLock sync.Mutex
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

func (sp *SerialPortList) Update() {
	if !sp.enumerationLock.TryLock() {
		// already enumerating...
		return
	}
	defer sp.enumerationLock.Unlock()

	ports, err := enumerateSerialPorts()
	if err != nil {
		// TODO: report error?

		// Empty port list if they can not be detected
		ports = []*OsSerialPort{}
	}

	// we have a full clean list of ports now. iterate thru them
	// to append the open/close state, baud rates, etc to make
	// a super clean nice list to send back to browser
	list := []SpPortItem{}
	for _, item := range ports {
		port := SpPortItem{
			Name:            item.Name,
			SerialNumber:    item.SerialNumber,
			IsOpen:          false,
			IsPrimary:       false,
			Baud:            0,
			BufferAlgorithm: "",
			Ver:             version,
			VendorID:        item.VID,
			ProductID:       item.PID,
		}

		// figure out if port is open
		if myport, isFound := sh.FindPortByName(item.Name); isFound {
			// and update data with the open port parameters
			port.IsOpen = true
			port.Baud = myport.portConf.Baud
			port.BufferAlgorithm = myport.BufferType
		}
		list = append(list, port)
	}

	serialPorts.portsLock.Lock()
	serialPorts.Ports = list
	serialPorts.portsLock.Unlock()
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
