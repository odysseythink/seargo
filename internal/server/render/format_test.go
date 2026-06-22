package render

import (
	"testing"
)

func TestResolveFormatQueryParam(t *testing.T) {
	f, err := ResolveFormat("csv", "*/*", []string{"json", "csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatCSV {
		t.Errorf("expected csv, got %s", f)
	}
	f, err = ResolveFormat("JSON", "*/*", []string{"json", "csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatJSON {
		t.Errorf("expected json, got %s", f)
	}
}

func TestResolveFormatQueryDisallowed(t *testing.T) {
	_, err := ResolveFormat("csv", "*/*", []string{"json"})
	if err == nil {
		t.Fatal("expected error for disallowed format")
	}
	_, err = ResolveFormat("xml", "*/*", []string{"json", "csv"})
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestResolveFormatAcceptHeader(t *testing.T) {
	f, err := ResolveFormat("", "application/json", []string{"json", "csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatJSON {
		t.Errorf("expected json, got %s", f)
	}
	f, err = ResolveFormat("", "text/csv", []string{"csv", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatCSV {
		t.Errorf("expected csv, got %s", f)
	}
	f, err = ResolveFormat("", "application/rss+xml", []string{"rss", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatRSS {
		t.Errorf("expected rss, got %s", f)
	}
	f, err = ResolveFormat("", "application/atom+xml", []string{"rss", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatRSS {
		t.Errorf("expected atom+xml → rss, got %s", f)
	}
}

func TestResolveFormatAcceptDisallowed(t *testing.T) {
	_, err := ResolveFormat("", "text/csv", []string{"json"})
	if err == nil {
		t.Fatal("expected error for csv not in allowed list")
	}
	_, err = ResolveFormat("", "application/rss+xml", []string{"json", "csv"})
	if err == nil {
		t.Fatal("expected error for rss not in allowed list")
	}
}

func TestResolveFormatWildcard(t *testing.T) {
	f, err := ResolveFormat("", "*/*", []string{"json", "csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatJSON {
		t.Errorf("expected json from */* fallback, got %s", f)
	}
	f, err = ResolveFormat("", "*/*", []string{"html", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatHTML {
		t.Errorf("expected html from */* with html in allowed, got %s", f)
	}
}

func TestResolveFormatNoAcceptNoQuery(t *testing.T) {
	f, err := ResolveFormat("", "", []string{"json", "csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatJSON {
		t.Errorf("expected json as default, got %s", f)
	}
	f, err = ResolveFormat("", "", []string{"html", "json", "csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatHTML {
		t.Errorf("expected html as default, got %s", f)
	}
}

func TestResolveFormatAcceptQualityValues(t *testing.T) {
	f, err := ResolveFormat("", "text/csv;q=0.9, application/json;q=0.8", []string{"json", "csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatCSV {
		t.Errorf("expected csv (higher q), got %s", f)
	}
	f, err = ResolveFormat("", "application/json;q=0.5, application/rss+xml;q=0.9", []string{"json", "rss"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != FormatRSS {
		t.Errorf("expected rss (higher q), got %s", f)
	}
}

func TestResolveFormatEmptyAllowed(t *testing.T) {
	_, err := ResolveFormat("json", "*/*", []string{})
	if err == nil {
		t.Fatal("expected error for empty allowed list")
	}
}
