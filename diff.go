package main

import (
	"strings"
	"fmt"
	"os"
	"crypto/sha256"
	"time"
	"slices"
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

func DiffForCommit(timeNow time.Time) ([]SimpleCommitStruct, error) {
	// STEP 1. It's starting by reading the .vc/keys directory to get all known keys. If a key is in the .vc/keys directory, but not in the working directory, it means it has been deleted.
	keys, err := os.ReadDir(".vc/keys")
	if err != nil {
		return nil, fmt.Errorf("failed to read .vc/keys directory: %w", err)
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
			return nil, fmt.Errorf("failed to read directory .vc/keys/%s: %w", key, err)
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
			return nil, fmt.Errorf("failed to read directory %s: %w", folder, err)
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
				return nil, fmt.Errorf("failed to read file %s: %w", key, err)
			}
			if fmt.Sprintf("%x", content) == "" {
				fmt.Printf("Skipping empty file %s\n", key)
				continue // skip empty files
			}
			// generate the hash of the content
			contentHash := sha256.Sum256(content)
			
			if IsBinary(content){ 
				// check the hashes
				commitsDir, err := os.ReadDir(fmt.Sprintf(".vc/keys/%s/.commits", key))
				if err != nil && !os.IsNotExist(err) {
					return nil, fmt.Errorf("failed to read commits directory for key %s: %w", key, err)
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
			if err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read commits directory for key %s: %w", key, err)
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
					return nil, fmt.Errorf("failed to read last commit for key %s: %w", key, err)
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


	// STEP 3. Now, it will check if there are any keys that have been deleted.

	// start by creating an array with the deleted keys
	for _, key := range knownKeys {
		if !slices.Contains(foundKeys, key) {
			// if the key is not found in the working directory, it means it has been deleted
			commits = append(commits, SimpleCommitStruct{
				Key: key,
				OldContent: "",
				NewContent: "",
				BinaryContent: nil, // no binary content for deleted keys
			})
		}
	}
	return commits, nil
}

func PrintDiffs(diffs []SimpleCommitStruct) {
	for _, diff := range diffs {
		if diff.BinaryContent != nil {
			fmt.Printf("Binary file %s has changed.\n", diff.Key)
			continue
		} else if diff.OldContent == "" && diff.NewContent != "" {
			fmt.Printf("New file %s added.\n", diff.Key)
			continue
		} else if diff.NewContent == "" {
			fmt.Printf("File %s deleted.\n", diff.Key)
			continue
		} else {
			fmt.Printf("Changes in file %s.\n", diff.Key)
		}
	}
}
