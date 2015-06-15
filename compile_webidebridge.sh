# git submodule init
# git submodule update

#dependencies
#go install github.com/sanbornm/go-selfupdate

VERSION=xxx
APP_NAME=Arduino_Create_Bridge

# OUTPUT-COLORING
red='\e[0;31m'
green='\e[0;32m'
NC='\e[0m' # No Color

extractVersionFromMain()
{
	VERSION=`grep versionFloat main.go | cut -d "(" -f2 | cut -d ")" -f1`
}

createZipEmbeddableFileArduino()
{
    echo 'In createZipEmbeddableFileArduino'
	GOOS=$1
	GOARCH=$2

	# start clean
	rm arduino/arduino.zip
	rm -r arduino/arduino
	mkdir arduino/arduino
	cp -r arduino/hardware arduino/tools\_$GOOS\_$GOARCH arduino/boards.json arduino/arduino
	cp config.ini arduino
	mv arduino/arduino/tools* arduino/arduino/tools
	cd arduino
	zip -r arduino.zip arduino/* config.ini > /dev/null
	cd ..
	cat arduino/arduino.zip >> $3
	zip  --adjust-sfx $3
	mkdir -p snapshot/$GOOS\_$GOARCH
	cp $3 snapshot/$GOOS\_$GOARCH/$3
	ls -la snapshot/$GOOS\_$GOARCH/$3
}

bootstrapPlatforms()
{
    echo 'In bootstrapPlatforms'
	#export PATH=$PATH:/home/martino/osxcross/target/bin
	cd $GOROOT/src
	env CC_FOR_TARGET=clang CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 ./make.bash --no-clean
	env CC_FOR_TARGET=gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 ./make.bash --no-clean
	env CC_FOR_TARGET=gcc CGO_ENABLED=1 GOOS=linux GOARCH=386 ./make.bash --no-clean
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm ./make.bash --no-clean
	env CC_FOR_TARGET=i686-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=386 ./make.bash --no-clean
}

set -x
compilePlatform()
{
    echo 'In compilePlatform'
	GOOS=$1
	GOARCH=$2
	CC=$3
	CGO_ENABLED=$4
	NAME=$APP_NAME
	if [ $GOOS == "windows" ]
	then
	NAME=$NAME".exe"
	fi
	env GOOS=$GOOS GOARCH=$GOARCH CC=$CC CXX=$CC CGO_ENABLED=$CGO_ENABLED go build -o=$NAME
	if [ $? != 0 ]
	then
	echo -e "${red}Target $GOOS, $GOARCH failed${NC}"
	exit 1
	fi
	echo createZipEmbeddableFileArduino $GOOS $GOARCH $NAME
	createZipEmbeddableFileArduino $GOOS $GOARCH $NAME
	GOOS=$GOOS GOARCH=$GOARCH go-selfupdate $NAME $VERSION
	rm -rf $NAME*
}

extractVersionFromMain
compilePlatform darwin amd64 clang 1
#compilePlatformLinux linux 386 gcc
compilePlatform linux amd64 gcc 1
compilePlatform linux arm 0
compilePlatform windows 386 i686-w64-mingw32-gcc 1


exit 0
