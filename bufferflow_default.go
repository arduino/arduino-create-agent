package main

import (
	"log"
	//"regexp"
	//"strconv"
	//"time"
)

type BufferflowDefault struct {
}

var ()

func (b *BufferflowDefault) Init() {
	log.Println("Initting default buffer flow (which means no buffering)")
}

func (b *BufferflowDefault) BlockUntilReady() {
	//log.Printf("BlockUntilReady() start\n")
}

func (b *BufferflowDefault) OnIncomingData(data string) {
	//log.Printf("OnIncomingData() start. data:%v\n", data)
}
