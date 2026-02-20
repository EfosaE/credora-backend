package test

import (
	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/rs/zerolog"
)

func SetupTestLogger() zerolog.Logger {
	testLogger := logger.Get()
	// // Get project root (based on current file location)
	// _, b, _, _ := runtime.Caller(0)
	// projectRoot := filepath.Join(filepath.Dir(b), "..")

	// logFilePath := filepath.Join(projectRoot, "logs", "test.log")
	// l, err := logger.NewLogger(logger.LoggerConfig{
	// 	LogFilePath:   logFilePath,
	// 	LogLevel:      logger.DEBUG,
	// 	EnableConsole: false,
	// 	EnableFile:    true,
	// 	IncludeSource: true,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	return testLogger.With().
		Str("component", "test").
		Logger()
}
