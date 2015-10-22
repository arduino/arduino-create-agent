package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/facchinm/go-serial"
	"github.com/mattn/go-shellwords"
	"github.com/sfreiberg/simplessh"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var compiling = false

func colonToUnderscore(input string) string {
	output := strings.Replace(input, ":", "_", -1)
	return output
}

type basicAuthData struct {
	UserName string
	Password string
}

type boardExtraInfo struct {
	use_1200bps_touch    bool
	wait_for_upload_port bool
	networkPort          bool
	authdata             basicAuthData
}

// Scp uploads sourceFile to remote machine like native scp console app.
func Scp(client *simplessh.Client, sourceFile, targetFile string) error {

	session, err := client.SSHClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	src, srcErr := os.Open(sourceFile)

	if srcErr != nil {
		return srcErr
	}

	srcStat, statErr := src.Stat()

	if statErr != nil {
		return statErr
	}

	go func() {
		w, _ := session.StdinPipe()

		fmt.Fprintln(w, "C0644", srcStat.Size(), filepath.Base(targetFile))

		if srcStat.Size() > 0 {
			io.Copy(w, src)
			fmt.Fprint(w, "\x00")
			w.Close()
		} else {
			fmt.Fprint(w, "\x00")
			w.Close()
		}

	}()

	if err := session.Run("scp -t " + targetFile); err != nil {
		return err
	}

	return nil
}

func spProgramSSHNetwork(portname string, boardname string, filePath string, commandline string, authdata basicAuthData) error {
	log.Println("Starting network upload")
	log.Println("Board Name: " + boardname)

	if authdata.UserName == "" {
		authdata.UserName = "root"
	}

	if authdata.Password == "" {
		authdata.Password = "arduino"
	}

	ssh_client, err := simplessh.ConnectWithPassword(portname+":22", authdata.UserName, authdata.Password)
	if err != nil {
		log.Println("Error connecting via ssh")
		return err
	}
	defer ssh_client.Close()

	err = Scp(ssh_client, filePath, "/tmp/sketch"+filepath.Ext(filePath))
	if err != nil {
		log.Printf("Upload: %s\n", err)
		return err
	}

	if commandline == "" {
		// very special case for Yun (remove once AVR boards.txt is fixed)
		commandline = "merge-sketch-with-bootloader.lua /tmp/sketch.hex && /usr/bin/run-avrdude /tmp/sketch.hex"
	}

	fmt.Println(commandline)

	ssh_output, err := ssh_client.Exec(commandline)
	if err == nil {
		log.Printf("Flash: %s\n", ssh_output)
		mapD := map[string]string{"ProgrammerStatus": "Busy", "Msg": string(ssh_output)}
		mapB, _ := json.Marshal(mapD)
		h.broadcastSys <- mapB
	}
	return err
}

func spProgramNetwork(portname string, boardname string, filePath string, authdata basicAuthData) error {

	log.Println("Starting network upload")
	log.Println("Board Name: " + boardname)

	if authdata.UserName == "" {
		authdata.UserName = "root"
	}

	if authdata.Password == "" {
		authdata.Password = "arduino"
	}

	// Prepare a form that you will submit to that URL.
	_url := "http://" + portname + "/data/upload_sketch_silent"
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add your image file
	filePath = strings.Trim(filePath, "\n")
	f, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file" + filePath + " err: " + err.Error())
		return err
	}
	fw, err := w.CreateFormFile("sketch_hex", filePath)
	if err != nil {
		log.Println("Error creating form file")
		return err
	}
	if _, err = io.Copy(fw, f); err != nil {
		log.Println("Error copying form file")
		return err
	}
	// Add the other fields
	if fw, err = w.CreateFormField("board"); err != nil {
		log.Println("Error creating form field")
		return err
	}
	if _, err = fw.Write([]byte(colonToUnderscore(boardname))); err != nil {
		log.Println("Error writing form field")
		return err
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", _url, &b)
	if err != nil {
		log.Println("Error creating post request")
		return err
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())
	if authdata.UserName != "" {
		req.SetBasicAuth(authdata.UserName, authdata.Password)
	}

	//h.broadcastSys <- []byte("Start flashing with command " + cmdString)
	log.Printf("Network flashing on " + portname)
	mapD := map[string]string{"ProgrammerStatus": "Starting", "Cmd": "POST"}
	mapB, _ := json.Marshal(mapD)
	h.broadcastSys <- mapB

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error during post request")
		return err
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		log.Errorf("bad status: %s", res.Status)
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return err
}

func spProgramLocal(portname string, boardname string, filePath string, commandline string, extraInfo boardExtraInfo) error {

	var err error
	if extraInfo.use_1200bps_touch {
		portname, err = touch_port_1200bps(portname, extraInfo.wait_for_upload_port)
	}

	if err != nil {
		log.Println("Could not touch the port")
		return err
	}

	log.Printf("Received commandline (unresolved):" + commandline)

	commandline = strings.Replace(commandline, "{build.path}", filepath.ToSlash(filepath.Dir(filePath)), 1)
	commandline = strings.Replace(commandline, "{build.project_name}", strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filepath.Base(filePath))), 1)
	commandline = strings.Replace(commandline, "{serial.port}", portname, 1)
	commandline = strings.Replace(commandline, "{serial.port.file}", filepath.Base(portname), 1)

	// search for runtime variables and replace with values from globalToolsMap
	var runtimeRe = regexp.MustCompile("\\{(.*?)\\}")
	runtimeVars := runtimeRe.FindAllString(commandline, -1)

	fmt.Println(runtimeVars)

	for _, element := range runtimeVars {
		commandline = strings.Replace(commandline, element, globalToolsMap[element], 1)
	}

	z, _ := shellwords.Parse(commandline)
	return spHandlerProgram(z[0], z[1:])
}

func spProgramRW(portname string, boardname string, filePath string, commandline string, extraInfo boardExtraInfo) {
	compiling = true

	defer func() {
		time.Sleep(1500 * time.Millisecond)
		compiling = false
	}()

	var err error

	if extraInfo.networkPort {
		err = spProgramNetwork(portname, boardname, filePath, extraInfo.authdata)
		if err != nil {
			// no http method available, try ssh upload
			err = spProgramSSHNetwork(portname, boardname, filePath, commandline, extraInfo.authdata)
		}
	} else {
		err = spProgramLocal(portname, boardname, filePath, commandline, extraInfo)
	}

	if err != nil {
		log.Printf("Command finished with error: %v", err)
		mapD := map[string]string{"ProgrammerStatus": "Error", "Msg": "Could not program the board"}
		mapB, _ := json.Marshal(mapD)
		h.broadcastSys <- mapB
	} else {
		log.Printf("Finished without error. Good stuff")
		mapD := map[string]string{"ProgrammerStatus": "Done", "Flash": "Ok"}
		mapB, _ := json.Marshal(mapD)
		h.broadcastSys <- mapB
		// analyze stdin
	}
}

var oscmd *exec.Cmd

func spHandlerProgram(flasher string, cmdString []string) error {

	// if runtime.GOOS == "darwin" {
	// 	sh, _ := exec.LookPath("sh")
	// 	// prepend the flasher to run it via sh
	// 	cmdString = append([]string{flasher}, cmdString...)
	// 	oscmd = exec.Command(sh, cmdString...)
	// } else {

	// remove quotes form flasher command and cmdString
	flasher = strings.Replace(flasher, "\"", "", -1)

	for index, _ := range cmdString {
		cmdString[index] = strings.Replace(cmdString[index], "\"", "", -1)
	}

	extension := ""
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}

	oscmd = exec.Command(flasher+extension, cmdString...)

	tellCommandNotToSpawnShell(oscmd)

	stdout, err := oscmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := oscmd.StderrPipe()
	if err != nil {
		return err
	}

	multi := io.MultiReader(stderr, stdout)

	// Stdout buffer
	//var cmdOutput []byte

	//h.broadcastSys <- []byte("Start flashing with command " + cmdString)
	log.Printf("Flashing with command:" + flasher + extension + " " + strings.Join(cmdString, " "))
	mapD := map[string]string{"ProgrammerStatus": "Starting", "Cmd": strings.Join(cmdString, " ")}
	mapB, _ := json.Marshal(mapD)
	h.broadcastSys <- mapB

	err = oscmd.Start()

	in := bufio.NewScanner(multi)

	in.Split(bufio.ScanLines)

	for in.Scan() {
		log.Info(in.Text())
		mapD := map[string]string{"ProgrammerStatus": "Busy", "Msg": in.Text()}
		mapB, _ := json.Marshal(mapD)
		h.broadcastSys <- mapB
	}

	err = oscmd.Wait()

	return err
}

func spHandlerProgramKill() {

	// Kill the process if there is one running
	if oscmd != nil && oscmd.Process.Pid > 0 {
		h.broadcastSys <- []byte("{\"ProgrammerStatus\": \"PreKilled\", \"Pid\": " + strconv.Itoa(oscmd.Process.Pid) + ", \"ProcessState\": \"" + oscmd.ProcessState.String() + "\"}")
		oscmd.Process.Kill()
		h.broadcastSys <- []byte("{\"ProgrammerStatus\": \"Killed\", \"Pid\": " + strconv.Itoa(oscmd.Process.Pid) + ", \"ProcessState\": \"" + oscmd.ProcessState.String() + "\"}")

	} else {
		if oscmd != nil {
			h.broadcastSys <- []byte("{\"ProgrammerStatus\": \"KilledError\", \"Msg\": \"No current process\", \"Pid\": " + strconv.Itoa(oscmd.Process.Pid) + ", \"ProcessState\": \"" + oscmd.ProcessState.String() + "\"}")
		} else {
			h.broadcastSys <- []byte("{\"ProgrammerStatus\": \"KilledError\", \"Msg\": \"No current process\"}")
		}
	}
}

func formatCmdline(cmdline string, boardOptions map[string]string) (string, bool) {

	list := strings.Split(cmdline, "{")
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
	log.Println(cmdline)
	return cmdline, true
}

func containsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func findNewPortName(slice1 []string, slice2 []string) string {
	m := map[string]int{}

	for _, s1Val := range slice1 {
		m[s1Val] = 1
	}
	for _, s2Val := range slice2 {
		m[s2Val] = m[s2Val] + 1
	}

	for mKey, mVal := range m {
		if mVal == 1 {
			return mKey
		}
	}

	return ""
}

func touch_port_1200bps(portname string, wait_for_upload_port bool) (string, error) {
	initialPortName := portname
	log.Println("Restarting in bootloader mode")

	mode := &serial.Mode{
		BaudRate: 1200,
		Vmin:     0,
		Vtimeout: 1,
	}
	port, err := serial.OpenPort(portname, mode)
	if err != nil {
		log.Println(err)
		return "", err
	}
	err = port.SetDTR(false)
	if err != nil {
		log.Println(err)
	}
	port.Close()
	time.Sleep(time.Second / 2.0)

	timeout := false
	go func() {
		time.Sleep(2 * time.Second)
		timeout = true
	}()

	// time.Sleep(time.Second / 4)
	// wait for port to reappear
	if wait_for_upload_port {
		after_reset_ports, _ := serial.GetPortsList()
		log.Println(after_reset_ports)
		var ports []string
		for {
			ports, _ = serial.GetPortsList()
			log.Println(ports)
			time.Sleep(time.Millisecond * 200)
			portname = findNewPortName(ports, after_reset_ports)
			if portname != "" {
				break
			}
			if timeout {
				break
			}
		}
	}

	if portname == "" {
		portname = initialPortName
	}
	return portname, nil
}
