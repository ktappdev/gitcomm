package analyzer

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/ktappdev/gitcomm/internal/diag"
	"github.com/ktappdev/gitcomm/internal/llm"
)

const (
	compactDiffThresholdChars = 12000
	maxCompactContextLines    = 6
	maxCompactChangeLines     = 12
	maxCompactLineLength      = 160
)

func AnalyzeChanges(diff string) (string, error) {
	fmt.Println("🤖 Generating commit message...")
	if strings.TrimSpace(diff) == "" {
		diag.Error("analyzer", "refusing to analyze empty diff")
		return "", fmt.Errorf("no staged diff content available to analyze")
	}

	client, err := llm.NewClient(llm.ClientConfig{MaxTokens: 400, Temperature: 0.7})
	if err != nil {
		return "", err
	}
	defer client.Close()

	analysisDiff, compacted := prepareDiffForAnalysis(diff)
	prompt := buildPrompt(analysisDiff)
	diag.Info("analyzer", "built prompt", "diff_chars", len(diff), "analysis_diff_chars", len(analysisDiff), "prompt_chars", len(prompt), "compacted", compacted)

	response, err := client.SendPrompt(prompt)
	if err != nil {
		return "", err
	}

	commitMessage, err := extractCommitMessage(response)
	if err != nil {
		diag.Error("analyzer", "failed to parse commit message", "error", err, "response_snippet", diag.Snippet(response, 300))
		return "", err
	}
	diag.Info("analyzer", "parsed commit message", "response_chars", len(response), "commit_chars", len(commitMessage))
	return commitMessage, nil
}

func buildPrompt(diff string) string {
	return `Analyze the following git diff and generate a proper Git commit message with both a subject line and detailed body.

Git Diff:
` + diff + `

Please follow these Git commit message best practices:
- Subject line: 50-72 characters, summarize the change
- Leave a blank line after the subject
- Body: Detailed explanation, wrap lines at 72 characters
- Explain WHAT changed and WHY (not just how)

Format your response as follows:
Generated Commit Message:
[Subject line - 50-72 characters]

[Detailed body explaining the changes, wrapped at 72 characters.
Include context about what was changed and why it was necessary.
You can use multiple paragraphs if needed.]`
}

func prepareDiffForAnalysis(diff string) (string, bool) {
	if len(diff) <= compactDiffThresholdChars {
		diag.Debug("analyzer", "using full diff for analysis", "diff_chars", len(diff), "threshold_chars", compactDiffThresholdChars)
		return diff, false
	}
	compacted := compactDiff(diff)
	if compacted == "" || len(compacted) >= len(diff) {
		diag.Warn("analyzer", "diff compaction skipped; no improvement", "diff_chars", len(diff), "compacted_chars", len(compacted), "threshold_chars", compactDiffThresholdChars)
		return diff, false
	}
	removed := len(diff) - len(compacted)
	diag.Info("analyzer", "using compact diff for analysis", "original_chars", len(diff), "compacted_chars", len(compacted), "reduced_chars", removed, "threshold_chars", compactDiffThresholdChars)
	return compacted, true
}

func compactDiff(diff string) string {
	lines := strings.Split(strings.TrimSuffix(diff, "\n"), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return ""
	}

	out := make([]string, 0, len(lines)/2)
	contextSeen := 0
	omittedContext := 0
	changeLines := make([]string, 0, maxCompactChangeLines)
	inHunk := false

	flushOmittedContext := func() {
		if omittedContext > 0 {
			out = append(out, fmt.Sprintf("[[gitcomm: %d context lines omitted in compact diff]]", omittedContext))
			omittedContext = 0
		}
	}
	flushHunkChanges := func() {
		if len(changeLines) == 0 {
			return
		}
		flushOmittedContext()
		out = append(out, sampleChangeLines(changeLines)...)
		changeLines = changeLines[:0]
	}
	resetHunk := func() {
		contextSeen = 0
		omittedContext = 0
		changeLines = changeLines[:0]
	}

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flushHunkChanges()
			out = append(out, line)
			resetHunk()
			inHunk = false
		case strings.HasPrefix(line, "index "), strings.HasPrefix(line, "new file mode "), strings.HasPrefix(line, "deleted file mode "), strings.HasPrefix(line, "similarity index "), strings.HasPrefix(line, "rename from "), strings.HasPrefix(line, "rename to "), strings.HasPrefix(line, "Binary files "):
			flushHunkChanges()
			out = append(out, line)
		case strings.HasPrefix(line, "--- "), strings.HasPrefix(line, "+++ "):
			flushHunkChanges()
			out = append(out, line)
		case strings.HasPrefix(line, "@@"):
			flushHunkChanges()
			out = append(out, line)
			resetHunk()
			inHunk = true
		case inHunk && strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			changeLines = append(changeLines, truncateCompactLine(line))
		case inHunk && strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			changeLines = append(changeLines, truncateCompactLine(line))
		case inHunk && strings.HasPrefix(line, " "):
			flushHunkChanges()
			if contextSeen < maxCompactContextLines {
				flushOmittedContext()
				out = append(out, truncateCompactLine(line))
				contextSeen++
			} else {
				omittedContext++
			}
		default:
			flushHunkChanges()
			out = append(out, truncateCompactLine(line))
		}
	}
	flushHunkChanges()
	flushOmittedContext()
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func sampleChangeLines(lines []string) []string {
	if len(lines) <= maxCompactChangeLines {
		return lines
	}
	frontCount := maxCompactChangeLines / 2
	backCount := maxCompactChangeLines - frontCount
	out := make([]string, 0, maxCompactChangeLines+1)
	out = append(out, lines[:frontCount]...)
	out = append(out, fmt.Sprintf("[[gitcomm: %d changed lines omitted in compact diff; showing early and late changes]]", len(lines)-maxCompactChangeLines))
	out = append(out, lines[len(lines)-backCount:]...)
	return out
}

func truncateCompactLine(line string) string {
	if len(line) <= maxCompactLineLength {
		return line
	}
	return line[:maxCompactLineLength] + " [[gitcomm: line truncated for compact diff]]"
}

func extractCommitMessage(response string) (string, error) {
	cleaned := strings.TrimSpace(response)
	if cleaned == "" {
		return "", fmt.Errorf("model returned empty analysis output")
	}

	markers := []string{
		"Generated Commit Message:",
		"Commit Message:",
		"**Generated Commit Message:**",
		"## Generated Commit Message",
	}
	for _, marker := range markers {
		if idx := strings.Index(cleaned, marker); idx >= 0 {
			return normalizeCommitMessage(cleaned[idx+len(marker):])
		}
	}

	return normalizeCommitMessage(cleaned)
}

func normalizeCommitMessage(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "`")
	lines := strings.Split(raw, "\n")
	cleanedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			continue
		}
		if len(cleanedLines) == 0 {
			trimmed = cleanSubjectPrefix(trimmed)
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
			trimmed = strings.Trim(trimmed, `"'`)
			cleanedLines = append(cleanedLines, trimmed)
			continue
		}
		cleanedLines = append(cleanedLines, strings.TrimRight(line, " \t"))
	}
	message := strings.TrimSpace(strings.Join(cleanedLines, "\n"))
	if message == "" {
		return "", fmt.Errorf("model response did not contain a commit message")
	}

	subject, body := splitSubjectBody(message)
	if err := validateSubject(subject); err != nil {
		return "", err
	}
	if body == "" {
		return subject, nil
	}
	return subject + "\n\n" + body, nil
}

func splitSubjectBody(message string) (string, string) {
	lines := strings.Split(message, "\n")
	subject := strings.TrimSpace(lines[0])
	bodyLines := make([]string, 0, len(lines)-1)
	seenContent := false
	for _, line := range lines[1:] {
		trimmed := strings.TrimRight(line, " \t")
		if strings.TrimSpace(trimmed) == "" {
			if seenContent && len(bodyLines) > 0 && strings.TrimSpace(bodyLines[len(bodyLines)-1]) != "" {
				bodyLines = append(bodyLines, "")
			}
			continue
		}
		seenContent = true
		bodyLines = append(bodyLines, trimmed)
	}
	body := strings.TrimSpace(strings.Join(bodyLines, "\n"))
	return subject, body
}

func validateSubject(subject string) error {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return fmt.Errorf("commit message subject is empty")
	}
	if len(subject) > 100 {
		return fmt.Errorf("commit message subject too long: %d characters", len(subject))
	}
	lower := strings.ToLower(subject)
	badPrefixes := []string{"here's", "here is", "generated commit message", "commit message", "explanation:", "you can use this commit message", "this commit message"}
	for _, prefix := range badPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return fmt.Errorf("model returned commentary instead of a commit subject: %q", subject)
		}
	}
	if strings.Contains(lower, "commit message") && strings.HasSuffix(subject, ":") {
		return fmt.Errorf("model returned commentary instead of a commit subject: %q", subject)
	}
	if !containsLetter(subject) {
		return fmt.Errorf("commit message subject must contain letters: %q", subject)
	}
	return nil
}

func cleanSubjectPrefix(subject string) string {
	prefixes := []string{"Subject:", "Title:"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(subject), strings.ToLower(prefix)) {
			return strings.TrimSpace(subject[len(prefix):])
		}
	}
	return subject
}

func containsLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
