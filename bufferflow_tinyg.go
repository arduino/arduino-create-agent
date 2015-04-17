package main

import (
	"encoding/json"
	"log"
	"regexp"
	//"strconv"
	"strings"
	"sync"
	//"time"
	"errors"
	"fmt"
	"runtime/debug"
	"time"
)

type BufferflowTinyg struct {
	Name   string
	Port   string
	Paused bool
	//StopSending     int
	//StartSending    int
	//PauseOnEachSend time.Duration // Amount of milliseconds to pause on each send to give TinyG time to send us a qr report
	sem        chan int // semaphore to wait on until given release
	LatestData string   // this holds the latest data across multiple serial reads so we can analyze it for qr responses
	//BypassMode      bool          // this means don't actually watch for qr responses until we know tinyg is in qr response mode
	//wg           sync.WaitGroup

	quit           chan int
	parent_serport *serport

	re                    *regexp.Regexp
	reNewLine             *regexp.Regexp
	reQrOff               *regexp.Regexp
	reQrOn                *regexp.Regexp
	reNoResponse          *regexp.Regexp
	reComment             *regexp.Regexp
	reComment2            *regexp.Regexp
	rePutBackInJsonMode   *regexp.Regexp
	reJsonVerbositySetTo0 *regexp.Regexp
	reCrLfSetTo1          *regexp.Regexp
	reRxResponse          *regexp.Regexp
	reFlowChar            *regexp.Regexp

	// slot counter approach
	reSlotDone *regexp.Regexp // the r:null cmd to look for back from tinyg indicating line processed
	//reCmdsWithNoRResponse *regexp.Regexp // since we're using slot approach, we expect an r:{} response, but some commands don't give that so just don't expect it
	//SlotMax               int            // queue into tinyg using slot approach
	//SlotCtr               int            // queue into tinyg using slot approach

	//lock *sync.Mutex // use a lock/unlock instead of sem chan int

	// do buffer size counting approach instead
	BufferMax int
	//BufferSize      int
	//BufferSizeArray []int
	//BufferCmdArray  []string
	q *Queue

	// use thread locking for b.Paused
	lock *sync.Mutex

	// use more thread locking for b.semLock
	semLock *sync.Mutex
}

type GcodeCmd struct {
	Cmd string
	Id  string
}

type BufFlowCmd struct {
	Cmd                          string
	Gcode                        string
	Resp                         string
	Id                           string
	HowMuchWeThinkWeShouldRemove int
	HowMuchTinyTellsUsToRemove   int
	IsMatchOnBufDecreaseCnt      bool
	IsErr                        bool
	Err                          string
	//TotalInBufPerSpjs            int
	//TotalInBufPerTinyG           int
}

type BufFlowRx struct {
	Cmd                string
	Resp               string
	IsMatchOnTotalBuf  bool
	IsErr              bool
	Err                string
	TotalInBufPerSpjs  int
	TotalInBufPerTinyG int
}

// RawString is a raw encoded JSON object.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
type RawString string

// MarshalJSON returns *m as the JSON encoding of m.
func (m *RawString) MarshalJSON() ([]byte, error) {
	return []byte(*m), nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawString) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("RawString: UnmarshalJSON on nil pointer")
	}
	*m += RawString(data)
	return nil
}

type RespMsg struct {
	R RawString `json:"r",sql:"type:json"`
	F []int     `json:"f"`
}

type RespRxMsg struct {
	R RxMsg `json:"r"`
	F []int `json:"f"`
}

type RxMsg struct {
	Rx int `json:"rx"`
}

func (b *BufferflowTinyg) Init() {

	b.Paused = false
	b.lock = &sync.Mutex{}
	b.semLock = &sync.Mutex{}
	//b.SetPaused(false, 2)

	/* Slot Approach */
	//b.SlotMax = 4 // at most queue up 2 slots, i.e. 2 gcode commands
	//b.SlotCtr = 0 // 0 indicates no gcode lines have been queued into tinyg
	// the regular expression to turn off the pause
	// this regexp will find the r:null response which indicates
	// a line of gcode was processed and thus we can send the next one
	// {"r":{},"f":[1,0,33,134]}
	// when we see this, decrement the b.SlotCtr
	b.reSlotDone, _ = regexp.Compile("{\"r\":{")
	// when we see the response to an rx query so we know how many chars
	// are sitting in the serial buffer
	b.reRxResponse, _ = regexp.Compile("{\"rx\":")
	b.reFlowChar, _ = regexp.Compile("\u0011|\u0013")

	//b.reCmdsWithNoRResponse, _ = regexp.Compile("[!~%]")
	//log.Printf("Using slot approach for TinyG buffering. slotMax:%v, slotCtr:%v\n", b.SlotMax, b.SlotCtr)

	/* End Slot Approach Items */

	/* Start Buffer Size Approach Items */
	b.BufferMax = 200 //max buffer size 254 bytes available
	//b.BufferSize = 0  //initialize buffer at zero bytes
	b.q = NewQueue()
	//b.lock = sync.Mutex
	/* End Buffer Size Approach */

	//b.StartSending = 20
	//b.StopSending = 18
	//b.PauseOnEachSend = 0 * time.Millisecond

	// make buffered channel big enough we won't overflow it
	// meaning we get told b.sem on incoming data, so at most this could
	// be the size of 1 character and the TinyG only allows 255, so just
	// go high to make sure it's high enough to never block
	// buffered
	b.sem = make(chan int, 1000)
	// non-buffered
	//b.sem = make(chan int)

	// start tinyg out in bypass mode because we don't really
	// know if user put tinyg into qr response mode. what we'll
	// do is watch for our first qr response and then assume we're
	// in active mode, i.e. b.BypassMode should then be set to false
	// the reason for this is if we think tinyg is going to send qr
	// responses and we don't get them, we end up holding up all data
	// and essentially break everything. so gotta really watch for this.
	//b.BypassMode = true
	// looking like bypassmode isn't very helpful
	//b.BypassMode = false

	// the regular expression to find the qr value
	// this regexp will find qr when in json mode or non-json mode on tinyg
	b.re, _ = regexp.Compile("\"{0,1}qr\"{0,1}:(\\d+)")

	//reWipeToQr, _ = regexp.Compile("(?s)^.*?\"qr\":\\d+")

	// we split the incoming data on newline using this regexp
	// tinyg seems to only send \n but look for \n\r optionally just in case
	b.reNewLine, _ = regexp.Compile("\\r{0,1}\\n")

	// Look for qr's being turned off by user to auto turn-on BypassMode
	/*
		$qv
		[qv]  queue report verbosity      2 [0=off,1=single,2=triple]
		$qv=0
		[qv]  queue report verbosity      0 [0=off,1=single,2=triple]
		{"qv":""}
		{"r":{"qv":0},"f":[1,0,10,5788]}
	*/
	b.reQrOff, _ = regexp.Compile("{\"qv\":0}|\\[qv\\]\\s+queue report verbosity\\s+0")

	// Look for qr's being turned ON by user to auto turn-off BypassMode
	/*
		$qv
		[qv]  queue report verbosity      3 [0=off,1=single,2=triple]
		{"qv":""}
		{"r":{"qv":3},"f":[1,0,10,5066]}
	*/
	b.reQrOn, _ = regexp.Compile("{\"qv\":[1-9]}|\\[qv\\]\\s+queue report verbosity\\s+[1-9]")

	// this regexp catches !, ~, %, \n, $ by itself, or $$ by itself and indicates
	// no r:{} response will come back so don't expect it
	b.reNoResponse, _ = regexp.Compile("^[!~%\n$?]")

	// if we get a cmd with a $ at the start or a ? at start, append
	// a new command that will put tinyg back in json mode
	b.rePutBackInJsonMode, _ = regexp.Compile("^[$?]")

	// see if they tried to turn off json verbosity, which will break things
	b.reJsonVerbositySetTo0, _ = regexp.Compile("(\\$jv\\=0|\\{\"jv\"\\:0\\})")

	// see if they tried to turn on CRLF, which will break things
	b.reCrLfSetTo1, _ = regexp.Compile("(\\$ec\\=1|\\{\"ec\"\\:1\\})")

	b.reComment, _ = regexp.Compile("\\(.*?\\)")
	b.reComment2, _ = regexp.Compile(";.*")

	//initialize query loop
	//b.rxQueryLoop(b.parent_serport)
}

// Serial buffer size approach
func (b *BufferflowTinyg) BlockUntilReady(cmd string, id string) (bool, bool) {
	log.Printf("BlockUntilReady(cmd:%v, id:%v) start\n", cmd, id)

	// Since BlockUntilReady is in the writer thread, lock so the reader
	// thread doesn't get messed up from all the bufferarray counting we're doing
	//b.lock.Lock()
	//defer b.lock.Unlock()

	// Here we add the length of the new command to the buffer size and append the length
	// to the buffer array.  Check if buffersize > buffermax and if so we pause and await free space before
	// sending the command to grbl.

	// Only increment if cmd is something we'll get an r:{} response to
	isReturnsNoResponse := b.SeeIfSpecificCommandsReturnNoResponse(cmd)
	if isReturnsNoResponse == false {

		b.q.Push(cmd, id)
		/*
			log.Printf("Going to lock inside BlockUntilReady to up the BufferSize and Arrays\n")
			b.lock.Lock()
			b.BufferSize += len(cmd)
			b.BufferSizeArray = append(b.BufferSizeArray, len(cmd))
			b.BufferCmdArray = append(b.BufferCmdArray, cmd)
			b.lock.Unlock()
			log.Printf("Done locking inside BlockUntilReady to up the BufferSize and Arrays\n")
		*/
	} else {
		// this is sketchy. could we overrun the buffer by not counting !~%\n
		// so to give extra room don't actually allow full serial buffer to
		// be used in b.BufferMax
		//log.Printf("Not incrementing buffer size for cmd:%v\n", cmd)

	}

	log.Printf("New line length: %v, buffer size increased to:%v\n", len(cmd), b.q.LenOfCmds())
	//log.Println(b.BufferSizeArray)
	//log.Println(b.BufferCmdArray)

	//b.lock.Lock()
	if b.q.LenOfCmds() >= b.BufferMax {
		b.SetPaused(true, 0) // b.Paused = true
		log.Printf("It looks like the buffer is over the allowed size, so we are going to paused. Then when some incoming responses come in a check will occur to see if there's room to send this command. Pausing...")
	}
	//b.lock.Lock()

	if b.GetPaused() {
		log.Println("It appears we are being asked to pause, so we will wait on b.sem")
		// We are being asked to pause our sending of commands

		// clear all b.sem signals so when we block below, we truly block
		b.ClearOutSemaphore()

		log.Println("Blocking on b.sem until told from OnIncomingData to go")
		unblockType, ok := <-b.sem // will block until told from OnIncomingData to go

		log.Printf("Done blocking cuz got b.sem semaphore release. ok:%v, unblockType:%v\n", ok, unblockType)

		// we get an unblockType of 1 for normal unblocks
		// we get an unblockType of 2 when we're being asked to wipe the buffer, i.e. from a % cmd
		if unblockType == 2 {
			log.Println("This was an unblock of type 2, which means we're being asked to wipe internal buffer. so return false.")
			// returning false asks the calling method to wipe the serial send once
			// this function returns
			return false, false
		}
	}

	// we will get here when we're done blocking and if we weren't cancelled
	// if this cmd returns no response, we need to generate a fake "Complete"
	// so do it now
	willHandleCompleteResponse := true
	if isReturnsNoResponse == true {
		willHandleCompleteResponse = false
	}

	log.Printf("BlockUntilReady(cmd:%v, id:%v) end\n", cmd, id)

	return true, willHandleCompleteResponse
}

// Serial buffer size approach
func (b *BufferflowTinyg) OnIncomingData(data string) {
	//log.Printf("OnIncomingData() start. data:%q\n", data)

	// Since OnIncomingData is in the reader thread, lock so the writer
	// thread doesn't get messed up from all the bufferarray counting we're doing
	//b.lock.Lock()
	//defer b.lock.Unlock()

	b.LatestData += data

	//it was found ok was only received with status responses until the grbl buffer is full.
	//b.LatestData = regexp.MustCompile(">\\r\\nok").ReplaceAllString(b.LatestData, ">") //remove oks from status responses

	arrLines := b.reNewLine.Split(b.LatestData, -1)
	//js, _ := json.Marshal(arrLines)
	//log.Printf("cnt:%v, arrLines:%v\n", len(arrLines), string(js))

	if len(arrLines) > 1 {
		// that means we found a newline and have 2 or greater array values
		// so we need to analyze our arrLines[] lines but keep last line
		// for next trip into OnIncomingData
		//log.Printf("We have data lines to analyze. numLines:%v\n", len(arrLines))

	} else {
		// we don't have a newline yet, so just exit and move on
		// we don't have to reset b.LatestData because we ended up
		// without any newlines so maybe we will next time into this method
		//log.Printf("Did not find newline yet, so nothing to analyze\n")
		return
	}

	// if we made it here we have lines to analyze
	// so analyze all of them except the last line
	for _, element := range arrLines[:len(arrLines)-1] {
		//log.Printf("Working on element:%v, index:%v", element, index)
		//log.Printf("Working on element:%v, index:%v", element)

		//check for r:{} response indicating a gcode line has been processed
		if b.reSlotDone.MatchString(element) {

			//log.Printf("Going to lock inside OnIncomingData to decrease the BufferSize and reset Arrays\n")
			//b.lock.Lock()

			//if b.BufferSizeArray != nil {
			// ok, a line has been processed, the if statement below better
			// be guaranteed to be true, cuz if its not we did something wrong
			if b.q.Len() > 0 {
				//b.BufferSize -= b.BufferSizeArray[0]
				doneCmd, id := b.q.Poll()

				//doneCmd := b.BufferCmdArray[0]
				// Send cmd:"Complete" back
				m := DataCmdComplete{"Complete", id, b.Port, b.q.LenOfCmds(), doneCmd}
				bm, err := json.Marshal(m)
				if err == nil {
					h.broadcastSys <- bm
				}

				// ok, here's the deal. it seems that sometimes we may miss
				// an r:{} coming back to us```

				/*
					if len(b.BufferSizeArray) > 1 {
						b.BufferSizeArray = b.BufferSizeArray[1:len(b.BufferSizeArray)]
						b.BufferCmdArray = b.BufferCmdArray[1:len(b.BufferCmdArray)]
					} else {
						b.BufferSizeArray = nil
						b.BufferCmdArray = nil
					}
				*/

				log.Printf("Buffer decreased to itemCnt:%v, lenOfBuf:%v\n", b.q.Len(), b.q.LenOfCmds())

				if *bufFlowDebugType == "on" {
					// let's report on how our buffer is doing
					// we need to unmarshall this r:{} response

					// do some initial cleanup to remove \u0011 or \u0013
					// that we're getting likely for flow control that is
					// throwing off the unmarshal call
					element2 := b.reFlowChar.ReplaceAllString(element, "")

					// unmarshall r:{} json
					var rm RespMsg
					err2 := json.Unmarshal([]byte(element2), &rm)

					bfc := BufFlowCmd{}
					bfc.Cmd = "BufFlowDebug"
					bfc.Gcode = doneCmd
					bfc.Resp = element
					bfc.Id = id
					bfc.HowMuchWeThinkWeShouldRemove = len(doneCmd)
					bfc.IsErr = false
					bfc.IsMatchOnBufDecreaseCnt = false

					if err2 != nil {
						log.Printf("Problem decoding json on r:{} response. giving up. json:%v, err:%v\n", element, err2)
						spErr(fmt.Sprintf("Problem decoding json on r:{} response. giving up. json:%v, err:%v", element, err2))
						bfc.IsErr = true
						bfc.Err = "Problem unmarshalling json which likely means we had dropped characters on the serial buffer. Giving up."
						//return
					} else {
						log.Printf("RespMsg:%v\n", rm)

						if len(rm.F) > 2 {
							bfc.HowMuchTinyTellsUsToRemove = rm.F[2]
							if rm.F[2] == len(doneCmd) {
								bfc.IsMatchOnBufDecreaseCnt = true
							} else {
								bfc.IsMatchOnBufDecreaseCnt = false
							}
						}

					}

					bfcm, err3 := json.Marshal(bfc)
					if err3 == nil {
						h.broadcastSys <- bfcm
					} else {
						log.Fatal(fmt.Sprintf("Could not marshal the buffer flow debug json response. We should never get here and since we did we are exiting so you can debug me. Giving up. json:%v, err:%v", element, err3))
					}

					// also check for rx value being returned so we can decide
					// if our serial buffer value is the same as what TinyG thinks
					// it should be.
					if b.reRxResponse.MatchString(element) {
						var rrxm RespRxMsg
						err4 := json.Unmarshal([]byte(element2), &rrxm)

						bfrx := BufFlowRx{}
						bfrx.Cmd = "BufFlowRxDebug"
						bfrx.Resp = element
						bfrx.IsErr = false
						bfrx.IsMatchOnTotalBuf = false
						bfrx.TotalInBufPerSpjs = b.q.LenOfCmds()

						if err4 != nil {
							bfrx.IsErr = true
							bfrx.Err = "Could not unmarshall the r:rx json string? huh?"
						} else {
							bfrx.TotalInBufPerTinyG = 254 - rrxm.R.Rx

							// do they match?
							if bfrx.TotalInBufPerSpjs == bfrx.TotalInBufPerTinyG {
								bfrx.IsMatchOnTotalBuf = true
							} else {
								bfrx.IsMatchOnTotalBuf = false
							}
						}

						bfrxm, err5 := json.Marshal(bfrx)
						if err5 == nil {
							h.broadcastSys <- bfrxm
						} else {
							log.Fatal(fmt.Sprintf("Could not marshal the buffer flow debug RX json response. We should never get here and since we did we are exiting so you can debug me. Giving up. json:%v, err:%v", element, err5))
						}

					}
				}

			} else {
				log.Printf("We should NEVER get here cuz we should have a command in the queue to dequeue when we get the r:{} response. If you see this debug stmt this is BAD!!!!")
			}

			//if b.BufferSize < b.BufferMax {
			// We should have our queue dequeued so lets see if we are now below
			// the allowed buffer room. If so go ahead and release the block on send
			// This if stmt still may not be true here because we could have had a tiny
			// cmd just get completed like "G0 X0" and the next cmd is long like "G2 X23.32342 Y23.535355 Z1.04345 I0.243242 J-0.232455"
			// So we'll have to wait until the next time in here for this test to pass
			if b.q.LenOfCmds() < b.BufferMax {

				log.Printf("tinyg just completed a line of gcode and there is room in buffer so setPaused(false)\n")

				// if we are paused, tell us to unpause cuz we have clean buffer room now
				if b.GetPaused() {
					b.SetPaused(false, 1) //set paused to false first, then release the hold on the buffer
				}

				/*
					// if we are paused, tell us to unpause cuz we have clean buffer room now
					b.lock.Lock()
					if b.Paused {

						b.Paused = false

						// send signal to the OnBlockUntilReady method
						// to let it start running again
						b.sem <- 1
						// do this in a goroutine because if multiple sends into the channel
						// occur then the write into the channel will block. we also want
						// to print out debug info when the channel gets consumed so this
						// helps us do that. however, this is a bit inefficient, so could
						// convert b.sem to a buffered channel and just not get debug output
						// or even move to a sync.lock.mutex

							go func() {
								gcodeline := element

								// changed b.SetPaused to here per version 1.75 and Jarret's testing
								//b.SetPaused(false) //set paused to false first, then release the hold on the buffer

								log.Printf("StartSending Semaphore goroutine created for gcodeline:%v\n", gcodeline)
								b.sem <- 1

								defer func() {
									gcodeline := gcodeline
									log.Printf("StartSending Semaphore just got consumed by the BlockUntilReady() thread for the gcodeline:%v\n", gcodeline)
								}()
							}()

					}
					b.lock.Unlock()
				*/
				// let's set that we are no longer paused
				// Not running b.SetPaused() here anymore per version 1.75
				//b.SetPaused(false) //b.Paused = false
			}
			//b.lock.Unlock()
			//log.Printf("Done locking inside OnIncomingData\n")
		}

		// handle communication back to client
		// for base serial data (this is not the cmd:"Write" or cmd:"Complete")
		m := DataPerLine{b.Port, element + "\n"}
		bm, err := json.Marshal(m)
		if err == nil {
			h.broadcastSys <- bm
		}

	} // for loop

	// now wipe the LatestData to only have the last line that we did not analyze
	// because we didn't know/think that was a full command yet
	b.LatestData = arrLines[len(arrLines)-1]

	// we are losing incoming serial data because of garbageCollection()
	// doing a "stop the world" and all this data queues up back on the
	// tinyg and we miss stuff coming in, which gets our serial counter off
	// and then causes stalling, so we're going to attempt to force garbageCollection
	// each time we get data so that we don't have pauses as long as we were having
	if *gcType == "max" {
		debug.FreeOSMemory()
	}

	//time.Sleep(3000 * time.Millisecond)
	//log.Printf("OnIncomingData() end.\n")
}

// Clean out b.sem so it can truly block
func (b *BufferflowTinyg) ClearOutSemaphore() {
	ctr := 0

	keepLooping := true
	for keepLooping {
		select {
		case d, ok := <-b.sem:
			log.Printf("Consuming b.sem queue to clear it before we block. ok:%v, d:%v\n", ok, string(d))
			ctr++
			if ok == false {
				keepLooping = false
			}
		default:
			keepLooping = false
			log.Println("Hit default in select clause")
		}
	}
	log.Printf("Done consuming b.sem queue so we're good to block on it now. ctr:%v\n", ctr)
	// ok, all b.sem signals are now consumed into la-la land

}

// break commands into individual commands
// so, for example, break on newlines to separate commands
// or, in the case of ~% break those onto separate commands
func (b *BufferflowTinyg) BreakApartCommands(cmd string) []string {
	// add newline after !~%
	reSingle := regexp.MustCompile("([!~%])")
	cmd = reSingle.ReplaceAllString(cmd, "$1\n")
	cmds := strings.Split(cmd, "\n")
	//log.Printf("Len of cmds array after split:%v\n", len(cmds))
	//json, _ := json.Marshal(cmds)
	//log.Printf("cmds after split:%v\n", json)
	finalCmds := []string{}
	if len(cmds) == 1 {
		item := cmds[0]
		// just put cmd back in with newline
		if reSingle.MatchString(item) {
			//log.Printf("len1. Added cmd back. Not re-adding newline cuz artificially added one earlier. item:'%v'\n", item)
			finalCmds = append(finalCmds, item)
		} else {
			item = item + "\n"
			//log.Printf("len1. Re-adding item to finalCmds with newline:'%v'\n", item)
			finalCmds = append(finalCmds, item)
		}
	} else {
		for index, item := range cmds {
			// since more than 1 cmd, loop thru
			if reSingle.MatchString(item) {
				//log.Printf("Added cmd back. Not re-adding newline cuz artificially added one earlier. item:'%v'\n", item)
				finalCmds = append(finalCmds, item)
			} else {
				// should we add back our newline? do this if there are elements after us
				if index < len(cmds)-1 {
					// there are cmds after me, so add newline
					//log.Printf("Re-adding newline to item:%v\n", item)
					s := item + "\n"
					finalCmds = append(finalCmds, s)
					//log.Printf("Added cmd back with newline. New cmd item:'%v'\n", s)
				} else {
					//log.Printf("Skipping adding cmd back cuz just empty newline. item:'%v'\n", item)
					//log.Printf("Re-adding item to finalCmds without adding newline:%v\n", item)
					//finalCmds = append(finalCmds, item)
				}

			}
		}
	}

	// loop 1 more time to do some rewriting
	newFinalCmds := []string{}
	for _, item := range finalCmds {
		// remove comments
		//item = b.reComment.ReplaceAllString(item, "")
		//item = b.reComment2.ReplaceAllString(item, "")

		// see if we need to override a cmd to not screw stuff up for us
		// if user sets json verbosity to 0, reset it back
		if match := b.reJsonVerbositySetTo0.MatchString(item); match {
			// they turned off json verbosity, shame on them, override it
			// by setting back
			newFinalCmds = append(newFinalCmds, "{\"jv\":1}\n")
		} else if match := b.reCrLfSetTo1.MatchString(item); match {
			// they turned off json verbosity, shame on them, override it
			// by setting back
			newFinalCmds = append(newFinalCmds, "{\"ec\":0}\n")

		} else {

			// just put the command back into the array without modifying
			newFinalCmds = append(newFinalCmds, item)
		}

		// see if need to put back in json mode
		if match := b.rePutBackInJsonMode.MatchString(item); match {
			// yes, this cmd needs to have us put tinyg back in json mode
			newFinalCmds = append(newFinalCmds, "{\"ej\":\"\"}\n")
		}
	}

	//log.Printf("Final array of cmds after BreakApartCommands(). newFinalCmds:%v\n", newFinalCmds)
	return newFinalCmds
}

func (b *BufferflowTinyg) Pause() {

	// Since we're tweaking b.Paused lock all threads
	//b.lock.Lock()
	//defer b.lock.Unlock()

	b.SetPaused(true, 0) //b.Paused = true
	//b.BypassMode = false // turn off bypassmode in case it's on
	//log.Println("Paused buffer on next BlockUntilReady() call")
}

func (b *BufferflowTinyg) Unpause() {

	// Since we're tweaking b.Paused lock all threads
	//b.lock.Lock()
	//defer b.lock.Unlock()

	b.SetPaused(false, 1) //b.Paused = false
	//log.Println("Unpause(), so we will send signal of 1 to b.sem to unpause the BlockUntilReady() thread")

	// do this as go-routine so we don't block on the b.sem <- 1 write
	/*
		go func() {

			log.Printf("Unpause() Semaphore goroutine created.\n")
			// this is an unbuffered channel, so we will
			// block here which is why this is a goroutine

			// sending a 1 asks BlockUntilReady() to move forward
			b.sem <- 1
			// when we get here that means a BlockUntilReady()
			// method consumed the signal, meaning we unblocked them
			// which is good because they're allowed to start sending
			// again
			defer func() {
				log.Printf("Unpause() Semaphore just got consumed by the BlockUntilReady()\n")
			}()
		}()
	*/
	log.Println("Unpaused buffer inside BlockUntilReady() call")
}

func (b *BufferflowTinyg) SeeIfSpecificCommandsShouldSkipBuffer(cmd string) bool {
	// remove comments
	cmd = b.reComment.ReplaceAllString(cmd, "")
	cmd = b.reComment2.ReplaceAllString(cmd, "")
	if match, _ := regexp.MatchString("[!~%]", cmd); match {
		log.Printf("Found cmd that should skip buffer. cmd:%v\n", cmd)
		return true
	}
	return false
}

func (b *BufferflowTinyg) SeeIfSpecificCommandsShouldPauseBuffer(cmd string) bool {
	// remove comments
	cmd = b.reComment.ReplaceAllString(cmd, "")
	cmd = b.reComment2.ReplaceAllString(cmd, "")
	if match, _ := regexp.MatchString("[!]", cmd); match {
		//log.Printf("Found cmd that should pause buffer. cmd:%v\n", cmd)
		return true
	}
	return false
}

func (b *BufferflowTinyg) SeeIfSpecificCommandsShouldUnpauseBuffer(cmd string) bool {
	// remove comments
	cmd = b.reComment.ReplaceAllString(cmd, "")
	cmd = b.reComment2.ReplaceAllString(cmd, "")
	if match, _ := regexp.MatchString("[~%]", cmd); match {
		//log.Printf("Found cmd that should unpause buffer. cmd:%v\n", cmd)
		return true
	}
	return false
}

func (b *BufferflowTinyg) SeeIfSpecificCommandsShouldWipeBuffer(cmd string) bool {
	// remove comments
	cmd = b.reComment.ReplaceAllString(cmd, "")
	cmd = b.reComment2.ReplaceAllString(cmd, "")
	if match, _ := regexp.MatchString("[%]", cmd); match {
		//log.Printf("Found cmd that should wipe out and reset buffer. cmd:%v\n", cmd)

		// Since we're tweaking b.Paused lock all threads
		//b.lock.Lock()
		//defer b.lock.Unlock()

		//b.BufferSize = 0
		//b.BufferSizeArray = nil
		//b.BufferCmdArray = nil
		//b.q.Delete()
		return true
	}
	return false
}

func (b *BufferflowTinyg) SeeIfSpecificCommandsReturnNoResponse(cmd string) bool {
	// remove comments
	//cmd = b.reComment.ReplaceAllString(cmd, "")
	//cmd = b.reComment2.ReplaceAllString(cmd, "")
	log.Printf("Checking cmd:%v for no response?", cmd)
	if match := b.reNoResponse.MatchString(cmd); match {
		//log.Printf("Found cmd that does not get a response from TinyG. cmd:%v\n", cmd)
		return true
	}
	return false
}

/*
func (b *BufferflowTinyg) RewriteCmd(cmd string) string {
	// remove comments from cmd. why bother sending them to tinyg and wasting
	// precious serial buffer?
	cmd = b.reComment.ReplaceAllString(cmd, "")
	cmd = b.reComment2.ReplaceAllString(cmd, "")
	// if cmd is $, ?, $x=1, etc then rewrap in json
	if match, _ := regexp.MatchString("^[$?]", cmd); match {
		log.Printf("Found cmd that should be wrapped in json. cmd:%v\n", cmd)

	}
	return cmd
}
*/

// This is called if user wiped entire buffer of gcode commands queued up
// which is up to 25,000 of them. So, we need to release the OnBlockUntilReady()
// in a way where the command will not get executed, so send unblockType of 2
func (b *BufferflowTinyg) ReleaseLock() {
	log.Println("Lock being released in TinyG buffer")

	b.q.Delete()
	b.SetPaused(false, 2)
	/*
		// Since we're tweaking b.Paused lock all threads
		b.lock.Lock()

		b.Paused = false
		b.SlotCtr = 0
		b.BufferSize = 0
		b.BufferSizeArray = nil
		b.BufferCmdArray = nil

		b.lock.Unlock()
	*/
	/*
		log.Println("ReleaseLock(), so we will send signal of 2 to b.sem to unpause the BlockUntilReady() thread")
		go func() {

			log.Printf("ReleaseLock() Semaphore goroutine created.\n")
			// this is an unbuffered channel, so we will
			// block here which is why this is a goroutine

			// sending a 2 asks BlockUntilReady() to cancel the send
			b.sem <- 2
			// when we get here that means a BlockUntilReady()
			// method consumed the signal, meaning we unblocked them
			// which is good because they're allowed to start sending
			// again
			defer func() {
				log.Printf("ReleaseLock() Semaphore just got consumed by the BlockUntilReady()\n")
			}()
		}()
	*/
}

func (b *BufferflowTinyg) IsBufferGloballySendingBackIncomingData() bool {
	// we want to send back incoming data as per line data
	// rather than having the default spjs implemenation that sends back data
	// as it sees it. the reason is that we were getting packets out of order
	// on the browser on bad internet connections. that will still happen with us
	// sending back per line data, but at least it will allow the browser to parse
	// correct json now.
	// TODO: The right way to solve this is to watch for an acknowledgement
	// from the browser and queue stuff up until the acknowledgement and then
	// send the full blast of ganged up data
	return true
}

//Use this function to open a connection, write directly to serial port and close connection.
//This is used for sending query requests outside of the normal buffered operations that will pause to wait for room in the grbl buffer
//'?' is asynchronous to the normal buffer load and does not need to be paused when buffer full
func (b *BufferflowTinyg) rxQueryLoop(p *serport) {
	b.parent_serport = p //make note of this port for use in clearing the buffer later, on error.
	ticker := time.NewTicker(5000 * time.Millisecond)
	b.quit = make(chan int)
	go func() {
		for {
			select {
			case <-ticker.C:

				// we'll write a lazy formatted version of json to reduce the amt of chars
				// chewed up since we're doing this outside the scope of the serial buffer counter
				n2, err := p.portIo.Write([]byte("{rx:n}\n"))

				log.Print("Just wrote ", n2, " bytes to serial: {rx:n}")

				if err != nil {
					errstr := "Error writing to " + p.portConf.Name + " " + err.Error() + " Closing port."
					log.Print(errstr)
					h.broadcastSys <- []byte(errstr)
					ticker.Stop() //stop query loop if we can't write to the port
					break
				}
			case <-b.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (b *BufferflowTinyg) Close() {
	//stop the rx query loop when the serial port is closed off.
	log.Println("Stopping the RX query loop")
	b.ReleaseLock()
	b.Unpause()
	go func() {
		b.quit <- 1
	}()
}

//	Gets the paused state of this buffer
//	go-routine safe.
func (b *BufferflowTinyg) GetPaused() bool {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.Paused
}

//	Sets the paused state of this buffer
//	go-routine safe.
func (b *BufferflowTinyg) SetPaused(isPaused bool, semRelease int) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.Paused = isPaused

	// only release semaphore if we are being told to unpause
	if b.Paused == false {
		// the BlockUntilReady thread should be sitting waiting
		// so when we send this should trigger it
		b.sem <- semRelease

		// since the first consuming of the semRelease will occur
		// by BlockUntilReady since it's sitting waiting then
		// we're good to go ahead and release the rest here
		// so our queue doesn't fill up
		// that's the theory anyway
		//b.ClearOutSemaphore()
	}
	//go func() {
	//log.Printf("StartSending Semaphore goroutine created for gcodeline:%v\n", gcodeline)
	//b.sem <- semRelease

	/*
		defer func() {
			//log.Printf("StartSending Semaphore just got consumed by the BlockUntilReady() thread for the gcodeline:%v\n", gcodeline)
		}()
	*/
	//}()
}
