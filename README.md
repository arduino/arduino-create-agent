arduino-create-agent
====================

Version X.XX

Get the latest version of the Agent for all supported platforms:

* [Windows](http://downloads.arduino.cc/CreateBridgeStable/ArduinoCreateAgent-1.0-windows-installer.exe)
* [MacOSX](http://downloads.arduino.cc/CreateBridgeStable/ArduinoCreateAgent-1.0-osx-installer.dmg)
* [Linux x64](http://downloads.arduino.cc/CreateBridgeStable/ArduinoCreateAgent-1.0-linux-x64-installer.run)

arduino-create-agent is a fork of @johnlauer's serial-port-json-server (which we really want to thank for his kindness and great work)

The history has been rewritten to keep the repo small (thus removing all binaries committed in the past)

# Contributing

Please use the current latest version:

* [Windows dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-windows-installer.exe)
* [MacOSX dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-osx-installer.dmg) 
* [Linux x64 dev](http://downloads.arduino.cc/CreateBridge/staging/ArduinoCreateAgent-1.0-linux-x64-installer.run)

## Submitting an issue

Please attach the output of the commands running at the debug console if useful.

## Submitting a pull request

We are glad you want to contribute with code: that's the best way to help this software.

Your contribution is adding or modifying existing behaviour, please always refer to an existing issue or open a new one before contributing. We are are trying to use [Test Driven Development](https://en.wikipedia.org/wiki/Test-driven_development) in the near future: please add one or more tests that prove that your contribution is good and is working as expected, it will help us a lot.

Be sure to use `go vet` and `go fmt` on every file before each commit: it ensures your code is properly formatted.

Also, for your contribution to be accepted, everyone of your commits must be "Signed-off". This is done by commiting using this command: `git commit --signoff`

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
