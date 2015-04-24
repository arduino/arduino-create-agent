// download.go
package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

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
	fmt.Println("The filePrefix is", filePrefix)
	tmpfile, err := ioutil.TempFile(tmpdir, filePrefix)
	if err != nil {
		fmt.Println("Error creating tempfile. ", err)
		return "", errors.New(err.Error() + " Could not create temp file to store downloaded file. Do you have permissions? Are you out of storage space?")
	}

	fileName := tmpfile.Name()
	fmt.Println("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return fileName, err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return fileName, err
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return fileName, err
	}

	fmt.Println(n, "bytes downloaded.")

	return fileName, nil
}
