// Code generated by goa v3.13.1, DO NOT EDIT.
//
// tools HTTP client CLI support package
//
// Command:
// $ goa gen github.com/arduino/arduino-create-agent/design

package client

import (
	"encoding/json"
	"fmt"

	tools "github.com/arduino/arduino-create-agent/gen/tools"
)

// BuildInstallPayload builds the payload for the tools install endpoint from
// CLI flags.
func BuildInstallPayload(toolsInstallBody string) (*tools.ToolPayload, error) {
	var err error
	var body InstallRequestBody
	{
		err = json.Unmarshal([]byte(toolsInstallBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, \nerror: %s, \nexample of valid JSON:\n%s", err, "'{\n      \"checksum\": \"SHA-256:1ae54999c1f97234a5c603eb99ad39313b11746a4ca517269a9285afa05f9100\",\n      \"name\": \"bossac\",\n      \"packager\": \"arduino\",\n      \"signature\": \"382898a97b5a86edd74208f10107d2fecbf7059ffe9cc856e045266fb4db4e98802728a0859cfdcda1c0b9075ec01e42dbea1f430b813530d5a6ae1766dfbba64c3e689b59758062dc2ab2e32b2a3491dc2b9a80b9cda4ae514fbe0ec5af210111b6896976053ab76bac55bcecfcececa68adfa3299e3cde6b7f117b3552a7d80ca419374bb497e3c3f12b640cf5b20875416b45e662fc6150b99b178f8e41d6982b4c0a255925ea39773683f9aa9201dc5768b6fc857c87ff602b6a93452a541b8ec10ca07f166e61a9e9d91f0a6090bd2038ed4427af6251039fb9fe8eb62ec30d7b0f3df38bc9de7204dec478fb86f8eb3f71543710790ee169dce039d3e0\",\n      \"url\": \"http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux64.tar.gz\",\n      \"version\": \"1.7.0-arduino3\"\n   }'")
		}
	}
	v := &tools.ToolPayload{
		Name:      body.Name,
		Version:   body.Version,
		Packager:  body.Packager,
		URL:       body.URL,
		Checksum:  body.Checksum,
		Signature: body.Signature,
	}

	return v, nil
}

// BuildRemovePayload builds the payload for the tools remove endpoint from CLI
// flags.
func BuildRemovePayload(toolsRemoveBody string, toolsRemovePackager string, toolsRemoveName string, toolsRemoveVersion string) (*tools.ToolPayload, error) {
	var err error
	var body RemoveRequestBody
	{
		err = json.Unmarshal([]byte(toolsRemoveBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, \nerror: %s, \nexample of valid JSON:\n%s", err, "'{\n      \"checksum\": \"SHA-256:1ae54999c1f97234a5c603eb99ad39313b11746a4ca517269a9285afa05f9100\",\n      \"signature\": \"382898a97b5a86edd74208f10107d2fecbf7059ffe9cc856e045266fb4db4e98802728a0859cfdcda1c0b9075ec01e42dbea1f430b813530d5a6ae1766dfbba64c3e689b59758062dc2ab2e32b2a3491dc2b9a80b9cda4ae514fbe0ec5af210111b6896976053ab76bac55bcecfcececa68adfa3299e3cde6b7f117b3552a7d80ca419374bb497e3c3f12b640cf5b20875416b45e662fc6150b99b178f8e41d6982b4c0a255925ea39773683f9aa9201dc5768b6fc857c87ff602b6a93452a541b8ec10ca07f166e61a9e9d91f0a6090bd2038ed4427af6251039fb9fe8eb62ec30d7b0f3df38bc9de7204dec478fb86f8eb3f71543710790ee169dce039d3e0\",\n      \"url\": \"http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux64.tar.gz\"\n   }'")
		}
	}
	var packager string
	{
		packager = toolsRemovePackager
	}
	var name string
	{
		name = toolsRemoveName
	}
	var version string
	{
		version = toolsRemoveVersion
	}
	v := &tools.ToolPayload{
		URL:       body.URL,
		Checksum:  body.Checksum,
		Signature: body.Signature,
	}
	v.Packager = packager
	v.Name = name
	v.Version = version

	return v, nil
}
