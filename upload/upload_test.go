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

var TestNetworkData = []struct {
	Name        string
	Port        string
	Board       string
	File        string
	Commandline string
	Auth        upload.Auth
}{
	{
		"yun", "", "", "",
		``, upload.Auth{}},
}

func TestNetwork(t *testing.T) {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	home, _ := homedir.Dir()

	for _, test := range TestNetworkData {
		commandline := strings.Replace(test.Commandline, "$HOME", home, -1)
		err := upload.Network(test.Port, test.Board, test.File, commandline, test.Auth, logger)
		log.Println(err)
	}
}

var TestResolveData = []struct {
	Port        string
	Board       string
	File        string
	Commandline string
	Extra       upload.Extra
	Result      string
}{
	{"/dev/ttyACM0", "arduino:avr:leonardo", "./upload_test.hex",
		`"{runtime.tools.avrdude.path}/bin/avrdude" "-C{runtime.tools.avrdude.path}/etc/avrdude.conf" {upload.verbose} {upload.verify} -patmega32u4 -cavr109 -P{serial.port} -b57600 -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"`, upload.Extra{Use1200bpsTouch: true, WaitForUploadPort: true},
		`"$loc$loc{runtime.tools.avrdude.path}/bin/avrdude" "-C{runtime.tools.avrdude.path}/etc/avrdude.conf"  $loc{upload.verify} -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:./upload_test.hex:i"`},
}

func TestResolve(t *testing.T) {
	for _, test := range TestResolveData {
		result, _ := upload.Resolve(test.Port, test.Board, test.File, test.Commandline, test.Extra, mockTools{})
		if result != test.Result {
			t.Error("expected " + test.Result + ", got " + result)
			continue
		}
	}
}
