// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/googollee/go-socket.io"
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

// Upload contains the data to upload a sketch onto a board
type Upload struct {
	Port        string         `json:"port"`
	Board       string         `json:"board"`
	Rewrite     string         `json:"rewrite"`
	Commandline string         `json:"commandline"`
	Signature   string         `json:"signature"`
	Extra       boardExtraInfo `json:"extra"`
	Hex         []byte         `json:"hex"`
	Filename    string         `json:"filename"`
}

func uploadHandler(c *gin.Context) {
	data := new(Upload)
	c.BindJSON(data)

	log.Printf("%+v", data)

	if data.Port == "" {
		c.String(http.StatusBadRequest, "port is required")
		return
	}

	if data.Board == "" {
		c.String(http.StatusBadRequest, "board is required")
		log.Error("board is required")
		return
	}

	if data.Signature == "" {
		c.String(http.StatusBadRequest, "signature is required")
		return
	}

	if extraInfo.networkPort {
		err := verifyCommandLine(data.Commandline, data.Signature)

		if err != nil {
			c.String(http.StatusBadRequest, "signature is invalid")
			return
		}
	}

	if data.Extra.Network == false && data.Commandline == "" {
		c.String(http.StatusBadRequest, "commandline is required for local board")
		return
	}

	buffer := bytes.NewBuffer(data.Hex)

	path, err := saveFileonTempDir(data.Filename, buffer)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	if data.Rewrite != "" {
		data.Board = data.Rewrite
	}

	go spProgramRW(data.Port, data.Board, path, data.Commandline, data.Extra)

	c.String(http.StatusAccepted, "")
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
