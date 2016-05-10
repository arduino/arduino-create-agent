package tools

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/arduino/arduino-create-agent/utilities"
	"github.com/pivotal-golang/archiver/extractor"
)

type system struct {
	Host     string `json:"host"`
	URL      string `json:"url"`
	Name     string `json:"archiveFileName"`
	CheckSum string `json:"checksum"`
}

type tool struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Systems     []system `json:"systems"`
	url         string
	destination string
}

type index struct {
	Packages []struct {
		Name  string `json:"name"`
		Tools []tool `json:"tools"`
	} `json:"packages"`
}

var systems = map[string]string{
	"linuxamd64":  "x86_64-linux-gnu",
	"linux386":    "i686-linux-gnu",
	"darwinamd64": "i386-apple-darwin11",
	"windows386":  "i686-mingw32",
}

// Download will parse the index at the indexURL for the tool to download.
// It will extract it in a folder in .arduino-create, and it will update the
// Installed map.
//
// name contains the name of the tool.
// version contains the version of the tool.
// behaviour contains the strategy to use when there is already a tool installed
//
// If version is "latest" it will always download the latest version (regardless
// of the value of behaviour)
//
// If version is not "latest" and behaviour is "replace", it will download the
// version again. If instead behaviour is "keep" it will not download the version
// if it already exists.
func (t *Tools) Download(name, version, behaviour string) error {
	var key string
	if version == "latest" {
		key = name
	} else {
		key = name + "-" + version
	}

	// Check if it already exists
	if version != "latest" && behaviour == "keep" {
		if _, ok := t.installed[key]; ok {
			t.Logger.Println("The tool is already present on the system")
			return nil
		}
	}
	// Fetch the index
	resp, err := http.Get(t.IndexURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var data index
	json.Unmarshal(body, &data)

	// Find the tool by name
	correctTool := findTool(name, version, data)

	if correctTool.Name == "" {
		return errors.New("We couldn't find a tool with the name " + name + " and version " + version)
	}

	// Find the url based on system
	var correctSystem system

	for _, s := range correctTool.Systems {
		if s.Host == systems[runtime.GOOS+runtime.GOARCH] {
			correctSystem = s
		}
	}

	// Download the tool
	t.Logger.Println("Downloading tool " + name + " from " + correctSystem.URL)
	resp, err = http.Get(correctSystem.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the body
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Checksum
	checksum := sha256.Sum256(body)
	checkSumString := "SHA-256:" + hex.EncodeToString(checksum[:sha256.Size])

	if checkSumString != correctSystem.CheckSum {
		return errors.New("Checksum doesn't match")
	}

	// Decompress
	t.Logger.Println("Unpacking tool " + name)

	location := path.Join(dir(), name, version)
	err = os.RemoveAll(location)

	if err != nil {
		return err
	}

	switch path.Ext(correctSystem.URL) {
	case ".zip":
		location, err = extractZip(body, location)
	case ".bz2":
		location, err = extractBz2(body, location)
	default:
		return errors.New("Unknown extension for file " + correctSystem.URL)
	}

	if err != nil {
		return err
	}

	// Ensure that the files are executable
	t.Logger.Println("Ensure that the files are executable")
	err = makeExecutable(location)
	if err != nil {
		return err
	}

	// Update the tool map
	t.Logger.Println("Updating map with location " + location)

	t.installed[key] = location
	return t.writeMap()
}

func findTool(name, version string, data index) tool {
	var correctTool tool

	for _, p := range data.Packages {
		for _, t := range p.Tools {
			if version != "latest" {
				if t.Name == name && t.Version == version {
					correctTool = t
				}
			} else {
				// Find latest
				if t.Name == name && t.Version > correctTool.Version {
					correctTool = t
				}
			}
		}
	}
	return correctTool
}

func extractZip(body []byte, location string) (string, error) {
	path, err := utilities.SaveFileonTempDir("tooldownloaded.zip", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	e := extractor.NewZip()
	err = e.Extract(path, location)
	if err != nil {
		return "", err
	}
	return "", nil
}

func extractBz2(body []byte, location string) (string, error) {
	tarFile := bzip2.NewReader(bytes.NewReader(body))
	tarReader := tar.NewReader(tarFile)

	var subfolder string

	i := 0
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}

		filePath := path.Join(location, header.Name)

		// We get the name of the subfolder
		if i == 0 {
			subfolder = filePath
		}

		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(filePath, os.FileMode(header.Mode))
		case tar.TypeReg:
			f, err := os.Create(filePath)
			if err != nil {
				break
			}
			defer f.Close()
			_, err = io.Copy(f, tarReader)
		case tar.TypeRegA:
			f, err := os.Create(filePath)
			if err != nil {
				break
			}
			defer f.Close()
			_, err = io.Copy(f, tarReader)
		case tar.TypeSymlink:
			err = os.Symlink(header.Linkname, filePath)
		default:
			err = errors.New("Unknown header in tar.bz2 file")
		}

		if err != nil {
			return "", err
		}

		i++
	}
	return subfolder, nil
}

func makeExecutable(location string) error {
	location = path.Join(location, "bin")
	files, err := ioutil.ReadDir(location)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = os.Chmod(path.Join(location, file.Name()), 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
