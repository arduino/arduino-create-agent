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

package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"strconv"
	"time"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"
	serial "go.bug.st/serial"
)

// SerialConfig is the serial port configuration
type SerialConfig struct {
	Name  string
	Baud  int
	RtsOn bool
	DtrOn bool
}

type serport struct {
	// The serial port connection.
	portConf *SerialConfig
	portIo   io.ReadWriteCloser

	// Keep track of whether we're being actively closed
	// just so we don't show scary error messages
	isClosing bool

	isClosingDueToError bool

	// counter incremented on queue, decremented on write
	itemsInBuffer int

	// buffered channel containing up to 25600 outbound messages.
	sendBuffered chan string

	// unbuffered channel of outbound messages that bypass internal serial port buffer
	sendNoBuf chan []byte

	// channel containing raw base64 encoded binary data (outbound messages)
	sendRaw chan string

	// Do we have an extra channel/thread to watch our buffer?
	BufferType string
	//bufferwatcher *BufferflowDummypause
	bufferwatcher Bufferflow
}

// SpPortMessage is the serial port message
type SpPortMessage struct {
	P string // the port, i.e. com22
	D string // the data, i.e. G0 X0 Y0
}

// SpPortMessageRaw is the raw serial port message
type SpPortMessageRaw struct {
	P string // the port, i.e. com22
	D []byte // the data, i.e. G0 X0 Y0
}

func (p *serport) reader(buftype string) {

	timeCheckOpen := time.Now()
	var bufferedCh bytes.Buffer

	serialBuffer := make([]byte, 1024)
	for {
		n, err := p.portIo.Read(serialBuffer)
		bufferPart := serialBuffer[:n]

		//if we detect that port is closing, break out of this for{} loop.
		if p.isClosing {
			strmsg := "Shutting down reader on " + p.portConf.Name
			log.Println(strmsg)
			h.broadcastSys <- []byte(strmsg)
			break
		}

		// read can return legitimate bytes as well as an error
		// so process the n bytes red, if n > 0
		if n > 0 && err == nil {

			log.Print("Read " + strconv.Itoa(n) + " bytes ch: " + string(bufferPart[:n]))

			data := ""
			switch buftype {
			case "timedraw", "timed":
				data = string(bufferPart[:n])
				// give the data to our bufferflow so it can do it's work
				// to read/translate the data to see if it wants to block
				// writes to the serialport. each bufferflow type will decide
				// this on its own based on its logic
				p.bufferwatcher.OnIncomingData(data)
			case "default": // the bufferbuftype is actually called default 🤷‍♂️
				// save the left out bytes for the next iteration due to UTF-8 encoding
				bufferPart = append(bufferedCh.Bytes(), bufferPart[:n]...)
				n += len(bufferedCh.Bytes())
				bufferedCh.Reset()
				for i, w := 0, 0; i < n; i += w {
					runeValue, width := utf8.DecodeRune(bufferPart[i:n]) // try to decode the first i bytes in the buffer (UTF8 runes do not have a fixed length)
					if runeValue == utf8.RuneError {
						bufferedCh.Write(bufferPart[i:n])
						break
					}
					if i == n {
						bufferedCh.Reset()
					}
					data += string(runeValue)
					w = width
				}
				p.bufferwatcher.OnIncomingData(data)
			default:
				log.Panicf("unknown buffer type %s", buftype)
			}
		}

		// double check that we got characters in the buffer
		// before deciding if an EOF is legitimately a reason
		// to close the port because we're seeing that on some
		// os's like Linux/Ubuntu you get an EOF when you open
		// the port. Perhaps the EOF was buffered from a previous
		// close and the OS doesn't clear out that buffer on a new
		// connect. This means we'll only catch EOF's when there are
		// other characters with it, but that seems to work ok
		if n <= 0 {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// hit end of file
				log.Println("Hit end of file on serial port")
				h.broadcastSys <- []byte("{\"Cmd\":\"OpenFail\",\"Desc\":\"Got EOF (End of File) on port which usually means another app other than Serial Port JSON Server is locking your port. " + err.Error() + "\",\"Port\":\"" + p.portConf.Name + "\",\"Baud\":" + strconv.Itoa(p.portConf.Baud) + "}")

			}

			if err != nil {
				log.Println(err)
				h.broadcastSys <- []byte("Error reading on " + p.portConf.Name + " " +
					err.Error() + " Closing port.")
				h.broadcastSys <- []byte("{\"Cmd\":\"OpenFail\",\"Desc\":\"Got error reading on port. " + err.Error() + "\",\"Port\":\"" + p.portConf.Name + "\",\"Baud\":" + strconv.Itoa(p.portConf.Baud) + "}")
				p.isClosingDueToError = true
				break
			}

			// Keep track of time difference between two consecutive read with n == 0 and err == nil
			// we get here if the port has been disconnected while open (cpu usage will jump to 100%)
			// let's close the port only if the events are extremely fast (<1ms)
			if err == nil {
				diff := time.Since(timeCheckOpen)
				if diff.Nanoseconds() < 1000000 {
					p.isClosingDueToError = true
					break
				}
				timeCheckOpen = time.Now()
			}
		}
	}
	if p.isClosingDueToError {
		spCloseReal(p)
	}
}

// this method runs as its own thread because it's instantiated
// as a "go" method. so if it blocks inside, it is ok
func (p *serport) writerBuffered() {

	// this method can panic if user closes serial port and something is
	// in BlockUntilReady() and then a send occurs on p.sendNoBuf

	defer func() {
		if e := recover(); e != nil {
			log.Println("Got panic: ", e)
		}
	}()

	// this for loop blocks on p.sendBuffered until that channel
	// sees something come in
	for data := range p.sendBuffered {

		// send to the non-buffered serial port writer
		//log.Println("About to send to p.sendNoBuf channel")
		p.sendNoBuf <- []byte(data)

	}
	msgstr := "writerBuffered just got closed. make sure you make a new one. port:" + p.portConf.Name
	log.Println(msgstr)
	h.broadcastSys <- []byte(msgstr)
}

// this method runs as its own thread because it's instantiated
// as a "go" method. so if it blocks inside, it is ok
func (p *serport) writerNoBuf() {
	// this for loop blocks on p.send until that channel
	// sees something come in
	for data := range p.sendNoBuf {

		// if we get here, we were able to write successfully
		// to the serial port because it blocks until it can write

		// decrement counter
		p.itemsInBuffer--
		log.Printf("itemsInBuffer:%v\n", p.itemsInBuffer)

		// FINALLY, OF ALL THE CODE IN THIS PROJECT
		// WE TRULY/FINALLY GET TO WRITE TO THE SERIAL PORT!
		n2, err := p.portIo.Write(data)

		log.Print("Just wrote ", n2, " bytes to serial: ", string(data))
		if err != nil {
			errstr := "Error writing to " + p.portConf.Name + " " + err.Error() + " Closing port."
			log.Print(errstr)
			h.broadcastSys <- []byte(errstr)
			break
		}
	}
	msgstr := "Shutting down writer on " + p.portConf.Name
	log.Println(msgstr)
	h.broadcastSys <- []byte(msgstr)
	p.portIo.Close()
	updateSerialPortList()
	spList()
}

// this method runs as its own thread because it's instantiated
// as a "go" method. so if it blocks inside, it is ok
func (p *serport) writerRaw() {
	// this method can panic if user closes serial port and something is
	// in BlockUntilReady() and then a send occurs on p.sendNoBuf

	defer func() {
		if e := recover(); e != nil {
			log.Println("Got panic: ", e)
		}
	}()

	// this for loop blocks on p.sendRaw until that channel
	// sees something come in
	for data := range p.sendRaw {

		// Decode stuff
		sDec, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			log.Println("Decoding error:", err)
		}
		log.Println(string(sDec))

		// send to the non-buffered serial port writer
		p.sendNoBuf <- sDec

	}
	msgstr := "writerRaw just got closed. make sure you make a new one. port:" + p.portConf.Name
	log.Println(msgstr)
	h.broadcastSys <- []byte(msgstr)
}

func spHandlerOpen(portname string, baud int, buftype string) {

	log.Print("Inside spHandler")

	var out bytes.Buffer

	out.WriteString("Opening serial port ")
	out.WriteString(portname)
	out.WriteString(" at ")
	out.WriteString(strconv.Itoa(baud))
	out.WriteString(" baud")
	log.Print(out.String())

	conf := &SerialConfig{Name: portname, Baud: baud, RtsOn: true}

	mode := &serial.Mode{
		BaudRate: baud,
	}

	sp, err := serial.Open(portname, mode)
	log.Print("Just tried to open port")
	if err != nil {
		//log.Fatal(err)
		log.Print("Error opening port " + err.Error())
		//h.broadcastSys <- []byte("Error opening port. " + err.Error())
		h.broadcastSys <- []byte("{\"Cmd\":\"OpenFail\",\"Desc\":\"Error opening port. " + err.Error() + "\",\"Port\":\"" + conf.Name + "\",\"Baud\":" + strconv.Itoa(conf.Baud) + "}")

		return
	}
	log.Print("Opened port successfully")
	//p := &serport{send: make(chan []byte, 256), portConf: conf, portIo: sp}
	// we can go up to 256,000 lines of gcode in the buffer
	p := &serport{
		sendBuffered: make(chan string, 256000),
		sendNoBuf:    make(chan []byte),
		sendRaw:      make(chan string),
		portConf:     conf,
		portIo:       sp,
		BufferType:   buftype}

	var bw Bufferflow

	switch buftype {
	case "timed":
		bw = NewBufferflowTimed(portname, h.broadcastSys)
	case "timedraw":
		bw = NewBufferflowTimedRaw(portname, h.broadcastSys)
	case "default":
		bw = NewBufferflowDefault(portname, h.broadcastSys)
	default:
		log.Panicf("unknown buffer type: %s", buftype)
	}

	bw.Init()
	p.bufferwatcher = bw

	sh.register <- p
	defer func() { sh.unregister <- p }()

	updateSerialPortList()
	spList()

	// this is internally buffered thread to not send to serial port if blocked
	go p.writerBuffered()
	// this is thread to send to serial port regardless of block
	go p.writerNoBuf()
	// this is thread to send to serial port but with base64 decoding
	go p.writerRaw()

	p.reader(buftype)

	updateSerialPortList()
	spList()
}

func spHandlerClose(p *serport) {
	p.isClosing = true
	h.broadcastSys <- []byte("Closing serial port " + p.portConf.Name)
	spCloseReal(p)
}

func spCloseReal(p *serport) {
	p.bufferwatcher.Close()
	p.portIo.Close()
	updateSerialPortList()
	spList()
}
