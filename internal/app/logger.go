package app

import "go.uber.org/zap"

func NewLogger() (*zap.Logger, error) {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level.SetLevel(zap.DebugLevel)

	return loggerConfig.Build()
}
