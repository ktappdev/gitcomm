package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// Constants for OpenAI API
const OPENAI_API_KEY_ENV = "OPENAI_API_KEY"
const OPENAI_API_URL = "https://api.openai.com/v1/chat/completions"

// getGitDiff retrieves the staged changes in the current Git repository
func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// sendToLLM sends a prompt to the OpenAI API and returns the response
func sendToLLM(prompt string) (string, error) {
	// Prepare the request body
	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	})

	// Create and send the HTTP request
	req, _ := http.NewRequest("POST", OPENAI_API_URL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	apiKey := os.Getenv(OPENAI_API_KEY_ENV)
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not set in environment variable %s", OPENAI_API_KEY_ENV)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Extract the content from the response
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected response format")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("content is not a string")
	}

	return content, nil
}

// analyzeChanges sends the git diff to the LLM for analysis and commit message generation
func analyzeChanges(diff string) (string, error) {
	prompt := `Analyze the following git diff and provide:
3. A generated commit message based on the changes

Git Diff:
` + diff + `

Please format your response as follows:
Generated Commit Message:
[Your generated commit message here]

	No othe comment or added text`

	return sendToLLM(prompt)
}

// autoCommit performs a git commit with the given message
func autoCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}

func main() {
	// Parse command-line flags
	autoFlag := flag.Bool("auto", false, "Automatically commit with the generated message")
	flag.Parse()

	// Get the git diff of staged changes
	diff, err := getGitDiff()
	if err != nil {
		fmt.Println("Error getting git diff:", err)
		return
	}

	// Check if there are any staged changes
	if diff == "" {
		fmt.Println("No staged changes. Please stage your changes before running gitcomm.")
		return
	}

	fmt.Println("Analyzing changes and generating commit message...")
	analysis, err := analyzeChanges(diff)
	if err != nil {
		fmt.Println("Error analyzing changes:", err)
		return
	}

	// Display the analysis results
	sections := strings.Split(analysis, "\n\n")
	for _, section := range sections {
		fmt.Println(section)
		fmt.Println() // Add an extra newline for readability
	}

	// Extract the generated commit message from the analysis
	var commitMessage string
	for _, section := range sections {
		if strings.HasPrefix(section, "Generated Commit Message:") {
			commitMessage = strings.TrimPrefix(section, "Generated Commit Message:")
			commitMessage = strings.TrimSpace(commitMessage)
			break
		}
	}

	// Handle auto-commit if the flag is set
	if *autoFlag {
		if commitMessage == "" {
			fmt.Println("Error: Could not extract a commit message from the analysis.")
			return
		}
		fmt.Println("Auto-committing with the generated message...")
		err = autoCommit(commitMessage)
		if err != nil {
			fmt.Println("Error committing:", err)
		} else {
			fmt.Println("Changes committed successfully!")
		}
	} else {
		// Provide instructions for manual commit
		fmt.Println("You can use this information to craft your commit message.")
		fmt.Println("Remember to review and adjust the suggested message as needed.")
		fmt.Println("To auto-commit, run getcomm with the --auto flag.")
	}
}
