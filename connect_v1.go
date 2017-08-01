package main

import (
	"log"

	"github.com/arduino/arduino-create-agent/app"
	"github.com/arduino/arduino-create-agent/connect"
	"github.com/goadesign/goa"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

type device struct {
	input  chan []byte
	output chan []byte
	cancel func()
}

// ConnectV1Controller implements the connect_v1 resource.
type ConnectV1Controller struct {
	*goa.Controller
	devices map[string]device
}

// NewConnectV1Controller creates a connect_v1 controller.
func NewConnectV1Controller(service *goa.Service) *ConnectV1Controller {
	return &ConnectV1Controller{
		Controller: service.NewController("ConnectV1Controller"),
		devices:    make(map[string]device),
	}
}

// Websocket runs the websocket action.
func (c *ConnectV1Controller) Websocket(ctx *app.WebsocketConnectV1Context) error {
	cont, cancel := context.WithCancel(context.Background())
	input, output, err := connect.Open(cont, ctx.Port, ctx.Baud)
	if err != nil {
		return ctx.BadRequest()
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(ctx.ResponseWriter, ctx.Request, nil)
	if err != nil {
		return ctx.BadRequest()
	}

	go func() {
		for msg := range output {
			log.Println(len(msg))

			conn.WriteMessage(websocket.TextMessage, msg)
		}
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		input <- msg
	}

	cancel()

	return nil
}
