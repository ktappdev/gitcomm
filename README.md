# GitComm

GitComm is a CLI tool that uses LLMs (Groq by default) to automatically generate meaningful Git commit messages by analyzing your staged changes.

## Features

- 🤖 Uses AI to analyze staged changes and generate commit messages
- ⚡ Powered by Groq's fast LLMs (with OpenAI fallback support)
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

1. Get a Groq API key from [Groq's website](https://console.groq.com) IT'S FREE!!

2. Set your API key as an environment variable:

```bash
# For Groq (default)
export GROQ_API_KEY=your_groq_api_key

# For OpenAI (optional fallback)
export OPENAI_API_KEY=your_openai_api_key
```

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

# Auto-commit with the generated message
gitcomm -auto

# Auto-commit and push to remote
gitcomm -ap
```

## Configuration

GitComm uses the following defaults:

- LLM Provider: Groq
- Model: llama-3.2-1b-preview
- Max Tokens: 50 (approximately 2 lines of text)
- Temperature: 0.7 (balanced between creativity and consistency)

## Environment Variables

- `GROQ_API_KEY`: Your Groq API key (required by default)
- `OPENAI_API_KEY`: Your OpenAI API key (optional, for fallback)

## Command Line Flags

- `-auto`: Automatically commit with the generated message
- `-ap`: Automatically commit and push to remote

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

- [Groq](https://groq.com) for their fast LLM API
- The Go community for the excellent tooling

## Troubleshooting

### Common Issues

1. **No API Key Set**

   ```
   Error: API key not set for provider groq
   ```

   Solution: Set your GROQ_API_KEY environment variable

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
