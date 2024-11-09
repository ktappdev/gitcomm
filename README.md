# gitcomm

A CLI tool that generates commit messages using LLMs by analyzing your git diff.

## Features

- Generates meaningful commit messages based on your changes
- Auto-commit functionality with `--auto` flag
- Uses GPT-4-mini for cost-effective and reliable results

## Requirements

1. Git installed and configured on your system
2. OpenAI API key set in your environment as `OPENAI_API_KEY`

## Usage

```bash
gitcomm              # Generate commit message only
gitcomm --auto      # Generate message and auto commit
gitcomm --ap      # Generate message and auto commit then auto push
```

## Installation

Currently, this tool requires manual setup. Future versions will include a proper installation process.

## Configuration

The tool uses GPT-4-mini by default for an optimal balance between cost and performance.

## Notes

This project was initially created for personal use but is being expanded to be more generally useful. Contributions and suggestions are welcome!

## Known Issues

- Commit message generation may occasionally need refinement
- Limited configuration options currently available
  Feel free to collab on this project - PRs welcomed
