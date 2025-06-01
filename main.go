package main

import (
	"fmt"
	"strings"
)

func main() {
	diff("Hello\nWorld\nThis is a test.\nGoodbye", "Hello\nWorld\nThis is a different test.\nGoodbye")
	diff("Line 1\nLine 2\nLine 3", "Line 1\nLine 2\nLine 4")
}
func diff(oldContent string, newContent string) {
	type Change struct {
		Op    string
		Value string
	}
	diffs := []Change{}

	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	lcs := make([][]int, len(oldLines)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(newLines)+1)
	}

	for i := 1; i <= len(oldLines); i++ {
		for j := 1; j <= len(newLines); j++ {
			if oldLines[i-1] == newLines[j-1] {
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else {
				lcs[i][j] = max(lcs[i-1][j], lcs[i][j-1])
			}
		}
	}

	i, j := len(oldLines), len(newLines)
	for i > 0 && j > 0 {
		if oldLines[i-1] == newLines[j-1] {
			diffs = append(diffs, Change{Op: "equal", Value: oldLines[i-1]})
			i--
			j--
		} else if lcs[i-1][j] >= lcs[i][j-1] {
			diffs = append(diffs, Change{Op: "delete", Value: oldLines[i-1]})
			i--
		} else {
			diffs = append(diffs, Change{Op: "insert", Value: newLines[j-1]})
			j--
		}
	}

	// Handle leftover lines in oldLines (deletions)
	for i > 0 {
		diffs = append(diffs, Change{Op: "delete", Value: oldLines[i-1]})
		i--
	}

	// Handle leftover lines in newLines (insertions)
	for j > 0 {
		diffs = append(diffs, Change{Op: "insert", Value: newLines[j-1]})
		j--
	}

	// Reverse the diffs slice since we built it backwards
	for left, right := 0, len(diffs)-1; left < right; left, right = left+1, right-1 {
		diffs[left], diffs[right] = diffs[right], diffs[left]
	}

	// Print diffs nicely
	for _, d := range diffs {
		fmt.Printf("%s: %s\n", d.Op, d.Value)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

