// Version 1.82
// Supports Windows, Linux, Mac, and Raspberry Pi, Beagle Bone Black

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/arduino/arduino-create-agent/systray"
	"github.com/arduino/arduino-create-agent/tools"
	"github.com/arduino/arduino-create-agent/updater"
	"github.com/arduino/arduino-create-agent/utilities"
	v2 "github.com/arduino/arduino-create-agent/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	cors "github.com/itsjamie/gin-cors"
	"github.com/kardianos/osext"
	log "github.com/sirupsen/logrus"
	//"github.com/sanbornm/go-selfupdate/selfupdate" #included in update.go to change heavily
)

var (
	version               = "x.x.x-dev" //don't modify it, Jenkins will take care
	git_revision          = "xxxxxxxx"  //don't modify it, Jenkins will take care
	embedded_autoextract  = false
	port                  string
	portSSL               string
	requiredToolsAPILevel = "v1"
)

// regular flags
var (
	hibernate        = flag.Bool("hibernate", false, "start hibernated")
	genCert          = flag.Bool("generateCert", false, "")
	additionalConfig = flag.String("additional-config", "config.ini", "config file path")
	isLaunchSelf     = flag.Bool("ls", false, "launch self 5 seconds later")

	// Ignored flags for compatibility
	_ = flag.String("gc", "std", "Deprecated. Use the config.ini file")
	_ = flag.String("regex", "usb|acm|com", "Deprecated. Use the config.ini file")
)

// iniflags
var (
	address      = iniConf.String("address", "127.0.0.1", "The address where to listen. Defaults to localhost")
	appName      = iniConf.String("appName", "", "")
	gcType       = iniConf.String("gc", "std", "Type of garbage collection. std = Normal garbage collection allowing system to decide (this has been known to cause a stop the world in the middle of a CNC job which can cause lost responses from the CNC controller and thus stalled jobs. use max instead to solve.), off = let memory grow unbounded (you have to send in the gc command manually to garbage collect or you will run out of RAM eventually), max = Force garbage collection on each recv or send on a serial port (this minimizes stop the world events and thus lost serial responses, but increases CPU usage)")
	hostname     = iniConf.String("hostname", "unknown-hostname", "Override the hostname we get from the OS")
	httpProxy    = iniConf.String("httpProxy", "", "Proxy server for HTTP requests")
	httpsProxy   = iniConf.String("httpsProxy", "", "Proxy server for HTTPS requests")
	indexURL     = iniConf.String("indexURL", "https://downloads.arduino.cc/packages/package_staging_index.json", "The address from where to download the index json containing the location of upload tools")
	iniConf      = flag.NewFlagSet("ini", flag.ContinueOnError)
	logDump      = iniConf.String("log", "off", "off = (default)")
	origins      = iniConf.String("origins", "", "Allowed origin list for CORS")
	regExpFilter = iniConf.String("regex", "usb|acm|com", "Regular expression to filter serial port list")
	signatureKey = iniConf.String("signatureKey", "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvc0yZr1yUSen7qmE3cxF\nIE12rCksDnqR+Hp7o0nGi9123eCSFcJ7CkIRC8F+8JMhgI3zNqn4cUEn47I3RKD1\nZChPUCMiJCvbLbloxfdJrUi7gcSgUXrlKQStOKF5Iz7xv1M4XOP3JtjXLGo3EnJ1\npFgdWTOyoSrA8/w1rck4c/ISXZSinVAggPxmLwVEAAln6Itj6giIZHKvA2fL2o8z\nCeK057Lu8X6u2CG8tRWSQzVoKIQw/PKK6CNXCAy8vo4EkXudRutnEYHEJlPkVgPn\n2qP06GI+I+9zKE37iqj0k1/wFaCVXHXIvn06YrmjQw6I0dDj/60Wvi500FuRVpn9\ntwIDAQAB\n-----END PUBLIC KEY-----", "Pem-encoded public key to verify signed commandlines")
	updateUrl    = iniConf.String("updateUrl", "", "")
	verbose      = iniConf.Bool("v", true, "show debug logging")
	crashreport  = iniConf.Bool("crashreport", false, "enable crashreport logging")
)

// global clients
var (
	Tools   tools.Tools
	Systray systray.Systray
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
	// prevents bad errors in OSX, such as '[NS...] is only safe to invoke on the main thread'.
	runtime.LockOSThread()

	// Parse regular flags
	flag.Parse()

	// Generate certificates
	if *genCert == true {
		generateCertificates()
		os.Exit(0)
	}

	// Launch main loop in a goroutine
	go loop()

	// SetupSystray is the main thread
	Systray = systray.Systray{
		Hibernate: *hibernate,
		Version:   version + "-" + git_revision,
		DebugURL: func() string {
			return "http://" + *address + port
		},
		AdditionalConfig: *additionalConfig,
	}

	path, err := osext.Executable()
	if err != nil {
		panic(err)
	}

	// If the executable is temporary, copy it to the full path, then restart
	if strings.Contains(path, "-temp") {
		err := copyExe(path, updater.BinPath(path))
		if err != nil {
			panic(err)
		}

		Systray.Restart()
	} else {
		// Otherwise copy to a path with -temp suffix
		err := copyExe(path, updater.TempPath(path))
		if err != nil {
			panic(err)
		}
	}

	Systray.Start()
}

func copyExe(from, to string) error {
	data, err := ioutil.ReadFile(from)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(to, data, 0755)
	if err != nil {
		return err
	}
	return nil
}

func loop() {
	if *hibernate {
		return
	}
	// autoextract self
	src, _ := osext.Executable()
	dest := filepath.Dir(src)

	if embedded_autoextract {
		// save the config.ini (if it exists)
		if _, err := os.Stat(filepath.Join(dest, "config.ini")); os.IsNotExist(err) {
			log.Println("First run, unzipping self")
			err := utilities.Unzip(src, dest)
			log.Println("Self extraction, err:", err)
		}
	}

	// Parse ini config
	args, err := parseIni(filepath.Join(dest, "config.ini"))
	if err != nil {
		panic(err)
	}
	err = iniConf.Parse(args)
	if err != nil {
		panic(err)
	}

	// Parse additional ini config
	args, err = parseIni(filepath.Join(dest, *additionalConfig))
	if err != nil {
		panic(err)
	}
	err = iniConf.Parse(args)
	if err != nil {
		panic(err)
	}

	// Instantiate Tools
	usr, _ := user.Current()
	directory := filepath.Join(usr.HomeDir, ".arduino-create")
	Tools = tools.Tools{
		Directory: directory,
		IndexURL:  *indexURL,
		Logger: func(msg string) {
			mapD := map[string]string{"DownloadStatus": "Pending", "Msg": msg}
			mapB, _ := json.Marshal(mapD)
			h.broadcastSys <- mapB
		},
	}
	Tools.Init(requiredToolsAPILevel)

	log.SetLevel(log.InfoLevel)

	log.SetOutput(os.Stdout)

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

	// If the httpProxy setting is set, use its value to override the
	// HTTP_PROXY environment variable. Setting this environment
	// variable ensures that all HTTP requests using net/http use this
	// proxy server.
	if *httpProxy != "" {
		log.Printf("Setting HTTP_PROXY variable to %v", *httpProxy)
		err := os.Setenv("HTTP_PROXY", *httpProxy)
		if err != nil {
			// The os.Setenv documentation doesn't specify how it can
			// fail, so I don't know how to handle this error
			// appropriately.
			panic(err)
		}
	}

	if *httpsProxy != "" {
		log.Printf("Setting HTTPS_PROXY variable to %v", *httpProxy)
		err := os.Setenv("HTTPS_PROXY", *httpProxy)
		if err != nil {
			// The os.Setenv documentation doesn't specify how it can
			// fail, so I don't know how to handle this error
			// appropriately.
			panic(err)
		}
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

	// save crashreport to file
	if *crashreport {
		logFilename := "crashreport_" + time.Now().Format("20060102150405") + ".log"
		currDir, err := osext.ExecutableFolder()
		if err != nil {
			panic(err)
		}
		// handle logs directory creation
		logsDir := filepath.Join(currDir, "logs")
		if _, err := os.Stat(logsDir); os.IsNotExist(err) {
			os.Mkdir(logsDir, 0700)
		}
		logFile, err := os.OpenFile(filepath.Join(logsDir, logFilename), os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0644)
		if err != nil {
			log.Print("Cannot create file used for crash-report")
		} else {
			redirectStderr(logFile)
		}
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

	extraOrigins := []string{
		"https://create.arduino.cc",
		"https://create-dev.arduino.cc", "https://create-intel.arduino.cc",
	}

	for i := 8990; i < 9001; i++ {
		port := strconv.Itoa(i)
		extraOrigins = append(extraOrigins, "http://localhost:"+port)
		extraOrigins = append(extraOrigins, "https://localhost:"+port)
		extraOrigins = append(extraOrigins, "http://127.0.0.1:"+port)
	}

	r.Use(cors.Middleware(cors.Config{
		Origins:         *origins + ", " + strings.Join(extraOrigins, ", "),
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

	// Mount goa handlers
	goa := v2.Server(directory)
	r.Any("/v2/*path", gin.WrapH(goa))

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
}

var homeTemplate = template.Must(template.New("home").Parse(homeTemplateHtml))

// If you navigate to this server's homepage, you'll get this HTML
// so you can directly interact with the serial port server
const homeTemplateHtml = `<!DOCTYPE html>
<html>
<head>
<title>Arduino Create Agent Debug Console</title>
<link href="https://fonts.googleapis.com/css?family=Open+Sans:400,600,700&display=swap" rel="stylesheet">
<link href="https://fonts.googleapis.com/css?family=Roboto+Mono:400,600,700&display=swap" rel="stylesheet">
<script type="text/javascript" src="https://ajax.googleapis.com/ajax/libs/jquery/1.4.2/jquery.min.js"></script>
<script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/socket.io/1.3.5/socket.io.min.js"></script>
<script type="text/javascript">
    $(function() {
	    var socket;
	    var input = $('#input');
	    var log = document.getElementById('log');
	    var autoscroll = document.getElementById('autoscroll');
	    var listenabled = document.getElementById('list');
	    var messages = [];
        var MESSAGES_MAX_COUNT = 2000;

	    function appendLog(msg) {
            var startsWithBracked = msg.indexOf('{') == 0;
            var startsWithList = msg.indexOf('list') == 0;

	        if (listenabled.checked || (typeof msg === 'string' && !startsWithBracked && !startsWithList)) {
	            messages.push(msg);
	            if (messages.length > MESSAGES_MAX_COUNT) {
	                messages.shift();
	            }
	            log.innerHTML = messages.join('<br>');
	            if (autoscroll.checked) {
	                log.scrollTop = log.scrollHeight - log.clientHeight;
	            }
	        }
	    }

	    $('#form').submit(function(e) {
	    	e.preventDefault();
	        if (!socket) {
	            return false;
	        }
	        if (!input.val()) {
	            return false;
	        }
	        socket.emit('command', input.val());
	    });

	    $('#export').click(function() {
	    	var link = document.createElement('a');
	    	link.setAttribute('download', 'agent-log.txt');
	    	var text = log.innerHTML.replace(/<br>/g, '\n');
	    	link.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(text));
    		link.click();
    	});

        $('#clear').click(function() {
            messages = [];
            log.innerHTML = '';
        });

	    if (window['WebSocket']) {
	        if (window.location.protocol === 'https:') {
	            socket = io('https://{{$}}')
	        } else {
	            socket = io('http://{{$}}');
	        }
	        socket.on('disconnect', function(evt) {
	            appendLog($('<div><b>Connection closed.</b></div>'))
	        });
	        socket.on('message', function(evt) {
	            appendLog(evt);
	        });
	    } else {
	        appendLog($('<div><b>Your browser does not support WebSockets.</b></div>'))
        }
        
        $("#input").focus();
	});
</script>
<style type="text/css">
html, body {
    overflow: hidden;
    height: 100%;
}

body {    
    margin: 0px;
    padding: 0px;
    background: #F8F9F9;
    font-size: 16px;
    font-family: "Open Sans", "Lucida Grande", Lucida, Verdana, sans-serif;
}

#container {
    display: flex;
    flex-direction: column;    
    height: 100%;
    width: 100%;
}

#log {
    flex-grow: 1;
    font-family: "Roboto Mono", "Courier", "Lucida Grande", Verdana, sans-serif;
    background-color: #DAE3E3;
    margin: 15px 15px 10px;
    padding: 8px 10px;
    overflow-y: auto;
}

#footer {    
    display: flex;    
    flex-wrap: wrap;
    align-items: flex-start;
    justify-content: space-between;
    margin: 0px 15px 0px;    
}

#form {    
    display: flex;
    flex-grow: 1;
    margin-bottom: 15px;
}

#input {
    flex-grow: 1;
}

#secondary-controls div {
    display: inline-block;        
    padding: 10px 15px;    
}

#autoscroll,
#list {
    vertical-align: middle;
    width: 20px;
    height: 20px;
}


#secondary-controls button {
    margin-bottom: 15px;
    vertical-align: top;
}

.button {
    background-color: #b5c8c9;
    border: 1px solid #b5c8c9;
    border-radius: 2px 2px 0 0;
    box-shadow: 0 4px #95a5a6;
    margin-bottom: 4px;
    color: #000;
    cursor: pointer;    
    font-size: 14px;
    letter-spacing: 1.28px;
    line-height: normal;
    outline: none;
    padding: 9px 18px;
    text-align: center;
    text-transform: uppercase;
    transition: box-shadow .1s ease-out, transform .1s ease-out;
}

.button:hover {
    box-shadow: 0 2px #95a5a6;    
    outline: none;
    transform: translateY(2px);
}

.button:active {
    box-shadow: none;    
    transform: translateY(4px);
}

.textfield {
    background-color: #dae3e3;
    width: auto;
    height: auto;    
    padding: 10px 8px;
    margin-left: 8px;
    vertical-align: top;
    border: none;
    font-family: "Open Sans", "Lucida Grande", Lucida, Verdana, sans-serif;
    font-size: 1em;
    outline: none;
}

</style>
</head>
    <body>
        <div id="container">
            <div id="log">This is some random text This is some random textThis is some random textThis is some random textThis is some random textThis is some random textThis is some random text<br />This is some random text<br />This is some random text<br /></div>
            <div id="footer">
                <form id="form">
                    <input type="submit" class="button" value="Send" />
                    <input type="text" id="input" class="textfield" />
                </form>
                <div id="secondary-controls">
                    <div>
                        <input name="pause" type="checkbox" checked id="autoscroll" />
                        <label for="autoscroll">Autoscroll</label>
                    </div>
                    <div>
                        <input name="list" type="checkbox" checked id="list" />
                        <label for="list">Enable&nbsp;List&nbsp;Command</label>
                    </div>
                    <button id="clear" class="button">Clear&nbsp;Log</button>
                    <button id="export" class="button">Export&nbsp;Log</button>
                </div>
            </div>
        </div>
    </body>
</html>

`

func parseIni(filename string) (args []string, err error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: false}, filename)
	if err != nil {
		return nil, err
	}

	for _, section := range cfg.Sections() {
		for key, val := range section.KeysHash() {
			// Ignore launchself
			if key == "ls" {
				continue
			} // Ignore configUpdateInterval
			if key == "configUpdateInterval" {
				continue
			} // Ignore name
			if key == "name" {
				continue
			}
			args = append(args, "-"+key+"="+val)
		}
	}

	return args, nil
}
