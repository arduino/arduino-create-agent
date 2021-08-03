package main

import (
	"encoding/json"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

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
}

var h = hub{
	broadcast:    make(chan []byte, 1000),
	broadcastSys: make(chan []byte, 1000),
	register:     make(chan *connection),
	unregister:   make(chan *connection),
	connections:  make(map[*connection]bool),
}

const commands = "{\"Commands\": [" +
	"\"list\", " +
	"\"open < portName > < baud > [bufferAlgorithm: ({default}, timed, timedraw, timedbinary)]\", " +
	"\"send < portName > < cmd >\", " +
	"\"sendnobuf < portName > < cmd >\", " +
	"\"close < portName >\", " +
	"\"restart\", " +
	"\"exit\", " +
	"\"killupload\", " +
	"\"downloadtool < tool > < toolVersion: {latest} > < pack: {arduino} > < behaviour: {keep} >\", " +
	"\"log\", " +
	"\"memorystats\", " +
	"\"gc\", " +
	"\"hostname\", " +
	"\"version\"]} "

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
			c.send <- []byte("{\"Version\" : \"" + version + "\"} ")
			c.send <- []byte(commands)
			c.send <- []byte("{\"Hostname\" : \"" + *hostname + "\"} ")
			c.send <- []byte("{\"OS\" : \"" + runtime.GOOS + "\"} ")
		case c := <-h.unregister:
			h.unregisterConnection(c)
		case m := <-h.broadcast:
			if len(m) > 0 {
				checkCmd(m)
				h.sendToRegisteredConnections(m)
			}
		case m := <-h.broadcastSys:
			h.sendToRegisteredConnections(m)
		}
	}
}

func checkCmd(m []byte) {
	//log.Print("Inside checkCmd")
	s := string(m[:])

	sl := strings.ToLower(strings.Trim(s, "\n"))

	if *hibernate == true {
		//do nothing
		return
	}

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
		bufferAlgorithm := "default" // use the default buffer if none is specified
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

	} else if strings.HasPrefix(sl, "killupload") {
		// kill the running process (assumes singleton for now)
		go func() {
			upload.Kill()
			h.broadcastSys <- []byte("{\"uploadStatus\": \"Killed\"}")
			log.Println("{\"uploadStatus\": \"Killed\"}")
		}()

	} else if strings.HasPrefix(sl, "send") {
		// will catch send and sendnobuf
		go spWrite(s)
	} else if strings.HasPrefix(sl, "list") {
		go spList(false)
		go spList(true)
	} else if strings.HasPrefix(sl, "downloadtool") {
		// Always delete root certificates when we receive a downloadtool command
		// Useful if the install procedure was not followed strictly (eg. manually)
		DeleteCertificates()
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

			err := Tools.Download(pack, tool, toolVersion, behaviour)
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
		go logAction(sl)
	} else if strings.HasPrefix(sl, "restart") {
		log.Println("Received restart from the daemon. Why? Boh")
		Systray.Restart()
	} else if strings.HasPrefix(sl, "exit") {
		Systray.Quit()
	} else if strings.HasPrefix(sl, "memstats") {
		memoryStats()
	} else if strings.HasPrefix(sl, "gc") {
		garbageCollection()
	} else if strings.HasPrefix(sl, "hostname") {
		getHostname()
	} else if strings.HasPrefix(sl, "version") {
		getVersion()
	} else {
		go spErr("Could not understand command.")
	}
}

func logAction(sl string) {
	if strings.HasPrefix(sl, "log on") {
		*logDump = "on"
		multi_writer := io.MultiWriter(&logger_ws, os.Stderr)
		log.SetOutput(multi_writer)
	} else if strings.HasPrefix(sl, "log off") {
		*logDump = "off"
		log.SetOutput(os.Stderr)
	} else if strings.HasPrefix(sl, "log show") {
		// TODO: send all the saved log to websocket
		//h.broadcastSys <- []byte("{\"BufFlowDebug\" : \"" + *logDump + "\"}")
	}
}

func memoryStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	json, _ := json.Marshal(memStats)
	log.Printf("memStats:%v\n", string(json))
	h.broadcastSys <- json
}

func getHostname() {
	h.broadcastSys <- []byte("{\"Hostname\" : \"" + *hostname + "\"}")
}

func getVersion() {
	h.broadcastSys <- []byte("{\"Version\" : \"" + version + "\"}")
}

func garbageCollection() {
	log.Printf("Starting garbageCollection()\n")
	h.broadcastSys <- []byte("{\"gc\":\"starting\"}")
	memoryStats()
	debug.SetGCPercent(100)
	debug.FreeOSMemory()
	debug.SetGCPercent(-1)
	log.Printf("Done with garbageCollection()\n")
	h.broadcastSys <- []byte("{\"gc\":\"done\"}")
	memoryStats()
}
