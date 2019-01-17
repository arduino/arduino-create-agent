package main

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type BufferflowTimedRaw struct {
	Name   string
	Port   string
	Output chan []byte
	Input  chan string
	ticker *time.Ticker
}

var (
	bufferedOutputRaw []byte
)

func (b *BufferflowTimedRaw) Init() {
	log.Println("Initting timed buffer flow (output once every 16ms)")

	go func() {
		for data := range b.Input {
			bufferedOutputRaw = append(bufferedOutputRaw, []byte(data)...)
		}
	}()

	go func() {
		b.ticker = time.NewTicker(16 * time.Millisecond)
		for _ = range b.ticker.C {
			if len(bufferedOutputRaw) != 0 {
				m := SpPortMessageRaw{bufferedOutputRaw}
				buf, _ := json.Marshal(m)
				// data is now encoded in base64 format
				// need a decoder on the other side
				b.Output <- []byte(buf)
				bufferedOutputRaw = nil
			}
		}
	}()

}

func (b *BufferflowTimedRaw) BlockUntilReady(cmd string, id string) (bool, bool) {
	//log.Printf("BlockUntilReady() start\n")
	return true, false
}

func (b *BufferflowTimedRaw) OnIncomingData(data string) {
	b.Input <- data
}

// Clean out b.sem so it can truly block
func (b *BufferflowTimedRaw) ClearOutSemaphore() {
}

func (b *BufferflowTimedRaw) BreakApartCommands(cmd string) []string {
	return []string{cmd}
}

func (b *BufferflowTimedRaw) Pause() {
	return
}

func (b *BufferflowTimedRaw) Unpause() {
	return
}

func (b *BufferflowTimedRaw) SeeIfSpecificCommandsShouldSkipBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimedRaw) SeeIfSpecificCommandsShouldPauseBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimedRaw) SeeIfSpecificCommandsShouldUnpauseBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimedRaw) SeeIfSpecificCommandsShouldWipeBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimedRaw) SeeIfSpecificCommandsReturnNoResponse(cmd string) bool {
	return false
}

func (b *BufferflowTimedRaw) ReleaseLock() {
}

func (b *BufferflowTimedRaw) IsBufferGloballySendingBackIncomingData() bool {
	return true
}

func (b *BufferflowTimedRaw) Close() {
	b.ticker.Stop()
	close(b.Input)
}
