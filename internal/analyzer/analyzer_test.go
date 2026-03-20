package analyzer

import (
	"fmt"
	"strings"
	"testing"
)

func TestPrepareDiffForAnalysisUsesFullDiffWhenSmall(t *testing.T) {
	diff := "diff --git a/a.txt b/a.txt\n@@ -1 +1 @@\n-old\n+new\n"
	got, compacted := prepareDiffForAnalysis(diff)
	if compacted {
		t.Fatal("expected full diff")
	}
	if got != diff {
		t.Fatalf("got %q want %q", got, diff)
	}
}

func TestPrepareDiffForAnalysisCompactsLargeDiff(t *testing.T) {
	largeContext := strings.Repeat(" context line that is long enough to count toward threshold\n", 400)
	diff := "diff --git a/app.go b/app.go\nindex 123..456 100644\n--- a/app.go\n+++ b/app.go\n@@ -1,20 +1,20 @@\n" + largeContext + "-old important line\n+new important line\n"
	got, compacted := prepareDiffForAnalysis(diff)
	if !compacted {
		t.Fatal("expected compacted diff")
	}
	if len(got) >= len(diff) {
		t.Fatalf("expected compacted diff to be smaller: %d >= %d", len(got), len(diff))
	}
	for _, want := range []string{"diff --git a/app.go b/app.go", "@@ -1,20 +1,20 @", "-old important line", "+new important line", "[[gitcomm:"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compacted diff missing %q\n%s", want, got)
		}
	}
}

func TestCompactDiffLimitsNoiseButPreservesStructure(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git a/service.go b/service.go",
		"index 111..222 100644",
		"--- a/service.go",
		"+++ b/service.go",
		"@@ -10,20 +10,24 @@",
		" context 1",
		" context 2",
		" context 3",
		" context 4",
		" context 5",
		" context 6",
		" context 7",
		"-remove one",
		"+add one",
		"-remove two",
		"+add two",
	}, "\n")

	got := compactDiff(diff)
	if !strings.Contains(got, "diff --git a/service.go b/service.go") {
		t.Fatalf("missing file header: %s", got)
	}
	if !strings.Contains(got, "@@ -10,20 +10,24 @@") {
		t.Fatalf("missing hunk header: %s", got)
	}
	if !strings.Contains(got, "-remove one") || !strings.Contains(got, "+add one") {
		t.Fatalf("missing representative changes: %s", got)
	}
	if !strings.Contains(got, "[[gitcomm: 1 context lines omitted in compact diff]]") {
		t.Fatalf("expected explicit omitted-context marker: %s", got)
	}
}

func TestCompactDiffPreservesLateMeaningfulChanges(t *testing.T) {
	lines := []string{"diff --git a/app.go b/app.go", "--- a/app.go", "+++ b/app.go", "@@ -1,40 +1,40 @@"}
	for i := 0; i < 16; i++ {
		lines = append(lines, fmt.Sprintf("-noise change %d", i), fmt.Sprintf("+noise change %d", i))
	}
	lines = append(lines, "-return nil", "+return err")
	got := compactDiff(strings.Join(lines, "\n"))
	if !strings.Contains(got, "+return err") || !strings.Contains(got, "-return nil") {
		t.Fatalf("expected late meaningful change to be preserved: %s", got)
	}
	if !strings.Contains(got, "[[gitcomm: ") || !strings.Contains(got, "showing early and late changes") {
		t.Fatalf("expected explicit changed-line omission marker: %s", got)
	}
}

func TestCompactDiffUsesExplicitTruncationMarker(t *testing.T) {
	line := "+" + strings.Repeat("x", maxCompactLineLength+20)
	got := truncateCompactLine(line)
	if !strings.Contains(got, "[[gitcomm: line truncated for compact diff]]") {
		t.Fatalf("expected explicit truncation marker: %q", got)
	}
}

func TestExtractCommitMessageMarker(t *testing.T) {
	got, err := extractCommitMessage("Generated Commit Message:\nAdd diagnostics logging\n\nLog model attempts and config failures.")
	if err != nil {
		t.Fatalf("extractCommitMessage() error = %v", err)
	}
	want := "Add diagnostics logging\n\nLog model attempts and config failures."
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExtractCommitMessageAllowsShortConventionalSubject(t *testing.T) {
	got, err := extractCommitMessage("fix: typo")
	if err != nil {
		t.Fatalf("extractCommitMessage() error = %v", err)
	}
	if got != "fix: typo" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractCommitMessageStripsQuotedSubject(t *testing.T) {
	got, err := extractCommitMessage(`"fix: typo"`)
	if err != nil {
		t.Fatalf("extractCommitMessage() error = %v", err)
	}
	if got != "fix: typo" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractCommitMessageHandlesBulletSubjectAndBody(t *testing.T) {
	got, err := extractCommitMessage("- Add diagnostics logging\n\nExplain model fallback failures clearly.")
	if err != nil {
		t.Fatalf("extractCommitMessage() error = %v", err)
	}
	want := "Add diagnostics logging\n\nExplain model fallback failures clearly."
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExtractCommitMessageHandlesSubjectPrefix(t *testing.T) {
	got, err := extractCommitMessage("Subject: Add diagnostics logging\n\nLog request sizing.")
	if err != nil {
		t.Fatalf("extractCommitMessage() error = %v", err)
	}
	want := "Add diagnostics logging\n\nLog request sizing."
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExtractCommitMessageHandlesTitlePrefix(t *testing.T) {
	got, err := extractCommitMessage("Title: Add diagnostics logging")
	if err != nil {
		t.Fatalf("extractCommitMessage() error = %v", err)
	}
	if got != "Add diagnostics logging" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractCommitMessageRejectsCommentary(t *testing.T) {
	_, err := extractCommitMessage("Here's a commit message you can use:\n\nAdd logging")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "commentary") {
		t.Fatalf("unexpected error: %v", err)
	}
}
