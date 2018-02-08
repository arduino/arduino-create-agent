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

var _ = Resource("manage_v1", func() {
	BasePath("/v1/manage")

	Action("info", func() {
		Description("Returns the info about the agent")
		Routing(GET("/info"))
		Response(OK, ManageInfoV1)
	})
	Action("pause", func() {
		Description("Restarts the agent in hibernation mode")
		Routing(POST("/pause"))
		Response(OK)
	})
	Action("update", func() {
		Description("Search for a new version, updates and restarts itself")
		Routing(POST("/update"))
		Response(OK)
	})
})

var ManageInfoV1 = MediaType("application/vnd.arduino.agent.manage.info+json", func() {
	Attributes(func() {
		Attribute("version", String, "Version", func() {
			Example("v1.0.0")
		})
		Attribute("revision", String, "Revision", func() {
			Example("4e498c5afc6304b6dc7a74534dbff96015954d48")
		})
	})

	Required("version", "revision")

	View("default", func() {
		Attribute("version")
		Attribute("revision")
	})
})
