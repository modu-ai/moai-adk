// Package sync implements the BAS (Blueprint-Anchored Synchronization)
// integration layer that joins the three existing Navigator chains (001 regen,
// 002 audit, 003 enrich) into a single addressable graph via three SSOT
// binding-token families (REQ-NS-001).
//
// The integration layer sits ON TOP of the three chains: it consumes their
// outputs and MUST NOT modify them (REQ-NS-012). It also MUST NOT modify
// internal/mx/ (REQ-NS-005) — it consumes the existing SpecAssociator output
// via an additive bridge.
package sync

// TokenFamily enumerates the three SSOT binding-token families recognized by
// the integration layer (REQ-NS-001).
type TokenFamily string

const (
	// FamilyNavDec is the `@NAV:DEC-<id>` design-decision token.
	FamilyNavDec TokenFamily = "NAV:DEC"
	// FamilyMxSpec is the `@MX:SPEC:<id>` code→SPEC back-pointer. This family
	// is NOT re-scanned by this layer; it is sourced from the existing
	// internal/mx.SpecAssociator output (REQ-NS-005).
	FamilyMxSpec TokenFamily = "MX:SPEC"
	// FamilyNavSym is the `@NAV:SYM:<symbol>` code-symbol token.
	FamilyNavSym TokenFamily = "NAV:SYM"
)

// EntityType enumerates the three graph-node entity types (REQ-NS-007).
type EntityType string

const (
	EntityDecision EntityType = "decision"
	EntitySpec     EntityType = "spec"
	EntitySymbol   EntityType = "symbol"
)

// EdgeType enumerates the three graph-edge types, one per token family
// (REQ-NS-008).
type EdgeType string

const (
	EdgeDec  EdgeType = "dec-edge"
	EdgeSpec EdgeType = "spec-edge"
	EdgeSym  EdgeType = "sym-edge"
)

// BindingRecord is the per-occurrence scanner output for one binding token
// (REQ-NS-002). Every record carries exactly the five fields below; scanners
// never emit a partial record (malformed occurrences are skipped with a
// diagnostic per REQ-NS-017).
type BindingRecord struct {
	TokenFamily TokenFamily `json:"token_family"`
	Identifier  string      `json:"identifier"`
	SourcePath  string      `json:"source_path"`
	LineNumber  int         `json:"line_number"`
	CommitSHA   string      `json:"commit_sha"`
}

// Provenance stamps a graph to a git baseline (REQ-NS-009). It carries
// ExtractCommitSHA (`git rev-parse HEAD`) and CapturedAt (the committer date
// of that SHA, `git log -1 --format=%cI`). No wall-clock timestamp is used,
// so two runs on the same HEAD produce byte-identical output. Carried forward
// byte-for-byte from 003's `internal/navigator/astx/enrich.go:21` model.
type Provenance struct {
	ExtractCommitSHA string `json:"extract_commit_sha"`
	CapturedAt       string `json:"captured_at"`
}

// Node is a graph node — one per unique entity (REQ-NS-007). The composite
// primary key is (entity_type, identifier); edges reference nodes via
// `<entity_type>:<identifier>`.
type Node struct {
	EntityType  EntityType `json:"entity_type"`
	Identifier  string     `json:"identifier"`
	DisplayName string     `json:"display_name"`
}

// Edge is a graph edge typed by one of the three token families (REQ-NS-008).
type Edge struct {
	EdgeType   EdgeType `json:"edge_type"`
	SourceNode string   `json:"source_node"`
	TargetNode string   `json:"target_node"`
	SourcePath string   `json:"source_path"`
	LineNumber int      `json:"line_number"`
}

// Graph is the single emitted artifact's in-memory shape (REQ-NS-006). It
// serializes to `{ "provenance": {...}, "nodes": [...], "edges": [...] }`
// with deterministic key/element ordering (sorted) for byte-identical re-runs.
type Graph struct {
	Provenance Provenance `json:"provenance"`
	Nodes      []Node     `json:"nodes"`
	Edges      []Edge     `json:"edges"`
}

// nodeKey returns the `<entity_type>:<identifier>` reference form used by
// Edge.SourceNode / Edge.TargetNode.
func nodeKey(t EntityType, id string) string {
	return string(t) + ":" + id
}
