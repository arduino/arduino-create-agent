package main

import (
	"bytes"
	"github.com/tarm/goserial"
	"io"
	"log"
	"strconv"
)

type serport struct {
	// The serial port connection.
	portConf *serial.Config
	portIo   io.ReadWriteCloser

	// Keep track of whether we're being actively closed
	// just so we don't show scary error messages
	isClosing bool

	// Buffered channel of outbound messages.
	send chan []byte
}

func (p *serport) reader() {
	//var buf bytes.Buffer
	for {
		ch := make([]byte, 1024)
		n, err := p.portIo.Read(ch)

		// read can return legitimate bytes as well as an error
		// so process the bytes if n > 0
		if n > 0 {
			log.Print("Read " + strconv.Itoa(n) + " bytes ch: " + string(ch))
			h.broadcastSys <- []byte("{p: '" + p.portConf.Name + "', d: '" + string(ch[:n]) + "'}\n")
		}

		if p.isClosing {
			strmsg := "Shutting down reader on " + p.portConf.Name
			log.Println(strmsg)
			h.broadcastSys <- []byte(strmsg)
			break
		}

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			// hit end of file
			log.Println("Hit end of file on serial port")
		}
		if err != nil {
			log.Println(err)
			h.broadcastSys <- []byte("Error reading on " + p.portConf.Name + " " +
				err.Error() + " Closing port.")
			break
		}

		// loop thru and look for a newline
		/*
			for i := 0; i < n; i++ {
				// see if we hit a newline
				if ch[i] == '\n' {
					// we are done with the line
					h.broadcastSys <- buf.Bytes()
					buf.Reset()
				} else {
					// append to buffer
					buf.WriteString(string(ch[:n]))
				}
			}*/
		/*
			buf.WriteString(string(ch[:n]))
			log.Print(string(ch[:n]))
			if string(ch[:n]) == "\n" {
				h.broadcastSys <- buf.Bytes()
				buf.Reset()
			}
		*/
	}
	p.portIo.Close()
}

func (p *serport) writer() {
	for data := range p.send {
		n2, err := p.portIo.Write(data)
		log.Print("Just wrote ")
		log.Print(n2)
		log.Print(" bytes to serial: ")
		log.Print(data)
		if err != nil {
			errstr := "Error writing to " + p.portConf.Name + " " + err.Error() + " Closing port."
			log.Fatal(errstr)
			h.broadcastSys <- []byte(errstr)
			break
		}
	}
	msgstr := "Shutting down writer on " + p.portConf.Name
	log.Println(msgstr)
	h.broadcastSys <- []byte(msgstr)
	p.portIo.Close()
}

func spHandlerOpen(portname string, baud int) {

	log.Print("Inside spHandler")

	var out bytes.Buffer

	out.WriteString("Opening serial port ")
	out.WriteString(portname)
	out.WriteString(" at ")
	out.WriteString(strconv.Itoa(baud))
	out.WriteString(" baud")
	log.Print(out.String())

	//h.broadcast <- []byte("Opened a serial port bitches")
	h.broadcastSys <- out.Bytes()

	conf := &serial.Config{Name: portname, Baud: baud}
	log.Print("Created config for port")
	log.Print(conf)

	sp, err := serial.OpenPort(conf)
	log.Print("Just tried to open port")
	if err != nil {
		//log.Fatal(err)
		log.Print("Error opening port " + err.Error())
		h.broadcastSys <- []byte("Error opening port. " + err.Error())
		return
	}
	log.Print("Opened port successfully")
	p := &serport{send: make(chan []byte, 256), portConf: conf, portIo: sp}
	sh.register <- p
	defer func() { sh.unregister <- p }()
	go p.writer()
	p.reader()
}

func spHandlerClose(p *serport) {
	p.isClosing = true
	// close the port
	p.portIo.Close()
	// unregister myself
	// we already have a deferred unregister in place from when
	// we opened. the only thing holding up that thread is the p.reader()
	// so if we close the reader we should get an exit
	h.broadcastSys <- []byte("Closing serial port " + p.portConf.Name)
}
