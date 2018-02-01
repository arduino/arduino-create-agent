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
	"flag"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/arduino/arduino-create-agent/icon"
	"github.com/getlantern/systray"
	"github.com/go-ini/ini"
	"github.com/kardianos/osext"
	"github.com/skratchdot/open-golang/open"
	"github.com/vharitonsky/iniflags"
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
				file := filepath.Join(dest, f.Name())
				cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, file)
				if err != nil {
					log.Printf("Error loading file %v: %v", file, err)
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

func applyEnvironment(filename string) {
	src, _ := osext.Executable()
	dest := filepath.Dir(src)
	cfg, _ := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, filepath.Join(dest, filename))
	defaultSection, err := cfg.GetSection("env")
	if err != nil {
		return
	}
	for _, env := range defaultSection.KeyStrings() {
		val := defaultSection.Key(env).String()
		log.Info("Applying env setting: " + env + "=" + val)
		os.Setenv(env, val)
	}
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
			if *configIni == config.Localtion {
				gliph = " 🗹 "
			}
			entry.SetTitle(gliph + config.Name)
		}
	} else {
		// apply env setting from first config immediately
		// applyEnvironment(configs[0].Localtion)
	}
	//mQuit := systray.AddMenuItem("Quit Plugin", "")

	menuVer.Disable()

	for i, _ := range mConfigCheckbox {
		go func(v int) {
			for {
				<-mConfigCheckbox[v].ClickedCh
				flag.Set("config", configs[v].Localtion)
				iniflags.UpdateConfig()
				applyEnvironment(configs[v].Localtion)
				mConfigCheckbox[v].SetTitle(" 🗹 " + configs[v].Name)
				//mConfigCheckbox[v].Check()
				for j, _ := range mConfigCheckbox {
					if j != v {
						mConfigCheckbox[j].SetTitle(" ☐ " + configs[j].Name)
					}
				}
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
		log.Println("Restart becayse setup went wrong?")
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
			open.Start("http://localhost" + port)
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
		log.Println("Restart for hubernation")
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
