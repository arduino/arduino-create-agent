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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arduino/arduino-create-agent/gen/indexes"
	"github.com/arduino/arduino-create-agent/v2/pkgs"
)

// TestIndexes performs a series of operations about indexes, ensuring it behaves as expected.
func TestIndexes(t *testing.T) {
	// Use local file as index
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/package_index.json")
	}))
	defer ts.Close()

	// Initialize indexes with a temp folder
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	// Create extraneous folder in temp folder
	os.MkdirAll(filepath.Join(tmp, "arduino"), 0755)

	service := pkgs.Indexes{
		Folder: tmp,
	}

	ctx := context.Background()

	// List indexes, they should be 0
	list, err := service.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected %d == %d (%s)", len(list), 0, "len(list)")
	}

	// Add a faulty index
	_, err = service.Add(ctx, &indexes.IndexPayload{URL: ":"})
	if err == nil || !strings.Contains(err.Error(), "missing protocol scheme") {
		t.Fatalf("expected [%v] == [%v] (%s)", err, "missing protocol scheme", "err")
	}

	// Add a new index
	_, err = service.Add(ctx, &indexes.IndexPayload{URL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}

	// List indexes, they should be 1
	list, err = service.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected %d == %d (%s)", len(list), 1, "len(list)")
	}
	if list[0] != ts.URL {
		t.Fatalf("expected %s == %s (%s)", list[0], "downloads.arduino.cc/packages/package_index.json", "list[0]")
	}

	// Remove the index
	_, err = service.Remove(ctx, &indexes.IndexPayload{URL: ts.URL})
	if err != nil {
		t.Fatal(err)
	}

	// List indexes, they should be 0
	list, err = service.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected %d == %d (%s)", len(list), 0, "len(list)")
	}
}
