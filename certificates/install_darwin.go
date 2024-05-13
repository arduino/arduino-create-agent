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

package certificates

/*
// Explicitly tell the GCC compiler that the language is Objective-C.
#cgo CFLAGS: -x objective-c

// Pass the list of macOS frameworks needed by this piece of Objective-C code.
// The "-ld_classic" is needed to avoid a wrong warning about duplicate libraries when building with XCode 15.
#cgo LDFLAGS: -framework Foundation -framework Security -framework AppKit -ld_classic

#import <Foundation/Foundation.h>
#include "certificates_darwin.h"
*/
import "C"
import (
	"errors"
	"time"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"github.com/arduino/arduino-create-agent/utilities"
	"github.com/arduino/go-paths-helper"
)

// InstallCertificate will install the certificates in the system keychain on macos,
// if something goes wrong will show a dialog with the error and return an error
func InstallCertificate(cert *paths.Path) error {
	log.Infof("Installing certificate: %s", cert)
	ccert := C.CString(cert.String())
	defer C.free(unsafe.Pointer(ccert))
	p := C.installCert(ccert)
	s := C.GoString(p)
	if len(s) != 0 {
		utilities.UserPrompt(s, "\"OK\"", "OK", "OK", "Arduino Agent: Error installing certificates")
		UninstallCertificates()
		return errors.New(s)
	}
	return nil
}

// UninstallCertificates will uninstall the certificates from the system keychain on macos,
// if something goes wrong will show a dialog with the error and return an error
func UninstallCertificates() error {
	log.Infof("Uninstalling certificates")
	p := C.uninstallCert()
	s := C.GoString(p)
	if len(s) != 0 {
		utilities.UserPrompt(s, "\"OK\"", "OK", "OK", "Arduino Agent: Error uninstalling certificates")
		return errors.New(s)
	}
	return nil
}

// GetExpirationDate returns the expiration date of a certificate stored in the keychain
func GetExpirationDate() (time.Time, error) {
	log.Infof("Retrieving certificate's expiration date")

	expirationDateLong := C.long(0)

	err := C.getExpirationDate(&expirationDateLong)
	errString := C.GoString(err)
	if len(errString) > 0 {
		utilities.UserPrompt(errString, "\"OK\"", "OK", "OK", "Arduino Agent: Error retrieving expiration date")
		return time.Time{}, errors.New(errString)
	}

	// The expirationDate is the number of seconds from the date of 1 Jan 2001 00:00:00 GMT.
	// Add 31 years to convert it to Unix Epoch.
	expirationDate := int64(expirationDateLong)
	return time.Unix(expirationDate, 0).AddDate(31, 0, 0), nil
}

// GetDefaultBrowserName returns the name of the default browser
func GetDefaultBrowserName() string {
	log.Infof("Retrieving default browser name")
	p := C.getDefaultBrowserName()
	return C.GoString(p)
}

// CertInKeychain checks if the certificate is stored inside the keychain
func CertInKeychain() bool {
	log.Infof("Checking if the Arduino certificate is in the keychain")

	certInKeychain := C.certInKeychain()
	return bool(certInKeychain)
}
