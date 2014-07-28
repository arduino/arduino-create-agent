package main

import (
//"log"
//"time"
)

type Bufferflow interface {
	BlockUntilReady()           // implement this method
	OnIncomingData(data string) // implement this method
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
