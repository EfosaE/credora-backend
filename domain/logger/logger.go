package logger


import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// RemoteLogger interface for future remote logging implementations
type RemoteLogger interface {
	Send(entry LogEntry) error
}

// Logger is the main logger struct
type Logger struct {
	mu             sync.RWMutex
	logFile        *os.File
	logLevel       LogLevel
	enableConsole  bool
	enableFile     bool
	enableRemote   bool
	remoteLogger   RemoteLogger
	maxFileSize    int64
	maxFiles       int
	logFilePath    string
	includeSource  bool
}

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	LogFilePath   string
	LogLevel      LogLevel
	EnableConsole bool
	EnableFile    bool
	MaxFileSize   int64 // in bytes
	MaxFiles      int
	IncludeSource bool
	Env           string // Environment (e.g., "dev", "test", "prod")
}


func NewLogger(config LoggerConfig) (*Logger, error) {
	// Determine default log file path if not provided
	if config.LogFilePath == "" {
		// base := "app"
		// if config.Env != "" {
		// 	base = config.Env // "test", "prod", "dev"
		// }
		// config.LogFilePath = fmt.Sprintf("%s.log", base)
		config.LogFilePath = "logs/app.log" // Default log file path
	}

	// Set sane defaults
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 10 * 1024 * 1024 // 10MB
	}
	if config.MaxFiles == 0 {
		config.MaxFiles = 5
	}

	logger := &Logger{
		logLevel:      config.LogLevel,
		enableConsole: config.EnableConsole,
		enableFile:    config.EnableFile,
		maxFileSize:   config.MaxFileSize,
		maxFiles:      config.MaxFiles,
		logFilePath:   config.LogFilePath,
		includeSource: config.IncludeSource,
	}

	if config.EnableFile {
		if err := logger.openLogFile(); err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	return logger, nil
}

// openLogFile opens or creates the log file
func (l *Logger) openLogFile() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(l.logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(l.logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.logFile = file
	return nil
}

// rotateLogFile rotates the log file if it exceeds max size
func (l *Logger) rotateLogFile() error {
	if l.logFile == nil {
		return nil
	}

	fileInfo, err := l.logFile.Stat()
	if err != nil {
		return err
	}

	if fileInfo.Size() < l.maxFileSize {
		return nil
	}

	// Close current file
	l.logFile.Close()

	// Rotate existing files
	for i := l.maxFiles - 1; i > 0; i-- {
		oldFile := fmt.Sprintf("%s.%d", l.logFilePath, i)
		newFile := fmt.Sprintf("%s.%d", l.logFilePath, i+1)
		
		if _, err := os.Stat(oldFile); err == nil {
			os.Rename(oldFile, newFile)
		}
	}

	// Move current log to .1
	rotatedFile := fmt.Sprintf("%s.1", l.logFilePath)
	if err := os.Rename(l.logFilePath, rotatedFile); err != nil {
		return err
	}

	// Open new log file
	return l.openLogFile()
}

// shouldLog checks if a message should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.logLevel
}

// getSourceInfo returns file and line number of the caller
func (l *Logger) getSourceInfo() (string, int) {
	if !l.includeSource {
		return "", 0
	}

	_, file, line, ok := runtime.Caller(3) // Skip log method, level method, and this method
	if !ok {
		return "", 0
	}
	return filepath.Base(file), line
}

// log is the core logging method
func (l *Logger) log(level LogLevel, message string, meta map[string]any) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	file, line := l.getSourceInfo()
	
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   message,
		File:      file,
		Line:      line,
		Meta:      meta,
	}

	// Write to console
	if l.enableConsole {
		l.writeToConsole(entry)
	}

	// Write to file
	if l.enableFile && l.logFile != nil {
		if err := l.rotateLogFile(); err != nil {
			log.Printf("Failed to rotate log file: %v", err)
		}
		l.writeToFile(entry)
	}

	// Send to remote logger
	if l.enableRemote && l.remoteLogger != nil {
		go func() {
			if err := l.remoteLogger.Send(entry); err != nil {
				log.Printf("Failed to send log to remote: %v", err)
			}
		}()
	}
}

// writeToConsole writes the log entry to console
func (l *Logger) writeToConsole(entry LogEntry) {
	var color string
	switch entry.Level {
	case "DEBUG":
		color = "\033[36m" // Cyan
	case "INFO":
		color = "\033[32m" // Green
	case "WARN":
		color = "\033[33m" // Yellow
	case "ERROR":
		color = "\033[31m" // Red
	default:
		color = "\033[0m" // Reset
	}

	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	sourceInfo := ""
	if entry.File != "" && entry.Line != 0 {
		sourceInfo = fmt.Sprintf(" [%s:%d]", entry.File, entry.Line)
	}

	metaStr := ""
	if len(entry.Meta) > 0 {
		if metaBytes, err := json.Marshal(entry.Meta); err == nil {
			metaStr = fmt.Sprintf(" %s", string(metaBytes))
		}
	}

	fmt.Printf("%s[%s] [%s]%s %s%s\033[0m\n", 
		color, timestamp, entry.Level, sourceInfo, entry.Message, metaStr)
}

// writeToFile writes the log entry to file
func (l *Logger) writeToFile(entry LogEntry) {
	if l.logFile == nil {
		return
	}

	logLine, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	if _, err := l.logFile.Write(append(logLine, '\n')); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}

// SetRemoteLogger sets the remote logger implementation
func (l *Logger) SetRemoteLogger(remoteLogger RemoteLogger) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.remoteLogger = remoteLogger
	l.enableRemote = remoteLogger != nil
}

// Debug logs a debug message
func (l *Logger) Debug(message string, meta ...map[string]any) {
	var metaMap map[string]any
	if len(meta) > 0 {
		metaMap = meta[0]
	}
	l.log(DEBUG, message, metaMap)
}

// Info logs an info message
func (l *Logger) Info(message string, meta ...map[string]any) {
	var metaMap map[string]any
	if len(meta) > 0 {
		metaMap = meta[0]
	}
	l.log(INFO, message, metaMap)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, meta ...map[string]any) {
	var metaMap map[string]any
	if len(meta) > 0 {
		metaMap = meta[0]
	}
	l.log(WARN, message, metaMap)
}

// Error logs an error message
func (l *Logger) Error(message string, meta ...map[string]any) {
	var metaMap map[string]any
	if len(meta) > 0 {
		metaMap = meta[0]
	}
	l.log(ERROR, message, metaMap)
}

// Close closes the logger and releases resources
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// Example remote logger implementation
type HTTPRemoteLogger struct {
	endpoint string
	// client   any // Would be *http.Client in real implementation
}

func NewHTTPRemoteLogger(endpoint string) *HTTPRemoteLogger {
	return &HTTPRemoteLogger{
		endpoint: endpoint,
		// client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (h *HTTPRemoteLogger) Send(entry LogEntry) error {
	// Implementation would send the log entry to remote endpoint
	// This is a placeholder for future implementation
	fmt.Printf("Sending to remote endpoint %s: %+v\n", h.endpoint, entry)
	return nil
}

