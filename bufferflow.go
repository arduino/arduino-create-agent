package main

import (
//"log"
//"time"
)

var availableBufferAlgorithms = []string{"default", "tinyg", "dummypause", "grbl"}

type BufferMsg struct {
	Cmd                string
	Port               string
	TriggeringResponse string
	//Desc string
	//Desc string
}

type Bufferflow interface {
	BlockUntilReady(cmd string) bool                          // implement this method
	OnIncomingData(data string)                               // implement this method
	BreakApartCommands(cmd string) []string                   // implement this method
	Pause()                                                   // implement this method
	Unpause()                                                 // implement this method
	SeeIfSpecificCommandsShouldSkipBuffer(cmd string) bool    // implement this method
	SeeIfSpecificCommandsShouldPauseBuffer(cmd string) bool   // implement this method
	SeeIfSpecificCommandsShouldUnpauseBuffer(cmd string) bool // implement this method
	SeeIfSpecificCommandsShouldWipeBuffer(cmd string) bool    // implement this method
	ReleaseLock()                                             // implement this method
	IsBufferGloballySendingBackIncomingData() bool            // implement this method
	Close()													  // implement this method
}
