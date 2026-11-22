//! #2617 (#2388 AC-2): two branches that each add an internal operation merge without a
//! conflict in `src/main.rs` or `docs/design-review/appendix/interfaces.md`, because a new
//! operation is one new `src/ops/<word>.rs` plus one `operations/<word>.md` fragment. The
//! negative control shows the old way (a variant appended at the end of `enum Command`)
//! still conflicts, so the positive result comes from the layout, not from luck.
//! Temporary repositories only; the real files are copied in, never written.

mod common;

use common::git::{command, git, git_raw};
use std::{fs, path::Path, process::Command};
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
    // An entry line of the `with_ops!` list: the file word, the variant name, the help lines in
    // brackets (SPEC-20261122-ops-registry-path-free).
    let declared: Vec<&str> = REGISTRY
        .lines()
        .map(str::trim)
        .filter(|l| l.ends_with(']') && !l.starts_with("//"))
        .filter_map(|l| l.split_whitespace().next())
        .collect();
    assert_eq!(declared, words);
}

// SPEC-20261122-ops-registry-path-free: the generated registry names no file path. A build
// folder that another worktree filled (shared, or warmed by an APFS clone) must not make
// this worktree compile the other worktree's operation files, so the registry carries names
// and help lines only and `src/ops/mod.rs` declares the modules next to itself. These tests
// live here because this file is the listed scan of `src/ops/` (tools/repo_scans.json).

/// The 1-based numbers and texts of the registry lines that hold a file path: a `#[path`
/// attribute, `src/ops/`, or the package folder `root`.
fn path_lines(text: &str, root: &str) -> Vec<(usize, String)> {
    text.lines()
        .enumerate()
        .filter(|(_, l)| {
            l.contains("#[path") || l.contains("src/ops/") || (!root.is_empty() && l.contains(root))
        })
        .map(|(i, l)| (i + 1, l.to_string()))
        .collect()
}

/// The operation file words under `src/ops/`, in file-name order.
fn op_words() -> Vec<String> {
    let mut words: Vec<String> = fs::read_dir(Path::new(ROOT).join("src/ops"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .filter_map(|n| n.strip_suffix(".rs").map(str::to_string))
        .filter(|w| w != "mod")
        .collect();
    words.sort();
    words
}

/// The words of the `with_ops!` list: an entry line starts with the file word, then the
/// variant name, then the help lines in brackets.
fn listed_words(text: &str) -> Vec<String> {
    text.lines()
        .map(str::trim)
        .filter(|l| l.ends_with(']'))
        .filter_map(|l| {
            let mut parts = l.split_whitespace();
            let word = parts.next()?;
            let variant = parts.next()?;
            let is_word = word
                .bytes()
                .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit() || b == b'_');
            let is_variant = variant.starts_with(|c: char| c.is_ascii_uppercase())
                && variant.bytes().all(|b| b.is_ascii_alphanumeric());
            (is_word && is_variant).then(|| word.to_string())
        })
        .collect()
}

/// The first `///` line right above `pub struct Args` in an operation file, if any.
fn first_help_line(text: &str) -> Option<String> {
    let lines: Vec<&str> = text.lines().collect();
    let at = lines
        .iter()
        .position(|l| l.trim_start().starts_with("pub struct Args"))?;
    let mut doc = Vec::new();
    for line in lines[..at].iter().rev() {
        let t = line.trim_start();
        if let Some(rest) = t.strip_prefix("///") {
            doc.push(rest.strip_prefix(' ').unwrap_or(rest).to_string());
        } else if t.starts_with("#[") {
            continue;
        } else {
            break;
        }
    }
    doc.reverse();
    doc.into_iter().next().filter(|l| !l.trim().is_empty())
}

/// AC-1: the registry is a `with_ops!` list of every operation file, in file-name order, and
/// names no path.
#[test]
fn the_registry_is_a_list_of_names_with_no_path() {
    assert!(
        REGISTRY.contains("macro_rules! with_ops"),
        "no with_ops! list in the registry:\n{REGISTRY}"
    );
    let words = op_words();
    assert!(!words.is_empty());
    assert_eq!(listed_words(REGISTRY), words);
    let found = path_lines(REGISTRY, ROOT);
    assert!(
        found.is_empty(),
        "registry lines that hold a path: {found:?}"
    );
}

/// AC-3: the path finder names each line that holds a path, and nothing in the real registry.
#[test]
fn a_path_in_the_registry_is_named() {
    let samples = [
        format!("#[path = \"{ROOT}/src/ops/alpha.rs\"]"),
        "#[path = \"/elsewhere/alpha.rs\"]".to_string(),
        "// see src/ops/alpha.rs".to_string(),
        format!("const HOME: &str = \"{ROOT}\";"),
    ];
    for sample in &samples {
        let text = format!("macro_rules! with_ops {{\n{sample}\n}}\n");
        let found = path_lines(&text, ROOT);
        assert_eq!(found, vec![(2, sample.clone())], "{sample}");
    }
    assert_eq!(path_lines("alpha Alpha [\"Alpha help\"]\n", ROOT), vec![]);
    assert_eq!(path_lines(REGISTRY, ROOT), vec![]);
}

/// AC-2: every operation answers `--help` under its kebab-case name, and the first help line
/// is the first `///` line above its `pub struct Args`.
#[test]
fn every_operation_keeps_its_name_and_help_line() {
    let exe = env!("CARGO_BIN_EXE_moai");
    for word in op_words() {
        let name = word.replace('_', "-");
        let out = Command::new(exe)
            .args([name.as_str(), "--help"])
            .env("NO_COLOR", "1")
            .output()
            .unwrap();
        let stdout = String::from_utf8_lossy(&out.stdout);
        assert!(
            out.status.success(),
            "{name} --help: {:?} {stdout} {}",
            out.status.code(),
            String::from_utf8_lossy(&out.stderr)
        );
        let text = fs::read_to_string(Path::new(ROOT).join(format!("src/ops/{word}.rs"))).unwrap();
        if let Some(line) = first_help_line(&text) {
            let first = stdout.lines().next().unwrap_or("");
            assert!(
                first.starts_with(line.trim_end()),
                "{name}: help starts {first:?}, the file says {line:?}"
            );
        }
    }
}
