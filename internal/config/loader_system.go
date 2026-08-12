package config

import (
	"log/slog"
	"path/filepath"
)

// loader_system.go — narrow system.yaml section loader (SPEC-CONFIG-KEY-HONESTY-001 M4).
//
// Prior to M4, audit_registry.go mapped "system" -> "SystemConfig" but no
// Loader.Load path read system.yaml, so the parity guard reported GREEN on a
// binding that did not exist (spec.md §A.2 finding F2). The two real consumers
// of the file (internal/hook/routing_ledger.go HookObserveOptInEnabled and
// internal/cli/update.go readHookOptInEnabled) each parsed it through a private
// inline anonymous struct, bypassing the loaded Config entirely.
//
// M4 adds a narrow loadSystemSection that binds the ONE block with a genuine
// SystemConfig consumer (hook.*), wired into Loader.Load. The shared helper
// LoadSystemHookOptInEnabled consolidates the two inline-struct readers onto a
// single parse path so the hook opt-in value is read the same way everywhere.
//
// The moai / github / document_management blocks in system.yaml have no
// SystemConfig field; binding them is out of scope (spec.md §C) and they remain
// classified R in the M1 inventory.

// loadSystemSection loads the hook block of the system configuration section
// from system.yaml. The wrapper is seeded with the populated default
// (cfg.System.Hook) so a system.yaml that declares the block but omits keys
// retains the construction-time defaults instead of collapsing them to zero
// (partial-override contract, parallel to loadGateSection / loadHandoffSection).
func (l *Loader) loadSystemSection(dir string, cfg *Config) {
	wrapper := &systemFileWrapper{Hook: cfg.System.Hook}
	loaded, err := loadYAMLFile(dir, "system.yaml", wrapper)
	if err != nil {
		slog.Warn("failed to load system config, using defaults", "error", err)
		return
	}
	if loaded {
		cfg.System.Hook = wrapper.Hook
		l.loadedSections["system"] = true
	}
}

// LoadSystemHookOptInEnabled reports whether the hook observe-opt-in master
// toggle (hook.opt_in.enabled in system.yaml) is on at the given project root.
//
// It is the single parse path shared by every consumer of that toggle (the
// routing-ledger gate and the `moai update` settings renderer). Fail-CLOSED,
// deliberately opposite to HarnessLearningEnabled: this is an opt-in toggle
// whose default state is OFF, so a missing file, an unreadable file, a parse
// error, or an absent key all yield false. Every L1 emission path therefore
// ships inert to distributed users.
//
// The value this returns is identical to cfg.System.Hook.OptIn.Enabled after
// Loader.Load reads the same file via loadSystemSection; callers that already
// hold a loaded *Config should read that field directly instead of re-reading
// the file here.
func LoadSystemHookOptInEnabled(projectRoot string) bool {
	dir := filepath.Join(projectRoot, ".moai", "config", "sections")
	wrapper := &systemFileWrapper{Hook: NewDefaultSystemConfig().Hook}
	loaded, err := loadYAMLFile(dir, "system.yaml", wrapper)
	if err != nil || !loaded {
		return false
	}
	return wrapper.Hook.OptIn.Enabled
}
