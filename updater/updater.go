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

package updater

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"

	log "github.com/sirupsen/logrus"
)

// Start checks if an update has been downloaded and if so returns the path to the
// binary to be executed to perform the update. If no update has been downloaded
// it returns an empty string.
func Start(src string) string {
	return start(src)
}

// CheckForUpdates checks if there is a new version of the binary available and
// if so downloads it.
func CheckForUpdates(currentVersion string, updateAPIURL, updateBinURL string, cmdName string) (string, error) {
	return checkForUpdates(currentVersion, updateAPIURL, updateBinURL, cmdName)
}

const (
	plat = runtime.GOOS + "-" + runtime.GOARCH
)

func fetchInfo(updateAPIURL string, cmdName string) (*availableUpdateInfo, error) {
	r, err := fetch(updateAPIURL + cmdName + "/" + plat + ".json")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var res availableUpdateInfo
	if err := json.NewDecoder(r).Decode(&res); err != nil {
		return nil, err
	}
	if len(res.Sha256) != sha256.Size {
		return nil, errors.New("bad cmd hash in info")
	}
	return &res, nil
}

type availableUpdateInfo struct {
	Version string
	Sha256  []byte
}

func fetch(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Errorf("bad http status from %s: %v", url, resp.Status)
		return nil, fmt.Errorf("bad http status from %s: %v", url, resp.Status)
	}
	return resp.Body, nil
}
