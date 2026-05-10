package tui

import (
	"embed"
	"fmt"
	"path"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"
)

// messages embeds all YAML message catalog files in internal/tui/messages/.
//
// @MX:NOTE: [AUTO] embed.FS is used to avoid filesystem dependency at runtime;
// catalogs are compiled into the binary so no external files are required.
//
//go:embed messages/*.yaml
var messagesFS embed.FS

// catalog is a flat map of key → translated string.
type catalog map[string]string

// catalogCache caches parsed catalogs by language code.
// Access is guarded by catalogMu.
var (
	catalogCache = map[string]catalog{}
	catalogMu    sync.RWMutex
)

// loadCatalog reads and parses messages/<lang>.yaml from the embedded FS.
// Results are cached after the first load.
//
// @MX:NOTE: [AUTO] loadCatalog is the single load point; sync.RWMutex ensures
// concurrent access is safe without duplicating parse work.
func loadCatalog(lang string) (catalog, error) {
	// Check cache first (read lock).
	catalogMu.RLock()
	if c, ok := catalogCache[lang]; ok {
		catalogMu.RUnlock()
		return c, nil
	}
	catalogMu.RUnlock()

	// Parse from embedded FS (write lock).
	catalogMu.Lock()
	defer catalogMu.Unlock()

	// Double-check after acquiring write lock.
	if c, ok := catalogCache[lang]; ok {
		return c, nil
	}

	p := path.Join("messages", lang+".yaml")
	data, err := messagesFS.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("i18n: catalog %q not found: %w", lang, err)
	}

	var raw map[string]string
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("i18n: parse %q: %w", lang, err)
	}

	c := catalog(raw)
	if c == nil {
		c = catalog{}
	}
	catalogCache[lang] = c
	return c, nil
}

// Translate returns the message for key in the requested language.
// Fallback chain: lang → "en" → key itself.
// If lang or the key is missing from lang catalog, the English value is returned.
// If the key is missing from the English catalog too, the key itself is returned.
//
// Key naming convention: <surface>.<section>.<role>
// Example keys: "doctor.summary.pass", "loop.phase.plan", "error.generic"
//
// @MX:NOTE: [AUTO] Translate is the single i18n entry point for all TUI surfaces.
// Callers must not embed raw strings; use a catalog key instead.
func Translate(key, lang string) string {
	if c, err := loadCatalog(lang); err == nil {
		if v, ok := c[key]; ok && v != "" {
			return v
		}
	}
	// Fallback to English.
	if lang != "en" {
		if c, err := loadCatalog("en"); err == nil {
			if v, ok := c[key]; ok && v != "" {
				return v
			}
		}
	}
	// Last resort: return the key itself so callers always get a non-empty string.
	return key
}

// CatalogKeys returns the sorted list of keys defined in the given language catalog.
// Returns an error if the catalog file does not exist or cannot be parsed.
// Placeholder catalogs (no keys) return an empty slice.
func CatalogKeys(lang string) ([]string, error) {
	c, err := loadCatalog(lang)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// CatalogValues returns a copy of all key→value pairs in the given language catalog.
// Returns an error if the catalog file does not exist or cannot be parsed.
func CatalogValues(lang string) (map[string]string, error) {
	c, err := loadCatalog(lang)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(c))
	for k, v := range c {
		out[k] = v
	}
	return out, nil
}
