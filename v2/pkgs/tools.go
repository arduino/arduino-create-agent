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
func (c *Tools) Install(ctx context.Context, payload *tools.ToolPayload) (*tools.Operation, error) {
	path := filepath.Join(payload.Packager, payload.Name, payload.Version)

	if payload.URL != nil {
		return c.install(ctx, path, *payload.URL, *payload.Checksum)
	}

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

					i := findSystem(tool)

					return c.install(ctx, path, tool.Systems[i].URL, tool.Systems[i].Checksum)
				}
			}
		}
	}

	return nil, tools.MakeNotFound(
		fmt.Errorf("tool not found with packager '%s', name '%s', version '%s'",
			payload.Packager, payload.Name, payload.Version))
}

func (c *Tools) install(ctx context.Context, path, url, checksum string) (*tools.Operation, error) {
	// Download
	res, err := http.Get(url)
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

	if sumString != checksum {
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

func writeInstalled(folder, path string) error {
	// read installed.json
	installed := map[string]string{}

	data, err := ioutil.ReadFile(filepath.Join(folder, "installed.json"))
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

	return ioutil.WriteFile(filepath.Join(folder, "installed.json"), data, 0644)
}
