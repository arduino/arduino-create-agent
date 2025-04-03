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
	"encoding/json"
	"fmt"
	"html"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/arduino/arduino-create-agent/systray"
	"github.com/arduino/arduino-create-agent/tools"
	"github.com/arduino/arduino-create-agent/upload"
	log "github.com/sirupsen/logrus"
	"go.bug.st/serial"
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

	// Serial hub to communicate with serial ports
	serialHub *serialhub

	serialPortList *serialPortList

	tools *tools.Tools

	systray *systray.Systray
}

func newHub(tools *tools.Tools, systray *systray.Systray) *hub {
	broadcastSys := make(chan []byte, 1000)

	onRegister := func(port *serport) {
		broadcastSys <- []byte("{\"Cmd\":\"Open\",\"Desc\":\"Got register/open on port.\",\"Port\":\"" + port.portConf.Name + "\",\"Baud\":" + strconv.Itoa(port.portConf.Baud) + ",\"BufferType\":\"" + port.BufferType + "\"}")
	}
	onUnregister := func(port *serport) {
		broadcastSys <- []byte("{\"Cmd\":\"Close\",\"Desc\":\"Got unregister/close on port.\",\"Port\":\"" + port.portConf.Name + "\",\"Baud\":" + strconv.Itoa(port.portConf.Baud) + "}")
	}
	serialHub := newSerialHub(onRegister, onUnregister)

	onList := func(data []byte) {
		broadcastSys <- data
	}
	onErr := func(err string) {
		broadcastSys <- []byte("{\"Error\":\"" + err + "\"}")
	}
	serialPortList := newSerialPortList(tools, onList, onErr)

	return &hub{
		broadcast:      make(chan []byte, 1000),
		broadcastSys:   broadcastSys,
		register:       make(chan *connection),
		unregister:     make(chan *connection),
		connections:    make(map[*connection]bool),
		serialHub:      serialHub,
		serialPortList: serialPortList,
		tools:          tools,
		systray:        systray,
	}
}

const commands = `{
  "Commands": [
    "list",
    "open <portName> <baud> [bufferAlgorithm: ({default}, timed, timedraw)]",
    "(send, sendnobuf, sendraw) <portName> <cmd>",
    "close <portName>",
    "restart",
    "exit",
    "killupload",
    "downloadtool <tool> <toolVersion: {latest}> <pack: {arduino}> <behaviour: {keep}>",
    "log",
    "memorystats",
    "gc",
    "hostname",
    "version"
  ]
}`

func (hub *hub) unregisterConnection(c *connection) {
	if _, contains := hub.connections[c]; !contains {
		return
	}
	delete(hub.connections, c)
	close(c.send)
}

func (hub *hub) sendToRegisteredConnections(data []byte) {
	for c := range hub.connections {
		select {
		case c.send <- data:
			//log.Print("did broadcast to ")
			//log.Print(c.ws.RemoteAddr())
			//c.send <- []byte("hello world")
		default:
			hub.unregisterConnection(c)
		}
	}
}

func (hub *hub) run() {
	go hub.serialPortList.Run()

	for {
		select {
		case c := <-hub.register:
			hub.connections[c] = true
			// send supported commands
			c.send <- []byte(fmt.Sprintf(`{"Version" : "%s"} `, version))
			c.send <- []byte(html.EscapeString(commands))
			c.send <- []byte(fmt.Sprintf(`{"Hostname" : "%s"} `, *hostname))
			c.send <- []byte(fmt.Sprintf(`{"OS" : "%s"} `, runtime.GOOS))
		case c := <-hub.unregister:
			hub.unregisterConnection(c)
		case m := <-hub.broadcast:
			if len(m) > 0 {
				hub.checkCmd(m)
				hub.sendToRegisteredConnections(m)
			}
		case m := <-hub.broadcastSys:
			hub.sendToRegisteredConnections(m)
		}
	}
}

func (hub *hub) checkCmd(m []byte) {
	//log.Print("Inside checkCmd")
	s := string(m[:])

	sl := strings.ToLower(strings.Trim(s, "\n"))

	if *hibernate {
		//do nothing
		return
	}

	if strings.HasPrefix(sl, "open") {

		args := strings.Split(s, " ")
		if len(args) < 3 {
			go hub.spErr("You did not specify a port and baud rate in your open cmd")
			return
		}
		if len(args[1]) < 1 {
			go hub.spErr("You did not specify a serial port")
			return
		}

		baudStr := strings.Replace(args[2], "\n", "", -1)
		baud, err := strconv.Atoi(baudStr)
		if err != nil {
			go hub.spErr("Problem converting baud rate " + args[2])
			return
		}
		// pass in buffer type now as string. if user does not
		// ask for a buffer type pass in empty string
		bufferAlgorithm := "default" // use the default buffer if none is specified
		if len(args) > 3 {
			// cool. we got a buffer type request
			buftype := strings.Replace(args[3], "\n", "", -1)
			bufferAlgorithm = buftype
		}
		go hub.spHandlerOpen(args[1], baud, bufferAlgorithm)

	} else if strings.HasPrefix(sl, "close") {

		args := strings.Split(s, " ")
		if len(args) > 1 {
			go hub.spClose(args[1])
		} else {
			go hub.spErr("You did not specify a port to close")
		}

	} else if strings.HasPrefix(sl, "killupload") {
		// kill the running process (assumes singleton for now)
		go func() {
			upload.Kill()
			hub.broadcastSys <- []byte("{\"uploadStatus\": \"Killed\"}")
			log.Println("{\"uploadStatus\": \"Killed\"}")
		}()

	} else if strings.HasPrefix(sl, "send") {
		// will catch send and sendnobuf and sendraw
		go hub.spWrite(s)
	} else if strings.HasPrefix(sl, "list") {
		go hub.serialPortList.List()
	} else if strings.HasPrefix(sl, "downloadtool") {
		go func() {
			args := strings.Split(s, " ")
			var tool, toolVersion, pack, behaviour string
			toolVersion = "latest"
			pack = "arduino"
			behaviour = "keep"
			if len(args) <= 1 {
				mapD := map[string]string{"DownloadStatus": "Error", "Msg": "Not enough arguments"}
				mapB, _ := json.Marshal(mapD)
				hub.broadcastSys <- mapB
				return
			}
			if len(args) > 1 {
				tool = args[1]
			}
			if len(args) > 2 {
				if strings.HasPrefix(args[2], "http") {
					//old APIs, ignore this field
				} else {
					toolVersion = args[2]
				}
			}
			if len(args) > 3 {
				pack = args[3]
			}
			if len(args) > 4 {
				behaviour = args[4]
			}

			reportPendingProgress := func(msg string) {
				mapD := map[string]string{"DownloadStatus": "Pending", "Msg": msg}
				mapB, _ := json.Marshal(mapD)
				hub.broadcastSys <- mapB
			}
			err := hub.tools.Download(pack, tool, toolVersion, behaviour, reportPendingProgress)
			if err != nil {
				mapD := map[string]string{"DownloadStatus": "Error", "Msg": err.Error()}
				mapB, _ := json.Marshal(mapD)
				hub.broadcastSys <- mapB
			} else {
				mapD := map[string]string{"DownloadStatus": "Success", "Msg": "Map Updated"}
				mapB, _ := json.Marshal(mapD)
				hub.broadcastSys <- mapB
			}
		}()
	} else if strings.HasPrefix(sl, "log") {
		go hub.logAction(sl)
	} else if strings.HasPrefix(sl, "restart") {
		// potentially, the sysStray dependencies can be removed  https://github.com/arduino/arduino-create-agent/issues/1013
		log.Println("Received restart from the daemon. Why? Boh")
		hub.systray.Restart()
	} else if strings.HasPrefix(sl, "exit") {
		hub.systray.Quit()
	} else if strings.HasPrefix(sl, "memstats") {
		hub.memoryStats()
	} else if strings.HasPrefix(sl, "gc") {
		hub.garbageCollection()
	} else if strings.HasPrefix(sl, "hostname") {
		hub.getHostname()
	} else if strings.HasPrefix(sl, "version") {
		hub.getVersion()
	} else {
		go hub.spErr("Could not understand command.")
	}
}

type logWriter struct {
	onWrite func([]byte)
}

func (hub *hub) logAction(sl string) {
	if strings.HasPrefix(sl, "log on") {
		*logDump = "on"

		logWriter := logWriter{}
		logWriter.onWrite = func(p []byte) {
			hub.broadcastSys <- p
		}

		multiWriter := io.MultiWriter(&logWriter, os.Stderr)
		log.SetOutput(multiWriter)
	} else if strings.HasPrefix(sl, "log off") {
		*logDump = "off"
		log.SetOutput(os.Stderr)
		// } else if strings.HasPrefix(sl, "log show") {
		// TODO: send all the saved log to websocket
		//hub.broadcastSys <- []byte("{\"BufFlowDebug\" : \"" + *logDump + "\"}")
	}
}

func (u *logWriter) Write(p []byte) (n int, err error) {
	u.onWrite(p)
	return len(p), nil
}

func (hub *hub) memoryStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	json, _ := json.Marshal(memStats)
	log.Printf("memStats:%v\n", string(json))
	hub.broadcastSys <- json
}

func (hub *hub) getHostname() {
	hub.broadcastSys <- []byte("{\"Hostname\" : \"" + *hostname + "\"}")
}

func (hub *hub) getVersion() {
	hub.broadcastSys <- []byte("{\"Version\" : \"" + version + "\"}")
}

func (hub *hub) garbageCollection() {
	log.Printf("Starting garbageCollection()\n")
	hub.broadcastSys <- []byte("{\"gc\":\"starting\"}")
	hub.memoryStats()
	debug.SetGCPercent(100)
	debug.FreeOSMemory()
	debug.SetGCPercent(-1)
	log.Printf("Done with garbageCollection()\n")
	hub.broadcastSys <- []byte("{\"gc\":\"done\"}")
	hub.memoryStats()
}

func (hub *hub) spHandlerOpen(portname string, baud int, buftype string) {

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
		//hub.broadcastSys <- []byte("Error opening port. " + err.Error())
		hub.broadcastSys <- []byte("{\"Cmd\":\"OpenFail\",\"Desc\":\"Error opening port. " + err.Error() + "\",\"Port\":\"" + conf.Name + "\",\"Baud\":" + strconv.Itoa(conf.Baud) + "}")

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
		portName:     portname,
		BufferType:   buftype,
	}

	p.OnMessage = func(msg []byte) {
		hub.broadcastSys <- msg
	}
	p.OnClose = func(port *serport) {
		hub.serialPortList.MarkPortAsClosed(p.portName)
		hub.serialPortList.List()
	}

	var bw Bufferflow

	switch buftype {
	case "timed":
		bw = NewBufferflowTimed(portname, hub.broadcastSys)
	case "timedraw":
		bw = NewBufferflowTimedRaw(portname, hub.broadcastSys)
	case "default":
		bw = NewBufferflowDefault(portname, hub.broadcastSys)
	default:
		log.Panicf("unknown buffer type: %s", buftype)
	}

	bw.Init()
	p.bufferwatcher = bw

	hub.serialHub.Register(p)
	defer hub.serialHub.Unregister(p)

	hub.serialPortList.MarkPortAsOpened(portname)
	hub.serialPortList.List()

	// this is internally buffered thread to not send to serial port if blocked
	go p.writerBuffered()
	// this is thread to send to serial port regardless of block
	go p.writerNoBuf()
	// this is thread to send to serial port but with base64 decoding
	go p.writerRaw()

	p.reader(buftype)

	hub.serialPortList.List()
}

func (hub *hub) spClose(portname string) {
	if myport, ok := hub.serialHub.FindPortByName(portname); ok {
		hub.broadcastSys <- []byte("Closing serial port " + portname)
		myport.Close()
	} else {
		hub.spErr("We could not find the serial port " + portname + " that you were trying to close.")
	}
}

func (hub *hub) spWrite(arg string) {
	// we will get a string of comXX asdf asdf asdf
	//log.Println("Inside spWrite arg: " + arg)
	arg = strings.TrimPrefix(arg, " ")
	//log.Println("arg after trim: " + arg)
	args := strings.SplitN(arg, " ", 3)
	if len(args) != 3 {
		errstr := "Could not parse send command: " + arg
		//log.Println(errstr)
		hub.spErr(errstr)
		return
	}
	bufferingMode := args[0]
	portname := strings.Trim(args[1], " ")
	data := args[2]

	//log.Println("The port to write to is:" + portname + "---")
	//log.Println("The data is:" + data + "---")

	// See if we have this port open
	port, ok := hub.serialHub.FindPortByName(portname)
	if !ok {
		// we couldn't find the port, so send err
		hub.spErr("We could not find the serial port " + portname + " that you were trying to write to.")
		return
	}

	// see if bufferingMode is valid
	switch bufferingMode {
	case "send", "sendnobuf", "sendraw":
		// valid buffering mode, go ahead
	default:
		hub.spErr("Unsupported send command:" + args[0] + ". Please specify a valid one")
		return
	}

	// send it to the write channel
	port.Write(data, bufferingMode)
}

func (hub *hub) spErr(err string) {
	//log.Println("Sending err back: ", err)
	hub.broadcastSys <- []byte("{\"Error\" : \"" + err + "\"}")
}
