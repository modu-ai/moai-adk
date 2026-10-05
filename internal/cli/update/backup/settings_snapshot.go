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
//
// The lifecycle itself lives in file_snapshot.go — .mcp.json is a sibling
// namespace on the same machinery (card t1029). This file declares only the
// settings.json namespace and the exported surface its callers use.

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

const (
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

// settingsSnapshot is the .claude/settings.json namespace. Its subdir mirrors
// the last segment of .claude/ without the leading dot, so the cache never
// holds a directory Claude Code might discover as a project .claude/ tree.
var settingsSnapshot = fileSnapshot{
	subdir:              "claude",
	name:                "settings.json",
	liveRel:             liveSettingsRel,
	writeFailedPrefix:   SettingsSnapshotWriteFailedPrefix,
	promoteFailedPrefix: SettingsSnapshotPromoteFailedPrefix,
}

// SettingsSnapshotPath returns the canonical base the next update merges
// .claude/settings.json against:
// <projectRoot>/.moai/cache/template-snapshot/claude/settings.json.
func SettingsSnapshotPath(projectRoot string) string {
	return settingsSnapshot.path(projectRoot)
}

// SettingsSnapshotPendingPath returns the staging copy of the render a flow
// deployed: the canonical path plus ".pending".
func SettingsSnapshotPendingPath(projectRoot string) string {
	return settingsSnapshot.pendingPath(projectRoot)
}

// LoadSettingsSnapshot returns the canonical base when it exists, reads, and
// decodes as a JSON object (REQ-USB-004). Anything else reports false and the
// merge keeps the derived base (REQ-USB-009).
func LoadSettingsSnapshot(projectRoot string) ([]byte, bool) {
	return settingsSnapshot.load(projectRoot)
}

// StageDeployedSettingsSnapshot records the render the deploy just wrote to
// .claude/settings.json as the staging copy (REQ-USB-001, REQ-USB-016).
//
// @MX:ANCHOR: [AUTO] settings.json staging entry — fan_in 3 (init, template sync, clean reinstall)
// @MX:REASON: the recording moment is the contract; calling this after any later rewrite of
// settings.json (autonomy bundle, merge, deny-rule strip) turns that rewrite into template content
func StageDeployedSettingsSnapshot(projectRoot string, m manifest.Manager, warn io.Writer) {
	settingsSnapshot.stageDeployed(projectRoot, m, warn)
}

// JudgeLeftoverSettingsSnapshot settles a staging copy an earlier flow left
// behind because it stopped before its promotion decision (REQ-USB-005).
//
// @MX:WARN: [AUTO] must run before the flow's first removal or rewrite of .claude/settings.json
// @MX:REASON: judged after the deploy or the deny-rule strip, the live file is already the new
// render, so "aborted with no revert → promote" silently becomes a discard (plan.md B8, M-D5g)
func JudgeLeftoverSettingsSnapshot(projectRoot string, warn io.Writer) {
	settingsSnapshot.judgeLeftover(projectRoot, warn)
}

// SettleSettingsSnapshot ends a flow's snapshot lifecycle at the flow's end,
// after its settings.json merge (REQ-USB-005). preserved reports whether that
// merge took a preserve path — wrote the pre-flow user file back wholesale —
// in which case the staging copy is discarded; otherwise it is promoted.
//
// @MX:WARN: [AUTO] promotion decision — call only after the flow's settings.json merge
// @MX:REASON: promoting before the merge makes base == updated (new keys read as user deletions);
// promoting a preserved flow's render makes its new keys read as user deletions next update
func SettleSettingsSnapshot(projectRoot string, preserved bool, warn io.Writer) {
	settingsSnapshot.settle(projectRoot, preserved, warn)
}

func warnLine(warn io.Writer, prefix string, err error) {
	if warn == nil {
		return
	}
	_, _ = fmt.Fprintf(warn, "%s %v\n", prefix, err)
}
