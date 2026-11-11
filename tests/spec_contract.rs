//! GEARS / legacy EARS requirement contract (Issue #2001, 05-A).
use moai_cli::spec::{
    amend, approve, content_digest, import_ears, parse_gears, record_semantic_review, validate,
    SortedKeys, SpecDocument,
};
use serde_json::{json, Value};
use std::{fs, path::Path};

fn fixture(rel: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("tests/fixtures/spec")
            .join(rel),
    )
    .unwrap_or_else(|_| panic!("missing fixture {rel}"))
}
fn json_fixture(rel: &str) -> Value {
    serde_json::from_str(&fixture(rel)).unwrap()
}
fn doc(rel: &str) -> SpecDocument {
    serde_json::from_str(&fixture(rel)).unwrap()
}

#[test]
fn gears_parse_splits_clauses_into_typed_records() {
    for name in ["gears_where_when", "gears_while", "gears_custom_subject"] {
        let input = json_fixture(&format!("{name}.json"));
        let req = parse_gears(
            input["id"].as_str().unwrap(),
            input["text"].as_str().unwrap(),
        )
        .unwrap_or_else(|e| panic!("{name}: {e:?}"));
        assert_eq!(
            serde_json::to_value(&req).unwrap(),
            json_fixture(&format!("golden/{name}.json")),
            "{name}"
        );
    }
}

#[test]
fn gears_parse_reports_unsplittable_text_with_its_position() {
    let cases = [
        ("the runner runs tests", "no `shall`"),
        (
            "The runner shall test and shall lint.",
            "more than one `shall`",
        ),
        (
            "When a PR opens, where Rust is used, the runner shall test.",
            "out of order",
        ),
        ("When a PR opens, shall test.", "empty subject"),
        ("The runner shall .", "empty behavior"),
    ];
    for (text, reason) in cases {
        let err = parse_gears("REQ-X", text).expect_err(text);
        assert!(err.reason.contains(reason), "{text}: {}", err.reason);
        assert!(err.offset <= text.len(), "{text}");
    }
    // Position points at the offending clause, not at the start of the text.
    let text = "When a PR opens, where Rust is used, the runner shall test.";
    assert_eq!(parse_gears("REQ-X", text).unwrap_err().offset, 17);
}

#[test]
fn gears_parse_refuses_secret_shaped_or_oversized_text_without_echo() {
    for text in [
        "The runner shall use token=abc123.".to_string(),
        "The runner shall read ghp_0123456789.".to_string(),
        "The runner shall\u{7} beep.".to_string(),
        format!("The runner shall {}.", "x".repeat(70 * 1024)),
        // Personal absolute paths (REQ-05A-08, #2001 r26).
        "The runner shall read /Users/someone/notes.md.".to_string(),
        "The runner shall read `/home/someone/notes.md`.".to_string(),
        "The runner shall read ~/notes.md.".to_string(),
        "The runner shall read C:\\Users\\someone.".to_string(),
    ] {
        let err = parse_gears("REQ-X", &text).unwrap_err();
        assert!(!err.reason.contains("abc123") && !err.reason.contains("ghp_"));
        assert!(!err.reason.contains("someone") && !err.reason.contains("notes"));
    }
    // Repository-relative paths, other absolute paths, `:` and ordinary slashes stay allowed.
    for text in [
        "The generator shall write skills/moai/SKILL.md.",
        "The runner shall write `/tmp/out` and read src/x.rs:12.",
        "The runner shall accept and/or reject 1/2 of the cases.",
    ] {
        assert!(parse_gears("REQ-X", text).is_ok(), "{text}");
    }
    // A word that merely ends in "sk-" is not a secret.
    assert!(parse_gears("REQ-X", "The task-runner shall start.").is_ok());
}

#[test]
fn ears_import_preserves_source_and_never_rewrites_if_then() {
    let corpus = fixture("ears_corpus.md");
    let reqs = import_ears(&corpus).unwrap();
    let ids: Vec<&str> = reqs.iter().map(|r| r.id.as_str()).collect();
    assert_eq!(
        ids,
        [
            "REQ-E-001",
            "REQ-E-002",
            "REQ-E-003",
            "REQ-E-004",
            "REQ-E-005",
            "REQ-E-006",
            "REQ-E-007"
        ]
    );
    let patterns: Vec<&str> = reqs
        .iter()
        .map(|r| r.import.as_ref().unwrap().pattern.as_str())
        .collect();
    assert_eq!(
        patterns,
        [
            "ubiquitous",
            "event-driven",
            "state-driven",
            "optional-feature",
            "unwanted-behaviour",
            "complex",
            "ubiquitous"
        ]
    );
    for r in &reqs {
        let imp = r.import.as_ref().unwrap();
        assert_eq!(r.syntax, "EARS-import");
        // The recorded range slices exactly the original text out of the corpus.
        assert_eq!(
            &corpus[imp.source_range.start_byte..imp.source_range.end_byte],
            r.text
        );
        // No grammar conversion happens on import.
        assert!(r.subject.is_none() && r.behavior.is_none() && r.when.is_none());
    }
    let unwanted = &reqs[4];
    assert!(unwanted.text.contains("**If** the target changed"));
    assert!(unwanted.text.contains("**then** the installer"));
    assert_eq!(
        unwanted.import.as_ref().unwrap().classification.as_deref(),
        Some("safety")
    );
    assert_eq!(
        unwanted.import.as_ref().unwrap().source_range.start_line,
        11
    );
    assert_eq!(unwanted.import.as_ref().unwrap().source_range.end_line, 12);
    assert!(
        !unwanted.unresolved.is_empty(),
        "If…then needs an event-or-state review"
    );
    assert_eq!(reqs[6].import.as_ref().unwrap().classification, None);
    // Re-running the import gives the same bytes.
    let again = import_ears(&corpus).unwrap();
    assert_eq!(
        serde_json::to_vec(&reqs).unwrap(),
        serde_json::to_vec(&again).unwrap()
    );
}

#[test]
fn validation_axes_do_not_promote() {
    let d = doc("snapshot_canonical_a.json");
    let mut v = validate(&d);
    let j = serde_json::to_value(&v).unwrap();
    assert_eq!(j["syntax"], "passed");
    assert_eq!(j["structure"], "passed");
    assert_eq!(j["semantic"], "unknown");
    assert_eq!(j["semantic_source"], "model-finding");
    assert_eq!(j["execution"], "not_run");
    record_semantic_review(&mut v, vec!["REQ-S-001 reads as observable".into()]);
    let j = serde_json::to_value(&v).unwrap();
    assert_eq!(j["semantic"], "reviewed");
    assert_eq!(j["execution"], "not_run");
    // A failing syntax axis leaves the other axes where they were as well.
    let mut bad = d.clone();
    bad.requirements[0].text = "the installer records receipts".into();
    let j = serde_json::to_value(validate(&bad)).unwrap();
    assert_eq!(j["syntax"], "failed");
    assert_eq!(j["semantic"], "unknown");
    assert_eq!(j["execution"], "not_run");
}

#[test]
fn structure_dangling_refs_fail_with_the_missing_ids() {
    let j = serde_json::to_value(validate(&doc("dangling_refs.json"))).unwrap();
    assert_eq!(j["structure"], "failed");
    assert_eq!(
        j["structure_report"]["dangling_refs"],
        json!([
            {"from": "REQ-D-001", "to": "AC-D-9"},
            {"from": "REQ-D-001", "to": "REQ-D-404"}
        ])
    );
    assert_eq!(
        j["structure_report"]["orphan_acceptance"],
        json!(["AC-D-2"])
    );
    assert_eq!(j["structure_report"]["duplicate_ids"], json!([]));
    assert_eq!(j["syntax"], "passed");
}

#[test]
fn snapshot_digest_is_canonical_sensitive_and_immutable() {
    let a = doc("snapshot_canonical_a.json");
    let b = doc("snapshot_canonical_b.json");
    let da = content_digest(&a, &SortedKeys);
    assert_eq!(
        da,
        content_digest(&b, &SortedKeys),
        "key order and whitespace"
    );
    assert!(da.starts_with("sha256:") && da.len() == 71);
    let mut c = a.clone();
    c.requirements[0].text = c.requirements[0].text.replace("receipt", "Receipt");
    assert_ne!(da, content_digest(&c, &SortedKeys), "one character");

    let approved = approve(&a, &SortedKeys);
    assert_eq!(approved.status, "approved");
    let snap = approved.snapshot.clone().unwrap();
    assert_eq!(snap.digest, da);
    assert_eq!(snap.reference, "SPEC-T-SNAP@1");

    let next = amend(&approved, c.requirements.clone(), c.acceptance.clone());
    assert_eq!(next.revision, 2);
    assert_eq!(next.status, "proposed");
    assert!(next.snapshot.is_none());
    // The approved revision and its digest are untouched by the amendment.
    assert_eq!(approved.snapshot.unwrap().digest, da);
    assert_eq!(content_digest(&approved_again(&a), &SortedKeys), da);
}

/// Issue #1992: serde_json now preserves key order, so SortedKeys sorts by itself. The digest
/// is pinned to the value measured before `preserve_order` was enabled (2026-11-08).
#[test]
fn sorted_keys_is_independent_of_map_order() {
    use moai_cli::spec::Canonical;
    let a = doc("snapshot_canonical_a.json");
    assert_eq!(
        content_digest(&a, &SortedKeys),
        "sha256:6326dbfddd4c2dad69536ca96b339b1625078b238a940f6014c00ed2dbdb8f6f"
    );
    let shuffled: serde_json::Value =
        serde_json::from_str(r#"{"b":1,"a":{"d":[{"y":1,"x":2}],"c":2}}"#).unwrap();
    assert_eq!(
        SortedKeys.bytes(&shuffled),
        br#"{"a":{"c":2,"d":[{"x":2,"y":1}]},"b":1}"#
    );
}

fn approved_again(a: &SpecDocument) -> SpecDocument {
    approve(a, &SortedKeys)
}
