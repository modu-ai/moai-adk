//! Ontology node/edge schema (C4 #2180) — the JSON-canonical record types with
//! `basis` (why the link exists) and `verification` (review state) kept as
//! separate fields, and the tombstone deletion record.
//!
//! Evidence base: `docs/design-review/appendix/dr07-ontology-table.md` (the
//! merged DR-07 appendix table). Every rule below cites its row.
//!
//! HOLD items from the appendix's "unverifiable / conflicts" section
//! (`docs/design-review/appendix/dr07-ontology-table.md`, from L121) — carried, not guessed:
//! 1. The leader instruction "8 node kinds" (the exact list) is unverifiable
//!    against [TO] 12 rows / [CB] L741 7 kinds — node `kind` is therefore
//!    validated structurally (non-empty identifier) until DR-07 supplies the
//!    list (appendix unverifiable row 1).
//! 2. [TO] mermaid `Other symbol` self-reference has no explicit node
//!    definition — not modeled (appendix unverifiable row 2).
//! 3. [CB] L2491 `included_in_context` has no supplied definition — excluded
//!    from the edge type list (appendix unverifiable row 3).

use serde::{Deserialize, Serialize};
use serde_json::Value;
use sha2::{Digest, Sha256};
use std::collections::BTreeMap;

type Result<T> = std::result::Result<T, String>;

/// `basis` — why the link exists ([TO] L392). Separate from verification:
/// explicit links are author intent; static_resolved/inferred/runtime_observed
/// are derived or observed, never reviewed intent.
pub const BASIS_VALUES: [&str; 4] = [
    "explicit",
    "static_resolved",
    "inferred",
    "runtime_observed",
];

/// `verification` — the review state of a link ([TO] L392). Orthogonal to
/// basis: an inferred link can be checked; an explicit link can be
/// contradicted. REQ-C4-01: the two axes never share one field.
pub const VERIFICATION_VALUES: [&str; 4] = ["unreviewed", "checked", "contradicted", "stale"];

/// The [TO] L375 minimal relation set — 17 types, the core edge vocabulary.
pub const CORE_EDGE_TYPES: [&str; 17] = [
    "contains",
    "declares",
    "imports",
    "calls",
    "annotated_by",
    "references_requirement",
    "implements",
    "verifies",
    "tracks",
    "changes",
    "generated_from",
    "supports",
    "contradicts",
    "supersedes",
    "derived_from",
    "authorized_by",
    "archived_as",
];

/// The [CB]-only relations measured in the appendix (L752-810, L2413-2433)
/// that the [TO] minimal set does not carry. `included_in_context` ([CB]
/// L2491) is excluded — its definition is outside the supplied range (HOLD 3).
pub const CB_ONLY_EDGE_TYPES: [&str; 10] = [
    "targets_host",
    "governed_by",
    "revision_of",
    "generated_by",
    "documents",
    "observed_execution",
    "reports_on",
    "mentions_issue",
    "restored_from",
    "conflicts_with",
];

/// Deletion record ([TO] L394-396): the current snapshot marks removal with a
/// tombstone or `valid_to`; immutable revisions that earlier sources and
/// evidence reference are preserved. A tombstone never deletes the record.
#[derive(Serialize, Deserialize, PartialEq, Debug, Clone)]
pub struct Tombstone {
    pub reason: String,
    /// RFC3339 observation time of the removal.
    pub observed_at: String,
}

#[derive(Serialize, Deserialize, PartialEq, Debug, Clone)]
pub struct Extractor {
    pub name: String,
    /// Pinned version ([TO] L383: extractor `{name, version: pinned}`).
    pub version: String,
}

#[derive(Serialize, Deserialize, PartialEq, Debug, Clone)]
pub struct OntologyNode {
    /// Logical identity ([TO] L354-369, per-kind identity rules).
    pub id: String,
    /// Kind — structurally validated only (HOLD 1: the "8 node kinds" list is
    /// unverifiable until DR-07 supplies it; measured vocabularies live in the
    /// appendix, not in this schema's enforced set).
    pub kind: String,
    /// Immutable revision digest of the node content.
    pub revision: String,
    #[serde(default, skip_serializing_if = "BTreeMap::is_empty")]
    pub properties: BTreeMap<String, String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub tombstone: Option<Tombstone>,
}

#[derive(Serialize, Deserialize, PartialEq, Debug, Clone)]
pub struct OntologyEdge {
    /// Deterministic hash of (type, endpoints, evidence) ([TO] L377).
    pub edge_id: String,
    /// One of CORE_EDGE_TYPES or CB_ONLY_EDGE_TYPES.
    pub kind: String,
    pub source_id: String,
    pub target_id: String,
    /// basis value — separate field (REQ-C4-01).
    pub basis: String,
    /// verification value — separate field (REQ-C4-01).
    pub verification: String,
    /// Evidence refs shaped `file-uri#span@content-hash` ([TO] L381).
    pub evidence_refs: Vec<String>,
    pub extractor: Extractor,
    pub snapshot_id: String,
    /// RFC3339 ([TO] L385).
    pub observed_at: String,
    /// Revision-or-time ([TO] L386).
    pub valid_from: String,
    /// null = currently valid ([TO] L387).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub valid_to: Option<String>,
    /// Categorical uncertainty, e.g. `semantic-match-not-proven` ([TO] L390).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub uncertainty: Option<String>,
}

/// Computes the deterministic `edge_id`: hash of type·endpoints·evidence
/// ([TO] L377). Evidence refs are sorted so ref order never changes identity.
pub fn edge_id(kind: &str, source_id: &str, target_id: &str, evidence_refs: &[String]) -> String {
    let mut refs: Vec<&str> = evidence_refs.iter().map(|s| s.as_str()).collect();
    refs.sort();
    let mut h = Sha256::new();
    h.update(b"moai-ontology-edge-v1\0");
    h.update((kind.len() as u64).to_be_bytes());
    h.update(kind.as_bytes());
    h.update((source_id.len() as u64).to_be_bytes());
    h.update(source_id.as_bytes());
    h.update((target_id.len() as u64).to_be_bytes());
    h.update(target_id.as_bytes());
    for r in refs {
        h.update((r.len() as u64).to_be_bytes());
        h.update(r.as_bytes());
    }
    format!("edge-{:x}", h.finalize())
}

/// Minimal RFC3339 check: `YYYY-MM-DDTHH:MM:SS(.frac)?(Z|±HH:MM)` with field
/// ranges verified. This is a format gate for the schema, not a calendar
/// library — month-day overflow beyond 31 is rejected here, not resolved.
pub fn is_rfc3339(s: &str) -> bool {
    let bytes = s.as_bytes();
    if bytes.len() < 20 {
        return false;
    }
    let digits = |r: std::ops::Range<usize>| bytes[r].iter().all(|b| b.is_ascii_digit());
    if !digits(0..4) || bytes[4] != b'-' || !digits(5..7) || bytes[7] != b'-' || !digits(8..10) {
        return false;
    }
    if bytes[10] != b'T' && bytes[10] != b't' {
        return false;
    }
    if !digits(11..13)
        || bytes[13] != b':'
        || !digits(14..16)
        || bytes[16] != b':'
        || !digits(17..19)
    {
        return false;
    }
    let month: u32 = s[5..7].parse().unwrap_or(0);
    let day: u32 = s[8..10].parse().unwrap_or(0);
    let hour: u32 = s[11..13].parse().unwrap_or(99);
    let minute: u32 = s[14..16].parse().unwrap_or(99);
    let second: u32 = s[17..19].parse().unwrap_or(99);
    if !(1..=12).contains(&month) || !(1..=31).contains(&day) {
        return false;
    }
    if hour > 23 || minute > 59 || second > 60 {
        return false;
    }
    let mut i = 19;
    if i < bytes.len() && bytes[i] == b'.' {
        i += 1;
        let start = i;
        while i < bytes.len() && bytes[i].is_ascii_digit() {
            i += 1;
        }
        if i == start {
            return false;
        }
    }
    match bytes.get(i) {
        Some(b'Z') | Some(b'z') => i + 1 == bytes.len(),
        Some(b'+') | Some(b'-') => {
            let rest = &s[i + 1..];
            rest.len() == 5
                && rest.as_bytes()[0..2].iter().all(|b| b.is_ascii_digit())
                && rest.as_bytes()[2] == b':'
                && rest.as_bytes()[3..5].iter().all(|b| b.is_ascii_digit())
        }
        _ => false,
    }
}

fn nonempty(s: &str, field: &str) -> Result<()> {
    if s.trim().is_empty() {
        return Err(format!("{field} must be non-empty"));
    }
    Ok(())
}

/// Valid values check against a documented vocabulary.
fn in_set(v: &str, set: &[&str], field: &str) -> Result<()> {
    if set.contains(&v) {
        Ok(())
    } else {
        Err(format!(
            "{field} {v:?} not in the documented vocabulary (allowed: {})",
            set.join(", ")
        ))
    }
}

fn validate_evidence_ref(r: &str) -> Result<()> {
    // `file-uri#span@content-hash` ([TO] L381): non-empty file part, non-empty
    // span between # and @, non-empty hash after @.
    let (file, rest) = r.split_once('#').ok_or_else(|| {
        format!("evidence ref {r:?} is not file-uri#span@content-hash (missing #)")
    })?;
    let (span, hash) = rest.split_once('@').ok_or_else(|| {
        format!("evidence ref {r:?} is not file-uri#span@content-hash (missing @)")
    })?;
    for (part, name) in [(file, "file-uri"), (span, "span"), (hash, "content-hash")] {
        if part.trim().is_empty() {
            return Err(format!("evidence ref {r:?} has an empty {name} part"));
        }
    }
    Ok(())
}

fn validate_tombstone(t: &Tombstone) -> Result<()> {
    nonempty(&t.reason, "tombstone.reason")?;
    if !is_rfc3339(&t.observed_at) {
        return Err(format!(
            "tombstone.observed_at {:?} is not RFC3339",
            t.observed_at
        ));
    }
    Ok(())
}

/// Validates a node record: structural rules only (HOLD 1 keeps `kind` out of
/// any enforced vocabulary until the DR-07 node list is supplied).
pub fn validate_node(node: &OntologyNode) -> Result<()> {
    nonempty(&node.id, "node.id")?;
    nonempty(&node.kind, "node.kind")?;
    nonempty(&node.revision, "node.revision")?;
    if let Some(t) = &node.tombstone {
        validate_tombstone(t)?;
    }
    Ok(())
}

/// Validates an edge record. Basis and verification are checked against their
/// own vocabularies in their own fields (REQ-C4-01).
pub fn validate_edge(edge: &OntologyEdge) -> Result<()> {
    // kind ∈ core ∪ CB-only; included_in_context is excluded (HOLD 3).
    if !CORE_EDGE_TYPES.contains(&edge.kind.as_str())
        && !CB_ONLY_EDGE_TYPES.contains(&edge.kind.as_str())
    {
        return Err(format!(
            "edge.kind {:?} not in the documented edge vocabulary (core: {}; CB-only: {})",
            edge.kind,
            CORE_EDGE_TYPES.join(", "),
            CB_ONLY_EDGE_TYPES.join(", ")
        ));
    }
    nonempty(&edge.source_id, "edge.source_id")?;
    nonempty(&edge.target_id, "edge.target_id")?;
    in_set(&edge.basis, &BASIS_VALUES, "edge.basis")?;
    in_set(
        &edge.verification,
        &VERIFICATION_VALUES,
        "edge.verification",
    )?;
    if edge.evidence_refs.is_empty() {
        return Err("edge.evidence_refs must carry at least one evidence ref".into());
    }
    for r in &edge.evidence_refs {
        validate_evidence_ref(r)?;
    }
    nonempty(&edge.extractor.name, "edge.extractor.name")?;
    nonempty(&edge.extractor.version, "edge.extractor.version")?;
    nonempty(&edge.snapshot_id, "edge.snapshot_id")?;
    if !is_rfc3339(&edge.observed_at) {
        return Err(format!(
            "edge.observed_at {:?} is not RFC3339",
            edge.observed_at
        ));
    }
    nonempty(&edge.valid_from, "edge.valid_from")?;
    if let Some(vt) = &edge.valid_to {
        nonempty(vt, "edge.valid_to")?;
    }
    let recomputed = edge_id(
        &edge.kind,
        &edge.source_id,
        &edge.target_id,
        &edge.evidence_refs,
    );
    if recomputed != edge.edge_id {
        return Err(format!(
            "edge.edge_id does not match the recomputed hash of (type, endpoints, evidence): expected {recomputed}"
        ));
    }
    Ok(())
}

/// Validates a JSON document carrying `{"nodes": [...], "edges": [...]}` (the
/// JSON canonical form — no TOON). Returns the counts on success; the first
/// invalid record fails the whole document with a naming message.
pub fn validate_document(doc: &Value) -> Result<(usize, usize)> {
    let nodes = doc
        .get("nodes")
        .and_then(|v| v.as_array())
        .ok_or("document must carry a nodes array")?;
    let edges = doc
        .get("edges")
        .and_then(|v| v.as_array())
        .ok_or("document must carry an edges array")?;
    for (i, n) in nodes.iter().enumerate() {
        let node: OntologyNode = serde_json::from_value(n.clone())
            .map_err(|e| format!("nodes[{i}] is not a valid node record: {e}"))?;
        validate_node(&node).map_err(|e| format!("nodes[{i}]: {e}"))?;
    }
    for (i, e) in edges.iter().enumerate() {
        let edge: OntologyEdge = serde_json::from_value(e.clone())
            .map_err(|e| format!("edges[{i}] is not a valid edge record: {e}"))?;
        validate_edge(&edge).map_err(|e| format!("edges[{i}]: {e}"))?;
    }
    Ok((nodes.len(), edges.len()))
}

/// Marks a node deleted for the current snapshot: sets the tombstone and, when
/// the caller recorded it, `valid_to` — without dropping the record ([TO]
/// L394-396: removal is marked, immutable revisions stay preserved).
pub fn tombstone_node(node: &mut OntologyNode, reason: &str, observed_at: &str) -> Result<()> {
    let t = Tombstone {
        reason: reason.to_string(),
        observed_at: observed_at.to_string(),
    };
    validate_tombstone(&t)?;
    node.tombstone = Some(t);
    Ok(())
}

#[cfg(test)]
mod unit {
    use super::*;

    fn ref_list() -> Vec<String> {
        vec!["src/lib.rs#L10-L20@abc123".into()]
    }

    fn base_edge() -> OntologyEdge {
        let refs = ref_list();
        OntologyEdge {
            edge_id: edge_id("calls", "s1", "t1", &refs),
            kind: "calls".into(),
            source_id: "s1".into(),
            target_id: "t1".into(),
            basis: "static_resolved".into(),
            verification: "checked".into(),
            evidence_refs: refs,
            extractor: Extractor {
                name: "atlas".into(),
                version: "1.0.0".into(),
            },
            snapshot_id: "snap-1".into(),
            observed_at: "2026-11-09T00:00:00Z".into(),
            valid_from: "2026-11-09T00:00:00Z".into(),
            valid_to: None,
            uncertainty: None,
        }
    }

    #[test]
    fn edge_id_is_order_insensitive_and_deterministic() {
        let a = edge_id("calls", "s", "t", &["p#sp@h".into(), "q#sp2@h2".into()]);
        let b = edge_id("calls", "s", "t", &["q#sp2@h2".into(), "p#sp@h".into()]);
        assert_eq!(a, b, "evidence ref order must not change identity");
        assert_ne!(a, edge_id("calls", "s", "t", &["p#sp@h".into()]));
    }

    #[test]
    fn rfc3339_accepts_offsets_and_fractions_rejects_shapes() {
        for good in [
            "2026-11-09T00:00:00Z",
            "2026-11-09t00:00:00.123456z",
            "2026-11-09T12:34:56+09:00",
            "2026-11-09T12:34:56-05:30",
        ] {
            assert!(is_rfc3339(good), "{good} should parse");
        }
        for bad in [
            "2026/11/09 00:00:00",
            "2026-13-09T00:00:00Z",
            "2026-11-32T00:00:00Z",
            "2026-11-09T24:00:00Z",
            "2026-11-09T00:00Z",
            "2026-11-09T00:00:00.Z",
            "2026-11-09T00:00:00+0900",
            "not-a-time",
        ] {
            assert!(!is_rfc3339(bad), "{bad} should be rejected");
        }
    }

    #[test]
    fn validates_and_rejects_edges() {
        let good = base_edge();
        assert!(validate_edge(&good).is_ok());

        let mut bad_basis = base_edge();
        bad_basis.basis = "guessed".into();
        assert!(validate_edge(&bad_basis).err().unwrap().contains("basis"));

        let mut bad_verify = base_edge();
        bad_verify.verification = "approved".into();
        assert!(validate_edge(&bad_verify)
            .err()
            .unwrap()
            .contains("verification"));

        let mut bad_id = base_edge();
        bad_id.edge_id = "edge-deadbeef".into();
        assert!(validate_edge(&bad_id).err().unwrap().contains("edge_id"));

        let mut bad_time = base_edge();
        bad_time.observed_at = "2026/11/09 3pm".into();
        assert!(validate_edge(&bad_time).err().unwrap().contains("RFC3339"));

        let mut bad_extractor = base_edge();
        bad_extractor.extractor.version = String::new();
        assert!(validate_edge(&bad_extractor)
            .err()
            .unwrap()
            .contains("extractor.version"));

        let mut bad_kind = base_edge();
        bad_kind.kind = "included_in_context".into();
        bad_kind.edge_id = edge_id(
            &bad_kind.kind,
            &bad_kind.source_id,
            &bad_kind.target_id,
            &bad_kind.evidence_refs,
        );
        assert!(validate_edge(&bad_kind)
            .err()
            .unwrap()
            .contains("edge.kind"));

        let mut bad_ref = base_edge();
        bad_ref.evidence_refs = vec!["just-a-file.md".into()];
        bad_ref.edge_id = edge_id(
            &bad_ref.kind,
            &bad_ref.source_id,
            &bad_ref.target_id,
            &bad_ref.evidence_refs,
        );
        assert!(validate_edge(&bad_ref)
            .err()
            .unwrap()
            .contains("evidence ref"));
    }

    #[test]
    fn tombstone_roundtrip_preserves_the_record() {
        let mut node = OntologyNode {
            id: "file:src/lib.rs@rev1".into(),
            kind: "file".into(),
            revision: "rev1".into(),
            properties: BTreeMap::new(),
            tombstone: None,
        };
        validate_node(&node).unwrap();
        tombstone_node(&mut node, "superseded by rev2", "2026-11-09T10:00:00Z").unwrap();
        let json = serde_json::to_string(&node).unwrap();
        let back: OntologyNode = serde_json::from_str(&json).unwrap();
        assert_eq!(back, node, "roundtrip must preserve the tombstone");
        let t = back.tombstone.expect("tombstone survives");
        assert_eq!(t.reason, "superseded by rev2");
        // The record itself is still present — deletion is a marker, not a drop.
        assert_eq!(back.id, "file:src/lib.rs@rev1");
    }
}
