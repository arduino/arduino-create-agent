// Copyright 2022 Arduino SA
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package upload allows to upload sketches into a board connected to the computer
// It can do it via serial port or via network
//
// **Usage for a serial upload**
//
// Make sure that you have a compiled sketch somewhere on your disk
//
//	```go
//	commandline = ``"/usr/bin/avrdude" "-C/usr/bin/avrdude.conf" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:./sketch.hex:i"``
//	err := upload.Serial("/dev/ttyACM0", commandline, upload.Extra{}, nil)
//	```
//
// note that the commandline contains the path of the sketch (sketch.hex)
//
// **Usage for a network upload**
//
// Make sure that you have a compiled sketch somewhere on your disk
//
//	```go
//	 err := upload.Network("127.0.10.120", "arduino:avr:yun, "./sketch.hex", "", upload.Auth{}, nil)
//	```
//
// The commandline can be empty or it can contain instructions (depends on the board)
//
// **Resolving commandlines**
//
// If you happen to have an unresolved commandline (full of {} parameters) you can resolve it
//
//	```go
//	 t := tools.Tools{}
//	 commandline = upload.Resolve("/dev/ttyACM0", "arduino:avr:leonardo", "./sketch.hex", commandline, upload.Extra{}, t)
//	 ```
//
// 't' must implement the locater interface (the Tools package does!)
//
// **Logging**
// If you're interested in the output of the commands, you can implement the logger interface. Here's an example:
//
//	```go
//	 logger := logrus.New()
//	 logger.Level = logrus.DebugLevel
//	 upload.Serial("/dev/ttyACM0", commandline, upload.Extra{}, logger)
//	 ```
package upload
