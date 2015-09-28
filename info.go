package main

import (
	"github.com/gin-gonic/gin"
)

func infoHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"http":  "http://localhost" + port,
		"https": "https://localhost" + port,
		"ws":    "ws://localhost" + port,
		"wss":   "wss://localhost" + portSSL,
	})
}
