/*
 * This file is part of arduino-create-agent.
 *
 * arduino-create-agent is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */
package agent

import (
	"bytes"

	"golang.org/x/net/context"

	serial "go.bug.st/serial.v1"

	"github.com/arduino/arduino-create-agent/app"
	"github.com/codeclysm/cc"
	"github.com/goadesign/goa"
	"github.com/gorilla/websocket"
)

type conns map[*websocket.Conn]*cc.Stoppable

// ConnectV1Controller implements the connect_v1 resource.
type ConnectV1Controller struct {
	*goa.Controller
	sockets conns
}

// NewConnectV1Controller creates a connect_v1 controller.
func NewConnectV1Controller(service *goa.Service) *ConnectV1Controller {
	return &ConnectV1Controller{
		Controller: service.NewController("ConnectV1Controller"),
		sockets:    make(conns),
	}
}

// Websocket runs the websocket action.
func (c *ConnectV1Controller) Websocket(ctx *app.WebsocketConnectV1Context) error {
	// Open port
	mode := &serial.Mode{
		BaudRate: ctx.Baud,
	}
	port, err := serial.Open(ctx.Port, mode)
	if err != nil {
		goa.LogError(ctx, err.Error())
		return ctx.BadRequest()
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(ctx.ResponseWriter, ctx.Request, nil)
	if err != nil {
		goa.LogError(ctx, err.Error())
		return ctx.BadRequest()
	}

	c.sockets[conn] = cc.Run(listen(ctx, conn, port))

	<-c.sockets[conn].Stopped
	delete(c.sockets, conn)

	conn.Close()

	return nil
}

// StopAll stops all websocket connections
func (c *ConnectV1Controller) StopAll() {
	for conn, stoppable := range c.sockets {
		conn.Close()
		stoppable.Stop()
		<-stoppable.Stopped
	}
}

func listen(ctx context.Context, conn *websocket.Conn, port serial.Port) (reader, writer cc.StoppableFunc) {
	reader = func(done chan struct{}) {
	L:
		for {
			select {
			case <-done:
				break L

			default:
				msg := make([]byte, 1024)
				n, err := port.Read(msg)
				if err != nil {
					goa.LogError(ctx, err.Error(), "when", "read from port")
					break L
				}
				if n > 0 {
					conn.WriteMessage(websocket.TextMessage, bytes.Trim(msg, "\x00"))
				}
			}
		}
	}

	writer = func(done chan struct{}) {
	L:
		for {
			select {
			case <-done:
				break L

			default:
				_, msg, err := conn.ReadMessage()
				if err != nil {
					goa.LogError(ctx, err.Error(), "when", "read from websocket")
					break L
				}
				_, err = port.Write(msg)
				if err != nil {
					goa.LogError(ctx, err.Error(), "when", "write on port")
					break L
				}
			}
		}
		port.Close()
		conn.Close()
	}

	return reader, writer
}
