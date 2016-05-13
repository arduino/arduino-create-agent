package utilities

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

// SaveFileonTempDir creates a temp directory and saves the file data as the
// filename in that directory.
// Returns an error if the agent doesn't have permission to create the folder.
// Returns an error if the filename doesn't form a valid path.
//
// Note that path could be defined and still there could be an error.
func SaveFileonTempDir(filename string, data io.Reader) (path string, err error) {
	// Create Temp Directory
	tmpdir, err := ioutil.TempDir("", "arduino-create-agent")
	if err != nil {
		return "", errors.New("Could not create temp directory to store downloaded file. Do you have permissions?")
	}

	// Determine filename
	filename, err = filepath.Abs(tmpdir + "/" + filename)
	if err != nil {
		return "", err
	}

	// Touch file
	output, err := os.Create(filename)
	if err != nil {
		return filename, err
	}
	defer output.Close()

	// Write file
	_, err = io.Copy(output, data)
	if err != nil {
		return filename, err
	}
	return filename, nil
}

// PipeCommands executes the commands received as input by feeding the output of
// one to the input of the other, exactly like Unix Pipe (|).
// Returns the output of the final command and the eventual error.
//
// code inspired by https://gist.github.com/tyndyll/89fbb2c2273f83a074dc
func PipeCommands(commands ...*exec.Cmd) ([]byte, error) {
	var errorBuffer, outputBuffer bytes.Buffer
	pipeStack := make([]*io.PipeWriter, len(commands)-1)
	i := 0
	for ; i < len(commands)-1; i++ {
		stdinPipe, stdoutPipe := io.Pipe()
		commands[i].Stdout = stdoutPipe
		commands[i].Stderr = &errorBuffer
		commands[i+1].Stdin = stdinPipe
		pipeStack[i] = stdoutPipe
	}
	commands[i].Stdout = &outputBuffer
	commands[i].Stderr = &errorBuffer

	if err := call(commands, pipeStack); err != nil {
		return nil, err
	}

	return outputBuffer.Bytes(), nil
}

func call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}
	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}
		defer func() {
			pipes[0].Close()
			err = call(stack[1:], pipes[1:])
		}()
	}
	return stack[0].Wait()
}

func Unzip(zippath string, destination string) (err error) {
	r, err := zip.OpenReader(zippath)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		fullname := path.Join(destination, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fullname, f.FileInfo().Mode().Perm())
		} else {
			os.MkdirAll(filepath.Dir(fullname), 0755)
			perms := f.FileInfo().Mode().Perm()
			out, err := os.OpenFile(fullname, os.O_CREATE|os.O_RDWR, perms)
			if err != nil {
				return err
			}
			rc, err := f.Open()
			if err != nil {
				return err
			}
			_, err = io.CopyN(out, rc, f.FileInfo().Size())
			if err != nil {
				return err
			}
			rc.Close()
			out.Close()

			mtime := f.FileInfo().ModTime()
			err = os.Chtimes(fullname, mtime, mtime)
			if err != nil {
				return err
			}
		}
	}
	return
}
