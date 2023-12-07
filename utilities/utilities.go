// Copyright 2022 Arduino SA
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package utilities

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-create-agent/globals"
)

// SaveFileonTempDir creates a temp directory and saves the file data as the
// filename in that directory.
// Returns an error if the agent doesn't have permission to create the folder.
// Returns an error if the filename doesn't form a valid path.
//
// Note that path could be defined and still there could be an error.
func SaveFileonTempDir(filename string, data io.Reader) (string, error) {
	tmpdir, err := os.MkdirTemp("", "arduino-create-agent")
	if err != nil {
		return "", errors.New("Could not create temp directory to store downloaded file. Do you have permissions?")
	}
	return saveFileonTempDir(tmpdir, filename, data)
}

func saveFileonTempDir(tmpDir, filename string, data io.Reader) (string, error) {
	path, err := SafeJoin(tmpDir, filename)
	if err != nil {
		return "", err
	}
	// Determine filename
	filename, err = filepath.Abs(path)
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

// SafeJoin performs a filepath.Join of 'parent' and 'subdir' but returns an error
// if the resulting path points outside of 'parent'.
func SafeJoin(parent, subdir string) (string, error) {
	res := filepath.Join(parent, subdir)
	if !strings.HasSuffix(parent, string(os.PathSeparator)) {
		parent += string(os.PathSeparator)
	}
	if !strings.HasPrefix(res, parent) {
		return res, fmt.Errorf("unsafe path join: '%s' with '%s'", parent, subdir)
	}
	return res, nil
}

// VerifyInput will verify an input against a signature
// A valid signature is indicated by returning a nil error.
func VerifyInput(input string, signature string) error {
	sign, _ := hex.DecodeString(signature)
	block, _ := pem.Decode([]byte(globals.SignatureKey))
	if block == nil {
		return errors.New("invalid key")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	rsaKey := key.(*rsa.PublicKey)
	h := sha256.New()
	h.Write([]byte(input))
	d := h.Sum(nil)
	return rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, d, sign)
}
