package pkgs_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arduino/arduino-create-agent/gen/indexes"
	"github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/arduino/arduino-create-agent/v2/pkgs"
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
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(tmp)
	// defer os.RemoveAll(tmp)

	indexesClient := pkgs.Indexes{
		Folder: tmp,
	}

	service := pkgs.Tools{
		Folder:  tmp,
		Indexes: &indexesClient,
	}

	ctx := context.Background()

	// Add a new index
	err = indexesClient.Add(ctx, &indexes.IndexPayload{URL: ts.URL})
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
	err = service.Install(ctx, &tools.ToolPayload{})
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

	err = service.Install(ctx, &tools.ToolPayload{
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

	// Remove tool
	err = service.Remove(ctx, &tools.ToolPayload{
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
