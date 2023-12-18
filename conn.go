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

// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-create-agent/upload"
	"github.com/arduino/arduino-create-agent/utilities"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	log "github.com/sirupsen/logrus"
)

type connection struct {
	// The websocket connection.
	ws socketio.Socket

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.Emit("message", string(message))
		if err != nil {
			break
		}
	}
}

// WsServer overrides socket.io server to set the CORS
type WsServer struct {
	Server *socketio.Server
}

func (s *WsServer) ServeHTTP(c *gin.Context) {
	s.Server.ServeHTTP(c.Writer, c.Request)
}

type additionalFile struct {
	Hex      []byte `json:"hex"`
	Filename string `json:"filename"`
}

// Upload contains the data to upload a sketch onto a board
type Upload struct {
	Port        string           `json:"port"`
	Board       string           `json:"board"`
	Rewrite     string           `json:"rewrite"`
	Commandline string           `json:"commandline"`
	Signature   string           `json:"signature"`
	Extra       upload.Extra     `json:"extra"`
	Hex         []byte           `json:"hex"`
	Filename    string           `json:"filename"`
	ExtraFiles  []additionalFile `json:"extrafiles"`
}

var uploadStatusStr = "ProgrammerStatus"

func uploadHandler(c *gin.Context) {

	data := new(Upload)
	c.BindJSON(data)

	log.Printf("%+v %+v %+v %+v %+v %+v", data.Port, data.Board, data.Rewrite, data.Commandline, data.Extra, data.Filename)

	if data.Port == "" {
		c.String(http.StatusBadRequest, "port is required")
		return
	}

	if data.Board == "" {
		c.String(http.StatusBadRequest, "board is required")
		log.Error("board is required")
		return
	}

	if !data.Extra.Network {
		if data.Signature == "" {
			c.String(http.StatusBadRequest, "signature is required")
			return
		}

		if data.Commandline == "" {
			c.String(http.StatusBadRequest, "commandline is required for local board")
			return
		}

		err := utilities.VerifyInput(data.Commandline, data.Signature)

		if err != nil {
			c.String(http.StatusBadRequest, "signature is invalid")
			return
		}
	}

	buffer := bytes.NewBuffer(data.Hex)

	filePath, err := utilities.SaveFileonTempDir(data.Filename, buffer)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	var filePaths []string
	filePaths = append(filePaths, filePath)

	tmpdir, err := os.MkdirTemp("", "extrafiles")
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	for _, extraFile := range data.ExtraFiles {
		path, err := utilities.SafeJoin(tmpdir, extraFile.Filename)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		filePaths = append(filePaths, path)
		log.Printf("Saving %s on %s", extraFile.Filename, path)

		err = os.MkdirAll(filepath.Dir(path), 0744)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		err = os.WriteFile(path, extraFile.Hex, 0644)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
	}

	if data.Rewrite != "" {
		data.Board = data.Rewrite
	}

	go func() {
		// Resolve commandline
		commandline, err := upload.PartiallyResolve(data.Board, filePath, tmpdir, data.Commandline, data.Extra, &Tools)
		if err != nil {
			send(map[string]string{uploadStatusStr: "Error", "Msg": err.Error()})
			return
		}

		l := PLogger{Verbose: true}

		// Upload
		if data.Extra.Network {
			err = errors.New("network upload is not supported anymore, pease use OTA instead")
		} else {
			send(map[string]string{uploadStatusStr: "Starting", "Cmd": "Serial"})
			err = upload.Serial(data.Port, commandline, data.Extra, l)
		}

		// Handle result
		if err != nil {
			send(map[string]string{uploadStatusStr: "Error", "Msg": err.Error()})
			return
		}
		send(map[string]string{uploadStatusStr: "Done", "Flash": "Ok"})
	}()

	c.String(http.StatusAccepted, "")
}

// PLogger sends the info from the upload to the websocket
type PLogger struct {
	Verbose bool
}

// Debug only sends messages if verbose is true (always true for now)
func (l PLogger) Debug(args ...interface{}) {
	if l.Verbose {
		l.Info(args...)
	}
}

// Info always send messages
func (l PLogger) Info(args ...interface{}) {
	output := fmt.Sprint(args...)
	log.Println(output)
	send(map[string]string{uploadStatusStr: "Busy", "Msg": output})
}

func send(args map[string]string) {
	mapB, _ := json.Marshal(args)
	h.broadcastSys <- mapB
}

func wsHandler() *WsServer {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		c := &connection{send: make(chan []byte, 256*10), ws: so}
		h.register <- c
		so.On("command", func(message string) {
			h.broadcast <- []byte(message)
		})

		so.On("disconnection", func() {
			h.unregister <- c
		})
		go c.writer()
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	wrapper := WsServer{
		Server: server,
	}

	return &wrapper
}
