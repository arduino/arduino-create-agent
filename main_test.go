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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-create-agent/config"
	"github.com/arduino/arduino-create-agent/gen/tools"
	"github.com/arduino/arduino-create-agent/index"
	"github.com/arduino/arduino-create-agent/upload"
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

func TestInstallToolV2(t *testing.T) {

	indexURL := "https://downloads.arduino.cc/packages/package_staging_index.json"
	// Instantiate Index
	Index := index.Init(indexURL, config.GetDataDir())

	r := gin.New()
	goa := v2.Server(config.GetDataDir().String(), Index)
	r.Any("/v2/*path", gin.WrapH(goa))
	ts := httptest.NewServer(r)

	type test struct {
		request      tools.ToolPayload
		responseCode int
		responseBody string
	}

	bossacURL := "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux64.tar.gz"
	bossacChecksum := "SHA-256:1ae54999c1f97234a5c603eb99ad39313b11746a4ca517269a9285afa05f9100"
	bossacSignature := "382898a97b5a86edd74208f10107d2fecbf7059ffe9cc856e045266fb4db4e98802728a0859cfdcda1c0b9075ec01e42dbea1f430b813530d5a6ae1766dfbba64c3e689b59758062dc2ab2e32b2a3491dc2b9a80b9cda4ae514fbe0ec5af210111b6896976053ab76bac55bcecfcececa68adfa3299e3cde6b7f117b3552a7d80ca419374bb497e3c3f12b640cf5b20875416b45e662fc6150b99b178f8e41d6982b4c0a255925ea39773683f9aa9201dc5768b6fc857c87ff602b6a93452a541b8ec10ca07f166e61a9e9d91f0a6090bd2038ed4427af6251039fb9fe8eb62ec30d7b0f3df38bc9de7204dec478fb86f8eb3f71543710790ee169dce039d3e0"
	bossacInstallURLOK := tools.ToolPayload{
		Name:      "bossac",
		Version:   "1.7.0-arduino3",
		Packager:  "arduino",
		URL:       &bossacURL,
		Checksum:  &bossacChecksum,
		Signature: &bossacSignature,
	}

	wrongSignature := "wr0ngs1gn4tur3"
	bossacInstallWrongSig := tools.ToolPayload{
		Name:      "bossac",
		Version:   "1.7.0-arduino3",
		Packager:  "arduino",
		URL:       &bossacURL,
		Checksum:  &bossacChecksum,
		Signature: &wrongSignature,
	}

	wrongChecksum := "wr0ngch3cksum"
	bossacInstallWrongCheck := tools.ToolPayload{
		Name:      "bossac",
		Version:   "1.7.0-arduino3",
		Packager:  "arduino",
		URL:       &bossacURL,
		Checksum:  &wrongChecksum,
		Signature: &bossacSignature,
	}

	bossacInstallNoURL := tools.ToolPayload{
		Name:     "bossac",
		Version:  "1.7.0-arduino3",
		Packager: "arduino",
	}

	tests := []test{
		{bossacInstallURLOK, http.StatusOK, "ok"},
		{bossacInstallWrongSig, http.StatusInternalServerError, "verification error"},
		{bossacInstallWrongCheck, http.StatusInternalServerError, "checksum of downloaded file doesn't match"},
		{bossacInstallNoURL, http.StatusOK, "ok"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Installing %s", test.request.Name), func(t *testing.T) {
			payload, err := json.Marshal(test.request)
			require.NoError(t, err)

			// for some reason the fronted sends requests with "text/plain" content type.
			// Even if the request body contains a json object.
			// With this test we verify is parsed correctly.
			for _, encoding := range []string{"encoding/json", "text/plain"} {
				resp, err := http.Post(ts.URL+"/v2/pkgs/tools/installed", encoding, bytes.NewBuffer(payload))
				require.NoError(t, err)
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), test.responseBody)
				require.Equal(t, test.responseCode, resp.StatusCode)
			}
		})
	}
}

func TestInstalledHead(t *testing.T) {
	indexURL := "https://downloads.arduino.cc/packages/package_staging_index.json"
	// Instantiate Index
	Index := index.Init(indexURL, config.GetDataDir())

	r := gin.New()
	goa := v2.Server(config.GetDataDir().String(), Index)
	r.Any("/v2/*path", gin.WrapH(goa))
	ts := httptest.NewServer(r)

	resp, err := http.Head(ts.URL + "/v2/pkgs/tools/installed")
	require.NoError(t, err)
	require.NotEqual(t, resp.StatusCode, http.StatusMethodNotAllowed)
	require.Equal(t, resp.StatusCode, http.StatusOK)
}
