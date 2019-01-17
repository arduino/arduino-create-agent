package design

import . "goa.design/goa/dsl"

var _ = Service("tools", func() {
	Description("The tools service managed the tools installed in the system.")
	Method("list", func() {
		Result(Tool)
		HTTP(func() {
			GET("/tools")
			Response(StatusOK)
		})
	})
})

var Tool = ResultType("application/vnd.arduino.tool", func() {
	Description("A tool is an executable program that can upload sketches.")
	TypeName("Tool")

	Attributes(func() {
		Attribute("name", String, "The name of the tool", func() {
			Example("avrdude")
		})
		Attribute("version", String, "The version of the tool", func() {
			Example("6.3.0-arduino9")
		})
		Attribute("packager", String, "The packager of the tool", func() {
			Example("arduino")
		})
	})

	Required("name", "version", "packager")
})
