package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ktappdev/gitcomm/internal/analyzer"
	"github.com/ktappdev/gitcomm/internal/config"
	"github.com/ktappdev/gitcomm/internal/diag"
	"github.com/ktappdev/gitcomm/internal/git"
)

const updateModule = "github.com/ktappdev/gitcomm@latest"

var (
	debug       = false
	execCommand = exec.Command
)

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
	setModelFlag := flag.String("set-model", "", "Set model at position (format: position:provider/model-name)")
	flag.Parse()

	debug = *debugFlag
	if path, err := diag.Init(debug); err == nil {
		diag.Info("main", "startup", "args", strings.Join(os.Args[1:], " "), "log_path", path)
	} else if debug {
		fmt.Printf("warning: failed to initialize diagnostics logging: %v\n", err)
	}
	logf("startup: flags setup=%v auto=%v ap=%v sa=%v debug=%v", *setupFlag, *autoFlag, *autoPushFlag, *stageAllFlag, *debugFlag)

	if flag.NArg() > 0 {
		switch flag.Arg(0) {
		case "update":
			if err := runSelfUpdate(); err != nil {
				diag.Error("main", "self-update failed", "error", err)
				fmt.Printf("❌ Update failed: %v\n", err)
				if diag.Path() != "" {
					fmt.Printf("   Diagnostics log: %s\n", diag.Path())
				}
				return
			}
			fmt.Println("✅ GitComm updated successfully.")
			fmt.Println("   This updates Go-installed copies of GitComm.")
			return
		default:
			fmt.Printf("❌ Unknown command: %s\n", flag.Arg(0))
			printHelp()
			return
		}
	}

	if *setupFlag {
		logf("runSetup: starting interactive setup")
		if err := runSetup(); err != nil {
			diag.Error("main", "setup failed", "error", err)
			fmt.Println("Setup failed:", err)
			printHelp()
			return
		}
		fmt.Println("Setup completed successfully!")
		return
	}

	if *setModelFlag != "" {
		handleSetModel(*setModelFlag)
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
		diag.Error("main", "failed to get staged changes", "error", err)
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
		diag.Warn("main", "no staged changes found")
		fmt.Println("⚠️  No staged changes. Please stage your changes before running gitcomm.")
		printHelp()
		return
	}

	logf("analyzer.AnalyzeChanges: begin")
	commitMessage, err := analyzer.AnalyzeChanges(diff)
	if err != nil {
		diag.Error("main", "analysis failed", "error", err)
		fmt.Printf("❌ Error analyzing changes: %v\n", err)
		if diag.Path() != "" {
			fmt.Printf("   Diagnostics log: %s\n", diag.Path())
		}
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

func runSelfUpdate() error {
	fmt.Println("⬆️  Updating GitComm via Go...")
	fmt.Printf("   Running: go install %s\n", updateModule)

	cmd := execCommand("go", "install", updateModule)
	output, err := cmd.CombinedOutput()
	outputText := strings.TrimSpace(string(output))
	if err != nil {
		if errorsIsExecNotFound(err) {
			return fmt.Errorf("Go is not installed or not on PATH; install Go to use `gitcomm update`")
		}
		if outputText != "" {
			return fmt.Errorf("go install failed: %s", outputText)
		}
		return fmt.Errorf("go install failed: %w", err)
	}
	if outputText != "" {
		diag.Info("main", "self-update command output", "output", diag.Snippet(outputText, 300))
	}
	return nil
}

func errorsIsExecNotFound(err error) bool {
	var execErr *exec.Error
	if errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound {
		return true
	}
	return errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist)
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

func handleSetModel(arg string) {
	parts := strings.SplitN(arg, ":", 2)
	if len(parts) != 2 {
		fmt.Println("❌ Error: Invalid format, expected position:model-name")
		fmt.Println("Example: 1:openai/gpt-4o-mini")
		fmt.Println("Example: 2:meta-llama/llama-3.3-8b-instruct:free")
		return
	}

	positionStr := strings.TrimSpace(parts[0])
	position, err := strconv.Atoi(positionStr)
	if err != nil {
		fmt.Printf("❌ Error: Invalid position '%s': must be an integer\n", positionStr)
		return
	}

	modelName := strings.TrimSpace(parts[1])
	if modelName == "" {
		fmt.Println("❌ Error: Model name cannot be empty")
		return
	}
	if err := config.ValidateModelName(modelName); err != nil {
		fmt.Printf("❌ Error: Invalid model name: %v\n", err)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("❌ Error: Failed to load configuration: %v\n", err)
		if diag.Path() != "" {
			fmt.Printf("   Diagnostics log: %s\n", diag.Path())
		}
		return
	}

	maxPosition := len(cfg.Models) + 1
	if position < 1 || position > maxPosition {
		fmt.Printf("❌ Error: Invalid position: %d\n", position)
		fmt.Printf("Valid positions are 1 to %d\n", maxPosition)
		return
	}

	if position <= len(cfg.Models) {
		cfg.Models[position-1] = modelName
		fmt.Printf("Updated model at position %d (primary = 1) to: %s\n", position, modelName)
	} else {
		cfg.Models = append(cfg.Models, modelName)
		fmt.Printf("Added new model at position %d: %s\n", position, modelName)
	}

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("❌ Error: Failed to save configuration: %v\n", err)
		return
	}
	fmt.Println("✅ Configuration saved successfully!")
}

func printHelp() {
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gitcomm [flags]")
	fmt.Println("  gitcomm update")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -setup      Run interactive setup to configure OpenRouter API key and defaults")
	fmt.Println("  -sa         Stage all changes before analyzing")
	fmt.Println("  -auto       Generate a commit message and auto-commit with it")
	fmt.Println("  -ap         Generate, auto-commit, and push to remote")
	fmt.Println("  -debug      Enable verbose debug logging")
	fmt.Println("  -set-model  Set model at position (format: position:provider/model-name)")
	fmt.Println("               Position: 1 = primary, 2 = first fallback, etc.")
	fmt.Println("               Example: 1:openai/gpt-4o-mini")
	fmt.Println("               Example: 2:meta-llama/llama-3.3-8b-instruct:free")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  update      Install the latest GitComm with `go install github.com/ktappdev/gitcomm@latest`")
	fmt.Println("              Only works for Go-installed copies of GitComm and requires `go` on PATH")
	fmt.Println()
	fmt.Println("Common examples:")
	fmt.Println("  gitcomm")
	fmt.Println("  gitcomm -sa")
	fmt.Println("  gitcomm -sa -auto")
	fmt.Println("  gitcomm -sa -ap")
	fmt.Println("  gitcomm update")
}
