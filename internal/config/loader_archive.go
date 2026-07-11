package config

import "log/slog"

// loader_archive.go — archive.yaml section loader + grace-window accessor.
//
// SPEC-SESSIONSTART-PERF-001 REQ-SSP-012 (Template-First config) / REQ-SSP-018
// (no hardcoded thresholds): the SPEC auto-archive grace window is a tunable
// value, so `moai spec archive` reads it from config rather than carrying an
// inline literal.

// loadArchiveSection loads the archive configuration section from archive.yaml.
//
// The wrapper is seeded with the populated default (cfg.Archive) so an
// archive.yaml that declares the section but omits grace_days retains the
// construction-time default instead of collapsing it to zero (partial-override
// contract, parallel to loadHandoffSection / loadFeedbackSection).
func (l *Loader) loadArchiveSection(dir string, cfg *Config) {
	wrapper := &archiveFileWrapper{Archive: cfg.Archive}
	loaded, err := loadYAMLFile(dir, "archive.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load archive config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.Archive = wrapper.Archive
		l.loadedSections["archive"] = true
	}
}

// ArchiveGraceDays returns the resolved SPEC auto-archive grace window in days.
//
// Resolution: the loaded ArchiveConfig.GraceDays when positive; otherwise the
// compiled default (DefaultArchiveGraceDays). The non-positive fallback is
// load-bearing, not cosmetic: a zero — from an absent archive.yaml, a section
// that omits grace_days, or an explicit `grace_days: 0` — must read as "unset",
// never as "no grace at all". Treating it literally would make every terminal
// SPEC instantly archive-eligible.
func (c *Config) ArchiveGraceDays() int {
	if c.Archive.GraceDays <= 0 {
		return DefaultArchiveGraceDays
	}
	return c.Archive.GraceDays
}
