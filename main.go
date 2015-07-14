// Version 1.82
// Supports Windows, Linux, Mac, and Raspberry Pi, Beagle Bone Black

package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	//"net/http/pprof"
	"github.com/kardianos/osext"
	//"github.com/sanbornm/go-selfupdate/selfupdate" #included in update.go to change heavily
	//"github.com/sanderhahn/gozip"
	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	"github.com/kardianos/service"
	"github.com/vharitonsky/iniflags"
	"runtime/debug"
	"text/template"
	"time"
)

var (
	version      = "1.83"
	versionFloat = float32(1.83)
	addr         = flag.String("addr", ":8989", "http service address")
	addrSSL      = flag.String("addrSSL", ":8990", "https service address")
	//assets       = flag.String("assets", defaultAssetPath(), "path to assets")
	verbose = flag.Bool("v", true, "show debug logging")
	//verbose = flag.Bool("v", false, "show debug logging")
	//homeTempl *template.Template
	isLaunchSelf = flag.Bool("ls", false, "launch self 5 seconds later")

	configIni = flag.String("configFile", "config.ini", "config file path")
	// regular expression to sort the serial port list
	// typically this wouldn't be provided, but if the user wants to clean
	// up their list with a regexp so it's cleaner inside their end-user interface
	// such as ChiliPeppr, this can make the massive list that Linux gives back
	// to you be a bit more manageable
	regExpFilter = flag.String("regex", "usb|acm|com", "Regular expression to filter serial port list")

	// allow garbageCollection()
	//isGC = flag.Bool("gc", false, "Is garbage collection on? Off by default.")
	//isGC = flag.Bool("gc", true, "Is garbage collection on? Off by default.")
	gcType = flag.String("gc", "std", "Type of garbage collection. std = Normal garbage collection allowing system to decide (this has been known to cause a stop the world in the middle of a CNC job which can cause lost responses from the CNC controller and thus stalled jobs. use max instead to solve.), off = let memory grow unbounded (you have to send in the gc command manually to garbage collect or you will run out of RAM eventually), max = Force garbage collection on each recv or send on a serial port (this minimizes stop the world events and thus lost serial responses, but increases CPU usage)")

	// whether to do buffer flow debugging
	bufFlowDebugType = flag.String("bufflowdebug", "off", "off = (default) We do not send back any debug JSON, on = We will send back a JSON response with debug info based on the configuration of the buffer flow that the user picked")

	// hostname. allow user to override, otherwise we look it up
	hostname = flag.String("hostname", "unknown-hostname", "Override the hostname we get from the OS")

	updateUrl = flag.String("updateUrl", "", "")
	appName   = flag.String("appName", "", "")
)

var globalConfigMap map[string]interface{}

type NullWriter int

func (NullWriter) Write([]byte) (int, error) { return 0, nil }

func defaultAssetPath() string {
	//p, err := build.Default.Import("gary.burd.info/go-websocket-chat", "", build.FindOnly)
	p, err := build.Default.Import("github.com/johnlauer/serial-port-json-server", "", build.FindOnly)
	if err != nil {
		return "."
	}
	return p.Dir
}

func homeHandler(c *gin.Context) {
	homeTemplate.Execute(c.Writer, c.Request.Host)
}

func launchSelfLater() {
	log.Println("Going to launch myself 5 seconds later.")
	time.Sleep(2 * 1000 * time.Millisecond)
	log.Println("Done waiting 5 secs. Now launching...")
}

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() {
	startDaemon()
}
func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	<-time.After(time.Second * 13)
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "ArduinoCreateBridge",
		DisplayName: "Arduino Create Bridge",
		Description: "A bridge that allows Arduino Create to operate on the boards connected to the computer",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		err = service.Control(s, os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Install()
	if err != nil {
		logger.Error(err)
	}

	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}

func startDaemon() {
	// setupSysTray()
	go func() {

		// autoextract self
		src, _ := osext.Executable()
		dest := filepath.Dir(src)

		// save the config.ini (if it exists)
		if _, err := os.Stat(dest + "/" + *configIni); os.IsNotExist(err) {
			fmt.Println("First run, unzipping self")
			err := Unzip(src, dest)
			fmt.Println("Self extraction, err:", err)
		}

		if _, err := os.Stat(dest + "/" + *configIni); os.IsNotExist(err) {
			flag.Parse()
			fmt.Println("No config.ini at", *configIni)
		} else {
			flag.Set("config", dest+"/"+*configIni)
			iniflags.Parse()
		}

		// setup logging
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

		// see if we are supposed to wait 5 seconds
		if *isLaunchSelf {
			launchSelfLater()
		}

		var updater = &Updater{
			CurrentVersion: version,
			ApiURL:         *updateUrl,
			BinURL:         *updateUrl,
			DiffURL:        "",
			Dir:            "update/",
			CmdName:        *appName,
		}

		if updater != nil {
			go updater.BackgroundRun()
		}

		// data, err := Asset("arduino.zip")
		// if err != nil {
		// 	log.Println("arduino tools not found")
		// }

		createGlobalConfigMap(&globalConfigMap)

		//getList()
		f := flag.Lookup("addr")
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

		ip := "0.0.0.0"
		log.Print("Starting server and websocket on " + ip + "" + f.Value.String())
		//homeTempl = template.Must(template.ParseFiles(filepath.Join(*assets, "home.html")))

		log.Println("The Serial Port JSON Server is now running.")
		log.Println("If you are using ChiliPeppr, you may go back to it and connect to this server.")

		// see if they provided a regex filter
		if len(*regExpFilter) > 0 {
			log.Printf("You specified a serial port regular expression filter: %v\n", *regExpFilter)
		}

		// list serial ports
		portList, _ := GetList(false)
		/*if errSys != nil {
			log.Printf("Got system error trying to retrieve serial port list. Err:%v\n", errSys)
			log.Fatal("Exiting")
		}*/
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

		r.Use(cors.Middleware(cors.Config{
			Origins:         "https://create.arduino.cc, http://create.arduino.cc, https://create-dev.arduino.cc, http://create-dev.arduino.cc, http://webide.arduino.cc:8080",
			Methods:         "GET, PUT, POST, DELETE",
			RequestHeaders:  "Origin, Authorization, Content-Type",
			ExposedHeaders:  "",
			MaxAge:          50 * time.Second,
			Credentials:     true,
			ValidateHeaders: false,
		}))

		r.GET("/", homeHandler)
		r.POST("/upload", uploadHandler)
		r.GET("/socket.io/", socketHandler)
		r.POST("/socket.io/", socketHandler)
		r.Handle("WS", "/socket.io/", socketHandler)
		r.Handle("WSS", "/socket.io/", socketHandler)
		go func() {
			if err := r.RunTLS(*addrSSL, filepath.Join(dest, "cert.pem"), filepath.Join(dest, "key.pem")); err != nil {
				fmt.Printf("Error trying to bind to port: %v, so exiting...", err)
				log.Fatal("Error ListenAndServe:", err)
			}
		}()

		if err := r.Run(*addr); err != nil {
			fmt.Printf("Error trying to bind to port: %v, so exiting...", err)
			log.Fatal("Error ListenAndServe:", err)
		}
	}()

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
        if (!socket) {
            return false;
        }
        if (!msg.val()) {
            return false;
        }
        socket.emit("command", msg.val());
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
            appendLog($("<div/>").text(evt))
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
</form>
</body>
</html>
`
