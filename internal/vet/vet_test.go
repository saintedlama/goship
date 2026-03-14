package vet

import (
	"strings"
	"testing"
)

const sampleOutput = `{}
{
	"example.com/foo": {
		"printf": [
			{
				"posn": "/work/example.com/foo/a.go:10:2",
				"end": "/work/example.com/foo/a.go:10:5",
				"message": "fmt.Printf format %d has arg of wrong type"
			}
		]
	}
}
{
	"example.com/bar": {
		"shadow": [
			{
				"posn": "/work/example.com/bar/b.go:5:1",
				"end": "/work/example.com/bar/b.go:5:3",
				"message": "declaration of err shadows declaration at b.go:1"
			}
		]
	}
}`

func TestParse_FindingCount(t *testing.T) {
	r, err := Parse(strings.NewReader(sampleOutput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Findings) != 2 {
		t.Errorf("got %d findings, want 2", len(r.Findings))
	}
}

func TestParse_HasIssues(t *testing.T) {
	r, _ := Parse(strings.NewReader(sampleOutput))
	if !r.HasIssues() {
		t.Error("expected HasIssues() = true")
	}
}

func TestParse_FindingContent(t *testing.T) {
	r, _ := Parse(strings.NewReader(sampleOutput))
	var found *Finding
	for _, f := range r.Findings {
		if f.Analyzer == "printf" {
			found = f
			break
		}
	}
	if found == nil {
		t.Fatal("expected a printf finding")
	}
	if found.Package != "example.com/foo" {
		t.Errorf("Package = %q, want %q", found.Package, "example.com/foo")
	}
	if !strings.Contains(found.Message, "format") {
		t.Errorf("unexpected message: %q", found.Message)
	}
}

func TestParse_Empty(t *testing.T) {
	r, err := Parse(strings.NewReader("{}"))
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
		t.Error("expected Truncated = false for 2 findings")
	}
}

func TestParse_Truncated(t *testing.T) {
	// Build a JSON document with MaxFindings+1 findings.
	var sb strings.Builder
	sb.WriteString(`{"example.com/big":{"printf":[`)
	for i := 0; i < MaxFindings+1; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(`{"posn":"a.go:1:1","message":"msg"}`)
	}
	sb.WriteString(`]}}`)

	r, err := Parse(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Findings) != MaxFindings {
		t.Errorf("got %d findings, want %d", len(r.Findings), MaxFindings)
	}
	if !r.Truncated {
		t.Error("expected Truncated = true")
	}
}
