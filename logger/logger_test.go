package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestLS(t *testing.T) {
	global := LS()
	if !global.Enabled(zapcore.DebugLevel) {
		t.Fatal("debug level should be enabled")
	}

	global.SetLevel(zapcore.WarnLevel)
	if global.Enabled(zapcore.DebugLevel) {
		t.Fatal("debug level should be disblaed")
	}

	if global.Enabled(zapcore.InfoLevel) {
		t.Fatal("info level should be disblaed")
	}

	if !global.Enabled(zapcore.WarnLevel) {
		t.Fatal("wanr level should be enabled")
	}
}
