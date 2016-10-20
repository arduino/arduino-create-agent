package programmer_test

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/arduino/arduino-create-agent/programmer"
	"github.com/stretchr/testify/suite"
)

type ProgrammerTestSuite struct {
	suite.Suite
}

func TestProgrammer(t *testing.T) {
	suite.Run(t, new(ProgrammerTestSuite))
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

func (t *ProgrammerTestSuite) TestSerial() {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	for _, test := range TestSerialData {
		programmer.Do(test.Port, test.Board, test.File, test.Commandline, test.Extra, logger)
	}
	t.Fail("")
}
