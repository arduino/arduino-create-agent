package main

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type BufferflowTimedBinary struct {
	Name   string
	Port   string
	Output chan []byte
	Input  chan []byte
	done   chan bool
	ticker *time.Ticker
}

var (
	bufferedOutputBinary []byte
	sPortBinary          string
)

func (b *BufferflowTimedBinary) Init() {
	log.Println("Initting timed buffer binary flow (output once every 16ms)")
	bufferedOutputBinary = nil
	sPortBinary = ""

	go func() {
		b.ticker = time.NewTicker(16 * time.Millisecond)
		b.done = make(chan bool)
	Loop:
		for {
			select {
			case data := <-b.Input:
				bufferedOutputBinary = append(bufferedOutputBinary, data...)
				sPortBinary = b.Port
			case <-b.ticker.C:
				if bufferedOutputBinary != nil {
					m := SpPortMessageRaw{sPortBinary, bufferedOutputBinary}
					buf, _ := json.Marshal(m)
					b.Output <- buf
					bufferedOutputBinary = nil
					sPortBinary = ""
				}
			case <-b.done:
				break Loop
			}
		}

		close(b.Input)
	}()
}

func (b *BufferflowTimedBinary) BlockUntilReady(cmd string, id string) (bool, bool) {
	//log.Printf("BlockUntilReady() start\n")
	return true, false
}

func (b *BufferflowTimedBinary) OnIncomingDataBinary(data []byte) {
	b.Input <- data
}

// not implemented, we are gonna use OnIncomingDataBinary
func (b *BufferflowTimedBinary) OnIncomingData(data string) {
}

// Clean out b.sem so it can truly block
func (b *BufferflowTimedBinary) ClearOutSemaphore() {
}

func (b *BufferflowTimedBinary) BreakApartCommands(cmd string) []string {
	return []string{cmd}
}

func (b *BufferflowTimedBinary) Pause() {
	return
}

func (b *BufferflowTimedBinary) Unpause() {
	return
}

func (b *BufferflowTimedBinary) SeeIfSpecificCommandsShouldSkipBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimedBinary) SeeIfSpecificCommandsShouldPauseBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimedBinary) SeeIfSpecificCommandsShouldUnpauseBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimedBinary) SeeIfSpecificCommandsShouldWipeBuffer(cmd string) bool {
	return false
}

func (b *BufferflowTimedBinary) SeeIfSpecificCommandsReturnNoResponse(cmd string) bool {
	return false
}

func (b *BufferflowTimedBinary) ReleaseLock() {
}

func (b *BufferflowTimedBinary) IsBufferGloballySendingBackIncomingData() bool {
	return true
}

func (b *BufferflowTimedBinary) Close() {
	b.ticker.Stop()
	close(b.Input)
}
