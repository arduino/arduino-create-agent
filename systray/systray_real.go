// +build !cli

// Systray_real gets compiled when the tag 'cli' is missing. This is the default case
package systray

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/arduino/arduino-create-agent/icon"
	"github.com/getlantern/systray"
	"github.com/go-ini/ini"
	"github.com/kardianos/osext"
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
	mUrl := systray.AddMenuItem("Go to Arduino Create", "Arduino Create")
	mDebug := systray.AddMenuItem("Open Debug Console", "Debug console")

	// Remove crash-reports
	mRmCrashes := systray.AddMenuItem("Remove crash reports", "")
	s.updateMenuItem(mRmCrashes, s.CrashesIsEmpty())

	// Add pause/quit
	mPause := systray.AddMenuItem("Pause Plugin", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Plugin", "")

	// Add configs
	s.addConfigs()

	// listen for events
	go func() {
		for {
			select {
			case <-mUrl.ClickedCh:
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
	currDir, err := osext.ExecutableFolder()
	if err != nil {
		log.Error("Cannot determine executable path: ", err)
	}
	logsDir := filepath.Join(currDir, "logs")
	if _, err := os.Stat(string(logsDir)); os.IsNotExist(err) {
		return true
	}
	return false
}

// RemoveCrashes removes the crash-reports from `logs` folder
func (s *Systray) RemoveCrashes() {
	currDir, err := osext.ExecutableFolder()
	if err != nil {
		log.Error("Cannot determine executable path: ", err)
	}
	logsDir := filepath.Join(currDir, "logs")
	pathErr := os.RemoveAll(logsDir)
	if pathErr != nil {
		log.Error("Cannot remove crashreports: ", pathErr)
	} else {
		log.Info("Removed crashreports inside: ", logsDir)
	}
}

// starthibernate creates a systray icon with menu options to resume/quit the agent
func (s *Systray) startHibernate() {
	systray.SetIcon(icon.GetIconHiber())

	mResume := systray.AddMenuItem("Resume Plugin", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Plugin", "")

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
	src, _ := osext.Executable()
	dest := filepath.Dir(src)

	var configs []configIni

	err := filepath.Walk(dest, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			if filepath.Ext(path) == ".ini" {
				cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, filepath.Join(dest, f.Name()))
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
