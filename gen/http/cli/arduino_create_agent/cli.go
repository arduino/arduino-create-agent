// Code generated by goa v3.13.2, DO NOT EDIT.
//
// arduino-create-agent HTTP client CLI support package
//
// Command:
// $ goa gen github.com/arduino/arduino-create-agent/design

package cli

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	toolsc "github.com/arduino/arduino-create-agent/gen/http/tools/client"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

// UsageCommands returns the set of commands and sub-commands using the format
//
//	command (subcommand1|subcommand2|...)
func UsageCommands() string {
	return `tools (available|installed|install|remove)
`
}

// UsageExamples produces an example of a valid invocation of the CLI tool.
func UsageExamples() string {
	return os.Args[0] + ` tools available` + "\n" +
		""
}

// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
) (goa.Endpoint, any, error) {
	var (
		toolsFlags = flag.NewFlagSet("tools", flag.ContinueOnError)

		toolsAvailableFlags = flag.NewFlagSet("available", flag.ExitOnError)

		toolsInstalledFlags = flag.NewFlagSet("installed", flag.ExitOnError)

		toolsInstallFlags    = flag.NewFlagSet("install", flag.ExitOnError)
		toolsInstallBodyFlag = toolsInstallFlags.String("body", "REQUIRED", "")

		toolsRemoveFlags        = flag.NewFlagSet("remove", flag.ExitOnError)
		toolsRemoveBodyFlag     = toolsRemoveFlags.String("body", "REQUIRED", "")
		toolsRemovePackagerFlag = toolsRemoveFlags.String("packager", "REQUIRED", "The packager of the tool")
		toolsRemoveNameFlag     = toolsRemoveFlags.String("name", "REQUIRED", "The name of the tool")
		toolsRemoveVersionFlag  = toolsRemoveFlags.String("version", "REQUIRED", "The version of the tool")
	)
	toolsFlags.Usage = toolsUsage
	toolsAvailableFlags.Usage = toolsAvailableUsage
	toolsInstalledFlags.Usage = toolsInstalledUsage
	toolsInstallFlags.Usage = toolsInstallUsage
	toolsRemoveFlags.Usage = toolsRemoveUsage

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, nil, err
	}

	if flag.NArg() < 2 { // two non flag args are required: SERVICE and ENDPOINT (aka COMMAND)
		return nil, nil, fmt.Errorf("not enough arguments")
	}

	var (
		svcn string
		svcf *flag.FlagSet
	)
	{
		svcn = flag.Arg(0)
		switch svcn {
		case "tools":
			svcf = toolsFlags
		default:
			return nil, nil, fmt.Errorf("unknown service %q", svcn)
		}
	}
	if err := svcf.Parse(flag.Args()[1:]); err != nil {
		return nil, nil, err
	}

	var (
		epn string
		epf *flag.FlagSet
	)
	{
		epn = svcf.Arg(0)
		switch svcn {
		case "tools":
			switch epn {
			case "available":
				epf = toolsAvailableFlags

			case "installed":
				epf = toolsInstalledFlags

			case "install":
				epf = toolsInstallFlags

			case "remove":
				epf = toolsRemoveFlags

			}

		}
	}
	if epf == nil {
		return nil, nil, fmt.Errorf("unknown %q endpoint %q", svcn, epn)
	}

	// Parse endpoint flags if any
	if svcf.NArg() > 1 {
		if err := epf.Parse(svcf.Args()[1:]); err != nil {
			return nil, nil, err
		}
	}

	var (
		data     any
		endpoint goa.Endpoint
		err      error
	)
	{
		switch svcn {
		case "tools":
			c := toolsc.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "available":
				endpoint = c.Available()
				data = nil
			case "installed":
				endpoint = c.Installed()
				data = nil
			case "install":
				endpoint = c.Install()
				data, err = toolsc.BuildInstallPayload(*toolsInstallBodyFlag)
			case "remove":
				endpoint = c.Remove()
				data, err = toolsc.BuildRemovePayload(*toolsRemoveBodyFlag, *toolsRemovePackagerFlag, *toolsRemoveNameFlag, *toolsRemoveVersionFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}

// toolsUsage displays the usage of the tools command and its subcommands.
func toolsUsage() {
	fmt.Fprintf(os.Stderr, `The tools service manages the available and installed tools
Usage:
    %[1]s [globalflags] tools COMMAND [flags]

COMMAND:
    available: Available implements available.
    installed: Installed implements installed.
    install: Install implements install.
    remove: Remove implements remove.

Additional help:
    %[1]s tools COMMAND --help
`, os.Args[0])
}
func toolsAvailableUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] tools available

Available implements available.

Example:
    %[1]s tools available
`, os.Args[0])
}

func toolsInstalledUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] tools installed

Installed implements installed.

Example:
    %[1]s tools installed
`, os.Args[0])
}

func toolsInstallUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] tools install -body JSON

Install implements install.
    -body JSON: 

Example:
    %[1]s tools install --body '{
      "checksum": "SHA-256:1ae54999c1f97234a5c603eb99ad39313b11746a4ca517269a9285afa05f9100",
      "name": "bossac",
      "packager": "arduino",
      "signature": "382898a97b5a86edd74208f10107d2fecbf7059ffe9cc856e045266fb4db4e98802728a0859cfdcda1c0b9075ec01e42dbea1f430b813530d5a6ae1766dfbba64c3e689b59758062dc2ab2e32b2a3491dc2b9a80b9cda4ae514fbe0ec5af210111b6896976053ab76bac55bcecfcececa68adfa3299e3cde6b7f117b3552a7d80ca419374bb497e3c3f12b640cf5b20875416b45e662fc6150b99b178f8e41d6982b4c0a255925ea39773683f9aa9201dc5768b6fc857c87ff602b6a93452a541b8ec10ca07f166e61a9e9d91f0a6090bd2038ed4427af6251039fb9fe8eb62ec30d7b0f3df38bc9de7204dec478fb86f8eb3f71543710790ee169dce039d3e0",
      "url": "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux64.tar.gz",
      "version": "1.7.0-arduino3"
   }'
`, os.Args[0])
}

func toolsRemoveUsage() {
	fmt.Fprintf(os.Stderr, `%[1]s [flags] tools remove -body JSON -packager STRING -name STRING -version STRING

Remove implements remove.
    -body JSON: 
    -packager STRING: The packager of the tool
    -name STRING: The name of the tool
    -version STRING: The version of the tool

Example:
    %[1]s tools remove --body '{
      "checksum": "SHA-256:1ae54999c1f97234a5c603eb99ad39313b11746a4ca517269a9285afa05f9100",
      "signature": "382898a97b5a86edd74208f10107d2fecbf7059ffe9cc856e045266fb4db4e98802728a0859cfdcda1c0b9075ec01e42dbea1f430b813530d5a6ae1766dfbba64c3e689b59758062dc2ab2e32b2a3491dc2b9a80b9cda4ae514fbe0ec5af210111b6896976053ab76bac55bcecfcececa68adfa3299e3cde6b7f117b3552a7d80ca419374bb497e3c3f12b640cf5b20875416b45e662fc6150b99b178f8e41d6982b4c0a255925ea39773683f9aa9201dc5768b6fc857c87ff602b6a93452a541b8ec10ca07f166e61a9e9d91f0a6090bd2038ed4427af6251039fb9fe8eb62ec30d7b0f3df38bc9de7204dec478fb86f8eb3f71543710790ee169dce039d3e0",
      "url": "http://downloads.arduino.cc/tools/bossac-1.7.0-arduino3-linux64.tar.gz"
   }' --packager "arduino" --name "bossac" --version "1.7.0-arduino3"
`, os.Args[0])
}
