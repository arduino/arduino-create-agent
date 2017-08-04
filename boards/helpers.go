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
	"strings"

	properties "github.com/dmotylev/goproperties"
)

type options map[string]string

// Merge returns an options object where the opts override the options
func (opts1 options) Merge(opts2 options) options {
	opts := options{}

	for key, opt := range opts1 {
		opts[key] = opt
	}

	for key, opt := range opts2 {
		opts[key] = opt
	}
	return opts
}

// findMenu returns the menu property of the boards.txt
func findMenu(props properties.Properties) string {
	for key := range props {
		parts := strings.Split(key, ".")

		// handle menu (there should be only one of this anyway)
		if parts[0] == "menu" {
			return parts[1]
		}
	}
	return "cpu"
}

func populate(parts []string, board *Board, menu, value string) {
	if parts[1] == "name" {
		board.Name = value
		return
	}

	if parts[1] == "vid" && len(parts) == 3 {
		board.Vid = append(board.Vid, value)
		return
	}

	if parts[1] == "pid" && len(parts) == 3 {
		board.Pid = append(board.Pid, value)
		return
	}

	// Populate variants
	if parts[1] == "menu" && len(parts) >= 4 {
		populateVariants(parts, board, menu, value)
		return
	}

	// Populate common patterns
	// These options will be appended to every bootloader variant at the end
	if len(parts) >= 3 {
		populateAction(board.Variants["default"], parts[1], strings.Join(parts[2:], "."), value)
		return
	}
}

func populateAction(variant *Variant, name, option, value string) {
	var action *Action
	var ok bool
	if action, ok = variant.Actions[name]; !ok {
		action = &Action{Options: options{}, Params: options{}}
		variant.Actions[name] = action
	}

	if option == "tool" {
		action.Tool = value
	}
	if action.Options[option] == "" {
		action.Options[option] = value
	}
}

func populateVariants(parts []string, board *Board, menu, value string) {
	name := parts[3]
	var variant *Variant
	var ok bool
	if variant, ok = board.Variants[name]; !ok {
		variant = &Variant{
			Fqbn:    board.Fqbn + ":" + menu + "=" + name,
			Actions: Actions{},
		}
		board.Variants[name] = variant
	}

	if len(parts) == 4 {
		variant.Name = value
		return
	}

	populateAction(variant, parts[4], parts[5], value)
}

// Len is the number of elements in the collection.
func (t Tools) Len() int {
	return len(t)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (t Tools) Less(i, j int) bool {
	return t[i].Name < t[j].Name
}

// Swap swaps the elements with indexes i and j.
func (t Tools) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// normalize sets one of the version as default. Usually the latest
func normalize(b *Board) {
	if len(b.Variants) < 2 {
		return
	}

	switch b.ID {
	case "atmegang":
		b.DefaultVariant = "atmega168"
	case "mega":
		b.DefaultVariant = "atmega2560"
	case "pro":
		b.DefaultVariant = "16MHzatmega328"
	case "bt", "diecimila", "lilypad", "mini", "nano":
		b.DefaultVariant = "atmega328"
	}
}

func in(a string, list []string) bool {
	a = strings.ToLower(a)
	for _, b := range list {
		b = strings.ToLower(b)
		if b == a {
			return true
		}
	}
	return false
}
