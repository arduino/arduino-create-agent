package tools

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
type Tools struct {
	Directory   string
	IndexURL    string
	LastRefresh time.Time
	Logger      func(msg string)
	installed   map[string]string
	mutex       sync.RWMutex
}

// Init creates the Installed map and populates it from a file in .arduino-create
func (t *Tools) Init(APIlevel string) {
	createDir(t.Directory)
	t.mutex.Lock()
	t.installed = make(map[string]string)
	t.mutex.Unlock()
	t.readMap()
	t.mutex.RLock()
	if t.installed["apilevel"] != APIlevel {
		t.mutex.RUnlock()
		// wipe the folder and reinitialize the data
		os.RemoveAll(t.Directory)
		createDir(t.Directory)
		t.mutex.Lock()
		t.installed = make(map[string]string)
		t.installed["apilevel"] = APIlevel
		t.mutex.Unlock()
		t.writeMap()
		t.readMap()
	} else {
		t.mutex.RUnlock()
	}
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
	t.mutex.RLock()
	b, err := json.Marshal(t.installed)
	t.mutex.RUnlock()
	if err != nil {
		return err
	}
	filePath := path.Join(dir(), "installed.json")
	return ioutil.WriteFile(filePath, b, 0644)
}

// readMap() reads the installed map from json file "installed.json"
func (t *Tools) readMap() error {
	filePath := path.Join(dir(), "installed.json")
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return json.Unmarshal(b, &t.installed)
}

func dir() string {
	usr, _ := user.Current()
	return path.Join(usr.HomeDir, ".arduino-create")
}

// createDir creates the directory where the tools will be stored
func createDir(directory string) {
	os.Mkdir(directory, 0777)
	hideFile(directory)
}
