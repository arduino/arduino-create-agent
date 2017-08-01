package main

import (
	"github.com/arduino/arduino-create-agent/app"
	"github.com/goadesign/goa"
	"golang.org/x/net/websocket"
	"io"
)

// ConnectV1Controller implements the connect_v1 resource.
type ConnectV1Controller struct {
	*goa.Controller
}

// NewConnectV1Controller creates a connect_v1 controller.
func NewConnectV1Controller(service *goa.Service) *ConnectV1Controller {
	return &ConnectV1Controller{Controller: service.NewController("ConnectV1Controller")}
}

// Websocket runs the websocket action.
func (c *ConnectV1Controller) Websocket(ctx *app.WebsocketConnectV1Context) error {
	c.WebsocketWSHandler(ctx).ServeHTTP(ctx.ResponseWriter, ctx.Request)
	return nil
}

// WebsocketWSHandler establishes a websocket connection to run the websocket action.
func (c *ConnectV1Controller) WebsocketWSHandler(ctx *app.WebsocketConnectV1Context) websocket.Handler {
	return func(ws *websocket.Conn) {
		// ConnectV1Controller_Websocket: start_implement

		// Put your logic here

		// ConnectV1Controller_Websocket: end_implement
		ws.Write([]byte("websocket connect_v1"))
		// Dummy echo websocket server
		io.Copy(ws, ws)
	}
}
