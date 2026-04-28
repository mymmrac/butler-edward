package logger

import "context"

type ctxKey struct{}

//nolint:gochecknoglobals
var ctxKeyValue ctxKey

// FromContext returns the logger from context.
func FromContext(ctx context.Context) Logger {
	log, ok := ctx.Value(ctxKeyValue).(Logger)
	if !ok {
		log = defaultLogger
	}
	return log
}

// ToContext adds logger to context.
func ToContext(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, ctxKeyValue, log)
}

// With adds logger to context with specified fields.
func With(ctx context.Context, args ...any) context.Context {
	return ToContext(ctx, FromContext(ctx).With(args...))
}
