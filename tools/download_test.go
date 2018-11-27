package tools

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
)

func Test_findBaseDir(t *testing.T) {
	cases := []struct {
		dirList []string
		want    string
	}{
		{[]string{"bin/bossac"}, "bin/"},
		{[]string{"bin/", "bin/bossac"}, "bin/"},
		{[]string{"bin/", "bin/bossac", "example"}, ""},
		{[]string{"avrdude/bin/",
			"avrdude/bin/avrdude.exe",
			"avrdude/bin/remove_giveio.bat",
			"avrdude/bin/status_giveio.bat",
			"avrdude/bin/giveio.sys",
			"avrdude/bin/loaddrv.exe",
			"avrdude/bin/libusb0.dll",
			"avrdude/bin/install_giveio.bat",
			"avrdude/etc/avrdude.conf"}, "avrdude/"},
		{[]string{"pax_global_header", "bin/", "bin/bossac"}, "bin/"},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintln(tt.dirList), func(t *testing.T) {
			if got := findBaseDir(tt.dirList); got != tt.want {
				t.Errorf("findBaseDir() = got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTools_DownloadAndUnpackBehaviour(t *testing.T) {
	urls := []string{
		"http://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-armhf-pc-linux-gnu.tar.bz2",
		"http://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-aarch64-pc-linux-gnu.tar.bz2",
		"http://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-i386-apple-darwin11.tar.bz2",
		"http://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-x86_64-pc-linux-gnu.tar.bz2",
		"http://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-i686-pc-linux-gnu.tar.bz2",
		"http://downloads.arduino.cc/tools/avrdude-6.3.0-arduino14-i686-w64-mingw32.zip",
	}
	expectedDirList := []string{"bin", "etc"}

	for _, url := range urls {
		t.Log("Downloading tool from " + url)
		resp, err := http.Get(url)
		if err != nil {
			t.Errorf("%v", err)
		}
		defer resp.Body.Close()

		// Read the body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%v", err)
		}

		location := path.Join("/tmp", dir(), "arduino", "avrdude", "6.3.0-arduino14")
		os.MkdirAll(location, os.ModePerm)
		err = os.RemoveAll(location)

		if err != nil {
			t.Errorf("%v", err)
		}

		srcType, err := mimeType(body)
		if err != nil {
			t.Errorf("%v", err)
		}

		switch srcType {
		case "application/zip":
			location, err = extractZip(func(msg string) { t.Log(msg) }, body, location)
		case "application/x-bz2":
		case "application/octet-stream":
			location, err = extractBz2(func(msg string) { t.Log(msg) }, body, location)
		case "application/x-gzip":
			location, err = extractTarGz(func(msg string) { t.Log(msg) }, body, location)
		default:
			t.Errorf("no suitable type found")
		}
		files, err := ioutil.ReadDir(location)
		if err != nil {
			t.Errorf("%v", err)
		}
		dirList := []string{}
		for _, f := range files {
			dirList = append(dirList, f.Name())
		}

		assert.ElementsMatchf(t, dirList, expectedDirList, "error message %s", "formatted")

	}

}
