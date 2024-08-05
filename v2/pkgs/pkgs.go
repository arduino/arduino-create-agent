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

import "regexp"

// Index is the go representation of a typical
// package-index file, stripped from every non-used field.
type Index struct {
	Packages []struct {
		Name  string `json:"name"`
		Tools []Tool `json:"tools"`
	} `json:"packages"`
}

// Tool is the go representation of the info about a
// tool contained in a package-index file, stripped from
// every non-used field.
type Tool struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Systems []System `json:"systems"`
}

// System is the go representation of the info needed to
// download a tool for a specific OS/Arch
type System struct {
	Host     string `json:"host"`
	URL      string `json:"url"`
	Name     string `json:"archiveFileName"`
	Checksum string `json:"checksum"`
}

// Source: https://github.com/arduino/arduino-cli/blob/master/internal/arduino/cores/tools.go#L129-L142
var (
	regexpLinuxArm   = regexp.MustCompile("arm.*-linux-gnueabihf")
	regexpLinuxArm64 = regexp.MustCompile("(aarch64|arm64)-linux-gnu")
	regexpLinux64    = regexp.MustCompile("x86_64-.*linux-gnu")
	regexpLinux32    = regexp.MustCompile("i[3456]86-.*linux-gnu")
	regexpWindows32  = regexp.MustCompile("i[3456]86-.*(mingw32|cygwin)")
	regexpWindows64  = regexp.MustCompile("(amd64|x86_64)-.*(mingw32|cygwin)")
	regexpMac64      = regexp.MustCompile("x86_64-apple-darwin.*")
	regexpMac32      = regexp.MustCompile("i[3456]86-apple-darwin.*")
	regexpMacArm64   = regexp.MustCompile("arm64-apple-darwin.*")
	regexpFreeBSDArm = regexp.MustCompile("arm.*-freebsd[0-9]*")
	regexpFreeBSD32  = regexp.MustCompile("i?[3456]86-freebsd[0-9]*")
	regexpFreeBSD64  = regexp.MustCompile("amd64-freebsd[0-9]*")
)

// Source: https://github.com/arduino/arduino-cli/blob/master/internal/arduino/cores/tools.go#L144-L176
func (s *System) isExactMatchWith(osName, osArch string) bool {
	if s.Host == "all" {
		return true
	}

	switch osName + "," + osArch {
	case "linux,arm", "linux,armbe":
		return regexpLinuxArm.MatchString(s.Host)
	case "linux,arm64":
		return regexpLinuxArm64.MatchString(s.Host)
	case "linux,amd64":
		return regexpLinux64.MatchString(s.Host)
	case "linux,386":
		return regexpLinux32.MatchString(s.Host)
	case "windows,386":
		return regexpWindows32.MatchString(s.Host)
	case "windows,amd64":
		return regexpWindows64.MatchString(s.Host)
	case "darwin,arm64":
		return regexpMacArm64.MatchString(s.Host)
	case "darwin,amd64":
		return regexpMac64.MatchString(s.Host)
	case "darwin,386":
		return regexpMac32.MatchString(s.Host)
	case "freebsd,arm":
		return regexpFreeBSDArm.MatchString(s.Host)
	case "freebsd,386":
		return regexpFreeBSD32.MatchString(s.Host)
	case "freebsd,amd64":
		return regexpFreeBSD64.MatchString(s.Host)
	}
	return false
}

// Source: https://github.com/arduino/arduino-cli/blob/master/internal/arduino/cores/tools.go#L178-L198
func (s *System) isCompatibleWith(osName, osArch string) (bool, int) {
	if s.isExactMatchWith(osName, osArch) {
		return true, 1000
	}

	switch osName + "," + osArch {
	case "windows,amd64":
		return regexpWindows32.MatchString(s.Host), 10
	case "darwin,amd64":
		return regexpMac32.MatchString(s.Host), 10
	case "darwin,arm64":
		// Compatibility guaranteed through Rosetta emulation
		if regexpMac64.MatchString(s.Host) {
			// Prefer amd64 version if available
			return true, 20
		}
		return regexpMac32.MatchString(s.Host), 10
	}

	return false, 0
}

// GetFlavourCompatibleWith returns the downloadable resource (System) compatible with the specified OS/Arch
// Source: https://github.com/arduino/arduino-cli/blob/master/internal/arduino/cores/tools.go#L206-L216
func (t *Tool) GetFlavourCompatibleWith(osName, osArch string) System {
	var correctSystem System
	maxSimilarity := -1

	for _, s := range t.Systems {
		if comp, similarity := s.isCompatibleWith(osName, osArch); comp && similarity > maxSimilarity {
			correctSystem = s
			maxSimilarity = similarity
		}
	}

	return correctSystem
}
