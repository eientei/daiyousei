package logger

import (
	"log"
	"os"
)

const (
	LOG_DEBUG = 1<<0
	LOG_INFO  = 1<<1
	LOG_WARN  = 1<<2
	LOG_ERROR = 1<<3
	LOG_ALL   = LOG_DEBUG | LOG_INFO | LOG_WARN | LOG_ERROR;
)

var (
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	logLevel    int
	logFlags    int
)

func init() {
	setupLoggers()
	Flags(log.LstdFlags)
}

func setupLoggers() {
	debugLogger = log.New(os.Stdout, "[DEBUG] ", logFlags)
	infoLogger  = log.New(os.Stdout, " [INFO] ", logFlags)
	warnLogger  = log.New(os.Stdout, " [WARN] ", logFlags)
	errorLogger = log.New(os.Stdout, "[ERROR] ", logFlags)
}

func Level(level int) {
	logLevel = level
}

func Flags(flags int) {
	logFlags = flags
	setupLoggers()
}

func Debug(args ...interface{}) {
	if logLevel & LOG_DEBUG != 0 {
		debugLogger.Println(args...)
	}
}

func Debugf(fmt string, args ...interface{}) {
	if logLevel & LOG_DEBUG != 0 {
		debugLogger.Printf(fmt, args...)
	}
}


func Info(args ...interface{}) {
	if logLevel & LOG_INFO != 0 {
		infoLogger.Println(args...)
	}
}

func Infof(fmt string, args ...interface{}) {
	if logLevel & LOG_INFO != 0 {
		infoLogger.Printf(fmt, args...)
	}
}

func Warn(args ...interface{}) {
	if logLevel & LOG_WARN != 0 {
		warnLogger.Println(args...)
	}
}

func Warnf(fmt string, args ...interface{}) {
	if logLevel & LOG_WARN != 0 {
		warnLogger.Printf(fmt, args...)
	}
}

func Error(args ...interface{}) {
	if logLevel & LOG_ERROR != 0 {
		errorLogger.Println(args...)
	}
}

func Errorf(fmt string, args ...interface{}) {
	if logLevel & LOG_ERROR != 0 {
		errorLogger.Printf(fmt, args...)
	}
}