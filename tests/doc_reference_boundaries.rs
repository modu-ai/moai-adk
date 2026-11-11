//! #2450 repair: literal auditor fixtures and adjacent lexical boundaries through CLI.
mod common;
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
        .env_clear()
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
    run(&root, &home, "index-build", json!({"protocol_version":1}));
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
fn fence_container_matrix_excludes_examples_and_keeps_following_prose() {
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
        assert_eq!(docs(&graph), 1, "{id}: following prose survives");
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
fn container_prose_is_not_discarded_to_hide_fence_gaps() {
    for (id, user) in [
        ("P-quote", "//! > See `crate::service::parse`.\n"),
        ("P-nested-quote", "//! > > See `crate::service::parse`.\n"),
        ("P-list", "//! - See `crate::service::parse`.\n"),
        ("P-ordered", "//! 1. See `crate::service::parse`.\n"),
        ("P-list-quote", "//! - > See `crate::service::parse`.\n"),
    ] {
        let (graph, _) = query(LIB, user, SERVICE);
        assert_eq!(docs(&graph), 1, "{id}");
        eprintln!("boundary {id}: real prose reference passed");
    }
}

#[test]
fn attributes_between_doc_comments_do_not_end_a_fenced_example() {
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
