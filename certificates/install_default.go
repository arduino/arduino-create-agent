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

/*
// Importing "certificates.h" here even if it is not used avoids building errors on Ubuntu and Windows.

// Explicitly tell the GCC compiler that the language is Objective-C.
#cgo CFLAGS: -x objective-c

// Pass the list of macOS frameworks needed by this piece of Objective-C code.
// The "-ld_classic" is needed to avoid a wrong warning about duplicate libraries when building with XCode 15.
#cgo LDFLAGS: -framework Foundation -framework Security -framework AppKit -ld_classic

#import <Foundation/Foundation.h>
#include "certificates.h"
*/

import (
	"errors"
	"time"

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
func GetExpirationDate() (time.Time, error) {
	log.Warn("platform not supported for retrieving certificates expiration date")
	return time.Time{}, errors.New("platform not supported for retrieving certificates expiration date")
}

// GetDefaultBrowserName won't do anything on unsupported Operative Systems
func GetDefaultBrowserName() string {
	log.Warn("platform not supported for retrieving default browser name")
	return ""
}

// CertInKeychain won't do anything on unsupported Operative Systems
func CertInKeychain() bool {
	log.Warn("platform not supported for verifying the certificate existence")
	return false
}
