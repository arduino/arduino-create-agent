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
