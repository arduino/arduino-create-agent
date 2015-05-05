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
	"fmt"
	"github.com/facchinm/systray"
	"github.com/facchinm/systray/example/icon"
	"github.com/skratchdot/open-golang/open"
)

func setupSysTray() {
	systray.Run(setupSysTrayReal)
}

func setupSysTrayReal() {

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetIcon(icon.Data)
		mUrl := systray.AddMenuItem("Open webide.arduino.cc", "Arduino Create Home")
		mQuit := systray.AddMenuItem("Quit", "Quit the bridge")
		for {
			select {
			case <-mUrl.ClickedCh:
				open.Run("http://webide.arduino.cc:8080")
			case <-mQuit.ClickedCh:
				systray.Quit()
				fmt.Println("Quit now...")
				exit()
			}
		}
	}()
}
