package config

import (
	"os"
	"testing"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigPath(t *testing.T) {
	t.Run("read config.ini from ARDUINO_CREATE_AGENT_CONFIG", func(t *testing.T) {
		os.Setenv("ARDUINO_CREATE_AGENT_CONFIG", "./testdata/fromenv/config.ini")
		defer os.Unsetenv("ARDUINO_CREATE_AGENT_CONFIG")
		configPath := GetConfigPath()
		assert.Equal(t, "./testdata/fromenv/config.ini", configPath.String())
	})

	t.Run("panic if config.ini does not exist", func(t *testing.T) {
		os.Setenv("ARDUINO_CREATE_AGENT_CONFIG", "./testdata/nonexistent_config.ini")
		defer os.Unsetenv("ARDUINO_CREATE_AGENT_CONFIG")

		defer func() {
			if r := recover(); r != nil {
				entry, ok := r.(*logrus.Entry)
				if !ok {
					t.Errorf("Expected panic of type *logrus.Entry but got %T", r)
				} else {
					assert.Equal(t, "config from env var ./testdata/nonexistent_config.ini does not exists", entry.Message)
				}
			} else {
				t.Errorf("Expected panic but did not get one")
			}
		}()

		GetConfigPath()
	})

	t.Run("read config.ini from $HOME/.config/ArduinoCreateAgent folder", func(t *testing.T) {
		os.Setenv("HOME", "./testdata/home")
		defer os.Unsetenv("HOME")
		configPath := GetConfigPath()
		assert.Equal(t, "testdata/home/.config/ArduinoCreateAgent/config.ini", configPath.String())
	})

	t.Run("legacy config are copied to new location", func(t *testing.T) {

		createLegacyConfig := func() string {
			// Create a "legacy" config.ini in the same directory as the binary executable
			src, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			legacyConfigPath, err := paths.New(src).Parent().Join("config.ini").Create()
			if err != nil {
				t.Fatal(err)
			}
			// adding a timestamp to the content to make it unique
			c := "hostname = legacy-config-file-" + time.Now().String()
			n, err := legacyConfigPath.WriteString(c)
			if err != nil || n <= 0 {
				t.Fatalf("Failed to write legacy config file: %v", err)
			}
			return c
		}

		wantContent := createLegacyConfig()

		// Expectation: it copies the "legacy" config.ini into the location pointed by $HOME
		os.Setenv("HOME", "./testdata/fromlegacy")
		defer os.Unsetenv("HOME")

		// remove any existing config.ini in the into the location pointed by $HOME
		err := os.Remove("./testdata/fromlegacy/.config/ArduinoCreateAgent/config.ini")
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}

		configPath := GetConfigPath()
		assert.Equal(t, "testdata/fromlegacy/.config/ArduinoCreateAgent/config.ini", configPath.String())

		given, err := paths.New(configPath.String()).ReadFile()
		assert.Nil(t, err)
		assert.Equal(t, wantContent, string(given))
	})

}
