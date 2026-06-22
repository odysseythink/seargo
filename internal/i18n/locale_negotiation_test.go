package i18n

import (
	"testing"
)

func TestNormalizeLocaleTag(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"zh_Hans_CN", "zh-Hans-CN"},
		{"en_us", "en-US"},
		{"fr", "fr"},
		{"ZH-CN", "zh-CN"},
		{"AR-EG", "ar-EG"},
	}
	for _, tt := range tests {
		got := NormalizeLocaleTag(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeLocaleTag(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNegotiateLocale_CookieWins(t *testing.T) {
	reg := NewLocaleRegistry()
	neg := NewNegotiator(reg)

	// Cookie with supported zh-CN, Accept-Language says en → cookie wins
	got := neg.Negotiate("en", "zh-CN", "en")
	if got != "zh-CN" {
		t.Errorf("cookie should win: got %q, want zh-CN", got)
	}
}

func TestNegotiateLocale_AcceptLanguage(t *testing.T) {
	reg := NewLocaleRegistry()
	neg := NewNegotiator(reg)

	// No cookie, Accept-Language zh-CN → zh-CN
	got := neg.Negotiate("zh-CN,en;q=0.8", "", "en")
	if got != "zh-CN" {
		t.Errorf("got %q, want zh-CN", got)
	}
}

func TestNegotiateLocale_UnsupportedCookieFallsBack(t *testing.T) {
	reg := NewLocaleRegistry()
	neg := NewNegotiator(reg)

	// Cookie has unsupported de-AT, Accept-Language supports en → en
	got := neg.Negotiate("en", "de-AT", "en")
	if got != "en" {
		t.Errorf("unsupported cookie should fall back: got %q, want en", got)
	}
}

func TestNegotiateLocale_DefaultFallback(t *testing.T) {
	reg := NewLocaleRegistry()
	neg := NewNegotiator(reg)

	// No cookie, no Accept-Language, default is xx-YY (unsupported) → en
	got := neg.Negotiate("", "", "xx-YY")
	if got != "en" {
		t.Errorf("should fall back to en: got %q", got)
	}

	// Default zh-CN is supported
	got = neg.Negotiate("", "", "zh-CN")
	if got != "zh-CN" {
		t.Errorf("default should be used: got %q", got)
	}
}

func TestNegotiateLocale_LanguageOnlyFallback(t *testing.T) {
	reg := NewLocaleRegistry()
	neg := NewNegotiator(reg)

	// Accept-Language ar-SA → ar is supported (language-only fallback)
	got := neg.Negotiate("ar-SA", "", "en")
	if got != "ar" {
		t.Errorf("language-only fallback for ar-SA: got %q, want ar", got)
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	tags := ParseAcceptLanguage("zh-CN,zh;q=0.9,en;q=0.8")
	if len(tags) == 0 {
		t.Fatal("expected non-empty tags")
	}
	if tags[0] != "zh-CN" {
		t.Errorf("first tag = %q, want zh-CN", tags[0])
	}
}
