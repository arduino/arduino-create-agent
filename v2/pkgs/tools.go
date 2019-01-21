package pkgs

import (
	"context"
	"io/ioutil"
	"os/user"
	"path/filepath"

	"github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/sirupsen/logrus"
)

type Tools struct {
	Log     *logrus.Logger
	Indexes interface {
		List(context.Context) ([]string, error)
		Get(context.Context, string) (Index, error)
	}
	Folder string
}

func (c *Tools) Available(ctx context.Context) (res tools.ToolCollection, err error) {
	list, err := c.Indexes.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, url := range list {
		index, err := c.Indexes.Get(ctx, url)
		if err != nil {
			return nil, err
		}

		for _, packager := range index.Packages {
			for _, tool := range packager.Tools {
				res = append(res, &tools.Tool{
					Packager: packager.Name,
					Name:     tool.Name,
					Version:  tool.Version,
				})
			}
		}
	}

	return res, nil
}

func (c *Tools) Installed(ctx context.Context) (tools.ToolCollection, error) {
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

func (c *Tools) Install(ctx context.Context, payload *tools.ToolPayload) error {
	return nil
}

func (c *Tools) Remove(ctx context.Context, payload *tools.ToolPayload) error {
	return nil
}
