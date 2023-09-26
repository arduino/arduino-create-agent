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

package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/arduino/arduino-create-agent/index"
	"github.com/xrash/smetrics"
)

// Tools handle the tools necessary for an upload on a board.
// It provides a means to download a tool from the arduino servers.
//
// - *Directory* contains the location where the tools are downloaded.
// - *IndexURL* contains the url where the tools description is contained.
// - *Logger* is a StdLogger used for reporting debug and info messages
// - *installed* contains a map of the tools and their exact location
//
// Usage:
// You have to instantiate the struct by passing it the required parameters:
//     _tools := tools.Tools{
//         Directory: "/home/user/.arduino-create",
//         IndexURL: "https://downloads.arduino.cc/packages/package_index.json"
//         Logger: log.Logger
//     }

// Tools will represent the installed tools
type Tools struct {
	Directory string
	Index     *index.Resource
	Logger    func(msg string)
	installed map[string]string
	mutex     sync.RWMutex
}

// Init creates the Installed map and populates it from a file in .arduino-create
func (t *Tools) Init() {
	t.mutex.Lock()
	t.installed = make(map[string]string)
	t.mutex.Unlock()
	t.readMap()
}

// GetLocation extracts the toolname from a command like
func (t *Tools) GetLocation(command string) (string, error) {
	command = strings.Replace(command, "{runtime.tools.", "", 1)
	command = strings.Replace(command, ".path}", "", 1)

	var location string
	var ok bool

	// Load installed
	t.mutex.RLock()
	fmt.Println(t.installed)
	t.mutex.RUnlock()

	err := t.readMap()
	if err != nil {
		return "", err
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()
	fmt.Println(t.installed)

	// use string similarity to resolve a runtime var with a "similar" map element
	if location, ok = t.installed[command]; !ok {
		maxSimilarity := 0.0
		for i, candidate := range t.installed {
			similarity := smetrics.Jaro(command, i)
			if similarity > 0.8 && similarity > maxSimilarity {
				maxSimilarity = similarity
				location = candidate
			}
		}
	}
	return filepath.ToSlash(location), nil
}

// writeMap() writes installed map to the json file "installed.json"
func (t *Tools) writeMap() error {
	t.mutex.Lock()
	b, err := json.Marshal(t.installed)
	defer t.mutex.Unlock()
	if err != nil {
		return err
	}
	filePath := path.Join(dir(), "installed.json")
	return os.WriteFile(filePath, b, 0644)
}

// readMap() reads the installed map from json file "installed.json"
func (t *Tools) readMap() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	filePath := path.Join(dir(), "installed.json")
	b, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &t.installed)
}

func dir() string {
	usr, _ := user.Current()
	return path.Join(usr.HomeDir, ".arduino-create")
}
