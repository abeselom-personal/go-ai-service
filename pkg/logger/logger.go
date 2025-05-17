// pkg/logger/logger.go
package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm/logger"
)

// Logger is the main logger struct
type Logger struct {
	*log.Logger
}

// NewLogger creates a new Logger instance
func NewLogger() *Logger {
	return &Logger{
		log.New(os.Stdout, "[APP] ", log.LstdFlags|log.Lshortfile),
	}
}

// Log levels
const (
	InfoLevel = iota
	WarnLevel
	ErrorLevel
	DebugLevel
)

func (l *Logger) Info(msg string) {
	l.Output(2, fmt.Sprintf("INFO: %s", msg))
}

func (l *Logger) Warn(msg string) {
	l.Output(2, fmt.Sprintf("WARN: %s", msg))
}

func (l *Logger) Error(msg string) {
	l.Output(2, fmt.Sprintf("ERROR: %s", msg))
}

func (l *Logger) Debug(msg string) {
	l.Output(2, fmt.Sprintf("DEBUG: %s", msg))
}

// GormLogger implements gorm's logger interface
type GormLogger struct {
	logger *Logger
}

func NewGormLogger(l *Logger) *GormLogger {
	return &GormLogger{logger: l}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Warn(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Error(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	if err != nil {
		l.logger.Error(fmt.Sprintf("SQL: %s [%d rows] | %v", sql, rows, err))
		return
	}
	l.logger.Debug(fmt.Sprintf("SQL: %s [%d rows] | %s", sql, rows, time.Since(begin)))
}
