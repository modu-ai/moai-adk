//! The guard on the reports folder of `moai optimize-plan`: a folder named like a repository
//! copy or a dependency folder, and a card folder over the evidence cap, are listed as
//! findings; a normal folder passes with an empty list. Temporary projects and HOME only, and
//! the guard itself never writes: the fixtures are read back unchanged.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const BIN: &str = crate::common::MOAI_BIN;
/// The per-card evidence cap as a literal, so a changed bound fails here.
const CAP: u64 = 100 * 1024 * 1024;
const REPORTS: &str = ".moai-cli/reports";

fn git(dir: &Path, args: &[&str]) {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all);
}

struct Fixture {
    _dir: tempfile::TempDir,
    root: PathBuf,
    home: PathBuf,
}

fn fixture() -> Fixture {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let (root, home) = (base.join("repo"), base.join("home"));
    for d in [&root, &home] {
        fs::create_dir_all(d).unwrap();
    }
    git(&root, &["init", "-q", "-b", "main"]);
    fs::write(root.join(".gitignore"), ".moai-cli/\n").unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", ".gitignore", "a.txt"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    git(&root, &["branch", "integ"]);
    fs::create_dir_all(root.join(REPORTS)).unwrap();
    Fixture {
        _dir: dir,
        root,
        home,
    }
}

fn put(fx: &Fixture, rel: &str, bytes: &[u8]) {
    let path = fx.root.join(REPORTS).join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, bytes).unwrap();
}

/// A sparse file of exactly `len` bytes (no disk is used for the hole).
fn sparse(fx: &Fixture, rel: &str, len: u64) {
    let path = fx.root.join(REPORTS).join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::File::create(path).unwrap().set_len(len).unwrap();
}

/// The `report` row of the project's reports folder.
fn report_row(fx: &Fixture) -> Value {
    let out = Command::new(BIN)
        .args([
            "optimize-plan",
            fx.root.to_str().unwrap(),
            "--integration-ref",
            "integ",
        ])
        .current_dir(&fx.root)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], true, "{v}");
    v["result"]["rows"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["kind"] == "report" && r["path"] == REPORTS)
        .unwrap_or_else(|| panic!("no report row: {v}"))
        .clone()
}

fn findings(row: &Value) -> Vec<Value> {
    row["findings"]
        .as_array()
        .unwrap_or_else(|| panic!("no findings list: {row}"))
        .clone()
}

fn names(row: &Value, kind: &str) -> Vec<String> {
    let mut out: Vec<String> = findings(row)
        .iter()
        .filter(|f| f["kind"] == kind)
        .map(|f| f["path"].as_str().unwrap().to_string())
        .collect();
    out.sort();
    out
}

#[test]
fn flags_a_repository_copy_and_its_dependency_folders() {
    let fx = fixture();
    put(&fx, "card-1/copy/.git/HEAD", b"ref: refs/heads/x\n");
    put(&fx, "card-1/copy/web/node_modules/pkg/index.js", b"x\n");
    put(&fx, "card-1/copy/.venv/bin/python", b"x\n");
    put(&fx, "card-1/copy/target/debug/x", b"x\n");
    put(&fx, "card-1/copy/web/.next/cache", b"x\n");
    put(&fx, "card-1/notes.md", b"fine\n");
    let row = report_row(&fx);
    assert_eq!(
        names(&row, "forbidden_folder"),
        [
            ".moai-cli/reports/card-1/copy/.git",
            ".moai-cli/reports/card-1/copy/.venv",
            ".moai-cli/reports/card-1/copy/target",
            ".moai-cli/reports/card-1/copy/web/.next",
            ".moai-cli/reports/card-1/copy/web/node_modules",
        ],
        "{row}"
    );
    // A finding is a note: the row stays `keep`, names the reason, and nothing was removed.
    assert_eq!(row["action"], "keep", "{row}");
    assert!(
        row["reasons"]
            .as_array()
            .unwrap()
            .contains(&json!("evidence_guard")),
        "{row}"
    );
    assert!(row["reasons"]
        .as_array()
        .unwrap()
        .contains(&json!("reports_kept")));
    assert!(fx
        .root
        .join(REPORTS)
        .join("card-1/copy/.git/HEAD")
        .is_file());
    assert!(fx
        .root
        .join(REPORTS)
        .join("card-1/copy/web/node_modules/pkg/index.js")
        .is_file());
}

#[test]
fn flags_a_link_named_like_a_dependency_folder_without_entering_it() {
    let fx = fixture();
    let outside = fx.home.join("elsewhere");
    fs::create_dir_all(outside.join(".git")).unwrap();
    fs::write(outside.join("big.bin"), b"x\n").unwrap();
    let card = fx.root.join(REPORTS).join("card-2");
    fs::create_dir_all(&card).unwrap();
    std::os::unix::fs::symlink(&outside, card.join("node_modules")).unwrap();
    let row = report_row(&fx);
    // The link is named, and the `.git` behind it is not: nothing is followed.
    assert_eq!(
        names(&row, "forbidden_folder"),
        [".moai-cli/reports/card-2/node_modules"],
        "{row}"
    );
    assert!(outside.join("big.bin").is_file());
}

#[test]
fn flags_a_card_folder_over_the_cap_with_its_size() {
    let fx = fixture();
    sparse(&fx, "card-3/log.bin", CAP + 1);
    put(&fx, "card-4/small.md", b"small\n");
    let row = report_row(&fx);
    let over: Vec<Value> = findings(&row)
        .into_iter()
        .filter(|f| f["kind"] == "over_cap")
        .collect();
    assert_eq!(over.len(), 1, "{row}");
    assert_eq!(over[0]["path"], ".moai-cli/reports/card-3");
    assert_eq!(over[0]["cap_bytes"], CAP);
    assert!(over[0]["bytes_at_least"].as_u64().unwrap() > CAP, "{row}");
    assert_eq!(row["action"], "keep");
}

#[test]
fn passes_exactly_the_cap_and_flags_one_byte_more() {
    let fx = fixture();
    sparse(&fx, "card-at/log.bin", CAP);
    assert_eq!(findings(&report_row(&fx)), Vec::<Value>::new());
    sparse(&fx, "card-over/log.bin", CAP + 1);
    let row = report_row(&fx);
    assert_eq!(
        names(&row, "over_cap"),
        [".moai-cli/reports/card-over"],
        "{row}"
    );
}

#[test]
fn passes_a_normal_reports_folder_with_an_empty_list() {
    let fx = fixture();
    put(&fx, "card-1/report.md", b"# report\n");
    put(&fx, "card-1/logs/run.txt", b"ok\n");
    put(&fx, "loose.md", b"top-level file\n");
    let row = report_row(&fx);
    assert_eq!(findings(&row), Vec::<Value>::new(), "{row}");
    assert_eq!(row["reasons"], json!(["reports_kept"]), "{row}");
    assert_eq!(row["action"], "keep");
}

#[test]
fn passes_names_that_only_contain_a_marker() {
    let fx = fixture();
    for dir in [
        "card-1/target-notes/a.md",
        "card-1/node_modules_old/a.md",
        "card-1/my.git/a.md",
        "card-1/.github/a.md",
        "card-1/venv/a.md",
        "card-1/.gitignore/a.md",
    ] {
        put(&fx, dir, b"x\n");
    }
    let row = report_row(&fx);
    assert_eq!(findings(&row), Vec::<Value>::new(), "{row}");
}

#[test]
fn flags_the_git_file_of_a_checked_out_worktree() {
    let fx = fixture();
    // In a worktree `.git` is a file that points at the main repository; it still marks a copy.
    put(
        &fx,
        "card-5/checkout/.git",
        b"gitdir: /elsewhere/.git/worktrees/x\n",
    );
    let row = report_row(&fx);
    assert_eq!(
        names(&row, "forbidden_folder"),
        [".moai-cli/reports/card-5/checkout/.git"],
        "{row}"
    );
}

#[test]
fn flags_a_dependency_folder_but_does_not_count_its_size() {
    let fx = fixture();
    // The folder is named and left alone; the bytes in it are not part of the card's evidence
    // size, so no second finding for the cap appears.
    sparse(&fx, "card-6/copy/node_modules/big.bin", CAP + 1);
    put(&fx, "card-6/copy/notes.md", b"small\n");
    let row = report_row(&fx);
    assert_eq!(
        names(&row, "forbidden_folder"),
        [".moai-cli/reports/card-6/copy/node_modules"],
        "{row}"
    );
    assert_eq!(names(&row, "over_cap"), Vec::<String>::new(), "{row}");
}

#[test]
fn adds_up_the_files_of_nested_folders_toward_the_cap() {
    let fx = fixture();
    // Three files of 40 MiB in three depths: each alone is far under the cap, the sum is over.
    // A guard that counted only the last file (or only the top folder) would pass this card.
    let part = 40 * 1024 * 1024;
    sparse(&fx, "card-7/a.bin", part);
    sparse(&fx, "card-7/x/b.bin", part);
    sparse(&fx, "card-7/x/y/c.bin", part);
    sparse(&fx, "card-8/x/y/only.bin", part);
    let row = report_row(&fx);
    assert_eq!(
        names(&row, "over_cap"),
        [".moai-cli/reports/card-7"],
        "{row}"
    );
    let over = findings(&row)
        .into_iter()
        .find(|f| f["kind"] == "over_cap")
        .unwrap();
    assert_eq!(over["bytes_at_least"], 3 * part, "{row}");
}

#[test]
fn flags_a_dependency_folder_and_a_git_file_directly_below_reports() {
    let fx = fixture();
    // A copy put straight into `reports`, not into a card folder: named, never entered, and the
    // folder beside it is checked as usual.
    sparse(&fx, "node_modules/big.bin", CAP + 1);
    put(&fx, ".git", b"gitdir: /elsewhere/.git/worktrees/x\n");
    put(&fx, "target/debug/x.o", b"x");
    put(&fx, "card-9/notes.md", b"small\n");
    let row = report_row(&fx);
    assert_eq!(
        names(&row, "forbidden_folder"),
        [
            ".moai-cli/reports/.git",
            ".moai-cli/reports/node_modules",
            ".moai-cli/reports/target"
        ],
        "{row}"
    );
    assert_eq!(names(&row, "over_cap"), Vec::<String>::new(), "{row}");
    assert_eq!(row["findings_partial"], false, "{row}");
}

#[test]
fn says_so_when_a_folder_cannot_be_read_and_still_checks_the_rest() {
    use std::os::unix::fs::PermissionsExt;
    let fx = fixture();
    put(&fx, "card-a/locked/inside.md", b"hidden\n");
    put(&fx, "card-b/copy/node_modules/x.js", b"x");
    let locked = fx.root.join(REPORTS).join("card-a/locked");
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o000)).unwrap();
    // A user that can read a mode-000 folder (root) cannot make the case: nothing to check.
    let readable = fs::read_dir(&locked).is_ok();
    let row = (!readable).then(|| report_row(&fx));
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o755)).unwrap();
    let Some(row) = row else { return };
    assert_eq!(row["findings_partial"], true, "{row}");
    // The unreadable folder does not stop the check of the card after it.
    assert_eq!(
        names(&row, "forbidden_folder"),
        [".moai-cli/reports/card-b/copy/node_modules"],
        "{row}"
    );
}
