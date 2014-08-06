package main

import (
	"log"
	//"regexp"
	//"strconv"
	//"time"
)

type BufferflowDefault struct {
	Name string
	Port string
}

var ()

func (b *BufferflowDefault) Init() {
	log.Println("Initting default buffer flow (which means no buffering)")
}

func (b *BufferflowDefault) BlockUntilReady() bool {
	//log.Printf("BlockUntilReady() start\n")
	return true
}

func (b *BufferflowDefault) OnIncomingData(data string) {
	//log.Printf("OnIncomingData() start. data:%v\n", data)
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

func (b *BufferflowDefault) ReleaseLock() {
}
