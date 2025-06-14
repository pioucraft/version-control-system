package main

import (
	"fmt"
	"os"
	"flag"
	"time"
)

func main() {
	// if the first argument is commmit, then call the commit function with the flag -m/--message for commit message
	if len(os.Args) < 2 || (os.Args[1] == "help") {
		fmt.Println("Usage: vc <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  commit -m <message>        Commit changes with a message")
		fmt.Println("  diff                       Show differences for the current commit")
		fmt.Println("  cat <key> [-c <commitId>]  Read content for a key at a specific commit")
		fmt.Println("  help                       Show this help message")

		return
	}
	if os.Args[1] == "commit" {
		var message string
		flagSet := flag.NewFlagSet("commit", flag.ExitOnError)
		flagSet.StringVar(&message, "m", "", "commit message")
		flagSet.StringVar(&message, "message", "", "commit message")
		if err := flagSet.Parse(os.Args[2:]); err != nil {
			fmt.Println("Error parsing flags:", err)
			return
		}
		if message == "" {
			fmt.Println("Commit message is required")
			return
		}
		err := FullCommit(message)
		if err != nil {
			fmt.Println("Error committing:", err)
			return
		}
		fmt.Println("Commit successful")
		return
	} else if os.Args[1] == "diff" {
		diffs, err := DiffForCommit(time.Now())	
		if err != nil {
			fmt.Println("Error getting diffs:", err)
			return
		}
		if len(diffs) == 0 {
			fmt.Println("No changes to commit")
			return
		}
		PrintDiffs(diffs)
	} else if os.Args[1] == "cat" {
		if len(os.Args) < 3 {
			fmt.Println("Usage : vc cat <key> -c <commitId>")
			return
		}
		key := os.Args[2]
		// make the commitId optional
		var commitId string
		flagSet := flag.NewFlagSet("cat", flag.ExitOnError)
		flagSet.StringVar(&commitId, "c", "", "commit ID to read")
		flagSet.StringVar(&commitId, "commit", "", "commit ID to read")
		if err := flagSet.Parse(os.Args[3:]); err != nil {
			fmt.Println("Error parsing flags:", err)
			return
		}
		if commitId == "" {
			content, err := LastCat(key)
			if err != nil {
				fmt.Println("Error reading last commit:", err)
				return
			}
			fmt.Printf("%s", content)
		} else {
			fmt.Println("Reading commit:", commitId)
			content, err := Cat(key, commitId)
			if err != nil {
				fmt.Println("Error reading commit:", err)
				return
			}
			fmt.Printf("%s", content)
		}

	 } else if os.Args[1] == "rebuild" {
		 err := Rebuild()
		 if err != nil {
			 fmt.Println("Error rebuilding:", err)
			 return
		 }
		 fmt.Println("Rebuild successful")
	 }
}

