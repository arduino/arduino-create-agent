/*
 * This file is part of arduino-create-agent.
 *
 * arduino-create-agent is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

var (
	host      = "localhost"
	validFrom = ""
	validFor  = 365 * 24 * time.Hour * 2 // 2 years
	rsaBits   = 2048
)

func main() {
	// Remove previous certificates and files
	fmt.Print("Remove previous certificates and files... ")
	os.Remove("ca.cert.pem")
	os.Remove("ca.key.pem")
	os.Remove("cert.pem")
	os.Remove("key.pem")
	fmt.Println("OK")

	// Generate ca.key.pem
	fmt.Print("Generate ca.key.pem... ")
	caKey, err := generateKey("P256")
	if err != nil {
		log.Fatal(err)
	}
	keyOut, err := os.OpenFile("ca.key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}
	pem.Encode(keyOut, pemBlockForKey(caKey))
	keyOut.Close()
	fmt.Println("OK")

	// Generate ca.key.pem
	fmt.Print("Generate ca.pem... ")
	caTemplate, err := generateSingleCertificate(true)
	if err != nil {
		log.Fatal(err)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, publicKey(caKey), caKey)

	certOut, err := os.Create("ca.cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	fmt.Println("OK")

	// Generate key.pem
	fmt.Print("Generate key.pem... ")
	key, err := generateKey("P256")
	if err != nil {
		log.Fatal(err)
	}
	keyOut, err = os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}
	pem.Encode(keyOut, pemBlockForKey(key))
	keyOut.Close()
	fmt.Println("OK")

	// Generate cert.pem
	fmt.Print("Generate key.pem... ")
	template, err := generateSingleCertificate(false)
	if err != nil {
		log.Fatal(err)
	}
	derBytes, err = x509.CreateCertificate(rand.Reader, template, caTemplate, publicKey(key), caKey)
	certOut, err = os.Create("cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	fmt.Println("OK")
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
		return nil, fmt.Errorf("Unrecognized elliptic curve: %q", ecdsaCurve)
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

func generateSingleCertificate(isCa bool) (*x509.Certificate, error) {
	var notBefore time.Time
	var err error
	if len(validFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", validFrom)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse creation date: %s\n", err.Error())
		}
	}

	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %s\n", err.Error())
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"Arduino LLC US"},
			Country:            []string{"US"},
			CommonName:         "localhost",
			OrganizationalUnit: []string{"IT"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCa {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		template.Subject.CommonName = "Arduino"
	}

	return &template, nil
}

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
