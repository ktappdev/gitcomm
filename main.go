package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ktappdev/gitcomm/internal/analyzer"
	"github.com/ktappdev/gitcomm/internal/config"
	"github.com/ktappdev/gitcomm/internal/git"
)

var debug = false

func logf(format string, args ...any) {
	if debug {
		fmt.Printf(format+"\n", args...)
	}
}

func main() {
	setupFlag := flag.Bool("setup", false, "Run interactive setup to configure API keys")
	autoFlag := flag.Bool("auto", false, "Automatically commit with the generated message")
	autoPushFlag := flag.Bool("ap", false, "Automatically commit and push with the generated message")
	stageAllFlag := flag.Bool("sa", false, "Stage all changes before analyzing")
	debugFlag := flag.Bool("debug", false, "Enable verbose debug logging")
	flag.Parse()

	debug = *debugFlag
	logf("startup: flags setup=%v auto=%v ap=%v sa=%v debug=%v", *setupFlag, *autoFlag, *autoPushFlag, *stageAllFlag, *debugFlag)

	if *setupFlag {
		logf("runSetup: starting interactive setup")
		if err := runSetup(); err != nil {
			fmt.Println("Setup failed:", err)
			return
		}
		fmt.Println("Setup completed successfully!")
		return
	}

	if *stageAllFlag {
		fmt.Println("Staging all changes...")
		logf("git.StageAll: invoking")
		if err := git.StageAll(); err != nil {
			fmt.Println("Error staging changes:", err)
			return
		}
		fmt.Println("All changes staged successfully!")
	}

	logf("git.GetStagedChanges: fetching staged diff")
	diff, err := git.GetStagedChanges()
	if err != nil {
		fmt.Println("Error getting git diff:", err)
		return
	}
	logf("git.GetStagedChanges: got %d bytes", len(diff))

	if diff == "" {
		fmt.Println("No staged changes. Please stage your changes before running gitcomm.")
		return
	}

	fmt.Println("Analyzing changes and generating commit message...")
	logf("analyzer.AnalyzeChanges: begin")
	commitMessage, err := analyzer.AnalyzeChanges(diff)
	if err != nil {
		fmt.Println("Error analyzing changes:", err)
		return
	}
	logf("analyzer.AnalyzeChanges: result length=%d", len(commitMessage))

	fmt.Println("Generated Commit Message:")
	fmt.Println(commitMessage)

	if *autoFlag || *autoPushFlag {
		if commitMessage == "" {
			fmt.Println("Error: Could not extract a commit message from the analysis.")
			return
		}
		fmt.Println("Auto-committing with the generated message...")
		logf("git.Commit: committing")
		err = git.Commit(commitMessage)
		if err != nil {
			fmt.Println("Error committing:", err)
			return
		}
		fmt.Println("Changes committed successfully!")

		if *autoPushFlag {
			fmt.Println("Pushing changes to remote repository...")
			logf("git.Push: pushing")
			err = git.Push()
			if err != nil {
				fmt.Println("Error pushing changes:", err)
				return
			}
			fmt.Println("Changes pushed successfully!")
		}
	}
}

func runSetup() error {
	cfg := &config.Config{}

	fmt.Println("Welcome to GitComm Setup!")
	fmt.Println("Press Enter to skip any provider you don't want to configure.")
	fmt.Println()

	fmt.Print("Enter Gemini API key (recommended): ")
	fmt.Scanln(&cfg.GeminiAPIKey)
	logf("setup: gemini set=%v", cfg.GeminiAPIKey != "")

	fmt.Print("Enter Groq API key (optional fallback): ")
	fmt.Scanln(&cfg.GroqAPIKey)
	logf("setup: groq set=%v", cfg.GroqAPIKey != "")

	fmt.Print("Enter OpenAI API key (optional fallback): ")
	fmt.Scanln(&cfg.OpenAIAPIKey)
	logf("setup: openai set=%v", cfg.OpenAIAPIKey != "")

	if cfg.GeminiAPIKey == "" && cfg.GroqAPIKey == "" && cfg.OpenAIAPIKey == "" {
		return fmt.Errorf("at least one API key must be provided")
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}

	fmt.Println("\nConfiguration saved successfully!")
	hd, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	fmt.Printf("Config file location: %s/.gitcomm/config.json\n", hd)

	return nil
}
