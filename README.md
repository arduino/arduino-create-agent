arduino-create-agent
====================

## Installation
Get the latest version of the Agent for all supported platforms:

* [Windows](http://downloads.arduino.cc/CreateBridgeStable/ArduinoCreateAgent-1.1-windows-installer.exe)
* [MacOSX](http://downloads.arduino.cc/CreateBridgeStable/ArduinoCreateAgent-1.1-osx-installer.dmg)
* [Linux x64](http://downloads.arduino.cc/CreateBridgeStable/ArduinoCreateAgent-1.1-linux-x64-installer.run)

arduino-create-agent is a fork of @johnlauer's serial-port-json-server (which we really want to thank for his kindness and great work)

The history has been rewritten to keep the repo small (thus removing all binaries committed in the past)

## Disable Autostart

### Windows
1. Type "Task Manager in the Windows Search Bar"

![Type "Task Manager in the Windows Search Bar"](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/windows/01.png)
2. Select the Startup tab

![Select the Startup tab](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/windows/02.png)
3. Select the autostart file

![Select the autostart file](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/windows/03.png)
4. Disable it

![Disable it](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/windows/04.png)

### Mac OSX
1. Open Finder, click on Go menu, select 'Go to Folder'

![Open Finder, click on Go menu, select 'Go to Folder'](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/mac/01.png)
2. Type the directory containing the autolauncher file, change <username> with your Mac username, by default the directory is /Users/username/Library/LaunchAgents

![Type the directory containing the autolauncher file, change <username> with your Mac username, by default the directory is /Users/username/Library/LaunchAgents](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/mac/02.png)
3. Select the ArduinoCreateAgent.plist file

![Select the ArduinoCreateAgent.plist file](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/mac/03.png)
4. Right click on the file name and select 'Move to Trash'

![Right click on the file name and select 'Move to Trash'](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/mac/04.png)

---
The command line way:
```
$ launchctl unload ~/Library/LaunchAgents/ArduinoCreateAgent.plist
```

### Linux
1. Show hidden files

![Show hidden files](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/linux/01.png)
2. Select the .config dir in your home

![Select the .config dir in your home](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/linux/02.png)
3. Select the autostart dir

![Select the autostart dir](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/linux/03.png)
4. Move the file to the trash

![Move the file to the trash](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/linux/04.png)

---
The command line way:

Just remove the autostart file in your desktop manager, in Ubuntu is:
```
$ rm $HOME/.config/autostart/arduino-create-agent.desktop
```
To start manually the agent you can open the file at:
```
$ nohup $HOME/ArduinoCreateAgent-1.1/Arduino_Create_Bridge &
```
or in the location selected during the installation

## Using multiple configurations

The agent support multiple configuration files. When multiple configurations are found by the agent, the tray icon menu is expanded to contain the different configurations:

![Agent multiple configuration tray ycon](https://raw.githubusercontent.com/arduino/arduino-create-agent/devel/images/linux/tray-icon-multiple-profiles.png)

The default `config.ini` file contains common configurations, every other config file inherit from it.

To create multiple configuration files:
- stop the agent: tray bar icon -> pause then tray bar icon -> kill
- find the `config.ini` file that is present in the `arduino-create-agent` installation folder 
- copy `config.ini` or create a new ini file (e.g. `example.ini`) with content:
    ```ini
    name = your configuration nanem
    ```
    add in this file other configuration options (you can override inherited values from `config.ini`)
- restart the agent
- click the tray bar icon and select the new configuration

# Contributing

Please use the current latest version:

### Windows
* [Windows with Edge dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-windows-installer-edge.exe)
* [Windows with Chrome dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-windows-installer.exe)
* [Windows with Firefox dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-windows-installer-firefox.exe)

### MacOSX
* [MacOSX with Safari dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-osx-installer-safari.dmg)
* [MacOSX with Chrome dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-osx-installer-chrome.dmg)
* [MacOSX with Firefox dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-osx-installer-firefox.dmg)

### Linux
* [Linux x64 with Chrome dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-linux-x64-installer-chrome.run)
* [Linux x64 with Firefox dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-linux-x64-installer-firefox.run)


## How to use it
The arduino create agent is a single binary that reads from a configuration file. Upon launching it will sit on the traybar and work in the background.

It will listen to http and websocket connections on a range of ports from `8990` to `9000`.

### Discover the port
You should make GET request to the `/info` endpoint on the possible ports, until you find a reply:

    $ curl http://127.0.0.1:8990/info
    curl: (7) Failed to connect to 127.0.0.1 port 8990: Connection refused
    $ curl http://127.0.0.1:8991/info
    
    $ curl http://127.0.0.1:8992/info
    {"http":"http://127.0.0.1:8992","https":"https://127.0.0.1:8991","version":"1.0.36","ws":"ws://127.0.0.1:8992","wss":"wss://127.0.0.1:8991"}

The reply will contain a json with info about the version and the http and https endpoints to use

### Open a websocket
Most of the commands can be performed with websocket instructions. We use a library called [socket.io](https://cdnjs.cloudflare.com/ajax/libs/socket.io/1.3.5/socket.io.min.js) to handle the messages.
Once you have the websocket endpoint you need you can:

```javascript
var socket = io(endpoint);
socket.on('connect', function () {
    socket.emit('command', yourCommand);

    socket.on('command', function () {
        // Your code to handle messages
    })
}
```

### Use the debug console
By clicking on the tray icon and going to the debug console you can try most of the websocket commands. The first command you should type in is:

    log on

### List the boards
To get a json list of the connected boards you can issue the command:

    list

You will receive an object of all the boards connected with USB or over the network:

```json
{
  "Ports":[
    {
      "Name":"/dev/ttyACM0",
      "SerialNumber":"",
      "DeviceClass":"",
      "IsOpen":false,
      "IsPrimary":false,
      "Baud":0,
      "BufferAlgorithm":"",
      "Ver":"1.0.36",
      "NetworkPort":false,
      "VendorID":"0x2341",
      "ProductID":"0x8036"
    }
  ],
  "Network":false
}{
  "Ports":[
    {
      "Name":"192.168.1.101",
      "SerialNumber":"",
      "DeviceClass":"",
      "IsOpen":false,
      "IsPrimary":false,
      "Baud":0,
      "BufferAlgorithm":"",
      "Ver":"1.0.36",
      "NetworkPort":true,
      "VendorID":"board=Arduino Y\\195\\186n Shield distro_version=0.1",
      "ProductID":"Shield"
    }
  ],
  "Network":true
}
```

## Open/Close ports

To read input from a board connected to USB you must first open the port with the command

    open /dev/ttyACM0 9600

where you should replace /dev/ttyACM0 with the actual port and 9600 with the baud.

You will receive a message like:

```json
{
  "Cmd":"Open",
  "Desc":"Got register/open on port.",
  "Port":"/dev/ttyACM0",
  "IsPrimary":true,
  "Baud":9600,
  "BufferType":""
}
```

or

```json
{
  "Cmd":"OpenFail",
  "Desc":"Error opening port. Serial port busy",
  "Port":"/dev/ttyACM0",
  "Baud":9600
}
```

You can then close the port with

    close /dev/ttyACM0

You will receive a message like:

```json
{
  "Cmd":"Close",
  "Desc":"Got unregister/close on port.",
  "Port":"/dev/ttyACM0",
  "Baud":9600
}
```

or


```json
{
  "Error":"We could not find the serial port /dev/ttyACM0 that you were trying to close."
}
```

### Receiving and sending data

While a port is open you can send input with

    send /dev/ttyACM0 hello

with a reply like

```json
{"Cmd":"Queued","QCnt":1,"Ids":[""],"D":["hello"],"Port":"/dev/ttyACM0"}
{"Cmd":"Write","QCnt":0,"Id":"","P":"/dev/ttyACM0"}
{"Cmd":"CompleteFake","Id":"","P":"/dev/ttyACM0"}
```

You can receive output from the serial port by listening to messages like this:

```json
{
  "D":"output string\r\n"
}
```

### Download a tool
You can download a tool on the computer with a command like

    downloadtool avrdude 6.0.1-arduino5 replace

receiving a reply like

```json
{
  "DownloadStatus": "Success",
  "Msg":"Map Updated"
}
```

The syntax of the command is:

    downloadtool {{name}} {{version}} {{behaviour}}

where `version` can be a version number of the string "latest", and `behaviour` can be
"keep" (which skips the download if the tool already exists) and "replace" (which will download it again).

### Upload
You can upload a binary sketch to a board connected to a port with a POST request to be made at the http endpoint.

The payload is a json object that looks like this:

```json
{
  "board":"arduino:avr:leonardo",
  "port":"/dev/ttyACM1",
  "commandline":"\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose} -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
  "signature":"97db97ced2c",
  "hex":"OjEwMDAwMDAwMEM5NEU1MDAwQzk0MEQwMTBDOTQwRDAxMEM5NDBEMDE2MQ0KOjEwMDAxMDAwMEM5NDBEMDEwQzk0M",
  "filename":"Blink.ino.hex",
  "extra":{
    "auth":{
      "password":null
    },
    "wait_for_upload_port":true,
    "use_1200bps_touch":true,
    "network":false,
    "params_verbose":"-v",
    "params_quiet":"-q -q",
    "verbose":true
  }
}
```

- commandline is the command to execute to perform the upload. This is, for example, avrdude on a Leonardo.

- hex contains the sketch hex encoded in base64

- signature is the signature of the commandline signed with the private key that matches the public key contained in the config.ini of the arduino-create-agent

The results of the upload will be delivered via websocket with messages that look like:

```json
{"Msg":"avrdude: verifying ...","ProgrammerStatus":"Busy"}
{"Msg":"avrdude done. Thank you.","ProgrammerStatus":"Busy"}
{"Flash":"Ok","ProgrammerStatus":"Done"}
```

---

## Development

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
* libappindicator (Linux only on Ubuntu `sudo apt-get install libappindicator1`)
* [go-selfupdate] (https://github.com/sanbornm/go-selfupdate) if you want to test automatic updates

## Submitting an issue

Please attach the output of the commands running at the debug console if useful.

## Submitting a pull request

We are glad you want to contribute with code: that's the best way to help this software.

Your contribution is adding or modifying existing behaviour, please always refer to an existing issue or open a new one before contributing. We are are trying to use [Test Driven Development](https://en.wikipedia.org/wiki/Test-driven_development) in the near future: please add one or more tests that prove that your contribution is good and is working as expected, it will help us a lot.

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


## Creating a release
Just create a new release on GitHub, and our drone server will build and upload
the compiled binaries for every architecture in a zip file in the release itself.
