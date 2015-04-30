// download.go
package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func saveFileonTempDir(filename string, sketch io.Reader) (path string, err error) {
	// create tmp dir
	tmpdir, err := ioutil.TempDir("", "serial-port-json-server")
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
	tmpdir, err := ioutil.TempDir("", "serial-port-json-server")
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
