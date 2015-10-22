// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/googollee/go-socket.io"
	"net/http"
	"strconv"
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

func uploadHandler(c *gin.Context) {
	log.Print("Received a upload")
	port := c.PostForm("port")
	if port == "" {
		c.String(http.StatusBadRequest, "port is required")
		return
	}
	board := c.PostForm("board")
	if board == "" {
		c.String(http.StatusBadRequest, "board is required")
		log.Error("board is required")
		return
	}
	board_rewrite := c.PostForm("board_rewrite")

	var extraInfo boardExtraInfo

	extraInfo.authdata.UserName = c.PostForm("auth_user")
	extraInfo.authdata.Password = c.PostForm("auth_pass")
	commandline := c.PostForm("commandline")
	if commandline == "undefined" {
		commandline = ""
	}
	extraInfo.use_1200bps_touch, _ = strconv.ParseBool(c.PostForm("use_1200bps_touch"))
	extraInfo.wait_for_upload_port, _ = strconv.ParseBool(c.PostForm("wait_for_upload_port"))
	extraInfo.networkPort, _ = strconv.ParseBool(c.PostForm("network"))

	if extraInfo.networkPort == false && commandline == "" {
		c.String(http.StatusBadRequest, "commandline is required for local board")
		log.Error("commandline is required for local board")
		return
	}

	sketch, header, err := c.Request.FormFile("sketch_hex")
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	if header != nil {
		path, err := saveFileonTempDir(header.Filename, sketch)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		go spProgramRW(port, board, board_rewrite, path, commandline, extraInfo)
	}
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
