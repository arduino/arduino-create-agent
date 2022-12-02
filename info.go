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

package main

import (
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"go.bug.st/serial"
)

func infoHandler(c *gin.Context) {
	host := c.Request.Host
	parts := strings.Split(host, ":")
	host = parts[0]

	c.JSON(200, gin.H{
		"version":    version,
		"http":       "http://" + host + port,
		"https":      "https://localhost" + portSSL,
		"ws":         "ws://" + host + port,
		"wss":        "wss://localhost" + portSSL,
		"origins":    origins,
		"update_url": updateURL,
		"os":         runtime.GOOS + ":" + runtime.GOARCH,
	})
}

func pauseHandler(c *gin.Context) {
	go func() {
		ports, _ := serial.GetPortsList()
		for _, element := range ports {
			spClose(element)
		}
		*hibernate = true
		Systray.Pause()
	}()
	c.JSON(200, nil)
}
