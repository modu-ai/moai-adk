//! `moai card guard-repoint` (SPEC-20261126-guard-record-repoint): a stale guard record, the file
//! `moai/integration-branch` in the common git folder, is shown as a preview and rewritten to the
//! profile's integration branch only with the preview's digest and a permission reference; the
//! receipt of the rewrite is the only way back. Every test runs a real git and the built binary in
//! a temporary repository and a temporary HOME; nothing here touches a real repository, hook,
//! ledger or user state folder. The hook that guards the branch is installed only through
//! `git-hooks-plan` and `git-hooks-apply`, the way a person gets it.

#![cfg(unix)]

mod common;

use common::home::TempHome;
use common::CleanEnv;
use moai_cli::card::guard;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const BRANCH: &str = "NeXT";
const TARGET_REF: &str = "refs/heads/NeXT";
const STALE_REF: &str = "refs/heads/main";
/// The stale record the fixtures start from: the old integration branch.
const STALE: &[u8] = b"refs/heads/main\n";
/// What the rewrite writes: the target ref and one newline, the bytes the seed writes.
const REWRITTEN: &[u8] = b"refs/heads/NeXT\n";
const REFERENCE: &str = "user-pick-1";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const KEY: &str = "0123456789ab";

fn profile(branch: Option<&str>) -> String {
    let flow = match branch {
        Some(b) => {
            format!(", \"work_flow\": {{\"flow\": \"next-flow\", \"integration_branch\": \"{b}\"}}")
        }
        None => String::new(),
    };
    format!(
        "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]{flow}}}"
    )
}

fn sha_hex(bytes: &[u8]) -> String {
    use sha2::{Digest, Sha256};
    let mut h = Sha256::new();
    h.update(bytes);
    h.finalize().iter().map(|b| format!("{b:02x}")).collect()
}

fn git_executable() -> PathBuf {
    let path = std::env::var_os("PATH").expect("PATH");
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
        .expect("a git on PATH")
}

/// The files and folders under the `moai` folder of the common git folder, with the bytes of each
/// file: a refusal must leave this exactly as it was.
type Snap = BTreeMap<String, Option<Vec<u8>>>;

fn walk(base: &Path, dir: &Path, out: &mut Snap) {
    let Ok(entries) = fs::read_dir(dir) else {
        return;
    };
    for entry in entries {
        let path = entry.unwrap().path();
        let rel = path.strip_prefix(base).unwrap().display().to_string();
        if path.is_dir() {
            out.insert(rel, None);
            walk(base, &path, out);
        } else {
            out.insert(rel, Some(fs::read(&path).unwrap()));
        }
    }
}

/// A temporary repository with a branch `NeXT` and the old branch `main`, the primary worktree
/// detached so both can move, a HOME that holds the `moai` of this build where the hook finds it,
/// and the profile of the project.
struct Fx {
    home: TempHome,
    dir: tempfile::TempDir,
    repo: PathBuf,
    a: String,
    /// A child of `a` made with `commit-tree`; no reference points at it.
    b: String,
}

impl Fx {
    fn new(branch: Option<&str>) -> Self {
        let home = TempHome::new();
        let dir = tempfile::Builder::new()
            .prefix("moai-test-repoint-")
            .tempdir()
            .unwrap();
        let repo = dir.path().join("r");
        fs::create_dir_all(&repo).unwrap();
        let mut fx = Fx {
            home,
            dir,
            repo: repo.clone(),
            a: String::new(),
            b: String::new(),
        };
        fx.git(&repo, &["init", "-q", "-b", "main"]);
        fs::write(repo.join("a.txt"), "a\n").unwrap();
        fx.git(&repo, &["add", "a.txt"]);
        fx.git(
            &repo,
            &[
                "commit",
                "-q",
                "-m",
                &format!("init\n\nMoai-Session: {S_A}:{KEY}"),
            ],
        );
        fx.a = fx.git(&repo, &["rev-parse", "HEAD"]);
        let a = fx.a.clone();
        fx.git(&repo, &["branch", BRANCH, &a]);
        fx.b = fx.commit_tree(&a, &[&a], "the lane head");
        fx.git(&repo, &["checkout", "-q", "--detach"]);
        if let Some(b) = branch {
            fx.write_profile(&repo, Some(b));
        }
        let bin = moai_cli::home_root::path(fx.home.path()).join("bin");
        fs::create_dir_all(&bin).unwrap();
        std::os::unix::fs::symlink(common::MOAI_BIN, bin.join("moai")).unwrap();
        fx
    }

    /// The fixture of most tests: the profile names `NeXT`, the record still names `main`.
    fn stale() -> Self {
        let fx = Fx::new(Some(BRANCH));
        fx.write_record(STALE);
        fx
    }

    fn write_profile(&self, at: &Path, branch: Option<&str>) {
        let cfg = moai_cli::state_root::path(at, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, profile(branch)).unwrap();
    }

    fn path_var(&self) -> String {
        std::env::join_paths([
            git_executable().parent().unwrap().to_path_buf(),
            PathBuf::from("/usr/bin"),
            PathBuf::from("/bin"),
        ])
        .unwrap()
        .to_string_lossy()
        .into_owned()
    }

    fn command(&self, dir: &Path, args: &[&str]) -> Command {
        let mut cmd = Command::new(git_executable());
        cmd.arg("-C")
            .arg(dir)
            .args([
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "-c",
                "gc.auto=0",
                "-c",
                "maintenance.auto=false",
            ])
            .args(args)
            .clean_env()
            .env("HOME", self.home.path())
            .env("PATH", self.path_var())
            .env("LC_ALL", "C")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null");
        cmd
    }

    fn git(&self, dir: &Path, args: &[&str]) -> String {
        let out = self.try_git(dir, args);
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    fn try_git(&self, dir: &Path, args: &[&str]) -> Output {
        self.command(dir, args)
            .stdin(Stdio::null())
            .output()
            .unwrap()
    }

    fn commit_tree(&self, tree_of: &str, parents: &[&str], message: &str) -> String {
        let tree = self.git(&self.repo, &["rev-parse", &format!("{tree_of}^{{tree}}")]);
        let mut args = vec!["commit-tree", tree.as_str()];
        for parent in parents {
            args.push("-p");
            args.push(parent);
        }
        let message = format!("{message}\n\nMoai-Session: {S_A}:{KEY}");
        args.push("-m");
        args.push(&message);
        self.git(&self.repo, &args)
    }

    fn rev(&self, reference: &str) -> Option<String> {
        let out = self.try_git(&self.repo, &["rev-parse", "-q", "--verify", reference]);
        out.status
            .success()
            .then(|| String::from_utf8_lossy(&out.stdout).trim().to_string())
    }

    fn hooks(&self) -> PathBuf {
        self.repo.join(".git/hooks")
    }

    /// The common git folder of the primary worktree.
    fn git_dir(&self) -> PathBuf {
        self.repo.join(".git")
    }

    /// The `moai` folder of the common git folder: the record, the intent file, the lock and the
    /// receipts live here.
    fn moai_dir(&self) -> PathBuf {
        self.git_dir().join("moai")
    }

    fn record(&self) -> PathBuf {
        self.git_dir().join(guard::RECORD_FILE)
    }

    fn lock(&self) -> PathBuf {
        self.git_dir().join(format!("{}.lock", guard::RECORD_FILE))
    }

    fn receipts(&self) -> PathBuf {
        self.moai_dir().join("record-repoint")
    }

    fn write_record(&self, bytes: &[u8]) {
        fs::create_dir_all(self.record().parent().unwrap()).unwrap();
        fs::write(self.record(), bytes).unwrap();
    }

    fn snapshot(&self) -> Snap {
        let base = self.moai_dir();
        let mut out = Snap::new();
        if base.is_dir() {
            out.insert(".".into(), None);
        }
        walk(&base, &base, &mut out);
        out
    }

    /// `moai <args>` in `cwd` with a clean environment, the HOME of this fixture and the git the
    /// fixture names. Answers (exit code, envelope).
    fn moai(&self, cwd: &Path, args: &[&str]) -> (i32, Value) {
        let mut cmd = self.home.command(args);
        cmd.current_dir(cwd)
            .env("PATH", self.path_var())
            .env("MOAI_GIT_BIN", git_executable())
            .env("LC_ALL", "C")
            .stdin(Stdio::null());
        let out = cmd.output().expect("run moai");
        let envelope: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "no envelope (exit {:?}): {} / {}",
                out.status.code(),
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (out.status.code().unwrap_or(-1), envelope)
    }

    fn plan(&self) -> Value {
        let (code, v) = self.moai(&self.repo, &["git-hooks-plan", self.repo.to_str().unwrap()]);
        assert_eq!(code, 0, "{v}");
        v["result"].clone()
    }

    fn apply_hooks(&self, plan: &Value) -> Value {
        let token = plan["confirm_token"].as_str().expect("a token").to_string();
        let (code, v) = self.moai(
            &self.repo,
            &[
                "git-hooks-apply",
                self.repo.to_str().unwrap(),
                "--confirm-git-hooks",
                &token,
            ],
        );
        assert_eq!(code, 0, "{v}");
        v["result"].clone()
    }

    /// `moai card guard-repoint <extra> --project <at>`, run in `at`.
    fn repoint(&self, at: &Path, extra: &[&str]) -> (i32, Value) {
        let mut args = vec!["card", "guard-repoint"];
        args.extend_from_slice(extra);
        args.extend_from_slice(&["--project", at.to_str().unwrap()]);
        self.moai(at, &args)
    }

    /// The preview of `at`: must succeed; answers the result.
    fn preview(&self, at: &Path) -> Value {
        let (code, v) = self.repoint(at, &[]);
        assert_eq!(code, 0, "{v}");
        assert_eq!(v["ok"], true, "{v}");
        v["result"].clone()
    }

    fn apply(&self, at: &Path, digest: &str, reference: &str) -> (i32, Value) {
        self.repoint(
            at,
            &[
                "--apply",
                "--plan-digest",
                digest,
                "--permission-reference",
                reference,
            ],
        )
    }

    fn revert(&self, at: &Path, stamp: &str, reference: &str) -> (i32, Value) {
        self.repoint(
            at,
            &["--revert", stamp, "--permission-reference", reference],
        )
    }

    /// A linked worktree of the repository, with the profile of the project in it.
    fn linked(&self) -> PathBuf {
        let linked = self.dir.path().join("linked");
        self.git(
            &self.repo,
            &[
                "worktree",
                "add",
                "-q",
                "--detach",
                linked.to_str().unwrap(),
            ],
        );
        self.write_profile(&linked, Some(BRANCH));
        linked
    }

    fn next(&self) -> String {
        self.rev(TARGET_REF).expect("NeXT exists")
    }

    /// A refused hand move: non-zero exit, the branch where it was.
    fn refused_move(&self, args: &[&str]) -> String {
        let before = self.next();
        let out = self.try_git(&self.repo, args);
        assert!(
            !out.status.success(),
            "git {args:?} was not refused: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        assert_eq!(self.next(), before, "git {args:?} moved the branch");
        String::from_utf8_lossy(&out.stderr).into_owned()
    }

    /// The preview digest and the stamp-shaped answer of a good apply: the fixture after one
    /// rewrite. Answers the stamp.
    fn rewritten(&self) -> String {
        let digest = digest_of(&self.preview(&self.repo));
        let (code, v) = self.apply(&self.repo, &digest, REFERENCE);
        assert_eq!(code, 0, "{v}");
        v["result"]["stamp"]
            .as_str()
            .unwrap_or_else(|| panic!("no stamp in the answer: {v}"))
            .to_string()
    }
}

fn names(rows: &Value) -> Vec<String> {
    rows.as_array()
        .map(|a| {
            a.iter()
                .filter_map(|r| r.as_str().or_else(|| r["name"].as_str()).map(String::from))
                .collect()
        })
        .unwrap_or_default()
}

fn digest_of(preview: &Value) -> String {
    let digest = preview["plan_digest"]
        .as_str()
        .unwrap_or_else(|| panic!("the preview carries no plan_digest: {preview}"));
    assert!(!digest.is_empty(), "{preview}");
    digest.to_string()
}

/// A digest that is not the preview's: the last character changed.
fn tamper(digest: &str) -> String {
    let mut text = digest.to_string();
    let last = text.pop().expect("a digest");
    text.push(if last == 'a' { 'b' } else { 'a' });
    text
}

/// The answer is a refusal: non-zero exit, `ok: false`, a message that names one of `reasons`
/// (any case; none given means only that a message exists), and nothing under the `moai` folder
/// changed. Answers the message.
fn assert_refused(fx: &Fx, before: &Snap, out: (i32, Value), reasons: &[&str]) -> String {
    let (code, v) = out;
    assert_ne!(code, 0, "a refusal exits non-zero: {v}");
    assert_eq!(v["ok"], false, "{v}");
    assert_ne!(
        v["error"]["kind"], "usage",
        "a refusal of the action itself, not a command-line usage error: {v}"
    );
    let message = v["error"]["message"]
        .as_str()
        .unwrap_or_else(|| panic!("a refusal carries a message: {v}"))
        .to_string();
    assert!(!message.is_empty(), "{v}");
    if !reasons.is_empty() {
        let lower = message.to_lowercase();
        assert!(
            reasons.iter().any(|r| lower.contains(r)),
            "the refusal names none of {reasons:?}: {message}"
        );
    }
    assert_eq!(
        &fx.snapshot(),
        before,
        "a refusal writes nothing: {message}"
    );
    message
}

// ---- AC-1 -----------------------------------------------------------------------------------

/// AC-1: a record that names `main` while the profile names `NeXT` is `stale`: the preview names
/// the record, the target, whether the target is a local branch, the hook, the landing in flight,
/// what is no longer guarded, a digest and the exact apply command, and writes nothing.
#[test]
fn ac1_a_stale_record_answers_stale_with_the_plan_and_writes_nothing() {
    let fx = Fx::stale();
    let before = fx.snapshot();
    let v = fx.preview(&fx.repo);
    assert_eq!(v["state"], "stale", "{v}");
    assert_eq!(v["record"], STALE_REF, "{v}");
    assert_eq!(v["target"], TARGET_REF, "{v}");
    assert_eq!(v["branch_exists"], true, "{v}");
    assert!(v["hook"].is_string(), "{v}");
    assert_eq!(v["intent_in_flight"], false, "{v}");
    assert!(
        v["unguarded_after"].to_string().contains(STALE_REF),
        "the old branch is named as unguarded after: {v}"
    );
    let digest = digest_of(&v);
    let apply = v["apply"]
        .as_str()
        .unwrap_or_else(|| panic!("no apply command: {v}"));
    for needle in [
        "moai card guard-repoint",
        "--apply",
        "--plan-digest",
        digest.as_str(),
        "--permission-reference",
    ] {
        assert!(apply.contains(needle), "{needle} is not in `{apply}`");
    }
    assert_eq!(v["writes"], false, "{v}");
    assert_eq!(fx.snapshot(), before, "the preview writes nothing");
    assert_eq!(fs::read(fx.record()).unwrap(), STALE);
    assert!(!fx.receipts().exists(), "the preview makes no receipt");
    assert!(!fx.lock().exists());
    assert_eq!(
        digest_of(&fx.preview(&fx.repo)),
        digest,
        "the same repository answers the same digest"
    );
}

/// AC-1: the digest binds the record line and bytes and the target ref, and not the tip of the
/// target branch, which moves with every landing.
#[test]
fn ac1_the_digest_binds_the_record_and_the_target_and_not_the_tip() {
    let fx = Fx::stale();
    let first = digest_of(&fx.preview(&fx.repo));
    let b = fx.b.clone();
    fx.git(&fx.repo, &["update-ref", TARGET_REF, &b]);
    assert_eq!(
        digest_of(&fx.preview(&fx.repo)),
        first,
        "a moved tip is the same plan"
    );
    fx.write_record(b"refs/heads/other\n");
    assert_ne!(
        digest_of(&fx.preview(&fx.repo)),
        first,
        "another record line is another plan"
    );
    fx.write_record(b"refs/heads/main\n# kept by hand\n");
    assert_ne!(
        digest_of(&fx.preview(&fx.repo)),
        first,
        "the same line with other bytes is another plan"
    );
    fx.write_record(STALE);
    fx.write_profile(&fx.repo, Some("Other"));
    assert_ne!(
        digest_of(&fx.preview(&fx.repo)),
        first,
        "another target is another plan"
    );
}

/// AC-1: a record equal to the profile's ref is `current`: no digest, nothing written.
#[test]
fn ac1_a_record_equal_to_the_profile_answers_current() {
    let fx = Fx::new(Some(BRANCH));
    fx.write_record(REWRITTEN);
    let before = fx.snapshot();
    let v = fx.preview(&fx.repo);
    assert_eq!(v["state"], "current", "{v}");
    assert!(v["plan_digest"].is_null(), "{v}");
    assert_eq!(v["writes"], false, "{v}");
    assert_eq!(fx.snapshot(), before);
}

/// AC-1: no record is `no_record`: no digest, and the preview does not seed one.
#[test]
fn ac1_no_record_answers_no_record_and_seeds_none() {
    let fx = Fx::new(Some(BRANCH));
    let before = fx.snapshot();
    let v = fx.preview(&fx.repo);
    assert_eq!(v["state"], "no_record", "{v}");
    assert!(v["plan_digest"].is_null(), "{v}");
    assert_eq!(v["writes"], false, "{v}");
    assert!(!fx.record().exists(), "the preview seeds no record");
    assert_eq!(fx.snapshot(), before);
}

/// AC-1: a profile that names no integration branch is `no_profile_branch`.
#[test]
fn ac1_a_profile_without_a_branch_answers_no_profile_branch() {
    let fx = Fx::new(None);
    fx.write_record(STALE);
    let before = fx.snapshot();
    let v = fx.preview(&fx.repo);
    assert_eq!(v["state"], "no_profile_branch", "{v}");
    assert!(v["plan_digest"].is_null(), "{v}");
    assert_eq!(v["writes"], false, "{v}");
    assert_eq!(fx.snapshot(), before);
}

/// AC-1: a linked worktree is `linked_worktree`: the profile is read in the primary worktree
/// only, so the same repository that is `stale` there answers no digest here.
#[test]
fn ac1_a_linked_worktree_answers_linked_worktree() {
    let fx = Fx::stale();
    let primary = fx.preview(&fx.repo);
    assert_eq!(primary["state"], "stale", "{primary}");
    let linked = fx.linked();
    let before = fx.snapshot();
    let v = fx.preview(&linked);
    assert_eq!(v["state"], "linked_worktree", "{v}");
    assert!(v["plan_digest"].is_null(), "{v}");
    assert_eq!(v["writes"], false, "{v}");
    assert_eq!(fx.snapshot(), before);
}

/// AC-1: a folder that is no repository is refused with a reason and nothing is written in it.
#[test]
fn ac1_a_folder_that_is_no_repository_is_refused() {
    let fx = Fx::new(Some(BRANCH));
    let plain = fx.dir.path().join("plain");
    fs::create_dir_all(&plain).unwrap();
    let (code, v) = fx.repoint(&plain, &[]);
    assert_ne!(code, 0, "{v}");
    assert_eq!(v["ok"], false, "{v}");
    assert_ne!(
        v["error"]["kind"], "usage",
        "a refusal of the action itself, not a command-line usage error: {v}"
    );
    let message = v["error"]["message"].as_str().unwrap_or("").to_lowercase();
    assert!(
        message.contains("git") || message.contains("repository"),
        "the refusal says why: {v}"
    );
    assert!(
        fs::read_dir(&plain).unwrap().next().is_none(),
        "nothing was written in the folder"
    );
}

// ---- AC-2 -----------------------------------------------------------------------------------

/// AC-2: with the preview's digest and a valid reference the record names the target, the receipt
/// is in the common git folder with both byte hashes, and the answer carries the receipt, the
/// stamp and the revert command; no lock stays and nothing else is written.
#[test]
fn ac2_apply_rewrites_the_record_and_writes_the_receipt() {
    let fx = Fx::stale();
    let v = fx.preview(&fx.repo);
    let digest = digest_of(&v);
    let profile_before = fs::read(moai_cli::state_root::path(&fx.repo, "project.yaml")).unwrap();
    let (code, answer) = fx.apply(&fx.repo, &digest, REFERENCE);
    assert_eq!(code, 0, "{answer}");
    assert_eq!(answer["ok"], true, "{answer}");
    let result = &answer["result"];
    let stamp = result["stamp"]
        .as_str()
        .unwrap_or_else(|| panic!("no stamp: {answer}"))
        .to_string();
    let shape: Vec<char> = stamp.chars().collect();
    assert_eq!(shape.len(), 16, "stamp is YYYYMMDDTHHMMSSZ: {stamp}");
    assert!(
        shape[8] == 'T'
            && shape[15] == 'Z'
            && shape
                .iter()
                .enumerate()
                .all(|(i, c)| i == 8 || i == 15 || c.is_ascii_digit()),
        "stamp is YYYYMMDDTHHMMSSZ: {stamp}"
    );
    assert!(
        !result["receipt"].is_null(),
        "no receipt in the answer: {answer}"
    );
    let revert = result["revert"]
        .as_str()
        .unwrap_or_else(|| panic!("no revert command: {answer}"));
    for needle in ["moai card guard-repoint", "--revert", stamp.as_str()] {
        assert!(revert.contains(needle), "{needle} is not in `{revert}`");
    }
    assert_eq!(fs::read(fx.record()).unwrap(), REWRITTEN);
    assert!(!fx.lock().exists(), "no lock is left");
    let receipt_path = fx.receipts().join(format!("{stamp}.json"));
    let receipt: Value = serde_json::from_slice(&fs::read(&receipt_path).unwrap()).unwrap();
    assert_eq!(receipt["kind"], "guard-repoint", "{receipt}");
    assert_eq!(receipt["old_record"], STALE_REF, "{receipt}");
    assert_eq!(receipt["new_record"], TARGET_REF, "{receipt}");
    assert_eq!(receipt["old_bytes_sha256"], sha_hex(STALE), "{receipt}");
    assert_eq!(receipt["new_bytes_sha256"], sha_hex(REWRITTEN), "{receipt}");
    assert_eq!(receipt["plan_digest"], digest.as_str(), "{receipt}");
    assert_eq!(receipt["permission_reference"], REFERENCE, "{receipt}");
    let files: Vec<String> = fs::read_dir(fx.receipts())
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(
        files,
        [format!("{stamp}.json")],
        "one receipt and nothing else"
    );
    assert!(
        !fx.hooks().join("reference-transaction").exists(),
        "the rewrite installs no hook"
    );
    assert_eq!(
        fs::read(moai_cli::state_root::path(&fx.repo, "project.yaml")).unwrap(),
        profile_before,
        "the profile is not touched"
    );
    // The repository agrees with the profile now.
    assert_eq!(fx.preview(&fx.repo)["state"], "current");
}

/// AC-2, the dangerous path itself: after the rewrite and the installer (`git-hooks-plan` then
/// `git-hooks-apply`, never a hand-written hook) the hand moves of `NeXT` are refused and the
/// old branch `main` moves freely; the installer leaves the rewritten record alone.
#[test]
fn ac2_after_the_repoint_the_installer_guards_the_new_branch_and_not_the_old_one() {
    let fx = Fx::stale();
    fx.rewritten();
    assert_eq!(fs::read(fx.record()).unwrap(), REWRITTEN);
    let plan = fx.plan();
    assert_eq!(names(&plan["install"]), ["reference-transaction"], "{plan}");
    let applied = fx.apply_hooks(&plan);
    assert_eq!(applied["guard"]["record"], "present", "{applied}");
    assert_eq!(fs::read(fx.record()).unwrap(), REWRITTEN);
    let b = fx.b.clone();
    let stderr = fx.refused_move(&["update-ref", TARGET_REF, &b]);
    assert!(stderr.contains("moai ref guard"), "{stderr}");
    fx.refused_move(&["branch", "-f", BRANCH, &b]);
    fx.git(&fx.repo, &["update-ref", STALE_REF, &b]);
    assert_eq!(
        fx.rev(STALE_REF).as_deref(),
        Some(b.as_str()),
        "the old branch is not guarded after the rewrite"
    );
}

/// AC-2: a digest that is not the preview's, and no digest, write nothing.
#[test]
fn ac2_a_wrong_or_absent_digest_writes_nothing() {
    let fx = Fx::stale();
    let digest = digest_of(&fx.preview(&fx.repo));
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.apply(&fx.repo, &tamper(&digest), REFERENCE),
        &["digest"],
    );
    assert_refused(
        &fx,
        &before,
        fx.repoint(&fx.repo, &["--apply", "--permission-reference", REFERENCE]),
        &["digest"],
    );
}

/// AC-2: no permission reference writes nothing, and names the flag.
#[test]
fn ac2_a_missing_permission_reference_writes_nothing() {
    let fx = Fx::stale();
    let digest = digest_of(&fx.preview(&fx.repo));
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.repoint(&fx.repo, &["--apply", "--plan-digest", &digest]),
        &["permission-reference", "permission reference"],
    );
}

/// AC-2: a reference with a character outside letters, digits and `. _ : / -`, an empty one and
/// one of 81 characters (the first over the limit of 80) write nothing.
#[test]
fn ac2_an_invalid_permission_reference_writes_nothing() {
    let fx = Fx::stale();
    let digest = digest_of(&fx.preview(&fx.repo));
    let before = fx.snapshot();
    let too_long = "a".repeat(81);
    for bad in [
        "",
        "has space",
        "semi;colon",
        "tab\there",
        "new\nline",
        "umlaut-\u{fc}",
        "star*",
        too_long.as_str(),
    ] {
        assert_refused(
            &fx,
            &before,
            fx.apply(&fx.repo, &digest, bad),
            &["permission-reference", "permission reference"],
        );
    }
}

/// AC-2: a reference of exactly 80 characters, built of every allowed mark, is accepted (the last
/// length that passes), and the receipt keeps it as given.
#[test]
fn ac2_a_reference_of_80_characters_with_every_allowed_mark_is_accepted() {
    let fx = Fx::stale();
    let digest = digest_of(&fx.preview(&fx.repo));
    let reference = "Ab9._:/-".repeat(10);
    assert_eq!(reference.chars().count(), 80);
    let (code, v) = fx.apply(&fx.repo, &digest, &reference);
    assert_eq!(code, 0, "{v}");
    let stamp = v["result"]["stamp"].as_str().expect("a stamp").to_string();
    let receipt: Value =
        serde_json::from_slice(&fs::read(fx.receipts().join(format!("{stamp}.json"))).unwrap())
            .unwrap();
    assert_eq!(
        receipt["permission_reference"],
        reference.as_str(),
        "{receipt}"
    );
    assert_eq!(fs::read(fx.record()).unwrap(), REWRITTEN);
}

/// AC-2: a record edited after the preview writes nothing, whatever the edit was.
#[test]
fn ac2_a_record_edited_after_the_preview_writes_nothing() {
    let fx = Fx::stale();
    let digest = digest_of(&fx.preview(&fx.repo));
    fx.write_record(b"refs/heads/other\n");
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.apply(&fx.repo, &digest, REFERENCE),
        &["record", "digest"],
    );
    assert_eq!(fs::read(fx.record()).unwrap(), b"refs/heads/other\n");
    // The same line with other bytes is an edit too.
    fx.write_record(b"refs/heads/main\n# kept by hand\n");
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.apply(&fx.repo, &digest, REFERENCE),
        &["record", "digest"],
    );
}

/// AC-2: a target that is no local `refs/heads/` ref writes nothing: a branch that does not
/// exist, a tag of that name and a remote-tracking ref whose short name is the profile's value.
#[test]
fn ac2_a_target_that_is_no_local_branch_writes_nothing() {
    let fx = Fx::stale();
    let a = fx.a.clone();
    fx.git(&fx.repo, &["tag", "Tagged", &a]);
    fx.git(&fx.repo, &["update-ref", "refs/remotes/origin/Remote", &a]);
    let before = fx.snapshot();
    for name in ["Missing", "Tagged", "origin/Remote"] {
        fx.write_profile(&fx.repo, Some(name));
        let v = fx.preview(&fx.repo);
        assert_eq!(v["state"], "stale", "{name}: {v}");
        assert_eq!(v["branch_exists"], false, "{name}: {v}");
        let digest = digest_of(&v);
        let lower = name.to_lowercase();
        assert_refused(
            &fx,
            &before,
            fx.apply(&fx.repo, &digest, REFERENCE),
            &[lower.as_str()],
        );
    }
}

/// AC-2: a landing in flight, the file `moai/land-intent`, writes nothing and the file stays.
#[test]
fn ac2_a_land_intent_file_blocks_the_apply() {
    let fx = Fx::stale();
    let intent = json!({
        "ref": STALE_REF, "old": fx.a, "new": fx.b, "by": "card land", "at": 0,
    });
    fs::write(fx.git_dir().join(guard::INTENT_FILE), intent.to_string()).unwrap();
    let v = fx.preview(&fx.repo);
    assert_eq!(v["state"], "stale", "{v}");
    assert_eq!(v["intent_in_flight"], true, "{v}");
    let digest = digest_of(&v);
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.apply(&fx.repo, &digest, REFERENCE),
        &["intent"],
    );
    assert_eq!(
        fs::read_to_string(fx.git_dir().join(guard::INTENT_FILE)).unwrap(),
        intent.to_string(),
        "the landing's intent file is left as it was"
    );
}

/// AC-2: a linked worktree cannot apply, even with the digest the primary worktree previewed.
#[test]
fn ac2_a_linked_worktree_cannot_apply() {
    let fx = Fx::stale();
    let digest = digest_of(&fx.preview(&fx.repo));
    let linked = fx.linked();
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.apply(&linked, &digest, REFERENCE),
        &["linked"],
    );
}

/// AC-2: a lock left by a process that died refuses the next apply, names the lock, and is not
/// removed by the refusal.
#[test]
fn ac2_a_leftover_lock_blocks_the_apply_and_the_refusal_names_it() {
    let fx = Fx::stale();
    let digest = digest_of(&fx.preview(&fx.repo));
    fs::write(fx.lock(), "held by another process\n").unwrap();
    let before = fx.snapshot();
    let message = assert_refused(&fx, &before, fx.apply(&fx.repo, &digest, REFERENCE), &[]);
    assert!(
        message.contains("integration-branch.lock"),
        "the refusal names the lock: {message}"
    );
    assert_eq!(
        fs::read_to_string(fx.lock()).unwrap(),
        "held by another process\n"
    );
}

/// AC-2: a step that fails after the lock was taken (the receipt folder cannot be made, because a
/// file stands where it must be) leaves the record as it was, no receipt and no lock.
#[test]
fn ac2_a_failing_receipt_write_leaves_the_record_and_no_lock() {
    let fx = Fx::stale();
    let digest = digest_of(&fx.preview(&fx.repo));
    fs::write(fx.receipts(), "a file where the receipt folder goes\n").unwrap();
    let before = fx.snapshot();
    assert_refused(&fx, &before, fx.apply(&fx.repo, &digest, REFERENCE), &[]);
    assert_eq!(fs::read(fx.record()).unwrap(), STALE);
    assert!(!fx.lock().exists(), "no lock is left");
}

// ---- AC-3 -----------------------------------------------------------------------------------

/// AC-3: the revert puts the old bytes back exactly (a record with a CRLF line and a second line
/// included) while the record still equals the receipt's new bytes, writes `<stamp>.revert.json`
/// beside the receipt, and leaves the first receipt as it was.
#[test]
fn ac3_revert_restores_the_old_bytes_and_writes_the_revert_receipt() {
    const ODD: &[u8] = b"refs/heads/main\r\n# kept by hand\n";
    let fx = Fx::stale();
    fx.write_record(ODD);
    let stamp = fx.rewritten();
    assert_eq!(fs::read(fx.record()).unwrap(), REWRITTEN);
    let receipt_path = fx.receipts().join(format!("{stamp}.json"));
    let receipt_bytes = fs::read(&receipt_path).unwrap();
    let receipt: Value = serde_json::from_slice(&receipt_bytes).unwrap();
    assert_eq!(receipt["old_bytes_sha256"], sha_hex(ODD), "{receipt}");
    let (code, v) = fx.revert(&fx.repo, &stamp, "user-pick-2");
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(
        fs::read(fx.record()).unwrap(),
        ODD,
        "the old bytes, exactly"
    );
    assert!(!fx.lock().exists(), "no lock is left");
    let revert_path = fx.receipts().join(format!("{stamp}.revert.json"));
    let revert: Value = serde_json::from_slice(&fs::read(&revert_path).unwrap()).unwrap();
    assert_eq!(revert["stamp"], stamp.as_str(), "{revert}");
    assert_eq!(revert["permission_reference"], "user-pick-2", "{revert}");
    assert_eq!(revert["restored_record"], STALE_REF, "{revert}");
    assert!(revert["reverted_at"].is_string(), "{revert}");
    assert_eq!(
        fs::read(&receipt_path).unwrap(),
        receipt_bytes,
        "receipts are only added to"
    );
    assert_eq!(
        fx.preview(&fx.repo)["state"],
        "stale",
        "the repository is as it was before the rewrite"
    );
}

/// AC-3: a record that changed since the rewrite is not reverted, and nothing is written.
#[test]
fn ac3_a_changed_record_refuses_the_revert() {
    let fx = Fx::stale();
    let stamp = fx.rewritten();
    fx.write_record(b"refs/heads/other\n");
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.revert(&fx.repo, &stamp, "user-pick-2"),
        &["record", "changed"],
    );
    assert_eq!(fs::read(fx.record()).unwrap(), b"refs/heads/other\n");
    assert!(!fx.receipts().join(format!("{stamp}.revert.json")).exists());
}

/// AC-3: a second revert refuses and writes nothing: first because the record is no longer the
/// receipt's new bytes, then, with the record put back to those bytes by hand, because the revert
/// receipt of that stamp already exists.
#[test]
fn ac3_a_second_revert_refuses() {
    let fx = Fx::stale();
    let stamp = fx.rewritten();
    let (code, v) = fx.revert(&fx.repo, &stamp, "user-pick-2");
    assert_eq!(code, 0, "{v}");
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.revert(&fx.repo, &stamp, "user-pick-3"),
        &[],
    );
    fx.write_record(REWRITTEN);
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.revert(&fx.repo, &stamp, "user-pick-3"),
        &["already", "revert"],
    );
    assert_eq!(fs::read(fx.record()).unwrap(), REWRITTEN);
}

/// AC-3: a linked worktree cannot revert, as it cannot apply; the record and the receipts stay.
#[test]
fn ac3_a_linked_worktree_cannot_revert() {
    let fx = Fx::stale();
    let stamp = fx.rewritten();
    let linked = fx.linked();
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.revert(&linked, &stamp, "user-pick-2"),
        &["linked"],
    );
    assert_eq!(fs::read(fx.record()).unwrap(), REWRITTEN);
}

/// AC-3: a landing in flight, the file `moai/land-intent`, blocks the revert and the file stays.
#[test]
fn ac3_a_land_intent_file_blocks_the_revert() {
    let fx = Fx::stale();
    let stamp = fx.rewritten();
    let intent = json!({
        "ref": TARGET_REF, "old": fx.a, "new": fx.b, "by": "card land", "at": 0,
    });
    fs::write(fx.git_dir().join(guard::INTENT_FILE), intent.to_string()).unwrap();
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.revert(&fx.repo, &stamp, "user-pick-2"),
        &["intent"],
    );
    assert_eq!(fs::read(fx.record()).unwrap(), REWRITTEN);
    assert_eq!(
        fs::read_to_string(fx.git_dir().join(guard::INTENT_FILE)).unwrap(),
        intent.to_string(),
        "the landing's intent file is left as it was"
    );
}

/// AC-3: a receipt that is missing or unreadable refuses the revert and writes nothing; so does a
/// stamp that climbs out of the receipt folder, and `--apply` together with `--revert`.
#[test]
fn ac3_a_missing_or_unreadable_receipt_refuses_the_revert() {
    let fx = Fx::stale();
    let stamp = fx.rewritten();
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.revert(&fx.repo, "20200101T000000Z", "user-pick-2"),
        &["receipt"],
    );
    assert_refused(
        &fx,
        &before,
        fx.revert(&fx.repo, "../integration-branch", "user-pick-2"),
        &[],
    );
    // A command-line conflict (a usage error) is a fair refusal here.
    let (code, v) = fx.repoint(
        &fx.repo,
        &[
            "--apply",
            "--revert",
            &stamp,
            "--plan-digest",
            "x",
            "--permission-reference",
            "user-pick-2",
        ],
    );
    assert_ne!(code, 0, "{v}");
    assert_eq!(v["ok"], false, "{v}");
    assert_eq!(fx.snapshot(), before, "nothing is written: {v}");
    fs::write(fx.receipts().join("20200101T000100Z.json"), "not json").unwrap();
    let before = fx.snapshot();
    assert_refused(
        &fx,
        &before,
        fx.revert(&fx.repo, "20200101T000100Z", "user-pick-2"),
        &["receipt"],
    );
    assert_eq!(fs::read(fx.record()).unwrap(), REWRITTEN);
}

// ---- AC-7 -----------------------------------------------------------------------------------

fn doc(name: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs/design-review/appendix/operations")
            .join(name),
    )
    .unwrap()
}

/// AC-7: the card operations document and the two git-hook operations documents name the new
/// action, the receipt folder and the update's withheld guard hook for a stale record, and the
/// apply document says the installer and the seed never rewrite an existing record.
#[test]
fn ac7_the_operations_documents_name_the_repoint_the_receipt_folder_and_the_withheld_hook() {
    for name in ["card.md", "git_hooks_plan.md", "git_hooks_apply.md"] {
        let text = doc(name);
        for needle in ["guard-repoint", "moai/record-repoint"] {
            assert!(text.contains(needle), "{name} does not name {needle}");
        }
        assert!(
            text.lines()
                .any(|l| l.contains("stale") && l.contains("withheld")),
            "{name} has no sentence that the guard hook is withheld for a stale record"
        );
    }
    assert!(
        doc("git_hooks_apply.md").contains("never rewrite"),
        "git_hooks_apply.md does not say the installer and the seed never rewrite the record"
    );
}
