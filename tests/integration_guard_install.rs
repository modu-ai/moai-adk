//! The installer of the integration branch guard (SPEC-20261125-integration-guard-install): the
//! third managed hook `reference-transaction`, the guard record the apply writes, and the
//! dangerous path itself — a repository set up only through `git-hooks-plan` and
//! `git-hooks-apply` refuses the hand moves of the integration branch and lets `card land`
//! through. Every test runs a real git and the built binary in a temporary repository and a
//! temporary HOME; nothing here touches a real repository, hook or ledger.

#![cfg(unix)]

mod common;

use common::home::TempHome;
use common::CleanEnv;
use moai_cli::card::guard;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const BRANCH: &str = "NeXT";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_LEAD: &str = "dddddddd-0000-4000-8000-00000000000d";
const KEY: &str = "0123456789ab";
const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const GO_PRE_COMMIT: &str = "#!/bin/sh\n\
# MoAI-ADK pre-commit hook — fast subset (gofmt + go vet) + opt-in heavy gate (moai gate)\n\
# Bypass via: SKIP_MOAI_PRECOMMIT=1 git commit\n\
if ! MOAI_PRECOMMIT=1 moai gate; then\n    exit 1\nfi\nexit 0\n";
const USER_HOOK: &str = "#!/bin/sh\n# my own reference-transaction hook, not MoAI's\nexit 0\n";

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

/// A temporary repository with a branch `NeXT`, the primary worktree detached so the branch can
/// move, a HOME that holds the `moai` of this build where the hook finds it, and the profile
/// of the project.
struct Fx {
    home: TempHome,
    dir: tempfile::TempDir,
    repo: PathBuf,
    /// The worktree that has `NeXT` checked out, when a test asked for one.
    wt: Option<PathBuf>,
    a: String,
    /// A child of `a` made with `commit-tree`; no reference points at it.
    b: String,
}

impl Fx {
    fn new(branch: Option<&str>) -> Self {
        let home = TempHome::new();
        let dir = tempfile::Builder::new()
            .prefix("moai-test-install-")
            .tempdir()
            .unwrap();
        let repo = dir.path().join("r");
        fs::create_dir_all(&repo).unwrap();
        let mut fx = Fx {
            home,
            dir,
            repo: repo.clone(),
            wt: None,
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

    fn record(&self) -> PathBuf {
        self.repo.join(".git").join(guard::RECORD_FILE)
    }

    fn put_hook(&self, name: &str, text: &str) {
        fs::create_dir_all(self.hooks()).unwrap();
        let path = self.hooks().join(name);
        fs::write(&path, text).unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }

    fn put_go_pre_commit(&self) {
        self.put_hook("pre-commit", GO_PRE_COMMIT);
        fs::write(
            self.hooks().join(".moai-pre-commit.sha256"),
            format!("{}  pre-commit\n", sha_hex(GO_PRE_COMMIT.as_bytes())),
        )
        .unwrap();
    }

    /// `moai <args>` in `cwd` with a clean environment, the HOME of this fixture and the git
    /// the fixture names. Answers (exit code, envelope).
    fn moai(&self, cwd: &Path, session: Option<&str>, args: &[&str]) -> (i32, Value) {
        let mut cmd = self.home.command(args);
        cmd.current_dir(cwd)
            .env("PATH", self.path_var())
            .env("MOAI_GIT_BIN", git_executable())
            .env("LC_ALL", "C")
            .stdin(Stdio::null());
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        let out = cmd.output().expect("run moai");
        let envelope: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "no envelope: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (out.status.code().unwrap_or(-1), envelope)
    }

    fn plan_at(&self, at: &Path) -> Value {
        let (code, v) = self.moai(at, None, &["git-hooks-plan", at.to_str().unwrap()]);
        assert_eq!(code, 0, "{v}");
        v["result"].clone()
    }

    fn plan(&self) -> Value {
        self.plan_at(&self.repo)
    }

    fn apply_at(&self, at: &Path, plan: &Value) -> Value {
        let token = plan["confirm_token"].as_str().expect("a token").to_string();
        let (code, v) = self.moai(
            at,
            None,
            &[
                "git-hooks-apply",
                at.to_str().unwrap(),
                "--confirm-git-hooks",
                &token,
            ],
        );
        assert_eq!(code, 0, "{v}");
        v["result"].clone()
    }

    /// Previews and applies in the primary worktree; answers the apply's result.
    fn install(&self) -> Value {
        let plan = self.plan();
        self.apply_at(&self.repo, &plan)
    }

    fn mode_of(&self, name: &str) -> u32 {
        fs::metadata(self.hooks().join(name))
            .unwrap()
            .permissions()
            .mode()
            & 0o777
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

// ---- AC-1 -----------------------------------------------------------------------------------

/// AC-1: the shipped file is the text the guard module provides, so a hook the installer writes
/// is the hook the guard tests run.
#[test]
fn ac1_the_shipped_file_equals_the_guard_text() {
    let shipped = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("install/git-hooks/reference-transaction"),
    )
    .unwrap();
    assert_eq!(shipped, guard::hook_text());
    assert_eq!(shipped, moai_cli::git_hooks::REFERENCE_TRANSACTION_TEXT);
    assert_eq!(moai_cli::git_hooks::MANAGED.len(), 3);
    assert!(moai_cli::git_hooks::MANAGED.contains(&"reference-transaction"));
}

/// AC-1: a Go-generation hook retires, and the apply installs the guard hook executable beside
/// the other two, with the shipped bytes.
#[test]
fn ac1_the_apply_installs_the_guard_hook_beside_the_other_two() {
    let fx = Fx::new(Some(BRANCH));
    fx.put_go_pre_commit();
    let plan = fx.plan();
    assert_eq!(
        names(&plan["install"]),
        ["pre-commit", "pre-push", "reference-transaction"],
        "{plan}"
    );
    let applied = fx.apply_at(&fx.repo, &plan);
    assert_eq!(applied["state"], "applied", "{applied}");
    for name in ["pre-commit", "pre-push", "reference-transaction"] {
        assert_eq!(fx.mode_of(name), 0o755, "{name}");
    }
    assert_eq!(
        fs::read_to_string(fx.hooks().join("reference-transaction")).unwrap(),
        guard::hook_text()
    );
}

/// AC-1: a repository with no Go hook still gets the guard hook when the profile names an
/// integration branch; the other two names stay free.
#[test]
fn ac1_a_repository_with_no_go_hook_gets_the_guard_hook_alone() {
    let fx = Fx::new(Some(BRANCH));
    let plan = fx.plan();
    assert_eq!(names(&plan["install"]), ["reference-transaction"], "{plan}");
    fx.apply_at(&fx.repo, &plan);
    assert!(fx.hooks().join("reference-transaction").is_file());
    assert!(!fx.hooks().join("pre-commit").exists());
    assert!(!fx.hooks().join("pre-push").exists());
}

/// AC-1: a user hook of that name is kept byte for byte and blocks only that name.
#[test]
fn ac1_a_user_hook_of_that_name_is_kept_and_blocks_only_that_name() {
    let fx = Fx::new(Some(BRANCH));
    fx.put_go_pre_commit();
    fx.put_hook("reference-transaction", USER_HOOK);
    let plan = fx.plan();
    assert_eq!(
        names(&plan["install"]),
        ["pre-commit", "pre-push"],
        "{plan}"
    );
    assert!(names(&plan["kept"]).contains(&"reference-transaction".to_string()));
    fx.apply_at(&fx.repo, &plan);
    assert_eq!(
        fs::read_to_string(fx.hooks().join("reference-transaction")).unwrap(),
        USER_HOOK
    );
    assert!(fx.hooks().join("pre-push").is_file());
}

/// AC-1: a guard hook edited after the install is kept and the next plan has nothing to write.
#[test]
fn ac1_an_edited_guard_hook_is_kept() {
    let fx = Fx::new(Some(BRANCH));
    fx.install();
    let path = fx.hooks().join("reference-transaction");
    let edited = format!("{}# edited by the user\n", guard::hook_text());
    fs::write(&path, &edited).unwrap();
    let plan = fx.plan();
    assert_eq!(plan["writes"], json!(false), "{plan}");
    assert!(names(&plan["kept"]).contains(&"reference-transaction".to_string()));
    assert_eq!(fs::read_to_string(&path).unwrap(), edited);
}

/// AC-1: the revert puts back what the apply retired and takes away the hooks it installed.
#[test]
fn ac1_the_revert_puts_back_what_it_retired() {
    let fx = Fx::new(Some(BRANCH));
    fx.put_go_pre_commit();
    let applied = fx.install();
    let stamp = applied["stamp"].as_str().unwrap().to_string();
    let (code, v) = fx.moai(
        &fx.repo,
        None,
        &["git-hooks-revert", &stamp, fx.repo.to_str().unwrap()],
    );
    assert_eq!(code, 0, "{v}");
    assert_eq!(
        fs::read_to_string(fx.hooks().join("pre-commit")).unwrap(),
        GO_PRE_COMMIT
    );
    assert!(!fx.hooks().join("reference-transaction").exists());
}

/// AC-1: a repository whose profile names no integration branch and that holds no record gets
/// no guard hook, and its plan is the plan the installer made before this change.
#[test]
fn ac1_a_repository_with_no_branch_and_no_record_gets_no_hook_and_an_unchanged_plan() {
    let fx = Fx::new(None);
    fx.put_go_pre_commit();
    let plan = fx.plan();
    assert_eq!(
        names(&plan["install"]),
        ["pre-commit", "pre-push"],
        "{plan}"
    );
    assert_eq!(
        plan,
        moai_cli::git_hooks::plan(&fx.repo).unwrap(),
        "the plan with the guard answer `no` is the old plan"
    );
    let applied = fx.apply_at(&fx.repo, &plan);
    assert!(
        !fx.hooks().join("reference-transaction").exists(),
        "{applied}"
    );
    assert!(!fx.record().exists());
}

/// AC-1: a repository that already holds the record gets the hook even when its profile names
/// no branch.
#[test]
fn ac1_a_repository_that_holds_the_record_gets_the_hook() {
    let fx = Fx::new(None);
    fs::create_dir_all(fx.record().parent().unwrap()).unwrap();
    fs::write(fx.record(), format!("refs/heads/{BRANCH}\n")).unwrap();
    let plan = fx.plan();
    assert_eq!(names(&plan["install"]), ["reference-transaction"], "{plan}");
}

// ---- AC-2 -----------------------------------------------------------------------------------

/// AC-2: the apply in the primary worktree writes the record from the profile and says
/// `written`; with the record in place, a later apply leaves it alone and says `present`.
#[test]
fn ac2_the_record_is_written_once_and_never_rewritten() {
    let fx = Fx::new(Some(BRANCH));
    let applied = fx.install();
    assert_eq!(applied["guard"]["record"], "written", "{applied}");
    assert_eq!(
        fs::read_to_string(fx.record()).unwrap(),
        format!("refs/heads/{BRANCH}\n")
    );
    assert!(
        fx.hooks().join("reference-transaction").is_file(),
        "the hook is there when the record is"
    );
    // The hook gone, the record kept: the next apply installs the hook again and does not touch
    // the record.
    fs::remove_file(fx.hooks().join("reference-transaction")).unwrap();
    fs::write(fx.record(), "refs/heads/other\n").unwrap();
    // The retire folder of an apply is named by the second of the clock.
    std::thread::sleep(std::time::Duration::from_millis(1100));
    let again = fx.install();
    assert_eq!(again["guard"]["record"], "present", "{again}");
    assert_eq!(
        fs::read_to_string(fx.record()).unwrap(),
        "refs/heads/other\n"
    );
}

/// AC-2: a linked worktree writes no record and says why. The answer is the record step of the
/// apply, called for the linked worktree (the installer's hooks folder of a linked worktree is a
/// separate question, see `docs/later.md`).
#[test]
fn ac2_a_linked_worktree_writes_no_record() {
    let fx = Fx::new(Some(BRANCH));
    let linked = fx.dir.path().join("linked");
    fx.git(
        &fx.repo,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            linked.to_str().unwrap(),
        ],
    );
    fx.write_profile(&linked, Some(BRANCH));
    let answer = guard::seed_record(&linked);
    assert_eq!(answer["record"], "not_written", "{answer}");
    assert!(
        answer["why"].as_str().unwrap().contains("linked"),
        "{answer}"
    );
    assert!(!fx.record().exists());
}

/// AC-2: a profile that names no branch writes no record and says why.
#[test]
fn ac2_a_profile_with_no_branch_writes_no_record() {
    let fx = Fx::new(None);
    fx.put_go_pre_commit();
    let applied = fx.install();
    assert_eq!(applied["guard"]["record"], "not_written", "{applied}");
    assert!(applied["guard"]["why"]
        .as_str()
        .unwrap()
        .contains("no integration branch"));
    assert!(!fx.record().exists());
}

// ---- AC-3 -----------------------------------------------------------------------------------

impl Fx {
    /// The branch `NeXT` checked out in a linked worktree, as a lane leaves it for the leader.
    fn check_out_next(&mut self) -> PathBuf {
        let wt = self.dir.path().join("next-wt");
        self.git(
            &self.repo,
            &["worktree", "add", "-q", wt.to_str().unwrap(), BRANCH],
        );
        self.wt = Some(wt.clone());
        wt
    }

    fn next(&self) -> String {
        self.rev(&format!("refs/heads/{BRANCH}"))
            .expect("NeXT exists")
    }

    /// A refused hand move: non-zero exit, the branch where it was.
    fn refused(&self, dir: &Path, args: &[&str]) -> String {
        let before = self.next();
        let out = self.try_git(dir, args);
        assert!(
            !out.status.success(),
            "git {args:?} was not refused: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        assert_eq!(self.next(), before, "git {args:?} moved the branch");
        String::from_utf8_lossy(&out.stderr).into_owned()
    }
}

/// AC-3, the dangerous path itself: a repository set up only through the plan and the apply
/// refuses `update-ref`, `branch -f`, `branch -D`, `merge --ff-only` where the branch is checked
/// out and `branch -M` onto the branch, and leaves the branch where it was.
#[test]
fn ac3_a_repository_set_up_by_the_installer_refuses_the_hand_moves() {
    let mut fx = Fx::new(Some(BRANCH));
    let wt = fx.check_out_next();
    let applied = fx.install();
    assert_eq!(applied["guard"]["record"], "written", "{applied}");
    let (a, b) = (fx.a.clone(), fx.b.clone());
    let stderr = fx.refused(
        &fx.repo,
        &["update-ref", &format!("refs/heads/{BRANCH}"), &b],
    );
    assert!(stderr.contains("moai ref guard"), "{stderr}");
    fx.refused(&fx.repo, &["branch", "-f", BRANCH, &b]);
    fx.refused(&fx.repo, &["branch", "-D", BRANCH]);
    // A fast-forward in the worktree that has the branch checked out: the lane head is a
    // descendant on a side branch.
    fx.git(&fx.repo, &["branch", "lane", &b]);
    fx.refused(&wt, &["merge", "--ff-only", "lane"]);
    // A rename onto the integration branch.
    fx.refused(&fx.repo, &["branch", "-M", "lane", BRANCH]);
    assert_eq!(fx.next(), a);
    assert!(
        fx.rev("refs/heads/lane").is_some(),
        "the rename left the source branch alone"
    );
}

impl Fx {
    fn insert_pass(&self, card: &str, head: &str, base: &str) {
        let conn =
            Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap();
        let payload = json!({
            "head": head, "base": base, "verdict": "pass", "tier": 1,
            "evidence": "the fixture's evidence",
            "receipt": {"head": head, "steps": [{"name": "test", "exit": 0}]},
            "reviewer": S_AUD,
        });
        let n: i64 = card.trim_start_matches("C-").parse().unwrap();
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES ('test-verdict-1', unixepoch(), ?1, ?2, 'card.verdict', ?3)",
            rusqlite::params![S_AUD, n, payload.to_string()],
        )
        .unwrap();
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Value {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", self.repo.to_str().unwrap()]);
        let (code, v) = self.moai(&self.repo, session, &all);
        assert_eq!(code, 0, "{all:?}: {v}");
        v["result"].clone()
    }
}

/// AC-3: the landing command moves the branch under the installed guard, on an audited head.
#[test]
fn ac3_card_land_of_an_audited_head_moves_the_branch_under_the_installed_guard() {
    let fx = Fx::new(Some(BRANCH));
    ledger::register(fx.home.path(), "demo").unwrap();
    fx.install();
    let tip = fx.next();
    // The candidate: a merge of the lane head on the tip, signed by the lander.
    let candidate = fx.commit_tree(&tip, &[&tip, &fx.b], "merge the lane head");
    let added = fx.card(
        None,
        &[
            "add",
            "--title",
            "fixture card",
            "--kind",
            "task",
            "--scope",
            "apps/a/",
            "--ac",
            "one",
        ],
    );
    let card = added["name"].as_str().unwrap().to_string();
    fx.card(None, &["ready", &card]);
    fx.card(Some(S_A), &["take", &card]);
    fx.card(Some(S_A), &["review", &card, "--generation", "1"]);
    fx.insert_pass(&card, &candidate, &tip);
    let landed = fx.card(
        Some(S_LEAD),
        &[
            "land",
            &card,
            "--head",
            &candidate,
            "--met",
            "1=the test",
            "--by",
            "leader",
        ],
    );
    assert_eq!(landed["moved"], json!(true), "{landed}");
    assert_eq!(fx.next(), candidate);
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

/// AC-7: the three operations documents say what a person now sees and types.
#[test]
fn ac7_the_operations_documents_name_the_third_hook_the_record_the_line_and_the_codes() {
    let plan = doc("git_hooks_plan.md");
    let apply = doc("git_hooks_apply.md");
    let card = doc("card.md");
    for text in [&plan, &apply] {
        assert!(text.contains("reference-transaction"));
    }
    assert!(plan.contains("is not guarded"), "the session-start line");
    for word in ["written", "present", "not_written"] {
        assert!(apply.contains(word), "{word}");
    }
    assert!(apply.contains("moai/integration-branch"));
    for code in ["no_verdict", "verdict_other_head", "verdict_fail"] {
        assert!(card.contains(code), "{code}");
    }
}
