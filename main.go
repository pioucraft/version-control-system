package main

import (
	"fmt"
	"strings"
)

type Change struct {
	Op string
	Value string
	Line int
}

/*

TO IMPLEMENT :
- Add to commits history
- Get diff between version of key in history and new version.
- Get key content (at specific version or at latest version)
- Get key history
*/


func main() {
	fmt.Println(diff("Hello\nWorld\nThis is a test.\nGoodbye", "Hello\nWorld\nThis is a different test.\nGoodbye"))
	fmt.Println(diff("Line 1\nLine 2\nLine 3", "Line 1\nLine 2\nLine 4"))
	fmt.Println(diff("Line A\nLine B", "Line C\nLine D\nLine E"))
}

func diff(oldContent string, newContent string) []Change {
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
			diffs = append(diffs, Change{Op: "equal", Value: oldLines[i-1], Line: i})
			i--
			j--
		} else if lcs[i-1][j] > lcs[i][j-1] {
			diffs = append(diffs, Change{Op: "delete", Value: oldLines[i-1], Line: i})
			i--
		} else {
			diffs = append(diffs, Change{Op: "insert", Value: newLines[j-1], Line: j})
			j--
		}
	}

	// Handle leftover lines in oldLines (deletions)
	for i > 0 {
		diffs = append(diffs, Change{Op: "delete", Value: oldLines[i-1], Line: i})
		i--
	}

	// Handle leftover lines in newLines (insertions)
	for j > 0 {
		diffs = append(diffs, Change{Op: "insert", Value: newLines[j-1], Line: j})
		j--
	}

	// Reverse the diffs slice since we built it backwards
	for left, right := 0, len(diffs)-1; left < right; left, right = left+1, right-1 {
		diffs[left], diffs[right] = diffs[right], diffs[left]
	}

	return diffs
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

