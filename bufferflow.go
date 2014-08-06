package main

import (
//"log"
//"time"
)

type BufferMsg struct {
	Cmd                string
	Port               string
	TriggeringResponse string
	//Desc string
	//Desc string
}

type Bufferflow interface {
	BlockUntilReady() bool                                    // implement this method
	OnIncomingData(data string)                               // implement this method
	Pause()                                                   // implement this method
	Unpause()                                                 // implement this method
	SeeIfSpecificCommandsShouldSkipBuffer(cmd string) bool    // implement this method
	SeeIfSpecificCommandsShouldPauseBuffer(cmd string) bool   // implement this method
	SeeIfSpecificCommandsShouldUnpauseBuffer(cmd string) bool // implement this method
	SeeIfSpecificCommandsShouldWipeBuffer(cmd string) bool    // implement this method
	ReleaseLock()                                             // implement this method
	//Name string
	//Port string
	//myvar mytype string
	//pause      bool   // keep track if we're paused from sending
	//buffertype string // is it tinyg, grbl, or other?
}

/*
// this method is a method of the struct above
func (b *bufferflow) blockUntilReady() {
	log.Printf("Blocking until ready. Buffertype is:%v\n", b.buffertype)
	//time.Sleep(3000 * time.Millisecond)
	if b.buffertype == "dummypause" {
		buf := bufferflow_dummypause{Name: "blah"}
		buf.blockUntilReady()
	}
	log.Printf("Done blocking. Buffertype is:%v\n", b.buffertype)
}

func (b *bufferflow) onIncomingData(data) {
	log.Printf("onIncomingData. data:%v", data)
	if b.buffertype == "dummypause" {
		buf := bufferflow_dummypause{Name: "blah"}
		buf.waitUntilReady()
	}
}
*/
