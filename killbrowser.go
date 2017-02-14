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
