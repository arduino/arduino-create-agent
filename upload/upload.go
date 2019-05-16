package upload

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

	"github.com/arduino/arduino-create-agent/utilities"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
	"github.com/sfreiberg/simplessh"
	serial "go.bug.st/serial.v1"
	"go.bug.st/serial.v1/enumerator"
)

// Busy tells wether the programmer is doing something
var Busy = false

// Auth contains username and password used for a network upload
type Auth struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	Port       int    `json:"port"`
}

// Extra contains some options used during the upload
type Extra struct {
	Use1200bpsTouch   bool   `json:"use_1200bps_touch"`
	WaitForUploadPort bool   `json:"wait_for_upload_port"`
	Network           bool   `json:"network"`
	Auth              Auth   `json:"auth"`
	Verbose           bool   `json:"verbose"`
	ParamsVerbose     string `json:"params_verbose"`
	ParamsQuiet       string `json:"params_quiet"`
	SSH               bool   `json:"ssh,omitempty"`
}

// PartiallyResolve replaces some symbols in the commandline with the appropriate values
// it can return an error when looking a variable in the Locater
func PartiallyResolve(board, file, platformPath, commandline string, extra Extra, t Locater) (string, error) {
	commandline = strings.Replace(commandline, "{build.path}", filepath.ToSlash(filepath.Dir(file)), -1)
	commandline = strings.Replace(commandline, "{build.project_name}", strings.TrimSuffix(filepath.Base(file), filepath.Ext(filepath.Base(file))), -1)
	commandline = strings.Replace(commandline, "{network.password}", extra.Auth.Password, -1)
	commandline = strings.Replace(commandline, "{runtime.platform.path}", filepath.ToSlash(platformPath), -1)

	if extra.Verbose == true {
		commandline = strings.Replace(commandline, "{upload.verbose}", extra.ParamsVerbose, -1)
	} else {
		commandline = strings.Replace(commandline, "{upload.verbose}", extra.ParamsQuiet, -1)
	}

	// search for runtime variables and replace with values from Locater
	var runtimeRe = regexp.MustCompile("\\{(.*?)\\}")
	runtimeVars := runtimeRe.FindAllString(commandline, -1)

	for _, element := range runtimeVars {

		location, err := t.GetLocation(element)
		if err != nil {
			return "", errors.Wrapf(err, "get location of %s", element)
		}
		if location != "" {
			commandline = strings.Replace(commandline, element, location, 1)
		}
	}

	return commandline, nil
}

func fixupPort(port, commandline string) string {
	commandline = strings.Replace(commandline, "{serial.port}", port, -1)
	commandline = strings.Replace(commandline, "{serial.port.file}", filepath.Base(port), -1)
	ports, err := enumerator.GetDetailedPortsList()
	if err == nil {
		for _, p := range ports {
			if p.Name == port {
				commandline = strings.Replace(commandline, "{serial.port.iserial}", p.SerialNumber, -1)
			}
		}
	}
	return commandline
}

// Network performs a network upload
func Network(port, board string, files []string, commandline string, auth Auth, l Logger, SSH bool) error {
	Busy = true

	// Defaults
	if auth.Username == "" {
		auth.Username = "root"
	}
	if auth.Password == "" {
		auth.Password = "arduino"
	}

	commandline = fixupPort(port, commandline)

	// try with ssh
	err := ssh(port, files, commandline, auth, l, SSH)
	if err != nil && !SSH {
		// fallback on form
		err = form(port, board, files[0], auth, l)
	}

	Busy = false
	return err
}

// Serial performs a serial upload
func Serial(port, commandline string, extra Extra, l Logger) error {
	Busy = true
	defer func() { Busy = false }()

	// some boards needs to be resetted
	if extra.Use1200bpsTouch {
		var err error
		port, err = reset(port, extra.WaitForUploadPort, l)
		if err != nil {
			return errors.Wrapf(err, "Reset before upload")
		}
	}

	commandline = fixupPort(port, commandline)

	z, err := shellwords.Parse(commandline)
	if err != nil {
		return errors.Wrapf(err, "Parse commandline")
	}

	return program(z[0], z[1:], l)
}

var cmds = map[*exec.Cmd]bool{}

// Kill stops any upload process as soon as possible
func Kill() {
	for cmd := range cmds {
		if cmd.Process.Pid > 0 {
			cmd.Process.Kill()
		}
	}
}

// reset opens the port at 1200bps. It returns the new port name (which could change
// sometimes) and an error (usually because the port listing failed)
func reset(port string, wait bool, l Logger) (string, error) {
	info(l, "Restarting in bootloader mode")

	// Get port list before reset
	ports, err := serial.GetPortsList()
	info(l, "Get port list before reset")
	if err != nil {
		return "", errors.Wrapf(err, "Get port list before reset")
	} else {
		info(l, ports)
	}

	// Touch port at 1200bps
	err = touchSerialPortAt1200bps(port, l)
	if err != nil {
		return "", errors.Wrapf(err, "1200bps Touch")
	}

	// Wait for port to disappear and reappear
	if wait {
		port = waitReset(ports, l, port)
	}

	return port, nil
}

func touchSerialPortAt1200bps(port string, l Logger) error {
	info(l, "Touching port ", port, " at 1200bps")

	// Open port
	p, err := serial.Open(port, &serial.Mode{BaudRate: 1200})
	if err != nil {
		return errors.Wrapf(err, "Open port %s", port)
	}
	defer p.Close()

	// Set DTR
	err = p.SetDTR(false)
	info(l, "Set DTR off")
	if err != nil {
		return errors.Wrapf(err, "Can't set DTR")
	}

	// Wait a bit to allow restart of the board
	time.Sleep(200 * time.Millisecond)

	return nil
}

// waitReset is meant to be called just after a reset. It watches the ports connected
// to the machine until a port disappears and reappears. The port name could be different
// so it returns the name of the new port.
func waitReset(beforeReset []string, l Logger, originalPort string) string {
	var port string
	timeout := false

	go func() {
		time.Sleep(10 * time.Second)
		timeout = true
	}()

	// Wait for the port to disappear
	debug(l, "Wait for the port to disappear")
	for {
		ports, err := serial.GetPortsList()
		port = differ(ports, beforeReset)
		debug(l, beforeReset, " -> ", ports)

		if port != "" {
			break
		}
		if timeout {
			debug(l, ports, err, port)
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	// Wait for the port to reappear
	debug(l, "Wait for the port to reappear")
	afterReset, _ := serial.GetPortsList()
	for {
		ports, _ := serial.GetPortsList()
		port = differ(ports, afterReset)
		debug(l, afterReset, " -> ", ports)
		if port != "" {
			debug(l, "Found upload port: ", port)
			time.Sleep(time.Millisecond * 500)
			break
		}
		if timeout {
			debug(l, "timeout")
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	// try to upload on the existing port if the touch was ineffective
	if port == "" {
		port = originalPort
	}

	return port
}

// program spawns the given binary with the given args, logging the sdtout and stderr
// through the Logger
func program(binary string, args []string, l Logger) error {
	// remove quotes form binary command and args
	binary = strings.Replace(binary, "\"", "", -1)

	for i := range args {
		args[i] = strings.Replace(args[i], "\"", "", -1)
	}

	// find extension
	extension := ""
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}

	cmd := exec.Command(binary, args...)

	// Add the command to the map of running commands
	cmds[cmd] = true
	defer func() {
		delete(cmds, cmd)
	}()

	utilities.TellCommandNotToSpawnShell(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrapf(err, "Retrieve output")
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrapf(err, "Retrieve output")
	}

	info(l, "Flashing with command:"+binary+extension+" "+strings.Join(args, " "))

	err = cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "Start command")
	}

	stdoutCopy := bufio.NewScanner(stdout)
	stderrCopy := bufio.NewScanner(stderr)

	stdoutCopy.Split(bufio.ScanLines)
	stderrCopy.Split(bufio.ScanLines)

	go func() {
		for stdoutCopy.Scan() {
			info(l, stdoutCopy.Text())
		}
	}()

	go func() {
		for stderrCopy.Scan() {
			info(l, stderrCopy.Text())
		}
	}()

	err = cmd.Wait()
	if err != nil {
		return errors.Wrapf(err, "Executing command")
	}
	return nil
}

func form(port, board, file string, auth Auth, l Logger) error {
	// Prepare a form that you will submit to that URL.
	_url := "http://" + port + "/data/upload_sketch_silent"
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add your image file
	file = strings.Trim(file, "\n")
	f, err := os.Open(file)
	if err != nil {
		return errors.Wrapf(err, "Open file %s", file)
	}
	fw, err := w.CreateFormFile("sketch_hex", file)
	if err != nil {
		return errors.Wrapf(err, "Create form file")
	}
	if _, err = io.Copy(fw, f); err != nil {
		return errors.Wrapf(err, "Copy form file")
	}

	// Add the other fields
	board = strings.Replace(board, ":", "_", -1)
	if fw, err = w.CreateFormField("board"); err != nil {
		return errors.Wrapf(err, "Create board field")
	}
	if _, err = fw.Write([]byte(board)); err != nil {
		return errors.Wrapf(err, "")
	}

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", _url, &b)
	if err != nil {
		return errors.Wrapf(err, "Create POST req")
	}

	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())
	if auth.Username != "" {
		req.SetBasicAuth(auth.Username, auth.Password)
	}

	info(l, "Network upload on ", port)

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error during post request")
		return errors.Wrapf(err, "")
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		return errors.New("Request error:" + string(body))
	}
	return nil
}

func ssh(port string, files []string, commandline string, auth Auth, l Logger, SSH bool) error {
	debug(l, "Connect via ssh ", files, commandline)

	if auth.Port == 0 {
		auth.Port = 22
	}

	// Connect via ssh
	var client *simplessh.Client
	var err error
	if auth.PrivateKey != "" {
		client, err = simplessh.ConnectWithKeyFile(port+":"+strconv.Itoa(auth.Port), auth.Username, auth.PrivateKey)
	} else {
		client, err = simplessh.ConnectWithPassword(port+":"+strconv.Itoa(auth.Port), auth.Username, auth.Password)
	}

	if err != nil {
		return errors.Wrapf(err, "Connect via ssh")
	}
	defer client.Close()

	// Copy the sketch
	for _, file := range files {
		fileName := "/tmp/sketch" + filepath.Ext(file)
		if SSH {
			// don't rename files
			fileName = "/tmp/" + filepath.Base(file)
		}
		err = scp(client, file, fileName)
		debug(l, "Copy "+file+" to "+fileName+" ", err)
		if err != nil {
			return errors.Wrapf(err, "Copy sketch")
		}
	}

	// very special case for Yun (remove once AVR boards.txt is fixed)
	if commandline == "" {
		commandline = "merge-sketch-with-bootloader.lua /tmp/sketch.hex && /usr/bin/run-avrdude /tmp/sketch.hex"
	}

	// Execute commandline
	output, err := client.Exec(commandline)
	info(l, string(output))
	debug(l, "Execute commandline ", commandline, string(output), err)
	if err != nil {
		return errors.Wrapf(err, "Execute commandline")
	}
	return nil
}

// scp uploads sourceFile to remote machine like native scp console app.
func scp(client *simplessh.Client, sourceFile, targetFile string) error {
	// open ssh session
	session, err := client.SSHClient.NewSession()
	if err != nil {
		return errors.Wrapf(err, "open ssh session")
	}
	defer session.Close()

	// open file
	src, err := os.Open(sourceFile)
	if err != nil {
		return errors.Wrapf(err, "open file %s", sourceFile)
	}

	// stat file
	srcStat, err := src.Stat()
	if err != nil {
		return errors.Wrapf(err, "stat file %s", sourceFile)
	}

	// Copy over ssh
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
		return errors.Wrapf(err, "Execute %s", "scp -t "+targetFile)
	}

	return nil
}
