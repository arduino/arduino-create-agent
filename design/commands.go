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

var _ = Resource("commands_v1", func() {
	Action("list", func() {
		Description("List the commands that the arduino-create-agent can perform on the machine")
		Routing(GET(""))
		Response(OK, CollectionOf(CommandV1))
	})
	Action("exec", func() {
		Description("Execute a command. Note that if you want to upload a sketch you'll probably be better off with the upload api. This one is very low-level")
		Routing(POST("/:id"))
		Params(func() {
			Param("id", String, "The id of the command")
		})
		Payload(ArrayOf(CommandParamV1))
		Response(OK, ExecResultV1)
	})
})

var CommandV1 = MediaType("application/vnd.arduino.agent.command+json", func() {
	Description("A command that the arduino-create-agent can perform on the machine")
	Attributes(func() {
		Attribute("id", String, "A unique identifier for the command", func() {
			Example("upload:arduino:avr:uno")
		})
		Attribute("pattern", String, "The command that will be executed", func() {
			Example(`"{runtime.tools.avrdude.path}/bin/avrdude" "-C{runtime.tools.avrdude.path}/etc/avrdude.conf" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b115200 -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"`)
		})
		Attribute("params", ArrayOf(String), "The command that will be executed", func() {
			Example([]string{"upload.verbose", "serial.port", "build.path", "build.project_name"})
		})
	})

	Required("id", "pattern", "params")

	View("default", func() {
		Attribute("id")
		Attribute("pattern")
		Attribute("params")
	})
})

var CommandParamV1 = Type("command.param", func() {
	Description("Option for the command to execute")
	Attribute("name", String, "The name of the param", func() {
		Example("serial.port")
	})
	Attribute("value", String, "The value of the option", func() {
		Example("/dev/ttyACM0")
	})
})

var ExecResultV1 = MediaType("application/vnd.arduino.agent.exec+json", func() {
	Description("The result of the command executed on the machine")
	Attributes(func() {
		Attribute("stdout", String, "The standard output returned by the command")
		Attribute("stderr", String, "The standard error returned by the command")
	})

	Required("stdout", "stderr")

	View("default", func() {
		Attribute("stdout")
		Attribute("stderr")
	})
})
