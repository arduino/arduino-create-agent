// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
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

func (s *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	s.Server.ServeHTTP(w, r)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Received a upload")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	port := r.FormValue("port")
	if port == "" {
		http.Error(w, "port is required", http.StatusBadRequest)
		return
	}
	board := r.FormValue("board")
	if board == "" {
		http.Error(w, "board is required", http.StatusBadRequest)
		return
	}
	board_rewrite := r.FormValue("board_rewrite")
	sketch, header, err := r.FormFile("sketch_hex")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if header != nil {
		path, err := saveFileonTempDir(header.Filename, sketch)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
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
