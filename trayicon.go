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

// +build !arm

package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/facchinm/go-serial"
	"github.com/facchinm/systray"
	"github.com/facchinm/systray/example/icon"
	"github.com/skratchdot/open-golang/open"
	"runtime"
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
		restart("")
	}()
}

func setupSysTrayReal() {

	systray.SetIcon(icon.Data)
	mUrl := systray.AddMenuItem("Go to Arduino Create (staging)", "Arduino Create")
	mDebug := systray.AddMenuItem("Open debug console", "Debug console")
	menuVer := systray.AddMenuItem("Agent version "+version+"-"+git_revision, "")
	mPause := systray.AddMenuItem("Pause Plugin", "")
	//mQuit := systray.AddMenuItem("Quit Plugin", "")

	menuVer.Disable()

	go func() {
		<-mPause.ClickedCh
		ports, _ := serial.GetPortsList()
		for _, element := range ports {
			spClose(element)
		}
		systray.Quit()
		*hibernate = true
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
			open.Start("http://create-staging.arduino.cc")
		}
	}()
}

func setupSysTrayHibernate() {

	systray.SetIcon(icon.DataHibernate)
	mOpen := systray.AddMenuItem("Open Plugin", "")
	mQuit := systray.AddMenuItem("Kill Plugin", "")

	go func() {
		<-mOpen.ClickedCh
		*hibernate = false
		restart("")
	}()

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
		exit()
	}()
}
