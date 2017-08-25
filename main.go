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
//go:generate go run cli/gen/main.go

package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/arduino/arduino-create-agent/app"
	"github.com/arduino/arduino-create-agent/discovery"
	"github.com/getlantern/systray"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	"github.com/kardianos/osext"
)

func main() {
	var (
		hibernate = flag.Bool("hibernate", false, "start hibernated")
		version   = "x.x.x-dev" //don't modify it, Jenkins will take care
		revision  = "xxxxxxxx"  //don't modify it, Jenkins will take care
	)

	flag.Parse()

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

	// Mount "tools" controller
	t := NewToolsV1Controller(service)
	app.MountToolsV1Controller(service, t)

	// Mount "public" controller
	public := NewPublicController(service)
	app.MountPublicController(service, public)

	// Mount systray
	restart := restartFunc("", !*hibernate)
	shutdown := func() {
		os.Exit(0)
	}

	http, https := findPorts()
	address := "http://localhost:" + strconv.Itoa(http)
	if !*hibernate {
		// Start http service
		go func() {
			if err := service.ListenAndServe(":" + strconv.Itoa(http)); err != nil {
				service.LogError("startup", "err", err)
			}
		}()
		// Start https service
		go func() {
			if err := service.ListenAndServeTLS(":"+strconv.Itoa(https), "cert.pem", "key.pem"); err != nil {
				service.LogError("startup", "err", err)
			}
		}()
	}

	setupSystray(*hibernate, version, revision, address, restart, shutdown)
}

// findPorts returns the first two available ports for http and https listening
func findPorts() (http, https int) {
	start := 8990
	end := 9000

	// http
	for i := start; i < end; i++ {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(i))
		if err == nil {
			ln.Close()
			http = i
			break
		}
	}

	// https
	for j := http + 1; j < end; j++ {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(j))
		if err == nil {
			ln.Close()
			https = j
			break
		}
	}

	return http, https
}

// RestartFunc launches itself before exiting. It works because we pass an option to tell it to wait for 5 seconds, which gives us time to exit and unbind from serial ports and TCP/IP
// sockets like :8989
func restartFunc(path string, hibernate bool) func() {
	return func() {
		// Quit systray
		systray.Quit()

		// figure out current path of executable so we know how to restart
		// this process using osext
		exePath, err := osext.Executable()
		if err != nil {
			log.Fatalf("Error getting exe path using osext lib. err: %v\n", err)
		}

		if path == "" {
			log.Printf("exePath using osext: %v\n", exePath)
		} else {
			exePath = path
		}
		exePath = strings.Trim(exePath, "\n")
		hiberString := ""
		if hibernate {
			hiberString = "-hibernate"
		}

		// Execute command
		cmd := exec.Command(exePath, hiberString)
		err = cmd.Start()
		if err != nil {
			log.Fatalf("Got err restarting spjs: %v\n", err)
		}
		os.Exit(0)
	}
}
