package httpx

import (
	"encoding/json"
	"math/rand"
	"os"
	"strings"
)

// useragentData is the JSON file format.
type useragentData struct {
	OS       []string `json:"os"`
	UA       string   `json:"ua"`
	Versions []string `json:"versions"`
}

// NewUserAgentPool loads UA data from a JSON file. If the file is missing
// or unreadable, it returns a built-in fallback pool.
func NewUserAgentPool(path string) (*UserAgentPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallbackPool(), nil
	}

	var raw useragentData
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	if len(raw.OS) == 0 || len(raw.Versions) == 0 || raw.UA == "" {
		return fallbackPool(), nil
	}

	return &UserAgentPool{
		OSes:     raw.OS,
		Template: raw.UA,
		Versions: raw.Versions,
	}, nil
}

// Random generates a random User-Agent string.
func (p *UserAgentPool) Random() string {
	p.mu.RLock()
	oses := p.OSes
	versions := p.Versions
	tmpl := p.Template
	p.mu.RUnlock()

	if len(oses) == 0 || len(versions) == 0 || tmpl == "" {
		return "SearGo/1.0"
	}

	os := oses[rand.Intn(len(oses))]
	version := versions[rand.Intn(len(versions))]

	ua := strings.ReplaceAll(tmpl, "{os}", os)
	ua = strings.ReplaceAll(ua, "{version}", version)
	return ua
}

// Reload reloads the pool from a new file path.
func (p *UserAgentPool) Reload(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var raw useragentData
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.OSes = raw.OS
	p.Template = raw.UA
	p.Versions = raw.Versions
	return nil
}

// fallbackPool returns a minimal built-in UA pool.
func fallbackPool() *UserAgentPool {
	return &UserAgentPool{
		OSes: []string{
			"Windows NT 10.0; Win64; x64",
			"X11; Linux x86_64",
		},
		Template: "Mozilla/5.0 ({os}; rv:{version}) Gecko/20100101 Firefox/{version}",
		Versions: []string{
			"151.0",
			"150.0",
		},
	}
}
