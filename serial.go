// Supports Windows, Linux, Mac, BeagleBone Black, and Raspberry Pi

package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/arduino-create-agent/upload"
)

type writeRequest struct {
	p      *serport
	d      string
	buffer bool
	id     string
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
}

type SpPortList struct {
	Ports   []SpPortItem
	Network bool
}

type SpPortItem struct {
	Name            string
	SerialNumber    string
	DeviceClass     string
	IsOpen          bool
	IsPrimary       bool
	Baud            int
	BufferAlgorithm string
	Ver             string
	NetworkPort     bool
	VendorID        string
	ProductID       string
}

// SerialPorts contains the ports attached to the machine
var SerialPorts SpPortList

// NetworkPorts contains the ports on the network
var NetworkPorts SpPortList

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
			//log.Print("Registering a port: ", p.portConf.Name)
			h.broadcastSys <- []byte("{\"Cmd\":\"Open\",\"Desc\":\"Got register/open on port.\",\"Port\":\"" + p.portConf.Name + "\",\"Baud\":" + strconv.Itoa(p.portConf.Baud) + ",\"BufferType\":\"" + p.BufferType + "\"}")
			sh.ports[p] = true
		case p := <-sh.unregister:
			//log.Print("Unregistering a port: ", p.portConf.Name)
			h.broadcastSys <- []byte("{\"Cmd\":\"Close\",\"Desc\":\"Got unregister/close on port.\",\"Port\":\"" + p.portConf.Name + "\",\"Baud\":" + strconv.Itoa(p.portConf.Baud) + "}")
			delete(sh.ports, p)
			close(p.sendBuffered)
			close(p.sendNoBuf)
		case wr := <-sh.write:
			// if user sent in the commands as one text mode line
			write(wr, "")
		}
	}
}

func write(wr writeRequest, id string) {
	if wr.buffer {
		//log.Println("Send was normal send, so sending to wr.p.sendBuffered")
		wr.p.sendBuffered <- wr.d
	} else {
		//log.Println("Send was sendnobuf, so sending to wr.p.sendNoBuf")
		wr.p.sendNoBuf <- wr.d
	}
}

// spList broadcasts a Json representation of the ports found
func spList(network bool) {
	var list SpPortList
	if network {
		list = NetworkPorts
	} else {
		list = SerialPorts
	}
	ls, err := json.MarshalIndent(list, "", "\t")
	if err != nil {
		//log.Println(err)
		h.broadcastSys <- []byte("Error creating json on port list " +
			err.Error())
	} else {
		h.broadcastSys <- ls
	}
}

// discoverLoop periodically update the list of ports found
func discoverLoop() {
	SerialPorts.Network = false
	SerialPorts.Ports = make([]SpPortItem, 0)
	NetworkPorts.Network = true
	NetworkPorts.Ports = make([]SpPortItem, 0)

	go func() {
		for {
			if !upload.Busy {
				spListDual(false)
			}
			time.Sleep(2 * time.Second)
		}
	}()
	go func() {
		for {
			spListDual(true)
			time.Sleep(2 * time.Second)
		}
	}()
}

func spListDual(network bool) {

	// call our os specific implementation of getting the serial list
	list, err := GetList(network)

	//log.Println(list)
	//log.Println(err)

	if err != nil {
		// avoid reporting dummy data if an error occurred
		return
	}

	// do a quick loop to see if any of our open ports
	// did not end up in the list port list. this can
	// happen on windows in a fallback scenario where an
	// open port can't be identified because it is locked,
	// so just solve that by manually inserting
	// if network {
	// 	for port := range sh.ports {

	// 		isFound := false
	// 		for _, item := range list {
	// 			if strings.ToLower(port.portConf.Name) == strings.ToLower(item.Name) {
	// 				isFound = true
	// 			}
	// 		}

	// 		if !isFound {
	// 			// artificially push to front of port list
	// 			log.Println(fmt.Sprintf("Did not find an open port in the serial port list. We are going to artificially push it onto the list. port:%v", port.portConf.Name))
	// 			var ossp OsSerialPort
	// 			ossp.Name = port.portConf.Name
	// 			ossp.FriendlyName = port.portConf.Name
	// 			list = append([]OsSerialPort{ossp}, list...)
	// 		}
	// 	}
	// }

	// we have a full clean list of ports now. iterate thru them
	// to append the open/close state, baud rates, etc to make
	// a super clean nice list to send back to browser
	n := len(list)
	spl := SpPortList{make([]SpPortItem, n, n), network}

	ctr := 0

	for _, item := range list {

		spl.Ports[ctr] = SpPortItem{
			Name:            item.Name,
			SerialNumber:    item.ISerial,
			DeviceClass:     item.DeviceClass,
			IsOpen:          false,
			IsPrimary:       false,
			Baud:            0,
			BufferAlgorithm: "",
			Ver:             version,
			NetworkPort:     item.NetworkPort,
			VendorID:        item.IdVendor,
			ProductID:       item.IdProduct,
		}

		// figure out if port is open
		myport, isFound := findPortByName(item.Name)

		if isFound {
			// we found our port
			spl.Ports[ctr].IsOpen = true
			spl.Ports[ctr].Baud = myport.portConf.Baud
			spl.Ports[ctr].BufferAlgorithm = myport.BufferType
		}
		ctr++
	}

	if network {
		NetworkPorts = spl
	} else {
		SerialPorts = spl
	}
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

	myport, isFound := findPortByName(portname)

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
	myport, isFound := findPortByName(portname)

	if !isFound {
		// we couldn't find the port, so send err
		spErr("We could not find the serial port " + portname + " that you were trying to write to.")
		return
	}

	// we found our port
	// create our write request
	var wr writeRequest
	wr.p = myport

	// see if args[0] is send or sendnobuf
	if args[0] != "sendnobuf" {
		// we were just given a "send" so buffer it
		wr.buffer = true
	} else {
		//log.Println("sendnobuf specified so wr.buffer is false")
		wr.buffer = false
	}

	// include newline or not in the write? that is the question.
	// for now lets skip the newline
	//wr.d = []byte(args[2] + "\n")
	wr.d = args[2] //[]byte(args[2])

	// send it to the write channel
	sh.write <- wr
}
