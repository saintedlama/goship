package tester

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	assert.Equal(t, 1, r.Passed())
	assert.Equal(t, 0, r.Failed())
	assert.False(t, r.HasFailures())
}

func TestParse_Failing(t *testing.T) {
	r, err := Parse(strings.NewReader(failingJSON))
	require.NoError(t, err)
	assert.Equal(t, 1, r.Failed())
	assert.True(t, r.HasFailures())
	assert.NotEmpty(t, r.Packages[0].Cases[0].Output)
}

func TestParse_Mixed(t *testing.T) {
	r, err := Parse(strings.NewReader(mixedJSON))
	require.NoError(t, err)
	assert.Equal(t, 1, r.Passed())
	assert.Equal(t, 1, r.Failed())
	assert.Len(t, r.Packages, 1)
}

func TestParse_PackageAction(t *testing.T) {
	r, err := Parse(strings.NewReader(passingJSON))
	require.NoError(t, err)
	assert.Equal(t, "pass", r.Packages[0].Action)
}
