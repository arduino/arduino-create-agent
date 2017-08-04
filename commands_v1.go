package main

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
	res := &app.ArduinoAgentExec{}
	return ctx.OK(res)
}

// List runs the list action.
func (c *CommandsV1Controller) List(ctx *app.ListCommandsV1Context) error {
	// CommandsV1Controller_List: start_implement

	// Put your logic here

	// CommandsV1Controller_List: end_implement
	res := app.ArduinoAgentCommandCollection{}
	return ctx.OK(res)
}
