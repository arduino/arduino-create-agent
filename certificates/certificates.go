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

// Generate a self-signed X.509 certificate for a TLS server. Outputs to
// 'cert.pem' and 'key.pem' and will overwrite existing files.

package certificates

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/arduino/go-paths-helper"
	log "github.com/sirupsen/logrus"
)

var (
	host      = "localhost"
	validFrom = ""
	validFor  = 365 * 24 * time.Hour * 2 // 2 years
	rsaBits   = 2048
)

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func generateKey(ecdsaCurve string) (interface{}, error) {
	switch ecdsaCurve {
	case "":
		return rsa.GenerateKey(rand.Reader, rsaBits)
	case "P224":
		return ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, fmt.Errorf("unrecognized elliptic curve: %q", ecdsaCurve)
	}
}

func generateSingleCertificate(isCa bool) (*x509.Certificate, error) {
	var notBefore time.Time
	var err error
	if len(validFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", validFrom)
		if err != nil {
			return nil, fmt.Errorf("failed to parse creation date: %s", err.Error())
		}
	}

	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %s", err.Error())
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"Arduino LLC US"},
			Country:            []string{"US"},
			CommonName:         "127.0.0.1",
			OrganizationalUnit: []string{"IT"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	template.IPAddresses = append(template.IPAddresses, net.ParseIP("127.0.0.1"))
	template.DNSNames = append(template.DNSNames, "localhost")

	if isCa {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		template.Subject.CommonName = "Arduino"
	}

	return &template, nil
}

// MigrateCertificatesGeneratedWithOldAgentVersions checks if certificates generated
// with an old version of the Agent needs to be migrated to the current certificates
// directory, and performs the migration if needed.
func MigrateCertificatesGeneratedWithOldAgentVersions(certsDir *paths.Path) {
	if certsDir.Join("ca.cert.pem").Exist() {
		// The new certificates are already set-up, nothing to do
		return
	}

	fileList := []string{
		"ca.key.pem",
		"ca.cert.pem",
		"ca.cert.cer",
		"key.pem",
		"cert.pem",
		"cert.cer",
	}
	oldCertsDirPath, _ := os.Executable()
	oldCertsDir := paths.New(oldCertsDirPath)
	for _, fileName := range fileList {
		oldCert := oldCertsDir.Join(fileName)
		if oldCert.Exist() {
			oldCert.CopyTo(certsDir.Join(fileName))
		}
	}
}

// GenerateCertificates will generate the required certificates useful for a HTTPS connection on localhost
func GenerateCertificates(certsDir *paths.Path) {

	// Create the key for the certification authority
	caKey, err := generateKey("P256")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	{
		keyOutPath := certsDir.Join("ca.key.pem").String()
		keyOut, err := os.OpenFile(keyOutPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) // Save key with user-only permission 0600
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		pem.Encode(keyOut, pemBlockForKey(caKey))
		keyOut.Close()
		log.Printf("written %s", keyOutPath)
	}

	// Create the certification authority
	caTemplate, err := generateSingleCertificate(true)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	derBytes, _ := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, publicKey(caKey), caKey)

	{
		caCertOutPath := certsDir.Join("ca.cert.pem")
		caCertOut, err := caCertOutPath.Create()
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		pem.Encode(caCertOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		caCertOut.Close()
		log.Printf("written %s", caCertOutPath)
	}

	{
		caCertPath := certsDir.Join("ca.cert.cer")
		caCertPath.WriteFile(derBytes)
		log.Printf("written %s", caCertPath)
	}

	// Create the key for the final certificate
	key, err := generateKey("P256")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	{
		keyOutPath := certsDir.Join("key.pem").String()
		keyOut, err := os.OpenFile(keyOutPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) // Save key with user-only permission 0600
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		pem.Encode(keyOut, pemBlockForKey(key))
		keyOut.Close()
		log.Printf("written %s", keyOutPath)
	}

	// Create the final certificate
	template, err := generateSingleCertificate(false)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	derBytes, _ = x509.CreateCertificate(rand.Reader, template, caTemplate, publicKey(key), caKey)

	{
		certOutPath := certsDir.Join("cert.pem").String()
		certOut, err := os.Create(certOutPath)
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		certOut.Close()
		log.Printf("written %s", certOutPath)
	}

	{
		certPath := certsDir.Join("cert.cer")
		certPath.WriteFile(derBytes)
		log.Printf("written %s", certPath)
	}
}

// DeleteCertificates will delete the certificates
func DeleteCertificates(certDir *paths.Path) {
	certDir.Join("ca.key.pem").Remove()
	certDir.Join("ca.cert.pem").Remove()
	certDir.Join("ca.cert.cer").Remove()
	certDir.Join("key.pem").Remove()
	certDir.Join("cert.pem").Remove()
	certDir.Join("cert.cer").Remove()
}
