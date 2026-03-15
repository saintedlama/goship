package action

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/saintedlama/goship/internal/coverage"
	"github.com/saintedlama/goship/internal/format"
	"github.com/saintedlama/goship/internal/report"
	"github.com/saintedlama/goship/internal/tester"
	"github.com/saintedlama/goship/internal/vet"
)

// StepStatus represents the outcome of a single action step.
type StepStatus string

const (
	StatusPassed   StepStatus = "passed"
	StatusFailed   StepStatus = "failed"
	StatusDisabled StepStatus = "disabled"
)

// Result holds the per-step outcome of a Run.
type Result struct {
	Test     StepStatus
	Coverage StepStatus
	Vet      StepStatus
	Fmt      StepStatus
}

// Passed reports whether every enabled step passed.
func (r Result) Passed() bool {
	return r.Test != StatusFailed &&
		r.Coverage != StatusFailed &&
		r.Vet != StatusFailed &&
		r.Fmt != StatusFailed
}

// Config holds all configuration parsed from GitHub Actions inputs.
type Config struct {
	Token            string
	WorkingDirectory string
	Test             bool // whether to run go test and report results
	Coverage         bool // whether to collect and report coverage
	Vet              bool // whether to run go vet and report results
	Fmt              bool // whether to run gofmt and report unformatted files
}

// Run executes the action. It returns a Result describing each step's
// outcome and a non-nil error only when the action itself cannot complete
// (e.g. cannot change directory, cannot write summary). All enabled steps
// are always run regardless of prior failures.
func Run(cfg Config) error {
	slog.Info("changing working directory", "dir", cfg.WorkingDirectory)
	if err := os.Chdir(cfg.WorkingDirectory); err != nil {
		slog.Error("failed to change directory", "err", err)
		return fmt.Errorf("change working directory to %q: %w", cfg.WorkingDirectory, err)
	}

	testResults, coverageResults, err := runTests(cfg)

	if err != nil {
		return fmt.Errorf("run tests: %w", err)
	}

	vetResults, err := runVet(cfg)
	if err != nil {
		return fmt.Errorf("run vet: %w", err)
	}

	fmtResults, err := runFmt(cfg)
	if err != nil {
		return fmt.Errorf("run fmt: %w", err)
	}

	// Write summary
	if testResults == nil {
		testResults = &tester.Results{}
	}
	if err := report.WriteStepSummary(testResults, coverageResults, vetResults, fmtResults); err != nil {
		slog.Error("failed to write step summary", "err", err)
		return fmt.Errorf("write step summary: %w", err)
	}

	return nil
}

// runTests runs go test and optionally collects a coverage profile.
// Returns nil results (not an error) when cfg.Test is disabled.
func runTests(cfg Config) (*tester.Results, *coverage.Results, error) {
	if !cfg.Test {
		return nil, nil, nil
	}

	slog.Info("running tests")
	testArgs := []string{"./..."}
	var profilePath string
	if cfg.Coverage {
		slog.Info("coverage enabled, preparing coverage profile")
		f, err := os.CreateTemp("", "goship-cover-*.out")
		if err != nil {
			slog.Error("failed to create coverage profile", "err", err)
			return nil, nil, fmt.Errorf("create coverage profile: %w", err)
		}
		f.Close()
		defer os.Remove(f.Name())
		profilePath = f.Name()
		testArgs = append(testArgs, "-coverprofile="+profilePath)
	}

	results, err := tester.Run(cfg.WorkingDirectory, testArgs)
	if err != nil {
		slog.Error("test execution failed", "err", err)
		return nil, nil, fmt.Errorf("run tests: %w", err)
	}
	slog.Info("tests completed", "passed", results.Passed(), "failed", results.Failed(), "skipped", results.Skipped())
	if results.HasFailures() {
		slog.Warn("some tests failed")
		for _, pkg := range results.Packages {
			for _, tc := range pkg.Cases {
				if tc.Action == "fail" {
					slog.Warn("test failed", "package", pkg.Name, "test", tc.Name)
				}
			}
		}
	}
	if results.BuildError != "" {
		slog.Error("build error during tests", "buildError", results.BuildError)
	}

	var coverageResults *coverage.Results
	if cfg.Coverage && profilePath != "" && results.BuildError == "" {
		slog.Info("collecting coverage")
		covResults, err := coverage.Run(cfg.WorkingDirectory, profilePath)
		if err != nil {
			slog.Error("failed to collect coverage", "err", err)
			return results, nil, fmt.Errorf("collect coverage: %w", err)
		}
		coverageResults = covResults
	}

	return results, coverageResults, nil
}

// runVet runs go vet.
// Returns nil results (not an error) when cfg.Vet is disabled.
func runVet(cfg Config) (*vet.Results, error) {
	if !cfg.Vet {
		return nil, nil
	}

	slog.Info("running vet")
	results, err := vet.Run(cfg.WorkingDirectory)
	if err != nil {
		slog.Error("vet execution failed", "err", err)
		return nil, fmt.Errorf("run vet: %w", err)
	}
	if results.HasIssues() {
		slog.Warn("vet found issues", "count", len(results.Findings))
		for _, f := range results.Findings {
			slog.Warn("vet finding", "analyzer", f.Analyzer, "location", f.Posn, "message", f.Message)
		}
	}
	if results.BuildError != "" {
		slog.Error("build error during vet", "buildError", results.BuildError)
	}
	return results, nil
}

// runFmt runs gofmt.
// Returns nil results (not an error) when cfg.Fmt is disabled.
func runFmt(cfg Config) (*format.Results, error) {
	if !cfg.Fmt {
		return nil, nil
	}

	slog.Info("running gofmt")
	results, err := format.Run(cfg.WorkingDirectory)
	if err != nil {
		slog.Error("gofmt execution failed", "err", err)
		return nil, fmt.Errorf("run fmt: %w", err)
	}
	if results.HasIssues() {
		slog.Warn("files need formatting", "count", len(results.Files))
		for _, f := range results.Files {
			slog.Warn("unformatted file", "file", f)
		}
	}
	return results, nil
}

// ParseBool returns true for "1", "true", "yes" (case-insensitive), else false.
func ParseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		return true
	}
	return false
}
