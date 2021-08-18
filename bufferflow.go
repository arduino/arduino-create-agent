package main

type Bufferflow interface {
	Init()
	OnIncomingData(data string) // implement this method
	Close()                     // implement this method
}
