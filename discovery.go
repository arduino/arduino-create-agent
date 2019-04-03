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
	"net"
	"strings"
	"time"

	"github.com/oleksandr/bonjour"
	log "github.com/sirupsen/logrus"
)

const timeoutConst = 2

// SavedNetworkPorts contains the ports which we know are already connected
var SavedNetworkPorts []OsSerialPort

// GetNetworkList returns a list of Network Ports
// The research of network ports is articulated in two phases. First we add new ports coming from
// the bonjour module, then we prune the boards who don't respond to a ping
func GetNetworkList() ([]OsSerialPort, error) {
	newPorts, err := getPorts()
	if err != nil {
		return nil, err
	}

	SavedNetworkPorts = Filter(SavedNetworkPorts, func(port OsSerialPort) bool {
		any := true
		for _, p := range newPorts {
			if p.Name == port.Name {
				any = false
				return any
			}
		}
		return any
	})

	SavedNetworkPorts, err = pruneUnreachablePorts(SavedNetworkPorts)
	if err != nil {
		return nil, err
	}

	SavedNetworkPorts = append(SavedNetworkPorts, newPorts...)

	return SavedNetworkPorts, nil
}

func checkAvailability(ip string) bool {
	timeout := time.Duration(1500 * time.Millisecond)
	// Check if the port 80 is open
	conn, err := net.DialTimeout("tcp", ip+":80", timeout)
	if err != nil {
		log.Println(err)
		// Check if the port 22 is open
		conn, err = net.DialTimeout("tcp", ip+":22", timeout)
		if err != nil {
			log.Println(err)
			return false
		}
		conn.Close()
		return true
	}
	conn.Close()
	return true
}

func pruneUnreachablePorts(ports []OsSerialPort) ([]OsSerialPort, error) {
	times := 2

	ports = Filter(ports, func(port OsSerialPort) bool {
		any := false
		for i := 0; i < times; i++ {
			if checkAvailability(port.Name) {
				any = true
			}
		}
		return any
	})

	return ports, nil
}

func getPorts() ([]OsSerialPort, error) {
	resolver, err := bonjour.NewResolver(nil)
	if err != nil {
		log.Println("Failed to initialize resolver:", err.Error())
		return nil, err
	}

	results := make(chan *bonjour.ServiceEntry)

	timeout := make(chan bool, 1)
	go func(exitCh chan<- bool) {
		time.Sleep(timeoutConst * time.Second)
		exitCh <- true
		close(results)
	}(resolver.Exit)

	arrPorts := []OsSerialPort{}
	go func(results chan *bonjour.ServiceEntry, exitCh chan<- bool) {
		for e := range results {
			log.Printf("%+v", e)
			if e.AddrIPv4 != nil {
				arrPorts = append(arrPorts, OsSerialPort{Name: e.AddrIPv4.String(), IdProduct: e.Instance, IdVendor: strings.Join(e.Text[:], " "), NetworkPort: true})
			}
		}
		timeout <- true
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

// Filter returns a new slice containing all OsSerialPort in the slice that satisfy the predicate f.
func Filter(vs []OsSerialPort, f func(OsSerialPort) bool) []OsSerialPort {
	var vsf []OsSerialPort
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
