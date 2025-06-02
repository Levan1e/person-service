package logger

import (
	"fmt"
	"person-service/internal/config"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger представляет обёртку над zap.Logger для удобного логирования.
type Logger struct {
	*zap.Logger
}

// NewLogger создаёт новый экземпляр логгера в режиме разработки.
func NewLogger(cfg *config.Config) (*Logger, error) {
	var zapCfg zap.Config
	switch cfg.LogLevel {
	case "debug":
		zapCfg = zap.NewDevelopmentConfig()
	case "info", "":
		zapCfg = zap.NewProductionConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "error":
		zapCfg = zap.NewProductionConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		return nil, fmt.Errorf("неподдерживаемый уровень логирования: %s", cfg.LogLevel)
	}

	zapCfg.EncoderConfig.CallerKey = "caller"
	zapCfg.EncoderConfig.StacktraceKey = "stacktrace"

	logger, err := zapCfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel), // Стек-трейсы для Error и выше
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать логгер: %w", err)
	}

	return &Logger{Logger: logger}, nil
}

// Sync сбрасывает буфер логов. На Windows возвращает nil, так как синхронизация не требуется.
func (l *Logger) Sync() error {
	if runtime.GOOS == "windows" {
		return nil
	}
	return l.Logger.Sync()
}

// Fatal логирует сообщение с уровнем Fatal и завершает программу.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// Fatalf форматирует и логирует сообщение с уровнем Fatal.
func (l *Logger) Fatalf(msg string, args ...any) {
	l.Logger.Fatal(fmt.Sprintf(msg, args...))
}

// FatalKV логирует сообщение с уровнем Fatal и дополнительными полями.
func (l *Logger) FatalKV(msg string, kv ...any) {
	l.Logger.Fatal(msg, parseKV(kv...)...)
}

// Panic логирует сообщение с уровнем Panic и вызывает панику.
func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.Logger.Panic(msg, fields...)
}

// Panicf форматирует и логирует сообщение с уровнем Panic.
func (l *Logger) Panicf(msg string, args ...any) {
	l.Logger.Panic(fmt.Sprintf(msg, args...))
}

// PanicKV логирует сообщение с уровнем Panic и дополнительными полями.
func (l *Logger) PanicKV(msg string, kv ...any) {
	l.Logger.Panic(msg, parseKV(kv...)...)
}

// Error логирует сообщение с уровнем Error.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Errorf форматирует и логирует сообщение с уровнем Error.
func (l *Logger) Errorf(msg string, args ...any) {
	l.Logger.Error(fmt.Sprintf(msg, args...))
}

// ErrorKV логирует сообщение с уровнем Error и дополнительными полями.
func (l *Logger) ErrorKV(msg string, kv ...any) {
	l.Logger.Error(msg, parseKV(kv...)...)
}

// Warn логирует сообщение с уровнем Warn.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Warnf форматирует и логирует сообщение с уровнем Warn.
func (l *Logger) Warnf(msg string, args ...any) {
	l.Logger.Warn(fmt.Sprintf(msg, args...))
}

// WarnKV логирует сообщение с уровнем Warn и дополнительными полями.
func (l *Logger) WarnKV(msg string, kv ...any) {
	l.Logger.Warn(msg, parseKV(kv...)...)
}

// Info логирует сообщение с уровнем Info.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Infof форматирует и логирует сообщение с уровнем Info.
func (l *Logger) Infof(msg string, args ...any) {
	l.Logger.Info(fmt.Sprintf(msg, args...))
}

// InfoKV логирует сообщение с уровнем Info и дополнительными полями.
func (l *Logger) InfoKV(msg string, kv ...any) {
	l.Logger.Info(msg, parseKV(kv...)...)
}

// Debug логирует сообщение с уровнем Debug.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Debugf форматирует и логирует сообщение с уровнем Debug.
func (l *Logger) Debugf(msg string, args ...any) {
	l.Logger.Debug(fmt.Sprintf(msg, args...))
}

// DebugKV логирует сообщение с уровнем Debug и дополнительными полями.
func (l *Logger) DebugKV(msg string, kv ...any) {
	l.Logger.Debug(msg, parseKV(kv...)...)
}

// ErrorKV возвращает поле для логирования ошибки.
func ErrorKV(key string, value any) zap.Field {
	return zap.Any(key, value)
}

// DebugKV возвращает поле для логирования отладочной информации.
func DebugKV(key string, value any) zap.Field {
	return zap.Any(key, value)
}

// InfoKV возвращает поле для логирования информационных сообщений.
func InfoKV(key string, value any) zap.Field {
	return zap.Any(key, value)
}

// parseKV преобразует пары ключ-значение в zap.Field.
func parseKV(kv ...any) []zap.Field {
	if len(kv)%2 != 0 {
		return []zap.Field{zap.String("error", "kv must be pairs")}
	}
	kvs := len(kv) / 2
	fields := make([]zap.Field, 0, kvs)
	for i := 0; i < kvs; i += 2 {
		k, ok := kv[i].(string)
		if !ok {
			return []zap.Field{zap.String("error", "kv key must be string")}
		}
		fields = append(fields, zap.Any(k, kv[i+1]))
	}
	return fields
}
