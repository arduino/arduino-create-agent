// Version 1.82
// Supports Windows, Linux, Mac, and Raspberry Pi, Beagle Bone Black

package main

import (
	"flag"
	"os"
	"os/user"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/arduino/arduino-create-agent/tools"
	"github.com/arduino/arduino-create-agent/utilities"
	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	"github.com/kardianos/osext"
	"github.com/vharitonsky/iniflags"
	//"github.com/sanbornm/go-selfupdate/selfupdate" #included in update.go to change heavily
)

var (
	version              = "x.x.x-dev" //don't modify it, Jenkins will take care
	git_revision         = "xxxxxxxx"  //don't modify it, Jenkins will take care
	embedded_autoextract = false
	hibernate            = flag.Bool("hibernate", false, "start hibernated")
	verbose              = flag.Bool("v", true, "show debug logging")
	//verbose = flag.Bool("v", false, "show debug logging")
	isLaunchSelf = flag.Bool("ls", false, "launch self 5 seconds later")
	configIni    = flag.String("configFile", "config.ini", "config file path")
	regExpFilter = flag.String("regex", "usb|acm|com", "Regular expression to filter serial port list")
	gcType       = flag.String("gc", "std", "Type of garbage collection. std = Normal garbage collection allowing system to decide (this has been known to cause a stop the world in the middle of a CNC job which can cause lost responses from the CNC controller and thus stalled jobs. use max instead to solve.), off = let memory grow unbounded (you have to send in the gc command manually to garbage collect or you will run out of RAM eventually), max = Force garbage collection on each recv or send on a serial port (this minimizes stop the world events and thus lost serial responses, but increases CPU usage)")
	logDump      = flag.String("log", "off", "off = (default)")
	// hostname. allow user to override, otherwise we look it up
	hostname     = flag.String("hostname", "unknown-hostname", "Override the hostname we get from the OS")
	updateUrl    = flag.String("updateUrl", "", "")
	appName      = flag.String("appName", "", "")
	genCert      = flag.Bool("generateCert", false, "")
	port         string
	portSSL      string
	origins      = flag.String("origins", "", "Allowed origin list for CORS")
	address      = flag.String("address", "127.0.0.1", "The address where to listen. Defaults to localhost")
	signatureKey = flag.String("signatureKey", "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvc0yZr1yUSen7qmE3cxF\nIE12rCksDnqR+Hp7o0nGi9123eCSFcJ7CkIRC8F+8JMhgI3zNqn4cUEn47I3RKD1\nZChPUCMiJCvbLbloxfdJrUi7gcSgUXrlKQStOKF5Iz7xv1M4XOP3JtjXLGo3EnJ1\npFgdWTOyoSrA8/w1rck4c/ISXZSinVAggPxmLwVEAAln6Itj6giIZHKvA2fL2o8z\nCeK057Lu8X6u2CG8tRWSQzVoKIQw/PKK6CNXCAy8vo4EkXudRutnEYHEJlPkVgPn\n2qP06GI+I+9zKE37iqj0k1/wFaCVXHXIvn06YrmjQw6I0dDj/60Wvi500FuRVpn9\ntwIDAQAB\n-----END PUBLIC KEY-----", "Pem-encoded public key to verify signed commandlines")
	Tools        tools.Tools
	indexURL     = flag.String("indexURL", "http://downloads.arduino.cc/packages/package_index.json", "The address from where to download the index json containing the location of upload tools")
)

type NullWriter int

func (NullWriter) Write([]byte) (int, error) { return 0, nil }

type logWriter struct{}

func (u *logWriter) Write(p []byte) (n int, err error) {
	h.broadcastSys <- p
	return len(p), nil
}

var logger_ws logWriter

func homeHandler(c *gin.Context) {
	homeTemplate.Execute(c.Writer, c.Request.Host)
}

func launchSelfLater() {
	log.Println("Going to launch myself 2 seconds later.")
	time.Sleep(2 * 1000 * time.Millisecond)
	log.Println("Done waiting 2 secs. Now launching...")
}

func main() {

	flag.Parse()

	if *genCert == true {
		generateCertificates()
		os.Exit(0)
	}

	if *hibernate == false {

		go func() {

			// autoextract self
			src, _ := osext.Executable()
			dest := filepath.Dir(src)

			// Instantiate Tools
			usr, _ := user.Current()
			directory := usr.HomeDir + "/.arduino-create"
			Tools = tools.Tools{
				Directory: directory,
				IndexURL:  *indexURL,
				Logger:    log.New(),
			}
			Tools.Init()

			if embedded_autoextract {
				// save the config.ini (if it exists)
				if _, err := os.Stat(dest + "/" + *configIni); os.IsNotExist(err) {
					log.Println("First run, unzipping self")
					err := utilities.Unzip(src, dest)
					log.Println("Self extraction, err:", err)
				}

				if _, err := os.Stat(dest + "/" + *configIni); os.IsNotExist(err) {
					flag.Parse()
					log.Println("No config.ini at", *configIni)
				} else {
					flag.Parse()
					flag.Set("config", dest+"/"+*configIni)
					iniflags.Parse()
				}
			} else {
				flag.Set("config", dest+"/"+*configIni)
				iniflags.Parse()
			}

			// move CORS to config file compatibility, Vagrant version
			if *origins == "" {
				log.Println("Patching config.ini for compatibility")
				f, err := os.OpenFile(dest+"/"+*configIni, os.O_APPEND|os.O_WRONLY, 0666)
				if err != nil {
					panic(err)
				}
				_, err = f.WriteString("\norigins = http://webide.arduino.cc:8080\n")
				if err != nil {
					panic(err)
				}
				f.Close()
				restart("")
			}
			//log.SetFormatter(&log.JSONFormatter{})

			log.SetLevel(log.InfoLevel)

			log.SetOutput(os.Stderr)

			// see if we are supposed to wait 5 seconds
			if *isLaunchSelf {
				launchSelfLater()
			}

			log.Println("Version:" + version)

			// hostname
			hn, _ := os.Hostname()
			if *hostname == "unknown-hostname" {
				*hostname = hn
			}
			log.Println("Hostname:", *hostname)

			// turn off garbage collection
			// this is dangerous, as u could overflow memory
			//if *isGC {
			if *gcType == "std" {
				log.Println("Garbage collection is on using Standard mode, meaning we just let Golang determine when to garbage collect.")
			} else if *gcType == "max" {
				log.Println("Garbage collection is on for MAXIMUM real-time collecting on each send/recv from serial port. Higher CPU, but less stopping of the world to garbage collect since it is being done on a constant basis.")
			} else {
				log.Println("Garbage collection is off. Memory use will grow unbounded. You WILL RUN OUT OF RAM unless you send in the gc command to manually force garbage collection. Lower CPU, but progressive memory footprint.")
				debug.SetGCPercent(-1)
			}

			// see if they provided a regex filter
			if len(*regExpFilter) > 0 {
				log.Printf("You specified a serial port regular expression filter: %v\n", *regExpFilter)
			}

			// list serial ports
			portList, _ := GetList(false)
			log.Println("Your serial ports:")
			if len(portList) == 0 {
				log.Println("\tThere are no serial ports to list.")
			}
			for _, element := range portList {
				log.Printf("\t%v\n", element)

			}

			if !*verbose {
				log.Println("You can enter verbose mode to see all logging by starting with the -v command line switch.")
				log.SetOutput(new(NullWriter)) //route all logging to nullwriter
			}

			// launch the hub routine which is the singleton for the websocket server
			go h.run()
			// launch our serial port routine
			go sh.run()
			// launch our dummy data routine
			//go d.run()

			go discoverLoop()

			r := gin.New()

			socketHandler := wsHandler().ServeHTTP

			extraOriginStr := "https://create.arduino.cc, http://create.arduino.cc, https://create-dev.arduino.cc, http://create-dev.arduino.cc"

			for i := 8990; i < 9001; i++ {
				extraOriginStr = extraOriginStr + ", http://localhost:" + strconv.Itoa(i) + ", https://localhost:" + strconv.Itoa(i)
			}

			r.Use(cors.Middleware(cors.Config{
				Origins:         *origins + ", " + extraOriginStr,
				Methods:         "GET, PUT, POST, DELETE",
				RequestHeaders:  "Origin, Authorization, Content-Type",
				ExposedHeaders:  "",
				MaxAge:          50 * time.Second,
				Credentials:     true,
				ValidateHeaders: false,
			}))

			r.LoadHTMLFiles("templates/nofirefox.html")

			r.GET("/", homeHandler)
			r.GET("/certificate.crt", certHandler)
			r.DELETE("/certificate.crt", deleteCertHandler)
			r.POST("/upload", uploadHandler)
			r.GET("/socket.io/", socketHandler)
			r.POST("/socket.io/", socketHandler)
			r.Handle("WS", "/socket.io/", socketHandler)
			r.Handle("WSS", "/socket.io/", socketHandler)
			r.GET("/info", infoHandler)
			r.POST("/killbrowser", killBrowserHandler)
			r.POST("/pause", pauseHandler)
			r.POST("/update", updateHandler)

			go func() {
				// check if certificates exist; if not, use plain http
				if _, err := os.Stat(filepath.Join(dest, "cert.pem")); os.IsNotExist(err) {
					return
				}

				start := 8990
				end := 9000
				i := start
				for i < end {
					i = i + 1
					portSSL = ":" + strconv.Itoa(i)
					if err := r.RunTLS(*address+portSSL, filepath.Join(dest, "cert.pem"), filepath.Join(dest, "key.pem")); err != nil {
						log.Printf("Error trying to bind to port: %v, so exiting...", err)
						continue
					} else {
						log.Print("Starting server and websocket (SSL) on " + *address + "" + port)
						break
					}
				}
			}()

			go func() {
				start := 8990
				end := 9000
				i := start
				for i < end {
					i = i + 1
					port = ":" + strconv.Itoa(i)
					if err := r.Run(*address + port); err != nil {
						log.Printf("Error trying to bind to port: %v, so exiting...", err)
						continue
					} else {
						log.Print("Starting server and websocket on " + *address + "" + port)
						break
					}
				}
			}()

		}()
	}
	setupSysTray()
}

var homeTemplate = template.Must(template.New("home").Parse(homeTemplateHtml))

// If you navigate to this server's homepage, you'll get this HTML
// so you can directly interact with the serial port server
const homeTemplateHtml = `<!DOCTYPE html>
<html>
<head>
<title>Serial Port Example</title>
<script type="text/javascript" src="https://ajax.googleapis.com/ajax/libs/jquery/1.4.2/jquery.min.js"></script>
<script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/socket.io/1.3.5/socket.io.min.js"></script>
<script type="text/javascript">
    $(function() {

    var socket;
    var msg = $("#msg");
    var log = document.getElementById('log');
    var pause = document.getElementById('myCheck');
    var messages = [];
    var only_log = true;

    function appendLog(msg) {

		if (!pause.checked && (only_log == false || (!(msg.indexOf("{") == 0) && !(msg.indexOf("list") == 0) && only_log == true))) {
			messages.push(msg);
			if (messages.length > 100) {
				messages.shift();
			}
			var doScroll = log.scrollTop == log.scrollHeight - log.clientHeight;
			log.innerHTML = messages.join("<br>");
			if (doScroll) {
				log.scrollTop = log.scrollHeight - log.clientHeight;
			}
		}
    }

    $("#form").submit(function() {
        if (!socket) {
            return false;
        }
        if (!msg.val()) {
            return false;
        }
        socket.emit("command", msg.val());
        if (msg.val().indexOf("log off") != -1) {only_log = true}
        if (msg.val().indexOf("log on") != -1) {only_log = false}
        msg.val("");
        return false
    });

    if (window["WebSocket"]) {
    	if (window.location.protocol === 'https:') {
    		socket = io('https://{{$}}')
    	} else {
    		socket = io("http://{{$}}");
    	}
        socket.on("disconnect", function(evt) {
            appendLog($("<div><b>Connection closed.</b></div>"))
        });
        socket.on("message", function(evt) {
            appendLog(evt);
        });
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
    <input name="pause" type="checkbox" value="pause" id="myCheck"/> Pause
    <!--<input type="button" value="Install Certificate" onclick="window.open('http://localhost:8991/certificate.crt')" />-->
</form>
</form>
</body>
</html>
`
