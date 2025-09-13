package analyzer

import (
	"fmt"
	"strings"

	"github.com/ktappdev/gitcomm/internal/llm"
)

func AnalyzeChanges(diff string) (string, error) {
	fmt.Println("🤖 Generating commit message...")
	
	client, err := llm.NewClient(llm.ClientConfig{
		MaxTokens:   400, // Allow for detailed commit messages
		Temperature: 0.7, // Keep same temperature
	})
	if err != nil {
		return "", err
	}
	defer client.Close()

	prompt := `Analyze the following git diff and generate a proper Git commit message with both a subject line and detailed body.

Git Diff:
` + diff + `

Please follow these Git commit message best practices:
- Subject line: 50-72 characters, summarize the change
- Leave a blank line after the subject
- Body: Detailed explanation, wrap lines at 72 characters
- Explain WHAT changed and WHY (not just how)

Format your response as follows:
Generated Commit Message:
[Subject line - 50-72 characters]

[Detailed body explaining the changes, wrapped at 72 characters.
Include context about what was changed and why it was necessary.
You can use multiple paragraphs if needed.]

Example output:
Generated Commit Message:
Add JWT-based user authentication system

Implement comprehensive authentication using JSON Web Tokens with
bcrypt password hashing for enhanced security. Add middleware for
protecting authenticated routes and validation for email/password
requirements.

Updates database schema to include user roles and timestamps for
better user management and audit trails.`

	response, err := client.SendPrompt(prompt)
	if err != nil {
		return "", err
	}

	commitMessage := extractCommitMessage(response)

	return commitMessage, nil
}

func extractCommitMessage(response string) string {
	// Look for the "Generated Commit Message:" marker
	marker := "Generated Commit Message:"
	idx := strings.Index(response, marker)
	if idx == -1 {
		return "update" // fallback if marker not found
	}

	// Extract everything after the marker
	commitMessage := strings.TrimSpace(response[idx+len(marker):])
	if commitMessage == "" {
		return "update" // fallback if empty
	}

	// Return the complete multi-line commit message
	// This preserves the Git format: subject + blank line + body
	return commitMessage
}
