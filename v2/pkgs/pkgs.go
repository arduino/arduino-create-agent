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

// Package pkgs implements the functions from
// github.com/arduino-create-agent/gen/indexes
// and github.com/arduino-create-agent/gen/tools.
//
// It allows to manage package indexes from arduino
// cores, and to download tools used for upload.
package pkgs

// Index is the go representation of a typical
// package-index file, stripped from every non-used field.
type Index struct {
	Packages []struct {
		Name  string `json:"name"`
		Tools []Tool `json:"tools"`
	} `json:"packages"`
}

// Tool is the go representation of the info about a
//tool contained in a package-index file, stripped from
//every non-used field.
type Tool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Systems []struct {
		Host     string `json:"host"`
		URL      string `json:"url"`
		Checksum string `json:"checksum"`
	} `json:"systems"`
}
