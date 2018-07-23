package iniflags

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

var (
	allowUnknownFlags    = flag.Bool("allowUnknownFlags", false, "Don't terminate the app if ini file contains unknown flags.")
	allowMissingConfig   = flag.Bool("allowMissingConfig", false, "Don't terminate the app if the ini file cannot be read.")
	config               = flag.String("config", "", "Path to ini config for using in go flags. May be relative to the current executable path.")
	configUpdateInterval = flag.Duration("configUpdateInterval", 0, "Update interval for re-reading config file set via -config flag. Zero disables config file re-reading.")
	dumpflags            = flag.Bool("dumpflags", false, "Dumps values for all flags defined in the app into stdout in ini-compatible syntax and terminates the app.")
)

var (
	flagChangeCallbacks = make(map[string][]FlagChangeCallback)
	importStack         []string
	parsed              bool
)

// Generation is flags' generation number.
//
// It is modified on each flags' modification
// via either -configUpdateInterval or SIGHUP.
var Generation int

// Parse obtains flag values from config file set via -config.
//
// It obtains flag values from command line like flag.Parse(), then overrides
// them by values parsed from config file set via -config.
//
// Path to config file can also be set via SetConfigFile() before Parse() call.
func Parse() {
	if parsed {
		logger.Panicf("iniflags: duplicate call to iniflags.Parse() detected")
	}

	parsed = true
	flag.Parse()
	_, ok := parseConfigFlags()
	if !ok {
		os.Exit(1)
	}

	if *dumpflags {
		dumpFlags()
		os.Exit(0)
	}

	for flagName := range flagChangeCallbacks {
		verifyFlagChangeFlagName(flagName)
	}
	Generation++
	issueAllFlagChangeCallbacks()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGHUP)
	go sighupHandler(ch)

	go configUpdater()
}

func configUpdater() {
	if *configUpdateInterval != 0 {
		for {
			// Use time.Sleep() instead of time.Tick() for the sake of dynamic flag update.
			time.Sleep(*configUpdateInterval)
			UpdateConfig()
		}
	}
}

func UpdateConfig() {
	if oldFlagValues, ok := parseConfigFlags(); ok && len(oldFlagValues) > 0 {
		modifiedFlags := make(map[string]string)
		for k := range oldFlagValues {
			modifiedFlags[k] = flag.Lookup(k).Value.String()
		}
		logger.Printf("iniflags: read updated config. Modified flags are: %v", modifiedFlags)
		Generation++
		issueFlagChangeCallbacks(oldFlagValues)
	}
}

// FlagChangeCallback is called when the given flag is changed.
//
// The callback may be registered for any flag via OnFlagChange().
type FlagChangeCallback func()

// OnFlagChange registers the callback, which is called after the given flag
// value is initialized and/or changed.
//
// Flag values are initialized during iniflags.Parse() call.
// Flag value can be changed on config re-read after obtaining SIGHUP signal
// or if periodic config re-read is enabled with -configUpdateInterval flag.
//
// Note that flags set via command-line cannot be overriden via config file modifications.
func OnFlagChange(flagName string, callback FlagChangeCallback) {
	if parsed {
		verifyFlagChangeFlagName(flagName)
	}
	flagChangeCallbacks[flagName] = append(flagChangeCallbacks[flagName], callback)
}

func verifyFlagChangeFlagName(flagName string) {
	if flag.Lookup(flagName) == nil {
		logger.Fatalf("iniflags: cannot register FlagChangeCallback for non-existing flag [%s]", flagName)
	}
}

func issueFlagChangeCallbacks(oldFlagValues map[string]string) {
	for flagName := range oldFlagValues {
		if fs, ok := flagChangeCallbacks[flagName]; ok {
			for _, f := range fs {
				f()
			}
		}
	}
}

func issueAllFlagChangeCallbacks() {
	for _, fs := range flagChangeCallbacks {
		for _, f := range fs {
			f()
		}
	}
}

func sighupHandler(ch <-chan os.Signal) {
	for _ = range ch {
		UpdateConfig()
	}
}

func parseConfigFlags() (oldFlagValues map[string]string, ok bool) {
	configPath := *config
	if !strings.HasPrefix(configPath, "./") {
		if configPath, ok = combinePath(os.Args[0], *config); !ok {
			return nil, false
		}
	}
	if configPath == "" {
		return nil, true
	}
	parsedArgs, ok := getArgsFromConfig(configPath)
	if !ok {
		return nil, false
	}
	missingFlags := getMissingFlags()

	ok = true
	oldFlagValues = make(map[string]string)
	for _, arg := range parsedArgs {
		f := flag.Lookup(arg.Key)
		if f == nil {
			logger.Printf("iniflags: unknown flag name=[%s] found at line [%d] of file [%s]", arg.Key, arg.LineNum, arg.FilePath)
			if !*allowUnknownFlags {
				ok = false
			}
			continue
		}

		if _, found := missingFlags[f.Name]; found {
			oldValue := f.Value.String()
			if oldValue == arg.Value {
				continue
			}
			if err := f.Value.Set(arg.Value); err != nil {
				logger.Printf("iniflags: error when parsing flag [%s] value [%s] at line [%d] of file [%s]: [%s]", arg.Key, arg.Value, arg.LineNum, arg.FilePath, err)
				ok = false
				continue
			}
			if oldValue != f.Value.String() {
				oldFlagValues[arg.Key] = oldValue
			}
		}
	}

	if !ok {
		// restore old flag values
		for k, v := range oldFlagValues {
			flag.Set(k, v)
		}
		oldFlagValues = nil
	}

	return oldFlagValues, ok
}

func checkImportRecursion(configPath string) bool {
	for _, path := range importStack {
		if path == configPath {
			logger.Printf("iniflags: import recursion found for [%s]: %v", configPath, importStack)
			return false
		}
	}
	return true
}

type flagArg struct {
	Key      string
	Value    string
	FilePath string
	LineNum  int
	Comment  string
}

func stripBOM(s string) string {
	if len(s) < 3 {
		return s
	}
	bom := s[:3]
	if bom == "\ufeff" || bom == "\ufffe" {
		return s[3:]
	}
	return s
}

func ReadIniFile(iniFilePath string) (args []flagArg, ok bool) {
	return getArgsFromConfig(iniFilePath)
}

func getArgsFromConfig(configPath string) (args []flagArg, ok bool) {
	if !checkImportRecursion(configPath) {
		return nil, false
	}
	importStack = append(importStack, configPath)
	defer func() {
		importStack = importStack[:len(importStack)-1]
	}()

	file, err := openConfigFile(configPath)
	if err != nil {
		return nil, *allowMissingConfig
	}
	defer file.Close()
	r := bufio.NewReader(file)

	var lineNum int
	var comment = ""
	var multilineFA flagArg
	for {
		lineNum++
		line, err := r.ReadString('\n')
		if err != nil && line == "" {
			if err == io.EOF {
				if len(multilineFA.Key) > 0 {
					// flush the last multiline arg
					args = append(args, multilineFA)
				}
				break
			}
			logger.Printf("iniflags: error when reading file [%s] at line %d: [%s]", configPath, lineNum, err)
			return nil, false
		}
		if lineNum == 1 {
			line = stripBOM(line)
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#import ") {
			importPath, _, ok := unquoteValue(line[7:], lineNum, configPath)
			if !ok {
				return nil, false
			}
			if importPath, ok = combinePath(configPath, importPath); !ok {
				return nil, false
			}
			importArgs, ok := getArgsFromConfig(importPath)
			if !ok {
				return nil, false
			}
			args = append(args, importArgs...)
			continue
		}
		if line == "" || line[0] == '[' {
			comment = ""
			continue
		}
		if line[0] == '#' || line[0] == ';' {
			//save the comment and move to the next line
			comment = line[1:]
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			logger.Printf("iniflags: cannot split [%s] at line %d into key and value in config file [%s]", line, lineNum, configPath)
			return nil, false
		}
		key := strings.TrimSpace(parts[0])

		value, cmt, ok := unquoteValue(parts[1], lineNum, configPath)
		if !ok {
			return nil, false
		}
		if comment == "" {
			comment = cmt
		}

		fa := flagArg{
			Key:      key,
			Value:    value,
			FilePath: configPath,
			LineNum:  lineNum,
			Comment:  comment,
		}

		comment = ""
		if !strings.HasSuffix(key, "}") {
			if len(multilineFA.Key) > 0 {
				// flush the last multiline arg
				args = append(args, multilineFA)
				multilineFA = flagArg{}
			}

			args = append(args, fa)
			continue
		}

		// multiline arg
		n := strings.LastIndex(key, "{")
		if n < 0 {
			log.Printf("iniflags: cannot find '{' in the multiline key [%s] at line %d, file [%s]", key, lineNum, configPath)
			return nil, false
		}
		switch multilineFA.Key {
		case "":
			// the first line for multiline arg
			multilineFA = fa
			multilineFA.Key = key[:n]
		case key[:n]:
			// the subsequent line for multiline arg
			delimiter := key[n+1 : len(key)-1]
			multilineFA.Value += delimiter
			multilineFA.Value += value
		default:
			// new multiline arg
			args = append(args, multilineFA)
			multilineFA = fa
			multilineFA.Key = key[:n]
		}
	}

	return args, true
}

func openConfigFile(path string) (io.ReadCloser, error) {
	if isHTTP(path) {
		resp, err := http.Get(path)
		if err != nil {
			logger.Printf("iniflags: cannot load config file at [%s]: [%s]\n", path, err)
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			logger.Printf("iniflags: unexpected http status code when obtaining config file [%s]: %d. Expected %d", path, resp.StatusCode, http.StatusOK)
			return nil, err
		}
		return resp.Body, nil
	}

	file, err := os.Open(path)
	if err != nil {
		if !(*allowMissingConfig) {
			logger.Printf("iniflags: cannot open config file at [%s]: [%s]", path, err)
		}
		return nil, err
	}
	return file, nil
}

func combinePath(basePath, relPath string) (string, bool) {
	if isHTTP(basePath) {
		base, err := url.Parse(basePath)
		if err != nil {
			logger.Printf("iniflags: error when parsing http base path [%s]: %s", basePath, err)
			return "", false
		}
		rel, err := url.Parse(relPath)
		if err != nil {
			logger.Printf("iniflags: error when parsing http rel path [%s] for base [%s]: %s", relPath, basePath, err)
			return "", false
		}
		return base.ResolveReference(rel).String(), true
	}

	if relPath == "" || relPath[0] == '/' || isHTTP(relPath) {
		return relPath, true
	}
	return path.Join(path.Dir(basePath), relPath), true
}

func isHTTP(path string) bool {
	return strings.HasPrefix(strings.ToLower(path), "http://") || strings.HasPrefix(strings.ToLower(path), "https://")
}

func getMissingFlags() map[string]bool {
	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	missingFlags := make(map[string]bool)
	flag.VisitAll(func(f *flag.Flag) {
		if _, ok := setFlags[f.Name]; !ok {
			missingFlags[f.Name] = true
		}
	})
	return missingFlags
}

func dumpFlags() {
	flag.VisitAll(func(f *flag.Flag) {
		if f.Name != "config" && f.Name != "dumpflags" {
			fmt.Printf("%s = %s  # %s\n", f.Name, quoteValue(f.Value.String()), escapeUsage(f.Usage))
		}
	})
}

func escapeUsage(s string) string {
	return strings.Replace(s, "\n", "\n    # ", -1)
}

func quoteValue(v string) string {
	if !strings.ContainsAny(v, "\n#;") && strings.TrimSpace(v) == v {
		return v
	}
	v = strings.Replace(v, "\\", "\\\\", -1)
	v = strings.Replace(v, "\n", "\\n", -1)
	v = strings.Replace(v, "\"", "\\\"", -1)
	return fmt.Sprintf("\"%s\"", v)
}

func unquoteValue(val string, lineNum int, configPath string) (string, string, bool) {
	v := strings.TrimSpace(val)
	if len(v) == 0 {
		return "", "", true
	}
	if v[0] != '"' {
		return removeTrailingComments(v), getTrailingComment(v), true
	}
	n := strings.LastIndex(v, "\"")
	if n == -1 {
		logger.Printf("iniflags: unclosed string found [%s] at line %d in config file [%s]", v, lineNum, configPath)
		return "", "", false
	}
	v = v[1:n]
	v = strings.Replace(v, "\\\"", "\"", -1)
	v = strings.Replace(v, "\\n", "\n", -1)
	v = strings.Replace(v, "\\\\", "\\", -1)

	//to get the comment remove the value from the original value and get the trailing comment
	comment := getTrailingComment(strings.Replace(val, fmt.Sprintf("%q", v), "", 1))

	return v, comment, true
}

func removeTrailingComments(v string) string {
	v = strings.Split(v, "#")[0]
	v = strings.Split(v, ";")[0]
	return strings.TrimSpace(v)
}

func getTrailingComment(v string) string {
	if len(v) == 0 {
		return ""
	}
	if v[0] == '"' {
		return ""
	}
	s := strings.Split(v, "#")
	if len(s) > 1 {
		return s[1]
	}
	s = strings.Split(v, ";")
	if len(s) > 1 {
		return s[1]
	}
	return ""
}

// SetConfigFile sets path to config file.
//
// Call this function before Parse() if you need default path to config file
// when -config command-line flag is not set.
func SetConfigFile(path string) {
	if parsed {
		logger.Panicf("iniflags: SetConfigFile() must be called before Parse()")
	}
	*config = path
}

func SetAllowMissingConfigFile(allowed bool) {
	if parsed {
		panic("iniflags: SetAllowMissingConfigFile() must be called before Parse()")
	}
	*allowMissingConfig = allowed
}

func SetAllowUnknownFlags(allowed bool) {
	if parsed {
		logger.Panicf("iniflags: SetAllowUnknownFlags() must be called before Parse()")
	}
	*allowUnknownFlags = allowed
}

func SetConfigUpdateInterval(interval time.Duration) {
	if parsed {
		logger.Panicf("iniflags: SetConfigUpdateInterval() must be called before Parse()")
	}
	*configUpdateInterval = interval
}

// Logger is a slimmed-down version of the log.Logger interface, which only includes the methods we use.
// This interface is accepted by SetLogger() to redirect log output to another destination.
type Logger interface {
	Printf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Panicf(format string, v ...interface{})
}

// logger is the global Logger used to output log messages.  By default, it outputs to the same place and with the same
// format as the standard libary log package calls.  It can be changed via SetLogger().
var logger Logger = log.New(os.Stderr, "", log.LstdFlags)

func SetLogger(l Logger) {
	logger = l
}
