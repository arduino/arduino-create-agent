// download.go
package main

import (
	"encoding/json"
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

func spCheckToolVersion(name string) {
	var outlist []string
	dirlist, err := ioutil.ReadDir(tempToolsPath + "/")
	if err == nil {
		for _, element := range dirlist {
			if element.IsDir() && strings.Contains(element.Name(), name) {
				outlist = append(outlist, element.Name())
			}
		}
	}
	mapD := map[string][]string{"ToolVersions": outlist}
	mapB, _ := json.Marshal(mapD)
	h.broadcastSys <- mapB
}

func spDownloadTool(name string, url string) {

	fileName, err := downloadFromUrl(url + "/" + name + "-" + runtime.GOOS + "-" + runtime.GOARCH + ".zip")
	if err != nil {
		log.Error("Could not download flashing tools!")
		mapD := map[string]string{"DownloadStatus": "Error", "Msg": err.Error()}
		mapB, _ := json.Marshal(mapD)
		h.broadcastSys <- mapB
		return
	}
	err = UnzipWrapper(fileName, tempToolsPath)
	if err != nil {
		log.Error("Could not unzip flashing tools!")
		mapD := map[string]string{"DownloadStatus": "Error", "Msg": err.Error()}
		mapB, _ := json.Marshal(mapD)
		h.broadcastSys <- mapB
		return
	}

	folders, _ := ioutil.ReadDir(tempToolsPath)
	for _, f := range folders {
		globalToolsMap["{runtime.tools."+f.Name()+".path}"] = filepath.ToSlash(tempToolsPath + "/" + f.Name())
	}

	log.Info("Map Updated")
	mapD := map[string]string{"DownloadStatus": "Success", "Msg": "Map Updated"}
	mapB, _ := json.Marshal(mapD)
	h.broadcastSys <- mapB
	return

}
