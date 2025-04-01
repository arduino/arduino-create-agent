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

func newHub(serialhub *serialhub, serialList *serialPortList, tools *tools.Tools, systray *systray.Systray) *hub {
	hub := &hub{
		broadcast:      make(chan []byte, 1000),
		broadcastSys:   make(chan []byte, 1000),
		register:       make(chan *connection),
		unregister:     make(chan *connection),
		connections:    make(map[*connection]bool),
		serialHub:      serialhub,
		serialPortList: serialList,
		tools:          tools,
		systray:        systray,
	}

	hub.serialHub.OnRegister = func(port *serport) {
		hub.broadcastSys <- []byte("{\"Cmd\":\"Open\",\"Desc\":\"Got register/open on port.\",\"Port\":\"" + port.portConf.Name + "\",\"Baud\":" + strconv.Itoa(port.portConf.Baud) + ",\"BufferType\":\"" + port.BufferType + "\"}")
	}

	hub.serialHub.OnUnregister = func(port *serport) {
		hub.broadcastSys <- []byte("{\"Cmd\":\"Close\",\"Desc\":\"Got unregister/close on port.\",\"Port\":\"" + port.portConf.Name + "\",\"Baud\":" + strconv.Itoa(port.portConf.Baud) + "}")
	}

	hub.serialPortList.OnList = func(data []byte) {
		hub.broadcastSys <- data
	}

	hub.serialPortList.OnErr = func(err string) {
		hub.broadcastSys <- []byte("{\"Error\":\"" + err + "\"}")
	}

	return hub
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

func (h *hub) unregisterConnection(c *connection) {
	if _, contains := h.connections[c]; !contains {
		return
	}
	delete(h.connections, c)
	close(c.send)
}

func (h *hub) sendToRegisteredConnections(data []byte) {
	for c := range h.connections {
		select {
		case c.send <- data:
			//log.Print("did broadcast to ")
			//log.Print(c.ws.RemoteAddr())
			//c.send <- []byte("hello world")
		default:
			h.unregisterConnection(c)
		}
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
			// send supported commands
			c.send <- []byte(fmt.Sprintf(`{"Version" : "%s"} `, version))
			c.send <- []byte(html.EscapeString(commands))
			c.send <- []byte(fmt.Sprintf(`{"Hostname" : "%s"} `, *hostname))
			c.send <- []byte(fmt.Sprintf(`{"OS" : "%s"} `, runtime.GOOS))
		case c := <-h.unregister:
			h.unregisterConnection(c)
		case m := <-h.broadcast:
			if len(m) > 0 {
				h.checkCmd(m)
				h.sendToRegisteredConnections(m)
			}
		case m := <-h.broadcastSys:
			h.sendToRegisteredConnections(m)
		}
	}
}

func (h *hub) checkCmd(m []byte) {
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
			go h.spErr("You did not specify a port and baud rate in your open cmd")
			return
		}
		if len(args[1]) < 1 {
			go h.spErr("You did not specify a serial port")
			return
		}

		baudStr := strings.Replace(args[2], "\n", "", -1)
		baud, err := strconv.Atoi(baudStr)
		if err != nil {
			go h.spErr("Problem converting baud rate " + args[2])
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
		go h.spHandlerOpen(args[1], baud, bufferAlgorithm)

	} else if strings.HasPrefix(sl, "close") {

		args := strings.Split(s, " ")
		if len(args) > 1 {
			go h.spClose(args[1])
		} else {
			go h.spErr("You did not specify a port to close")
		}

	} else if strings.HasPrefix(sl, "killupload") {
		// kill the running process (assumes singleton for now)
		go func() {
			upload.Kill()
			h.broadcastSys <- []byte("{\"uploadStatus\": \"Killed\"}")
			log.Println("{\"uploadStatus\": \"Killed\"}")
		}()

	} else if strings.HasPrefix(sl, "send") {
		// will catch send and sendnobuf and sendraw
		go h.spWrite(s)
	} else if strings.HasPrefix(sl, "list") {
		go h.serialPortList.List()
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
				h.broadcastSys <- mapB
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
				h.broadcastSys <- mapB
			}
			err := h.tools.Download(pack, tool, toolVersion, behaviour, reportPendingProgress)
			if err != nil {
				mapD := map[string]string{"DownloadStatus": "Error", "Msg": err.Error()}
				mapB, _ := json.Marshal(mapD)
				h.broadcastSys <- mapB
			} else {
				mapD := map[string]string{"DownloadStatus": "Success", "Msg": "Map Updated"}
				mapB, _ := json.Marshal(mapD)
				h.broadcastSys <- mapB
			}
		}()
	} else if strings.HasPrefix(sl, "log") {
		go h.logAction(sl)
	} else if strings.HasPrefix(sl, "restart") {
		// potentially, the sysStray dependencies can be removed  https://github.com/arduino/arduino-create-agent/issues/1013
		log.Println("Received restart from the daemon. Why? Boh")
		h.systray.Restart()
	} else if strings.HasPrefix(sl, "exit") {
		h.systray.Quit()
	} else if strings.HasPrefix(sl, "memstats") {
		h.memoryStats()
	} else if strings.HasPrefix(sl, "gc") {
		h.garbageCollection()
	} else if strings.HasPrefix(sl, "hostname") {
		h.getHostname()
	} else if strings.HasPrefix(sl, "version") {
		h.getVersion()
	} else {
		go h.spErr("Could not understand command.")
	}
}

type logWriter struct {
	onWrite func([]byte)
}

func (u *logWriter) Write(p []byte) (n int, err error) {
	u.onWrite(p)
	return len(p), nil
}

func (h *hub) logAction(sl string) {
	if strings.HasPrefix(sl, "log on") {
		*logDump = "on"

		logWriter := logWriter{}
		logWriter.onWrite = func(p []byte) {
			h.broadcastSys <- p
		}

		multiWriter := io.MultiWriter(&logWriter, os.Stderr)
		log.SetOutput(multiWriter)
	} else if strings.HasPrefix(sl, "log off") {
		*logDump = "off"
		log.SetOutput(os.Stderr)
		// } else if strings.HasPrefix(sl, "log show") {
		// TODO: send all the saved log to websocket
		//h.broadcastSys <- []byte("{\"BufFlowDebug\" : \"" + *logDump + "\"}")
	}
}

func (h *hub) memoryStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	json, _ := json.Marshal(memStats)
	log.Printf("memStats:%v\n", string(json))
	h.broadcastSys <- json
}

func (h *hub) getHostname() {
	h.broadcastSys <- []byte("{\"Hostname\" : \"" + *hostname + "\"}")
}

func (h *hub) getVersion() {
	h.broadcastSys <- []byte("{\"Version\" : \"" + version + "\"}")
}

func (h *hub) garbageCollection() {
	log.Printf("Starting garbageCollection()\n")
	h.broadcastSys <- []byte("{\"gc\":\"starting\"}")
	h.memoryStats()
	debug.SetGCPercent(100)
	debug.FreeOSMemory()
	debug.SetGCPercent(-1)
	log.Printf("Done with garbageCollection()\n")
	h.broadcastSys <- []byte("{\"gc\":\"done\"}")
	h.memoryStats()
}
