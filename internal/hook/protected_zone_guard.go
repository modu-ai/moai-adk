package hook

// protected_zone_guard.go — the manifest-driven half of the FROZEN-zone guard
// (SPEC-SELF-IMPROVE-PROTECTED-ZONE-001).
//
// The guard applies only to the compiled self-improvement identity set and only
// to the Write/Edit branch of the PreToolUse handler; the Bash half arrives with
// the shell rule in plan.md M3. Order inside the check, per plan.md M2: identity
// gate → normalize → compiled baseline first (so P1–P4 and S1 keep their legacy
// sentinels) → manifest. A present-but-invalid manifest fails closed for the
// identity; an absent one degrades visibly and is recorded in the audit log.
//
// The manifest is opened only after the identity and tool gates (REQ-SIPZ-008):
// a caller outside the identity set, or a tool other than Write/Edit, never
// reaches loadZone. No environment variable, tool-input field, or file the
// identity can write changes a denial (REQ-SIPZ-013) — the guard reads none.

import (
	"encoding/json"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	// SentinelHarnessFrozenProtectedZone is emitted when a self-improvement
	// identity attempts to modify a path the protected-zone manifest declares
	// and no compiled baseline entry covers. A baseline match keeps its legacy
	// sentinel and reason byte for byte and carries no routing field
	// (REQ-SIPZ-011, REQ-SIPZ-016).
	SentinelHarnessFrozenProtectedZone = "HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION"

	// zoneReasonMax bounds a whole denial reason in bytes; the trailing path
	// field is the only one truncated to fit (REQ-SIPZ-011).
	zoneReasonMax = 240

	// zoneBaselineCategory labels an audit row whose denial came from the
	// compiled baseline floor rather than a manifest entry.
	zoneBaselineCategory = "baseline"

	// zoneAuditRel is the project-relative path of the guard's audit log
	// (REQ-SIPZ-014).
	zoneAuditRel = ".moai/logs/protected-zone-audit.jsonl"
)

// zoneIdentities is the compiled self-improvement identity set (REQ-SIPZ-008).
// Widening it is a reviewed code change, not a manifest entry (spec §C.3).
var zoneIdentities = map[string]bool{harnessLearnerIdentity: true}

// isZoneIdentity reports whether agentID is in the compiled self-improvement
// identity set (REQ-SIPZ-008).
func isZoneIdentity(agentID string) bool {
	return zoneIdentities[agentID]
}

// zoneAuditRow is one audit-log line (REQ-SIPZ-014). An empty ManifestState on
// a baseline denial means the manifest was not consulted: the compiled floor
// decided before the manifest step ran.
type zoneAuditRow struct {
	TS            string `json:"ts"`
	Identity      string `json:"identity"`
	Tool          string `json:"tool"`
	Path          string `json:"path"`
	Category      string `json:"category"`
	Decision      string `json:"decision"`
	ManifestState string `json:"manifest_state"`
}

// zoneDenyReason formats a routed denial reason: the sentinel, the identity,
// one field=value pair, the two routing fields, then the project-relative path
// as the only field that may be truncated — never mid-rune, all within
// zoneReasonMax bytes (REQ-SIPZ-011). The path is project-relative: a reason
// must not leak an absolute path or manifest content.
//
// @MX:NOTE: [AUTO] the 240-byte bound and the fixed field order are a wire
// contract — the orchestrator pattern-matches the routing fields.
// @MX:SPEC:SPEC-SELF-IMPROVE-PROTECTED-ZONE-001
func zoneDenyReason(identity, field, value, relPath string) string {
	prefix := SentinelHarnessFrozenProtectedZone + ": " + identity + " " + field + "=" + value +
		" route=human next=return-blocker-report path="
	p := relPath
	if budget := zoneReasonMax - len(prefix); len(p) > budget {
		if budget < 0 {
			budget = 0
		}
		p = p[:budget]
		for len(p) > 0 && !utf8.ValidString(p) {
			p = p[:len(p)-1]
		}
	}
	return prefix + p
}

// zoneAppendAudit appends one audit row under the project root, creating the
// logs directory when needed. A failed append is returned, never swallowed
// here: the decision it accompanies never changes (REQ-SIPZ-014).
func zoneAppendAudit(root string, row zoneAuditRow) error {
	if row.TS == "" {
		row.TS = time.Now().UTC().Format(time.RFC3339)
	}
	dir := filepath.Join(root, filepath.FromSlash(path.Dir(zoneAuditRel)))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(root, filepath.FromSlash(zoneAuditRel)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	line, err := json.Marshal(row)
	if err != nil {
		return err
	}
	line = append(line, '\n')
	_, err = f.Write(line)
	return err
}

// recordZoneAudit appends one audit row, failing open with a stderr notice: a
// failed append never changes the decision it accompanies (REQ-SIPZ-014).
func (h *preToolHandler) recordZoneAudit(root string, row zoneAuditRow) {
	if err := zoneAppendAudit(root, row); err != nil {
		slog.Warn("protected zone audit append failed", "error", err)
	}
}

// loadZone reads both manifest files through the handler's loader seam. Tests
// replace zoneLoader to count reads; production leaves it nil.
func (h *preToolHandler) loadZone(root string) config.ProtectedZoneLoad {
	if h.zoneLoader != nil {
		return h.zoneLoader(root)
	}
	return config.LoadProtectedZone(root)
}

// checkProtectedZone decides one identity Write/Edit call against the zone.
// The path was already extracted from the tool input; rawPath may be relative
// (resolved the way the existing file-access check resolves it, against the
// hook process cwd) or absolute. Returns (sentinel, deny-reason); ("", "")
// allows the call through to the next guard.
//
// @MX:SPEC:SPEC-SELF-IMPROVE-PROTECTED-ZONE-001
func (h *preToolHandler) checkProtectedZone(agentID, toolName, rawPath string) (string, string) {
	root := h.projectRoot()
	if root == "" {
		return "", ""
	}
	forms := resolveZoneTarget(root, rawPath)

	// Compiled baseline first, on every relative form the normalization
	// produced: a baseline match keeps its legacy sentinel and reason byte for
	// byte and carries no routing field (REQ-SIPZ-011, REQ-SIPZ-016). Exact
	// case first so canonical inputs keep their byte-for-byte reason, then the
	// case-folded form so a letter-case variant earns the same legacy sentinel
	// (AC-SIPZ-002 P4). The legacy helper re-checks the identity, which today
	// is the whole set.
	for _, form := range forms {
		sentinel, reason := h.checkHarnessFrozenZone(agentID, form.Display)
		if sentinel == "" {
			sentinel, reason = h.checkHarnessFrozenZone(agentID, form.Folded)
		}
		if sentinel != "" {
			h.recordZoneAudit(root, zoneAuditRow{
				Identity: agentID, Tool: toolName, Path: form.Display,
				Category: zoneBaselineCategory, Decision: "deny",
			})
			return sentinel, reason
		}
	}

	load := h.loadZone(root)
	switch load.State {
	case config.ZoneStateInvalid:
		// Fail closed for the identity: the zone's extent is unknown, so
		// every identity Write/Edit is denied and the failing file is named
		// (REQ-SIPZ-009).
		reason := zoneDenyReason(agentID, "manifest", "invalid", load.InvalidFile)
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: toolName, Path: load.InvalidFile,
			Decision: "deny", ManifestState: config.ZoneStateInvalid,
		})
		return SentinelHarnessFrozenProtectedZone, reason
	case config.ZoneStateAbsent:
		// Degrade visibly: the compiled floor above is all that held; the
		// evaluation is recorded so the degraded state has a trail
		// (REQ-SIPZ-010). A healthy allow below appends nothing.
		rel := ""
		if len(forms) > 0 {
			rel = forms[0].Display
		}
		h.recordZoneAudit(root, zoneAuditRow{
			Identity: agentID, Tool: toolName, Path: rel,
			Decision: "allow", ManifestState: config.ZoneStateAbsent,
		})
		return "", ""
	}
	for _, form := range forms {
		for i := range load.Zone.Entries {
			entry := load.Zone.Entries[i]
			if entry.Match(form.Folded) {
				reason := zoneDenyReason(agentID, "category", entry.Category, form.Display)
				h.recordZoneAudit(root, zoneAuditRow{
					Identity: agentID, Tool: toolName, Path: form.Display,
					Category: entry.Category, Decision: "deny", ManifestState: config.ZoneStateOK,
				})
				return SentinelHarnessFrozenProtectedZone, reason
			}
		}
	}
	return "", ""
}
