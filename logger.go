package loggerfactory

import (
	"os"
	"runtime"
	"strings"

	"github.com/alecthomas/log4go"
)

const (
	PanicLevel int = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

var loggers = make(map[string]log4go.Logger)
var defaultLogger log4go.Logger

var initConfiguration = false
var defaultConfFilePath = "log4go/log4go.xml"
var defaultConfFilePath2 = "log4go.xml"
var defaultLoggerName = "default"

func init() {
	if !initConfiguration {
		if PathExists(defaultConfFilePath) {
			println("log4go init with log4go/log4go.xml")
			LoadConfiguration(defaultConfFilePath)
		} else if PathExists(defaultConfFilePath2) {
			println("log4go init with log4go.xml")
			LoadConfiguration(defaultConfFilePath2)
		} else {
			println("log4go init with default. use console output and debug level")
			InitDefault()
		}
		initConfiguration = true
	}
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func GetLogger() log4go.Logger {
	_, file, _, ok := runtime.Caller(1)
	if ok {
		for path, logger := range loggers {
			if strings.Index(file, path) > -1 {
				return logger
			}
		}
	}
	return defaultLogger
}
