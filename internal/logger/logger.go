package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Fields log field type
type Fields map[string]interface{}

// Logger log interface
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	DebugWithFields(fields Fields, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	InfoWithFields(fields Fields, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	WarnWithFields(fields Fields, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	ErrorWithFields(fields Fields, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	FatalWithFields(fields Fields, args ...interface{})
	WithError(err error) Logger
	WithFields(fields Fields) Logger
	GetLevel() logrus.Level
}

// logrusLogger wrapper for logrus.Logger
type logrusLogger struct {
	entry *logrus.Entry
}

var (
	globalLogger Logger
)

// GetLogger gets global logger
func GetLogger() Logger {
	if globalLogger == nil {
		InitLogger("info")
	}
	return globalLogger
}

// InitLogger initializes logger
func InitLogger(logLevel string) {
	logger := logrus.New()

	// Set log format
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})

	// Set log level
	if logLevel != "" {
		level, err := logrus.ParseLevel(logLevel)
		if err == nil {
			logger.SetLevel(level)
		} else {
			logger.WithFields(logrus.Fields{
				"log_level": logLevel,
				"error":     err.Error(),
			}).Warn("Invalid log level, using default level: info")
			logger.SetLevel(logrus.InfoLevel)
		}
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	// Set output
	logger.SetOutput(os.Stderr)

	// Create wrapper
	globalLogger = &logrusLogger{
		entry: logrus.NewEntry(logger),
	}

	globalLogger.WithFields(Fields{
		"level": logger.GetLevel().String(),
	}).Debug("Logger initialization completed")
}

// Debug outputs debug log
func (l *logrusLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Debugf formats and outputs debug log
func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// DebugWithFields outputs debug log with fields
func (l *logrusLogger) DebugWithFields(fields Fields, args ...interface{}) {
	l.entry.WithFields(convertFields(fields)).Debug(args...)
}

// Info outputs info log
func (l *logrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Infof formats and outputs info log
func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// InfoWithFields outputs info log with fields
func (l *logrusLogger) InfoWithFields(fields Fields, args ...interface{}) {
	l.entry.WithFields(convertFields(fields)).Info(args...)
}

// Warn outputs warning log
func (l *logrusLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Warnf formats and outputs warning log
func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

// WarnWithFields outputs warning log with fields
func (l *logrusLogger) WarnWithFields(fields Fields, args ...interface{}) {
	l.entry.WithFields(convertFields(fields)).Warn(args...)
}

// Error outputs error log
func (l *logrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Errorf formats and outputs error log
func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// ErrorWithFields outputs error log with fields
func (l *logrusLogger) ErrorWithFields(fields Fields, args ...interface{}) {
	l.entry.WithFields(convertFields(fields)).Error(args...)
}

// Fatal outputs fatal error log
func (l *logrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Fatalf formats and outputs fatal error log
func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// FatalWithFields outputs fatal error log with fields
func (l *logrusLogger) FatalWithFields(fields Fields, args ...interface{}) {
	l.entry.WithFields(convertFields(fields)).Fatal(args...)
}

// WithError creates log entry with error
func (l *logrusLogger) WithError(err error) Logger {
	return &logrusLogger{entry: l.entry.WithError(err)}
}

// WithFields creates log entry with fields
func (l *logrusLogger) WithFields(fields Fields) Logger {
	return &logrusLogger{entry: l.entry.WithFields(convertFields(fields))}
}

// GetLevel gets log level
func (l *logrusLogger) GetLevel() logrus.Level {
	return l.entry.Logger.GetLevel()
}

// convertFields converts Fields to logrus.Fields
func convertFields(fields Fields) logrus.Fields {
	if fields == nil {
		return logrus.Fields{}
	}

	result := make(logrus.Fields, len(fields))
	for k, v := range fields {
		result[k] = v
	}
	return result
}
