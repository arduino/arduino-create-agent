//
//  update.go
//
//  Created by Martino Facchin
//  Code from github.com/sanbornm and github.com/sanderhahn
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

package main

import (
	"github.com/arduino/arduino-create-agent/updater"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/osext"
)

func updateHandler(c *gin.Context) {

	path, err := osext.Executable()

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var up = &updater.Updater{
		CurrentVersion: version,
		APIURL:         *updateUrl,
		BinURL:         *updateUrl,
		DiffURL:        "",
		Dir:            "update/",
		CmdName:        *appName,
	}

	err = up.BackgroundRun()

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": "Please wait a moment while the agent reboots itself"})
	go restart(path)
}
