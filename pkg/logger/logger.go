package logger

import (
	"log"
)

// InfoC logs an info message with a module name.
func InfoC(module, msg string) {
	log.Printf("[%s] INFO: %s", module, msg)
}

// InfoCF logs an info message with a module name and format data.
func InfoCF(module, msg string, data map[string]any) {
	log.Printf("[%s] INFO: %s %+v", module, msg, data)
}

// WarnCF logs a warning message with a module name and format data.
func WarnCF(module, msg string, data map[string]any) {
	log.Printf("[%s] WARN: %s %+v", module, msg, data)
}

// ErrorCF logs an error message with a module name and format data.
func ErrorCF(module, msg string, data map[string]any) {
	log.Printf("[%s] ERROR: %s %+v", module, msg, data)
}

// DebugCF logs a debug message with a module name and format data.
func DebugCF(module, msg string, data map[string]any) {
	log.Printf("[%s] DEBUG: %s %+v", module, msg, data)
}
