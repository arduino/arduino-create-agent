package design

import . "goa.design/goa/dsl"

var _ = Service("indexes", func() {
	Description("The indexes service manages the package_index files")

	Error("invalid_url", ErrorResult, "url invalid")
	HTTP(func() {
		Response("invalid_url", StatusBadRequest)
	})

	Method("list", func() {
		Result(ArrayOf(String))
		HTTP(func() {
			GET("/pkgs/indexes")
			Response(StatusOK)
		})
	})

	Method("add", func() {
		Payload(IndexPayload)
		Result(Operation)
		HTTP(func() {
			POST("/pkgs/indexes/add")
			Response(StatusOK)
		})
	})

	Method("remove", func() {
		Payload(IndexPayload)
		Result(Operation)
		HTTP(func() {
			POST("/pkgs/indexes/delete")
			Response(StatusOK)
		})
	})
})

var _ = Service("tools", func() {
	Description("The tools service manages the available and installed tools")

	Method("available", func() {
		Result(CollectionOf(Tool))
		HTTP(func() {
			GET("/pkgs/tools/available")
			Response(StatusOK)
		})
	})

	Method("installed", func() {
		Result(CollectionOf(Tool))
		HTTP(func() {
			GET("/pkgs/tools/installed")
			Response(StatusOK)
		})
	})

	Method("install", func() {
		Error("not_found", ErrorResult, "tool not found")
		HTTP(func() {
			Response("not_found", StatusBadRequest)
		})
		Payload(ToolPayload)
		Result(Operation)
		HTTP(func() {
			POST("/pkgs/tools/installed")
			Response(StatusOK)
		})
	})

	Method("remove", func() {
		Payload(ToolPayload)
		Result(Operation)

		HTTP(func() {
			DELETE("/pkgs/tools/installed/{packager}/{name}/{version}")
			Response(StatusOK)
		})
	})
})

var IndexPayload = Type("arduino.index", func() {
	TypeName("IndexPayload")

	Attribute("url", String, "The url of the index file", func() {
		Example("http://downloads.arduino.cc/packages/package_index.json")
	})
	Required("url")
})

var ToolPayload = Type("arduino.tool", func() {
	Description(`A tool is an executable program that can upload sketches. 
	If url is absent the tool will be searched among the package index installed`)
	TypeName("ToolPayload")

	Attribute("name", String, "The name of the tool", func() {
		Example("avrdude")
	})
	Attribute("version", String, "The version of the tool", func() {
		Example("6.3.0-arduino9")
	})
	Attribute("packager", String, "The packager of the tool", func() {
		Example("arduino")
	})

	Attribute("url", String, `The url where the package can be found. Optional. 
	If present checksum must also be present.`)

	Attribute("checksum", String, `A checksum of the archive. Mandatory when url is present. 
	This ensures that the package is downloaded correcly.`)

	Required("name", "version", "packager")
})

var Tool = ResultType("application/vnd.arduino.tool", func() {
	Description("A tool is an executable program that can upload sketches.")
	TypeName("Tool")
	Reference(ToolPayload)

	Attribute("name")
	Attribute("version")
	Attribute("packager")

	Required("name", "version", "packager")
})

var Operation = ResultType("application/vnd.arduino.operation", func() {
	Description("Describes the result of an operation.")
	TypeName("Operation")

	Attribute("status", String, "The status of the operation", func() {
		Example("ok")
	})
	Required("status")
})
