package backup

// mcp_snapshot.go keeps the base the per-file 3-way merge needs for .mcp.json
// (card t1029), the sibling of the .claude/settings.json snapshot.
//
// The defect it closes splits by delivery kind rather than by file: with a
// derived base, a key the template ADDS reaches the project while a value the
// template CHANGES on a key the user never touched does not. The update
// therefore looks successful — the new server appears — while the changed
// launcher command stays stale. Remembering the render a flow deployed makes
// that change visible, because the base then carries the OLD value and only
// the template side moved.
//
// The lifecycle, the manifest gate, and the staging/promotion split are the
// machinery in file_snapshot.go, unchanged from the settings.json sibling.
//
// What this does NOT deliver: a key the template RETIRES still stays in the
// user's file. That cell fails identically on the already-shipped
// settings.json snapshot path, which makes it a pre-existing defect of its own
// rather than a gap in this namespace; it is tracked separately.

import (
	"io"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

const (
	// liveMCPRel is the project-relative path of the file the snapshot records,
	// in the slash form the manifest keys use.
	liveMCPRel = ".mcp.json"

	// MCPSnapshotWriteFailedPrefix starts the single warning line printed when
	// the staging copy cannot be written. It is distinct from the settings and
	// sections prefixes so the three failures never read alike.
	MCPSnapshotWriteFailedPrefix = "mcp-snapshot-write-failed:"
	// MCPSnapshotPromoteFailedPrefix starts the single warning line printed when
	// the staging copy cannot replace the canonical base.
	MCPSnapshotPromoteFailedPrefix = "mcp-snapshot-promote-failed:"
)

// mcpSnapshot is the .mcp.json namespace: the copies sit directly under the
// snapshot root as a sibling of sections/ and claude/, under a name that drops
// the live path's leading dot so the cache never holds a file another tool
// could discover as MCP configuration.
var mcpSnapshot = fileSnapshot{
	name:                "mcp.json",
	liveRel:             liveMCPRel,
	writeFailedPrefix:   MCPSnapshotWriteFailedPrefix,
	promoteFailedPrefix: MCPSnapshotPromoteFailedPrefix,
}

// MCPSnapshotPath returns the canonical base the next update merges .mcp.json
// against: <projectRoot>/.moai/cache/template-snapshot/mcp.json.
func MCPSnapshotPath(projectRoot string) string {
	return mcpSnapshot.path(projectRoot)
}

// MCPSnapshotPendingPath returns the staging copy of the render a flow
// deployed: the canonical path plus ".pending".
func MCPSnapshotPendingPath(projectRoot string) string {
	return mcpSnapshot.pendingPath(projectRoot)
}

// LoadMCPSnapshot returns the canonical base when it exists, reads, and decodes
// as a JSON object. Anything else reports false and the merge keeps the derived
// base, byte-for-byte the pre-snapshot behaviour.
func LoadMCPSnapshot(projectRoot string) ([]byte, bool) {
	return mcpSnapshot.load(projectRoot)
}

// StageDeployedMCPSnapshot records the render the deploy just wrote to
// .mcp.json as the staging copy.
//
// @MX:ANCHOR: [AUTO] .mcp.json staging entry — fan_in 3 (init, template sync, clean reinstall)
// @MX:REASON: the recording moment is the contract; calling this after any later rewrite of
// .mcp.json turns that rewrite into template content
func StageDeployedMCPSnapshot(projectRoot string, m manifest.Manager, warn io.Writer) {
	mcpSnapshot.stageDeployed(projectRoot, m, warn)
}

// JudgeLeftoverMCPSnapshot settles a staging copy an earlier flow left behind
// because it stopped before its promotion decision.
//
// @MX:WARN: [AUTO] must run before the flow's first removal or rewrite of .mcp.json
// @MX:REASON: judged after the deploy the live file is already the new render, so "aborted with
// no revert → promote" silently becomes a discard
func JudgeLeftoverMCPSnapshot(projectRoot string, warn io.Writer) {
	mcpSnapshot.judgeLeftover(projectRoot, warn)
}

// SettleMCPSnapshot ends a flow's snapshot lifecycle at the flow's end, after
// its .mcp.json merge. preserved reports whether that merge took a preserve
// path, in which case the staging copy is discarded; otherwise it is promoted.
//
// @MX:WARN: [AUTO] promotion decision — call only after the flow's .mcp.json merge
// @MX:REASON: promoting before the merge makes base == updated (new keys read as user deletions);
// promoting a preserved flow's render makes its new keys read as user deletions next update
func SettleMCPSnapshot(projectRoot string, preserved bool, warn io.Writer) {
	mcpSnapshot.settle(projectRoot, preserved, warn)
}
