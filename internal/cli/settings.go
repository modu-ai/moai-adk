package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// mutateSettingsLocal performs a single locked + atomic read-modify-write on a
// .claude/settings.local.json file. It is the canonical mutation helper for the
// launch path (REQ-CGH-005): concurrent moai cc/glm/cg launches that race on the
// same settings file are serialized via a file lock, and the write is atomic
// (temp-file + os.Rename, mirroring saveLLMSection in glm.go) so a partial write
// can never leave a truncated or clobbered file on disk.
//
// The mutate closure receives the current settings (or a zero value when the file
// is absent or empty — EC-1/EC-2) and mutates it in place. User-only keys
// (defaultMode, env.PATH, and any unrecognized keys) are preserved by the
// SettingsLocal struct round-trip, so a closure that only touches GLM credential
// keys + teammateMode leaves everything else intact (per internal/cli/CLAUDE.md
// settings-helper convention + CLAUDE.local.md §22).
//
// The lock is advisory (flock on POSIX, a process-local mutex on Windows — see
// team_spawn_lock_{unix,windows}.go). It is acquired on a sibling lock file so the
// atomic os.Rename of the settings file itself is never blocked by an open handle
// on the target path.
func mutateSettingsLocal(path string, mutate func(*SettingsLocal)) error {
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

	if err := lockFile(lockF); err != nil {
		return fmt.Errorf("acquire settings lock: %w", err)
	}
	defer func() { _ = unlockFile(lockF) }()

	// Read current settings, tolerating an absent (EC-1) or empty (EC-2) file.
	var settings SettingsLocal
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read settings.local.json: %w", err)
		}
		// Absent file → zero-value settings.
	} else if len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parse settings.local.json: %w", err)
		}
	}

	mutate(&settings)

	out, err := json.MarshalIndent(settings, "", "  ")
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
// never a partial write (mirrors the saveLLMSection pattern in glm.go).
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
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
func stripGLMCredsAndSetTeammateMode(s *SettingsLocal) {
	if s.Env != nil {
		// Restore backed-up OAuth token before removing GLM vars.
		if backup, ok := s.Env["MOAI_BACKUP_AUTH_TOKEN"]; ok && backup != "" {
			s.Env["ANTHROPIC_AUTH_TOKEN"] = backup
			delete(s.Env, "MOAI_BACKUP_AUTH_TOKEN")
		} else {
			delete(s.Env, "ANTHROPIC_AUTH_TOKEN")
		}
		delete(s.Env, "ANTHROPIC_BASE_URL")
		delete(s.Env, "ANTHROPIC_DEFAULT_HAIKU_MODEL")
		delete(s.Env, "ANTHROPIC_DEFAULT_SONNET_MODEL")
		delete(s.Env, "ANTHROPIC_DEFAULT_OPUS_MODEL")
		delete(s.Env, "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS")
		delete(s.Env, "API_TIMEOUT_MS")
		delete(s.Env, "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC")
		delete(s.Env, "CLAUDE_CODE_TEAMMATE_DISPLAY")
		delete(s.Env, "MOAI_STATUSLINE_CONTEXT_SIZE")

		if len(s.Env) == 0 {
			s.Env = nil
		}
	}

	// CG mode requires tmux display so teammates spawn in panes and inherit the
	// GLM session env (#468). Set it in this same RMW so there is no window where
	// teammateMode is absent.
	s.TeammateMode = "tmux"
}
