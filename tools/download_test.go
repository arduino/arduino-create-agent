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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
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
		{"windows", "amd64", "i686-mingw32"},
		{"linux", "arm", "arm-linux-gnueabihf"},
	}
	testIndex := paths.New("testdata", "test_tool_index.json")
	buf, err := testIndex.ReadFile()
	require.NoError(t, err)

	var data index
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

func Test_findBaseDir(t *testing.T) {
	cases := []struct {
		dirList []string
		want    string
	}{
		{[]string{"bin/bossac"}, "bin/"},
		{[]string{"bin/", "bin/bossac"}, "bin/"},
		{[]string{"bin/", "bin/bossac", "example"}, ""},
		{[]string{"avrdude/bin/",
			"avrdude/bin/avrdude.exe",
			"avrdude/bin/remove_giveio.bat",
			"avrdude/bin/status_giveio.bat",
			"avrdude/bin/giveio.sys",
			"avrdude/bin/loaddrv.exe",
			"avrdude/bin/libusb0.dll",
			"avrdude/bin/install_giveio.bat",
			"avrdude/etc/avrdude.conf"}, "avrdude/"},
		{[]string{"pax_global_header", "bin/", "bin/bossac"}, "bin/"},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintln(tt.dirList), func(t *testing.T) {
			if got := findBaseDir(tt.dirList); got != tt.want {
				t.Errorf("findBaseDir() = got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTools_DownloadAndUnpackBehaviour(t *testing.T) {
	urls := []string{
		"https://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-armhf-pc-linux-gnu.tar.bz2",
		"https://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-aarch64-pc-linux-gnu.tar.bz2",
		"https://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-i386-apple-darwin11.tar.bz2",
		"https://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-x86_64-pc-linux-gnu.tar.bz2",
		"https://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-i686-pc-linux-gnu.tar.bz2",
		"https://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-i686-w64-mingw32.zip",
	}
	expectedDirList := []string{"bin", "etc"}

	tmpDir, err := ioutil.TempDir("", "download_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	for _, url := range urls {
		t.Log("Downloading tool from " + url)
		resp, err := http.Get(url)
		if err != nil {
			t.Errorf("%v", err)
		}
		defer resp.Body.Close()

		// Read the body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%v", err)
		}

		location := path.Join(tmpDir, "username", "arduino", "avrdude", "6.3.0-arduino14")
		os.MkdirAll(location, os.ModePerm)
		err = os.RemoveAll(location)

		if err != nil {
			t.Errorf("%v", err)
		}

		srcType, err := mimeType(body)
		if err != nil {
			t.Errorf("%v", err)
		}

		switch srcType {
		case "application/zip":
			location, err = extractZip(func(msg string) { t.Log(msg) }, body, location)
		case "application/x-bz2":
		case "application/octet-stream":
			location, err = extractBz2(func(msg string) { t.Log(msg) }, body, location)
		case "application/x-gzip":
			location, err = extractTarGz(func(msg string) { t.Log(msg) }, body, location)
		default:
			t.Errorf("no suitable type found")
		}
		files, err := ioutil.ReadDir(location)
		if err != nil {
			t.Errorf("%v", err)
		}
		dirList := []string{}
		for _, f := range files {
			dirList = append(dirList, f.Name())
		}

		assert.ElementsMatchf(t, dirList, expectedDirList, "error message %s", "formatted")
	}

}
