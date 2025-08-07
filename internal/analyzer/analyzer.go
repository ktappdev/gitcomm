package analyzer

import (
	"fmt"
	"strings"

	"github.com/ktappdev/gitcomm/internal/llm"
)

func AnalyzeChanges(diff string) (string, error) {
	fmt.Println("[debug] analyzer: create llm client")
	client, err := llm.NewClient(llm.ClientConfig{
		Provider: llm.ProviderGemini,
		Model:    "gemini-1.5-flash",
	})
	if err != nil {
		fmt.Println("[debug] analyzer: NewClient error:", err)
		return "", err
	}
	defer func() { fmt.Println("[debug] analyzer: closing llm client"); client.Close() }()

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

	fmt.Println("[debug] analyzer: sending prompt to llm (bytes)", len(prompt))
	response, err := client.SendPrompt(prompt)
	if err != nil {
		fmt.Println("[debug] analyzer: SendPrompt error:", err)
		return "", err
	}
	fmt.Println("[debug] analyzer: got response (bytes)", len(response))

	commitMessage := extractCommitMessage(response)
	fmt.Println("[debug] analyzer: extracted commit message length", len(commitMessage))

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
