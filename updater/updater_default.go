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

//go:build !darwin

package updater

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/inconshreveable/go-update.v0"
)

// Update protocol:
//
//   GET hk.heroku.com/hk/linux-amd64.json
//
//   200 ok
//   {
//       "Version": "2",
//       "Sha256": "..." // base64
//   }
//
// then
//
//   GET hkpatch.s3.amazonaws.com/hk/1/2/linux-amd64
//
//   200 ok
//   [bsdiff data]
//
// or
//
//   GET hkdist.s3.amazonaws.com/hk/2/linux-amd64.gz
//
//   200 ok
//   [gzipped executable data]
//
//

var errHashMismatch = errors.New("new file hash mismatch after patch")
var up = update.New()

func start(src string) string {
	// If the executable is temporary, copy it to the full path, then restart
	if strings.Contains(src, "-temp") {
		newPath := removeTempSuffixFromPath(src)
		if err := copyExe(src, newPath); err != nil {
			log.Println("Copy error: ", err)
			panic(err)
		}
		return newPath
	}

	// Otherwise copy to a path with -temp suffix
	if err := copyExe(src, addTempSuffixToPath(src)); err != nil {
		panic(err)
	}
	return ""
}

func checkForUpdates(currentVersion string, updateURL string, cmdName string) (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	var up = &Updater{
		CurrentVersion: currentVersion,
		UpdateURL:      updateURL,
		Dir:            "update/",
		CmdName:        cmdName,
	}

	if err := up.BackgroundRun(); err != nil {
		return "", err
	}
	return addTempSuffixToPath(path), nil
}

// Updater is the configuration and runtime data for doing an update.
//
// Note that ApiURL, BinURL and DiffURL should have the same value if all files are available at the same location.
//
// Example:
//
//	updater := &selfupdate.Updater{
//		CurrentVersion: version,
//		UpdateURL:      "http://updates.yourdomain.com/",
//		Dir:            "update/",
//		CmdName:        "myapp", // app name
//	}
//	if updater != nil {
//		go updater.BackgroundRun()
//	}
type Updater struct {
	CurrentVersion string               // Currently running version.
	UpdateURL      string               // Base URL for API requests (json files).
	CmdName        string               // Command name is appended to the ApiURL like http://apiurl/CmdName/. This represents one binary.
	Dir            string               // Directory to store selfupdate state.
	Info           *availableUpdateInfo // Information about the available update.
}

// BackgroundRun starts the update check and apply cycle.
func (u *Updater) BackgroundRun() error {
	os.MkdirAll(u.getExecRelativeDir(u.Dir), 0777)
	if err := up.CanUpdate(); err != nil {
		log.Println(err)
		return err
	}
	//self, err := os.Executable()
	//if err != nil {
	// fail update, couldn't figure out path to self
	//return
	//}
	// TODO(bgentry): logger isn't on Windows. Replace w/ proper error reports.
	if err := u.update(); err != nil {
		return err
	}
	return nil
}

func verifySha(bin []byte, sha []byte) bool {
	h := sha256.New()
	h.Write(bin)
	return bytes.Equal(h.Sum(nil), sha)
}

func (u *Updater) fetchAndVerifyFullBin() ([]byte, error) {
	bin, err := u.fetchBin()
	if err != nil {
		return nil, err
	}
	verified := verifySha(bin, u.Info.Sha256)
	if !verified {
		return nil, errHashMismatch
	}
	return bin, nil
}

func (u *Updater) fetchBin() ([]byte, error) {
	r, err := fetch(u.UpdateURL + u.CmdName + "/" + u.Info.Version + "/" + plat + ".gz")
	if err != nil {
		return nil, err
	}
	defer r.Close()
	buf := new(bytes.Buffer)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(buf, gz); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (u *Updater) getExecRelativeDir(dir string) string {
	filename, _ := os.Executable()
	path := filepath.Join(filepath.Dir(filename), dir)
	return path
}

func (u *Updater) update() error {
	path, err := os.Executable()
	if err != nil {
		return err
	}

	path = addTempSuffixToPath(path)

	old, err := os.Open(path)
	if err != nil {
		return err
	}
	defer old.Close()

	infoURL := u.UpdateURL + u.CmdName + "/" + plat + ".json"
	info, err := fetchInfo(infoURL)
	if err != nil {
		log.Println(err)
		return err
	}
	u.Info = info
	if u.Info.Version == u.CurrentVersion {
		return nil
	}

	bin, err := u.fetchAndVerifyFullBin()
	if err != nil {
		if err == errHashMismatch {
			log.Println("update: hash mismatch from full binary")
		} else {
			log.Println("update: fetching full binary,", err)
		}
		return err
	}

	// close the old binary before installing because on windows
	// it can't be renamed if a handle to the file is still open
	old.Close()

	up.TargetPath = path
	err, errRecover := up.FromStream(bytes.NewBuffer(bin))
	if errRecover != nil {
		log.Errorf("update and recovery errors: %q %q", err, errRecover)
		return fmt.Errorf("update and recovery errors: %q %q", err, errRecover)
	}
	if err != nil {
		return err
	}

	return nil
}
