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

package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-create-agent/upload"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestValidSignatureKey(t *testing.T) {
	testfile := filepath.Join("tests", "testdata", "test.ini")
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

func TestUploadHandlerAgainstEvilFileNames(t *testing.T) {
	r := gin.New()
	r.POST("/", uploadHandler)
	ts := httptest.NewServer(r)

	uploadEvilFileName := Upload{
		Port:       "/dev/ttyACM0",
		Board:      "arduino:avr:uno",
		Extra:      upload.Extra{Network: true},
		Hex:        []byte("test"),
		Filename:   "../evil.txt",
		ExtraFiles: []additionalFile{{Hex: []byte("test"), Filename: "../evil.txt"}},
	}
	uploadEvilExtraFile := Upload{
		Port:       "/dev/ttyACM0",
		Board:      "arduino:avr:uno",
		Extra:      upload.Extra{Network: true},
		Hex:        []byte("test"),
		Filename:   "file.txt",
		ExtraFiles: []additionalFile{{Hex: []byte("test"), Filename: "../evil.txt"}},
	}

	for _, request := range []Upload{uploadEvilFileName, uploadEvilExtraFile} {
		payload, err := json.Marshal(request)
		require.NoError(t, err)

		resp, err := http.Post(ts.URL, "encoding/json", bytes.NewBuffer(payload))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "unsafe path join")
	}
}
