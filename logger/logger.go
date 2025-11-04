package logger

import (
	"go.uber.org/zap"
)

func New() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "json"
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	log, _ := cfg.Build()
	return log
}
