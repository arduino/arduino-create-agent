package upload

import "testing"

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
