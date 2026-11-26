//! SPEC-20261126-manifest-merge-driver-rust: `moai manifest-merge-driver %O %A %B %P`, the git
//! merge driver of the skill source manifest.
//!
//! The driver cases of `tests/source_manifest_test.py` live here. Every case builds a temporary
//! repository with a small package under `skills/moai/`, runs every git and `moai` child with a
//! clean environment and a temporary HOME, and reads or writes no real HOME. The expected hashes
//! come from `check_fix::render_manifest`, the function `moai check fix` writes the manifest with.
//!
//! The five name prefixes are the selectors of the SPEC's verify commands: `merged_`, `refused_`,
//! `registered_`, `order_` and `text_`.
#![cfg(unix)]

mod common;
#[path = "common/prop.rs"]
mod prop;

use common::home::TempHome;
use common::{CleanEnv, MOAI_BIN};
use moai_cli::check_fix::render_manifest;
use serde_json::{json, Map, Value};
use std::collections::{BTreeMap, BTreeSet};
use std::ffi::{OsStr, OsString};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use tempfile::TempDir;

const ROOT: &str = env!("CARGO_MANIFEST_DIR");
const PACKAGE: &str = "skills/moai";
/// The manifest, relative to the package folder.
const IN_PACKAGE: &str = "references/source-manifest.json";
/// The manifest, relative to the repository root: the `%P` git passes.
const MANIFEST: &str = "skills/moai/references/source-manifest.json";
const COMMAND: &str = "manifest-merge-driver";
/// The driver name `.gitattributes` gives the manifest, and its configuration key.
const DRIVER: &str = "moai-manifest-rs";
const DRIVER_KEY: &str = "merge.moai-manifest-rs.driver";
/// The earlier driver name, which a clone may still have registered. It is a prefix of `DRIVER`.
const OLD_DRIVER: &str = "moai-manifest";
const OLD_DRIVER_KEY: &str = "merge.moai-manifest.driver";
/// What a refusal tells the person to run.
const REPAIR: &str = "moai check fix";
/// The document that holds the registration command.
const DOCUMENT: &str = "docs/agents-md-maintenance.md";
/// The decisions log, a history record that may keep naming what was removed.
const DECISIONS: &str = "docs/decisions-2026-11-08.md";
/// What stands before the registered command on its line of the document.
const REGISTRATION: &str = "git config --local merge.moai-manifest-rs.driver '";
/// Git's own text merge as a driver command: what the registration falls back to.
const TEXT_MERGE: &str = "git merge-file -L ours -L base -L theirs %A %O %B";
const CONFLICT_MARKER: &str = "<<<<<<< ";

const A: &str = "workflows/a.md";
const B: &str = "workflows/b.md";

/// One side of a merge: what it changes in the package.
type Step = fn(&Repo);

fn process_path() -> OsString {
    std::env::var_os("PATH").expect("the test process has a PATH")
}

/// Everything but the clearing and HOME that a git child, or a `moai` child that starts git, gets.
fn scrub(command: &mut Command, path: &OsStr) {
    command
        .env("PATH", path)
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("LC_ALL", "C")
        .stdin(Stdio::null());
}

/// A guarded git command (`common::git`) in `cwd` with a clean environment, the temporary HOME
/// and a fixture identity.
fn git_command(home: &TempHome, path: &OsStr, cwd: &Path) -> Command {
    let mut command = common::git::command();
    command
        .clean_env()
        .env("HOME", home.path())
        .env("USERPROFILE", home.path());
    scrub(&mut command, path);
    command
        .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
        .args(["-c", "commit.gpgsign=false", "-c", "core.autocrlf=false"])
        .args(["-c", "merge.conflictStyle=merge"])
        .current_dir(cwd);
    command
}

fn show(out: &Output) -> String {
    format!(
        "exit {:?}\n--- stdout\n{}--- stderr\n{}",
        out.status.code(),
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn text(bytes: &[u8]) -> String {
    String::from_utf8_lossy(bytes).into_owned()
}

fn origin(rel: &str) -> Value {
    json!({"kind": "design-spec", "template_id": rel, "adoption": "new", "reason": "first"})
}

/// `<stem> line 1` to `<stem> line <count>`: no two files share a line.
fn lines(stem: &str, count: usize) -> String {
    (1..=count).map(|i| format!("{stem} line {i}\n")).collect()
}

fn pretty(value: &Value) -> String {
    serde_json::to_string_pretty(value).unwrap() + "\n"
}

/// The entries of a manifest text.
fn entries(manifest: &[u8]) -> Map<String, Value> {
    let value: Value = serde_json::from_slice(manifest)
        .unwrap_or_else(|e| panic!("the manifest is not JSON ({e}):\n{}", text(manifest)));
    value["files"]
        .as_object()
        .expect("the manifest has a `files` object")
        .clone()
}

/// Every file under the package of `root` except the manifest, relative to the package.
fn package_files(root: &Path) -> BTreeSet<String> {
    let package = root.join(PACKAGE);
    let mut out = BTreeSet::new();
    let mut folders = vec![package.clone()];
    while let Some(folder) = folders.pop() {
        for entry in fs::read_dir(folder).unwrap() {
            let path = entry.unwrap().path();
            if path.is_dir() {
                folders.push(path);
            } else {
                let rel = path.strip_prefix(&package).unwrap();
                out.insert(rel.to_str().unwrap().to_string());
            }
        }
    }
    out.remove(IN_PACKAGE);
    out
}

/// A temporary repository holding a small package, with the manifest driver named for it in
/// `.gitattributes`. Paths given to its methods are relative to the package folder.
struct Repo {
    dir: TempDir,
    root: PathBuf,
    home: TempHome,
    /// The PATH of every git child, and through it of the driver.
    path: OsString,
}

impl Repo {
    /// `main` with one commit; no driver is registered.
    fn new() -> Repo {
        let dir = tempfile::Builder::new()
            .prefix("moai-test-manifest-")
            .tempdir()
            .unwrap();
        let root = dir.path().join("repo");
        fs::create_dir_all(&root).unwrap();
        let repo = Repo {
            dir,
            root,
            home: TempHome::new(),
            path: process_path(),
        };
        repo.git(&["init", "-q", "-b", "main"]);
        repo.name_driver(DRIVER);
        let contents = [
            ("SKILL.md", "# moai\n".to_string()),
            (A, lines("a", 20)),
            (B, lines("b", 20)),
        ];
        let mut files = Map::new();
        for (rel, content) in &contents {
            repo.put(rel, content);
            files.insert(
                rel.to_string(),
                json!({"sha256": "", "origin": origin(rel)}),
            );
        }
        repo.save(&json!({
            "schema_version": 1,
            "covers": "every package file except this manifest",
            "files": files,
        }));
        repo.seal("base");
        repo
    }

    /// `new` with the built binary registered as the driver.
    fn registered() -> Repo {
        let repo = Repo::new();
        repo.register(&format!("'{MOAI_BIN}' {COMMAND} %O %A %B %P >/dev/null"));
        repo
    }

    fn register(&self, command: &str) {
        self.git(&["config", "--local", DRIVER_KEY, command]);
    }

    /// Writes the `.gitattributes` that gives the manifest the merge driver `driver`.
    fn name_driver(&self, driver: &str) {
        let line = format!("{MANIFEST} merge={driver}\n");
        fs::write(self.root.join(".gitattributes"), line).unwrap();
    }

    fn run_in(&self, cwd: &Path, args: &[&str]) -> Output {
        git_command(&self.home, &self.path, cwd)
            .args(args)
            .output()
            .expect("spawn git")
    }

    fn run(&self, args: &[&str]) -> Output {
        self.run_in(&self.root, args)
    }

    /// The raw standard output of a git call that must succeed.
    fn git(&self, args: &[&str]) -> Vec<u8> {
        let out = self.run(args);
        assert!(out.status.success(), "git {args:?}: {}", show(&out));
        out.stdout
    }

    fn file(&self, rel: &str) -> PathBuf {
        self.root.join(PACKAGE).join(rel)
    }

    fn put(&self, rel: &str, content: &str) {
        let path = self.file(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, content).unwrap();
    }

    fn read(&self, rel: &str) -> String {
        fs::read_to_string(self.file(rel)).unwrap_or_else(|e| panic!("read {rel}: {e}"))
    }

    fn manifest(&self) -> Value {
        serde_json::from_str(&self.read(IN_PACKAGE)).unwrap()
    }

    fn save(&self, manifest: &Value) {
        self.put(IN_PACKAGE, &pretty(manifest));
    }

    /// Writes the manifest the way `moai check fix` does.
    fn render(&self) {
        let rendered = render_manifest(&self.root).expect("render the fixture manifest");
        self.put(IN_PACKAGE, &rendered);
    }

    fn commit(&self, message: &str) {
        self.git(&["add", "-A"]);
        self.git(&["commit", "-q", "-m", message]);
    }

    fn seal(&self, message: &str) {
        self.render();
        self.commit(message);
    }

    /// Replaces line `n` (from 1) of a package file.
    fn edit_line(&self, rel: &str, n: usize, new: &str) {
        let old = self.read(rel);
        let mut all: Vec<String> = old.split_inclusive('\n').map(str::to_string).collect();
        all[n - 1] = format!("{new}\n");
        self.put(rel, &all.concat());
    }

    /// A new package file with its manifest entry.
    fn add(&self, rel: &str, content: &str) {
        self.put(rel, content);
        let mut manifest = self.manifest();
        manifest["files"][rel] = json!({"sha256": "", "origin": origin(rel)});
        self.save(&manifest);
    }

    /// A package file removed with its entry.
    fn delete(&self, rel: &str) {
        fs::remove_file(self.file(rel)).unwrap();
        let mut manifest = self.manifest();
        manifest["files"].as_object_mut().unwrap().remove(rel);
        self.save(&manifest);
    }

    /// `git mv` of a package file; its entry moves to the end under the new name.
    fn rename(&self, old: &str, new: &str) {
        fs::create_dir_all(self.file(new).parent().unwrap()).unwrap();
        let from = format!("{PACKAGE}/{old}");
        let to = format!("{PACKAGE}/{new}");
        self.git(&["mv", &from, &to]);
        let mut manifest = self.manifest();
        let files = manifest["files"].as_object_mut().unwrap();
        let entry = files.remove(old).expect("the renamed file has an entry");
        files.insert(new.to_string(), entry);
        self.save(&manifest);
    }

    fn set_origin(&self, rel: &str, key: &str, value: &str) {
        let mut manifest = self.manifest();
        manifest["files"][rel]["origin"][key] = json!(value);
        self.save(&manifest);
    }

    fn set_sha(&self, rel: &str, value: &str) {
        let mut manifest = self.manifest();
        manifest["files"][rel]["sha256"] = json!(value);
        self.save(&manifest);
    }

    /// Commits `theirs` on `side`, then `ours` on `main`. Each step leaves the manifest as it
    /// wants it committed.
    fn diverge_raw(&self, ours: impl FnOnce(&Repo), theirs: impl FnOnce(&Repo)) {
        self.git(&["checkout", "-q", "-b", "side"]);
        theirs(self);
        self.commit("theirs");
        self.git(&["checkout", "-q", "main"]);
        ours(self);
        self.commit("ours");
    }

    /// `diverge_raw` with the manifest rendered on each side.
    fn diverge(&self, ours: impl FnOnce(&Repo), theirs: impl FnOnce(&Repo)) {
        self.diverge_raw(
            |r| {
                ours(r);
                r.render();
            },
            |r| {
                theirs(r);
                r.render();
            },
        );
    }

    /// `git merge side` into `main`, started in the repository root.
    fn merge(&self) -> Output {
        self.run(&["merge", "-q", "--no-edit", "side"])
    }

    fn branches(&self, ours: impl FnOnce(&Repo), theirs: impl FnOnce(&Repo)) -> Output {
        self.diverge(ours, theirs);
        self.merge()
    }

    fn merge_tree(&self) -> Output {
        self.run(&["merge-tree", "--write-tree", "main", "side"])
    }

    fn unmerged(&self) -> Vec<String> {
        let out = text(&self.git(&["ls-files", "-u", "-z"]));
        let paths: BTreeSet<String> = out
            .split('\0')
            .filter_map(|record| record.split_once('\t'))
            .map(|(_, path)| path.to_string())
            .collect();
        paths.into_iter().collect()
    }

    fn blob(&self, rev: &str, path: &str) -> Vec<u8> {
        self.git(&["show", &format!("{rev}:{path}")])
    }

    /// The three manifests git hands the driver for `main` and `side`: common, ours, theirs.
    fn sides(&self) -> [Vec<u8>; 3] {
        let base = text(&self.git(&["merge-base", "main", "side"]));
        [
            self.blob(base.trim(), MANIFEST),
            self.blob("main", MANIFEST),
            self.blob("side", MANIFEST),
        ]
    }

    /// Git's text merge of three texts (common, ours, theirs): the number of conflicts and the
    /// result.
    fn text_merge(&self, sides: &[Vec<u8>; 3]) -> (i32, Vec<u8>) {
        let scratch = tempfile::TempDir::new_in(self.dir.path()).unwrap();
        for (name, bytes) in ["base", "ours", "theirs"].iter().zip(sides) {
            fs::write(scratch.path().join(name), bytes).unwrap();
        }
        let out = self.run_in(
            scratch.path(),
            &[
                "merge-file",
                "-p",
                "-L",
                "ours",
                "-L",
                "base",
                "-L",
                "theirs",
                "ours",
                "base",
                "theirs",
            ],
        );
        let code = out.status.code().expect("git merge-file was not killed");
        assert!((0..128).contains(&code), "git merge-file: {}", show(&out));
        (code, out.stdout)
    }

    /// Git's text merge of the three index stages of the unmerged manifest.
    fn text_merge_of_stages(&self) -> Vec<u8> {
        let stages = [1, 2, 3].map(|n| self.git(&["show", &format!(":{n}:{MANIFEST}")]));
        self.text_merge(&stages).1
    }

    /// The package of a tree object, written into a fresh folder.
    fn materialize(&self, tree: &str) -> TempDir {
        let out = tempfile::TempDir::new_in(self.dir.path()).unwrap();
        let names = text(&self.git(&["ls-tree", "-r", "-z", "--name-only", tree, "--", PACKAGE]));
        for name in names.split('\0').filter(|n| !n.is_empty()) {
            let path = out.path().join(name);
            fs::create_dir_all(path.parent().unwrap()).unwrap();
            fs::write(path, self.blob(tree, name)).unwrap();
        }
        out
    }

    /// The entries whose committed sha256 differs from the file in the checked-out tree.
    fn stale_entries(&self) -> Vec<String> {
        let held = entries(&self.blob("HEAD", MANIFEST));
        let right = entries(render_manifest(&self.root).unwrap().as_bytes());
        assert_eq!(
            held.keys().collect::<Vec<_>>(),
            right.keys().collect::<Vec<_>>()
        );
        held.iter()
            .filter(|(rel, entry)| right[rel.as_str()]["sha256"] != entry["sha256"])
            .map(|(rel, _)| rel.clone())
            .collect()
    }
}

/// A merge that stopped for no skill file: exit 0, nothing unmerged, a merge commit. Returns the
/// committed manifest without judging its content.
fn assert_merge_clean(repo: &Repo, merge: &Output) -> Vec<u8> {
    let unmerged = repo.unmerged();
    let skill: Vec<&String> = unmerged.iter().filter(|p| *p != MANIFEST).collect();
    assert!(
        skill.is_empty(),
        "fixture: the skill files must merge cleanly, unmerged {skill:?}\n{}",
        show(merge)
    );
    assert_eq!(
        merge.status.code(),
        Some(0),
        "the merge must end clean: {}",
        show(merge)
    );
    assert_eq!(unmerged, Vec::<String>::new());
    let parents = text(&repo.git(&["rev-list", "--parents", "-n", "1", "HEAD"]));
    assert_eq!(parents.split_whitespace().count(), 3, "no merge commit");
    assert_eq!(text(&repo.git(&["status", "--porcelain"])), "");
    repo.blob("HEAD", MANIFEST)
}

/// A clean merge whose committed manifest is, byte for byte, what `moai check fix` writes for
/// the merged tree, with one entry per package file.
fn assert_clean_and_correct(repo: &Repo, merge: &Output) {
    let committed = assert_merge_clean(repo, merge);
    assert!(
        !text(&committed).contains(CONFLICT_MARKER),
        "{}",
        text(&committed)
    );
    assert_eq!(
        text(&committed),
        render_manifest(&repo.root).expect("render the merged tree"),
        "the committed manifest is not the one of the merged tree"
    );
    let keys: BTreeSet<String> = entries(&committed).keys().cloned().collect();
    assert_eq!(
        keys,
        package_files(&repo.root),
        "one entry per package file"
    );
}

/// What a refusal says on standard error: the repair command, the command or the path, and each
/// of `reasons` (a path, a setting or a key). Git's own `warning:` lines do not count.
fn assert_names_the_refusal(out: &Output, reasons: &[&str]) {
    let said: String = text(&out.stderr)
        .lines()
        .filter(|line| !line.starts_with("warning:"))
        .map(|line| format!("{line}\n"))
        .collect();
    assert!(
        said.contains(REPAIR),
        "standard error must name `{REPAIR}`: {}",
        show(out)
    );
    assert!(
        said.contains(COMMAND) || said.contains(MANIFEST),
        "standard error must name `{COMMAND}` or the manifest path: {}",
        show(out)
    );
    for reason in reasons {
        assert!(
            said.contains(reason),
            "standard error must name the reason `{reason}`: {}",
            show(out)
        );
    }
}

/// A merge that left the manifest as a conflict with git's text merge of the three versions in it.
fn assert_text_conflict(repo: &Repo, merge: &Output) {
    assert_eq!(
        merge.status.code(),
        Some(1),
        "the merge must stop with a conflict, not succeed or fail: {}",
        show(merge)
    );
    assert!(
        repo.unmerged().iter().any(|p| p == MANIFEST),
        "the manifest must stay unmerged, unmerged {:?}: {}",
        repo.unmerged(),
        show(merge)
    );
    assert_eq!(
        repo.read(IN_PACKAGE),
        text(&repo.text_merge_of_stages()),
        "the manifest must hold git's text merge of the three versions: {}",
        show(merge)
    );
}

fn assert_refused(repo: &Repo, merge: &Output, reasons: &[&str]) {
    assert_text_conflict(repo, merge);
    assert_names_the_refusal(merge, reasons);
}

/// The built command run as git would run it, on three temporary files holding `sides`, from
/// `cwd`, for the path `path`. It must refuse: an exit status from 1 to 127, git's text merge in
/// `%A`, `%O` and `%B` untouched, and the refusal on standard error.
fn assert_direct_refusal(
    repo: &Repo,
    cwd: &Path,
    sides: &[Vec<u8>; 3],
    path: &str,
    reasons: &[&str],
) {
    let scratch = tempfile::TempDir::new_in(repo.dir.path()).unwrap();
    let files = ["O", "A", "B"].map(|name| scratch.path().join(name));
    for (file, bytes) in files.iter().zip(sides) {
        fs::write(file, bytes).unwrap();
    }
    let expected = repo.text_merge(sides).1;
    let out = drive(repo, cwd, &files, path);
    let code = out.status.code();
    assert!(
        code.is_some_and(|c| (1..=127).contains(&c)),
        "a refusal exits from 1 to 127: {}",
        show(&out)
    );
    assert_eq!(
        text(&fs::read(&files[1]).unwrap()),
        text(&expected),
        "%A must hold git's text merge: {}",
        show(&out)
    );
    assert_eq!(fs::read(&files[0]).unwrap(), sides[0], "%O changed");
    assert_eq!(fs::read(&files[2]).unwrap(), sides[2], "%B changed");
    assert_names_the_refusal(&out, reasons);
}

/// `moai manifest-merge-driver %O %A %B %P` with a clean environment and the temporary HOME.
fn drive(repo: &Repo, cwd: &Path, files: &[PathBuf; 3], path: &str) -> Output {
    let mut command = repo.home.command(&[COMMAND]);
    scrub(&mut command, &repo.path);
    command
        .args(files)
        .arg(path)
        .env("GIT_CEILING_DIRECTORIES", ceiling(repo))
        .current_dir(cwd)
        .output()
        .expect("spawn moai")
}

/// The fixture folder with links resolved: git looks for no repository at or above it.
fn ceiling(repo: &Repo) -> PathBuf {
    fs::canonicalize(repo.dir.path()).unwrap()
}

// --- merged: the driver decides every entry ---------------------------------------------------

fn ours_first_line(r: &Repo) {
    r.edit_line(A, 1, "ours");
}

fn theirs_last_line(r: &Repo) {
    r.edit_line(A, 20, "theirs");
}

fn assert_both_edits(content: &str) {
    assert_eq!(content.lines().next(), Some("ours"), "{content}");
    assert_eq!(content.lines().last(), Some("theirs"), "{content}");
}

#[test]
fn merged_one_file_edited_apart_on_both_sides() {
    let r = Repo::registered();
    let merge = r.branches(ours_first_line, theirs_last_line);
    assert_clean_and_correct(&r, &merge);
    assert_both_edits(&r.read(A));
}

#[test]
fn merged_different_files_on_each_side() {
    let r = Repo::registered();
    let merge = r.branches(ours_first_line, |r| r.edit_line(B, 20, "theirs"));
    assert_clean_and_correct(&r, &merge);
}

#[test]
fn merged_a_merge_started_in_a_subfolder() {
    let r = Repo::registered();
    r.diverge(ours_first_line, theirs_last_line);
    let merge = r.run_in(&r.file("workflows"), &["merge", "-q", "--no-edit", "side"]);
    assert_clean_and_correct(&r, &merge);
    assert_both_edits(&r.read(A));
}

#[test]
fn merged_new_files_on_both_sides_keep_sorted_keys() {
    let r = Repo::registered();
    let merge = r.branches(
        |r| r.add("references/x.md", "x\n"),
        |r| r.add("references/y.md", "y\n"),
    );
    assert_clean_and_correct(&r, &merge);
    let keys: Vec<String> = entries(&r.blob("HEAD", MANIFEST)).keys().cloned().collect();
    assert_eq!(
        keys,
        ["SKILL.md", "references/x.md", "references/y.md", A, B]
    );
}

#[test]
fn merged_the_same_new_file_on_both_sides() {
    // Added on both sides with the same contents is no refusal; the other edits make the two
    // manifests differ, so git asks the driver.
    let r = Repo::registered();
    let merge = r.branches(
        |r| {
            r.add("references/x.md", "same\n");
            r.edit_line(A, 1, "ours");
        },
        |r| {
            r.add("references/x.md", "same\n");
            r.edit_line(B, 20, "theirs");
        },
    );
    assert_clean_and_correct(&r, &merge);
}

#[test]
fn merged_a_rename_on_one_side() {
    let r = Repo::registered();
    let merge = r.branches(ours_first_line, |r| {
        r.rename(B, "workflows/c.md");
        r.edit_line(A, 20, "theirs");
    });
    assert_clean_and_correct(&r, &merge);
    assert_both_edits(&r.read(A));
    let files = entries(&r.blob("HEAD", MANIFEST));
    assert!(!files.contains_key(B));
    assert_eq!(files["workflows/c.md"]["origin"], origin(B));
}

#[test]
fn merged_a_folder_renamed_on_one_side_alone() {
    let r = Repo::registered();
    let merge = r.branches(
        |r| r.add("references/x.md", "x\n"),
        |r| {
            r.rename(A, "flows/a.md");
            r.rename(B, "flows/b.md");
        },
    );
    assert_clean_and_correct(&r, &merge);
    let keys: Vec<String> = entries(&r.blob("HEAD", MANIFEST)).keys().cloned().collect();
    assert_eq!(
        keys,
        ["SKILL.md", "flows/a.md", "flows/b.md", "references/x.md"]
    );
}

#[test]
fn merged_a_file_deleted_on_one_side_alone_drops_its_entry() {
    let r = Repo::registered();
    let merge = r.branches(ours_first_line, |r| r.delete(B));
    assert_clean_and_correct(&r, &merge);
    assert!(!entries(&r.blob("HEAD", MANIFEST)).contains_key(B));
}

#[test]
fn merged_origin_keys_changed_apart_are_both_kept() {
    let r = Repo::registered();
    let merge = r.branches(
        |r| {
            r.edit_line(A, 1, "ours");
            r.set_origin(A, "reason", "first, then ours");
        },
        |r| {
            r.edit_line(A, 20, "theirs");
            r.set_origin(A, "adoption", "adapted by theirs");
        },
    );
    assert_clean_and_correct(&r, &merge);
    let origin = entries(&r.blob("HEAD", MANIFEST))[A]["origin"].clone();
    assert_eq!(origin["reason"], "first, then ours");
    assert_eq!(origin["adoption"], "adapted by theirs");
}

#[test]
fn merged_two_branches_change_only_their_own_hash_lines() {
    let r = Repo::registered();
    let base = text(&r.blob("HEAD", MANIFEST));
    let merge = r.branches(
        |r| r.edit_line(A, 1, "card one"),
        |r| r.edit_line(B, 1, "card two"),
    );
    assert_clean_and_correct(&r, &merge);
    let merged = text(&r.blob("HEAD", MANIFEST));
    assert_eq!(base.lines().count(), merged.lines().count());
    let moved: Vec<&str> = base
        .lines()
        .zip(merged.lines())
        .filter(|(old, new)| old != new)
        .map(|(_, new)| new.trim())
        .collect();
    assert_eq!(moved.len(), 2, "{moved:?}");
    assert!(
        moved.iter().all(|l| l.starts_with("\"sha256\":")),
        "{moved:?}"
    );
    let (old, new) = (entries(base.as_bytes()), entries(merged.as_bytes()));
    let changed: Vec<&String> = new
        .keys()
        .filter(|k| new[k.as_str()]["sha256"] != old[k.as_str()]["sha256"])
        .collect();
    assert_eq!(changed, [A, B]);
}

#[test]
fn merged_by_merge_tree_with_the_hashes_of_the_merged_files() {
    // `merge-tree` writes no work tree: a hash read from disk would be our side's.
    let r = Repo::registered();
    r.diverge(ours_first_line, theirs_last_line);
    let out = r.merge_tree();
    let tree = text(&out.stdout)
        .lines()
        .next()
        .unwrap_or_default()
        .to_string();
    assert_both_edits(&text(&r.blob(&tree, &format!("{PACKAGE}/{A}"))));
    assert_eq!(
        out.status.code(),
        Some(0),
        "merge-tree must end clean: {}",
        show(&out)
    );
    let written = r.materialize(&tree);
    assert_eq!(
        text(&r.blob(&tree, MANIFEST)),
        render_manifest(written.path()).unwrap()
    );
    let keys: BTreeSet<String> = entries(&r.blob(&tree, MANIFEST)).keys().cloned().collect();
    assert_eq!(keys, package_files(written.path()));
}

#[test]
fn merged_with_the_histogram_diff_of_git() {
    // On this triple histogram and myers both merge cleanly but give different text; git's own
    // merge gives the histogram text, so a myers hash would not match the merged file.
    let (base, ours, theirs) = ("\nb\nx\nx\nx\n", "\nb\nb\nx\nx\n", "\nb\nx\nx\n");
    let merged = "\nb\nb\nx\nx\n";
    let r = Repo::registered();
    let t = "references/t.md";
    r.add(t, base);
    r.seal("t");

    for (name, content) in [("o", base), ("a", ours), ("b", theirs)] {
        fs::write(r.dir.path().join(name), content).unwrap();
    }
    let by = |algorithm: &str| {
        let flag = format!("--diff-algorithm={algorithm}");
        r.run_in(r.dir.path(), &["merge-file", "-p", &flag, "a", "o", "b"])
    };
    let (histogram, myers) = (by("histogram"), by("myers"));
    assert_eq!(histogram.status.code(), Some(0), "{}", show(&histogram));
    assert_eq!(myers.status.code(), Some(0), "{}", show(&myers));
    assert_eq!(text(&histogram.stdout), merged, "fixture: histogram text");
    assert_ne!(text(&myers.stdout), merged, "fixture: myers must differ");

    let merge = r.branches(|r| r.put(t, ours), |r| r.put(t, theirs));
    assert_clean_and_correct(&r, &merge);
    assert_eq!(r.read(t), merged);
}

#[test]
fn merged_with_the_git_settings_the_driver_repeats_set_explicitly() {
    for (key, value) in [
        ("diff.algorithm", "histogram"),
        ("merge.renormalize", "false"),
    ] {
        let r = Repo::registered();
        r.git(&["config", "--local", key, value]);
        let merge = r.branches(ours_first_line, theirs_last_line);
        assert_clean_and_correct(&r, &merge);
    }
}

#[test]
fn merged_a_side_with_an_unwritten_edit_is_not_repaired_by_the_merge() {
    // The driver trusts each side's sha256: it merges the versions the manifests name, so an
    // edit committed without writing the manifest still shows after the merge.
    let r = Repo::registered();
    r.git(&["checkout", "-q", "-b", "side"]);
    r.edit_line(A, 20, "theirs");
    r.seal("theirs, manifest written");
    r.edit_line(A, 19, "theirs, not written");
    r.commit("theirs, manifest stale");
    r.git(&["checkout", "-q", "main"]);
    r.edit_line(A, 1, "ours");
    r.seal("ours");
    let merge = r.merge();
    assert_merge_clean(&r, &merge);
    assert_eq!(r.stale_entries(), [A]);
}

#[test]
fn merged_a_hand_typed_hash_is_not_repaired_by_the_merge() {
    let r = Repo::registered();
    let typed = "f".repeat(64);
    r.diverge_raw(
        |r| {
            r.edit_line(A, 1, "card one");
            r.render();
        },
        |r| {
            r.edit_line(B, 1, "card two");
            r.render();
            r.set_sha(B, &typed);
        },
    );
    let merge = r.merge();
    let committed = assert_merge_clean(&r, &merge);
    assert_eq!(r.stale_entries(), [B]);
    assert_eq!(entries(&committed)[B]["sha256"], typed.as_str());
}

#[test]
fn merged_a_side_the_driver_wrote_never_gives_a_later_text_merge_a_key_twice() {
    let x = "references/x.md";
    let r = Repo::new();
    let base = text(&r.git(&["rev-parse", "HEAD"])).trim().to_string();
    r.git(&["checkout", "-q", "-b", "feat"]);
    r.add(x, "x\n");
    r.seal("feat adds x");
    r.git(&["checkout", "-q", "main"]);
    r.add(x, "x\n");
    r.seal("main adds x on its own");
    r.git(&["checkout", "-q", "-b", "side", &base]);
    r.register(&format!("'{MOAI_BIN}' {COMMAND} %O %A %B %P >/dev/null"));
    r.add("references/y.md", "y\n");
    r.seal("side adds y");
    let through_driver = r.run(&["merge", "-q", "--no-edit", "feat"]);
    assert_clean_and_correct(&r, &through_driver);
    // The last merge as a server without the driver does it: text only.
    r.git(&["config", "--local", "--unset-all", DRIVER_KEY]);
    r.git(&["checkout", "-q", "main"]);
    let by_text = r.merge();
    let held = r.read(IN_PACKAGE);
    if by_text.status.success() {
        assert_eq!(held.matches("\"references/x.md\": {").count(), 1, "{held}");
        assert_eq!(held, render_manifest(&r.root).unwrap());
    } else {
        assert!(held.contains(CONFLICT_MARKER), "{held}");
    }
}

#[test]
fn refused_a_file_with_its_own_merge_attribute_changed_on_one_side_alone() {
    // The attribute refuses every skill file the merge changes. Here both sides change A, while
    // B, which has the attribute, changes on one side: git takes B as it is, and the manifest
    // is still left to a person.
    let own_attribute: Step = |r| {
        let path = r.root.join(".gitattributes");
        let mut attributes = fs::read_to_string(&path).unwrap();
        attributes.push_str(&format!("{PACKAGE}/{B} merge=binary\n"));
        fs::write(path, attributes).unwrap();
    };
    for b_changed_by_ours in [true, false] {
        let r = Repo::registered();
        let merge = r.branches(
            |r| {
                own_attribute(r);
                ours_first_line(r);
                if b_changed_by_ours {
                    r.edit_line(B, 1, "ours");
                }
            },
            |r| {
                theirs_last_line(r);
                if !b_changed_by_ours {
                    r.edit_line(B, 20, "theirs");
                }
            },
        );
        assert_eq!(r.unmerged(), [MANIFEST], "ours: {b_changed_by_ours}");
        assert_refused(&r, &merge, &[B, "attribute"]);
    }
}

// --- refused: the manifest stays a conflict ---------------------------------------------------

#[test]
fn refused_a_real_conflict_in_a_skill_file() {
    let r = Repo::registered();
    let merge = r.branches(
        |r| r.edit_line(A, 10, "ours"),
        |r| r.edit_line(A, 10, "theirs"),
    );
    assert_eq!(
        r.unmerged(),
        [MANIFEST.to_string(), format!("{PACKAGE}/{A}")]
    );
    assert!(r.read(A).contains(CONFLICT_MARKER));
    assert_refused(&r, &merge, &[A]);
    assert!(r.read(IN_PACKAGE).contains(CONFLICT_MARKER));
}

#[test]
fn refused_a_rename_against_an_edit() {
    let edit: Step = |r| r.edit_line(B, 1, "edited");
    let rename: Step = |r| r.rename(B, "workflows/c.md");
    for (ours, theirs) in [(edit, rename), (rename, edit)] {
        let r = Repo::registered();
        let merge = r.branches(ours, theirs);
        assert_eq!(r.unmerged(), [MANIFEST]);
        assert_eq!(r.read("workflows/c.md").lines().next(), Some("edited"));
        assert_refused(&r, &merge, &[B]);
    }
}

#[test]
fn refused_an_origin_changed_on_both_sides() {
    let r = Repo::registered();
    let merge = r.branches(
        |r| {
            r.edit_line(A, 1, "ours");
            r.set_origin(A, "reason", "first, ours");
        },
        |r| {
            r.edit_line(A, 20, "theirs");
            r.set_origin(A, "reason", "first, theirs");
        },
    );
    assert_eq!(r.unmerged(), [MANIFEST]);
    assert_refused(&r, &merge, &[A, "origin"]);
    assert!(r.read(IN_PACKAGE).contains(CONFLICT_MARKER));
}

#[test]
fn refused_a_skill_file_with_its_own_merge_attribute() {
    let r = Repo::registered();
    let merge = r.branches(
        |r| {
            let path = r.root.join(".gitattributes");
            let mut attributes = fs::read_to_string(&path).unwrap();
            attributes.push_str(&format!("{PACKAGE}/{A} merge=binary\n"));
            fs::write(path, attributes).unwrap();
            r.edit_line(A, 1, "ours");
        },
        theirs_last_line,
    );
    assert!(r.unmerged().contains(&format!("{PACKAGE}/{A}")));
    assert_refused(&r, &merge, &[A, "attribute"]);
}

/// Both sides edit one file apart under a git setting the driver cannot repeat.
fn assert_setting_refused(key: &str, value: &str) {
    let r = Repo::registered();
    r.git(&["config", "--local", key, value]);
    let merge = r.branches(ours_first_line, theirs_last_line);
    assert_eq!(r.unmerged(), [MANIFEST], "{key}={value}");
    assert_refused(&r, &merge, &[key]);
}

#[test]
fn refused_a_diff_algorithm_other_than_histogram() {
    for value in ["myers", "patience"] {
        assert_setting_refused("diff.algorithm", value);
    }
}

#[test]
fn refused_a_diff_algorithm_given_on_the_git_command_line() {
    // `git -c` reaches the command only through the environment git hands the driver.
    let r = Repo::registered();
    r.diverge(ours_first_line, theirs_last_line);
    let myers = "diff.algorithm=myers";
    let merge = r.run(&["-c", myers, "merge", "-q", "--no-edit", "side"]);
    assert_eq!(r.unmerged(), [MANIFEST]);
    assert_refused(&r, &merge, &["diff.algorithm"]);
}

#[test]
fn refused_when_merge_renormalize_is_on() {
    for value in ["true", "yes"] {
        assert_setting_refused("merge.renormalize", value);
    }
}

#[test]
fn refused_a_folder_renamed_on_one_side_while_the_other_adds_a_file_in_it() {
    // Git's directory-rename detection may move the new file, which the manifests cannot show.
    let add: Step = |r| r.add("workflows/n.md", "n\n");
    let rename: Step = |r| {
        r.rename(A, "flows/a.md");
        r.rename(B, "flows/b.md");
    };
    for (ours, theirs) in [(add, rename), (rename, add)] {
        let r = Repo::registered();
        let merge = r.branches(ours, theirs);
        assert_refused(&r, &merge, &["workflows/n.md"]);
    }
}

#[test]
fn refused_a_file_deleted_on_one_side_and_changed_on_the_other() {
    let change: Step = |r| r.edit_line(B, 1, "changed");
    let delete: Step = |r| r.delete(B);
    for (ours, theirs) in [(change, delete), (delete, change)] {
        let r = Repo::registered();
        let merge = r.branches(ours, theirs);
        assert!(r.unmerged().contains(&format!("{PACKAGE}/{B}")));
        assert_refused(&r, &merge, &[B]);
    }
}

#[test]
fn refused_a_file_added_on_both_sides_with_different_contents() {
    let r = Repo::registered();
    let x = "references/x.md";
    let merge = r.branches(|r| r.add(x, "ours\n"), |r| r.add(x, "theirs\n"));
    assert!(r.unmerged().contains(&format!("{PACKAGE}/{x}")));
    assert_refused(&r, &merge, &[x]);
}

#[test]
fn refused_a_version_not_found_in_the_history() {
    let plain: Step = |r| {
        r.edit_line(A, 1, "plain");
        r.render();
    };
    let made_up: Step = |r| {
        r.edit_line(A, 20, "made up");
        r.render();
        r.set_sha(A, &"e".repeat(64));
    };
    for (ours, theirs) in [(plain, made_up), (made_up, plain)] {
        let r = Repo::registered();
        r.diverge_raw(ours, theirs);
        let merge = r.merge();
        assert_eq!(r.unmerged(), [MANIFEST]);
        assert_refused(&r, &merge, &[A]);
    }
}

/// One side commits the manifest `broken` makes of its own; the other side changes `b.md` and
/// writes the manifest. Both directions.
fn assert_broken_side_refused(broken: Step, reasons: &[&str]) {
    let sound: Step = |r| {
        r.edit_line(B, 20, "sound");
        r.render();
    };
    for (ours, theirs) in [(sound, broken), (broken, sound)] {
        let r = Repo::registered();
        r.diverge_raw(ours, theirs);
        let merge = r.merge();
        assert_eq!(r.unmerged(), [MANIFEST]);
        assert_refused(&r, &merge, reasons);
    }
}

#[test]
fn refused_a_side_with_a_key_twice() {
    // The text merge of such a side can be clean; the manifest still stays unmerged.
    assert_broken_side_refused(
        |r| {
            let held = r.read(IN_PACKAGE);
            let first = held.find("    \"workflows/a.md\"").unwrap();
            let next = held.find("    \"workflows/b.md\"").unwrap();
            let entry = &held[first..next];
            let twice = format!("{}{entry}{}", &held[..next], &held[next..]);
            assert_eq!(twice.matches("\"workflows/a.md\": {").count(), 2);
            r.put(IN_PACKAGE, &twice);
        },
        &[A],
    );
}

#[test]
fn refused_a_side_that_is_not_json() {
    assert_broken_side_refused(|r| r.put(IN_PACKAGE, "{ not json\n"), &["JSON"]);
}

#[test]
fn refused_a_side_without_a_files_object() {
    assert_broken_side_refused(
        |r| r.put(IN_PACKAGE, "{\n  \"schema_version\": 1\n}\n"),
        &["files"],
    );
}

#[test]
fn refused_a_direct_refusal_exits_between_1_and_127() {
    let r = Repo::new();
    r.diverge(
        |r| r.set_origin(A, "reason", "first, ours"),
        |r| r.set_origin(A, "reason", "first, theirs"),
    );
    assert_direct_refusal(&r, &r.root, &r.sides(), MANIFEST, &[A]);
}

#[test]
fn refused_a_side_that_is_not_utf8() {
    let r = Repo::new();
    r.diverge(ours_first_line, theirs_last_line);
    let mut sides = r.sides();
    sides[2].extend_from_slice(b"\xff\xfe\n");
    // The reason is asked for: bytes read lossily would be refused as broken JSON instead.
    assert_direct_refusal(&r, &r.root, &sides, MANIFEST, &["their side", "UTF-8"]);
}

#[test]
fn refused_a_path_that_is_not_a_package_manifest() {
    let r = Repo::new();
    r.diverge(ours_first_line, theirs_last_line);
    // The reason is asked for: another path would also fail later, on versions not found.
    let reason = ["docs/other.json is not a package manifest"];
    assert_direct_refusal(&r, &r.root, &r.sides(), "docs/other.json", &reason);
}

#[test]
fn refused_an_entry_the_common_version_holds_as_null() {
    // Both sides hold the same entry, so only the common version can refuse it.
    let r = Repo::new();
    let held = r.read(IN_PACKAGE);
    let mut base: Value = serde_json::from_str(&held).unwrap();
    base["files"][A] = Value::Null;
    let side = held.into_bytes();
    let sides = [pretty(&base).into_bytes(), side.clone(), side];
    assert_direct_refusal(&r, &r.root, &sides, MANIFEST, &[A, "not an object"]);
}

#[test]
fn refused_when_run_outside_a_repository_keeps_the_text_merge_and_exits_below_128() {
    // Both sides changed one file, so deciding its hash needs the history: git fails here.
    let r = Repo::new();
    r.diverge(ours_first_line, theirs_last_line);
    let sides = r.sides();
    assert!(
        text(&r.text_merge(&sides).1).contains(CONFLICT_MARKER),
        "fixture: the two hash lines conflict as text"
    );
    let outside = r.dir.path().join("outside");
    fs::create_dir_all(&outside).unwrap();
    let probe = git_command(&r.home, &r.path, &outside)
        .env("GIT_CEILING_DIRECTORIES", ceiling(&r))
        .args(["rev-parse", "--git-dir"])
        .output()
        .unwrap();
    assert!(
        !probe.status.success(),
        "fixture: {} is inside a repository",
        outside.display()
    );
    assert_direct_refusal(&r, &outside, &sides, MANIFEST, &[]);
}

// --- registered: the command the document tells a person to register ---------------------------

/// The command of the document's registration line: the text between the quote that ends
/// `REGISTRATION` and the last quote of that line.
fn registration_command() -> String {
    let document = fs::read_to_string(Path::new(ROOT).join(DOCUMENT)).unwrap();
    let found: Vec<&str> = document
        .lines()
        .filter(|line| line.contains(REGISTRATION))
        .collect();
    assert_eq!(
        found.len(),
        1,
        "registration line missing from {DOCUMENT} (or written twice): want exactly one line \
         with `{REGISTRATION}...'`, found {found:?}"
    );
    let line = found[0];
    let start = line.find(REGISTRATION).unwrap() + REGISTRATION.len();
    let end = line.rfind('\'').unwrap();
    assert!(end > start, "the registration line has no closing quote");
    let command = &line[start..end];
    assert!(!command.contains('\''), "no single quote inside: {command}");
    command.to_string()
}

/// What stands for `moai` on the PATH of a registered case.
enum Moai {
    Built,
    Absent,
    /// A script that records each call in this file and exits 2.
    Unknowing(PathBuf),
}

/// Sets the repository's PATH to a folder of its own (`git`, and `moai` as asked) followed by
/// the system folders, which must hold no `moai`. Returns that folder.
fn limit_path(repo: &mut Repo, moai: Moai) -> PathBuf {
    let system = ["/usr/bin", "/bin"];
    for folder in system {
        let installed = Path::new(folder).join("moai");
        assert!(
            fs::symlink_metadata(&installed).is_err(),
            "fixture: {} exists",
            installed.display()
        );
    }
    let shim = repo.dir.path().join("shim");
    fs::create_dir_all(&shim).unwrap();
    let git = std::env::split_paths(&process_path())
        .map(|folder| folder.join("git"))
        .find(|candidate| candidate.is_file())
        .expect("git is on the PATH");
    std::os::unix::fs::symlink(git, shim.join("git")).unwrap();
    match moai {
        Moai::Built => std::os::unix::fs::symlink(MOAI_BIN, shim.join("moai")).unwrap(),
        Moai::Absent => {}
        Moai::Unknowing(calls) => common::write_executable(
            &shim.join("moai"),
            &format!(
                "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '{}'\nexit 2\n",
                calls.display()
            ),
        ),
    }
    let folders = std::iter::once(shim.clone()).chain(system.map(PathBuf::from));
    repo.path = std::env::join_paths(folders).unwrap();
    shim
}

#[test]
fn registered_a_clone_with_only_the_old_driver_line_runs_no_python() {
    // A clone that still has the earlier installer's line, and nothing under the new name, must
    // get git's own text merge: no attribute of this layout names the earlier driver.
    let interpreter = format!("{}{}", "python", 3);
    let mut r = Repo::new();
    r.diverge(ours_first_line, theirs_last_line);
    let shim = limit_path(&mut r, Moai::Absent);
    let marker = r.dir.path().join("interpreter-ran");
    let fake = shim.join(&interpreter);
    common::write_executable(
        &fake,
        &format!(
            "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '{}'\nexit 1\n",
            marker.display()
        ),
    );
    // Asked the way a driver command is run: only the script above may answer, never an
    // installed interpreter.
    let lookup = format!("alias.which-interpreter=!command -v {interpreter}");
    let resolved = text(&r.git(&["-c", &lookup, "which-interpreter"]));
    assert_eq!(resolved.trim(), fake.to_str().unwrap(), "fixture");

    // What git itself does with no driver registered at all.
    let unregistered = r.merge();
    assert_eq!(
        unregistered.status.code(),
        Some(1),
        "{}",
        show(&unregistered)
    );
    let by_git = r.read(IN_PACKAGE);
    r.git(&["merge", "--abort"]);

    // The earlier installer's registration, with the copied script it tests for in place.
    let copy = r.root.join(".git").join("moai-manifest-driver");
    fs::create_dir_all(&copy).unwrap();
    let script = copy.join(format!("{}.py", python_driver()));
    fs::write(&script, "").unwrap();
    let script = script.display();
    let installed = format!(
        "if test -f '{script}' && command -v {interpreter} >/dev/null 2>&1; \
         then {interpreter} -I '{script}' %O %A %B %P; \
         else {TEXT_MERGE}; fi"
    );
    r.git(&["config", "--local", OLD_DRIVER_KEY, &installed]);

    let merge = r.merge();
    assert!(
        !marker.exists(),
        "the earlier driver ran the interpreter: {}",
        show(&merge)
    );
    assert_eq!(merge.status.code(), Some(1), "{}", show(&merge));
    assert_eq!(r.unmerged(), [MANIFEST], "{}", show(&merge));
    assert!(r.read(IN_PACKAGE).contains(CONFLICT_MARKER));
    assert_eq!(r.read(IN_PACKAGE), by_git, "not git's own text merge");
    r.git(&["merge", "--abort"]);

    // Positive control: with the earlier attribute the same setup does call the interpreter.
    r.name_driver(OLD_DRIVER);
    r.commit("the earlier attribute");
    let control = r.merge();
    assert!(
        marker.is_file(),
        "control: the setup never calls the interpreter, so the case above proves nothing: {}",
        show(&control)
    );
}

#[test]
fn registered_the_attributes_of_the_checkout_that_merges_decide_the_driver() {
    // Measures git: each registered driver only records its name. The base carries the earlier
    // attribute and one side changes it to the new one.
    let ran_when = |main_names: &str, side_names: &str| {
        let r = Repo::new();
        r.name_driver(OLD_DRIVER);
        r.commit("the earlier attribute");
        let marker = r.dir.path().join("ran");
        for (key, word) in [(OLD_DRIVER_KEY, "old"), (DRIVER_KEY, "new")] {
            let record = format!("echo {word} >> '{}'; exit 1", marker.display());
            r.git(&["config", "--local", key, &record]);
        }
        r.diverge(
            |r| {
                r.name_driver(main_names);
                ours_first_line(r);
            },
            |r| {
                r.name_driver(side_names);
                theirs_last_line(r);
            },
        );
        let merge = r.merge();
        assert_eq!(r.unmerged(), [MANIFEST], "{}", show(&merge));
        // The merged work tree holds the new attribute whichever driver ran.
        let attributes = fs::read_to_string(r.root.join(".gitattributes")).unwrap();
        assert_eq!(attributes, format!("{MANIFEST} merge={DRIVER}\n"));
        fs::read_to_string(&marker).unwrap_or_default()
    };
    let measured = [
        // The checked-out branch has the earlier line, the incoming branch the new one.
        ran_when(OLD_DRIVER, DRIVER),
        // The reverse.
        ran_when(DRIVER, OLD_DRIVER),
    ];
    assert_eq!(measured, ["old\n", "new\n"]);
}

#[test]
fn registered_command_merges_cleanly_with_moai_on_the_path() {
    let command = registration_command();
    let mut r = Repo::new();
    r.register(&command);
    r.diverge(ours_first_line, theirs_last_line);
    limit_path(&mut r, Moai::Built);
    let merge = r.merge();
    assert_clean_and_correct(&r, &merge);
    assert_both_edits(&r.read(A));
}

#[test]
fn registered_command_keeps_the_text_conflict_without_moai() {
    let command = registration_command();
    let mut r = Repo::new();
    r.register(&command);
    r.diverge(ours_first_line, theirs_last_line);
    limit_path(&mut r, Moai::Absent);
    let merge = r.merge();
    assert_eq!(r.unmerged(), [MANIFEST], "{}", show(&merge));
    assert_text_conflict(&r, &merge);
    assert!(r.read(IN_PACKAGE).contains(CONFLICT_MARKER));
}

#[test]
fn registered_command_keeps_the_text_conflict_when_moai_does_not_know_it() {
    let command = registration_command();
    let mut r = Repo::new();
    r.register(&command);
    r.diverge(ours_first_line, theirs_last_line);
    let calls = r.dir.path().join("calls");
    limit_path(&mut r, Moai::Unknowing(calls.clone()));
    let merge = r.merge();
    assert!(
        calls.is_file(),
        "the moai on the PATH was never asked: {}",
        show(&merge)
    );
    assert_eq!(r.unmerged(), [MANIFEST], "{}", show(&merge));
    assert_text_conflict(&r, &merge);
    assert!(r.read(IN_PACKAGE).contains(CONFLICT_MARKER));
}

#[test]
fn registered_command_keeps_a_refusal_whose_text_merge_is_clean() {
    // The command refuses under this setting although git can merge the two manifests as text:
    // a registration that ran the text merge after a refusal would resolve what was refused.
    let command = registration_command();
    let mut r = Repo::new();
    r.register(&command);
    r.git(&["config", "--local", "diff.algorithm", "myers"]);
    r.diverge(ours_first_line, |r| r.edit_line(B, 20, "theirs"));
    assert_eq!(
        r.text_merge(&r.sides()).0,
        0,
        "fixture: the text merge of the manifests is clean"
    );
    limit_path(&mut r, Moai::Built);
    let merge = r.merge();
    assert_eq!(r.unmerged(), [MANIFEST], "{}", show(&merge));
    assert_refused(&r, &merge, &["diff.algorithm"]);
}

#[test]
fn registered_command_leaves_merge_tree_output_as_the_text_merge_does() {
    let command = registration_command();
    let mut r = Repo::new();
    r.diverge(
        |r| r.edit_line(A, 10, "ours"),
        |r| r.edit_line(A, 10, "theirs"),
    );
    limit_path(&mut r, Moai::Built);
    r.register(&command);
    let registered = r.merge_tree();
    r.register(TEXT_MERGE);
    let by_text = r.merge_tree();
    assert_eq!(by_text.status.code(), Some(1), "{}", show(&by_text));
    assert_eq!(registered.status.code(), Some(1), "{}", show(&registered));
    assert_eq!(text(&registered.stdout), text(&by_text.stdout));
    // The built command was asked and refused; the equal output is not the fallback's.
    assert_names_the_refusal(&registered, &[A]);
}

#[test]
fn registered_command_adds_nothing_to_a_clean_merge_tree() {
    let command = registration_command();
    let mut r = Repo::new();
    r.register(&command);
    r.diverge(ours_first_line, theirs_last_line);
    limit_path(&mut r, Moai::Built);
    let out = r.merge_tree();
    assert_eq!(out.status.code(), Some(0), "{}", show(&out));
    let said = text(&out.stdout);
    let tree = said.strip_suffix('\n').unwrap_or(&said);
    assert!(
        tree.len() == 40 && tree.bytes().all(|b| b.is_ascii_hexdigit()),
        "standard output must be the tree id alone: {said:?}"
    );
    let written = r.materialize(tree);
    assert_eq!(
        text(&r.blob(tree, MANIFEST)),
        render_manifest(written.path()).unwrap()
    );
}

// --- order: the bytes do not depend on the order of the entries --------------------------------

const ORDER_CASES: u64 = 30;
const ORDER_SEED: u64 = 0x0C55_4202_6112_6000;

/// One change a side of a generated case makes.
enum Change {
    Edit(String, usize, String),
    Add(String, String),
    Delete(String),
    Origin(String, &'static str, String),
}

fn apply(repo: &Repo, changes: &[Change]) {
    for change in changes {
        match change {
            Change::Edit(rel, line, new) => repo.edit_line(rel, *line, new),
            Change::Add(rel, content) => repo.add(rel, content),
            Change::Delete(rel) => repo.delete(rel),
            Change::Origin(rel, key, value) => repo.set_origin(rel, key, value),
        }
    }
}

/// The manifest text with its `files` entries in the order of `keys`.
fn reordered(manifest: &[u8], keys: &[String]) -> Vec<u8> {
    let mut value: Value = serde_json::from_slice(manifest).unwrap();
    let held = value["files"].as_object().unwrap();
    let files: Map<String, Value> = keys.iter().map(|k| (k.clone(), held[k].clone())).collect();
    value["files"] = Value::Object(files);
    pretty(&value).into_bytes()
}

/// A case whose sides merge cleanly: `a.md` is edited apart by both, other files are left,
/// edited or deleted, each side adds files, and `origin` keys change apart. No folder
/// disappears (`a.md` and `keep.md` stay).
fn order_case(case: u64) {
    let context = format!("case {case} of {ORDER_CASES}, seed {ORDER_SEED:#x}");
    let mut rng = prop::Rng::new(ORDER_SEED ^ case);
    let r = Repo::new();
    let mut model: BTreeMap<String, Value> = ["SKILL.md", A, B]
        .iter()
        .map(|rel| (rel.to_string(), origin(rel)))
        .collect();
    let mut optional = vec![(B.to_string(), 20)];
    let mut base_files = vec!["references/keep.md".to_string()];
    for i in 0..1 + rng.below(4) {
        let rel = format!("references/r{i}.md");
        optional.push((rel.clone(), 10));
        base_files.push(rel);
    }
    for rel in &base_files {
        r.add(rel, &lines(rel, 10));
        model.insert(rel.clone(), origin(rel));
    }
    r.seal("the case's base");

    let mut ours = vec![Change::Edit(A.into(), 1, format!("ours {case}"))];
    let mut theirs = vec![Change::Edit(A.into(), 20, format!("theirs {case}"))];
    for (rel, count) in optional {
        let first = Change::Edit(rel.clone(), 1, format!("ours {rel}"));
        let last = Change::Edit(rel.clone(), count, format!("theirs {rel}"));
        match rng.below(6) {
            0 => {}
            1 => ours.push(first),
            2 => theirs.push(last),
            3 => {
                ours.push(first);
                theirs.push(last);
            }
            side => {
                model.remove(&rel);
                let change = Change::Delete(rel);
                if side == 4 {
                    ours.push(change);
                } else {
                    theirs.push(change);
                }
            }
        }
    }
    let kept: Vec<String> = model.keys().cloned().collect();
    for (changes, who, key) in [
        (&mut ours, "ours", "reason"),
        (&mut theirs, "theirs", "adoption"),
    ] {
        for i in 0..1 + rng.below(3) {
            let folder = rng.pick(&["references", "workflows", "docs", "references/deep"]);
            let rel = format!("{folder}/{who}-{i}-{}.md", rng.name(6));
            changes.push(Change::Add(rel.clone(), lines(&rel, 2)));
            model.insert(rel.clone(), origin(&rel));
        }
        if rng.bool() {
            let rel = rng.pick(&kept).clone();
            let value = format!("{who} {case}");
            model.get_mut(&rel).unwrap()[key] = json!(value);
            changes.push(Change::Origin(rel, key, value));
        }
    }
    if rng.bool() {
        let rel = "references/same.md".to_string();
        for changes in [&mut ours, &mut theirs] {
            changes.push(Change::Add(rel.clone(), lines(&rel, 2)));
        }
        model.insert(rel.clone(), origin(&rel));
    }
    r.diverge(|r| apply(r, &ours), |r| apply(r, &theirs));

    // The merged skill files, from git alone: `true` keeps our manifest and exits 0.
    r.register("true");
    let out = r.merge_tree();
    assert_eq!(
        out.status.code(),
        Some(0),
        "fixture ({context}): the skill files merge cleanly: {}",
        show(&out)
    );
    let tree = text(&out.stdout).trim().to_string();
    let merged = r.materialize(&tree);
    let keys: BTreeSet<String> = model.keys().cloned().collect();
    assert_eq!(
        package_files(merged.path()),
        keys,
        "fixture ({context}): the merged files"
    );
    let files: Map<String, Value> = model
        .iter()
        .map(|(rel, origin)| (rel.clone(), json!({"sha256": "", "origin": origin})))
        .collect();
    fs::write(
        merged.path().join(MANIFEST),
        pretty(&json!({
            "schema_version": 1,
            "covers": "every package file except this manifest",
            "files": files,
        })),
    )
    .unwrap();
    let expected = render_manifest(merged.path()).unwrap();

    let sides = r.sides();
    for order in 0..3 {
        let given = sides.clone().map(|side| {
            let mut keys: Vec<String> = entries(&side).keys().cloned().collect();
            match order {
                0 => {}
                1 => keys.reverse(),
                _ => keys = rng.shuffled(&keys),
            }
            reordered(&side, &keys)
        });
        let scratch = tempfile::TempDir::new_in(r.dir.path()).unwrap();
        let paths = ["O", "A", "B"].map(|name| scratch.path().join(name));
        for (path, bytes) in paths.iter().zip(&given) {
            fs::write(path, bytes).unwrap();
        }
        let out = drive(&r, &r.root, &paths, MANIFEST);
        assert_eq!(
            out.status.code(),
            Some(0),
            "{context}, order {order}: the command must decide: {}",
            show(&out)
        );
        assert_eq!(
            text(&fs::read(&paths[1]).unwrap()),
            expected,
            "{context}, order {order}"
        );
    }
}

#[test]
fn order_of_the_entries_never_changes_the_merged_bytes() {
    for case in 0..ORDER_CASES {
        order_case(case);
    }
}

#[test]
fn order_rendering_the_merged_manifest_again_moves_no_byte() {
    let r = Repo::registered();
    let merge = r.branches(
        |r| {
            r.edit_line(A, 1, "ours");
            r.set_origin(A, "reason", "first — 우리 쪽 \"quoted\"\ttab");
        },
        |r| {
            r.edit_line(A, 20, "theirs");
            r.add("references/z.md", "z\n");
        },
    );
    assert_clean_and_correct(&r, &merge);
    let merged = r.blob("HEAD", MANIFEST);
    assert!(text(&merged).contains("우리 쪽"), "{}", text(&merged));

    // The command given its own output as every side writes it back as it is.
    let scratch = tempfile::TempDir::new_in(r.dir.path()).unwrap();
    let files = ["O", "A", "B"].map(|name| scratch.path().join(name));
    for file in &files {
        fs::write(file, &merged).unwrap();
    }
    let out = drive(&r, &r.root, &files, MANIFEST);
    assert_eq!(out.status.code(), Some(0), "{}", show(&out));
    assert_eq!(text(&fs::read(&files[1]).unwrap()), text(&merged));

    // So does `moai check fix` rendering the manifest that holds those bytes.
    r.put(IN_PACKAGE, &text(&merged));
    assert_eq!(render_manifest(&r.root).unwrap(), text(&merged));
}

// --- text: what this repository says -----------------------------------------------------------

/// The Python driver's file stem, built so this file does not name it.
fn python_driver() -> String {
    ["merge", "driver", "manifest"].join("_")
}

#[test]
fn text_gitattributes_names_the_new_driver_and_the_command_and_not_the_python_driver() {
    let attributes = fs::read_to_string(Path::new(ROOT).join(".gitattributes")).unwrap();
    // Whole words of the attribute lines: the earlier name is a prefix of the new one.
    let rules: Vec<Vec<&str>> = attributes
        .lines()
        .filter(|line| !line.trim_start().starts_with('#'))
        .map(|line| line.split_whitespace().collect())
        .collect();
    let named = format!("merge={DRIVER}");
    assert!(
        rules.iter().any(|rule| rule == &[MANIFEST, named.as_str()]),
        ".gitattributes has no line `{MANIFEST} {named}`:\n{attributes}"
    );
    let earlier = format!("merge={OLD_DRIVER}");
    assert!(
        !rules.iter().flatten().any(|word| *word == earlier),
        ".gitattributes still has the attribute `{earlier}`:\n{attributes}"
    );
    assert!(
        attributes.contains(COMMAND),
        ".gitattributes does not name `{COMMAND}`:\n{attributes}"
    );
    assert!(
        !attributes.contains(&python_driver()),
        ".gitattributes still names the Python driver:\n{attributes}"
    );
}

#[test]
fn text_document_holds_the_registration_line() {
    let command = registration_command();
    let call = format!("{COMMAND} %O %A %B %P");
    assert!(
        command.contains(&call),
        "the registered command does not run `{call}`: {command}"
    );
}

#[test]
fn text_no_tracked_file_names_the_python_driver() {
    let driver = python_driver();
    let interpreter = format!("{}{}", "python", 3);
    let specs = format!(".{}/specs/", "moai");
    let home = TempHome::new();
    let root = Path::new(ROOT);
    let out = git_command(&home, &process_path(), root)
        .args(["ls-files", "-z"])
        .output()
        .unwrap();
    assert!(out.status.success(), "git ls-files: {}", show(&out));
    let tracked = text(&out.stdout);
    let mut found = Vec::new();
    let mut read = 0;
    for rel in tracked.split('\0').filter(|p| !p.is_empty()) {
        if rel.starts_with(&specs) || rel == DECISIONS {
            continue;
        }
        let path = root.join(rel);
        // A file deleted but not yet committed is still listed; it holds nothing.
        let Ok(meta) = fs::symlink_metadata(&path) else {
            continue;
        };
        let held = if meta.file_type().is_symlink() {
            fs::read_link(&path).unwrap().to_string_lossy().into_owned()
        } else if meta.is_file() {
            text(&fs::read(&path).unwrap())
        } else {
            continue;
        };
        read += 1;
        if rel.contains(&driver) {
            found.push(format!("{rel} (the file itself)"));
        }
        for (n, line) in held.lines().enumerate() {
            if line.contains(&driver)
                || (line.contains(&interpreter) && line.contains("moai-manifest"))
            {
                found.push(format!("{rel}:{}", n + 1));
            }
        }
    }
    assert!(read > 100, "the scan read only {read} tracked files");
    assert!(
        found.is_empty(),
        "tracked files outside the history records still name the Python driver or register it \
         with {interpreter}:\n{}",
        found.join("\n")
    );
}
