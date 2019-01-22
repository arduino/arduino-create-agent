package pkgs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/codeclysm/extract"
	"github.com/sirupsen/logrus"
	"github.com/xrash/smetrics"
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
	list, err := c.Indexes.List(ctx)
	if err != nil {
		return err
	}

	for _, url := range list {
		index, err := c.Indexes.Get(ctx, url)
		if err != nil {
			return err
		}

		for _, packager := range index.Packages {
			if packager.Name != payload.Packager {
				continue
			}

			for _, tool := range packager.Tools {
				if tool.Name == payload.Name &&
					tool.Version == payload.Version {
					return c.install(ctx, payload.Packager, tool)
				}
			}
		}
	}

	return tools.MakeNotFound(
		fmt.Errorf("tool not found with packager '%s', name '%s', version '%s'",
			payload.Packager, payload.Name, payload.Version))
}

func (c *Tools) install(ctx context.Context, packager string, tool Tool) error {
	i := findSystem(tool)

	// Download
	fmt.Println(tool.Systems[i].URL)
	res, err := http.Get(tool.Systems[i].URL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Use a teereader to only read once
	var buffer bytes.Buffer
	reader := io.TeeReader(res.Body, &buffer)

	basepath := filepath.Join(packager, tool.Name, tool.Version)
	err = extract.Archive(ctx, reader, c.Folder, rename(basepath))
	if err != nil {
		return err
	}

	checksum := sha256.Sum256(buffer.Bytes())
	checkSumString := "SHA-256:" + hex.EncodeToString(checksum[:sha256.Size])

	if checkSumString != tool.Systems[i].Checksum {
		os.RemoveAll(basepath)
		return errors.New("checksum doesn't match")
	}

	return nil
}

func (c *Tools) Remove(ctx context.Context, payload *tools.ToolPayload) error {
	return nil
}

func rename(base string) extract.Renamer {
	return func(path string) string {
		parts := strings.Split(path, string(filepath.Separator))
		path = strings.Join(parts[1:], string(filepath.Separator))
		path = filepath.Join(base, path)
		return path
	}
}

func findSystem(tool Tool) int {
	var systems = map[string]string{
		"linuxamd64":   "x86_64-linux-gnu",
		"linux386":     "i686-linux-gnu",
		"darwinamd64":  "apple-darwin",
		"windows386":   "i686-mingw32",
		"windowsamd64": "i686-mingw32",
		"linuxarm":     "arm-linux-gnueabihf",
	}

	var correctSystem int
	maxSimilarity := 0.7

	for i, system := range tool.Systems {
		similarity := smetrics.Jaro(system.Host, systems[runtime.GOOS+runtime.GOARCH])
		if similarity > maxSimilarity {
			correctSystem = i
			maxSimilarity = similarity
		}
	}

	return correctSystem
}
