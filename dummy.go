package main

import (
	"log"
	"time"
)

type dummy struct {
	//myvar mytype string
}

var d = dummy{
//myvar: make(mytype string),
}

func (d *dummy) run() {
	for {
		//h.broadcast <- message
		log.Print("dummy data")
		h.broadcast <- []byte("dummy data")
		time.Sleep(15000 * time.Millisecond)
	}
}
