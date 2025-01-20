package config

import (
	"fmt"
	"os"
	"testing"

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

	t.Run("read config.ini from $HOME", func(t *testing.T) {
		os.Setenv("HOME", "./testdata/home")
		defer os.Unsetenv("HOME")
		configPath := GetConfigPath()
		assert.Equal(t, "testdata/home/.config/ArduinoCreateAgent/config.ini", configPath.String())
	})

	t.Run("fallback old : read config.ini where the binary is launched", func(t *testing.T) {
		src, _ := os.Executable()
		paths.New(src).Parent().Join("config.ini").Create() // create a config.ini in the same directory as the binary
		// The fallback path is the directory where the binary is launched
		fmt.Println(src)
		os.Setenv("HOME", "./testdata/noconfig") // force to not have a config in the home directory
		defer os.Unsetenv("HOME")

		// expect it creates a config.ini in the same directory as the binary
		configPath := GetConfigPath()
		assert.Equal(t, "testdata/home/.config/ArduinoCreateAgent/config.ini", configPath.String())
	})

}
