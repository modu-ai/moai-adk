// Package cache provides a TTL-based, version-aware diagnostic cache for the
// MoAI LSP aggregator layer. Diagnostics are stored per file URI, keyed by
// (uri, version) tuple, and expire automatically after a configurable duration.
//
// Usage overview:
//
//	cache := cache.NewDiagnosticCache(5 * time.Second)
//	cache.Set("file:///main.go", 3, diagnostics)
//	diags, ok := cache.Get("file:///main.go", 3)
package cache
