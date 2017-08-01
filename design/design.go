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
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = API("arduino-create-agent", func() {
	Title("Arduino Create Agent")
	Description("A bridge from the user's computer and the Create platform")
	Host("localhost:9000")
	Scheme("http")
	BasePath("/")
	Consumes("application/json")
	Produces("application/json")

	Origin("*", func() {
		Methods("GET", "PUT", "POST", "DELETE")
		Headers("Authorization", "Origin", "X-Requested-With", "Content-Type", "Accept")
		Credentials()
	})
})

var _ = Resource("public", func() {
	Metadata("swagger:generate", "false")

	Files("swagger.json", "swagger/swagger.json")
	Files("docs", "templates/docs.html")
	Files("debug", "templates/debug.html")
})
