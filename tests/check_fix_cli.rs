//! SPEC-20261125-check-fast-and-fix: `check fix <base>..<head>` writes the table lines and hashes
//! that need no human choice, prints the lines that do, and writes nothing through a link.
//! Temporary git repositories only: the real tables, the real HOME and the real cargo are never
//! used.
mod common;
use common::git::{git, git_ok};
use moai_cli::check_affected::{self, Pick};
use moai_cli::check_fix::{self, Options};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::fs;
use std::path::Path;

const REGISTRY: &str = "tools/repo_scans.json";
const AREAS: &str = "tests/areas.toml";
const ALLOW: &str = "tests/fixtures/state_folder_literals/allow.tsv";
const MANIFEST: &str = "skills/moai/references/source-manifest.json";

fn put(root: &Path, rel: &str, text: &[u8]) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

fn text(root: &Path, rel: &str) -> String {
    fs::read_to_string(root.join(rel)).unwrap()
}

fn sha(bytes: &[u8]) -> String {
    Sha256::digest(bytes)
        .iter()
        .map(|b| format!("{b:02x}"))
        .collect()
}

fn init(root: &Path) {
    git_ok(root, &["init", "-q", "-b", "main"]);
    git_ok(root, &["config", "user.email", "t@example.invalid"]);
    git_ok(root, &["config", "user.name", "t"]);
}

/// Commit everything in the working tree and answer the new head.
fn commit_all(root: &Path, message: &str) -> String {
    git_ok(root, &["add", "-A"]);
    git_ok(root, &["commit", "-q", "-m", message]);
    git(root, &["rev-parse", "HEAD"])
}

/// The earlier state-folder name, spelled at run time so that no source line of this test file
/// holds it.
fn earlier_name() -> String {
    format!(".{}-{}", "moai", "cli")
}

/// The scan registry as the tool's own table writes it: pretty JSON, one string per line. One
/// `always` scan lists `rule` as a rule file and runs only `--test card_other`.
fn registry(rule: &str) -> String {
    let doc = json!({
        "about": "fixture",
        "scans": [{
            "id": "zz-rule",
            "always": true,
            "tests": [format!("{rule}::t")],
            "checks": ["cargo test --offline --locked --test card_other"],
            "rules": [rule],
            "scope": [{"prefix": "src/", "suffixes": [".rs"]}],
            "rejects": "fixture",
        }],
    });
    serde_json::to_string_pretty(&doc).unwrap() + "\n"
}

/// A manifest of the files under `skills/moai/`, each entry holding the given sha256 text.
fn manifest(entries: &[(&str, &str)]) -> String {
    let mut files = serde_json::Map::new();
    for (rel, hash) in entries {
        files.insert(
            rel.to_string(),
            json!({"origin": "fixture", "sha256": hash}),
        );
    }
    let doc = json!({"schema_version": 1, "covers": "fixture", "files": files});
    serde_json::to_string_pretty(&doc).unwrap() + "\n"
}

const AREA_TABLE: &str =
    "[areas]\ncard = [\"card\"]\nrepo = [\"repo\", \"atlas\"]\n\n[solo]\nfiles = []\n";

/// The files every fixture starts from: the area table, an empty allow list and a README.
fn base_tables(root: &Path, areas: &str) {
    put(root, "README.md", b"fixture\n");
    put(root, AREAS, areas.as_bytes());
    put(root, ALLOW, b"# path\tearlier\tnew\n");
}

fn printed(answer: &Value) -> Vec<String> {
    answer["print"]
        .as_array()
        .unwrap()
        .iter()
        .map(|l| l.as_str().unwrap().to_string())
        .collect()
}

fn wrote(answer: &Value) -> Vec<String> {
    answer["wrote"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| w["path"].as_str().unwrap().to_string())
        .collect()
}

fn area_opts(pairs: &[(&str, &str)]) -> Options {
    Options {
        areas: pairs
            .iter()
            .map(|(w, a)| (w.to_string(), a.to_string()))
            .collect(),
        integration_ref: None,
    }
}

/// AC-4: the mechanical facts are written (the word of a new test file into the area named by
/// `--area`, the sha256 of the manifest, a `--test` option of an `always` scan that the registry
/// reads as growth); the lines that need a human choice are printed and not written; the fixed
/// tree, once the choices are made, has nothing left to do.
#[test]
fn check_fix_writes_what_is_mechanical_and_prints_the_rest() {
    let dir = tempfile::tempdir().unwrap();
    let root = dir.path();
    init(root);
    base_tables(root, AREA_TABLE);
    put(root, REGISTRY, registry("tests/zz_new_cli.rs").as_bytes());
    put(root, "skills/moai/references/a.md", b"alpha\n");
    put(
        root,
        MANIFEST,
        manifest(&[("references/a.md", &"0".repeat(64))]).as_bytes(),
    );
    let base = commit_all(root, "base");
    let spelled = format!("// the layout of {}\nfn t() {{}}\n", earlier_name());
    put(root, "tests/zz_new_cli.rs", spelled.as_bytes());
    put(root, "tests/qq_other.rs", b"fn t() {}\n");
    put(root, "skills/moai/references/new.md", b"new\n");
    let head = commit_all(root, "change");
    let range = format!("{base}..{head}");
    let old_registry = text(root, REGISTRY);
    let old_allow = text(root, ALLOW);

    let answer = check_fix::run(root, &range, &area_opts(&[("zz", "repo")])).unwrap();

    assert_eq!(answer["action"], "fix", "{answer}");
    assert_eq!(answer["writes"], true, "{answer}");
    assert_eq!(answer["result"], "FAIL", "{answer}");
    assert_eq!(answer["nothing_to_do"], false, "{answer}");
    let mut written = wrote(&answer);
    written.sort();
    let mut want = vec![AREAS, REGISTRY, MANIFEST];
    want.sort();
    assert_eq!(written, want, "{answer}");
    // The word is appended at the end of the one-line array; the other lines stay as they were.
    assert_eq!(
        text(root, AREAS),
        AREA_TABLE.replace("\"atlas\"]", "\"atlas\", \"zz\"]")
    );
    // The hash is the file's own; no entry was added for the new file.
    let fixed: Value = serde_json::from_str(&text(root, MANIFEST)).unwrap();
    assert_eq!(
        fixed["files"]["references/a.md"]["sha256"],
        sha(b"alpha\n"),
        "{fixed}"
    );
    assert_eq!(fixed["files"].as_object().unwrap().len(), 1, "{fixed}");
    // The registry gained `--test zz_new_cli` on the existing line, and reads as growth.
    let new_registry = text(root, REGISTRY);
    assert!(
        new_registry.contains("--test card_other --test zz_new_cli\""),
        "{new_registry}"
    );
    assert!(moai_cli::gate::registry_only_grew(
        REGISTRY,
        &old_registry,
        &new_registry
    ));
    // Only the registry differs from the old text by the one added option.
    assert_eq!(
        new_registry.replace(" --test zz_new_cli", ""),
        old_registry,
        "the edit is in place and nothing else changed"
    );
    // The choices of a human are printed as the missing line and not written.
    let lines = printed(&answer);
    let row = "tests/zz_new_cli.rs\t1\t0";
    assert!(lines.iter().any(|l| l.contains(row)), "{lines:?}");
    assert!(
        lines
            .iter()
            .any(|l| l.contains("tests/qq_other.rs") && l.contains("`qq`") && l.contains("repo")),
        "an unknown first word with no --area is printed with the area names: {lines:?}"
    );
    assert!(
        lines
            .iter()
            .any(|l| l.contains("references/new.md") && l.contains("origin")),
        "{lines:?}"
    );
    assert!(
        !lines.iter().any(|l| l.contains("`zz`")),
        "the word that --area named is written, not printed: {lines:?}"
    );
    assert_eq!(
        text(root, ALLOW),
        old_allow,
        "an allow-list row is never written"
    );
    assert!(!text(root, AREAS).contains("qq"), "no area is guessed");
    // The tool's own writes leave a modified tree: a second run before a commit is refused.
    let refused = check_fix::run(root, &range, &area_opts(&[])).unwrap_err();
    assert!(refused.contains("refused"), "{refused}");

    // The lane commits the tool's writes, makes the choices that were printed, commits again.
    commit_all(root, "fix");
    put(
        root,
        ALLOW,
        format!("{old_allow}tests/zz_new_cli.rs\t1\t0\n").as_bytes(),
    );
    put(
        root,
        AREAS,
        text(root, AREAS)
            .replace("\"card\"]", "\"card\", \"qq\"]")
            .as_bytes(),
    );
    let mut with_entry: Value = serde_json::from_str(&text(root, MANIFEST)).unwrap();
    with_entry["files"]["references/new.md"] =
        json!({"origin": "fixture", "sha256": sha(b"new\n")});
    put(
        root,
        MANIFEST,
        (serde_json::to_string_pretty(&with_entry).unwrap() + "\n").as_bytes(),
    );
    let fixed_head = commit_all(root, "choices");

    let again = check_fix::run(root, &format!("{base}..{fixed_head}"), &area_opts(&[])).unwrap();
    assert_eq!(again["result"], "PASS", "{again}");
    assert_eq!(again["wrote"], json!([]), "{again}");
    assert_eq!(again["print"], json!([]), "{again}");
    assert_eq!(again["nothing_to_do"], true, "{again}");
}

/// SPEC-20261125-repo-scans-convert-test-targets: on a `reads-` scan that keeps its targets in
/// `test_targets` and has a filtered area line, the fix writes the new name into the list. The
/// check line stays as it was and `check affected --list` selects every target it selected before.
#[test]
fn check_fix_adds_to_test_targets_and_the_selection_does_not_shrink() {
    const LINE: &str = "cargo test --offline --locked --test card mod_a:: -- mod_b::";
    let dir = tempfile::tempdir().unwrap();
    let root = dir.path();
    init(root);
    put(root, "README.md", b"fixture\n");
    put(root, "Cargo.toml", b"[package]\nname = \"fixture\"\n");
    put(root, "src/zz_stub.rs", b"// a\n");
    put(root, "tests/card/main.rs", b"mod mod_a;\nmod mod_b;\n");
    for file in ["tests/card/mod_a.rs", "tests/card/mod_b.rs"] {
        put(root, file, b"#[test]\nfn t() {}\n");
    }
    for file in ["tests/zz_base_cli.rs", "tests/zz_zeta_cli.rs"] {
        put(root, file, b"fn t() {}\n");
    }
    let doc = json!({
        "about": "fixture",
        "scans": [{
            "id": "reads-zz",
            "always": true,
            "tests": ["tests/zz_new_cli.rs::t"],
            "test_targets": ["zz_base_cli", "zz_zeta_cli"],
            "checks": [LINE],
            "rules": ["tests/zz_new_cli.rs"],
            "scope": [{"prefix": "src/", "suffixes": [".rs"]}],
            "rejects": "fixture",
        }],
    });
    let old_registry = serde_json::to_string_pretty(&doc).unwrap() + "\n";
    put(root, REGISTRY, old_registry.as_bytes());
    let base = commit_all(root, "base");
    put(root, "src/zz_stub.rs", b"// b\n");
    put(root, "tests/zz_new_cli.rs", b"fn t() {}\n");
    let head = commit_all(root, "change");
    let selected = |range: &str| -> Vec<String> {
        let answer = check_affected::list(root, range, Pick::Full, None).unwrap();
        answer["rust_targets"]
            .as_array()
            .unwrap()
            .iter()
            .map(|t| t.as_str().unwrap().to_string())
            .collect()
    };
    let before = selected(&format!("{base}..{head}"));
    for want in ["mod_a", "mod_b", "zz_base_cli", "zz_zeta_cli"] {
        assert!(before.iter().any(|t| t == want), "{want} in {before:?}");
    }

    let answer = check_fix::run(root, &format!("{base}..{head}"), &area_opts(&[])).unwrap();

    assert_eq!(wrote(&answer), vec![REGISTRY], "{answer}");
    let new_registry = text(root, REGISTRY);
    let fixed = commit_all(root, "fix");
    let after = selected(&format!("{base}..{fixed}"));
    let lost: Vec<&String> = before.iter().filter(|t| !after.contains(t)).collect();
    assert!(lost.is_empty(), "lost {lost:?}: {before:?} -> {after:?}");
    assert!(after.iter().any(|t| t == "zz_new_cli"), "{after:?}");
    assert!(
        answer.to_string().contains("to `test_targets` of the scan"),
        "{answer}"
    );
    assert_eq!(
        new_registry,
        old_registry.replace(
            "\"zz_zeta_cli\"",
            "\"zz_new_cli\",\n        \"zz_zeta_cli\""
        ),
        "one new line at its place; the check line is as it was"
    );
    assert!(new_registry.contains(LINE), "{new_registry}");
    assert!(moai_cli::gate::registry_only_grew(
        REGISTRY,
        &old_registry,
        &new_registry
    ));
}

/// REQ-13: a registry edit that the registry's own judge would not read as growth is printed and
/// not written. A name with a capital letter is no target name.
#[test]
fn check_fix_prints_a_registry_edit_that_would_not_read_as_growth() {
    let dir = tempfile::tempdir().unwrap();
    let root = dir.path();
    init(root);
    base_tables(
        root,
        "[areas]\ncard = [\"card\"]\nrepo = [\"repo\", \"zz\"]\n\n[solo]\nfiles = []\n",
    );
    put(root, REGISTRY, registry("tests/zz_Bad_cli.rs").as_bytes());
    let base = commit_all(root, "base");
    put(root, "tests/zz_Bad_cli.rs", b"fn t() {}\n");
    let head = commit_all(root, "change");
    let before = text(root, REGISTRY);

    let answer = check_fix::run(root, &format!("{base}..{head}"), &area_opts(&[])).unwrap();

    assert_eq!(answer["result"], "FAIL", "{answer}");
    assert_eq!(answer["wrote"], json!([]), "{answer}");
    assert_eq!(answer["nothing_to_do"], false, "{answer}");
    let lines = printed(&answer);
    assert!(
        lines
            .iter()
            .any(|l| l.contains("zz-rule") && l.contains("--test zz_Bad_cli")),
        "{lines:?}"
    );
    assert_eq!(text(root, REGISTRY), before, "nothing was written");
}

/// An `always` scan without `test_targets` whose check lines all filter by name is not written:
/// the lines stay and the answer prints what a person adds, and where. A scan with a line that has
/// no filter still gets the name on that line.
#[test]
fn check_fix_leaves_filtered_lines_and_prints_where_to_add() {
    // The two filtered lines are those of the scans `include-str-release-suite` and
    // `test-seams-read-only-in-debug-builds` of this repository's table.
    const MODULE: &str = "cargo test --offline --locked --test card release_suite::";
    const NAMED: &str = "cargo test --offline --locked --test state_folder_literal_scan the_test_seams_are_read_only_in_debug_builds";
    const PLAIN: &str = "cargo test --offline --locked --test card_other";
    const ADDED: &str = "cargo test --offline --locked --test zz_new_cli";
    let dir = tempfile::tempdir().unwrap();
    let root = dir.path();
    init(root);
    base_tables(
        root,
        "[areas]\ncard = [\"card\"]\nrepo = [\"repo\", \"zz\"]\n\n[solo]\nfiles = []\n",
    );
    let scan = |id: &str, check: &str| {
        json!({
            "id": id,
            "always": true,
            "tests": ["tests/zz_new_cli.rs::t"],
            "checks": [check],
            "rules": ["tests/zz_new_cli.rs"],
            "scope": [{"prefix": "src/", "suffixes": [".rs"]}],
            "rejects": "fixture",
        })
    };
    let mut doc = json!({
        "about": "fixture",
        "scans": [
            scan("zz-module", MODULE),
            scan("zz-named", NAMED),
            scan("zz-plain", PLAIN),
        ],
    });
    let old_registry = serde_json::to_string_pretty(&doc).unwrap() + "\n";
    put(root, REGISTRY, old_registry.as_bytes());
    let base = commit_all(root, "base");
    put(root, "tests/zz_new_cli.rs", b"fn t() {}\n");
    let head = commit_all(root, "change");

    let answer = check_fix::run(root, &format!("{base}..{head}"), &area_opts(&[])).unwrap();

    assert_eq!(answer["result"], "FAIL", "{answer}");
    assert_eq!(wrote(&answer), vec![REGISTRY], "{answer}");
    assert_eq!(
        text(root, REGISTRY),
        old_registry.replace(PLAIN, &format!("{PLAIN} --test zz_new_cli")),
        "only the line without a filter gained the name"
    );
    let lines = printed(&answer);
    assert_eq!(
        lines.len(),
        2,
        "one line for each scan left alone: {lines:?}"
    );
    for id in ["zz-module", "zz-named"] {
        let line = lines
            .iter()
            .find(|l| l.starts_with(&format!("{id}: ")))
            .unwrap_or_else(|| panic!("no line for {id}: {lines:?}"));
        for part in ["wrote nothing", ADDED, REGISTRY] {
            assert!(line.contains(part), "{part} in {line}");
        }
        assert!(!line.contains("test_targets"), "{line}");
    }

    // A person adds the printed line to both scans: nothing is left to print.
    for at in [0, 1] {
        doc["scans"][at]["checks"]
            .as_array_mut()
            .unwrap()
            .push(json!(ADDED));
    }
    doc["scans"][2]["checks"][0] = json!(format!("{PLAIN} --test zz_new_cli"));
    put(
        root,
        REGISTRY,
        (serde_json::to_string_pretty(&doc).unwrap() + "\n").as_bytes(),
    );
    let by_hand = commit_all(root, "by hand");
    let again = check_fix::run(root, &format!("{base}..{by_hand}"), &area_opts(&[])).unwrap();
    assert_eq!(again["result"], "PASS", "{again}");
    assert_eq!(again["wrote"], json!([]), "{again}");
}

#[cfg(unix)]
fn link(target: &str, at: &Path) {
    std::os::unix::fs::symlink(target, at).unwrap();
}

/// REQ-10: the registry is a link, or lies under a linked folder: nothing is written there, the
/// path is named and the answer is FAIL. The link itself is committed, so the checkout is clean.
#[cfg(unix)]
#[test]
fn check_fix_refuses_a_linked_table() {
    for linked_folder in [false, true] {
        let dir = tempfile::tempdir().unwrap();
        let root = dir.path();
        init(root);
        base_tables(root, AREA_TABLE);
        let real = if linked_folder {
            "real_tools/repo_scans.json"
        } else {
            "real/repo_scans.json"
        };
        put(root, real, registry("tests/zz_new_cli.rs").as_bytes());
        if linked_folder {
            link("real_tools", &root.join("tools"));
        } else {
            fs::create_dir_all(root.join("tools")).unwrap();
            link("../real/repo_scans.json", &root.join(REGISTRY));
        }
        // The word is already listed, so the registry is the one thing there is to write.
        put(
            root,
            AREAS,
            AREA_TABLE
                .replace("\"atlas\"]", "\"atlas\", \"zz\"]")
                .as_bytes(),
        );
        let base = commit_all(root, "base");
        put(root, "tests/zz_new_cli.rs", b"fn t() {}\n");
        let head = commit_all(root, "change");
        let before = text(root, real);

        let answer = check_fix::run(root, &format!("{base}..{head}"), &area_opts(&[])).unwrap();

        let case = format!("linked_folder={linked_folder}: {answer}");
        assert_eq!(answer["result"], "FAIL", "{case}");
        assert_eq!(answer["wrote"], json!([]), "{case}");
        assert!(
            printed(&answer)
                .iter()
                .any(|l| l.contains(REGISTRY) && l.contains("is a link")),
            "the path is named and the link is the reason: {case}"
        );
        assert_eq!(text(root, real), before, "{case}");
        let linked = if linked_folder { "tools" } else { REGISTRY };
        assert!(
            fs::symlink_metadata(root.join(linked))
                .unwrap()
                .file_type()
                .is_symlink(),
            "the link is still a link: {case}"
        );
    }
}
