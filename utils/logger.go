package utils

import (
	"log"
	"os"
)

// Initialize logger
var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// LogInfo logs informational messages
func LogInfo(message string) {
	InfoLogger.Println(message)
}

// LogError logs error messages
func LogError(message string) {
	ErrorLogger.Println(message)
}

// LogDebug logs debug messages
func LogDebug(message string) {
	// Uncomment if you want debug-level logs
	// InfoLogger.Println("DEBUG: " + message)
}
