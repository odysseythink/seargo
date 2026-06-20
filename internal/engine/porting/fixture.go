package porting

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FixtureRequest represents the search request parameters for a fixture.
type FixtureRequest struct {
	Query    string `yaml:"query"`
	Category string `yaml:"category"`
	Language string `yaml:"language,omitempty"`
	Page     int    `yaml:"page,omitempty"`
}

// FixtureMockResponse represents the mock HTTP response for a fixture.
type FixtureMockResponse struct {
	Status  int               `yaml:"status"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body"`
}

// FixtureExpectedResult represents an expected result in a fixture.
type FixtureExpectedResult struct {
	Title   string `yaml:"title"`
	URL     string `yaml:"url"`
	Content string `yaml:"content,omitempty"`
}

// Fixture represents a complete golden test fixture.
type Fixture struct {
	Engine          string                  `yaml:"engine"`
	Request         FixtureRequest          `yaml:"request"`
	MockResponse    FixtureMockResponse     `yaml:"mock_response"`
	ExpectedResults []FixtureExpectedResult `yaml:"expected_results"`
}

// ParseFixture parses a YAML fixture from raw bytes.
func ParseFixture(data []byte) (*Fixture, error) {
	var f Fixture
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse fixture: %w", err)
	}

	if err := f.Validate(); err != nil {
		return nil, err
	}

	return &f, nil
}

// Validate checks that the fixture has the required fields.
func (f *Fixture) Validate() error {
	if f.Engine == "" {
		return fmt.Errorf("fixture engine name is empty")
	}
	return nil
}

// LoadFixture loads a fixture from a YAML file.
func LoadFixture(path string) (*Fixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read fixture file: %w", err)
	}
	return ParseFixture(data)
}

// FixtureResult holds the outcome of running a single fixture.
type FixtureResult struct {
	Path   string
	Passed bool
	Error  string
}

// RunFixtures loads and validates all YAML files in a directory.
// Returns a slice of FixtureResult for each file found.
func RunFixtures(dir string) []FixtureResult {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []FixtureResult{{Path: dir, Passed: false, Error: fmt.Sprintf("read dir: %v", err)}}
	}

	var results []FixtureResult
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		f, err := LoadFixture(path)
		if err != nil {
			results = append(results, FixtureResult{
				Path:   path,
				Passed: false,
				Error:  err.Error(),
			})
			continue
		}

		if err := f.Validate(); err != nil {
			results = append(results, FixtureResult{
				Path:   path,
				Passed: false,
				Error:  err.Error(),
			})
			continue
		}

		results = append(results, FixtureResult{
			Path:   path,
			Passed: true,
		})
		_ = f
	}

	return results
}
