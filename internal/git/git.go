package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// StageAll stages all changes in the current Git repository
func StageAll() error {
	cmd := exec.Command("git", "add", ".")
	return cmd.Run()
}

// GetStagedChanges retrieves the staged changes in the current Git repository
func GetStagedChanges() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return limitDiffSize(string(output), 1500), nil
}

func limitDiffSize(diff string, maxLines int) string {
	if maxLines <= 0 {
		return diff
	}

	lines := strings.Split(diff, "\n")
	if len(lines) <= maxLines {
		return diff
	}

	return strings.Join(lines[:maxLines], "\n") + fmt.Sprintf("\n... (truncated, %d more lines)", len(lines)-maxLines)
}

// Commit performs a git commit with the given message
func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}

// Push performs a git push to the remote repository
func Push() error {
	cmd := exec.Command("git", "push")
	return cmd.Run()
}
