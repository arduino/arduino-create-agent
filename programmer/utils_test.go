package programmer

import (
	"log"
	"testing"
)

var TestDifferData = []struct {
	Before []string
	After  []string
	Result string
}{
	{[]string{}, []string{}, ""},
	{[]string{"/dev/ttyACM0", "/dev/sys"}, []string{"/dev/sys"}, "/dev/ttyACM0"},
	{[]string{"/dev/sys"}, []string{"/dev/ttyACM0", "/dev/sys"}, "/dev/ttyACM0"},
}

func TestDiffer(t *testing.T) {
	for _, test := range TestDifferData {
		result := differ(test.Before, test.After)
		if result != test.Result {
			t.Error("expected " + test.Result + ", got " + result)
			continue
		}
	}
}

var TestResolveData = []struct {
	Port        string
	Board       string
	File        string
	Commandline string
	Extra       Extra
	Result      string
}{
	{"/dev/ttyACM0", "arduino:avr:leonardo", "./programmer_test.hex",
		`"{runtime.tools.avrdude.path}/bin/avrdude" "-C{runtime.tools.avrdude.path}/etc/avrdude.conf" {upload.verbose} {upload.verify} -patmega32u4 -cavr109 -P{serial.port} -b57600 -D "-Uflash:w:{build.path}/{build.project_name}.hex:i"`, Extra{Use1200bpsTouch: true, WaitForUploadPort: true},
		`"{runtime.tools.avrdude.path}/bin/avrdude" "-C{runtime.tools.avrdude.path}/etc/avrdude.conf"  {upload.verify} -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:./programmer_test.hex:i"`},
}

func TestResolve(t *testing.T) {
	for _, test := range TestResolveData {
		result := resolve(test.Port, test.Board, test.File, test.Commandline, test.Extra)
		if result != test.Result {
			log.Println(result)
			log.Println(test.Result)
			t.Error("expected " + test.Result + ", got " + result)
			continue
		}
	}
}
