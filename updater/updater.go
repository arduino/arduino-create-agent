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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"

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
func CheckForUpdates(currentVersion string, updateURL string, cmdName string) (string, error) {
	return checkForUpdates(currentVersion, updateURL, cmdName)
}

const (
	plat = runtime.GOOS + "-" + runtime.GOARCH
)

func fetchInfo(updateAPIURL string) (*availableUpdateInfo, error) {
	r, err := fetch(updateAPIURL)
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

// addTempSuffixToPath adds the "-temp" suffix to the path to an executable file (a ".exe" extension is replaced with "-temp.exe")
func addTempSuffixToPath(path string) string {
	if filepath.Ext(path) == ".exe" {
		// Windows
		path = strings.Replace(path, ".exe", "-temp.exe", -1)
	} else {
		// Unix
		path = path + "-temp"
	}

	return path
}

// removeTempSuffixFromPath removes "-temp" suffix from the path to an executable file (a "-temp.exe" extension is replaced with ".exe")
func removeTempSuffixFromPath(path string) string {
	return strings.Replace(path, "-temp", "", -1)
}

func copyExe(from, to string) error {
	data, err := os.ReadFile(from)
	if err != nil {
		log.Println("Cannot read file: ", from)
		return err
	}
	err = os.WriteFile(to, data, 0755)
	if err != nil {
		log.Println("Cannot write file: ", to)
		return err
	}
	return nil
}

// requestElevation requests this program to rerun as administrator, for when we don't have permission over the update files
func requestElevation() {
	log.Println("Permission denied. Requesting elevated privileges...")
	var err error
	if runtime.GOOS == "windows" {
		err = elevateWindows()
	} else {
		err = elevateUnix()
	}

	if err != nil {
		log.Println("Failed to request elevation:", err)
		return
	}
}

func elevateWindows() error {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, err := syscall.UTF16PtrFromString(verb)
	if err != nil {
		return err
	}
	exePtr, err := syscall.UTF16PtrFromString(exe)
	if err != nil {
		return err
	}
	cwdPtr, err := syscall.UTF16PtrFromString(cwd)
	if err != nil {
		return err
	}
	argPtr, _ := syscall.UTF16PtrFromString(args)
	var showCmd int32 = 1
	return windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
}

func elevateUnix() error {
	args := append([]string{os.Args[0]}, os.Args[1:]...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
