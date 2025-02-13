package paradigm

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
	LogPath string
	Debug   bool
	Logger  *log.Logger
	Printer *log.Logger
	LogFile *os.File
	//LogLevels map[string]func(string)
}

// NewLogWriter creates a new LogWriter instance
func NewLogWriter(logPath string, debug bool) *LogWriter {
	return &LogWriter{
		LogPath: logPath,
		Debug:   debug,
		//LogLevels: make(map[string]func(string)),
	}
}

// Init initializes the logger
func (lw *LogWriter) Init() error {
	var logWriters []io.Writer
	var printWriters []io.Writer
	printWriters = append(printWriters, os.Stdout) // 默认输出到终端

	if !lw.Debug {
		currentTime := time.Now().Format("2006-01-02_15-04-05")
		logFile := fmt.Sprintf("%s/%s.log", lw.LogPath, currentTime)
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		lw.LogFile = file
		printWriters = append(printWriters, file) // 写入日志文件
		logWriters = append(logWriters, file)
	}

	logMultiWriter := io.MultiWriter(logWriters...)
	lw.Logger = log.New(logMultiWriter, "", log.LstdFlags)
	printMultiWriter := io.MultiWriter(printWriters...)
	lw.Printer = log.New(printMultiWriter, "", log.LstdFlags)
	//lw.setupLogLevels()
	return nil
}

// setupLogLevels sets up custom log levels
//func (lw *LogWriter) setupLogLevels() {
//	lw.LogLevels["DEBUG"] = func(msg string) { lw.log("DEBUG", msg) }
//	lw.LogLevels["INFO"] = func(msg string) { lw.log("INFO", msg) }
//	lw.LogLevels["WARNING"] = func(msg string) { lw.log("WARNING", msg) }
//	lw.LogLevels["ERROR"] = func(msg string) { lw.log("ERROR", msg) }
//	lw.LogLevels["CRITICAL"] = func(msg string) { lw.log("CRITICAL", msg) }
//	lw.LogLevels["NETWORK"] = func(msg string) { lw.log("NETWORK", msg) }
//	lw.LogLevels["HTTP"] = func(msg string) { lw.log("HTTP", msg) }
//
//	lw.LogLevels["SCHEDULE"] = func(msg string) { lw.log("SCHEDULE", msg) }
//	lw.LogLevels["CHAINUP"] = func(msg string) { lw.log("CHAINUP", msg) }
//	lw.LogLevels["COORDINATOR"] = func(msg string) { lw.log("COORDINATOR", msg) }
//	lw.LogLevels["TRACKER"] = func(msg string) { lw.log("TRACKER", msg) }
//	lw.LogLevels["VOTE"] = func(msg string) { lw.log("VOTE", msg) }
//	lw.LogLevels["EPOCH"] = func(msg string) { lw.log("EPOCH", msg) }
//	lw.LogLevels["COLLECT"] = func(msg string) { lw.log("COLLECT", msg) }
//	lw.LogLevels["ORACLE"] = func(msg string) { lw.log("ORACLE", msg) }
//}

// log logs a message with a specific level
func (lw *LogWriter) log(level, msg string) {
	lw.Logger.Printf("[%s] %s", level, msg)
}

// Log writes a log message at a specific level (only logs to file, not to console)
func (lw *LogWriter) Log(level, message string) {
	//if logFunc, exists := lw.LogLevels[level]; exists {
	//	logFunc(message) // 只写入日志
	//} else {
	//	lw.Logger.Printf("[UNKNOWN] %s", message) // 只写入日志
	//}
	lw.Logger.Printf("[%s] %s", level, message) // 只写入日志
}

// Print writes a log message at a specific level and also prints to console
func (lw *LogWriter) Print(level, message string) {
	//if logFunc, exists := lw.LogLevels[level]; exists {
	//	logFunc(message)     // 写入日志文件
	//	fmt.Println(message) // 同时打印到控制台
	//} else {
	//	lw.Logger.Printf("[UNKNOWN] %s", message) // 写入日志文件
	//	fmt.Println(message)                      // 同时打印到控制台
	//}
	lw.Printer.Printf("[%s] %s", level, message)
}

// Error handles error logging and printing
func (lw *LogWriter) Error(errorEnum ErrorEnum, errorMessage string) RappaError {
	error := NewRappaError(errorEnum, errorMessage)
	lw.Print("ERROR", error.Error()) // 写入日志文件
	return error
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

// Log is a package-level function to write logs directly (only logs to file, not to console)
func Log(level, message string) {
	if globalLogWriter == nil {
		fmt.Println("Global LogWriter is not initialized")
		return
	}
	globalLogWriter.Log(level, message)
}

// Print is a package-level function to write logs and also print to console
func Print(level, message string) {
	if globalLogWriter == nil {
		fmt.Println("Global LogWriter is not initialized")
		return
	}
	globalLogWriter.Print(level, message)
}

// Error is a package-level function to handle error logging and printing
func Error(errorEnum ErrorEnum, errorMessage string) RappaError {
	if globalLogWriter == nil {
		fmt.Println("Global LogWriter is not initialized")
		return RappaError{
			errorType:    RuntimeError,
			errorMessage: "Global LogWriter is not initialized",
		}
	}
	return globalLogWriter.Error(errorEnum, errorMessage)
}

// CloseGlobalLogWriter closes the global LogWriter instance
func CloseGlobalLogWriter() {
	if globalLogWriter != nil {
		globalLogWriter.Close()
	}
}
