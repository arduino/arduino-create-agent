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

package config

import (
	// we need this for the ArduinoCreateAgent.plist in this package
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/arduino/go-paths-helper"
	log "github.com/sirupsen/logrus"
)

//go:embed ArduinoCreateAgent.plist
var launchdAgentDefinition []byte

// getLaunchdAgentPath will return the path of the launchd agent default path
func getLaunchdAgentPath() *paths.Path {
	return GetDefaultHomeDir().Join("Library", "LaunchAgents", "ArduinoCreateAgent.plist")
}

// WritePlistFile function will write the required plist file to $HOME/Library/LaunchAgents/ArduinoCreateAgent.plist
// it will return nil in case of success,
// it will error if the file is already there or in any other case
func WritePlistFile() error {

	launchdAgentPath := getLaunchdAgentPath()
	if launchdAgentPath.Exist() {
		// we already have an existing launchd plist file, so we don't have to do anything
		return fmt.Errorf("the autostart file %s already exists", launchdAgentPath)
	}

	src, err := os.Executable()
	if err != nil {
		return err
	}
	data := struct {
		Program   string
		RunAtLoad bool
	}{
		Program:   src,
		RunAtLoad: false,
	}

	t := template.Must(template.New("launchdConfig").Parse(string(launchdAgentDefinition)))

	// we need to create a new launchd plist file
	plistFile, _ := launchdAgentPath.Create()
	return t.Execute(plistFile, data)
}

// LoadLaunchdAgent will use launchctl to load the agent, will return an error if something goes wrong
func LoadLaunchdAgent() error {
	// https://www.launchd.info/
	oscmd := exec.Command("launchctl", "load", getLaunchdAgentPath().String())
	err := oscmd.Run()
	return err
}

// UnloadLaunchdAgent will use launchctl to load the agent, will return an error if something goes wrong
func UnloadLaunchdAgent() error {
	// https://www.launchd.info/
	oscmd := exec.Command("launchctl", "unload", getLaunchdAgentPath().String())
	err := oscmd.Run()
	return err
}

// RemovePlistFile function will remove the plist file from $HOME/Library/LaunchAgents/ArduinoCreateAgent.plist and return an error
// it will not do anything if the file is not there
func RemovePlistFile() error {
	launchdAgentPath := getLaunchdAgentPath()
	if launchdAgentPath.Exist() {
		log.Infof("removing: %s", launchdAgentPath)
		return launchdAgentPath.Remove()
	}
	return nil
}
