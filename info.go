package main

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.bug.st/serial.v1"
)

func infoHandler(c *gin.Context) {
	host := c.Request.Host
	parts := strings.Split(host, ":")
	host = parts[0]

	c.JSON(200, gin.H{
		"version": version,
		"http":    "http://" + host + port,
		"https":   "https://" + host + portSSL,
		"ws":      "ws://" + host + port,
		"wss":     "wss://" + host + portSSL,
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
