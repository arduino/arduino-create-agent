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
package agent

import (
	"github.com/arduino/arduino-create-agent/app"
	"github.com/goadesign/goa"
)

// CommandsV1Controller implements the commands_v1 resource.
type CommandsV1Controller struct {
	*goa.Controller
}

// NewCommandsV1Controller creates a commands_v1 controller.
func NewCommandsV1Controller(service *goa.Service) *CommandsV1Controller {
	return &CommandsV1Controller{Controller: service.NewController("CommandsV1Controller")}
}

// Exec runs the exec action.
func (c *CommandsV1Controller) Exec(ctx *app.ExecCommandsV1Context) error {
	// CommandsV1Controller_Exec: start_implement

	// Put your logic here

	// CommandsV1Controller_Exec: end_implement
	return ctx.Accepted()
}

// List runs the list action.
func (c *CommandsV1Controller) List(ctx *app.ListCommandsV1Context) error {
	// CommandsV1Controller_List: start_implement

	// Put your logic here

	// CommandsV1Controller_List: end_implement
	res := app.ArduinoAgentCommandCollection{}
	return ctx.OK(res)
}
