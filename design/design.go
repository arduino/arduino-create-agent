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

var _ = API("arduino-create-agent", func() {
	Title("Arduino Create Agent")
	Description(`A companion of Arduino Create. 
	Allows the website to perform operations on the user computer, 
	such as detecting which boards are connected and upload sketches on them.`)
	HTTP(func() {
		Path("/v2")
		Consumes("application/json")
		Consumes("plain/text")
	})
})
