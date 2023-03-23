// Copyright 2023 Arduino SA
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
	_ "embed"
	"os"

	"github.com/arduino/go-paths-helper"
	log "github.com/sirupsen/logrus"
)

// GetCertificatesDir return the directory where SSL certificates are saved
func GetCertificatesDir() *paths.Path {
	return GetDataDir()
}

// GetDataDir returns the full path to the default Arduino Create Agent data directory.
func GetDataDir() *paths.Path {
	userDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("Could not get user dir: %s", err)
	}
	dataDir := paths.New(userDir, ".arduino-create")
	if err := dataDir.MkdirAll(); err != nil {
		log.Panicf("Could not create data dir: %s", err)
	}
	return dataDir
}

// GetLogsDir return the directory where logs are saved
func GetLogsDir() *paths.Path {
	logsDir := GetDataDir().Join("logs")
	if err := logsDir.MkdirAll(); err != nil {
		log.Panicf("Can't create logs dir: %s", err)
	}
	return logsDir
}

// GetDefaultConfigDir returns the full path to the default Arduino Create Agent configuration directory.
func GetDefaultConfigDir() *paths.Path {
	// UserConfigDir returns the default root directory to use
	// for user-specific configuration data. Users should create
	// their own application-specific subdirectory within this
	// one and use that.
	//
	// On Unix systems, it returns $XDG_CONFIG_HOME as specified by
	// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
	// if non-empty, else $HOME/.config.
	//
	// On Darwin, it returns $HOME/Library/Application Support.
	// On Windows, it returns %AppData%.
	// On Plan 9, it returns $home/lib.
	//
	// If the location cannot be determined (for example, $HOME
	// is not defined), then it will return an error.
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Panicf("Can't get user home dir: %s", err)
	}

	agentConfigDir := paths.New(configDir, "ArduinoCreateAgent")
	if err := agentConfigDir.MkdirAll(); err != nil {
		log.Panicf("Can't create config dir: %s", err)
	}
	return agentConfigDir
}

// https://github.com/golang/go/issues/46056
//
//go:generate cp -r ../config.ini config.ini
//go:embed config.ini
var configContent []byte

// GenerateConfig function will take a directory path as an input
// and will write the default config,ini file to that directory,
// it will panic if something goes wrong
func GenerateConfig(destDir *paths.Path) *paths.Path {
	configPath := destDir.Join("config.ini")

	// generate the config.ini file directly in destDir
	if err := configPath.WriteFile(configContent); err != nil {
		// if we do not have a config there's nothing else we can do
		panic("cannot generate config: " + err.Error())
	}
	log.Infof("generated config in %s", configPath)
	return configPath
}
