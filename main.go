package main

import (
	"fmt"
	"os"
	"flag"
)

func main() {
	// if the first argument is commmit, then call the commit function with the flag -m/--message for commit message
	if len(os.Args) < 2 {
		fmt.Println("Usage: vc commit -m <message>")
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
	}
}

