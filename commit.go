package main

import (
	"fmt"
	"os"
	"strings"
	"crypto/sha256"
	"time"
)

// The FullCommit function will iterate over all files, check for changes and create a simple commit for each file that has changed. Then, it will add the commits to the history folder.

func FullCommit(message string) error {
	commits, err := DiffForCommit(timeNow)
	if err != nil {
		return fmt.Errorf("failed to get diffs for commit: %w", err)
	}

	// create the deletedKeys array from the commitIds
	deletedKeys := []string{}
	for _, commit := range commits {
		if commit.NewContent == "" {
			deletedKeys = append(deletedKeys, commit.Key)
		}
	}
	newCommits := []SimpleCommitStruct{}
	for _, commit := range commits {
		if (commit.NewContent != "" && commit.BinaryContent == nil) {
			newCommits = append(newCommits, commit)
		}
	}
	commits = newCommits

	commitIds := []string{}

	for _, key := range deletedKeys {
		os.MkdirAll(fmt.Sprintf(".vc/deleted/%d", timeNow.Unix()), 0755)
		keyParentFolder := strings.TrimSuffix(key, "/"+strings.Split(key, "/")[len(strings.Split(key, "/"))-1])
		err = os.MkdirAll(fmt.Sprintf(".vc/deleted/%d/%s", timeNow.Unix(), keyParentFolder), 0755)
		err := os.Rename(fmt.Sprintf(".vc/keys/%s", key), fmt.Sprintf(".vc/deleted/%d/%s", timeNow.Unix(), key))
		if err != nil {
			return fmt.Errorf("failed to move deleted key %s: %w", key, err)
		}

		commitId := fmt.Sprintf("d%d+%s", timeNow.Unix(), "deleted")
		commitId = key + "/.commits/" + commitId 
		commitIds = append(commitIds, commitId)
	}

	// STEP 4. Finally, it will create a commit for each file that has changed.
	for _, commit := range commits {
		if commit.BinaryContent != nil {
			commitId, err := BinarySimpleCommit(commit.Key, commit.BinaryContent)
			if err != nil {
				return fmt.Errorf("failed to create binary commit for key %s: %w", commit.Key, err)
			}
			commitId = commit.Key + "/.commits/" + commitId
			commitIds = append(commitIds, commitId)
			continue
		}
		commitId, err := SimpleCommit(commit.Key, commit.OldContent, commit.NewContent)
		if err != nil {
			return fmt.Errorf("failed to create commit for key %s: %w", commit.Key, err)
		}
		commitId = commit.Key + "/.commits/" + commitId
		commitIds = append(commitIds, commitId)
	}
	if len(commitIds) == 0 {
		return fmt.Errorf("no commits created, check if there are changes")
	}
	numberOfLines := len(strings.Split(message, "\n"))
	commitContent := fmt.Sprintf("%d\n%s\n%s", numberOfLines, message, strings.Join(commitIds, "\n"))
	path := fmt.Sprintf(".vc/history/%d", timeNow.Unix())
	err = os.WriteFile(path, []byte(commitContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write commit history file: %w", err)
	}
	return nil
}

// these are the simple commit functions. They are used for single files

func BinarySimpleCommit(key string, content []byte) (string, error) {
	hash := sha256.Sum256(content)
	hashString := fmt.Sprintf("%x", hash)
	timestamp := timeNow.Unix()
	timestampString := fmt.Sprintf("%d", timestamp)
	commitId := "b" + timestampString + "+" + hashString 
	commitFilePath := fmt.Sprintf(".vc/keys/%s/.commits/%s", key, commitId)
	if err := os.MkdirAll(fmt.Sprintf(".vc/keys/%s/.commits", key), 0755); err != nil {
		return "", fmt.Errorf("failed to create commits directory for key %s: %w", key, err)
	}
	err := os.WriteFile(commitFilePath, content, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write binary commit file: %w", err)
	}
	return commitId, nil
}

func SimpleCommit(key string, oldContent string, newContent string) (string, error) {
	if oldContent == newContent {
		return "", fmt.Errorf("no changes detected, commit not created")
	}
	diffs := Diff(oldContent, newContent)
	stringDiff := "" 
	for _, change := range diffs {
		switch change.Op {
		case "=":
			stringDiff += fmt.Sprintf("=%d\n", change.Line)
		case "+":
			stringDiff += fmt.Sprintf("+%s\n", change.Value)
		}
	}

	hash := sha256.Sum256([]byte(newContent))
	hashString := fmt.Sprintf("%x", hash)
	timestamp := timeNow.Unix()
	timestampString := fmt.Sprintf("%d", timestamp)
	commitId := "d" + timestampString + "+" + hashString 

	commitFilePath := fmt.Sprintf(".vc/keys/%s/.commits/%s", key, commitId)
	if err := os.MkdirAll(fmt.Sprintf(".vc/keys/%s/.commits", key), 0755); err != nil {
		return "", fmt.Errorf("failed to create commits directory for key %s: %w", key, err)
	}
	err := os.WriteFile(commitFilePath, []byte(stringDiff), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write commit file: %w", err)
	}
	return commitId, nil
}

var timeNow time.Time

func init() {
	timeNow = time.Now() 
}
