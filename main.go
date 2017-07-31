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

	// Mount "public" controller
	public := NewPublicController(service)
	app.MountPublicController(service, public)

	// Start service
	if err := service.ListenAndServe(":9000"); err != nil {
		service.LogError("startup", "err", err)
	}

}
