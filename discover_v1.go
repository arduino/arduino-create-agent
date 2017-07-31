package main

import (
	"github.com/arduino/arduino-create-agent/app"
	"github.com/arduino/arduino-create-agent/discovery"
	"github.com/goadesign/goa"
)

// DiscoverV1Controller implements the discover_v1 resource.
type DiscoverV1Controller struct {
	*goa.Controller
	Monitor *discovery.Monitor
}

// NewDiscoverV1Controller creates a discover_v1 controller.
func NewDiscoverV1Controller(service *goa.Service, m *discovery.Monitor) *DiscoverV1Controller {
	return &DiscoverV1Controller{
		Controller: service.NewController("DiscoverV1Controller"),
		Monitor:    m,
	}
}

// List runs the list action.
func (c *DiscoverV1Controller) List(ctx *app.ListDiscoverV1Context) error {
	serial := c.Monitor.Serial()
	network := c.Monitor.Network()

	res := &app.ArduinoAgentDiscover{
		Serial:  app.ArduinoAgentDiscoverSerialCollection{},
		Network: app.ArduinoAgentDiscoverNetworkCollection{},
	}

	for i := range serial {
		s := &app.ArduinoAgentDiscoverSerial{
			Vid:  serial[i].VendorID,
			Pid:  serial[i].ProductID,
			Port: serial[i].Port,
		}

		if serial[i].SerialNumber != "" {
			s.Serial = &serial[i].SerialNumber
		}

		res.Serial = append(res.Serial, s)
	}

	for i := range network {
		n := &app.ArduinoAgentDiscoverNetwork{
			Address: network[i].Address,
			Port:    network[i].Port,
			Info:    network[i].Info,
			Name:    network[i].Name,
		}

		res.Network = append(res.Network, n)
	}

	return ctx.OK(res)
}
