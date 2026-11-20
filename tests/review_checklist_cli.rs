//! `review-package` checklist (SPEC-20261120-review-checklist-by-diff): the package carries, in
//! a `checklist` key, the items of `tools/review_checklist.json` whose trigger the diff meets.
//! The table is read from the BASE commit only, so a pull request that edits it is still judged
//! by the table it started from. Temporary repositories only; the binary reads git and writes
//! nothing. The repository's own table is checked by `real_file`, `moved_rules`,
//! `every_item_has_cases` and `protected_list`.

mod common;

use common::git::{command, git, git_ok};
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    collections::BTreeSet,
    fs,
    path::{Path, PathBuf},
    process::Output,
};
use tempfile::TempDir;

const TABLE: &str = "tools/review_checklist.json";
const BYTE_LIMIT: usize = 150_000;

fn write(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

fn commit(repo: &Path, message: &str) -> String {
    git_ok(repo, &["add", "-A"]);
    let mut cmd = command();
    cmd.args([
        "-c",
        "user.name=Fixture",
        "-c",
        "user.email=fixture@example.invalid",
        "commit",
        "-q",
        "--allow-empty",
        "-m",
        message,
    ])
    .current_dir(repo)
    .env("GIT_AUTHOR_DATE", "2026-01-01T00:00:00 +0000")
    .env("GIT_COMMITTER_DATE", "2026-01-01T00:00:00 +0000");
    assert!(cmd.status().unwrap().success());
    git(repo, &["rev-parse", "HEAD"])
}

struct Fx {
    _tmp: TempDir,
    repo: PathBuf,
    base: String,
    head: String,
}

/// A repository whose base commit holds `base` and whose head commit adds or overwrites `head`
/// and deletes `removed`.
fn fixture(base: &[(&str, &str)], head: &[(&str, &str)], removed: &[&str]) -> Fx {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    fs::create_dir_all(&repo).unwrap();
    git_ok(&repo, &["init", "-q", "-b", "main"]);
    write(&repo, "README.md", "readme\n");
    for (path, text) in base {
        write(&repo, path, text);
    }
    let base_commit = commit(&repo, "base");
    for (path, text) in head {
        write(&repo, path, text);
    }
    for path in removed {
        fs::remove_file(repo.join(path)).unwrap();
    }
    let head_commit = commit(&repo, "head");
    Fx {
        _tmp: tmp,
        repo,
        base: base_commit,
        head: head_commit,
    }
}

/// The binary runs through a temporary HOME with a clean environment, never the real one.
fn run(fx: &Fx) -> Output {
    let home = TempHome::new();
    home.command(&[
        "review-package",
        fx.repo.to_str().unwrap(),
        "--base",
        &fx.base,
        "--head",
        &fx.head,
    ])
    .output()
    .unwrap()
}

fn package(fx: &Fx) -> Value {
    let out = run(fx);
    assert!(
        out.status.success(),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn ids(result: &Value) -> Vec<String> {
    result["checklist"]
        .as_array()
        .expect("the package has a `checklist` list")
        .iter()
        .map(|e| e["id"].as_str().unwrap().to_string())
        .collect()
}

fn table_skips(result: &Value) -> Vec<String> {
    result["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["path"] == json!(TABLE))
        .map(|e| e["reason"].as_str().unwrap().to_string())
        .collect()
}

fn item(id: &str, paths: Value, patterns: Option<&[&str]>) -> Value {
    let mut trigger = json!({ "paths": paths });
    if let Some(p) = patterns {
        trigger["diff_pattern"] = json!(p);
    }
    json!({
        "id": id,
        "trigger": trigger,
        "check": format!("check {id}"),
        "source": "Issue #1",
        "added": "2026-11-20",
    })
}

fn table(items: &[Value]) -> String {
    serde_json::to_string_pretty(&json!({ "schema_version": 1, "items": items })).unwrap()
}

fn prefix(p: &str) -> Value {
    json!([{ "prefix": p }])
}

// --- base only -----------------------------------------------------------------------------

/// SPEC REQ-1, REQ-2, REQ-12: only the base commit's table counts; a table in the head, or a
/// head edit of the base table, changes nothing; no table at base is an empty list, no skip.
#[test]
fn base_only() {
    let at_base = table(&[
        item("first", prefix("src/"), None),
        item("second", prefix("docs/"), None),
        item("third", prefix(""), None),
    ]);
    // The head edits the table: it would drop `first` and add `injected`.
    let in_head = table(&[item("injected", prefix(""), None)]);
    let fx = fixture(
        &[(TABLE, &at_base), ("src/a.rs", "a\n")],
        &[(TABLE, &in_head), ("src/a.rs", "b\n")],
        &[],
    );
    let r = package(&fx);
    // Table order, only the items whose trigger the diff meets (the table edit is itself a
    // changed path, which `third` meets).
    assert_eq!(ids(&r), ["first", "third"], "{r}");
    let first = &r["checklist"][0];
    assert_eq!(first["check"], json!("check first"));
    assert_eq!(first["source"], json!("Issue #1"));
    assert_eq!(first["added"], json!("2026-11-20"));
    assert_eq!(first["matched"]["paths"], json!(["src/a.rs"]));
    assert_eq!(first["matched"]["pattern"], Value::Null);
    assert!(table_skips(&r).is_empty(), "{}", r["skipped"]);

    // No table at base: an empty list and no skipped entry, even when the head adds one.
    let fx = fixture(&[], &[(TABLE, &in_head), ("src/a.rs", "b\n")], &[]);
    let r = package(&fx);
    assert_eq!(ids(&r), Vec::<String>::new(), "{r}");
    assert!(table_skips(&r).is_empty(), "{}", r["skipped"]);
}

// --- triggers ------------------------------------------------------------------------------

#[test]
fn trigger() {
    let t = table(&[
        item(
            "rs-in-src",
            json!([{ "prefix": "src/", "suffixes": [".rs"] }]),
            None,
        ),
        item(
            "docs-read",
            json!([{ "prefix": "tests/", "suffixes": [".rs", ".py"] }]),
            Some(&["docs/", "README"]),
        ),
        item("anything", prefix(""), None),
    ]);
    let base = [(TABLE, t.as_str()), ("tests/old.rs", "old docs/ line\n")];

    // Path only: the suffix decides, the prefix is a prefix and nothing else.
    let fx = fixture(&base, &[("src/a.rs", "x\n"), ("src/b.md", "x\n")], &[]);
    let r = package(&fx);
    assert_eq!(ids(&r), ["rs-in-src", "anything"], "{r}");
    assert_eq!(r["checklist"][0]["matched"]["paths"], json!(["src/a.rs"]));
    assert_eq!(
        r["checklist"][1]["matched"]["paths"],
        json!(["src/a.rs", "src/b.md"])
    );
    let fx = fixture(&base, &[("srcx/a.rs", "x\n"), ("src/a.md", "x\n")], &[]);
    assert_eq!(ids(&package(&fx)), ["anything"]);

    // Path and pattern: the pattern is looked for in the added lines of a path in scope.
    let fx = fixture(
        &base,
        &[
            ("tests/a.rs", "let p = \"docs/x.md\";\n"),
            ("tests/b.rs", "nothing\n"),
        ],
        &[],
    );
    let r = package(&fx);
    assert_eq!(ids(&r), ["docs-read", "anything"], "{r}");
    assert_eq!(r["checklist"][0]["matched"]["paths"], json!(["tests/a.rs"]));
    assert_eq!(r["checklist"][0]["matched"]["pattern"], json!("docs/"));
    // The first pattern of the list that is found, not the first line that matches.
    let fx = fixture(&base, &[("tests/a.rs", "README\ndocs/\n")], &[]);
    assert_eq!(
        package(&fx)["checklist"][0]["matched"]["pattern"],
        json!("docs/")
    );

    // The same text in a path out of scope does not fire the item.
    let fx = fixture(&base, &[("src/a.rs", "docs/ in src\n")], &[]);
    assert_eq!(ids(&package(&fx)), ["rs-in-src", "anything"]);

    // A line that was removed or only context is not an added line.
    let fx = fixture(&base, &[("tests/old.rs", "new line\n")], &[]);
    assert_eq!(ids(&package(&fx)), ["anything"]);
    let fx = fixture(
        &[
            (TABLE, t.as_str()),
            ("tests/c.rs", "docs/ context\nkeep\nchange me\n"),
        ],
        &[("tests/c.rs", "docs/ context\nkeep\nchanged\n")],
        &[],
    );
    assert_eq!(ids(&package(&fx)), ["anything"]);

    // A deleted file has no added line: its pattern item stays quiet, its path-only item fires.
    let fx = fixture(&base, &[], &["tests/old.rs"]);
    let r = package(&fx);
    assert_eq!(ids(&r), ["anything"], "{r}");
    assert_eq!(
        r["checklist"][0]["matched"]["paths"],
        json!(["tests/old.rs"])
    );

    // No changed path, nothing fires (even the item that meets every path).
    let fx = fixture(&base, &[], &[]);
    assert_eq!(ids(&package(&fx)), Vec::<String>::new());
}

// --- invalid tables ------------------------------------------------------------------------

fn valid_item() -> Value {
    item("ok", prefix(""), None)
}

fn with(mut v: Value, f: impl FnOnce(&mut Value)) -> Value {
    f(&mut v);
    v
}

#[test]
fn invalid() {
    let cases: Vec<(&str, String)> = vec![
        ("not json", "{ nope".to_string()),
        (
            "schema_version",
            serde_json::to_string(&json!({ "schema_version": 2, "items": [valid_item()] }))
                .unwrap(),
        ),
        (
            "items",
            serde_json::to_string(&json!({ "schema_version": 1, "items": {} })).unwrap(),
        ),
        (
            "unknown top-level key",
            serde_json::to_string(
                &json!({ "schema_version": 1, "items": [valid_item()], "extra": 1 }),
            )
            .unwrap(),
        ),
        (
            "missing source",
            table(&[with(valid_item(), |v| {
                v.as_object_mut().unwrap().remove("source");
            })]),
        ),
        (
            "missing trigger",
            table(&[with(valid_item(), |v| {
                v.as_object_mut().unwrap().remove("trigger");
            })]),
        ),
        (
            "missing check",
            table(&[with(valid_item(), |v| {
                v.as_object_mut().unwrap().remove("check");
            })]),
        ),
        (
            "missing added",
            table(&[with(valid_item(), |v| {
                v.as_object_mut().unwrap().remove("added");
            })]),
        ),
        (
            "unknown item key",
            table(&[with(valid_item(), |v| v["sourc"] = json!("typo"))]),
        ),
        (
            "unknown trigger key",
            table(&[with(valid_item(), |v| {
                v["trigger"]["diff_patern"] = json!(["x"]);
            })]),
        ),
        ("duplicate id", table(&[valid_item(), valid_item()])),
        (
            "bad id",
            table(&[with(valid_item(), |v| v["id"] = json!("Not Ok"))]),
        ),
        (
            "source names nothing",
            table(&[with(valid_item(), |v| v["source"] = json!("C-12"))]),
        ),
        (
            "issue without number",
            table(&[with(valid_item(), |v| v["source"] = json!("Issue #"))]),
        ),
        (
            "bad date",
            table(&[with(valid_item(), |v| v["added"] = json!("2026-13-01"))]),
        ),
        (
            "bad date shape",
            table(&[with(valid_item(), |v| v["added"] = json!("2026-1-1"))]),
        ),
        (
            "empty paths",
            table(&[with(valid_item(), |v| v["trigger"]["paths"] = json!([]))]),
        ),
        (
            "path entry without prefix",
            table(&[with(valid_item(), |v| {
                v["trigger"]["paths"] = json!([{ "suffixes": [".rs"] }]);
            })]),
        ),
        (
            "empty suffix",
            table(&[with(valid_item(), |v| {
                v["trigger"]["paths"] = json!([{ "prefix": "", "suffixes": [""] }]);
            })]),
        ),
        (
            "empty diff_pattern list",
            table(&[with(valid_item(), |v| {
                v["trigger"]["diff_pattern"] = json!([])
            })]),
        ),
        (
            "empty diff_pattern string",
            table(&[with(valid_item(), |v| {
                v["trigger"]["diff_pattern"] = json!([""]);
            })]),
        ),
        (
            "empty check",
            table(&[with(valid_item(), |v| v["check"] = json!(" "))]),
        ),
    ];
    for (name, text) in cases {
        let fx = fixture(&[(TABLE, &text)], &[("src/a.rs", "x\n")], &[]);
        let r = package(&fx);
        assert_eq!(ids(&r), Vec::<String>::new(), "{name}: {r}");
        let skips = table_skips(&r);
        assert_eq!(skips.len(), 1, "{name}: {}", r["skipped"]);
        assert!(
            skips[0].starts_with("invalid-checklist: "),
            "{name}: {}",
            skips[0]
        );
        // The package itself is still built.
        assert_eq!(r["changed"].as_array().unwrap().len(), 1, "{name}");
    }

    // The three source shapes that are accepted.
    for source in [
        "Issue #2401",
        "PR #12 and Issue #3",
        "SPEC-20261120-review-checklist-by-diff",
    ] {
        let text = table(&[with(valid_item(), |v| v["source"] = json!(source))]);
        let fx = fixture(&[(TABLE, &text)], &[("src/a.rs", "x\n")], &[]);
        let r = package(&fx);
        assert_eq!(ids(&r), ["ok"], "{source}: {r}");
        assert!(table_skips(&r).is_empty(), "{source}");
    }
}

/// An enforced limit has tests on both sides of it: 150,000 bytes are read, 150,001 are not.
#[test]
fn invalid_byte_limit() {
    let build = |pad: usize| {
        serde_json::to_string(&json!({
            "schema_version": 1,
            "comment": "x".repeat(pad),
            "items": [valid_item()],
        }))
        .unwrap()
    };
    let floor = build(0).len();
    let at_limit = build(BYTE_LIMIT - floor);
    let over = build(BYTE_LIMIT + 1 - floor);
    assert_eq!((at_limit.len(), over.len()), (BYTE_LIMIT, BYTE_LIMIT + 1));

    let fx = fixture(&[(TABLE, &at_limit)], &[("src/a.rs", "x\n")], &[]);
    let r = package(&fx);
    assert_eq!(ids(&r), ["ok"], "{r}");
    assert!(table_skips(&r).is_empty());

    let fx = fixture(&[(TABLE, &over)], &[("src/a.rs", "x\n")], &[]);
    let r = package(&fx);
    assert_eq!(ids(&r), Vec::<String>::new());
    assert_eq!(table_skips(&r), ["byte-limit"], "{}", r["skipped"]);
}

/// Anything that is not a regular file at base is skipped with git's own kind, never followed.
#[cfg(unix)]
#[test]
fn invalid_symlink_table() {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    fs::create_dir_all(&repo).unwrap();
    git_ok(&repo, &["init", "-q", "-b", "main"]);
    write(&repo, "real.json", &table(&[valid_item()]));
    fs::create_dir_all(repo.join("tools")).unwrap();
    std::os::unix::fs::symlink("../real.json", repo.join(TABLE)).unwrap();
    let base = commit(&repo, "base");
    write(&repo, "src/a.rs", "x\n");
    let head = commit(&repo, "head");
    let fx = Fx {
        _tmp: tmp,
        repo,
        base,
        head,
    };
    let r = package(&fx);
    assert_eq!(ids(&r), Vec::<String>::new());
    assert_eq!(table_skips(&r), ["not-a-regular-file"], "{}", r["skipped"]);
}

/// SPEC REQ-5: a type change shows two diff sections for one changed path. The pattern items
/// are then not judged and the skip says so; the path-only items still are.
#[cfg(unix)]
#[test]
fn invalid_diff_sections() {
    let t = table(&[
        item("by-path", prefix("a"), None),
        item("by-pattern", prefix("a"), Some(&["target"])),
    ]);
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    fs::create_dir_all(&repo).unwrap();
    git_ok(&repo, &["init", "-q", "-b", "main"]);
    write(&repo, TABLE, &t);
    write(&repo, "a", "plain file\n");
    let base = commit(&repo, "base");
    fs::remove_file(repo.join("a")).unwrap();
    std::os::unix::fs::symlink("target", repo.join("a")).unwrap();
    let head = commit(&repo, "head");
    let fx = Fx {
        _tmp: tmp,
        repo,
        base,
        head,
    };
    let r = package(&fx);
    assert_eq!(r["changed"], json!([{ "path": "a", "status": "T" }]), "{r}");
    assert_eq!(ids(&r), ["by-path"], "{r}");
    assert_eq!(table_skips(&r), ["diff-sections-mismatch"]);
}

// --- the repository's own table --------------------------------------------------------------

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn real_text() -> String {
    fs::read_to_string(root().join(TABLE)).expect("tools/review_checklist.json exists")
}

fn real_items() -> Vec<Value> {
    let doc: Value = serde_json::from_str(&real_text()).expect("the table is JSON");
    doc["items"].as_array().expect("`items` is a list").clone()
}

/// SPEC REQ-6: this repository's table is valid, every item has a source and a trigger, and a
/// copy without one is refused (so the test can fail).
#[test]
fn real_file() {
    let text = real_text();
    let fx = fixture(&[(TABLE, &text)], &[("src/a.rs", "x\n")], &[]);
    let r = package(&fx);
    assert!(table_skips(&r).is_empty(), "{}", r["skipped"]);
    assert!(!ids(&r).is_empty(), "{r}");
    for it in real_items() {
        let id = it["id"].as_str().unwrap();
        assert!(
            it["source"].as_str().is_some_and(|s| !s.trim().is_empty()),
            "{id} has no source"
        );
        assert!(
            it["trigger"]["paths"]
                .as_array()
                .is_some_and(|p| !p.is_empty()),
            "{id} has no trigger"
        );
    }
    // The mutant: the first item without `source`.
    let mut doc: Value = serde_json::from_str(&text).unwrap();
    doc["items"][0].as_object_mut().unwrap().remove("source");
    let broken = serde_json::to_string(&doc).unwrap();
    let fx = fixture(&[(TABLE, &broken)], &[("src/a.rs", "x\n")], &[]);
    let r = package(&fx);
    assert_eq!(ids(&r), Vec::<String>::new());
    assert_eq!(table_skips(&r).len(), 1, "{}", r["skipped"]);
}

/// The rules of `verify.md` ("Change-specific rules" and "Shell and helper rules"): each has an
/// item whose `check` is the bullet's text word for word.
const VERIFY_RULES: [(&str, &str); 9] = [
    ("base-drift", "Base drift (WF-OBS #9):"),
    (
        "platform-specific-code",
        "Platform-specific code (WF-OBS #25):",
    ),
    ("checker-changes", "Checker changes (WF-OBS #26):"),
    ("absence-in-tests", "Absence in tests (WF-OBS #29):"),
    (
        "manifest-after-conflict",
        "Manifest after a conflict (WF-OBS #30):",
    ),
    ("guard-refusal", "Guard refusal (WF-OBS #17):"),
    ("host-worktree-guard", "Host worktree guard (Issue #2836):"),
    (
        "helper-auth-failure",
        "Helper authentication failure (WF-OBS #19):",
    ),
    ("zsh-word-splitting", "zsh word splitting (WF-OBS #16):"),
];

/// The five path-fixed bullets of `REVIEW.md` at the base of the change that moved them. They
/// leave `REVIEW.md`, so their text is pinned here; an item's `check` carries it verbatim after
/// its section label.
const REVIEW_RULES: [(&str, &str); 5] = [
    (
        "test-git-and-binary-helpers",
        "Test git calls that bypass `tests/common/git.rs`, or copying the binary with `fs::copy(EXE…)` instead of `common::copy_executable`.",
    ),
    (
        "test-reads-docs",
        "A test that reads a file under `docs/` without that path denied in `tools/ci_scope.py`.",
    ),
    (
        "skills-manifest-sha256",
        "A change under `skills/moai/**` without the matching `sha256` in `skills/moai/references/source-manifest.json`.",
    ),
    (
        "golden-files-rerecorded",
        "Re-recorded golden files under `tests/fixtures/**/golden/`: check only that each differs from the base by the intended change.",
    ),
    (
        "cargo-lock-follows-manifest",
        "`Cargo.lock` changes that only follow a `Cargo.toml` change in the same pull request.",
    ),
];

fn verify_bullets() -> Vec<String> {
    let text = fs::read_to_string(root().join("skills/moai/workflows/verify.md")).unwrap();
    let mut in_section = false;
    let mut out = Vec::new();
    for line in text.lines() {
        if let Some(head) = line.strip_prefix("## ") {
            in_section = head == "Change-specific rules" || head == "Shell and helper rules";
        } else if in_section {
            if let Some(bullet) = line.strip_prefix("- ") {
                out.push(bullet.to_string());
            }
        }
    }
    out
}

/// SPEC REQ-7: nothing is lost by the move.
#[test]
fn moved_rules() {
    let items = real_items();
    let check_of = |id: &str| -> String {
        items
            .iter()
            .find(|i| i["id"] == json!(id))
            .unwrap_or_else(|| panic!("no checklist item `{id}`"))["check"]
            .as_str()
            .unwrap()
            .to_string()
    };
    // Every rule is an item, whether or not verify.md still holds its text.
    for (id, _) in VERIFY_RULES.iter().chain(REVIEW_RULES.iter()) {
        let _ = check_of(id);
    }
    // While verify.md holds a rule, its item carries the text exactly; and every bullet that is
    // still there is one of the nine.
    let bullets = verify_bullets();
    for bullet in &bullets {
        let known = VERIFY_RULES
            .iter()
            .find(|(_, lead)| bullet.starts_with(lead));
        let (id, _) = known.unwrap_or_else(|| panic!("a verify.md rule has no item: {bullet}"));
        assert_eq!(&check_of(id), bullet, "item `{id}` differs from verify.md");
    }
    for (id, lead) in VERIFY_RULES {
        let text = check_of(id);
        assert!(text.starts_with(lead), "item `{id}` lost its label: {text}");
    }
    for (id, text) in REVIEW_RULES {
        assert!(
            check_of(id).contains(text),
            "item `{id}` lost the REVIEW.md text"
        );
    }
}

/// SPEC REQ-11: `REVIEW.md` points to the `checklist` key and no longer carries the five rules.
#[test]
fn review_md_points_to_checklist() {
    let text = fs::read_to_string(root().join("REVIEW.md")).unwrap();
    assert!(text.contains("`checklist`"), "REVIEW.md names the key");
    for (id, rule) in REVIEW_RULES {
        assert!(
            !text.contains(rule),
            "REVIEW.md still carries the text of `{id}`"
        );
    }
}

/// One firing diff and one quiet diff per item: (id, files that fire it, files that do not).
/// A quiet diff of `None` means "the empty diff": an item that meets every path has no other.
type Case = (
    &'static str,
    &'static [(&'static str, &'static str)],
    Option<&'static [(&'static str, &'static str)]>,
);

const CASES: [Case; 14] = [
    ("base-drift", &[("README.md", "x\n")], None),
    (
        "platform-specific-code",
        &[("src/a.rs", "#[cfg(windows)]\nfn f() {}\n")],
        Some(&[("src/a.rs", "fn f() {}\n")]),
    ),
    (
        "checker-changes",
        &[("tools/ci_scope.py", "x = 1\n")],
        Some(&[("tools/other.py", "x = 1\n")]),
    ),
    (
        "absence-in-tests",
        &[("tests/t.rs", "cmd.env_remove(\"PATH\");\n")],
        Some(&[("tests/t.rs", "fn t() {}\n")]),
    ),
    (
        "manifest-after-conflict",
        &[("skills/moai/references/source-manifest.json", "{}\n")],
        Some(&[("docs/manifest.json", "{}\n")]),
    ),
    ("guard-refusal", &[("README.md", "x\n")], None),
    ("host-worktree-guard", &[("README.md", "x\n")], None),
    ("helper-auth-failure", &[("README.md", "x\n")], None),
    ("zsh-word-splitting", &[("README.md", "x\n")], None),
    (
        "test-git-and-binary-helpers",
        &[("tests/t.rs", "let o = Command::new(\"git\");\n")],
        Some(&[("tests/t.rs", "fn ok() {}\n")]),
    ),
    (
        "test-reads-docs",
        &[("tests/t.rs", "let t = read(\"docs/x.md\");\n")],
        Some(&[("src/a.rs", "let t = read(\"docs/x.md\");\n")]),
    ),
    (
        "skills-manifest-sha256",
        &[("skills/moai/workflows/x.md", "x\n")],
        Some(&[("docs/x.md", "x\n")]),
    ),
    (
        "golden-files-rerecorded",
        &[("tests/fixtures/a/golden/x.txt", "x\n")],
        Some(&[("tests/other.rs", "x\n")]),
    ),
    (
        "cargo-lock-follows-manifest",
        &[("Cargo.lock", "x\n")],
        Some(&[("Cargo.toml", "x\n")]),
    ),
];

/// SPEC REQ-8: every item of the real table has a firing and a quiet case, and the cases do
/// what they say.
#[test]
fn every_item_has_cases() {
    let text = real_text();
    let in_table: BTreeSet<String> = real_items()
        .iter()
        .map(|i| i["id"].as_str().unwrap().to_string())
        .collect();
    let in_cases: BTreeSet<String> = CASES.iter().map(|c| c.0.to_string()).collect();
    assert_eq!(CASES.len(), in_cases.len(), "an id has two cases");
    assert_eq!(in_table, in_cases, "table and cases name different items");

    for (id, fire, quiet) in CASES {
        let fx = fixture(&[(TABLE, &text)], fire, &[]);
        let r = package(&fx);
        let fired = ids(&r);
        assert!(fired.contains(&id.to_string()), "{id} did not fire: {r}");
        let it = r["checklist"]
            .as_array()
            .unwrap()
            .iter()
            .find(|e| e["id"] == json!(id))
            .unwrap();
        for (path, _) in fire {
            assert!(
                it["matched"]["paths"]
                    .as_array()
                    .unwrap()
                    .contains(&json!(path)),
                "{id}: {path} is not in matched.paths: {it}"
            );
        }

        let quiet_files = quiet.unwrap_or(&[]);
        let fx = fixture(&[(TABLE, &text)], quiet_files, &[]);
        let r = package(&fx);
        assert!(
            !ids(&r).contains(&id.to_string()),
            "{id} fired on its quiet diff: {r}"
        );
    }
}

/// SPEC REQ-9: the table is on the protected list (an exact entry or a `/**` prefix of it).
#[test]
fn protected_list() {
    let list: Value = serde_json::from_str(
        &fs::read_to_string(root().join("tools/protected_checks.json")).unwrap(),
    )
    .unwrap();
    let covered = list["protected"].as_array().unwrap().iter().any(|e| {
        let e = e.as_str().unwrap();
        e == TABLE
            || e.strip_suffix("/**")
                .is_some_and(|dir| TABLE.starts_with(&format!("{dir}/")))
    });
    assert!(covered, "tools/protected_checks.json does not list {TABLE}");
}
