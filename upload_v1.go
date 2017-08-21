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
package main

import (
	"github.com/arduino/arduino-create-agent/app"
	"github.com/goadesign/goa"
)

// UploadV1Controller implements the upload_v1 resource.
type UploadV1Controller struct {
	*goa.Controller
}

// NewUploadV1Controller creates a upload_v1 controller.
func NewUploadV1Controller(service *goa.Service) *UploadV1Controller {
	return &UploadV1Controller{Controller: service.NewController("UploadV1Controller")}
}

// Serial runs the serial action.
func (c *UploadV1Controller) Serial(ctx *app.SerialUploadV1Context) error {
	// UploadV1Controller_Serial: start_implement

	// Put your logic here

	// UploadV1Controller_Serial: end_implement
	return nil
}

// Show runs the show action.
func (c *UploadV1Controller) Show(ctx *app.ShowUploadV1Context) error {
	// UploadV1Controller_Show: start_implement

	// Put your logic here

	// UploadV1Controller_Show: end_implement
	res := &app.ArduinoAgentExec{}
	return ctx.OK(res)
}
