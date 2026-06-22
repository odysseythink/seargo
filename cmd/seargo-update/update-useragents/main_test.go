package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const mockReleasesHTML = `<!DOCTYPE html>
<html>
<body>
<a href="/pub/firefox/releases/152.0/">152.0/</a>
<a href="/pub/firefox/releases/151.1/">151.1/</a>
<a href="/pub/firefox/releases/151.0/">151.0/</a>
<a href="/pub/firefox/releases/150.0/">150.0/</a>
<a href="/pub/firefox/releases/149.0/">149.0/</a>
<a href="/pub/firefox/releases/150.0b1/">150.0b1/</a>
<a href="/pub/firefox/releases/">../</a>
</body>
</html>`

func TestRun_MockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, mockReleasesHTML)
	}))
	defer srv.Close()

	dir := t.TempDir()
	out := filepath.Join(dir, "useragents.json")

	if err := Run(out, srv.Client(), srv.URL+"/"); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var ua userAgentData
	if err := json.Unmarshal(data, &ua); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	if len(ua.Versions) == 0 {
		t.Fatal("expected versions")
	}
	// Latest major is 152, keep 152 and 151.
	found := map[string]bool{}
	for _, v := range ua.Versions {
		found[v] = true
	}
	for _, want := range []string{"152.0", "151.1", "151.0"} {
		if !found[want] {
			t.Errorf("expected version %s in %v", want, ua.Versions)
		}
	}
	for _, notWant := range []string{"150.0", "149.0"} {
		if found[notWant] {
			t.Errorf("did not expect version %s in %v", notWant, ua.Versions)
		}
	}
	if ua.UA == "" {
		t.Error("expected UA template")
	}
}
