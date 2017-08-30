// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/arduino/arduino-create-agent/upload"
	"github.com/arduino/arduino-create-agent/utilities"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

type connection struct {
	// The websocket connection.
	ws socketio.Socket

	// Buffered channel of outbound messages.
	send     chan []byte
	incoming chan []byte
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

type AdditionalFile struct {
	Hex      []byte `json:"hex"`
	Filename string `json:"filename"`
}

// Upload contains the data to upload a sketch onto a board
type Upload struct {
	Port        string           `json:"port"`
	Board       string           `json:"board"`
	Rewrite     string           `json:"rewrite"`
	Commandline string           `json:"commandline"`
	Extra       upload.Extra     `json:"extra"`
	Hex         []byte           `json:"hex"`
	Filename    string           `json:"filename"`
	ExtraFiles  []AdditionalFile `json:"extrafiles"`
}

var uploadStatusStr string = "ProgrammerStatus"

func uploadHandler(c *gin.Context) {

	data := new(Upload)
	c.BindJSON(data)

	log.Printf("%+v", data)

	if data.Port == "" {
		c.String(http.StatusBadRequest, "port is required")
		return
	}

	// Check that the port is available
	found := false
	for i := range SerialPorts.Ports {
		if data.Port == SerialPorts.Ports[i].Name {
			found = true
			break
		}
	}

	for i := range NetworkPorts.Ports {
		if data.Port == NetworkPorts.Ports[i].Name {
			found = true
			break
		}
	}

	if !found {
		c.String(http.StatusBadRequest, "port is not available")
		log.Error("port is not available")
		return
	}

	if data.Board == "" {
		c.String(http.StatusBadRequest, "board is required")
		log.Error("board is required")
		return
	}

	if data.Extra.Network == false {
		log.Println(data.Commandline)

		if data.Extra.Verbose {
			data.Commandline = commands["verbose"+data.Commandline]
		} else {
			data.Commandline = commands["quiet"+data.Commandline]
		}

		log.Println(data.Commandline)

		if data.Commandline == "" {
			c.String(http.StatusBadRequest, "commandline is required for local board")
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

	for _, extraFile := range data.ExtraFiles {
		path := filepath.Join(filepath.Dir(filePath), extraFile.Filename)
		filePaths = append(filePaths, path)
		log.Printf("Saving %s on %s", extraFile.Filename, path)
		err := ioutil.WriteFile(path, extraFile.Hex, 0644)
		if err != nil {
			log.Printf(err.Error())
		}
	}

	if data.Rewrite != "" {
		data.Board = data.Rewrite
	}

	go func() {
		// Resolve commandline
		commandline, err := upload.PartiallyResolve(data.Board, filePath, data.Commandline, data.Extra, &Tools)
		if err != nil {
			send(map[string]string{uploadStatusStr: "Error", "Msg": err.Error()})
			return
		}

		l := PLogger{Verbose: data.Extra.Verbose}

		// Upload
		if data.Extra.Network {
			send(map[string]string{uploadStatusStr: "Starting", "Cmd": "Network"})
			err = upload.Network(data.Port, data.Board, filePaths, commandline, data.Extra.Auth, l, data.Extra.SSH)
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

// Debug only sends messages if verbose is true
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

func verifyCommandLine(input string, signature string) error {
	sign, _ := hex.DecodeString(signature)
	block, _ := pem.Decode([]byte(*signatureKey))
	if block == nil {
		return errors.New("invalid key")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	rsaKey := key.(*rsa.PublicKey)
	h := sha256.New()
	h.Write([]byte(input))
	d := h.Sum(nil)
	return rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, d, sign)
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
