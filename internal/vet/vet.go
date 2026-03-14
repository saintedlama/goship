package vet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// MaxFindings is the maximum number of findings included in Results.
// If more findings exist, Results.Truncated is set to true.
const MaxFindings = 100

// diagnostic mirrors a single entry from `go vet -json` output.
type diagnostic struct {
	Posn    string `json:"posn"`
	Message string `json:"message"`
}

// Finding represents a single vet warning.
type Finding struct {
	Package  string
	Analyzer string
	Posn     string // file:line:col (relative to working directory)
	Message  string
}

// Results holds all parsed vet findings.
type Results struct {
	Packages   int // total number of packages checked
	Findings   []*Finding
	Truncated  bool   // true when more than MaxFindings were found
	BuildError string // non-empty when go vet failed due to a build failure
}

// HasIssues reports whether any vet issues were found.
func (r *Results) HasIssues() bool {
	return len(r.Findings) > 0
}

// Run executes `go vet -json ./...` in dir and returns parsed results.
// With -json, go vet exits 0 even when issues are present; issues appear in
// the JSON output and are captured in the returned Results.
func Run(dir string) (*Results, error) {
	cmd := exec.Command("go", "vet", "-json", "./...")
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start go vet: %w", err)
	}

	results, parseErr := Parse(stdout)
	if exitErr := cmd.Wait(); exitErr != nil {
		// go vet -json exits 0 for vet findings; any non-zero exit means build failed.
		errMsg := strings.TrimSpace(stderrBuf.String())
		if errMsg == "" {
			errMsg = exitErr.Error()
		}
		results.BuildError = errMsg
	}

	// Strip the working-directory prefix from posns for readable output.
	if dir != "" {
		prefix := dir + "/"
		for _, f := range results.Findings {
			f.Posn = strings.TrimPrefix(f.Posn, prefix)
		}
	}
	return results, parseErr
}

// Parse reads `go vet -json` output from r and returns parsed results.
// Exported to allow testing without running a real process.
//
// go vet -json emits one JSON document per package:
//
//	{"pkg/name": {"analyzer": [{"posn": "file:line:col", "message": "..."}]}}
//
// An empty document {} means the package had no issues.
func Parse(r io.Reader) (*Results, error) {
	dec := json.NewDecoder(r)
	var all []*Finding
	var pkgs int

	for {
		var record map[string]map[string][]diagnostic
		if err := dec.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("parse vet output: %w", err)
		}
		pkgs++
		for pkg, analyzers := range record {
			for analyzer, diags := range analyzers {
				for _, d := range diags {
					all = append(all, &Finding{
						Package:  pkg,
						Analyzer: analyzer,
						Posn:     d.Posn,
						Message:  d.Message,
					})
				}
			}
		}
	}

	truncated := false
	if len(all) > MaxFindings {
		all = all[:MaxFindings]
		truncated = true
	}
	return &Results{Packages: pkgs, Findings: all, Truncated: truncated}, nil
}
