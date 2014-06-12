serial-port-json-server
=======================

A serial port JSON websocket &amp; web server that runs from the command line on 
Windows, Mac, Linux, or Raspberry Pi that lets you communicate with your serial 
port from a web application. This enables web apps to be written that can 
communicate with your local serial device such as an Arduino, CNC controller, or 
any device that communicates over the serial port.

The app is written in Go. It has an embedded web server and websocket server.
The server runs on the standard port of localhost:8989. You can connect to
it locally with your browser to interact by visiting http://localhost:8989.
The websocket is technically running at ws://localhost/ws.

The app is one executable with everything you need and is available ready-to-go
for every major platform.

If you are a web developer and want to write a web application that connects
to somebody's local or remote serial port server, then you simply need to create a 
websocket connection to the localhost or remote host and you will be directly 
interacting with that user's serial port.

For example, if you wanted to create a Gcode Sender web app to enable people to send
3D print or milling commands from your site, this would be a perfect use case. Or if
you've created an oscilloscope web app that connects to an Arduino, it would be another
great use case. Finally you can write web apps that interact with a user's local hardware.

Thanks go to gary.burd.info for the websocket example in Go. Thanks also go to 
tarm/goserial for the serial port base implementation.