package vet

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	assert.Len(t, r.Findings, 2)
}

func TestParse_HasIssues(t *testing.T) {
	r, err := Parse(strings.NewReader(sampleOutput))
	require.NoError(t, err)
	assert.True(t, r.HasIssues())
}

func TestParse_FindingContent(t *testing.T) {
	r, err := Parse(strings.NewReader(sampleOutput))
	require.NoError(t, err)
	var found *Finding
	for _, f := range r.Findings {
		if f.Analyzer == "printf" {
			found = f
			break
		}
	}
	require.NotNil(t, found, "expected a printf finding")
	assert.Equal(t, "example.com/foo", found.Package)
	assert.Contains(t, found.Message, "format")
}

func TestParse_Empty(t *testing.T) {
	r, err := Parse(strings.NewReader("{}"))
	require.NoError(t, err)
	assert.False(t, r.HasIssues())
}

func TestParse_NotTruncated(t *testing.T) {
	r, err := Parse(strings.NewReader(sampleOutput))
	require.NoError(t, err)
	assert.False(t, r.Truncated)
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
	require.NoError(t, err)
	assert.Len(t, r.Findings, MaxFindings)
	assert.True(t, r.Truncated)
}
