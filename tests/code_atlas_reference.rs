//! SPEC-20261117-atlas-in-read-plan: the READ and PLAN instruction for the code atlas
//! (`skills/moai/references/internal/code-atlas.md`) names commands and requests that really
//! work, says which freshness states allow its use, and never builds the atlas itself. Each
//! case runs the instruction's own requests against a temporary git repository.
mod common;

use common::CleanEnv;
use serde_json::Value;
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn reference() -> String {
    fs::read_to_string(package().join("references/internal/code-atlas.md")).unwrap()
}

/// The requests the instruction gives, with their placeholders; each must appear verbatim.
const VERSION: &str = r#"{"protocol_version":1}"#;
const GRAPH: &str = r#"{"protocol_version":1,"query":{"start":["file:<path>"]}}"#;
const MAP: &str = r#"{"protocol_version":1,"query":{"paths":["<file>"]}}"#;
const SEARCH: &str = r#"{"protocol_version":1,"query":{"literals":["<text>"]}}"#;

fn fixture() -> (tempfile::TempDir, PathBuf) {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("project");
    fs::create_dir_all(project.join("src")).unwrap();
    fs::write(
        project.join("src/a.rs"),
        "pub fn alpha() -> u8 {\n    7\n}\n",
    )
    .unwrap();
    fs::write(
        project.join("src/b.rs"),
        "use crate::a::alpha;\n\npub fn beta() -> u8 {\n    alpha() // ATLAS_PROBE_LITERAL\n}\n",
    )
    .unwrap();
    fs::write(project.join("src/lib.rs"), "pub mod a;\npub mod b;\n").unwrap();
    common::git::git_ok(&project, &["init", "-q", "--initial-branch=main"]);
    common::git::git_ok(&project, &["add", "."]);
    common::git::git_ok(
        &project,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            "init",
        ],
    );
    (tmp, project)
}

/// One command with `request` on stdin; returns the envelope's `result`.
fn run(home: &Path, project: &Path, command: &str, request: &str) -> Value {
    let mut child = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("HOME", home)
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .args([command, &project.to_string_lossy()])
        .current_dir(project)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(request.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(
        out.status.success(),
        "{command}: {text}{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_str(&text).unwrap_or_else(|e| panic!("{command}: {e}: {text}"));
    assert_eq!(v["ok"], true, "{command}: {v}");
    v["result"].clone()
}

fn index_dirs(project: &Path) -> Vec<PathBuf> {
    [".moai-cli/index", ".moai/index"]
        .iter()
        .map(|p| project.join(p))
        .filter(|p| p.exists())
        .collect()
}

#[test]
fn the_instruction_names_the_requests_the_states_and_the_no_build_rule() {
    let text = reference();
    assert!(text.starts_with("# Internal: code-atlas\n\nWhen to read: "));
    assert!(text.contains(&format!("`moai index-query <project>` with `{VERSION}`")));
    assert!(text.contains(&format!("`moai atlas-graph` with `{GRAPH}`")));
    assert!(text.contains(&format!("`moai atlas-map` with `{MAP}`")));
    assert!(text.contains(&format!("`moai search` with `{SEARCH}`")));
    for state in [
        "`fresh`",
        "`stale`",
        "`partial`",
        "`unavailable`",
        "branch_switched",
    ] {
        assert!(text.contains(state), "{state}");
    }
    assert!(text.contains("**Never build or rebuild it here.**"));
    assert!(text.contains("never build it silently"));
    assert!(text.contains("evidence_refs"));
    // Audit B1: a stale answer is incomplete for the changed files, dependents above all.
    let stale = stale_line(&text);
    for phrase in [
        "An answer that lists files — dependents above all — is incomplete",
        "drop the `deleted` files from it",
        "with `rg` for the path and the import name of what you asked about",
        // Re-audit N1: a grouped import names the module only as a bare word.
        "the bare module name as a whole word, `rg -w x`",
        "grouped imports such as `use crate::{a, x::y}`",
        "add each file that matches",
    ] {
        assert!(
            stale.contains(phrase),
            "the stale rule lacks {phrase:?}: {stale}"
        );
    }
    // Audit N1: `branch_switched` means HEAD moved, by a commit or a branch switch.
    assert!(text.contains("HEAD moved since the build (a commit or a branch switch)"));
    // Audit N2: no exact ratio from one repository.
    assert!(text.contains("two hops can be tens of times larger than one"));
    assert!(!text.contains("21 times"));
}

/// The `stale` line of step 1.
fn stale_line(text: &str) -> &str {
    text.lines()
        .find(|l| l.trim_start().starts_with("- `stale`:"))
        .expect("step 1 has a `stale` line")
}

/// The `changes` keys the stale rule says to search, in the order it names them.
fn keys_to_search(stale: &str) -> Vec<&str> {
    let from = stale
        .find("search every ")
        .expect("the stale rule searches");
    let to = stale[from..]
        .find(" with `rg`")
        .expect("the stale rule uses rg")
        + from;
    stale[from..to].split('`').skip(1).step_by(2).collect()
}

/// `rg -w`: `word` appears as a whole identifier word.
fn has_word(body: &str, word: &str) -> bool {
    body.split(|c: char| !(c.is_ascii_alphanumeric() || c == '_'))
        .any(|w| w == word)
}

/// Audit B1: files added after the build depend on `src/a.rs` — one with a plain import, one
/// with a grouped import (re-audit N1). The atlas, being stale, lists neither; the stale rule's
/// own procedure — search the `changes` files the rule names for the path and the bare module
/// name as a whole word, then merge — finds both.
#[test]
fn a_dependent_added_after_the_build_is_found_by_the_stale_rule() {
    let (tmp, project) = fixture();
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    assert_eq!(
        run(&home, &project, "index-build", VERSION)["status"],
        "complete"
    );
    fs::write(
        project.join("src/c.rs"),
        "use crate::a::alpha;\n\npub fn gamma() -> u8 {\n    alpha() + 1\n}\n",
    )
    .unwrap();
    // A grouped import: no `crate::a` anywhere in the file.
    let grouped =
        "use crate::{a::alpha, b::beta};\n\npub fn delta() -> u8 {\n    alpha() + beta()\n}\n";
    assert!(!grouped.contains("crate::a"));
    fs::write(project.join("src/d.rs"), grouped).unwrap();
    fs::write(
        project.join("src/lib.rs"),
        "pub mod a;\npub mod b;\npub mod c;\npub mod d;\n",
    )
    .unwrap();
    let q = run(&home, &project, "index-query", VERSION);
    assert_eq!(q["state"], "stale", "{q}");
    assert_eq!(q["changes"]["branch_switched"], false, "{q}");

    // What the atlas alone says depends on src/a.rs.
    let g = run(
        &home,
        &project,
        "atlas-graph",
        &GRAPH.replace("<path>", "src/a.rs"),
    );
    let mut dependents: Vec<String> = g["graph"]["edges"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["target_id"] == "file:src/a.rs")
        .filter_map(|e| e["source_id"].as_str()?.strip_prefix("file:"))
        .map(str::to_string)
        .collect();
    assert!(dependents.contains(&"src/b.rs".to_string()), "{g}");
    for new in ["src/c.rs", "src/d.rs"] {
        assert!(
            !dependents.contains(&new.to_string()),
            "the stale atlas cannot know {new}: {g}"
        );
    }

    // The rule as written: drop `deleted`, search the files of the keys it names.
    let text = reference();
    let stale = stale_line(&text);
    let keys = keys_to_search(stale);
    assert_eq!(keys, ["added", "modified", "new_files"], "{stale}");
    let deleted: Vec<&str> = q["changes"]["deleted"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(Value::as_str)
        .collect();
    dependents.retain(|d| !deleted.contains(&d.as_str()));
    for key in keys {
        for path in q["changes"][key].as_array().unwrap() {
            let path = path.as_str().unwrap();
            let body = fs::read_to_string(project.join(path)).unwrap_or_default();
            if (body.contains("src/a.rs") || has_word(&body, "a"))
                && !dependents.iter().any(|d| d == path)
            {
                dependents.push(path.to_string());
            }
        }
    }
    for new in ["src/c.rs", "src/d.rs"] {
        assert!(
            dependents.contains(&new.to_string()),
            "the stale rule missed the new dependent {new}: {dependents:?} ({q})"
        );
    }
}

#[test]
fn read_and_plan_point_to_the_instruction_in_one_line_and_project_offers_the_build() {
    for workflow in ["read", "plan"] {
        let text = fs::read_to_string(package().join(format!("workflows/{workflow}.md"))).unwrap();
        let lines = text
            .lines()
            .filter(|l| l.contains("internal/code-atlas"))
            .count();
        assert_eq!(lines, 1, "{workflow}.md");
        assert!(
            !text.contains("index-build"),
            "{workflow}.md must not build"
        );
    }
    let project = fs::read_to_string(package().join("workflows/project.md")).unwrap();
    assert!(project.contains("**Offer the code map.**"));
    assert!(project.contains("run it only after the user agrees"));
}

/// The instruction's own requests, run in its order: before a build every read answer is
/// `unavailable` and nothing is written; after a build the answers are `fresh` and carry what
/// the instruction relies on; after an edit the change shows as `stale` with the file named.
#[test]
fn the_requests_work_and_the_states_follow_the_tree() {
    let (tmp, project) = fixture();
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    let graph = GRAPH.replace("<path>", "src/a.rs");
    let map = MAP.replace("<file>", "src/b.rs");
    let search = SEARCH.replace("<text>", "ATLAS_PROBE_LITERAL");

    let q = run(&home, &project, "index-query", VERSION);
    assert_eq!(q["state"], "unavailable", "{q}");
    assert_eq!(
        run(&home, &project, "atlas-map", &map)["state"],
        "unavailable"
    );
    assert_eq!(
        run(&home, &project, "atlas-graph", &graph)["state"],
        "unavailable"
    );
    assert!(
        index_dirs(&project).is_empty(),
        "a read built the atlas: {:?}",
        index_dirs(&project)
    );

    let b = run(&home, &project, "index-build", VERSION);
    assert_eq!(b["status"], "complete", "{b}");
    let q = run(&home, &project, "index-query", VERSION);
    assert_eq!(q["state"], "fresh", "{q}");
    let generation = q["generation_id"].clone();
    assert!(generation.is_string(), "{q}");

    let g = run(&home, &project, "atlas-graph", &graph);
    assert_eq!(g["state"], "fresh", "{g}");
    assert_eq!(g["generation_id"], generation);
    assert!(g["coverage"]["symbols"]
        .as_str()
        .unwrap()
        .starts_with("unavailable"));
    let edges = g["graph"]["edges"].as_array().unwrap();
    assert!(
        edges.iter().any(|e| e["target_id"] == "file:src/a.rs"
            && e["source_id"] == "file:src/b.rs"
            && e["evidence_refs"].as_array().is_some_and(|r| !r.is_empty())),
        "{g}"
    );
    let m = run(&home, &project, "atlas-map", &map);
    assert_eq!(m["state"], "fresh", "{m}");
    assert_eq!(m["map"]["nodes"][0]["id"], "file:src/b.rs", "{m}");
    let s = run(&home, &project, "search", &search);
    let hits = s["results"].as_array().unwrap();
    assert!(
        hits.iter()
            .any(|h| h["path"] == "src/b.rs" && h["start_line"] == 4 && h["state"] == "fresh"),
        "{s}"
    );

    fs::write(
        project.join("src/a.rs"),
        "pub fn alpha() -> u8 {\n    8\n}\n",
    )
    .unwrap();
    let q = run(&home, &project, "index-query", VERSION);
    assert_eq!(q["state"], "stale", "{q}");
    assert_eq!(
        q["changes"]["modified"],
        serde_json::json!(["src/a.rs"]),
        "{q}"
    );
    assert_eq!(q["changes"]["branch_switched"], false, "{q}");
    assert_eq!(run(&home, &project, "atlas-map", &map)["state"], "stale");
}
