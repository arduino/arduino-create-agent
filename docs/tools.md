# tools
--
    import "github.com/arduino/arduino-create-agent/tools"


## Usage

#### type Tools

```go
type Tools struct {
	Directory string
	IndexURL  string
	Logger    log.StdLogger
}
```

Tools handle the tools necessary for an upload on a board. It provides a means
to download a tool from the arduino servers.

- *Directory* contains the location where the tools are downloaded.
- *IndexURL* contains the url where the tools description is contained.
- *Logger* is a StdLogger used for reporting debug and info messages
- *installed* contains a map of the tools and their exact location

Usage: You have to instantiate the struct by passing it the required parameters:

    _tools := tools.Tools{
        Directory: "/home/user/.arduino-create",
        IndexURL: "http://downloads.arduino.cc/packages/package_index.json"
        Logger: log.Logger
    }

#### func (*Tools) Download

```go
func (t *Tools) Download(name, version, behaviour string) error
```
Download will parse the index at the indexURL for the tool to download. It will
extract it in a folder in .arduino-create, and it will update the Installed map.

name contains the name of the tool. version contains the version of the tool.
behaviour contains the strategy to use when there is already a tool installed

If version is "latest" it will always download the latest version (regardless of
the value of behaviour)

If version is not "latest" and behaviour is "replace", it will download the
version again. If instead behaviour is "keep" it will not download the version
if it already exists.

#### func (*Tools) GetLocation

```go
func (t *Tools) GetLocation(command string) (string, error)
```
GetLocation extracts the toolname from a command like

#### func (*Tools) Init

```go
func (t *Tools) Init()
```
Init creates the Installed map and populates it from a file in .arduino-create
