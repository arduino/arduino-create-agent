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
//go:generate go run cli/gen/main.go

package main

import (
	"time"

	"golang.org/x/net/context"

	"github.com/arduino/arduino-create-agent/app"
	"github.com/arduino/arduino-create-agent/discovery"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
)

func main() {
	// Create service
	service := goa.New("arduino-create-agent")

	// Start monitor
	monitor := discovery.New(1 * time.Second)
	monitor.Start(context.Background())

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	// Mount "discovery" controller
	d := NewDiscoverV1Controller(service, monitor)
	app.MountDiscoverV1Controller(service, d)

	// Mount "connect" controller
	c := NewConnectV1Controller(service)
	app.MountConnectV1Controller(service, c)

	// Mount "public" controller
	public := NewPublicController(service)
	app.MountPublicController(service, public)

	// Start service
	if err := service.ListenAndServe(":9000"); err != nil {
		service.LogError("startup", "err", err)
	}

}
