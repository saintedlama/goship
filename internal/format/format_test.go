package format

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleOutput = `internal/foo/foo.go
internal/bar/bar.go
`

func TestParse_FileCount(t *testing.T) {
	r, err := Parse(strings.NewReader(sampleOutput))
	require.NoError(t, err)
	assert.Len(t, r.Files, 2)
}

func TestParse_HasIssues(t *testing.T) {
	r, err := Parse(strings.NewReader(sampleOutput))
	require.NoError(t, err)
	assert.True(t, r.HasIssues())
}

func TestParse_FileNames(t *testing.T) {
	r, err := Parse(strings.NewReader(sampleOutput))
	require.NoError(t, err)
	assert.Equal(t, "internal/foo/foo.go", r.Files[0])
}

func TestParse_Empty(t *testing.T) {
	r, err := Parse(strings.NewReader(""))
	require.NoError(t, err)
	assert.False(t, r.HasIssues())
}

func TestParse_NotTruncated(t *testing.T) {
	r, err := Parse(strings.NewReader(sampleOutput))
	require.NoError(t, err)
	assert.False(t, r.Truncated)
}

func TestParse_Truncated(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < MaxFindings+1; i++ {
		sb.WriteString("internal/pkg/file.go\n")
	}
	r, err := Parse(strings.NewReader(sb.String()))
	require.NoError(t, err)
	assert.Len(t, r.Files, MaxFindings)
	assert.True(t, r.Truncated)
}
