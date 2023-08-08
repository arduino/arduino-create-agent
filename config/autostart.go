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
	"bytes"
	// we need this for the ArduinoCreateAgent.plist in this package
	_ "embed"
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
	homeDir := GetDefaultHomeDir()
	launchAgentsPath := homeDir.Join("Library", "LaunchAgents")
	agentPlistPath := launchAgentsPath.Join("ArduinoCreateAgent.plist")

	if err := os.MkdirAll(launchAgentsPath.String(), 0755); err != nil {
		log.Panicf("Could not create ~/Library/LaunchAgents directory: %s", err)
	}

	return agentPlistPath
}

// InstallPlistFile will handle the process of creating the plist file required for the autostart
// and loading it using launchd
func InstallPlistFile() {
	launchdAgentPath := getLaunchdAgentPath()
	if !launchdAgentPath.Exist() {
		writeAndLoadPlistFile(launchdAgentPath)
		log.Info("Quitting, another instance of the agent has been started by launchd")
		os.Exit(0)
	} else {
		// we already have an existing launchd plist file, so we check if it's updated
		launchAgentContent, _ := launchdAgentPath.ReadFile()
		launchAgentContentNew, _ := getLaunchdAgentDefinition()
		if bytes.Equal(launchAgentContent, launchAgentContentNew) {
			log.Infof("the autostart file %s already exists: nothing to do", launchdAgentPath)
		} else {
			log.Infof("the autostart file %s needs to be updated", launchdAgentPath)
			removePlistFile()
			writeAndLoadPlistFile(launchdAgentPath)
		}

	}
}

// writeAndLoadPlistFile function will write the plist file, load it, and then exit, because launchd will start a new instance.
func writeAndLoadPlistFile(launchdAgentPath *paths.Path) {
	err := writePlistFile(launchdAgentPath)
	if err != nil {
		log.Error(err)
	} else {
		err = loadLaunchdAgent() // this will load the agent: basically starting a new instance
		if err != nil {
			log.Error(err)
		}
	}
}

// writePlistFile function will write the required plist file to launchdAgentPath
// it will return nil in case of success,
// it will error in any other case
func writePlistFile(launchdAgentPath *paths.Path) error {
	definition, err := getLaunchdAgentDefinition()
	if err != nil {
		return err
	}
	// we need to create a new launchd plist file
	return launchdAgentPath.WriteFile(definition)
}

// getLaunchdAgentDefinition will return the definition of the new LaunchdAgent
func getLaunchdAgentDefinition() ([]byte, error) {
	src, err := os.Executable()

	if err != nil {
		return nil, err
	}
	data := struct {
		Program   string
		RunAtLoad bool
	}{
		Program:   src,
		RunAtLoad: true, // This will start the agent right after login (and also after `launchctl load ...`)
	}

	t := template.Must(template.New("launchdConfig").Parse(string(launchdAgentDefinition)))

	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// loadLaunchdAgent will use launchctl to load the agent, will return an error if something goes wrong
func loadLaunchdAgent() error {
	// https://www.launchd.info/
	oscmd := exec.Command("launchctl", "load", getLaunchdAgentPath().String())
	err := oscmd.Run()
	return err
}

// UninstallPlistFile will handle the process of unloading (unsing launchd) the file required for the autostart
// and removing the file
func UninstallPlistFile() {
	err := unloadLaunchdAgent()
	if err != nil {
		log.Error(err)
	} else {
		err = removePlistFile()
		if err != nil {
			log.Error(err)
		}
	}
}

// unloadLaunchdAgent will use launchctl to load the agent, will return an error if something goes wrong
func unloadLaunchdAgent() error {
	// https://www.launchd.info/
	oscmd := exec.Command("launchctl", "unload", getLaunchdAgentPath().String())
	err := oscmd.Run()
	return err
}

// removePlistFile function will remove the plist file from $HOME/Library/LaunchAgents/ArduinoCreateAgent.plist and return an error
// it will not do anything if the file is not there
func removePlistFile() error {
	launchdAgentPath := getLaunchdAgentPath()
	if launchdAgentPath.Exist() {
		log.Infof("removing: %s", launchdAgentPath)
		return launchdAgentPath.Remove()
	}
	log.Infof("the autostart file %s do not exists: nothing to do", launchdAgentPath)
	return nil
}
