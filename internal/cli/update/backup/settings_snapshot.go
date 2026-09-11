package backup

// settings_snapshot.go keeps the base the per-file 3-way merge needs for
// .claude/settings.json (SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001).
//
// Without a stored base the merge derives one from the freshly deployed
// template, which makes a template value change to a key the user never
// touched invisible. The fix is to remember the render a flow deployed and
// merge the next update against it. One file cannot serve both moments,
// though: a render recorded right after the deploy would overwrite the base
// that same flow's merge is about to read, and with base == updated every
// template key the user lacks reads as a deletion. So the render is kept in
// two states:
//
//   - staging (claude/settings.json.pending) — written right after the deploy
//     wrote the render, read only by the promotion decision;
//   - canonical (claude/settings.json) — what the next flow's merge reads.
//
// A staging copy becomes canonical only by promotion, and never before the
// flow's merge has run. Every step is best-effort: a failure prints one
// prefixed warning line and the enclosing init/update carries on.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

const (
	// settingsSnapshotSubdir mirrors the last segment of .claude/ without the
	// leading dot, so the cache never holds a directory Claude Code might
	// discover as a project .claude/ tree.
	settingsSnapshotSubdir = "claude"
	settingsSnapshotName   = "settings.json"
	settingsPendingSuffix  = ".pending"

	// liveSettingsRel is the project-relative path of the file the snapshot
	// records, in the slash form the manifest keys use.
	liveSettingsRel = ".claude/settings.json"

	// SettingsSnapshotWriteFailedPrefix starts the single warning line printed
	// when the staging copy cannot be written (REQ-USB-010). It is distinct
	// from the sections snapshot warning so the two failures never read alike.
	SettingsSnapshotWriteFailedPrefix = "settings-snapshot-write-failed:"
	// SettingsSnapshotPromoteFailedPrefix starts the single warning line printed
	// when the staging copy cannot replace the canonical base (REQ-USB-010).
	SettingsSnapshotPromoteFailedPrefix = "settings-snapshot-promote-failed:"
)

// SettingsSnapshotPath returns the canonical base the next update merges
// .claude/settings.json against:
// <projectRoot>/.moai/cache/template-snapshot/claude/settings.json.
func SettingsSnapshotPath(projectRoot string) string {
	return filepath.Join(SnapshotDir(projectRoot), settingsSnapshotSubdir, settingsSnapshotName)
}

// SettingsSnapshotPendingPath returns the staging copy of the render a flow
// deployed: the canonical path plus ".pending".
func SettingsSnapshotPendingPath(projectRoot string) string {
	return SettingsSnapshotPath(projectRoot) + settingsPendingSuffix
}

// LoadSettingsSnapshot returns the canonical base when it exists, reads, and
// decodes as a JSON object (REQ-USB-004). Anything else reports false and the
// merge keeps the derived base (REQ-USB-009).
func LoadSettingsSnapshot(projectRoot string) ([]byte, bool) {
	data, err := os.ReadFile(SettingsSnapshotPath(projectRoot))
	if err != nil {
		return nil, false
	}
	var doc map[string]any
	if json.Unmarshal(data, &doc) != nil || doc == nil {
		return nil, false
	}
	return data, true
}

// StageDeployedSettingsSnapshot records the render the deploy just wrote to
// .claude/settings.json as the staging copy (REQ-USB-001).
//
// Whether the deploy wrote the file is decided from the deployer's own
// manifest record, never from the file alone (REQ-USB-016, plan.md D7): the
// entry must be template-managed and its template hash — the hash of the bytes
// the deployer wrote — must match the file. A deploy that skipped an existing
// user file records it as user-created or leaves the user's provenance in
// place, so the skip records nothing. The hash match also means the staged
// bytes are exactly the deployer's render, not a later rewrite (REQ-USB-003).
//
// @MX:ANCHOR: [AUTO] settings.json staging entry — fan_in 3 (init, template sync, clean reinstall)
// @MX:REASON: the recording moment is the contract; calling this after any later rewrite of
// settings.json (autonomy bundle, merge, deny-rule strip) turns that rewrite into template content
func StageDeployedSettingsSnapshot(projectRoot string, m manifest.Manager, warn io.Writer) {
	if m == nil {
		return
	}
	entry, ok := m.GetEntry(liveSettingsRel)
	if !ok || entry.Provenance != manifest.TemplateManaged {
		return
	}
	render, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(liveSettingsRel)))
	if err != nil || manifest.HashBytes(render) != entry.TemplateHash {
		return
	}
	pending := SettingsSnapshotPendingPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(pending), defs.DirPerm); err != nil {
		warnLine(warn, SettingsSnapshotWriteFailedPrefix, err)
		return
	}
	if err := os.WriteFile(pending, render, defs.FilePerm); err != nil {
		warnLine(warn, SettingsSnapshotWriteFailedPrefix, err)
	}
}

// JudgeLeftoverSettingsSnapshot settles a staging copy an earlier flow left
// behind because it stopped before its promotion decision (REQ-USB-005).
//
// The live file still equal to the leftover means nothing reverted the render
// after the stop, so it is promoted; any difference — a hand revert, or any
// other intervening write — discards it and keeps the prior base (spec.md §E
// N-08, fail-safe).
//
// @MX:WARN: [AUTO] must run before the flow's first removal or rewrite of .claude/settings.json
// @MX:REASON: judged after the deploy or the deny-rule strip, the live file is already the new
// render, so "aborted with no revert → promote" silently becomes a discard (plan.md B8, M-D5g)
func JudgeLeftoverSettingsSnapshot(projectRoot string, warn io.Writer) {
	leftover, err := os.ReadFile(SettingsSnapshotPendingPath(projectRoot))
	if err != nil {
		return
	}
	live, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(liveSettingsRel)))
	if err == nil && bytes.Equal(live, leftover) {
		promoteSettingsSnapshot(projectRoot, warn)
		return
	}
	discardSettingsSnapshot(projectRoot)
}

// SettleSettingsSnapshot ends a flow's snapshot lifecycle at the flow's end,
// after its settings.json merge (REQ-USB-005). preserved reports whether that
// merge took a preserve path — wrote the pre-flow user file back wholesale —
// in which case the live file does not reflect the render and the staging copy
// is discarded; otherwise it is promoted. A flow without a staging copy is a
// no-op.
//
// @MX:WARN: [AUTO] promotion decision — call only after the flow's settings.json merge
// @MX:REASON: promoting before the merge makes base == updated (new keys read as user deletions);
// promoting a preserved flow's render makes its new keys read as user deletions next update
func SettleSettingsSnapshot(projectRoot string, preserved bool, warn io.Writer) {
	if _, err := os.Stat(SettingsSnapshotPendingPath(projectRoot)); err != nil {
		return
	}
	if preserved {
		discardSettingsSnapshot(projectRoot)
		return
	}
	promoteSettingsSnapshot(projectRoot, warn)
}

// promoteSettingsSnapshot moves the staging copy onto the canonical path with a
// same-directory rename, which replaces an existing file on every platform.
func promoteSettingsSnapshot(projectRoot string, warn io.Writer) {
	if err := os.Rename(SettingsSnapshotPendingPath(projectRoot), SettingsSnapshotPath(projectRoot)); err != nil {
		warnLine(warn, SettingsSnapshotPromoteFailedPrefix, err)
	}
}

func discardSettingsSnapshot(projectRoot string) {
	_ = os.Remove(SettingsSnapshotPendingPath(projectRoot))
}

func warnLine(warn io.Writer, prefix string, err error) {
	if warn == nil {
		return
	}
	_, _ = fmt.Fprintf(warn, "%s %v\n", prefix, err)
}
