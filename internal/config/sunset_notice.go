package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// sunsetNoticeMu guards access to sunsetNoticeOnce so that test-only reset
// (resetSunsetNoticeOnce) cannot race with concurrent emission via
// emitSunsetDormantNotice. SPEC-V3R6-CI-BASELINE-DRIFT-001 M3.
var sunsetNoticeMu sync.Mutex

// sunsetNoticeOnce guards the once-per-process DORMANT notice emission.
// REQ-MIG003-018: emitted at most once per process lifetime.
// Protected by sunsetNoticeMu for test-reset safety (SPEC-V3R6-CI-BASELINE-DRIFT-001).
var sunsetNoticeOnce sync.Once

// emitSunsetDormantNotice emits a one-shot session log per REQ-MIG003-018.
// Guarded by sync.Once so it fires at most once per process lifetime.
// The notice is advisory only — it reminds future maintainers that
// SunsetConfig is template-only and has no runtime hot path.
//
// @MX:NOTE: [AUTO] DORMANT notice helper — fires when sunset.yaml exists,
// reminds future maintainers that SunsetConfig is template-only.
// REQ-MIG003-018, AC-MIG003-15.
//
// Concurrency: sunsetNoticeMu is held for the entire Do() invocation so that
// the package-level sync.Once is never accessed concurrently with the
// resetSunsetNoticeOnce() test-only struct rewrite. The slog.Info side-effect
// is held under the mutex; this is acceptable because the emission path is
// once-per-process and the lock is uncontended after first invocation.
// SPEC-V3R6-CI-BASELINE-DRIFT-001 M3.
func emitSunsetDormantNotice(sectionsDir string) {
	sunsetNoticeMu.Lock()
	defer sunsetNoticeMu.Unlock()
	sunsetNoticeOnce.Do(func() {
		path := filepath.Join(sectionsDir, "sunset.yaml")
		if _, err := os.Stat(path); err == nil {
			slog.Info("SUNSET_CONFIG_DORMANT_NOTICE",
				"spec", "SPEC-V3R2-MIG-003 REQ-018",
				"yaml_path", path,
				"advice", "SunsetConfig has no runtime hot path; activation requires a new SPEC")
		}
	})
}

// resetSunsetNoticeOnce resets the once guard. FOR TESTING ONLY.
// Tests that need to verify the notice fires must call this before Loader.Load().
//
// Concurrency: protected by sunsetNoticeMu against concurrent
// emitSunsetDormantNotice readers. SPEC-V3R6-CI-BASELINE-DRIFT-001 M3.
func resetSunsetNoticeOnce() {
	sunsetNoticeMu.Lock()
	defer sunsetNoticeMu.Unlock()
	sunsetNoticeOnce = sync.Once{}
}
