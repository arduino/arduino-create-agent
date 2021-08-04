package main

import (
	log "github.com/sirupsen/logrus"
)

type BufferflowDefault struct {
	port string
}

func NewBufferflowDefault(port string) *BufferflowDefault {
	return &BufferflowDefault{
		port: port,
	}
}

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

func (b *BufferflowDefault) IsBufferGloballySendingBackIncomingData() bool {
	return false
}

func (b *BufferflowDefault) Close() {
}
