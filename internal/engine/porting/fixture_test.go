package porting

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixture_ParseYAML(t *testing.T) {
	yamlData := `
engine: test_engine
request:
  query: "hello world"
  category: general
  language: en
  page: 1
mock_response:
  status: 200
  headers:
    Content-Type: text/html
  body: "<html><body>OK</body></html>"
expected_results:
  - title: "OK"
    url: "https://example.com"
    content: ""
`
	f, err := ParseFixture([]byte(yamlData))
	require.NoError(t, err)
	assert.Equal(t, "test_engine", f.Engine)
	assert.Equal(t, "hello world", f.Request.Query)
	assert.Equal(t, 200, f.MockResponse.Status)
	assert.Len(t, f.ExpectedResults, 1)
	assert.Equal(t, "OK", f.ExpectedResults[0].Title)
}

func TestFixture_Validation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{"empty engine", "engine: ''\nmock_response: {}\nexpected_results: []", true},
		{"missing mock_response", "engine: x\nmock_response: null\nexpected_results: []", false},
		{"valid", "engine: x\nrequest: {query: q}\nmock_response: {status: 200}\nexpected_results: []", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseFixture([]byte(tc.yaml))
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFixture_Run(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "test_runner.yaml")
	err := os.WriteFile(fixturePath, []byte(`
engine: test_runner
request:
  query: test
  category: general
mock_response:
  status: 200
  headers:
    Content-Type: text/html
  body: "<html><body>ok</body></html>"
expected_results: []
`), 0644)
	require.NoError(t, err)

	data, err := os.ReadFile(fixturePath)
	require.NoError(t, err)

	f, err := ParseFixture(data)
	require.NoError(t, err)
	assert.Equal(t, "test_runner", f.Engine)
	assert.Equal(t, "test", f.Request.Query)
}

func TestRunFixtures_Directory(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "fixture1.yaml"), []byte(`
engine: engine1
request: {query: q}
mock_response: {status: 200, body: "ok"}
expected_results: []
`), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, "fixture2.yaml"), []byte(`
engine: engine2
request: {query: q}
mock_response: {status: 200}
expected_results: []
`), 0644)
	require.NoError(t, err)

	results := RunFixtures(dir)
	assert.Len(t, results, 2)
}

func TestFixture_LoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "simple_fixture.yaml")
	err := os.WriteFile(path, []byte(`
engine: simple
request:
  query: hello
  category: general
mock_response:
  status: 200
  body: "<html/>"
expected_results: []
`), 0644)
	require.NoError(t, err)

	f, err := LoadFixture(path)
	require.NoError(t, err)
	assert.Equal(t, "simple", f.Engine)
	assert.Equal(t, "hello", f.Request.Query)
}
