/*
 * This file is part of arduino-create-agent.
 *
 * arduino-create-agent is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */
// Package connect allows to open connections to devices connected to a serial port to
// read and write from them
//
// Usage
// 	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
// 	input, output, err := connect.Open(ctx, "/dev/ttyACM0", 9600)
// 	if err != nil {
// 		panic(err)
// 	}
// 	go func() {
// 		for msg := range output {
// 			fmt.Print(string(msg))
// 		}
// 	}()
// 	input <- []byte("some message")
package connect

import (
	"bytes"
	"log"

	"github.com/pkg/errors"

	serial "go.bug.st/serial.v1"
	"golang.org/x/net/context"
)

// Open will establish a connection to a device connected to a serial port
// Will return two channels that will be used to communicate. Use input to send data,
// read from output to retrieve data.
// It accepts a cancelable context, so you can close the connection when you're finished
// Errors during read or write will result in a the error being sent to the output channel
// and the connection being closed
func Open(ctx context.Context, name string, baud int) (input, output chan []byte, err error) {
	mode := &serial.Mode{
		BaudRate: baud,
	}

	p, err := serial.Open(name, mode)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "open %s", name)
	}

	input = make(chan []byte)
	output = make(chan []byte)

	// reader
	go func() {
		for {
			ch := make([]byte, 1024)
			n, err := p.Read(ch)
			if err != nil {
				output <- []byte(err.Error())
				break
			}
			if n > 0 {
				log.Println(len(bytes.Trim(ch, "\x00")))
				output <- bytes.Trim(ch, "\x00")
			}
		}

		close(output)
	}()

	go func() {
	L:
		for {
			select {
			case <-ctx.Done():
				break L
			case msg := <-input:
				_, err := p.Write(msg)
				if err != nil {
					output <- []byte(err.Error())
					break L
				}
			}
		}

		p.Close()
		close(input)
	}()

	return input, output, nil
}
