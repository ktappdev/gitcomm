# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

GitComm is a CLI tool that uses LLMs to automatically generate meaningful Git commit messages by analyzing staged changes. The tool uses OpenRouter to access multiple LLM models with automatic fallback support.

## Development Commands

### Build and Run
```bash
# Build the binary
go build ./...

# Run directly
go run .

# Install as CLI tool globally
go install ./...

# Clean up dependencies
go mod tidy
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Lint (requires golangci-lint)
golangci-lint run

# Install golangci-lint if needed
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Testing
```bash
# Run all tests
go test ./...

# Test specific package
go test ./internal/analyzer -v

# Test single function
go test ./internal/analyzer -run TestName -v
```

## Architecture Overview

### Core Components

**Entry Point**: `main.go`
- CLI flag parsing (`-setup`, `-auto`, `-ap`, `-sa`, `-debug`)
- Orchestrates the entire flow from git diff to commit

**Key Packages**:
- `internal/git/` - Git operations wrapper (stage, diff, commit, push)
- `internal/analyzer/` - Orchestrates LLM analysis of git diffs
- `internal/llm/` - Multi-provider LLM client abstraction
- `internal/config/` - Configuration management for API keys
- `config/` - Constants for API URLs and environment variables

### Data Flow
1. Parse CLI flags and handle setup mode if requested
2. Stage changes if `-sa` flag is used
3. Get staged git diff using `git diff --staged`
4. Send diff to LLM with structured prompt for commit message generation
5. Extract commit message from LLM response
6. Optionally auto-commit (`-auto`) and/or push (`-ap`)

### LLM Provider System

**Provider**: OpenRouter (https://openrouter.ai)
**Model Fallback Chain**:
1. `meta-llama/llama-3.3-8b-instruct:free` (Primary: Free and capable)
2. `meta-llama/llama-4-scout` (Fallback 1: Strong performance)  
3. `google/gemini-2.5-flash-lite` (Fallback 2: Fast and capable)

**Provider Configuration**:
- OpenRouter: `OPEN_ROUTER_API_KEY` environment variable

**LLM Settings** (optimized for proper Git commit format):
- Max tokens: 400
- Temperature: 0.7
- Diff size limit: 1,500 lines (truncated with note if exceeded)
- Timeout: 30 seconds per model attempt
- Commit format: Subject line (50-72 chars) + blank line + detailed body (72 chars/line)

## Configuration System

**Configuration Precedence**:
1. Environment variables (highest priority)
2. `~/.gitcomm/config.json` file

**Setup Options**:
```bash
# Interactive setup (recommended)
gitcomm -setup

# Manual environment setup
export OPEN_ROUTER_API_KEY=your_openrouter_api_key
```

**Config File Format** (`~/.gitcomm/config.json`):
```json
{
    "open_router_api_key": "your_key_here"
}
```

## CLI Usage Patterns

```bash
# Basic usage - analyze staged changes and suggest commit message
gitcomm

# Stage all changes first, then analyze
gitcomm -sa

# Auto-commit with generated message
gitcomm -auto

# Auto-commit and push
gitcomm -ap

# Combined: stage all, auto-commit, and push
gitcomm -sa -ap

# Enable debug logging
gitcomm -debug
```

## Debug and Troubleshooting

### Debug Mode
Use the `-debug` flag to enable verbose logging:
- Flag parsing and startup sequence
- Git command execution and output sizes  
- LLM client initialization and API calls
- Response processing and extraction

### Common Issues

**No API Key Set**:
```
OpenRouter API key not set in config file or OPEN_ROUTER_API_KEY environment variable
```
Solution: Run `gitcomm -setup` or set `OPEN_ROUTER_API_KEY` environment variable

**No Staged Changes**:
```  
No staged changes. Please stage your changes before running gitcomm.
```
Solution: Use `git add` to stage changes, or use `gitcomm -sa` to stage all changes

**LLM Provider Failures**: 
The system automatically tries multiple models in sequence with proper fallback handling

## Code Conventions

### Go Standards
- **Go Version**: 1.23.x with modules enabled (from go.mod)
- **Import Order**: Standard library → External packages → Internal packages
- **Error Handling**: Return `(T, error)` pattern, wrap with context using `fmt.Errorf("...: %w", err)`
- **File Length**: Keep files under 250 lines when practical
- **Formatting**: Use `go fmt`, minimal comments unless necessary

### Project-Specific Patterns
- **CLI Integration**: Shell out to git commands, avoid interactive prompts in library code
- **Security**: Never log API keys, use environment variables or config file
- **Logging**: Use `log/slog` for structured logging in `internal/llm`, simple `fmt.Println` for CLI output
- **Prompt Engineering**: Structure LLM prompts with clear format expectations for single-line commit extraction
- **Debug Output**: All debug messages prefixed with `[debug]` and package name

### LLM Response Processing
The analyzer expects LLM responses in proper Git commit format:
```
Generated Commit Message:
[Subject line - 50-72 characters]

[Detailed body explaining the changes, wrapped at 72 characters.
Include context about what was changed and why it was necessary.]
```

Example:
```
Generated Commit Message:
Add JWT-based user authentication system

Implement comprehensive authentication using JSON Web Tokens with
bcrypt password hashing for enhanced security. Add middleware for
protecting authenticated routes and validation for requirements.
```

If the expected format is not found, falls back to "update" as the commit message.

## Dependencies

**Core Dependencies**:
- Standard library HTTP client for OpenRouter API
- No external LLM-specific dependencies required

**Go Version**: Requires Go 1.23.2+ (specified in go.mod)

## References

- [README.md](README.md) - Installation, setup, and basic usage
- [CLAUDE.md](CLAUDE.md) - Detailed architecture and development conventions  
- [CRUSH.md](CRUSH.md) - Quick command reference and project conventions