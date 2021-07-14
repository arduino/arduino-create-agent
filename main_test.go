package main

import (
	"crypto/x509"
	"encoding/pem"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidSignatureKey(t *testing.T) {
	testfile := filepath.Join("test", "testdata", "test.ini")
	args, err := parseIni(testfile)
	require.NoError(t, err)
	require.NotNil(t, args)
	err = iniConf.Parse(args)
	require.NoError(t, err)
	print(*signatureKey)
	block, _ := pem.Decode([]byte(*signatureKey))
	require.NotNil(t, block)
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	require.NoError(t, err)
	require.NotNil(t, key)
}
