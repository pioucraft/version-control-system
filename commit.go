package main

import (
	"fmt"
	"os"
	"strings"
	"crypto/sha256"
	"time"
	"slices"
)

// The FullCommit function will iterate over all files, check for changes and create a simple commit for each file that has changed. Then, it will add the commits to the history folder.

func FullCommit(message string) error {
	// STEP 1. It's starting by reading the .vc/keys directory to get all known keys. If a key is in the .vc/keys directory, but not in the working directory, it means it has been deleted.
	keys, err := os.ReadDir(".vc/keys")
	if err != nil {
		return fmt.Errorf("failed to read .vc/keys directory: %w", err)
	}

	knownKeys := []string{}
	for _, key := range keys {
		knownKeys = append(knownKeys, "./"+key.Name())
	}

	keysWithCommits := []string{}

	for ki := 0; ki < len(knownKeys); ki++ {
		key := knownKeys[ki]
		// if the folder has a .commit folder, skip it
		if _, err := os.Stat(fmt.Sprintf(".vc/keys/%s/.commits", key)); err == nil {
			keysWithCommits = append(keysWithCommits, key)
			continue // skip this key, it has a .commits folder
		}
		// then add every children of the key folder to the keys to itterate over
		filesAndFolders, err := os.ReadDir(fmt.Sprintf(".vc/keys/%s", key))
		if err != nil {
			return fmt.Errorf("failed to read directory .vc/keys/%s: %w", key, err)
		}
		for _, fileOrFolder := range filesAndFolders {
			knownKeys = append(knownKeys, key + "/" + fileOrFolder.Name())
		}
	}

	knownKeys = keysWithCommits 

	foundKeys := []string{}

	// STEP 2. after that, it will iterate over all files in the working directory and check if they have changed. If they have changed, it will create a simple commit for each file.
	commits := []SimpleCommitStruct{}
	// this is the list of folders to navigate to find files
	foldersToNavigate := []string{"./"}
	for folderi := 0; folderi < len(foldersToNavigate); folderi++ { 
		folder := foldersToNavigate[folderi]
		filesAndFolders, err := os.ReadDir(folder)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", folder, err)
		}
		for _, fileOrFolder := range filesAndFolders {
			// if it's a directory, add it to the folders to navigate.
			if fileOrFolder.IsDir() {
				if fileOrFolder.Name() == ".vc" || fileOrFolder.Name() == ".git" || fileOrFolder.Name() == ".commits" {
					continue
				}
				foldersToNavigate = append(foldersToNavigate, folder+fileOrFolder.Name()+"/")
				continue		
			}
			// if it's a file, check if it has changed
			oldContent := ""
			key := folder + fileOrFolder.Name()
			content, err := os.ReadFile(key)
			foundKeys = append(foundKeys, key)
			
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", key, err)
			}
			// generate the hash of the content
			contentHash := sha256.Sum256(content)
			
			if IsBinary(content){ 
				// check the hashes
				commitsDir, err := os.ReadDir(fmt.Sprintf(".vc/keys/%s/.commits", key))
				if err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to read commits directory for key %s: %w", key, err)
				}
				// find the last commit hash
				hasChanges := true
				for _, commit := range commitsDir {
					commitHash := commit.Name()[1:] 
					commitHash = strings.Split(commitHash, "+")[1] // split by '+' and take the hash part
					if commitHash == fmt.Sprintf("%x", contentHash) {
						hasChanges = false
						break
					}
				}
				if !hasChanges {
					continue // no changes detected
				}
				// add a new simple commit
				commitPath := fmt.Sprintf(".vc/keys/%s/.commits", key)
				err = os.MkdirAll(commitPath, 0755)
				if err != nil && !os.IsExist(err) {
					return fmt.Errorf("failed to create commits directory for key %s: %w", key, err)
				}
				commits = append(commits, SimpleCommitStruct{
					Key: key,
					OldContent: "",
					NewContent: "",
					BinaryContent: content, // store binary content directly
				})
				continue
			}
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
				oldContent, err = Cat(key, latestCommit)
				if err != nil {
					return fmt.Errorf("failed to read last commit for key %s: %w", key, err)
				}
			} else {
				oldContent = "" // no previous commit found
			}
			commits = append(commits, SimpleCommitStruct{	
				Key: key,
				OldContent: oldContent,
				NewContent: string(content),
				BinaryContent: nil, // will be set if the file is binary
			})
		}
	}

	// a list of commit IDs to be created
	commitIds := []string{}

	// STEP 3. Now, it will check if there are any keys that have been deleted.
	for _, key := range knownKeys {
		if !slices.Contains(foundKeys, key) {
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
