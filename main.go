// Version 1.5
// Supports Windows, Linux, Mac, and Raspberry Pi, Beagle Bone Black

package main

import (
	"flag"
	"go/build"
	"log"
	"net/http"
	//"path/filepath"
	"text/template"
)

var (
	version = "1.5"
	addr    = flag.String("addr", ":8989", "http service address")
	assets  = flag.String("assets", defaultAssetPath(), "path to assets")
	//homeTempl *template.Template
)

func defaultAssetPath() string {
	//p, err := build.Default.Import("gary.burd.info/go-websocket-chat", "", build.FindOnly)
	p, err := build.Default.Import("github.com/johnlauer/serial-port-json-server", "", build.FindOnly)
	if err != nil {
		return "."
	}
	return p.Dir
}

func homeHandler(c http.ResponseWriter, req *http.Request) {
	homeTemplate.Execute(c, req.Host)
}

func main() {

	// setup logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	//getList()
	flag.Parse()
	f := flag.Lookup("addr")
	log.Println("Version:" + version)
	log.Print("Started server and websocket on localhost" + f.Value.String())
	//homeTempl = template.Must(template.ParseFiles(filepath.Join(*assets, "home.html")))

	// launch the hub routine which is the singleton for the websocket server
	go h.run()
	// launch our serial port routine
	go sh.run()
	// launch our dummy data routine
	//go d.run()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", wsHandler)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("Error ListenAndServe:", err)
	}
}

var homeTemplate = template.Must(template.New("home").Parse(homeTemplateHtml))

// If you navigate to this server's homepage, you'll get this HTML
// so you can directly interact with the serial port server
const homeTemplateHtml = `<!DOCTYPE html>
<html>
<head>
<title>Serial Port Example</title>
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.4.2/jquery.min.js"></script>
<script type="text/javascript">
    $(function() {

    var conn;
    var msg = $("#msg");
    var log = $("#log");

    function appendLog(msg) {
        var d = log[0]
        var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
        msg.appendTo(log)
        if (doScroll) {
            d.scrollTop = d.scrollHeight - d.clientHeight;
        }
    }

    $("#form").submit(function() {
        if (!conn) {
            return false;
        }
        if (!msg.val()) {
            return false;
        }
        conn.send(msg.val() + "\n");
        msg.val("");
        return false
    });

    if (window["WebSocket"]) {
        conn = new WebSocket("ws://{{$}}/ws");
        conn.onclose = function(evt) {
            appendLog($("<div><b>Connection closed.</b></div>"))
        }
        conn.onmessage = function(evt) {
            appendLog($("<div/>").text(evt.data))
        }
    } else {
        appendLog($("<div><b>Your browser does not support WebSockets.</b></div>"))
    }
    });
</script>
<style type="text/css">
html {
    overflow: hidden;
}

body {
    overflow: hidden;
    padding: 0;
    margin: 0;
    width: 100%;
    height: 100%;
    background: gray;
}

#log {
    background: white;
    margin: 0;
    padding: 0.5em 0.5em 0.5em 0.5em;
    position: absolute;
    top: 0.5em;
    left: 0.5em;
    right: 0.5em;
    bottom: 3em;
    overflow: auto;
}

#form {
    padding: 0 0.5em 0 0.5em;
    margin: 0;
    position: absolute;
    bottom: 1em;
    left: 0px;
    width: 100%;
    overflow: hidden;
}

</style>
</head>
<body>
<div id="log"></div>
<form id="form">
    <input type="submit" value="Send" />
    <input type="text" id="msg" size="64"/>
</form>
</body>
</html>
`
