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

	// Install a tool by specifying url and checksum
	_, err = service.Install(ctx, &tools.ToolPayload{
		Packager: "arduino",
		Name:     "avrdude",
		Version:  "6.0.1-arduino2",
		URL:      strpoint(url()),
		Checksum: strpoint(checksum()),
	})
	if err != nil {
		t.Fatal(err)
	}

	installed, err = service.Installed(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(installed) != 1 {
		t.Fatalf("expected %d == %d (%s)", len(installed), 1, "len(installed)")
	}

	t.Run("payload containing evil names", func(t *testing.T) {
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
		for _, evilFileName := range evilFileNames {
			// Here we could inject malicious name also in the Packager and Version field.
			// Since the path is made by joining all of these 3 fields, we're using only the Name,
			// as it won't change the result and let us keep the test small and readable.
			_, err := service.Remove(ctx, &tools.ToolPayload{Name: evilFileName})
			require.Error(t, err, evilFileName)
			require.ErrorContains(t, err, "unsafe path join")
		}
	})
}

func strpoint(s string) *string {
	return &s
}

func url() string {
	urls := map[string]string{
		"linuxamd64":   "https://downloads.arduino.cc/tools/avrdude-6.0.1-arduino2-x86_64-pc-linux-gnu.tar.bz2",
		"linux386":     "https://downloads.arduino.cc/tools/avrdude-6.0.1-arduino2-i686-pc-linux-gnu.tar.bz2",
		"darwinamd64":  "https://downloads.arduino.cc/tools/avrdude-6.0.1-arduino2-i386-apple-darwin11.tar.bz2",
		"windows386":   "https://downloads.arduino.cc/tools/avrdude-6.0.1-arduino2-i686-mingw32.zip",
		"windowsamd64": "https://downloads.arduino.cc/tools/avrdude-6.0.1-arduino2-i686-mingw32.zip",
	}

	return urls[runtime.GOOS+runtime.GOARCH]
}

func checksum() string {
	checksums := map[string]string{
		"linuxamd64":   "SHA-256:2489004d1d98177eaf69796760451f89224007c98b39ebb5577a9a34f51425f1",
		"linux386":     "SHA-256:6f633dd6270ad0d9ef19507bcbf8697b414a15208e4c0f71deec25ef89cdef3f",
		"darwinamd64":  "SHA-256:71117cce0096dad6c091e2c34eb0b9a3386d3aec7d863d2da733d9e5eac3a6b1",
		"windows386":   "SHA-256:6c5483800ba753c80893607e30cade8ab77b182808fcc5ea15fa3019c63d76ae",
		"windowsamd64": "SHA-256:6c5483800ba753c80893607e30cade8ab77b182808fcc5ea15fa3019c63d76ae",
	}
	return checksums[runtime.GOOS+runtime.GOARCH]

}
