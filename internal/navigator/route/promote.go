// Package route implements the M2 Route layer of the BAS Falconer loop
// (SPEC-NAVIGATOR-SYNC-004): promote audit missing/orphan findings + M1
// detect findings into actionable work items, each owner-bound to a code
// path or design-doc path (never a person — REQ-NS4-004 falconer binding).
//
// The package CONSUMES the M0 graph types from internal/navigator/sync
// (REQ-NS4-005 bridge-not-absorb): it imports sync.Graph / sync.Edge /
// sync.Node read-only and NEVER mutates the M0 producer surface. It also
// consumes the audit-report.json schema (002) and the M1 detect JSONL
// schema read-only.
package route

import (
	"fmt"
	"sort"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// --- Source kinds + confidence taxonomy (REQ-NS4-003/004) ---

// SourceKind enumerates the three work-item source kinds (REQ-NS4-003).
type SourceKind string

const (
	SourceAuditMissing SourceKind = "audit-missing"
	SourceAuditOrphan  SourceKind = "audit-orphan"
	SourceDetect       SourceKind = "detect"
)

// Confidence enumerates the three owner-resolution confidence levels
// (REQ-NS4-004, plan.md §C.2).
//
//   - high   — owner resolved from a direct path field (implementation_path,
//     changed_path). No graph traversal, no heuristic.
//   - medium — owner resolved via M0 graph traversal (@NAV:SYM symbol →
//     declaration code path). One hop through the graph.
//   - low    — owner fell back to a doc/spec path with no code binding.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// --- Consumed input schemas (read-only, 002 audit + M1 detect) ---

// AuditReport is the parsed audit-report.json emitted by 002
// (.claude/skills/moai-workflow-project/scripts/navigator-audit.sh:553-591).
// The Route layer reads this read-only (REQ-NS4-002a).
type AuditReport struct {
	AuditAt     string         `json:"audit_at"`
	AuditCommit string         `json:"audit_commit"`
	Inputs      AuditInputs    `json:"inputs"`
	Missing     []MissingEntry `json:"missing"`
	Orphan      []OrphanEntry  `json:"orphan"`
	Matched     []MatchedEntry `json:"matched"`
}

// AuditInputs is the inputs block of audit-report.json.
type AuditInputs struct {
	DesignDocs    []string `json:"design_docs"`
	CapabilityMap string   `json:"capability_map"`
	OverrideFile  any      `json:"override_file"`
}

// MissingEntry is one element of audit-report.json's missing[] array — a
// design feature with no matching SPEC. The Source.File + Source.HeadingPath
// is the owner anchor for REQ-NS4-004's missing-resolution path.
type MissingEntry struct {
	DesignName   string      `json:"design_name"`
	Source       AuditSource `json:"source"`
	ClosestMatch *string     `json:"closest_match"`
}

// AuditSource is the design-doc location inside a MissingEntry.
type AuditSource struct {
	File        string `json:"file"`
	HeadingPath string `json:"heading_path"`
}

// OrphanEntry is one element of audit-report.json's orphan[] array — a SPEC
// with no matching design feature. The ImplementationPath (002 emits it at
// navigator-audit.sh:575) is the load-bearing owner anchor for
// REQ-NS4-004's orphan-resolution path.
type OrphanEntry struct {
	SpecID             string `json:"spec_id"`
	Title              string `json:"title"`
	ImplementationPath string `json:"implementation_path"`
}

// MatchedEntry is one element of audit-report.json's matched[] array. Matched
// entries are NOT findings (they are already-resolved) and are excluded from
// work-item promotion.
type MatchedEntry struct {
	DesignName string `json:"design_name"`
	SpecID     string `json:"spec_id"`
	MatchBasis string `json:"match_basis"`
}

// DetectRecord is the parsed M1 detect JSONL impact record
// (internal/hook/navigator_detect.go impactRecord). The Route layer reads
// ALL *.jsonl files across sessions, deduplicating by ChangedPath (latest
// ChangedAt wins) — REQ-NS4-002b.
type DetectRecord struct {
	ChangedPath   string         `json:"changed_path"`
	ChangedAt     string         `json:"changed_at"`
	AffectedNodes []DetectNode   `json:"affected_nodes"`
	AffectedEdges []navsync.Edge `json:"affected_edges"`
}

// DetectNode is the per-node entry inside a DetectRecord's affected_nodes.
type DetectNode struct {
	EntityType navsync.EntityType `json:"entity_type"`
	Identifier string             `json:"identifier"`
}

// --- Work item (REQ-NS4-003: five fields, no more, no less) ---

// WorkItem is a promoted finding with an owner bound to a code/doc path.
// The five fields are load-bearing: source_kind identifies the input chain,
// source_entry preserves the original entry verbatim (provenance), owner_path
// is the falconer binding (always a path, never a person — REQ-NS4-004),
// action is the one-line closing directive, and confidence reflects how
// directly the owner was resolved.
type WorkItem struct {
	SourceKind  SourceKind `json:"source_kind"`
	SourceEntry any        `json:"source_entry"`
	OwnerPath   string     `json:"owner_path"`
	Action      string     `json:"action"`
	Confidence  Confidence `json:"confidence"`
}

// Promote transforms parsed audit + detect + graph inputs into a sorted,
// deduplicated work-item slice (REQ-NS4-003). Pure function: no I/O, no
// side effects, deterministic. The caller owns input loading (including
// detect-row deduplication by changed_path) and fail-open policy.
func Promote(audit *AuditReport, detectRows []DetectRecord, graph *navsync.Graph, projectRoot string) []WorkItem {
	var items []WorkItem

	if audit != nil {
		for _, entry := range audit.Orphan {
			owner, conf := resolveOrphanOwner(entry, projectRoot)
			items = append(items, WorkItem{
				SourceKind:  SourceAuditOrphan,
				SourceEntry: entry,
				OwnerPath:   owner,
				Action:      actionOrphan,
				Confidence:  conf,
			})
		}
		for _, entry := range audit.Missing {
			owner, conf := resolveMissingOwner(entry, graph, projectRoot)
			items = append(items, WorkItem{
				SourceKind:  SourceAuditMissing,
				SourceEntry: entry,
				OwnerPath:   owner,
				Action:      actionMissing,
				Confidence:  conf,
			})
		}
	}

	for _, record := range detectRows {
		owner, conf := resolveDetectOwner(record, projectRoot)
		items = append(items, WorkItem{
			SourceKind:  SourceDetect,
			SourceEntry: record,
			OwnerPath:   owner,
			Action:      actionDetect,
			Confidence:  conf,
		})
	}

	return dedupAndSort(items)
}

// identifierOf returns the natural identifier of a work item's source entry,
// used for deduplication and sorting (REQ-NS4-003b). The identifier is
// source-kind-specific: spec_id for orphans, design_name for missing,
// changed_path for detect.
func identifierOf(item WorkItem) string {
	switch item.SourceKind {
	case SourceAuditOrphan:
		if e, ok := item.SourceEntry.(OrphanEntry); ok {
			return e.SpecID
		}
	case SourceAuditMissing:
		if e, ok := item.SourceEntry.(MissingEntry); ok {
			return e.DesignName
		}
	case SourceDetect:
		if e, ok := item.SourceEntry.(DetectRecord); ok {
			return e.ChangedPath
		}
	}
	return ""
}

// dedupAndSort applies cross-source dedup (audit wins over detect when
// identifiers collide), within-source-kind dedup (same owner_path +
// identifier), and deterministic sort by (source_kind, owner_path,
// identifier) so two runs on the same inputs produce byte-identical output
// (REQ-NS4-003b idempotence).
func dedupAndSort(items []WorkItem) []WorkItem {
	if len(items) == 0 {
		return []WorkItem{}
	}

	// Cross-source dedup: if an audit item and a detect item share the same
	// identifier, the audit source wins — audit is the authoritative roll-up,
	// detect is the real-time supplement (REQ-NS4-003b).
	auditIDs := make(map[string]bool, len(items))
	for _, item := range items {
		if item.SourceKind == SourceAuditMissing || item.SourceKind == SourceAuditOrphan {
			auditIDs[identifierOf(item)] = true
		}
	}
	filtered := make([]WorkItem, 0, len(items))
	for _, item := range items {
		if item.SourceKind == SourceDetect && auditIDs[identifierOf(item)] {
			continue
		}
		filtered = append(filtered, item)
	}

	// Within-source-kind dedup by (source_kind, owner_path, identifier).
	seen := make(map[string]bool, len(filtered))
	deduped := make([]WorkItem, 0, len(filtered))
	for _, item := range filtered {
		key := fmt.Sprintf("%s|%s|%s", item.SourceKind, item.OwnerPath, identifierOf(item))
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, item)
	}

	// Deterministic sort by (source_kind, owner_path, identifier).
	sort.SliceStable(deduped, func(i, j int) bool {
		if deduped[i].SourceKind != deduped[j].SourceKind {
			return deduped[i].SourceKind < deduped[j].SourceKind
		}
		if deduped[i].OwnerPath != deduped[j].OwnerPath {
			return deduped[i].OwnerPath < deduped[j].OwnerPath
		}
		return identifierOf(deduped[i]) < identifierOf(deduped[j])
	})

	return deduped
}
