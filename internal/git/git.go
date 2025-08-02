package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func StageAll() error {
	fmt.Println("[debug] git: running 'git add .'")
	cmd := exec.Command("git", "add", ".")
	err := cmd.Run()
	if err != nil {
		fmt.Println("[debug] git: add error:", err)
	}
	return err
}

func GetStagedChanges() (string, error) {
	fmt.Println("[debug] git: running 'git diff --staged'")
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("[debug] git: diff error:", err)
		return "", err
	}
	res := limitDiffSize(string(output), 1500)
	fmt.Println("[debug] git: staged diff bytes", len(res))
	return res, nil
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

func Commit(message string) error {
	fmt.Println("[debug] git: running 'git commit -m <message>' len", len(message))
	cmd := exec.Command("git", "commit", "-m", message)
	err := cmd.Run()
	if err != nil {
		fmt.Println("[debug] git: commit error:", err)
	}
	return err
}

func Push() error {
	fmt.Println("[debug] git: running 'git push'")
	cmd := exec.Command("git", "push")
	err := cmd.Run()
	if err != nil {
		fmt.Println("[debug] git: push error:", err)
	}
	return err
}
