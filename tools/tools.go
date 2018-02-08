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
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */
package tools

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeclysm/extract"
)

type Tool struct {
	Name     string
	Version  string
	Packager string
	Path     string
}

func (t *Tool) Download(url, signature string, opts *Opts) error {
	opts = opts.fill()

	// Download
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := opts.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Checksum
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	h := sha256.New()
	h.Write(body)
	sum := h.Sum(nil)
	signature = strings.Split(signature, ":")[1]

	if string(hex.EncodeToString(sum)) != signature {
		return errors.New("signature doesn't match")
	}

	// Remove folder
	path := filepath.Join(opts.Location, t.Packager, t.Name, t.Version)
	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	// Extract
	err = extract.Archive(bytes.NewReader(body), path, func(file string) string {
		// Remove the first part of the path if it matches the name
		parts := strings.Split(file, string(filepath.Separator))
		if len(parts) > 0 && strings.HasPrefix(parts[0], t.Name) {
			parts = parts[1:]
			file = strings.Join(parts, string(filepath.Separator))
		}

		return file
	})
	if err != nil {
		return err
	}
	t.Path = path

	return nil
}

// Opts contain options to pass to the Download function
type Opts struct {
	Location string
	Client   *http.Client
}

// fill fills itself with default values
func (o *Opts) fill() *Opts {
	if o == nil {
		o = &Opts{}
	}

	if o.Location == "" {
		usr, _ := user.Current()
		o.Location = filepath.Join(usr.HomeDir, ".arduino-create")
	}

	if o.Client == nil {
		o.Client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	fmt.Println(o)

	return o
}

// Installed returns a list of the installed tools
func Installed(opts *Opts) ([]Tool, error) {
	opts = opts.fill()

	packagers, err := ioutil.ReadDir(opts.Location)
	if err != nil {
		return nil, err
	}

	installed := []Tool{}

	for _, packager := range packagers {
		if !packager.IsDir() {
			continue
		}

		path := filepath.Join(opts.Location, packager.Name())
		tools, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, tool := range tools {
			if !tool.IsDir() {
				continue
			}

			path := filepath.Join(opts.Location, packager.Name(), tool.Name())
			versions, err := ioutil.ReadDir(path)
			if err != nil {
				return nil, err
			}

			for _, version := range versions {
				if !version.IsDir() {
					continue
				}

				installed = append(installed, Tool{
					Name:     tool.Name(),
					Packager: packager.Name(),
					Version:  version.Name(),
					Path:     path,
				})
			}
		}
	}

	return installed, nil
}
