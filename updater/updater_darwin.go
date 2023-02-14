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
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract/v3"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func start(src string) string {
	if strings.Contains(src, "-temp") {
		// This is required to transition from the previous auto-update system to the new one.
		// Updating from version <=1.2.7 will produce the `ArduinoCreateAgent-temp` file and we should
		// complete the upgrade by copying the `ArduinoCreateAgent-temp` executable back to the original.
		//
		// This procedure will be automatically skipped starting from version >1.2.7.

		newPath := removeTempSuffixFromPath(src)
		if err := copyExe(src, newPath); err != nil {
			log.Println("Copy error: ", err)
			panic(err)
		}
		return newPath
	}

	return ""
}

func checkForUpdates(currentVersion string, updateURL string, cmdName string) (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not app path: %w", err)
	}
	currentAppPath := paths.New(executablePath).Parent().Parent().Parent()
	if currentAppPath.Ext() != ".app" {
		return "", fmt.Errorf("could not find app root in %s", executablePath)
	}
	oldAppPath := currentAppPath.Parent().Join("ArdiunoCreateAgent.old.app")
	if oldAppPath.Exist() {
		return "", fmt.Errorf("temp app already exists: %s, cannot update", oldAppPath)
	}

	// Fetch information about updates

	// updateURL: "https://downloads.arduino.cc/"
	// cmdName: "CreateAgent/Stable"
	// plat: "darwin-amd64"
	// info URL: "https://downloads.arduino.cc/CreateAgent/Stable/darwin-amd64-bundle.json"
	infoURL := updateURL + cmdName + "/" + plat + "-bundle.json"
	info, err := fetchInfo(infoURL)
	if err != nil {
		return "", err
	}
	if info.Version == currentVersion {
		// No updates available, bye bye
		return "", nil
	}

	tmp := paths.TempDir().Join("arduino-create-agent")
	if err := tmp.MkdirAll(); err != nil {
		return "", err
	}
	tmpZip := tmp.Join("update.zip")
	tmpAppPath := tmp.Join("ArduinoCreateAgent-update.app")
	defer tmp.RemoveAll()

	// Download the update.

	// updateURL: "https://downloads.arduino.cc/"
	// cmdName == "CreateAgent/Stable"
	// info.Version == "1.2.8"
	// runtime.GOARCH: "amd64"
	// downloadURL: "https://downloads.arduino.cc/CreateAgent/Stable/1.2.8/ArduinoCreateAgent.app_arm64_notarized.zip"
	downloadURL := updateURL + cmdName + "/" + info.Version + "/ArduinoCreateAgent.app_" + runtime.GOARCH + "_notarized.zip"
	logrus.WithField("url", downloadURL).Info("Downloading update")
	download, err := fetch(downloadURL)
	if err != nil {
		return "", err
	}
	defer download.Close()

	f, err := tmpZip.Create()
	if err != nil {
		return "", err
	}
	defer f.Close()

	sha := sha256.New()
	if _, err := io.Copy(io.MultiWriter(sha, f), download); err != nil {
		return "", err
	}
	f.Close()

	// Check the hash
	if s := sha.Sum(nil); !bytes.Equal(s, info.Sha256) {
		return "", fmt.Errorf("bad hash: %s (expected %s)", s, info.Sha256)
	}

	// Unzip the update
	logrus.WithField("tmpDir", tmpAppPath).Info("Unzipping update")
	if err := tmpAppPath.MkdirAll(); err != nil {
		return "", fmt.Errorf("could not create tmp dir to unzip update: %w", err)
	}

	f, err = tmpZip.Open()
	if err != nil {
		return "", fmt.Errorf("could not open archive for unzip: %w", err)
	}
	defer f.Close()
	if err := extract.Archive(context.Background(), f, tmpAppPath.String(), nil); err != nil {
		return "", fmt.Errorf("extracting archive: %w", err)
	}

	// Rename current app as .old
	logrus.WithField("from", currentAppPath).WithField("to", oldAppPath).Info("Renaming old app")
	if err := currentAppPath.Rename(oldAppPath); err != nil {
		return "", fmt.Errorf("could not rename old app as .old: %w", err)
	}

	// Install new app
	logrus.WithField("from", tmpAppPath).WithField("to", currentAppPath).Info("Copying updated app")
	if err := tmpAppPath.CopyDirTo(currentAppPath); err != nil {
		// Try rollback changes
		_ = currentAppPath.RemoveAll()
		_ = oldAppPath.Rename(currentAppPath)
		return "", fmt.Errorf("could not install app: %w", err)
	}

	// Remove old app
	logrus.WithField("to", oldAppPath).Info("Removing old app")
	_ = oldAppPath.RemoveAll()

	// Restart agent
	logrus.WithField("path", currentAppPath).Info("Running new app")

	// Close old agent
	return currentAppPath.String(), nil
}
