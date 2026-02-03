package statusline

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// LatestVersionFunc is a function that fetches the latest available version.
// It may perform network I/O and should respect context cancellation.
type LatestVersionFunc func(ctx context.Context) (string, error)

// UpdateChecker checks for new ADK versions with caching support.
// It implements the UpdateProvider interface.
type UpdateChecker struct {
	currentVersion string
	cacheTTL       time.Duration
	fetchLatest    LatestVersionFunc

	mu           sync.RWMutex
	cachedResult *VersionData
	cachedAt     time.Time
}

// NewUpdateChecker creates an UpdateChecker with the given current version,
// cache TTL, and a function to fetch the latest version.
// If fetchFn is nil, CheckUpdate returns current version without update info.
func NewUpdateChecker(currentVersion string, cacheTTL time.Duration, fetchFn LatestVersionFunc) *UpdateChecker {
	return &UpdateChecker{
		currentVersion: currentVersion,
		cacheTTL:       cacheTTL,
		fetchLatest:    fetchFn,
	}
}

// CheckUpdate returns version data with update availability.
// Results are cached for cacheTTL duration to avoid repeated network calls.
// On fetch failure, returns cached data if available, or current version only.
func (u *UpdateChecker) CheckUpdate(ctx context.Context) (*VersionData, error) {
	// Check cache first
	if cached, ok := u.getCachedIfFresh(); ok {
		return cached, nil
	}

	// No fetch function: return current version only
	if u.fetchLatest == nil {
		return &VersionData{
			Current:   u.currentVersion,
			Available: true,
		}, nil
	}

	// Fetch latest version
	latest, err := u.fetchLatest(ctx)
	if err != nil {
		slog.Debug("update check failed", "error", err)
		// Return stale cached data if available
		if stale := u.getCachedStale(); stale != nil {
			return stale, nil
		}
		// No cache: return current version without update info
		return &VersionData{
			Current:   u.currentVersion,
			Available: true,
		}, nil
	}

	data := &VersionData{
		Current:         u.currentVersion,
		Latest:          latest,
		UpdateAvailable: latest != "" && latest != u.currentVersion,
		Available:       true,
	}

	u.updateCache(data)

	return data, nil
}

// getCachedIfFresh returns the cached result if it exists and is within TTL.
func (u *UpdateChecker) getCachedIfFresh() (*VersionData, bool) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if u.cachedResult != nil && time.Since(u.cachedAt) < u.cacheTTL {
		return u.cachedResult, true
	}
	return nil, false
}

// getCachedStale returns any cached result regardless of TTL, or nil.
func (u *UpdateChecker) getCachedStale() *VersionData {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.cachedResult
}

// updateCache stores the result in the cache. Thread-safe.
func (u *UpdateChecker) updateCache(data *VersionData) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.cachedResult = data
	u.cachedAt = time.Now()
}
