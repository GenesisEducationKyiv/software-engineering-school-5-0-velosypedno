package logging

import (
	"go.uber.org/zap"
)

type LoggerFactory struct {
	base *zap.Logger
}

func NewFactory(base *zap.Logger, service string) *LoggerFactory {
	return &LoggerFactory{base: base.With(zap.String("service", service))}
}

func (f *LoggerFactory) ForPackage(pack string) *zap.Logger {
	return f.base.Named(pack)
}
