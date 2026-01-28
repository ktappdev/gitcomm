package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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
			printHelp()
			return
		}
		fmt.Println("Setup completed successfully!")
		return
	}

	if *stageAllFlag {
		fmt.Println("📁 Staging all changes...")
		logf("git.StageAll: invoking")
		if err := git.StageAll(); err != nil {
			if strings.Contains(err.Error(), "not a git repository") {
				fmt.Println("❌ This directory is not a Git repository.")
				fmt.Println("   Run `git init` to create one, or run gitcomm inside an existing repo.")
			} else {
				fmt.Printf("❌ Error staging changes: %v\n", err)
			}
			printHelp()
			return
		}
		fmt.Println("✅ All changes staged successfully!")
	}

	logf("git.GetStagedChanges: fetching staged diff")
	diff, err := git.GetStagedChanges()
	if err != nil {
		if strings.Contains(err.Error(), "not a git repository") {
			fmt.Println("❌ This directory is not a Git repository.")
			fmt.Println("   Run `git init` to create one, or run gitcomm inside an existing repo.")
		} else {
			fmt.Printf("❌ Error getting git diff: %v\n", err)
		}
		printHelp()
		return
	}
	logf("git.GetStagedChanges: got %d bytes", len(diff))

	if diff == "" {
		fmt.Println("⚠️  No staged changes. Please stage your changes before running gitcomm.")
		printHelp()
		return
	}

	logf("analyzer.AnalyzeChanges: begin")
	commitMessage, err := analyzer.AnalyzeChanges(diff)
	if err != nil {
		fmt.Printf("❌ Error analyzing changes: %v\n", err)
		printHelp()
		return
	}
	logf("analyzer.AnalyzeChanges: result length=%d", len(commitMessage))

	fmt.Println("\n📝 Generated Commit Message:")
	fmt.Println("┌" + strings.Repeat("─", 50))
	fmt.Println(commitMessage)
	fmt.Println("└" + strings.Repeat("─", 50))

	if *autoFlag || *autoPushFlag {
		if commitMessage == "" {
			fmt.Println("❌ Error: Could not extract a commit message from the analysis.")
			printHelp()
			return
		}
		fmt.Println("\n💾 Auto-committing with the generated message...")
		logf("git.Commit: committing")
		err = git.Commit(commitMessage)
		if err != nil {
			fmt.Printf("❌ Error committing: %v\n", err)
			printHelp()
			return
		}
		fmt.Println("✅ Changes committed successfully!")

		if *autoPushFlag {
			fmt.Println("🚀 Pushing changes to remote repository...")
			logf("git.Push: pushing")
			err = git.Push()
			if err != nil {
				fmt.Printf("❌ Error pushing changes: %v\n", err)
				printHelp()
				return
			}
			fmt.Println("✅ Changes pushed successfully!")
		}
	}
}

func runSetup() error {
	configPath, err := config.Path()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config file already exists at %s\n", configPath)
		fmt.Print("Overwrite? (y = overwrite, k = keep, b = backup+overwrite): ")

		var choice string
		fmt.Scanln(&choice)
		choice = strings.ToLower(strings.TrimSpace(choice))

		switch choice {
		case "k", "":
			fmt.Println("Keeping existing config. Setup cancelled.")
			return nil
		case "b":
			backupPath := configPath + ".bak"
			if err := os.Rename(configPath, backupPath); err != nil {
				return fmt.Errorf("failed to backup existing config: %w", err)
			}
			fmt.Printf("Backed up existing config to %s\n", backupPath)
		case "y":
			// continue to overwrite
		default:
			fmt.Println("Unrecognized choice. Keeping existing config. Setup cancelled.")
			return nil
		}
	}

	cfg := config.DefaultConfig()

	envKey := os.Getenv(config.OpenRouterAPIKeyEnvPrimary)
	envName := config.OpenRouterAPIKeyEnvPrimary
	if envKey == "" {
		envKey = os.Getenv(config.OpenRouterAPIKeyEnvLegacy)
		envName = config.OpenRouterAPIKeyEnvLegacy
	}
	useEnvKey := false
	if envKey != "" {
		fmt.Printf("Found %s in environment. Use it? (Y/n): ", envName)
		var choice string
		fmt.Scanln(&choice)
		choice = strings.ToLower(strings.TrimSpace(choice))
		if choice == "" || choice == "y" || choice == "yes" {
			useEnvKey = true
		}
	}

	fmt.Println("Welcome to GitComm Setup!")
	fmt.Println("This will configure your OpenRouter API key.")
	fmt.Println()
	if useEnvKey {
		fmt.Printf("Using API key from %s. It will not be stored in config.\n", envName)
		logf("setup: openrouter use_env=true")
	} else {
		fmt.Print("Enter OpenRouter API key: ")
		fmt.Scanln(&cfg.OpenRouterAPIKey)
		logf("setup: openrouter set=%v", cfg.OpenRouterAPIKey != "")

		if cfg.OpenRouterAPIKey == "" {
			return fmt.Errorf("OpenRouter API key is required")
		}
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

func printHelp() {
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gitcomm [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -setup   Run interactive setup to configure OpenRouter API key and defaults")
	fmt.Println("  -sa      Stage all changes before analyzing")
	fmt.Println("  -auto    Generate a commit message and auto-commit with it")
	fmt.Println("  -ap      Generate, auto-commit, and push to remote")
	fmt.Println("  -debug   Enable verbose debug logging for troubleshooting")
	fmt.Println()
	fmt.Println("Common examples:")
	fmt.Println("  gitcomm")
	fmt.Println("  gitcomm -sa")
	fmt.Println("  gitcomm -sa -auto")
	fmt.Println("  gitcomm -sa -ap")
}
