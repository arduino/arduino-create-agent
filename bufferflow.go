package main

type Bufferflow interface {
	Init()
	BlockUntilReady(cmd string, id string) (bool, bool) // implement this method
	OnIncomingData(data string)                         // implement this method
	Close()                                             // implement this method
}
