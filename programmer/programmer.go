package programmer

type logger interface {
	Info()
	Debug()
}

// Auth contains username and password used for a network upload
type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Extra contains some options used during the upload
type Extra struct {
	Use1200bpsTouch   bool   `json:"use_1200bps_touch"`
	WaitForUploadPort bool   `json:"wait_for_upload_port"`
	Network           bool   `json:"network"`
	Auth              Auth   `json:"auth"`
	Verbose           bool   `json:"verbose"`
	ParamsVerbose     string `json:"params_verbose"`
	ParamsQuiet       string `json:"params_quiet"`
}

// Do performs a command on a port with a board attached to it
func Do(port, board, path, commandline string, extra Extra, l logger) {

}
