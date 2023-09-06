// Copyright 2022 Arduino SA
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package design

import . "goa.design/goa/v3/dsl"

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
		Example("https://downloads.arduino.cc/packages/package_index.json")
	})
	Required("url")
})

var ToolPayload = Type("arduino.tool", func() {
	Description(`A tool is an executable program that can upload sketches. 
	If url is absent the tool will be searched among the package index installed`)
	TypeName("ToolPayload")

	Attribute("name", String, "The name of the tool", func() {
		Example("bossac")
	})
	Attribute("version", String, "The version of the tool", func() {
		Example("1.7.0-arduino3")
	})
	Attribute("packager", String, "The packager of the tool", func() {
		Example("arduino")
	})

	Attribute("url", String, `The url where the package can be found. Optional. 
	If present checksum must also be present.`, func() {
		Example("http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux64.tar.gz")
	})

	Attribute("checksum", String, `A checksum of the archive. Mandatory when url is present. 
	This ensures that the package is downloaded correcly.`, func() {
		Example("SHA-256:1ae54999c1f97234a5c603eb99ad39313b11746a4ca517269a9285afa05f9100")
	})

	Attribute("signature", String, `The signature used to sign the url. Mandatory when url is present.
	This ensure the security of the file downloaded`, func() {
		Example("382898a97b5a86edd74208f10107d2fecbf7059ffe9cc856e045266fb4db4e98802728a0859cfdcda1c0b9075ec01e42dbea1f430b813530d5a6ae1766dfbba64c3e689b59758062dc2ab2e32b2a3491dc2b9a80b9cda4ae514fbe0ec5af210111b6896976053ab76bac55bcecfcececa68adfa3299e3cde6b7f117b3552a7d80ca419374bb497e3c3f12b640cf5b20875416b45e662fc6150b99b178f8e41d6982b4c0a255925ea39773683f9aa9201dc5768b6fc857c87ff602b6a93452a541b8ec10ca07f166e61a9e9d91f0a6090bd2038ed4427af6251039fb9fe8eb62ec30d7b0f3df38bc9de7204dec478fb86f8eb3f71543710790ee169dce039d3e0")
	})

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
