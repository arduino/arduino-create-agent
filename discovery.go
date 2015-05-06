//
//  discovery.go
//
//  Created by Martino Facchin
//  Copyright (c) 2015 Arduino LLC
//
//  Permission is hereby granted, free of charge, to any person
//  obtaining a copy of this software and associated documentation
//  files (the "Software"), to deal in the Software without
//  restriction, including without limitation the rights to use,
//  copy, modify, merge, publish, distribute, sublicense, and/or sell
//  copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following
//  conditions:
//
//  The above copyright notice and this permission notice shall be
//  included in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
//  OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
//  HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
//  FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
//

package main

import (
	"github.com/oleksandr/bonjour"
	"log"
	"strings"
	"time"
)

const timeoutConst = 2

func GetNetworkList() ([]OsSerialPort, error) {
	resolver, err := bonjour.NewResolver(nil)
	if err != nil {
		log.Println("Failed to initialize resolver:", err.Error())
		return nil, err
	}

	timeout := make(chan bool, 1)
	go func(exitCh chan<- bool) {
		time.Sleep(timeoutConst * time.Second)
		timeout <- true
		exitCh <- true
	}(resolver.Exit)

	results := make(chan *bonjour.ServiceEntry)
	arrPorts := []OsSerialPort{}
	go func(results chan *bonjour.ServiceEntry, exitCh chan<- bool) {
		for e := range results {
			log.Printf("%s %s %d %s", e.Instance, e.AddrIPv4, e.Port, e.Text)
			var boardInfosSlice []string
			for _, element := range e.Text {
				if strings.Contains(element, "board=yun") {
					boardInfosSlice = append(boardInfosSlice, "arduino:avr:yun")
				}
			}
			arrPorts = append(arrPorts, OsSerialPort{Name: e.AddrIPv4.String(), FriendlyName: e.Instance, NetworkPort: true, RelatedNames: boardInfosSlice})
		}
	}(results, resolver.Exit)

	err = resolver.Browse("_arduino._tcp", "", results)
	if err != nil {
		log.Println("Failed to browse:", err.Error())
		return nil, err
	}
	// wait for some kind of timeout and return arrPorts
	select {
	case <-timeout:
		return arrPorts, nil
	}
}
