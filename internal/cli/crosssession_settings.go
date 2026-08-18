package cli

// crosssession_settings.go — the crosssession.yaml → transient --settings
// translation consumed by every launcher (moai cc / glm / cg).
//
// The user's cross-session messaging preferences live in
// .moai/config/sections/crosssession.yaml (edited by hand or through the web
// console's settings seam). The launchers translate them into a session-private
// settings file passed via the backend's --settings flag — a trusted source in
// Claude Code's precedence order, so an `accept` injected there applies over
// user/project values on the accept < hold < refuse ladder.
//
// Translation table (moai key → Claude Code settings key):
//
//	inbound         → crossSessionInbound   (closed set: accept|hold|refuse)
//	isolate_machines→ isolatePeerMachines   (true ONLY on explicit opt-in)
//	dialog_expiry   → dialogExpiry          (closed set: 60s|5m|10m|never)
//
// The isolate_machines column is guard-pinned: a `true` from ANY Claude Code
// settings scope applies and cannot be turned off from a lower scope, so the
// launcher emits isolatePeerMachines only when the user wrote
// `isolate_machines: true` in their own crosssession.yaml — never from a
// template default or a launcher fallback. TestTemplateNeverShipsIsolatePeerMachines
// and TestLauncherNeverInjectsIsolatePeerMachinesByDefault pin both halves.
//
// Fail-open throughout: an unreadable or invalid config degrades to injecting
// nothing (launch proceeds with Claude Code's own defaults), never to blocking
// the launch.

import (
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
)

// crossSessionConfigRootFn resolves the project root the crosssession.yaml
// read uses. Tests override it to point at a t.TempDir() so launcher tests
// never read the host project's real config.
var crossSessionConfigRootFn = launchProjectRoot

// crossSessionSettingsPayload reads the project's crosssession.yaml and
// translates it into the Claude Code settings payload. Neutral or invalid
// values are omitted; an all-neutral (or absent) config yields an EMPTY map,
// which callers treat as "inject nothing". The closed value sets are the
// shared config accessors (config.ValidCrossSession*) — the same sets the
// web console select options derive from.
func crossSessionSettingsPayload(root string) map[string]any {
	sections := filepath.Join(root, ".moai", "config", "sections")
	cfg, err := config.LoadCrossSessionConfig(sections)
	if err != nil {
		// Fail-open: unreadable config injects nothing.
		return map[string]any{}
	}
	payload := map[string]any{}
	if config.IsValidCrossSessionInbound(cfg.Inbound) {
		payload["crossSessionInbound"] = cfg.Inbound
	}
	if config.IsValidCrossSessionDialogExpiry(cfg.DialogExpiry) {
		payload["dialogExpiry"] = cfg.DialogExpiry
	}
	if cfg.IsolateMachines {
		// The ONLY emission path: the user's explicit opt-in. Never a
		// template default, never a launcher fallback — see the file comment.
		payload["isolatePeerMachines"] = true
	}
	return payload
}

// appendCrossSessionSettings is the general-launch injection consumed by
// unifiedLaunchDefault (covers moai cc / glm / cg). It no-ops when the
// operator supplied --settings themselves (their intent wins, and this also
// covers args that already carry the kanban-injected flag) or when the config
// is neutral. The transient file is session-private under os.TempDir();
// on POSIX the launch execs away before any cleanup could run, leaving the
// file to the OS, exactly like the kanban injection.
//
// @MX:NOTE: [AUTO] guards TestTemplateNeverShipsIsolatePeerMachines + TestLauncherNeverInjectsIsolatePeerMachinesByDefault
func appendCrossSessionSettings(root string, args []string) []string {
	if operatorSuppliedSettings(args) {
		return args
	}
	payload := crossSessionSettingsPayload(root)
	if len(payload) == 0 {
		return args
	}
	path, err := writeTransientSettingsFile(payload, "moai-crosssession")
	if err != nil {
		// Fail-open: launch without the injected --settings.
		return args
	}
	return append(args, settingsFlagLong, path)
}
