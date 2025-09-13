# GitComm

GitComm is a CLI tool that uses LLMs to automatically generate meaningful Git commit messages by analyzing your staged changes. It uses OpenRouter to access multiple models with automatic fallback support.

## Features

- 🤖 Uses AI to analyze staged changes and generate commit messages
- ⚡ Powered by OpenRouter with multiple model fallback support
- 🔄 Automatic model fallback: GPT OSS → Llama → Gemini
- 🚀 Auto-commit and push capabilities
- 💻 Cross-platform support (Windows, macOS, Linux)

## Installation

### Using Go Install

```bash
go install github.com/ktappdev/gitcomm@latest
```

### From Releases

Download the appropriate binary for your system from the [releases page](https://github.com/ktappdev/gitcomm/releases).

### Building from Source

```bash
# Clone the repository
git clone https://github.com/ktappdev/gitcomm.git

# Enter the directory
cd gitcomm

# Build
go build

# Or use the build script for all platforms
./build.sh
```

## Setup

You have two options to configure your OpenRouter API key:

1. Interactive Setup (Recommended):
   ```bash
   gitcomm -setup
   ```
   This will guide you through setting up your OpenRouter API key.

2. Environment Variable:
   ```bash
   export OPEN_ROUTER_API_KEY=your_openrouter_api_key
   ```

API keys are stored securely in `~/.gitcomm/config.json`. Environment variables will override stored configuration.

### Getting an OpenRouter API Key

1. Visit [OpenRouter.ai](https://openrouter.ai)
2. Sign up for an account
3. Generate an API key from your dashboard
4. Use the key in the setup process above

## Usage

1. Stage your changes as normal:

```bash
git add .
```

2. Generate a commit message:

```bash
gitcomm
```

3. Auto-commit with the generated message:

```bash
gitcomm -auto
```

4. Auto-commit and push:

```bash
gitcomm -ap
```

## Examples

```bash
# Basic usage - will analyze changes and suggest a commit message
gitcomm

# Stage all changes and generate a commit message
gitcomm -sa

# Stage all changes, generate message, and auto-commit
gitcomm -sa -auto

# Stage all changes, generate message, auto-commit, and push
gitcomm -sa -ap
```

## Configuration

GitComm uses the following defaults:

- LLM Provider: OpenRouter
- Models (with fallback):
  1. `openai/gpt-oss-20b` (Primary)
  2. `meta-llama/llama-4-scout` (Fallback 1)
  3. `google/gemini-2.5-flash-lite` (Fallback 2)
- Max Tokens: 400 (allows for proper Git commit format with subject + body)
- Temperature: 0.7 (balanced between creativity and consistency)
- Diff size limit: 1,500 lines (truncated with note if exceeded)
- Timeout: 30 seconds per model attempt
- Commit Format: Subject line (50-72 chars) + blank line + detailed body

### Model Fallback System

GitComm automatically tries multiple models if one fails:

1. **Primary Model**: `openai/gpt-oss-20b` - Reliable and high quality for most commit messages
2. **Fallback 1**: `meta-llama/llama-4-scout` - Strong performance if GPT OSS is unavailable  
3. **Fallback 2**: `google/gemini-2.5-flash-lite` - Fast and capable final option

Commit messages follow proper Git format with a concise subject line and detailed body explaining the changes. If all models fail, you'll see an error message.

### Example Output

**Basic Usage:**
```
📄 Analyzed 45 lines of diff
🤖 Generating commit message...
⚡ Using GPT OSS 20B

📝 Generated Commit Message:
┌──────────────────────────────────────────────────
Add user authentication system

Implement JWT-based authentication with bcrypt password hashing
for enhanced security. Add middleware for protecting authenticated
routes and validation for email/password requirements.

Updates database schema to include user roles and timestamps for
better user management and audit trails.
└──────────────────────────────────────────────────
```

**With Model Fallback:**
```
📄 Analyzed 127 lines of diff (truncated from 890 lines)
🤖 Generating commit message...
⚡ Using GPT OSS 20B
⚠️  GPT OSS 20B failed, trying next model...
🔄 Falling back to Llama 4 Scout

📝 Generated Commit Message:
┌──────────────────────────────────────────────────
Refactor database connection handling

Replace deprecated connection pooling with modern async patterns.
Improves performance and reduces memory usage under high load.
└──────────────────────────────────────────────────
```

**Auto-commit and Push:**
```
📁 Staging all changes...
✅ All changes staged successfully!
📄 Analyzed 23 lines of diff
🤖 Generating commit message...
⚡ Using GPT OSS 20B

📝 Generated Commit Message:
┌──────────────────────────────────────────────────
Fix bug in user login validation

Correct email format validation regex that was rejecting valid
email addresses with subdomain patterns.
└──────────────────────────────────────────────────

💾 Auto-committing with the generated message...
✅ Changes committed successfully!
🚀 Pushing changes to remote repository...
✅ Changes pushed successfully!
```

## Environment Variables

- `OPEN_ROUTER_API_KEY`: Your OpenRouter API key (required)

## Command Line Flags

- `-auto`: Automatically commit with the generated message
- `-ap`: Automatically commit and push to remote
- `-sa`: Stage all changes before analyzing (equivalent to `git add .`)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

- [OpenRouter](https://openrouter.ai) for their unified LLM API with model fallback support
- The Go community for the excellent tooling

## Troubleshooting

### Common Issues

1. **No API Key Set**

   ```
   OpenRouter API key not set in config file or OPEN_ROUTER_API_KEY environment variable
   ```

   Solution: Set your OPEN_ROUTER_API_KEY environment variable or run `gitcomm -setup`

2. **No Staged Changes**

   ```
   No staged changes. Please stage your changes before running gitcomm.
   ```

   Solution: Stage your changes using `git add`

3. **Push Failed**
   ```
   Error pushing changes
   ```
   Solution: Check your remote repository configuration and permissions

### Getting Help

If you encounter any issues:

1. Check the troubleshooting section above
2. Search existing GitHub issues
3. Create a new issue with:
   - Your OS and version
   - Command used
   - Full error message
   - Steps to reproduce
