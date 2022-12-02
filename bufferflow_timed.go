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
	"time"

	log "github.com/sirupsen/logrus"
)

// BufferflowTimed sends data once every 16ms
type BufferflowTimed struct {
	port           string
	output         chan<- []byte
	input          chan string
	done           chan bool
	ticker         *time.Ticker
	sPort          string
	bufferedOutput string
}

// NewBufferflowTimed will create a new timed bufferflow
func NewBufferflowTimed(port string, output chan<- []byte) *BufferflowTimed {
	return &BufferflowTimed{
		port:           port,
		output:         output,
		input:          make(chan string),
		done:           make(chan bool),
		ticker:         time.NewTicker(16 * time.Millisecond),
		sPort:          "",
		bufferedOutput: "",
	}
}

// Init will initialize the bufferflow
func (b *BufferflowTimed) Init() {
	log.Println("Initting timed buffer flow (output once every 16ms)")
	go b.consumeInput()
}

func (b *BufferflowTimed) consumeInput() {
Loop:
	for {
		select {
		case data := <-b.input: // use the buffer and append data to it
			b.bufferedOutput = b.bufferedOutput + data
			b.sPort = b.port
		case <-b.ticker.C: // after 16ms send the buffered output message
			if b.bufferedOutput != "" {
				m := SpPortMessage{b.sPort, b.bufferedOutput}
				buf, _ := json.Marshal(m)
				b.output <- buf
				// reset the buffer and the port
				b.bufferedOutput = ""
				b.sPort = ""
			}
		case <-b.done:
			break Loop //this is required, a simple break statement would only exit the innermost switch statement
		}
	}
	close(b.input)
}

// OnIncomingData will forward the data
func (b *BufferflowTimed) OnIncomingData(data string) {
	b.input <- data
}

// Close will close the bufferflow
func (b *BufferflowTimed) Close() {
	b.ticker.Stop()
	b.done <- true
	close(b.done)
}
