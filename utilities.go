// utilities.go

package main

import (
	"github.com/kardianos/osext"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

func pipe_commands(commands ...*exec.Cmd) ([]byte, error) {
	for i, command := range commands[:len(commands)-1] {
		out, err := command.StdoutPipe()
		if err != nil {
			return nil, err
		}
		command.Start()
		commands[i+1].Stdin = out
	}
	final, err := commands[len(commands)-1].Output()
	if err != nil {
		return nil, err
	}
	return final, nil
}

func getBoardName(pid string) (string, string, string, error) {
	execPath, _ := osext.Executable()
	findcmd := exec.Command("grep", "-r", pid, filepath.Dir(execPath)+"/arduino/hardware/")
	findOut, _ := findcmd.Output()

	log.Println(findcmd)
	log.Println(string(findOut))

	boardName := strings.Split(string(findOut), "\n")[0]
	// in this moment arch is the complete path of board.txt
	arch := strings.Split(boardName, ":")[0]
	boardName = strings.Split(boardName, ":")[1]
	boardName = strings.Split(boardName, ".")[0]
	archBoardName := boardName

	// get board.name from board.txt
	boardcmd := exec.Command("grep", "-r", boardName+".name", filepath.Dir(execPath)+"/arduino/hardware/")
	boardOut, _ := boardcmd.Output()
	boardName = string(boardOut)
	boardName = strings.Split(boardName, "=")[1]

	arch = filepath.Dir(arch)
	arch_arr := strings.Split(arch, "/")
	arch = arch_arr[len(arch_arr)-1]

	log.Println(arch, archBoardName, boardName)

	return arch, archBoardName, boardName, nil
}
