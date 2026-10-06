package hook

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The bind cache records, per session, the binding the last successful
// prompt-submit bind established. A later prompt whose session, run, owner
// PID, process start, role, and slot all match — and whose run-state probe
// reports the same live run — skips the broker open and the peer query. The
// probe itself always runs, so a retired or changed run is never served from
// the cache. The hook is the cache's only writer.

type factoryBindCacheEntry struct {
	Session string `json:"session"`
	Run     string `json:"run"`
	PID     int    `json:"pid"`
	Start   string `json:"process_start"`
	Role    string `json:"role"`
	Slot    string `json:"slot"`
	BoundAt string `json:"bound_at"`
}

var factoryCacheSessionPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

func factoryBindCacheDir(root string) (string, error) {
	dir, err := homestate.FactoryDir(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "bind-cache"), nil
}

func factoryBindCachePath(root, session, suffix string) (string, bool) {
	if !factoryCacheSessionPattern.MatchString(session) || strings.Contains(session, "..") {
		return "", false
	}
	dir, err := factoryBindCacheDir(root)
	if err != nil {
		return "", false
	}
	return filepath.Join(dir, session+suffix), true
}

// factoryBindCacheHit reports whether the cached binding matches want exactly.
func factoryBindCacheHit(root string, want factoryBindCacheEntry) bool {
	path, ok := factoryBindCachePath(root, want.Session, ".json")
	if !ok {
		return false
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var got factoryBindCacheEntry
	if json.Unmarshal(raw, &got) != nil {
		return false
	}
	return got.Session == want.Session && got.Run == want.Run && got.PID == want.PID &&
		got.Start == want.Start && got.Role == want.Role && got.Slot == want.Slot
}

// writeFactoryBindCache records a successful binding. A write failure only
// costs the next prompt a full bind, so it is ignored.
func writeFactoryBindCache(root string, e factoryBindCacheEntry) {
	path, ok := factoryBindCachePath(root, e.Session, ".json")
	if !ok {
		return
	}
	e.BoundAt = time.Now().UTC().Format(time.RFC3339)
	raw, err := json.Marshal(e)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

// dropFactoryBindCache invalidates a session's cached binding.
func dropFactoryBindCache(root, session string) {
	if path, ok := factoryBindCachePath(root, session, ".json"); ok {
		_ = os.Remove(path)
	}
}

// dropFactoryBindCacheForSlot invalidates every other session's cached
// binding to run/slot. A bind that takes a slot over (a handoff, a launch
// binding, a registration) replaces whoever held it; their cache must not
// keep answering for an endpoint the broker has since tombstoned.
func dropFactoryBindCacheForSlot(root, run, slot, keepSession string) {
	dir, err := factoryBindCacheDir(root)
	if err != nil {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, de := range entries {
		name := de.Name()
		if !strings.HasSuffix(name, ".json") || strings.TrimSuffix(name, ".json") == keepSession {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var e factoryBindCacheEntry
		if json.Unmarshal(raw, &e) != nil || (e.Run == run && e.Slot == slot) {
			_ = os.Remove(filepath.Join(dir, name))
		}
	}
}

// surfaceFactoryInboxState logs a degraded inbox claim at warn on every
// occurrence and returns a notice for the session at most once per
// factoryDegradedNoticeInterval. Any other state returns "".
func surfaceFactoryInboxState(root, session, state string) string {
	detail, degraded := strings.CutPrefix(state, "degraded:")
	if !degraded {
		return ""
	}
	detail = strings.TrimSpace(detail)
	if strings.HasPrefix(detail, "no-broker:") {
		return "" // nothing was ever sent to this run; the inbox is empty
	}
	slog.Warn("factory inbox degraded", "session", session, "state", detail)
	marker, ok := factoryBindCachePath(root, session, ".degraded")
	if !ok {
		return ""
	}
	if info, err := os.Stat(marker); err == nil && time.Since(info.ModTime()) < factoryDegradedNoticeInterval {
		return ""
	}
	if err := os.MkdirAll(filepath.Dir(marker), 0o700); err == nil {
		_ = os.WriteFile(marker, []byte(time.Now().UTC().Format(time.RFC3339)), 0o600)
	}
	return "factory messaging degraded: inbox " + detail
}
