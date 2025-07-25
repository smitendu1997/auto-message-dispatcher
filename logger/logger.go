// Package logger contains logger related functions that are used in different packages
package logger

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// InitLogger sets up
func InitLogger() {
	// Set application-wide logger
	// Set log level based on environment variable or config
	level := zerolog.InfoLevel // default
	// Example: read from env "LOG_LEVEL"
	if lvlStr := viper.GetString("LOG_LEVEL"); lvlStr != "" {
		switch lvlStr {
		case "trace":
			level = zerolog.TraceLevel
		case "debug":
			level = zerolog.DebugLevel
		case "info":
			level = zerolog.InfoLevel
		case "warn":
			level = zerolog.WarnLevel
		case "error":
			level = zerolog.ErrorLevel
		case "fatal":
			level = zerolog.FatalLevel
		case "panic":
			level = zerolog.PanicLevel
		}
	}
	zerolog.SetGlobalLevel(level)
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// Level ... log level
type Level int

const (
	// TRACE ... trace log level
	TRACE Level = -1
	// DEBUG ... debug log level
	DEBUG Level = iota
	// INFO ... info log level
	INFO
	// WARN ... warn log level
	WARN
	// ERROR ... error log level
	ERROR
	// FATAL ... fatal log level
	FATAL
	// PANIC ... panic log level
	PANIC
)

// var level Level = DEBUG

// SetLevel ... sets the level of the logger
func SetLevel(l Level) {
	fmt.Print("SetLevel:", l)
	switch l {
	case PANIC:
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case FATAL:
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case WARN:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case DEBUG:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case INFO:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case ERROR:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case TRACE:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	// level = l
}

// debug logs a debug message
func debug(msg interface{}) {
	log.Debug().Msgf("%v", msg)
}

// info logs an info message
func info(msg interface{}) {
	log.Info().Msgf("%v", msg)
}

// warn logs a warning message
func warn(msg interface{}) {
	log.Warn().Msgf("%v", msg)
}

// errormsg logs an error message
func errormsg(msg interface{}) {
	log.Error().Msgf("%v", msg)
}

// fatal logs a fatal message
func fatal(msg interface{}) {
	log.Fatal().Msgf("%v", msg)
}

// panic logs a panic message
func panic(msg interface{}) {
	log.Panic().Msgf("%v", msg)
}

// trace logs a trace message
func trace(msg interface{}) {
	log.Trace().Msgf("%v", msg)
}

// D ... debug logs
func Debug(i ...interface{}) {
	debug(i)
}

// I ... Info logs
func Info(i ...interface{}) {
	info(i)
}

// W ... Warn logs
func Warn(i ...interface{}) {
	warn(i)
}

// E ... Error logs
func Error(i ...interface{}) {
	errormsg(i)
}

// F ... Fatal logs
func Fatal(i ...interface{}) {
	fatal(i)
}

// P ... Panic logs
func Panic(i ...interface{}) {
	panic(i)
}

// T ... Trace logs
func Trace(i ...interface{}) {
	trace(i)
}
