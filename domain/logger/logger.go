package logger

import (
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once

var log zerolog.Logger

func Get() zerolog.Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := strconv.Atoi(config.App.LogLevel)
		if err != nil {
			logLevel = int(zerolog.InfoLevel) // default to INFO
		}

		var output io.Writer

		// File logger for both dev and prod
		fileLogger := &lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    5,
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		}

		if config.App.Env == "development" {
			// Pretty console output + file logging for development
			consoleWriter := zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
			}
			output = zerolog.MultiLevelWriter(consoleWriter, fileLogger)
		} else {
			// JSON to stderr + file logging for production
			output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)
		}

		var gitRevision string

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "vcs.revision" {
					gitRevision = v.Value
					break
				}
			}
		}

		log = zerolog.New(output).
			Level(zerolog.Level(logLevel)).
			With().
			Timestamp().
			Caller().                // Add caller info (file:line)
			Int("pid", os.Getpid()). // Add process ID
			Str("git_revision", gitRevision).
			Str("go_version", buildInfo.GoVersion).
			Logger()
	})

	return log
}
