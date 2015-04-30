// Supports Windows, Linux, Mac, and Raspberry Pi

package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		h.broadcast <- message
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Received a upload")
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
	sketch, header, err := r.FormFile("sketch_hex")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	path, err := saveFileonTempDir(header.Filename, sketch)

	go spProgram(port, board, path)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Started a new websocket handler")
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		return
	}
	//c := &connection{send: make(chan []byte, 256), ws: ws}
	c := &connection{send: make(chan []byte, 256*10), ws: ws}
	h.register <- c
	defer func() { h.unregister <- c }()
	go c.writer()
	c.reader()
}
