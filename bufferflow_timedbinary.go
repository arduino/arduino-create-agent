package main

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type BufferflowTimedBinary struct {
	port                 string
	output               chan []byte
	input                chan []byte
	done                 chan bool
	ticker               *time.Ticker
	bufferedOutputBinary []byte
	sPortBinary          string
}

func NewBufferflowTimedBinary(port string, output chan []byte) *BufferflowTimedBinary {
	return &BufferflowTimedBinary{
		port:                 port,
		output:               output,
		input:                make(chan []byte),
		done:                 make(chan bool),
		ticker:               time.NewTicker(16 * time.Millisecond),
		bufferedOutputBinary: nil,
		sPortBinary:          "",
	}
}

func (b *BufferflowTimedBinary) Init() {
	log.Println("Initting timed buffer binary flow (output once every 16ms)")
	go b.consumeInput()
}

func (b *BufferflowTimedBinary) consumeInput() {
Loop:
	for {
		select {
		case data := <-b.input: // use the buffer and append data to it
			b.bufferedOutputBinary = append(b.bufferedOutputBinary, data...)
			b.sPortBinary = b.port
		case <-b.ticker.C: // after 16ms send the buffered output message
			if b.bufferedOutputBinary != nil {
				m := SpPortMessageRaw{b.sPortBinary, b.bufferedOutputBinary}
				buf, _ := json.Marshal(m)
				// data is now encoded in base64 format
				// need a decoder on the other side
				b.output <- buf
				// reset the buffer and the port
				b.bufferedOutputBinary = nil
				b.sPortBinary = ""
			}
		case <-b.done:
			break Loop //this is required, a simple break statement would only exit the innermost switch statement
		}
	}
	close(b.input)
}

func (b *BufferflowTimedBinary) BlockUntilReady(cmd string, id string) (bool, bool) {
	//log.Printf("BlockUntilReady() start\n")
	return true, false
}

func (b *BufferflowTimedBinary) OnIncomingData(data string) {
	b.input <- []byte(data)
}

func (b *BufferflowTimedBinary) IsBufferGloballySendingBackIncomingData() bool {
	return true
}

func (b *BufferflowTimedBinary) Close() {
	b.ticker.Stop()
	b.done <- true
	close(b.input)
}
