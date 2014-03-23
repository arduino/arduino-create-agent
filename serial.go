package main

import (
	"log"
	"strings"
)

type writeRequest struct {
	p *serport
	d []byte
}

type serialhub struct {
	// Opened serial ports.
	ports map[*serport]bool

	//open chan *io.ReadWriteCloser
	//write chan *serport, chan []byte
	write chan writeRequest
	//read chan []byte

	// Register requests from the connections.
	register chan *serport

	// Unregister requests from connections.
	unregister chan *serport
}

var sh = serialhub{
	//write:   	make(chan *serport, chan []byte),
	write:      make(chan writeRequest),
	register:   make(chan *serport),
	unregister: make(chan *serport),
	ports:      make(map[*serport]bool),
}

func (sh *serialhub) run() {

	log.Print("Inside run of serialhub")
	//s := ser.open()
	//ser.s := s
	//ser.write(s, []byte("hello serial data"))
	for {
		select {
		case p := <-sh.register:
			log.Print("Registering a port ")
			log.Print(p)
			sh.ports[p] = true
		case p := <-sh.unregister:
			delete(sh.ports, p)
			close(p.send)
		case wr := <-sh.write:
			log.Print("Got a write to a port")
			log.Print("Port is ")
			log.Print(wr.p)
			log.Print("Data is ")
			log.Print(wr.d)
			log.Print("Data as string:" + string(wr.d))
			log.Print("-----")
			select {
			case wr.p.send <- wr.d:
				log.Print("Did write to serport")
			default:
				delete(sh.ports, wr.p)
				close(wr.p.send)
				//wr.p.port.Close()
				//go wr.p.port.Close()
			}
		}
	}
}

func spList() {
	ls := "{serialports:[\n"
	list, _ := getList()
	for _, item := range list {
		ls += "{name:'" + item.Name + "', friendly:'" + item.FriendlyName + "'}\n"
	}
	ls += "]}\n"
	h.broadcastSys <- []byte(ls)
}

func spErr(err string) {
	log.Println("Sending err back: ", err)
	h.broadcastSys <- []byte(err)
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
	log.Println("Inside spWrite arg: " + arg)
	arg = strings.TrimPrefix(arg, " ")
	log.Println("arg after trim: " + arg)
	args := strings.SplitN(arg, " ", 3)
	if len(args) != 3 {
		errstr := "Could not parse send command: " + arg
		log.Println(errstr)
		spErr(errstr)
		return
	}
	portname := strings.Trim(args[1], " ")
	log.Println("The port to write to is:" + portname + "---")
	log.Println("The data is:" + args[2])

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
	wr.d = []byte(args[2] + "\n")

	// send it to the write channel
	sh.write <- wr

}

func findPortByName(portname string) (*serport, bool) {
	portnamel := strings.ToLower(portname)
	for port := range sh.ports {
		if strings.ToLower(port.portConf.Name) == portnamel {
			// we found our port
			//spHandlerClose(port)
			return port, true
		}
	}
	return nil, false
}
