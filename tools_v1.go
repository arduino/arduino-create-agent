package main

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
