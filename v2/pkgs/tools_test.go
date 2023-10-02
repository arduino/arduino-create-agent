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
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/arduino/arduino-create-agent/gen/indexes"
	"github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/arduino/arduino-create-agent/v2/pkgs"
	"github.com/stretchr/testify/require"
)

// TestTools performs a series of operations about tools, ensuring it behaves as expected.
// This test depends on the internet so it could fail unexpectedly
func TestTools(t *testing.T) {
	// Use local file as index
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/package_index.json")
	}))
	defer ts.Close()

	// Initialize indexes with a temp folder
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	indexesClient := pkgs.Indexes{
		Folder: tmp,
	}

	service := pkgs.Tools{
		Folder:  tmp,
		Indexes: &indexesClient,
	}

	ctx := context.Background()

	// Add a new index
	_, err = indexesClient.Add(ctx, &indexes.IndexPayload{URL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}

	// List available tools
	available, err := service.Available(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(available) != 61 {
		t.Fatalf("expected %d == %d (%s)", len(available), 61, "len(available)")
	}

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

	service := pkgs.Tools{
		Folder: tmp,
		Indexes: &pkgs.Indexes{
			Folder: tmp,
		},
	}

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

func strpoint(s string) *string {
	return &s
}
