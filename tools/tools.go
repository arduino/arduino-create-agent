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
	"path/filepath"
	"strings"
	"sync"

	"github.com/arduino/arduino-create-agent/index"
	"github.com/arduino/go-paths-helper"
	"github.com/xrash/smetrics"
)

// Tools handle the tools necessary for an upload on a board.
// It provides a means to download a tool from the arduino servers.
//
// - *directory* contains the location where the tools are downloaded.
// - *indexURL* contains the url where the tools description is contained.
// - *logger* is a StdLogger used for reporting debug and info messages
// - *installed* contains a map[string]string of the tools installed and their exact location
//
// Usage:
// You have to call the New() function passing it the required parameters:
//
// 	index = index.Init("https://downloads.arduino.cc/packages/package_index.json", dataDir)
// 	tools := tools.New(dataDir, index, logger)

// Tools will represent the installed tools
type Tools struct {
	directory *paths.Path
	index     *index.Resource
	logger    func(msg string)
	installed map[string]string
	mutex     sync.RWMutex
}

// New will return a Tool object, allowing the caller to execute operations on it.
// The New functions accept the directory to use to host the tools,
// an index (used to download the tools),
// and a logger to log the operations
func New(directory *paths.Path, index *index.Resource, logger func(msg string)) *Tools {
	t := &Tools{
		directory: directory,
		index:     index,
		logger:    logger,
		installed: map[string]string{},
		mutex:     sync.RWMutex{},
	}
	_ = t.readMap()
	return t
}

func (t *Tools) setMapValue(key, value string) {
	t.mutex.Lock()
	t.installed[key] = value
	t.mutex.Unlock()
}

func (t *Tools) getMapValue(key string) (string, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	value, ok := t.installed[key]
	return value, ok
}

// readMap() reads the installed map from json file "installed.json"
func (t *Tools) readMap() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	filePath := t.directory.Join("installed.json")
	b, err := filePath.ReadFile()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &t.installed)
}

// GetLocation extracts the toolname from a command like
func (t *Tools) GetLocation(command string) (string, error) {
	command = strings.Replace(command, "{runtime.tools.", "", 1)
	command = strings.Replace(command, ".path}", "", 1)

	var location string
	var ok bool

	// Load installed
	err := t.readMap()
	if err != nil {
		return "", err
	}

	// use string similarity to resolve a runtime var with a "similar" map element
	if location, ok = t.getMapValue(command); !ok {
		maxSimilarity := 0.0
		t.mutex.RLock()
		defer t.mutex.RUnlock()
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
