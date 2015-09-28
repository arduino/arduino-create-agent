package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"time"
)

type BufferflowTimed struct {
	Name   string
	Port   string
	Output chan []byte
	Input  chan string
}

var (
	bufferedOutput string
)

func (b *BufferflowTimed) Init() {
	log.Println("Initting timed buffer flow (output once every 16ms)")
	bufferedOutput = ""

	go func() {
		for data := range b.Input {
			bufferedOutput = bufferedOutput + data
		}
	}()

	go func() {
		c := time.Tick(16 * time.Millisecond)
		log.Println(bufferedOutput)
		for _ = range c {
			m := SpPortMessage{bufferedOutput}
			buf, _ := json.Marshal(m)
			b.Output <- []byte(buf)
			bufferedOutput = ""
		}
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
}
