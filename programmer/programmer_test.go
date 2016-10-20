package programmer_test

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/arduino/arduino-create-agent/programmer"
)

type mockTools struct{}

func (mockTools) GetLocation(el string) (string, error) {
	return "$loc" + el, nil
}

var TestSerialData = []struct {
	Name        string
	Port        string
	Board       string
	File        string
	Commandline string
	Extra       programmer.Extra
}{
	{"leonardo", "/dev/ttyACM0", "arduino:avr:leonardo", "./programmer_test.hex",
		`"{runtime.tools.avrdude.path}/bin/avrdude" "-C{runtime.tools.avrdude.path}/etc/avrdude.conf" {upload.verbose} {upload.verify} -patmega32u4 -cavr109 -P{serial.port} -b57600 -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"`, programmer.Extra{Use1200bpsTouch: true, WaitForUploadPort: true}},
}

func TestSerial(t *testing.T) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	for _, test := range TestSerialData {
		programmer.Do(test.Port, test.Board, test.File, test.Commandline, test.Extra, mockTools{}, logger)
	}
	t.Fail()
}

var TestResolveData = []struct {
	Port        string
	Board       string
	File        string
	Commandline string
	Extra       programmer.Extra
	Result      string
}{
	{"/dev/ttyACM0", "arduino:avr:leonardo", "./programmer_test.hex",
		`"{runtime.tools.avrdude.path}/bin/avrdude" "-C{runtime.tools.avrdude.path}/etc/avrdude.conf" {upload.verbose} {upload.verify} -patmega32u4 -cavr109 -P{serial.port} -b57600 -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"`, programmer.Extra{Use1200bpsTouch: true, WaitForUploadPort: true},
		`"$loc$loc{runtime.tools.avrdude.path}/bin/avrdude" "-C{runtime.tools.avrdude.path}/etc/avrdude.conf"  $loc{upload.verify} -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:./programmer_test.hex:i"`},
}

func TestResolve(t *testing.T) {
	for _, test := range TestResolveData {
		result, _ := programmer.Resolve(test.Port, test.Board, test.File, test.Commandline, test.Extra, mockTools{})
		if result != test.Result {
			t.Error("expected " + test.Result + ", got " + result)
			continue
		}
	}
}
