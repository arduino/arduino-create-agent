// Copyright 2022 Arduino SA
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package upload

import (
	"bufio"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-create-agent/utilities"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
	"go.bug.st/serial/enumerator"
)

// Busy tells wether the programmer is doing something
var Busy = false

// Extra contains some options used during the upload
type Extra struct {
	Use1200bpsTouch   bool `json:"use_1200bps_touch"`
	WaitForUploadPort bool `json:"wait_for_upload_port"`
	Network           bool `json:"network"`
}

// PartiallyResolve replaces some symbols in the commandline with the appropriate values
// it can return an error when looking a variable in the Locater
func PartiallyResolve(board, file, platformPath, commandline string, extra Extra, t Locater) (string, error) {
	commandline = strings.Replace(commandline, "{build.path}", filepath.ToSlash(filepath.Dir(file)), -1)
	commandline = strings.Replace(commandline, "{build.project_name}", strings.TrimSuffix(filepath.Base(file), filepath.Ext(filepath.Base(file))), -1)
	commandline = strings.Replace(commandline, "{runtime.platform.path}", filepath.ToSlash(platformPath), -1)

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

// reset wraps arduino-cli's serialutils
// it opens the port at 1200bps. It returns the new port name (which could change
// sometimes) and an error (usually because the port listing failed)
func reset(port string, wait bool, l Logger) (string, error) {
	info(l, "Restarting in bootloader mode")
	newPort, err := serialutils.Reset(port, wait, nil, false) // TODO use callbacks to print as the cli does
	if err != nil {
		info(l, err)
		return "", err
	}
	if newPort != "" {
		port = newPort
	}
	return port, nil
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
