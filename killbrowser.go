package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func killBrowserHandler(c *gin.Context) {

	var data struct {
		Action  string `json:"action"`
		Process string `json:"process"`
		URL     string `json:"url"`
	}

	c.BindJSON(&data)

	command, err := findBrowser(data.Process)

	log.Println(command)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}

	if data.Action == "kill" || data.Action == "restart" {
		// _, err := killBrowser(data.Process)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, err.Error())
		// }
	}

	if data.Action == "restart" {
		_, err := startBrowser(command, data.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
		}
	}

}
