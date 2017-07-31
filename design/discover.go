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
