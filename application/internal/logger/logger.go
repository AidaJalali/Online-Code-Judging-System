package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func Init() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	// Create log files
	infoFile, err := os.OpenFile("logs/info.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open info log file:", err)
	}

	errorFile, err := os.OpenFile("logs/error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open error log file:", err)
	}

	// Initialize loggers
	InfoLogger = log.New(infoFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	InfoLogger.Printf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message)
}

func Error(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	ErrorLogger.Printf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), message)
}
