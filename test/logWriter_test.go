package test

import (
	"BHLayer2node/LogWriter"
	"BHLayer2node/utils"
	"os"
	"path/filepath"
	"testing"
)

// TestLogWriter tests the LogWriter functionality
func TestLogWriter(t *testing.T) {
	root_path, err := utils.GetProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find root path: %v", err)
	}
	logPath := filepath.Join(root_path, "logs")
	debug := false

	// Ensure test log directory exists
	if err := os.MkdirAll(logPath, 0755); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}

	// Create and initialize LogWriter
	logWriter := LogWriter.NewLogWriter(logPath, debug)
	if err := logWriter.Init(); err != nil {
		t.Fatalf("Failed to initialize LogWriter: %v", err)
	}
	defer logWriter.Close()

	// Log messages of different levels
	logWriter.Log("DEBUG", "This is a debug message for testing")
	logWriter.Log("INFO", "This is an info message for testing")
	logWriter.Log("WARNING", "This is a warning message for testing")
	logWriter.Log("ERROR", "This is an error message for testing")
	logWriter.Log("CRITICAL", "This is a critical message for testing")
	logWriter.Log("NETWORK", "This is a network message for testing")
	logWriter.Log("SCHEDULE", "This is an schedule message for testing")
	logWriter.Log("CHAINUP", "This is a chainup message for testing")

	// Verify log file is created
	files, err := os.ReadDir(logPath)
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("No log file was created in the directory: %s", logPath)
	}

	// Check log file content
	logFile := logPath + "/" + files[0].Name()
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	expectedMessages := []string{
		"This is a debug message for testing",
		"This is an info message for testing",
		"This is a warning message for testing",
		"This is an error message for testing",
		"This is a critical message for testing",
		"This is a network message for testing",
		"This is an schedule message for testing",
		"This is a chainup message for testing",
	}

	for _, msg := range expectedMessages {
		if !contains(string(content), msg) {
			t.Errorf("Log file does not contain expected message: %s", msg)
		}
	}

	//// Cleanup
	//if err := os.RemoveAll(logPath); err != nil {
	//	t.Logf("Failed to clean up log directory: %v", err)
	//}
}

// Helper function to check if a string is contained in another string
func contains(haystack, needle string) bool {
	return len(haystack) > 0 && len(needle) > 0 && (len(haystack) >= len(needle) && haystack[:len(needle)] == needle || contains(haystack[1:], needle))
}
