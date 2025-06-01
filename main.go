package main

import (
)

func main() {
	diff("Hello\nWorld\nThis is a test.\nGoodbye", "Hello\nWorld\nThis is a different test.\nGoodbye")
	diff("Line 1\nLine 2\nLine 3", "Line 1\nLine 2\nLine 4")
}

func diff(oldContent string, newContent string) {
}
