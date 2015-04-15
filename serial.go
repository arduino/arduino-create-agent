// Supports Windows, Linux, Mac, BeagleBone Black, and Raspberry Pi

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path/filepath"
	//"github.com/kballard/go-shellquote"
	//"github.com/johnlauer/goserial"
	"github.com/mikepb/go-serial"
	"log"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type writeRequest struct {
	p      *serport
	d      string
	buffer bool
	id     string
}

type writeRequestJson struct {
	p    *serport
	P    string
	Data []writeRequestJsonData
}

type writeRequestJsonData struct {
	D   string
	Id  string
	Buf string
}

type qReportJson struct {
	Cmd  string
	QCnt int
	P    string
	Data []qReportJsonData
}

type qReportJsonData struct {
	D     string
	Id    string
	Buf   string `json:"-"`
	Parts int    `json:"-"`
}

type qReport struct {
	Cmd  string
	QCnt int
	Type []string `json:"-"`
	Ids  []string
	D    []string //`json:"-"`
	Port string
}

type serialhub struct {
	// Opened serial ports.
	ports map[*serport]bool

	//open chan *io.ReadWriteCloser
	//write chan *serport, chan []byte
	write chan writeRequest
	//read chan []byte

	writeJson chan writeRequestJson

	// Register requests from the connections.
	register chan *serport

	// Unregister requests from connections.
	unregister chan *serport

	// regexp for json trimming
	reJsonTrim *regexp.Regexp
}

type SpPortList struct {
	SerialPorts []SpPortItem
}

type SpPortItem struct {
	Name                      string
	Friendly                  string
	SerialNumber              string
	DeviceClass               string
	IsOpen                    bool
	IsPrimary                 bool
	RelatedNames              []string
	Baud                      int
	RtsOn                     bool
	DtrOn                     bool
	BufferAlgorithm           string
	AvailableBufferAlgorithms []string
	Ver                       float32
}

var sh = serialhub{
	//write:   	make(chan *serport, chan []byte),
	write:      make(chan writeRequest),
	writeJson:  make(chan writeRequestJson),
	register:   make(chan *serport),
	unregister: make(chan *serport),
	ports:      make(map[*serport]bool),
	reJsonTrim: regexp.MustCompile("sendjson"),
}

func (sh *serialhub) run() {

	log.Print("Inside run of serialhub")
	//cmdIdCtr := 0

	//s := ser.open()
	//ser.s := s
	//ser.write(s, []byte("hello serial data"))
	for {
		select {
		case p := <-sh.register:
			log.Print("Registering a port: ", p.portConf.Name)
			isPrimary := "false"
			if p.IsPrimary {
				isPrimary = "true"
			}
			h.broadcastSys <- []byte("{\"Cmd\":\"Open\",\"Desc\":\"Got register/open on port.\",\"Port\":\"" + p.portConf.Name + "\",\"IsPrimary\":" + isPrimary + ",\"Baud\":" + strconv.Itoa(p.portConf.Baud) + ",\"BufferType\":\"" + p.BufferType + "\"}")
			//log.Print(p.portConf.Name)
			sh.ports[p] = true
		case p := <-sh.unregister:
			log.Print("Unregistering a port: ", p.portConf.Name)
			h.broadcastSys <- []byte("{\"Cmd\":\"Close\",\"Desc\":\"Got unregister/close on port.\",\"Port\":\"" + p.portConf.Name + "\",\"Baud\":" + strconv.Itoa(p.portConf.Baud) + "}")
			delete(sh.ports, p)
			close(p.sendBuffered)
			close(p.sendNoBuf)
		case wrj := <-sh.writeJson:
			// if the user sent in the commands as json
			writeJson(wrj)

		case wr := <-sh.write:
			// if user sent in the commands as one text mode line
			write(wr, "")
		}
	}
}

func writeJson(wrj writeRequestJson) {
	// we'll parse this json request and then do a write() as if
	// the cmd was sent in as text mode

	// create array to hold our qReportJsonData
	qReportDataArr := []qReportJsonData{}

	for _, cmdJson := range wrj.Data {
		var wr writeRequest
		wr.d = cmdJson.D //[]byte(cmdJson.D)
		//wr.id = cmdJson.Id
		wr.p = wrj.p
		if cmdJson.Buf == "Buf" {
			wr.buffer = true
		} else if cmdJson.Buf == "NoBuf" {
			wr.buffer = false
		} else {
			wr.buffer = true
		}
		//write(wr, cmdJson.Id, true)

		// we are sending 1 cmd in, but we may get back multiple cmds
		// because the BreakApartCommands() can add/modify stuff, so keep
		// that in mind
		cmds, idArr, bufTypeArr := createCommands(wr, cmdJson.Id)

		for index, _ := range cmds {
			// create q report data
			cmdId := idArr[index]
			// append stuff to the id if this cmd was one line and just got broken up
			// the first line will have a normal id like "123"
			// the 2nd, 3rd,e tc line will have id like "123-part-2-3", "123-part-3-3"
			if index > 0 {
				cmdId = fmt.Sprintf("%v-part-%v-%v", cmdId, (index + 1), len(cmds))
			}
			qrd := qReportJsonData{D: cmds[index], Id: cmdId, Parts: len(cmds)}
			// if user forced the buffer type, just use it
			if cmdJson.Buf == "Buf" || cmdJson.Buf == "NoBuf" {
				qrd.Buf = cmdJson.Buf
			} else {
				// else use the buffer type figured out in createCommands()
				qrd.Buf = bufTypeArr[index]
			}
			qReportDataArr = append(qReportDataArr, qrd)

		}

	}

	// do our own report
	qr := qReportJson{
		Cmd:  "Queued",
		Data: qReportDataArr,
		QCnt: wrj.p.itemsInBuffer,
		P:    wrj.p.portConf.Name,
	}
	json, _ := json.Marshal(qr)
	h.broadcastSys <- json

	// now send off the commands to the appropriate channel
	for _, qrd := range qReportDataArr {

		if qrd.Buf == "Buf" {
			//log.Println("Json sending to wr.p.sendBuffered")
			wrj.p.sendBuffered <- Cmd{qrd.D, qrd.Id, false, false}
		} else {
			log.Println("Json sending to wr.p.sendNoBuf")
			wrj.p.sendNoBuf <- Cmd{qrd.D, qrd.Id, true, false}
		}
	}

	// garbage collect
	if *gcType == "max" {
		debug.FreeOSMemory()
	}
}

func write(wr writeRequest, id string) {
	cmds, idArr, bufTypeArr := createCommands(wr, id)

	qr := qReport{
		Cmd: "Queued",
		//Type: bufTypeArr,
		Ids:  idArr,
		D:    cmds,
		QCnt: wr.p.itemsInBuffer,
		Port: wr.p.portConf.Name,
	}
	json, _ := json.Marshal(qr)
	h.broadcastSys <- json

	// now send off the commands to the appropriate channel
	for index, cmdToSendToChannel := range cmds {
		//cmdIdCtr++
		//cmdId := "fakeid-" + strconv.Itoa(cmdIdCtr)
		cmdId := idArr[index]
		if bufTypeArr[index] == "Buf" {
			log.Println("Send was normal send, so sending to wr.p.sendBuffered")
			wr.p.sendBuffered <- Cmd{cmdToSendToChannel, cmdId, false, false}
		} else {
			log.Println("Send was sendnobuf, so sending to wr.p.sendNoBuf")
			wr.p.sendNoBuf <- Cmd{cmdToSendToChannel, cmdId, true, false}
		}
	}

	// garbage collect
	if *gcType == "max" {
		debug.FreeOSMemory()
	}

}

func createCommands(wr writeRequest, id string) ([]string, []string, []string) {
	//log.Print("Got a write to a port")
	//log.Print("Port: ", string(wr.p.portConf.Name))
	//log.Print(wr.p)
	//log.Print("Data is: ")
	//log.Println(wr.d)
	//log.Print("Data:" + string(wr.d))
	//log.Print("-----")
	log.Printf("Got write to id:%v, port:%v, buffer:%v, data:%v", id, wr.p.portConf.Name, wr.buffer, string(wr.d))

	dataCmd := string(wr.d)

	// break the data into individual commands for queuing
	// this is important because:
	// 1) we could be sent multiple serial commands at once and the
	//    serial device may want them sent in smaller chunks to give
	//    better feedback. For example, if we're sent G0 X0\nG0 Y10\n we
	//    could happily send that to a CNC controller like a TinyG. However,
	//    on something like TinyG that would chew up 2 buffer planners. So,
	//    to better match what will happen, we break those into 2 commands
	//    so we get a better granularity of getting back qr responses or
	//    other feedback.
	// 2) we need to break apart specific commands potentially that do
	//    not need newlines. For example, on TinyG we need !~% to be different
	//    commands because we need to pivot off of what they mean. ! means pause
	//    the sending. So, we need that command as its own command in order of
	//    how they were sent to us.
	cmds := wr.p.bufferwatcher.BreakApartCommands(dataCmd)
	dataArr := []string{}
	bufTypeArr := []string{}
	idArr := []string{}
	for _, cmd := range cmds {

		// push this cmd onto dataArr for reporting
		dataArr = append(dataArr, cmd)
		idArr = append(idArr, id)

		// do extra check to see if certain command should wipe out
		// the entire internal serial port buffer we're holding in wr.p.sendBuffered
		wipeBuf := wr.p.bufferwatcher.SeeIfSpecificCommandsShouldWipeBuffer(cmd)
		if wipeBuf {
			log.Printf("We got a command that is asking us to wipe the sendBuffered buf. cmd:%v\n", cmd)
			// just wipe out the current channel and create new
			// hopefully garbage collection works here

			// close the channel
			//close(wr.p.sendBuffered)

			// consume all stuff queued
			func() {
				ctr := 0
				/*
					for data := range wr.p.sendBuffered {
						log.Printf("Consuming sendBuffered queue. d:%v\n", string(data))
						ctr++
					}*/

				keepLooping := true
				for keepLooping {
					select {
					case d, ok := <-wr.p.sendBuffered:
						log.Printf("Consuming sendBuffered queue. ok:%v, d:%v, id:%v\n", ok, string(d.data), string(d.id))
						ctr++
						// since we just consumed a buffer item, we need to decrement bufcount
						// we are doing this artificially because we artifically threw
						// away what was in the bufer
						wr.p.itemsInBuffer--
						if ok == false {
							keepLooping = false
						}
					default:
						keepLooping = false
						log.Println("Hit default in select clause")
					}
				}
				log.Printf("Done consuming sendBuffered cmds. ctr:%v\n", ctr)
			}()

			// we still will likely have a sendBuffered that is in the BlockUntilReady()
			// that we have to deal with so it doesn't send to the serial port
			// when we release it
			// send semaphore release if there is one on the BlockUntilReady()
			// this method will release the BlockUntilReady() but with an unblock
			// of type 2 which means cancel the send
			wr.p.bufferwatcher.ReleaseLock()

			// let user know we wiped queue
			log.Printf("itemsInBuffer:%v\n", wr.p.itemsInBuffer)
			h.broadcastSys <- []byte("{\"Cmd\":\"WipedQueue\",\"QCnt\":" + strconv.Itoa(wr.p.itemsInBuffer) + ",\"Port\":\"" + wr.p.portConf.Name + "\"}")

		}

		// do extra check to see if any specific commands should pause
		// the buffer. this means we'll trigger a BlockUntilReady() block
		pauseBuf := wr.p.bufferwatcher.SeeIfSpecificCommandsShouldPauseBuffer(cmd)
		if pauseBuf {
			log.Printf("We need to pause our internal buffer.\n")
			wr.p.bufferwatcher.Pause()
		}

		// do extra check to see if any specific commands should unpause
		// the buffer. this means we'll release the BlockUntilReady() block
		unpauseBuf := wr.p.bufferwatcher.SeeIfSpecificCommandsShouldUnpauseBuffer(cmd)
		if unpauseBuf {
			log.Printf("We need to unpause our internal buffer.\n")
			wr.p.bufferwatcher.Unpause()
		}

		// do extra check to see if certain commands for this buffer type
		// should skip the internal serial port buffering
		// for example ! on tinyg and grbl should skip
		typeBuf := "" // set in if stmt below for reporting afterwards

		if wr.buffer {
			bufferSkip := wr.p.bufferwatcher.SeeIfSpecificCommandsShouldSkipBuffer(cmd)
			if bufferSkip {
				log.Printf("Forcing this cmd to skip buffer. cmd:%v\n", cmd)
				//wr.buffer = false
				typeBuf = "NoBuf"
			} else {
				typeBuf = "Buf"
			}
		} else {
			typeBuf = "NoBuf"
		}

		/*
			if wr.buffer {
				//log.Println("Send was normal send, so sending to wr.p.sendBuffered")
				//wr.p.sendBuffered <- []byte(cmd)
				typeBuf = "Buf"
			} else {
				//log.Println("Send was sendnobuf, so sending to wr.p.sendNoBuf")
				//wr.p.sendNoBuf <- []byte(cmd)
				typeBuf = "NoBuf"
			}
		*/
		// increment queue counter for reporting
		wr.p.itemsInBuffer++
		log.Printf("itemsInBuffer:%v\n", wr.p.itemsInBuffer)

		// push the type of this command to bufTypeArr
		bufTypeArr = append(bufTypeArr, typeBuf)

	} // for loop on broken apart commands

	return cmds, idArr, bufTypeArr
}

func writeToChannels(cmds []string, idArr []string, bufTypeArr []string) {

}

func spList() {

	// call our os specific implementation of getting the serial list
	list, _ := GetList()

	// do a quick loop to see if any of our open ports
	// did not end up in the list port list. this can
	// happen on windows in a fallback scenario where an
	// open port can't be identified because it is locked,
	// so just solve that by manually inserting
	for port := range sh.ports {

		isFound := false
		for _, item := range list {
			if strings.ToLower(port.portConf.Name) == strings.ToLower(item.Name) {
				isFound = true
			}
		}

		if !isFound {
			// artificially push to front of port list
			log.Println(fmt.Sprintf("Did not find an open port in the serial port list. We are going to artificially push it onto the list. port:%v", port.portConf.Name))
			var ossp OsSerialPort
			ossp.Name = port.portConf.Name
			ossp.FriendlyName = port.portConf.Name
			list = append([]OsSerialPort{ossp}, list...)
		}
	}

	// we have a full clean list of ports now. iterate thru them
	// to append the open/close state, baud rates, etc to make
	// a super clean nice list to send back to browser
	n := len(list)
	spl := SpPortList{make([]SpPortItem, n, n)}

	ctr := 0
	for _, item := range list {

		/*
			Name                      string
			Friendly                  string
			IsOpen                    bool
			IsPrimary                 bool
			RelatedNames              []string
			Baud                      int
			RtsOn                     bool
			DtrOn                     bool
			BufferAlgorithm           string
			AvailableBufferAlgorithms []string
			Ver                       float32
		*/
		spl.SerialPorts[ctr] = SpPortItem{
			Name:                      item.Name,
			Friendly:                  item.FriendlyName,
			SerialNumber:              item.SerialNumber,
			DeviceClass:               item.DeviceClass,
			IsOpen:                    false,
			IsPrimary:                 false,
			RelatedNames:              item.RelatedNames,
			Baud:                      0,
			RtsOn:                     false,
			DtrOn:                     false,
			BufferAlgorithm:           "",
			AvailableBufferAlgorithms: availableBufferAlgorithms,
			Ver: versionFloat,
		}

		// figure out if port is open
		//spl.SerialPorts[ctr].IsOpen = false
		myport, isFound := findPortByName(item.Name)

		if isFound {
			// we found our port
			spl.SerialPorts[ctr].IsOpen = true
			spl.SerialPorts[ctr].Baud = myport.portConf.Baud
			spl.SerialPorts[ctr].RtsOn = myport.portConf.RtsOn
			spl.SerialPorts[ctr].DtrOn = myport.portConf.DtrOn
			spl.SerialPorts[ctr].BufferAlgorithm = myport.BufferType
			spl.SerialPorts[ctr].IsPrimary = myport.IsPrimary
		}
		//ls += "{ \"name\" : \"" + item.Name + "\", \"friendly\" : \"" + item.FriendlyName + "\" },\n"
		ctr++
	}

	ls, err := json.MarshalIndent(spl, "", "\t")
	if err != nil {
		log.Println(err)
		h.broadcastSys <- []byte("Error creating json on port list " +
			err.Error())
	} else {
		//log.Print("Printing out json byte data...")
		//log.Print(ls)
		h.broadcastSys <- ls
	}
}

func spListOld() {
	ls := "{\"serialports\" : [\n"
	list, _ := getList()
	for _, item := range list {
		ls += "{ \"name\" : \"" + item.Name + "\", \"friendly\" : \"" + item.FriendlyName + "\" },\n"
	}
	ls = strings.TrimSuffix(ls, "},\n")
	ls += "}\n"
	ls += "]}\n"
	h.broadcastSys <- []byte(ls)
}

func spErr(err string) {
	log.Println("Sending err back: ", err)
	//h.broadcastSys <- []byte(err)
	h.broadcastSys <- []byte("{\"Error\" : \"" + err + "\"}")
}

func spClose(portname string) {
	// look up the registered port by name
	// then call the close method inside serialport
	// that should cause an unregister channel call back
	// to myself

	myport, isFound := findPortByName(portname)

	if isFound {
		// we found our port
		spHandlerClose(myport)
	} else {
		// we couldn't find the port, so send err
		spErr("We could not find the serial port " + portname + " that you were trying to close.")
	}
}

func spProgram(portname string, boardname string, filePath string) {

	isFound, flasher, mycmd := assembleCompilerCommand(boardname, portname, filePath)

	log.Printf("is Found: %v", isFound)

	if isFound {
		spHandlerProgram(flasher, mycmd)
	} else {
		spErr("We could not find the serial port " + portname + " or the board " + boardname + "  that you were trying to program.")
	}
}

func spWriteJson(arg string) {

	log.Printf("spWriteJson. arg:%v\n", arg)

	// remove sendjson string
	arg = sh.reJsonTrim.ReplaceAllString(arg, "")
	//log.Printf("string we're going to parse:%v\n", arg)

	// this is a structured command now for sending in serial commands multiple at a time
	// with an ID so we can send back the ID when the command is done
	var m writeRequestJson
	/*
		m.P = "COM22"
		var data writeRequestJsonData
		data.Id = "234"
		str := "yeah yeah"
		data.D = str //[]byte(str) //[]byte(string("blah blah"))
		m.Data = append(m.Data, data)
		//m.Data = append(m.Data, data)
		bm, err2 := json.Marshal(m)
		if err2 == nil {
			log.Printf("Test json serialize:%v\n", string(bm))
		}
	*/

	err := json.Unmarshal([]byte(arg), &m)

	if err != nil {
		log.Printf("Problem decoding json. giving up. json:%v, err:%v\n", arg, err)
		spErr(fmt.Sprintf("Problem decoding json. giving up. json:%v, err:%v", arg, err))
		return
	}

	// see if we have this port open
	portname := m.P
	myport, isFound := findPortByName(portname)

	if !isFound {
		// we couldn't find the port, so send err
		spErr("We could not find the serial port " + portname + " that you were trying to write to.")
		return
	}

	// we found our port
	m.p = myport

	// send it to the writeJson channel
	sh.writeJson <- m
}

func spWrite(arg string) {
	// we will get a string of comXX asdf asdf asdf
	log.Println("Inside spWrite arg: " + arg)
	arg = strings.TrimPrefix(arg, " ")
	//log.Println("arg after trim: " + arg)
	args := strings.SplitN(arg, " ", 3)
	if len(args) != 3 {
		errstr := "Could not parse send command: " + arg
		log.Println(errstr)
		spErr(errstr)
		return
	}
	portname := strings.Trim(args[1], " ")
	//log.Println("The port to write to is:" + portname + "---")
	//log.Println("The data is:" + args[2] + "---")

	// see if we have this port open
	myport, isFound := findPortByName(portname)

	if !isFound {
		// we couldn't find the port, so send err
		spErr("We could not find the serial port " + portname + " that you were trying to write to.")
		return
	}

	// we found our port
	// create our write request
	var wr writeRequest
	wr.p = myport

	// see if args[0] is send or sendnobuf
	if args[0] != "sendnobuf" {
		// we were just given a "send" so buffer it
		wr.buffer = true
	} else {
		log.Println("sendnobuf specified so wr.buffer is false")
		wr.buffer = false
	}

	// include newline or not in the write? that is the question.
	// for now lets skip the newline
	//wr.d = []byte(args[2] + "\n")
	wr.d = args[2] //[]byte(args[2])

	// send it to the write channel
	sh.write <- wr

}

func formatCmdline(cmdline string, boardOptions map[string]string) (string, bool) {

	list := strings.Split(cmdline, "{")
	fmt.Println("%v", list)
	if len(list) == 1 {
		return cmdline, false
	}
	cmdline = ""
	for _, item := range list {
		item_s := strings.Split(item, "}")
		item = boardOptions[item_s[0]]
		if len(item_s) == 2 {
			cmdline += item + item_s[1]
		} else {
			if item != "" {
				cmdline += item
			} else {
				cmdline += item_s[0]
			}
		}
	}
	fmt.Println(cmdline)
	return cmdline, true
}

func assembleCompilerCommand(boardname string, portname string, filePath string) (bool, string, string) {
	// walk across the local filesystem, find boards.txt files, search for the board in it

	boardFields := strings.Split(boardname, ":")
	if len(boardFields) != 3 {
		h.broadcastSys <- []byte("Board need to be specified in core:architecture:name format")
		return false, "", ""
	}
	file, err := os.Open(boardFields[0] + "/hardware/" + boardFields[1] + "/boards.txt")
	if err != nil {
		h.broadcastSys <- []byte("Could not find board: " + boardname)
		fmt.Println("Error:", err)
		return false, "", ""
	}
	scanner := bufio.NewScanner(file)
	cmdline := ""
	//ide_tools_dir := "./" + boardFields[0] + "/tools"

	boardOptions := make(map[string]string)
	uploadOptions := make(map[string]string)

	for scanner.Scan() {
		// map everything matching with boardname
		if strings.Contains(scanner.Text(), boardFields[2]) {
			arr := strings.Split(scanner.Text(), "=")
			arr[0] = strings.Replace(arr[0], boardFields[2]+".", "", 1)
			boardOptions[arr[0]] = arr[1]
		}
	}

	boardOptions["serial.port"] = portname
	boardOptions["build.project_name"] = strings.Trim(filePath, "\n")

	//fmt.Printf("boardOptions %v %T", boardOptions, boardOptions)

	file.Close()
	file, err = os.Open(boardFields[0] + "/hardware/" + boardFields[1] + "/platform.txt")
	if err != nil {
		h.broadcastSys <- []byte("Could not find board: " + boardname)
		fmt.Println("Error:", err)
		return false, "", ""
	}
	scanner = bufio.NewScanner(file)

	tool := boardOptions["upload.tool"]

	for scanner.Scan() {
		//fmt.Println(scanner.Text());
		// map everything matching with upload
		if strings.Contains(scanner.Text(), tool) {
			arr := strings.Split(scanner.Text(), "=")
			uploadOptions[arr[0]] = arr[1]
			arr[0] = strings.Replace(arr[0], "tools."+tool+".", "", 1)
			boardOptions[arr[0]] = arr[1]
		}
	}
	file.Close()

	version := uploadOptions["runtime.tools."+tool+".version"]
	path, err := filepath.Abs(boardFields[0] + "/tools/" + tool + "/" + version)
	if err != nil {
		h.broadcastSys <- []byte("Could not find board: " + boardname)
		fmt.Println("Error:", err)
		return false, "", ""
	}

	boardOptions["runtime.tools.avrdude.path"] = path

	//boardOptions["config.path"] = uploadOptions["tools."+tool+".config.path"]
	//boardOptions["path"] = uploadOptions["tools."+tool+".path"]

	var winded = true

	cmdline = uploadOptions["tools."+tool+".upload.pattern"]
	// remove cmd.path as it is handles differently
	cmdline = strings.Replace(cmdline, "\"{cmd.path}\"", " ", 1)
	cmdline = strings.Replace(cmdline, "\"", "", -1)

	for winded != false {
		cmdline, winded = formatCmdline(cmdline, boardOptions)
	}

	// cmdline := "-C" + ide_tools_dir + "etc/avrdude.conf" +
	// 	" -c" + protocol +
	// 	" -b" + speed +
	// 	" -p" + mcu +
	// 	" -P" + portname +
	// 	" -D" +
	// 	" -Uflash:w:" + filePath + ":i "

	if boardOptions["upload.use_1200bps_touch"] == "true" {
		// triggers bootloader mode
		log.Println("Restarting in bootloader mode")

		options := serial.RawOptions
		options.BitRate = 1200
		options.FlowControl = serial.FLOWCONTROL_RTSCTS
		p, err := options.Open(portname)
		if err != nil {
			log.Panic(err)
		}
		time.Sleep(time.Second / 2)
		p.Close()
		time.Sleep(time.Second)
	}

	return (tool != ""), tool, cmdline
}

func findPortByName(portname string) (*serport, bool) {
	portnamel := strings.ToLower(portname)
	for port := range sh.ports {
		if strings.ToLower(port.portConf.Name) == portnamel {
			// we found our port
			//spHandlerClose(port)
			return port, true
		}
	}
	return nil, false
}

func spBufferAlgorithms() {
	//arr := []string{"default", "tinyg", "dummypause"}
	arr := availableBufferAlgorithms
	json := "{\"BufferAlgorithm\" : ["
	for _, elem := range arr {
		json += "\"" + elem + "\", "
	}
	json = regexp.MustCompile(", $").ReplaceAllString(json, "]}")
	h.broadcastSys <- []byte(json)
}

func spBaudRates() {
	arr := []string{"2400", "4800", "9600", "19200", "38400", "57600", "115200", "230400"}
	json := "{\"BaudRate\" : ["
	for _, elem := range arr {
		json += "" + elem + ", "
	}
	json = regexp.MustCompile(", $").ReplaceAllString(json, "]}")
	h.broadcastSys <- []byte(json)
}
