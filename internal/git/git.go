package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func StageAll() error {
	cmd := exec.Command("git", "add", ".")
	err := cmd.Run()
	return err
}

func GetStagedChanges() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	res, wasTruncated := limitDiffSizeWithInfo(string(output), 1500)
	
	// Show useful diff information
	originalLines := len(strings.Split(string(output), "\n"))
	if wasTruncated {
		fmt.Printf("📄 Analyzed %d lines of diff (truncated from %d lines)\n", 1500, originalLines)
	} else {
		fmt.Printf("📄 Analyzed %d lines of diff\n", originalLines)
	}
	
	return res, nil
}

func limitDiffSizeWithInfo(diff string, maxLines int) (string, bool) {
	if maxLines <= 0 {
		return diff, false
	}

	lines := strings.Split(diff, "\n")
	if len(lines) <= maxLines {
		return diff, false
	}

	truncated := strings.Join(lines[:maxLines], "\n") + fmt.Sprintf("\n... (truncated, %d more lines)", len(lines)-maxLines)
	return truncated, true
}

// Keep old function for backward compatibility
func limitDiffSize(diff string, maxLines int) string {
	result, _ := limitDiffSizeWithInfo(diff, maxLines)
	return result
}

func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}

func Push() error {
	cmd := exec.Command("git", "push")
	return cmd.Run()
}
