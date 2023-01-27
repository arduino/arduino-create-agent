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
