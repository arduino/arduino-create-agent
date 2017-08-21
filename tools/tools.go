/*
 * This file is part of arduino-create-agent.
 *
 * arduino-create-agent is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */
package tools

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/codeclysm/extract"
	"github.com/pkg/errors"
	"github.com/xrash/smetrics"

	"golang.org/x/crypto/openpgp"
)

// Opts contain options to pass to the Download function
type Opts struct {
	IndexURL string
	Key      string
	Location string
	Client   *http.Client
}

// fill fills itself with default values
func (o *Opts) fill() *Opts {
	if o == nil {
		o = &Opts{}
	}

	if o.IndexURL == "" {
		o.IndexURL = "https://downloads.arduino.cc/packages/package_index.json"
	}

	if o.Key == "" {
		o.Key = gpgpHex
	}

	return o
}

// Download by default parses a cached version (refreshed every hour) of
// https://downloads.arduino.cc/packages/package_index.json
// for a suitable download for the user OS and unpacks it in the ~/.arduino-create folder
// replacing the existing files if existing
// version can be a valid version or the string "latest"
func Download(packager, name, version string, opts *Opts) error {
	opts = opts.fill()

	// download index
	err := downloadIndex(opts.IndexURL, opts.Location, opts.Key, opts.Client)
	if err != nil {
		return errors.Wrap(err, "download index")
	}

	// parse index
	data, err := ioutil.ReadFile(filepath.Join(opts.Location, filepath.Base(opts.IndexURL)))
	if err != nil {
		return errors.Wrap(err, "parse index")
	}

	var i index
	err = json.Unmarshal(data, &i)
	if err != nil {
		return errors.Wrap(err, "parse index")
	}

	tool, system := i.find(packager, name, version)
	if tool.Name == "" || system.URL == "" {
		return errors.New("tool not found")
	}

	// Download
	resp, err := http.Get(system.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Remove folder
	path := filepath.Join(opts.Location, tool.Name, tool.Version)
	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	// Extract
	err = extract.Archive(resp.Body, path, func(file string) string {
		// Remove the first part of the path if it matches the name
		parts := strings.Split(file, string(filepath.Separator))
		if len(parts) > 0 && parts[0] == tool.Name {
			parts = parts[1:]
			file = strings.Join(parts, string(filepath.Separator))
		}

		return file
	})
	if err != nil {
		return err
	}

	return nil
}

// List returns a list of the installed tools
func List() {

}

// downloadIndex parses https://downloads.arduino.cc/packages/package_index.json checking the signature and saving both files into location
func downloadIndex(url, location, keyString string, client *http.Client) error {
	if client == nil {
		client = &http.Client{}
		client.Timeout = 30 * time.Second
	}
	// Fetch the index
	index, err := client.Get(url)
	if err != nil {
		return err
	}
	defer index.Body.Close()

	// Fetch the signature
	sig, err := client.Get(url + ".sig")
	if err != nil {
		return err
	}
	defer sig.Body.Close()

	var bodyBuf, sigBuf bytes.Buffer
	body := io.TeeReader(index.Body, &bodyBuf)
	sigBody := io.TeeReader(sig.Body, &sigBuf)

	// Check signature
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return err
	}

	err = checkGPGSig(body, sigBody, key)
	if err != nil {
		return err
	}

	filename := filepath.Join(location, filepath.Base(url))

	bodyFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer bodyFile.Close()

	_, err = io.Copy(bodyFile, &bodyBuf)
	if err != nil {
		return err
	}

	sigFile, err := os.Create(filename + ".sig")
	if err != nil {
		return err
	}
	defer sigFile.Close()

	_, err = io.Copy(sigFile, &sigBuf)
	if err != nil {
		return err
	}

	return nil
}

// checkGPGSig validates the signature (sig) of the body with the given key
func checkGPGSig(body, sig io.Reader, key []byte) error {
	keyring, err := openpgp.ReadKeyRing(bytes.NewReader(key))
	if err != nil {
		return err
	}

	_, err = openpgp.CheckDetachedSignature(keyring, body, sig)
	if err != nil {
		return err
	}

	return nil
}

type index struct {
	Packages []struct {
		Name  string `json:"name"`
		Tools []tool `json:"tools"`
	} `json:"packages"`
}

func (i index) find(packager, name, version string) (correctTool tool, correctSystem system) {
	correctTool.Version = "0.0"

	for _, p := range i.Packages {
		if p.Name != packager {
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
	maxSimilarity := 0.7

	for _, s := range correctTool.Systems {
		similarity := smetrics.Jaro(s.Host, systems[runtime.GOOS+runtime.GOARCH])
		if similarity > maxSimilarity {
			correctSystem = s
			maxSimilarity = similarity
		}
	}

	return correctTool, correctSystem
}

type tool struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Systems     []system `json:"systems"`
	url         string
	destination string
}
type system struct {
	Host     string `json:"host"`
	URL      string `json:"url"`
	Name     string `json:"archiveFileName"`
	CheckSum string `json:"checksum"`
}

var systems = map[string]string{
	"linuxamd64":  "x86_64-linux-gnu",
	"linux386":    "i686-linux-gnu",
	"darwinamd64": "apple-darwin",
	"windows386":  "i686-mingw32",
}
