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
package main

import (
	"flag"
	"log"
	"regexp"
	"strings"

	"github.com/arduino/arduino-create-agent/boards"
	. "github.com/dave/jennifer/jen"
)

func main() {
	var folder = flag.String("folder", "/opt/cores", "Location of the arduino cores")
	var target = flag.String("target", "boards.go", "Location of the generated file")
	flag.Parse()

	// Retrieve boards
	boards, err := boards.Find(*folder)
	if err != nil {
		log.Fatal(err)
	}

	// Generate boards.go
	// log.Println(boards)

	boardsVar := Var().Id("boards").Op("=").Map(String()).Qual("github.com/arduino/arduino-create-agent/exec", "Command")

	values := []Code{}

	re := regexp.MustCompile(`({[.\w-]+})`)

	for _, board := range boards {
		for _, variant := range board.Variants {
			for name, action := range variant.Actions {
				if name != "upload" || action.Command == "" {
					continue // We must draw the line somewhere...
				}
				params := re.FindAllString(action.Command, -1)

				paramValues := []Code{}
				for _, param := range params {
					// Avoid runtime
					if strings.Contains(param, "runtime") {
						continue
					}
					paramValues = append(paramValues, Lit(param))
				}
				value := Lit(name+":"+variant.Fqbn).Op(":").Qual("github.com/arduino/arduino-create-agent/exec", "Command").Values(Dict{
					Id("Pattern"): Lit(action.Command),
					Id("Params"):  Index().String().Values(paramValues...),
				})
				values = append(values, value)
			}

		}
	}

	boardsVar = boardsVar.Values(values...)

	f := NewFile("main")

	f.Add(boardsVar)

	err = f.Save(*target)
	if err != nil {
		log.Fatal(err)
	}
}
