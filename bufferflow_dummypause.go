package main

import (
	"log"
	"time"
)

type BufferflowDummypause struct {
	Name     string
	Port     string
	NumLines int
	Paused   bool
}

func (b *BufferflowDummypause) Init() {
}

func (b *BufferflowDummypause) BlockUntilReady() bool {
	log.Printf("BlockUntilReady() start. numLines:%v\n", b.NumLines)
	log.Printf("buffer:%v\n", b)
	//for b.Paused {
	log.Println("We are paused for 3 seconds. Yeilding send.")
	time.Sleep(3000 * time.Millisecond)
	//}
	log.Printf("BlockUntilReady() end\n")
	return true
}

func (b *BufferflowDummypause) OnIncomingData(data string) {
	log.Printf("OnIncomingData() start. data:%v\n", data)
	b.NumLines++
	//time.Sleep(3000 * time.Millisecond)
	log.Printf("OnIncomingData() end. numLines:%v\n", b.NumLines)
}

func (b *BufferflowDummypause) BreakApartCommands(cmd string) []string {
	return []string{cmd}
}

func (b *BufferflowDummypause) Pause() {
	return
}

func (b *BufferflowDummypause) Unpause() {
	return
}

func (b *BufferflowDummypause) SeeIfSpecificCommandsShouldSkipBuffer(cmd string) bool {
	return false
}

func (b *BufferflowDummypause) SeeIfSpecificCommandsShouldPauseBuffer(cmd string) bool {
	return false
}

func (b *BufferflowDummypause) SeeIfSpecificCommandsShouldUnpauseBuffer(cmd string) bool {
	return false
}

func (b *BufferflowDummypause) SeeIfSpecificCommandsShouldWipeBuffer(cmd string) bool {
	return false
}

func (b *BufferflowDummypause) ReleaseLock() {
}

func (b *BufferflowDummypause) IsBufferGloballySendingBackIncomingData() bool {
	return false
}
