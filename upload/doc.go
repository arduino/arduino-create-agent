// Package upload allows to upload sketches into a board connected to the computer
// It can do it via serial port or via network
//
// **Usage for a serial upload**
//
// Make sure that you have a compiled sketch somewhere on your disk
//
//   ```go
//   commandline = ``"/usr/bin/avrdude" "-C/usr/bin/avrdude.conf" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:./sketch.hex:i"``
//   err := upload.Serial("/dev/ttyACM0", commandline, upload.Extra{}, nil)
//   ```
//
// note that the commandline contains the path of the sketch (sketch.hex)
//
// **Usage for a network upload**
//
// Make sure that you have a compiled sketch somewhere on your disk
//
//   ```go
//    err := upload.Network("127.0.10.120", "arduino:avr:yun, "./sketch.hex", "", upload.Auth{}, nil)
//   ```
//
// The commandline can be empty or it can contain instructions (depends on the board)
//
// **Resolving commandlines**
//
// If you happen to have an unresolved commandline (full of {} parameters) you can resolve it
//
//   ```go
//    t := tools.Tools{}
//    commandline = upload.Resolve("/dev/ttyACM0", "arduino:avr:leonardo", "./sketch.hex", commandline, upload.Extra{}, t)
//    ```
//
// t must implement the locater interface (the Tools package does!)
//
// **Logging**
// If you're interested in the output of the commands, you can implement the logger interface. Here's an example:
//
//   ```go
//    logger := logrus.New()
//    logger.Level = logrus.DebugLevel
//    upload.Serial("/dev/ttyACM0", commandline, upload.Extra{}, logger)
//    ```
package upload
