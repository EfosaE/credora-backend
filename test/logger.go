package test

import (
	"path/filepath"
	"runtime"

	"github.com/EfosaE/credora-backend/domain/logger"
)

func SetupTestLogger() *logger.Logger {
	// Get project root (based on current file location)
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "..") 

	logFilePath := filepath.Join(projectRoot, "logs", "test.log")
	l, err := logger.NewLogger(logger.LoggerConfig{
		LogFilePath:   logFilePath,
		LogLevel:      logger.DEBUG,
		EnableConsole: false,
		EnableFile:    true,
		IncludeSource: true,
	})
	if err != nil {
		panic(err)
	}
	return l
}
