package report

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
	assert.Contains(t, md, "🔭 Tests")
	assert.Contains(t, md, "2 total")
	assert.Contains(t, md, "2 passed")
}

func TestBuildMarkdown_FailHeading(t *testing.T) {
	md := BuildMarkdown(singleFailResult())
	assert.Contains(t, md, "1 failed")
}

func TestBuildMarkdown_PackageTable(t *testing.T) {
	md := BuildMarkdown(singlePassResult())
	assert.Contains(t, md, "example.com/foo")
}

func TestBuildMarkdown_FailedTestDetails(t *testing.T) {
	md := BuildMarkdown(singleFailResult())
	assert.Contains(t, md, "TestBad")
	assert.Contains(t, md, "expected 1 got 2")
}

func TestBuildMarkdown_NoFailedSection_WhenAllPass(t *testing.T) {
	md := BuildMarkdown(singlePassResult())
	assert.NotContains(t, md, "Failed Tests")
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
	assert.Contains(t, md, "example.com/mymod")
}

func TestBuildMarkdown_StripsPrefix(t *testing.T) {
	md := BuildMarkdown(multiPackagePassResult())
	assert.NotContains(t, md, "`example.com/mymod/foo`")
	assert.Contains(t, md, "`foo`")
}

func TestBuildCoverageMarkdown_CommonRootSubtitle(t *testing.T) {
	md := BuildCoverageMarkdown(multiPackageCoverageResults())
	assert.Contains(t, md, "example.com/mymod")
}

func TestBuildCoverageMarkdown_StripsPrefix(t *testing.T) {
	md := BuildCoverageMarkdown(multiPackageCoverageResults())
	assert.NotContains(t, md, "`example.com/mymod/foo`")
	assert.Contains(t, md, "`foo`")
}

func TestCommonRoot_TwoPackages(t *testing.T) {
	label, prefix := commonRoot([]string{"github.com/foo/bar", "github.com/foo/baz"})
	assert.Equal(t, "github.com/foo", label)
	assert.Equal(t, "github.com/foo/", prefix)
}

func TestCommonRoot_SinglePackage(t *testing.T) {
	label, prefix := commonRoot([]string{"github.com/foo/bar"})
	assert.Empty(t, label)
	assert.Empty(t, prefix)
}

func TestCommonRoot_NoCommonPrefix(t *testing.T) {
	label, prefix := commonRoot([]string{"github.com/foo", "gitlab.com/bar"})
	assert.Empty(t, label)
	assert.Empty(t, prefix)
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
	assert.Contains(t, md, "82.5%")
	assert.Contains(t, md, "Coverage")
}

func TestBuildCoverageMarkdown_PackageTable(t *testing.T) {
	md := BuildCoverageMarkdown(coverageResults(72.0))
	// common prefix "example.com/" is stripped; short names appear in table
	assert.Contains(t, md, "`foo`")
	assert.Contains(t, md, "`bar`")
	// the common root should appear as a subtitle
	assert.Contains(t, md, "example.com")
}

func TestBuildCoverageMarkdown_Icons(t *testing.T) {
	md := BuildCoverageMarkdown(coverageResults(72.0))
	// foo is 90% → anchor, bar is 55% → lifebuoy
	assert.Contains(t, md, "⚓️")
	assert.Contains(t, md, "🛟")
}
