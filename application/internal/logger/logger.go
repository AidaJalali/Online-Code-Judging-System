package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

var (
	AppLogger *log.Logger
)

func Init() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	// Find the next available try-<index>.log filename
	files, err := os.ReadDir("logs")
	if err != nil {
		log.Fatal("Failed to read logs directory:", err)
	}

	maxIndex := 0
	pattern := regexp.MustCompile(`^try-(\\d+)\\.log$`)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		matches := pattern.FindStringSubmatch(file.Name())
		if len(matches) == 2 {
			var idx int
			fmt.Sscanf(matches[1], "%d", &idx)
			if idx > maxIndex {
				maxIndex = idx
			}
		}
	}
	newIndex := maxIndex + 1
	logFileName := filepath.Join("logs", fmt.Sprintf("try-%d.log", newIndex))

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	AppLogger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func Printf(format string, v ...interface{}) {
	if AppLogger != nil {
		AppLogger.Printf(format, v...)
	}
}

func Println(v ...interface{}) {
	if AppLogger != nil {
		AppLogger.Println(v...)
	}
}
