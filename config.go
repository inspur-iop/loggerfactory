package loggerfactory

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/log4go"
)

type xmlLogger struct {
	Path     string          `xml:"path"`
	Target   []string        `xml:"target"`
	Level    string          `xml:"level"`
}

type xmlProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type xmlFilter struct {
	Enabled  string        `xml:"enabled,attr"`
	Tag      string        `xml:"tag"`
	Type     string        `xml:"type"`
	Property []xmlProperty `xml:"property"`
}

type xmlLoggerConfig struct {
	Logger []xmlLogger `xml:"logger"`
	Filter []xmlFilter `xml:"filter"`
}

func InitDefault() {
	format := "[%D %T] [%L] (%S) %M"
	clw := log4go.NewConsoleLogWriter()
	clw.SetFormat(format)
	logger := make(log4go.Logger)
	logger.AddFilter("default",  log4go.DEBUG, clw)
	defaultLogger = logger
}

// Load XML configuration; see examples/example.xml for documentation
func LoadConfiguration(filename string) {
	// Open the configuration file
	fd, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Could not open %q for reading: %s\n", filename, err)
		os.Exit(1)
	}

	contents, err := ioutil.ReadAll(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Could not read %q: %s\n", filename, err)
		os.Exit(1)
	}

	xc := new(xmlLoggerConfig)
	if err := xml.Unmarshal(contents, xc); err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Could not parse XML configuration in %q: %s\n", filename, err)
		os.Exit(1)
	}
	bad := false
	targetWriters := make(map[string]log4go.LogWriter)
	for _, xmlFilter := range xc.Filter {
		good, enabled := true, false

		// Check required children
		if len(xmlFilter.Enabled) == 0 {
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required attribute %s for filter missing in %s\n", "enabled", filename)
			bad = true
		} else {
			enabled = xmlFilter.Enabled != "false"
		}
		if len(xmlFilter.Tag) == 0 {
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required child <%s> for filter missing in %s\n", "tag", filename)
			bad = true
		}
		if len(xmlFilter.Type) == 0 {
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required child <%s> for filter missing in %s\n", "type", filename)
			bad = true
		}

		var writer log4go.LogWriter
		switch xmlFilter.Type {
		case "console":
			writer, good = xmlToConsoleLogWriter(filename, xmlFilter.Property, enabled)
		case "file":
			writer, good = xmlToFileLogWriter(filename, xmlFilter.Property, enabled)
		default:
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Could not load XML configuration in %s: unknown filter type \"%s\"\n", filename, xmlFilter.Type)
			os.Exit(1)
		}
		// Just so all of the required params are errored at the same time if wrong
		if !good {
			os.Exit(1)
		}
		// If we're disabled (syntax and correctness checks only), don't add to logger
		if !enabled {
			continue
		}
		targetWriters[xmlFilter.Tag] = writer
	}

	for _, xmlLogger := range xc.Logger {
		logger := make(log4go.Logger)
		var lvl log4go.Level
		if len(xmlLogger.Level) == 0 {
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required child <%s> for logger missing in %s\n", "level", filename)
			bad = true
		}
		switch xmlLogger.Level {
		case "FINEST":
			lvl = log4go.FINEST
		case "FINE":
			lvl = log4go.FINE
		case "DEBUG":
			lvl = log4go.DEBUG
		case "TRACE":
			lvl = log4go.TRACE
		case "INFO":
			lvl = log4go.INFO
		case "WARNING":
			lvl = log4go.WARNING
		case "ERROR":
			lvl = log4go.ERROR
		case "CRITICAL":
			lvl = log4go.CRITICAL
		default:
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required child <%s> for filter has unknown value in %s: %s\n", "level", filename, xmlLogger.Level)
			bad = true
		}
		if len(xmlLogger.Path) == 0 {
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required child <%s> for logger missing in %s\n", "path", filename)
			bad = true
		}
		if len(xmlLogger.Target) == 0 {
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required child <%s> for logger missing in %s\n", "target", filename)
			bad = true
		}
		for _, target := range xmlLogger.Target {
			if len(target) == 0 {
				fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required child <%s> for logger missing in %s\n", "target", filename)
				bad = true
			}
			if targetWriters[target] == nil {
				fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required target filter <%s> for logger missing in %s\n", target, filename)
				bad = true
			}
			logger.AddFilter(target, lvl, targetWriters[target])
		}
		loggers[xmlLogger.Path] = logger
		if xmlLogger.Path == defaultLoggerName {
			defaultLogger = logger
		}
	}
	if defaultLogger == nil {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required default logger missing in %s\n", filename)
		bad = true
	}

	// Just so all of the required attributes are errored at the same time if missing
	if bad {
		os.Exit(1)
	}
}


func xmlToConsoleLogWriter(filename string, props []xmlProperty, enabled bool) (*log4go.ConsoleLogWriter, bool) {

	format := "[%D %T] [%L] (%S) %M"

	// Parse properties
	for _, prop := range props {
		switch prop.Name {
		case "format":
			format = strings.Trim(prop.Value, " \r\n")
		default:
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Warning: Unknown property \"%s\" for console filter in %s\n", prop.Name, filename)
		}
	}

	// If it's disabled, we're just checking syntax
	if !enabled {
		return nil, true
	}

	clw := log4go.NewConsoleLogWriter()
	clw.SetFormat(format)

	return clw, true
}

func xmlToFileLogWriter(filename string, props []xmlProperty, enabled bool) (*log4go.FileLogWriter, bool) {
	file := ""
	format := "[%D %T] [%L] (%S) %M"
	maxlines := 0
	maxsize := 0
	daily := false
	rotate := false

	// Parse properties
	for _, prop := range props {
		switch prop.Name {
		case "filename":
			file = strings.Trim(prop.Value, " \r\n")
		case "format":
			format = strings.Trim(prop.Value, " \r\n")
		case "maxlines":
			maxlines = strToNumSuffix(strings.Trim(prop.Value, " \r\n"), 1000)
		case "maxsize":
			maxsize = strToNumSuffix(strings.Trim(prop.Value, " \r\n"), 1024)
		case "daily":
			daily = strings.Trim(prop.Value, " \r\n") != "false"
		case "rotate":
			rotate = strings.Trim(prop.Value, " \r\n") != "false"
		default:
			fmt.Fprintf(os.Stderr, "LoadConfiguration: Warning: Unknown property \"%s\" for file filter in %s\n", prop.Name, filename)
		}
	}

	// Check properties
	if len(file) == 0 {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Error: Required property \"%s\" for file filter missing in %s\n", "filename", filename)
		return nil, false
	}

	// If it's disabled, we're just checking syntax
	if !enabled {
		return nil, true
	}

	flw := log4go.NewFileLogWriter(file, rotate)
	flw.SetFormat(format)
	flw.SetRotateLines(maxlines)
	flw.SetRotateSize(maxsize)
	flw.SetRotateDaily(daily)
	return flw, true
}

// Parse a number with K/M/G suffixes based on thousands (1000) or 2^10 (1024)
func strToNumSuffix(str string, mult int) int {
	num := 1
	if len(str) > 1 {
		switch str[len(str)-1] {
		case 'G', 'g':
			num *= mult
			fallthrough
		case 'M', 'm':
			num *= mult
			fallthrough
		case 'K', 'k':
			num *= mult
			str = str[0 : len(str)-1]
		}
	}
	parsed, _ := strconv.Atoi(str)
	return parsed * num
}