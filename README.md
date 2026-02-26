# GitComm

GitComm is a CLI tool that uses LLMs to automatically generate meaningful Git commit messages by analyzing your staged changes. It uses OpenRouter to access multiple models with automatic fallback support.

GitComm also aims to be free-friendly: by default it uses OpenRouter models that are typically available on a free tier. The first two default models (`meta-llama/llama-3.3-8b-instruct:free` and `meta-llama/llama-4-scout`) are usually free on OpenRouter; the third model is a stronger fallback. You only need an OpenRouter API key (a free key works as long as those models remain free), and you can always update `~/.gitcomm/config.json` later to point at other free models if pricing changes.

## Features

- 🤖 Uses AI to analyze staged changes and generate commit messages
- ⚡ Powered by OpenRouter with multiple model fallback support
- 🔄 Automatic model fallback: Llama 3.3 → Llama 4 → Gemini
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
   This collects your OpenRouter API key and seeds `~/.gitcomm/config.json` with the default LLM models, token/temperature settings, API URL, and timeout so you can edit them later. If `OPEN_ROUTER_API_KEY` is already set in your environment, setup will offer to use it without storing the key in the config file.
   If a config file already exists, setup will prompt you to keep it, overwrite it, or back it up before overwriting.

2. Environment Variable (set `OPENROUTER_API_KEY` in your shell; `OPEN_ROUTER_API_KEY` is also supported):
   ```bash
   export OPENROUTER_API_KEY=your_openrouter_api_key
   ```

API keys are stored securely in `~/.gitcomm/config.json`. When you choose to reuse an environment key during setup, the key isn’t written to disk; GitComm will read it from the environment on each run. Environment variables always override stored configuration. `OPENROUTER_API_KEY` is the primary name; `OPEN_ROUTER_API_KEY` is accepted for compatibility. If you prefer to edit manually, copy `config.example.json` to `~/.gitcomm/config.json` and fill in your API key.

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

# Customize the primary model to use GPT-4o Mini
gitcomm -set-model "1:openai/gpt-4o-mini"

# Change the first fallback model to a free Llama model
gitcomm -set-model "2:meta-llama/llama-3.3-8b-instruct:free"

# Add Claude 3.5 Sonnet as an additional fallback option
gitcomm -set-model "4:anthropic/claude-3.5-sonnet"

# Use custom model and then generate a commit message
gitcomm -set-model "1:openai/gpt-4o-mini" && gitcomm -sa -auto
```

## Configuration

GitComm uses config values first (from `~/.gitcomm/config.json`) and falls back to built-in defaults when fields are missing.

You can start from the sample config in `config.example.json`.

```json
{
  "open_router_api_key": "your_openrouter_api_key",
  "models": [
    "meta-llama/llama-3.3-8b-instruct:free",
    "meta-llama/llama-4-scout",
    "google/gemini-2.5-flash-lite"
  ],
  "max_tokens": 400,
  "temperature": 0.7,
  "api_url": "https://openrouter.ai/api/v1/chat/completions",
  "timeout_seconds": 30
}
```

GitComm uses the following defaults:

- LLM Provider: OpenRouter
- Models (with fallback):
  1. `meta-llama/llama-3.3-8b-instruct:free` (Primary)
  2. `meta-llama/llama-4-scout` (Fallback 1)
  3. `google/gemini-2.5-flash-lite` (Fallback 2)
- Max Tokens: 400 (allows for proper Git commit format with subject + body)
- Temperature: 0.7 (balanced between creativity and consistency)
- Diff size limit: 1,500 lines (truncated with note if exceeded)
- Timeout: 30 seconds per model attempt
- Commit Format: Subject line (50-72 chars) + blank line + detailed body

The first two models in this list are chosen because they are typically available on OpenRouter's free tier; the third is a stronger fallback that may require paid credits. If OpenRouter's free offerings change, you can edit the `models` array in `~/.gitcomm/config.json` to point to other free models, or use the `-set-model` flag to update models from the command line.

### Model Fallback System

GitComm automatically tries multiple models if one fails:

1. **Primary Model**: `meta-llama/llama-3.3-8b-instruct:free` - Free and capable for most commit messages
2. **Fallback 1**: `meta-llama/llama-4-scout` - Strong performance if Llama 3.3 is unavailable  
3. **Fallback 2**: `google/gemini-2.5-flash-lite` - Fast and capable final option

Commit messages follow proper Git format with a concise subject line and detailed body explaining the changes. If all models fail, you'll see an error message.

### Customizing Models

You can customize the models used by GitComm using the `-set-model` flag. This allows you to:

1. Replace existing models in the fallback chain
2. Add new models to the end of the chain
3. Use different models from OpenRouter's catalog

**Usage:**
```bash
gitcomm -set-model "position:provider/model-name"
```

**Position numbers:**
- `1` = Primary model (first tried)
- `2` = First fallback model (tried if primary fails)
- `3` = Second fallback model (tried if first fallback fails)
- `4+` = Additional fallback models (appended to the end)

**Examples:**
```bash
# Change the primary model to GPT-4o Mini
gitcomm -set-model "1:openai/gpt-4o-mini"

# Change the first fallback model to Llama 3.3 8B Instruct (free tier)
gitcomm -set-model "2:meta-llama/llama-3.3-8b-instruct:free"

# Add a new model as the fourth fallback option
gitcomm -set-model "4:anthropic/claude-3.5-sonnet"

# View current configuration
cat ~/.gitcomm/config.json
```

**Model name format:**
- Use the full model identifier from OpenRouter
- Format: `provider/model-name` (e.g., `openai/gpt-4o-mini`)
- Must contain a `/` character to separate provider from model name
- Some models include qualifiers like `:free` for free tier models
- Check [OpenRouter's model list](https://openrouter.ai/models) for available models

**Note:** When you set a model at a position beyond the current list length, it will be appended as a new fallback option. For example, if you have 3 models and set position 5, the new model will be added as the 4th model (positions 1-3 remain unchanged, position 4 becomes the new model).

**Important:** The `-set-model` flag is mutually exclusive with other GitComm operations. When you use `-set-model`, GitComm will only update the configuration and exit; it will not generate commit messages, stage changes, or perform any other operations. To use a custom model and then generate a commit message, you need to run two separate commands:

```bash
# First, update the model configuration
gitcomm -set-model "1:openai/gpt-4o-mini"

# Then, use GitComm normally with the new model
gitcomm -sa -auto
```

### Example Output

**Basic Usage:**
```
📄 Analyzed 45 lines of diff
🤖 Generating commit message...
⚡ Using Llama 3.3 8B Instruct (Free)

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
⚡ Using Llama 3.3 8B Instruct (Free)
⚠️  Llama 3.3 8B Instruct (Free) failed, trying next model...
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
⚡ Using Llama 3.3 8B Instruct (Free)

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

- `OPENROUTER_API_KEY`: Your OpenRouter API key (preferred)
- `OPEN_ROUTER_API_KEY`: Your OpenRouter API key (legacy/compatibility)

## Command Line Flags

- `-auto`: Automatically commit with the generated message
- `-ap`: Automatically commit and push to remote
- `-sa`: Stage all changes before analyzing (equivalent to `git add .`)
- `-set-model`: Set model at position (format: position:provider/model-name)

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
