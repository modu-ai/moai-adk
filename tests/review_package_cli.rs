//! `review-package` (#2614): the deterministic review package an AUDIT hands to a reviewer —
//! the diff from the merge base to the head, the changed files, and the review instructions
//! (`REVIEW.md`, its `@` includes, subdirectory `REVIEW.md` files) and project context
//! (`AGENTS.md`, `CLAUDE.md`) read from the BASE commit only, with the limits of Claude Code's
//! Code Review documentation (code.claude.com/docs/en/code-review-customize). Temporary
//! repositories only; the binary reads git and writes nothing.

mod common;

use common::git::{command, git, git_ok};
use serde_json::{json, Value};
use std::{fs, path::Path, process::Output};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

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
        "-m",
        message,
    ])
    .current_dir(repo)
    .env("GIT_AUTHOR_DATE", "2026-01-01T00:00:00 +0000")
    .env("GIT_COMMITTER_DATE", "2026-01-01T00:00:00 +0000");
    assert!(cmd.status().unwrap().success());
    git(repo, &["rev-parse", "HEAD"])
}

fn init(repo: &Path) {
    fs::create_dir_all(repo).unwrap();
    git_ok(repo, &["init", "-q", "-b", "main"]);
}

/// A base commit with review instructions and context, and a head commit on top that changes
/// code and also edits REVIEW.md (the PR's own edit must not reach the package).
fn fixture() -> (TempDir, String, String) {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    init(&repo);
    write(
        &repo,
        "REVIEW.md",
        "# Review\n\nImportant: a missing trailer.\n\n@rules.md\n@sub/deep.md\n@.claude/skills/a/SKILL.md\n@missing.md\n",
    );
    write(&repo, "rules.md", "Rule: use tests/common/git.rs.\n");
    write(&repo, "sub/deep.md", "not beside the root\n");
    write(
        &repo,
        ".claude/skills/a/SKILL.md",
        "---\nname: a\n---\nbody\n",
    );
    write(&repo, "AGENTS.md", "root agents\n");
    write(&repo, "CLAUDE.md", "root claude\n");
    write(&repo, "src/AGENTS.md", "src agents\n");
    write(&repo, "src/REVIEW.md", "src review\n");
    write(&repo, "src/lib.rs", "pub fn a() {}\n");
    write(&repo, "tests/REVIEW.md", "on the skip list\n");
    write(&repo, "tests/t.rs", "#[test] fn t() {}\n");
    write(&repo, "other/REVIEW.md", "not on a changed path\n");
    let base = commit(&repo, "base");
    write(&repo, "src/lib.rs", "pub fn a() {}\npub fn b() {}\n");
    write(&repo, "tests/t.rs", "#[test] fn t() { assert!(true) }\n");
    write(&repo, "REVIEW.md", "# Review\n\nchanged by the PR\n");
    let head = commit(&repo, "head");
    (tmp, base, head)
}

fn run_in(repo: &Path, args: &[&str], envs: &[(&str, &str)]) -> Output {
    let mut cmd = std::process::Command::new(EXE);
    cmd.arg("review-package").arg(repo).args(args);
    for (k, v) in envs {
        cmd.env(k, v);
    }
    cmd.output().unwrap()
}

fn package(repo: &Path, base: &str, head: &str) -> Value {
    let out = run_in(repo, &["--base", base, "--head", head], &[]);
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

fn paths(list: &Value) -> Vec<String> {
    list.as_array()
        .unwrap()
        .iter()
        .map(|e| e["path"].as_str().unwrap().to_string())
        .collect()
}

fn skipped(result: &Value) -> Vec<(String, String)> {
    result["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| {
            (
                e["path"].as_str().unwrap().to_string(),
                e["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

#[test]
fn instructions_and_context_come_from_the_base_commit() {
    let (tmp, base, head) = fixture();
    let repo = tmp.path().join("repo");
    let r = package(&repo, &base, &head);
    assert_eq!(r["base"], json!(base));
    assert_eq!(r["head"], json!(head));
    assert_eq!(r["merge_base"], json!(base));
    assert_eq!(r["protection"], json!("not_checked"));
    assert_eq!(
        r["changed"],
        json!([
            {"path": "REVIEW.md", "status": "M"},
            {"path": "src/lib.rs", "status": "M"},
            {"path": "tests/t.rs", "status": "M"},
        ])
    );
    let diff = r["diff"].as_str().unwrap();
    assert!(diff.contains("+pub fn b() {}"), "{diff}");

    // Root file and its one include, from the base; the PR's REVIEW.md edit is not used.
    assert_eq!(
        paths(&r["review"]),
        ["REVIEW.md", "rules.md", "src/REVIEW.md"]
    );
    assert_eq!(r["review"][0]["role"], json!("root"));
    assert!(r["review"][0]["text"]
        .as_str()
        .unwrap()
        .contains("Important: a missing trailer."));
    assert_eq!(r["review"][1]["role"], json!("include"));
    assert_eq!(r["review"][1]["included_by"], json!("REVIEW.md"));
    assert_eq!(r["review"][2]["role"], json!("subdirectory"));
    let sk = skipped(&r);
    for want in [
        ("sub/deep.md", "include-not-beside-root"),
        (".claude/skills/a/SKILL.md", "skill-include-not-supported"),
        ("missing.md", "missing"),
        ("tests/REVIEW.md", "skipped-directory-name"),
    ] {
        assert!(
            sk.contains(&(want.0.to_string(), want.1.to_string())),
            "{want:?} not in {sk:?}"
        );
    }
    // A REVIEW.md off the changed paths is neither read nor reported.
    assert!(!paths(&r["review"]).contains(&"other/REVIEW.md".to_string()));
    assert!(!sk.iter().any(|(p, _)| p == "other/REVIEW.md"));

    // Context: AGENTS.md and CLAUDE.md at the root and above each changed file.
    assert_eq!(
        paths(&r["context"]),
        ["AGENTS.md", "CLAUDE.md", "src/AGENTS.md"]
    );
    for entry in r["review"]
        .as_array()
        .unwrap()
        .iter()
        .chain(r["context"].as_array().unwrap())
    {
        let text = entry["text"].as_str().unwrap();
        assert_eq!(entry["bytes"], json!(text.len()), "{entry}");
        assert_eq!(
            entry["sha256"],
            json!(moai_cli::digest(text.as_bytes())),
            "{entry}"
        );
    }
    assert!(
        r["package_digest"].as_str().unwrap().starts_with("sha256:"),
        "{r}"
    );
}

/// AC-2: the same base and head give the same bytes whatever the caller's environment —
/// another HOME with a user git config that changes diff output, another locale, git
/// variables in the environment, and another working directory.
#[test]
fn the_package_is_byte_identical_across_environments() {
    let (tmp, base, head) = fixture();
    let repo = tmp.path().join("repo");
    let plain = run_in(&repo, &["--base", &base, "--head", &head], &[]);
    assert!(plain.status.success());

    let home = tmp.path().join("home");
    // `diff.orderFile` has no command-line override in the package's fixed options, so it
    // shows whether the caller's configuration is ignored at all.
    let order = tmp.path().join("order.txt");
    write(tmp.path(), "order.txt", "tests/*\nREVIEW.md\n");
    write(
        &home,
        ".gitconfig",
        &format!(
            "[diff]\n\tnoprefix = true\n\tcontext = 9\n\trenames = copies\n\talgorithm = histogram\n\tmnemonicPrefix = true\n\torderFile = {}\n[color]\n\tui = always\n[core]\n\tquotePath = false\n",
            order.display()
        ),
    );
    let other = run_in(
        &repo,
        &["--base", &base, "--head", &head],
        &[
            ("HOME", home.to_str().unwrap()),
            ("XDG_CONFIG_HOME", home.to_str().unwrap()),
            ("LANG", "ko_KR.UTF-8"),
            ("LC_ALL", "ko_KR.UTF-8"),
            ("GIT_CONFIG_PARAMETERS", "'diff.noprefix'='true'"),
            ("GIT_EXTERNAL_DIFF", "false"),
        ],
    );
    assert!(
        other.status.success(),
        "{}",
        String::from_utf8_lossy(&other.stderr)
    );
    assert_eq!(
        String::from_utf8_lossy(&plain.stdout),
        String::from_utf8_lossy(&other.stdout)
    );

    // The same package from another working directory and a relative project path.
    let mut cmd = std::process::Command::new(EXE);
    cmd.current_dir(tmp.path())
        .args(["review-package", "repo", "--base", &base, "--head", &head]);
    let relative = cmd.output().unwrap();
    assert!(relative.status.success());
    assert_eq!(plain.stdout, relative.stdout);
}

/// The include count: eight different files are read, the ninth is skipped whole.
#[test]
fn includes_stop_at_eight_different_files() {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    init(&repo);
    let mut root = String::from("# r\n");
    for i in 1..=9 {
        root.push_str(&format!("@i{i}.md\n"));
        write(&repo, &format!("i{i}.md"), &format!("include {i}\n"));
    }
    root.push_str("@i1.md\n"); // the same file twice counts once
    write(&repo, "REVIEW.md", &root);
    write(&repo, "a.txt", "a\n");
    let base = commit(&repo, "base");
    write(&repo, "a.txt", "b\n");
    let head = commit(&repo, "head");
    let r = package(&repo, &base, &head);
    let includes: Vec<String> = paths(&r["review"]).into_iter().skip(1).collect();
    assert_eq!(
        includes,
        (1..=8).map(|i| format!("i{i}.md")).collect::<Vec<_>>()
    );
    assert_eq!(
        skipped(&r),
        [("i9.md".to_string(), "include-count".to_string())]
    );
}

fn byte_case(total: usize) -> Value {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    init(&repo);
    let root = "@big.md\n";
    write(&repo, "REVIEW.md", root);
    write(&repo, "big.md", &"x".repeat(total - root.len()));
    write(&repo, "a.txt", "a\n");
    let base = commit(&repo, "base");
    write(&repo, "a.txt", "b\n");
    let head = commit(&repo, "head");
    package(&repo, &base, &head)
}

/// The root file plus its includes stay under 150,000 bytes combined: exactly 150,000 is
/// read, one byte more skips the include whole (never cut short).
#[test]
fn includes_share_a_150000_byte_limit() {
    let at = byte_case(150_000);
    assert_eq!(paths(&at["review"]), ["REVIEW.md", "big.md"]);
    assert!(at["skipped"].as_array().unwrap().is_empty(), "{at}");
    let over = byte_case(150_001);
    assert_eq!(paths(&over["review"]), ["REVIEW.md"]);
    assert_eq!(
        skipped(&over),
        [("big.md".to_string(), "byte-limit".to_string())]
    );
}

fn link_case(links: usize) -> Value {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    init(&repo);
    let mut root =
        String::from("# r\n[web](https://example.com) [mail](mailto:a@b.c) [here](#x)\n");
    for i in 0..links {
        root.push_str(&format!("[l{i}](sub/l{i}.md)\n"));
    }
    write(&repo, "REVIEW.md", &root);
    write(&repo, "a.txt", "a\n");
    let base = commit(&repo, "base");
    write(&repo, "a.txt", "b\n");
    let head = commit(&repo, "head");
    package(&repo, &base, &head)
}

/// More than 64 Markdown links (http, https, mailto and `#` links not counted) skip the whole
/// root file.
#[test]
fn more_than_64_links_skip_the_root_file() {
    let at = link_case(64);
    assert_eq!(paths(&at["review"]), ["REVIEW.md"]);
    let over = link_case(65);
    assert!(over["review"].as_array().unwrap().is_empty(), "{over}");
    assert_eq!(
        skipped(&over),
        [("REVIEW.md".to_string(), "too-many-links".to_string())]
    );
}

/// A REVIEW.md that is a symbolic link in the base commit is not read.
#[cfg(unix)]
#[test]
fn a_linked_review_file_is_not_read() {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    init(&repo);
    write(&repo, "real.md", "real\n");
    std::os::unix::fs::symlink("real.md", repo.join("REVIEW.md")).unwrap();
    write(&repo, "a.txt", "a\n");
    let base = commit(&repo, "base");
    write(&repo, "a.txt", "b\n");
    let head = commit(&repo, "head");
    let r = package(&repo, &base, &head);
    assert!(r["review"].as_array().unwrap().is_empty(), "{r}");
    assert_eq!(
        skipped(&r),
        [("REVIEW.md".to_string(), "not-a-regular-file".to_string())]
    );
}

/// An unknown revision is an error, never an empty package.
#[test]
fn an_unknown_revision_is_refused() {
    let (tmp, base, _) = fixture();
    let repo = tmp.path().join("repo");
    let out = run_in(&repo, &["--base", &base, "--head", "no-such-rev"], &[]);
    assert!(!out.status.success());
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
}

/// Audit FAIL 6839909483 B1: the repository's own `.git/config` and the uncommitted worktree
/// `.gitattributes` must not change the package either — only BASE and HEAD do.
#[test]
fn repository_settings_and_worktree_attributes_do_not_change_the_package() {
    let (tmp, base, head) = fixture();
    let repo = tmp.path().join("repo");
    let plain = run_in(&repo, &["--base", &base, "--head", &head], &[]);
    assert!(plain.status.success());

    write(tmp.path(), "order.txt", "tests/*\nREVIEW.md\n");
    let order = tmp.path().join("order.txt");
    for (key, value) in [
        ("diff.orderFile", order.to_str().unwrap()),
        ("diff.suppressBlankEmpty", "true"),
        ("diff.interHunkContext", "9"),
        ("diff.indentHeuristic", "false"),
        ("diff.noprefix", "true"),
        ("diff.relative", "true"),
        ("core.abbrev", "12"),
        ("core.bigFileThreshold", "1"),
        ("core.quotePath", "false"),
        ("diff.fn.xfuncname", "^#"),
    ] {
        git_ok(&repo, &["config", key, value]);
    }
    // Uncommitted attributes in the worktree: binary for Rust files, a hunk-header driver.
    write(&repo, ".gitattributes", "*.rs -diff\n*.md diff=fn\n");
    let other = run_in(&repo, &["--base", &base, "--head", &head], &[]);
    assert!(
        other.status.success(),
        "{}",
        String::from_utf8_lossy(&other.stdout)
    );
    assert_eq!(
        String::from_utf8_lossy(&plain.stdout),
        String::from_utf8_lossy(&other.stdout)
    );
}

/// `.git/info/attributes` is read whatever the attribute source, so a package built under it
/// could differ; it is refused instead of returned.
#[test]
fn repository_info_attributes_are_refused() {
    let (tmp, base, head) = fixture();
    let repo = tmp.path().join("repo");
    write(&repo, ".git/info/attributes", "*.rs -diff\n");
    let out = run_in(&repo, &["--base", &base, "--head", &head], &[]);
    assert!(!out.status.success());
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("info/attributes"),
        "{v}"
    );
}

/// Audit FAIL 6839909483 B2: `@.` and `@..` are not includes beside the root file; they are
/// skipped like any other include that does not name a file there, and never fail the package.
#[test]
fn dot_includes_are_skipped_not_followed() {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    init(&repo);
    write(&repo, ".aaa-first.md", "not an include\n");
    write(&repo, "REVIEW.md", "# r\n@.\n@..\n");
    write(&repo, "a.txt", "a\n");
    let base = commit(&repo, "base");
    write(&repo, "a.txt", "b\n");
    let head = commit(&repo, "head");
    let r = package(&repo, &base, &head);
    assert_eq!(paths(&r["review"]), ["REVIEW.md"]);
    assert_eq!(
        skipped(&r),
        [
            (".".to_string(), "include-not-beside-root".to_string()),
            ("..".to_string(), "include-not-beside-root".to_string()),
        ]
    );
}

/// B2, the layers behind the name rule: `@:` is a pathspec that lists every root entry (only
/// an entry named exactly `:` would count, and there is none), and `@:!a.md` makes git itself
/// fail — that skips the one include, never the package.
#[test]
fn pathspec_like_includes_are_skipped_not_followed() {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    init(&repo);
    write(&repo, "a.md", "not an include\n");
    write(&repo, "REVIEW.md", "# r\n@:\n@:!a.md\n");
    write(&repo, "a.txt", "a\n");
    let base = commit(&repo, "base");
    write(&repo, "a.txt", "b\n");
    let head = commit(&repo, "head");
    let r = package(&repo, &base, &head);
    assert_eq!(paths(&r["review"]), ["REVIEW.md"]);
    assert_eq!(
        skipped(&r),
        [
            (":".to_string(), "missing".to_string()),
            (":!a.md".to_string(), "unreadable".to_string()),
        ]
    );
}

/// Follow-up 6840613078 (agent-25): a fixture where each fixed option changes the bytes when
/// the repository's own setting wins — two hunks nine lines apart (`diff.interHunkContext`),
/// a function inserted between similar blocks (`diff.indentHeuristic`), a file with a Korean
/// name (`core.quotePath`), and PROJECT given as a subfolder (`diff.relative`).
fn options_fixture() -> (TempDir, String, String) {
    let tmp = TempDir::new().unwrap();
    let repo = tmp.path().join("repo");
    init(&repo);
    let long: String = (1..=20).map(|i| format!("line {i}\n")).collect();
    write(&repo, "src/long.rs", &long);
    write(
        &repo,
        "src/blocks.rs",
        "fn a() {\n    x();\n}\n\nfn c() {\n    z();\n}\n",
    );
    write(&repo, "src/한글.rs", "a\n");
    write(&repo, "REVIEW.md", "root review\n");
    write(&repo, "src/REVIEW.md", "src review\n");
    let base = commit(&repo, "base");
    let long = long
        .replace("line 3\n", "line three\n")
        .replace("line 12\n", "line twelve\n");
    write(&repo, "src/long.rs", &long);
    write(
        &repo,
        "src/blocks.rs",
        "fn a() {\n    x();\n}\n\nfn b() {\n    y();\n}\n\nfn c() {\n    z();\n}\n",
    );
    write(&repo, "src/한글.rs", "b\n");
    let head = commit(&repo, "head");
    (tmp, base, head)
}

#[test]
fn every_fixed_option_holds_against_the_repository_setting() {
    let (tmp, base, head) = options_fixture();
    let repo = tmp.path().join("repo");
    let plain = run_in(&repo, &["--base", &base, "--head", &head], &[]);
    assert!(plain.status.success());
    let text = String::from_utf8_lossy(&plain.stdout).to_string();
    assert!(
        text.contains("\\\\355\\\\225\\\\234"),
        "quoted Korean path expected: {text}"
    );
    for (key, value) in [
        ("diff.interHunkContext", "9"),
        ("diff.indentHeuristic", "false"),
        ("core.quotePath", "false"),
        ("diff.relative", "true"),
    ] {
        git_ok(&repo, &["config", key, value]);
    }
    let configured = run_in(&repo, &["--base", &base, "--head", &head], &[]);
    assert!(configured.status.success());
    assert_eq!(text, String::from_utf8_lossy(&configured.stdout));

    // PROJECT given as a subfolder is the same repository: the same bytes, the root REVIEW.md
    // still the root file.
    let sub = run_in(&repo.join("src"), &["--base", &base, "--head", &head], &[]);
    assert!(
        sub.status.success(),
        "{}",
        String::from_utf8_lossy(&sub.stdout)
    );
    assert_eq!(text, String::from_utf8_lossy(&sub.stdout));
}
