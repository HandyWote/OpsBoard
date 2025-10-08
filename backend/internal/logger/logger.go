package logger

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// New 创建带有统一格式的 zap.Logger。
func New() (*zap.Logger, error) {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     encodeTime,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    encoderCfg,
	}

	return cfg.Build()
}

func encodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format(time.RFC3339Nano))
}

// WithRequestID 将请求 ID 写入日志上下文，便于检索。
func WithRequestID(ctx context.Context, logger *zap.Logger) *zap.Logger {
	if logger == nil {
		return zap.NewNop()
	}
	if reqID, ok := ctx.Value(requestIDKey).(string); ok && reqID != "" {
		return logger.With(zap.String("request_id", reqID))
	}
	return logger
}

// InjectRequestID 在 context 中压入请求 ID。
func InjectRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// SyncOnShutdown 在退出前刷新日志缓冲。
func SyncOnShutdown(log *zap.Logger) {
	if log == nil {
		return
	}
	_ = log.Sync() // #nosec G104 -- 忽略同步错误
}

// ReplaceGlobals 将 zap logger 设为全局 logger。
func ReplaceGlobals(log *zap.Logger) {
	if log == nil {
		return
	}
	zap.ReplaceGlobals(log)
}

// Must 将错误视为致命。
func Must(log *zap.Logger, err error) {
	if err != nil {
		log.Fatal("logger initialization failed", zap.Error(err))
		os.Exit(1)
	}
}
