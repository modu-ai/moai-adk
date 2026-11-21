//! `moai card finish` (SPEC-20261121-card-finish): a leader closes a ready, active or review card in
//! one step when its landing is proven (a head inside the integration branch) or, for a research
//! or audit card, when a report file or a verdict note is named; `card next` stops offering a
//! card that already landed and `card list` shows a `landed` key. Every refusal leaves every
//! ledger table as it was.
//!
//! Every command runs against a temporary HOME and a temporary git repository; the real `~/.moai`
//! is not read or written. The caller is the session id in the environment
//! (`CLAUDE_CODE_SESSION_ID`). `card pr` reads the pull request through a stub `gh`
//! (`MOAI_FORGE_GH_BIN`); no test reaches the network. A merged pull request link is made through
//! the CLI (`card pr`), so no test writes the ledger by hand except to end a lease, which has no
//! command (its length is at least 60 seconds).
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use moai_cli::ledger::reservation::Reservation;
use moai_cli::ledger::workflow::{self, finish_gate, FinishEvidence, WorkflowError};
use rusqlite::Connection;
use serde_json::{json, Value};
use std::cell::RefCell;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Output, Stdio};
use std::sync::{Arc, Barrier};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const S_L: &str = "cccccccc-0000-4000-8000-00000000000c";
const MERGE: &str = "0123456789abcdef0123456789abcdef01234567";
const PR_HEAD: &str = "89abcdef0123456789abcdef0123456789abcdef";
const UNKNOWN_HEAD: &str = "fedcba9876543210fedcba9876543210fedcba98";

/// The profile of an m-flow project whose integration branch is `integ`.
const MFLOW_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";
/// The same profile with no integration branch named.
const MFLOW_NO_BRANCH_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\"}}";
/// The profile a project has without any work flow (the stub `TempHome::project` writes).
const PLAIN_PROFILE: &[u8] = b"{\"schema_version\":2,\"project_id\":\"demo\"}\n";

/// The keys of one `card list` entry before this change.
const LIST_KEYS: [&str; 10] = [
    "name",
    "card_id",
    "title",
    "kind",
    "state",
    "priority",
    "rev",
    "updated_at",
    "source",
    "spec",
];

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

/// A commit with the tree of `parent`, the paragraphs as its message and `parent` as its only
/// parent; `branch` is moved to it without a checkout, so the working tree never changes.
fn commit_on(project: &Path, branch: &str, parent: &str, paragraphs: &[&str]) -> String {
    let tree = git(project, &["rev-parse", &format!("{parent}^{{tree}}")]);
    let mut args = vec!["commit-tree", tree.as_str(), "-p", parent];
    for paragraph in paragraphs {
        args.push("-m");
        args.push(paragraph);
    }
    let sha = git(project, &args);
    git(
        project,
        &["update-ref", &format!("refs/heads/{branch}"), &sha],
    );
    sha
}

/// A commit on `branch` (moved to it without a checkout) that adds `path` with `content` to the
/// tree of `parent`; returns the new commit.
fn put_on(project: &Path, branch: &str, parent: &str, path: &str, content: &str) -> String {
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
    let sha = git(
        project,
        &["commit-tree", &tree, "-p", parent, "-m", "docs: add a file"],
    );
    git(
        project,
        &["update-ref", &format!("refs/heads/{branch}"), &sha],
    );
    sha
}

/// An absolute `git` on this runner's `PATH`, for the child that has none.
fn git_executable() -> Option<PathBuf> {
    let path = std::env::var_os("PATH")?;
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

/// Asserts exit 0 and a success envelope, and unwraps the result.
fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn dump_tables(conn: &Connection, tables: &[&str]) -> Vec<String> {
    let mut out = Vec::new();
    for table in tables {
        let mut st = conn
            .prepare(&format!("SELECT * FROM {table} ORDER BY 1, 2"))
            .unwrap();
        let n = st.column_count();
        let lines: Vec<String> = st
            .query_map([], |r| {
                let mut line = format!("{table}:");
                for i in 0..n {
                    line.push_str(&format!("{:?}|", r.get_ref(i)?));
                }
                Ok(line)
            })
            .unwrap()
            .map(Result::unwrap)
            .collect();
        out.extend(lines);
    }
    out
}

/// A template's text with every `<...>` placeholder filled with the word `filled`.
fn filled(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(open) = rest.find('<') {
        let Some(close) = rest[open..].find('>') else {
            break;
        };
        out.push_str(&rest[..open]);
        out.push_str("filled");
        rest = &rest[open + close + 1..];
    }
    out.push_str(rest);
    out
}

/// A met reference of exactly `n` characters made of short words (so no word has the shape of
/// a secret), starting and ending with a letter.
fn reference_of(n: usize) -> String {
    assert!(n >= 3);
    let mut s = String::new();
    while s.len() + 2 < n {
        s.push_str("x ");
    }
    while s.len() < n {
        s.push('y');
    }
    assert_eq!(s.len(), n);
    s
}

/// The names of the entries of a `ready` or `skipped` list.
fn names(list: &Value) -> Vec<String> {
    list.as_array()
        .unwrap_or_else(|| panic!("not a list: {list}"))
        .iter()
        .map(|e| e["name"].as_str().unwrap().to_string())
        .collect()
}

/// The entry of `skipped` for one card, when there is one.
fn skipped_entry<'a>(listing: &'a Value, name: &str) -> Option<&'a Value> {
    listing["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["name"] == json!(name))
}

/// Asserts that a card of `card next` was skipped as already landed, with the proof named.
fn assert_already_landed(listing: &Value, name: &str) {
    let entry = skipped_entry(listing, name)
        .unwrap_or_else(|| panic!("{name} is not under skipped: {listing}"));
    assert_eq!(entry["reason"], json!("already_landed"), "{entry}");
    let via = entry["via"].as_str().unwrap_or("");
    assert!(!via.is_empty(), "the skip entry names its proof: {entry}");
    assert!(
        !names(&listing["ready"]).contains(&name.to_string()),
        "{name} is still offered under ready: {listing}"
    );
}

/// The world: a temporary HOME with a registered ledger, a temporary git repository as the
/// project and a stub `gh`. The repository has `main` at `head` (the card's commit), `integ`
/// one commit ahead of `head` (`tip`), and `scratch` with one commit outside `integ` (`orphan`).
/// The profile (m-flow with `integ` unless a test changes it) is written again before every
/// `finish` and `next` call, because some card commands rewrite it.
struct World {
    home: TempHome,
    _host: tempfile::TempDir,
    dir: PathBuf,
    project: PathBuf,
    p: String,
    gh: PathBuf,
    git_bin: Option<PathBuf>,
    profile: RefCell<Vec<u8>>,
    head: String,
    tip: String,
    orphan: String,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        git(&project, &["add", "a.txt"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        fs::write(project.join("b.txt"), "b\n").unwrap();
        git(&project, &["add", "b.txt"]);
        git(&project, &["commit", "-q", "-m", "the card's change"]);
        let head = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["branch", "integ"]);
        let tip = commit_on(&project, "integ", &head, &["integration moved on"]);
        let orphan = commit_on(&project, "scratch", &head, &["not integrated"]);
        let host = tempfile::Builder::new()
            .prefix("moai-test-host-")
            .tempdir()
            .unwrap();
        let dir = host.path().to_path_buf();
        let script = format!(
            "#!/bin/sh\nurl=''\nfor a in \"$@\"; do case \"$a\" in https://*) url=\"$a\";; esac; done\n\
             case \"$url\" in\n\
               */pulls/*) n=${{url##*/pulls/}}; if [ -f '{dir}/pull-'$n'.json' ]; then printf 'HTTP/2.0 200 OK\\r\\n\\r\\n'; cat '{dir}/pull-'$n'.json'; exit 0; fi;;\n\
             esac\n\
             printf 'HTTP/2.0 404 Not Found\\r\\n\\r\\n{{}}'; exit 1\n",
            dir = dir.display()
        );
        let gh = dir.join("gh-stub");
        common::write_executable(&gh, &script);
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            _host: host,
            dir,
            project,
            gh,
            git_bin: git_executable(),
            profile: RefCell::new(MFLOW_PROFILE.to_vec()),
            head,
            tip,
            orphan,
        }
    }

    /// Sets the profile `finish` and `next` run with.
    fn set_profile(&self, bytes: &[u8]) {
        *self.profile.borrow_mut() = bytes.to_vec();
    }

    fn write_profile(&self) {
        let cfg = self.project.join(".moai-cli/project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, &*self.profile.borrow()).unwrap();
    }

    /// `moai <args>` as `session`, with the stub `gh` and an absolute `git` named, not yet run.
    fn command(&self, session: Option<&str>, args: &[&str]) -> Command {
        let mut cmd = self.home.command(args);
        cmd.env("MOAI_FORGE_GH_BIN", &self.gh);
        if let Some(git) = &self.git_bin {
            cmd.env("MOAI_GIT_BIN", git);
        }
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd
    }

    /// `moai <args>` as `session`, run to its end.
    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        self.command(session, args).output().expect("run moai")
    }

    /// `card finish` as `session`, started and not waited for (the profile written last), so two
    /// of them can run at the same time.
    fn spawn_finish(&self, session: &str, card: &str, extra: &[&str]) -> Child {
        self.write_profile();
        let mut all = vec!["card", "finish", card];
        all.extend_from_slice(extra);
        all.extend_from_slice(&["--project", &self.p]);
        self.command(Some(session), &all)
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .expect("run moai")
    }

    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        self.run(session, &all)
    }

    /// `moai card finish <card> <extra> --project <project>`, as `session`, with the profile
    /// written last.
    fn finish(&self, session: Option<&str>, card: &str, extra: &[&str]) -> Output {
        self.write_profile();
        let mut all = vec!["card", "finish", card];
        all.extend_from_slice(extra);
        all.extend_from_slice(&["--project", &self.p]);
        self.run(session, &all)
    }

    /// `moai card next <project>` as `session`, with the profile written last.
    fn next(&self, session: &str) -> Value {
        self.write_profile();
        result(&self.run(Some(session), &["card", "next", &self.p]))
    }

    /// `moai card list <project>`.
    fn list(&self) -> Value {
        result(&self.run(None, &["card", "list", &self.p]))
    }

    /// `card note` with the text on stdin, as `S_B`.
    fn note(&self, card: &str, kind: &str, reference: &str, text: &str) {
        let mut child = self
            .home
            .command(&[
                "card",
                "note",
                card,
                "--kind",
                kind,
                "--ref",
                reference,
                "--by",
                "leader",
                "--project",
                &self.p,
            ])
            .env(SESSION_VAR, S_B)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .expect("run moai");
        common::feed_stdin(&mut child, text.as_bytes());
        result(&child.wait_with_output().unwrap());
    }

    /// A ready card of `kind` with one write scope and one open criterion per text; returns
    /// its name.
    fn card_named(&self, kind: &str, scope: &str, criteria: &[&str]) -> String {
        let mut args = vec![
            "add",
            "--title",
            "fixture card",
            "--kind",
            kind,
            "--scope",
            scope,
        ];
        for text in criteria {
            args.push("--ac");
            args.push(text);
        }
        let added = result(&self.card(None, &args));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    /// `S_A` takes the card, sends it to review and records `head` as landed. With `release`
    /// it then reopens and gives the card back (a ready card with a landing event); without,
    /// it stays in review with the reservation held.
    fn land(&self, card: &str, head: &str, release: bool) {
        result(&self.card(Some(S_A), &["take", card]));
        result(&self.card(Some(S_A), &["review", card, "--generation", "1"]));
        result(&self.card(
            Some(S_A),
            &["landed", card, "--head", head, "--by", "leader"],
        ));
        if release {
            result(&self.card(Some(S_A), &["reopen", card, "--generation", "1"]));
            result(&self.card(Some(S_A), &["release", card, "--generation", "1"]));
        }
    }

    /// What the host says about pull request `number` from now on: merged.
    fn host_merged_pr(&self, number: u64) {
        let pull = json!({
            "number": number, "state": "closed", "merged": true,
            "merged_at": "2026-11-17T00:00:00Z",
            "merge_commit_sha": MERGE,
            "head": {"sha": PR_HEAD},
        });
        fs::write(
            self.dir.join(format!("pull-{number}.json")),
            pull.to_string(),
        )
        .unwrap();
    }

    /// `S_A` takes the card, links the merged pull request `number` and gives the card back:
    /// a ready card with a merged pull request link.
    fn link_merged_pr(&self, card: &str, number: u64) {
        self.host_merged_pr(number);
        result(&self.card(Some(S_A), &["take", card]));
        let n = number.to_string();
        result(&self.card(
            Some(S_A),
            &["pr", card, &n, "--generation", "1", "--repo", "octo/demo"],
        ));
        result(&self.card(Some(S_A), &["release", card, "--generation", "1"]));
    }

    /// A draft SPEC with every placeholder filled, so `spec check` passes it; returns its id.
    fn draft_spec(&self, slug: &str) -> String {
        let out = self.run(
            None,
            &[
                "spec",
                "new",
                slug,
                "--title",
                "t",
                "--minimal",
                "--date",
                "2026-11-21",
                "--project",
                &self.p,
            ],
        );
        let id = result(&out)["id"].as_str().unwrap().to_string();
        let path = self.project.join(".moai/specs").join(&id).join("spec.md");
        fs::write(&path, filled(&fs::read_to_string(&path).unwrap())).unwrap();
        id
    }

    /// The reports folder of the temporary HOME.
    fn reports_root(&self) -> PathBuf {
        self.home.path().join(".moai/reports")
    }

    /// A regular file under the reports folder; returns its path.
    fn report_file(&self, relative: &str) -> String {
        let path = self.reports_root().join(relative);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(&path, "the findings\n").unwrap();
        path.to_str().unwrap().to_string()
    }

    /// Runs `statement` against the ledger directly (the only way to end a lease).
    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
    }

    fn end_lease(&self, card: i64) {
        self.sql(&format!(
            "UPDATE reservation SET lease_until = unixepoch() - 5 \
             WHERE card_id = {card} AND status = 'held'"
        ));
    }

    /// Every row of the ledger's tables, as text.
    fn dump(&self) -> Vec<String> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        dump_tables(
            l.conn(),
            &[
                "card",
                "card_ac",
                "card_note",
                "card_link",
                "reservation",
                "reservation_scope",
                "event",
            ],
        )
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    fn text(&self, sql: &str) -> Option<String> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    /// The `n` newest events as (kind, payload with the actor added), newest first.
    fn newest_events(&self, n: i64) -> Vec<(String, Value)> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l
            .conn()
            .prepare("SELECT kind, payload, actor FROM event ORDER BY event_id DESC LIMIT ?1")
            .unwrap();
        st.query_map([n], |r| {
            let payload: String = r.get(1)?;
            let mut v: Value = serde_json::from_str(&payload).unwrap();
            v["actor"] = json!(r.get::<_, String>(2)?);
            Ok((r.get::<_, String>(0)?, v))
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
    }

    /// Asserts a refusal and that it wrote nothing (AC-7): exit 1, `ok` false, the error kind,
    /// a reason that names the real cause (one of `any_of`), every table equal to `before` and
    /// no `fence_reject` event. A missing subcommand is exit 2, so it can never pass here.
    fn refuse(&self, out: &Output, kind: &str, any_of: &[&str], before: &[String]) -> String {
        assert_eq!(
            out.status.code(),
            Some(1),
            "{} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        let v = envelope(out);
        assert_eq!(v["ok"], json!(false), "{v}");
        assert_eq!(v["error"]["kind"], json!(kind), "{v}");
        let message = v["error"]["message"].as_str().unwrap().to_string();
        assert!(
            any_of.iter().any(|needle| message.contains(needle)),
            "the reason names none of {any_of:?}: {message}"
        );
        self.assert_unchanged(before);
        message
    }

    /// Every table is byte equal to `before` and no `fence_reject` event was written.
    fn assert_unchanged(&self, before: &[String]) {
        assert_eq!(self.dump(), before, "the call wrote to the ledger");
        assert_eq!(
            self.count("SELECT count(*) FROM event WHERE kind = 'fence_reject'"),
            0,
            "the call left a fence_reject event"
        );
    }
}

/// The tables other than `event`: a call that only records a landing leaves all of them as they
/// were.
const STATE_TABLES: [&str; 6] = [
    "card",
    "card_ac",
    "card_note",
    "card_link",
    "reservation",
    "reservation_scope",
];

/// The lines of one table in a dump.
fn table_lines(dump: &[String], table: &str) -> Vec<String> {
    let prefix = format!("{table}:");
    dump.iter()
        .filter(|l| l.starts_with(&prefix))
        .cloned()
        .collect()
}

/// Asserts that `after` differs from `before` by exactly `added` new `event` rows and by nothing
/// in any other table.
fn assert_only_events_added(before: &[String], after: &[String], added: usize) {
    for table in STATE_TABLES {
        assert_eq!(
            table_lines(after, table),
            table_lines(before, table),
            "the call changed the table {table}"
        );
    }
    assert_eq!(
        table_lines(after, "event").len(),
        table_lines(before, "event").len() + added,
        "the call wrote another number of events than {added}"
    );
}

// ---- AC-1: the landing proof --------------------------------------------------------------

/// AC-1: a ready card with two open criteria and a head inside the integration tip (behind
/// it, not equal to it) closes to done with a landing event and a finish event.
#[test]
fn ac1_a_ready_card_with_a_head_inside_the_tip_closes_to_done() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one", "two"]);
    let events_before = w.count("SELECT count(*) FROM event");
    assert_ne!(w.head, w.tip, "the fixture's head is behind the tip");

    let got = result(&w.finish(
        Some(S_L),
        &card,
        &[
            "--head", &w.head, "--met", "1=r1", "--met", "2=r2", "--by", "leader",
        ],
    ));
    assert_eq!(got["action"], json!("finish"), "{got}");
    assert_eq!(got["name"], json!(card), "{got}");
    assert_eq!(got["card_id"], json!(1), "{got}");
    assert_eq!(got["from"], json!("ready"), "{got}");
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(got["head"], json!(w.head), "{got}");
    assert_eq!(got["tip"], json!(w.tip), "{got}");
    assert_eq!(got["report"], Value::Null, "{got}");
    assert_eq!(got["by"], json!("leader"), "{got}");
    assert_eq!(got["session"], json!(S_L), "{got}");
    assert_eq!(
        got["previous"],
        Value::Null,
        "a ready card has no reservation"
    );
    assert_eq!(got["released"], json!(false), "{got}");
    assert_eq!(got["events"], json!(2), "{got}");
    assert_eq!(got["writes"], json!(true), "{got}");
    assert!(got["closed_at"].is_i64(), "{got}");
    assert!(
        got["via"].as_str().is_some_and(|v| !v.is_empty()),
        "the answer names the proof route: {got}"
    );
    let met = got["met"].as_array().unwrap();
    assert_eq!(met.len(), 2, "{got}");
    let met_text = got["met"].to_string();
    assert!(met_text.contains("r1") && met_text.contains("r2"), "{got}");

    // The ledger: done, both criteria met with their references, the two newest events, no
    // held reservation.
    assert_eq!(
        w.text("SELECT state FROM card WHERE card_id = 1")
            .as_deref(),
        Some("done")
    );
    assert!(
        w.count("SELECT closed_at FROM card WHERE card_id = 1") > 0,
        "closed_at is set"
    );
    assert_eq!(
        w.text(
            "SELECT group_concat(status || ':' || evidence_ref, ',') FROM \
             (SELECT status, evidence_ref FROM card_ac WHERE card_id = 1 ORDER BY seq)"
        )
        .as_deref(),
        Some("met:r1,met:r2")
    );
    assert_eq!(
        w.count("SELECT count(*) FROM reservation WHERE card_id = 1 AND status = 'held'"),
        0
    );
    assert_eq!(
        w.count("SELECT count(*) FROM event"),
        events_before + 2,
        "exactly two events"
    );
    let newest = w.newest_events(2);
    assert_eq!(newest[0].0, "card.finish", "{newest:?}");
    assert_eq!(newest[1].0, "card.landed", "{newest:?}");
    assert_eq!(newest[1].1["head"], json!(w.head), "{newest:?}");
    assert_eq!(newest[1].1["by"], json!("leader"), "{newest:?}");
    assert_eq!(newest[0].1["by"], json!("leader"), "{newest:?}");
    assert_eq!(newest[0].1["from"], json!("ready"), "{newest:?}");

    // A done card is not finished again.
    let before = w.dump();
    let again = w.finish(Some(S_L), &card, &["--head", &w.head, "--by", "leader"]);
    w.refuse(
        &again,
        "conflict",
        &["done", "ready, active or review"],
        &before,
    );
}

/// AC-1: the landing event carries the head, the tip that contained it when the call ran, the
/// generation of the card's newest reservation and the caller; a card that never had a
/// reservation is at generation 0, and one that had one keeps the number it had.
#[test]
fn ac1_the_landing_event_carries_the_head_the_tip_the_generation_and_the_caller() {
    let w = World::new();
    let never = w.card_named("task", "apps/a/", &["one"]);
    let given_back = w.card_named("task", "apps/b/", &["one"]);
    result(&w.card(Some(S_A), &["take", &given_back]));
    result(&w.card(Some(S_A), &["release", &given_back, "--generation", "1"]));
    assert_eq!(
        w.text("SELECT state FROM card WHERE card_id = 2")
            .as_deref(),
        Some("ready"),
        "the fixture gave the card back"
    );

    for (card, generation) in [(&never, 0), (&given_back, 1)] {
        let got = result(&w.finish(
            Some(S_L),
            card,
            &["--head", &w.head, "--met", "1=r1", "--by", "leader"],
        ));
        assert_eq!(got["state"], json!("done"), "{got}");
        assert_eq!(got["events"], json!(2), "{got}");
        let newest = w.newest_events(2);
        assert_eq!(newest[0].0, "card.finish", "{newest:?}");
        assert_eq!(newest[1].0, "card.landed", "{newest:?}");
        let landed = &newest[1].1;
        assert_eq!(landed["head"], json!(w.head), "{newest:?}");
        assert_eq!(landed["tip"], json!(w.tip), "{newest:?}");
        assert_eq!(landed["generation"], json!(generation), "{newest:?}");
        assert_eq!(landed["by"], json!("leader"), "{newest:?}");
        assert_eq!(landed["actor"], json!(S_L), "{newest:?}");
    }
}

/// AC-1: the head was fast-forwarded onto the integration branch and the ledger write never
/// happened (no landing event in the ledger at all): the same evidence closes the card and the
/// call records the landing itself.
#[test]
fn ac1_a_head_on_the_integration_branch_with_no_landing_event_closes_the_card() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one", "two"]);
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.landed'"),
        0,
        "the ledger has no landing event yet"
    );
    assert_eq!(
        git(
            &w.project,
            &["merge-base", "--is-ancestor", &w.head, "integ"]
        ),
        "",
        "the head is already on the integration branch"
    );

    let got = result(&w.finish(
        Some(S_L),
        &card,
        &[
            "--head", &w.head, "--met", "1=r1", "--met", "2=r2", "--by", "leader",
        ],
    ));
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(got["via"], json!("landing"), "{got}");
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.landed' AND card_id = 1"),
        1,
        "finish recorded the landing"
    );
    assert_eq!(
        w.text("SELECT json_extract(payload, '$.head') FROM event WHERE kind = 'card.landed'")
            .as_deref(),
        Some(w.head.as_str())
    );
}

/// AC-1, AC-7: a head that is not an ancestor of the integration tip is refused, naming the
/// rule, and nothing is written.
#[test]
fn ac1_a_head_outside_the_integration_tip_refuses() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one", "two"]);
    let before = w.dump();
    let out = w.finish(
        Some(S_L),
        &card,
        &[
            "--head", &w.orphan, "--met", "1=r1", "--met", "2=r2", "--by", "leader",
        ],
    );
    w.refuse(
        &out,
        "invalid_input",
        &["not contained in the integration branch"],
        &before,
    );
    assert_eq!(
        w.text("SELECT state FROM card WHERE card_id = 1")
            .as_deref(),
        Some("ready")
    );
}

/// AC-1, AC-7: a 40-hex head that names no commit of the repository is refused and nothing is
/// written.
#[test]
fn ac1_an_unknown_head_refuses() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one", "two"]);
    let before = w.dump();
    let out = w.finish(
        Some(S_L),
        &card,
        &[
            "--head",
            UNKNOWN_HEAD,
            "--met",
            "1=r1",
            "--met",
            "2=r2",
            "--by",
            "leader",
        ],
    );
    w.refuse(
        &out,
        "invalid_input",
        &[
            "not contained in the integration branch",
            "not a commit",
            "unknown commit",
            "does not exist",
        ],
        &before,
    );
}

/// AC-1, AC-7: a profile that names no integration branch (no work flow at all, or an m-flow
/// without the branch) refuses a landing proof, naming the missing branch, and writes nothing.
#[test]
fn ac1_a_profile_without_an_integration_branch_refuses() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one", "two"]);
    for profile in [PLAIN_PROFILE, MFLOW_NO_BRANCH_PROFILE] {
        w.set_profile(profile);
        let before = w.dump();
        let out = w.finish(
            Some(S_L),
            &card,
            &[
                "--head", &w.head, "--met", "1=r1", "--met", "2=r2", "--by", "leader",
            ],
        );
        w.refuse(&out, "invalid_input", &["integration branch"], &before);
    }
}

// ---- AC-2: one met option per open criterion ----------------------------------------------

/// AC-2: two open criteria with two met options close the card; a duplicate number and a number
/// that is not an open criterion are refused, each naming its cause, with every table unchanged.
/// (One met option for two open criteria is a landing-only call: AC-13.)
#[test]
fn ac2_two_met_options_close_two_open_criteria_and_a_duplicate_or_a_stranger_refuses() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one", "two"]);
    let before = w.dump();

    let twice = w.finish(
        Some(S_L),
        &card,
        &[
            "--head", &w.head, "--met", "1=r1", "--met", "1=again", "--met", "2=r2", "--by",
            "leader",
        ],
    );
    w.refuse(
        &twice,
        "invalid_input",
        &["twice", "more than once", "duplicate"],
        &before,
    );

    let stranger = w.finish(
        Some(S_L),
        &card,
        &[
            "--head", &w.head, "--met", "1=r1", "--met", "2=r2", "--met", "3=r3", "--by", "leader",
        ],
    );
    w.refuse(
        &stranger,
        "invalid_input",
        &["not an open criterion"],
        &before,
    );

    // With two met options naming the two open criteria the card closes.
    let got = result(&w.finish(
        Some(S_L),
        &card,
        &[
            "--head", &w.head, "--met", "1=r1", "--met", "2=r2", "--by", "leader",
        ],
    ));
    assert_eq!(got["state"], json!("done"), "{got}");
}

/// AC-2: a met reference of 500 characters passes (N) and one of 501 is refused naming the
/// limit (N+1); the refusal writes nothing.
#[test]
fn ac2_a_met_reference_of_500_characters_passes_and_501_refuses() {
    let w = World::new();
    let long = w.card_named("task", "apps/a/", &["one"]);
    let longer = w.card_named("task", "apps/b/", &["one"]);
    let n = reference_of(500);
    let n_plus_1 = reference_of(501);

    let before = w.dump();
    let refused = w.finish(
        Some(S_L),
        &longer,
        &[
            "--head",
            &w.head,
            "--met",
            &format!("1={n_plus_1}"),
            "--by",
            "leader",
        ],
    );
    let message = w.refuse(&refused, "invalid_input", &["500"], &before);
    assert!(message.contains("500"), "{message}");

    let got = result(&w.finish(
        Some(S_L),
        &long,
        &[
            "--head",
            &w.head,
            "--met",
            &format!("1={n}"),
            "--by",
            "leader",
        ],
    ));
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(
        w.text("SELECT evidence_ref FROM card_ac WHERE card_id = 1 AND seq = 1")
            .as_deref(),
        Some(n.as_str()),
        "the whole 500-character reference is stored"
    );
}

// ---- AC-3: the reservation gate -----------------------------------------------------------

/// A held reservation of `holder` whose lease ends at second `lease_until`.
fn held_until(holder: &str, lease_until: i64) -> Reservation {
    Reservation {
        reservation_id: 1,
        card_id: 1,
        generation: 1,
        holder: holder.to_string(),
        status: "held".to_string(),
        acquired_at: 0,
        lease_until,
        updated_at: 0,
        worktree: None,
        branch: None,
        write_scope: Vec::new(),
    }
}

/// AC-3 (N and N+1): the gate passes another session's lease that ends at the very second of the
/// call and refuses one that ends a second later; the caller's own lease passes either way.
#[test]
fn ac3_the_gate_passes_a_lease_that_ends_now_and_refuses_one_that_ends_a_second_later() {
    let now = 1_000;
    assert_eq!(
        finish_gate(&held_until(S_A, now), Some(S_L), now),
        Ok(()),
        "a lease that ends at the moment of the call has ended"
    );
    assert_eq!(
        finish_gate(&held_until(S_A, now - 1), Some(S_L), now),
        Ok(())
    );
    let refused = finish_gate(&held_until(S_A, now + 1), Some(S_L), now).unwrap_err();
    assert!(
        matches!(&refused, WorkflowError::FinishHeld { holder, .. } if holder == S_A),
        "{refused:?}"
    );
    let text = refused.to_string();
    assert!(
        text.contains("is held by") && text.contains(S_A) && text.contains("an ended lease"),
        "{text}"
    );
    assert_eq!(
        finish_gate(&held_until(S_L, now + 1), Some(S_L), now),
        Ok(()),
        "the caller's own live lease passes"
    );
}

/// AC-3: a caller without a session id holds nothing, so it passes only when the lease has
/// ended (also at the very second it ends).
#[test]
fn ac3_a_caller_without_a_session_id_passes_only_when_the_lease_has_ended() {
    let now = 1_000;
    assert_eq!(finish_gate(&held_until(S_A, now), None, now), Ok(()));
    assert_eq!(finish_gate(&held_until(S_A, now - 1), None, now), Ok(()));
    assert!(matches!(
        finish_gate(&held_until(S_A, now + 1), None, now),
        Err(WorkflowError::FinishHeld { .. })
    ));
}

/// AC-3: a live lease of another session refuses naming the holder (also for a caller with no
/// session id), an ended lease of another session passes, a live lease of the caller passes.
#[test]
fn ac3_a_live_lease_of_another_session_refuses_and_an_ended_one_passes() {
    let w = World::new();
    let held = w.card_named("task", "apps/a/", &["one"]);
    result(&w.card(Some(S_A), &["take", &held]));
    let args = ["--head", &w.head, "--met", "1=r1", "--by", "leader"];

    let before = w.dump();
    let by_other = w.finish(Some(S_L), &held, &args);
    let message = w.refuse(&by_other, "conflict", &["is held by"], &before);
    assert!(message.contains(S_A), "the holder is named: {message}");
    assert!(
        message.contains("an ended lease"),
        "the rule is named: {message}"
    );

    let no_session = w.finish(None, &held, &args);
    let message = w.refuse(&no_session, "conflict", &["is held by"], &before);
    assert!(message.contains(S_A), "{message}");

    // The lease ends: the same call by the same other session passes and gives the
    // reservation back.
    w.end_lease(1);
    let got = result(&w.finish(Some(S_L), &held, &args));
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(got["from"], json!("active"), "{got}");
    assert_eq!(got["previous"]["holder"], json!(S_A), "{got}");
    assert_eq!(got["released"], json!(true), "{got}");
    assert_eq!(
        w.count("SELECT count(*) FROM reservation WHERE card_id = 1 AND status = 'held'"),
        0
    );
    assert_eq!(
        w.text("SELECT status FROM reservation WHERE card_id = 1")
            .as_deref(),
        Some("released")
    );
}

/// AC-3: a live lease of the caller itself passes.
#[test]
fn ac3_a_live_lease_of_the_caller_passes() {
    let w = World::new();
    let own = w.card_named("task", "apps/a/", &["one"]);
    result(&w.card(Some(S_L), &["take", &own]));
    let got = result(&w.finish(
        Some(S_L),
        &own,
        &["--head", &w.head, "--met", "1=r1", "--by", "leader"],
    ));
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(got["previous"]["holder"], json!(S_L), "{got}");
    assert_eq!(got["released"], json!(true), "{got}");
    assert_eq!(
        w.count("SELECT count(*) FROM reservation WHERE card_id = 1 AND status = 'held'"),
        0
    );
}

// ---- AC-4: report and verdict evidence ------------------------------------------------------

/// AC-4: a research card closes with a regular file under the reports folder of the HOME (and,
/// on another card, one under the project folder), and refuses a missing file, a folder, a path
/// outside both, a path that leaves the reports folder by `..`, a link out of it and a path of
/// more than 500 characters, each with every table unchanged.
#[test]
fn ac4_a_research_card_closes_with_a_report_file_and_refuses_the_others() {
    let w = World::new();
    let research = w.card_named("research", "study/r/", &["report written"]);
    let report = w.report_file("demo/answer.md");
    let reports = w.reports_root();

    // A file the refusals point at: outside the reports folder and the project folder.
    let elsewhere = w.home.path().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::write(elsewhere.join("x.md"), "x\n").unwrap();
    fs::write(w.home.path().join(".moai/outside.md"), "x\n").unwrap();
    std::os::unix::fs::symlink(elsewhere.join("x.md"), reports.join("demo/link.md")).unwrap();
    // A regular file whose path is longer than 500 characters.
    let deep = reports
        .join("demo")
        .join("d".repeat(200))
        .join("e".repeat(200))
        .join("f".repeat(200));
    fs::create_dir_all(&deep).unwrap();
    fs::write(deep.join("long.md"), "x\n").unwrap();

    let by_path = |path: &str| {
        w.finish(
            Some(S_L),
            &research,
            &[
                "--evidence-report",
                path,
                "--met",
                "1=report written",
                "--by",
                "leader",
            ],
        )
    };
    let before = w.dump();
    let missing = reports.join("demo/missing.md");
    let cases: Vec<(String, Vec<&str>)> = vec![
        (
            missing.to_str().unwrap().to_string(),
            vec![
                "not a regular file",
                "does not exist",
                "not found",
                "no such file",
                "cannot read",
            ],
        ),
        (
            reports.join("demo").to_str().unwrap().to_string(),
            vec!["not a regular file"],
        ),
        (
            elsewhere.join("x.md").to_str().unwrap().to_string(),
            vec!["reports folder", "project folder"],
        ),
        (
            format!("{}/demo/../../outside.md", reports.display()),
            vec!["reports folder", "project folder"],
        ),
        (
            reports.join("demo/link.md").to_str().unwrap().to_string(),
            vec!["reports folder", "project folder"],
        ),
        (
            deep.join("long.md").to_str().unwrap().to_string(),
            vec!["500"],
        ),
    ];
    for (path, any_of) in &cases {
        let out = by_path(path);
        w.refuse(&out, "invalid_input", any_of, &before);
    }

    let got = result(&by_path(&report));
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(got["from"], json!("ready"), "{got}");
    assert_eq!(
        got["head"],
        Value::Null,
        "a report close has no head: {got}"
    );
    assert!(
        got["report"].as_str().unwrap().ends_with("answer.md"),
        "{got}"
    );
    assert_eq!(got["events"], json!(2), "{got}");
    let newest = w.newest_events(2);
    assert_eq!(newest[0].0, "card.finish", "{newest:?}");
    assert_eq!(newest[1].0, "card.landed", "{newest:?}");
    assert_eq!(newest[1].1["no_code"], json!(true), "{newest:?}");
    assert!(
        newest[1].1["head"].is_null(),
        "the landing event of a report close names no head: {newest:?}"
    );
    assert!(
        newest[1].1["report"]
            .as_str()
            .is_some_and(|r| r.ends_with("answer.md")),
        "{newest:?}"
    );

    // A report file under the project folder is accepted as well.
    let second = w.card_named("research", "study/s/", &["report written"]);
    let notes = w.project.join("notes");
    fs::create_dir_all(&notes).unwrap();
    fs::write(notes.join("report.md"), "findings\n").unwrap();
    let got = result(&w.finish(
        Some(S_L),
        &second,
        &[
            "--evidence-report",
            notes.join("report.md").to_str().unwrap(),
            "--met",
            "1=report written",
            "--by",
            "leader",
        ],
    ));
    assert_eq!(got["state"], json!("done"), "{got}");
}

/// AC-4 (N and N+1): a report path typed in 500 characters passes the length rule and one typed
/// in 501 is refused naming the limit; both name a regular file under the reports folder, so only
/// the length decides, and the refusal writes nothing.
#[test]
fn ac4_a_report_path_of_500_characters_passes_and_501_refuses() {
    let w = World::new();
    let research = w.card_named("research", "study/r/", &["report written"]);
    let base = w.reports_root().join("demo");
    // `<reports>/demo/<dir>/<stem>.md`, `n` characters in all.
    let file_of = |n: usize| -> String {
        let used = base.to_str().unwrap().len() + 2 + ".md".len();
        let free = n - used;
        let dir = base.join("d".repeat(free / 2));
        fs::create_dir_all(&dir).unwrap();
        let file = dir.join(format!("{}.md", "f".repeat(free - free / 2)));
        fs::write(&file, "x\n").unwrap();
        let text = file.to_str().unwrap().to_string();
        assert_eq!(text.len(), n);
        text
    };
    let n = file_of(500);
    let n_plus_1 = file_of(501);
    let by_path = |path: &str| {
        w.finish(
            Some(S_L),
            &research,
            &[
                "--evidence-report",
                path,
                "--met",
                "1=report written",
                "--by",
                "leader",
            ],
        )
    };

    let before = w.dump();
    let refused = by_path(&n_plus_1);
    let message = w.refuse(&refused, "invalid_input", &["500"], &before);
    assert!(
        message.contains("501"),
        "the refusal names the length: {message}"
    );

    let got = result(&by_path(&n));
    assert_eq!(got["state"], json!("done"), "{got}");
    assert!(
        got["report"].as_str().unwrap().ends_with(".md"),
        "the report is stored: {got}"
    );
}

/// AC-4: an audit card closes with a verdict note of its own (a decision or a report note) and
/// refuses a note that does not exist, a note of kind comment and a note of another card.
#[test]
fn ac4_an_audit_card_closes_with_a_verdict_note_and_refuses_the_others() {
    let w = World::new();
    let audit = w.card_named("audit", "study/a/", &["verdict written"]);
    let other = w.card_named("audit", "study/b/", &["verdict written"]);
    w.note(&audit, "comment", "chatter", "just talking");
    w.note(&other, "decision", "theirs", "a verdict of the other card");
    w.note(&audit, "decision", "verdict", "pass: nothing found");
    let note_refusal = [
        "report or decision",
        "no note",
        "not found",
        "does not exist",
    ];

    let by_note = |reference: &str| {
        w.finish(
            Some(S_L),
            &audit,
            &[
                "--evidence-note",
                reference,
                "--met",
                "1=verdict written",
                "--by",
                "leader",
            ],
        )
    };
    let before = w.dump();
    for reference in ["nothing-like-it", "chatter", "theirs"] {
        let out = by_note(reference);
        w.refuse(&out, "invalid_input", &note_refusal, &before);
    }

    let got = result(&by_note("verdict"));
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(got["head"], Value::Null, "{got}");
    let newest = w.newest_events(2);
    assert_eq!(newest[0].0, "card.finish", "{newest:?}");
    assert_eq!(newest[1].0, "card.landed", "{newest:?}");
    assert_eq!(newest[1].1["no_code"], json!(true), "{newest:?}");
    assert_eq!(newest[1].1["note"], json!("verdict"), "{newest:?}");

    // A report note is accepted as well.
    let third = w.card_named("audit", "study/c/", &["verdict written"]);
    w.note(&third, "report", "findings", "the findings");
    let got = result(&w.finish(
        Some(S_L),
        &third,
        &[
            "--evidence-note",
            "findings",
            "--met",
            "1=verdict written",
            "--by",
            "leader",
        ],
    ));
    assert_eq!(got["state"], json!("done"), "{got}");
}

/// AC-4: a research card and an audit card with an open criterion that has no met option refuse
/// (no landing-only answer for them: their evidence is a report or a verdict), with every table
/// unchanged, whether one criterion of two or all of them lack the option.
#[test]
fn ac4_a_research_or_audit_card_with_an_open_criterion_and_no_met_option_refuses() {
    let w = World::new();
    let research = w.card_named("research", "study/r/", &["found", "written"]);
    let audit = w.card_named("audit", "study/a/", &["checked"]);
    let report = w.report_file("demo/answer.md");
    w.note(&audit, "decision", "verdict", "pass: nothing found");
    let before = w.dump();
    let open_criterion = ["open criterion", "open criteria", "criterion"];

    let one_of_two = w.finish(
        Some(S_L),
        &research,
        &[
            "--evidence-report",
            &report,
            "--met",
            "1=found",
            "--by",
            "leader",
        ],
    );
    w.refuse(&one_of_two, "invalid_input", &open_criterion, &before);

    let none_of_two = w.finish(
        Some(S_L),
        &research,
        &["--evidence-report", &report, "--by", "leader"],
    );
    w.refuse(&none_of_two, "invalid_input", &open_criterion, &before);

    let audit_none = w.finish(
        Some(S_L),
        &audit,
        &["--evidence-note", "verdict", "--by", "leader"],
    );
    w.refuse(&audit_none, "invalid_input", &open_criterion, &before);
    for card in [1, 2] {
        assert_eq!(
            w.text(&format!("SELECT state FROM card WHERE card_id = {card}"))
                .as_deref(),
            Some("ready")
        );
    }
}

// ---- AC-5: the evidence kind follows the card kind ------------------------------------------

/// AC-5: a code card (kind fix) refuses report evidence and note evidence even with a good head;
/// a research card refuses a head even with a good report; both evidence options at once are
/// refused; every refusal writes nothing.
#[test]
fn ac5_a_code_card_refuses_report_evidence_and_a_research_card_refuses_a_head() {
    let w = World::new();
    let code = w.card_named("fix", "apps/a/", &["one"]);
    let research = w.card_named("research", "study/r/", &["report written"]);
    let report = w.report_file("demo/answer.md");
    w.note(&code, "decision", "verdict", "a verdict on a code card");
    w.note(&research, "decision", "verdict", "a verdict");
    let before = w.dump();

    let code_with_report = w.finish(
        Some(S_L),
        &code,
        &[
            "--head",
            &w.head,
            "--evidence-report",
            &report,
            "--met",
            "1=r1",
            "--by",
            "leader",
        ],
    );
    w.refuse(
        &code_with_report,
        "invalid_input",
        &["--evidence-report", "evidence report", "report evidence"],
        &before,
    );

    let code_with_note = w.finish(
        Some(S_L),
        &code,
        &[
            "--head",
            &w.head,
            "--evidence-note",
            "verdict",
            "--met",
            "1=r1",
            "--by",
            "leader",
        ],
    );
    w.refuse(
        &code_with_note,
        "invalid_input",
        &["--evidence-note", "evidence note", "note evidence"],
        &before,
    );

    let research_with_head = w.finish(
        Some(S_L),
        &research,
        &[
            "--head",
            &w.head,
            "--evidence-report",
            &report,
            "--met",
            "1=report written",
            "--by",
            "leader",
        ],
    );
    w.refuse(
        &research_with_head,
        "invalid_input",
        &["--head", "landing head"],
        &before,
    );

    let both = w.finish(
        Some(S_L),
        &research,
        &[
            "--evidence-report",
            &report,
            "--evidence-note",
            "verdict",
            "--met",
            "1=report written",
            "--by",
            "leader",
        ],
    );
    w.refuse(
        &both,
        "invalid_input",
        &["exactly one", "only one", "not both"],
        &before,
    );
}

// ---- AC-6: --by -----------------------------------------------------------------------------

/// AC-6: finish without `--by`, with an empty one and with a blank one refuses and writes
/// nothing; a given `--by` is stored in both events.
#[test]
fn ac6_finish_needs_a_by_and_stores_it_in_both_events() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one"]);
    let before = w.dump();

    let missing = w.finish(Some(S_L), &card, &["--head", &w.head, "--met", "1=r1"]);
    assert_eq!(
        missing.status.code(),
        Some(2),
        "{} {}",
        String::from_utf8_lossy(&missing.stdout),
        String::from_utf8_lossy(&missing.stderr)
    );
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&missing.stdout),
        String::from_utf8_lossy(&missing.stderr)
    );
    // An unknown subcommand answers "invalid command line"; a missing required option answers
    // this sentence and the usage error on stderr names the option.
    assert!(
        text.contains("required arguments were not provided") && text.contains("--by"),
        "the usage error names the missing option: {text}"
    );
    w.assert_unchanged(&before);

    for blank in ["", "   "] {
        let out = w.finish(
            Some(S_L),
            &card,
            &["--head", &w.head, "--met", "1=r1", "--by", blank],
        );
        w.refuse(&out, "invalid_input", &["--by"], &before);
    }

    let got = result(&w.finish(
        Some(S_L),
        &card,
        &["--head", &w.head, "--met", "1=r1", "--by", "release lead"],
    ));
    assert_eq!(got["by"], json!("release lead"), "{got}");
    let newest = w.newest_events(2);
    assert_eq!(newest[0].0, "card.finish", "{newest:?}");
    assert_eq!(newest[1].0, "card.landed", "{newest:?}");
    assert_eq!(newest[0].1["by"], json!("release lead"), "{newest:?}");
    assert_eq!(newest[1].1["by"], json!("release lead"), "{newest:?}");
}

// ---- AC-7: a refusal writes nothing ---------------------------------------------------------

/// AC-7: one refusal of each kind (landing proof, met options, reservation gate, report
/// evidence, `--by`, a card that is already done, a stale generation) leaves every table byte
/// equal to its dump before the call and writes no `fence_reject` event. The refusal helper also
/// checks the exit code, the error kind and the reason, so the check cannot pass on a command
/// that does not exist.
#[test]
fn ac7_every_kind_of_refusal_leaves_the_ledger_as_it_was() {
    let w = World::new();
    let code = w.card_named("task", "apps/a/", &["one", "two"]);
    let held = w.card_named("task", "apps/b/", &["one"]);
    let research = w.card_named("research", "study/r/", &["report written"]);
    let finished = w.card_named("task", "apps/d/", &["one"]);
    result(&w.card(Some(S_A), &["take", &held]));
    result(&w.finish(
        Some(S_L),
        &finished,
        &["--head", &w.head, "--met", "1=r1", "--by", "leader"],
    ));
    let before = w.dump();

    let landing = w.finish(
        Some(S_L),
        &code,
        &[
            "--head", &w.orphan, "--met", "1=r1", "--met", "2=r2", "--by", "leader",
        ],
    );
    w.refuse(
        &landing,
        "invalid_input",
        &["not contained in the integration branch"],
        &before,
    );

    let met = w.finish(
        Some(S_L),
        &code,
        &["--head", &w.head, "--met", "9=r9", "--by", "leader"],
    );
    w.refuse(&met, "invalid_input", &["not an open criterion"], &before);

    let lease = w.finish(
        Some(S_L),
        &held,
        &["--head", &w.head, "--met", "1=r1", "--by", "leader"],
    );
    w.refuse(&lease, "conflict", &["is held by"], &before);

    let missing = w.reports_root().join("demo/missing.md");
    let report = w.finish(
        Some(S_L),
        &research,
        &[
            "--evidence-report",
            missing.to_str().unwrap(),
            "--met",
            "1=report written",
            "--by",
            "leader",
        ],
    );
    w.refuse(
        &report,
        "invalid_input",
        &[
            "not a regular file",
            "does not exist",
            "not found",
            "no such file",
            "cannot read",
        ],
        &before,
    );

    let by = w.finish(
        Some(S_L),
        &code,
        &[
            "--head", &w.head, "--met", "1=r1", "--met", "2=r2", "--by", "  ",
        ],
    );
    w.refuse(&by, "invalid_input", &["--by"], &before);

    let done = w.finish(
        Some(S_L),
        &finished,
        &["--head", &w.head, "--met", "1=r1", "--by", "leader"],
    );
    w.refuse(&done, "conflict", &["`done`"], &before);

    let stale = w.finish(
        Some(S_L),
        &code,
        &[
            "--head",
            &w.head,
            "--met",
            "1=r1",
            "--met",
            "2=r2",
            "--generation",
            "4",
            "--by",
            "leader",
        ],
    );
    w.refuse(&stale, "invalid_input", &["generation"], &before);
}

// ---- AC-8: card next skips a card that already landed --------------------------------------

/// AC-8: a ready card with a recorded landed head inside the tip is listed under skipped as
/// already_landed with its proof and leaves ready; a ready card whose head is outside the tip
/// and a ready card with no landing stay in ready.
#[test]
fn ac8_next_skips_a_ready_card_whose_landed_head_is_inside_the_tip() {
    let w = World::new();
    let landed = w.card_named("task", "apps/a/", &["one"]);
    let plain = w.card_named("task", "apps/b/", &["one"]);
    let outside = w.card_named("task", "apps/c/", &["one"]);
    w.land(&landed, &w.head, true);
    w.land(&outside, &w.orphan, true);

    let listing = w.next(S_B);
    assert_already_landed(&listing, &landed);
    assert_eq!(
        skipped_entry(&listing, &landed).unwrap()["head"],
        json!(w.head),
        "{listing}"
    );
    assert_eq!(
        names(&listing["ready"]),
        vec![plain.clone(), outside.clone()],
        "{listing}"
    );
    assert!(
        skipped_entry(&listing, &outside).is_none(),
        "a head outside the tip is no proof: {listing}"
    );
    assert_eq!(listing["next"], json!(plain), "{listing}");
    assert_eq!(listing["landing"]["checked"], json!(true), "{listing}");
    assert_eq!(listing["landing"]["branch"], json!("integ"), "{listing}");
    assert_eq!(listing["writes"], json!(false), "{listing}");
}

/// AC-8: a ready card with a merged pull request link is skipped as already_landed with its
/// proof; a card with no link stays in ready. The link is made through `card pr` with the stub
/// host, then the card is given back.
#[test]
fn ac8_next_skips_a_ready_card_with_a_merged_pull_request_link() {
    let w = World::new();
    let merged = w.card_named("task", "apps/a/", &["one"]);
    let plain = w.card_named("task", "apps/b/", &["one"]);
    w.link_merged_pr(&merged, 7);

    let listing = w.next(S_B);
    assert_already_landed(&listing, &merged);
    assert_eq!(names(&listing["ready"]), vec![plain], "{listing}");
}

/// AC-8: a `Moai-Spec` trailer of a card's linked SPEC on the tip is a hint, never a proof: the
/// card stays in ready and the landing key names the trailer commit; a card whose SPEC's trailer
/// sits only on a commit outside the tip, and a card whose number appears in a commit message
/// but has no trailer, stay in ready with no hint (a card number in a message is never a proof).
#[test]
fn ac8_next_reports_a_spec_trailer_on_the_tip_as_a_hint_and_never_skips_for_it() {
    let w = World::new();
    let shipped_spec = w.draft_spec("shipped-fix");
    let unmerged_spec = w.draft_spec("unmerged-fix");
    let shipped = w.card_named("task", "apps/a/", &["one"]);
    let unmerged = w.card_named("task", "apps/b/", &["one"]);
    let mentioned = w.card_named("task", "apps/c/", &["one"]);
    result(&w.card(
        Some(S_L),
        &["spec", &shipped, &shipped_spec, "--by", "leader"],
    ));
    result(&w.card(
        Some(S_L),
        &["spec", &unmerged, &unmerged_spec, "--by", "leader"],
    ));
    let shipped_commit = commit_on(
        &w.project,
        "integ",
        &w.tip,
        &[
            "feat: the shipped change",
            &format!("Moai-Spec: {shipped_spec}"),
        ],
    );
    commit_on(
        &w.project,
        "scratch",
        &w.orphan,
        &[
            "feat: not integrated",
            &format!("Moai-Spec: {unmerged_spec}"),
        ],
    );
    let integ_tip = git(&w.project, &["rev-parse", "integ"]);
    commit_on(
        &w.project,
        "integ",
        &integ_tip,
        &[&format!("fix: {mentioned} was mentioned here")],
    );

    let listing = w.next(S_B);
    assert!(
        skipped_entry(&listing, &shipped).is_none(),
        "a trailer is no proof: {listing}"
    );
    assert_eq!(
        names(&listing["ready"]),
        vec![shipped.clone(), unmerged.clone(), mentioned.clone()],
        "{listing}"
    );
    let hints = listing["landing"]["spec_trailer_hints"].as_array().unwrap();
    assert_eq!(hints.len(), 1, "{listing}");
    assert_eq!(hints[0]["name"], json!(shipped), "{listing}");
    assert_eq!(hints[0]["spec"], json!(shipped_spec), "{listing}");
    assert_eq!(hints[0]["commit"], json!(shipped_commit), "{listing}");
    assert_eq!(listing["landing"]["checked"], json!(true), "{listing}");
}

/// AC-8: for a profile that is not m-flow, `landing.checked` is false with a reason and no git
/// proof is used: a card whose recorded head is inside the tip stays in ready, while a merged
/// pull request link, which needs no git, still skips its card.
#[test]
fn ac8_next_says_git_proofs_were_not_checked_for_a_profile_that_is_not_m_flow() {
    let w = World::new();
    let landed = w.card_named("task", "apps/a/", &["one"]);
    let merged = w.card_named("task", "apps/b/", &["one"]);
    let plain = w.card_named("task", "apps/c/", &["one"]);
    w.land(&landed, &w.head, true);
    w.link_merged_pr(&merged, 7);
    w.set_profile(PLAIN_PROFILE);

    let listing = w.next(S_B);
    assert_eq!(listing["landing"]["checked"], json!(false), "{listing}");
    let reason = listing["landing"]["reason"]
        .as_str()
        .or_else(|| listing["landing"]["note"].as_str())
        .unwrap_or("");
    assert!(
        !reason.is_empty(),
        "the landing key says why nothing was checked: {listing}"
    );
    assert_already_landed(&listing, &merged);
    assert_eq!(
        names(&listing["ready"]),
        vec![landed, plain],
        "no git proof is used: {listing}"
    );
}

// ---- AC-9: card list ------------------------------------------------------------------------

/// AC-9: `card list` keeps every existing key and adds `landed`: true for a card with a landing
/// event and for a card with a merged pull request link, false for another card.
#[test]
fn ac9_list_adds_a_landed_key_and_keeps_the_others() {
    let w = World::new();
    let landed = w.card_named("task", "apps/a/", &["one"]);
    let plain = w.card_named("task", "apps/b/", &["one"]);
    let merged = w.card_named("task", "apps/c/", &["one"]);
    w.land(&landed, &w.head, false);
    w.link_merged_pr(&merged, 7);

    let listing = w.list();
    let cards = listing["cards"].as_array().unwrap();
    assert_eq!(cards.len(), 3, "{listing}");
    for entry in cards {
        let mut keys: Vec<&str> = entry
            .as_object()
            .unwrap()
            .keys()
            .map(String::as_str)
            .collect();
        keys.sort_unstable();
        let mut expected: Vec<&str> = LIST_KEYS.to_vec();
        expected.push("landed");
        expected.sort_unstable();
        assert_eq!(keys, expected, "{entry}");
        assert!(entry["landed"].is_boolean(), "{entry}");
    }
    let by_name = |name: &str| {
        cards
            .iter()
            .find(|c| c["name"] == json!(name))
            .unwrap_or_else(|| panic!("{name} is not listed: {listing}"))
    };
    assert_eq!(by_name(&landed)["landed"], json!(true), "{listing}");
    assert_eq!(by_name(&merged)["landed"], json!(true), "{listing}");
    assert_eq!(by_name(&plain)["landed"], json!(false), "{listing}");
    assert_eq!(by_name(&landed)["state"], json!("review"), "{listing}");
}

// ---- AC-10: the scenario of a landed card ---------------------------------------------------

/// The scenario: card A has two criteria; `S_A` takes it, reviews it and records a head inside
/// the tip as landed; then it gives the card back (`release`) or its lease ends. Card B is ready
/// and not landed. A new session `S_B` reads `card next`, finishes A with one met option per
/// criterion and takes B.
fn landed_card_scenario(release: bool) {
    let w = World::new();
    let a = w.card_named("task", "apps/a/", &["one", "two"]);
    let b = w.card_named("task", "apps/b/", &["one"]);
    w.land(&a, &w.head, release);
    if !release {
        w.end_lease(1);
    }

    let listing = w.next(S_B);
    if release {
        assert_already_landed(&listing, &a);
    } else {
        // A still carries its ended reservation; it is not offered either way.
        assert!(
            !names(&listing["ready"]).contains(&a),
            "the landed card is offered again: {listing}"
        );
    }
    assert_eq!(names(&listing["ready"]), vec![b.clone()], "{listing}");
    assert_eq!(listing["next"], json!(b), "{listing}");
    assert_eq!(listing["next_is"], json!("ready"), "{listing}");

    let events_before = w.count("SELECT count(*) FROM event");
    let got = result(&w.finish(
        Some(S_B),
        &a,
        &[
            "--head", &w.head, "--met", "1=r", "--met", "2=r", "--by", "leader",
        ],
    ));
    assert_eq!(got["name"], json!(a), "{got}");
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(got["from"], json!(if release { "ready" } else { "review" }));
    assert_eq!(got["session"], json!(S_B), "{got}");
    if release {
        assert_eq!(got["previous"], Value::Null, "{got}");
        assert_eq!(got["released"], json!(false), "{got}");
    } else {
        assert_eq!(got["previous"]["holder"], json!(S_A), "{got}");
        assert_eq!(got["released"], json!(true), "{got}");
    }
    assert_eq!(
        w.text("SELECT state FROM card WHERE card_id = 1")
            .as_deref(),
        Some("done")
    );
    assert_eq!(
        w.count("SELECT count(*) FROM reservation WHERE card_id = 1 AND status = 'held'"),
        0
    );
    let newest = w.newest_events(2);
    assert_eq!(newest[0].0, "card.finish", "{newest:?}");
    assert_eq!(newest[0].1["by"], json!("leader"), "{newest:?}");
    // The holder's landing event stays; finish may add none for the same head.
    let landed_events =
        w.count("SELECT count(*) FROM event WHERE kind = 'card.landed' AND card_id = 1");
    assert!(
        landed_events >= 1,
        "the landing event of the holder stays: {newest:?}"
    );
    assert!(
        w.count("SELECT count(*) FROM event") > events_before,
        "finish wrote its event"
    );

    // The next real card is the new session's to take.
    let taken = result(&w.card(Some(S_B), &["take", &b]));
    assert_eq!(taken["action"], json!("take"), "{taken}");
    assert_eq!(taken["reservation"]["name"], json!(b), "{taken}");
    let after = w.next(S_B);
    assert_eq!(after["next"], json!(b), "{after}");
    assert_eq!(after["next_is"], json!("mine"), "{after}");
}

/// AC-10: the holder gave the card back after the landing.
#[test]
fn ac10_a_landed_card_given_back_is_skipped_finished_and_the_next_card_is_taken() {
    landed_card_scenario(true);
}

/// AC-10: the holder's lease ended after the landing and the card stayed in review.
#[test]
fn ac10_a_landed_card_with_an_ended_lease_is_finished_and_the_next_card_is_taken() {
    landed_card_scenario(false);
}

// ---- AC-12: help and document ---------------------------------------------------------------

/// AC-12: the help of `card finish` starts with `Writes the ledger.` and names its options;
/// the card operations document names the command, its two evidence kinds and the
/// already_landed reason.
#[test]
fn ac12_the_help_and_the_card_document_name_the_finish_command() {
    let home = TempHome::new();
    let out = home
        .command(&["card", "finish", "--help"])
        .stdin(Stdio::null())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "card finish --help: {} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    let head = text.lines().find(|l| !l.trim().is_empty()).unwrap_or("");
    assert!(
        head.starts_with("Writes the ledger."),
        "card finish --help starts with `{head}`"
    );
    for option in [
        "--head",
        "--met",
        "--evidence-report",
        "--evidence-note",
        "--generation",
        "--by",
    ] {
        assert!(text.contains(option), "the help names {option}: {text}");
    }
    assert!(
        text.contains("open_criteria") && text.contains("Writes the ledger"),
        "the help says it writes the ledger and what a landing-only call answers: {text}"
    );

    let doc = fs::read_to_string(concat!(
        env!("CARGO_MANIFEST_DIR"),
        "/docs/design-review/appendix/operations/card.md"
    ))
    .unwrap();
    let missing: Vec<&str> = [
        "card finish",
        "already_landed",
        "--evidence-report",
        "--evidence-note",
        "--generation",
        "open_criteria",
    ]
    .into_iter()
    .filter(|word| !doc.contains(word))
    .collect();
    assert!(
        missing.is_empty(),
        "the card operations document lacks: {missing:?}"
    );
}

// ---- AC-13: an open criterion without a met option records the landing only -----------------

/// AC-13: a code card with two open criteria and one met option records one landing event and
/// changes nothing else (state, criteria, reservation, generation, rev); the answer says closed
/// false and names the open criterion; a second identical call writes nothing; with no met option
/// at all both criteria are named; with a reservation of the caller the reservation and its
/// generation stay.
#[test]
fn ac13_an_open_criterion_without_a_met_option_records_the_landing_once_and_closes_nothing() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one", "two"]);
    let held = w.card_named("task", "apps/b/", &["one", "two"]);
    let before = w.dump();

    let args = ["--head", &w.head, "--met", "1=r1", "--by", "leader"];
    let got = result(&w.finish(Some(S_L), &card, &args));
    assert_eq!(got["action"], json!("finish"), "{got}");
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(got["open_criteria"], json!([2]), "{got}");
    assert_eq!(got["state"], json!("ready"), "{got}");
    assert_eq!(got["events"], json!(1), "{got}");
    assert_eq!(got["writes"], json!(true), "{got}");
    assert_eq!(got["via"], json!("landing"), "{got}");
    assert_eq!(got["head"], json!(w.head), "{got}");
    assert_eq!(got["tip"], json!(w.tip), "{got}");
    assert_eq!(got["by"], json!("leader"), "{got}");
    let after = w.dump();
    assert_only_events_added(&before, &after, 1);
    let newest = w.newest_events(1);
    assert_eq!(newest[0].0, "card.landed", "{newest:?}");
    assert_eq!(newest[0].1["head"], json!(w.head), "{newest:?}");
    assert_eq!(newest[0].1["tip"], json!(w.tip), "{newest:?}");
    assert_eq!(newest[0].1["generation"], json!(0), "{newest:?}");
    assert_eq!(newest[0].1["by"], json!("leader"), "{newest:?}");
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.finish'"),
        0
    );
    assert_eq!(
        w.text(
            "SELECT group_concat(status, ',') FROM \
             (SELECT status FROM card_ac WHERE card_id = 1 ORDER BY seq)"
        )
        .as_deref(),
        Some("open,open"),
        "the met option is not applied by a landing-only call"
    );

    // The same call again: the landing of this head is recorded already, nothing is written.
    let again = result(&w.finish(Some(S_L), &card, &args));
    assert_eq!(again["closed"], json!(false), "{again}");
    assert_eq!(again["open_criteria"], json!([2]), "{again}");
    assert_eq!(again["events"], json!(0), "{again}");
    assert_eq!(again["writes"], json!(false), "{again}");
    assert_eq!(w.dump(), after, "the second call wrote to the ledger");

    // No met option at all: both open criteria are named.
    let none = result(&w.finish(Some(S_L), &card, &["--head", &w.head, "--by", "leader"]));
    assert_eq!(none["closed"], json!(false), "{none}");
    assert_eq!(none["open_criteria"], json!([1, 2]), "{none}");
    assert_eq!(w.dump(), after, "the call without a met option wrote");

    // The caller's own reservation stays, with its generation.
    result(&w.card(Some(S_L), &["take", &held]));
    let before_held = w.dump();
    let got = result(&w.finish(
        Some(S_L),
        &held,
        &["--head", &w.head, "--met", "2=r2", "--by", "leader"],
    ));
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(got["open_criteria"], json!([1]), "{got}");
    assert_eq!(got["state"], json!("active"), "{got}");
    assert_eq!(got["events"], json!(1), "{got}");
    assert_only_events_added(&before_held, &w.dump(), 1);
    assert_eq!(
        w.count("SELECT count(*) FROM reservation WHERE card_id = 2 AND status = 'held'"),
        1
    );
    assert_eq!(w.newest_events(1)[0].1["generation"], json!(1));
}

/// AC-13: a card with no proven head refuses (no head given, no landing recorded, no merged link;
/// or a head that git does not prove), and a met option that names a number that is not open or
/// names one twice refuses even when the call would only record a landing; every table is
/// unchanged.
#[test]
fn ac13_a_card_with_no_proven_head_or_a_bad_met_option_refuses_and_writes_nothing() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one", "two"]);
    let before = w.dump();

    let no_head = w.finish(Some(S_L), &card, &["--met", "1=r1", "--by", "leader"]);
    w.refuse(
        &no_head,
        "invalid_input",
        &["no landing is recorded"],
        &before,
    );

    let orphan = w.finish(
        Some(S_L),
        &card,
        &["--head", &w.orphan, "--met", "1=r1", "--by", "leader"],
    );
    w.refuse(
        &orphan,
        "invalid_input",
        &["not contained in the integration branch"],
        &before,
    );

    let stranger = w.finish(
        Some(S_L),
        &card,
        &["--head", &w.head, "--met", "3=r3", "--by", "leader"],
    );
    w.refuse(
        &stranger,
        "invalid_input",
        &["not an open criterion"],
        &before,
    );

    let twice = w.finish(
        Some(S_L),
        &card,
        &[
            "--head", &w.head, "--met", "1=r1", "--met", "1=again", "--by", "leader",
        ],
    );
    w.refuse(&twice, "invalid_input", &["twice"], &before);
}

// ---- AC-14: a done card is not finished again; a retry closes it once -------------------------

/// AC-14: a second finish on a done card refuses naming `done`, writes nothing and leaves one
/// landing event and one finish event.
#[test]
fn ac14_a_second_finish_on_a_done_card_refuses_naming_done_and_writes_nothing() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one"]);
    let args = ["--head", &w.head, "--met", "1=r1", "--by", "leader"];
    let got = result(&w.finish(Some(S_L), &card, &args));
    assert_eq!(got["state"], json!("done"), "{got}");

    let before = w.dump();
    let again = w.finish(Some(S_L), &card, &args);
    let message = w.refuse(&again, "conflict", &["`done`"], &before);
    assert!(message.contains("done"), "{message}");
    for (kind, expected) in [("card.finish", 1), ("card.landed", 1), ("card.done", 0)] {
        assert_eq!(
            w.count(&format!(
                "SELECT count(*) FROM event WHERE kind = '{kind}' AND card_id = 1"
            )),
            expected,
            "events of kind {kind}"
        );
    }
}

/// AC-14: a call that refuses before any write (a met option that names a number that is not
/// open) is followed by one correct call that closes the card, with exactly one finish event.
#[test]
fn ac14_a_call_that_failed_before_any_write_is_followed_by_one_call_that_closes_the_card_once() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one"]);
    let before = w.dump();
    let failed = w.finish(
        Some(S_L),
        &card,
        &["--head", &w.head, "--met", "9=r9", "--by", "leader"],
    );
    w.refuse(
        &failed,
        "invalid_input",
        &["not an open criterion"],
        &before,
    );

    let got = result(&w.finish(
        Some(S_L),
        &card,
        &["--head", &w.head, "--met", "1=r1", "--by", "leader"],
    ));
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.finish' AND card_id = 1"),
        1
    );
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.landed' AND card_id = 1"),
        1
    );
}

// ---- AC-15: two callers at once close a card once ---------------------------------------------

/// How many rounds the race tests run, each on a card of its own.
const RACE_ROUNDS: usize = 20;

/// AC-15 (in process): two writers, each with its own connection to the same ledger file, call
/// `workflow::finish` on one ready card at the same moment (a barrier). Exactly one succeeds; the
/// other is refused (the card is done, or the ledger is busy) and is never a silent success; the
/// card has one finish event and no release event.
#[test]
fn ac15_two_writers_on_one_ledger_close_a_card_once() {
    let w = World::new();
    for i in 0..RACE_ROUNDS {
        w.card_named("task", &format!("apps/r{i}/"), &["one"]);
    }
    for round in 0..RACE_ROUNDS {
        let card_id = round as i64 + 1;
        let barrier = Arc::new(Barrier::new(2));
        let handles: Vec<_> = (0..2)
            .map(|_| {
                let barrier = Arc::clone(&barrier);
                let home = w.home.path().to_path_buf();
                let (head, tip) = (w.head.clone(), w.tip.clone());
                std::thread::spawn(move || {
                    let mut own = ledger::open_write(&home, "demo").unwrap();
                    barrier.wait();
                    workflow::finish(
                        &mut own,
                        card_id,
                        Some(S_L),
                        "leader",
                        &FinishEvidence::Landing {
                            head: &head,
                            tip: &tip,
                        },
                        &[(1, "r1".to_string())],
                        None,
                    )
                })
            })
            .collect();
        let results: Vec<_> = handles.into_iter().map(|h| h.join().unwrap()).collect();
        let ok = results.iter().filter(|r| r.is_ok()).count();
        assert_eq!(
            ok, 1,
            "round {round}: exactly one call succeeds: {results:?}"
        );
        for refused in results.iter().filter_map(|r| r.as_ref().err()) {
            assert!(
                matches!(refused, WorkflowError::WrongState { state, .. } if state == "done")
                    || matches!(refused, WorkflowError::Ledger(_)),
                "round {round}: the other call is refused, not a silent success: {refused:?}"
            );
        }
        assert_eq!(
            w.count(&format!(
                "SELECT count(*) FROM event WHERE kind = 'card.finish' AND card_id = {card_id}"
            )),
            1,
            "round {round}: one finish event"
        );
        assert_eq!(
            w.count(&format!(
                "SELECT count(*) FROM event WHERE kind LIKE '%release%' AND card_id = {card_id}"
            )),
            0,
            "round {round}: no release event"
        );
        assert_eq!(
            w.text(&format!("SELECT state FROM card WHERE card_id = {card_id}"))
                .as_deref(),
            Some("done")
        );
    }
}

/// AC-15 (two processes): two `card finish` processes started together on one card give exactly
/// one exit 0 and one refusal, with one finish event and one landing event.
#[test]
fn ac15_two_processes_that_finish_one_card_at_once_give_one_success_and_one_refusal() {
    let w = World::new();
    let cards: Vec<String> = (0..RACE_ROUNDS)
        .map(|i| w.card_named("task", &format!("apps/p{i}/"), &["one"]))
        .collect();
    for (round, card) in cards.iter().enumerate() {
        let card_id = round as i64 + 1;
        let args = ["--head", &w.head, "--met", "1=r1", "--by", "leader"];
        let first = w.spawn_finish(S_L, card, &args);
        let second = w.spawn_finish(S_L, card, &args);
        let outs = [
            first.wait_with_output().unwrap(),
            second.wait_with_output().unwrap(),
        ];
        let codes: Vec<Option<i32>> = outs.iter().map(|o| o.status.code()).collect();
        assert_eq!(
            codes.iter().filter(|c| **c == Some(0)).count(),
            1,
            "round {round}: exactly one process succeeds: {codes:?}"
        );
        let refused = outs.iter().find(|o| o.status.code() != Some(0)).unwrap();
        assert!(refused.status.code().is_some(), "the loser was killed");
        assert_eq!(envelope(refused)["ok"], json!(false), "round {round}");
        assert_eq!(
            w.count(&format!(
                "SELECT count(*) FROM event WHERE kind = 'card.finish' AND card_id = {card_id}"
            )),
            1,
            "round {round}: one finish event"
        );
        assert_eq!(
            w.count(&format!(
                "SELECT count(*) FROM event WHERE kind = 'card.landed' AND card_id = {card_id}"
            )),
            1,
            "round {round}: one landing event"
        );
        assert_eq!(
            w.count(&format!(
                "SELECT count(*) FROM event WHERE kind LIKE '%release%' AND card_id = {card_id}"
            )),
            0,
            "round {round}: no release event"
        );
    }
}

// ---- AC-16: --generation ---------------------------------------------------------------------

/// `base` and `--generation <generation>`.
fn with_generation<'a>(base: &[&'a str], generation: &'a str) -> Vec<&'a str> {
    let mut args = base.to_vec();
    args.extend_from_slice(&["--generation", generation]);
    args
}

/// AC-16: a generation lower or higher than the generation of the card's newest reservation
/// refuses with one line naming both numbers, every table unchanged and no fence_reject event;
/// the current generation passes; a call without the option passes. A card that never had a
/// reservation is at generation 0.
#[test]
fn ac16_a_generation_that_is_not_the_cards_refuses_and_writes_nothing() {
    let w = World::new();
    // Taken and given back once: ready, newest reservation at generation 1 (released).
    let given_back = w.card_named("task", "apps/a/", &["one"]);
    result(&w.card(Some(S_A), &["take", &given_back]));
    result(&w.card(Some(S_A), &["release", &given_back, "--generation", "1"]));
    // Never reserved: generation 0.
    let never = w.card_named("task", "apps/b/", &["one"]);
    let absent = w.card_named("task", "apps/c/", &["one"]);
    let base = ["--head", &w.head, "--met", "1=r1", "--by", "leader"];

    let before = w.dump();
    for (card, given, expected) in [
        (&given_back, "0", "1"),
        (&given_back, "2", "1"),
        (&never, "1", "0"),
    ] {
        let out = w.finish(Some(S_L), card, &with_generation(&base, given));
        let message = w.refuse(&out, "invalid_input", &["generation"], &before);
        assert!(
            message.contains(&format!("generation {given}"))
                && message.contains(&format!("generation {expected}")),
            "the refusal names both numbers: {message}"
        );
        assert_eq!(message.lines().count(), 1, "one line: {message}");
    }

    let got = result(&w.finish(Some(S_L), &given_back, &with_generation(&base, "1")));
    assert_eq!(got["state"], json!("done"), "{got}");
    let got = result(&w.finish(Some(S_L), &never, &with_generation(&base, "0")));
    assert_eq!(got["state"], json!("done"), "{got}");
    let got = result(&w.finish(Some(S_L), &absent, &base));
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'fence_reject'"),
        0
    );
}

// ---- AC-17: release answers landed ------------------------------------------------------------

/// AC-17: releasing a card that has a landing event, and one that has a merged pull request link,
/// answers `landed` true and a `next_step` that names `card finish`; releasing a card with no
/// landing keeps the old answer (no new key); no version control command runs (a stand-in that
/// records its call and fails replaces the real one).
#[test]
fn ac17_releasing_a_landed_card_answers_landed_and_names_the_finish_command() {
    let w = World::new();
    let landed = w.card_named("task", "apps/a/", &["one"]);
    let merged = w.card_named("task", "apps/b/", &["one"]);
    let plain = w.card_named("task", "apps/c/", &["one"]);
    let marker = w.dir.join("vcs-was-called");
    let stub = w.dir.join("vcs-stub");
    common::write_executable(
        &stub,
        &format!("#!/bin/sh\ntouch '{}'\nexit 1\n", marker.display()),
    );
    // `release` as `S_A` with a stand-in that must never run.
    let release = |card: &str| -> Value {
        let out = w
            .command(
                Some(S_A),
                &[
                    "card",
                    "release",
                    card,
                    "--generation",
                    "1",
                    "--project",
                    &w.p,
                ],
            )
            .env("MOAI_GIT_BIN", &stub)
            .output()
            .unwrap();
        result(&out)
    };

    result(&w.card(Some(S_A), &["take", &landed]));
    result(&w.card(Some(S_A), &["review", &landed, "--generation", "1"]));
    result(&w.card(
        Some(S_A),
        &["landed", &landed, "--head", &w.head, "--by", "leader"],
    ));
    result(&w.card(Some(S_A), &["reopen", &landed, "--generation", "1"]));
    let got = release(&landed);
    assert_eq!(got["action"], json!("release"), "{got}");
    assert_eq!(got["state"], json!("ready"), "{got}");
    assert_eq!(got["writes"], json!(true), "{got}");
    assert_eq!(got["landed"], json!(true), "{got}");
    let step = got["next_step"].as_str().unwrap_or("");
    assert!(
        step.contains("card finish") && step.contains(&landed),
        "next_step names the finish command: {got}"
    );

    w.host_merged_pr(7);
    result(&w.card(Some(S_A), &["take", &merged]));
    result(&w.card(
        Some(S_A),
        &[
            "pr",
            &merged,
            "7",
            "--generation",
            "1",
            "--repo",
            "octo/demo",
        ],
    ));
    let got = release(&merged);
    assert_eq!(got["landed"], json!(true), "{got}");
    assert!(
        got["next_step"]
            .as_str()
            .is_some_and(|s| s.contains("card finish")),
        "{got}"
    );

    result(&w.card(Some(S_A), &["take", &plain]));
    let got = release(&plain);
    let mut keys: Vec<&str> = got
        .as_object()
        .unwrap()
        .keys()
        .map(String::as_str)
        .collect();
    keys.sort_unstable();
    assert_eq!(
        keys,
        vec!["action", "project_id", "reservation", "state", "writes"],
        "a card with no landing keeps the old answer: {got}"
    );
    assert!(!marker.exists(), "release ran a version control command");
}

// ---- AC-18 to AC-21: the repairs of the first audit -----------------------------------------

/// AC-18: on an m-flow profile a merged pull request link alone is not evidence, the same as for
/// `card done`: the call is refused with every table unchanged, and the same card closes on a
/// head that git proves. On a profile that is not m-flow the link alone still closes the card.
#[test]
fn ac18_on_an_m_flow_profile_a_merged_link_alone_does_not_close_the_card() {
    let w = World::new();
    let linked = w.card_named("task", "apps/a/", &["one"]);
    w.link_merged_pr(&linked, 7);
    let before = w.dump();
    let out = w.finish(Some(S_L), &linked, &["--met", "1=r1", "--by", "leader"]);
    let message = w.refuse(&out, "invalid_input", &["no landing is recorded"], &before);
    assert!(
        message.contains("merged pull request link alone"),
        "the reason says the link is not enough: {message}"
    );
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind IN ('card.landed', 'card.finish')"),
        0
    );

    let args = ["--head", &w.head, "--met", "1=r1", "--by", "leader"];
    let got = result(&w.finish(Some(S_L), &linked, &args));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["via"], json!("landing"), "{got}");

    let other = w.card_named("task", "apps/b/", &["one"]);
    w.link_merged_pr(&other, 8);
    w.set_profile(PLAIN_PROFILE);
    let got = result(&w.finish(Some(S_L), &other, &["--met", "1=r1", "--by", "leader"]));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["via"], json!("merged_pr"), "{got}");
    assert_eq!(got["events"], json!(1), "only the finish event: {got}");
}

/// AC-19: the SPEC a card names must pass the check at finish; a card with no criterion met and
/// none named refuses; a card with a criterion met earlier closes with no met option.
#[test]
fn ac19_a_failing_spec_and_a_card_with_nothing_met_refuse_and_an_earlier_met_closes() {
    let w = World::new();
    let broken = w.card_named("task", "apps/a/", &["one"]);
    let made = w.run(
        None,
        &[
            "spec",
            "new",
            "broken-fix",
            "--title",
            "t",
            "--minimal",
            "--date",
            "2026-11-21",
            "--project",
            &w.p,
        ],
    );
    let id = result(&made)["id"].as_str().unwrap().to_string();
    let spec_file = w.project.join(".moai/specs").join(&id).join("spec.md");
    fs::write(&spec_file, filled(&fs::read_to_string(&spec_file).unwrap())).unwrap();
    result(&w.card(None, &["spec", &broken, &id, "--by", "leader"]));
    // The link is made while the SPEC passes; the SPEC is damaged afterwards.
    fs::write(&spec_file, "not a SPEC: no front matter, no requirements\n").unwrap();
    let before = w.dump();
    let args = ["--head", &w.head, "--met", "1=r1", "--by", "leader"];
    let out = w.finish(Some(S_L), &broken, &args);
    w.refuse(&out, "invalid_input", &["its SPEC"], &before);

    let withdrawn = w.card_named("task", "apps/b/", &["one"]);
    result(&w.card(Some(S_A), &["take", &withdrawn]));
    result(&w.card(
        Some(S_A),
        &[
            "criterion",
            &withdrawn,
            "1",
            "--generation",
            "1",
            "--withdraw",
            "--reason",
            "not needed",
        ],
    ));
    result(&w.card(Some(S_A), &["release", &withdrawn, "--generation", "1"]));
    let before = w.dump();
    let out = w.finish(
        Some(S_L),
        &withdrawn,
        &["--head", &w.head, "--by", "leader"],
    );
    w.refuse(
        &out,
        "invalid_input",
        &["no acceptance criterion that is met"],
        &before,
    );

    let earlier = w.card_named("task", "apps/c/", &["one"]);
    result(&w.card(Some(S_A), &["take", &earlier]));
    result(&w.card(
        Some(S_A),
        &[
            "criterion",
            &earlier,
            "1",
            "--generation",
            "1",
            "--met",
            "--evidence",
            "r1",
        ],
    ));
    result(&w.card(Some(S_A), &["release", &earlier, "--generation", "1"]));
    let got = result(&w.finish(Some(S_L), &earlier, &["--head", &w.head, "--by", "leader"]));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["state"], json!("done"), "{got}");
}

/// AC-20: a landing-only call is for the leader (`--by leader`) or the holder of the card; a
/// third session is refused with every table unchanged. `card next` still shows a card whose
/// recorded head is not on the integration branch.
#[test]
fn ac20_a_landing_only_call_is_for_the_leader_or_the_holder() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one", "two"]);
    let before = w.dump();
    let args = ["--head", &w.head, "--met", "1=r1", "--by", "agent-b"];
    let out = w.finish(Some(S_B), &card, &args);
    w.refuse(
        &out,
        "invalid_input",
        &["the leader or the holder"],
        &before,
    );
    let listing = w.next(S_B);
    assert!(
        skipped_entry(&listing, &card).is_none(),
        "a refused landing hides nothing: {listing}"
    );

    let by_leader = ["--head", &w.head, "--met", "1=r1", "--by", "leader"];
    let got = result(&w.finish(Some(S_L), &card, &by_leader));
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(got["events"], json!(1), "{got}");

    let held = w.card_named("task", "apps/b/", &["one", "two"]);
    result(&w.card(Some(S_A), &["take", &held]));
    let by_holder = ["--head", &w.head, "--met", "1=r1", "--by", "agent-a"];
    let got = result(&w.finish(Some(S_A), &held, &by_holder));
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(got["events"], json!(1), "{got}");

    let stray = w.card_named("task", "apps/c/", &["one"]);
    w.land(&stray, &w.orphan, true);
    let listing = w.next(S_B);
    assert!(
        skipped_entry(&listing, &stray).is_none(),
        "a head outside the tip is no proof: {listing}"
    );
    assert!(
        names(&listing["ready"]).contains(&stray),
        "the card is still offered: {listing}"
    );
}

/// AC-21: a tag named like the integration branch changes no proof: the branch is read as
/// `refs/heads/<branch>`. A head only on the tag refuses, a head on the branch closes the card,
/// and `card next` proves a landed card against the branch.
#[test]
fn ac21_a_tag_named_like_the_integration_branch_does_not_win() {
    let w = World::new();
    git(&w.project, &["tag", "integ", &w.orphan]);
    let on_tag = w.card_named("task", "apps/a/", &["one"]);
    let before = w.dump();
    let out = w.finish(
        Some(S_L),
        &on_tag,
        &["--head", &w.orphan, "--met", "1=r1", "--by", "leader"],
    );
    w.refuse(
        &out,
        "invalid_input",
        &["not contained in the integration branch"],
        &before,
    );

    let on_branch = w.card_named("task", "apps/b/", &["one"]);
    let got = result(&w.finish(
        Some(S_L),
        &on_branch,
        &["--head", &w.tip, "--met", "1=r1", "--by", "leader"],
    ));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["tip"], json!(w.tip), "{got}");

    let landed = w.card_named("task", "apps/c/", &["one"]);
    w.land(&landed, &w.tip, true);
    let listing = w.next(S_B);
    assert_already_landed(&listing, &landed);
}

/// AC-22: a review card whose lease another session holds closes for the leader on a head that
/// git proves, and the holder's reservation is released in the same call; an active card held by
/// another live session still refuses, and so does a review card whose head git does not prove.
#[test]
fn ac22_a_review_card_with_a_live_lease_closes_on_a_proven_head() {
    let w = World::new();
    let review = w.card_named("task", "apps/a/", &["one"]);
    result(&w.card(Some(S_A), &["take", &review]));
    result(&w.card(Some(S_A), &["review", &review, "--generation", "1"]));
    let active = w.card_named("task", "apps/b/", &["one"]);
    result(&w.card(Some(S_A), &["take", &active]));
    let unproven = w.card_named("task", "apps/c/", &["one"]);
    result(&w.card(Some(S_A), &["take", &unproven]));
    result(&w.card(Some(S_A), &["review", &unproven, "--generation", "1"]));

    let by_leader = |card: &str, head: &str| {
        w.finish(
            Some(S_L),
            card,
            &["--head", head, "--met", "1=r1", "--by", "leader"],
        )
    };
    let before = w.dump();
    let out = by_leader(&active, &w.head);
    w.refuse(&out, "conflict", &["is held by"], &before);
    let out = by_leader(&unproven, &w.orphan);
    w.refuse(
        &out,
        "invalid_input",
        &["not contained in the integration branch"],
        &before,
    );

    let got = result(&by_leader(&review, &w.head));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["from"], json!("review"), "{got}");
    assert_eq!(
        w.text("SELECT status FROM reservation WHERE card_id = 1 ORDER BY generation DESC LIMIT 1")
            .as_deref(),
        Some("released"),
        "the holder's reservation is released in the same call"
    );
    assert_eq!(
        w.text("SELECT state FROM card WHERE card_id = 2")
            .as_deref(),
        Some("active")
    );
}

/// AC-23: one SPEC with two slices: after the first slice its trailer is on the tip, and the
/// card stays offered by `card next`; releasing it does not name `card finish`.
#[test]
fn ac23_a_two_slice_card_stays_offered_after_its_first_slice() {
    let w = World::new();
    let spec = w.draft_spec("two-slices");
    let card = w.card_named("task", "apps/a/", &["slice one", "slice two"]);
    result(&w.card(Some(S_L), &["spec", &card, &spec, "--by", "leader"]));
    commit_on(
        &w.project,
        "integ",
        &w.tip,
        &["feat: slice one", &format!("Moai-Spec: {spec}")],
    );
    let listing = w.next(S_B);
    assert!(
        skipped_entry(&listing, &card).is_none(),
        "the first slice does not land the card: {listing}"
    );
    assert!(names(&listing["ready"]).contains(&card), "{listing}");

    result(&w.card(Some(S_A), &["take", &card]));
    let got = result(&w.card(Some(S_A), &["release", &card, "--generation", "1"]));
    assert!(
        got.get("landed").is_none() && got.get("next_step").is_none(),
        "release names no finish step: {got}"
    );
}

/// AC-24: the SPEC is read from the proven tip. A SPEC that passes on the tip closes the card
/// while the checkout holds a damaged copy; a SPEC that fails on the tip refuses while the
/// checkout holds a good copy.
#[test]
fn ac24_the_spec_is_read_from_the_proven_tip_and_not_from_the_checkout() {
    let w = World::new();
    let good = w.draft_spec("good-on-tip");
    let bad = w.draft_spec("bad-on-tip");
    let good_path = format!(".moai/specs/{good}/spec.md");
    let bad_path = format!(".moai/specs/{bad}/spec.md");
    let good_text = fs::read_to_string(w.project.join(&good_path)).unwrap();
    let broken = "not a SPEC: no front matter, no requirements\n";
    let on_good = w.card_named("task", "apps/a/", &["one"]);
    let on_bad = w.card_named("task", "apps/b/", &["one"]);
    result(&w.card(Some(S_L), &["spec", &on_good, &good, "--by", "leader"]));
    result(&w.card(Some(S_L), &["spec", &on_bad, &bad, "--by", "leader"]));
    // The tip holds the SPEC files as they are now: the first good, the second damaged.
    let step = put_on(&w.project, "integ", &w.tip, &good_path, &good_text);
    let tip = put_on(&w.project, "integ", &step, &bad_path, broken);
    // The checkout holds the opposite copies.
    fs::write(w.project.join(&good_path), broken).unwrap();
    fs::write(w.project.join(&bad_path), good_text.replace(&good, &bad)).unwrap();

    let before = w.dump();
    let args = ["--head", &w.head, "--met", "1=r1", "--by", "leader"];
    let out = w.finish(Some(S_L), &on_bad, &args);
    w.refuse(&out, "invalid_input", &["its SPEC"], &before);
    let got = result(&w.finish(Some(S_L), &on_good, &args));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["tip"], json!(tip), "{got}");
}
