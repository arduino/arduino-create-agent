package main

import (
//"log"
//"time"
)

var availableBufferAlgorithms = []string{"default", "timed", "timedraw"}

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
	//JustQueue(cmd string, id string) bool                     // implement this method
	OnIncomingData(data string)                               // implement this method
	ClearOutSemaphore()                                       // implement this method
	BreakApartCommands(cmd string) []string                   // implement this method
	Pause()                                                   // implement this method
	Unpause()                                                 // implement this method
	SeeIfSpecificCommandsShouldSkipBuffer(cmd string) bool    // implement this method
	SeeIfSpecificCommandsShouldPauseBuffer(cmd string) bool   // implement this method
	SeeIfSpecificCommandsShouldUnpauseBuffer(cmd string) bool // implement this method
	SeeIfSpecificCommandsShouldWipeBuffer(cmd string) bool    // implement this method
	SeeIfSpecificCommandsReturnNoResponse(cmd string) bool    // implement this method
	ReleaseLock()                                             // implement this method
	IsBufferGloballySendingBackIncomingData() bool            // implement this method
	Close()                                                   // implement this method
}

/*data packets returned to client*/
type DataCmdComplete struct {
	Cmd     string
	Id      string
	P       string
	BufSize int    `json:"-"`
	D       string `json:"-"`
}

type DataPerLine struct {
	P string
	D string
}
