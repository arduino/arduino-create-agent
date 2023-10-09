package index

import (
	"net/url"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	indexURL := "https://downloads.arduino.cc/packages/package_staging_index.json"
	// Instantiate Index
	tempDir := paths.New(t.TempDir()).Join(".arduino-create")
	Index := Init(indexURL, tempDir)
	require.DirExists(t, tempDir.String())
	fileName := "package_staging_index.json"
	signatureName := fileName + ".sig"
	parsedURL, _ := url.Parse(indexURL)
	require.Equal(t, Index.IndexURL, *parsedURL)
	require.Contains(t, Index.IndexFile.String(), fileName)
	require.Contains(t, Index.IndexSignature.String(), signatureName)
	require.FileExists(t, tempDir.Join(fileName).String())
	require.FileExists(t, tempDir.Join(signatureName).String())
}
