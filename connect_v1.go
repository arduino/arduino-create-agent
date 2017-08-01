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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */
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
