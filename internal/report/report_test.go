package report

import (
	"strings"
	"testing"

	"github.com/saintedlama/goship/internal/coverage"
	"github.com/saintedlama/goship/internal/tester"
)

func singlePassResult() *tester.Results {
	return &tester.Results{
		Packages: []*tester.PackageResult{
			{
				Name:    "example.com/foo",
				Action:  "pass",
				Elapsed: 0.012,
				Cases: []*tester.TestCase{
					{Name: "TestAdd", Action: "pass", Elapsed: 0.001},
					{Name: "TestSub", Action: "pass", Elapsed: 0.001},
				},
			},
		},
	}
}

func singleFailResult() *tester.Results {
	return &tester.Results{
		Packages: []*tester.PackageResult{
			{
				Name:    "example.com/foo",
				Action:  "fail",
				Elapsed: 0.015,
				Cases: []*tester.TestCase{
					{Name: "TestAdd", Action: "pass", Elapsed: 0.001},
					{
						Name:    "TestBad",
						Action:  "fail",
						Elapsed: 0.002,
						Output:  []string{"    foo_test.go:10: expected 1 got 2\n"},
					},
				},
			},
		},
	}
}

func TestBuildMarkdown_PassHeading(t *testing.T) {
	md := BuildMarkdown(singlePassResult())
	if !strings.Contains(md, "🔭 Tests") {
		t.Error("expected 🔭 Tests heading")
	}
	if !strings.Contains(md, "2 total") {
		t.Error("expected \"2 total\" in summary")
	}
	if !strings.Contains(md, "2 passed") {
		t.Error("expected \"2 passed\" in summary")
	}
}

func TestBuildMarkdown_FailHeading(t *testing.T) {
	md := BuildMarkdown(singleFailResult())
	if !strings.Contains(md, "1 failed") {
		t.Error("expected \"1 failed\" in heading")
	}
}

func TestBuildMarkdown_PackageTable(t *testing.T) {
	md := BuildMarkdown(singlePassResult())
	if !strings.Contains(md, "example.com/foo") {
		t.Error("expected package name in table")
	}
}

func TestBuildMarkdown_FailedTestDetails(t *testing.T) {
	md := BuildMarkdown(singleFailResult())
	if !strings.Contains(md, "TestBad") {
		t.Error("expected failed test name in details")
	}
	if !strings.Contains(md, "expected 1 got 2") {
		t.Error("expected failure output in details")
	}
}

func TestBuildMarkdown_NoFailedSection_WhenAllPass(t *testing.T) {
	md := BuildMarkdown(singlePassResult())
	if strings.Contains(md, "Failed Tests") {
		t.Error("did not expect \"Failed Tests\" section when all pass")
	}
}

func multiPackagePassResult() *tester.Results {
	return &tester.Results{
		Packages: []*tester.PackageResult{
			{Name: "example.com/mymod/foo", Action: "pass", Elapsed: 0.01,
				Cases: []*tester.TestCase{{Name: "TestA", Action: "pass"}}},
			{Name: "example.com/mymod/bar", Action: "pass", Elapsed: 0.01,
				Cases: []*tester.TestCase{{Name: "TestB", Action: "pass"}}},
		},
	}
}

func multiPackageCoverageResults() *coverage.Results {
	return &coverage.Results{
		Total: 80.0,
		Packages: []*coverage.PackageCoverage{
			{Name: "example.com/mymod/foo", Percent: 90.0},
			{Name: "example.com/mymod/bar", Percent: 55.0},
		},
	}
}

func TestBuildMarkdown_CommonRootSubtitle(t *testing.T) {
	md := BuildMarkdown(multiPackagePassResult())
	if !strings.Contains(md, "example.com/mymod") {
		t.Error("expected common root label in output")
	}
}

func TestBuildMarkdown_StripsPrefix(t *testing.T) {
	md := BuildMarkdown(multiPackagePassResult())
	if strings.Contains(md, "`example.com/mymod/foo`") {
		t.Error("expected full path to be stripped in package table")
	}
	if !strings.Contains(md, "`foo`") {
		t.Error("expected short package name \"foo\" in table")
	}
}

func TestBuildCoverageMarkdown_CommonRootSubtitle(t *testing.T) {
	md := BuildCoverageMarkdown(multiPackageCoverageResults())
	if !strings.Contains(md, "example.com/mymod") {
		t.Error("expected common root label in coverage output")
	}
}

func TestBuildCoverageMarkdown_StripsPrefix(t *testing.T) {
	md := BuildCoverageMarkdown(multiPackageCoverageResults())
	if strings.Contains(md, "`example.com/mymod/foo`") {
		t.Error("expected full path to be stripped in coverage table")
	}
	if !strings.Contains(md, "`foo`") {
		t.Error("expected short package name \"foo\" in coverage table")
	}
}

func TestCommonRoot_TwoPackages(t *testing.T) {
	label, prefix := commonRoot([]string{"github.com/foo/bar", "github.com/foo/baz"})
	if label != "github.com/foo" {
		t.Errorf("label = %q, want \"github.com/foo\"", label)
	}
	if prefix != "github.com/foo/" {
		t.Errorf("prefix = %q, want \"github.com/foo/\"", prefix)
	}
}

func TestCommonRoot_SinglePackage(t *testing.T) {
	label, prefix := commonRoot([]string{"github.com/foo/bar"})
	if label != "" || prefix != "" {
		t.Errorf("expected empty strings for single package, got %q %q", label, prefix)
	}
}

func TestCommonRoot_NoCommonPrefix(t *testing.T) {
	label, prefix := commonRoot([]string{"github.com/foo", "gitlab.com/bar"})
	if label != "" || prefix != "" {
		t.Errorf("expected empty strings for no common prefix, got %q %q", label, prefix)
	}
}

func coverageResults(total float64) *coverage.Results {
	return &coverage.Results{
		Total: total,
		Packages: []*coverage.PackageCoverage{
			{Name: "example.com/foo", Percent: 90.0},
			{Name: "example.com/bar", Percent: 55.0},
		},
	}
}

func TestBuildCoverageMarkdown_Heading(t *testing.T) {
	md := BuildCoverageMarkdown(coverageResults(82.5))
	if !strings.Contains(md, "82.5%") {
		t.Error("expected total percentage in summary")
	}
	if !strings.Contains(md, "Coverage") {
		t.Error("expected Coverage in heading")
	}
}

func TestBuildCoverageMarkdown_PackageTable(t *testing.T) {
	md := BuildCoverageMarkdown(coverageResults(72.0))
	// common prefix "example.com/" is stripped; short names appear in table
	if !strings.Contains(md, "`foo`") {
		t.Error("expected short package name 'foo' in table")
	}
	if !strings.Contains(md, "`bar`") {
		t.Error("expected short package name 'bar' in table")
	}
	// the common root should appear as a subtitle
	if !strings.Contains(md, "example.com") {
		t.Error("expected common root shown as subtitle")
	}
}

func TestBuildCoverageMarkdown_Icons(t *testing.T) {
	md := BuildCoverageMarkdown(coverageResults(72.0))
	// foo is 90% → anchor, bar is 55% → lifebuoy
	if !strings.Contains(md, "⚓️") {
		t.Error("expected ⚓️ icon for ≥80% package")
	}
	if !strings.Contains(md, "🛟") {
		t.Error("expected 🛟 icon for <60% package")
	}
}
