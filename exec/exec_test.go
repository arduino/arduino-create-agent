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
package exec_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/arduino/arduino-create-agent/exec"
)

func Example() {
	opts := map[string]string{"interpolate.string": "hello world"}
	stdout, stderr, err := exec.Local("echo {interpolate.string}", opts)
	fmt.Println(err) // nil
	out, _ := ioutil.ReadAll(stdout)
	stdout.Close()
	stderr.Close()
	fmt.Println(string(out)) // hello world
	// Output:
	// <nil>
	// hello world
}

func TestLocal(t *testing.T) {
	cases := []struct {
		ID        string
		Pattern   string
		Options   map[string]string
		ExpStdout string
		ExpStderr string
		ExpErr    error
	}{
		{
			"command not found",
			"foo",
			nil,
			"",
			"",
			errors.New(`exec: "foo": executable file not found in $PATH`),
		},
		{
			"error",
			"foo '",
			nil,
			"",
			"",
			errors.New("interpolate: invalid command line string"),
		},
		{
			"hello world",
			"echo {interpolate.string}",
			map[string]string{
				"interpolate.string": "hello world",
				"ignored":            "ignored",
			},
			"hello world",
			"",
			nil,
		},
		{
			"inject ; to perform other commands has no effect",
			"echo {interpolate.string}",
			map[string]string{
				"interpolate.string": "hello world; ls",
				"ignored":            "ignored",
			},
			"hello world",
			"",
			nil,
		},
		{
			"bash expansion do not work",
			"echo {interpolate.string}",
			map[string]string{
				"interpolate.string": "`date`",
				"ignored":            "ignored",
			},
			"`date`",
			"",
			nil,
		},
		{
			"env vars do not work",
			"echo {interpolate.string}",
			map[string]string{
				"interpolate.string": "$USER",
				"ignored":            "ignored",
			},
			"$USER",
			"",
			nil,
		},
		{
			"return stderr",
			"ls /notexist",
			map[string]string{
				"interpolate.string": "$USER",
				"ignored":            "ignored",
			},
			"",
			"ls: cannot access '/notexist': No such file or directory",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s", tc.ID), func(t *testing.T) {
			stdout, stderr, e := exec.Local(tc.Pattern, tc.Options)
			if !errEq(e, tc.ExpErr) {
				t.Errorf("expected e to be '%s', was '%s'", tc.ExpErr, e)
				return
			}
			if e != nil {
				return
			}
			if stdout == nil || stderr == nil {
				t.Errorf("stdout and stderr should not be nil")
				return
			}
			out, _ := ioutil.ReadAll(stdout)
			stdout.Close()

			if strings.TrimSpace(string(out)) != tc.ExpStdout {
				t.Errorf("expected stdout to be '%s', was '%s'", tc.ExpStdout, out)
			}
			err, _ := ioutil.ReadAll(stderr)
			stderr.Close()
			if strings.TrimSpace(string(err)) != tc.ExpStderr {
				t.Errorf("expected stderr to be '%s', was '%s'", tc.ExpStderr, err)
			}
		})
	}
}

func errEq(err1, err2 error) bool {
	return fmt.Sprintf("%s", err1) == fmt.Sprintf("%s", err2)
}
