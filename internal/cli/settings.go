package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/lockfile"
)

// readSettingsMap reads .claude/settings.local.json as a generic map, preserving
// every top-level key (including keys the SettingsLocal struct does not model —
// hooks, outputStyle, model, defaultMode, future keys). Returns a fresh empty map
// when the file is absent (EC-1) or empty (EC-2).
//
// SPEC-CLIFIX-CRITICAL-001 REQ-CRIT-001-001: every write-back path must round-trip
// the file as map[string]any so unknown top-level keys are never wiped.
func readSettingsMap(path string) (map[string]any, error) {
	m := make(map[string]any)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, fmt.Errorf("read settings.local.json: %w", err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse settings.local.json: %w", err)
		}
	}
	return m, nil
}

// settingsEnvMap returns the env sub-map from m, creating it if absent. The env
// values are stored as any (strings at runtime); callers type-assert when reading.
func settingsEnvMap(m map[string]any) map[string]any {
	if em, ok := m["env"].(map[string]any); ok {
		return em
	}
	em := make(map[string]any)
	m["env"] = em
	return em
}

// mutateSettingsLocal performs a single locked + atomic read-modify-write on a
// .claude/settings.local.json file. It is the canonical mutation helper for the
// launch path (REQ-CGH-005): concurrent moai cc/glm/cg launches that race on the
// same settings file are serialized via a file lock, and the write is atomic
// (temp-file + os.Rename) so a partial write can never leave a truncated or
// clobbered file on disk.
//
// SPEC-CLIFIX-CRITICAL-001 REQ-CRIT-001-001: the mutate closure receives the
// file as a map[string]any so every top-level key — including keys the
// SettingsLocal struct does not model (hooks, outputStyle, model, defaultMode,
// and any future keys) — survives the round-trip. The closed SettingsLocal struct
// remains available for read-only convenience but is NEVER marshaled back to disk.
//
// The lock is advisory (flock on POSIX, a process-local mutex on Windows — see
// team_spawn_lock_{unix,windows}.go). It is acquired on a sibling lock file so the
// atomic os.Rename of the settings file itself is never blocked by an open handle
// on the target path.
func mutateSettingsLocal(path string, mutate func(map[string]any)) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Acquire the advisory lock on a sibling .lock file. Locking a separate file
	// keeps the target free for the temp-file + os.Rename atomic swap below.
	lockPath := path + ".lock"
	lockF, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("open settings lock file: %w", err)
	}
	defer func() { _ = lockF.Close() }()

	if err := lockfile.Lock(lockF); err != nil {
		return fmt.Errorf("acquire settings lock: %w", err)
	}
	defer func() { _ = lockfile.Unlock(lockF) }()

	m, err := readSettingsMap(path)
	if err != nil {
		return err
	}

	mutate(m)

	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := writeFileAtomic(path, out, 0o600); err != nil {
		return fmt.Errorf("write settings.local.json: %w", err)
	}

	return nil
}

// writeFileAtomic writes data to path atomically via a temp file in the same
// directory followed by os.Rename. os.Rename within a single filesystem is
// atomic, so a concurrent reader observes either the old or the new file —
// never a partial write.
//
// SPEC-CLIFIX-CONCURRENCY-001 M3: this is the single consolidated atomic-writer
// helper for internal/cli. All former atomic-write sites (atomicWriteJSON in
// update_noise.go, writeClaudeJSONAtomic + writeClaudeJSONBytes in glm_tools.go,
// the inline tmp+rename blocks in harness_mute.go and glm.go saveLLMSection)
// route through this helper. The perm parameter + tmp.Chmod preserve the
// per-caller file-mode contract (0600 for credential-bearing files like
// settings.local.json / ~/.claude.json; 0644 for non-credential config like
// workflow.yaml). os.MkdirAll is a safe superset added during consolidation
// (writeClaudeJSONBytes and preference/atomicWrite already had it). No fsync —
// none of the former callers relied on crash-durability beyond rename atomicity.
//
// @MX:ANCHOR: [AUTO] consolidated atomic-writer helper (internal/cli)
// @MX:REASON: union-of-guarantees contract — single temp+rename+Chmod mechanism shared by 5 former atomic-write sites; perm param preserves per-caller file-mode (0600 credential vs 0644 config)
// @MX:SPEC: SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-004 (M3 consolidation)
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir for atomic write: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".settings-local-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	return os.Rename(tmpName, path)
}

// stripGLMCredsAndSetTeammateMode is the CG-leader-cleanup mutate closure used by
// applyCGMode (REQ-CGH-002 ordering + REQ-CGH-003 single write): it strips the GLM
// credential keys from the leader's env block AND sets teammateMode="tmux" in one
// read-modify-write, so no intermediate file state exists where teammateMode is
// absent. The credential-stripping logic mirrors removeGLMEnv's env handling, but
// teammateMode is set to "tmux" (CG leader stays clean of GLM creds but keeps the
// tmux display mode) rather than cleared.
//
// SPEC-CLIFIX-CRITICAL-001: operates on map[string]any so unknown top-level keys
// survive the round-trip.
func stripGLMCredsAndSetTeammateMode(m map[string]any) {
	if env, ok := m["env"].(map[string]any); ok {
		// Restore backed-up OAuth token before removing GLM vars.
		if backup, bok := env["MOAI_BACKUP_AUTH_TOKEN"].(string); bok && backup != "" {
			env["ANTHROPIC_AUTH_TOKEN"] = backup
			delete(env, "MOAI_BACKUP_AUTH_TOKEN")
		} else {
			delete(env, "ANTHROPIC_AUTH_TOKEN")
		}
		delete(env, "ANTHROPIC_BASE_URL")
		delete(env, "ANTHROPIC_DEFAULT_HAIKU_MODEL")
		delete(env, "ANTHROPIC_DEFAULT_SONNET_MODEL")
		delete(env, "ANTHROPIC_DEFAULT_OPUS_MODEL")
		delete(env, "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS")
		delete(env, "API_TIMEOUT_MS")
		delete(env, "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC")
		delete(env, "CLAUDE_CODE_TEAMMATE_DISPLAY")
		delete(env, "MOAI_STATUSLINE_CONTEXT_SIZE")

		if len(env) == 0 {
			delete(m, "env")
		}
	}

	// CG mode requires tmux display so teammates spawn in panes and inherit the
	// GLM session env (#468). Set it in this same RMW so there is no window where
	// teammateMode is absent.
	m["teammateMode"] = "tmux"
}
