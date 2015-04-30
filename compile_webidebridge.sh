# git submodule init
# git submodule update

if [ `uname` == Linux ]
then
cp -r arduino/tools_linux_64  arduino/tools
CGO_ENABLED=1 goxc -os="linux" -arch="amd64" --include="arduino/hardware,arduino/tools,arduino/resources" -n="Arduino_WebIDE_Bridge" -d=.
rm -rf arduino/tools
exit 0
cp -r arduino/tools_linux_32  arduino/tools
CGO_ENABLED=1 goxc -os="linux" -arch="386" --include="arduino/hardware,arduino/tools,arduino/resources" -n="Arduino_WebIDE_Bridge" -d=.
rm -rf arduino/tools

cp -r arduino/tools_linux_arm  arduino/tools
CGO_ENABLED=1 goxc -os="linux" -arch="arm" --include="arduino/hardware,arduino/tools,arduino/resources" -n="Arduino_WebIDE_Bridge" -d=.
rm -rf arduino/tools

cp -r arduino/tools_windows  arduino/tools
goxc -os="windows" -arch="386" --include="arduino/hardware,arduino/tools,arduino/resources" -n="Arduino_WebIDE_Bridge" -d=. -build-ldflags="-H=windowsgui"
rm -rf arduino/tools
fi

if [ `uname` == Darwin ]
then
cp -r arduino/tools_darwin  arduino/tools
CGO_ENABLED=1 goxc -os="darwin" -arch="amd64" --include="arduino/hardware,arduino/tools,arduino/resources" -n="Arduino_WebIDE_Bridge" -d=.
rm -rf arduino/tools
fi
