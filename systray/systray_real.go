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

//go:build !cli

// Systray_real gets compiled when the tag 'cli' is missing. This is the default case

package systray

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/arduino/arduino-create-agent/icon"
	"github.com/arduino/go-paths-helper"
	"github.com/getlantern/systray"
	"github.com/go-ini/ini"
	"github.com/skratchdot/open-golang/open"
)

// Start sets up the systray icon with its menus
func (s *Systray) Start() {
	if s.Hibernate {
		systray.Run(s.startHibernate, s.end)
	} else {
		systray.Run(s.start, s.end)
	}
}

// Quit simply exits the program
func (s *Systray) Quit() {
	systray.Quit()
}

// start creates a systray icon with menu options to go to arduino create, open debug, pause/quit the agent
func (s *Systray) start() {
	systray.SetIcon(icon.GetIcon())

	// Add version
	menuVer := systray.AddMenuItem("Agent version "+s.Version, "")
	menuVer.Disable()

	// Add links
	mURL := systray.AddMenuItem("Go to Arduino Create", "Arduino Create")
	mDebug := systray.AddMenuItem("Open Debug Console", "Debug console")

	// Remove crash-reports
	mRmCrashes := systray.AddMenuItem("Remove crash reports", "")
	s.updateMenuItem(mRmCrashes, s.CrashesIsEmpty())

	// Add pause/quit
	mPause := systray.AddMenuItem("Pause Agent", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Agent", "")

	// Add configs
	s.addConfigs()

	// listen for events
	go func() {
		for {
			select {
			case <-mURL.ClickedCh:
				_ = open.Start("https://create.arduino.cc")
			case <-mDebug.ClickedCh:
				_ = open.Start(s.DebugURL())
			case <-mRmCrashes.ClickedCh:
				s.RemoveCrashes()
				s.updateMenuItem(mRmCrashes, s.CrashesIsEmpty())
			case <-mPause.ClickedCh:
				s.Pause()
			case <-mQuit.ClickedCh:
				s.Quit()
			}
		}
	}()
}

// updateMenuItem will enable or disable an item in the tray icon menu id disable is true
func (s *Systray) updateMenuItem(item *systray.MenuItem, disable bool) {
	if disable {
		item.Disable()
	} else {
		item.Enable()
	}
}

// CrashesIsEmpty checks if the folder containing crash-reports is empty
func (s *Systray) CrashesIsEmpty() bool {
	logsDir := getLogsDir()
	return logsDir.NotExist() // if the logs directory is empty we assume there are no crashreports
}

// RemoveCrashes removes the crash-reports from `logs` folder
func (s *Systray) RemoveCrashes() {
	logsDir := getLogsDir()
	pathErr := logsDir.RemoveAll()
	if pathErr != nil {
		log.Errorf("Cannot remove crashreports: %s", pathErr)
	} else {
		log.Infof("Removed crashreports inside: %s", logsDir)
	}
}

// getLogsDir simply returns the folder containing the logs
func getLogsDir() *paths.Path {
	usr, _ := user.Current()
	usrDir := paths.New(usr.HomeDir) // The user folder, on linux/macos /home/<usr>/
	agentDir := usrDir.Join(".arduino-create")
	return agentDir.Join("logs")
}

// starthibernate creates a systray icon with menu options to resume/quit the agent
func (s *Systray) startHibernate() {
	systray.SetIcon(icon.GetIconHiber())

	mResume := systray.AddMenuItem("Resume Agent", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Agent", "")

	// listen for events
	go func() {
		for {
			select {
			case <-mResume.ClickedCh:
				s.Resume()
			case <-mQuit.ClickedCh:
				s.Quit()
			}
		}
	}()
}

// end simply exits the program
func (s *Systray) end() {
	os.Exit(0)
}

func (s *Systray) addConfigs() {
	var mConfigCheckbox []*systray.MenuItem

	configs := getConfigs()
	if len(configs) > 1 {
		for _, config := range configs {
			entry := systray.AddMenuItem(config.Name, "")
			mConfigCheckbox = append(mConfigCheckbox, entry)
			// decorate configs
			gliph := " ☐ "
			if s.AdditionalConfig == config.Location {
				gliph = " 🗹 "
			}
			entry.SetTitle(gliph + config.Name)
		}
	}

	// It would be great to use the select channel here,
	// but unfortunately there's no clean way to do it with an array of channels, so we start a single goroutine for each of them
	for i := range mConfigCheckbox {
		go func(v int) {
			<-mConfigCheckbox[v].ClickedCh
			s.AdditionalConfig = configs[v].Location
			s.Restart()
		}(i)
	}
}

type configIni struct {
	Name     string
	Location string
}

// getconfigs parses all config files in the executable folder
func getConfigs() []configIni {
	// config.ini must be there, so call it Default
	src, _ := os.Executable() // TODO change path
	dest := filepath.Dir(src)

	var configs []configIni

	err := filepath.Walk(dest, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			if filepath.Ext(path) == ".ini" {
				cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true, AllowPythonMultilineValues: true}, filepath.Join(dest, f.Name()))
				if err != nil {
					return err
				}
				defaultSection, err := cfg.GetSection("")
				name := defaultSection.Key("name").String()
				if name == "" || err != nil {
					name = "Default config"
				}
				conf := configIni{Name: name, Location: f.Name()}
				configs = append(configs, conf)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("error walking through executable configuration: %w", err)
	}

	return configs
}
