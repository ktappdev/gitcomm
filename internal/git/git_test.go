package git

import (
	"strings"
	"testing"
)

func TestLimitDiffSizeWithInfoTruncates(t *testing.T) {
	diff := strings.Join([]string{"a", "b", "c", "d"}, "\n")
	got, truncated := limitDiffSizeWithInfo(diff, 2)
	if !truncated {
		t.Fatal("expected truncation")
	}
	if !strings.Contains(got, "truncated, 2 more lines") {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestLimitDiffSizeWithInfoExactLimitWithTrailingNewline(t *testing.T) {
	diff := "a\nb\n"
	got, truncated := limitDiffSizeWithInfo(diff, 2)
	if truncated {
		t.Fatalf("did not expect truncation: %q", got)
	}
	if got != diff {
		t.Fatalf("expected original diff, got %q", got)
	}
	if countLines(diff) != 2 {
		t.Fatalf("expected 2 lines, got %d", countLines(diff))
	}
}
