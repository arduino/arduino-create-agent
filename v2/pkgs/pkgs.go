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
