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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/codeclysm/extract"
	"github.com/xrash/smetrics"
)

// Tools is a client that implements github.com/arduino/arduino-create-agent/gen/tools.Service interface.
// It saves tools in a specified folder with this structure: packager/name/version
// For example:
//   folder
//   └── arduino
//       └── bossac
//           ├── 1.6.1-arduino
//           │   └── bossac
//           └── 1.7.0
//               └── bossac
// It requires an Indexes client to list and read package index files: use the Indexes struct
type Tools struct {
	Indexes interface {
		List(context.Context) ([]string, error)
		Get(context.Context, string) (Index, error)
	}
	Folder string
}

// Available crawles the downloaded package index files and returns a list of tools that can be installed.
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

// Installed crawles the Tools Folder and finds the installed tools.
func (c *Tools) Installed(ctx context.Context) (tools.ToolCollection, error) {
	res := tools.ToolCollection{}

	// Find packagers
	packagers, err := ioutil.ReadDir(c.Folder)
	if err != nil {
		return nil, err
	}

	for _, packager := range packagers {
		if !packager.IsDir() {
			continue
		}

		// Find tools
		toolss, err := ioutil.ReadDir(filepath.Join(c.Folder, packager.Name()))
		if err != nil {
			return nil, err
		}

		for _, tool := range toolss {
			// Find versions
			path := filepath.Join(c.Folder, packager.Name(), tool.Name())
			versions, err := ioutil.ReadDir(path)
			if err != nil {
				continue // we ignore errors because the folders could be dirty
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

// Install crawles the Index folder, downloads the specified tool, extracts the archive in the Tools Folder.
// It checks for the Signature specified in the package index.
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

// Remove deletes the tool folder from Tools Folder
func (c *Tools) Remove(ctx context.Context, payload *tools.ToolPayload) error {
	path := filepath.Join(payload.Packager, payload.Name, payload.Version)

	return os.RemoveAll(filepath.Join(c.Folder, path))
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
