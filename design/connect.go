package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("connect_v1", func() {
	Action("websocket", func() {
		Routing(GET("/v1/connect"))
		Scheme("ws")
		Description("Opens a websocket connection to the device, allowing to read/write")
		Params(func() {
			Param("port", String, "The port where the device resides", func() {
				Example("/dev/ttyACM0")
			})
			Param("baud", Integer, "The speed of the connection. Defaults to 9600", func() {
				Example(9600)
				Default(9600)
			})
			Required("port", "baud")
		})
		Response(BadRequest)
		Response(SwitchingProtocols)
	})
})
