//
//  trayicon.go
//
//  Created by Martino Facchin
//  Copyright (c) 2015 Arduino LLC
//
//  Permission is hereby granted, free of charge, to any person
//  obtaining a copy of this software and associated documentation
//  files (the "Software"), to deal in the Software without
//  restriction, including without limitation the rights to use,
//  copy, modify, merge, publish, distribute, sublicense, and/or sell
//  copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following
//  conditions:
//
//  The above copyright notice and this permission notice shall be
//  included in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
//  OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
//  HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
//  FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
//

// +build !cli

package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/arduino/arduino-create-agent/icon"
	"github.com/facchinm/systray"
	"github.com/go-ini/ini"
	"github.com/kardianos/osext"
	log "github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
	"go.bug.st/serial.v1"
)

func setupSysTray() {
	runtime.LockOSThread()
	if *hibernate == true {
		systray.Run(setupSysTrayHibernate)
	} else {
		systray.Run(setupSysTrayReal)
	}
}

func addRebootTrayElement() {
	reboot_tray := systray.AddMenuItem("Reboot to update", "")

	go func() {
		<-reboot_tray.ClickedCh
		systray.Quit()
		log.Println("Restarting now...")
		log.Println("Restart because addReebotTrayElement")
		restart("")
	}()
}

type ConfigIni struct {
	Name      string
	Localtion string
}

func getConfigs() []ConfigIni {
	// parse all configs in executable folder
	// config.ini must be there, so call it Default
	src, _ := osext.Executable()
	dest := filepath.Dir(src)

	var configs []ConfigIni

	filepath.Walk(dest, func(path string, f os.FileInfo, _ error) error {
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
				conf := ConfigIni{Name: name, Localtion: f.Name()}
				configs = append(configs, conf)
			}
		}
		return nil
	})
	return configs
}

func setupSysTrayReal() {

	systray.SetIcon(icon.GetIcon())
	mUrl := systray.AddMenuItem("Go to Arduino Create", "Arduino Create")
	mDebug := systray.AddMenuItem("Open debug console", "Debug console")
	menuVer := systray.AddMenuItem("Agent version "+version+"-"+git_revision, "")
	mPause := systray.AddMenuItem("Pause Plugin", "")
	var mConfigCheckbox []*systray.MenuItem

	configs := getConfigs()

	if len(configs) > 1 {
		for _, config := range configs {
			entry := systray.AddMenuItem(config.Name, "")
			mConfigCheckbox = append(mConfigCheckbox, entry)
			// decorate configs
			gliph := " ☐ "
			if *additionalConfig == config.Localtion {
				gliph = " 🗹 "
			}
			entry.SetTitle(gliph + config.Name)
		}
	}
	//mQuit := systray.AddMenuItem("Quit Plugin", "")

	menuVer.Disable()

	for i, _ := range mConfigCheckbox {
		go func(v int) {
			for {
				<-mConfigCheckbox[v].ClickedCh

				restart("", "-additional-config", configs[v].Localtion)
			}
		}(i)
	}

	go func() {
		<-mPause.ClickedCh
		ports, _ := serial.GetPortsList()
		for _, element := range ports {
			spClose(element)
		}
		systray.Quit()
		*hibernate = true
		log.Println("Restart because setup went wrong?")
		restart("")
	}()

	// go func() {
	// 	<-mQuit.ClickedCh
	// 	systray.Quit()
	// 	exit()
	// }()

	go func() {
		for {
			<-mDebug.ClickedCh
			logAction("log on")
			open.Start("http://127.0.0.1" + port)
		}
	}()

	// We can manipulate the systray in other goroutines
	go func() {
		for {
			<-mUrl.ClickedCh
			open.Start("http://create.arduino.cc")
		}
	}()
}

func setupSysTrayHibernate() {

	systray.SetIcon(icon.GetIconHiber())
	mOpen := systray.AddMenuItem("Open Plugin", "")
	mQuit := systray.AddMenuItem("Kill Plugin", "")

	go func() {
		<-mOpen.ClickedCh
		*hibernate = false
		log.Println("Restart for hibernation")
		systray.Quit()
		restart("")
	}()

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
		exit()
	}()
}

func quitSysTray() {
	systray.Quit()
}
