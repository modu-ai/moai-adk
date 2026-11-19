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

/// C-17 #2450 G-006 (user decision A 2026-11-15): the contracted machine token in a
/// code-level string literal resolves through the explicit alias contract to the file
/// that defines the concept, as an ordinary `documents` edge. The seven-code and G-002
/// producer rules and the plain-doc block contract are untouched. The helper's decode
/// round-trip also pins the codec: a `concept` row without a contract token makes the
/// stored scan unreadable.
fn concept_resolve(files: &[(&str, &str)]) -> (Vec<ontology::OntologyEdge>, Vec<Value>) {
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
        code_refs::edges(&refs, "generation", "2026-11-19T00:00:00Z"),
        misses,
    )
}

#[test]
fn contracted_concept_token_produces_a_documents_edge_to_the_definition_file() {
    let files = [
        ("src/lib.rs", "pub mod claim_verify;\npub mod handoff;\n"),
        (
            "src/claim_verify.rs",
            "//! Two machine lines close a claim.\nconst HANDOFF_PREFIX: &str = \"moai-claim-handoff:\";\n",
        ),
        (
            "src/handoff.rs",
            "//! The handoff protocol module.\npub fn claim_handoff() {}\n",
        ),
    ];
    let (edges, misses) = concept_resolve(&files);
    let doc = edges
        .iter()
        .find(|e| e.kind == "documents")
        .expect("the concept mention resolves to the definition file");
    ontology::validate_edge(doc).unwrap();
    assert_eq!(doc.source_id, "file:src/claim_verify.rs");
    assert_eq!(doc.target_id, "file:src/handoff.rs");
    assert_eq!(doc.basis, "inferred");
    assert_eq!(doc.verification, "unreviewed");
    assert_eq!(
        doc.evidence_refs,
        [
            "file:src/claim_verify.rs#L2@sha256:src/claim_verify.rs",
            "file:src/handoff.rs#L2@sha256:src/handoff.rs",
        ]
    );
    assert!(!misses
        .iter()
        .any(|m| m["reference"] == "moai-claim-handoff:"));
}

#[test]
fn concept_resolution_needs_the_definition_and_reports_reasoned_gaps() {
    // The definition file is not in the file set: a reasoned gap, never an edge.
    let absent = [
        ("src/lib.rs", "pub mod claim_verify;\n"),
        (
            "src/claim_verify.rs",
            "const HANDOFF_PREFIX: &str = \"moai-claim-handoff:\";\n",
        ),
    ];
    let (edges, misses) = concept_resolve(&absent);
    assert!(!edges.iter().any(|e| e.kind == "documents"));
    assert!(misses.iter().any(|m| {
        m["kind"] == "doc_reference_missing" && m["reason"] == "alias_definition_absent"
    }));
    // The confirm symbol is absent (renamed away) or declared more than once.
    for (definition, reason) in [
        (
            "//! protocol.\npub fn renamed_away() {}\n",
            "alias_definition_absent",
        ),
        (
            "pub fn claim_handoff() {}\npub fn claim_handoff<C: Clock>() {}\n",
            "alias_definition_ambiguous",
        ),
    ] {
        let files = [
            ("src/lib.rs", "pub mod claim_verify;\npub mod handoff;\n"),
            (
                "src/claim_verify.rs",
                "const HANDOFF_PREFIX: &str = \"moai-claim-handoff:\";\n",
            ),
            ("src/handoff.rs", definition),
        ];
        let (edges, misses) = concept_resolve(&files);
        assert!(!edges.iter().any(|e| e.kind == "documents"), "{definition}");
        assert!(
            misses.iter().any(|m| {
                m["kind"] == "doc_reference_missing"
                    && m["reason"] == reason
                    && m["reference"] == "moai-claim-handoff:"
            }),
            "{definition}: {misses:?}"
        );
    }
}

#[test]
fn a_same_file_concept_mention_stays_silent() {
    let files = [
        ("src/lib.rs", "pub mod handoff;\n"),
        (
            "src/handoff.rs",
            "const HANDOFF_PREFIX: &str = \"moai-claim-handoff:\";\npub fn claim_handoff() {}\n",
        ),
    ];
    let (edges, misses) = concept_resolve(&files);
    assert!(!edges.iter().any(|e| e.kind == "documents"));
    assert!(!misses
        .iter()
        .any(|m| m["reference"] == "moai-claim-handoff:"));
}

#[test]
fn only_string_literals_only_contracted_tokens_become_concept_references() {
    // A mention needs a string literal whose body starts with the token: without the
    // colon, in another literal, in a doc backtick span or inside a macro token tree
    // is prose or opaque, not a mention. A raw string literal does mention it.
    for source in [
        "const S: &str = \"the moai-claim-handoff prefix without the colon\";\n",
        "const S: &str = \"moai-claim-release: comment=<id>\";\n",
        "//! `moai-claim-handoff: comment=<id>`\n",
        "const T: &str = stringify!(\"moai-claim-handoff: x\");\n",
    ] {
        let files = [
            ("src/lib.rs", "pub mod user;\npub mod handoff;\n"),
            ("src/user.rs", source),
            ("src/handoff.rs", "pub fn claim_handoff() {}\n"),
        ];
        let (edges, misses) = concept_resolve(&files);
        assert!(!edges.iter().any(|e| e.kind == "documents"), "{source}");
        assert!(
            !misses
                .iter()
                .any(|m| m["reference"] == "moai-claim-handoff:"),
            "{source}: {misses:?}"
        );
    }
    let raw = [
        ("src/lib.rs", "pub mod user;\npub mod handoff;\n"),
        (
            "src/user.rs",
            "const R: &str = r#\"moai-claim-handoff: x\"#;\n",
        ),
        ("src/handoff.rs", "pub fn claim_handoff() {}\n"),
    ];
    let (edges, _) = concept_resolve(&raw);
    let doc = edges
        .iter()
        .find(|e| e.kind == "documents")
        .expect("a raw string literal mentions the token the same way");
    assert_eq!(doc.source_id, "file:src/user.rs");
    // A container doc block keeps its exclusion count and produces no concept reference.
    let files = [
        ("src/lib.rs", "pub mod user;\npub mod handoff;\n"),
        ("src/user.rs", "//! > `moai-claim-handoff: comment=<id>`\n"),
        ("src/handoff.rs", "pub fn claim_handoff() {}\n"),
    ];
    let (_, misses) = concept_resolve(&files);
    assert!(misses
        .iter()
        .any(|m| m["kind"] == "doc_reference_unsupported"));
}

#[test]
fn a_stored_concept_row_needs_a_contract_token() {
    assert!(code_refs::decode("r\tconcept\t\tmoai-claim-handoff:\t\t7\n").is_some());
    assert!(code_refs::decode("r\tconcept\t\tmoai-claim-release:\t\t7\n").is_none());
    assert!(code_refs::decode("r\tconcept\t\tunbound-token\t\t7\n").is_none());
}
