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
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */
package boards

import (
	"path/filepath"
	"strings"

	properties "github.com/dmotylev/goproperties"
	"github.com/juju/errors"
)

// Pattern is a commandline that can be expanded with various options and params
type Pattern struct {
	Command string  `json:"command"`
	Params  options `json:"params"`
}

// Patterns is a map of Patterns
type Patterns map[string]*Pattern

// Tool is a program that can program boards with different patterns
type Tool struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Packager string   `json:"packager"`
	Path     string   `json:"path"`
	Patterns Patterns `json:"patterns"`
	Options  options  `json:"options"`
}

// Tools is a slice of Tools
type Tools []*Tool

// Platform is a single platform in a package
type Platform struct {
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	Architecture string  `json:"architecture"`
	Packager     string  `json:"packager"`
	Version      string  `json:"version"`
	Options      options `json:"options"`
	Tools        Tools   `json:"toolsDependencies"`
}

// Platforms is a map of Platforms
type Platforms map[string]*Platform

// Tool searches for a tool with the given package, architecture and name
func (p Platforms) Tool(pack, arch, name string) *Tool {
	parts := strings.Split(name, ":")
	if len(parts) == 2 {
		pack = parts[0]
		name = parts[1]
	}

	for _, plat := range p {
		if plat.Architecture != arch || plat.Packager != pack {
			continue
		}
		for _, tool := range plat.Tools {
			if pack != "" && pack != tool.Packager {
				continue
			}
			if tool.Name == name {
				return tool
			}
		}
	}
	return nil
}

// Tool searches for a tool with the given name
func (p *Platform) Tool(name string) *Tool {
	for _, tool := range p.Tools {
		if tool.Name == name {
			return tool
		}
	}

	tool := Tool{Name: name, Options: options{}, Patterns: Patterns{}}
	p.Tools = append(p.Tools, &tool)

	return &tool
}

// ParsePlatformTXT parses a platform.txt
func (p *Platform) ParsePlatformTXT(path string) error {
	p.Path = filepath.Dir(path)
	p.Architecture = filepath.Base(filepath.Dir(path))
	p.Packager = filepath.Base(filepath.Dir(filepath.Dir(path)))

	props, err := properties.Load(path)
	if err != nil {
		return errors.Annotatef(err, "parse properties of platforms.txt %s", path)
	}

	if p.Options == nil {
		p.Options = options{}
	}
	if p.Tools == nil {
		p.Tools = Tools{}
	}

	for key, value := range props {
		if key == "version" {
			p.Version = value
			continue
		}
		if key == "name" {
			p.Name = value
			continue
		}

		parts := strings.Split(key, ".")
		if parts[0] == "tools" && len(parts) > 2 {
			name := parts[1]
			tool := p.Tool(name)

			if tool.Patterns == nil {
				tool.Patterns = Patterns{}
			}
			if tool.Options == nil {
				tool.Options = options{}
			}

			if len(parts) == 3 && parts[2] == "path" {
				tool.Path = value
				continue
			}

			if len(parts) == 4 && parts[3] == "pattern" {
				var pattern *Pattern
				var ok bool
				name := parts[2]
				if pattern, ok = tool.Patterns[name]; !ok {
					pattern = &Pattern{Params: options{}}
					tool.Patterns[name] = pattern
				}

				pattern.Command = value
				continue
			}

			if len(parts) == 5 && parts[3] == "params" {
				var pattern *Pattern
				var ok bool
				name := parts[2]
				if pattern, ok = tool.Patterns[name]; !ok {
					pattern = &Pattern{Params: options{}}
					tool.Patterns[name] = pattern
				}

				pattern.Params[parts[4]] = value
				continue
			}

			if len(parts) > 2 {
				prop := strings.Join(parts[2:], ".")
				tool.Options[prop] = value
				continue
			}
		}

		p.Options[key] = value
	}
	return nil
}
