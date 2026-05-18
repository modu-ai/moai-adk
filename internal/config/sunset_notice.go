package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// sunsetNoticeOnce guards the once-per-process DORMANT notice emission.
// REQ-MIG003-018: emitted at most once per process lifetime.
var sunsetNoticeOnce sync.Once

// emitSunsetDormantNotice emits a one-shot session log per REQ-MIG003-018.
// Guarded by sync.Once so it fires at most once per process lifetime.
// The notice is advisory only — it reminds future maintainers that
// SunsetConfig is template-only and has no runtime hot path.
//
// @MX:NOTE: [AUTO] DORMANT notice helper — fires when sunset.yaml exists,
// reminds future maintainers that SunsetConfig is template-only.
// REQ-MIG003-018, AC-MIG003-15.
func emitSunsetDormantNotice(sectionsDir string) {
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
func resetSunsetNoticeOnce() {
	sunsetNoticeOnce = sync.Once{}
}
