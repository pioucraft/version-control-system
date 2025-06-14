package main

import (
	"fmt"
	"os"
	"strings"
)

func Rebuild() error {
	// Rebuild the workspace by reading all keys and their commits
	keys := []string{}
	foldersToNavigate := []string{"./"}
	for i := 0; i < len(foldersToNavigate); i++ {
		folder := foldersToNavigate[i]
		filesOrFolders, err := os.ReadDir(fmt.Sprintf(".vc/keys/%s", folder))
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", fmt.Sprintf(".vc/keys/%s", folder), err)
		}
		for _, fileOrFolder := range filesOrFolders {
			fileOrFolderContent, err := os.ReadDir(fmt.Sprintf(".vc/keys/%s/%s", folder, fileOrFolder.Name()))
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", fmt.Sprintf(".vc/keys/%s/%s", folder, fileOrFolder.Name()), err)
			}
			// if the first folder inside is .commits, then it is a key
			if len(fileOrFolderContent) > 0 && fileOrFolderContent[0].Name() == ".commits" {
				keys = append(keys, fmt.Sprintf("%s/%s", folder, fileOrFolder.Name()))
			} else {
				// if it is not a key, then it is a folder, so add it to the foldersToNavigate
				foldersToNavigate = append(foldersToNavigate, fmt.Sprintf("%s/%s", folder, fileOrFolder.Name()))
			}
		}
	}

	for _, key := range keys {
		lastCat, err := LastCat(key)
		if err != nil {
			return fmt.Errorf("failed to get last cat for key %s: %w", key, err)
		}
		parentDirSlice := strings.Split(key, "/")
		parentDir := strings.Join(parentDirSlice[:len(parentDirSlice)-1], "/")
		os.MkdirAll(parentDir, 0755)
		os.WriteFile(key, []byte(lastCat), 0644)
	}
	return nil
}
