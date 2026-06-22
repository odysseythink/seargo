package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
)

// DefaultCatalogFS is the embedded filesystem containing catalog JSON files.
//
//go:embed catalogs/*.json
var DefaultCatalogFS embed.FS

// CatalogLoader loads translation catalogs from an embed.FS.
type CatalogLoader struct {
	fs       embed.FS
	mu       sync.RWMutex
	catalogs map[string]catalog.Catalog
	locales  []string
}

// NewCatalogLoader creates a CatalogLoader from the given embed.FS.
func NewCatalogLoader(fs embed.FS) *CatalogLoader {
	return &CatalogLoader{
		fs:       fs,
		catalogs: make(map[string]catalog.Catalog),
	}
}

// Load loads a catalog JSON file for the given locale.
func (l *CatalogLoader) Load(locale string) (catalog.Catalog, error) {
	l.mu.RLock()
	if cat, ok := l.catalogs[locale]; ok {
		l.mu.RUnlock()
		return cat, nil
	}
	l.mu.RUnlock()

	data, err := l.fs.ReadFile(fmt.Sprintf("catalogs/%s.json", locale))
	if err != nil {
		return nil, err
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("catalog %s: %w", locale, err)
	}

	langTag, err := language.Parse(locale)
	if err != nil {
		langTag = language.English
	}

	builder := catalog.NewBuilder()
	for msgid, msgstr := range messages {
		builder.SetString(langTag, msgid, msgstr)
	}

	l.mu.Lock()
	l.catalogs[locale] = builder
	l.locales = append(l.locales, locale)
	l.mu.Unlock()

	return builder, nil
}

// SupportedLocales returns the list of locales loaded so far.
func (l *CatalogLoader) SupportedLocales() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]string, len(l.locales))
	copy(result, l.locales)
	return result
}
