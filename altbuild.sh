#!/bin/bash

echo "About to cross-compile Serial Port JSON Server"
#echo '$0 = ' $0
#echo '$1 = ' $1
#echo '$2 = ' $2

if [ "$1" = "" ]; then
        echo "You need to pass in the version number as the first parameter."
        exit
fi

# turn on echo
set -x
#set -v

# Windows x32 and x64, Linux
goxc -bc=windows,linux -d="." -pv=$1 -tasks-=pkg-build default -GOARM=6

# Rename arm to arm6
#set +x
FILE=$1'/serial-port-json-server_'$1'_linux_arm.tar.gz'
FILE2=$1'/serial-port-json-server_'$1'_linux_armv6.tar.gz'
#set -x
mv $FILE $FILE2

# Special build for armv7 for BBB and Raspi2
goxc -bc=linux,arm -d="." -pv=$1 -tasks-=pkg-build default -GOARM=7
FILE3=$1'/serial-port-json-server_'$1'_linux_armv7.tar.gz'
mv $FILE $FILE3

# Special build for armv8
goxc -bc=linux,arm -d="." -pv=$1 -tasks-=pkg-build default -GOARM=8
FILE4=$1'/serial-port-json-server_'$1'_linux_armv8.tar.gz'
mv $FILE $FILE4

