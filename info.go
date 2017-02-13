package main

import (
	"github.com/gin-gonic/gin"
	"go.bug.st/serial.v1"
)

func infoHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"version": version,
		"http":    "http://localhost" + port,
		"https":   "https://localhost" + portSSL,
		"ws":      "ws://localhost" + port,
		"wss":     "wss://localhost" + portSSL,
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
