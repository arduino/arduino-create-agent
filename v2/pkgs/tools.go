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
	"github.com/arduino/arduino-create-agent/index"
	"github.com/arduino/arduino-create-agent/utilities"
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
// It requires an Index Resource to search for tools
type Tools struct {
	index  *index.Resource
	folder string
}

// New will return a Tool object, allowing the caller to execute operations on it.
// The New function will accept an index as parameter (used to download the indexes)
// and a folder used to download the indexes
func New(index *index.Resource, folder string) *Tools {
	return &Tools{
		index:  index,
		folder: folder,
	}
}

// Available crawles the downloaded package index files and returns a list of tools that can be installed.
func (t *Tools) Available(ctx context.Context) (res tools.ToolCollection, err error) {
	body, err := t.index.Read()
	if err != nil {
		return nil, err
	}

	var index Index
	json.Unmarshal(body, &index)

	for _, packager := range index.Packages {
		for _, tool := range packager.Tools {
			res = append(res, &tools.Tool{
				Packager: packager.Name,
				Name:     tool.Name,
				Version:  tool.Version,
			})
		}
	}

	return res, nil
}

// Installed crawles the Tools Folder and finds the installed tools.
func (t *Tools) Installed(ctx context.Context) (tools.ToolCollection, error) {
	res := tools.ToolCollection{}

	// Find packagers
	packagers, err := os.ReadDir(t.folder)
	if err != nil {
		if !strings.Contains(err.Error(), "no such file") {
			return nil, err
		}
		err = os.MkdirAll(t.folder, 0755)
		if err != nil {
			return nil, err
		}
	}

	for _, packager := range packagers {
		if !packager.IsDir() {
			continue
		}

		// Find tools
		toolss, err := os.ReadDir(filepath.Join(t.folder, packager.Name()))
		if err != nil {
			return nil, err
		}

		for _, tool := range toolss {
			// Find versions
			path := filepath.Join(t.folder, packager.Name(), tool.Name())
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
func (t *Tools) Install(ctx context.Context, payload *tools.ToolPayload) (*tools.Operation, error) {
	path := filepath.Join(payload.Packager, payload.Name, payload.Version)

	//if URL is defined and is signed we verify the signature and override the name, payload, version parameters
	if payload.URL != nil && payload.Signature != nil && payload.Checksum != nil {
		err := utilities.VerifyInput(*payload.URL, *payload.Signature)
		if err != nil {
			return nil, err
		}
		return t.install(ctx, path, *payload.URL, *payload.Checksum)
	}

	// otherwise we install from the default index
	body, err := t.index.Read()
	if err != nil {
		return nil, err
	}

	var index Index
	json.Unmarshal(body, &index)

	for _, packager := range index.Packages {
		if packager.Name != payload.Packager {
			continue
		}

		for _, tool := range packager.Tools {
			if tool.Name == payload.Name &&
				tool.Version == payload.Version {

				sys := tool.GetFlavourCompatibleWith(runtime.GOOS, runtime.GOARCH)

				return t.install(ctx, path, sys.URL, sys.Checksum)
			}
		}
	}

	return nil, tools.MakeNotFound(
		fmt.Errorf("tool not found with packager '%s', name '%s', version '%s'",
			payload.Packager, payload.Name, payload.Version))
}

func (t *Tools) install(ctx context.Context, path, url, checksum string) (*tools.Operation, error) {
	// Download
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Use a teereader to only read once
	var buffer bytes.Buffer
	reader := io.TeeReader(res.Body, &buffer)

	safePath, err := utilities.SafeJoin(t.folder, path)
	if err != nil {
		return nil, err
	}

	// Cleanup
	err = os.RemoveAll(safePath)
	if err != nil {
		return nil, err
	}

	err = extract.Archive(ctx, reader, t.folder, rename(path))
	if err != nil {
		os.RemoveAll(safePath)
		return nil, err
	}

	sum := sha256.Sum256(buffer.Bytes())
	sumString := "SHA-256:" + hex.EncodeToString(sum[:sha256.Size])

	if sumString != checksum {
		os.RemoveAll(safePath)
		return nil, errors.New("checksum doesn't match")
	}

	// Write installed.json for retrocompatibility with v1
	err = writeInstalled(t.folder, path)
	if err != nil {
		return nil, err
	}

	return &tools.Operation{Status: "ok"}, nil
}

// Remove deletes the tool folder from Tools Folder
func (t *Tools) Remove(ctx context.Context, payload *tools.ToolPayload) (*tools.Operation, error) {
	path := filepath.Join(payload.Packager, payload.Name, payload.Version)
	pathToRemove, err := utilities.SafeJoin(t.folder, path)
	if err != nil {
		return nil, err
	}

	err = os.RemoveAll(pathToRemove)
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

	installedFile, err := utilities.SafeJoin(folder, "installed.json")
	if err != nil {
		return err
	}
	data, err := os.ReadFile(installedFile)
	if err == nil {
		err = json.Unmarshal(data, &installed)
		if err != nil {
			return err
		}
	}

	parts := strings.Split(path, string(filepath.Separator))
	tool := parts[len(parts)-2]
	toolWithVersion := fmt.Sprint(tool, "-", parts[len(parts)-1])
	toolFile, err := utilities.SafeJoin(folder, path)
	if err != nil {
		return err
	}
	installed[tool] = toolFile
	installed[toolWithVersion] = toolFile

	data, err = json.Marshal(installed)
	if err != nil {
		return err
	}

	return os.WriteFile(installedFile, data, 0644)
}
