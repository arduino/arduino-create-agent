package main

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type BufferflowTimed struct {
	Name   string
	Port   string
	Output chan []byte
	Input  chan string
	done   chan bool
	ticker *time.Ticker
}

var (
	bufferedOutput string
	sPort string
)

func (b *BufferflowTimed) Init() {
	log.Println("Initting timed buffer flow (output once every 16ms)")
	bufferedOutput = ""
	sPort = ""

	go func() {
		b.ticker = time.NewTicker(16 * time.Millisecond)
		b.done = make(chan bool)
	Loop:
		for {
			select {
			case data := <-b.Input:
				bufferedOutput = bufferedOutput + data
				sPort = b.Port
			case <-b.ticker.C:
				if bufferedOutput != "" {
					m := SpPortMessage{sPort, bufferedOutput}
					buf, _ := json.Marshal(m)
					// data is now encoded in base64 format
					// need a decoder on the other side
					b.Output <- []byte(buf)
					bufferedOutput = ""
					sPort = ""
				}
			case <-b.done:
				break Loop
			}
		}

		close(b.Input)

	}()

}

func (b *BufferflowTimed) BlockUntilReady(cmd string, id string) (bool, bool) {
	//log.Printf("BlockUntilReady() start\n")
	return true, false
}

func (b *BufferflowTimed) OnIncomingData(data string) {
	b.Input <- data
}

// Clean out b.sem so it can truly block
func (b *BufferflowTimed) ClearOutSemaphore() {
}

func (b *BufferflowTimed) BreakApartCommands(cmd string) []string {
	return []string{cmd}
}

func (b *BufferflowTimed) Pause() {
	return
}

func (b *BufferflowTimed) Unpause() {
	return
}

func (b *BufferflowTimed) SeeIfSpecificCommandsShouldSkipBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimed) SeeIfSpecificCommandsShouldPauseBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimed) SeeIfSpecificCommandsShouldUnpauseBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimed) SeeIfSpecificCommandsShouldWipeBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimed) SeeIfSpecificCommandsReturnNoResponse(cmd string) bool {
	return false
}

func (b *BufferflowTimed) ReleaseLock() {
}

func (b *BufferflowTimed) IsBufferGloballySendingBackIncomingData() bool {
	return true
}

func (b *BufferflowTimed) Close() {
	b.ticker.Stop()
	b.done <- true
	close(b.done)
}
