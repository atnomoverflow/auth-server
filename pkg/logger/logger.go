package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type ILogger interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

type Logger struct {
	logger *zerolog.Logger
}

var _ ILogger = (*Logger)(nil)

type LogLevel int

const (
	// it bugs if the intial value is 0
	DEBUG LogLevel = iota + 1
	INFO
	WARN
	ERROR
	FATAL
)

func (lvl LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[lvl-1]
}

func (lvl *LogLevel) UnmarshalText(text []byte) error {
	str := strings.ToUpper(string(text))
	switch str {
	case "DEBUG":
		*lvl = DEBUG
	case "INFO":
		*lvl = INFO
	case "WARN":
		*lvl = WARN
	case "ERROR":
		*lvl = ERROR
	case "FATAL":
		*lvl = FATAL
	default:
		return errors.New("invalid log level")
	}
	return nil
}

func (f *LogLevel) SetValue(s string) error {
	if s == "" {
		return fmt.Errorf("missing log level. choose between [\"DEBUG\", \"INFO\", \"WARN\", \"ERROR\", \"FATAL\"] ")
	}
	switch strings.ToUpper(s) {
	case "DEBUG":
		*f = DEBUG
	case "INFO":
		*f = INFO
	case "WARN":
		*f = WARN
	case "ERROR":
		*f = ERROR
	case "FATAL":
		*f = FATAL
	default:
		fmt.Printf("%s unknown log level. falling down to INFO", s)
		*f = INFO
	}
	return nil
}

func New(lvl LogLevel) *Logger {
	var l zerolog.Level
	switch lvl {
	case DEBUG:
		l = zerolog.DebugLevel
	case INFO:
		l = zerolog.InfoLevel
	case WARN:
		l = zerolog.WarnLevel
	case ERROR:
		l = zerolog.ErrorLevel
	case FATAL:
		l = zerolog.FatalLevel
	}

	zerolog.SetGlobalLevel(l)

	logger := zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + 3).Logger()

	return &Logger{
		logger: &logger,
	}
}

func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg(DEBUG, message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.msg(INFO, message, args...)
}

func (l *Logger) Warn(message string, args ...interface{}) {
	l.msg(WARN, message, args...)
}

func (l *Logger) Error(message interface{}, args ...interface{}) {
	if l.logger.GetLevel() == zerolog.DebugLevel {
		l.Debug(message, args...)
	}
	l.msg(ERROR, message, args...)
}

func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.msg(FATAL, message, args...)
	os.Exit(1)
}

func (l *Logger) msg(lvl LogLevel, message interface{}, args ...interface{}) {
	logEvent := l.logger.WithLevel(zerolog.Level(lvl - 1))
	switch msg := message.(type) {
	case error:
		if len(args) == 0 {
			logEvent.Err(msg).Msg(msg.Error())
		} else {
			logEvent.Err(msg).Msgf(msg.Error(), args...)
		}
	case string:
		if len(args) == 0 {
			logEvent.Msg(msg)
		} else {
			logEvent.Msgf(msg, args...)
		}
	default:
		logEvent.Msg(fmt.Sprintf("%s message %v has no known Type: %v", lvl.String(), message, msg))
	}
}
