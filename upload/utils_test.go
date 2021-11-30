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
