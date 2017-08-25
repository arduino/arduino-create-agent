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
package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("upload_v1", func() {
	Action("show", func() {
		Description("Retrieve the status of a running command")
		Routing(GET("/:id"))
		Response(OK, ExecResultV1)
	})
	Action("serial", func() {
		Description("Performs an upload of a sketch over the serial port")
		Routing(POST(""))
		Payload(ArrayOf(UploadSerialV1))
		Response(Accepted, func() {
			Headers(func() {
				Header("Location", String, "Contains the location of the show resource")
				Required("Location")
			})
		})
	})
})

var UploadSerialV1 = Type("upload.serial", func() {
	Description("The necessary info to upload a sketch over a serial port")
	Attribute("port", String, "The serial port", func() {
		Example("/dev/ttyACM0")
	})
	Attribute("command", String, "The id of the command to use (See commands#list)", func() {
		Example("upload:arduino:avr:uno")
	})
	Attribute("bin", String, "Base64-encoded binary file", func() {
		Example("QmFzZTY0IGlzIGEgZ2VuZ...")
	})
	Attribute("filename", String, "The name of the binary file", func() {
		Example("QmFzZTY0IGlzIGEgZ2VuZ...")
	})
	Attribute("params", ArrayOf(CommandParamV1), "Params")
})
