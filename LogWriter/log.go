package LogWriter

import (
	"BHLayer2Node/utils"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	// 全局 LogWriter 实例
	globalLogWriter *LogWriter
	once            sync.Once
)

// LogWriter handles logging to console and file
type LogWriter struct {
	LogPath   string
	Debug     bool
	Logger    *log.Logger
	LogFile   *os.File
	LogLevels map[string]func(string)
}

// NewLogWriter creates a new LogWriter instance
func NewLogWriter(logPath string, debug bool) *LogWriter {
	return &LogWriter{
		LogPath:   logPath,
		Debug:     debug,
		LogLevels: make(map[string]func(string)),
	}
}

// Init initializes the logger
func (lw *LogWriter) Init() error {
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	if !lw.Debug {
		currentTime := time.Now().Format("2006-01-02_15-04-05")
		logFile := fmt.Sprintf("%s/%s.log", lw.LogPath, currentTime)
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		lw.LogFile = file
		writers = append(writers, file)
	}

	multiWriter := io.MultiWriter(writers...)
	lw.Logger = log.New(multiWriter, "", log.LstdFlags)

	lw.setupLogLevels()
	return nil
}

// setupLogLevels sets up custom log levels
func (lw *LogWriter) setupLogLevels() {
	lw.LogLevels["DEBUG"] = func(msg string) { lw.log("DEBUG", msg) }
	lw.LogLevels["INFO"] = func(msg string) { lw.log("INFO", msg) }
	lw.LogLevels["WARNING"] = func(msg string) { lw.log("WARNING", msg) }
	lw.LogLevels["ERROR"] = func(msg string) { lw.log("ERROR", msg) }
	lw.LogLevels["CRITICAL"] = func(msg string) { lw.log("CRITICAL", msg) }
	lw.LogLevels["NETWORK"] = func(msg string) { lw.log("NETWORK", msg) }
	lw.LogLevels["SCHEDULE"] = func(msg string) { lw.log("SCHEDULE", msg) }
	lw.LogLevels["CHAINUP"] = func(msg string) { lw.log("CHAINUP", msg) }
	lw.LogLevels["COORDINATOR"] = func(msg string) { lw.log("COORDINATOR", msg) }
	lw.LogLevels["TRACKER"] = func(msg string) { lw.log("TRACKER", msg) }
	lw.LogLevels["VOTE"] = func(msg string) { lw.log("VOTE", msg) }
	lw.LogLevels["EPOCH"] = func(msg string) { lw.log("EPOCH", msg) }
	lw.LogLevels["COLLECT"] = func(msg string) { lw.log("COLLECT", msg) }
	lw.LogLevels["QUERY"] = func(msg string) { lw.log("QUERY", msg) }
}

// log logs a message with a specific level
func (lw *LogWriter) log(level, msg string) {
	lw.Logger.Printf("[%s] %s", level, msg)
}

// Log writes a log message at a specific level
func (lw *LogWriter) Log(level, message string) {
	if logFunc, exists := lw.LogLevels[level]; exists {
		logFunc(message)
	} else {
		lw.Logger.Printf("[UNKNOWN] %s", message)
	}
}

// Close closes the log file if opened
func (lw *LogWriter) Close() {
	if lw.LogFile != nil {
		_ = lw.LogFile.Close()
	}
}

// InitGlobalLogWriter initializes the global LogWriter instance
func InitGlobalLogWriter(logPath string, debug bool) {
	once.Do(func() {
		root, err := utils.GetProjectRoot()
		if err != nil {
			log.Fatalf("Failed to find root path: %v", err)
		}
		path := filepath.Join(root, logPath)
		globalLogWriter = NewLogWriter(path, debug)
		err = globalLogWriter.Init()
		if err != nil {
			log.Fatalf("Error initializing global LogWriter: %v", err)
		}
		globalLogWriter.Log("INFO", "Global LogWriter initialized successfully")
	})
}

// Log is a package-level function to write logs directly
func Log(level, message string) {
	if globalLogWriter == nil {
		fmt.Println("Global LogWriter is not initialized")
		return
	}
	globalLogWriter.Log(level, message)
}

// CloseGlobalLogWriter closes the global LogWriter instance
func CloseGlobalLogWriter() {
	if globalLogWriter != nil {
		globalLogWriter.Close()
	}
}
