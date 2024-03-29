Package upload
=====================

    import "github.com/arduino/arduino-create-agent/upload"

Package upload allows to upload sketches into a board connected to the
computer It can do it via serial port

**Usage for a serial upload**

Make sure that you have a compiled sketch somewhere on your disk

```go
commandline = ``"/usr/bin/avrdude" "-C/usr/bin/avrdude.conf" -v -patmega32u4 -cavr109 -P/dev/ttyACM0 -b57600 -D "-Uflash:w:./sketch.hex:i"``
err := upload.Serial("/dev/ttyACM0", commandline, upload.Extra{}, nil)
```

note that the commandline contains the path of the sketch (sketch.hex)

**Resolving commandlines**

If you happen to have an unresolved commandline (full of {} parameters) you can
resolve it

```go
 t := tools.Tools{}
 commandline = upload.Resolve("/dev/ttyACM0", "arduino:avr:leonardo", "./sketch.hex", commandline, upload.Extra{}, t)
 ```

t must implement the locater interface (the Tools package does!)

**Logging** If you're interested in the output of the commands, you can
implement the logger interface. Here's an example:

```go
 logger := logrus.New()
 logger.Level = logrus.DebugLevel
 upload.Serial("/dev/ttyACM0", commandline, upload.Extra{}, logger)
 ```



Variables
---------


```go
var Busy = false
```
Busy tells wether the upload is doing something

Functions
---------


```go
func Kill()
```

Kill stops any upload process as soon as possible


```go
func Resolve(port, board, file, commandline string, extra Extra, t Locater) (string, error)
```

Resolve replaces some symbols in the commandline with the appropriate values it
can return an error when looking a variable in the Locater


```go
func Serial(port, commandline string, extra Extra, l Logger) error
```

Serial performs a serial upload

Types
-----


```go
type Extra struct {
    Use1200bpsTouch   bool   `json:"use_1200bps_touch"`
    WaitForUploadPort bool   `json:"wait_for_upload_port"`
    Network           bool   `json:"network"`
}
```
Extra contains some options used during the upload


```go
type Locater interface {
    GetLocation(command string) (string, error)
}
```
Locater can return the location of a tool in the system


```go
type Logger interface {
    Debug(args ...interface{})
    Info(args ...interface{})
}
```
Logger is an interface implemented by most loggers (like logrus)
