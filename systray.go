/*
 * This file is part of arduino-create-agent.
 *
 * arduino-create-agent is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */
package agent

import (
	"os"
	"runtime"

	"github.com/arduino/arduino-create-agent/icon"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

func setupSystray(hibernate bool, version, revision, address string, restart, shutdown func()) {
	runtime.LockOSThread()
	if !hibernate {
		systray.Run(setupSystrayReal(version, revision, address, restart), nil)
	} else {
		systray.Run(setupSysTrayHibernate(restart, shutdown), nil)
	}
}

func setupSystrayReal(version, revision, address string, restart func()) func() {
	return func() {
		systray.SetIcon(icon.GetIcon())
		mURL := systray.AddMenuItem("Go to Arduino Create", "Arduino Create")
		mDebug := systray.AddMenuItem("Open debug console", "Debug console")
		menuVer := systray.AddMenuItem("Agent version "+version+"-"+revision, "")
		mPause := systray.AddMenuItem("Pause Plugin", "")

		menuVer.Disable()

		// Listen for events
		go func() {
			for {
				select {
				case <-mPause.ClickedCh:
					systray.Quit()
					restart()
				case <-mDebug.ClickedCh:
					open.Start(address + "/debug")
				case <-mURL.ClickedCh:
					open.Start("https://create.arduino.cc")
				}
			}
		}()
	}
}

func setupSysTrayHibernate(restart, shutdown func()) func() {
	return func() {
		systray.SetIcon(icon.GetIconHiber())
		mOpen := systray.AddMenuItem("Open Plugin", "")
		mQuit := systray.AddMenuItem("Kill Plugin", "")

		// Listen for events
		go func() {
			for {
				select {
				case <-mOpen.ClickedCh:
					systray.Quit()
					restart()
				case <-mQuit.ClickedCh:
					systray.Quit()
					os.Exit(0)
				}
			}
		}()
	}
}
