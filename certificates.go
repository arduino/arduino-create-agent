// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Generate a self-signed X.509 certificate for a TLS server. Outputs to
// 'cert.pem' and 'key.pem' and will overwrite existing files.

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
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
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
		return nil, fmt.Errorf("Unrecognized elliptic curve: %q", ecdsaCurve)
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
	}

	return &template, nil
}

func generateCertificates() {

	os.Remove("ca.cert.pem")
	os.Remove("ca.key.pem")
	os.Remove("cert.pem")
	os.Remove("key.pem")

	// Create the key for the certification authority
	caKey, err := generateKey("")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	keyOut, err := os.OpenFile("ca.key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	pem.Encode(keyOut, pemBlockForKey(caKey))
	keyOut.Close()
	log.Println("written ca.key.pem")

	// Create the certification authority
	caTemplate, err := generateSingleCertificate(true)

	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, publicKey(caKey), caKey)

	certOut, err := os.Create("ca.cert.pem")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	log.Print("written ca.cert.pem")

	ioutil.WriteFile("ca.cert.cer", derBytes, 0644)
	log.Print("written ca.cert.cer")

	// Create the key for the final certificate
	key, err := generateKey("")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	keyOut, err = os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	pem.Encode(keyOut, pemBlockForKey(key))
	keyOut.Close()
	log.Println("written key.pem")

	// Create the final certificate
	template, err := generateSingleCertificate(false)

	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	derBytes, err = x509.CreateCertificate(rand.Reader, template, caTemplate, publicKey(key), caKey)

	certOut, err = os.Create("cert.pem")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	log.Print("written cert.pem")
}
