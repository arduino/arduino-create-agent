package utilities

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveFileonTemp(t *testing.T) {
	filename := "file"
	tmpDir := t.TempDir()

	path, err := saveFileonTempDir(tmpDir, filename, bytes.NewBufferString("TEST"))
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tmpDir, filename), path)
}

func TestSaveFileonTempDirWithEvilName(t *testing.T) {
	evilFileNames := []string{
		"/",
		"..",
		"../",
		"../evil.txt",
		"../../../../../../../../../../../../../../../../../../../../tmp/evil.txt",
		"some/path/../../../../../../../../../../../../../../../../../../../../tmp/evil.txt",
		"/../../../../../../../../../../../../../../../../../../../../tmp/evil.txt",
		"/some/path/../../../../../../../../../../../../../../../../../../../../tmp/evil.txt",
	}
	if runtime.GOOS == "windows" {
		evilFileNames = []string{
			"..\\",
			"..\\evil.txt",
			"..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\tmp\\evil.txt",
			"some\\path\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\tmp\\evil.txt",
			"\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\tmp\\evil.txt",
			"\\some\\path\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\..\\tmp\\evil.txt",
		}
	}
	for _, evilFileName := range evilFileNames {
		_, err := saveFileonTempDir(t.TempDir(), evilFileName, bytes.NewBufferString("TEST"))
		require.Error(t, err, fmt.Sprintf("with filename: '%s'", evilFileName))
		require.ErrorContains(t, err, "unsafe path join")
	}
}
