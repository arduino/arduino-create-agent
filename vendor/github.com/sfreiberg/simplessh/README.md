# simplessh
[![GoDoc](https://godoc.org/github.com/sfreiberg/simplessh?status.png)](https://godoc.org/github.com/sfreiberg/simplessh)

SimpleSSH is a simple wrapper around go ssh and sftp libraries.

## Features
* Multiple authentication methods (password, private key and ssh-agent)
* Sudo support
* Simple file upload/download support

## Installation
`go get github.com/sfreiberg/simplessh`

## Example

```
package main

import (
	"fmt"
	
	"github.com/sfreiberg/simplessh"
)

func main() {
	/*
		Leave privKeyPath empty to use $HOME/.ssh/id_rsa.
		If username is blank simplessh will attempt to use the current user.
	*/
	client, err := simplessh.ConnectWithKeyFile("localhost:22", "root", "/home/user/.ssh/id_rsa")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	output, err := client.Exec("uptime")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Uptime: %s\n", output)
}

```

## License
SimpleSSH is licensed under the MIT license.