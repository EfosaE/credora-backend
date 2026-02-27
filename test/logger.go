package test

import (
	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/rs/zerolog"
)

func SetupTestLogger() zerolog.Logger {
	testLogger := logger.Get()
	return testLogger.With().
		Str("component", "test").
		Logger()
}
