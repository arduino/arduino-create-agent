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
	"runtime"
	"testing"
	"time"

	"github.com/arduino/arduino-create-agent/index"
	"github.com/arduino/arduino-create-agent/v2/pkgs"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestDownloadCorrectPlatform(t *testing.T) {
	testCases := []struct {
		hostOS        string
		hostArch      string
		correctOSArch string
	}{
		{"linux", "amd64", "x86_64-linux-gnu"},
		{"linux", "386", "i686-linux-gnu"},
		{"darwin", "amd64", "x86_64-apple-darwin"},
		{"darwin", "arm64", "arm64-apple-darwin"},
		{"windows", "386", "i686-mingw32"},
		{"windows", "amd64", "x86_64-mingw32"},
		{"linux", "arm", "arm-linux-gnueabihf"},
	}
	defer func() {
		OS = runtime.GOOS     // restore `runtime.OS`
		Arch = runtime.GOARCH // restore `runtime.ARCH`
	}()
	testIndex := paths.New("testdata", "test_tool_index.json")
	buf, err := testIndex.ReadFile()
	require.NoError(t, err)

	var data pkgs.Index
	err = json.Unmarshal(buf, &data)
	require.NoError(t, err)
	for _, tc := range testCases {
		t.Run(tc.hostOS+tc.hostArch, func(t *testing.T) {
			OS = tc.hostOS     // override `runtime.OS` for testing purposes
			Arch = tc.hostArch // override `runtime.ARCH` for testing purposes
			// Find the tool by name
			correctTool, correctSystem := findTool("arduino-test", "arduino-fwuploader", "2.2.2", data)
			require.NotNil(t, correctTool)
			require.NotNil(t, correctSystem)
			require.Equal(t, correctTool.Name, "arduino-fwuploader")
			require.Equal(t, correctTool.Version, "2.2.2")
			require.Equal(t, correctSystem.Host, tc.correctOSArch)
		})
	}
}

func TestDownloadFallbackPlatform(t *testing.T) {
	testCases := []struct {
		hostOS        string
		hostArch      string
		correctOSArch string
	}{
		{"darwin", "amd64", "i386-apple-darwin11"},
		{"darwin", "arm64", "i386-apple-darwin11"},
		{"windows", "amd64", "i686-mingw32"},
	}
	defer func() {
		OS = runtime.GOOS     // restore `runtime.OS`
		Arch = runtime.GOARCH // restore `runtime.ARCH`
	}()
	testIndex := paths.New("testdata", "test_tool_index.json")
	buf, err := testIndex.ReadFile()
	require.NoError(t, err)

	var data pkgs.Index
	err = json.Unmarshal(buf, &data)
	require.NoError(t, err)
	for _, tc := range testCases {
		t.Run(tc.hostOS+tc.hostArch, func(t *testing.T) {
			OS = tc.hostOS     // override `runtime.OS` for testing purposes
			Arch = tc.hostArch // override `runtime.ARCH` for testing purposes
			// Find the tool by name
			correctTool, correctSystem := findTool("arduino-test", "arduino-fwuploader", "2.2.0", data)
			require.NotNil(t, correctTool)
			require.NotNil(t, correctSystem)
			require.Equal(t, correctTool.Name, "arduino-fwuploader")
			require.Equal(t, correctTool.Version, "2.2.0")
			require.Equal(t, correctSystem.Host, tc.correctOSArch)
		})
	}
}

func TestDownload(t *testing.T) {
	testCases := []struct {
		name         string
		version      string
		filesCreated []string
	}{
		{"avrdude", "6.3.0-arduino17", []string{"bin", "etc"}},
		{"bossac", "1.6.1-arduino", []string{"bossac"}},
		{"bossac", "1.7.0-arduino3", []string{"bossac"}},
		{"bossac", "1.9.1-arduino2", []string{"bossac"}},
		{"openocd", "0.11.0-arduino2", []string{"bin", "share"}},
		{"dfu-util", "0.10.0-arduino1", []string{"dfu-prefix", "dfu-suffix", "dfu-util"}},
		{"rp2040tools", "1.0.6", []string{"elf2uf2", "picotool", "pioasm", "rp2040load"}},
		{"esptool_py", "4.5.1", []string{"esptool"}},
		{"arduino-fwuploader", "2.2.2", []string{"arduino-fwuploader"}},
		{"fwupdater", "0.1.12", []string{"firmwares", "FirmwareUploader"}}, // old legacy tool
	}
	// prepare the test environment
	tempDir := t.TempDir()
	tempDirPath := paths.New(tempDir)
	testIndex := index.Resource{
		IndexFile:   *paths.New("testdata", "test_tool_index.json"),
		LastRefresh: time.Now(),
	}
	testTools := New(tempDirPath, &testIndex, func(msg string) { t.Log(msg) })

	for _, tc := range testCases {
		t.Run(tc.name+"-"+tc.version, func(t *testing.T) {
			// Download the tool
			err := testTools.Download("arduino-test", tc.name, tc.version, "replace")
			require.NoError(t, err)

			// Check that the tool has been downloaded
			toolDir := tempDirPath.Join("arduino-test", tc.name, tc.version)
			require.DirExists(t, toolDir.String())

			// Check that the files have been created
			for _, file := range tc.filesCreated {
				filePath := toolDir.Join(file)
				if filePath.IsDir() {
					require.DirExists(t, filePath.String())
				} else {
					if OS == "windows" {
						require.FileExists(t, filePath.String()+".exe")
					} else {
						require.FileExists(t, filePath.String())
					}
				}
			}

			// Check that the tool has been installed
			_, ok := testTools.getMapValue(tc.name + "-" + tc.version)
			require.True(t, ok)
		})
	}
}
