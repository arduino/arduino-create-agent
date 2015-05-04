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
	"github.com/facchinm/trayhost"
	"github.com/kardianos/osext"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

var notificationThumbnail trayhost.Image

func setupSysTray() {

	menuItems := []trayhost.MenuItem{
		trayhost.MenuItem{
			Title: "Launch webide.arduino.cc",
			Handler: func() {
				open.Run("http://webide.arduino.cc:8080")
			},
		},
		trayhost.SeparatorMenuItem(),
		trayhost.MenuItem{
			Title: "Quit",
			Handler: func() {
				trayhost.Exit()
				exit()
			},
		},
	}

	runtime.LockOSThread()

	execPath, _ := osext.Executable()
	b, err := ioutil.ReadFile(filepath.Dir(execPath) + "/arduino/resources/icons/icon.png")
	if err != nil {
		panic(err)
	}

	trayhost.Initialize("WebIDEBridge", b, menuItems)
	trayhost.EnterLoop()

	// systray.SetIcon(IconData)
	// systray.SetTitle("Arduino WebIDE Bridge")

	// // We can manipulate the systray in other goroutines
	// go func() {
	// 	systray.SetIcon(IconData)
	// 	mUrl := systray.AddMenuItem("Open webide.arduino.cc", "WebIDE Home")
	// 	mQuit := systray.AddMenuItem("Quit", "Quit the bridge")
	// 	for {
	// 		select {
	// 		case <-mUrl.ClickedCh:
	// 			open.Run("http://webide.arduino.cc:8080")
	// 		case <-mQuit.ClickedCh:
	// 			systray.Quit()
	// 			fmt.Println("Quit now...")
	// 			exit()
	// 		}
	// 	}
	// }()
}
