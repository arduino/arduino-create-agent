package main

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"time"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"
	serial "go.bug.st/serial.v1"
)

type SerialConfig struct {
	Name string
	Baud int

	// Size     int // 0 get translated to 8
	// Parity   SomeNewTypeToGetCorrectDefaultOf_None
	// StopBits SomeNewTypeToGetCorrectDefaultOf_1

	// RTSFlowControl bool
	// DTRFlowControl bool
	// XONFlowControl bool

	// CRLFTranslate bool
	// TimeoutStuff int
	RtsOn bool
	DtrOn bool
}

type serport struct {
	// The serial port connection.
	portConf *SerialConfig
	portIo   io.ReadWriteCloser

	done chan bool // signals the end of this request

	// Keep track of whether we're being actively closed
	// just so we don't show scary error messages
	isClosing bool

	isClosingDueToError bool

	// counter incremented on queue, decremented on write
	itemsInBuffer int

	// buffered channel containing up to 25600 outbound messages.
	sendBuffered chan string

	// unbuffered channel of outbound messages that bypass internal serial port buffer
	sendNoBuf chan string

	// Do we have an extra channel/thread to watch our buffer?
	BufferType string
	//bufferwatcher *BufferflowDummypause
	bufferwatcher Bufferflow
}

type Cmd struct {
	data                       string
	id                         string
	skippedBuffer              bool
	willHandleCompleteResponse bool
}

type CmdComplete struct {
	Cmd     string
	Id      string
	P       string
	BufSize int    `json:"-"`
	D       string `json:"-"`
}

type qwReport struct {
	Cmd  string
	QCnt int
	Id   string
	D    string `json:"-"`
	Buf  string `json:"-"`
	P    string
}

type SpPortMessage struct {
	// P string // the port, i.e. com22
	D string // the data, i.e. G0 X0 Y0
}

type SpPortMessageRaw struct {
	// P string // the port, i.e. com22
	D []byte // the data, i.e. G0 X0 Y0
}

func (p *serport) reader() {

	//var buf bytes.Buffer
	ch := make([]byte, 1024)
	timeCheckOpen := time.Now()
	var buffered_ch bytes.Buffer

	for {

		n, err := p.portIo.Read(ch)

		//if we detect that port is closing, break out o this for{} loop.
		if p.isClosing {
			strmsg := "Shutting down reader on " + p.portConf.Name
			log.Println(strmsg)
			h.broadcastSys <- []byte(strmsg)
			break
		}

		if err == nil {
			ch = append(buffered_ch.Bytes(), ch[:n]...)
			n += len(buffered_ch.Bytes())
			buffered_ch.Reset()
		}

		// read can return legitimate bytes as well as an error
		// so process the bytes if n > 0
		if n > 0 {
			//log.Print("Read " + strconv.Itoa(n) + " bytes ch: " + string(ch))

			data := ""

			for i, w := 0, 0; i < n; i += w {
				runeValue, width := utf8.DecodeRune(ch[i:n])
				if runeValue == utf8.RuneError {
					buffered_ch.Write(append(ch[i:n]))
					break
				}
				if i == n {
					buffered_ch.Reset()
				}
				data += string(runeValue)
				w = width
			}

			//log.Print("The data i will convert to json is:")
			//log.Print(data)

			// give the data to our bufferflow so it can do it's work
			// to read/translate the data to see if it wants to block
			// writes to the serialport. each bufferflow type will decide
			// this on its own based on its logic, i.e. tinyg vs grbl vs others
			//p.b.bufferwatcher..OnIncomingData(data)
			p.bufferwatcher.OnIncomingData(data)

			// see if the OnIncomingData handled the broadcast back
			// to the user. this option was added in case the OnIncomingData wanted
			// to do something fancier or implementation specific, i.e. TinyG Buffer
			// actually sends back data on a perline basis rather than our method
			// where we just send the moment we get it. the reason for this is that
			// the browser was sometimes getting back packets out of order which
			// of course would screw things up when parsing

			if p.bufferwatcher.IsBufferGloballySendingBackIncomingData() == false {
				//m := SpPortMessage{"Alice", "Hello"}
				m := SpPortMessage{data}
				//log.Print("The m obj struct is:")
				//log.Print(m)

				//b, err := json.MarshalIndent(m, "", "\t")
				b, err := json.Marshal(m)
				if err != nil {
					log.Println(err)
					h.broadcastSys <- []byte("Error creating json on " + p.portConf.Name + " " +
						err.Error() + " The data we were trying to convert is: " + string(ch[:n]))
					break
				}
				//log.Print("Printing out json byte data...")
				//log.Print(string(b))
				h.broadcastSys <- b
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

		// we want to block here if we are being asked to pause.
		goodToGo, _ := p.bufferwatcher.BlockUntilReady(string(data), "")

		if goodToGo == false {
			log.Println("We got back from BlockUntilReady() but apparently we must cancel this cmd")
			// since we won't get a buffer decrement in p.sendNoBuf, we must do it here
			p.itemsInBuffer--
		} else {
			// send to the non-buffered serial port writer
			//log.Println("About to send to p.sendNoBuf channel")
			p.sendNoBuf <- data
		}
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
		n2, err := p.portIo.Write([]byte(data))

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
	spListDual(false)
	spList(false)
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
	p := &serport{sendBuffered: make(chan string, 256000), sendNoBuf: make(chan string), portConf: conf, portIo: sp, BufferType: buftype}

	var bw Bufferflow

	if buftype == "timed" {
		bw = &BufferflowTimed{Name: "timed", Port: portname, Output: h.broadcastSys, Input: make(chan string)}
	} else if buftype == "timedraw" {
		bw = &BufferflowTimedRaw{Name: "timedraw", Port: portname, Output: h.broadcastSys, Input: make(chan string)}
	} else {
		bw = &BufferflowDefault{Port: portname}
	}

	bw.Init()
	p.bufferwatcher = bw

	sh.register <- p
	defer func() { sh.unregister <- p }()

	spListDual(false)
	spList(false)

	// this is internally buffered thread to not send to serial port if blocked
	go p.writerBuffered()
	// this is thread to send to serial port regardless of block
	go p.writerNoBuf()
	p.reader()

	spListDual(false)
	spList(false)

	//go p.reader()
	//p.done = make(chan bool)
	//<-p.done
}

func spHandlerClose(p *serport) {
	p.isClosing = true
	h.broadcastSys <- []byte("Closing serial port " + p.portConf.Name)
	spCloseReal(p)
}

func spCloseReal(p *serport) {
	p.bufferwatcher.Close()
	p.portIo.Close()
	spListDual(false)
	spList(false)
}
