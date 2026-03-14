package tester

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// Event mirrors a single line of `go test -json` output.
type Event struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"`
	Output  string    `json:"Output"`
}

// TestCase holds the result of a single test function.
type TestCase struct {
	Name    string
	Action  string   // "pass", "fail", "skip"
	Elapsed float64  // seconds
	Output  []string // raw output lines; populated for all actions
}

// PackageResult holds the aggregated result of one Go package.
type PackageResult struct {
	Name    string
	Action  string  // "pass", "fail", "skip"
	Elapsed float64 // seconds
	Cases   []*TestCase
}

// Results holds all parsed test results.
type Results struct {
	Packages   []*PackageResult
	BuildError string // non-empty when go test failed to build (not a test failure)
}

// Passed returns the total number of passing test cases.
func (r *Results) Passed() int { return r.countByAction("pass") }

// Failed returns the total number of failing test cases.
func (r *Results) Failed() int { return r.countByAction("fail") }

// Skipped returns the total number of skipped test cases.
func (r *Results) Skipped() int { return r.countByAction("skip") }

// HasFailures reports whether any test case failed.
func (r *Results) HasFailures() bool { return r.Failed() > 0 }

func (r *Results) countByAction(action string) int {
	n := 0
	for _, p := range r.Packages {
		for _, c := range p.Cases {
			if c.Action == action {
				n++
			}
		}
	}
	return n
}

// Run executes `go test -json <args>` in dir and returns parsed results.
// A non-zero exit from go test due to test failures is not treated as an error;
// failures are captured in the returned Results instead.
func Run(dir string, args []string) (*Results, error) {
	cmdArgs := append([]string{"test", "-json"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start go test: %w", err)
	}

	results, parseErr := Parse(stdout)
	if exitErr := cmd.Wait(); exitErr != nil && !results.HasFailures() {
		// Non-zero exit not explained by test failures means the code failed to build.
		errMsg := strings.TrimSpace(stderrBuf.String())
		if errMsg == "" {
			errMsg = exitErr.Error()
		}
		results.BuildError = errMsg
	}
	return results, parseErr
}

// Parse reads `go test -json` output from r and returns parsed results.
// Exposed to allow testing without running a real process.
func Parse(r io.Reader) (*Results, error) {
	pkgMap := map[string]*PackageResult{}
	pkgOrder := []string{}
	caseMap := map[string]*TestCase{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var ev Event
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			// skip non-JSON lines (build output, etc.)
			continue
		}

		if _, ok := pkgMap[ev.Package]; !ok {
			pkgMap[ev.Package] = &PackageResult{Name: ev.Package}
			pkgOrder = append(pkgOrder, ev.Package)
		}
		pkg := pkgMap[ev.Package]

		if ev.Test == "" {
			// Package-level event.
			switch ev.Action {
			case "pass", "fail", "skip":
				pkg.Action = ev.Action
				pkg.Elapsed = ev.Elapsed
			}
			continue
		}

		// Test-level event.
		key := ev.Package + "/" + ev.Test
		if _, ok := caseMap[key]; !ok {
			tc := &TestCase{Name: ev.Test}
			caseMap[key] = tc
			pkg.Cases = append(pkg.Cases, tc)
		}
		tc := caseMap[key]

		switch ev.Action {
		case "pass", "fail", "skip":
			tc.Action = ev.Action
			tc.Elapsed = ev.Elapsed
		case "output":
			tc.Output = append(tc.Output, ev.Output)
		}
	}

	results := &Results{}
	for _, name := range pkgOrder {
		results.Packages = append(results.Packages, pkgMap[name])
	}
	return results, nil
}
