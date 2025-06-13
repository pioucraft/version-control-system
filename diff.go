package main

import (
	"strings"
)

func Diff(oldContent string, newContent string) []Change {
	diffs := []Change{}

	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	for i, newLine := range newLines {
		found := false
		for j, oldLine := range oldLines {
			if newLine == oldLine {
				diffs = append(diffs, Change{Op: "=", Value: "", Line: j + 1})
				found = true
				break
			}
		}
		if !found {
			diffs = append(diffs, Change{Op: "+", Value: newLine, Line: i + 1})
		}
	}

	return diffs
}
