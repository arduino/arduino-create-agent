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
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/arduino/arduino-create-agent/utilities"
	"github.com/arduino/arduino-create-agent/v2/pkgs"
)

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
//
// At the moment the value of behaviour is ignored.
func (t *Tools) Download(pack, name, version, behaviour string) error {

	tool := pkgs.New(t.index, t.directory.String())
	_, err := tool.Install(context.Background(), &tools.ToolPayload{Name: name, Version: version, Packager: pack})
	if err != nil {
		return err
	}

	path := filepath.Join(pack, name, version)
	safePath, err := utilities.SafeJoin(t.directory.String(), path)
	if err != nil {
		return err
	}

	// if the tool contains a post_install script, run it: it means it is a tool that needs to install drivers
	// AFAIK this is only the case for the windows-driver tool
	err = t.installDrivers(safePath)
	if err != nil {
		return err
	}

	// Ensure that the files are executable
	t.logger("Ensure that the files are executable")

	// Update the tool map
	t.logger("Updating map with location " + safePath)

	t.setMapValue(name, safePath)
	t.setMapValue(name+"-"+version, safePath)

	return nil
}

func (t *Tools) installDrivers(location string) error {
	OkPressed := 6
	extension := ".bat"
	// add .\ to force locality
	preamble := ".\\"
	if runtime.GOOS != "windows" {
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
			if runtime.GOOS != "linux" {
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
