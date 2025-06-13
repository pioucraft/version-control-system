package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func Cat(key string, commitId string) (string, error) {
	fileslistPath := fmt.Sprintf(".vc/keys/%s/.commits", key)
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
			continue
		}
		breakAfter := file == commitId
		diffPath := fmt.Sprintf(".vc/keys/%s/.commits/%s", key, file)
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
	filesList, err := os.ReadDir(fmt.Sprintf(".vc/keys/%s/.commits", key))
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
