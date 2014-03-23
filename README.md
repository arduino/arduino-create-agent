serial-port-json-server
=======================

A serial port JSON websocket &amp; web server that runs in your system tray on 
Windows, Mac, or Linux that lets you communicate with your serial port from a 
web application. This enables web apps to be written that can communicate with 
your local serial device such as an Arduino, CNC controller, or any device that 
communicates over the serial port.

The app is written in Go. It has an embedded web server and websocket server.
The server runs on the standard port of localhost:8989. You can connect to
it locally with your browser to interact.

The app is one executable with everything you need and is available ready-to-go
for every major platform.

If you are a web developer and want to write a web application that connects
to somebody's local serial port server, then you simply need to create a new DNS entry
under your domain called something like serialjson.mydomain.com and map it to 
127.0.0.1 to solve for the cross-domain Ajax policy. Then simply create a websocket
connection to serialjson.mydomain.com and you will be directly interacting with
that user's serial port.

For example, if you wanted to create a Gcode Sender web app to enable people to send
3D print or milling commands from your site, this would be a perfect use case. Or if
you've created an oscilloscope web app that connects to an Arduino, it would be another
great use case. Finally you can write web apps that interact with a user's local hardware.
