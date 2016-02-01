package main

import (
	"github.com/gin-gonic/gin"
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
