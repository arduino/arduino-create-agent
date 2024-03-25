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

package pkgs_test

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/arduino/arduino-create-agent/config"
	"github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/arduino/arduino-create-agent/index"
	"github.com/arduino/arduino-create-agent/v2/pkgs"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

// TestTools performs a series of operations about tools, ensuring it behaves as expected.
// This test depends on the internet so it could fail unexpectedly
func TestTools(t *testing.T) {
	// Initialize indexes with a temp folder
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	indexURL := "https://downloads.arduino.cc/packages/package_index.json"
	// Instantiate Index
	Index := index.Init(indexURL, config.GetDataDir())

	service := pkgs.New(Index, tmp)

	ctx := context.Background()

	// List available tools
	available, err := service.Available(ctx)
	if err != nil {
		t.Fatal(err)
	}
	require.NotEmpty(t, available)

	// Try to install a non-existent tool
	_, err = service.Install(ctx, &tools.ToolPayload{})
	if err == nil || !strings.Contains(err.Error(), "tool not found with packager '', name '', version ''") {
		t.Fatalf("expected '%v' == '%v' (%s)", err, "tool not found with packager '', name '', version ''", "err")
	}

	// Install a tool
	installed, err := service.Installed(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(installed) != 0 {
		t.Fatalf("expected %d == %d (%s)", len(installed), 0, "len(installed)")
	}

	_, err = service.Install(ctx, &tools.ToolPayload{
		Packager: "arduino",
		Name:     "avrdude",
		Version:  "6.0.1-arduino2",
	})
	if err != nil {
		t.Fatal(err)
	}
	// List installed tools
	installed, err = service.Installed(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(installed) != 1 {
		t.Fatalf("expected %d == %d (%s)", len(installed), 1, "len(installed)")
	}

	// Install the tool again
	_, err = service.Install(ctx, &tools.ToolPayload{
		Packager: "arduino",
		Name:     "avrdude",
		Version:  "6.0.1-arduino2",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Remove tool
	_, err = service.Remove(ctx, &tools.ToolPayload{
		Packager: "arduino",
		Name:     "avrdude",
		Version:  "6.0.1-arduino2",
	})
	if err != nil {
		t.Fatal(err)
	}

	installed, err = service.Installed(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(installed) != 0 {
		t.Fatalf("expected %d == %d (%s)", len(installed), 0, "len(installed)")
	}
}

func TestEvilFilename(t *testing.T) {

	// Initialize indexes with a temp folder
	tmp := t.TempDir()

	indexURL := "https://downloads.arduino.cc/packages/package_index.json"
	// Instantiate Index
	Index := index.Init(indexURL, config.GetDataDir())

	service := pkgs.New(Index, tmp)

	ctx := context.Background()

	type test struct {
		fileName string
		errBody  string
	}

	evilFileNames := []string{
		"/",
		"..",
		"../",
		"../evil.txt",
		"../../../../../../../../../../../../../../../../../../../../tmp/evil.txt",
		"some/path/../../../../../../../../../../../../../../../../../../../../tmp/evil.txt",
	}
	if runtime.GOOS == "windows" {
		evilFileNames = []string{
			"..\\",
			"..\\evil.txt",
			"..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\tmp\\evil.txt",
			"some\\path\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\tmp\\evil.txt",
		}
	}
	tests := []test{}
	for _, evilFileName := range evilFileNames {
		tests = append(tests, test{fileName: evilFileName,
			errBody: "unsafe path join"})
	}

	toolsTemplate := tools.ToolPayload{
		// We'll replace the name directly in the test
		Checksum:  strpoint("SHA-256:1ae54999c1f97234a5c603eb99ad39313b11746a4ca517269a9285afa05f9100"),
		Signature: strpoint("382898a97b5a86edd74208f10107d2fecbf7059ffe9cc856e045266fb4db4e98802728a0859cfdcda1c0b9075ec01e42dbea1f430b813530d5a6ae1766dfbba64c3e689b59758062dc2ab2e32b2a3491dc2b9a80b9cda4ae514fbe0ec5af210111b6896976053ab76bac55bcecfcececa68adfa3299e3cde6b7f117b3552a7d80ca419374bb497e3c3f12b640cf5b20875416b45e662fc6150b99b178f8e41d6982b4c0a255925ea39773683f9aa9201dc5768b6fc857c87ff602b6a93452a541b8ec10ca07f166e61a9e9d91f0a6090bd2038ed4427af6251039fb9fe8eb62ec30d7b0f3df38bc9de7204dec478fb86f8eb3f71543710790ee169dce039d3e0"),
		URL:       strpoint("http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux64.tar.gz"),
	}

	for _, test := range tests {
		t.Run("REMOVE payload containing evil names: "+test.fileName, func(t *testing.T) {
			// Here we could inject malicious name also in the Packager and Version field.
			// Since the path is made by joining all of these 3 fields, we're using only the Name,
			// as it won't change the result and let us keep the test small and readable.
			_, err := service.Remove(ctx, &tools.ToolPayload{Name: test.fileName})
			require.Error(t, err, test)
			require.ErrorContains(t, err, test.errBody)
		})
	}
	for _, test := range tests {
		toolsTemplate.Name = test.fileName
		t.Run("INSTALL payload containing evil names: "+toolsTemplate.Name, func(t *testing.T) {
			// Here we could inject malicious name also in the Packager and Version field.
			// Since the path is made by joining all of these 3 fields, we're using only the Name,
			// as it won't change the result and let us keep the test small and readable.
			_, err := service.Install(ctx, &toolsTemplate)
			require.Error(t, err, test)
			require.ErrorContains(t, err, test.errBody)
		})
	}
}

func TestInstalledHead(t *testing.T) {
	// Initialize indexes with a temp folder
	tmp := t.TempDir()

	indexURL := "https://downloads.arduino.cc/packages/package_index.json"
	// Instantiate Index
	Index := index.Init(indexURL, config.GetDataDir())

	service := pkgs.New(Index, tmp)

	ctx := context.Background()

	err := service.Installedhead(ctx)
	require.NoError(t, err)
}

func strpoint(s string) *string {
	return &s
}

func TestInstall(t *testing.T) {
	// Initialize indexes with a temp folder
	tmp := t.TempDir()

	testIndex := &index.Resource{
		IndexFile:   *paths.New("testdata", "test_tool_index.json"),
		LastRefresh: time.Now(),
	}

	tool := pkgs.New(testIndex, tmp)

	ctx := context.Background()

	testCases := []tools.ToolPayload{
		// https://github.com/arduino/arduino-create-agent/issues/920
		{Name: "avrdude", Version: "6.3.0-arduino17", Packager: "arduino-test", URL: nil, Checksum: nil, Signature: nil},
		{Name: "bossac", Version: "1.6.1-arduino", Packager: "arduino-test", URL: nil, Checksum: nil, Signature: nil},
		{Name: "bossac", Version: "1.7.0-arduino3", Packager: "arduino-test", URL: nil, Checksum: nil, Signature: nil},
		{Name: "bossac", Version: "1.9.1-arduino2", Packager: "arduino-test", URL: nil, Checksum: nil, Signature: nil},
		{Name: "openocd", Version: "0.11.0-arduino2", Packager: "arduino-test", URL: nil, Checksum: nil, Signature: nil},
		{Name: "dfu-util", Version: "0.10.0-arduino1", Packager: "arduino-test", URL: nil, Checksum: nil, Signature: nil},
		{Name: "rp2040tools", Version: "1.0.6", Packager: "arduino-test", URL: nil, Checksum: nil, Signature: nil},
		{Name: "esptool_py", Version: "4.5.1", Packager: "arduino-test", URL: nil, Checksum: nil, Signature: nil},
		// At the moment we don't install these ones because they are packaged in a different way: they do not have a top level dir, causing the rename funcion to behave incorrectly
		// {Name: "fwupdater", Version: "0.1.12", Packager: "arduino-test", URL: nil, Checksum: nil, Signature: nil},
		// {Name: "arduino-fwuploader", Version: "2.2.2", Packager: "arduino-test", URL: nil, Checksum: nil, Signature: nil},
	}

	expectedFiles := map[string][]string{
		"avrdude-6.3.0-arduino17":  {"bin", "etc"},
		"bossac-1.6.1-arduino":     {"bossac"},
		"bossac-1.7.0-arduino3":    {"bossac"},
		"bossac-1.9.1-arduino2":    {"bossac"},
		"openocd-0.11.0-arduino2":  {"bin", "share"},
		"dfu-util-0.10.0-arduino1": {"dfu-prefix", "dfu-suffix", "dfu-util"},
		"rp2040tools-1.0.6":        {"elf2uf2", "picotool", "pioasm", "rp2040load"},
		"esptool_py-4.5.1":         {"esptool"},
		// "fwupdater-0.1.12":         {"firmwares", "FirmwareUploader"}, // old legacy tool
		// "arduino-fwuploader-2.2.2": {"arduino-fwuploader"},
	}
	for _, tc := range testCases {
		t.Run(tc.Name+"-"+tc.Version, func(t *testing.T) {
			// Install the Tool
			_, err := tool.Install(ctx, &tc)
			require.NoError(t, err)

			// Check that the tool has been downloaded
			toolDir := paths.New(tmp).Join("arduino-test", tc.Name, tc.Version)
			require.DirExists(t, toolDir.String())

			// Check that the files have been created
			for _, file := range expectedFiles[tc.Name+"-"+tc.Version] {
				filePath := toolDir.Join(file)
				if filePath.IsDir() {
					require.DirExists(t, filePath.String())
				} else {
					if runtime.GOOS == "windows" {
						require.FileExists(t, filePath.String()+".exe")
					} else {
						require.FileExists(t, filePath.String())
					}
				}
			}
		})
	}

}
