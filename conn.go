// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
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
		return
	}
	board_rewrite := c.PostForm("board_rewrite")
	sketch, header, err := c.Request.FormFile("sketch_hex")
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}

	if header != nil {
		path, err := saveFileonTempDir(header.Filename, sketch)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		go spProgramRW(port, board, board_rewrite, path)
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
