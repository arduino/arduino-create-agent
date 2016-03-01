package tools

import (
	"os"
	"os/user"
)

func dir() string {
	usr, _ := user.Current()
	return usr.HomeDir + "/.arduino-create"
}

// CreateDir creates the directory where the tools will be stored
func CreateDir() {
	directory := dir()
	os.Mkdir(directory, 0777)
	hideFile(directory)
}

// Download will parse the index at the indexURL for the tool to download
func Download(name, indexURL string) {
	if _, err := os.Stat(dir() + "/" + name); err != nil {
	}
}
