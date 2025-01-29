package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigPathFromXDG_CONFIG_HOME(t *testing.T) {
	// read config from $XDG_CONFIG_HOME/ArduinoCreateAgent/config.ini
	os.Setenv("XDG_CONFIG_HOME", "./testdata/fromxdghome")
	defer os.Unsetenv("XDG_CONFIG_HOME")
	configPath := GetConfigPath()
	assert.Equal(t, "testdata/fromxdghome/ArduinoCreateAgent/config.ini", configPath.String())
}

func TestGetConfigPathFromHOME(t *testing.T) {
	// Test case 2: read config from $HOME/.config/ArduinoCreateAgent/config.ini "
	os.Setenv("HOME", "./testdata/fromhome")
	defer os.Unsetenv("HOME")
	configPath := GetConfigPath()
	assert.Equal(t, "testdata/fromhome/.config/ArduinoCreateAgent/config.ini", configPath.String())

}

func TestGetConfigPathFromARDUINO_CREATE_AGENT_CONFIG(t *testing.T) {
	// read config from ARDUINO_CREATE_AGENT_CONFIG/config.ini"
	os.Setenv("HOME", "./fromhome")
	os.Setenv("ARDUINO_CREATE_AGENT_CONFIG", "./testdata/fromenv/config.ini")
	defer os.Unsetenv("ARDUINO_CREATE_AGENT_CONFIG")

	configPath := GetConfigPath()
	assert.Equal(t, "./testdata/fromenv/config.ini", configPath.String())
}

// func TestGetConfigPathFromLegacyConfig(t *testing.T) {
// 	// If no config is found, copy the legacy config to the new location
// 	src, err := os.Executable()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	legacyConfigPath, err := paths.New(src).Parent().Join("config.ini").Create()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// adding a timestamp to the content to make it unique
// 	legacyContent := "hostname = legacy-config-file-" + time.Now().String()
// 	n, err := legacyConfigPath.WriteString(legacyContent)
// 	if err != nil || n <= 0 {
// 		t.Fatalf("Failed to write legacy config file: %v", err)
// 	}

// 	// remove any existing config.ini in the into the location pointed by $HOME
// 	err = os.Remove("./testdata/fromlegacy/.config/ArduinoCreateAgent/config.ini")
// 	if err != nil && !os.IsNotExist(err) {
// 		t.Fatal(err)
// 	}

// 	// Expectation: it copies the "legacy" config.ini into the location pointed by $HOME
// 	os.Setenv("HOME", "./testdata/fromlegacy")
// 	defer os.Unsetenv("HOME")

// 	configPath := GetConfigPath()
// 	assert.Equal(t, "testdata/fromlegacy/.config/ArduinoCreateAgent/config.ini", configPath.String())

// 	given, err := paths.New(configPath.String()).ReadFile()
// 	assert.Nil(t, err)
// 	assert.Equal(t, legacyContent, string(given))
// }

// func TestGetConfigPathCreateDefaultConfig(t *testing.T) {
// 	os.Setenv("HOME", "./testdata/noconfig")
// 	os.Unsetenv("ARDUINO_CREATE_AGENT_CONFIG")

// 	// ensure the config.ini does not exist in HOME directory
// 	os.Remove("./testdata/noconfig/.config/ArduinoCreateAgent/config.ini")
// 	// ensure the config.ini does not exist in target directory
// 	os.Remove("./testdata/fromdefault/.config/ArduinoCreateAgent/config.ini")

// 	configPath := GetConfigPath()

// 	assert.Equal(t, "testdata/fromdefault/.config/ArduinoCreateAgent/config.ini", configPath.String())

// 	givenContent, err := paths.New(configPath.String()).ReadFile()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	assert.Equal(t, string(configContent), string(givenContent))

// }
