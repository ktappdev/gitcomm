package main

import (
	"flag"
	"fmt"
	"github.com/ktappdev/gitcomm/internal/analyzer"
	"github.com/ktappdev/gitcomm/internal/git"
)

func main() {
	// Parse command-line flags
	autoFlag := flag.Bool("auto", false, "Automatically commit with the generated message")
	autoPushFlag := flag.Bool("ap", false, "Automatically commit and push with the generated message")
	flag.Parse()

	// Get the git diff of staged changes
	diff, err := git.GetStagedChanges()
	if err != nil {
		fmt.Println("Error getting git diff:", err)
		return
	}

	// Check if there are any staged changes
	if diff == "" {
		fmt.Println("No staged changes. Please stage your changes before running gitcomm.")
		return
	}

	fmt.Println("Analyzing changes and generating commit message...")
	commitMessage, err := analyzer.AnalyzeChanges(diff)
	if err != nil {
		fmt.Println("Error analyzing changes:", err)
		return
	}

	fmt.Println("Generated Commit Message:")
	fmt.Println(commitMessage)

	// Handle auto-commit and auto-push if the flags are set
	if *autoFlag || *autoPushFlag {
		if commitMessage == "" {
			fmt.Println("Error: Could not extract a commit message from the analysis.")
			return
		}
		fmt.Println("Auto-committing with the generated message...")
		err = git.Commit(commitMessage)
		if err != nil {
			fmt.Println("Error committing:", err)
			return
		}
		fmt.Println("Changes committed successfully!")

		if *autoPushFlag {
			fmt.Println("Pushing changes to remote repository...")
			err = git.Push()
			if err != nil {
				fmt.Println("Error pushing changes:", err)
				return
			}
			fmt.Println("Changes pushed successfully!")
		}
	}
}
