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

package upload_test

import (
	"log"
	"strings"
	"testing"

	"github.com/arduino/arduino-create-agent/upload"
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
	Extra       upload.Extra
}{
	{
		"leonardo", "/dev/ttyACM0",
		`"$HOME/.arduino-create/avrdude/6.3.0-arduino6/bin/avrdude" "-C$HOME/.arduino-create/avrdude/6.3.0-arduino6/etc/avrdude.conf" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:./upload_test.hex:i"`, upload.Extra{Use1200bpsTouch: true, WaitForUploadPort: true}},
}

func TestSerial(t *testing.T) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	home, _ := homedir.Dir()

	for _, test := range TestSerialData {
		commandline := strings.Replace(test.Commandline, "$HOME", home, -1)
		err := upload.Serial(test.Port, commandline, test.Extra, logger)
		log.Println(err)
	}
}

var TestResolveData = []struct {
	Board        string
	File         string
	PlatformPath string
	Commandline  string
	Extra        upload.Extra
	Result       string
}{
	{"arduino:avr:leonardo", "./upload_test.hex", "",
		`"{runtime.tools.avrdude.path}/bin/avrdude" "-C{runtime.tools.avrdude.path}/etc/avrdude.conf" -v {upload.verify} -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"`, upload.Extra{Use1200bpsTouch: true, WaitForUploadPort: true},
		`"$loc$loc{runtime.tools.avrdude.path}/bin/avrdude" "-C{runtime.tools.avrdude.path}/etc/avrdude.conf" -v $loc{upload.verify} -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:./upload_test.hex:i"`},
}

func TestResolve(t *testing.T) {
	for _, test := range TestResolveData {
		result, _ := upload.PartiallyResolve(test.Board, test.File, test.PlatformPath, test.Commandline, test.Extra, mockTools{})
		if result != test.Result {
			t.Error("expected " + test.Result + ", got " + result)
			continue
		}
	}
}
