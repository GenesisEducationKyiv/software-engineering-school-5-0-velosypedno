package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logPerm os.FileMode = 0644
const logDirPerm os.FileMode = 0755

type Factory interface {
	ForPackage(pkg string) *zap.Logger
	Sync() error
}

type LoggerFactory struct {
	base *zap.Logger
}

func NewFactory(logDir string, service string) (*LoggerFactory, error) {
	if err := os.MkdirAll(logDir, logDirPerm); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFilePath := filepath.Join(logDir, service+".log")
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logPerm) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(encoderCfg)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
	fileWriter := zapcore.AddSync(file)
	consoleWriter := zapcore.Lock(os.Stdout)

	level := zapcore.InfoLevel
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, level)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, level)
	core := zapcore.NewTee(consoleCore, fileCore)
	logger := zap.New(core, zap.AddCaller()).With(zap.String("service", service))
	return &LoggerFactory{base: logger}, nil
}

func (f *LoggerFactory) ForPackage(pkg string) *zap.Logger {
	return f.base.Named(pkg)
}

func (f *LoggerFactory) Sync() error {
	return f.base.Sync()
}

func NewFakeFactory() *LoggerFactory {
	return &LoggerFactory{base: zap.NewNop()}
}
