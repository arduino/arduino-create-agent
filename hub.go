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
		case c := <-h.unregister:
			delete(h.connections, c)
			close(c.send)
		case m := <-h.broadcast:
			log.Print("Got a broadcast")
			log.Print(m)
			//log.Print(h.broadcast)
			checkCmd(m)
			log.Print("-----")

			for c := range h.connections {
				select {
				case c.send <- m:
					log.Print("did broadcast to ")
					log.Print(c.ws.RemoteAddr())
					//c.send <- []byte("hello world")
				default:
					delete(h.connections, c)
					close(c.send)
					go c.ws.Close()
				}
			}
		case m := <-h.broadcastSys:
			log.Print("Got a system broadcast")
			log.Print(m)
			log.Print("-----")

			for c := range h.connections {
				select {
				case c.send <- m:
					log.Print("did broadcast to ")
					log.Print(c.ws.RemoteAddr())
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
	log.Print("Inside checkCmd")
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
		baud, err := strconv.Atoi(args[2])
		if err != nil {
			go spErr("Problem converting baud rate " + args[2])
			return
		}
		go spHandlerOpen(args[1], baud)

	} else if strings.HasPrefix(sl, "close") {

		args := strings.Split(s, " ")
		go spClose(args[1])

	} else if strings.HasPrefix(sl, "send ") {

		//args := strings.Split(s, "send ")
		go spWrite(s)

	} else if s == "list" {
		go spList()
		//go getListViaWmiPnpEntity()
	}

	log.Print("Done with checkCmd")
}
