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

package browser

import (
	"os/exec"
	"strings"

	"github.com/arduino/arduino-create-agent/utilities"
)

// Find will find the browser
func Find(process string) ([]byte, error) {
	ps := exec.Command("ps", "-A", "-o", "command")
	grep := exec.Command("grep", process)
	head := exec.Command("head", "-n", "1")

	return utilities.PipeCommands(ps, grep, head)
}

// Kill will kill a process
func Kill(process string) ([]byte, error) {
	cmd := exec.Command("pkill", "-9", process)
	return cmd.Output()
}

// Start will start a command
func Start(command []byte, url string) ([]byte, error) {
	parts := strings.Split(string(command), " ")
	cmd := exec.Command(parts[0], url)
	return cmd.Output()
}
