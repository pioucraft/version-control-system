/*
	For CLI usage. I don't know why anyone would use this for CLI usage, but here it is. :
	<program> commit --key/-k <key> --old-content/-o <oldContent> --new-content/-n <newContent>
	<program> cat --key/-k <key> --commit-id/-c <optional: commitId>
	<program> diff --old-content/-o <oldContent> --new-content/-n <newContent>
*/
package main

import (
	"strings"
	"fmt"
	"crypto/sha256"
	"time"
	"os"
	"sort"
	"strconv"
)

type Change struct {
	Op string
	Value string
	Line int
}

type simpleCommitStruct struct {
	Key string
	OldContent string
	NewContent string
	BinaryContent []byte // Optional, for binary commits
}

func main() {
	
}

func Cat(key string, commitId string) (string, error) {
	fileslistPath := fmt.Sprintf(".vc/keys/%s/commits", key)
	fileslist, err := os.ReadDir(fileslistPath)
	if err != nil {
		return "", fmt.Errorf("failed to read commits directory for key %s: %w", key, err)
	}
	var sortedFiles []string
	for _, file := range fileslist {
		sortedFiles = append(sortedFiles, file.Name())
	}
	sortedFiles = sort.StringSlice(sortedFiles) 
	content := []string{""}
	for _, file := range sortedFiles {
		if strings.HasPrefix(file, "b") {
			return "", fmt.Errorf("binary commits are not supported for cat operation")
		}
		breakAfter := file == commitId
		diffPath := fmt.Sprintf(".vc/keys/%s/commits/%s", key, file)
		diff, err := os.ReadFile(diffPath)
		if err != nil {
			return "", fmt.Errorf("failed to read commit file %s for key %s: %w", file, key, err)
		}
		lines := strings.Split(string(diff), "\n")
		newContent := []string{}
		for _, line := range lines {
			if line == "" {
				continue
			}
			op := line[:1]
			if op == "=" {
				lineToAdd := line[1:]
				if lineToAdd != "" {
					lineToAddInt, err := strconv.Atoi(lineToAdd)
					if err != nil {
						return "", fmt.Errorf("invalid line number in commit diff: %s", lineToAdd)
					}
					newContent = append(newContent, content[lineToAddInt-1])	
				}
			} else if op == "+" {
				lineToAdd := line[1:]
				if lineToAdd != "" {
					newContent = append(newContent, lineToAdd)
				}
			}
		}
		content = newContent
		if breakAfter {
			break
		}
	}
	return strings.Join(content, "\n"), nil
}

func LastCat(key string) (string, error) {
	filesList, err := os.ReadDir(fmt.Sprintf(".vc/keys/%s/commits", key))
	if err != nil {
		return "", fmt.Errorf("failed to read commits directory for key %s: %w", key, err)
	}
	if len(filesList) == 0 {
		return "", fmt.Errorf("no commits found for key %s", key)
	}
	var latestCommit string
	for _, file := range filesList {
		if latestCommit == "" || file.Name() > latestCommit {
			latestCommit = file.Name()
		}
	}
	cat, err := Cat(key, latestCommit)
	return cat, err
}

func FullCommit(commits []simpleCommitStruct, message string) error {
	commitIds := []string{}
	for _, commit := range commits {
		if commit.BinaryContent != nil {
			commitId, err := binarySimpleCommit(commit.Key, commit.BinaryContent)
			if err != nil {
				return fmt.Errorf("failed to create binary commit for key %s: %w", commit.Key, err)
			}
			commitIds = append(commitIds, commitId)
			continue
		}
		commitId, err := simpleCommit(commit.Key, commit.OldContent, commit.NewContent)
		if err != nil {
			return fmt.Errorf("failed to create commit for key %s: %w", commit.Key, err)
		}
		commitId = commit.Key + "/" + commitId
		commitIds = append(commitIds, commitId)
	}
	if len(commitIds) == 0 {
		return fmt.Errorf("no commits created, check if there are changes")
	}
	numberOfLines := len(strings.Split(message, "\n"))
	commitContent := fmt.Sprintf("%d\n%s\n%s", numberOfLines, message, strings.Join(commitIds, "\n"))
	path := fmt.Sprintf(".vc/history/%d", time.Now().Unix())
	err := os.WriteFile(path, []byte(commitContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write commit history file: %w", err)
	}
	return nil
}

func binarySimpleCommit(key string, content []byte) (string, error) {
	hash := sha256.Sum256(content)
	hashString := fmt.Sprintf("%x", hash)
	timestamp := time.Now().Unix()
	timestampString := fmt.Sprintf("%d", timestamp)
	commitId := "b" + timestampString + "+" + hashString 
	commitFilePath := fmt.Sprintf(".vc/keys/%s/commits/%s", key, commitId)
	err := os.WriteFile(commitFilePath, content, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write binary commit file: %w", err)
	}
	return commitId, nil
}

func simpleCommit(key string, oldContent string, newContent string) (string, error) {
	if oldContent == newContent {
		return "", fmt.Errorf("no changes detected, commit not created")
	}
	diffs := Diff(oldContent, newContent)
	stringDiff := "" 
	for _, change := range diffs {
		switch change.Op {
		case "equal":
			stringDiff += fmt.Sprintf("=%d\n", change.Line)
		case "insert":
			stringDiff += fmt.Sprintf("+%s\n", change.Value)
		case "delete":
			stringDiff += fmt.Sprintf("-%d\n", change.Line)
		}
	}

	hash := sha256.Sum256([]byte(newContent))
	hashString := fmt.Sprintf("%x", hash)
	timestamp := time.Now().Unix()
	timestampString := fmt.Sprintf("%d", timestamp)
	commitId := "d" + timestampString + "+" + hashString 

	commitFilePath := fmt.Sprintf(".vc/keys/%s/commits/%s", key, commitId)
	err := os.WriteFile(commitFilePath, []byte(stringDiff), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write commit file: %w", err)
	}
	return commitId, nil
}

func Diff(oldContent string, newContent string) []Change {
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
			diffs = append(diffs, Change{Op: "insert", Value: newLines[j-1], Line: i})
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
		diffs = append(diffs, Change{Op: "insert", Value: newLines[j-1], Line: i})
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

