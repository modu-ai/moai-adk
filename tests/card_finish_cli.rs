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
use rusqlite::Connection;
use serde_json::{json, Value};
use std::cell::RefCell;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Output, Stdio};

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

    /// `moai <args>` as `session`, with the stub `gh` and an absolute `git` named.
    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env("MOAI_FORGE_GH_BIN", &self.gh);
        if let Some(git) = &self.git_bin {
            cmd.env("MOAI_GIT_BIN", git);
        }
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
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

/// AC-2: two open criteria need two met options; a duplicate number and a number that is not
/// an open criterion are refused too, each naming its cause, with every table unchanged.
#[test]
fn ac2_each_open_criterion_needs_exactly_one_met_option() {
    let w = World::new();
    let card = w.card_named("task", "apps/a/", &["one", "two"]);
    let before = w.dump();

    let one = w.finish(
        Some(S_L),
        &card,
        &["--head", &w.head, "--met", "1=r1", "--by", "leader"],
    );
    w.refuse(
        &one,
        "invalid_input",
        &["open criterion", "open criteria", "criterion 2"],
        &before,
    );

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

// ac3 gate edges: added with the function in the code step

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
    let research = w.card_named("research", "docs/r/", &["report written"]);
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
    let second = w.card_named("research", "docs/s/", &["report written"]);
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

/// AC-4: an audit card closes with a verdict note of its own (a decision or a report note) and
/// refuses a note that does not exist, a note of kind comment and a note of another card.
#[test]
fn ac4_an_audit_card_closes_with_a_verdict_note_and_refuses_the_others() {
    let w = World::new();
    let audit = w.card_named("audit", "docs/a/", &["verdict written"]);
    let other = w.card_named("audit", "docs/b/", &["verdict written"]);
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
    let third = w.card_named("audit", "docs/c/", &["verdict written"]);
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

// ---- AC-5: the evidence kind follows the card kind ------------------------------------------

/// AC-5: a code card (kind fix) refuses report evidence and note evidence even with a good head;
/// a research card refuses a head even with a good report; both evidence options at once are
/// refused; every refusal writes nothing.
#[test]
fn ac5_a_code_card_refuses_report_evidence_and_a_research_card_refuses_a_head() {
    let w = World::new();
    let code = w.card_named("fix", "apps/a/", &["one"]);
    let research = w.card_named("research", "docs/r/", &["report written"]);
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
/// evidence, `--by`) leaves every table byte equal to its dump before the call and writes no
/// `fence_reject` event. The refusal helper also checks the exit code, the error kind and the
/// reason, so the check cannot pass on a command that does not exist.
#[test]
fn ac7_every_kind_of_refusal_leaves_the_ledger_as_it_was() {
    let w = World::new();
    let code = w.card_named("task", "apps/a/", &["one", "two"]);
    let held = w.card_named("task", "apps/b/", &["one"]);
    let research = w.card_named("research", "docs/r/", &["report written"]);
    result(&w.card(Some(S_A), &["take", &held]));
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
        &["--head", &w.head, "--met", "1=r1", "--by", "leader"],
    );
    w.refuse(
        &met,
        "invalid_input",
        &["open criterion", "open criteria", "criterion 2"],
        &before,
    );

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

/// AC-8: a ready card whose linked SPEC id is the `Moai-Spec` trailer of a commit reachable from
/// the tip is skipped as already_landed; a card whose SPEC's trailer sits only on a commit
/// outside the tip, and a card whose number appears in a commit message but has no trailer,
/// stay in ready (a card number in a message is never a proof).
#[test]
fn ac8_next_skips_a_ready_card_whose_spec_trailer_is_reachable_from_the_tip() {
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
    commit_on(
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
    assert_already_landed(&listing, &shipped);
    assert_eq!(
        names(&listing["ready"]),
        vec![unmerged.clone(), mentioned.clone()],
        "{listing}"
    );
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
    assert_eq!(taken["name"], json!(b), "{taken}");
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
        "--by",
    ] {
        assert!(text.contains(option), "the help names {option}: {text}");
    }

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
    ]
    .into_iter()
    .filter(|word| !doc.contains(word))
    .collect();
    assert!(
        missing.is_empty(),
        "the card operations document lacks: {missing:?}"
    );
}
