//! #2617 (#2388 AC-2): two branches that each add an internal operation merge without a
//! conflict in `src/main.rs` or `docs/design-review/appendix/interfaces.md`, because a new
//! operation is one new `src/ops/<word>.rs` plus one `operations/<word>.md` fragment. The
//! negative control shows the old way (a variant appended at the end of `enum Command`)
//! still conflicts, so the positive result comes from the layout, not from luck.
//! Temporary repositories only; the real files are copied in, never written.

mod common;

use common::git::{command, git, git_raw};
use std::{fs, path::Path};
use tempfile::TempDir;

const ROOT: &str = env!("CARGO_MANIFEST_DIR");
const MAIN: &str = "src/main.rs";
const DOC: &str = "docs/design-review/appendix/interfaces.md";
const FRAGMENTS: &str = "docs/design-review/appendix/operations";
const REGISTRY: &str = include_str!(concat!(env!("OUT_DIR"), "/ops_registry.rs"));

fn copy_in(repo: &Path, rel: &str) {
    let to = repo.join(rel);
    fs::create_dir_all(to.parent().unwrap()).unwrap();
    fs::copy(Path::new(ROOT).join(rel), to).unwrap();
}

fn write(repo: &Path, rel: &str, text: &str) {
    let to = repo.join(rel);
    fs::create_dir_all(to.parent().unwrap()).unwrap();
    fs::write(to, text).unwrap();
}

fn commit(repo: &Path, message: &str) {
    git(repo, &["add", "-A"]);
    let ok = command()
        .args(["-c", "user.name=F", "-c", "user.email=f@example.invalid"])
        .args(["commit", "-q", "-m", message])
        .current_dir(repo)
        .status()
        .unwrap()
        .success();
    assert!(ok, "commit {message}");
}

/// A repository holding today's `src/main.rs`, `interfaces.md` and `src/ops/`.
fn base_repo() -> TempDir {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path();
    git(repo, &["init", "-q", "-b", "main"]);
    copy_in(repo, MAIN);
    copy_in(repo, DOC);
    for entry in fs::read_dir(Path::new(ROOT).join("src/ops")).unwrap() {
        let name = entry.unwrap().file_name().to_string_lossy().into_owned();
        copy_in(repo, &format!("src/ops/{name}"));
    }
    for entry in fs::read_dir(Path::new(ROOT).join(FRAGMENTS)).unwrap() {
        let name = entry.unwrap().file_name().to_string_lossy().into_owned();
        copy_in(repo, &format!("{FRAGMENTS}/{name}"));
    }
    commit(repo, "base");
    tmp
}

fn op_file(word: &str) -> String {
    format!(
        "//! `{w}` (fixture).\n\n/// {w} help.\n#[derive(clap::Args)]\npub struct Args {{}}\n\npub const OPERATION: &str = \"{word}.run\";\n\npub fn run(_args: Args) -> Result<serde_json::Value, String> {{\n    Ok(serde_json::json!({{}}))\n}}\n",
        w = word.replace('_', "-")
    )
}

/// Adds one operation on its own branch the #2617 way: two new files, nothing else.
fn add_op(repo: &Path, branch: &str, word: &str) {
    git(repo, &["checkout", "-q", "-b", branch, "main"]);
    write(repo, &format!("src/ops/{word}.rs"), &op_file(word));
    write(
        repo,
        &format!("{FRAGMENTS}/{word}.md"),
        &format!("# `{}`\n\nfixture\n", word.replace('_', "-")),
    );
    commit(repo, branch);
}

/// Merges `b` into `a` and returns (merged cleanly, files git reported as conflicted).
fn merge(repo: &Path, a: &str, b: &str) -> (bool, String) {
    git(repo, &["checkout", "-q", a]);
    let out = git_raw(
        repo,
        &[
            "-c",
            "user.name=F",
            "-c",
            "user.email=f@example.invalid",
            "merge",
            "--no-edit",
            "-q",
            b,
        ],
    );
    let conflicts = git_raw(repo, &["diff", "--name-only", "--diff-filter=U"]).stdout;
    (out.success, conflicts)
}

#[test]
fn two_operation_files_on_two_branches_merge_without_conflict() {
    let tmp = base_repo();
    let repo = tmp.path();
    let main_before = fs::read(repo.join(MAIN)).unwrap();
    let doc_before = fs::read(repo.join(DOC)).unwrap();
    // Neighbouring names on purpose: the old alphabetical list put them in the same gap.
    add_op(repo, "lane-a", "zz_alpha_plan");
    add_op(repo, "lane-b", "zz_alpha_apply");
    let (ok, conflicts) = merge(repo, "lane-a", "lane-b");
    assert!(ok && conflicts.is_empty(), "conflicted: {conflicts}");
    assert_eq!(
        fs::read(repo.join(MAIN)).unwrap(),
        main_before,
        "src/main.rs changed"
    );
    assert_eq!(
        fs::read(repo.join(DOC)).unwrap(),
        doc_before,
        "interfaces.md changed"
    );
    for word in ["zz_alpha_plan", "zz_alpha_apply"] {
        assert!(repo.join(format!("src/ops/{word}.rs")).is_file());
        assert!(repo.join(format!("{FRAGMENTS}/{word}.md")).is_file());
    }
}

/// Negative control: the same two operations added the old way — a variant at the end of
/// `enum Command` — conflict in `src/main.rs`.
#[test]
fn the_old_way_still_conflicts() {
    let tmp = base_repo();
    let repo = tmp.path();
    let anchor = "    Op(ops::Ops),\n";
    for (branch, variant) in [("old-a", "ZzAlphaPlan"), ("old-b", "ZzAlphaApply")] {
        git(repo, &["checkout", "-q", "-b", branch, "main"]);
        let text = fs::read_to_string(repo.join(MAIN)).unwrap();
        assert!(text.contains(anchor), "anchor missing in src/main.rs");
        let text = text.replacen(
            anchor,
            &format!("{anchor}    {variant} {{ project: PathBuf }},\n"),
            1,
        );
        fs::write(repo.join(MAIN), text).unwrap();
        commit(repo, branch);
    }
    let (ok, conflicts) = merge(repo, "old-a", "old-b");
    assert!(
        !ok && conflicts.contains(MAIN),
        "expected a src/main.rs conflict: {conflicts}"
    );
}

/// The generated registry holds exactly the files under `src/ops/`, so adding a file is
/// enough for the operation to exist.
#[test]
fn the_registry_holds_every_operation_file() {
    let mut words: Vec<String> = fs::read_dir(Path::new(ROOT).join("src/ops"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .filter_map(|n| n.strip_suffix(".rs").map(str::to_string))
        .filter(|w| w != "mod")
        .collect();
    words.sort();
    assert!(!words.is_empty());
    let declared: Vec<&str> = REGISTRY
        .lines()
        .filter_map(|l| l.strip_prefix("pub mod "))
        .map(|l| l.trim_end_matches(';'))
        .collect();
    assert_eq!(declared, words);
}
