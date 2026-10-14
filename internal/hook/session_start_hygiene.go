package hook

import (
	"log/slog"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hygiene"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// hygieneRunFn is the SessionStart hygiene seam (SPEC-MOAI-HYGIENE-001
// REQ-HYG-014, AC-HYG-013): the deferred advisory pass invokes it
// best-effort; tests stub it to fail and assert the hook still allows the
// launch. Production installs runHygieneBestEffort.
//
// @MX:WARN @MX:REASON: the seam runs inside the SessionStart deferred
// pass — a production caller installing a blocking or mutating function
// here would put hygiene on the launch critical path this wiring exists
// to keep clear.
var hygieneRunFn = runSessionHygiene

// runSessionHygiene executes one hygiene pass for the project in the
// configured mode. The SessionStart auto path is governed by
// workflow.hygiene.mode alone — apply (the shipped default as of
// SPEC-CAPACITY-RECLAIM-001 AC-CAP-008) runs the non-destructive categories
// (audit-log rotation, closed-registry state GC); report never rotates or
// deletes anything (REQ-HYG-013). Every failure is logged and swallowed: the
// session launch is never blocked or delayed by hygiene (REQ-HYG-014's
// best-effort contract).
func runSessionHygiene(projectDir string) error {
	settings := hygiene.LoadSettingsFrom(projectDir)
	if err := settings.Validate(); err != nil {
		// D30: a config-invalid value refuses the pass — logged, never
		// propagated, never mutated.
		slog.Warn("session start (deferred): hygiene config invalid; pass skipped",
			"error", err.Error())
		return nil
	}
	mode := settings.ApplyMode(false)
	engine := &hygiene.Engine{
		LogDir:           filepath.Join(projectDir, ".moai", "logs"),
		MoaiRoot:         filepath.Join(projectDir, ".moai"),
		RegistryPath:     filepath.Join(projectDir, ".moai", "state", "active-sessions.json"),
		TranscriptRoots:  hygiene.DefaultTranscriptRoots(),
		MaxBytes:         settings.AuditLogMaxBytes,
		KeptRotations:    settings.AuditLogKeptRotations,
		MinAge:           settings.MinAge(),
		TranscriptWindow: settings.TranscriptActivityWindow,
		HeartbeatWindow:  settings.HeartbeatStaleWindow,
	}
	_, _, rotErr, gcErr := engine.Run(mode)
	if rotErr != nil {
		slog.Warn("session start (deferred): hygiene rotator failed (non-blocking)",
			"error", rotErr.Error())
	}
	if gcErr != nil {
		slog.Warn("session start (deferred): hygiene gc failed (non-blocking)",
			"error", gcErr.Error())
	}
	runProfileCapAdvisory()
	return nil
}

// runProfileCapAdvisory is the SessionStart profile-cap advisory check
// (SPEC-CAPACITY-RECLAIM-001 REQ-CAP-009): it reads the HOME-tier
// state.profile_max_bytes and logs one advisory per profile above the cap.
// Warning-only — nothing is deleted; space is reclaimed with
// `moai clean --home --force`. Time-boxed (Advisory-Check Discipline): the
// profile sizing walk runs under config.ProfileCapAdvisoryTimeout, and every
// failure (absent key, unreadable config, deadline) is logged and swallowed —
// the session launch is never blocked or delayed by this check.
func runProfileCapAdvisory() {
	capBytes, ok, err := hygiene.LoadProfileCapMaxBytes()
	if err != nil {
		slog.Debug("session start (deferred): profile-cap advisory skipped (config unreadable)",
			"error", err.Error())
		return
	}
	if !ok {
		return // absent key: no warning (REQ-CAP-009); no compiled fallback
	}
	profilesDir, err := paths.ProfilesDir()
	if err != nil {
		slog.Debug("session start (deferred): profile-cap advisory skipped (home unresolved)",
			"error", err.Error())
		return
	}
	advisories, advErr := hygiene.ProfileCapAdvisories(profilesDir, capBytes, config.ProfileCapAdvisoryTimeout)
	for _, msg := range advisories {
		slog.Warn("session start (deferred): " + msg)
	}
	if advErr != nil {
		slog.Debug("session start (deferred): profile-cap advisory incomplete (time-boxed, non-blocking)",
			"error", advErr.Error())
	}
}
