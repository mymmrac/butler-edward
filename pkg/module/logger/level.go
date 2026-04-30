package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//nolint:gochecknoglobals
var atomicLevel = zap.NewAtomicLevelAt(zap.DebugLevel)

// AtomicLevel returns global log level.
func AtomicLevel() zap.AtomicLevel {
	return atomicLevel
}

// SetLevel sets global log level.
func SetLevel(lvl string) {
	l := zapcore.DebugLevel
	switch strings.ToLower(lvl) {
	case "debug":
		l = zapcore.DebugLevel
	case "info":
		l = zapcore.InfoLevel
	case "warn":
		l = zapcore.WarnLevel
	case "error":
		l = zapcore.ErrorLevel
	case "fatal":
		l = zapcore.FatalLevel
	default:
		// Fallthrough with default level
	}
	atomicLevel.SetLevel(l)
}

// Level is log level.
type Level string

// Log levels.
const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelFatal Level = "fatal"
)
