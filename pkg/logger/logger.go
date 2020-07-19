package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger .
type Logger struct {
	*zap.Logger
}

// NewLogger creates a wrapper around zap
func NewLogger(logLevel string) (*Logger, error) {
	zap, err := newZap(logLevel)
	return &Logger{
		zap,
	}, err
}

// Debugf debug logs in printf style
func (l *Logger) Debugf(msg string, values ...interface{}) {
	l.Debug(fmt.Sprintf(msg, values...))
}

// Infof info logs in printf style
func (l *Logger) Infof(msg string, values ...interface{}) {
	l.Info(fmt.Sprintf(msg, values...))
}

// Errorf error logs in printf style
func (l *Logger) Errorf(msg string, values ...interface{}) {
	l.Error(fmt.Sprintf(msg, values...))
}

// newZap creates a new zapcore logger instance
func newZap(logLevel string) (*zap.Logger, error) {
	atom := zap.NewAtomicLevel()
	err := atom.UnmarshalText([]byte(logLevel))
	if err != nil {
		return nil, err
	}
	encoderCfg := zap.NewProductionEncoderConfig()

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))

	return logger, nil
}
