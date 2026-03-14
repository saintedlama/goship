package tester

import (
	"strings"
	"testing"
)

const passingJSON = `
{"Action":"run","Package":"example.com/foo","Test":"TestAdd"}
{"Action":"output","Package":"example.com/foo","Test":"TestAdd","Output":"--- PASS: TestAdd (0.00s)\n"}
{"Action":"pass","Package":"example.com/foo","Test":"TestAdd","Elapsed":0.001}
{"Action":"pass","Package":"example.com/foo","Elapsed":0.001}
`

const failingJSON = `
{"Action":"run","Package":"example.com/foo","Test":"TestBad"}
{"Action":"output","Package":"example.com/foo","Test":"TestBad","Output":"    foo_test.go:10: expected 1 got 2\n"}
{"Action":"output","Package":"example.com/foo","Test":"TestBad","Output":"--- FAIL: TestBad (0.00s)\n"}
{"Action":"fail","Package":"example.com/foo","Test":"TestBad","Elapsed":0.002}
{"Action":"fail","Package":"example.com/foo","Elapsed":0.002}
`

const mixedJSON = `
{"Action":"run","Package":"example.com/foo","Test":"TestOK"}
{"Action":"pass","Package":"example.com/foo","Test":"TestOK","Elapsed":0.001}
{"Action":"run","Package":"example.com/foo","Test":"TestBad"}
{"Action":"output","Package":"example.com/foo","Test":"TestBad","Output":"--- FAIL\n"}
{"Action":"fail","Package":"example.com/foo","Test":"TestBad","Elapsed":0.002}
{"Action":"fail","Package":"example.com/foo","Elapsed":0.003}
`

func TestParse_Passing(t *testing.T) {
	r, err := Parse(strings.NewReader(passingJSON))
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Passed(); got != 1 {
		t.Errorf("Passed() = %d, want 1", got)
	}
	if got := r.Failed(); got != 0 {
		t.Errorf("Failed() = %d, want 0", got)
	}
	if r.HasFailures() {
		t.Error("HasFailures() = true, want false")
	}
}

func TestParse_Failing(t *testing.T) {
	r, err := Parse(strings.NewReader(failingJSON))
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Failed(); got != 1 {
		t.Errorf("Failed() = %d, want 1", got)
	}
	if !r.HasFailures() {
		t.Error("HasFailures() = false, want true")
	}
	if len(r.Packages[0].Cases[0].Output) == 0 {
		t.Error("expected output lines on failed test case")
	}
}

func TestParse_Mixed(t *testing.T) {
	r, err := Parse(strings.NewReader(mixedJSON))
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Passed(); got != 1 {
		t.Errorf("Passed() = %d, want 1", got)
	}
	if got := r.Failed(); got != 1 {
		t.Errorf("Failed() = %d, want 1", got)
	}
	if len(r.Packages) != 1 {
		t.Errorf("len(Packages) = %d, want 1", len(r.Packages))
	}
}

func TestParse_PackageAction(t *testing.T) {
	r, err := Parse(strings.NewReader(passingJSON))
	if err != nil {
		t.Fatal(err)
	}
	pkg := r.Packages[0]
	if pkg.Action != "pass" {
		t.Errorf("Package.Action = %q, want \"pass\"", pkg.Action)
	}
}
