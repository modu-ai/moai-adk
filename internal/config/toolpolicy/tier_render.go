package toolpolicy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/config/atomicfile"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// SPEC-AUTONOMY-TIERS-001 M3 — tier → permission-bundle renderer.
//
// This renderer is a NEW CALLER of the existing toolpolicy codegen
// (BuildPermissions + RenderSettingsJSON), NOT a parallel writer (AP-2). It
// splits the tier bundle across the two Claude Code config scopes per
// REQ-003 / spec.md §C:
//   - USER scope (~/.claude/settings.json): permissions.defaultMode.
//     auto/bypassPermissions are USER-scope-only per Claude Code v2.1.142+
//     (PROJECT/local cannot grant them downward) — writing them to PROJECT
//     would silently fail (AP-6).
//   - PROJECT scope (<project>/.claude/settings.json): permissions.deny +
//     permissions.ask. These merge across all sessions and are tier-INVARIANT
//     (REQ-004) — the deny/ask rule set is identical at every tier.
//
// The caller resolves the tier → defaultMode (via config.TierDefaultMode +
// config.EffectiveTierWithGates) and passes the resolved defaultMode string.
// This keeps the tier logic in internal/config and the rendering here, with no
// import cycle (config does not import toolpolicy).

// @MX:NOTE: [AUTO] tier→permission-bundle renderer — USER scope (defaultMode) + PROJECT scope (deny/ask) split per REQ-003 / v2.1.142+ scope rules

// TierRenderResult reports the two files written by RenderTierPermissions.
type TierRenderResult struct {
	UserPath    string
	ProjectPath string
}

// RenderTierPermissions writes the tier bundle across the two config scopes:
//
//   - userSettingsPath (USER scope, typically ~/.claude/settings.json):
//     permissions.defaultMode is set to defaultMode. Existing deny/ask arrays
//     in the USER file are CLEARED (deny/ask are PROJECT-scoped per REQ-003;
//     leaving a stale USER deny/ask would shadow the PROJECT rule set).
//
//   - projectSettingsPath (PROJECT scope, <project>/.claude/settings.json):
//     permissions.deny + permissions.ask are regenerated from doc. The
//     defaultMode in PROJECT is reset to "default" (or removed) — auto /
//     bypassPermissions MUST NOT land in PROJECT (AP-6 / v2.1.142+).
//
// deny/ask are tier-INVARIANT (AC-004): doc is loaded once and the same deny/
// ask arrays are written regardless of defaultMode. semi-auto (defaultMode=
// "default") produces byte-identical output to today's template (AC-007).
func RenderTierPermissions(projectSettingsPath, userSettingsPath, defaultMode string, doc *PolicyDocument) (*TierRenderResult, error) {
	if doc == nil {
		return nil, fmt.Errorf("nil policy document")
	}

	// Build the full permissions block once (defaultMode + deny + ask), then
	// split it across the two scopes.
	full, _, err := BuildPermissions(doc, defaultMode)
	if err != nil {
		return nil, fmt.Errorf("build permissions: %w", err)
	}

	// USER scope: defaultMode ONLY. deny/ask are PROJECT-scoped, so a USER
	// block carries just defaultMode (plus any preserved Raw extras).
	userBlock := &PermissionsBlock{
		DefaultMode: full.DefaultMode,
		Raw:         map[string]json.RawMessage{},
	}

	if err := renderIntoFile(userSettingsPath, userBlock); err != nil {
		return nil, fmt.Errorf("USER scope render: %w", err)
	}

	// Record-protocol serialization (card t1630 sync-audit F2): the whole
	// span — record read, Q5 classification, stage-ahead, PROJECT settings
	// write, converge/rollback — is serialized so a concurrent render cannot
	// interleave a stale converge past a newer one and invert the record
	// against the settings file (which resurrects revoked entries). One
	// package-level mutex is deliberately coarser than a per-project lock
	// registry: record spans are rare file writes, and a single primitive
	// avoids a global registry map. BuildInto holds the same mutex for its
	// TargetJSON (project settings) span.
	recordProtocolMu.Lock()
	defer recordProtocolMu.Unlock()

	// Inter-process lock (card t1630 sync-audit F2'): two concurrent moai OS
	// processes interleave outside recordProtocolMu, so the same span is also
	// guarded by the project-scoped lockfile. Mutex first, file lock second —
	// intra-process callers queue on the mutex and only one touches the
	// lockfile.
	releaseRecordLock, err := acquireProjectRecordLock(projectSettingsPath)
	if err != nil {
		return nil, fmt.Errorf("acquire record lock: %w", err)
	}
	defer releaseRecordLock()

	// The Q5 classification reads the PREVIOUS render's record and the
	// current project file — inside the protocol span, per F2.
	prior := readManagedAllowRecordList(projectSettingsPath)
	projectBlock := &PermissionsBlock{
		Allow: mergeProjectAllow(projectSettingsPath, full.Allow), // regenerated ∪ user-added
		Ask:   full.Ask,
		Deny:  full.Deny,
		Raw:   map[string]json.RawMessage{},
	}

	// Record-ahead (card t1630 card-review round 2, P2-1): stage the record
	// BEFORE the PROJECT settings mutation. If the project write later fails,
	// the record is rolled back to prior (F3) — see the failure path below.
	if err := stageManagedAllowRecord(projectSettingsPath, full.Allow); err != nil {
		return nil, fmt.Errorf("stage managed-allow record: %w", err)
	}
	if err := renderIntoFile(projectSettingsPath, projectBlock); err != nil {
		// Rollback (card t1630 sync-audit F3): a failed settings write must
		// not leave the staged union behind — an entry it upgraded to managed
		// (because the policy coincidentally generates it) would be dropped
		// by a later policy revision, deleting a user-authored allow entry.
		if rbErr := rollbackManagedAllowRecord(projectSettingsPath, prior); rbErr != nil {
			return nil, fmt.Errorf("PROJECT scope render: %w (record rollback failed: %v)", err, rbErr)
		}
		return nil, fmt.Errorf("PROJECT scope render: %w", err)
	}
	// Converge the record to exactly this render's generation now that the
	// project file actually carries it — the closing half of the
	// stage/converge pair (card t1630 card-review P1).
	if err := persistManagedAllowRecord(projectSettingsPath, full.Allow); err != nil {
		return nil, fmt.Errorf("persist managed-allow record: %w", err)
	}

	return &TierRenderResult{UserPath: userSettingsPath, ProjectPath: projectSettingsPath}, nil
}

// recordProtocolMu serializes the sidecar record protocol — record read, Q5
// classification, stage-ahead, project settings write, converge/rollback —
// across RenderTierPermissions and BuildInto (card t1630 sync-audit F2).
// Concurrent renders that interleave a stale converge past a newer one invert
// the record against the settings file: an entry the policy has since revoked
// then reads as never-generated and persists as user-added. One package-level
// mutex is deliberately coarser than a per-project lock registry — record
// spans are rare file writes, and a single primitive avoids a global
// registry map. BuildInto holds it only for the TargetJSON (project settings)
// target.
var recordProtocolMu sync.Mutex

// Inter-process record-lock parameters (card t1630 sync-audit F2'): the
// lockfile is retried at this interval for at most this budget; on budget
// exhaustion the holder is probed, and a holder that is provably dead — or a
// lockfile whose mtime age exceeds the stale bound (windows has no portable
// liveness probe, so the age bound is its takeover test) — is taken over.
// A LIVE holder beyond the budget surfaces an error: bounded retry, never an
// infinite wedge.
const (
	recordLockRetryInterval = 50 * time.Millisecond
	recordLockRetryBudget   = 5 * time.Second
	recordLockStaleAge      = 30 * time.Second
)

// managedAllowLockPath derives the inter-process lockfile path: beside the
// sidecar record, `<projectRoot>/.moai/state/tool-policy/managed-allow.lock`.
func managedAllowLockPath(projectSettingsPath string) string {
	return filepath.Join(filepath.Dir(managedAllowRecordPath(projectSettingsPath)), "managed-allow.lock")
}

// parseRecordLockPid extracts the recorded holder pid from a lockfile body.
// ok is false for an unparseable body (the age bound then governs takeover).
func parseRecordLockPid(body []byte) (int, bool) {
	for _, line := range strings.Split(string(body), "\n") {
		for _, field := range strings.Fields(line) {
			if !strings.HasPrefix(field, "pid=") {
				continue
			}
			pid, err := strconv.Atoi(strings.TrimPrefix(field, "pid="))
			if err != nil {
				return 0, false
			}
			return pid, true
		}
	}
	return 0, false
}

// managedAllowRecordPath derives the sidecar record path from the project
// settings path: <project>/.claude/settings.json names the project root two
// levels up, so the record lives at
// <projectRoot>/.moai/state/tool-policy/managed-allow.json. Read and write
// sides share this one derivation.
func managedAllowRecordPath(projectSettingsPath string) string {
	root := filepath.Dir(filepath.Dir(projectSettingsPath))
	return filepath.Join(root, ".moai", "state", "tool-policy", "managed-allow.json")
}

// readManagedAllowRecordList returns the sidecar record's last_generated list,
// or nil when the record is absent or unreadable — the same fail-open stance
// the classification takes (an unreadable record classifies every existing
// entry as user-added, which never removes anything).
func readManagedAllowRecordList(projectSettingsPath string) []string {
	body, err := os.ReadFile(managedAllowRecordPath(projectSettingsPath))
	if err != nil {
		return nil
	}
	var rec managedAllowRecord
	if json.Unmarshal(body, &rec) != nil {
		return nil
	}
	return rec.LastGenerated
}

// writeManagedAllowRecord overwrites the sidecar record with entries,
// creating the directory and file when absent (mode-preserving atomic write).
func writeManagedAllowRecord(projectSettingsPath string, entries []string) error {
	if entries == nil {
		entries = []string{}
	}
	out, err := json.MarshalIndent(managedAllowRecord{LastGenerated: entries}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode managed-allow record: %w", err)
	}
	path := managedAllowRecordPath(projectSettingsPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create managed-allow record dir: %w", err)
	}
	if err := atomicfile.Write(path, out, defs.FilePerm); err != nil {
		return fmt.Errorf("write managed-allow record: %w", err)
	}
	return nil
}

// unionStrings merges b into a order-stably, dropping duplicates.
func unionStrings(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// stageManagedAllowRecord records the union of the record's prior content and
// the list a render is about to generate, written BEFORE the project settings
// are mutated. Staging ahead means: new managed entries are known to the
// record before the settings file can gain them, and an entry the settings
// file still holds keeps its managed classification even when this render's
// settings write fails — the next successful render then either re-adds the
// still-generated entries as managed or drops the revoked one. A staging
// failure fails the render before any settings mutation.
func stageManagedAllowRecord(projectSettingsPath string, generated []string) error {
	return writeManagedAllowRecord(projectSettingsPath,
		unionStrings(readManagedAllowRecordList(projectSettingsPath), generated))
}

// persistManagedAllowRecord converges the sidecar record to exactly the
// generation that landed in the project file — the closing half of the
// stage/converge pair. After a fully successful render the record reads
// exactly "the managed allow list last generated"; a converge failure leaves
// the staged superset, which is the safe direction (it can only mark more
// entries as managed).
func persistManagedAllowRecord(projectSettingsPath string, generated []string) error {
	return writeManagedAllowRecord(projectSettingsPath, generated)
}

// rollbackManagedAllowRecord restores the sidecar record to its pre-render
// content after a FAILED settings write (card t1630 sync-audit F3): the
// stage-ahead must not upgrade entries to managed unless the settings write
// that would land them succeeds — otherwise a later policy revision dropping
// the entry deletes a user-authored allow (a tool-initiated removal of a
// user-owned entry, against the decided Q5 rule). A nil prior (no record
// existed before the stage) rolls back to record-absent by removing the
// staged file. It runs inside the record-protocol span, so the restore cannot
// race a concurrent render.
func rollbackManagedAllowRecord(projectSettingsPath string, prior []string) error {
	if prior == nil {
		if err := os.Remove(managedAllowRecordPath(projectSettingsPath)); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return writeManagedAllowRecord(projectSettingsPath, prior)
}

// managedAllowRecord mirrors the sidecar record at
// <projectRoot>/.moai/state/tool-policy/managed-allow.json, which carries the
// managed allow list last generated into the project file
// (SPEC-USER-SETTINGS-PROTECT-001 decision Q5; detection pinned in
// progress.md §G, G-20). Every policy-generation path maintains it:
// RenderTierPermissions stages prior ∪ generated before the PROJECT settings
// mutation and converges to the exact generation once the write lands
// (stageManagedAllowRecord / persistManagedAllowRecord); BuildInto does the
// same for the TargetJSON project-settings target. Across a failed settings
// write the record is rolled back to its pre-render content
// (rollbackManagedAllowRecord) — a failed render leaves no trace. The union
// classification reads the PREVIOUS render's record before the overwrite.
type managedAllowRecord struct {
	LastGenerated []string `json:"last_generated"`
}

// mergeProjectAllow keeps user-added permissions.allow entries across a
// PROJECT-scope regeneration (REQ-005, decided Q5). A user-added entry is an
// existing allow entry absent from the sidecar record's last_generated list;
// with no record every existing entry counts as user-added, so nothing is
// removed. The result is the regenerated list with the user-added entries
// appended in their existing-file order (order-stable set union). A
// last_generated entry the document no longer generates is dropped — only the
// user removes a user-added entry.
//
// @MX:NOTE: [AUTO] Q5 set-union — user-added allow entries survive the managed-block regeneration; a last_generated entry the doc no longer generates is dropped (SPEC-USER-SETTINGS-PROTECT-001)
func mergeProjectAllow(projectSettingsPath string, regenerated []string) []string {
	merged := append([]string{}, regenerated...)
	seen := make(map[string]bool, len(regenerated))
	for _, entry := range regenerated {
		seen[entry] = true
	}

	body, err := os.ReadFile(projectSettingsPath)
	if err != nil {
		return merged
	}
	existing, extrErr := extractPermissions(body)
	if extrErr != nil || existing == nil {
		return merged
	}

	lastGenerated := map[string]bool{}
	for _, entry := range readManagedAllowRecordList(projectSettingsPath) {
		lastGenerated[entry] = true
	}

	for _, entry := range existing.Allow {
		if lastGenerated[entry] {
			// Managed entry the document no longer generates: dropped.
			continue
		}
		if !seen[entry] {
			seen[entry] = true
			merged = append(merged, entry)
		}
	}
	return merged
}

// WriteUserDefaultMode splices ONLY the defaultMode into the USER-scope
// settings.json at userPath, preserving every other region (PATH, hooks, env,
// allow, deny, ask). It is the init-time fallback used when no tool-policy.yaml
// is deployed: the deny/ask arrays already ship in the PROJECT template, so the
// only missing piece is the USER-scope defaultMode that the selected tier maps
// to. The codegen reuses extractPermissions + RenderSettingsJSON (the same
// rendering pipeline RenderTierPermissions uses), so this is NOT a parallel
// writer — it is a scoped entry point into the same splice logic.
//
// If the file does not exist, a minimal permissions skeleton is created so the
// region matcher has a target (mirrors renderIntoFile's create-on-absent path).
// The parent directory is created on demand so the init path can write the
// USER-scope ~/.claude/settings.json even when the directory does not yet exist.
//
// SPEC-USER-SETTINGS-PROTECT-001 (decision Q1, DECIDED): an existing USER-scope
// permissions.defaultMode is KEPT on every path — the tier default is written
// only when the key is absent. A present value means the user owns the setting,
// so the writer returns without touching the file (REQ-002: the bytes stay
// identical when nothing would change). The write path (absent defaultMode)
// preserves the user's allow/ask/deny entries and every unmodelled permissions
// key (REQ-001).
//
// @MX:NOTE: [AUTO] Q1 keep-rule — an existing USER defaultMode is never overwritten; the tier default is written only when the key is absent, and the write path preserves the user's lists (SPEC-USER-SETTINGS-PROTECT-001)
func WriteUserDefaultMode(userPath, defaultMode string) error {
	if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
		return fmt.Errorf("create user settings dir: %w", err)
	}
	if body, readErr := os.ReadFile(userPath); readErr == nil {
		if existing, extrErr := extractPermissions(body); extrErr == nil && existing != nil {
			if existing.DefaultMode != "" {
				// Q1 (SPEC-USER-SETTINGS-PROTECT-001): the user's value is
				// kept; the file is left byte-identical.
				return nil
			}
			// Absent defaultMode: write the tier default while preserving the
			// user's allow/ask/deny and every unmodelled key (REQ-001).
			// extractPermissions parsed the lists from the same body that
			// renderIntoFile re-reads, so the splice keeps them.
			existing.DefaultMode = defaultMode
			return renderIntoFile(userPath, existing)
		}
	} else if !os.IsNotExist(readErr) {
		return fmt.Errorf("read %q: %w", userPath, readErr)
	}
	block := &PermissionsBlock{
		DefaultMode: defaultMode,
		Raw:         map[string]json.RawMessage{},
	}
	return renderIntoFile(userPath, block)
}

// renderIntoFile reads the target settings file, splices block into its
// permissions region, and writes it back. It preserves all non-permissions
// regions (PATH, hooks, env, Go-template directives). If the file does not
// exist, it is created with a minimal skeleton.
func renderIntoFile(path string, block *PermissionsBlock) error {
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create a minimal skeleton so the region matcher has a target.
			// Ensure the parent directory exists so the subsequent WriteFile
			// succeeds (the USER-scope ~/.claude/ dir may not exist yet at
			// init time).
			if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
				return fmt.Errorf("create dir for %q: %w", path, mkErr)
			}
			body = []byte("{\n  \"permissions\": {\n  }\n}")
		} else {
			return fmt.Errorf("read %q: %w", path, err)
		}
	}

	// Preserve any extra Raw keys from the existing block (additionalDirectories
	// etc.) so the codegen does not drop non-list permission settings.
	existing, extrErr := extractPermissions(body)
	if extrErr == nil && existing != nil {
		block.Raw = existing.Raw
	}

	out, err := RenderSettingsJSON(body, block)
	if err != nil {
		return fmt.Errorf("render %q: %w", path, err)
	}
	if err := atomicfile.Write(path, out, defs.FilePerm); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}
