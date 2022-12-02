#/bin/sh

if [ -z "$GOPATH" ]; then
    echo GOPATH environment variable not set
    exit
fi

if [ ! -e "$GOPATH/bin/2goarray" ]; then
    echo "Installing 2goarray..."
    go install github.com/cratonica/2goarray@latest
    if [ $? -ne 0 ]; then
        echo Failure executing go install github.com/cratonica/2goarray@latest
        exit
    fi
fi

if [ -z "$1" ]; then
    echo Please specify a PNG file
    exit
fi

if [ ! -f "$1" ]; then
    echo $1 is not a valid file
    exit
fi    

if [ -z "$2" ]; then
    OUTPUT="$1.go"
else
    OUTPUT=$2
fi

echo Generating $OUTPUT
echo "// +build linux darwin" > $OUTPUT
echo >> $OUTPUT
cat "$1" | $GOPATH/bin/2goarray Data icon >> $OUTPUT
if [ $? -ne 0 ]; then
    echo Failure generating $OUTPUT
    exit
fi
echo Finished
