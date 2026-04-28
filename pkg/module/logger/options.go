package logger

type Option interface {
	option()
}

type SkipCallerOption struct {
	Skip int
}

func (s *SkipCallerOption) option() {}

func WithSkipCaller(skip int) Option {
	return &SkipCallerOption{Skip: skip}
}
