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
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/arduino/arduino-create-agent/utilities"
	"github.com/arduino/arduino-create-agent/v2/pkgs"
	"github.com/blang/semver"
)

// public vars to allow override in the tests
var (
	OS   = runtime.GOOS
	Arch = runtime.GOARCH
)

func mimeType(data []byte) (string, error) {
	return http.DetectContentType(data[0:512]), nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// Download will parse the index at the indexURL for the tool to download.
// It will extract it in a folder in .arduino-create, and it will update the
// Installed map.
//
// pack contains the packager of the tool
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
func (t *Tools) Download(pack, name, version, behaviour string) error {

	body, err := t.index.Read()
	if err != nil {
		return err
	}

	var data pkgs.Index
	json.Unmarshal(body, &data)

	// Find the tool by name
	correctTool, correctSystem := findTool(pack, name, version, data)

	if correctTool.Name == "" || correctSystem.URL == "" {
		t.logger("We couldn't find a tool with the name " + name + " and version " + version + " packaged by " + pack)
		return nil
	}

	key := correctTool.Name + "-" + correctTool.Version

	// Check if it already exists
	if behaviour == "keep" {
		location, ok := t.getMapValue(key)
		if ok && pathExists(location) {
			// overwrite the default tool with this one
			t.setMapValue(correctTool.Name, location)
			t.logger("The tool is already present on the system")
			return t.writeMap()
		}
	}

	// Download the tool
	t.logger("Downloading tool " + name + " from " + correctSystem.URL)
	resp, err := http.Get(correctSystem.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the body
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Checksum
	checksum := sha256.Sum256(body)
	checkSumString := "SHA-256:" + hex.EncodeToString(checksum[:sha256.Size])

	if checkSumString != correctSystem.Checksum {
		return errors.New("checksum doesn't match")
	}

	// Decompress
	t.logger("Unpacking tool " + name)

	location := t.directory.Join(pack, correctTool.Name, correctTool.Version).String()
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
		location, err = extractZip(t.logger, body, location)
	case "application/x-bz2":
	case "application/octet-stream":
		location, err = extractBz2(t.logger, body, location)
	case "application/x-gzip":
		location, err = extractTarGz(t.logger, body, location)
	default:
		return errors.New("Unknown extension for file " + correctSystem.URL)
	}

	if err != nil {
		t.logger("Error extracting the archive: " + err.Error())
		return err
	}

	err = t.installDrivers(location)
	if err != nil {
		return err
	}

	// Ensure that the files are executable
	t.logger("Ensure that the files are executable")

	// Update the tool map
	t.logger("Updating map with location " + location)

	t.setMapValue(name, location)
	t.setMapValue(name+"-"+correctTool.Version, location)
	return t.writeMap()
}

func findTool(pack, name, version string, data pkgs.Index) (pkgs.Tool, pkgs.System) {
	var correctTool pkgs.Tool
	correctTool.Version = "0.0"

	for _, p := range data.Packages {
		if p.Name != pack {
			continue
		}
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

	// Find the url based on system
	correctSystem := correctTool.GetFlavourCompatibleWith(OS, Arch)

	return correctTool, correctSystem
}

func commonPrefix(sep byte, paths []string) string {
	// Handle special cases.
	switch len(paths) {
	case 0:
		return ""
	case 1:
		return path.Clean(paths[0])
	}

	c := []byte(path.Clean(paths[0]))

	// We add a trailing sep to handle: common prefix directory is included in the path list
	// (e.g. /home/user1, /home/user1/foo, /home/user1/bar).
	// path.Clean will have cleaned off trailing / separators with
	// the exception of the root directory, "/" making it "//"
	// but this will get fixed up to "/" below).
	c = append(c, sep)

	// Ignore the first path since it's already in c
	for _, v := range paths[1:] {
		// Clean up each path before testing it
		v = path.Clean(v) + string(sep)

		// Find the first non-common byte and truncate c
		if len(v) < len(c) {
			c = c[:len(v)]
		}
		for i := 0; i < len(c); i++ {
			if v[i] != c[i] {
				c = c[:i]
				break
			}
		}
	}

	// Remove trailing non-separator characters and the final separator
	for i := len(c) - 1; i >= 0; i-- {
		if c[i] == sep {
			c = c[:i]
			break
		}
	}

	return string(c)
}

func removeStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func findBaseDir(dirList []string) string {
	if len(dirList) == 1 {
		return path.Dir(dirList[0]) + "/"
	}

	// https://github.com/backdrop-ops/contrib/issues/55#issuecomment-73814500
	dontdiff := []string{"pax_global_header"}
	for _, v := range dontdiff {
		dirList = removeStringFromSlice(dirList, v)
	}

	commonBaseDir := commonPrefix('/', dirList)
	if commonBaseDir != "" {
		commonBaseDir = commonBaseDir + "/"
	}
	return commonBaseDir
}

func extractZip(log func(msg string), body []byte, location string) (string, error) {
	path, _ := utilities.SaveFileonTempDir("tooldownloaded.zip", bytes.NewReader(body))
	r, err := zip.OpenReader(path)
	if err != nil {
		return location, err
	}

	var dirList []string

	for _, f := range r.File {
		dirList = append(dirList, f.Name)
	}

	basedir := findBaseDir(dirList)
	log(fmt.Sprintf("selected baseDir %s from Zip Archive Content: %v", basedir, dirList))

	for _, f := range r.File {
		fullname := filepath.Join(location, strings.Replace(f.Name, basedir, "", -1))
		log(fmt.Sprintf("generated fullname %s removing %s from %s", fullname, basedir, f.Name))
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

func extractTarGz(log func(msg string), body []byte, location string) (string, error) {
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
	log(fmt.Sprintf("selected baseDir %s from TarGz Archive Content: %v", basedir, dirList))

	tarFile, _ = gzip.NewReader(bytes.NewReader(bodyCopy))
	tarReader = tar.NewReader(tarFile)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return location, err
		}

		path := filepath.Join(location, strings.Replace(header.Name, basedir, "", -1))
		info := header.FileInfo()

		// Create parent folder
		dirmode := info.Mode() | os.ModeDir | 0700
		if err = os.MkdirAll(filepath.Dir(path), dirmode); err != nil {
			return location, err
		}

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return location, err
			}
			continue
		}

		if header.Typeflag == tar.TypeSymlink {
			_ = os.Symlink(header.Linkname, path)
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			continue
		}
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return location, err
		}
		file.Close()
	}
	return location, nil
}

func extractBz2(log func(msg string), body []byte, location string) (string, error) {
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
	log(fmt.Sprintf("selected baseDir %s from Bz2 Archive Content: %v", basedir, dirList))

	tarFile = bzip2.NewReader(bytes.NewReader(bodyCopy))
	tarReader = tar.NewReader(tarFile)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			continue
			//return location, err
		}

		path := filepath.Join(location, strings.Replace(header.Name, basedir, "", -1))
		info := header.FileInfo()

		// Create parent folder
		dirmode := info.Mode() | os.ModeDir | 0700
		if err = os.MkdirAll(filepath.Dir(path), dirmode); err != nil {
			return location, err
		}

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return location, err
			}
			continue
		}

		if header.Typeflag == tar.TypeSymlink {
			_ = os.Symlink(header.Linkname, path)
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			continue
			//return location, err
		}
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return location, err
		}
		file.Close()
	}
	return location, nil
}

func (t *Tools) installDrivers(location string) error {
	OkPressed := 6
	extension := ".bat"
	preamble := ""
	if OS != "windows" {
		extension = ".sh"
		// add ./ to force locality
		preamble = "./"
	}
	if _, err := os.Stat(filepath.Join(location, "post_install"+extension)); err == nil {
		t.logger("Installing drivers")
		ok := MessageBox("Installing drivers", "We are about to install some drivers needed to use Arduino/Genuino boards\nDo you want to continue?")
		if ok == OkPressed {
			os.Chdir(location)
			t.logger(preamble + "post_install" + extension)
			oscmd := exec.Command(preamble + "post_install" + extension)
			if OS != "linux" {
				// spawning a shell could be the only way to let the user type his password
				TellCommandNotToSpawnShell(oscmd)
			}
			err = oscmd.Run()
			return err
		}
		return errors.New("could not install drivers")
	}
	return nil
}
