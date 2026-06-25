package logging

import (
	"context"
	"os"

	"github.com/gianglt1/short-link/internal/common"
	"github.com/gianglt1/short-link/internal/config"
	"github.com/gianglt1/short-link/internal/utils"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
	serviceName string
}

type LoggerParams struct {
	fx.In

	Config *config.Config
}

var Module = fx.Provide(NewLogger)

// NewLogger creates a new logger instance
func NewLogger(p LoggerParams) *Logger {
	var coreArr []zapcore.Core

	config := p.Config

	if config.Server.Environment == "production" {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder := zapcore.NewJSONEncoder(encoderConfig)

		consoleCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), zapcore.InfoLevel)
		coreArr = append(coreArr, consoleCore)
	} else {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), zapcore.InfoLevel) // The third and subsequent parameters are the log levels for writing to the file. In ErrorLevel mode, only error - level logs are recorded.

		coreArr = append(coreArr, consoleCore)
	}

	log := zap.New(zapcore.NewTee(coreArr...), zap.AddCaller()) // zap.AddCaller() is used to display the file name and line number and can be omitted.
	defer func() { _ = log.Sync() }()

	log = log.With(zap.String("service", config.Server.Name), zap.String("environment", config.Server.Environment))
	return &Logger{
		serviceName: config.Server.Name,
		Logger:      log,
	}
}

func NewNoopLogger() *Logger {
	return &Logger{
		Logger:      zap.NewNop(),
		serviceName: "test",
	}
}

func (l *Logger) GetLogger() *zap.Logger { return l.Logger }

type WrappedLogger struct {
	*zap.Logger
}

func (l *Logger) GetWrappedLogger(ctx context.Context) *WrappedLogger {
	return &WrappedLogger{
		Logger: l.With(l.extractContext(ctx)...),
	}
}

func (l *Logger) extractContext(ctx context.Context) []zap.Field {
	fields := []zap.Field{
		zap.String("service_name", l.serviceName),
	}

	requestID := utils.GetRequestIDFromCtx(ctx)

	if requestID != "" {
		fields = append(fields, zap.String(string(common.KEY_REQUEST_ID), requestID))
	}

	return fields
}

// For PGX logging
func (l *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	log := l.GetWrappedLogger(ctx)

	fields := make([]zap.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		log.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		log.Info(msg, fields...)
	case tracelog.LogLevelWarn:
		log.Warn(msg, fields...)
	case tracelog.LogLevelError:
		log.Error(msg, fields...)
	}
}
