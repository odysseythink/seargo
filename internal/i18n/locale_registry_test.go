package i18n

import (
	"testing"
)

func TestLocaleRegistry_SupportedLocales(t *testing.T) {
	r := NewLocaleRegistry()

	if !r.IsSupported("en") {
		t.Error("en should be supported")
	}
	if !r.IsSupported("zh-CN") {
		t.Error("zh-CN should be supported")
	}
	if r.IsSupported("xx-YY") {
		t.Error("xx-YY should not be supported")
	}
}

func TestLocaleRegistry_RTL(t *testing.T) {
	r := NewLocaleRegistry()

	if r.IsRTL("ar") != true {
		t.Error("ar should be RTL")
	}
	if r.IsRTL("en") != false {
		t.Error("en should not be RTL")
	}
}

func TestLocaleRegistry_BestMatch(t *testing.T) {
	r := NewLocaleRegistry()

	tests := []struct {
		acceptTags []string
		want       string
	}{
		{[]string{"zh-CN", "en"}, "zh-CN"},
		{[]string{"fr", "de"}, "en"},        // unsupported → en fallback
		{[]string{"zh-Hans-CN"}, "en"},       // no fuzzy match → en fallback
		{[]string{}, "en"},
	}
	for _, tt := range tests {
		got := r.BestMatch(tt.acceptTags)
		if got != tt.want {
			t.Errorf("BestMatch(%v) = %q, want %q", tt.acceptTags, got, tt.want)
		}
	}
}

func TestLocaleRegistry_LocaleInfo(t *testing.T) {
	r := NewLocaleRegistry()
	infos := r.Supported
	if len(infos) < 2 {
		t.Fatalf("expected at least 2 supported locales, got %d", len(infos))
	}
	if infos[0].Tag != "en" {
		t.Errorf("first locale = %q, want en", infos[0].Tag)
	}
}
