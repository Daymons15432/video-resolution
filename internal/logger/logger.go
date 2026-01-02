package logger

import "fmt"

func Info(tag, msg string) {
	fmt.Printf("[%s] %s\n", tag, msg)
}

func Inline(msg string) {
	fmt.Printf("\r%s", msg)
}
