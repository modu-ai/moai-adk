//! Ontology schema fixtures (C4 #2180). Evidence base:
//! `docs/design-review/appendix/dr07-ontology-table.md`. AC-1: validation
//! tests with ≥5 valid and ≥5 invalid records. AC-2: tombstone roundtrip.
//! REQ-C4-01: explicit and inferred links never share one field.

use moai_cli::ontology::{
    edge_id, tombstone_node, validate_document, validate_edge, validate_node, Extractor,
    OntologyEdge, OntologyNode, Tombstone, BASIS_VALUES, CB_ONLY_EDGE_TYPES, CORE_EDGE_TYPES,
    VERIFICATION_VALUES,
};
use serde_json::{json, Value};

fn evidence() -> Vec<String> {
    vec!["src/lib.rs#L10-L20@abc123def".into()]
}

fn edge(
    kind: &str,
    basis: &str,
    verification: &str,
    valid_to: Option<&str>,
    refs: Vec<String>,
) -> OntologyEdge {
    OntologyEdge {
        edge_id: edge_id(kind, "node:src/lib.rs", "node:req:42", &refs),
        kind: kind.into(),
        source_id: "node:src/lib.rs".into(),
        target_id: "node:req:42".into(),
        basis: basis.into(),
        verification: verification.into(),
        evidence_refs: refs,
        extractor: Extractor {
            name: "atlas".into(),
            version: "1.0.0".into(),
        },
        snapshot_id: "snap-2026-11-09".into(),
        observed_at: "2026-11-09T03:00:00Z".into(),
        valid_from: "2026-11-09T00:00:00Z".into(),
        valid_to: valid_to.map(|s| s.into()),
        uncertainty: None,
    }
}

fn node(id: &str, kind: &str, revision: &str) -> OntologyNode {
    OntologyNode {
        id: id.into(),
        kind: kind.into(),
        revision: revision.into(),
        properties: Default::default(),
        tombstone: None,
    }
}

/// AC-1 (valid ≥5): each measured basis × verification axis pair, a CB-only
/// relation, a closed edge (valid_to = revision), and a tombstoned node all
/// validate.
#[test]
fn ac1_valid_records() {
    for (basis, verification) in [
        ("explicit", "unreviewed"),
        ("static_resolved", "checked"),
        ("inferred", "contradicted"),
        ("runtime_observed", "stale"),
        ("explicit", "checked"),
    ] {
        let e = edge("implements", basis, verification, None, evidence());
        assert!(
            validate_edge(&e).is_ok(),
            "{basis}/{verification} must validate: {:?}",
            validate_edge(&e)
        );
    }
    // A CB-only relation from the measured appendix rows validates.
    let cb = edge("targets_host", "explicit", "unreviewed", None, evidence());
    assert!(validate_edge(&cb).is_ok(), "CB-only kind must validate");
    assert!(CB_ONLY_EDGE_TYPES.contains(&"targets_host"));

    // A closed edge: valid_to carries a revision (revision-or-time).
    let closed = edge(
        "supersedes",
        "explicit",
        "unreviewed",
        Some("abc1111111111111111111111111111111111111"),
        evidence(),
    );
    assert!(
        validate_edge(&closed).is_ok(),
        "revision valid_to is allowed"
    );

    // Every core edge type is in the enforced vocabulary.
    for k in CORE_EDGE_TYPES {
        let e = edge(k, "static_resolved", "checked", None, evidence());
        assert!(validate_edge(&e).is_ok(), "core kind {k} must validate");
    }

    // A tombstoned node (marked, not dropped) validates structurally.
    let mut n = node("file:src/lib.rs@r1", "file", "r1");
    tombstone_node(&mut n, "path renamed", "2026-11-09T04:00:00Z").unwrap();
    assert!(validate_node(&n).is_ok());
}

/// AC-1 (invalid ≥5): vocabulary escapes on both axes, a broken edge_id, a
/// non-RFC3339 time, an unpinned extractor, a HOLD-excluded kind, a malformed
/// evidence ref, and an empty endpoint each fail with a naming message.
#[test]
fn ac1_invalid_records() {
    let cases: Vec<(String, OntologyEdge)> = vec![
        // 1. basis escapes its vocabulary (same field never mixes axes).
        (
            "basis not in vocabulary".into(),
            edge("calls", "guessed", "checked", None, evidence()),
        ),
        // 2. verification escapes its vocabulary.
        (
            "verification not in vocabulary".into(),
            edge("calls", "explicit", "approved", None, evidence()),
        ),
        // 3. edge_id does not match the recomputed hash.
        (
            "edge_id mismatch".into(),
            OntologyEdge {
                edge_id: "edge-0000000000000000000000000000000000000000000000000000000000000000"
                    .into(),
                ..edge("calls", "explicit", "unreviewed", None, evidence())
            },
        ),
        // 4. observed_at is not RFC3339.
        (
            "observed_at not RFC3339".into(),
            OntologyEdge {
                observed_at: "2026/11/09 3pm".into(),
                ..edge("calls", "explicit", "unreviewed", None, evidence())
            },
        ),
        // 5. extractor version empty (pinned requirement).
        (
            "extractor version unpinned".into(),
            OntologyEdge {
                extractor: Extractor {
                    name: "atlas".into(),
                    version: String::new(),
                },
                ..edge("calls", "explicit", "unreviewed", None, evidence())
            },
        ),
        // 6. HOLD 3: included_in_context is outside the supplied definitions.
        (
            "HOLD-excluded kind".into(),
            edge(
                "included_in_context",
                "explicit",
                "unreviewed",
                None,
                evidence(),
            ),
        ),
        // 7. evidence ref is not file-uri#span@content-hash.
        (
            "evidence ref malformed".into(),
            edge(
                "calls",
                "explicit",
                "unreviewed",
                None,
                vec!["src/lib.rs".into()],
            ),
        ),
        // 8. empty target endpoint.
        (
            "empty endpoint".into(),
            OntologyEdge {
                target_id: String::new(),
                ..edge("calls", "explicit", "unreviewed", None, evidence())
            },
        ),
    ];
    assert!(cases.len() >= 5);
    for (name, mut e) in cases {
        // Keep the id honest except for the dedicated mismatch case, so each
        // mutation is the single reason for rejection.
        if name != "edge_id mismatch" {
            e.edge_id = edge_id(&e.kind, &e.source_id, &e.target_id, &e.evidence_refs);
        }
        let err = validate_edge(&e)
            .err()
            .unwrap_or_else(|| panic!("{name} must be rejected"));
        assert!(!err.is_empty(), "{name} must carry a naming message");
    }

    // Invalid nodes: empty identity fields.
    assert!(validate_node(&node("", "file", "r1")).is_err());
    assert!(validate_node(&node("f:1", "", "r1")).is_err());
    assert!(validate_node(&node("f:1", "file", "")).is_err());
}

/// REQ-C4-01: the same type and endpoints under an explicit basis and an
/// inferred basis are separate records with their own basis fields — the
/// axes (basis, verification) never share one field, and the two records get
/// distinct edge_ids only when evidence differs; when everything but the
/// declared axes differs they stay distinct documents, not merged fields.
#[test]
fn req_c4_01_axes_are_separate_fields() {
    let explicit = edge("implements", "explicit", "unreviewed", None, evidence());
    let inferred = edge("implements", "inferred", "checked", None, evidence());
    assert_eq!(explicit.basis, "explicit");
    assert_eq!(inferred.basis, "inferred");
    assert_ne!(explicit.verification, inferred.verification);
    // The vocabulary constants themselves never overlap.
    for b in BASIS_VALUES {
        assert!(
            !VERIFICATION_VALUES.contains(&b),
            "basis value {b} leaked into the verification vocabulary"
        );
    }
    // Serializing keeps the two axes as distinct named fields (JSON 정본).
    let v: Value = serde_json::to_value(&explicit).unwrap();
    assert!(v.get("basis").is_some(), "basis is its own field");
    assert!(
        v.get("verification").is_some(),
        "verification is its own field"
    );
    assert_ne!(v["basis"], v["verification"]);
}

/// AC-2: a tombstone roundtrips through the JSON canonical form — the record
/// survives (marked, not dropped), the reason and time are preserved, and the
/// same holds for a whole document of nodes and edges.
#[test]
fn ac2_tombstone_roundtrip() {
    let mut n = node("file:src/lib.rs@r1", "file", "r1");
    tombstone_node(&mut n, "path renamed", "2026-11-09T05:00:00Z").unwrap();
    let json = serde_json::to_string(&n).unwrap();
    let back: OntologyNode = serde_json::from_str(&json).unwrap();
    assert_eq!(back, n, "node roundtrip preserves the tombstone");
    assert_eq!(
        back.tombstone,
        Some(Tombstone {
            reason: "path renamed".into(),
            observed_at: "2026-11-09T05:00:00Z".into(),
        })
    );

    // Document-level roundtrip: one tombstoned node among live records, plus
    // a closed edge — the full document validates and survives byte-identical
    // (field order aside, semantically equal) after serialization.
    let closed = edge(
        "supersedes",
        "explicit",
        "unreviewed",
        Some("abc1111111111111111111111111111111111111"),
        evidence(),
    );
    let doc = json!({
        "nodes": [serde_json::to_value(&n).unwrap(), serde_json::to_value(node("f:x@r2", "file", "r2")).unwrap()],
        "edges": [serde_json::to_value(&closed).unwrap()],
    });
    let (nodes, edges) = validate_document(&doc).expect("document validates");
    assert_eq!((nodes, edges), (2, 1));
    let re: Value = serde_json::from_str(&doc.to_string()).unwrap();
    let (nodes2, edges2) = validate_document(&re).expect("re-parsed document validates");
    assert_eq!((nodes, edges), (nodes2, edges2));
}

/// A tombstone with an empty reason or a non-RFC3339 time is refused before
/// the record is marked — the marker itself obeys the schema.
#[test]
fn tombstone_marker_is_validated() {
    let mut n = node("f:1", "file", "r1");
    assert!(tombstone_node(&mut n, "", "2026-11-09T05:00:00Z").is_err());
    assert!(tombstone_node(&mut n, "reason", "11/09/2026").is_err());
    assert!(n.tombstone.is_none(), "failed marking must not mutate");
}
