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

// Logger is an interface implemented by most loggers (like logrus)
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
}

func debug(l Logger, args ...interface{}) {
	if l != nil {
		l.Debug(args...)
	}
}

func info(l Logger, args ...interface{}) {
	if l != nil {
		l.Info(args...)
	}
}

// Locater can return the location of a tool in the system
type Locater interface {
	GetLocation(command string) (string, error)
}

// differ returns the first item that differ between the two input slices
func differ(slice1 []string, slice2 []string) string {
	m := map[string]int{}

	for _, s1Val := range slice1 {
		m[s1Val] = 1
	}
	for _, s2Val := range slice2 {
		m[s2Val] = m[s2Val] + 1
	}

	for mKey, mVal := range m {
		if mVal == 1 {
			return mKey
		}
	}

	return ""
}
