package logger

// Option represents a logger option.
type Option interface {
	option()
}

// SkipCallerOption represents a logger option to skip caller.
type SkipCallerOption struct {
	Skip int
}

func (s *SkipCallerOption) option() {}

// WithSkipCaller returns a new logger option to skip caller.
func WithSkipCaller(skip int) Option {
	return &SkipCallerOption{Skip: skip}
}

// IncreasedLevelOption represents a logger option to increase the log level.
type IncreasedLevelOption struct {
	Level Level
}

func (i *IncreasedLevelOption) option() {}

// WithIncreasedLevel returns a new logger option to increase the log level.
func WithIncreasedLevel(level Level) Option {
	return &IncreasedLevelOption{Level: level}
}
