package logger

import (
	"log"
	"os"
)

const (
	LOG_DEBUG   = 1<<0
	LOG_INFO    = 1<<1
	LOG_WARN    = 1<<2
	LOG_ERROR   = 1<<3
	LOG_ALL     = LOG_DEBUG | LOG_INFO | LOG_WARN | LOG_ERROR
	LOG_DEFAULT = LOG_INFO | LOG_WARN | LOG_ERROR
)

const (
	PREFIX_DEBUG = "[DEBUG] "
	PREFIX_INFO  = " [INFO] "
	PREFIX_WARN  = " [WARN] "
	PREFIX_ERROR = "[ERROR] "
)

var (
	level  int         = LOG_DEFAULT
	flags  int         = log.LstdFlags
	logger *log.Logger = log.New(os.Stdout, "", flags);
)

func Level(newlevel int) {
	level = newlevel
}

func Flags(newflags int) {
	flags = newflags
	logger = log.New(os.Stdout, "", flags);
}

func Debug(args ...interface{}) {
	if level & LOG_DEBUG != 0 {
		logger.SetPrefix(PREFIX_DEBUG)
		logger.Println(args...)
	}
}

func Debugf(fmt string, args ...interface{}) {
	if level & LOG_DEBUG != 0 {
		logger.SetPrefix(PREFIX_DEBUG)
		logger.Printf(fmt, args...)
	}
}

func Info(args ...interface{}) {
	if level & LOG_INFO != 0 {
		logger.SetPrefix(PREFIX_INFO)
		logger.Println(args...)
	}
}

func Infof(fmt string, args ...interface{}) {
	if level & LOG_INFO != 0 {
		logger.SetPrefix(PREFIX_INFO)
		logger.Printf(fmt, args...)
	}
}

func Warn(args ...interface{}) {
	if level & LOG_WARN != 0 {
		logger.SetPrefix(PREFIX_WARN)
		logger.Println(args...)
	}
}

func Warnf(fmt string, args ...interface{}) {
	if level & LOG_WARN != 0 {
		logger.SetPrefix(PREFIX_WARN)
		logger.Printf(fmt, args...)
	}
}

func Error(args ...interface{}) {
	if level & LOG_ERROR != 0 {
		logger.SetPrefix(PREFIX_ERROR)
		logger.Println(args...)
	}
}

func Errorf(fmt string, args ...interface{}) {
	if level & LOG_ERROR != 0 {
		logger.SetPrefix(PREFIX_ERROR)
		logger.Printf(fmt, args...)
	}
}