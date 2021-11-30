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
	"errors"
	"net/http"

	"github.com/arduino/arduino-create-agent/killbrowser"
	"github.com/gin-gonic/gin"
)

func killBrowserHandler(c *gin.Context) {

	var data struct {
		Action  string `json:"action"`
		Process string `json:"process"`
		URL     string `json:"url"`
	}

	c.BindJSON(&data)

	if data.Process != "chrome" && data.Process != "chrom" {
		c.JSON(http.StatusBadRequest, errors.New("You can't kill the process"+data.Process))
		return
	}

	command, err := browser.Find(data.Process)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if data.Action == "kill" || data.Action == "restart" {
		_, err := browser.Kill(data.Process)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	}

	if data.Action == "restart" {
		_, err := browser.Start(command, data.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	}

}
