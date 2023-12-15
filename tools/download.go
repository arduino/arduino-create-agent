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

package tools

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
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/arduino/arduino-create-agent/v2/pkgs"
	"github.com/arduino/go-paths-helper"
	"github.com/blang/semver"
	"github.com/codeclysm/extract/v3"
)

// public vars to allow override in the tests
var (
	OS   = runtime.GOOS
	Arch = runtime.GOARCH
)

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// Download will parse the index at the indexURL for the tool to download.
// It will extract it in a folder in .arduino-create, and it will update the
// Installed map.
//
// pack contains the packager of the tool
// name contains the name of the tool.
// version contains the version of the tool.
// behaviour contains the strategy to use when there is already a tool installed
//
// If version is "latest" it will always download the latest version (regardless
// of the value of behaviour)
//
// If version is not "latest" and behaviour is "replace", it will download the
// version again. If instead behaviour is "keep" it will not download the version
// if it already exists.
func (t *Tools) Download(pack, name, version, behaviour string) error {

	body, err := t.index.Read()
	if err != nil {
		return err
	}

	var data pkgs.Index
	json.Unmarshal(body, &data)

	// Find the tool by name
	correctTool, correctSystem := findTool(pack, name, version, data)

	if correctTool.Name == "" || correctSystem.URL == "" {
		t.logger("We couldn't find a tool with the name " + name + " and version " + version + " packaged by " + pack)
		return nil
	}

	key := correctTool.Name + "-" + correctTool.Version

	// Check if it already exists
	if behaviour == "keep" {
		location, ok := t.getMapValue(key)
		if ok && pathExists(location) {
			// overwrite the default tool with this one
			t.setMapValue(correctTool.Name, location)
			t.logger("The tool is already present on the system")
			return t.writeMap()
		}
	}

	// Download the tool
	t.logger("Downloading tool " + name + " from " + correctSystem.URL)
	resp, err := http.Get(correctSystem.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the body
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Checksum
	checksum := sha256.Sum256(body)
	checkSumString := "SHA-256:" + hex.EncodeToString(checksum[:sha256.Size])

	if checkSumString != correctSystem.Checksum {
		return errors.New("checksum doesn't match")
	}

	tempPath := paths.TempDir()
	// Create a temporary dir to extract package
	if err := tempPath.MkdirAll(); err != nil {
		return fmt.Errorf("creating temp dir for extraction: %s", err)
	}
	tempDir, err := tempPath.MkTempDir("package-")
	if err != nil {
		return fmt.Errorf("creating temp dir for extraction: %s", err)
	}
	defer tempDir.RemoveAll()

	t.logger("Unpacking tool " + name)
	ctx := context.Background()
	reader := bytes.NewReader(body)
	// Extract into temp directory
	if err := extract.Archive(ctx, reader, tempDir.String(), nil); err != nil {
		return fmt.Errorf("extracting archive: %s", err)
	}

	location := t.directory.Join(pack, correctTool.Name, correctTool.Version)
	err = location.RemoveAll()
	if err != nil {
		return err
	}

	// Check package content and find package root dir
	root, err := findPackageRoot(tempDir)
	if err != nil {
		return fmt.Errorf("searching package root dir: %s", err)
	}

	if err := root.Rename(location); err != nil {
		if err := root.CopyDirTo(location); err != nil {
			return fmt.Errorf("moving extracted archive to destination dir: %s", err)
		}
	}

	// if the tool contains a post_install script, run it: it means it is a tool that needs to install drivers
	// AFAIK this is only the case for the windows-driver tool
	err = t.installDrivers(location.String())
	if err != nil {
		return err
	}

	// Ensure that the files are executable
	t.logger("Ensure that the files are executable")

	// Update the tool map
	t.logger("Updating map with location " + location.String())

	t.setMapValue(name, location.String())
	t.setMapValue(name+"-"+correctTool.Version, location.String())
	return t.writeMap()
}

func findPackageRoot(parent *paths.Path) (*paths.Path, error) {
	files, err := parent.ReadDir()
	if err != nil {
		return nil, fmt.Errorf("reading package root dir: %s", err)
	}
	files.FilterOutPrefix("__MACOSX")

	// if there is only one dir, it is the root dir
	if len(files) == 1 && files[0].IsDir() {
		return files[0], nil
	}
	return parent, nil
}

func findTool(pack, name, version string, data pkgs.Index) (pkgs.Tool, pkgs.System) {
	var correctTool pkgs.Tool
	correctTool.Version = "0.0"

	for _, p := range data.Packages {
		if p.Name != pack {
			continue
		}
		for _, t := range p.Tools {
			if version != "latest" {
				if t.Name == name && t.Version == version {
					correctTool = t
				}
			} else {
				// Find latest
				v1, _ := semver.Make(t.Version)
				v2, _ := semver.Make(correctTool.Version)
				if t.Name == name && v1.Compare(v2) > 0 {
					correctTool = t
				}
			}
		}
	}

	// Find the url based on system
	correctSystem := correctTool.GetFlavourCompatibleWith(OS, Arch)

	return correctTool, correctSystem
}

func (t *Tools) installDrivers(location string) error {
	OkPressed := 6
	extension := ".bat"
	preamble := ""
	if OS != "windows" {
		extension = ".sh"
		// add ./ to force locality
		preamble = "./"
	}
	if _, err := os.Stat(filepath.Join(location, "post_install"+extension)); err == nil {
		t.logger("Installing drivers")
		ok := MessageBox("Installing drivers", "We are about to install some drivers needed to use Arduino/Genuino boards\nDo you want to continue?")
		if ok == OkPressed {
			os.Chdir(location)
			t.logger(preamble + "post_install" + extension)
			oscmd := exec.Command(preamble + "post_install" + extension)
			if OS != "linux" {
				// spawning a shell could be the only way to let the user type his password
				TellCommandNotToSpawnShell(oscmd)
			}
			err = oscmd.Run()
			return err
		}
		return errors.New("could not install drivers")
	}
	return nil
}
