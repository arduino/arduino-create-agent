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

	"github.com/arduino/arduino-create-agent/config"
	"github.com/arduino/arduino-create-agent/gen/tools"
	v2 "github.com/arduino/arduino-create-agent/v2"
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

func TestInstallToolDifferentContentType(t *testing.T) {
	r := gin.New()
	goa := v2.Server(config.GetDataDir().String())
	r.Any("/v2/*path", gin.WrapH(goa))
	ts := httptest.NewServer(r)

	URL := "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux64.tar.gz"
	Checksum := "SHA-256:1ae54999c1f97234a5c603eb99ad39313b11746a4ca517269a9285afa05f9100"
	request := tools.ToolPayload{
		Name:     "bossac",
		Version:  "1.7.0-arduino3",
		Packager: "arduino",
		URL:      &URL,
		Checksum: &Checksum,
	}

	payload, err := json.Marshal(request)
	require.NoError(t, err)

	// for some reason the fronted sends requests with "text/plain" content type.
	// Even if the request body contains a json object.
	// With this test we verify is parsed correctly.
	for _, encoding := range []string{"encoding/json", "text/plain"} {
		resp, err := http.Post(ts.URL+"/v2/pkgs/tools/installed", encoding, bytes.NewBuffer(payload))
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "ok")
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}
}
