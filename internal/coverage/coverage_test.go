package coverage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sample profile: foo has 3+2=5 stmts, 3 covered (first block only).
// baz has 1 stmt, 1 covered.
// Total: 4/6 = 66.7%
const sampleProfile = `mode: set
github.com/foo/bar/file.go:10.20,15.5 3 1
github.com/foo/bar/file.go:16.3,18.12 2 0
github.com/foo/baz/baz.go:5.10,8.3 1 1
`

func TestParseProfile_Total(t *testing.T) {
	r, err := ParseProfile(strings.NewReader(sampleProfile))
	require.NoError(t, err)
	want := float64(4) / float64(6) * 100
	assert.Equal(t, want, r.Total)
}

func TestParseProfile_PackageCount(t *testing.T) {
	r, err := ParseProfile(strings.NewReader(sampleProfile))
	require.NoError(t, err)
	assert.Len(t, r.Packages, 2)
}

func TestParseProfile_PackageNames(t *testing.T) {
	r, err := ParseProfile(strings.NewReader(sampleProfile))
	require.NoError(t, err)
	assert.Equal(t, "github.com/foo/bar", r.Packages[0].Name)
	assert.Equal(t, "github.com/foo/baz", r.Packages[1].Name)
}

func TestParseProfile_PackagePercent_Weighted(t *testing.T) {
	r, err := ParseProfile(strings.NewReader(sampleProfile))
	require.NoError(t, err)
	// bar: 3 covered out of 5 = 60%
	want := float64(3) / float64(5) * 100
	assert.Equal(t, want, r.Packages[0].Percent)
}

func TestParseProfile_FullCoverage(t *testing.T) {
	profile := "mode: set\ngithub.com/foo/bar/f.go:1.1,2.1 5 1\n"
	r, err := ParseProfile(strings.NewReader(profile))
	require.NoError(t, err)
	assert.Equal(t, 100.0, r.Total)
	assert.Equal(t, 100.0, r.Packages[0].Percent)
}

func TestParseProfile_ZeroCoverage(t *testing.T) {
	profile := "mode: set\ngithub.com/foo/bar/f.go:1.1,2.1 5 0\n"
	r, err := ParseProfile(strings.NewReader(profile))
	require.NoError(t, err)
	assert.Equal(t, 0.0, r.Total)
}

func TestParseProfile_Empty(t *testing.T) {
	r, err := ParseProfile(strings.NewReader("mode: set\n"))
	require.NoError(t, err)
	assert.Empty(t, r.Packages)
}
