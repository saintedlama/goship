package action

import (
	"fmt"
	"os"
	"strings"

	"github.com/saintedlama/goship/internal/coverage"
	"github.com/saintedlama/goship/internal/format"
	"github.com/saintedlama/goship/internal/report"
	"github.com/saintedlama/goship/internal/tester"
	"github.com/saintedlama/goship/internal/vet"
)

// Config holds all configuration parsed from GitHub Actions inputs.
type Config struct {
	Token            string
	WorkingDirectory string
	Test             bool // whether to run go test and report results
	Coverage         bool // whether to collect and report coverage
	Vet              bool // whether to run go vet and report results
	Fmt              bool // whether to run gofmt and report unformatted files
}

// Run executes the action. It returns false when any check fails (so the
// caller can exit with a non-zero code) and a non-nil error only when the
// action itself cannot complete (e.g. cannot change directory, cannot write
// summary). All enabled steps are always run regardless of prior failures.
func Run(cfg Config) (passed bool, err error) {
	if err := os.Chdir(cfg.WorkingDirectory); err != nil {
		return false, fmt.Errorf("change working directory to %q: %w", cfg.WorkingDirectory, err)
	}

	allPassed := true

	if cfg.Test {
		testArgs := []string{"./..."}

		// When coverage is requested, create a temp file and append -coverprofile so
		// the single test run produces both JSON output and a coverage profile.
		var profilePath string
		if cfg.Coverage {
			f, err := os.CreateTemp("", "goship-cover-*.out")
			if err != nil {
				return false, fmt.Errorf("create coverage profile: %w", err)
			}
			f.Close()
			defer os.Remove(f.Name())
			profilePath = f.Name()
			testArgs = append(testArgs, "-coverprofile="+profilePath)
		}

		results, err := tester.Run(cfg.WorkingDirectory, testArgs)
		if err != nil {
			return false, fmt.Errorf("run tests: %w", err)
		}
		if err := report.WriteStepSummary(results); err != nil {
			return false, fmt.Errorf("write test summary: %w", err)
		}
		if results.HasFailures() {
			allPassed = false
		}
		if results.BuildError != "" {
			allPassed = false
		}

		if cfg.Coverage && profilePath != "" && results.BuildError == "" {
			covResults, err := coverage.Run(cfg.WorkingDirectory, profilePath)
			if err != nil {
				return false, fmt.Errorf("collect coverage: %w", err)
			}
			if err := report.WriteCoverageSection(covResults); err != nil {
				return false, fmt.Errorf("write coverage summary: %w", err)
			}
		}
	}

	if cfg.Vet {
		vetResults, err := vet.Run(cfg.WorkingDirectory)
		if err != nil {
			return false, fmt.Errorf("run vet: %w", err)
		}
		if err := report.WriteVetSection(vetResults); err != nil {
			return false, fmt.Errorf("write vet summary: %w", err)
		}
		if vetResults.HasIssues() {
			allPassed = false
		}
		if vetResults.BuildError != "" {
			allPassed = false
		}
	}

	if cfg.Fmt {
		fmtResults, err := format.Run(cfg.WorkingDirectory)
		if err != nil {
			return false, fmt.Errorf("run fmt: %w", err)
		}
		if err := report.WriteFmtSection(fmtResults); err != nil {
			return false, fmt.Errorf("write fmt summary: %w", err)
		}
		if fmtResults.HasIssues() {
			allPassed = false
		}
	}

	return allPassed, nil
}

// ParseBool returns true for "1", "true", "yes" (case-insensitive), else false.
func ParseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		return true
	}
	return false
}
