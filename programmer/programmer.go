package programmer

import (
	"time"

	"github.com/facchinm/go-serial"
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

// Do performs a command on a port with a board attached to it
func Do(port, board, file, commandline string, extra Extra, l logger) {
	debug(l, port, board, file, commandline)
	if extra.Network {
		doNetwork()
	} else {
		doSerial(port, board, file, commandline, extra, l)
	}
}

func doNetwork() {}

func doSerial(port, board, file, commandline string, extra Extra, l logger) error {
	// some boards needs to be resetted
	if extra.Use1200bpsTouch {
		var err error
		port, err = reset(port, extra.WaitForUploadPort, l)
		if err != nil {
			return errors.Wrapf(err, "Reset before upload")
		}
	}

	// resolve commandline
	info(l, "unresolved commandline ", commandline)
	commandline = resolve(port, board, file, commandline, extra)
	info(l, "resolved commandline ", commandline)

	return nil
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
