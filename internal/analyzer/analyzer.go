package analyzer

import (
	// "log/slog"
	"strings"

	"github.com/ktappdev/gitcomm/internal/llm"
)

func AnalyzeChanges(diff string) (string, error) {
	client, err := llm.NewClient(llm.ClientConfig{
		Provider: llm.ProviderOpenAI, // Default to Groq
		Model:    "gpt-4o-mini",
	})
	if err != nil {
		return "", err
	}
	// slog.Info("LLM Client: ", client)

	prompt := `Analyze the following git diff and provide:
A generated a one line commit message based on the changes

Git Diff:
` + diff + `
Please format your response as follows:
Generated Commit Message:
[Your generated commit message here]`

	response, err := client.SendPrompt(prompt)
	// slog.Info("LLM Response: ", response)
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
	return "update"
}
