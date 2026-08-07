// Package tiers implements the 4-tier addressable map overlay (BAS Epic M4,
// SPEC-NAVIGATOR-SYNC-003). It emits a single tiers.json OVERLAY at
// .moai/project/navigator/tiers.json that consumers JOIN with M0's
// nav-graph.json on the composite key (entity_type, identifier).
//
// The layer is strictly a CONSUMER of M0/M1 producer outputs (REQ-NS3-016):
// it reads nav-graph.json and the M1 detect impact record; it MUST NOT modify
// internal/navigator/sync/, internal/navigator/detect/, internal/hook/, or
// internal/navigator/astx/. M4 NEVER overwrites nav-graph.json (REQ-NS3-018).
//
// The schema below is the Go projection of the normative JSON example in
// design.md §2. JSON tags are the stable contract: field names are additive
// only (forward-compatible per .claude/rules/moai/workflow/nav-tokens.md).
package tiers

// ContractKind enumerates the Tier 0 contract-declaration kinds
// (REQ-NS3-001). String-typed for additive forward-compat (no custom
// marshaler — a yet-unknown kind degrades to an unvalidated string rather
// than failing to parse).
type ContractKind string

const (
	// ContractKindSchema is a JSON-schema-style declaration validated
	// against an artifact's shape (e.g. nav-graph.json's schema).
	ContractKindSchema ContractKind = "schema"
	// ContractKindAllowlist is a capability-allowlist declaration (e.g.
	// a Tauri allowlist) validated against the declared surface.
	ContractKindAllowlist ContractKind = "allowlist"
	// ContractKindOpenAPI is an OpenAPI specification declaration.
	ContractKindOpenAPI ContractKind = "openapi"
)

// DriftStatus enumerates the Tier 0 contract drift states (REQ-NS3-002).
type DriftStatus string

const (
	// DriftUnknown means the drift check has not yet run for this contract.
	DriftUnknown DriftStatus = "unknown"
	// DriftAligned means the declared contract and implementation agree.
	DriftAligned DriftStatus = "aligned"
	// DriftCollapsed means the declared contract and implementation diverge.
	// A drift finding is fail-open with respect to graph emission
	// (REQ-NS3-002): it logs and (in CI) exits non-zero, but it does NOT
	// block tiers.json emission.
	DriftCollapsed DriftStatus = "drifted"
)

// Layer enumerates the four moai-adk-go module layers used by the Tier 1
// blueprint (REQ-NS3-004). The names are language-neutralized for template
// distribution (the implementing language is Go, but a template-distributed
// user's project may be any of the 16 supported languages).
type Layer string

const (
	LayerPresentation   Layer = "presentation"
	LayerDomain         Layer = "domain"
	LayerInfrastructure Layer = "infrastructure"
	LayerMeasurement    Layer = "measurement"
)

// TierEdgeType enumerates the three NEW edge types the 4-tier layer wires
// into the graph additively (REQ-NS3-007 / REQ-NS3-009 / REQ-NS3-010).
// These do NOT rename or redefine M0's existing dec-edge/spec-edge/sym-edge.
type TierEdgeType string

const (
	// EdgeModule is a blueprint→blueprint edge sourced from a module's
	// depends_on list (REQ-NS3-007).
	EdgeModule TierEdgeType = "module-edge"
	// EdgeOwns is a blueprint→symbol edge joining an authored module to
	// the symbols extracted within it (REQ-NS3-007).
	EdgeOwns TierEdgeType = "owns-edge"
	// EdgeSupersededBy is a decision→decision edge recording the ADR
	// supersede chain (REQ-NS3-009).
	EdgeSupersededBy TierEdgeType = "superseded_by"
)

// Provenance stamps a tier artifact to a git baseline (REQ-NS3-019).
// ExtractCommitSHA is `git rev-parse HEAD`; CapturedAt is the committer date
// of that SHA (`git log -1 --format=%cI`). NO wall-clock timestamp is used,
// so two runs on the same HEAD produce byte-identical output. Carried forward
// from M0's internal/navigator/sync/schema.go Provenance model byte-for-byte.
type Provenance struct {
	ExtractCommitSHA string `json:"extract_commit_sha"`
	CapturedAt       string `json:"captured_at"`
}

// ContractNode is a Tier 0 contract declaration node (REQ-NS3-001 /
// REQ-NS3-003). Each node declares a build-enforced contract surface and
// its current drift status.
type ContractNode struct {
	Identifier       string       `json:"identifier"`
	ContractKind     ContractKind `json:"contract_kind"`
	ContractPath     string       `json:"contract_path"`
	ValidatorCommand string       `json:"validator_command"`
	DriftStatus      DriftStatus  `json:"drift_status"`
}

// BlueprintNode is a Tier 1 authored module entry (REQ-NS3-004 /
// REQ-NS3-007). It is scaffolded from /moai codemaps dependencies.md and
// refined by human or agent (blueprint-first, NOT auto-generate-and-replace).
type BlueprintNode struct {
	Identifier     string   `json:"identifier"`
	DisplayName    string   `json:"display_name"`
	Layer          Layer    `json:"layer"`
	Responsibility string   `json:"responsibility"`
	DependsOn      []string `json:"depends_on"`
	OverviewPath   string   `json:"overview_path"`
}

// DecisionEnrichment is the Tier 2 additive enrichment of an M0 decision
// node (REQ-NS3-008 / REQ-NS3-010). AdrPath is empty when no ADR file exists
// for the identifier (graceful degrade). SupersededBy is empty when current.
type DecisionEnrichment struct {
	Identifier   string `json:"identifier"`
	AdrPath      string `json:"adr_path"`
	SupersededBy string `json:"superseded_by"`
	Supersedes   string `json:"supersedes"`
}

// SymbolRef is one reference (caller) location within the per-symbol
// references list (REQ-NS3-011).
type SymbolRef struct {
	Path string `json:"path"`
	Line int    `json:"line"`
}

// SymbolEnrichment is the Tier 3 additive enrichment of an M0 symbol node
// (REQ-NS3-011 / REQ-NS3-014). The deterministic structure fields
// (Signature, DeclarationPath, DeclarationLine, References) are produced by
// the astx/Go path (REQ-NS3-013); NarrativePath is the LLM narrative slot
// (empty when no narrative exists — 2-tier separable, REQ-NS3-015).
type SymbolEnrichment struct {
	Identifier      string      `json:"identifier"`
	Signature       string      `json:"signature"`
	DeclarationPath string      `json:"declaration_path"`
	DeclarationLine int         `json:"declaration_line"`
	References      []SymbolRef `json:"references"`
	NarrativePath   string      `json:"narrative_path"`
}

// TierEdge is a graph edge wired by the 4-tier layer (REQ-NS3-007 /
// REQ-NS3-009). SourceNode and TargetNode use the `<entity_type>:<identifier>`
// reference form (blueprint:<id>, decision:<id>, symbol:<id>).
type TierEdge struct {
	EdgeType   TierEdgeType `json:"edge_type"`
	SourceNode string       `json:"source_node"`
	TargetNode string       `json:"target_node"`
}

// TiersOverlay is the in-memory shape of the tiers.json overlay artifact
// (REQ-NS3-018). It serializes to the normative JSON shape in design.md §2.
// The overlay is self-contained — it carries its own provenance plus the
// enriched nodes and the new edge types; consumers JOIN with nav-graph.json
// on (entity_type, identifier) and MUST NOT require M0's artifact to carry
// these fields.
type TiersOverlay struct {
	Provenance      Provenance           `json:"provenance"`
	Tier0Contracts  []ContractNode       `json:"tier0_contracts"`
	Tier1Blueprints []BlueprintNode      `json:"tier1_blueprints"`
	Tier2Decisions  []DecisionEnrichment `json:"tier2_decisions"`
	Tier3Symbols    []SymbolEnrichment   `json:"tier3_symbols"`
	TierEdges       []TierEdge           `json:"tier_edges"`
}
