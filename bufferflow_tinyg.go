package main

import (
	"log"
	"regexp"
	"strconv"
	"time"
)

type BufferflowTinyg struct {
	Name         string
	Port         string
	Paused       bool
	StopSending  int
	StartSending int
	sem          chan int
}

var (
	// the regular expression to find the qr value
	re, _ = regexp.Compile("\"qr\":(\\d+)")
)

func (b *BufferflowTinyg) Init() {
	b.StartSending = 16
	b.StopSending = 14
	b.sem = make(chan int)
}

func (b *BufferflowTinyg) BlockUntilReady() {
	log.Printf("BlockUntilReady() start\n")
	//log.Printf("buffer:%v\n", b)
	if b.Paused {
		//<-b.sem // will block until told from OnIncomingData to go

		for b.Paused {
			//log.Println("We are paused. Yeilding send.")
			time.Sleep(5 * time.Millisecond)
		}

	} else {
		// still yeild a bit cuz seeing we need to let tinyg
		// have a chance to respond
		time.Sleep(15 * time.Millisecond)
	}
	log.Printf("BlockUntilReady() end\n")
}

func (b *BufferflowTinyg) OnIncomingData(data string) {
	//log.Printf("OnIncomingData() start. data:%v\n", data)
	if re.Match([]byte(data)) {
		// we have a qr value
		//log.Printf("Found a qr value:%v", re)
		res := re.FindStringSubmatch(data)
		qr, err := strconv.Atoi(res[1])
		if err != nil {
			log.Printf("Got error converting qr value. huh? err:%v\n", err)
		} else {
			log.Printf("The qr val is:\"%v\"", qr)
			if qr <= b.StopSending {
				b.Paused = true

				log.Println("Paused sending gcode")
			} else if qr >= b.StartSending {
				b.Paused = false
				//b.sem <- 1 // send channel a val to trigger the unblocking in BlockUntilReady()
				log.Println("Started sending gcode again")
			} else {
				log.Println("In a middle state where we're paused sending gcode but watching for the buffer to get high enough to start sending again")
			}
		}
	}
	// Look for {"qr":28}
	// Actually, if we hit qr:10, stop sending
	// when hit qr:16 start again
	//time.Sleep(3000 * time.Millisecond)
	//log.Printf("OnIncomingData() end.\n")
}
