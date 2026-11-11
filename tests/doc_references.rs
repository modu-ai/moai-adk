//! C6d #2450: document references are evidence-backed edges, never calls.
use moai_cli::{code_refs, ontology};
use serde_json::Value;
use std::collections::{BTreeMap, BTreeSet};

fn resolve(source: &str, target: &str) -> (Vec<ontology::OntologyEdge>, Vec<Value>) {
    let files = [
        ("src/lib.rs", "pub mod user; pub mod service;"),
        ("src/user.rs", source),
        ("src/service.rs", target),
    ];
    let scans = files
        .iter()
        .map(|(p, text)| {
            let scan = code_refs::scan(text);
            assert_eq!(
                code_refs::decode(&code_refs::encode(&scan)),
                Some(scan.clone())
            );
            (p.to_string(), (format!("sha256:{p}"), scan))
        })
        .collect::<BTreeMap<_, _>>();
    let paths = files
        .iter()
        .map(|(p, _)| p.to_string())
        .collect::<BTreeSet<_>>();
    let (refs, misses) = code_refs::resolve(&scans, &paths, &BTreeMap::new());
    (
        code_refs::edges(&refs, "generation", "2026-11-11T00:00:00Z"),
        misses,
    )
}

#[test]
fn qualified_doc_reference_checks_the_definition_and_both_evidence_spans() {
    let (edges, _) = resolve("//! See `service::parse`.\n", "pub fn parse() {}\n");
    let doc = edges
        .iter()
        .find(|e| e.kind == "documents")
        .expect("direct doc edge");
    ontology::validate_edge(doc).unwrap();
    assert_eq!(doc.source_id, "file:src/user.rs");
    assert_eq!(doc.target_id, "file:src/service.rs");
    assert_eq!(doc.basis, "inferred");
    assert_eq!(doc.verification, "unreviewed");
    assert!(doc.uncertainty.as_deref().unwrap().contains("doc"));
    assert_eq!(
        doc.evidence_refs,
        [
            "file:src/service.rs#L1@sha256:src/service.rs",
            "file:src/user.rs#L1@sha256:src/user.rs",
        ]
    );
    assert!(!edges.iter().any(|e| e.kind == "calls"));
}

#[test]
fn unresolved_or_ambiguous_doc_item_is_missing_not_a_module_only_match() {
    for (target, reason) in [
        ("pub fn other() {}", "unresolved"),
        ("pub fn parse() {}\npub fn parse() {}", "ambiguous"),
        ("pub fn parse() {} pub fn parse() {}", "ambiguous"),
        ("impl Service { pub fn parse() {} }", "unresolved"),
        ("// pub fn parse() {}", "unresolved"),
        ("pub fn unrelated<const parse: usize>() {}", "unresolved"),
    ] {
        let (edges, misses) = resolve("//! `crate::service::parse`\n", target);
        assert!(!edges.iter().any(|e| e.kind == "documents"), "{target}");
        assert!(
            misses
                .iter()
                .any(|m| m["kind"] == "doc_reference_missing" && m["reason"] == reason),
            "{misses:?}"
        );
    }
}

#[test]
fn a_malformed_stored_doc_path_is_refused() {
    assert!(code_refs::decode("r\tdocuments\t\t\t\t1\n").is_none());
    assert!(code_refs::decode("r\tdocuments\t\tparse\t\t1\n").is_none());
}

#[test]
fn doc_path_and_definition_mutants_remove_the_edge() {
    for (source, target) in [
        ("//! `crate::absent::parse`", "pub fn parse() {}"),
        ("//! `crate::service::renamed`", "pub fn parse() {}"),
        ("//! `crate::service::parse`", "pub fn renamed() {}"),
    ] {
        let (edges, misses) = resolve(source, target);
        assert!(!edges.iter().any(|e| e.kind == "documents"));
        assert!(
            misses.iter().any(|m| m["kind"] == "doc_reference_missing"),
            "{misses:?}"
        );
    }
}

#[test]
fn doc_scanner_ignores_strings_normal_comments_fenced_examples_and_bare_names() {
    for source in [
        "// `crate::service::parse`\n",
        "//// `crate::service::parse`\n",
        "/* //! `crate::service::parse` */\n",
        "const S: &str = \"//! `crate::service::parse`\";",
        "const S: &str = r###\"\n//! `crate::service::parse`\n\"###;",
        "//! ```rust\n//! `crate::service::parse`\n//! ```\n",
        "//! ~~~~rust\n//! `crate::service::parse`\n//! ~~~~\n",
        "//! `parse`\n",
        "//!     `crate::service::parse`\n",
    ] {
        let (edges, _) = resolve(source, "pub fn parse() {}");
        assert!(!edges.iter().any(|e| e.kind == "documents"), "{source}");
    }
}

#[test]
fn local_bindings_do_not_fall_back_to_a_same_named_root_module() {
    for source in [
        "//! `service::parse`\nfn service() {}",
        "//! `service::parse`\nuse crate::other as service;",
    ] {
        let (edges, misses) = resolve(source, "pub fn parse() {}");
        assert!(!edges.iter().any(|e| e.kind == "documents"));
        assert!(misses.iter().any(|m| m["kind"] == "doc_reference_missing"));
    }
}

#[test]
fn line_block_and_inline_module_doc_comments_keep_their_scope() {
    for source in [
        "/// [`crate::service::parse`]\npub fn user() {}",
        "/** `crate::service::parse` */\npub fn user() {}",
        "/*! `crate::service::parse` */",
        "pub mod inner {\n//! `super::super::service::parse`\n}",
    ] {
        let (edges, _) = resolve(source, "pub fn parse() {}");
        assert!(edges.iter().any(|e| e.kind == "documents"), "{source}");
    }
}
