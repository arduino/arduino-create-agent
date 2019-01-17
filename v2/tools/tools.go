package tools

import (
	"context"
	"io/ioutil"
	"os/user"
	"path/filepath"

	"github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/sirupsen/logrus"
)

type Tools struct {
	Log *logrus.Logger
}

func (c *Tools) List(ctx context.Context) (tools.ToolCollection, error) {
	res := tools.ToolCollection{}

	// Find packagers
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, ".arduino-create")
	packagers, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, packager := range packagers {
		if !packager.IsDir() {
			continue
		}
		// Find tools
		toolss, err := ioutil.ReadDir(filepath.Join(path, packager.Name()))
		if err != nil {
			return nil, err
		}
		for _, tool := range toolss {
			// Find versions
			versions, err := ioutil.ReadDir(filepath.Join(path, packager.Name(), tool.Name()))
			if err != nil {
				return nil, err
			}

			for _, version := range versions {
				res = append(res, &tools.Tool{
					Packager: packager.Name(),
					Name:     tool.Name(),
					Version:  version.Name(),
				})
			}
		}
	}

	return res, nil
}
