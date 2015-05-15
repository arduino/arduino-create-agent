// utilities.go

package main

import (
	"encoding/json"
	"fmt"
	"github.com/kardianos/osext"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Combo struct {
	vendors map[string]Vendor
}

type Vendor struct {
	architectures map[string]Architecture
}

type Architecture struct {
	boards   map[string]Board
	platform map[string]Platform
}

type Platform struct {
}

type Board struct {
	name       string
	vid        map[int]int
	pid        map[int]int
	upload     Upload
	build      Build
	bootloader Bootloader
}

type Upload struct {
	protocol             string
	disable_flushing     bool
	maximum_data_size    int
	tool                 string
	use_1200bps_touch    bool
	speed                int
	maximum_size         int
	wait_for_upload_port int
}

type Build struct {
	core        string
	f_cpu       int
	board       string
	vid         int
	pid         int
	usb_product string
	mcu         string
	extra_flags string
	variant     string
}

type Bootloader struct {
	extended_fuses int
	high_fuses     int
	file           string
	low_fuses      int
	lock_bits      int
	tool           string
	unlock_bits    int
}

type ConfigProperty struct {
	value string
	path  string
}

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

func getBoardName(pid string) (string, string, error) {
	//execPath, _ := osext.Executable()

	//avr := (m["arduino"].(map[string]interface{})["avr"].(map[string]interface{})["boards"])
	//var uno Board
	//uno = avr["uno"]

	findAllPIDs(globalConfigMap)

	list, _ := searchFor(globalConfigMap, []string{"pid"}, pid)
	fmt.Println(list)

	archBoardNameSlice := strings.Split(list[0].path, ":")[:5]
	archBoardName := archBoardNameSlice[1] + ":" + archBoardNameSlice[2] + ":" + archBoardNameSlice[4]

	boardPath := append(archBoardNameSlice, "name")
	fmt.Println(boardPath)

	boardName := getElementFromMapWithList(globalConfigMap, boardPath).(string)
	fmt.Println(boardName)

	return archBoardName, boardName, nil
}

func matchStringWithSlice(str string, match []string) bool {
	for _, elem := range match {
		if !strings.Contains(str, elem) {
			return false
		}
	}
	return true
}

func recursivelyIterateConfig(m map[string]interface{}, fullpath string, match []string, mapOut *[]ConfigProperty) {

	for k, v := range m {
		switch vv := v.(type) {
		case string:
			if matchStringWithSlice(fullpath+":"+k, match) {
				//fmt.Println(k, "is string", vv, "path", fullpath)
				if mapOut != nil {
					*mapOut = append(*mapOut, ConfigProperty{path: fullpath, value: vv})
					//fmt.Println(getElementFromMapWithList(globalConfigMap, strings.Split(fullpath, ":")))
				}
			}
		case map[string]interface{}:
			//fmt.Println(k, "is a map:", fullpath)
			recursivelyIterateConfig(m[k].(map[string]interface{}), fullpath+":"+k, match, mapOut)
		default:
			//fmt.Println(k, "is of a type I don't know how to handle ", vv)
		}
	}
}

func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

func findAllVIDs(m map[string]interface{}) []ConfigProperty {
	var vidList []ConfigProperty
	recursivelyIterateConfig(m, "", []string{"vid"}, &vidList)
	//fmt.Println(vidList)
	return vidList
}

func findAllPIDs(m map[string]interface{}) []ConfigProperty {
	var pidList []ConfigProperty
	recursivelyIterateConfig(m, "", []string{"pid"}, &pidList)
	//fmt.Println(pidList)
	return pidList
}

func searchFor(m map[string]interface{}, args []string, element string) ([]ConfigProperty, bool) {
	var uList []ConfigProperty
	var results []ConfigProperty
	recursivelyIterateConfig(m, "", args, &uList)
	//fmt.Println(uList)
	for _, elm := range uList {
		if elm.value == element {
			results = append(results, elm)
		}
	}
	return results, len(results) != 0
}

func getElementFromMapWithList(m map[string]interface{}, listStr []string) interface{} {
	var k map[string]interface{}
	k = m
	for _, element := range listStr {
		switch k[element].(type) {
		case string:
			return k[element]
		default:
			if element != "" {
				k = k[element].(map[string]interface{})
			}
		}
	}
	return k
}

func createGlobalConfigMap(m *map[string]interface{}) {
	execPath, _ := osext.Executable()

	file, e := ioutil.ReadFile(filepath.Dir(execPath) + "/arduino/boards.json")

	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	//var config Combo
	json.Unmarshal(file, m)
}
