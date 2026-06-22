package i18n

import (
	"embed"
	"testing"
)

//go:embed catalogs/*.json
var testCatalogFS embed.FS

func TestCatalogLoader_LoadEN(t *testing.T) {
	loader := NewCatalogLoader(testCatalogFS)
	cache := NewPrinterCache(loader)
	printer, err := cache.GetPrinter("en")
	if err != nil {
		t.Fatalf("GetPrinter(en) failed: %v", err)
	}

	msg := printer.Sprintf("search.placeholder")
	if msg == "search.placeholder" {
		t.Error("translation key returned as-is; catalog may be empty or missing key")
	}
	if msg == "" {
		t.Error("empty translation for search.placeholder")
	}
}

func TestCatalogLoader_FallbackToEN(t *testing.T) {
	loader := NewCatalogLoader(testCatalogFS)
	cache := NewPrinterCache(loader)
	printer, err := cache.GetPrinter("xx-YY")
	if err != nil {
		t.Fatalf("GetPrinter(xx-YY) should fallback to en, got error: %v", err)
	}

	msg := printer.Sprintf("search.placeholder")
	if msg == "search.placeholder" {
		t.Error("fallback translation key returned as-is")
	}
	if msg == "" {
		t.Error("empty fallback translation for search.placeholder")
	}
}

func TestCatalogLoader_SupportedLocales(t *testing.T) {
	loader := NewCatalogLoader(testCatalogFS)
	_, err := loader.Load("en")
	if err != nil {
		t.Fatalf("failed to load en: %v", err)
	}
	locales := loader.SupportedLocales()
	hasEN := false
	for _, l := range locales {
		if l == "en" {
			hasEN = true
			break
		}
	}
	if !hasEN {
		t.Error("SupportedLocales should include en")
	}
}

func TestCatalogLoader_LoadZHCN(t *testing.T) {
	loader := NewCatalogLoader(testCatalogFS)
	cache := NewPrinterCache(loader)
	printer, err := cache.GetPrinter("zh-CN")
	if err != nil {
		t.Fatalf("GetPrinter(zh-CN) failed: %v", err)
	}

	msg := printer.Sprintf("search.placeholder")
	if msg == "search.placeholder" {
		t.Error("translation key returned as-is for zh-CN")
	}
	if msg == "" {
		t.Error("empty translation for search.placeholder in zh-CN")
	}
	if msg != "搜索网络..." {
		t.Errorf("unexpected translation for zh-CN: got %q, want %q", msg, "搜索网络...")
	}
}
