package logger

import (
	"go.uber.org/zap"
)

var (
	emptyLevel zap.AtomicLevel

	ls *Logger
)

func init() {
	var err error
	ls, err = NewDefault()
	if err != nil {
		panic(err)
	}
}

// Setup sets the given logger as the global variable
func Setup(l *Logger) {
	if l != nil {
		ls = l
	}
}

// LS returns the global logger variable
func LS() *Logger {
	return ls
}

// New contructs a Logger with given Config and Options
func New(cfg zap.Config, opts ...zap.Option) (*Logger, error) {
	level := cfg.Level

	if level == emptyLevel {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	cfg.Level = level

	raw, err := cfg.Build(opts...)
	if err != nil {
		return nil, err
	}

	return &Logger{
		raw:           raw,
		SugaredLogger: raw.Sugar(),
		AtomicLevel:   level,
	}, nil
}

// NewDefault constructs a Logger with default Config
func NewDefault(opts ...zap.Option) (*Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true

	return New(cfg, opts...)
}

// Logger Common Logger
type Logger struct {
	raw *zap.Logger
	*zap.SugaredLogger
	zap.AtomicLevel
}
