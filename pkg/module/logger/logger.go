package logger

import (
	"context"
)

// Logger representation.
type Logger interface {
	// With adds a variadic number of fields to the logging context.
	With(args ...any) Logger
	// WithOptions applies the supplied options.
	WithOptions(opts ...Option) Logger

	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)

	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Fatalf(template string, args ...any)

	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)
}

func Debug(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Debug(args...)
}

func Info(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Info(args...)
}

func Warn(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Warn(args...)
}

func Error(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Error(args...)
}

func Fatal(ctx context.Context, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Fatal(args...)
}

func Debugf(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Debugf(template, args...)
}

func Infof(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Infof(template, args...)
}

func Warnf(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Warnf(template, args...)
}

func Errorf(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Errorf(template, args...)
}

func Fatalf(ctx context.Context, template string, args ...any) {
	FromContext(ctx).WithOptions(skipOne).Fatalf(template, args...)
}

func Debugw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Debugw(msg, keysAndValues...)
}

func Infow(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Infow(msg, keysAndValues...)
}

func Warnw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Warnw(msg, keysAndValues...)
}

func Errorw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Errorw(msg, keysAndValues...)
}

func Fatalw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).WithOptions(skipOne).Fatalw(msg, keysAndValues...)
}
