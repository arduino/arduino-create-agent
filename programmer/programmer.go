package programmer

import (
	"bufio"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/arduino/arduino-create-agent/utilities"
	"github.com/facchinm/go-serial"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
)

type logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
}

func debug(l logger, args ...interface{}) {
	if l != nil {
		l.Debug(args...)
	}
}

func info(l logger, args ...interface{}) {
	if l != nil {
		l.Info(args...)
	}
}

// locater can return the location of a tool in the system
type locater interface {
	GetLocation(command string) (string, error)
}

// Auth contains username and password used for a network upload
type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
}

// Resolve replaces some symbols in the commandline with the appropriate values
// it can return an error when looking a variable in the locater
func Resolve(port, board, file, commandline string, extra Extra, t locater) (string, error) {
	commandline = strings.Replace(commandline, "{build.path}", filepath.ToSlash(filepath.Dir(file)), -1)
	commandline = strings.Replace(commandline, "{build.project_name}", strings.TrimSuffix(filepath.Base(file), filepath.Ext(filepath.Base(file))), -1)
	commandline = strings.Replace(commandline, "{serial.port}", port, -1)
	commandline = strings.Replace(commandline, "{serial.port.file}", filepath.Base(port), -1)

	if extra.Verbose == true {
		commandline = strings.Replace(commandline, "{upload.verbose}", extra.ParamsVerbose, -1)
	} else {
		commandline = strings.Replace(commandline, "{upload.verbose}", extra.ParamsQuiet, -1)
	}

	// search for runtime variables and replace with values from locater
	var runtimeRe = regexp.MustCompile("\\{(.*?)\\}")
	runtimeVars := runtimeRe.FindAllString(commandline, -1)

	for _, element := range runtimeVars {

		location, err := t.GetLocation(element)
		if err != nil {
			return "", errors.Wrapf(err, "get location of %s", element)
		}
		commandline = strings.Replace(commandline, element, location, 1)
	}

	return commandline, nil
}

// Do performs a command on a port with a board attached to it
func Do(port, commandline string, extra Extra, l logger) error {
	if extra.Network {
		doNetwork()
	} else {
		return doSerial(port, commandline, extra, l)
	}
	return nil
}

func doNetwork() {}

func doSerial(port, commandline string, extra Extra, l logger) error {
	// some boards needs to be resetted
	if extra.Use1200bpsTouch {
		var err error
		port, err = reset(port, extra.WaitForUploadPort, l)
		if err != nil {
			return errors.Wrapf(err, "Reset before upload")
		}
	}

	z, err := shellwords.Parse(commandline)
	if err != nil {
		return errors.Wrapf(err, "Parse commandline")
	}

	return program(z[0], z[1:], l)
}

// reset opens the port at 1200bps. It returns the new port name (which could change
// sometimes) and an error (usually because the port listing failed)
func reset(port string, wait bool, l logger) (string, error) {
	info(l, "Restarting in bootloader mode")

	// Get port list before reset
	ports, err := serial.GetPortsList()
	debug(l, "Get port list before reset")
	debug(l, ports, err)
	if err != nil {
		return "", errors.Wrapf(err, "Get port list before reset")
	}

	// Open port
	mode := &serial.Mode{
		BaudRate: 1200,
		Vmin:     0,
		Vtimeout: 1,
	}
	p, err := serial.OpenPort(port, mode)
	debug(l, "Open port", port)
	debug(l, p, err)
	if err != nil {
		return "", errors.Wrapf(err, "Open port %s", port)
	}

	// Set DTR
	err = p.SetDTR(false)
	debug(l, "Set DTR")
	debug(l, err)
	p.Close()

	// Wait for port to disappear and reappear
	if wait {
		port = waitReset(ports, l)
	}

	return port, nil
}

// waitReset is meant to be called just after a reset. It watches the ports connected
// to the machine until a port disappears and reappears. The port name could be different
// so it returns the name of the new port.
func waitReset(beforeReset []string, l logger) string {
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
		debug(l, "..", ports, beforeReset, err, port)

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
		ports, err := serial.GetPortsList()
		port = differ(ports, afterReset)
		debug(l, "..", ports, afterReset, err, port)
		if port != "" {
			time.Sleep(time.Millisecond * 500)
			break
		}
		if timeout {
			debug(l, "timeout")
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	return port
}

// program spawns the given binary with the given args, logging the sdtout and stderr
// through the logger
func program(binary string, args []string, l logger) error {
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

	oscmd := exec.Command(binary, args...)

	utilities.TellCommandNotToSpawnShell(oscmd)

	stdout, err := oscmd.StdoutPipe()
	if err != nil {
		return errors.Wrapf(err, "Retrieve output")
	}

	stderr, err := oscmd.StderrPipe()
	if err != nil {
		return errors.Wrapf(err, "Retrieve output")
	}

	info(l, "Flashing with command:"+binary+extension+" "+strings.Join(args, " "))

	err = oscmd.Start()

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

	err = oscmd.Wait()

	return errors.Wrapf(err, "Executing command")
}
