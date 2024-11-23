package analyzer

import (
	"strings"

	"github.com/ktappdev/gitcomm/internal/llm"
)

func AnalyzeChanges(diff string) (string, error) {
	client, err := llm.NewClient(llm.ClientConfig{
		Provider: llm.ProviderGemini, // Use Gemini as default
		Model:    "gemini-1.5-flash-8b",
	})
	if err != nil {
		return "", err
	}
	defer client.Close()

	prompt := `Analyze the following git diff and provide a single-line commit message based on the changes. 
Please ensure that your response strictly follows the specified format below.

Git Diff:
` + diff + `

Format your response as follows, including the exact wording:
Generated Commit Message:
[Your generated commit message here]

Example output:
Generated Commit Message:
Fix bug in user login process

Make sure to provide a commit message that accurately reflects the changes made in the git diff. Thank you!`

	response, err := client.SendPrompt(prompt)
	if err != nil {
		return "", err
	}

	commitMessage := extractCommitMessage(response)

	return commitMessage, nil
}

func extractCommitMessage(response string) string {
	sections := strings.Split(response, "\n\n")
	for _, section := range sections {
		if strings.HasPrefix(section, "Generated Commit Message:") {
			return strings.TrimSpace(strings.TrimPrefix(section, "Generated Commit Message:"))
		}
	}
	return "update"
}
