# GitComm

GitComm is a CLI tool that uses LLMs to automatically generate meaningful Git commit messages by analyzing your staged changes. It uses OpenRouter to access multiple models with automatic fallback support.

GitComm is also designed to stay CLI-friendly when providers misbehave: model failures are summarized in the terminal, while detailed diagnostics are written to a persistent log file so you can troubleshoot without noisy output on successful runs.

## Features

- 🤖 Uses AI to analyze staged changes and generate commit messages
- ⚡ Powered by OpenRouter with multiple model fallback support
- 🔄 Automatic model fallback: Llama 3.3 → Llama 4 → Gemini
- 📝 Persistent diagnostics log for provider/config/runtime failures
- 🛡️ Runtime-tolerant config loading with safe defaults when possible
- 🚀 Auto-commit and push capabilities
- 💻 Cross-platform support (Windows, macOS, Linux)

## Installation

### Using Go Install

```bash
go install github.com/ktappdev/gitcomm@latest
```

If you installed GitComm this way, you can later update it with:

```bash
gitcomm update
```

`gitcomm update` runs `go install github.com/ktappdev/gitcomm@latest` for convenience. It only works for copies installed via `go install`, requires `go` to be available on your PATH, and does not update binaries installed via release downloads or other package managers.

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
   This collects your OpenRouter API key and seeds `~/.gitcomm/config.json` with the default models, token/temperature settings, API URL, and timeout so you can edit them later. If `OPENROUTER_API_KEY` is already set in your environment, setup will offer to use it without storing the key in the config file. `OPEN_ROUTER_API_KEY` is also supported for legacy compatibility.
   If a config file already exists, setup will prompt you to keep it, overwrite it, or back it up before overwriting.

2. Environment Variable (set `OPENROUTER_API_KEY` in your shell; `OPEN_ROUTER_API_KEY` is also supported):
   ```bash
   export OPENROUTER_API_KEY=your_openrouter_api_key
   ```

API keys are stored in `~/.gitcomm/config.json` when you choose to save them. When you reuse an environment key during setup, the key is not written to disk. `OPENROUTER_API_KEY` is the primary name; `OPEN_ROUTER_API_KEY` is accepted for compatibility.

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

5. Update a Go-installed GitComm binary (requires `go` on your PATH):

```bash
gitcomm update
```

## Examples

```bash
# Basic usage - analyze staged changes and suggest a commit message
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

# Update GitComm if you installed it with `go install` and still have `go` on your PATH
gitcomm update
```

## Configuration

GitComm reads `~/.gitcomm/config.json` when present, applies environment overrides, and fills any missing fields with built-in defaults.

You can start from `config.example.json`:

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

### Canonical defaults

GitComm currently defaults to:

- Models (with fallback):
  1. `meta-llama/llama-3.3-8b-instruct:free`
  2. `meta-llama/llama-4-scout`
  3. `google/gemini-2.5-flash-lite`
- Max tokens: `400`
- Temperature: `0.7`
- API URL: `https://openrouter.ai/api/v1/chat/completions`
- Timeout: `30` seconds per model attempt
- Diff size limit: `1,500` lines, with truncation noted in CLI output
- Large diffs are compacted before sending to the model so file paths, hunk headers, and representative changes are preserved while bulk context is reduced
- Compacted diffs may include explicit `[[gitcomm: ...]]` omission markers so skipped context is clearly editorial rather than real patch content

The first two models are intended to be free-friendly on OpenRouter when available. Availability and pricing can change, so you can update the `models` array at any time.

### Config precedence and runtime behavior

GitComm resolves settings in this order:

1. Built-in defaults
2. `~/.gitcomm/config.json` values, if the file exists
3. Environment API key overrides:
   - `OPENROUTER_API_KEY`
   - `OPEN_ROUTER_API_KEY`

Important runtime behavior:

- Missing config fields do not break the app; GitComm fills them from defaults.
- Empty or invalid model entries are ignored at runtime.
- If all configured models are invalid, GitComm falls back to the built-in default model list.
- Negative numeric config values are sanitized at runtime.
- If the config file cannot be parsed, GitComm logs the failure and may still continue with runtime defaults if an API key is available via environment variables.
- Environment API keys override any key stored in the config file.

### Model fallback system

GitComm automatically tries multiple models if one fails:

1. **Primary Model**: `meta-llama/llama-3.3-8b-instruct:free`
2. **Fallback 1**: `meta-llama/llama-4-scout`
3. **Fallback 2**: `google/gemini-2.5-flash-lite`

If one provider request fails, GitComm prints a short fallback message in the terminal and records more detail in the diagnostics log.

### Customizing models

Use `-set-model` to replace or append models in the fallback chain:

```bash
gitcomm -set-model "position:provider/model-name"
```

Positions:

- `1` = primary model
- `2` = first fallback
- `3` = second fallback
- `4+` = append additional fallbacks

Examples:

```bash
# Change the primary model to GPT-4o Mini
gitcomm -set-model "1:openai/gpt-4o-mini"

# Change the first fallback model to Llama 3.3 8B Instruct (free tier)
gitcomm -set-model "2:meta-llama/llama-3.3-8b-instruct:free"

# Add a new model as the fourth fallback option
gitcomm -set-model "4:anthropic/claude-3.5-sonnet"
```

Model names must use OpenRouter's `provider/model-name` format. Some models include qualifiers such as `:free`.

**Important:** `-set-model` only updates configuration and exits. Run GitComm again afterward to generate a commit message. Unlike normal runtime usage, config-editing commands require a parseable `~/.gitcomm/config.json`; if the file is malformed, fix or remove it first, then rerun the command.

## Diagnostics and logging

GitComm writes diagnostics to:

```text
~/.gitcomm/logs/diagnostics.log
```

Use this log when the CLI reports model/config errors but keeps the terminal output brief.

What is logged:

- Config load and parse failures
- Runtime fallback decisions
- Model attempt start/success/failure
- Provider HTTP status failures
- Response parsing failures
- Analyzer parsing failures

Sensitive values are sanitized before logging.

### Quick log inspection

```bash
# Show recent diagnostics
tail -n 50 ~/.gitcomm/logs/diagnostics.log

# Focus on provider failures
grep 'component="llm"' ~/.gitcomm/logs/diagnostics.log | tail -n 20

# Focus on config issues
grep 'component="config"' ~/.gitcomm/logs/diagnostics.log | tail -n 20
```

## Troubleshooting

### Quick flow

1. Re-run `gitcomm`
2. Read the terminal summary to see whether it was a config, auth, or model failure
3. Open `~/.gitcomm/logs/diagnostics.log`
4. Look for the latest `component="config"`, `component="llm"`, or `component="analyzer"` entries
5. If a model repeatedly fails, replace it with another OpenRouter model using `-set-model`

### Common provider failures

GitComm now surfaces clearer provider errors. Common examples:

- **401 / 403 authentication failure**
  - Usually means the OpenRouter API key is missing, invalid, or not being picked up.
  - Check `OPENROUTER_API_KEY`, `OPEN_ROUTER_API_KEY`, and `~/.gitcomm/config.json`.

- **402 payment/credits failure**
  - Usually means the selected model requires credits or is not available for your plan.
  - Switch to another model or add credits in OpenRouter.

- **429 rate limited**
  - Usually means the provider or model is temporarily throttling requests.
  - Retry shortly or move that model later in the fallback chain.

- **400 malformed or too-large request**
  - Usually means the diff/prompt was too large or the provider rejected the request format.
  - Try staging a smaller set of changes and rerun.

- **All models failed**
  - GitComm will report that all models failed and point you to the diagnostics log.
  - The last error shown in the CLI may not be the only failure; inspect the log to see the full fallback chain.

### Free or cheap OpenRouter models keep failing

If low-cost or free models are unreliable for your account or workload:

1. Check `~/.gitcomm/logs/diagnostics.log` for the exact status codes and provider messages.
2. If you see `402`, the model likely is not actually available on your plan.
3. If you see `429`, the model is likely rate-limited; retry or move it later in the chain.
4. If you see `400`, try a smaller staged diff.
5. Replace unstable models with alternatives:

```bash
gitcomm -set-model "1:openai/gpt-4o-mini"
gitcomm -set-model "2:meta-llama/llama-3.3-8b-instruct:free"
```

If you prefer to stay on free-tier options, keep your first few entries pointed at currently available free OpenRouter models and treat later entries as paid fallbacks.

### Common issues

1. **No API Key Set**

   ```
   OpenRouter API key not set in config file or OPENROUTER_API_KEY/OPEN_ROUTER_API_KEY environment variables
   ```

   Solution: Set an environment variable or run `gitcomm -setup`

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

## Environment Variables

- `OPENROUTER_API_KEY`: OpenRouter API key (preferred)
- `OPEN_ROUTER_API_KEY`: OpenRouter API key (legacy compatibility)

## Command Line Flags

- `-auto`: Automatically commit with the generated message
- `-ap`: Automatically commit and push to remote
- `-sa`: Stage all changes before analyzing (equivalent to `git add .`)
- `-debug`: Enable verbose debug logging to the diagnostics log
- `-set-model`: Set model at position (`position:provider/model-name`)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

- [OpenRouter](https://openrouter.ai) for their unified LLM API with model fallback support
- The Go community for the excellent tooling
