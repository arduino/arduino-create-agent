package main

import (
	"github.com/gin-gonic/gin"
	"go.bug.st/serial.v1"
)

func infoHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"version": version,
		"http":    "http://" + c.Request.Host + ":" + port,
		"https":   "https://" + c.Request.Host + ":" + portSSL,
		"ws":      "ws://" + c.Request.Host + ":" + port,
		"wss":     "wss://" + c.Request.Host + ":" + portSSL,
	})
}

func pauseHandler(c *gin.Context) {
	go func() {
		ports, _ := serial.GetPortsList()
		for _, element := range ports {
			spClose(element)
		}
		*hibernate = true
		restart("")
	}()
	c.JSON(200, nil)
}
