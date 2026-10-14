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
	runProfileCapAdvisoryAsync(projectDir)
	return nil
}

// profileCapAdvisoryFn is the advisory body (the seam tests stub to model a
// slow or failing walk).
var profileCapAdvisoryFn = runProfileCapAdvisory

// profileCapSpawnFn installs the profile-cap advisory OFF the deferred pass's
// critical path (sync-audit P2-1, Advisory-Check Discipline): production
// spawns it in its OWN goroutine, so the profile-sizing walk — the one step
// whose runtime grows with the very profiles the 8 GiB default targets —
// never consumes the deferredScanJoinBound (250ms) window that the deferred
// pass's sibling steps share. The walk keeps its own time-box
// (config.ProfileCapAdvisoryTimeout) inside the goroutine: on deadline the
// advisory degrades to a skip notice and the goroutine exits cleanly. The
// split-behavior test installs a synchronous variant for determinism.
//
// @MX:DEBT: the advisory goroutine is detached, so it races the short-lived
// hook process's exit (session_start.go spawnDeferredAdvisoryScans documents
// the same gap): a very large profile walk can be torn down before its
// warning lands — the advisory is silently undelivered for that session.
// @MX:CEILING: warning-only loss — the walk deletes nothing, so the worst
// case is a skipped one-session advisory; every session retries for free.
// @MX:UPGRADE: a separate-process or early-session cached writer (the
// Advisory-Check Discipline's cached clause — key the result on a cheap
// invariant and fill it out of band, as the drift cache does) closes the
// race; revisit when a real miss is observed in the wild.
// @MX:SPEC: SPEC-CAPACITY-RECLAIM-001
var profileCapSpawnFn = func(projectDir string) {
	go profileCapAdvisoryFn(projectDir)
}

// runProfileCapAdvisoryAsync installs the advisory for one deferred pass.
func runProfileCapAdvisoryAsync(projectDir string) {
	profileCapSpawnFn(projectDir)
}

// runProfileCapAdvisory is the SessionStart profile-cap advisory check
// (SPEC-CAPACITY-RECLAIM-001 REQ-CAP-009): it reads state.profile_max_bytes
// (HOME tier first, then the project tier) and logs one advisory per profile
// above the cap. Warning-only — nothing is deleted; space is reclaimed with
// `moai clean --home --force`. Time-boxed (Advisory-Check Discipline): the
// profile sizing walk runs under config.ProfileCapAdvisoryTimeout, and every
// failure (absent key, unreadable config, deadline) is logged and swallowed —
// the session launch is never blocked or delayed by this check.
func runProfileCapAdvisory(projectDir string) {
	capBytes, ok, err := hygiene.LoadProfileCapMaxBytes(projectDir)
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
