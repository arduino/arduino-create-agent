// Copyright 2023 Arduino SA
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

package certificates

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/arduino/go-paths-helper"
)

// InstallCertificate won't do anything on unsupported Operative Systems
func InstallCertificate(cert *paths.Path) error {
	log.Warn("platform not supported for the certificate install")
	return errors.New("platform not supported for the certificate install")
}

// UninstallCertificates won't do anything on unsupported Operative Systems
func UninstallCertificates() error {
	log.Warn("platform not supported for the certificates uninstall")
	return errors.New("platform not supported for the certificates uninstall")
}

// GetExpirationDate won't do anything on unsupported Operative Systems
func GetExpirationDate() (string, error) {
	log.Warn("platform not supported for retrieving certificates expiration date")
	return "", errors.New("platform not supported for retrieving certificates expiration date")
}
