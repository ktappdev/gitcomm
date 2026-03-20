package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/ktappdev/gitcomm/internal/diag"
)

const MaxDiffLines = 1500

func StageAll() error {
	cmd := exec.Command("git", "add", ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		diag.Error("git", "git add failed", "error", err, "output", diag.Snippet(msg, 300))
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}
	return nil
}

func GetStagedChanges() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg != "" {
			lines := strings.Split(msg, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if strings.HasPrefix(line, "fatal:") || strings.HasPrefix(line, "error:") {
					msg = line
					break
				}
			}
			diag.Error("git", "git diff --cached failed", "error", err, "output", diag.Snippet(msg, 300))
			return "", fmt.Errorf("%s", msg)
		}
		return "", err
	}

	res, wasTruncated := limitDiffSizeWithInfo(string(output), MaxDiffLines)
	originalLines := countLines(string(output))
	returnedLines := countLines(res)
	fmt.Printf("📄 Analyzed %d lines of diff", minInt(originalLines, MaxDiffLines))
	if wasTruncated {
		fmt.Printf(" (truncated from %d lines)", originalLines)
	}
	fmt.Println()

	if strings.TrimSpace(res) == "" {
		diag.Warn("git", "staged diff is empty", "bytes", len(output), "lines", originalLines)
	} else {
		diag.Info("git", "collected staged diff", "bytes", len(output), "lines", originalLines, "returned_lines", returnedLines, "truncated", wasTruncated)
	}
	return res, nil
}

func limitDiffSizeWithInfo(diff string, maxLines int) (string, bool) {
	if maxLines <= 0 {
		return diff, false
	}

	lines := splitDiffLines(diff)
	if len(lines) <= maxLines {
		return diff, false
	}

	truncated := strings.Join(lines[:maxLines], "\n") + fmt.Sprintf("\n... (truncated, %d more lines)", len(lines)-maxLines)
	return truncated, true
}

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

func countLines(s string) int {
	return len(splitDiffLines(s))
}

func splitDiffLines(s string) []string {
	if s == "" {
		return nil
	}
	trimmed := strings.TrimSuffix(s, "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
