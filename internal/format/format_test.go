package format

import (
	"strings"
	"testing"
)

const sampleOutput = `internal/foo/foo.go
internal/bar/bar.go
`

func TestParse_FileCount(t *testing.T) {
	r, err := Parse(strings.NewReader(sampleOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Files) != 2 {
		t.Errorf("got %d files, want 2", len(r.Files))
	}
}

func TestParse_HasIssues(t *testing.T) {
	r, _ := Parse(strings.NewReader(sampleOutput))
	if !r.HasIssues() {
		t.Error("expected HasIssues() = true")
	}
}

func TestParse_FileNames(t *testing.T) {
	r, _ := Parse(strings.NewReader(sampleOutput))
	if r.Files[0] != "internal/foo/foo.go" {
		t.Errorf("got %q, want %q", r.Files[0], "internal/foo/foo.go")
	}
}

func TestParse_Empty(t *testing.T) {
	r, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.HasIssues() {
		t.Error("expected no issues for empty output")
	}
}

func TestParse_NotTruncated(t *testing.T) {
	r, _ := Parse(strings.NewReader(sampleOutput))
	if r.Truncated {
		t.Error("expected Truncated = false for 2 files")
	}
}

func TestParse_Truncated(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < MaxFindings+1; i++ {
		sb.WriteString("internal/pkg/file.go\n")
	}
	r, err := Parse(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Files) != MaxFindings {
		t.Errorf("got %d files, want %d", len(r.Files), MaxFindings)
	}
	if !r.Truncated {
		t.Error("expected Truncated = true")
	}
}
