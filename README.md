[![License: GPL v2](https://img.shields.io/badge/License-GPL%20v2-blue.svg)](https://www.gnu.org/licenses/old-licenses/gpl-2.0.en.html)

arduino-create-agent
====================


## GOA 2 refactoring
The agent is currently transitioning to the v2 of the GOA framework for API management, please refer to the following 
[documentation](https://github.com/goadesign/goa/tree/v2) in order to install tools and libraries


i.e. to regenerate code from design use:
```bash
goa gen github.com/arduino/arduino-create-agent/design
```

## Installation
Get the [latest version](https://github.com/arduino/arduino-create-agent/releases) of the Agent for all supported platforms

arduino-create-agent is a fork of @[johnlauer](https://github.com/johnlauer)'s [serial-port-json-server](https://github.com/johnlauer/serial-port-json-server) (which we really want to thank for his kindness and great work)

The history has been rewritten to keep the repo small (thus removing all binaries committed in the past)

## Development

Please remember that for compile the project, you need go version >= 1.10.x (older versions are not supported for compile)

To clone the repository, run the following command:
```
go get github.com/arduino/arduino-create-agent
```

This will clone the repository into your [Go workspace](https://golang.org/doc/code.html#Workspaces) or create a new workspace, if one doesn't exist. You can set `$GOPATH` to define where your Go workspace is located.

Now you can go to the project directory and compile it:
```
cd $GOPATH/src/github.com/arduino/arduino-create-agent
go build
```

This will create the `arduino-create-agent` binary.

Other prerequisites are:
* libappindicator (Linux only on Ubuntu `sudo apt-get install libappindicator1 libappindicator3-0.1-cil libappindicator3-0.1-cil-dev libappindicator3-1 libappindicator3-dev libgtk-3-0 libgtk-3-dev`)
* [go-selfupdate] (https://github.com/sanbornm/go-selfupdate) if you want to test automatic updates

### Windows
Since we are using the https://github.com/lxn/walk library, we need to ship a manifest.xml file, otherwise the error would be:

```
panic: Unable to create main window: TTM_ADDTOOL failed
```

To do it make sure to install the required tool:

```
$ go get github.com/akavel/rsrc
```

and build it with

```
$ rsrc -arch=386 -manifest=manifest.xml
$ go build
```

Keep in mind that the presence of rsrc.syso file will break other builds, for example

```
$ GOOS=linux go build
# github.com/arduino/arduino-create-agent
/usr/lib/go/pkg/tool/linux_amd64/link: running gcc failed: exit status 1
/usr/sbin/ld: i386 architecture of input file `/tmp/go-link-084341451/000000.o' is incompatible with i386:x86-64 output
collect2: error: ld returned 1 exit status
```

## Submitting an issue

Please attach the output of the commands running at the debug console if useful.

## Submitting a pull request

We are glad you want to contribute with code: that's the best way to help this software.

Your contribution is adding or modifying existing behaviour, please always refer to an existing issue or open a new one before contributing. We are trying to use [Test Driven Development](https://en.wikipedia.org/wiki/Test-driven_development) in the near future: please add one or more tests that prove that your contribution is good and is working as expected, it will help us a lot.

Be sure to use `go vet` and `go fmt` on every file before each commit: it ensures your code is properly formatted.

Also, for your contribution to be accepted, every one of your commits must be "Signed-off". This is done by committing using this command: `git commit --signoff`

By signing off your commits, you agree to the following agreement, also known as [Developer Certificate of Origin](http://developercertificate.org/): it assures everyone that the code you're submitting is yours or that you have rights to submit it.

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```
