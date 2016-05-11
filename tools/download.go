package tools

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/arduino/arduino-create-agent/utilities"
	"github.com/blang/semver"
	"github.com/xrash/smetrics"
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

func mimeType(data []byte) (string, error) {
	return http.DetectContentType(data[0:512]), nil
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

	t.Logger.Println(string(body))

	// Find the tool by name
	correctTool := findTool(name, version, data)

	if correctTool.Name == "" {
		return errors.New("We couldn't find a tool with the name " + name + " and version " + version)
	}

	// Find the url based on system
	var correctSystem system
	max_similarity := 0.8

	for _, s := range correctTool.Systems {
		similarity := smetrics.Jaro(s.Host, systems[runtime.GOOS+runtime.GOARCH])
		if similarity > max_similarity {
			correctSystem = s
			max_similarity = similarity
		}
	}

	key := correctTool.Name + "-" + correctTool.Version

	// Check if it already exists
	if behaviour == "keep" {
		if _, ok := t.installed[key]; ok {
			t.Logger.Println("The tool is already present on the system")
			return nil
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

	location := path.Join(dir(), correctTool.Name, correctTool.Version)
	err = os.RemoveAll(location)

	if err != nil {
		return err
	}

	srcType, err := mimeType(body)
	if err != nil {
		return err
	}

	switch srcType {
	case "application/zip":
		location, err = extractZip(body, location)
	case "application/x-bz2":
	case "application/octet-stream":
		location, err = extractBz2(body, location)
	case "application/x-gzip":
		location, err = extractTarGz(body, location)
	default:
		return errors.New("Unknown extension for file " + correctSystem.URL)
	}

	if err != nil {
		return err
	}

	t.installDrivers(location)

	// Ensure that the files are executable
	t.Logger.Println("Ensure that the files are executable")

	// Update the tool map
	t.Logger.Println("Updating map with location " + location)

	t.installed[name] = location
	t.installed[name+"-"+correctTool.Version] = location
	return t.writeMap()
}

func findTool(name, version string, data index) tool {
	var correctTool tool
	correctTool.Version = "0.0"

	for _, p := range data.Packages {
		for _, t := range p.Tools {
			if version != "latest" {
				if t.Name == name && t.Version == version {
					correctTool = t
				}
			} else {
				// Find latest
				v1, _ := semver.Make(t.Version)
				v2, _ := semver.Make(correctTool.Version)
				if t.Name == name && v1.Compare(v2) > 0 {
					correctTool = t
				}
			}
		}
	}
	return correctTool
}

func findBaseDir(dirList []string) string {
	baseDir := ""
	for index, _ := range dirList {
		candidateBaseDir := dirList[index]
		for i := index; i < len(dirList); i++ {
			if !strings.Contains(dirList[i], candidateBaseDir) {
				return baseDir
			}
		}
		// avoid setting the candidate if it is the last file
		if dirList[len(dirList)-1] != candidateBaseDir {
			baseDir = candidateBaseDir
		}
	}
	return baseDir
}

func extractZip(body []byte, location string) (string, error) {
	path, err := utilities.SaveFileonTempDir("tooldownloaded.zip", bytes.NewReader(body))
	r, err := zip.OpenReader(path)
	if err != nil {
		return location, err
	}

	var dirList []string

	for _, f := range r.File {
		dirList = append(dirList, f.Name)
	}

	basedir := findBaseDir(dirList)

	for _, f := range r.File {
		fullname := filepath.Join(location, strings.Replace(f.Name, basedir, "", -1))
		if f.FileInfo().IsDir() {
			os.MkdirAll(fullname, f.FileInfo().Mode().Perm())
		} else {
			os.MkdirAll(filepath.Dir(fullname), 0755)
			perms := f.FileInfo().Mode().Perm()
			out, err := os.OpenFile(fullname, os.O_CREATE|os.O_RDWR, perms)
			if err != nil {
				return location, err
			}
			rc, err := f.Open()
			if err != nil {
				return location, err
			}
			_, err = io.CopyN(out, rc, f.FileInfo().Size())
			if err != nil {
				return location, err
			}
			rc.Close()
			out.Close()

			mtime := f.FileInfo().ModTime()
			err = os.Chtimes(fullname, mtime, mtime)
			if err != nil {
				return location, err
			}
		}
	}
	return location, nil
}

func extractTarGz(body []byte, location string) (string, error) {
	bodyCopy := make([]byte, len(body))
	copy(bodyCopy, body)
	tarFile, _ := gzip.NewReader(bytes.NewReader(body))
	tarReader := tar.NewReader(tarFile)

	var dirList []string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		dirList = append(dirList, header.Name)
	}

	basedir := findBaseDir(dirList)

	tarFile, _ = gzip.NewReader(bytes.NewReader(bodyCopy))
	tarReader = tar.NewReader(tarFile)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			//return location, err
		}

		path := filepath.Join(location, strings.Replace(header.Name, basedir, "", -1))
		info := header.FileInfo()

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return location, err
			}
			continue
		}

		if header.Typeflag == tar.TypeSymlink {
			err = os.Symlink(header.Linkname, path)
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			//return location, err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			//return location, err
		}
	}
	return location, nil
}

func (t *Tools) installDrivers(location string) {
	if runtime.GOOS == "windows" {
		if _, err := os.Stat(filepath.Join(location, "post_install.bat")); err == nil {
			t.Logger.Println("Installing drivers")
			oscmd := exec.Command(filepath.Join(location, "post_install.bat"))
			TellCommandNotToSpawnShell(oscmd)
			oscmd.Run()
		}
	}
}

func extractBz2(body []byte, location string) (string, error) {
	bodyCopy := make([]byte, len(body))
	copy(bodyCopy, body)
	tarFile := bzip2.NewReader(bytes.NewReader(body))
	tarReader := tar.NewReader(tarFile)

	var dirList []string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		dirList = append(dirList, header.Name)
	}

	basedir := findBaseDir(dirList)

	tarFile = bzip2.NewReader(bytes.NewReader(bodyCopy))
	tarReader = tar.NewReader(tarFile)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			//return location, err
		}

		path := filepath.Join(location, strings.Replace(header.Name, basedir, "", -1))
		info := header.FileInfo()

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return location, err
			}
			continue
		}

		if header.Typeflag == tar.TypeSymlink {
			err = os.Symlink(header.Linkname, path)
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			//return location, err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			//return location, err
		}
	}
	return location, nil
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
