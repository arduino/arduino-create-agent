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
	"log"
	"strings"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

type mockTools struct{}

func (mockTools) GetLocation(el string) (string, error) {
	return "$loc" + el, nil
}

// TestSerialData requires a leonardo connected to the /dev/ttyACM0 port
var TestSerialData = []struct {
	Name        string
	Port        string
	Commandline string
	Extra       Extra
}{
	{
		"leonardo", "/dev/ttyACM0",
		`"$HOME/.arduino-create/avrdude/6.3.0-arduino6/bin/avrdude" "-C$HOME/.arduino-create/avrdude/6.3.0-arduino6/etc/avrdude.conf" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:./upload_test.hex:i"`, Extra{Use1200bpsTouch: true, WaitForUploadPort: true}},
}

func TestSerial(t *testing.T) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	home, _ := homedir.Dir()

	for _, test := range TestSerialData {
		commandline := strings.Replace(test.Commandline, "$HOME", home, -1)
		err := Serial(test.Port, commandline, test.Extra, logger)
		log.Println(err)
	}
}

var TestResolveData = []struct {
	Board        string
	File         string
	PlatformPath string
	CommandLine  string
	Extra        Extra
	Result       string
}{
	{
		Board:        "arduino:avr:leonardo",
		File:         "./upload_test.hex",
		PlatformPath: "",
		CommandLine:  `{runtime.tools.avrdude.path}/bin/avrdude -C{runtime.tools.avrdude.path}/etc/avrdude.conf -v {upload.verify} -patmega32u4 -cavr109 -P{serial.port} -b57600 -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"`,
		Extra:        Extra{Use1200bpsTouch: true, WaitForUploadPort: true},
		Result:       `$loc$loc{runtime.tools.avrdude.path}/bin/avrdude -C{runtime.tools.avrdude.path}/etc/avrdude.conf -v $loc{upload.verify} -patmega32u4 -cavr109 -P$loc{serial.port} -b57600 -D "-Uflash:w:./upload_test.hex:i"`,
	},
	{
		Board:        "arduino:renesas_uno:unor4wifi",
		File:         "UpdateFirmware.bin",
		PlatformPath: "",
		CommandLine:  `{runtime.tools.arduino-fwuploader.path}/arduino-fwuploader firmware flash -a {serial.port} -b {fqbn} -v --retries 5"`,
		Extra:        Extra{Use1200bpsTouch: true, WaitForUploadPort: true},
		Result:       `$loc{runtime.tools.arduino-fwuploader.path}/arduino-fwuploader firmware flash -a $loc{serial.port} -b arduino:renesas_uno:unor4wifi -v --retries 5"`,
	},
}

func TestResolve(t *testing.T) {
	for _, test := range TestResolveData {
		result, _ := PartiallyResolve(test.Board, test.File, test.PlatformPath, test.CommandLine, test.Extra, mockTools{})
		if result != test.Result {
			t.Error("expected " + test.Result + ", got " + result)
			continue
		}
	}
}

var TestFixupData = []struct {
	Port        string
	CommandLine string
	Result      string
}{
	{
		Port:        "/dev/ttyACM0",
		CommandLine: `{runtime.tools.avrdude.path}/bin/avrdude -C{runtime.tools.avrdude.path}/etc/avrdude.conf -v {upload.verify} -patmega32u4 -cavr109 -P{serial.port} -b57600 -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"`,
		Result:      `{runtime.tools.avrdude.path}/bin/avrdude -C{runtime.tools.avrdude.path}/etc/avrdude.conf -v {upload.verify} -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"`,
	},
	{
		Port:        "/dev/cu.usbmodemDC5475C5557C2",
		CommandLine: `{runtime.tools.arduino-fwuploader.path}/arduino-fwuploader firmware flash -a {serial.port} -b arduino:renesas_uno:unor4wifi -v --retries 5"`,
		Result:      `{runtime.tools.arduino-fwuploader.path}/arduino-fwuploader firmware flash -a /dev/cu.usbmodemDC5475C5557C2 -b arduino:renesas_uno:unor4wifi -v --retries 5"`,
	},
}

func TestFixupPort(t *testing.T) {
	for _, test := range TestFixupData {
		result := fixupPort(test.Port, test.CommandLine)
		if result != test.Result {
			t.Error("expected " + test.Result + ", got " + result)
			continue
		}
	}
}
