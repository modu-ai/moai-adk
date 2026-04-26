package core

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// LSP document synchronization method constants (prevent hard-coding).
const (
	methodDidOpen   = "textDocument/didOpen"
	methodDidChange = "textDocument/didChange"
	methodDidClose  = "textDocument/didClose"
	methodDidSave   = "textDocument/didSave"
)

// docEntry represents the state of an individual file in the documentCache.
type docEntry struct {
	languageID   string
	version      int32
	content      string
	lastActivity time.Time
}

// documentCache tracks the state of files open in the language server to prevent
// duplicate didOpen sends (REQ-LC-006).
//
// @MX:ANCHOR: [AUTO] documentCache — central state store for all document synchronization operations
// @MX:REASON: fan_in >= 3 — openOrChange, reapIdle, didSave, OpenFile, DidSave, and Manager all manage document state through this struct
type documentCache struct {
	mu      sync.RWMutex
	entries map[string]docEntry
}

// newDocumentCache creates an empty documentCache.
func newDocumentCache() *documentCache {
	return &documentCache{
		entries: make(map[string]docEntry),
	}
}

// openOrChange sends textDocument/didOpen or textDocument/didChange depending on the URI's state.
//
// Rules:
//   - Unregistered URI: send didOpen (version=1) and add to cache.
//   - Registered URI + same content: update lastActivity only (no-op).
//   - Registered URI + changed content: send didChange (increment version, full document sync) and update cache.
func (c *documentCache) openOrChange(ctx context.Context, tr transport.Transport, uri, languageID, content string) error {
	c.mu.Lock()

	entry, exists := c.entries[uri]

	if !exists {
		// New file: send didOpen.
		newEntry := docEntry{
			languageID:   languageID,
			version:      1,
			content:      content,
			lastActivity: time.Now(),
		}
		c.entries[uri] = newEntry
		c.mu.Unlock()

		params := map[string]any{
			"textDocument": map[string]any{
				"uri":        uri,
				"languageId": languageID,
				"version":    int32(1),
				"text":       content,
			},
		}
		if err := tr.Notify(ctx, methodDidOpen, params); err != nil {
			return transport.WrapCallError(methodDidOpen, uri, languageID, err)
		}
		return nil
	}

	if entry.content == content {
		// Same content: update lastActivity only (no-op).
		entry.lastActivity = time.Now()
		c.entries[uri] = entry
		c.mu.Unlock()
		return nil
	}

	// Content changed: send didChange.
	newVersion := entry.version + 1
	entry.version = newVersion
	entry.content = content
	entry.lastActivity = time.Now()
	c.entries[uri] = entry
	c.mu.Unlock()

	params := map[string]any{
		"textDocument": map[string]any{
			"uri":     uri,
			"version": newVersion,
		},
		"contentChanges": []map[string]any{
			{"text": content},
		},
	}
	if err := tr.Notify(ctx, methodDidChange, params); err != nil {
		return transport.WrapCallError(methodDidChange, uri, languageID, err)
	}
	return nil
}

// touch updates the lastActivity timestamp for the URI to the current time.
// If the URI is not in the cache, this is a no-op.
func (c *documentCache) touch(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.entries[uri]; ok {
		entry.lastActivity = time.Now()
		c.entries[uri] = entry
	}
}

// snapshot returns a copy of the current cache. Used by the idle reaper and similar consumers.
func (c *documentCache) snapshot() map[string]docEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]docEntry, len(c.entries))
	for k, v := range c.entries {
		out[k] = v
	}
	return out
}

// remove deletes the URI from the cache.
func (c *documentCache) remove(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, uri)
}

// reapIdle sends textDocument/didClose for entries that have exceeded idleTimeout and removes them from the cache (REQ-LC-022).
// Errors are logged but do not stop reaping.
// Returns the number of entries removed.
func (c *documentCache) reapIdle(ctx context.Context, tr transport.Transport, idleTimeout time.Duration) int {
	now := time.Now()

	// Collect expired entries while holding the read lock (minimize lock duration).
	c.mu.RLock()
	var expired []struct {
		uri        string
		languageID string
	}
	for uri, entry := range c.entries {
		if now.Sub(entry.lastActivity) > idleTimeout {
			expired = append(expired, struct {
				uri        string
				languageID string
			}{uri: uri, languageID: entry.languageID})
		}
	}
	c.mu.RUnlock()

	if len(expired) == 0 {
		return 0
	}

	reaped := 0
	for _, e := range expired {
		params := map[string]any{
			"textDocument": map[string]any{
				"uri": e.uri,
			},
		}
		// Ignore errors and continue (REQ-LC-022 requirement).
		_ = tr.Notify(ctx, methodDidClose, params)
		c.remove(e.uri)
		reaped++
	}
	return reaped
}

// didSave sends textDocument/didSave for a tracked URI (REQ-LC-023).
// Returns an error if the URI is not in the cache.
func (c *documentCache) didSave(ctx context.Context, tr transport.Transport, uri string) error {
	c.mu.RLock()
	entry, ok := c.entries[uri]
	c.mu.RUnlock()

	if !ok {
		return transport.WrapCallError(methodDidSave, uri, "", fmt.Errorf("file not tracked: %q", uri))
	}

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	}
	if err := tr.Notify(ctx, methodDidSave, params); err != nil {
		return transport.WrapCallError(methodDidSave, uri, entry.languageID, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// pathToURI — convert a file path to an LSP URI
// ---------------------------------------------------------------------------

// pathToURI converts a file path to an LSP file:// URI.
//
// Conversion rules:
//   - If the path already has a "file://" prefix, return it as-is.
//   - Absolute path: resolve symlinks, then return "file://" + slash-normalized path (macOS /var → /private/var).
//   - Relative path: return "file://" + path (as-is).
func pathToURI(path string) string {
	if strings.HasPrefix(path, "file://") {
		return path
	}
	if filepath.IsAbs(path) {
		// Resolve symlinks: on macOS /var/folders → /private/var/folders.
		// Prevents URI mismatch when the language server returns the real path.
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		}
		// Unix: /foo/bar.go → file:///foo/bar.go
		// Windows: C:\foo\bar.go → file:///C:/foo/bar.go
		return "file://" + filepath.ToSlash(path)
	}
	// Relative path: return as-is (test compatibility).
	return "file://" + path
}

// resolveLanguageID determines the LSP languageId from ServerConfig.Language.
// Unknown languages are returned unchanged from cfg.Language.
func resolveLanguageID(language string) string {
	switch language {
	case "go":
		return "go"
	case "python":
		return "python"
	case "typescript":
		return "typescript"
	case "javascript":
		return "javascript"
	case "rust":
		return "rust"
	case "java":
		return "java"
	default:
		return language
	}
}
