package main

import (
	"log"

	"github.com/facchinm/go-serial"
	"github.com/facchinm/systray"
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

func pauseHandler(c *gin.Context) {
	go func() {
		ports, _ := serial.GetPortsList()
		for _, element := range ports {
			spClose(element)
		}
		systray.Quit()
		*hibernate = true
		log.Println("Restart becayse setup went wrong?")
		restart("")
	}()
	c.JSON(200, nil)
}
