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
package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("discover_v1", func() {
	BasePath("/v1/discover")

	Action("list", func() {
		Description("Returns a list of devices connected to the PC")
		Routing(GET(""))
		Response(OK, DiscoverV1)
	})
})

var DiscoverV1 = MediaType("application/vnd.arduino.agent.discover+json", func() {
	Attributes(func() {
		Attribute("serial", CollectionOf(DiscoverSerialV1), "A list of devices connected through the serial port")
		Attribute("network", CollectionOf(DiscoverNetworkV1), "A list of devices connected through the network")
	})

	Required("serial", "network")

	View("default", func() {
		Attribute("serial")
		Attribute("network")
	})
})

var DiscoverSerialV1 = MediaType("application/vnd.arduino.agent.discover.serial+json", func() {
	Attributes(func() {
		Attribute("vid", String, "Vendor ID", func() {
			Example("0x2341")
		})
		Attribute("pid", String, "Vendor ID", func() {
			Example("0x8036")
		})
		Attribute("serial", String, "The Serial Number")
		Attribute("port", String, "The port through which it's connected", func() {
			Example("/dev/ttyACM0")
		})
	})

	Required("vid", "pid", "port")

	View("default", func() {
		Attribute("vid")
		Attribute("pid")
		Attribute("port")
		Attribute("serial")
	})

})

var DiscoverNetworkV1 = MediaType("application/vnd.arduino.agent.discover.network+json", func() {
	Attributes(func() {
		Attribute("address", String, "IP Address", func() {
			Example("192.168.1.107")
		})
		Attribute("port", Integer, "IP Port", func() {
			Example(80)
		})
		Attribute("info", String, "Informations about the device", func() {
			Example(`board=Arduino Y\195\186n Shield distro_version=0.1`)
		})
		Attribute("name", String, "The friendly name given to the device", func() {
			Example("MyShield")
		})
	})

	Required("address", "port", "info", "name")

	View("default", func() {
		Attribute("address")
		Attribute("port")
		Attribute("info")
		Attribute("name")
	})
})
