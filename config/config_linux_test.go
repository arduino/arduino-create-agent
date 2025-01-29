package config

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/go-ini/ini"
	"github.com/stretchr/testify/assert"
)

// TestGetConfigPathFromXDG_CONFIG_HOME tests the case when the config.ini is read from XDG_CONFIG_HOME/ArduinoCreateAgent/config.ini
func TestGetConfigPathFromXDG_CONFIG_HOME(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-linux OS")
	}
	// read config from $XDG_CONFIG_HOME/ArduinoCreateAgent/config.ini
	os.Setenv("XDG_CONFIG_HOME", "./testdata/fromxdghome")
	defer os.Unsetenv("XDG_CONFIG_HOME")
	configPath := GetConfigPath()

	assert.Equal(t, "testdata/fromxdghome/ArduinoCreateAgent/config.ini", configPath.String())
	checkIniName(t, configPath, "this-is-a-config-file-from-xdghome-dir")
}

// TestGetConfigPathFromHOME tests the case when the config.ini is read from $HOME/.config/ArduinoCreateAgent/config.ini
func TestGetConfigPathFromHOME(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-linux OS")
	}
	os.Setenv("HOME", "./testdata/fromhome")
	defer os.Unsetenv("HOME")
	configPath := GetConfigPath()

	assert.Equal(t, "testdata/fromhome/.config/ArduinoCreateAgent/config.ini", configPath.String())
	checkIniName(t, configPath, "this-is-a-config-file-from-home-di")
}

// TestGetConfigPathFromARDUINO_CREATE_AGENT_CONFIG tests the case when the config.ini is read from ARDUINO_CREATE_AGENT_CONFIG env variable
func TestGetConfigPathFromARDUINO_CREATE_AGENT_CONFIG(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-linux OS")
	}
	//  $HOME must be always set, otherwise panic
	os.Setenv("HOME", "./testdata/dummyhome")

	os.Setenv("ARDUINO_CREATE_AGENT_CONFIG", "./testdata/from-arduino-create-agent-config-env/config.ini")
	defer os.Unsetenv("ARDUINO_CREATE_AGENT_CONFIG")

	configPath := GetConfigPath()
	assert.Equal(t, "./testdata/from-arduino-create-agent-config-env/config.ini", configPath.String())
	checkIniName(t, configPath, "this-is-a-config-file-from-home-dir-from-ARDUINO_CREATE_AGENT_CONFIG-env")
}

// TestIfHomeDoesNotContainConfigTheDefaultConfigAreCopied tests the case when the default config.ini is copied into $HOME/.config/ArduinoCreateAgent/config.ini
// from the default config.ini
// If the ARDUINO_CREATE_AGENT_CONFIG is NOT set and the config.ini does not exist in HOME directory
// then it copies the default config (the config.ini) into the HOME directory
func TestIfHomeDoesNotContainConfigTheDefaultConfigAreCopied(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-linux OS")
	}
	//  $HOME must be always set, otherwise panic
	os.Setenv("HOME", "./testdata/home-without-config")

	os.Unsetenv("ARDUINO_CREATE_AGENT_CONFIG")
	// we want to test the case when the config does not exist in the home directory
	err := os.Remove("./testdata/home-without-config/.config/ArduinoCreateAgent/config.ini")
	if err != nil {
		t.Fatal(err)
	}

	configPath := GetConfigPath()

	assert.Equal(t, "testdata/home-without-config/.config/ArduinoCreateAgent/config.ini", configPath.String())
	checkIniName(t, configPath, "") // the name of the default config is missing (an empty string)

	givenContent, err := paths.New(configPath.String()).ReadFile()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(configContent), string(givenContent))
}

// TestGetConfigPathPanicIfPathDoesNotExist tests that it panics if the ARDUINO_CREATE_AGENT_CONFIG env variable point to an non-existing path
func TestGetConfigPathPanicIfPathDoesNotExist(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-linux OS")
	}
	os.Setenv("HOME", "./testdata/dummyhome")
	defer os.Unsetenv("HOME")

	os.Setenv("ARDUINO_CREATE_AGENT_CONFIG", "./testdata/a-not-existing-path/config.ini")

	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, fmt.Sprintf("config from env var %s does not exists", "./testdata/a-not-existing-path/config.ini"), r)
		} else {
			t.Fatal("Expected panic but did not occur")
		}
	}()

	configPath := GetConfigPath()

	assert.Equal(t, "testdata/fromxdghome/ArduinoCreateAgent/config.ini", configPath.String())
	checkIniName(t, configPath, "this-is-a-config-file-from-xdghome-dir")
}

func checkIniName(t *testing.T, confipath *paths.Path, expected string) {
	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true, AllowPythonMultilineValues: true}, confipath.String())
	if err != nil {
		t.Fatal(err)
	}
	defaultSection, err := cfg.GetSection("")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected, defaultSection.Key("name").String())
}
