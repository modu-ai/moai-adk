//! #2450 repair: literal auditor fixtures and adjacent lexical boundaries through CLI.
mod common;
use common::CleanEnv;
use moai_cli::code_refs;
use serde_json::{json, Value};
use std::{
    fs,
    path::Path,
    process::{Command, Stdio},
};

const LIB: &str = "pub mod user;\npub mod service;\n";
const USER: &str = "//! `crate::service::parse`\n";
const SERVICE: &str = "pub fn parse() {}\n";

fn write(root: &Path, path: &str, text: &str) {
    let path = root.join(path);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

fn run(root: &Path, home: &Path, op: &str, input: Value) -> Value {
    // Only the separate mutation runner supplies an alternate freshly built CLI.
    let binary = std::env::var_os("MOAI_C6D_BOUNDARY_BINARY")
        .unwrap_or_else(|| env!("CARGO_BIN_EXE_moai-cli").into());
    let mut child = Command::new(binary)
        .args([op, root.to_str().unwrap()])
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", home)
        .env("USERPROFILE", home)
        .env("XDG_CONFIG_HOME", home.join("config"))
        .env("XDG_DATA_HOME", home.join("data"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_TERMINAL_PROMPT", "0")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let output = child.wait_with_output().unwrap();
    assert!(
        output.status.success(),
        "{op}: {}",
        String::from_utf8_lossy(&output.stderr)
    );
    let v: Value = serde_json::from_slice(&output.stdout).unwrap();
    v.get("result").cloned().unwrap_or(v)
}

fn query(lib: &str, user: &str, service: &str) -> (Value, Value) {
    let temp = tempfile::tempdir().unwrap();
    let root = temp.path().join("project");
    let home = temp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    write(&root, ".gitignore", ".moai-cli/index/\n");
    write(&root, "src/lib.rs", lib);
    write(&root, "src/user.rs", user);
    write(&root, "src/service.rs", service);
    if lib.contains("#[path = \"real.rs\"]") {
        write(&root, "src/real.rs", SERVICE);
    }
    for args in [
        vec!["init", "-q"],
        vec!["add", "-A"],
        vec!["commit", "-q", "-m", "fixture"],
    ] {
        assert!(common::git::command()
            .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
            .args(args)
            .current_dir(&root)
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status()
            .unwrap()
            .success());
    }
    let built = run(&root, &home, "index-build", json!({"protocol_version":1}));
    let graph = run(
        &root,
        &home,
        "atlas-graph",
        json!({"protocol_version":1,"query":{"start":["file:src/user.rs"]},"limits":{"max_hops":1,"max_edges":5000}}),
    );
    let records = moai_cli::index::index_records(
        &root,
        &moai_cli::index::IndexQueryRequest {
            protocol_version: 1,
        },
    )
    .unwrap();
    assert_eq!(graph["state"], "fresh");
    assert_eq!(records["state"], "fresh");
    assert_eq!(graph["generation_id"], built["generation_id"]);
    assert_eq!(records["generation_id"], built["generation_id"]);
    (graph["graph"].clone(), records["graph"].clone())
}

fn docs(graph: &Value) -> usize {
    graph["edges"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["kind"] == "documents")
        .count()
}
fn excluded_blocks(graph: &Value) -> u64 {
    graph["missing"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|m| m["kind"] == "doc_reference_unsupported")
        .map(|m| m["count"].as_u64().unwrap())
        .sum()
}

fn doc_misses(graph: &Value) -> usize {
    graph["missing"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|m| m["kind"] == "doc_reference_missing")
        .count()
}

#[test]
fn d1_literal_stringify_input_is_not_a_definition() {
    let service = "const TEXT: &str = stringify!(fn parse() {});\n";
    let (graph, _) = query(LIB, USER, service);
    eprintln!("D1 CLI fresh; false documents={}", docs(&graph));
    assert_eq!(docs(&graph), 0, "opaque macro input is not a declaration");
    assert_eq!(doc_misses(&graph), 1);
    assert!(!code_refs::scan(service)
        .definitions
        .iter()
        .any(|d| d.name == "parse"));
}

#[test]
fn d2_literal_path_attribute_with_doc_never_follows_the_decoy() {
    let lib = "pub mod user;\n#[path = \"real.rs\"]\n/// documentation\npub mod service;\n";
    let (graph, records) = query(lib, USER, SERVICE);
    eprintln!("D2 CLI fresh; false documents={}", docs(&graph));
    assert_eq!(
        docs(&graph),
        0,
        "unsupported path must not resolve to service.rs"
    );
    assert!(records["missing"]
        .as_array()
        .unwrap()
        .iter()
        .any(|m| m["kind"] == "module_path_attribute"));
    assert!(
        code_refs::scan(lib)
            .mods
            .iter()
            .find(|m| m.name == "service")
            .unwrap()
            .path_attr
    );
}

#[test]
fn d3_literal_blockquote_fence_produces_no_candidate_edge_or_missing() {
    let user = "//! > ~~~rust\n//! > `crate::service::parse`\n//! > ~~~\n";
    let (graph, _) = query(LIB, user, SERVICE);
    eprintln!("D3 CLI fresh; false documents={}", docs(&graph));
    assert_eq!(docs(&graph), 0);
    assert_eq!(doc_misses(&graph), 0);
    assert!(!code_refs::scan(user)
        .refs
        .iter()
        .any(|r| r.kind == "documents"));
}

#[test]
fn macro_delimiter_matrix_excludes_only_opaque_declarations() {
    for (id, service) in [
        ("M-paren", "discard!(fn parse() {});"),
        ("M-bracket", "discard![fn parse() {}];"),
        ("M-brace", "discard!{fn parse() {}}"),
        ("M-nested", "discard!([ { fn parse() {} } ]);"),
        ("M-qualified", "tools::discard!(fn parse() {});"),
        ("M-attribute", "#[opaque(fn parse() {})]\npub fn other() {}"),
        (
            "M-definition",
            "macro_rules! discard { () => (fn parse() {}); }",
        ),
    ] {
        let (graph, _) = query(LIB, USER, service);
        assert_eq!(docs(&graph), 0, "{id}");
        assert_eq!(doc_misses(&graph), 1, "{id}");
        let actual = format!("{service}\npub fn parse() {{}}\n");
        let (graph, _) = query(LIB, USER, &actual);
        assert_eq!(docs(&graph), 1, "{id}: real following definition survives");
        eprintln!("boundary {id}: negative + real-definition control passed");
    }
}

#[test]
fn attribute_doc_order_matrix_preserves_path_refusal_and_normal_modules() {
    for (id, attributes) in [
        ("A-path-doc", "#[path = \"real.rs\"]\n/// documentation\n"),
        ("A-doc-path", "/// documentation\n#[path = \"real.rs\"]\n"),
        (
            "A-mixed",
            "/// before\n#[path = \"real.rs\"]\n#[cfg(any())]\n/// after\n",
        ),
        (
            "A-block",
            "#[path = \"real.rs\"]\n/** block\n * documentation */\n",
        ),
        (
            "A-restricted",
            "#[path = \"real.rs\"]\n/// documentation\npub(crate) ",
        ),
    ] {
        let lib = format!("pub mod user;\n{attributes}mod service;\n");
        let user = "//! `crate::service::parse`\nuse crate::service::parse;\nfn user() { crate::service::parse(); }\n";
        let (graph, records) = query(&lib, user, SERVICE);
        assert!(
            graph["edges"]
                .as_array()
                .unwrap()
                .iter()
                .all(|e| e["target_id"] != "file:src/service.rs"),
            "{id}"
        );
        assert!(
            records["missing"]
                .as_array()
                .unwrap()
                .iter()
                .any(|m| m["kind"] == "module_path_attribute"),
            "{id}"
        );
        eprintln!("boundary {id}: documents/imports/calls refusal passed");
    }
    let (graph, _) = query(
        "pub mod user;\n/// documentation\npub mod service;\n",
        USER,
        SERVICE,
    );
    assert_eq!(docs(&graph), 1, "A-normal control");
}

#[test]
fn fence_blocks_exclude_following_prose_until_a_physical_block_boundary() {
    for (id, opening, body, closing) in [
        ("F-plain-tilde", "~~~rust", "`crate::service::parse`", "~~~"),
        ("F-plain-tick", "```rust", "`crate::service::parse`", "```"),
        (
            "F-quote-tilde",
            "> ~~~rust",
            "> `crate::service::parse`",
            "> ~~~",
        ),
        (
            "F-quote-tick",
            "> ```rust",
            "> `crate::service::parse`",
            "> ```",
        ),
        (
            "F-nested-tilde",
            "> > ~~~rust",
            "> > `crate::service::parse`",
            "> > ~~~",
        ),
        (
            "F-nested-tick",
            "> > ```rust",
            "> > `crate::service::parse`",
            "> > ```",
        ),
        (
            "F-list-tilde",
            "- ~~~rust",
            "  `crate::service::parse`",
            "  ~~~",
        ),
        (
            "F-list-tick",
            "- ```rust",
            "  `crate::service::parse`",
            "  ```",
        ),
        (
            "F-ordered",
            "1. ~~~rust",
            "   `crate::service::parse`",
            "   ~~~",
        ),
        (
            "F-ordered-tick",
            "1. ```rust",
            "   `crate::service::parse`",
            "   ```",
        ),
        (
            "F-list-quote",
            "- > ```rust",
            "  > `crate::service::parse`",
            "  > ```",
        ),
        (
            "F-list-quote-tilde",
            "- > ~~~rust",
            "  > `crate::service::parse`",
            "  > ~~~",
        ),
        (
            "F-quote-list",
            "> - ~~~rust",
            ">   `crate::service::parse`",
            ">   ~~~",
        ),
        (
            "F-quote-list-tick",
            "> - ```rust",
            ">   `crate::service::parse`",
            ">   ```",
        ),
        (
            "F-nested-list",
            "- - ```rust",
            "    `crate::service::parse`",
            "    ```",
        ),
        (
            "F-nested-list-tilde",
            "- - ~~~rust",
            "    `crate::service::parse`",
            "    ~~~",
        ),
    ] {
        let user = format!("//! {opening}\n//! {body}\n//! {closing}\n");
        let (graph, _) = query(LIB, &user, SERVICE);
        assert_eq!(docs(&graph), 0, "{id}");
        assert_eq!(doc_misses(&graph), 0, "{id}");
        let after = format!("{user}//! See `crate::service::parse`.\n");
        let (graph, _) = query(LIB, &after, SERVICE);
        assert_eq!(docs(&graph), 0, "{id}: the whole doc block is excluded");
        let separate = format!("{user}\n//! See `crate::service::parse`.\n");
        let (graph, _) = query(LIB, &separate, SERVICE);
        assert_eq!(docs(&graph), 1, "{id}: a separate plain block survives");
        assert!(
            !graph["edges"][0]["evidence_refs"]
                .as_array()
                .unwrap()
                .iter()
                .any(|e| e.as_str().unwrap().starts_with("file:src/user.rs#L2@")),
            "{id}"
        );
        eprintln!("boundary {id}: example exclusion + following-prose control passed");
    }
}

#[test]
fn fence_body_markers_and_decorated_block_indentation_are_not_reinterpreted() {
    for (id, user) in [
        (
            "F-extra-quote",
            "//! > ~~~\n//! > > ~~~\n//! > `crate::service::parse`\n//! > ~~~\n",
        ),
        (
            "F-extra-list",
            "//! - ```\n//!   - ```\n//!   `crate::service::parse`\n//!   ```\n",
        ),
        (
            "F-short-close",
            "//! > ~~~~\n//! > ~~~\n//! > `crate::service::parse`\n//! > ~~~~\n",
        ),
        (
            "F-wrong-marker",
            "//! > ```\n//! > ~~~\n//! > `crate::service::parse`\n//! > ```\n",
        ),
        (
            "F-block-indent",
            "/** Prose\n *\n *     // See `crate::service::parse`\n */\n",
        ),
        (
            "F-block-fence",
            "/**\n * > ~~~rust\n * > `crate::service::parse`\n * > ~~~\n */\n",
        ),
        (
            "F-block-fence-tick",
            "/**\n * > ```rust\n * > `crate::service::parse`\n * > ```\n */\n",
        ),
        (
            "F-plain-body-quote",
            "//! ~~~\n//! > ~~~\n//! `crate::service::parse`\n//! ~~~\n",
        ),
        (
            "F-plain-body-list",
            "//! ```\n//! - ```\n//! `crate::service::parse`\n//! ```\n",
        ),
        (
            "F-list-continuation",
            "//! - Topic\n//!     ~~~\n//!   `crate::service::parse`\n//!     ~~~\n",
        ),
    ] {
        let (graph, _) = query(LIB, user, SERVICE);
        assert_eq!(docs(&graph), 0, "{id}");
        assert_eq!(doc_misses(&graph), 0, "{id}");
        eprintln!("boundary {id}: no candidate/edge/doc-missing passed");
    }
}

#[test]
fn container_prose_is_explicitly_outside_the_narrow_contract() {
    for (id, user) in [
        ("P-quote", "//! > See `crate::service::parse`.\n"),
        ("P-nested-quote", "//! > > See `crate::service::parse`.\n"),
        ("P-list", "//! - See `crate::service::parse`.\n"),
        ("P-ordered", "//! 1. See `crate::service::parse`.\n"),
        ("P-list-quote", "//! - > See `crate::service::parse`.\n"),
    ] {
        let (graph, _) = query(LIB, user, SERVICE);
        assert_eq!(docs(&graph), 0, "{id}");
        assert_eq!(excluded_blocks(&graph), 1, "{id}");
        eprintln!("boundary {id}: container block exclusion passed");
    }
}

#[test]
fn attributes_between_doc_comments_delimit_separate_excluded_blocks() {
    for (id, user) in [
        ("F-attribute-between-docs", "/// > ~~~rust\n#[allow(dead_code)]\n/// > `crate::service::parse`\n/// > ~~~\npub fn user() {}\n"),
        ("F-attribute-between-docs-tick", "/// - ```rust\n#[allow(dead_code)]\n///   `crate::service::parse`\n///   ```\npub fn user() {}\n"),
    ] {
        let (graph, _) = query(LIB, user, SERVICE);
        eprintln!("D4 CLI fresh; false documents={}", docs(&graph));
        assert_eq!(docs(&graph), 0, "{id}");
        assert_eq!(doc_misses(&graph), 0, "{id}");
        eprintln!("boundary {id}: intervening attribute preserves doc fence");
    }
}

#[test]
fn auditor_eight_container_lifetime_literals() {
    let cases = [
        (
            "N1-quote-exit",
            "//! > ~~~rust\n//! > example\n//!\n//! See `crate::service::parse`.\n",
            1,
            Some(4),
        ),
        (
            "N1-list-dedent",
            "//! - ~~~rust\n//!   example\n//!\n//! See `crate::service::parse`.\n",
            1,
            Some(4),
        ),
        (
            "N2-quote-list-indent-leak",
            "//! > - Topic\n//!\n//!     `crate::service::parse`\n",
            0,
            None,
        ),
        (
            "C-N1-plain-prose",
            "//! See `crate::service::parse`.\n",
            1,
            Some(1),
        ),
        (
            "C-N1-closed-quote",
            "//! > ~~~rust\n//! > example\n//! > ~~~\n//! See `crate::service::parse`.\n",
            1,
            Some(4),
        ),
        (
            "C-N2-no-prior-list",
            "//! > Topic\n//!\n//!     `crate::service::parse`\n",
            0,
            None,
        ),
        (
            "C-N2-plain-indent",
            "//!     `crate::service::parse`\n",
            0,
            None,
        ),
        (
            "C-N2-valid-list-continuation",
            "//! - Topic\n//!     ~~~\n//!   `crate::service::parse`\n//!     ~~~\n",
            0,
            None,
        ),
    ];
    lifetime_ledger(
        "auditor-eight",
        cases
            .into_iter()
            .map(|(id, text, expected, line)| (id.to_string(), text.to_string(), expected, line))
            .collect(),
    );
}

fn lifetime_cell(
    id: &str,
    user: &str,
    expected: usize,
    line: Option<usize>,
    excluded: u64,
) -> Value {
    use sha2::{Digest, Sha256};
    let (graph, records) = query(LIB, user, SERVICE);
    let candidates = code_refs::scan(user)
        .refs
        .iter()
        .filter(|r| r.kind == "documents")
        .count();
    let stored = records["nodes"]
        .as_array()
        .unwrap()
        .iter()
        .find(|n| n["id"] == "code-scan:src/user.rs")
        .unwrap();
    let scan = code_refs::decode(stored["properties"]["scan"].as_str().unwrap()).unwrap();
    let stored_candidates = scan.refs.iter().filter(|r| r.kind == "documents").count();
    let user_sha = format!("sha256:{:x}", Sha256::digest(user.as_bytes()));
    let service_sha = format!("sha256:{:x}", Sha256::digest(SERVICE.as_bytes()));
    let expected_refs = line.map(|n| {
        json!([
            format!("file:src/service.rs#L1@{service_sha}"),
            format!("file:src/user.rs#L{n}@{user_sha}")
        ])
    });
    let evidence_matches = expected_refs.is_none_or(|want| {
        graph["edges"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|e| e["kind"] == "documents")
            .all(|e| {
                e["source_id"] == "file:src/user.rs"
                    && e["target_id"] == "file:src/service.rs"
                    && e["evidence_refs"] == want
            })
    });
    let actual = json!({"component_candidates":candidates,"stored_candidates":stored_candidates,"documents":docs(&graph),"stored_documents":docs(&records),"missing":doc_misses(&graph),"stored_missing":doc_misses(&records),"evidence_matches":evidence_matches,"excluded_blocks":excluded_blocks(&graph),"stored_excluded_blocks":excluded_blocks(&records)});
    let passed = candidates == expected
        && stored_candidates == expected
        && docs(&graph) == expected
        && docs(&records) == expected
        && doc_misses(&graph) == 0
        && doc_misses(&records) == 0
        && evidence_matches
        && excluded_blocks(&graph) == excluded
        && excluded_blocks(&records) == excluded;
    eprintln!("lifetime {id}: expected={expected} actual={actual} passed={passed}");
    json!({"id":id,"user":user,"user_sha256":user_sha,"expected":expected,"line":line,"actual":actual,"passed":passed})
}

fn lifetime_ledger(name: &str, cases: Vec<(String, String, usize, Option<usize>)>) {
    let rows: Vec<Value> = cases
        .iter()
        .map(|(id, user, _previous_want, line)| {
            // These 251 historical input bytes are unchanged. The user explicitly
            // replaced the old Markdown-lifetime oracle in decision 6801346885.
            let (want, excluded) = match id.as_str() {
                "C-N1-plain-prose" => (1, 0),
                "T-html-then-new-doc-group" | "T-tab-then-new-doc-group" => (1, 1),
                _ => (0, 1),
            };
            lifetime_cell(
                id,
                user,
                want,
                if want == 1 { *line } else { None },
                excluded,
            )
        })
        .collect();
    let failed: Vec<&Value> = rows.iter().filter(|r| r["passed"] == false).collect();
    if let Some(dir) = std::env::var_os("MOAI_C6D_LIFETIME_LEDGER") {
        let dir = Path::new(&dir);
        fs::create_dir_all(dir).unwrap();
        fs::write(
            dir.join(format!("{name}.json")),
            serde_json::to_vec_pretty(&rows).unwrap(),
        )
        .unwrap();
    }
    assert!(
        failed.is_empty(),
        "lifetime failures: {}",
        serde_json::to_string(&failed).unwrap()
    );
}

#[test]
fn legacy_container_transition_inputs_are_now_whole_block_gaps() {
    let mut cases = Vec::new();
    let mut add = |id: String, raw: String, expected: usize| {
        let user: String = raw
            .lines()
            .map(|line| {
                if line.is_empty() {
                    "//!\n".into()
                } else {
                    format!("//! {line}\n")
                }
            })
            .collect();
        let line = (expected == 1).then(|| {
            user.lines()
                .enumerate()
                .filter(|(_, l)| l.contains("crate::service::parse"))
                .map(|(n, _)| n + 1)
                .last()
                .unwrap()
        });
        cases.push((id, user, expected, line));
    };
    let contexts = [
        ("Q", "> ", "> "),
        ("L", "- ", "  "),
        ("QQ", "> > ", "> > "),
        ("LL", "- - ", "    "),
        ("QL", "> - ", ">   "),
        ("LQ", "- > ", "  > "),
    ];
    for (context, opening, continuation) in contexts {
        for (marker_name, marker) in [("tilde", "~~~"), ("tick", "```")] {
            for blank in [false, true] {
                let gap = if blank { "\n" } else { "" };
                let head =
                    format!("{opening}{marker}rust\n{continuation}example\n{gap}End outside.\n");
                for (tail, text, expected) in [
                    ("prose", "See `crate::service::parse`.\n", 1),
                    ("code", "\n    `crate::service::parse`\n", 0),
                    ("new-quote", "> See `crate::service::parse`.\n", 1),
                ] {
                    add(
                        format!("T-root-{context}-{marker_name}-b{}-{tail}", u8::from(blank)),
                        format!("{head}{text}"),
                        expected,
                    );
                }
            }
        }
        for blank in [false, true] {
            let gap = if blank { "\n" } else { "" };
            let expected = usize::from(!blank || matches!(context, "L" | "LL" | "LQ"));
            add(
                format!("T-indent-{context}-b{}", u8::from(blank)),
                format!("{opening}Topic\n{gap}    `crate::service::parse`\n"),
                expected,
            );
        }
    }
    for (context, opening, continuation, outer) in [
        ("QQ", "> > ", "> > ", "> "),
        ("LL", "- - ", "    ", "  "),
        ("QL", "> - ", ">   ", "> "),
        ("LQ", "- > ", "  > ", "  "),
    ] {
        for (marker_name, marker) in [("tilde", "~~~"), ("tick", "```")] {
            for blank in [false, true] {
                let gap = if blank {
                    format!("{outer}\n")
                } else {
                    String::new()
                };
                let head = format!(
                    "{opening}{marker}rust\n{continuation}example\n{gap}{outer}End outer.\n"
                );
                for (tail, text, expected) in [
                    ("prose", format!("{outer}See `crate::service::parse`.\n"), 1),
                    (
                        "code",
                        format!("{outer}\n{outer}    `crate::service::parse`\n"),
                        0,
                    ),
                    (
                        "new-quote",
                        format!("{outer}> See `crate::service::parse`.\n"),
                        1,
                    ),
                ] {
                    add(
                        format!(
                            "T-pop-child-{context}-{marker_name}-b{}-{tail}",
                            u8::from(blank)
                        ),
                        format!("{head}{text}"),
                        expected,
                    );
                }
            }
        }
    }
    for (name, marker) in [("tilde", "~~~"), ("tick", "```")] {
        add(
            format!("T-root-fence-stays-{name}"),
            format!("{marker}rust\n`crate::service::parse`\n"),
            0,
        );
    }
    add(
        "T-unsupported-tab".into(),
        "\tSee `crate::service::parse`.\n".into(),
        0,
    );
    add(
        "T-unsupported-html".into(),
        "<div>\nSee `crate::service::parse`.\n</div>\n".into(),
        0,
    );
    add(
        "T-unsupported-html-comment".into(),
        "<!--\nSee `crate::service::parse`.\n-->\n".into(),
        0,
    );
    cases.push(("T-html-then-new-doc-group".into(), "//! <div>\n//! opaque\npub fn user() {}\n/// See `crate::service::parse`.\npub fn other() {}\n".into(), 1, Some(4)));
    cases.push(("T-tab-then-new-doc-group".into(), "//! \tSee `crate::service::parse`.\npub fn user() {}\n/// See `crate::service::parse`.\npub fn other() {}\n".into(), 1, Some(3)));
    lifetime_ledger("transitions", cases);
}

#[test]
fn legacy_relative_fence_inputs_are_now_whole_block_gaps() {
    let mut cases = Vec::new();
    for (context, opening, outer) in [
        ("LQ", "- > ", "  "),
        ("LLQ", "- - > ", "    "),
        ("QLQ", "> - > ", ">   "),
        ("LQQ", "- > > ", "  "),
    ] {
        for indent in 0..4 {
            for (name, marker) in [("tilde", "~~~"), ("tick", "```")] {
                for blank in [false, true] {
                    let mut raw = vec![format!("{opening}Topic")];
                    if blank {
                        raw.push(outer.trim_end().to_string());
                    }
                    raw.extend([
                        format!("{outer}{}{marker}rust", " ".repeat(indent)),
                        format!("{outer}`crate::service::parse`"),
                        format!("{outer}{}{marker}", " ".repeat(indent)),
                    ]);
                    let user: String = raw
                        .into_iter()
                        .map(|line| {
                            if line.is_empty() {
                                "//!\n".into()
                            } else {
                                format!("//! {line}\n")
                            }
                        })
                        .collect();
                    cases.push((
                        format!("T-relative-{context}-i{indent}-{name}-b{}", u8::from(blank)),
                        user,
                        0,
                        None,
                    ));
                }
            }
        }
    }
    lifetime_ledger("relative-fences", cases);
}

#[test]
fn legacy_new_list_inputs_are_now_whole_block_gaps() {
    let mut cases = Vec::new();
    let mut add = |id: String, raw: String| {
        let user: String = raw
            .lines()
            .map(|line| {
                if line.is_empty() {
                    "//!\n".into()
                } else {
                    format!("//! {line}\n")
                }
            })
            .collect();
        let line = user.lines().count();
        cases.push((id, user, 1, Some(line)));
    };
    for (context, opening, continuation) in [
        ("Q", "> ", "> "),
        ("L", "- ", "  "),
        ("QQ", "> > ", "> > "),
        ("LL", "- - ", "    "),
        ("QL", "> - ", ">   "),
        ("LQ", "- > ", "  > "),
    ] {
        for (name, marker) in [("tilde", "~~~"), ("tick", "```")] {
            for blank in [false, true] {
                let gap = if blank { "\n" } else { "" };
                add(format!("T-new-list-root-{context}-{name}-b{}", u8::from(blank)), format!("{opening}{marker}rust\n{continuation}example\n{gap}End outside.\n- See `crate::service::parse`.\n"));
            }
        }
    }
    for (context, opening, continuation, outer) in [
        ("QQ", "> > ", "> > ", "> "),
        ("LL", "- - ", "    ", "  "),
        ("QL", "> - ", ">   ", "> "),
        ("LQ", "- > ", "  > ", "  "),
    ] {
        for (name, marker) in [("tilde", "~~~"), ("tick", "```")] {
            for blank in [false, true] {
                let gap = if blank {
                    format!("{outer}\n")
                } else {
                    String::new()
                };
                add(format!("T-new-list-child-{context}-{name}-b{}", u8::from(blank)), format!("{opening}{marker}rust\n{continuation}example\n{gap}{outer}End outer.\n{outer}- See `crate::service::parse`.\n"));
            }
        }
    }
    lifetime_ledger("new-lists", cases);
}

/// Decision 6801346885 supersedes the container-lifetime interpretation.
#[test]
fn narrow_contract_auditor_literals_exclude_the_entire_block() {
    let mut failures = Vec::new();
    for (id, user) in [
        (
            "D3",
            "//! > ~~~rust\n//! > `crate::service::parse`\n//! > ~~~\n",
        ),
        (
            "N1-quote",
            "//! > ~~~rust\n//! > example\n//!\n//! See `crate::service::parse`.\n",
        ),
        (
            "N1-list",
            "//! - ~~~rust\n//!   example\n//!\n//! See `crate::service::parse`.\n",
        ),
        (
            "N2",
            "//! > - Topic\n//!\n//!     `crate::service::parse`\n",
        ),
        ("empty-list", "//! -\n//!       `crate::service::parse`\n"),
        (
            "empty-quote-list",
            "//! > -\n//! >       `crate::service::parse`\n",
        ),
        (
            "empty-list-quote",
            "//! - >\n//!   >     `crate::service::parse`\n",
        ),
        (
            "empty-ordered",
            "//! 1.\n//!        `crate::service::parse`\n",
        ),
        ("empty-quote", "//! >\n//! >     `crate::service::parse`\n"),
        (
            "plain-before-container",
            "//! See `crate::service::parse`.\n//! > example\n",
        ),
        (
            "plain-before-indent",
            "//! See `crate::service::parse`.\n//!     example\n",
        ),
    ] {
        let (graph, records) = query(LIB, user, SERVICE);
        let candidates = code_refs::scan(user)
            .refs
            .iter()
            .filter(|r| r.kind == "documents")
            .count();
        let excluded = |g: &Value| {
            g["missing"]
                .as_array()
                .unwrap()
                .iter()
                .filter(|m| m["kind"] == "doc_reference_unsupported")
                .map(|m| m["count"].as_u64().unwrap_or(0))
                .sum::<u64>()
        };
        let actual = json!({"candidates":candidates,"documents":docs(&graph),"stored_documents":docs(&records),"excluded":excluded(&graph),"stored_excluded":excluded(&records)});
        let passed = candidates == 0
            && docs(&graph) == 0
            && docs(&records) == 0
            && excluded(&graph) == 1
            && excluded(&records) == 1;
        eprintln!("narrow {id}: {actual} passed={passed}");
        if !passed {
            failures.push((id, actual));
        }
    }
    assert!(
        failures.is_empty(),
        "narrow contract failures: {failures:?}"
    );
}

#[test]
fn narrow_physical_block_boundaries_and_syntax_matrix() {
    for (id, user, want, excluded) in [
        ("plain", "//! See `crate::service::parse`.\n", 1, 0),
        ("crlf-plain", "//! See `crate::service::parse`.\r\n", 1, 0),
        ("crlf-html-start", "//! <script\r\n//! See `crate::service::parse`.\r\n", 0, 1),
        ("plain-decorated-block", "/**\n * See `crate::service::parse`.\n */\n", 1, 0),
        ("first-line-star-list", "/** * See `crate::service::parse`. */\n", 0, 1),
        ("first-line-dash-list", "/** - See `crate::service::parse`. */\n", 0, 1),
        ("inline-html-is-prose", "//! See `crate::service::parse` (`<!-- marker -->`).\n", 1, 0),
        ("autolink-is-prose", "//! <https://example.test> See `crate::service::parse`.\n", 1, 0),
        ("not-a-list", "//! -not-a-list See `crate::service::parse`.\n", 1, 0),
        ("empty-plus-list", "//! +\n//! See `crate::service::parse`.\n", 0, 1),
        ("star-list", "//! * See `crate::service::parse`.\n", 0, 1),
        ("ordered-parenthesis", "//! 2) See `crate::service::parse`.\n", 0, 1),
        ("backtick-fence", "//! See `crate::service::parse`.\n//! ```rust\n", 0, 1),
        ("tilde-fence", "//! See `crate::service::parse`.\n//! ~~~rust\n", 0, 1),
        ("trailing-tab", "//! See `crate::service::parse`.\n//! prose\t\n", 0, 1),
        ("blank-indented-line", "//! See `crate::service::parse`.\n//!     \n", 0, 1),
        ("html-after-prose", "//! See `crate::service::parse`.\n//! <div>\n", 0, 1),
        ("html-comment", "//! <!--\n//! See `crate::service::parse`.\n", 0, 1),
        ("physical-blank-boundary", "//! > excluded\n\n//! See `crate::service::parse`.\n", 1, 1),
        ("doc-blank-is-same-block", "//! > excluded\n//!\n//! See `crate::service::parse`.\n", 0, 1),
        ("ordinary-comment-boundary", "//! > excluded\n// ordinary\n//! See `crate::service::parse`.\n", 1, 1),
        ("comment-before-doc-boundary", "//! > excluded\n/* ordinary */ //! See `crate::service::parse`.\n", 1, 1),
        ("attribute-boundary", "/// See `crate::service::parse`.\n#[allow(dead_code)]\n/// > excluded\npub fn user() {}\n", 1, 1),
        ("adjacent-block-comments", "/** > excluded */\n/** See `crate::service::parse`. */\n", 1, 1),
        ("block-then-line", "/** > excluded */\n//! See `crate::service::parse`.\n", 1, 1),
        ("two-excluded-blocks", "//! > excluded\npub fn user() {}\n/// - See `crate::service::parse`.\npub fn other() {}\n", 0, 2),
    ] {
        let (graph, stored) = query(LIB, user, SERVICE);
        assert_eq!(docs(&graph), want, "{id}");
        assert_eq!(docs(&stored), want, "{id}");
        assert_eq!(excluded_blocks(&graph), excluded, "{id}");
        assert_eq!(excluded_blocks(&stored), excluded, "{id}");
        assert_eq!(doc_misses(&graph), 0, "{id}");
        eprintln!("narrow boundary {id}: documents={want}, excluded={excluded}");
    }
}
