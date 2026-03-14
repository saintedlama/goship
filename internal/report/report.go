package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/saintedlama/goship/internal/coverage"
	"github.com/saintedlama/goship/internal/format"
	"github.com/saintedlama/goship/internal/tester"
	"github.com/saintedlama/goship/internal/vet"
)

// WriteStepSummary appends a Markdown test report to $GITHUB_STEP_SUMMARY.
// When the env var is not set (local development), it writes to stdout.
func WriteStepSummary(results *tester.Results) error {
	md := BuildMarkdown(results)

	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile == "" {
		fmt.Print(md)
		return nil
	}

	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open GITHUB_STEP_SUMMARY: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprint(f, md)
	return err
}

// BuildMarkdown renders results as a Markdown string suitable for GitHub Step Summary.
// Exported so it can be used in tests.
func BuildMarkdown(results *tester.Results) string {
	var b strings.Builder

	pkgNames := make([]string, len(results.Packages))
	for i, p := range results.Packages {
		pkgNames[i] = p.Name
	}
	root, prefix := commonRoot(pkgNames)

	passed := results.Passed()
	failed := results.Failed()
	skipped := results.Skipped()
	total := passed + failed + skipped

	fmt.Fprintln(&b, "## 🔭 Tests")
	fmt.Fprintln(&b)
	if results.BuildError != "" {
		fmt.Fprintln(&b, "build failed")
	} else if failed > 0 {
		fmt.Fprintf(&b, "%d total, %d failed, %d passed", total, failed, passed)
		if skipped > 0 {
			fmt.Fprintf(&b, ", %d skipped", skipped)
		}
		fmt.Fprintln(&b)
	} else {
		fmt.Fprintf(&b, "%d total, %d passed", total, passed)
		if skipped > 0 {
			fmt.Fprintf(&b, ", %d skipped", skipped)
		}
		fmt.Fprintln(&b)
	}
	if root != "" {
		fmt.Fprintf(&b, "_%s_\n", root)
	}
	fmt.Fprintln(&b)

	// Per-package table.
	fmt.Fprintln(&b, "| | Package | Tests | Duration |")
	fmt.Fprintln(&b, "|---|---------|-------|----------|")
	for _, pkg := range results.Packages {
		passed, failed, skipped := packageCounts(pkg)
		counts := formatCounts(passed, failed, skipped)
		short := stripPrefix(pkg.Name, prefix)
		fmt.Fprintf(&b, "| %s | `%s` | %s | %.3fs |\n", actionIcon(pkg.Action), short, counts, pkg.Elapsed)
	}

	// Details for failed test cases.
	var failedCases []failedCase
	for _, pkg := range results.Packages {
		for _, tc := range pkg.Cases {
			if tc.Action == "fail" {
				failedCases = append(failedCases, failedCase{pkg: pkg, tc: tc})
			}
		}
	}

	if len(failedCases) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "### 🌊 Failed Tests")
		for _, fc := range failedCases {
			fmt.Fprintln(&b)
			fmt.Fprintf(&b, "#### `%s` (%.3fs)\n", fc.tc.Name, fc.tc.Elapsed)
			fmt.Fprintf(&b, "_Package: `%s`_\n", fc.pkg.Name)
			if len(fc.tc.Output) > 0 {
				fmt.Fprintln(&b)
				fmt.Fprintln(&b, "```")
				for _, line := range fc.tc.Output {
					fmt.Fprint(&b, line)
				}
				fmt.Fprintln(&b, "```")
			}
		}
	}

	if results.BuildError != "" {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "### 🌊 Build Output")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "```")
		fmt.Fprintln(&b, results.BuildError)
		fmt.Fprintln(&b, "```")
	}

	return b.String()
}

type failedCase struct {
	pkg *tester.PackageResult
	tc  *tester.TestCase
}

func packageCounts(pkg *tester.PackageResult) (passed, failed, skipped int) {
	for _, tc := range pkg.Cases {
		switch tc.Action {
		case "pass":
			passed++
		case "fail":
			failed++
		case "skip":
			skipped++
		}
	}
	return
}

func formatCounts(passed, failed, skipped int) string {
	var parts []string
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	if passed > 0 {
		parts = append(parts, fmt.Sprintf("%d passed", passed))
	}
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", skipped))
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, ", ")
}

func actionIcon(action string) string {
	switch action {
	case "pass":
		return "⚓️"
	case "fail":
		return "🌊"
	case "skip":
		return "⏭️"
	default:
		return "❓"
	}
}

// WriteVetSection appends a Markdown vet report to $GITHUB_STEP_SUMMARY.
// When the env var is not set (local development), it writes to stdout.
func WriteVetSection(results *vet.Results) error {
	md := BuildVetMarkdown(results)

	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile == "" {
		fmt.Print(md)
		return nil
	}

	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open GITHUB_STEP_SUMMARY: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprint(f, md)
	return err
}

// BuildVetMarkdown renders vet results as a Markdown string.
// Exported so it can be used in tests.
func BuildVetMarkdown(results *vet.Results) string {
	var b strings.Builder

	fmt.Fprintln(&b, "## 🧭 Vet")
	fmt.Fprintln(&b)

	// Clean pass — no issues and no build error.
	if !results.HasIssues() && results.BuildError == "" {
		fmt.Fprintf(&b, "⚓️ passed for %d package(s)\n", results.Packages)
		fmt.Fprintln(&b)
		return b.String()
	}

	// Summary paragraph.
	switch {
	case results.BuildError != "" && !results.HasIssues():
		fmt.Fprintln(&b, "build failed")
	case results.BuildError != "":
		fmt.Fprintf(&b, "%d issue(s), build failed\n", len(results.Findings))
	default:
		fmt.Fprintf(&b, "%d issue(s)\n", len(results.Findings))
	}
	fmt.Fprintln(&b)

	// Findings table.
	if results.HasIssues() {
		fmt.Fprintln(&b, "| Analyzer | Location | Message |")
		fmt.Fprintln(&b, "|----------|----------|---------|")
		for _, f := range results.Findings {
			fmt.Fprintf(&b, "| `%s` | `%s` | %s |\n", f.Analyzer, f.Posn, f.Message)
		}
		if results.Truncated {
			fmt.Fprintln(&b)
			fmt.Fprintf(&b, "> ⚠️ Output limited to %d findings. Fix these first to see all issues.\n", vet.MaxFindings)
		}
	}

	// Build error output.
	if results.BuildError != "" {
		if results.HasIssues() {
			fmt.Fprintln(&b)
		}
		fmt.Fprintln(&b, "### 🌊 Build Output")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "```")
		fmt.Fprintln(&b, results.BuildError)
		fmt.Fprintln(&b, "```")
	}

	fmt.Fprintln(&b)
	return b.String()
}

// WriteFmtSection appends a Markdown gofmt report to $GITHUB_STEP_SUMMARY.
// When the env var is not set (local development), it writes to stdout.
func WriteFmtSection(results *format.Results) error {
	md := BuildFmtMarkdown(results)

	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile == "" {
		fmt.Print(md)
		return nil
	}

	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open GITHUB_STEP_SUMMARY: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprint(f, md)
	return err
}

// BuildFmtMarkdown renders gofmt results as a Markdown string.
// Exported so it can be used in tests.
func BuildFmtMarkdown(results *format.Results) string {
	var b strings.Builder

	fmt.Fprintln(&b, "## 🏄🏾 Fmt")
	fmt.Fprintln(&b)

	if !results.HasIssues() {
		fmt.Fprintf(&b, "⚓️ passed for %d file(s)\n", results.Total)
		fmt.Fprintln(&b)
		return b.String()
	}

	fmt.Fprintf(&b, "%d file(s) need formatting\n", len(results.Files))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| File |")
	fmt.Fprintln(&b, "|------|")
	for _, f := range results.Files {
		fmt.Fprintf(&b, "| `%s` |\n", f)
	}
	if results.Truncated {
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "> ⚠️ Output limited to %d files. Run `gofmt -w .` to fix all at once.\n", format.MaxFindings)
	}
	fmt.Fprintln(&b)
	return b.String()
}

// WriteCoverageSection appends a Markdown coverage section to $GITHUB_STEP_SUMMARY.
// When the env var is not set (local development), it writes to stdout.
func WriteCoverageSection(results *coverage.Results) error {
	md := BuildCoverageMarkdown(results)

	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile == "" {
		fmt.Print(md)
		return nil
	}

	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open GITHUB_STEP_SUMMARY: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprint(f, md)
	return err
}

// BuildCoverageMarkdown renders coverage results as a Markdown string.
// Exported so it can be used in tests.
func BuildCoverageMarkdown(results *coverage.Results) string {
	var b strings.Builder

	pkgNames := make([]string, len(results.Packages))
	for i, p := range results.Packages {
		pkgNames[i] = p.Name
	}
	root, prefix := commonRoot(pkgNames)

	fmt.Fprintln(&b, "## \U0001f5faï¸ Coverage")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "%d package(s), %.1f%% coverage\n", len(results.Packages), results.Total)
	if root != "" {
		fmt.Fprintf(&b, "_%s_\n", root)
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "| | Package | Coverage |")
	fmt.Fprintln(&b, "|---|---------|----------|")
	for _, pkg := range results.Packages {
		short := stripPrefix(pkg.Name, prefix)
		fmt.Fprintf(&b, "| %s | `%s` | %.1f%% |\n", coverageIcon(pkg.Percent), short, pkg.Percent)
	}
	fmt.Fprintln(&b)

	return b.String()
}

// commonRoot finds the longest common slash-delimited prefix shared by all
// package names and returns a human-readable label (e.g. "github.com/foo/bar")
// and the raw prefix string to strip from each name (including trailing slash).
// Returns ("", "") when there is only one package or no common prefix.
func commonRoot(names []string) (label, prefix string) {
	if len(names) <= 1 {
		return "", ""
	}
	parts := strings.Split(names[0], "/")
	for _, name := range names[1:] {
		other := strings.Split(name, "/")
		i := 0
		for i < len(parts) && i < len(other) && parts[i] == other[i] {
			i++
		}
		parts = parts[:i]
	}
	if len(parts) == 0 {
		return "", ""
	}
	root := strings.Join(parts, "/")
	return root, root + "/"
}

// stripPrefix removes a known prefix from a package name. If the name equals
// the prefix root exactly (trailing slash stripped), it returns "." so the
// root package itself is still legible in the table.
func stripPrefix(name, prefix string) string {
	if prefix == "" {
		return name
	}
	trimmed := strings.TrimPrefix(name, prefix)
	if trimmed == "" || trimmed == name {
		return name
	}
	return trimmed
}

// coverageIcon returns a nautical emoji based on the coverage percentage.
// ≥80% ⚓️  ≥60% ⛵️  <60% 🛟
func coverageIcon(pct float64) string {
	switch {
	case pct >= 80:
		return "⚓️"
	case pct >= 60:
		return "⛵️"
	default:
		return "🛟"
	}
}
