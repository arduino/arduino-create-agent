package tools

import (
	"context"

	"github.com/arduino/arduino-create-agent/gen/tools"
)

type Tools struct {
}

func (c *Tools) List(ctx context.Context) (tools.ToolCollection, error) {
	return tools.ToolCollection{}, nil
}
