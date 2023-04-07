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

// code inspired by https://github.com/FiloSottile/mkcert licenced under BSD3

package certificates

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"syscall"
	"unsafe"

	"github.com/arduino/go-paths-helper"
	log "github.com/sirupsen/logrus"
)

var (
	modcrypt32                           = syscall.NewLazyDLL("crypt32.dll")
	procCertAddEncodedCertificateToStore = modcrypt32.NewProc("CertAddEncodedCertificateToStore")
	procCertCloseStore                   = modcrypt32.NewProc("CertCloseStore")
	procCertOpenSystemStoreW             = modcrypt32.NewProc("CertOpenSystemStoreW")
)

// // InstallCertificate will install the certificates in the system keychain on windows
func InstallCertificate(certFile *paths.Path) {
	// Load cert
	cert, err := ioutil.ReadFile(certFile.String())
	if err != nil {
		log.Errorf("failed to read root certificate: %s", err)
	}
	// Decode PEM
	if certBlock, _ := pem.Decode(cert); certBlock == nil || certBlock.Type != "CERTIFICATE" {
		log.Error("invalid PEM data: decode pem")
	} else {
		cert = certBlock.Bytes
	}
	// Open root store
	store, err := openWindowsRootStore()
	if err != nil {
		log.Errorf("cannot open root store %s", err)
	} else {
		log.Info("opened Root Store")
	}
	defer store.close()
	// Add cert
	err = store.addCert(cert)
	if err != nil {
		log.Errorf("cannot install certificate in the system keychain: %s", err)
	} else {
		log.Info("certificate installed")
	}
}

type windowsRootStore uintptr

func openWindowsRootStore() (windowsRootStore, error) {
	rootStr, err := syscall.UTF16PtrFromString("ROOT")
	if err != nil {
		return 0, err
	}
	store, _, err := procCertOpenSystemStoreW.Call(0, uintptr(unsafe.Pointer(rootStr)))
	if store != 0 {
		return windowsRootStore(store), nil
	}
	return 0, fmt.Errorf("failed to open windows root store: %s", err)
}

func (w windowsRootStore) close() error {
	ret, _, err := procCertCloseStore.Call(uintptr(w), 0)
	if ret != 0 {
		return nil
	}
	return fmt.Errorf("failed to close windows root store: %s", err)
}

func (w windowsRootStore) addCert(cert []byte) error {
	// this will always override
	ret, _, err := procCertAddEncodedCertificateToStore.Call(
		uintptr(w), // HCERTSTORE hCertStore
		uintptr(syscall.X509_ASN_ENCODING|syscall.PKCS_7_ASN_ENCODING), // DWORD dwCertEncodingType
		uintptr(unsafe.Pointer(&cert[0])),                              // const BYTE *pbCertEncoded
		uintptr(len(cert)),                                             // DWORD cbCertEncoded
		3,                                                              // DWORD dwAddDisposition (CERT_STORE_ADD_REPLACE_EXISTING is 3)
		0,                                                              // PCCERT_CONTEXT *ppCertContext
	)
	if ret != 0 {
		return nil
	}
	return fmt.Errorf("failed adding cert: %s", err)
}
