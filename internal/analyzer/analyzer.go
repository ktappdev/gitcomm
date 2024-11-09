package analyzer

import (
	"strings"

	"github.com/ktappdev/gitcomm/internal/llm"
)

func AnalyzeChanges(diff string) (string, error) {
	client, err := llm.NewClient(llm.ClientConfig{
		Provider: llm.ProviderGroq, // Default to Groq
		Model:    "llama2-70b-4096",
	})
	if err != nil {
		return "", err
	}

	prompt := `Analyze the following git diff and provide:
A generated a one line commit message based on the changes

Git Diff:
` + diff

	response, err := client.SendPrompt(prompt)
	if err != nil {
		return "", err
	}

	return extractCommitMessage(response), nil
}

func extractCommitMessage(response string) string {
	sections := strings.Split(response, "\n\n")
	for _, section := range sections {
		if strings.HasPrefix(section, "Generated Commit Message:") {
			return strings.TrimSpace(strings.TrimPrefix(section, "Generated Commit Message:"))
		}
	}
	return ""
}
