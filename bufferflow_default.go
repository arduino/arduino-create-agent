package main

import (
	log "github.com/sirupsen/logrus"
)

type BufferflowDefault struct {
	Name string
	Port string
}

var ()

func (b *BufferflowDefault) Init() {
	log.Println("Initting default buffer flow (which means no buffering)")
}

func (b *BufferflowDefault) BlockUntilReady(cmd string, id string) (bool, bool) {
	//log.Printf("BlockUntilReady() start\n")
	return true, false
}

func (b *BufferflowDefault) OnIncomingData(data string) {
	//log.Printf("OnIncomingData() start. data:%v\n", data)
}

// Clean out b.sem so it can truly block
func (b *BufferflowDefault) ClearOutSemaphore() {
}

func (b *BufferflowDefault) BreakApartCommands(cmd string) []string {
	return []string{cmd}
}

func (b *BufferflowDefault) Pause() {
	return
}

func (b *BufferflowDefault) Unpause() {
	return
}

func (b *BufferflowDefault) SeeIfSpecificCommandsShouldSkipBuffer(cmd string) bool {
	return false
}

func (b *BufferflowDefault) SeeIfSpecificCommandsShouldPauseBuffer(cmd string) bool {
	return false
}

func (b *BufferflowDefault) SeeIfSpecificCommandsShouldUnpauseBuffer(cmd string) bool {
	return false
}

func (b *BufferflowDefault) SeeIfSpecificCommandsShouldWipeBuffer(cmd string) bool {
	return false
}

func (b *BufferflowDefault) SeeIfSpecificCommandsReturnNoResponse(cmd string) bool {
	return false
}

func (b *BufferflowDefault) ReleaseLock() {
}

func (b *BufferflowDefault) IsBufferGloballySendingBackIncomingData() bool {
	return false
}

func (b *BufferflowDefault) Close() {
}
