package main

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type BufferflowTimedRaw struct {
	port              string
	output            chan<- []byte
	input             chan string
	done              chan bool
	ticker            *time.Ticker
	bufferedOutputRaw []byte
	sPortRaw          string
}

func NewBufferflowTimedRaw(port string, output chan<- []byte) *BufferflowTimedRaw {
	return &BufferflowTimedRaw{
		port:              port,
		output:            output,
		input:             make(chan string),
		done:              make(chan bool),
		ticker:            time.NewTicker(16 * time.Millisecond),
		bufferedOutputRaw: nil,
		sPortRaw:          "",
	}
}

func (b *BufferflowTimedRaw) Init() {
	log.Println("Initting timed buffer raw flow (output once every 16ms)")
	go b.consumeInput()
}

func (b *BufferflowTimedRaw) consumeInput() {
Loop:
	for {
		select {
		case data := <-b.input: // use the buffer and append data to it
			b.bufferedOutputRaw = append(b.bufferedOutputRaw, []byte(data)...)
			b.sPortRaw = b.port
		case <-b.ticker.C: // after 16ms send the buffered output message
			if b.bufferedOutputRaw != nil {
				m := SpPortMessageRaw{b.sPortRaw, b.bufferedOutputRaw}
				buf, _ := json.Marshal(m)
				// data is now encoded in base64 format
				// need a decoder on the other side
				b.output <- buf
				// reset the buffer and the port
				b.bufferedOutputRaw = nil
				b.sPortRaw = ""
			}
		case <-b.done:
			break Loop //this is required, a simple break statement would only exit the innermost switch statement
		}
	}
	close(b.input)
}

func (b *BufferflowTimedRaw) BlockUntilReady(cmd string, id string) (bool, bool) {
	//log.Printf("BlockUntilReady() start\n")
	return true, false
}

func (b *BufferflowTimedRaw) OnIncomingData(data string) {
	b.input <- data
}

func (b *BufferflowTimedRaw) Close() {
	b.ticker.Stop()
	b.done <- true
	close(b.done)
}
