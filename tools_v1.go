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
	"github.com/arduino/arduino-create-agent/tools"
	"github.com/goadesign/goa"
)

// ToolsV1Controller implements the tools_v1 resource.
type ToolsV1Controller struct {
	*goa.Controller
}

// NewToolsV1Controller creates a tools_v1 controller.
func NewToolsV1Controller(service *goa.Service) *ToolsV1Controller {
	return &ToolsV1Controller{Controller: service.NewController("ToolsV1Controller")}
}

// Download runs the download action.
func (c *ToolsV1Controller) Download(ctx *app.DownloadToolsV1Context) error {
	tool, err := tools.Download(ctx.Packager, ctx.Name, ctx.Version, nil)
	if err != nil {
		return err
	}
	res := &app.ArduinoAgentToolsTool{
		Name:     tool.Name,
		Packager: tool.Packager,
		Version:  tool.Version,
		Path:     tool.Path,
	}
	return ctx.OK(res)
}

// List runs the list action.
func (c *ToolsV1Controller) List(ctx *app.ListToolsV1Context) error {
	list, err := tools.Installed(nil)
	if err != nil {
		return err
	}
	res := app.ArduinoAgentToolsToolCollection{}

	for _, tool := range list {
		res = append(res, &app.ArduinoAgentToolsTool{
			Name:     tool.Name,
			Packager: tool.Packager,
			Version:  tool.Version,
			Path:     tool.Path,
		})
	}

	return ctx.OK(res)
}
