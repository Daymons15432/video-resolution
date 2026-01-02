package logger

import "fmt"

func Info(tag, msg string) {
	fmt.Printf("[%s] %s\n", tag, msg)
}

func Inline(msg string) {
	fmt.Printf("\r%s", msg)
}

func Warn(tag, msg string) {
	fmt.Printf("[%s] ⚠ %s\n", tag, msg)
}

func Error(tag, msg string) {
	fmt.Printf("[%s] ✗ %s\n", tag, msg)
}

func Success(tag, msg string) {
	fmt.Printf("[%s] ✓ %s\n", tag, msg)
}
