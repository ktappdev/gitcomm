package git

import (
	"os/exec"
)

// GetStagedChanges retrieves the staged changes in the current Git repository
func GetStagedChanges() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
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
