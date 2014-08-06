package main

import (
	"log"
	"strconv"
	"strings"
)

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Inbound messages from the system
	broadcastSys chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var h = hub{
	broadcast:    make(chan []byte),
	broadcastSys: make(chan []byte),
	register:     make(chan *connection),
	unregister:   make(chan *connection),
	connections:  make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
			// send supported commands
			c.send <- []byte("{\"Version\" : \"" + version + "\"} ")
			c.send <- []byte("{\"Commands\" : [\"list\", \"open [portName] [baud] [bufferAlgorithm (optional)]\", \"send [portName] [cmd]\", \"sendnobuf [portName] [cmd]\", \"close [portName]\", \"bufferalgorithms\", \"baudrates\"]} ")
		case c := <-h.unregister:
			delete(h.connections, c)
			close(c.send)
		case m := <-h.broadcast:
			//log.Print("Got a broadcast")
			//log.Print(m)
			//log.Print(len(m))
			if len(m) > 0 {
				//log.Print(string(m))
				//log.Print(h.broadcast)
				checkCmd(m)
				//log.Print("-----")

				for c := range h.connections {
					select {
					case c.send <- m:
						//log.Print("did broadcast to ")
						//log.Print(c.ws.RemoteAddr())
						//c.send <- []byte("hello world")
					default:
						delete(h.connections, c)
						close(c.send)
						go c.ws.Close()
					}
				}
			}
		case m := <-h.broadcastSys:
			log.Printf("Got a system broadcast: %v\n", string(m))
			//log.Print(string(m))
			//log.Print("-----")

			for c := range h.connections {
				select {
				case c.send <- m:
					//log.Print("did broadcast to ")
					//log.Print(c.ws.RemoteAddr())
					//c.send <- []byte("hello world")
				default:
					delete(h.connections, c)
					close(c.send)
					go c.ws.Close()
				}
			}
		}
	}
}

func checkCmd(m []byte) {
	//log.Print("Inside checkCmd")
	s := string(m[:])
	log.Print(s)

	sl := strings.ToLower(s)

	if strings.HasPrefix(sl, "open") {

		args := strings.Split(s, " ")
		if len(args) < 3 {
			go spErr("You did not specify a port and baud rate in your open cmd")
			return
		}
		if len(args[1]) < 1 {
			go spErr("You did not specify a serial port")
			return
		}

		baudStr := strings.Replace(args[2], "\n", "", -1)
		baud, err := strconv.Atoi(baudStr)
		if err != nil {
			go spErr("Problem converting baud rate " + args[2])
			return
		}
		// pass in buffer type now as string. if user does not
		// ask for a buffer type pass in empty string
		bufferAlgorithm := ""
		if len(args) > 3 {
			// cool. we got a buffer type request
			buftype := strings.Replace(args[3], "\n", "", -1)
			bufferAlgorithm = buftype
		}
		go spHandlerOpen(args[1], baud, bufferAlgorithm)

	} else if strings.HasPrefix(sl, "close") {

		args := strings.Split(s, " ")
		if len(args) > 1 {
			go spClose(args[1])
		} else {
			go spErr("You did not specify a port to close")
		}

	} else if strings.HasPrefix(sl, "send") {
		// will catch send and sendnobuf

		//args := strings.Split(s, "send ")
		go spWrite(s)

	} else if strings.HasPrefix(sl, "list") {
		go spList()
		//go getListViaWmiPnpEntity()
	} else if strings.HasPrefix(sl, "bufferalgorithm") {
		go spBufferAlgorithms()
	} else if strings.HasPrefix(sl, "baudrate") {
		go spBaudRates()
	} else {
		go spErr("Could not understand command.")
	}

	//log.Print("Done with checkCmd")
}
