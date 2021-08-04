package main

// availableBufferAlgorithms = {"default", "timed", "timedraw", "timedbinary"}

type BufferMsg struct {
	Cmd                string
	Port               string
	TriggeringResponse string
	//Desc string
	//Desc string
}

type Bufferflow interface {
	Init()
	BlockUntilReady(cmd string, id string) (bool, bool) // implement this method
	OnIncomingData(data string)                         // implement this method
	Close()                                             // implement this method
}
