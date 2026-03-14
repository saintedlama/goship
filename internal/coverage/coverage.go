package coverage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

// PackageCoverage holds statement-weighted coverage for one package.
type PackageCoverage struct {
	Name    string
	Covered int // number of covered statements
	Total   int // total number of statements
	Percent float64
}

// Results holds parsed coverage for all packages and the overall total.
type Results struct {
	Packages []*PackageCoverage
	Total    float64 // statement-weighted percentage across all packages
}

// Run parses the coverage profile at profilePath and returns structured results.
// It reads the raw profile directly, so no extra subprocess is needed.
func Run(_, profilePath string) (*Results, error) {
	f, err := os.Open(profilePath)
	if err != nil {
		return nil, fmt.Errorf("open coverage profile: %w", err)
	}
	defer f.Close()
	return ParseProfile(f)
}

// ParseProfile reads a `go test -coverprofile` file and returns structured results.
// Exported so it can be tested without running a real process.
//
// Profile format:
//
//	mode: set
//	github.com/foo/bar/file.go:10.20,15.5 3 1
//
// Each block line: file:startLine.startCol,endLine.endCol numStatements count
// count > 0 means the block was executed at least once.
func ParseProfile(r io.Reader) (*Results, error) {
	pkgMap := map[string]*PackageCoverage{}
	pkgOrder := []string{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}

		// "github.com/foo/bar/file.go:10.2,12.14 3 1"
		//  ─────────────── file ──────────────────  ^ count
		//                                        ^ numStmt
		colonIdx := strings.LastIndex(line, ":")
		if colonIdx < 0 {
			continue
		}
		filePath := line[:colonIdx]
		rest := line[colonIdx+1:]

		// rest is "startLine.startCol,endLine.endCol numStmt count"
		spaceIdx := strings.Index(rest, " ")
		if spaceIdx < 0 {
			continue
		}
		fields := strings.Fields(rest[spaceIdx+1:])
		if len(fields) < 2 {
			continue
		}
		numStmt, err := strconv.Atoi(fields[0])
		if err != nil || numStmt <= 0 {
			continue
		}
		count, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		pkgName := path.Dir(filePath)
		if _, ok := pkgMap[pkgName]; !ok {
			pkgMap[pkgName] = &PackageCoverage{Name: pkgName}
			pkgOrder = append(pkgOrder, pkgName)
		}
		pkg := pkgMap[pkgName]
		pkg.Total += numStmt
		if count > 0 {
			pkg.Covered += numStmt
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	results := &Results{}
	allTotal, allCovered := 0, 0
	for _, name := range pkgOrder {
		pkg := pkgMap[name]
		if pkg.Total > 0 {
			pkg.Percent = float64(pkg.Covered) / float64(pkg.Total) * 100
		}
		results.Packages = append(results.Packages, pkg)
		allTotal += pkg.Total
		allCovered += pkg.Covered
	}
	if allTotal > 0 {
		results.Total = float64(allCovered) / float64(allTotal) * 100
	}
	return results, nil
}
