//! SPEC-20261125-land-sync-merge: `card land-sync <FROM> --spec <SPEC-ID> --by leader` merges a
//! listed local branch into the integration branch with git plumbing, runs the wide check itself
//! and moves the branch by one compare-and-swap through the guard's intent record; only the
//! session of the registered leader entry may call it. A throwaway repository, a throwaway HOME,
//! a fake `cargo` and shell scripts as the check program: the real checks, HOME and slots are
//! never used. The repository: `integ` is the integration branch, `hotfix` the one listed source,
//! `main` stays checked out in the primary checkout.
#![cfg(unix)]

use crate::common;
use common::home::TempHome;
use moai_cli::card::guard;
use moai_cli::ledger;
use moai_cli::spec_doc::{self, Options};
use rusqlite::Connection;
use serde_json::{json, Value};
use std::cell::RefCell;
use std::ffi::OsString;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_LEAD: &str = "dddddddd-0000-4000-8000-00000000000d";
const WORKTREE_KEY: &str = "0123456789ab";
/// The approved SPEC every sync names, and a draft SPEC the tip holds as well.
const SPEC: &str = "SPEC-20261117-sync-fixture";
const DRAFT_SPEC: &str = "SPEC-20261117-sync-draft";
const BRANCH: &str = "integ";
const SOURCE: &str = "hotfix";
/// How every refusal of the command starts and ends.
const PREFIX: &str = "card land-sync refused: ";
const SUFFIX: &str = "; nothing was changed";
/// One scan every commit of the fixture needs, so the wide run of the command plans every step.
const SCAN_TABLE: &str = "{\"scans\": [{\"id\": \"notes-rule\", \"always\": true, \"tests\": [\"tests/notes.rs::notes_rule\"], \"checks\": [\"cargo test --offline --locked --test notes notes_rule\"], \"rules\": [\"tests/notes.rs\"], \"scope\": [{\"prefix\": \"notes/\", \"suffixes\": [\".md\"]}], \"rejects\": \"a note the rule refuses\"}]}";
/// A check program that logs the range it was given and fails.
const RED: &str = "#!/bin/sh\necho \"$1\" >> \"$SYNC_LOG\"\nexit 3\n";
/// A check program that logs the range it was given and passes.
const GREEN: &str = "#!/bin/sh\necho \"$1\" >> \"$SYNC_LOG\"\nexit 0\n";

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

fn signed(subject: &str, session: &str) -> String {
    format!(
        "{subject}\n\nMoai-Spec: {SPEC}\nMoai-Session: {session}:{WORKTREE_KEY}\nMoai-Verified: none"
    )
}

/// A commit on `parent` that adds or changes `path`; no ref moves.
fn put_on(project: &Path, parent: &str, path: &str, content: &str, message: &str) -> String {
    let blob_file = project.join("put-on-blob.tmp");
    fs::write(&blob_file, content).unwrap();
    let blob = git(project, &["hash-object", "-w", "put-on-blob.tmp"]);
    fs::remove_file(&blob_file).unwrap();
    let index = project.join(".git/put-on-index");
    let with_index = |args: &[&str]| -> String {
        let out = common::git::command()
            .arg("-C")
            .arg(project)
            .args(args)
            .env("GIT_INDEX_FILE", &index)
            .output()
            .unwrap();
        assert!(out.status.success(), "git {args:?}");
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    };
    with_index(&["read-tree", parent]);
    with_index(&[
        "update-index",
        "--add",
        "--cacheinfo",
        &format!("100644,{blob},{path}"),
    ]);
    let tree = with_index(&["write-tree"]);
    fs::remove_file(&index).unwrap();
    git(
        project,
        &["commit-tree", &tree, "-p", parent, "-m", message],
    )
}

fn git_executable() -> PathBuf {
    let path = std::env::var_os("PATH").expect("PATH");
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
        .expect("a git on PATH")
}

fn said(out: &Output) -> String {
    format!(
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("stdout is not JSON ({e}): {}", said(out)))
}

fn result(out: &Output) -> Value {
    assert_eq!(out.status.code(), Some(0), "{}", said(out));
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// A refusal: exit 1, `ok` false and a message that starts as every refusal of the command
/// starts. Answers the message.
fn refusal_text(out: &Output) -> String {
    assert_eq!(out.status.code(), Some(1), "{}", said(out));
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    let message = v["error"]["message"]
        .as_str()
        .unwrap_or_else(|| panic!("no error message: {v}"))
        .to_string();
    assert!(message.starts_with(PREFIX), "{message}");
    message
}

/// [`refusal_text`] of a refusal that changed nothing, so the text also ends as such.
fn refusal(out: &Output) -> String {
    let message = refusal_text(out);
    assert!(message.ends_with(SUFFIX), "{message}");
    message
}

/// The text of the SPEC `id` in `status`: an approved one carries the approval record with the
/// digest of its own body, so it reads with no error.
fn spec_text(id: &str, status: &str) -> String {
    let body =
        "## Intent\n\nLogins fail.\n\n## Requirements\n\n- REQ-1: When a user sends a valid \
                password, the server shall accept the login. · failure: none\n\n## Acceptance\n\n\
                - AC-1 (REQ-1): it logs in · verify: `cargo test --test login`\n\n## Scope\n\n\
                - write: `src/login.rs`\n\n## Plan\n\n1. Fix it.\n";
    let front = |status: &str, extra: &str| {
        format!(
            "---\nid: {id}\ntitle: \"Sync fixture\"\nstatus: {status}\nrevision: 1\n\
             created: 2026-11-17\n{extra}---\n\n# Sync fixture\n\n"
        )
    };
    let approval = if status == "draft" {
        String::new()
    } else {
        let draft = spec_doc::parse(
            &format!("{}{body}", front("draft", "")),
            &Options::default(),
        );
        let digest = draft.digest.expect("a GEARS draft has a digest");
        format!(
            "approval:\n  by: user\n  at: 2026-11-17\n  basis: test selection\n  digest: {digest}\n"
        )
    };
    format!("{}{body}", front(status, &approval))
}

/// The ledger and ref state a refused call must leave as it found it.
#[derive(Debug, PartialEq)]
struct Snap {
    refs: String,
    events: (i64, i64),
}

struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    real_git: PathBuf,
    /// The git program the `moai` children run (a wrapper in the tests that need one).
    git_bin: RefCell<PathBuf>,
    scripts: tempfile::TempDir,
    c0: String,
    /// `work_flow.sync_sources` of the profile written before each call (none: no key).
    sources: RefCell<Option<Vec<String>>>,
    extra_env: RefCell<Vec<(&'static str, OsString)>>,
    /// The file the fake `cargo` logs its calls to.
    calls: PathBuf,
}

impl World {
    /// A world whose leader session `S_LEAD` is registered, as `/moai leader` registers it.
    fn new() -> Self {
        let w = World::bare();
        w.register_leader(S_LEAD, true);
        w
    }

    /// A world with no leader registered.
    fn bare() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        git(&project, &["config", "user.name", "fixture"]);
        git(
            &project,
            &["config", "user.email", "fixture@example.invalid"],
        );
        fs::write(project.join("a.txt"), "a\n").unwrap();
        // A root `Cargo.toml` makes the fixture a Rust repository, so a wide `check affected` run
        // plans the (fake) cargo steps.
        fs::write(project.join("Cargo.toml"), "[workspace]\n").unwrap();
        fs::create_dir_all(project.join("tools")).unwrap();
        fs::write(project.join("tools/repo_scans.json"), SCAN_TABLE).unwrap();
        for (id, status) in [(SPEC, "approved"), (DRAFT_SPEC, "draft")] {
            let dir = project.join(spec_doc::SPECS_DIR).join(id);
            fs::create_dir_all(&dir).unwrap();
            fs::write(dir.join("spec.md"), spec_text(id, status)).unwrap();
        }
        let approved = spec_doc::parse(
            &spec_text(SPEC, "approved"),
            &Options {
                folder: Some(SPEC.to_string()),
                closure: false,
            },
        );
        assert_eq!(approved.errors(), 0, "{:?}", approved.findings);
        git(
            &project,
            &["add", "a.txt", "Cargo.toml", "tools", spec_doc::SPECS_DIR],
        );
        git(&project, &["commit", "-q", "-m", "init"]);
        let c0 = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["update-ref", "refs/heads/integ", &c0]);
        let scripts = tempfile::tempdir().unwrap();
        let w = World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            real_git: git_executable(),
            git_bin: RefCell::new(git_executable()),
            calls: scripts.path().join("cargo-calls.log"),
            scripts,
            c0,
            sources: RefCell::new(None),
            extra_env: RefCell::new(Vec::new()),
        };
        w.install_fake_cargo();
        w
    }

    /// Puts a fake `cargo` first on the PATH of every `moai` child, so that the command's own
    /// wide `check affected` runs end to end on the merge: its git, table and summary steps are
    /// real and only the cargo steps are replaced by a script that logs its call and exits 0.
    /// The machine numbers are the debug-only roomy seam, so the real disk and memory never
    /// decide the outcome.
    fn install_fake_cargo(&self) {
        let bin = self.scripts.path().join("fakes");
        fs::create_dir_all(&bin).unwrap();
        self.set_cargo_hook("");
        let path = std::env::join_paths([
            bin,
            self.real_git.parent().unwrap().to_path_buf(),
            PathBuf::from("/usr/bin"),
            PathBuf::from("/bin"),
        ])
        .unwrap();
        let mut env = self.extra_env.borrow_mut();
        env.push(("PATH", path));
        env.push(("FAKE_CARGO_CALLS", self.calls.clone().into_os_string()));
        env.push(("MOAI_TEST_MACHINE", "free_gb=1000,swap_pct=0".into()));
    }

    /// Rewrites the fake `cargo`: it logs its arguments and folder, runs the shell text `hook`
    /// and exits 0.
    fn set_cargo_hook(&self, hook: &str) {
        let script = self.scripts.path().join("fakes").join("cargo");
        fs::write(
            &script,
            format!(
                "#!/bin/sh\necho \"cargo $* @ $(pwd)\" >> \"$FAKE_CARGO_CALLS\"\n{hook}\nexit 0\n"
            ),
        )
        .unwrap();
        fs::set_permissions(&script, fs::Permissions::from_mode(0o755)).unwrap();
    }

    /// The calls the fake `cargo` logged, one line each.
    fn cargo_calls(&self) -> Vec<String> {
        fs::read_to_string(&self.calls)
            .unwrap_or_default()
            .lines()
            .map(String::from)
            .collect()
    }

    fn cargo_runs(&self, word: &str) -> usize {
        let prefix = format!("cargo {word}");
        self.cargo_calls()
            .iter()
            .filter(|l| l.starts_with(&prefix))
            .count()
    }

    fn set_sources(&self, list: &[&str]) {
        *self.sources.borrow_mut() = Some(list.iter().map(|s| s.to_string()).collect());
    }

    fn clear_sources(&self) {
        *self.sources.borrow_mut() = None;
    }

    /// The profile of an m-flow project whose integration branch is `integ`.
    fn profile(&self) -> Vec<u8> {
        let mut flow = json!({"flow": "m-flow", "integration_branch": BRANCH});
        if let Some(list) = &*self.sources.borrow() {
            flow["sync_sources"] = json!(list);
        }
        serde_json::to_vec(&json!({
            "schema_version": 1, "project_id": "demo", "purpose": "fixture", "tools": ["rust"],
            "development_methods": ["sdd"], "hosts": ["claude"], "work_flow": flow
        }))
        .unwrap()
    }

    fn write_profile(&self) {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, self.profile()).unwrap();
    }

    /// `session-register --name leader` as the host session `session`, in mode `leader` or with
    /// no mode (a plain lane entry).
    fn register_leader(&self, session: &str, mode_leader: bool) -> Output {
        let mut args = vec!["session-register", self.p.as_str(), "--name", "leader"];
        if mode_leader {
            args.extend(["--mode", "leader"]);
        }
        let out = self.run(Some(session), &args);
        assert_eq!(out.status.code(), Some(0), "{}", said(&out));
        out
    }

    /// The registry entry of the leader.
    fn leader_entry(&self) -> PathBuf {
        moai_cli::clone_root::path(&self.project.join(".git"))
            .join("sessions")
            .join("leader.json")
    }

    fn script(&self, name: &str, text: &str) -> PathBuf {
        let path = self.scripts.path().join(name);
        fs::write(&path, text).unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        path
    }

    fn log(&self) -> PathBuf {
        self.scripts.path().join("sync.log")
    }

    /// The ranges the check program was given, one per run, in order.
    fn ranges(&self) -> Vec<String> {
        fs::read_to_string(self.log())
            .unwrap_or_default()
            .lines()
            .map(String::from)
            .collect()
    }

    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        self.run_with(session, args, None)
    }

    /// `run`, with the check program that the test seam `MOAI_TEST_LAND_SYNC_CHECK` names.
    fn run_with(&self, session: Option<&str>, args: &[&str], check: Option<&Path>) -> Output {
        let mut cmd: Command = self.home.command(args);
        cmd.current_dir(&self.project);
        cmd.env("MOAI_GIT_BIN", &*self.git_bin.borrow());
        cmd.env("SYNC_LOG", self.log());
        for (key, value) in self.extra_env.borrow().iter() {
            cmd.env(key, value);
        }
        if let Some(program) = check {
            cmd.env("MOAI_TEST_LAND_SYNC_CHECK", program);
        }
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    /// `moai card <args> --project <project>` with the profile written last.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        self.card_with(session, args, None)
    }

    fn card_with(&self, session: Option<&str>, args: &[&str], check: Option<&Path>) -> Output {
        self.write_profile();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        self.run_with(session, &all, check)
    }

    /// `card land-sync <tail>` as `session`.
    fn sync(&self, session: Option<&str>, tail: &[&str]) -> Output {
        let mut all = vec!["land-sync"];
        all.extend_from_slice(tail);
        self.card(session, &all)
    }

    /// `card land-sync <from> --spec <SPEC> --by leader` as the registered leader session.
    fn lead(&self, from: &str) -> Output {
        self.sync(Some(S_LEAD), &[from, "--spec", SPEC, "--by", "leader"])
    }

    /// [`World::lead`] with the check program at `check`.
    fn lead_check(&self, from: &str, check: &Path) -> Output {
        let all = ["land-sync", from, "--spec", SPEC, "--by", "leader"];
        self.card_with(Some(S_LEAD), &all, Some(check))
    }

    fn integ(&self) -> String {
        self.rev(&format!("refs/heads/{BRANCH}"))
    }

    fn rev(&self, reference: &str) -> String {
        git(&self.project, &["rev-parse", reference])
    }

    /// A commit on `parent` that adds `path`; no ref moves.
    fn commit_on(&self, parent: &str, path: &str, content: &str) -> String {
        put_on(
            &self.project,
            parent,
            path,
            content,
            &signed("a change", S_A),
        )
    }

    /// A commit with the tree of `parent` and `parent` as its only parent; no ref moves.
    fn empty_child(&self, parent: &str) -> String {
        let tree = git(&self.project, &["rev-parse", &format!("{parent}^{{tree}}")]);
        git(
            &self.project,
            &["commit-tree", &tree, "-p", parent, "-m", "an empty change"],
        )
    }

    fn point(&self, branch: &str, commit: &str) {
        git(
            &self.project,
            &["update-ref", &format!("refs/heads/{branch}"), commit],
        );
    }

    /// The integration branch and the source each get a commit of their own on the base, so the
    /// merge is a real one: answers (the tip, the source tip).
    fn diverge(&self) -> (String, String) {
        let tip = self.commit_on(&self.c0, "notes/tip.md", "t\n");
        let src = self.commit_on(&self.c0, "notes/h.md", "h\n");
        self.point(BRANCH, &tip);
        self.point(SOURCE, &src);
        (tip, src)
    }

    fn parents(&self, commit: &str) -> Vec<String> {
        git(&self.project, &["rev-list", "--parents", "-n", "1", commit])
            .split_whitespace()
            .skip(1)
            .map(String::from)
            .collect()
    }

    /// The number of reflog entries of the integration branch.
    fn reflog(&self) -> usize {
        git(
            &self.project,
            &[
                "reflog",
                "show",
                "--format=%H",
                &format!("refs/heads/{BRANCH}"),
            ],
        )
        .lines()
        .count()
    }

    fn snap(&self) -> Snap {
        let refs = git(
            &self.project,
            &["for-each-ref", "--format=%(refname) %(objectname)"],
        );
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let events = l
            .conn()
            .query_row(
                "SELECT count(*), COALESCE(max(event_id), 0) FROM event",
                [],
                |r| Ok((r.get(0)?, r.get(1)?)),
            )
            .unwrap();
        Snap { refs, events }
    }

    /// The file the landing writes before the move and removes after it.
    fn intent(&self) -> PathBuf {
        self.project.join(".git").join(guard::INTENT_FILE)
    }

    /// The events of `kind`, oldest first: (event id, card id, payload).
    fn rows(&self, kind: &str) -> Vec<(i64, Option<i64>, Value)> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let conn = l.conn();
        let mut st = conn
            .prepare(
                "SELECT event_id, card_id, payload FROM event WHERE kind = ?1 ORDER BY event_id",
            )
            .unwrap();
        let found: Vec<(i64, Option<i64>, Value)> = st
            .query_map([kind], |r| {
                let payload: String = r.get(2)?;
                Ok((
                    r.get(0)?,
                    r.get(1)?,
                    serde_json::from_str(&payload).unwrap(),
                ))
            })
            .unwrap()
            .map(Result::unwrap)
            .collect();
        found
    }

    fn sync_rows(&self) -> usize {
        self.rows("card.sync_started").len() + self.rows("card.synced").len()
    }

    /// A refusal that changed nothing: the text starts and ends as the command's refusals do,
    /// names one of `any_of`, and the refs, the ledger and the intent file are as `before` had
    /// them. Answers the message.
    fn refused(&self, out: &Output, any_of: &[&str], before: &Snap) -> String {
        let message = refusal(out);
        assert!(
            any_of.iter().any(|needle| message.contains(needle)),
            "the reason names none of {any_of:?}: {message}"
        );
        assert_eq!(
            &self.snap(),
            before,
            "the refused call changed refs or the ledger"
        );
        assert!(
            !self.intent().exists(),
            "the refused call left an intent file"
        );
        assert_eq!(self.sync_rows(), 0, "the refused call wrote a sync row");
        message
    }

    fn worktrees(&self) -> String {
        git(&self.project, &["worktree", "list", "--porcelain"])
    }

    // ---- cards, for the tests that need an earlier landing -----------------------------------

    /// A card with the write scope `apps/<n>/`, linked to the SPEC, taken by the lane and sent to
    /// review; returns its name.
    fn review_card(&self, n: usize) -> String {
        let added = result(&self.card(
            None,
            &[
                "add",
                "--title",
                "fixture card",
                "--kind",
                "task",
                "--scope",
                &format!("apps/{n}/"),
                "--ac",
                "one",
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        self.insert_event(
            &name,
            S_LEAD,
            "card.spec",
            json!({"from": null, "to": SPEC}),
        );
        result(&self.card(None, &["ready", &name]));
        result(&self.card(Some(S_A), &["take", &name]));
        result(&self.card(Some(S_A), &["review", &name, "--generation", "1"]));
        name
    }

    /// An event of `card`, written as the command that owns the kind writes it.
    fn insert_event(&self, card: &str, actor: &str, kind: &str, payload: Value) {
        let conn =
            Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap();
        let rows: i64 = conn
            .query_row("SELECT count(*) FROM event", [], |r| r.get(0))
            .unwrap();
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, ?4, ?5)",
            rusqlite::params![
                format!("test-{kind}-{rows}"),
                actor,
                card.trim_start_matches("C-").parse::<i64>().unwrap(),
                kind,
                payload.to_string()
            ],
        )
        .unwrap();
    }

    /// An event with no card, written as the command that owns the kind writes it.
    fn insert_cardless_event(&self, actor: &str, kind: &str, payload: Value) {
        let conn =
            Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap();
        let rows: i64 = conn
            .query_row("SELECT count(*) FROM event", [], |r| r.get(0))
            .unwrap();
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, NULL, ?3, ?4)",
            rusqlite::params![
                format!("test-{kind}-{rows}"),
                actor,
                kind,
                payload.to_string()
            ],
        )
        .unwrap();
    }

    /// The folders under `<project>/target/land-batch/` that the private worktrees of the wide
    /// check are made in (the logs beside them stay, so the files are left out).
    fn private_worktree_folders(&self) -> Vec<String> {
        let base = self.project.join("target").join("land-batch");
        let Ok(entries) = fs::read_dir(&base) else {
            return Vec::new();
        };
        entries
            .map(|e| e.unwrap())
            .filter(|e| e.path().is_dir())
            .map(|e| e.file_name().to_string_lossy().into_owned())
            .collect()
    }

    /// A new card with a lane head on the integration tip and an independent PASS of it, landed
    /// by `card land`; answers the card, the head and the landing's output.
    fn land_next(&self, n: usize) -> (String, String, Output) {
        let card = self.review_card(n);
        let parent = self.integ();
        let head = self.commit_on(&parent, &format!("notes/{n}.md"), "ok\n");
        self.insert_event(
            &card,
            S_AUD,
            "card.verdict",
            json!({
                "head": head, "base": parent, "verdict": "pass", "tier": 0,
                "evidence": "the fixture's evidence", "receipt": null, "reviewer": S_AUD,
            }),
        );
        let out = self.card(
            Some(S_LEAD),
            &["land", &card, "--head", &head, "--by", "leader"],
        );
        (card, head, out)
    }

    /// [`World::land_next`] that must land.
    fn land_one(&self, n: usize) -> (String, String) {
        let (card, head, out) = self.land_next(n);
        let got = result(&out);
        assert_eq!(got["moved"], json!(true), "{got}");
        (card, head)
    }

    fn state(&self, card: &str) -> String {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn()
            .query_row(
                "SELECT state FROM card WHERE card_id = ?1",
                [card.trim_start_matches("C-").parse::<i64>().unwrap()],
                |r| r.get(0),
            )
            .unwrap()
    }

    // ---- the installed hook ------------------------------------------------------------------

    fn path_var(&self) -> OsString {
        std::env::join_paths([
            self.real_git.parent().unwrap().to_path_buf(),
            PathBuf::from("/usr/bin"),
            PathBuf::from("/bin"),
        ])
        .unwrap()
    }

    /// A git call of a person, in `dir`, with the HOME the hook finds `moai` in.
    fn hand(&self, dir: &Path, args: &[&str]) -> Output {
        common::git::command()
            .arg("-C")
            .arg(dir)
            .args([
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
            ])
            .args(args)
            .env("HOME", self.home.path())
            .env("PATH", self.path_var())
            .env_remove(SESSION_VAR)
            .stdin(Stdio::null())
            .output()
            .expect("run git")
    }

    /// Installs the guard hook the way a user does, and only that way: `git-hooks-plan`, then
    /// `git-hooks-apply` with the token the plan prints, with a `moai` where the hook looks.
    fn install_hooks(&self) {
        let bin = moai_cli::home_root::path(self.home.path()).join("bin");
        fs::create_dir_all(&bin).unwrap();
        std::os::unix::fs::symlink(common::MOAI_BIN, bin.join("moai")).unwrap();
        // The plan reads the profile; a build that does not know `sync_sources` yet reads the
        // profile of the install without it.
        let saved = self.sources.replace(None);
        self.write_profile();
        let plan = result(&self.run(None, &["git-hooks-plan", &self.p]));
        let token = plan["confirm_token"].as_str().expect("a token").to_string();
        let applied = result(&self.run(
            None,
            &["git-hooks-apply", &self.p, "--confirm-git-hooks", &token],
        ));
        assert_eq!(applied["guard"]["record"], "written", "{applied}");
        *self.sources.borrow_mut() = saved;
    }

    /// The hand moves of the integration branch: each refused with the branch where it was.
    /// `wt` is the worktree that has the branch checked out.
    fn hand_moves_stay_refused(&self, wt: &Path) {
        let refuse = |dir: &Path, args: &[&str]| -> String {
            let before = self.integ();
            let out = self.hand(dir, args);
            assert!(
                !out.status.success(),
                "git {args:?} was not refused: {}",
                String::from_utf8_lossy(&out.stdout)
            );
            assert_eq!(self.integ(), before, "git {args:?} moved the branch");
            String::from_utf8_lossy(&out.stderr).into_owned()
        };
        let lane = self.rev("refs/heads/lane");
        let stderr = refuse(&self.project, &["update-ref", "refs/heads/integ", &lane]);
        assert!(stderr.contains("moai ref guard"), "{stderr}");
        refuse(&self.project, &["branch", "-f", BRANCH, &lane]);
        refuse(&self.project, &["branch", "-D", BRANCH]);
        refuse(wt, &["merge", "--no-ff", "-m", "a hand merge", "lane"]);
        // The refused merge leaves its MERGE_HEAD in that worktree, and a worktree with an
        // operation in progress is one the sync's follow step refuses to touch; a person who
        // met the refusal aborts the merge, so the fixture does too.
        let aborted = self.hand(wt, &["merge", "--abort"]);
        assert!(
            aborted.status.success(),
            "git merge --abort failed: {}",
            String::from_utf8_lossy(&aborted.stderr)
        );
        refuse(&self.project, &["branch", "-M", "lane", BRANCH]);
        assert!(
            !self.rev("refs/heads/lane").is_empty(),
            "the rename left the source branch alone"
        );
    }
}

// ---- AC-1 -----------------------------------------------------------------------------------

/// AC-1: a registered leader merges a listed source: the branch moves once, the new tip's parents
/// are the old tip and the source tip, one reflog entry is added and nothing of the source is
/// left outside the branch.
#[test]
fn ac1_a_registered_leader_merges_a_listed_source_in_one_guarded_move() {
    let w = World::new();
    let (tip, src) = w.diverge();
    let reflog_before = w.reflog();
    w.set_sources(&[SOURCE]);
    let r = result(&w.lead(SOURCE));
    let merge = r["merge"].as_str().expect("a merge commit").to_string();
    assert_eq!(merge.len(), 40, "{r}");
    assert_eq!(r["action"], json!("land-sync"), "{r}");
    assert_eq!(r["moved"], json!(true), "{r}");
    assert_eq!(r["already_synced"], json!(false), "{r}");
    assert_eq!(r["from"], json!(SOURCE), "{r}");
    assert_eq!(r["from_tip"], json!(src), "{r}");
    assert_eq!(r["spec"], json!(SPEC), "{r}");
    assert_eq!(r["to_tip"], json!(merge), "{r}");
    assert!(r["branch"].as_str().unwrap().contains(BRANCH), "{r}");
    assert_eq!(r["rev_list_count"], json!(0), "{r}");
    let fields = r.as_object().unwrap();
    for key in [
        "variants",
        "profile_disagrees",
        "worktree_follow",
        "runs",
        "writes",
    ] {
        assert!(fields.contains_key(key), "{key} is missing: {r}");
    }
    // The branch moved once, to the merge, with the old tip first and the source second.
    assert_eq!(w.integ(), merge);
    assert_eq!(w.parents(&merge), vec![tip.clone(), src.clone()]);
    assert_eq!(w.reflog(), reflog_before + 1);
    let behind = git(
        &w.project,
        &["rev-list", "--count", &format!("{BRANCH}..{SOURCE}")],
    );
    assert_eq!(behind, "0");
    let files = git(&w.project, &["ls-tree", "-r", "--name-only", &merge]);
    assert!(
        files.contains("notes/tip.md") && files.contains("notes/h.md"),
        "{files}"
    );
    let body = git(&w.project, &["cat-file", "-p", &merge]);
    assert!(body.contains(&format!("Moai-Spec: {SPEC}")), "{body}");
    assert!(!w.intent().exists(), "the intent file stays after the call");
}

/// AC-1: a source the tip already holds answers `moved: false` and writes nothing, and no check
/// runs.
#[test]
fn ac1_a_source_already_inside_the_tip_answers_moved_false_and_writes_nothing() {
    let w = World::new();
    let (tip, _src) = w.diverge();
    w.point(SOURCE, &w.c0);
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    let r = result(&w.lead(SOURCE));
    assert_eq!(r["moved"], json!(false), "{r}");
    assert_eq!(r["already_synced"], json!(true), "{r}");
    assert_eq!(r["to_tip"], json!(tip), "{r}");
    assert_eq!(w.snap(), before);
    assert_eq!(w.sync_rows(), 0);
    assert!(!w.intent().exists());
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
}

/// AC-1: a tip that moved after the merge was built, while the wide check ran, is refused as
/// `stale_base` and the branch stays where the other writer put it.
#[test]
fn ac1_a_tip_that_moved_after_the_merge_was_built_is_refused_stale_base() {
    let w = World::new();
    let (tip, _src) = w.diverge();
    let other = w.commit_on(&tip, "notes/x.md", "x\n");
    // The command's own wide run is the only check that can cover the merge, so the other writer
    // moves the branch from inside its first cargo step.
    w.set_cargo_hook(&format!(
        "if [ ! -e \"$FAKE_CARGO_CALLS.moved\" ]; then\n  : > \"$FAKE_CARGO_CALLS.moved\"\n  '{}' -C '{}' update-ref refs/heads/{BRANCH} {other}\nfi",
        w.real_git.display(),
        w.p
    ));
    w.set_sources(&[SOURCE]);
    let events = w.snap().events;
    let out = w.lead(SOURCE);
    let message = refusal(&out);
    assert!(message.contains("stale_base"), "{message}");
    assert!(!w.cargo_calls().is_empty(), "the check never ran");
    assert_eq!(
        w.integ(),
        other,
        "the command moved a branch it no longer owned"
    );
    assert_eq!(
        w.snap().events,
        events,
        "a refused call wrote to the ledger"
    );
    assert_eq!(w.sync_rows(), 0);
    assert!(!w.intent().exists());
}

// ---- AC-2 -----------------------------------------------------------------------------------

/// AC-2: a lane session and a call with no session id are refused, and the text names the
/// land-sync command; the registered leader session then merges.
#[test]
fn ac2_a_lane_session_and_a_call_with_no_session_id_are_refused() {
    let w = World::new();
    w.diverge();
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    for (session, cause) in [
        (Some(S_A), "not the session of the registered leader entry"),
        (None, "no session id"),
    ] {
        let out = w.sync(session, &[SOURCE, "--spec", SPEC, "--by", "leader"]);
        let message = w.refused(&out, &[cause], &before);
        assert!(message.contains(cause), "{message}");
        assert!(!message.contains("land-batch"), "{message}");
    }
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
    let r = result(&w.lead(SOURCE));
    assert_eq!(r["moved"], json!(true), "{r}");
}

/// AC-2: a missing `--by` and a `--by` that is not `leader` are refused, from the leader session
/// itself.
#[test]
fn ac2_a_missing_or_wrong_by_word_is_refused() {
    let w = World::new();
    w.diverge();
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    let out = w.sync(Some(S_LEAD), &[SOURCE, "--spec", SPEC]);
    assert!(!out.status.success(), "{}", said(&out));
    let text = said(&out);
    assert!(text.contains("--by"), "{text}");
    assert!(
        !text.contains("unrecognized"),
        "the command does not exist yet: {text}"
    );
    assert_eq!(w.snap(), before);
    let out = w.sync(Some(S_LEAD), &[SOURCE, "--spec", SPEC, "--by", "agent-1"]);
    let message = w.refused(&out, &["--by leader"], &before);
    assert!(!message.contains("land-batch"), "{message}");
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
}

/// AC-2: no leader entry, an entry with no leader mode and an entry whose lease has ended
/// confirm nobody, whatever session calls.
#[test]
fn ac2_no_entry_no_leader_mode_or_an_ended_lease_confirms_nobody() {
    // No entry at all.
    let w = World::bare();
    w.diverge();
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    let out = w.lead(SOURCE);
    let message = w.refused(&out, &["no live leader entry"], &before);
    assert!(!message.contains("land-batch"), "{message}");
    // An entry of the same name with no leader mode.
    w.register_leader(S_LEAD, false);
    let out = w.lead(SOURCE);
    w.refused(&out, &["no live leader entry"], &before);
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());

    // A leader entry whose lease ended.
    let w = World::new();
    w.diverge();
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    let file = w.leader_entry();
    let mut entry: Value = serde_json::from_str(&fs::read_to_string(&file).unwrap()).unwrap();
    entry["lease_until"] = json!(1);
    fs::write(&file, entry.to_string()).unwrap();
    let out = w.lead(SOURCE);
    let message = w.refused(&out, &["no live leader entry"], &before);
    assert!(!message.contains("land-batch"), "{message}");
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
}

// ---- AC-3 -----------------------------------------------------------------------------------

/// AC-3: a source the list does not hold, an empty list and a profile with no key are each
/// refused with a reason that names `sync_sources`.
#[test]
fn ac3_an_unlisted_source_and_an_empty_or_missing_list_name_the_key() {
    let w = World::new();
    w.diverge();
    let before = w.snap();
    w.set_sources(&["other"]);
    w.refused(&w.lead(SOURCE), &["sync_sources"], &before);
    w.set_sources(&[]);
    w.refused(&w.lead(SOURCE), &["sync_sources"], &before);
    w.clear_sources();
    w.refused(&w.lead(SOURCE), &["sync_sources"], &before);
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
}

/// AC-3: only a local branch the list holds exactly is a source: a tag, a remote-tracking name,
/// the full ref spelling, the integration branch, a missing branch and a malformed name are each
/// refused.
#[test]
fn ac3_only_a_local_branch_the_list_holds_exactly_is_a_source() {
    let w = World::new();
    let (_tip, src) = w.diverge();
    git(&w.project, &["update-ref", "refs/tags/v1", &src]);
    git(
        &w.project,
        &["update-ref", "refs/remotes/origin/hotfix", &src],
    );
    let before = w.snap();
    let cases: [(&[&str], &str); 6] = [
        (&[SOURCE, "v1"], "v1"),
        (&[SOURCE, "origin/hotfix"], "origin/hotfix"),
        (&[SOURCE], "refs/heads/hotfix"),
        (&[SOURCE], BRANCH),
        (&["ghost"], "ghost"),
        (&[SOURCE], "bad..name"),
    ];
    for (list, from) in cases {
        w.set_sources(list);
        let out = w.lead(from);
        w.refused(&out, &[from, "sync_sources"], &before);
    }
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
}

/// AC-3: an unapproved, a missing and a malformed SPEC are each refused with a reason that names
/// the SPEC.
#[test]
fn ac3_an_unapproved_missing_or_malformed_spec_is_refused() {
    let w = World::new();
    w.diverge();
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    for spec in [DRAFT_SPEC, "SPEC-20261117-ghost", "not-a-spec"] {
        let out = w.sync(Some(S_LEAD), &[SOURCE, "--spec", spec, "--by", "leader"]);
        w.refused(&out, &[spec], &before);
    }
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
}

/// AC-3: a source that changes the same lines as the tip is refused as `conflicts`, with no
/// check run.
#[test]
fn ac3_a_conflicting_source_is_refused_as_conflicts() {
    let w = World::new();
    let tip = w.commit_on(&w.c0, "a.txt", "tip side\n");
    let src = w.commit_on(&w.c0, "a.txt", "source side\n");
    w.point(BRANCH, &tip);
    w.point(SOURCE, &src);
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    let out = w.lead(SOURCE);
    w.refused(&out, &["conflicts"], &before);
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
}

/// AC-3: a source with no common history with the tip is refused as `join_failed`.
#[test]
fn ac3_an_unrelated_history_is_refused_as_join_failed() {
    let w = World::new();
    let tip = w.commit_on(&w.c0, "notes/tip.md", "t\n");
    w.point(BRANCH, &tip);
    let tree = git(&w.project, &["rev-parse", &format!("{}^{{tree}}", w.c0)]);
    let orphan = git(
        &w.project,
        &["commit-tree", &tree, "-m", "an unrelated root"],
    );
    w.point(SOURCE, &orphan);
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    let out = w.lead(SOURCE);
    w.refused(&out, &["join_failed"], &before);
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
}

/// AC-3: a tip somebody moved outside any landing is refused as `drift`: the ledger knows one
/// landing, the branch stands elsewhere.
#[test]
fn ac3_a_drifted_tip_is_refused_as_drift() {
    let w = World::new();
    let (_card, head) = w.land_one(1);
    let src = w.commit_on(&w.c0, "notes/h.md", "h\n");
    w.point(SOURCE, &src);
    let stray = w.commit_on(&head, "notes/stray.md", "s\n");
    w.point(BRANCH, &stray);
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    let out = w.lead(SOURCE);
    w.refused(&out, &["drift"], &before);
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
}

// ---- AC-4 -----------------------------------------------------------------------------------

/// AC-4: a check that cannot start gives `check_not_started`; the branch, the ledger and the
/// private worktree are as they were.
#[test]
fn ac4_a_check_that_cannot_start_gives_check_not_started() {
    let w = World::new();
    w.diverge();
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    let out = w.lead_check(SOURCE, &w.scripts.path().join("missing.sh"));
    w.refused(&out, &["check_not_started"], &before);
    assert!(!w.worktrees().contains("land-sync"), "{}", w.worktrees());
}

/// AC-4: a red check gives `check_failed`, after it ran once on the merge.
#[test]
fn ac4_a_red_check_gives_check_failed() {
    let w = World::new();
    w.diverge();
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    let check = w.script("red.sh", RED);
    let out = w.lead_check(SOURCE, &check);
    w.refused(&out, &["check_failed"], &before);
    assert_eq!(w.ranges().len(), 1, "{:?}", w.ranges());
    assert!(!w.worktrees().contains("land-sync"), "{}", w.worktrees());
}

/// AC-4: a replaced check program never counts as a full-selection PASS, even when it leaves the
/// line a full wide run would leave: `not_full_pass`.
#[test]
fn ac4_a_replaced_check_program_gives_not_full_pass_even_with_a_pass_line() {
    let line = "#!/bin/sh\necho \"$1\" >> \"$SYNC_LOG\"\nH=$(\"$MOAI_GIT_BIN\" rev-parse HEAD | cut -c1-12)\nmkdir -p \"target/check-affected-$H\"\necho \"check-affected base=000000000000 head=$H targets=1 result=PASS\" > \"target/check-affected-$H/summary.txt\"\nexit 0\n";
    for (why, text) in [("a bare exit 0", GREEN), ("a full pass line", line)] {
        let w = World::new();
        w.diverge();
        w.set_sources(&[SOURCE]);
        let before = w.snap();
        let check = w.script("check.sh", text);
        let out = w.lead_check(SOURCE, &check);
        let message = w.refused(&out, &["not_full_pass"], &before);
        assert!(message.contains("not_full_pass"), "{why}: {message}");
        assert_eq!(w.ranges().len(), 1, "{why}: {:?}", w.ranges());
        assert!(
            !w.worktrees().contains("land-sync"),
            "{why}: {}",
            w.worktrees()
        );
    }
}

/// AC-4: the full-selection PASS of the command's own in-process wide run lets the move happen:
/// the cargo steps ran once each, in a folder under `target/`, and the branch is the merge.
#[test]
fn ac4_the_wide_run_of_the_command_itself_lets_the_move_happen() {
    let w = World::new();
    w.diverge();
    w.set_sources(&[SOURCE]);
    let r = result(&w.lead(SOURCE));
    let merge = r["merge"].as_str().expect("a merge commit").to_string();
    assert_eq!(w.integ(), merge, "{r}");
    assert_eq!(w.cargo_runs("clippy"), 1, "{:?}", w.cargo_calls());
    assert_eq!(w.cargo_runs("fmt"), 1, "{:?}", w.cargo_calls());
    assert!(
        w.cargo_runs("nextest run") + w.cargo_runs("test") >= 1,
        "{:?}",
        w.cargo_calls()
    );
    assert!(
        w.cargo_calls().iter().all(|l| l.contains("/target/")),
        "the check ran in the checkout and not in a private worktree: {:?}",
        w.cargo_calls()
    );
    assert!(!w.worktrees().contains("land-sync"), "{}", w.worktrees());
}

// ---- AC-5 -----------------------------------------------------------------------------------

/// AC-5, the dangerous path itself: a repository whose hook was installed only through
/// `git-hooks-plan` and `git-hooks-apply` refuses every hand move of the integration branch
/// before and after a sync, and the sync as the registered leader moves the branch under it,
/// follows the worktree that has the branch checked out and leaves no intent file.
#[test]
fn ac5_the_installed_hook_refuses_every_hand_move_before_and_after_a_sync() {
    let w = World::new();
    let (tip, src) = w.diverge();
    let wt = w.scripts.path().join("integ-wt");
    git(
        &w.project,
        &["worktree", "add", "-q", wt.to_str().unwrap(), BRANCH],
    );
    let lane = w.empty_child(&tip);
    w.point("lane", &lane);
    w.install_hooks();
    w.hand_moves_stay_refused(&wt);
    assert_eq!(w.integ(), tip);

    w.set_sources(&[SOURCE]);
    let r = result(&w.lead(SOURCE));
    let merge = r["merge"].as_str().expect("a merge commit").to_string();
    assert_eq!(r["moved"], json!(true), "{r}");
    assert_eq!(w.integ(), merge);
    assert_eq!(w.parents(&merge), vec![tip, src]);
    assert!(!w.intent().exists(), "the intent file stays after the sync");
    // The worktree that has the branch checked out follows it and stays clean.
    assert_eq!(git(&wt, &["status", "--porcelain"]), "");
    assert!(wt.join("notes/h.md").is_file());

    w.hand_moves_stay_refused(&wt);
    assert_eq!(w.integ(), merge);
}

/// AC-5: a tip lost at the swap (another writer moves the branch between the last read and the
/// move) is refused as `stale_base`: nothing moved, the intent file is gone, no close row.
#[test]
fn ac5_a_tip_lost_at_the_swap_is_refused_stale_base_and_removes_the_intent_file() {
    let w = World::new();
    let (tip, _src) = w.diverge();
    let other = w.commit_on(&tip, "notes/x.md", "x\n");
    let marker = w.scripts.path().join("moved");
    let wrapper = w.script(
        "git-moves-first",
        &format!(
            "#!/bin/sh\nfor a in \"$@\"; do\n  if [ \"$a\" = update-ref ] && [ ! -e '{m}' ]; then\n    : > '{m}'\n    '{git}' -C '{p}' update-ref refs/heads/{BRANCH} {other}\n    break\n  fi\ndone\nexec '{git}' \"$@\"\n",
            m = marker.display(),
            git = w.real_git.display(),
            p = w.p
        ),
    );
    *w.git_bin.borrow_mut() = wrapper;
    w.set_sources(&[SOURCE]);
    let out = w.lead(SOURCE);
    // The start row is written before the move, so this refusal cannot promise "nothing was
    // changed"; it promises the branch, the close row and the intent file.
    let message = refusal_text(&out);
    assert!(message.contains("stale_base"), "{message}");
    assert!(!w.cargo_calls().is_empty(), "the loss was before the check");
    assert_eq!(w.integ(), other, "the command moved a branch it had lost");
    assert!(w.rows("card.synced").is_empty());
    assert!(!w.intent().exists(), "the refused move left an intent file");
}

// ---- AC-6 -----------------------------------------------------------------------------------

/// AC-6: the sync writes the card-less rows `card.sync_started` and `card.synced`, each naming the
/// merge, the source tip, the old tip and the SPEC.
#[test]
fn ac6_the_sync_writes_card_less_rows_that_name_the_merge_the_source_the_old_tip_and_the_spec() {
    let w = World::new();
    let (tip, src) = w.diverge();
    w.set_sources(&[SOURCE]);
    let r = result(&w.lead(SOURCE));
    let merge = r["merge"].as_str().expect("a merge commit").to_string();
    let started = w.rows("card.sync_started");
    let synced = w.rows("card.synced");
    assert_eq!(started.len(), 1, "{started:?}");
    assert_eq!(synced.len(), 1, "{synced:?}");
    assert!(started[0].0 < synced[0].0, "the start row comes first");
    for (_, card, payload) in started.iter().chain(synced.iter()) {
        assert_eq!(*card, None, "{payload}");
        assert_eq!(payload["merge"], json!(merge), "{payload}");
        assert_eq!(payload["from_tip"], json!(src), "{payload}");
        assert_eq!(payload["old_tip"], json!(tip), "{payload}");
        assert_eq!(payload["spec"], json!(SPEC), "{payload}");
        assert_eq!(payload["from"], json!(SOURCE), "{payload}");
        assert!(
            payload["branch"].as_str().unwrap().contains(BRANCH),
            "{payload}"
        );
    }
}

/// AC-6: after a landing and a sync, the next `card land` of an audited head is not refused as
/// `drift`: the merge the sync wrote is the tip the ledger expects.
#[test]
fn ac6_the_next_landing_of_an_audited_head_is_not_refused_as_drift() {
    let w = World::new();
    // The ledger holds a landing, so the drift check has a tip to compare with.
    w.land_one(1);
    let src = w.commit_on(&w.c0, "notes/h.md", "h\n");
    w.point(SOURCE, &src);
    w.set_sources(&[SOURCE]);
    let r = result(&w.lead(SOURCE));
    let merge = r["merge"].as_str().expect("a merge commit").to_string();
    assert_eq!(w.integ(), merge);
    let (card, head, out) = w.land_next(2);
    let text = said(&out);
    assert!(
        !text.contains("drift"),
        "the landing after a sync is refused as drift: {text}"
    );
    let got = result(&out);
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(w.integ(), head);
    assert_eq!(w.state(&card), "done");
}

/// AC-6: the same call again, after the sync, answers `already_synced` and writes nothing: no
/// ref, no row and no check.
#[test]
fn ac6_the_same_call_again_answers_already_synced_and_writes_nothing() {
    let w = World::new();
    w.diverge();
    w.set_sources(&[SOURCE]);
    let first = result(&w.lead(SOURCE));
    let merge = first["merge"].as_str().expect("a merge commit").to_string();
    let before = w.snap();
    let calls = w.cargo_calls().len();
    let again = result(&w.lead(SOURCE));
    assert_eq!(again["moved"], json!(false), "{again}");
    assert_eq!(again["already_synced"], json!(true), "{again}");
    assert_eq!(again["to_tip"], json!(merge), "{again}");
    assert_eq!(w.snap(), before);
    assert_eq!(w.cargo_calls().len(), calls, "the second call ran a check");
    assert_eq!(w.rows("card.sync_started").len(), 1);
    assert_eq!(w.rows("card.synced").len(), 1);
    assert!(!w.intent().exists());
}

/// AC-6: `card export` reads a ledger that holds the card-less sync rows and carries them.
#[test]
fn ac6_card_export_reads_a_ledger_that_holds_sync_rows() {
    let w = World::new();
    w.diverge();
    w.set_sources(&[SOURCE]);
    result(&w.lead(SOURCE));
    let out_dir = w.scripts.path().join("export");
    let exported = w.card(None, &["export", "--out", out_dir.to_str().unwrap()]);
    assert_eq!(exported.status.code(), Some(0), "{}", said(&exported));
    let events = fs::read_to_string(out_dir.join("events.jsonl")).unwrap();
    assert!(events.contains("card.sync_started"), "{events}");
    assert!(events.contains("card.synced"), "{events}");
}

// ---- the cases the first set left open --------------------------------------------------------

/// AC-4: the command's private worktree is removed after every outcome of the wide check: the
/// folder it was made in (`run_check` names it under `target/land-batch/`) and git's own list of
/// worktrees hold nothing of it. The `land-sync` word the tests above look for in the list never
/// appears, because the folder is the batch's, so this one reads the folder itself.
#[test]
fn ac4_the_private_worktree_is_removed_after_every_outcome_of_the_check() {
    let cases: [(&str, Option<&str>); 4] = [
        ("cannot start", None),
        ("red", Some(RED)),
        ("replaced", Some(GREEN)),
        ("full pass", None),
    ];
    for (why, text) in cases {
        let w = World::new();
        w.diverge();
        w.set_sources(&[SOURCE]);
        let out = match (why, text) {
            ("cannot start", _) => w.lead_check(SOURCE, &w.scripts.path().join("missing.sh")),
            ("full pass", _) => w.lead(SOURCE),
            (_, Some(text)) => w.lead_check(SOURCE, &w.script("check.sh", text)),
            _ => unreachable!(),
        };
        assert_eq!(
            out.status.code(),
            Some(if why == "full pass" { 0 } else { 1 }),
            "{why}: {}",
            said(&out)
        );
        assert_eq!(
            w.private_worktree_folders(),
            Vec::<String>::new(),
            "{why}: a private worktree was left"
        );
        let listed = w.worktrees();
        assert_eq!(
            listed.matches("worktree ").count(),
            1,
            "{why}: only the checkout is listed: {listed}"
        );
    }
}

/// AC-5 (REQ-7): a worktree that has the branch checked out and holds a tracked change at a path
/// the sync changes is refused before the wide check and before any write: the change is the
/// owner's and nothing deletes it. (A staged file at the path the source adds is such a change.)
#[test]
fn ac5_a_dirty_checkout_of_the_branch_is_refused_before_the_check_and_any_write() {
    let w = World::new();
    w.diverge();
    let wt = w.scripts.path().join("integ-wt");
    git(
        &w.project,
        &["worktree", "add", "-q", wt.to_str().unwrap(), BRANCH],
    );
    fs::write(wt.join("notes/h.md"), "my work\n").unwrap();
    git(&wt, &["add", "notes/h.md"]);
    w.set_sources(&[SOURCE]);
    let before = w.snap();
    let out = w.lead(SOURCE);
    let message = w.refused(&out, &["overlapping tracked changes"], &before);
    assert!(message.contains("notes/h.md"), "{message}");
    assert!(
        w.cargo_calls().is_empty(),
        "the check ran: {:?}",
        w.cargo_calls()
    );
    assert_eq!(
        fs::read_to_string(wt.join("notes/h.md")).unwrap(),
        "my work\n",
        "the refusal touched the owner's change"
    );
}

/// AC-5 (REQ-7): a tracked change of the checkout at a path the sync does not change does not stop
/// the sync: the branch moves, the worktree follows and keeps the owner's change, and the answer
/// counts it in `kept_changes`.
#[test]
fn ac5_a_tracked_change_the_sync_does_not_touch_is_kept_and_the_sync_goes_on() {
    let w = World::new();
    let (tip, src) = w.diverge();
    let wt = w.scripts.path().join("integ-wt");
    git(
        &w.project,
        &["worktree", "add", "-q", wt.to_str().unwrap(), BRANCH],
    );
    fs::write(wt.join("notes/tip.md"), "my work\n").unwrap();
    w.set_sources(&[SOURCE]);
    let r = result(&w.lead(SOURCE));
    let merge = r["merge"].as_str().expect("a merge commit").to_string();
    assert_eq!(r["moved"], json!(true), "{r}");
    assert_eq!(w.integ(), merge);
    assert_eq!(w.parents(&merge), vec![tip, src]);
    assert_eq!(r["worktree_follow"]["kept_changes"], json!(1), "{r}");
    assert_eq!(
        fs::read_to_string(wt.join("notes/tip.md")).unwrap(),
        "my work\n",
        "the sync touched the owner's change"
    );
    assert!(
        wt.join("notes/h.md").is_file(),
        "the worktree did not follow"
    );
}

/// AC-6 (REQ-6): a sync that stopped after the move (the branch is on the merge, the start row is
/// written, the closing row is not, and the worktree that has the branch checked out is still on
/// the old tree) is closed by the same call: no merge, no check, no second start row; the
/// worktree follows, the closing row is written and the answer says `closed_previous`. The call
/// after it writes nothing.
#[test]
fn ac6_a_sync_stopped_after_the_move_is_closed_by_the_same_call() {
    let w = World::new();
    let (tip, src) = w.diverge();
    let wt = w.scripts.path().join("integ-wt");
    git(
        &w.project,
        &["worktree", "add", "-q", wt.to_str().unwrap(), BRANCH],
    );
    // The sync as the interrupted call left it: the merge exists and the branch stands on it.
    let tree = git(&w.project, &["merge-tree", "--write-tree", &tip, &src]);
    let tree = tree.lines().next().unwrap().to_string();
    let merge = git(
        &w.project,
        &[
            "commit-tree",
            &tree,
            "-p",
            &tip,
            "-p",
            &src,
            "-m",
            &signed("a sync merge", S_LEAD),
        ],
    );
    w.point(BRANCH, &merge);
    w.insert_cardless_event(
        "leader",
        "card.sync_started",
        json!({
            "branch": BRANCH, "merge": merge, "from": SOURCE, "from_tip": src,
            "old_tip": tip, "spec": SPEC, "by": "leader",
        }),
    );
    assert!(
        !wt.join("notes/h.md").exists(),
        "the worktree is still behind"
    );
    w.set_sources(&[SOURCE]);
    let r = result(&w.lead(SOURCE));
    assert_eq!(r["moved"], json!(false), "{r}");
    assert_eq!(r["already_synced"], json!(true), "{r}");
    assert_eq!(r["closed_previous"], json!(true), "{r}");
    assert_eq!(r["merge"], json!(merge), "{r}");
    assert_eq!(r["to_tip"], json!(merge), "{r}");
    assert_eq!(r["writes"], json!(true), "{r}");
    assert_eq!(r["worktree_follow"]["status"], json!("finished"), "{r}");
    assert_eq!(w.integ(), merge);
    assert!(w.cargo_calls().is_empty(), "{:?}", w.cargo_calls());
    assert!(!w.intent().exists());
    assert!(
        wt.join("notes/h.md").is_file(),
        "the worktree did not follow"
    );
    assert_eq!(git(&wt, &["status", "--porcelain"]), "");
    let started = w.rows("card.sync_started");
    let synced = w.rows("card.synced");
    assert_eq!(started.len(), 1, "no second start row: {started:?}");
    assert_eq!(synced.len(), 1, "{synced:?}");
    let (_, card, payload) = &synced[0];
    assert_eq!(*card, None, "{payload}");
    assert_eq!(payload["merge"], json!(merge), "{payload}");
    assert_eq!(payload["old_tip"], json!(tip), "{payload}");
    assert_eq!(payload["from_tip"], json!(src), "{payload}");
    // The call after the close writes nothing.
    let before = w.snap();
    let again = result(&w.lead(SOURCE));
    assert_eq!(again["moved"], json!(false), "{again}");
    assert_eq!(again["closed_previous"], json!(false), "{again}");
    assert_eq!(again["writes"], json!(false), "{again}");
    assert_eq!(w.snap(), before);
}
