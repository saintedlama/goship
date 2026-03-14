package format

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MaxFindings is the maximum number of unformatted files included in Results.
// If more files are unformatted, Results.Truncated is set to true.
const MaxFindings = 100

// Results holds all files reported as unformatted by gofmt.
type Results struct {
	Total     int      // total number of .go files checked
	Files     []string // paths relative to the working directory
	Truncated bool     // true when more than MaxFindings files exist
}

// HasIssues reports whether any unformatted files were found.
func (r *Results) HasIssues() bool {
	return len(r.Files) > 0
}

// Run executes `gofmt -l .` in dir and returns parsed results.
// gofmt always exits 0; unformatted files appear one per line on stdout.
func Run(dir string) (*Results, error) {
	cmd := exec.Command("gofmt", "-l", ".")
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start gofmt: %w", err)
	}

	results, parseErr := Parse(stdout)
	cmd.Wait() // gofmt -l exits 0 even when files are unformatted

	// Count total .go files so the clean message can show "passed for N files".
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".go") {
			results.Total++
		}
		return nil
	})

	// Strip an explicit dir prefix so paths are relative to the project root.
	if dir != "" && dir != "." {
		prefix := strings.TrimSuffix(dir, "/") + "/"
		for i, f := range results.Files {
			results.Files[i] = strings.TrimPrefix(f, prefix)
		}
	}
	return results, parseErr
}

// Parse reads `gofmt -l` output from r and returns parsed results.
// Exported to allow testing without running a real process.
// Each line of output is the path of one unformatted file.
func Parse(r io.Reader) (*Results, error) {
	var files []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			files = append(files, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read gofmt output: %w", err)
	}

	truncated := false
	if len(files) > MaxFindings {
		files = files[:MaxFindings]
		truncated = true
	}
	return &Results{Files: files, Truncated: truncated}, nil
}
