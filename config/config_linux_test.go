package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/arduino/go-paths-helper"
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
	os.Remove("./testdata/home-without-config/.config/ArduinoCreateAgent/config.ini")

	configPath := GetConfigPath()

	assert.Equal(t, "testdata/home-without-config/.config/ArduinoCreateAgent/config.ini", configPath.String())

	givenContent, err := paths.New(configPath.String()).ReadFile()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(configContent), string(givenContent))

}
