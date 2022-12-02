// Copyright 2022 Arduino SA
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// BufferflowDefault is the default bufferflow, whick means no buffering
type BufferflowDefault struct {
	port   string
	output chan<- []byte
	input  chan string
	done   chan bool
}

// NewBufferflowDefault create a new default bufferflow
func NewBufferflowDefault(port string, output chan<- []byte) *BufferflowDefault {
	return &BufferflowDefault{
		port:   port,
		output: output,
		input:  make(chan string),
		done:   make(chan bool),
	}
}

// Init will initialize the bufferflow
func (b *BufferflowDefault) Init() {
	log.Println("Initting default buffer flow (which means no buffering)")
	go b.consumeInput()
}

func (b *BufferflowDefault) consumeInput() {
Loop:
	for {
		select {
		case data := <-b.input:
			m := SpPortMessage{b.port, data}
			message, _ := json.Marshal(m)
			b.output <- message
		case <-b.done:
			break Loop //this is required, a simple break statement would only exit the innermost switch statement
		}
	}
	close(b.input) // close the input channel at the end of the computation
}

// OnIncomingData will forward the data
func (b *BufferflowDefault) OnIncomingData(data string) {
	b.input <- data
}

// Close will close the bufferflow
func (b *BufferflowDefault) Close() {
	b.done <- true
	close(b.done)
}
