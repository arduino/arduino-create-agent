# git submodule init
# git submodule update

bootstrapPlatforms()
{
	#export PATH=$PATH:/home/martino/osxcross/target/bin
	cd $GOROOT/src
	env CC_FOR_TARGET=o64-clang CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 ./make.bash --no-clean
	env CC_FOR_TARGET=gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 ./make.bash --no-clean
	env CC_FOR_TARGET=gcc CGO_ENABLED=1 GOOS=linux GOARCH=386 ./make.bash --no-clean
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm ./make.bash --no-clean
	env CC_FOR_TARGET=i686-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=386 ./make.bash --no-clean
}

compilePlatform()
{
	GOOS=$1
	GOARCH=$2
	CC=$3
	cp -r arduino/tools_$GOOS  arduino/tools
	echo "compiling for $GOOS, $GOARCH"
	rm snapshot/Arduino_WebIDE_Bridge-$GOOS-$GOARCH.zip
	env GOOS=$GOOS GOARCH=$GOARCH CC=$CC CGO_ENABLED=1 go build -o="Arduino_WebIDE_Bridge"
	if [ $? != 0 ]
	then
	echo "Target $GOOS, $GOARCH failed"
	exit 1
	fi
	zip -r snapshot/Arduino_WebIDE_Bridge-$GOOS-$GOARCH.zip arduino/hardware arduino/tools arduino/resources Arduino_WebIDE_Bridge > /dev/null
	rm -rf arduino/tools
	ls -la snapshot/Arduino_WebIDE_Bridge-$GOOS-$GOARCH.zip
}

compilePlatformLinux()
{
	GOOS=$1
	GOARCH=$2
	CC=$3
	if [ $GOARCH == "386" ]
	then
	TOOLS_DIR=32
	fi
	if [ $GOARCH == "amd64" ]
	then
	TOOLS_DIR=64
	fi
	cp -r arduino/tools_$GOOS\_$TOOLS_DIR  arduino/tools
	echo "compiling for $GOOS, $GOARCH"
	rm snapshot/Arduino_WebIDE_Bridge-$GOOS-$GOARCH.zip
	env GOOS=$GOOS GOARCH=$GOARCH CC=$CC CGO_ENABLED=1 go build -o="Arduino_WebIDE_Bridge"
	if [ $? != 0 ]
	then
	echo "Target $GOOS, $GOARCH failed"
	exit 1
	fi
	zip -r snapshot/Arduino_WebIDE_Bridge-$GOOS-$GOARCH.zip arduino/hardware arduino/tools arduino/resources Arduino_WebIDE_Bridge > /dev/null
	rm -rf arduino/tools
	ls -la snapshot/Arduino_WebIDE_Bridge-$GOOS-$GOARCH.zip
}

compilePlatformNoCGO()
{
	GOOS=$1
	GOARCH=$2
	cp -r arduino/tools_$GOOS\_$GOARCH  arduino/tools
	echo "compiling for $GOOS, $GOARCH"
	rm snapshot/Arduino_WebIDE_Bridge-$GOOS-$GOARCH.zip
	if [ $GOARCH == "arm" ]
	then
	env GOARM=6 GOOS=$GOOS GOARCH=$GOARCH go build -o="Arduino_WebIDE_Bridge"
	else
	env GOOS=$GOOS GOARCH=$GOARCH go build -o="Arduino_WebIDE_Bridge"
	fi
	if [ $? != 0 ]
	then
	echo "Target $GOOS, $GOARCH failed"
	exit 1
	fi
	zip -r snapshot/Arduino_WebIDE_Bridge-$GOOS-$GOARCH.zip arduino/hardware arduino/tools arduino/resources Arduino_WebIDE_Bridge > /dev/null
	rm -rf arduino/tools
	ls -la snapshot/Arduino_WebIDE_Bridge-$GOOS-$GOARCH.zip
}

compilePlatform darwin amd64 o64-clang
compilePlatformLinux linux 386 gcc
compilePlatformLinux linux amd64 gcc
compilePlatformNoCGO linux arm
compilePlatform windows 386 i686-w64-mingw32-gcc


exit 0