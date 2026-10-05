package hook

import (
	"log/slog"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/hygiene"
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
// workflow.hygiene.mode alone — report (the shipped default) never rotates
// or deletes anything; apply is the operator's explicit config opt-in
// (REQ-HYG-013). Every failure is logged and swallowed: the session launch
// is never blocked or delayed by hygiene (REQ-HYG-014's best-effort
// contract).
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
	return nil
}
