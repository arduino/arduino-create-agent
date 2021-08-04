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

	go func() {
	Loop:
		for {
			select {
			case data := <-b.input:
				b.bufferedOutput = b.bufferedOutput + data
				b.sPort = b.port
			case <-b.ticker.C:
				if b.bufferedOutput != "" {
					m := SpPortMessage{b.sPort, b.bufferedOutput}
					buf, _ := json.Marshal(m)
					// data is now encoded in base64 format
					// need a decoder on the other side
					b.output <- []byte(buf)
					b.bufferedOutput = ""
					b.sPort = ""
				}
			case <-b.done:
				break Loop
			}
		}

		close(b.input)

	}()

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
