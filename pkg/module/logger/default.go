package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mymmrac/butler-edward/pkg/module/logger/colors"
)

//nolint:gochecknoglobals
var defaultLogger Logger

// SetDefaultLogger sets default logger.
func SetDefaultLogger(log Logger) {
	defaultLogger = log
}

func init() { //nolint:gochecknoinits
	var cores []zapcore.Core //nolint:prealloc

	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "time"
	cfg.MessageKey = "message"
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	cfg.EncodeDuration = zapcore.StringDurationEncoder
	cfg.ConsoleSeparator = " "

	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncodeCaller = func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(colors.Green("<" + caller.TrimmedPath() + ">"))
	}
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.Lock(zapcore.AddSync(os.Stdout)),
		atomicLevel,
	)
	cores = append(cores, consoleCore)

	log := zap.New(zapcore.NewSamplerWithOptions(
		zapcore.NewTee(cores...), time.Second, 100, 100,
	), zap.AddCaller())
	zap.RedirectStdLog(log)
	zap.ReplaceGlobals(log)
	defaultLogger = NewZapLogger(log)
}

type ZapLogger struct {
	*zap.SugaredLogger
}

func NewZapLogger(log *zap.Logger) Logger {
	sugar := log.Sugar()
	return &ZapLogger{
		SugaredLogger: sugar,
	}
}

func (l *ZapLogger) With(args ...any) Logger {
	return &ZapLogger{
		SugaredLogger: l.SugaredLogger.With(args...),
	}
}

func (l *ZapLogger) WithOptions(opts ...Option) Logger {
	zOps := make([]zap.Option, 0, len(opts))
	for _, opt := range opts {
		switch o := opt.(type) {
		case *SkipCallerOption:
			zOps = append(zOps, zap.AddCallerSkip(o.Skip))
		case *IncreasedLevelOption:
			var level zapcore.Level
			switch o.Level {
			case LevelInfo:
				level = zapcore.InfoLevel
			case LevelWarn:
				level = zapcore.WarnLevel
			case LevelError:
				level = zapcore.ErrorLevel
			case LevelFatal:
				level = zapcore.FatalLevel
			default:
				level = zapcore.DebugLevel
			}
			zOps = append(zOps, zap.IncreaseLevel(level))
		default:
			l.Errorw("unsupported option type", "type", fmt.Sprintf("%T", o))
		}
	}
	return &ZapLogger{
		SugaredLogger: l.SugaredLogger.WithOptions(zOps...),
	}
}

//nolint:gochecknoglobals
var skipOne = WithSkipCaller(1)
