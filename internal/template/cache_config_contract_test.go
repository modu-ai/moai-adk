package template

import (
	"io/fs"
	"strings"
	"testing"
)

// TestCacheYAMLFourKeyContract guards the cacheStrategy four-key contract:
// the shipped template cache.yaml must expose every key that
// internal/config DefaultCacheConfig() defines, so user projects see the
// full tunable surface rather than a partial one.
func TestCacheYAMLFourKeyContract(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".moai/config/sections/cache.yaml")
	if err != nil {
		t.Fatalf("ReadFile(cache.yaml) error: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "cacheStrategy:") {
		t.Fatalf("cache.yaml missing cacheStrategy block")
	}
	for _, key := range []string{"enabled:", "session_ttl:", "spec_ttl:", "min_cacheable_tokens:"} {
		if !strings.Contains(content, key) {
			t.Errorf("cache.yaml missing cacheStrategy key %q", key)
		}
	}
}
