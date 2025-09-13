# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GitComm is a CLI tool that uses LLMs to automatically generate meaningful Git commit messages by analyzing staged changes. The tool uses OpenRouter to access multiple LLM models with automatic fallback support.

## Development Commands

### Build and Run
- **Build**: `go build ./...`
- **Run**: `go run .`
- **Install CLI**: `go install ./...`
- **Tidy dependencies**: `go mod tidy`

### Code Quality
- **Lint**: `golangci-lint run` (install with `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- **Format**: `go fmt ./...`
- **Vet**: `go vet ./...`

### Testing
- **All tests**: `go test ./...`
- **Package tests**: `go test ./internal/analyzer -v`
- **Single test**: `go test ./internal/analyzer -run TestName -v`

## Architecture

### Core Components

1. **main.go** - Entry point handling CLI flags and orchestration
2. **internal/analyzer** - Orchestrates LLM analysis of git diffs
3. **internal/llm** - LLM client abstraction supporting multiple providers
4. **internal/config** - Configuration management for API keys
5. **internal/git** - Git operations wrapper

### Data Flow

1. Main parses CLI flags (`-setup`, `-auto`, `-ap`, `-sa`, `-debug`)
2. Git operations get staged changes or stage all files
3. Analyzer sends diff to LLM with structured prompt
4. LLM response is parsed to extract commit message
5. Optionally commits and/or pushes changes

### Configuration System

API keys are loaded in this order (environment variables override file):
1. `~/.gitcomm/config.json` file
2. Environment variables: `GEMINI_API_KEY`, `GROQ_API_KEY`, `OPENAI_API_KEY`

### LLM Provider Architecture

- **Primary**: Gemini (gemini-1.5-flash)
- **Fallback**: Groq (llama-3.1-70b-versatile) or OpenAI (gpt-4o-mini)
- **Settings**: Max tokens: 400, Temperature: 0.7 (optimized for proper Git commit format)

## Code Conventions

### Go Standards
- **Go version**: 1.23.x with modules enabled
- **Imports**: Standard library, then external, then internal packages
- **Formatting**: Use `go fmt`, keep files under 250 lines when practical
- **Error handling**: Return `(T, error)`, wrap with context using `fmt.Errorf("...: %w", err)`
- **Logging**: Use `log/slog` for structured logging in `internal/llm`, simple `fmt.Println` for CLI output

### Project-Specific Patterns
- **CLI flags**: Maintain backwards compatibility with existing flags
- **Git operations**: Shell out to git commands, avoid interactive prompts in library code
- **API keys**: Never log secrets, use environment variables or config file
- **Prompt engineering**: Structure LLM prompts for single-line commit message extraction
- **Fallback logic**: Gemini is always available as fallback when other providers fail

## Debug Mode

Use `-debug` flag to enable verbose logging that shows:
- Flag parsing and startup sequence
- Git command execution and output sizes
- LLM client initialization and API calls
- Response processing and extraction

## Testing Strategy

The project focuses on integration testing with real APIs due to the nature of LLM interactions. When adding new LLM providers or prompt formats, test with actual API calls rather than mocks.