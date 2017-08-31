/*
 * This file is part of arduino-create-agent.
 *
 * arduino-create-agent is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */
// Package exec is just syntactic sugar over os/exec. Allows to execute predefined
// parametric commands in a safe way
//
// Usage:
// 	cmd := exec.Command{
// 		Pattern: "echo {interpolate.string}",
// 		Params:  []string{"interpolate.string"},
// 	}
// 	opts := map[string]string{"interpolate.string": "hello world"}
// 	stdout, stderr, err := exec.Local(cmd, opts)
// 	fmt.Println(err) // nil
// 	out, _ := ioutil.ReadAll(stdout)
// 	stdout.Close()
// 	stderr.Close()
// 	fmt.Println(string(out)) // hello world
package exec

import (
	"io"
	"os/exec"
	"strings"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
)

// Interpolate substitutes the params with the corresponding options
func Interpolate(pattern string, opts map[string]string) (interpolated []string, err error) {
	for key, value := range opts {
		pattern = strings.Replace(pattern, "{"+key+"}", value, -1)
	}

	z, err := shellwords.Parse(pattern)
	if err != nil {
		return nil, err
	}
	return z, nil
}

// Local executes a command on the local machine, interpolating the command with the
// given options
func Local(pattern string, opts map[string]string) (stdout, stderr io.ReadCloser, err error) {
	// interpolate
	inter, err := Interpolate(pattern, opts)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "interpolate")
	}

	// create command
	executable, args := inter[0], inter[1:]
	cmd := exec.Command(executable, args...)

	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "retrieve output")
	}

	stderr, err = cmd.StderrPipe()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "retrieve output")
	}

	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}

	return stdout, stderr, nil
}
