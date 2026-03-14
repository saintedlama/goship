package coverage

import (
	"strings"
	"testing"
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
	if err != nil {
		t.Fatal(err)
	}
	want := float64(4) / float64(6) * 100
	if r.Total != want {
		t.Errorf("Total = %.4f, want %.4f", r.Total, want)
	}
}

func TestParseProfile_PackageCount(t *testing.T) {
	r, err := ParseProfile(strings.NewReader(sampleProfile))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Packages) != 2 {
		t.Errorf("len(Packages) = %d, want 2", len(r.Packages))
	}
}

func TestParseProfile_PackageNames(t *testing.T) {
	r, err := ParseProfile(strings.NewReader(sampleProfile))
	if err != nil {
		t.Fatal(err)
	}
	if r.Packages[0].Name != "github.com/foo/bar" {
		t.Errorf("Packages[0].Name = %q, want \"github.com/foo/bar\"", r.Packages[0].Name)
	}
	if r.Packages[1].Name != "github.com/foo/baz" {
		t.Errorf("Packages[1].Name = %q, want \"github.com/foo/baz\"", r.Packages[1].Name)
	}
}

func TestParseProfile_PackagePercent_Weighted(t *testing.T) {
	r, err := ParseProfile(strings.NewReader(sampleProfile))
	if err != nil {
		t.Fatal(err)
	}
	// bar: 3 covered out of 5 = 60%
	want := float64(3) / float64(5) * 100
	got := r.Packages[0].Percent
	if got != want {
		t.Errorf("Packages[0].Percent = %.4f, want %.4f", got, want)
	}
}

func TestParseProfile_FullCoverage(t *testing.T) {
	profile := "mode: set\ngithub.com/foo/bar/f.go:1.1,2.1 5 1\n"
	r, err := ParseProfile(strings.NewReader(profile))
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 100.0 {
		t.Errorf("Total = %.1f, want 100.0", r.Total)
	}
	if r.Packages[0].Percent != 100.0 {
		t.Errorf("Percent = %.1f, want 100.0", r.Packages[0].Percent)
	}
}

func TestParseProfile_ZeroCoverage(t *testing.T) {
	profile := "mode: set\ngithub.com/foo/bar/f.go:1.1,2.1 5 0\n"
	r, err := ParseProfile(strings.NewReader(profile))
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 0.0 {
		t.Errorf("Total = %.1f, want 0.0", r.Total)
	}
}

func TestParseProfile_Empty(t *testing.T) {
	r, err := ParseProfile(strings.NewReader("mode: set\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Packages) != 0 {
		t.Errorf("expected no packages for empty profile")
	}
}
