// download.go
package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func saveFileonTempDir(filename string, sketch io.Reader) (path string, err error) {
	// create tmp dir
	tmpdir, err := ioutil.TempDir("", "arduino-create-agent")
	if err != nil {
		return "", errors.New("Could not create temp directory to store downloaded file. Do you have permissions?")
	}

	filename, _ = filepath.Abs(tmpdir + "/" + filename)

	output, err := os.Create(filename)
	if err != nil {
		log.Println("Error while creating", filename, "-", err)
		return filename, err
	}
	defer output.Close()

	n, err := io.Copy(output, sketch)
	if err != nil {
		log.Println("Error while copying", err)
		return filename, err
	}

	log.Println(n, "bytes saved")

	return filename, nil

}

func downloadFromUrl(url string) (filename string, err error) {

	// clean up url
	// remove newlines and space at end
	url = strings.TrimSpace(url)

	// create tmp dir
	tmpdir, err := ioutil.TempDir("", "arduino-create-agent")
	if err != nil {
		return "", errors.New("Could not create temp directory to store downloaded file. Do you have permissions?")
	}
	tokens := strings.Split(url, "/")
	filePrefix := tokens[len(tokens)-1]
	log.Println("The filePrefix is", filePrefix)

	fileName, _ := filepath.Abs(tmpdir + "/" + filePrefix)
	log.Println("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		log.Println("Error while creating", fileName, "-", err)
		return fileName, err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		log.Println("Error while downloading", url, "-", err)
		return fileName, err
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		log.Println("Error while downloading", url, "-", err)
		return fileName, err
	}

	log.Println(n, "bytes downloaded.")

	return fileName, nil
}

func spDownloadTool(name string, url string) {

	if _, err := os.Stat(tempToolsPath + "/" + name); err != nil {

		fileName, err := downloadFromUrl(url + "/" + name + "-" + runtime.GOOS + "-" + runtime.GOARCH + ".zip")
		if err != nil {
			log.Error("Could not download flashing tools!")
			return
		}
		Unzip(fileName, tempToolsPath)
	} else {
		log.Info("Tool already present, skipping download")
	}

	// will be something like ${tempfolder}/avrdude/bin/avrdude
	globalToolsMap["{runtime.tools."+name+".path}"] = tempToolsPath + "/" + name
}
