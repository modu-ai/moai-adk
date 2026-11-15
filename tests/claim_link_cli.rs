//! #3030: `claim-link` re-points an active claim record at the worktree and branch the lane
//! really works in, so `claim-verify --hook` guards the commits made there.
//!
//! Every test runs the real binary in a temporary repository with a temporary HOME and a stub
//! `gh` (`MOAI_FORGE_GH_BIN`): no network, no real GitHub, and the real HOME is never touched.
//! The claim journal is seeded through `Journal::commit`, as `claim-acquire` would write it.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use moai_cli::claim::{ClaimRecord, Journal, Owner, SpecRef, TaskRef, WorktreeId};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const KEY: &str = "0123456789abcdef";
const OTHER_KEY: &str = "fedcba9876543210";
const LANE: &str = "lane-a";

struct Fx {
    _dir: tempfile::TempDir,
    /// The base folder: the project, its sibling worktrees, the home and the stubs.
    base: PathBuf,
    root: PathBuf,
    home: PathBuf,
    gh: PathBuf,
    gh_log: PathBuf,
}

fn secs(offset: i64) -> String {
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64;
    moai_cli::forge::policy::format_utc(now + offset)
}

fn claim_line(session: &str, lease: &str) -> String {
    format!("moai-claim: session={session} epoch=1 lease_until={lease}")
}

fn comments(items: &[(i64, String)]) -> String {
    Value::Array(
        items
            .iter()
            .map(|(id, body)| json!({"id": id, "user": {"login": "GoosLab"}, "body": body}))
            .collect(),
    )
    .to_string()
}

/// The Issue's comments when lane `session` owns comment 100.
fn owner_body(session: &str) -> String {
    comments(&[(100, claim_line(session, &secs(3600)))])
}

fn git(dir: &Path, args: &[&str]) -> String {
    let full = [
        [
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "-c",
            "init.defaultBranch=main",
        ]
        .as_slice(),
        args,
    ]
    .concat();
    common::git::git(dir, &full)
}

impl Fx {
    /// A clone with one commit on `main` and a GitHub `origin`, a home, and a `gh` stub that
    /// answers with `$STUB_JSON` (or fails when `$STUB_FAIL` is set) and logs its arguments.
    fn new() -> Fx {
        let dir = tempfile::TempDir::new().unwrap();
        let base = fs::canonicalize(dir.path()).unwrap();
        let root = base.join("project");
        let home = base.join("home");
        fs::create_dir_all(&root).unwrap();
        fs::create_dir_all(&home).unwrap();
        git(&root, &["init", "-q", "-b", "main"]);
        git(&root, &["commit", "-q", "--allow-empty", "-m", "base"]);
        git(
            &root,
            &[
                "remote",
                "add",
                "origin",
                "https://github.com/acme/widgets.git",
            ],
        );
        let gh = base.join("gh-stub");
        let gh_log = base.join("gh.log");
        common::write_executable(
            &gh,
            "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$GH_LOG\"\nif [ -n \"$STUB_FAIL\" ]; then echo 'gh: boom' >&2; exit 1; fi\nprintf 'HTTP/2.0 200 OK\\r\\n\\r\\n%s' \"$STUB_JSON\"\n",
        );
        Fx {
            _dir: dir,
            base,
            root,
            home,
            gh,
            gh_log,
        }
    }

    fn head(&self, dir: &Path) -> String {
        git(dir, &["rev-parse", "HEAD"])
    }

    /// The record `claim-acquire` would have written from the project folder.
    fn record(&self, key: &str, issue: i64, comment: i64, lane: &str) -> ClaimRecord {
        let id = moai_cli::identity::identity(&self.root).unwrap();
        let text = |k: &str| id[k].as_str().unwrap().to_string();
        ClaimRecord {
            schema_version: 1,
            task_ref: TaskRef::new(
                "github",
                "github.com",
                "acme",
                "widgets",
                1,
                70 + issue,
                issue,
            ),
            task_key: key.into(),
            logical_key: format!("L{issue}"),
            owner: Owner::new("session", lane),
            assignee: None,
            provider_session: None,
            integration_owner: None,
            attempt: None,
            epoch: 1,
            base_revision: self.head(&self.root),
            worktree: WorktreeId::new(
                text("clone_anchor"),
                text("workspace_git_dir"),
                text("work_tree"),
                self.head(&self.root),
            )
            .branch("main"),
            write_scope: vec!["src/x.rs".into()],
            spec_ref: SpecRef {
                id: "SPEC-L".into(),
                revision: "r1".into(),
                digest: None,
            },
            comment_id: Some(comment),
            claim_token: Some(format!("Moai-Claim: {issue}#{comment}")),
            lease_until: Some(secs(3600)),
            state: "assigned".into(),
            effects: vec![],
            last_observed_at: "2026-11-09T05:00:00Z".into(),
        }
    }

    /// Seeds Issue 7, claim comment 100, owned by [`LANE`], at the project folder.
    fn seed(&self) -> ClaimRecord {
        let r = self.record(KEY, 7, 100, LANE);
        self.commit("acquire", &r);
        r
    }

    fn commit(&self, event: &str, r: &ClaimRecord) {
        let j = Journal::open(&moai_cli::claim::journal_root(&self.root).unwrap()).unwrap();
        assert!(j.try_lock().unwrap());
        j.commit(event, r).unwrap();
    }

    fn journal_dir(&self) -> PathBuf {
        moai_cli::claim::journal_root(&self.root).unwrap()
    }

    fn log(&self) -> String {
        fs::read_to_string(self.journal_dir().join("journal.log")).unwrap_or_default()
    }

    fn record_bytes(&self, key: &str) -> Vec<u8> {
        fs::read(self.journal_dir().join(format!("{key}.json"))).unwrap_or_default()
    }

    /// A worktree beside the project, made with `git worktree add -b`.
    fn worktree(&self, name: &str, branch: &str, from: &str) -> PathBuf {
        let path = self.base.join(name);
        git(
            &self.root,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                branch,
                path.to_str().unwrap(),
                from,
            ],
        );
        fs::canonicalize(path).unwrap()
    }

    /// Runs the binary in `cwd`; the stub `gh` answers `body`.
    fn run_in(&self, cwd: &Path, args: &[&str], body: &str, stdin: &str) -> Output {
        self.run_os(
            cwd,
            &args
                .iter()
                .map(std::ffi::OsString::from)
                .collect::<Vec<_>>(),
            body,
            stdin,
        )
    }

    fn run_os(&self, cwd: &Path, args: &[std::ffi::OsString], body: &str, stdin: &str) -> Output {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env("MOAI_FORGE_GH_BIN", &self.gh)
            .env("GH_LOG", &self.gh_log)
            .env("STUB_JSON", body)
            .args(args)
            .current_dir(cwd)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        if body == "FAIL" {
            cmd.env("STUB_FAIL", "1");
        }
        let mut child = cmd.spawn().unwrap();
        common::feed_stdin(&mut child, stdin.as_bytes());
        child.wait_with_output().unwrap()
    }

    /// `claim-link` with the usual flags for Issue 7, comment 100 and [`LANE`], for `worktree`,
    /// run in the project folder.
    fn link(&self, worktree: &Path, body: &str) -> Output {
        self.link_repo(worktree, "acme/widgets", body)
    }

    fn link_repo(&self, worktree: &Path, repo: &str, body: &str) -> Output {
        let mut args: Vec<std::ffi::OsString> =
            ["claim-link", ".", "--issue", "7", "--comment", "100"]
                .iter()
                .map(Into::into)
                .collect();
        let flags = ["--lane", LANE, "--repo", repo];
        args.extend(flags.iter().map(std::ffi::OsString::from));
        args.push("--worktree".into());
        args.push(worktree.into());
        self.run_os(&self.root, &args, body, "")
    }

    fn calls(&self) -> usize {
        fs::read_to_string(&self.gh_log)
            .map(|t| t.lines().count())
            .unwrap_or(0)
    }

    /// `claim-show` of [`KEY`]: the current record.
    fn shown(&self) -> Value {
        let out = self.run_in(&self.root, &["claim-show", ".", "--task", KEY], "", "");
        assert_eq!(out.status.code(), Some(0), "{out:?}");
        envelope(&out)["result"]["record"].clone()
    }

    /// `claim-verify --hook` for `git commit` in `cwd`.
    fn hook(&self, cwd: &Path, body: &str) -> Output {
        let input = json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
                           "tool_input": {"command": "git commit -m x"}, "cwd": cwd})
        .to_string();
        self.run_in(&self.root, &["claim-verify", ".", "--hook"], body, &input)
    }
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {out:?}"))
}

fn ok(out: &Output) -> Value {
    assert_eq!(out.status.code(), Some(0), "{out:?}");
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("claim.link"), "{v}");
    v["result"].clone()
}

/// A refusal: exit 1, `ok: false`, the pinned `error.kind`, and the message.
fn refused(out: &Output, kind: &str) -> String {
    assert_eq!(out.status.code(), Some(1), "{out:?}");
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(v["error"]["kind"], json!(kind), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
}

/// L1: a worktree made after the claim is not guarded until it is linked. Before the link the
/// hook sees no record for the folder, the branch is not `issue/<n>`, and it allows without a
/// check; after the link it verifies the Issue and `claim-show` names the new path and branch.
#[test]
fn worktree_made_after_the_claim_is_unguarded_until_linked() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let body = owner_body(LANE);

    let shown = fx.shown();
    assert_eq!(shown["worktree"]["path"], json!(fx.root.to_str().unwrap()));
    // The binary's hook prints nothing when it allows, so the verdict is read in process.
    let v = hook_verdict(&fx, &w, None);
    assert!(
        v["reason"]
            .as_str()
            .unwrap()
            .contains("not a lane worktree"),
        "{v}"
    );
    let out = fx.hook(&w, &body);
    assert_eq!(out.status.code(), Some(0), "{out:?}");
    assert_eq!(
        fx.calls(),
        0,
        "the unlinked folder is allowed with no check"
    );

    // Normal use: inside the new worktree, the project is `.` and `--worktree` is omitted.
    let args = [
        "claim-link",
        ".",
        "--issue",
        "7",
        "--comment",
        "100",
        "--lane",
        LANE,
        "--repo",
        "acme/widgets",
    ];
    let out = fx.run_in(&w, &args, &body, "");
    let r = ok(&out);
    assert_eq!(r["outcome"], json!("linked"));
    assert_eq!(r["to"]["path"], json!(w.to_str().unwrap()));

    let shown = fx.shown();
    assert_eq!(shown["worktree"]["path"], json!(w.to_str().unwrap()));
    assert_eq!(shown["worktree"]["branch"], json!("feat/x"));
    let v = hook_verdict(&fx, &w, Some(&body));
    assert_eq!(v["source"], json!("journal"), "{v}");
    assert_eq!(v["issue"], json!(7));
    assert_eq!(v["status"], json!("owner"));
    let before_calls = fx.calls();
    let out = fx.hook(&w, &body);
    assert_eq!(out.status.code(), Some(0), "{out:?}");
    assert_eq!(
        fx.calls(),
        before_calls + 1,
        "the linked folder verifies the Issue"
    );
}

/// The hook's verdict for `git commit` in `cwd`, in process: `Some(body)` answers the one read
/// of Issue 7's comments; `None` fails the test if the hook reads at all.
fn hook_verdict(fx: &Fx, cwd: &Path, body: Option<&str>) -> Value {
    use moai_cli::forge::transport::{Exchange, RecordedTransport};
    let exchanges = body
        .map(|b| {
            vec![Exchange {
                request: moai_cli::forge::transport::HttpRequest::new(
                    "GET",
                    "https://api.github.com/repos/acme/widgets/issues/7/comments?per_page=100",
                ),
                response: moai_cli::forge::transport::HttpResponse::new(200)
                    .with_body(serde_json::from_str(b).unwrap()),
            }]
        })
        .unwrap_or_default();
    let t = RecordedTransport::new(exchanges);
    let input = json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
                       "tool_input": {"command": "git commit -m x"}, "cwd": cwd})
    .to_string();
    let v = moai_cli::claim_verify::hook_cli(
        &fx.root,
        input.as_bytes(),
        &t,
        &moai_cli::forge::policy::SystemClock,
    )
    .unwrap();
    assert_eq!(t.sent.borrow().len(), usize::from(body.is_some()), "{v}");
    v
}

/// L2: a renamed branch is synced; the path does not change and one event is appended.
#[test]
fn renamed_branch_is_synced_by_link() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/old", "HEAD");
    let body = owner_body(LANE);
    ok(&fx.link(&w, &body));
    let lines = fx.log().lines().count();
    git(&w, &["branch", "-m", "feat/new"]);
    let r = ok(&fx.link(&w, &body));
    assert_eq!(r["outcome"], json!("linked"));
    assert_eq!(r["changed"], json!(["branch"]));
    assert_eq!(r["from"]["path"], r["to"]["path"]);
    assert_eq!(fx.log().lines().count(), lines + 1);
    assert!(fx
        .log()
        .lines()
        .last()
        .unwrap()
        .contains("\"event\":\"worktree_linked\""));
    assert_eq!(fx.shown()["worktree"]["branch"], json!("feat/new"));
}

/// L3: a worktree whose HEAD lacks the claim base is refused with the one git command that
/// fixes it, and the link works once that command ran.
#[test]
fn stale_head_is_refused_with_the_merge_command() {
    let fx = Fx::new();
    let old = fx.head(&fx.root);
    git(&fx.root, &["commit", "-q", "--allow-empty", "-m", "newer"]);
    let rec = fx.seed();
    let base = rec.base_revision.clone();
    assert_ne!(base, old);
    let w = fx.worktree("w-stale", "feat/stale", &old);
    let body = owner_body(LANE);
    let lines = fx.log().lines().count();
    let msg = refused(&fx.link(&w, &body), "conflict");
    let fix = format!("run `git -C '{}' merge --ff-only {base}`", w.display());
    assert!(
        msg.starts_with("conflict in claim link: base_not_in_head: "),
        "{msg}"
    );
    assert!(msg.ends_with(&fix), "{msg}");
    assert_eq!(fx.log().lines().count(), lines, "a refusal writes nothing");
    git(&w, &["merge", "-q", "--ff-only", &base]);
    assert_eq!(ok(&fx.link(&w, &body))["outcome"], json!("linked"));
}

/// N13: HEAD equal to the base and one commit later link; a lagging and a diverged HEAD are
/// refused, each with its own command.
#[test]
fn base_boundary_cases() {
    let fx = Fx::new();
    let old = fx.head(&fx.root);
    git(&fx.root, &["commit", "-q", "--allow-empty", "-m", "newer"]);
    let rec = fx.seed();
    let base = rec.base_revision;
    let body = owner_body(LANE);
    let equal = fx.worktree("w-equal", "feat/equal", &base);
    let r = ok(&fx.link(&equal, &body));
    assert_eq!(r["base_relation"], json!("equal"));
    git(&equal, &["commit", "-q", "--allow-empty", "-m", "plus one"]);
    let later = ok(&fx.link(&equal, &body));
    assert_eq!(later["base_relation"], json!("descends"));
    assert_eq!(later["outcome"], json!("unchanged"));
    let behind = fx.worktree("w-behind", "feat/behind", &old);
    let msg = refused(&fx.link(&behind, &body), "conflict");
    assert!(
        msg.ends_with(&format!(
            "`git -C '{}' merge --ff-only {base}`",
            behind.display()
        )),
        "{msg}"
    );
    git(
        &behind,
        &["commit", "-q", "--allow-empty", "-m", "own work"],
    );
    let msg = refused(&fx.link(&behind, &body), "conflict");
    assert!(
        msg.ends_with(&format!("`git -C '{}' merge {base}`", behind.display())),
        "{msg}"
    );
    assert!(!msg.contains("--ff-only"), "{msg}");
}

/// L4: a second link writes nothing; neither does a link after a new commit in the worktree.
#[test]
fn link_twice_writes_one_event() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let body = owner_body(LANE);
    ok(&fx.link(&w, &body));
    let log = fx.log();
    let bytes = fx.record_bytes(KEY);
    let r = ok(&fx.link(&w, &body));
    assert_eq!(r["outcome"], json!("unchanged"));
    assert_eq!(r["applied"], json!(false));
    assert_eq!(r["changed"], json!([]));
    assert_eq!(r["writes"], json!("none"));
    assert_eq!(fx.log(), log);
    assert_eq!(fx.record_bytes(KEY), bytes);
}

#[test]
fn a_new_commit_does_not_make_link_rewrite() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let body = owner_body(LANE);
    ok(&fx.link(&w, &body));
    let log = fx.log();
    let bytes = fx.record_bytes(KEY);
    git(&w, &["commit", "-q", "--allow-empty", "-m", "more work"]);
    let r = ok(&fx.link(&w, &body));
    assert_eq!(r["outcome"], json!("unchanged"));
    assert_eq!(
        r["to"]["head"],
        json!(fx.head(&w)),
        "the answer shows the measured HEAD"
    );
    assert_eq!(fx.log(), log, "the HEAD is not part of the idempotence key");
    assert_eq!(fx.record_bytes(KEY), bytes);
}

/// L5: after a link the record equals the old one except `worktree`, `last_observed_at` and
/// one added effect.
#[test]
fn link_changes_only_the_worktree() {
    let fx = Fx::new();
    fx.seed();
    let before = fx.shown();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    ok(&fx.link(&w, &owner_body(LANE)));
    let after = fx.shown();
    let strip = |v: &Value| {
        let mut v = v.clone();
        let o = v.as_object_mut().unwrap();
        o.remove("worktree");
        o.remove("last_observed_at");
        o.remove("effects");
        v
    };
    assert_eq!(strip(&before), strip(&after));
    let effects = after["effects"].as_array().unwrap();
    assert_eq!(
        effects.len(),
        before["effects"].as_array().unwrap().len() + 1
    );
    let e = effects.last().unwrap();
    assert_eq!(e["kind"], json!("worktree_linked"));
    assert_eq!(e["from"]["path"], before["worktree"]["path"]);
    assert_eq!(e["to"]["path"], json!(w.to_str().unwrap()));
    assert!(
        e.get("operation_id").is_none(),
        "bookkeeping, not a publish effect"
    );
    assert_ne!(before["last_observed_at"], after["last_observed_at"]);
}

/// N1-N4: an ownership refusal keeps `claim-verify`'s kind and writes nothing.
#[test]
fn non_owner_cannot_link() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let log = fx.log();
    let lower = comments(&[
        (50, claim_line("other-lane", &secs(3600))),
        (100, claim_line(LANE, &secs(3600))),
    ]);
    let msg = refused(&fx.link(&w, &lower), "conflict");
    assert!(msg.starts_with("claim verify: not owner"), "{msg}");
    assert_eq!(fx.log(), log);
}

#[test]
fn closed_claim_cannot_link() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let log = fx.log();
    let closed = comments(&[
        (100, claim_line(LANE, &secs(3600))),
        (110, "moai-claim-release: comment=100 by=leader".into()),
    ]);
    let msg = refused(&fx.link(&w, &closed), "conflict");
    assert!(msg.starts_with("claim verify: closed"), "{msg}");
    let handed = comments(&[
        (100, claim_line(LANE, &secs(3600))),
        (
            110,
            "moai-claim-handoff: comment=100 to=lane-z by=leader".into(),
        ),
    ]);
    let msg = refused(&fx.link(&w, &handed), "conflict");
    assert!(msg.starts_with("claim verify: closed"), "{msg}");
    assert_eq!(fx.log(), log);
}

#[test]
fn stale_claim_cannot_link() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let log = fx.log();
    let stale = comments(&[(100, claim_line(LANE, &secs(-3600)))]);
    let msg = refused(&fx.link(&w, &stale), "blocked");
    assert!(msg.starts_with("claim verify: stale"), "{msg}");
    assert_eq!(fx.log(), log);
}

#[test]
fn uncertain_claim_cannot_link() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let log = fx.log();
    let elsewhere = comments(&[(999, claim_line(LANE, &secs(3600)))]);
    let msg = refused(&fx.link(&w, &elsewhere), "blocked");
    assert!(msg.starts_with("claim verify: uncertain"), "{msg}");
    assert_eq!(fx.log(), log);
}

/// N5: `--lane` must be the record's owner, checked before any forge read.
#[test]
fn lane_must_match_the_record() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let log = fx.log();
    let args: Vec<std::ffi::OsString> = [
        "claim-link",
        ".",
        "--issue",
        "7",
        "--comment",
        "100",
        "--lane",
        "lane-b",
        "--repo",
        "acme/widgets",
        "--worktree",
    ]
    .iter()
    .map(Into::into)
    .chain([w.clone().into_os_string()])
    .collect();
    let out = fx.run_os(&fx.root, &args, &owner_body("lane-b"), "");
    let msg = refused(&out, "conflict");
    assert!(
        msg.starts_with("conflict in claim link: lane_differs: "),
        "{msg}"
    );
    assert_eq!(fx.calls(), 0);
    assert_eq!(fx.log(), log);
}

/// N6: `--comment` must be the record's claim comment.
#[test]
fn comment_must_match_the_record() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let args: Vec<std::ffi::OsString> = [
        "claim-link",
        ".",
        "--issue",
        "7",
        "--comment",
        "101",
        "--lane",
        LANE,
        "--repo",
        "acme/widgets",
        "--worktree",
    ]
    .iter()
    .map(Into::into)
    .chain([w.into_os_string()])
    .collect();
    let out = fx.run_os(&fx.root, &args, &owner_body(LANE), "");
    let msg = refused(&out, "conflict");
    assert!(
        msg.starts_with("conflict in claim link: comment_differs: "),
        "{msg}"
    );
    assert_eq!(fx.calls(), 0);
}

/// N7: the claim comment's `session=` must be `--lane`, even when the comment id is right.
#[test]
fn lane_must_match_the_claim_comment() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let log = fx.log();
    let other_session = comments(&[(100, claim_line("lane-b", &secs(3600)))]);
    let msg = refused(&fx.link(&w, &other_session), "conflict");
    assert!(
        msg.starts_with("conflict in claim link: session_differs: "),
        "{msg}"
    );
    assert_eq!(fx.log(), log);
}

/// N8: no active record for the Issue (none, another Issue, another repository, a closed
/// record) is `record_not_found`, with no forge call.
#[test]
fn no_record_is_refused() {
    let fx = Fx::new();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    // No journal at all.
    let msg = refused(&fx.link(&w, &owner_body(LANE)), "conflict");
    assert!(
        msg.starts_with("conflict in claim link: record_not_found: "),
        "{msg}"
    );
    fx.seed();
    // Another repository.
    let msg = refused(
        &fx.link_repo(&w, "acme/other", &owner_body(LANE)),
        "conflict",
    );
    assert!(msg.contains("record_not_found"), "{msg}");
    // A released record is not active.
    let mut released = fx.record(KEY, 7, 100, LANE);
    released.state = "cancelled".into();
    fx.commit("release", &released);
    let msg = refused(&fx.link(&w, &owner_body(LANE)), "conflict");
    assert!(msg.contains("record_not_found"), "{msg}");
    assert_eq!(fx.calls(), 0, "no forge call without a record");
}

/// N14: a record waiting for a handoff is not linked by its old owner.
#[test]
fn handoff_pending_cannot_link() {
    let fx = Fx::new();
    let mut r = fx.seed();
    r.state = "handoff_pending".into();
    fx.commit("handoff_pending", &r);
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let log = fx.log();
    let msg = refused(&fx.link(&w, &owner_body(LANE)), "conflict");
    assert!(msg.contains("state_not_linkable"), "{msg}");
    assert_eq!(fx.log(), log);
    assert_eq!(fx.calls(), 0);
}

/// N9: only a registered worktree of this clone is linked: a plain folder, another clone's
/// worktree, a deleted folder, a pruned worktree and a forged `.git` file are all refused and
/// the journal does not change.
#[test]
fn target_must_be_a_registered_worktree_of_this_repo() {
    let fx = Fx::new();
    fx.seed();
    let body = owner_body(LANE);
    let log = fx.log();
    // A plain folder.
    let plain = fx.base.join("plain");
    fs::create_dir(&plain).unwrap();
    let msg = refused(&fx.link(&plain, &body), "conflict");
    assert!(msg.contains("not_a_worktree"), "{msg}");
    // A worktree of another clone.
    let other = fx.base.join("other-clone");
    fs::create_dir(&other).unwrap();
    git(&other, &["init", "-q", "-b", "main"]);
    git(
        &other,
        &["commit", "-q", "--allow-empty", "-m", "elsewhere"],
    );
    let msg = refused(&fx.link(&other, &body), "conflict");
    assert!(msg.contains("common_dir_differs"), "{msg}");
    // A deleted folder.
    let gone = fx.base.join("never-made");
    let msg = refused(&fx.link(&gone, &body), "not_found");
    assert!(msg.contains("path unavailable"), "{msg}");
    // A pruned worktree: the folder stays, git's record of it is gone.
    let pruned = fx.worktree("w-pruned", "feat/pruned", "HEAD");
    let admin = fs::read_dir(fx.root.join(".git").join("worktrees"))
        .unwrap()
        .map(|e| e.unwrap().path())
        .find(|p| {
            fs::read_to_string(p.join("gitdir"))
                .is_ok_and(|g| g.trim().starts_with(pruned.to_str().unwrap()))
        })
        .unwrap();
    fs::remove_dir_all(admin).unwrap();
    let msg = refused(&fx.link(&pruned, &body), "conflict");
    assert!(msg.contains("not_a_worktree"), "{msg}");
    // A forged `.git` file that points at this clone's git directory.
    let forged = fx.base.join("forged");
    fs::create_dir(&forged).unwrap();
    fs::write(
        forged.join(".git"),
        format!("gitdir: {}\n", fx.root.join(".git").display()),
    )
    .unwrap();
    let msg = refused(&fx.link(&forged, &body), "conflict");
    assert!(
        msg.contains("not_a_worktree") || msg.contains("common_dir_differs"),
        "{msg}"
    );
    assert_eq!(fx.log(), log, "no refusal wrote anything");
}

/// N10: a symlink to a worktree is refused.
#[test]
fn symlinked_target_is_refused() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let link = fx.base.join("w-link");
    std::os::unix::fs::symlink(&w, &link).unwrap();
    let log = fx.log();
    let msg = refused(&fx.link(&link, &owner_body(LANE)), "invalid_input");
    assert!(msg.contains("path is a symlink; refused"), "{msg}");
    assert_eq!(fx.log(), log);
}

/// N11: another active record already holds the branch; the message names its Issue number.
#[test]
fn branch_held_by_another_record() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let mut other = fx.record(OTHER_KEY, 8, 200, "lane-b");
    other.worktree.path = fx.base.join("elsewhere").display().to_string();
    other.worktree.branch = Some("feat/x".into());
    fx.commit("acquire", &other);
    let log = fx.log();
    let msg = refused(&fx.link(&w, &owner_body(LANE)), "conflict");
    assert!(
        msg.starts_with("conflict in claim link: branch_held: "),
        "{msg}"
    );
    assert!(msg.contains("#8") && !msg.contains("lane-b"), "{msg}");
    assert_eq!(fx.log(), log);
}

/// N12: another active record already holds the path.
#[test]
fn path_held_by_another_record() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let mut other = fx.record(OTHER_KEY, 8, 200, "lane-b");
    other.worktree.path = w.display().to_string();
    other.worktree.branch = Some("feat/other".into());
    fx.commit("acquire", &other);
    let log = fx.log();
    let msg = refused(&fx.link(&w, &owner_body(LANE)), "conflict");
    assert!(
        msg.starts_with("conflict in claim link: path_held: "),
        "{msg}"
    );
    assert!(msg.contains("#8") && !msg.contains("lane-b"), "{msg}");
    assert_eq!(fx.log(), log);
}

/// N16: a failing `gh` is a transport error and writes nothing.
#[test]
fn forge_failure_writes_nothing() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let log = fx.log();
    let bytes = fx.record_bytes(KEY);
    let out = fx.link(&w, "FAIL");
    assert_eq!(out.status.code(), Some(1), "{out:?}");
    let v = envelope(&out);
    assert_eq!(v["error"]["kind"], json!("transport"), "{v}");
    assert_eq!(fx.log(), log);
    assert_eq!(fx.record_bytes(KEY), bytes);
}

/// B1: the lane bound is 128 bytes, a path with spaces links, and a path that is not UTF-8 is
/// refused.
#[test]
fn lane_and_path_limits() {
    // A 128-byte lane is accepted as the owner, a 129-byte lane is refused before any read.
    let fx = Fx::new();
    let long = "a".repeat(128);
    let r = fx.record(KEY, 7, 100, &long);
    fx.commit("acquire", &r);
    let w = fx.worktree("my work tree", "feat/spaces", "HEAD");
    let body = owner_body(&long);
    let args: Vec<std::ffi::OsString> = [
        "claim-link",
        ".",
        "--issue",
        "7",
        "--comment",
        "100",
        "--lane",
        &long,
        "--repo",
        "acme/widgets",
        "--worktree",
    ]
    .iter()
    .map(|s| std::ffi::OsString::from(*s))
    .chain([w.clone().into_os_string()])
    .collect();
    let out = fx.run_os(&fx.root, &args, &body, "");
    let res = ok(&out);
    assert_eq!(
        res["to"]["path"],
        json!(w.to_str().unwrap()),
        "a path with spaces links"
    );
    let too_long = "a".repeat(129);
    let calls = fx.calls();
    let args: Vec<std::ffi::OsString> = [
        "claim-link",
        ".",
        "--issue",
        "7",
        "--comment",
        "100",
        "--lane",
        &too_long,
        "--repo",
        "acme/widgets",
    ]
    .iter()
    .map(|s| std::ffi::OsString::from(*s))
    .collect();
    let out = fx.run_os(&fx.root, &args, &body, "");
    let msg = refused(&out, "invalid_input");
    assert!(
        msg.contains("--lane must be a bounded string (1 to 128 bytes)"),
        "{msg}"
    );
    assert_eq!(
        fx.calls(),
        calls,
        "the bound is checked before any forge call"
    );
    // A path that is not UTF-8 is covered by the unit test
    // `a_path_that_is_not_utf8_is_refused_before_git_runs`: the binary cannot be driven with
    // one, because every command panics in `std::env::args()` on a non-UTF-8 argument today.
}

/// B2: `--issue 0`, `--issue -1` and `--comment 0` are usage errors; the largest `i64` is
/// accepted by the parser and then fails on the record.
#[test]
fn numeric_flag_bounds() {
    let fx = Fx::new();
    fx.seed();
    let w = fx.worktree("w1", "feat/x", "HEAD");
    let body = owner_body(LANE);
    let wp = w.to_str().unwrap();
    for (flag, value) in [
        ("--issue", "0"),
        ("--issue", "-1"),
        ("--comment", "0"),
        ("--comment", "-5"),
    ] {
        let (issue, comment) = if flag == "--issue" {
            (value, "100")
        } else {
            ("7", value)
        };
        let args = [
            "claim-link",
            ".",
            "--issue",
            issue,
            "--comment",
            comment,
            "--lane",
            LANE,
            "--repo",
            "acme/widgets",
            "--worktree",
            wp,
        ];
        let out = fx.run_in(&fx.root, &args, &body, "");
        assert_eq!(out.status.code(), Some(2), "{flag} {value}: {out:?}");
        assert_eq!(fx.calls(), 0);
    }
    let max = i64::MAX.to_string();
    let args = [
        "claim-link",
        ".",
        "--issue",
        &max,
        "--comment",
        "100",
        "--lane",
        LANE,
        "--repo",
        "acme/widgets",
        "--worktree",
        wp,
    ];
    let msg = refused(&fx.run_in(&fx.root, &args, &body, ""), "conflict");
    assert!(msg.contains("record_not_found"), "{msg}");
    let args = [
        "claim-link",
        ".",
        "--issue",
        "7",
        "--comment",
        &max,
        "--lane",
        LANE,
        "--repo",
        "acme/widgets",
        "--worktree",
        wp,
    ];
    let msg = refused(&fx.run_in(&fx.root, &args, &body, ""), "conflict");
    assert!(msg.contains("comment_differs"), "{msg}");
    assert_eq!(fx.calls(), 0);
}
