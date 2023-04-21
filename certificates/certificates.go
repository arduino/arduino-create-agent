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
	"strings"
	"text/template"
	"time"

	"github.com/arduino/arduino-create-agent/config"
	"github.com/arduino/go-paths-helper"
	"github.com/gin-gonic/gin"
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

// CertHandler will expone the certificate (we do not know why this was required)
func CertHandler(c *gin.Context) {
	if strings.Contains(c.Request.UserAgent(), "Firefox") {
		c.Header("content-type", "application/x-x509-ca-cert")
		c.File("ca.cert.cer")
		return
	}
	noFirefoxTemplate.Execute(c.Writer, gin.H{
		"url": "http://" + c.Request.Host + c.Request.URL.String(),
	})
}

// DeleteCertHandler will delete the certificates
func DeleteCertHandler(c *gin.Context) {
	DeleteCertificates(config.GetCertificatesDir())
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

const noFirefoxTemplateHTML = `<!DOCTYPE html>
<html>
  <head>
  <style>
html {
    background-color: #0ca1a6;
    background-repeat: no-repeat;
    background-position: 2% 11%;
    background-size: 10%;
    background-image: url(data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz48IURPQ1RZUEUgc3ZnIFBVQkxJQyAiLS8vVzNDLy9EVEQgU1ZHIDEuMS8vRU4iICJodHRwOi8vd3d3LnczLm9yZy9HcmFwaGljcy9TVkcvMS4xL0RURC9zdmcxMS5kdGQiPjxzdmcgdmVyc2lvbj0iMS4xIiBpZD0iTGF5ZXJfMSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxuczp4bGluaz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94bGluayIgeD0iMHB4IiB5PSIwcHgiIHdpZHRoPSIxMjguN3B4IiBoZWlnaHQ9IjE0MS4zcHgiIHZpZXdCb3g9IjAgMCAxMjguNyAxNDEuMyIgZW5hYmxlLWJhY2tncm91bmQ9Im5ldyAwIDAgMTI4LjcgMTQxLjMiIHhtbDpzcGFjZT0icHJlc2VydmUiPjxnIGlkPSJjcmVhdGVsb2dvIj48cGF0aCBmaWxsPSIjN0ZDQkNEIiBkPSJNMTI0LDMyLjljMC0xNi0xMy41LTI4LjktMzAtMjguOWMtMS41LDAtMy4xLDAuMS00LjYsMC4zQzc2LjUsNi4xLDY3LjksMTUuNCw2My4xLDIyLjVDNTguMywxNS40LDQ5LjcsNi4xLDM2LjgsNC4zYy0xLjUtMC4yLTMuMS0wLjMtNC42LTAuM2MtMTYuNiwwLTMwLDEzLTMwLDI4LjljMCwxNiwxMy41LDI5LDMwLDI5YzEuNSwwLDMuMS0wLjEsNC42LTAuM2MxMi45LTEuOCwyMS41LTExLjEsMjYuMy0xOC4yYzQuOCw3LjEsMTMuNCwxNi40LDI2LjMsMTguMmMxLjUsMC4yLDMuMSwwLjMsNC42LDAuM0MxMTAuNiw2MS44LDEyNCw0OC45LDEyNCwzMi45eiBNMzUuNCw1MS40Yy0xLjEsMC4yLTIuMSwwLjItMy4yLDAuMmMtMTAuOSwwLTE5LjgtOC40LTE5LjgtMTguN2MwLTEwLjMsOC45LTE4LjcsMTkuOC0xOC43YzEsMCwyLjEsMC4xLDMuMiwwLjJjMTIuMSwxLjcsMTkuNSwxMy43LDIyLDE4LjVDNTQuOSwzNy43LDQ3LjUsNDkuNiwzNS40LDUxLjR6IE02OC44LDMyLjljMi41LTQuOCw5LjktMTYuNywyMi0xOC41YzEuMS0wLjEsMi4xLTAuMiwzLjItMC4yYzEwLjksMCwxOS44LDguNCwxOS44LDE4LjdjMCwxMC4zLTguOSwxOC43LTE5LjcsMTguN2MtMSwwLTIuMS0wLjEtMy4yLTAuMkM3OC43LDQ5LjYsNzEuMywzNy43LDY4LjgsMzIuOXoiLz48cmVjdCB4PSIyMy45IiB5PSIzMC4xIiBmaWxsPSIjN0ZDQkNEIiB3aWR0aD0iMTgiIGhlaWdodD0iNS44Ii8+PHBvbHlnb24gZmlsbD0iIzdGQ0JDRCIgcG9pbnRzPSI5NiwzNS45IDEwMi4xLDM1LjkgMTAyLjEsMzAuMSA5NiwzMC4xIDk2LDI0IDkwLjIsMjQgOTAuMiwzMC4xIDg0LjEsMzAuMSA4NC4xLDM1LjkgOTAuMiwzNS45IDkwLjIsNDIgOTYsNDIgIi8+PHBhdGggZmlsbD0iIzdGQ0JDRCIgZD0iTTEwLjcsNzAuNUw1LjUsODcuMmg0LjJsMC45LTMuNGg0LjhsMC45LDMuNGg0LjRsLTUuMy0xNi44SDEwLjd6IE0xMS40LDgwLjZsMS41LTYuMWwxLjcsNi4xSDExLjR6Ii8+PHBhdGggZmlsbD0iIzdGQ0JDRCIgZD0iTTM2LjUsNzYuMWMwLTMuMi0xLjctNS42LTYuNC01LjZoLTYuOHYxNi44aDQuMnYtNS4zaDIuMWwyLjUsNS4zSDM3bC0zLjMtNi4xQzM1LjYsODAuMSwzNi41LDc4LjMsMzYuNSw3Ni4xeiBNMjkuNyw3OC42aC0yLjJ2LTQuN2gyLjFjMS45LDAsMi41LDAuOCwyLjUsMi4yQzMyLjEsNzcuOSwzMS4yLDc4LjYsMjkuNyw3OC42eiIvPjxwYXRoIGZpbGw9IiM3RkNCQ0QiIGQ9Ik00NS40LDcwLjVoLTUuNXYxNi44aDUuNWM3LjEsMCw4LjMtNC42LDguMy04LjVDNTMuNyw3My40LDUxLjYsNzAuNSw0NS40LDcwLjV6IE00NS42LDgzLjhINDR2LTkuOWgxLjJjMy43LDAsNCwxLjgsNCw1QzQ5LjMsODEuNyw0OSw4My44LDQ1LjYsODMuOHoiLz48cGF0aCBmaWxsPSIjN0ZDQkNEIiBkPSJNNjYsODAuN2MwLDIuNi0wLjcsMy40LTIuNiwzLjRjLTEuNiwwLTIuNi0wLjYtMi42LTMuNFY3MC41aC00LjJ2MTFjMCw1LjEsMy43LDYsNi41LDZjMywwLDYuNy0xLjEsNi43LTZ2LTExaC00VjgwLjd6Ii8+PHBvbHlnb24gZmlsbD0iIzdGQ0JDRCIgcG9pbnRzPSI3My43LDczLjkgNzgsNzMuOSA3OCw4My44IDczLjcsODMuOCA3My43LDg3LjIgODYuNiw4Ny4yIDg2LjYsODMuOCA4Mi4yLDgzLjggODIuMiw3My45IDg2LjYsNzMuOSA4Ni42LDcwLjUgNzMuNyw3MC41ICIvPjxwb2x5Z29uIGZpbGw9IiM3RkNCQ0QiIHBvaW50cz0iOTkuOSw4MC41IDk0LjUsNzAuNSA5MCw3MC41IDkwLDg3LjIgOTMuOSw4Ny4yIDkzLjksNzYuNSA5OS42LDg3LjIgMTAzLjgsODcuMiAxMDMuOCw3MC41IDk5LjksNzAuNSAiLz48cGF0aCBmaWxsPSIjN0ZDQkNEIiBkPSJNMTEzLjYsNzAuMmMtNS4xLDAtNywzLjctNyw4LjhjMCw1LjYsMi4yLDguNSw3LDguNWM1LjQsMCw3LjItMy44LDcuMi04LjdDMTIwLjgsNzIuNCwxMTguMyw3MC4yLDExMy42LDcwLjJ6IE0xMTMuNiw4NC4xYy0yLjQsMC0yLjYtMS43LTIuNi01LjFjMC00LjEsMC40LTUuNCwyLjYtNS40YzIuNCwwLDIuOCwxLjEsMi44LDUuMkMxMTYuNCw4Mi43LDExNiw4NC4xLDExMy42LDg0LjF6Ii8+PHBhdGggZmlsbD0iIzdGQ0JDRCIgZD0iTTEyNi4yLDUuMWMtMC4yLTAuNS0wLjUtMS0wLjktMS4zYy0wLjQtMC40LTAuOC0wLjctMS4zLTAuOWMtMC41LTAuMi0xLjEtMC4zLTEuNi0wLjNjLTAuNiwwLTEuMSwwLjEtMS42LDAuM2MtMC41LDAuMi0wLjksMC41LTEuMywwLjljLTAuNCwwLjQtMC43LDAuOC0wLjksMS4zYy0wLjIsMC41LTAuMywxLjEtMC4zLDEuN2MwLDAuNiwwLjEsMS4xLDAuMywxLjdjMC4yLDAuNSwwLjUsMSwwLjksMS40YzAuNCwwLjQsMC44LDAuNywxLjMsMC45YzAuNSwwLjIsMS4xLDAuMywxLjYsMC4zYzAuNiwwLDEuMS0wLjEsMS42LTAuM2MwLjUtMC4yLDAuOS0wLjUsMS4zLTAuOWMwLjQtMC40LDAuNy0wLjgsMC45LTEuNGMwLjItMC41LDAuMy0xLjEsMC4zLTEuN0MxMjYuNSw2LjIsMTI2LjQsNS42LDEyNi4yLDUuMXogTTEyNS4zLDguMWMtMC4yLDAuNC0wLjQsMC43LTAuNywxYy0wLjMsMC4zLTAuNiwwLjUtMSwwLjdjLTAuNCwwLjItMC44LDAuMy0xLjIsMC4zYy0wLjQsMC0wLjktMC4xLTEuMi0wLjNjLTAuNC0wLjItMC43LTAuNC0xLTAuN2MtMC4zLTAuMy0wLjUtMC42LTAuNy0xYy0wLjItMC40LTAuMy0wLjgtMC4zLTEuM2MwLTAuNSwwLjEtMC45LDAuMy0xLjNjMC4yLTAuNCwwLjQtMC44LDAuNy0xLjFjMC4zLTAuMywwLjYtMC41LDEtMC43YzAuNC0wLjIsMC44LTAuMywxLjItMC4zYzAuNCwwLDAuOCwwLjEsMS4yLDAuM2MwLjQsMC4yLDAuNywwLjQsMSwwLjdjMC4zLDAuMywwLjUsMC42LDAuNywxLjFjMC4yLDAuNCwwLjMsMC44LDAuMywxLjNDMTI1LjUsNy4yLDEyNS40LDcuNywxMjUuMyw4LjF6Ii8+PHBhdGggZmlsbD0iIzdGQ0JDRCIgZD0iTTEyMy4yLDcuMmMwLjItMC4xLDAuNC0wLjMsMC41LTAuNWMwLjEtMC4yLDAuMi0wLjQsMC4yLTAuN2MwLTAuNC0wLjEtMC43LTAuNC0xYy0wLjMtMC4zLTAuOC0wLjQtMS40LTAuNGMtMC4zLDAtMC41LDAtMC43LDBjLTAuMiwwLTAuNCwwLTAuNiwwdjQuMmgxLjFWNy40aDAuNGwwLjcsMS40aDEuMkwxMjMuMiw3LjJ6IE0xMjIuMiw2LjZjLTAuMSwwLTAuMiwwLTAuMywwVjUuNWMwLjEsMCwwLjIsMCwwLjMsMGMwLjUsMCwwLjcsMC4yLDAuNywwLjZDMTIyLjksNi40LDEyMi43LDYuNiwxMjIuMiw2LjZ6Ii8+PHBhdGggZmlsbD0iIzdGQ0JDRCIgZD0iTTM5LjgsMTE2LjVjMC0yLjctMS4yLTMuNy00LjEtMy43aC0zLjV2Ny4yaDMuN0MzOC4xLDEyMC4xLDM5LjgsMTE5LjIsMzkuOCwxMTYuNXoiLz48cG9seWdvbiBmaWxsPSIjN0ZDQkNEIiBwb2ludHM9IjY3LjksMTIyLjQgNzQsMTIyLjQgNzAuOSwxMTMuMyAiLz48cGF0aCBmaWxsPSIjN0ZDQkNEIiBkPSJNNS41LDEwMS4ydjM3LjZoMTE1LjF2LTM3LjZINS41eiBNMjAuMSwxMjkuM2MtNSwwLTguNC0yLjctOC40LTkuM2MwLTYuMSwzLjUtOS41LDguNi05LjVjMS44LDAsMy43LDAuNCw0LjgsMS4xbC0wLjgsMi4xYy0xLTAuOC0yLjctMS4yLTQuMS0xLjJjLTQuMSwwLTYuMiwyLjctNi4yLDcuNGMwLDQuOSwyLDcuNCw2LjIsNy40YzEuOCwwLDMuNy0wLjUsNC42LTEuMWwwLjYsMS45QzI0LDEyOSwyMi4xLDEyOS4zLDIwLjEsMTI5LjN6IE00MC40LDEyOS4xbC0zLjktNy4xaC00LjN2Ny4xSDMwdi0xOC4yaDUuOWM0LjYsMCw2LjMsMS45LDYuMyw1LjZjMCwyLjUtMS4zLDQuMy0zLjQsNS4xbDQuMyw3LjVINDAuNHogTTU5LjksMTEyLjhINDkuOHY1LjloOS4xdjEuOWgtOS4xdjYuNGgxMC4xdjJINDcuNnYtMTguMmgxMi4zVjExMi44eiBNNzYuMiwxMjkuMWwtMS42LTQuOGgtNy4zbC0xLjYsNC44aC0yLjNsNi4zLTE4LjJoMi40bDYuNCwxOC4ySDc2LjJ6IE05NS45LDExMi44aC02LjN2MTYuMmgtMi4zdi0xNi4yaC02LjN2LTJoMTQuOFYxMTIuOHogTTExMi4zLDExMi44aC0xMC4xdjUuOWg5LjF2MS45aC05LjF2Ni40aDEwLjF2MkgxMDB2LTE4LjJoMTIuM1YxMTIuOHoiLz48L2c+PC9zdmc+);
    margin: 0;
    padding: 0;
    height: 100%;
}

body {
    margin: 0;
    height: 100%;
}

.container {
    font-family: TypoNine;
    font-weight: 400;
    border: 0;
    color: #4D4D4D;
    background-color: #ecf1f1;
    display: block;
    min-height: 100%;
    height: auto;
    margin: 0 auto;
    min-width: 600px;
    padding: 2% 4%;
    position: relative;
    width: 65%;
    box-shadow: 0 0 10px 0 rgba(21,110,114,.8);
}

h1 {
    text-align: left;
    font-weight: 400;
    color: #4D4D4D;
    margin: 0;
    font-family: "TyponineSans Regular 18";
    font-size: 2.5em;
    -webkit-font-smoothing: antialiased;
    text-transform: none;
    letter-spacing: .05em;
    line-height: 1.65em;
}

.image {
    text-align: center;
    margin: 2em 0;
}

p {
    line-height: 1.5em;
    font-size: 1.35em;
    letter-spacing: .05em;
    font-family: "TyponineSans Light 17";
}

    </style>
    <script type="text/javascript">
       WebFontConfig = {
        custom: {
          families: [
            'TyponineSans Light 17',
            'TyponineSans Regular 18',
            'TyponineSans Monospace Medium 3',
          ],
          urls: [
            '//arduino.cc/fonts/fonts.css',
            '//arduino.cc/css/arduino-icons.css'
          ]
        }
      };
      (function() {
        var wf = document.createElement('script');
        wf.src = ('https:' == document.location.protocol ? 'https' : 'http') +
          '://ajax.googleapis.com/ajax/libs/webfont/1/webfont.js';
        wf.type = 'text/javascript';
        wf.async = 'true';
        var s = document.getElementsByTagName('script')[0];
        s.parentNode.insertBefore(wf, s);
      })();
    </script>
  </head>
  <body>
    <div class="container">
        <h1>Oops, this is not Firefox</h1>
        <div class="image">
            <img width="418" height="257" title="" alt="" src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAaIAAAEBCAYAAAA6g6EvAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAALjBJREFUeNrsnQl8FEXaxqtzECBgggoiisQbREkADwSVAUQuIeFQV3RN8FhRUcK6rrufuxJ21+PbVYmLunisCavCp1yBRQQUmHCjHEFEERQS5JDTQBICIcl89Xa6w8yke9LdM9PHzPPnV0xmpqe7q7vqfeqteqta8Hg8DEQunzw99J7q09X9q8/UXFFx/FQqfXb8cEWymecw9Imehn87bvK3tr/Gb0y4xvBvz92y0Pb5O5Y62MzDFUqvbp6KeSrqPGJKEWpyZBOHSxBZzHpm2C2nKqqePll2useJIxUXVFfVCLgqwEH09ntl2+Y8eVwSJkoFXJiKcZkgRMCm4nNk7/E7+SvuKYg0knhKl9JkLkwl/DUXogQhAjaAV8iMFZ9syTO7qw0Ai+lAgiSJEnXl5XJBKsBlgRABcwUoi7/kUIXkIoQLAqIZ6sLrLXlJOVyQ8nFJIEQgzB4Qq+uS6ICrAUADLymP1xFqoGXDQ4IQgdALUJokQL1xNQBoVJDmSl122Yi4cwYxuAS2FyFq4W2GCAGgC6ovm6X6A+ARAYMClMJfqHshFVcDAMNMlLq0MxBhB48I6BMhqjhFECEAQgLVoyKpXgEIEdAgQjQWNJfVzZ0AAIQGqk9zpfoFIEQggAjl85fxZhxr1y9l7IvdB9i3h0tx4YFlLP1yF3vr4y/ZV9v2mXXI8VI9AzYCY0T2ECCakOpmJnXFkQAt5UmmV/s27M4rL8aNAKby8vsr2Yefbql78wljj999I3v8nhvNOHSmFInq6jxiClpi8IiA2SJErP7pUMD3AJhBvQhJ/GeBqZHWVN/cUv0DEKKox1QRIk5V1+CqA9tRfrLK7EOmSvUPWAy65qz1hvLtIkIXXHE1S7rgQtaGv8qcLi9jB3/4nu3ZshE3CxgisfWVrGnSxSw2vln9Z5XH97IzFccUty+rOM1aJiaYKkZUDzuPmJKFuwUhikYRouidTNOPqxCc0OW669iYt6cH/N3O1W62g6eti/+LmwcCktzhJnZOuy6s5YVdAm7XpctH7Ouvt/p8RsELGX06mX3KNGZUysUoG3fPGtA1Z40I0XyG8VYce+OBow0+69qtW6O/u7KXiw35fQ57bPoCdt2AobiJQFGArho4iV3U/f5GRUgsd10blruC5dutOv3xmGcEIYomEUrhL/lWHJtCtXeXljf4PCNDe/2j7jsSpNGvvcOatmiJGwrEbreU28aLAhTf/FzNv1Mqdxu27RO9IovIl+ongBBFPLRsj+mTVQ+UV7K5uw42+LxDhw66hEjmktTubMw7M8SxJRC90PgPiVDi+Vfo/i2VOyp//vz5rRVs++4jVmQnSaqfAEIU0d5QDrNg2R6avPr+tj2svKLhs4uys413i5N3BM8o2j2hp7gYXWR4Hzk5OQ0+O1FWzh6ctMDMSa7epGKhVAhRJIsQTaCbaLYXNPO7Evbu5p2KIpSamhqUEBEJiS1EMQLRB3lC3tFwRsjKyhLLoZIYjXl+LntuyhdWeEcTpfoKTAJRc+YR9jWuKBBBDkb45VQVKz2lPi8jKSmJFRSEpheizeVXsVsyH2Wrpr2NuxwltOk0OChPyBsqh2lpaez48eMNvpvn3i6mdq1bsnZtzhE/y+jT0YzIOqqvLtxpeESR5A3RIExYnydEAjSLez8UjECpMRFyu90sJSUlZMe/5YHfsKS27XCzowAKSGjdaVDoPCteDqk8UrlUY//hMjGQgdKf3ljKCpZ/F+5s9kYUHYQI3pABIdICdYNQpacWaKghrwhEhzcUaqg8UrlU6qZT9KLMCfPGSt0QoojxhrJY3eOLw0qrZo3PRp84cWLYRIi4qpcLNzwKaNmuS1j2K4sRldPGOCexiRlZ7SDVXwAhcjw5ZhzklvZtWNO4WEUPaPLkyeyXX34RI5SSk8O3xiMFLkCMIl+Egg1QCASVTyqnVF6p3Cp5SC2aN2GP331TRNXfaAfBCuH1hjLM8IaIC1s0Y8/2vJbtLzspvk8893w2cdZC3cJDLVJK9DuKaNL7e1qrjpYCApFJsyT9jwspLS1l+fn54qvL5RKTFkGiiE45qvPDv2ey6tNl4t8dU843cz068oroMeOYXwQhciymrl1FHtFlrerm9FzS5VrdIkLCM23atLNNQd4yLSoq0hXUQBNdQeTSvPWVurYvLi72iYibNGkSy8zMFIVJl8d/c3dWcXinlfUYQhRG0DUXPm+IrHdvq46vd8UDEhxvESLIeChNOARAK1R+/MOyqZxRedODnqWDwkBvLP0DIXIqWVYePEHnagdqc4qoRQuAUdTKj945bE2an2d1VhDKDSGCEIUbtUi6cAY3gOglXJGbYQSPiIAQOQvJje9g5TnsKdqgr7mXkaEYoRTsEkAgulHq2qVypneh3YojO63OSgd0z0GInIblbvzpinLdv6FoufHjx7PevXuz9PR0tnz5ck0RTt4c+nEH7n4Ec6p0r67tqfxQOaLyROWKyheVM73UVFWiXkcwiJoLDy6rT4Ae8U1iRHN7tELdcLm5uUEfF0SwEB3fq78yaAzZVhWhM5WGjhumeo3VFuARQYj0YMV8np2YQxTRlO3/OiqOaed6DSECjbJtzpNdmYkPvhMEQTWZLQo71xSKXpj/eQRVQAX7pyBvoP2Tn3dSdsBcYSjds94u1TtJqt8AQmR7zA0HEtTTzjVuduLgAdNOZcOc6crnEiahtUuKNo7+YF4D58zJY1ZOZLW+fkOIgEEuNVWHGmnMLnwlxzRv6KevNzbWoI50h0E3Hq7Udk/+kDCY5RXt2/hhVNdvCBEwistcIQrcWv/p603sm88XhPUcqDvus1cnhcVjiHQhcir7NnwodtOFtUuuZL3dvCHT6zeECEQMy6a+xg7tCl+FnpvzO3a6vCw8QuuAFI1KSyL007p3G1rpZV+w0jNVQd/3U8f3sZ+/no3KGyVERfj2jI9nujweTwr/M4W/pvGUzBN91Zte5SR2lWh8r/Zdnws8puZNi8dRxT2Wj38/lv3q71NZ68uuDOnxF736F7Z366bA5+Exfk1iYhzgcniCuefOzV/FoR2iZ3TR9feL7/N372KFhw+xnG+2styuxhe/JREqLswNu8dlkFv/+cZbHn+PX+19oO8U3hdK70t5KuKpmL8tHv2ru90QIocxc/ZcEhmXnPhHqabWWbMzrNGOna4oY//HxWjQ08+zK24Ofi1W6o4r+Msz4rhQY+cQ4WY6qPx5HJ6/X0rWsaqTR9k5149h2Zs3ip+9vuN7ln1VR5aSmKj7WCf2f83F7QO7ilC48a6Y6V4NaRKnLfxPN38V092jRpRCiGzG7LnzyNPJoMQsXPHaCvSMUVSdLGPz/voM655xL7v5vkd0TXb15od1haInROIW7jGSaB2DcRI0jvPqZ++x49Vt6z/LWr+WufvernkfJDyHvl3Ijv6wHBdUmVQpjZca3OQ9FXBRKhg1IqMYQmQhcwrmZ7G6ZTfSUU61s7FghhjA0D3jV6xz/zvZOW0u1PS7bV98ytOCOi/IhkLrRDzM+Rnc70lgb9a09fmMuujchw4yV5sLAv6WwrPJqzq6c3m0ekHBeE+UJvOG+DwuSPkjMoY59plJjhOigvkLaDnobO79kAh1sNXFjKkJcrzAiKE2ZsiqTpaztdPfE1Pry65i7a/rxtpcflUDUfpp6yZ2eNcOMfqOPCBDxwzimjhCiIK55xGQv7dq2yt+/viqz5m747ks0e9hemdOHmWVpXtFT0rv2nVAEWqIp/OGeQkJEk+5GcPudFTXnWOEaP6ChTT2Q0tBU0qybZ01+4AhMGSHd+8QU7iOhTEip3ttAbwhlsDmedoofvfdmRj2zravWLrwKaTCHKhhPpFsJG+w55IgpQ8d4ghBckT49n8//YzEp1i6yLYVoeraWFSFkHt8mNBqtwmt3izzBH5g3XZPIgqx+SRJtrKYN+Ad8RwXW3tECxYuosg3Wu021TFFwOzmcYS7DALuub3z10jey8jERPr1sbcgTeYN+SzykoYOGeSGEOlg4aIl1A2Xw6QIEWfVWZPHiCL8mjhhjCiY/Hkcfv+uF0oDZuJ6dtz0OgEaQA355bxh/7ogCDlDBg2wXXed7brmuAjRooJFThQhS4jwpQcif4kfZ9/Aq1kFu0/Yr/pdunAQddQ+kE0tkmwsPCI1Plv8OfVnTg7T7rdIAlfMk+yiFj9w/+jiUB7k608eW8vbfz3gE3m1qD2R3SIOJn+eCMjfM+xH1pJVs3nsAnaAJbAW/O90dpCNFfZE3L3ntW3dU+MevznU+/3Ph9NT+EuK9NYl/Z3GQj8sQQENm7kYjRk88I58CJEfi5Z8QRclM4S7LOGJ4uoL7r3nLrd53RhmV4zINvKRPp81Uu7eWKGEjRWrHO6XEaQGsdwo9rFXtEQZq5svmcFCN2Uljzf8XYMG9M+CEHEWf740WbrwoVB+WXxy77lrZLFFWVrJm4HmeURCZK/FFvHziBjy56yWkbDS7ENKDWlK2R/PnE2eUnaIRCmTOwBpgiC4BvTvZ+m4kaVCtOSLZaESIVruIveukcMLbFBpi1FvQ2fHIl2HnFBWUJ59eiCKrTy+1MAW51POnD03Q/o7mGXNyPa6uUNgqRhZJkSfL10eChEiAcoZOTzdbSPXfYOpHS6C05fNjACPKMKj5hB/7X0lhA12ORep4V0we+48F9nBIARJFCPuGLjuuL2vJWJkSdRcCESIuuCGj8gY5uLJbaeCmnbP1C95C7KcWpFmJKd4RMbz57F9wv2LmlRO9dtu94ca4mQLySZKttGwGEm2OfKF6Itl7mBFaBLZ++HpQ228wJ9nnYeZ888J0b/B5C9cIdd9+g9gnbukBrWPFi1asLvve4DFNGkWzB1kTVNuZC3ShttZavGv7kqss3ODQVr0NE2ykY4SIyu65oyKECl9RsawO4sc0IhcwZtPt5txIGf0XHlslb/nX36NXXNtF/HvDevWsFf+9rzufdzQoxd7bMLvWfPERFZz5i5WsjaPle7drHs/be95gyV27Cv+3TI1g+3Lz4yo+xdRCMIKu58ib6BT11pOwfwFYrcd0x/QkCrZaFPnGpnqES1dXphvUITm0YVJHzrECSJE9fZj07oLnNB5FUT+Qu2g3TX6gXoRIq7v0ZMNTh+pax+JiS1Y5m8eF0WIiI1vxi7ufg+LiW+uK2+tr769XoSIZik3sHNdT0TU/Yuw9LFTNFNqsKdJtlO3Z/TFMnd+RArRMvcKiu4w0tybNHTIIH5dBztmWfNuo9/ZwYvtATNqhiPm5Qc1iBK61PqCC9ggLjr+jOLiVCcq2vYz6r4H2Pl+z9lpkngea3fdUM35IvFq12VYg3NJ7vEAi0++CINEdku8PtfVa+dAK29zu5lhsKsukzsO2RElRFyESJmNrJgwhotQjkMd+anoVQ/+XygFUe5K84c+e5x/p2UfKZddwQYNG6F4w9t0vJ01a9VeU75Sbh4jilGDCtm0Jbsg40V72WH8E+uzU3sUJRs6xsBPJ0u22/lCtLxwZbLUV6lbhO4cPDDfqTefmdQ955QhIsP5C5EKXX9zL9bpWvVe4e49eonbNLafsVywApHS88FG85R8cVeW3L6r6j6oiy6xYz90zdnKIXJOt5yKGOUbFKMCLkbJjhciVhffrnfAbMyQQQPynXzju9//3ve89G6ktbbCmZwhRMbzF4on6iRyj2ds9rM+57Rn94/syCHfBTkfeOQJcVu1/VAXXodLL/f5zXffbPH1rrhHRF1uavmJ4V5Q+xvu9fnNmdL97PTP3/t8duHwF0XvyB5C5InqVHOmqriuPjsbqWGvV4w6SDbcuULkXrHKxfSvom2rxfiMUvBsataxvcUpcImY5WNgJDD+XXIfvPsmezv3f30+o3GfUaMzFffRpk1bNnDYqAYi9MIfJ7Cyg742qm2nO1hC4nmK+bkoNb3uOy/2z/0f9vNnLzXoomsz6I828Qai2x06tnd3yqzfdiqe88y1WU63S1IDX68YjV9euNLlZI9Ir6BMcLoILfjzja75z3UvYoKQV3Hs8Hk11WfC7BE5YcKndfnrdF0qu7XfAJ97tHj+bPbd1s1i2rh+tc93A4aNZB0uu7zBfn4z4dkGYvbO5JfF735c+a7P57FNmrFLez3UIC80ftS2U3+fbcu2L2MVu78U07G1H/h8l5SWwZpfeiM8Imu9IcbrMa/OAvcMhLyCP6QV8frtcrKNkmzsBJ0/y3WkEBWuXJ2ts0tu2qAB/XOdenMX/fXW5IWTeubyErucF9hUXnCp8LKywwfCK0ROmNBqYdfjo9l/8Hl/sqKCzZ5+tq3zNhcT+izQb7r3uKXB+NKcGdPY4UM/i3+fKjvM9hb5DoOe07Yja9W+m09eLr/lYZ9taqoq2b7Zf6j//tCyKazmVJnPNu2Gvxj1XWNWpuMH99bPYJbqNC8IwvIFz9+Uy+t7soPFiGztNB0/SXWvWJXtKCHiIpSss1+RJqtmO/WmLnmpTxovoUW8kI6XBUguvOVHDrJwekVOmEhk1RjRyNGZDcKs38l9iVVWlNVvQ39/+O4Un20uufRy8bfy+JK/MNH40tzpefX7oPP8adMc3nLe47Pd5bc+LEbG0fcXk3dz7iU+3+9e/yGrriyr7wWiv/fN/qPPNvHJ7VjrPuMs7p2KXm/oxKEDdQLEBH9BorpetOhvt6Ux55LN9C0JlCPZdsd4RJTBJB3bZwy84/ZSJ97JpX/vn8UL5GbJdT9bUNlZQTrBW1Xh84jsn6zwiDpcdgUbfq9vl/6m9avZxnWrGmy7cukitt0v6GDAsLu4iLVl9z8yrkGXnL9wyedavO5Dvy665qzDTfeJc4wu7OzbPXji5+3s0I4VDXT7+HdL2QmevGnT9wmWcGFHdM2ZnEr376mrwyr1mv/Xger+5y/3deTY0aAB/cnmZuj4SVK4HIaQC9GKVWuSdZ7spAH9+xU58UYuf21QPo0FMVbfSvLxhuTCe/KXo6z6VGWYvA1HhL4aTka9IRIQ/y65D9/5p+r270x+0Wd7Ep/nXnqd3dpvoN/40iy2fetmn9/K53p8/3fswDeLfUXkylvZtUOeE0XJm52Fb6vev/2fvtSwi27w/zjy/jk1namsYOXHDnl7QA3qdf3nTMhb+g9njm1zB4Bsr54Jr9nh8IrC4RHp8YZKuAjlOO3mrXwjPbnw9aFU8DIbFFJvUfJy508c3FdUy2t1qJMTarXZ+RuQPop1vNa3x2TujPfZkUMHVH9D382dkefzG/9uPRKzudPfD5i/ko2zWXXVSZ/fJbQ437drb9McVnnisLi90r/Tv+xlB5e+4fObxEtvYOf3fMCSyZzhKLd2T+XHDlNPR6FPlxzzER9/Ucqkhim3C44bN5JssNYuurB4ReEQoqwwbWsLVv9rZDIvdG5e9jLPChCTCimTCinzLqAlPPW58y9ru8YIwpvMU8tCmpxAEPnTGxtB68CNuPdBvzGdH9iSeTMb/W3B9DxxWzXey32RVVaUKyxhdPZ8q0+Xs51u9Un4FUdL2J4Ns+q3V/NCDq2exioPbPf57QX9xrGYhHMs8IhqoypRPR3y/Mppd7++w8UrcB+pDnt1yTGpbjOpsVlf37lNYO5Vb2Y4MYjBUrsdUiFauXotnaDWSLnCO27v63bSnVr37j2iCFFU3FmhkcRHLJ9n39c1moRJGS8XpYz4x1Yxnz0f/mAc/3AXwrfDl78B6XexZn5jOu/mvqj59x+9+0/Fe79p/Sq2cd0KTfk7svsrdrRY+flpP675j9/26gJQMtM3cCG2aUvWutcDCFYI63QIYVddPa1j1Gvfunn9TeGfT2KCX/2WRMjXDohRde41U0c5Soy4V0Q2qlDrEOyKVWuybCtEOge+HNclxwtYLoVv+raOBKV+ZO7mCl2HvbChQR5jYmOG8uJ+MmSVxwlSZKIQfbd1k8/1LpjxPtuza4fm32/nv18yf6bPPsgL+uid13UJ7Y/co/Hvotu39TNWum+b3/1T/3fywLfsgF8XXdmu9eZ3zkWLEPF6SfVT0bC9vDmHV+2uvF6X+I8TKdgBshFOnIqSEyZbb54QrVqzjrcaWLpWb6h/vz6O8oa+yrs/n1zvxsaC+L9C/lnanX9ZrxiA0fOhD76NjYl5KnTh2/ZPZuZvz487znbHzf+ELS74WPc+5n70Htv+zeazIvTu6+zIwf268ld54pCPGJXu/5YVfzWzwXaNjVX8vCqfHeT7OSl105Xv/8708ZJoESKql1Q/1WzA0L9toCkaaRrHjjLXv3evowIYpB4qrV5R+srVa1NC1sQP1XplXIiy+b4me4fd+ofher0fzoWowCk3aON/srJqa2vyamtqmIcn/jejv2ulv70+m9b/D8s0uawr3753ek1Nzb3BntumJTttf/263XGl4d/mLdxn2XlfctmVXIAOsJNcjAIxZrD6YxviEhJZ05atWfmRYsXvE0vsHzBa0cHJU2W0ERsbO+PWR2eM1rr94hd658fExmbyxGJiYpkgvYrvvT+LjR1zQ9YHjhGkz5cuJ09nrm9UIGMq7yfc2uvmkHh+oeya09pnWOIkEdo8/WFqAeX5u+FCQ9dcswgRVOgFIeaz2lre6gwiOeFfMPkLxaKnRtNPu3b4TH5VS4HOv6qynJ04tFv9/jnAqQ22jNo9UT3UI0LimMpzhRRVN82/a15o2E2Xt+mDMY5Rcsk2l4TY5psjRNwbooE5rU9edUzr4OuPx1JwQkGAsSBZlKb1e2aJ7pvS+7H/G8x3sRvBCtGbPyRrE9U/qodG7MPtzy4VxUglnNt77KigaMYjTgpe0GqjUyXbbxuPyBWGTFqPQOdKs6eZ2jIfogi5Jiw03DKIi4vvEawYRXQUU4SPgdWSx2HzFMkiRPUvGBPR93eLs5gkRv6N1LNjRxRJLORHoBDptf22EaItt/d1FTvhTmybPS6DF6J03zDNBqGbhbc9NT8o9/TWR6cfio/nYsRYsae2lulNTsBIvs7mzwEeURD5c4TRDiZ/Nk1U36jeUf0L2vhlfypOflWewlH/WfrWmY9lOKG+9u/Xh2z0FjOFKM5kIXI7QoTmPJnMKPzSr+/X4zNZlZUIQkxICtatj86gynDp0imjPquprh6oy8g74JlETnmAnxX5kx5DjftnIrFxcYv6PTlrUIh7T3jDNaZI9H4aTmqXPaVcblvcnUdMccK6mm6Nwy228oi0jg85JUhBfISFd1ecQrh2Rq/HZoe0QFHliI2N/QRjKF7dJw5I6JpzUIg2r18hFyFqTI6bV7eAaKClgBgtjOyYpwwUhNj2h1eIVq9dr1kR+/XpbXuPiLdYUvjLRJUWjfzZhJt/80lY4m77PTX7nvgm8WN5pamsra1ljSUnVH4t+VDNnyOiAoPIX4TfP7skqk9Ur6h+hct28IYpzTOaIKgtjlrnKU2UbIytub2vS7OtXrVmnctyIeJovahbmDPIqfe2lZeAL7zxoelhnTXdd9ystxOaJtwQExNTEu4Hx5nVtYOoOSSrEtUjqk9Ur8Jd1ns88jF16RcqPDLi7DqUzllVZkuINSCsY0RaT8L2M/eklkqmT9evVKA8Z1s5prjWfZ6YuY2u7ee5w9+tOn36YUePEQUTVBHh+auN9PtnMU0SEt7rnz33ETOPKZCNEITN/tFzXmRyW5PTecSUYptfviKNXW9BC1EoPCKtk7WKHVBucxRKlbd7/fr1D0wzVVCpEjVt1qxvTGzst8rrY9XaPwU1RlRr+xRcRJoDkhO9IF5fqN6YLULEDWM+LOLW4nWFbrnAtsZ+FIdYA8LqEWmd0GTr8SExUk5hIb+z0XLCcf7GksLT76nZy/lL5yW5wx+rOnX6r7W1teedba7avzTT7HXDrXEW2fmrjfD7ZzYxMTFHmzRN+PMd2XP/ZeV5CGQrBCGL/5XEhAYeEZFBNsfmEXRksyeGUANsIUR2hxca5Qf6Sa2a3K6j37O00EiV61+LX0t/rur06QkkSM4I/w2ma642ovPnjPB7+98DUYASEiYP+O28F+xwPt1+nVdaNOMRWq1/ooIIMcnWkM1x4irdthQireF7du+ayw7QvOHekH0KjFTZXlj02rDHmiU2feVk+cnm9m5RB2PI7G+oayN8jKjWxmNEcXFx38UnNJky8Lfz/2W/s6O5iIzGi5IC2Bw7C5FWmx10CHeMWTnq67rNtkLEXWQXC/BAP96eKehy91TbudBU+bI/2ZLY/vL2A1tf2LowPj6+1o6BZcGMTzhiiZ9IH3+x2ZgV935+SmjadHJiixZdhv5x0TX2FCHG0u59p5TWqhTUN+kg2R5b0q9Pb9NsdhwDcrdcgIaNNWNDWvn1lGWL+Qsl9sGTfQdUVlT+sfLkqdST5RVJtTW1guUt6qBa/Q7wGDwei66NE+5fCPwKQTgVHx+3NTYubhX3gPIG/HbeVsdYlrqxosxGbI872g0whKiOQEv1bHFAmKWiKBH5j/d+9EzVmQHVZ6ov4wIlPhjI7K68YMJ/Iz083QmR0WaGb8c3iRefAMhFZ12MELNLiIlZM/iZ/y51qmG5btRbxdzr2RKg+yqDAQiR5BonBdjE0YOJWW8V0iS+t516/nehjgLnQzYkT+W7JLJBvLEb1V5RDMpIoy2SAlwiAEAQNGZDot4rghAFXj220CEr5QIAbIpkQwoN2iAIUaQjTWJNhTcEALDQK0qVbBGECN6QIm7UHwBACGjMlkS1VxTtQhRojaTj3KUuQv0BAASLZEuOG7RFEKIo9ojgDQEAzPKK4BHBI1IE3hAAIJQUwSOCEPkgDQ4mwSMCANjAI0qK5oCFaPaIGmuBwCMCAJjlEUW1VwQhUgHzhwAAoUSDTYEQRSGB3OBCVBsAQBgoNGiTIERR6BHBGwIAhINSeEQQIq2tD4wPAQDCQRE8IggRAAAACJH1SGGSNwXYpBxFAwAQBgLZlpuiNYQ76oSI3+gcVvcs9qYBNvuHtB0AAITS9vwjwCZkk4qj0fbERVEhoJYGrYDbW+NPJvLf0HNCXAjlBgAEaXvcLPBK/zJJku1x8deMaLE90eQR5eoQIZlUhkdBAACCo0CjCHnTmzn86dAQooYtEvJsMg3+vDf/fTbqEgDAgO3JNtAAlsmUbBeEKIK8oWDIQZUCAFhgO6LCK4p4IeItCpok1iHI3SRFS8sEABAy20M2IynI3XSQbBiEyOGESkDSULUAABbYjIhvBGNCK4QIAACbYSlxuATa2HCmRfq1U9/34EoAALSQ36oFuz4ec+PhEYWQL6ta4CIAAGAzIESGCMkCpvtrElBaAABW2IyIX4Q5GoTIzdPxYHey9HQSahYAwGybcZwFfsQ4hMgJSEtkBBWLP+/UuazME4uaBQDQDNkMsh1BkhsNy/xE04TWEiM/LOeF6eWyi1GrAAC6IdtRbrwRW8IwoTXivKIMprOLjgpQ1i9XwhsCABj2isiGGBAjslVY9DQCxYgG/Fw8bdGy/f6aJmIB2l7dDLUJAGAYsiFkS8imaIRslEuyWVFBVM0jkm5s2rY5T2bxV1qMsMGKuN/zQvPByTasIPi+XQAAqBejO452ZhlNj7FfNz/Ero6rVBMgGhPKj7brE5UTWqUbnS89J0Sc/Tym9Mrl1GLZp73VAgAAuqAGLqWLYqtYO57yknf2kb4qiubnnkX1ygrSjXfT319NfR+1BABgCvukRi+3QW5cDaysAAAAAEIEAAAAQgQAAABAiAAAAECIAAAAAAgRAAAACBEAAAAAIQIAAAAhAgAAACBEAAAAIEQAAAAAhAgAAACECAAAAIAQAQAAgBABAAAAECIAAACRQBwuATBK1/POZX3btWV9L2zLLm3ZQkxEaVUV23TkGNtdXs4289fZxXvEz5AfX+hY4jH58VolNBH/llm2/+e64x49Jv69mb9GO8lNmrCRKZewruefyy5t0YJ146/0mXx/cL2ci+DxeILaweq16z3yPuhVTv7v+7puE2x9Iaa+7zHjOH/rnsaevu6aBp/vLitn18yer3t/lVmjdW1PFfSX01Vs2YGfDVdYMqBPX3uN+KoFMhCvbP2W/fv7HzQZcKU8/WljEXuV7yMchDs//jx01RViGZCFTgtUPv694wfFY9J5f3pH35Bek2b500NS3gKh9Z6S2PyOXy+lehOue6SWV7XrYhTP2AdtbReXuVeIdlEQhPqk9P6Wnj2Cygc8IpMZwVt0SpBRohZxuFtycqtbNLrd61reZBC0HpeM6Bs9b9TdkiUBptbs4MXLbOUdmZkfuvZv8mN5ez5aofJBxpgMa7R53QsH9BWvuZF71I3//j73Khgem4MxIpMrVaBW8EgVkQonJEhU0ckgh8Noh8Ko2EmEjOZH3taICMk4pYvTahHyZhO66BwBPCKTjX5j3hJ5J2Yjth6vTxMrrZpnROeuZrSp24i6+uhVbr3L4yxKxoVa9lbk06r80PUNZFDpOPIx5eNRS95ftMh7jSbIe1S7ZnQt5OvVqkmT+vE9f+Zw8QYQIqDD4wlV95x3JfXfP52DUuWWuzKGLFmmuE8aQ1EiUD8/9enTPpU+p/EO2dBbgZn5+ch1i+I1J+/mTxuKxN+qlQfaN3lutG81o0rfBRJCpXNWKyOhLG+NEaicU6NMyXuk31BXm9K1pmv80NVXiA0D+lsWeAAhAn4i42+I/A0UCUXQQsSNgpoxHbfmS1WDSi1KOk//yitGkim0NhsbbJa/UzoWGVervCIz86N2LLr3nWbND9jVRveB7hcJUKAuPdou0HkrClGAMhLK8maUbir5fYJfDzVxoWv5qhSgQOKPyDnngDEik1AKUqBKpWW7UEOVVa11rXR8pc/kSq/lWEqGY4QF42FW5Edt7I1a9VrHe8jjCFfEoF1RE14t4kLXlTz7V6LsmkGIQKP4d8vJXS3+xkjJcwoHat1BrRS6kJS6FGfr6HtXEj3veTpW34tw5Ye8XSWBEruyomy8J9wCpSZIAEIEAoiLbMyUDKAZ0XNqhtD/POm90vgGTezUilrkkhmCq3RMs/Kjdh8xgN44ap5PoAAGACECAVBqFcvGTEkQrOy28ketr55WGQi2ZWqFR2RmftTyF6oggUhGbRyIxH7N0IGaphsACBFopGUst4qVhMis7jktJCcotz436fEgjthn0NjM/KjdQ0RyNU6gOVNUPyj0/tuRw0RBgofkfBA1F2YCdcvJrWt67+8FhSJ6LhBqc5r8W+tqrXo9/e9q23azQGzNzE8rBdGLxEgummOlFyrzgQRZXqJHKeLPX5BoDhxFyoV6SoB/VCvGnCBEjiVQt5z3e//twj25Va37z78i0+KSYfNOLGjJmpkfJY+I1vmLOCFSCVEPBAlyY6JBkYIkNo11w9F1f1pai47ESC2yUS/k+Xrny06efaSBrjkLWov+g9VqUVjh6p6TJ0kqtQARzQXsBM2j0tMgo3KNMSR4RMCvpebfUlSa7S1/5t9tZLR7Tq2rpJUUTqzWPUVdIeh+sA9mr0JuV+R5b9RNpyWQh+odddnR4yLGKczVAxCiqEIpSEEtYooqmv8y90a754x2lUTbys7AOVBDjSYBey97pMU7orB8tTlzAEIUFSiJgVrXFwmUvxCZ9WgI2r8Vj2fQEzLtCGNZ3tDT9fc+WyVEXoSXmV6avOwRrdFH68qR2ASaBkCBDEZXLac66V2HEXYPIXIcarPqaQ0svV5VOIWIjEigh4f5V0YjqBkKK8KYzcyPkhDZJSzf6chLMlGiBpy80KlSPSTBehXL/UCIopFQrY4gGs2NofV+xEcOcM8smOfb0HlpDWxQDZm2UQRZOPKj5vGZ4eVGEyQydO/UHrVBY6ZGhIjukXfXOO4ZhMhxBNvq9jZaSitim9VVolb59IReq3kBVjy0zMz8qN0zKhswaqG/r9Rdp/SMqW7nG/NCsSageSB8OwyodcsZxcolf9TmTuiZjKrkQZAnZoUxNjM/akZspI2WcIok1IISsPICPKKoRMnQiIOsaxsPJaUQVaXFM63q45YNrP85aY3oIyMQ7GrXTs0PHUfpmVN0bDoeFj8NjDhBNcD4pZ57DuARRR2K0XIHfq539QMlJYMmd89ZxWyVCbf+UX6Kwnp9mmKL1EojbGZ+1J6JQ6tII3AhMBSAQJNTtXZzq22HFRHgEUUlSl1pWh8zIHbndFfep1VeEbVKlaKS5HXAlM6Ltv2dynwPI33vRtYzUzs3M/NDAqV0LHpPg+uBHhWuFRJRvd23jV1PPWXN6L0hj1Htuo2QHmlP6dM7+orXiO6bWncuiZBaRKrRsGs6ByoTciOQzoHuFzwsCJEjRUhPZZCj2pRWWbCye05tAUr6jIwz5U8enKfxFjIMan3zZk3SVTOoZuaH9qF2LHkFAPKyyCDLwQ60AkZXnWNWgRYHNXI9dQmRwXtD101NiPz3R/eEEl1PqiPe14q2VbtedK+NTNSma+ovbHIjBKs1QIhsj1KF1LLAo79oPdTSt+VtJHoulJBhopavUv7EhSlbalvbiyqxHSLGzMxPY4t3ysEtIxDE4NPwUhMIPR7gE/z+GPFgAj36HUIUejBGZEIF0tsNFai7wkpoiZVgwlmpAttpuRUz8xOKvAfqyookQjEeSuJD9zfUY5HoloMQ2R65X9sfvfNl7Br2S5VwyJJlYpeKngpJ+en530W2W/PL7PyQGJFx1OvVkjGl39Ixo2H+EV2fTrPmi/fFSA8A3Rf6fTAipPbb2Yh0DAuCx+MJager1673yPugVzn5v+/ruk2w9YWY+r4HxUG/8NL4CXUb0qRBWYTl1cTlSEGnGE8z89NVGnei7kG5u8lb6GhVBjpmYw+Qiwbka0X3xn9Fe/neyN3fwawWolQeqCuOjiffi1CvUO8Z+6Ct7eIy9wrRLgqCUJ+U3t/Ss0dQ+YAQQYgAABYBIaoDXXMAAAAsBUIEAAAAQgQAAABCBAAAAECIAAAAQIgAAAAACBEAAAAIEQAAAAAhAgAAACECAAAAIEQAAAAgRAAAAACECAAAAIQIAAAAgBABAACAEAEAAAAQIgAAABAiAAAAAEIEAAAAQgQAAABAiAAAAECIAAAAAAgRAAAA5xCHS1CHZ+yDAq4CAADAIwIAAAAhAgAAACBEAAAAIEQAAAAAhAgAAACECAAAAIAQAQAAgBABAAAAECIAAAAQIgAAAABCBAAAAEIEAAAAQIgAAABAiAAAAAAIEQAAAAgRAAAAACECAAAAIQIAAAAgRAAAACBEAAAAAIQIAAAAhAgAAACAEAEAAIAQAQAAABAiAAAAECIAAAAAQgQAAABCBAAAAECIAAAAQIgAAABAiAAAAAAIEQAAAAgRAAAAACECAAAAIQIAAAAgRAAAACBEAAAAAIQIAAAAhAgAAACAEAEAAIAQAQAAABAiAAAAECIAAAAAQgQAAABCBAAAAECIAAAAQIgAAAAACBEAAAAnEuf9xu12Z/GXbJ5Ste4gPqEZriIAAEQxXDs8OjbfwlOuy+XKb+AR8R3Rh3l6RAgAAADQCWlMnqQ5Z4VI8oQycX0AAACYRKakPfUeUTauCQAAAJPJ9hYidMcBAAAwm1RvIQIAAAAsQRaiLUZ34PHUatmsEpcaAAAcR6O22+PxBLP/Ld5ClGt0L7U11Y1uIzDPdNxPAABwFlpst0eDBgRA1B5BVjMplM5Q5FxsXDwTYmJFZZSTrJTcY/qxT+/brsAtBQAA57HcXfgDN/CXC4LAKInCIf/tqWU11WeM7nqay+XK8vaImPTBGGagm45OpJYnXxfNU8lP8t8QIQAAcC59XL25Dff8W7Tp9ebdI/aGGRQh0pgxsgj5eEQAAACAFfy/AAMARmI+mTf4ZycAAAAASUVORK5CYII=" />
        </div>
        <p>You need to open this link in Firefox to trust this certificate: {{.host}}{{.url}}</p>
    </div>
  </body>

</html>`

var noFirefoxTemplate = template.Must(template.New("home").Parse(noFirefoxTemplateHTML))
