package colors

import "fmt"

// Green color ANSI code
func Green(text string) string {
	return Colored(text, 32)
}

// Cyan color ANSI code
func Cyan(text string) string {
	return Colored(text, 36)
}

// Colored returns a colored text using ANSI code
func Colored(text string, ansiColor uint8) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", ansiColor, text)
}
