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
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/arduino/arduino-create-agent/config"
	"github.com/arduino/arduino-create-agent/icon"
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
	mConfig := systray.AddMenuItem("Open Configuration", "Config File")

	// Remove crash-reports
	mRmCrashes := systray.AddMenuItem("Remove crash reports", "")
	s.updateMenuItem(mRmCrashes, config.LogsIsEmpty())

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
			case <-mConfig.ClickedCh:
				_ = open.Start(s.currentConfigFilePath.String())
			case <-mRmCrashes.ClickedCh:
				s.RemoveCrashes()
				s.updateMenuItem(mRmCrashes, config.LogsIsEmpty())
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

// RemoveCrashes removes the crash-reports from `logs` folder
func (s *Systray) RemoveCrashes() {
	logsDir := config.GetLogsDir()
	pathErr := logsDir.RemoveAll()
	if pathErr != nil {
		log.Errorf("Cannot remove crashreports: %s", pathErr)
	} else {
		log.Infof("Removed crashreports inside: %s", logsDir)
	}
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

	configs := s.getConfigs()
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

// getConfigs parses all config files in the .arduino-create folder
func (s *Systray) getConfigs() []configIni {
	var configs []configIni

	files, err := s.ConfigDir.ReadDir()
	if err != nil {
		log.Errorf("cannot read the content of %s", s.ConfigDir)
		return nil
	}
	files.FilterOutDirs()
	files.FilterSuffix(".ini")
	for _, file := range files {
		cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true, AllowPythonMultilineValues: true}, file.String())
		if err != nil {
			log.Errorf("error walking through executable configuration: %s", err)
		} else {
			defaultSection, err := cfg.GetSection("")
			name := defaultSection.Key("name").String()
			if name == "" || err != nil {
				name = "Default config"
			}
			conf := configIni{Name: name, Location: file.String()}
			configs = append(configs, conf)
		}
	}

	return configs
}
