[![License: GPL v2](https://img.shields.io/badge/License-GPL%20v2-blue.svg)](https://www.gnu.org/licenses/old-licenses/gpl-2.0.en.html)

arduino-create-agent
====================

The Arduino Create Agent is a single binary that will sit on the traybar and work in the background. It allows you to use the [Arduino Create applications](https://create.arduino.cc) to seamlessly upload code to any USB connected Arduino board (or Yún in LAN) directly from the browser.

## Architecture
```
+-------------------------------+
|                               |
|            Browser            |
|                               |   Web socket   +----------------------+   flashes   +---------------+
| +---------------------------+ |<-------------->|                      +------------>|               |
| |                           | |                | Arduino Create Agent |             | Arduino Board |
| | Arduino Create Web Editor | +--------------->|                      |<------------+               |
| |                           | |   REST API     +----------------------+   serial    +---------------+
| +---------------------------+ |
+-------------------------------+
```

## Installation
Get the [latest version](https://github.com/arduino/arduino-create-agent/releases) of the Agent for all supported platforms or complete the [Getting Started](https://create.arduino.cc/getting-started/plugin/welcome).

## Apple M1 support
At the moment the new Apple Silicon Macs released in November 2020, like the new [MacBook Pro 13"](https://www.apple.com/macbook-pro-13/), [MacBook Air](https://www.apple.com/macbook-air/) and [Mac mini](https://www.apple.com/mac-mini/) models with the [Apple M1](https://www.apple.com/mac/m1/) chip are currently NOT supported by the Create Agent.

## Documentation
The documentation has been moved to the [wiki](https://github.com/arduino/arduino-create-agent/wiki) page. There you can find:
- [Advanced usage](https://github.com/arduino/arduino-create-agent/wiki/Advanced-usage): explaining how to use multiple configurations and how to use the agent with a proxy.
- [Agent Beta Program](https://github.com/arduino/arduino-create-agent/wiki/Agent-Beta-Program)
- [Developement](https://github.com/arduino/arduino-create-agent/wiki/Developement): containing useful info to help in development
- [Disable Autostart](https://github.com/arduino/arduino-create-agent/wiki/Disable-Autostart)
- [How to compile on Raspberry Pi](https://github.com/arduino/arduino-create-agent/wiki/How-to-compile-on-Raspberry-Pi)
- [How to use crashreport functionality](https://github.com/arduino/arduino-create-agent/wiki/How-to-use-crashreport-functionality)
- [How to use the agent](https://github.com/arduino/arduino-create-agent/wiki/How-to-use-the-agent)

## Contributing
### Submitting an issue

When submitting a new issue please search for duplicates before creating a new one. Help us by providing  useful context and information. Please attach the output of the commands running at the debug console or attach [crash reports](https://github.com/arduino/arduino-create-agent/wiki/How-to-use-crashreport-functionality) if useful.

### Submitting a pull request
We are glad you want to contribute with code: that's the best way to help this software.

Your contribution is adding or modifying existing behaviour, please always refer to an existing issue or open a new one before contributing. We are trying to use [Test Driven Development](https://en.wikipedia.org/wiki/Test-driven_development) in the near future: please add one or more tests that prove that your contribution is good and is working as expected, it will help us a lot.

Be sure to use `go vet` and `go fmt` on every file before each commit: it ensures your code is properly formatted.

Also, for your contribution to be accepted, every one of your commits must be "Signed-off". This is done by committing using this command: `git commit --signoff`

By signing off your commits, you agree to the following agreement, also known as [Developer Certificate of Origin](http://developercertificate.org/): it assures everyone that the code you're submitting is yours or that you have rights to submit it.

## Authors and acknowledgment
arduino-create-agent is a fork of @[johnlauer](https://github.com/johnlauer)'s [serial-port-json-server](https://github.com/johnlauer/serial-port-json-server) (which we really want to thank for his kindness and great work)

The history has been rewritten to keep the repo small (thus removing all binaries committed in the past)

## License
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
