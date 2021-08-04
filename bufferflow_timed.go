package main

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type BufferflowTimed struct {
	port           string
	output         chan []byte
	input          chan string
	done           chan bool
	ticker         *time.Ticker
	sPort          string
	bufferedOutput string
}

func NewBufferflowTimed(port string, output chan []byte) *BufferflowTimed {
	return &BufferflowTimed{
		port:           port,
		output:         output,
		input:          make(chan string),
		done:           make(chan bool),
		ticker:         time.NewTicker(16 * time.Millisecond),
		sPort:          "",
		bufferedOutput: "",
	}
}

func (b *BufferflowTimed) Init() {
	log.Println("Initting timed buffer flow (output once every 16ms)")
	go b.consumeInput()
}

func (b *BufferflowTimed) consumeInput() {
Loop:
	for {
		select {
		case data := <-b.input: // use the buffer and append data to it
			b.bufferedOutput = b.bufferedOutput + data
			b.sPort = b.port
		case <-b.ticker.C: // after 16ms send the buffered output message
			if b.bufferedOutput != "" {
				m := SpPortMessage{b.sPort, b.bufferedOutput}
				buf, _ := json.Marshal(m)
				// data is now encoded in base64 format
				// need a decoder on the other side
				b.output <- buf
				// reset the buffer and the port
				b.bufferedOutput = ""
				b.sPort = ""
			}
		case <-b.done:
			break Loop //this is required, a simple break statement would only exit the innermost switch statement
		}
	}
	close(b.input)
}

func (b *BufferflowTimed) BlockUntilReady(cmd string, id string) (bool, bool) {
	//log.Printf("BlockUntilReady() start\n")
	return true, false
}

func (b *BufferflowTimed) OnIncomingData(data string) {
	b.input <- data
}

func (b *BufferflowTimed) IsBufferGloballySendingBackIncomingData() bool {
	return true
}

func (b *BufferflowTimed) Close() {
	b.ticker.Stop()
	b.done <- true
	close(b.done)
}
