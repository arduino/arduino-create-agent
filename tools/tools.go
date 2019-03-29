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
//         IndexURL: "http://downloads.arduino.cc/packages/package_index.json"
//         Logger: log.Logger
//     }
type Tools struct {
	Directory   string
	IndexURL    string
	LastRefresh time.Time
	Logger      func(msg string)
	installed   map[string]string
}

// Init creates the Installed map and populates it from a file in .arduino-create
func (t *Tools) Init(APIlevel string) {
	createDir(t.Directory)
	t.installed = make(map[string]string)
	t.readMap()
	if t.installed["apilevel"] != APIlevel {
		// wipe the folder and reinitialize the data
		os.RemoveAll(t.Directory)
		createDir(t.Directory)
		t.installed = make(map[string]string)
		t.installed["apilevel"] = APIlevel
		t.writeMap()
		t.readMap()
	}
}

// GetLocation extracts the toolname from a command like
func (t *Tools) GetLocation(command string) (string, error) {
	command = strings.Replace(command, "{runtime.tools.", "", 1)
	command = strings.Replace(command, ".path}", "", 1)

	var location string
	var ok bool

	// Load installed
	fmt.Println(t.installed)

	err := t.readMap()
	if err != nil {
		return "", err
	}

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

func (t *Tools) writeMap() error {
	b, err := json.Marshal(t.installed)
	if err != nil {
		return err
	}
	filePath := path.Join(dir(), "installed.json")
	return ioutil.WriteFile(filePath, b, 0644)
}

func (t *Tools) readMap() error {
	filePath := path.Join(dir(), "installed.json")
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
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
