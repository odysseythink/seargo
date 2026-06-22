package i18n

import (
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

// PrinterCache holds cached message.Printer instances per locale.
type PrinterCache struct {
	loader   *CatalogLoader
	mu       sync.RWMutex
	printers map[string]*message.Printer
}

// NewPrinterCache creates a PrinterCache backed by the given CatalogLoader.
func NewPrinterCache(loader *CatalogLoader) *PrinterCache {
	return &PrinterCache{
		loader:   loader,
		printers: make(map[string]*message.Printer),
	}
}

// GetPrinter returns a message.Printer for the given locale.
// Falls back to en if the locale catalog is not available.
func (p *PrinterCache) GetPrinter(locale string) (*message.Printer, error) {
	p.mu.RLock()
	if pr, ok := p.printers[locale]; ok {
		p.mu.RUnlock()
		return pr, nil
	}
	p.mu.RUnlock()

	cat, err := p.loader.Load(locale)
	if err != nil {
		// Fallback to en
		cat, err = p.loader.Load("en")
		if err != nil {
			// Last resort: empty builder (satisfies catalog.Catalog)
			cat = catalog.NewBuilder()
		}
		locale = "en"
	}

	langTag, _ := language.Parse(locale)
	printer := message.NewPrinter(langTag, message.Catalog(cat))

	p.mu.Lock()
	p.printers[locale] = printer
	p.mu.Unlock()

	return printer, nil
}
