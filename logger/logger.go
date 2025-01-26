// Package logger provides minimal logging features.
package logger

import "log"

func LogError(message string, err error) {
	log.Printf("💥 %s\n", message)
	log.Print(err)
}

func LogInfo(message string) {
	log.Printf("ℹ️ %s\n", message)
}

func LogSuccess(message string) {
	log.Printf("🎉 %s\n", message)
}
