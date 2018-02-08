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
package tools_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-create-agent/tools"
)

func TestDownload(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(tmp)

	opts := tools.Opts{
		Location: tmp,
	}

	tool := tools.Tool{
		Name:     "avrdude",
		Version:  "6.0.1-arduino2",
		Packager: "arduino",
	}

	err := tool.Download("http://downloads.arduino.cc/tools/avrdude-6.0.1-arduino2-x86_64-pc-linux-gnu.tar.bz2", "SHA-256:2489004d1d98177eaf69796760451f89224007c98b39ebb5577a9a34f51425f1", &opts)
	if err != nil {
		t.Error(err.Error())
	}

	_, err = os.Open(filepath.Join(tool.Path, "bin", "avrdude"))
	if err != nil {
		t.Error(err.Error())
	}
}
func TestInstalled(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(tmp)

	opts := tools.Opts{
		Location: tmp,
	}

	list, err := tools.Installed(&opts)
	if err != nil {
		t.Error(err.Error())
	}

	if len(list) != 0 {
		t.Error("Expected len(list) to be 0, got", len(list))
	}

	tool := tools.Tool{
		Name:     "avrdude",
		Version:  "6.0.1-arduino2",
		Packager: "arduino",
	}

	err = tool.Download("http://downloads.arduino.cc/tools/avrdude-6.0.1-arduino2-x86_64-pc-linux-gnu.tar.bz2", "SHA-256:2489004d1d98177eaf69796760451f89224007c98b39ebb5577a9a34f51425f1", &opts)
	if err != nil {
		t.Error(err.Error())
	}

	list, err = tools.Installed(&opts)
	if err != nil {
		t.Error(err.Error())
	}

	if len(list) != 1 {
		t.Error("Expected len(list) to be 1, got", len(list))
	}
}
