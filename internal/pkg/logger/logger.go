package logger

import (
	"context"
	"io"
	"os"

	"github.com/gofiber/fiber/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
)

// Initialize initializes the global logger
func Initialize(level string, environment string) (*zap.Logger, error) {
	// Parse log level
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Configure encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	if environment == "development" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.MessageKey = "message"
	encoderConfig.LevelKey = "level"
	encoderConfig.CallerKey = "caller"

	// Create encoder
	var encoder zapcore.Encoder
	if environment == "development" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	// Create logger with options
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if environment == "development" {
		opts = append(opts, zap.Development())
	}

	logger := zap.New(core, opts...)

	globalLogger = logger
	return logger, nil
}

// Get returns the global logger instance
func Get() *zap.Logger {
	if globalLogger == nil {
		// Fallback to a basic logger if not initialized
		logger, _ := zap.NewProduction()
		return logger
	}
	return globalLogger
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// FiberLogger is an adapter that implements fiber's AllLogger interface using zap
type FiberLogger struct {
	logger *zap.Logger
}

// NewFiberLogger creates a new fiber logger from a zap logger
func NewFiberLogger(zapLogger *zap.Logger) *FiberLogger {
	return &FiberLogger{logger: zapLogger}
}

// Debug logs a debug message
func (l *FiberLogger) Debug(v ...interface{}) {
	l.logger.Sugar().Debug(v...)
}

// Debugf logs a formatted debug message
func (l *FiberLogger) Debugf(format string, v ...interface{}) {
	l.logger.Sugar().Debugf(format, v...)
}

// Info logs an info message
func (l *FiberLogger) Info(v ...interface{}) {
	l.logger.Sugar().Info(v...)
}

// Infof logs a formatted info message
func (l *FiberLogger) Infof(format string, v ...interface{}) {
	l.logger.Sugar().Infof(format, v...)
}

// Warn logs a warning message
func (l *FiberLogger) Warn(v ...interface{}) {
	l.logger.Sugar().Warn(v...)
}

// Warnf logs a formatted warning message
func (l *FiberLogger) Warnf(format string, v ...interface{}) {
	l.logger.Sugar().Warnf(format, v...)
}

// Error logs an error message
func (l *FiberLogger) Error(v ...interface{}) {
	l.logger.Sugar().Error(v...)
}

// Errorf logs a formatted error message
func (l *FiberLogger) Errorf(format string, v ...interface{}) {
	l.logger.Sugar().Errorf(format, v...)
}

// Fatal logs a fatal message and exits
func (l *FiberLogger) Fatal(v ...interface{}) {
	l.logger.Sugar().Fatal(v...)
}

// Fatalf logs a formatted fatal message and exits
func (l *FiberLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Sugar().Fatalf(format, v...)
}

// Panic logs a panic message and panics
func (l *FiberLogger) Panic(v ...interface{}) {
	l.logger.Sugar().Panic(v...)
}

// Panicf logs a formatted panic message and panics
func (l *FiberLogger) Panicf(format string, v ...interface{}) {
	l.logger.Sugar().Panicf(format, v...)
}

// Debugw logs a debug message with structured fields
func (l *FiberLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.logger.Sugar().Debugw(msg, keysAndValues...)
}

// Infow logs an info message with structured fields
func (l *FiberLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.logger.Sugar().Infow(msg, keysAndValues...)
}

// Warnw logs a warning message with structured fields
func (l *FiberLogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.logger.Sugar().Warnw(msg, keysAndValues...)
}

// Errorw logs an error message with structured fields
func (l *FiberLogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.logger.Sugar().Errorw(msg, keysAndValues...)
}

// Fatalw logs a fatal message with structured fields and exits
func (l *FiberLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.logger.Sugar().Fatalw(msg, keysAndValues...)
}

// Panicw logs a panic message with structured fields and panics
func (l *FiberLogger) Panicw(msg string, keysAndValues ...interface{}) {
	l.logger.Sugar().Panicw(msg, keysAndValues...)
}

// SetLevel sets the logging level (no-op for zap logger as level is set during initialization)
func (l *FiberLogger) SetLevel(level log.Level) {
	// This is a no-op for zap logger as level is typically set during initialization
	// To properly implement this, you would need to recreate the logger with the new level
}

// SetOutput sets the logger output (no-op for zap logger)
func (l *FiberLogger) SetOutput(writer io.Writer) {
	// This is a no-op for zap logger as output is set during initialization
}

// Trace logs a trace message (zap doesn't have trace, using debug instead)
func (l *FiberLogger) Trace(v ...interface{}) {
	l.logger.Sugar().Debug(v...)
}

// Tracef logs a formatted trace message (zap doesn't have trace, using debug instead)
func (l *FiberLogger) Tracef(format string, v ...interface{}) {
	l.logger.Sugar().Debugf(format, v...)
}

// Tracew logs a trace message with structured fields (zap doesn't have trace, using debug instead)
func (l *FiberLogger) Tracew(msg string, keysAndValues ...interface{}) {
	l.logger.Sugar().Debugw(msg, keysAndValues...)
}

// WithContext returns a logger with context (zap doesn't use context in the same way, so we just return the same logger)
func (l *FiberLogger) WithContext(ctx context.Context) log.CommonLogger {
	// Zap doesn't use context in the same way as some other loggers
	// In a more advanced implementation, you could extract trace IDs from context
	return l
}
