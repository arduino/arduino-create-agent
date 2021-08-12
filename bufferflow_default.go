package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type BufferflowDefault struct {
	port   string
	output chan<- []byte
	input  chan string
	done   chan bool
}

func NewBufferflowDefault(port string, output chan<- []byte) *BufferflowDefault {
	return &BufferflowDefault{
		port:   port,
		output: output,
		input:  make(chan string),
		done:   make(chan bool),
	}
}

func (b *BufferflowDefault) Init() {
	log.Println("Initting default buffer flow (which means no buffering)")
	go b.consumeInput()
}

func (b *BufferflowDefault) consumeInput() {
Loop:
	for {
		select {
		case data := <-b.input:
			m := SpPortMessage{b.port, data}
			message, _ := json.Marshal(m)
			b.output <- message
		case <-b.done:
			break Loop //this is required, a simple break statement would only exit the innermost switch statement
		}
	}
	close(b.input) // close the input channel at the end of the computation
}

func (b *BufferflowDefault) BlockUntilReady(cmd string, id string) (bool, bool) {
	//log.Printf("BlockUntilReady() start\n")
	return true, false
}

func (b *BufferflowDefault) OnIncomingData(data string) {
	b.input <- data
}

func (b *BufferflowDefault) Close() {
	b.done <- true
	close(b.done)
}
