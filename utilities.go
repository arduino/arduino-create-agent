// utilities.go

package main

import (
	"bytes"
	"crypto/md5"
	"io"
	"os"
	"os/exec"
)

func computeMd5(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
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

// code inspired by https://gist.github.com/tyndyll/89fbb2c2273f83a074dc
func pipe_commands(commands ...*exec.Cmd) ([]byte, error) {
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

// Filter returns a new slice containing all OsSerialPort in the slice that satisfy the predicate f.
func Filter(vs []OsSerialPort, f func(OsSerialPort) bool) []OsSerialPort {
	var vsf []OsSerialPort
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
