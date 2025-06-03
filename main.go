/*
	Important to do:
	- Add support for deletions
	- STILL EMPTY LINE ISSUE
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
	err := fullCommit("Initial commit")
	if err != nil {
		fmt.Printf("Error during commit: %v\n", err)
		return
	}
}

func cat(key string, commitId string) (string, error) {
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

func lastCat(key string) (string, error) {
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
	cat, err := cat(key, latestCommit)
	return cat, err
}

func fullCommit(message string) error {
	// check for diffs
	commits := []simpleCommitStruct{}
	foldersToNavigate := []string{"./"}
	for _, folder := range foldersToNavigate {
		filesAndFolders, err := os.ReadDir(folder)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", folder, err)
		}
		for _, fileOrFolder := range filesAndFolders {
			if fileOrFolder.IsDir() {
				if fileOrFolder.Name() == ".vc" || fileOrFolder.Name() == ".git" || fileOrFolder.Name() == ".commits" {
					continue
				}
				foldersToNavigate = append(foldersToNavigate, folder+fileOrFolder.Name()+"/")
				continue		
			}
			oldContent := ""
			key := folder + fileOrFolder.Name()
			content, err := os.ReadFile(key)
			
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", key, err)
			}
			contentHash := sha256.Sum256(content)
			if isBinary(content){ 
				// check the hashes
				commitsDir, err := os.ReadDir(fmt.Sprintf(".vc/keys/%s/.commits", key))
				if err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to read commits directory for key %s: %w", key, err)
				}
				hasChanges := true
				for _, commit := range commitsDir {
					if strings.HasPrefix(commit.Name(), "b") {
						commitHash := commit.Name()[1:] // remove the 'b' prefix
						commitHash = strings.Split(commitHash, "+")[1] // split by '+' and take the hash part
						if commitHash == fmt.Sprintf("%x", contentHash) {
							hasChanges = false
							break
						}
					}
				}
				if !hasChanges {
					continue // no changes detected
				}
				commitPath := fmt.Sprintf(".vc/keys/%s/.commits", key)
				err = os.MkdirAll(commitPath, 0755)
				if err != nil && !os.IsExist(err) {
					return fmt.Errorf("failed to create commits directory for key %s: %w", key, err)
				}
				commits = append(commits, simpleCommitStruct{
					Key: key,
					OldContent: "",
					NewContent: "",
					BinaryContent: content, // store binary content directly
				})
				continue
			}
			// get key last commit
			commitPath := fmt.Sprintf(".vc/keys/%s/.commits", key)
			commitFiles, err := os.ReadDir(commitPath)
			if err != nil {
				if err != os.ErrNotExist {
					// create the directory
					err = os.MkdirAll(commitPath, 0755)
				} else {
					return fmt.Errorf("failed to read commits directory for key %s: %w", key, err)
				}
			}
			var latestCommit string
			for _, commitFile := range commitFiles {
				if commitFile.Name() > latestCommit {
					latestCommit = commitFile.Name()
				}
			}
			if latestCommit != "" {
				latestCommitHash := latestCommit[1:]
				latestCommitHash = strings.Split(latestCommitHash, "+")[1]
				if latestCommitHash == fmt.Sprintf("%x", contentHash) {
					continue // no changes detected
				}
				// read the last commit content
				oldContent, err = cat(key, latestCommit)
				if err != nil {
					return fmt.Errorf("failed to read last commit for key %s: %w", key, err)
				}
			} else {
				oldContent = "" // no previous commit found
			}
			commits = append(commits, simpleCommitStruct{	
				Key: key,
				OldContent: oldContent,
				NewContent: string(content),
				BinaryContent: nil, // will be set if the file is binary
			})
		}
	}


	commitIds := []string{}
	for _, commit := range commits {
		if commit.BinaryContent != nil {
			commitId, err := binarySimpleCommit(commit.Key, commit.BinaryContent)
			if err != nil {
				return fmt.Errorf("failed to create binary commit for key %s: %w", commit.Key, err)
			}
			commitId = commit.Key + "/.commits/" + commitId
			commitIds = append(commitIds, commitId)
			continue
		}
		commitId, err := simpleCommit(commit.Key, commit.OldContent, commit.NewContent)
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
	commitFilePath := fmt.Sprintf(".vc/keys/%s/.commits/%s", key, commitId)
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
	diffs := diff(oldContent, newContent)
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
	timestamp := time.Now().Unix()
	timestampString := fmt.Sprintf("%d", timestamp)
	commitId := "d" + timestampString + "+" + hashString 

	commitFilePath := fmt.Sprintf(".vc/keys/%s/.commits/%s", key, commitId)
	err := os.WriteFile(commitFilePath, []byte(stringDiff), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write commit file: %w", err)
	}
	return commitId, nil
}

func diff(oldContent string, newContent string) []Change {
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func isBinary(data []byte) bool {
	maxCheck := 8000
	if len(data) < maxCheck {
		maxCheck = len(data)
	}
	for i := 0; i < maxCheck; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

