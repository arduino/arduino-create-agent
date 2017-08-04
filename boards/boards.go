/*
 * This file is part of arduino-create-agent.
 *
 * arduino-create-agent is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */
package boards

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bcmi-labs/arduino-modules/fs"
	properties "github.com/dmotylev/goproperties"
	"github.com/juju/errors"
)

// Board is a physical board belonging to a certain architecture in a package.
// The most obvious package is arduino, which contains architectures avr, sam
// and samd
// It can contain multiple variants, but at least one that it's the default
type Board struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Vid            []string            `json:"vid"`
	Pid            []string            `json:"pid"`
	Package        string              `json:"package"`
	Architecture   string              `json:"architecture"`
	Fqbn           string              `json:"fqbn"`
	Variants       map[string]*Variant `json:"variants"`
	DefaultVariant string              `json:"default_variant"`
}

// Boards is a map of Boards
type Boards map[string]*Board

// Variant is a board that differ slightly from the others with the same model
type Variant struct {
	Name    string  `json:"name"`
	Fqbn    string  `json:"fqbn"`
	Actions Actions `json:"actions"`
}

// Action is a command that a tool can execute on a board
type Action struct {
	Tool        string    `json:"tool"`
	ToolVersion string    `json:"tool_version"`
	Ext         string    `json:"ext"`
	Command     string    `json:"command"`
	Params      options   `json:"params"`
	Options     options   `json:"options"`
	Files       []fs.File `json:"files,omitempty"`
}

// Actions is a map of Actions
type Actions map[string]*Action

// Lister contains methods to retrieve a slice of boards
type Lister interface {
	List()
}

// Retriever contains methods to retrieve a single board
type Retriever interface {
	ById()
	ByVidPid()
}

// Client parses the boards.txt, platform.txt and platform.json files to build a
// representation of the boards
type Client struct {
}

// New returns a new client by parsing the files contained in the given folder
func New() *Client {
	return nil
}

// ByID returns the board with the given id, or nil if it doesn't exists
func (list Boards) ByID(id string) *Board {
	return list[id]
}

// ByVidPid return the board with the correct combination of vid and pid, or nil
// if it doesn't exists
func (list Boards) ByVidPid(vid, pid string) *Board {
	for _, board := range list {
		if in(vid, board.Vid) && in(pid, board.Pid) {
			return board
		}
	}

	return nil
}

// New returns a new board with the given id
func (list Boards) New(id string) *Board {
	board := &Board{
		Fqbn:           id,
		Vid:            []string{},
		Pid:            []string{},
		DefaultVariant: "default",
		Variants: map[string]*Variant{
			"default": &Variant{
				Fqbn:    id,
				Actions: Actions{},
			},
		},
	}
	list[id] = board
	return board
}

// ParseBoardsTXT parses a bords.txt adding the boards to itself
func (list Boards) ParseBoardsTXT(path string) error {
	arch := filepath.Base(filepath.Dir(path))
	pack := filepath.Base(filepath.Dir(filepath.Dir(path)))

	props, err := properties.Load(path)
	if err != nil {
		return errors.Annotatef(err, "parse properties of %s", path)
	}

	menu := findMenu(props)

	temp := Boards{}

	// discover which boards are present
	for key, value := range props {
		parts := strings.Split(key, ".")

		// Discard menus
		if parts[0] == "menu" {
			continue
		}

		// The first part is always the id
		id := parts[0]
		if id == "" {
			continue
		}

		fqbn := pack + ":" + arch + ":" + id

		// Get or create the board
		var board *Board
		if board = temp.ByID(fqbn); board == nil {
			board = temp.New(fqbn)
			board.ID = id
			board.Package = pack
			board.Architecture = arch
		}

		if len(parts) < 2 {
			continue
		}

		// Populate fields
		populate(parts, board, menu, value)
	}

	// Upgrade the variants with the common options
	for fqbn, board := range temp {
		defVariant := board.Variants["default"]
		delete(board.Variants, "default")
		if len(board.Variants) == 0 {
			board.Variants["default"] = defVariant
		} else {
			for _, variant := range board.Variants {
				for name, action := range defVariant.Actions {
					populateAction(variant, name, "tool", action.Tool)
					for opt, value := range action.Options {
						populateAction(variant, name, opt, value)
					}
				}
			}
		}

		// Set the default variant
		normalize(board)

		// Append the board
		list[fqbn] = board
	}

	return nil
}

// Find parses all subfolders of a location, computing the results
func Find(location string) (Boards, error) {
	folders, err := ioutil.ReadDir(location)
	if err != nil {
		return nil, errors.Annotatef(err, "while reading the contents of folder %s", location)
	}

	list := Boards{}
	plats := Platforms{}

	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}
		files, err := ioutil.ReadDir(filepath.Join(location, folder.Name()))
		if err != nil {
			return nil, errors.Annotatef(err, "while reading the contents of folder %s", folder)
		}

		for _, file := range files {
			path := filepath.Join(location, folder.Name(), file.Name())

			if file.IsDir() {
				// Parse boards.txt
				list.ParseBoardsTXT(filepath.Join(path, "boards.txt"))

				// Parse platform.json
				f, _ := ioutil.ReadFile(filepath.Join(path, "platform.json"))
				var plat Platform
				json.Unmarshal(f, &plat)

				// Parse platform.txt
				plat.ParsePlatformTXT(filepath.Join(path, "platform.txt"))
				plats[plat.Architecture+":"+plat.Packager] = &plat
			}
		}
	}

	Compute(list, plats)

	return list, nil
}

// Compute fills the fields of the boards that need to be calculated from the platform info, such as the fully expanded commandline or the tools versions
func Compute(brds Boards, plats Platforms) {
	extRe := regexp.MustCompile(`{build.project_name}(\.bin|\.hex|\.bin)`)

	for _, board := range brds {
		for _, variant := range board.Variants {
			for name, action := range variant.Actions {
				tool := plats.Tool(board.Package, board.Architecture, action.Tool)
				if tool == nil {
					continue
				}

				if tool.Patterns[name] == nil {
					continue
				}

				action.ToolVersion = tool.Version

				// Find the expected extension
				action.Command = expand(tool, variant, name)

				// Arduino Zero Debug port is strange, so it's handled as a special case
				if variant.Fqbn == "arduino:samd:arduino_zero_edbg" && name == "upload" {
					action.Command = `"{runtime.tools.openocd.path}/bin/openocd" {upload.verbose} -s "{runtime.tools.openocd.path}/share/openocd/scripts/" -f "{build.path}/arduino_zero.cfg" -c "telnet_port disabled; program {build.path}/{build.project_name}.bin verify reset 0x00002000; shutdown"`
				}

				// Arduino M0 Debug port is strange, so it's handled as a special case
				if variant.Fqbn == "arduino:samd:mzero_pro_bl_dbg" && name == "upload" {
					action.Tool = "openocd"
					action.Command = `"{runtime.tools.openocd.path}/bin/openocd" {upload.verbose} -s "{runtime.tools.openocd.path}/share/openocd/scripts/" -f "{build.path}/arduino_zero.cfg" -c "telnet_port disabled; program {build.path}/{build.project_name}.bin verify reset 0x00004000; shutdown"`
				}

				match := extRe.FindString(action.Command)
				action.Ext = filepath.Ext(match)
				for param, value := range tool.Patterns[name].Params {
					action.Params[param] = value
				}

				// Find the files to include
				findFiles(action, plats[board.Architecture+":"+board.Package])
			}
		}
	}

}

func findFiles(action *Action, plat *Platform) {
	action.Files = []fs.File{}
	re := regexp.MustCompile(`{runtime.platform.path}[\w\/\.\-]*`)

	action.Command = re.ReplaceAllStringFunc(action.Command, func(file string) string {
		filename := strings.Replace(file, `{runtime.platform.path}`, plat.Path, -1)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return file
		}

		filename = filepath.Base(filename)

		action.Files = append(action.Files, fs.File{Name: filename, Data: data})

		return "{runtime.platform.path}/" + filename
	})

}

func expand(tool *Tool, variant *Variant, pattern string) string {
	re := regexp.MustCompile(`{([\S{}]*?)}`)

	if tool.Patterns[pattern] == nil {
		return ""
	}

	command := tool.Patterns[pattern].Command
	oldCommand := ""

	for command != oldCommand {
		oldCommand = command
		command = re.ReplaceAllStringFunc(command, replace(tool, variant, pattern))
	}
	return command
}

func replace(tool *Tool, variant *Variant, pattern string) func(string) string {
	return func(value string) string {
		// Remove parenthesis
		key := strings.Replace(value, "{", "", 1)
		key = strings.Replace(key, "}", "", 1)

		// Update path
		if key == "path" {
			return tool.Path
		}

		// Search in tool options
		if prop, ok := tool.Options[key]; ok {
			return prop
		}

		for name, action := range variant.Actions {
			actionkey := strings.Replace(key, name+".", "", 1)
			if prop, ok := action.Options[actionkey]; ok {
				return prop
			}
		}

		return value
	}
}
