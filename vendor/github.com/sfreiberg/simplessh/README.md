simplessh
=========

SimpleSSH is a simple wrapper around go ssh and sftp libraries.

License
=======

SimpleSSH is licensed under the MIT license.

Installation
============
`go get github.com/sfreiberg/simplessh`

Documentation
=============
[GoDoc](http://godoc.org/github.com/sfreiberg/simplessh)

Example
=======

```
package main

import (
	"github.com/sfreiberg/simplessh"

	"fmt"
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