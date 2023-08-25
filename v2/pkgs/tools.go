// Copyright 2022 Arduino SA
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package pkgs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/codeclysm/extract/v3"
)

// Tools is a client that implements github.com/arduino/arduino-create-agent/gen/tools.Service interface.
// It saves tools in a specified folder with this structure: packager/name/version
// For example:
//
//	folder
//	└── arduino
//	    └── bossac
//	        ├── 1.6.1-arduino
//	        │   └── bossac
//	        └── 1.7.0
//	            └── bossac
//
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
	packagers, err := os.ReadDir(c.Folder)
	if err != nil {
		if !strings.Contains(err.Error(), "no such file") {
			return nil, err
		}
		err = os.MkdirAll(c.Folder, 0755)
		if err != nil {
			return nil, err
		}
	}

	for _, packager := range packagers {
		if !packager.IsDir() {
			continue
		}

		// Find tools
		toolss, err := os.ReadDir(filepath.Join(c.Folder, packager.Name()))
		if err != nil {
			return nil, err
		}

		for _, tool := range toolss {
			// Find versions
			path := filepath.Join(c.Folder, packager.Name(), tool.Name())
			versions, err := os.ReadDir(path)
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
func (c *Tools) Install(ctx context.Context, payload *tools.ToolPayload) (*tools.Operation, error) {
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

	return nil, tools.MakeNotFound(
		fmt.Errorf("tool not found with packager '%s', name '%s', version '%s'",
			payload.Packager, payload.Name, payload.Version))
}

func (c *Tools) install(ctx context.Context, packager string, tool Tool) (*tools.Operation, error) {
	sys := tool.GetFlavourCompatibleWith(runtime.GOOS, runtime.GOARCH)
	path := filepath.Join(packager, tool.Name, tool.Version)

	// Download
	res, err := http.Get(sys.URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Use a teereader to only read once
	var buffer bytes.Buffer
	reader := io.TeeReader(res.Body, &buffer)

	// Cleanup
	err = os.RemoveAll(filepath.Join(c.Folder, path))
	if err != nil {
		return nil, err
	}

	err = extract.Archive(ctx, reader, c.Folder, rename(path))
	if err != nil {
		os.RemoveAll(path)
		return nil, err
	}

	sum := sha256.Sum256(buffer.Bytes())
	sumString := "SHA-256:" + hex.EncodeToString(sum[:sha256.Size])

	if sumString != sys.Checksum {
		os.RemoveAll(path)
		return nil, errors.New("checksum doesn't match")
	}

	// Write installed.json for retrocompatibility with v1
	err = writeInstalled(c.Folder, path)
	if err != nil {
		return nil, err
	}

	return &tools.Operation{Status: "ok"}, nil
}

// Remove deletes the tool folder from Tools Folder
func (c *Tools) Remove(ctx context.Context, payload *tools.ToolPayload) (*tools.Operation, error) {
	path := filepath.Join(payload.Packager, payload.Name, payload.Version)

	err := os.RemoveAll(filepath.Join(c.Folder, path))
	if err != nil {
		return nil, err
	}

	return &tools.Operation{Status: "ok"}, nil
}

func rename(base string) extract.Renamer {
	return func(path string) string {
		parts := strings.Split(filepath.ToSlash(path), "/")
		path = strings.Join(parts[1:], "/")
		path = filepath.Join(base, path)
		return path
	}
}

func writeInstalled(folder, path string) error {
	// read installed.json
	installed := map[string]string{}

	data, err := os.ReadFile(filepath.Join(folder, "installed.json"))
	if err == nil {
		err = json.Unmarshal(data, &installed)
		if err != nil {
			return err
		}
	}

	parts := strings.Split(path, string(filepath.Separator))
	tool := parts[len(parts)-2]
	toolWithVersion := fmt.Sprint(tool, "-", parts[len(parts)-1])
	installed[tool] = filepath.Join(folder, path)
	installed[toolWithVersion] = filepath.Join(folder, path)

	data, err = json.Marshal(installed)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(folder, "installed.json"), data, 0644)
}
