package main

import (
	"fmt"
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
		//h.broadcast <- []byte("dummy data")
		time.Sleep(8000 * time.Millisecond)
		h.broadcast <- []byte("list")

		// open com4 (tinyg)
		h.broadcast <- []byte("open com4 115200 tinyg")
		time.Sleep(1000 * time.Millisecond)

		// send some commands
		//h.broadcast <- []byte("send com4 ?\n")
		//time.Sleep(3000 * time.Millisecond)
		h.broadcast <- []byte("send com4 {\"qr\":\"\"}\n")
		h.broadcast <- []byte("send com4 g21 g90\n") // mm
		//h.broadcast <- []byte("send com4 {\"qr\":\"\"}\n")
		//h.broadcast <- []byte("send com4 {\"sv\":0}\n")
		//time.Sleep(3000 * time.Millisecond)
		for i := 0.0; i < 10.0; i = i + 0.001 {
			h.broadcast <- []byte("send com4 G1 X" + fmt.Sprintf("%.3f", i) + " F100\n")
			time.Sleep(10 * time.Millisecond)
		}
		/*
			h.broadcast <- []byte("send com4 G1 X1\n")
			h.broadcast <- []byte("send com4 G1 X2\n")
			h.broadcast <- []byte("send com4 G1 X3\n")
			h.broadcast <- []byte("send com4 G1 X4\n")
			h.broadcast <- []byte("send com4 G1 X5\n")
			h.broadcast <- []byte("send com4 G1 X6\n")
			h.broadcast <- []byte("send com4 G1 X7\n")
			h.broadcast <- []byte("send com4 G1 X8\n")
			h.broadcast <- []byte("send com4 G1 X9\n")
			h.broadcast <- []byte("send com4 G1 X10\n")
			h.broadcast <- []byte("send com4 G1 X1\n")
			h.broadcast <- []byte("send com4 G1 X2\n")
			h.broadcast <- []byte("send com4 G1 X3\n")
			h.broadcast <- []byte("send com4 G1 X4\n")
			h.broadcast <- []byte("send com4 G1 X5\n")
			h.broadcast <- []byte("send com4 G1 X6\n")
			h.broadcast <- []byte("send com4 G1 X7\n")
			h.broadcast <- []byte("send com4 G1 X8\n")
			h.broadcast <- []byte("send com4 G1 X9\n")
			h.broadcast <- []byte("send com4 G1 X10\n")
			h.broadcast <- []byte("send com4 G1 X1\n")
			h.broadcast <- []byte("send com4 G1 X2\n")
			h.broadcast <- []byte("send com4 G1 X3\n")
			h.broadcast <- []byte("send com4 G1 X4\n")
			h.broadcast <- []byte("send com4 G1 X5\n")
			h.broadcast <- []byte("send com4 G1 X6\n")
			h.broadcast <- []byte("send com4 G1 X7\n")
			h.broadcast <- []byte("send com4 G1 X8\n")
			h.broadcast <- []byte("send com4 G1 X9\n")
			h.broadcast <- []byte("send com4 G1 X10\n")
		*/
		break
	}
	log.Println("dummy process exited")
}
