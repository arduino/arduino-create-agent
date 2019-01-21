package pkgs

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-create-agent/gen/indexes"
	"github.com/sirupsen/logrus"
	"go.bug.st/downloader"
)

type Indexes struct {
	Log    *logrus.Logger
	Folder string
}

func (c *Indexes) Add(ctx context.Context, payload *indexes.IndexPayload) error {
	// Parse url
	indexURL, err := url.Parse(payload.URL)
	if err != nil {
		return indexes.MakeInvalidURL(err)
	}

	// Download tmp file
	filename := url.PathEscape(payload.URL)
	path := filepath.Join(c.Folder, filename+".tmp")
	d, err := downloader.Download(path, indexURL.String())
	if err != nil {
		return err
	}
	err = d.Run()
	if err != nil {
		return err
	}

	// Move tmp file
	err = os.Rename(path, filepath.Join(c.Folder, filename))
	if err != nil {
		return err
	}

	return nil
}

func (c *Indexes) Get(ctx context.Context, uri string) (index Index, err error) {
	filename := url.PathEscape(uri)
	path := filepath.Join(c.Folder, filename)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return index, err
	}

	err = json.Unmarshal(data, &index)
	if err != nil {
		return index, err
	}

	return index, nil
}

func (c *Indexes) List(context.Context) ([]string, error) {
	// Read files
	files, err := ioutil.ReadDir(c.Folder)
	if err != nil {
		return nil, err
	}

	res := make([]string, len(files))
	for i, file := range files {
		path, err := url.PathUnescape(file.Name())
		if err != nil {
			c.Log.Warn(err)
		}
		res[i] = path
	}

	return res, nil
}

func (c *Indexes) Remove(ctx context.Context, payload *indexes.IndexPayload) error {
	filename := url.PathEscape(payload.URL)
	return os.RemoveAll(filepath.Join(c.Folder, filename))
}
