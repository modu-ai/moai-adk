//! `moai card next | note | criterion | pr | review | done | renew | check` (card C-53 S1): a
//! lane finds, works and closes a card with no GitHub Issue. The ledger rules are
//! `tests/ledger_workflow.rs`; here the commands, their answers and their refusal kinds.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`). `card pr` reads the
//! pull request through a stub `gh` (`MOAI_FORGE_GH_BIN`) that answers from files and logs every
//! call; no test reaches the network.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::card::work;
use moai_cli::forge::connector::{Call, FakeConnector, HostKind, PrPhase, PrState};
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const MERGE: &str = "0123456789abcdef0123456789abcdef01234567";
const HEAD: &str = "89abcdef0123456789abcdef0123456789abcdef";

struct World {
    home: TempHome,
    /// The stub `gh`, its call log and its answers live outside the HOME.
    _host: tempfile::TempDir,
    dir: PathBuf,
    project: String,
    gh: PathBuf,
    log: PathBuf,
}

/// A HOME with a project folder and a stub `gh`; `registered` also makes the ledger.
fn world(registered: bool) -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    if registered {
        ledger::register(home.path(), "demo").unwrap();
    }
    let host = tempfile::Builder::new()
        .prefix("moai-test-host-")
        .tempdir()
        .unwrap();
    let dir = host.path().to_path_buf();
    let log = dir.join("gh-calls.log");
    let script = format!(
        "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '{log}'\nurl=''\nfor a in \"$@\"; do case \"$a\" in https://*) url=\"$a\";; esac; done\n\
         case \"$url\" in\n\
           */pulls/*) n=${{url##*/pulls/}}; if [ -f '{dir}/pull-'$n'.json' ]; then printf 'HTTP/2.0 200 OK\\r\\n\\r\\n'; cat '{dir}/pull-'$n'.json'; exit 0; fi;;\n\
         esac\n\
         printf 'HTTP/2.0 404 Not Found\\r\\n\\r\\n{{}}'; exit 1\n",
        log = log.display(),
        dir = dir.display()
    );
    let gh = dir.join("gh-stub");
    common::write_executable(&gh, &script);
    World {
        home,
        _host: host,
        dir,
        project: project.to_str().unwrap().to_string(),
        gh,
        log,
    }
}

impl World {
    fn command(&self, session: Option<&str>, args: &[&str]) -> std::process::Command {
        let mut cmd = self.home.command(args);
        cmd.env("MOAI_FORGE_GH_BIN", &self.gh);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd
    }

    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        self.command(session, &all).output().expect("run moai")
    }

    /// The same with `text` on stdin.
    fn card_stdin(&self, session: Option<&str>, args: &[&str], text: &[u8]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut child = self
            .command(session, &all)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .expect("run moai");
        common::feed_stdin(&mut child, text);
        child.wait_with_output().unwrap()
    }

    /// `moai card next <project>`.
    fn next(&self, session: Option<&str>) -> Value {
        result(
            &self
                .command(session, &["card", "next", &self.project])
                .output()
                .expect("run moai"),
        )
    }

    /// What the host says about pull request `number` from now on.
    fn host_pr(&self, number: u64, state: &str, merged: bool) {
        let pull = json!({
            "number": number, "state": state, "merged": merged,
            "merged_at": if merged { json!("2026-11-17T00:00:00Z") } else { Value::Null },
            "merge_commit_sha": if merged { json!(MERGE) } else { Value::Null },
            "head": {"sha": HEAD},
        });
        fs::write(
            self.dir.join(format!("pull-{number}.json")),
            pull.to_string(),
        )
        .unwrap();
    }

    fn gh_calls(&self) -> Vec<String> {
        fs::read_to_string(&self.log)
            .unwrap_or_default()
            .lines()
            .map(str::to_string)
            .collect()
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    /// Adds a card with `scope` and `criteria` criteria, and makes it ready. Returns its name.
    fn ready_card(&self, scope: &str, criteria: usize, priority: &str) -> String {
        let mut args = vec![
            "add",
            "--title",
            "work",
            "--scope",
            scope,
            "--priority",
            priority,
        ];
        for _ in 0..criteria {
            args.extend_from_slice(&["--ac", "it works"]);
        }
        let added = result(&self.card(None, &args));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    /// A ready card taken by `session`; returns `(name, generation)`.
    fn taken(&self, session: &str, scope: &str, criteria: usize) -> (String, String) {
        let name = self.ready_card(scope, criteria, "p2");
        let taken = result(&self.card(Some(session), &["take", &name]));
        (name, taken["reservation"]["generation"].to_string())
    }
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
    assert_eq!(v["operation"], json!("card"), "{v}");
    v["result"].clone()
}

fn refusal(out: &Output) -> (String, String) {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

fn names(list: &Value) -> Vec<String> {
    list.as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap().to_string())
        .collect()
}

#[test]
fn a_lane_finds_takes_works_and_closes_a_card_with_no_issue() {
    let w = world(true);
    let name = w.ready_card("src/a/", 2, "p1");
    assert_eq!(name, "C-1");

    // Find it.
    let next = w.next(Some(S_A));
    assert_eq!(next["action"], json!("next"));
    assert_eq!(next["ledger"], json!("ready"));
    assert_eq!(next["holder"], json!(S_A));
    assert_eq!(next["next"], json!("C-1"));
    assert_eq!(next["next_is"], json!("ready"));
    assert_eq!(next["mine"], json!([]));
    assert_eq!(next["ready"][0]["write_scope"], json!(["src/a/"]));
    assert_eq!(next["writes"], json!(false));

    // Take it; from then on it is this session's card.
    let taken = result(&w.card(
        Some(S_A),
        &["take", "C-1", "--worktree", "/w/a", "--branch", "card/1-a"],
    ));
    assert_eq!(taken["reservation"]["generation"], json!(1));
    let next = w.next(Some(S_A));
    assert_eq!(next["next"], json!("C-1"), "the card it holds comes first");
    assert_eq!(next["next_is"], json!("mine"));
    assert_eq!(next["ready"], json!([]));
    let mine = &next["mine"][0];
    assert_eq!(mine["name"], json!("C-1"));
    assert_eq!(mine["generation"], json!(1));
    assert_eq!(mine["lease_until"], taken["reservation"]["lease_until"]);
    assert_eq!(mine["worktree"], json!("/w/a"));
    assert_eq!(mine["branch"], json!("card/1-a"));
    assert_eq!(mine["state"], json!("active"));
    assert_eq!(mine["lease_live"], json!(true));

    // The plan is a note.
    let note = result(&w.card_stdin(
        Some(S_A),
        &["note", "C-1", "--kind", "report"],
        b"PLAN C-1\n1. write the test\n",
    ));
    assert_eq!(note["action"], json!("note"));
    assert_eq!(note["note_id"], json!(1));
    assert_eq!(note["author"], json!(S_A));
    assert_eq!(note["kind"], json!("report"));
    assert!(note["ref"].as_str().unwrap().starts_with("note:1:"));
    assert_eq!(note["events"], json!(1));

    // The criteria.
    let met = result(&w.card(
        Some(S_A),
        &[
            "criterion",
            "C-1",
            "1",
            "--generation",
            "1",
            "--met",
            "--evidence",
            "cargo test --test x: 3 passed",
        ],
    ));
    assert_eq!(met["action"], json!("criterion"));
    assert_eq!(met["seq"], json!(1));
    assert_eq!(met["from"], json!("open"));
    assert_eq!(met["status"], json!("met"));
    assert_eq!(met["evidence_ref"], json!("cargo test --test x: 3 passed"));
    let withdrawn = result(&w.card(
        Some(S_A),
        &[
            "criterion",
            "C-1",
            "2",
            "--generation",
            "1",
            "--withdraw",
            "--reason",
            "moved to C-2",
        ],
    ));
    assert_eq!(withdrawn["status"], json!("withdrawn"));
    assert_eq!(withdrawn["withdrawn_reason"], json!("moved to C-2"));

    // The pull request, open at first.
    w.host_pr(7, "open", false);
    let pr = result(&w.card(
        Some(S_A),
        &["pr", "C-1", "7", "--generation", "1", "--repo", "Octo/Demo"],
    ));
    assert_eq!(pr["action"], json!("pr"));
    assert_eq!(pr["to"], json!("github:octo/demo#7"));
    assert_eq!(pr["pr_state"], json!("open"));
    assert_eq!(pr["merge_ref"], Value::Null);
    assert_eq!(pr["head"], json!(HEAD));
    assert_eq!(pr["created"], json!(true));

    // Review, and a close that is too early.
    let review = result(&w.card(Some(S_A), &["review", "C-1", "--generation", "1"]));
    assert_eq!(review["from"], json!("active"));
    assert_eq!(review["state"], json!("review"));
    assert_eq!(review["reservation"]["status"], json!("held"));
    let (kind, message) = refusal(&w.card(Some(S_A), &["done", "C-1", "--generation", "1"]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert_eq!(
        message,
        "card done refused: cannot close: no merged pull request is linked"
    );

    // Merged on the host: read it again, then close.
    w.host_pr(7, "closed", true);
    let pr = result(&w.card(
        Some(S_A),
        &["pr", "C-1", "7", "--generation", "1", "--repo", "octo/demo"],
    ));
    assert_eq!(pr["pr_state"], json!("merged"));
    assert_eq!(pr["merge_ref"], json!(MERGE));
    assert_eq!(pr["created"], json!(false));
    let done = result(&w.card(Some(S_A), &["done", "C-1", "--generation", "1"]));
    assert_eq!(done["action"], json!("done"));
    assert_eq!(done["from"], json!("review"));
    assert_eq!(done["state"], json!("done"));
    assert!(done["closed_at"].is_i64());
    assert_eq!(done["reservation"]["status"], json!("released"));

    let show = result(&w.card(None, &["show", "C-1"]));
    assert_eq!(show["state"], json!("done"));
    assert_eq!(show["reservation"], Value::Null);
    assert_eq!(show["links"][0]["pr_state"], json!("merged"));
    assert_eq!(show["links"][0]["to"], json!("github:octo/demo#7"));
    assert_eq!(
        show["notes"][0]["body"],
        json!("PLAN C-1\n1. write the test")
    );
    let next = w.next(Some(S_A));
    assert_eq!(next["mine"], json!([]));
    assert_eq!(next["ready"], json!([]));
    assert_eq!(next["skipped"], json!([]));
    assert_eq!(next["next"], Value::Null);
    assert_eq!(next["next_is"], Value::Null);

    // The host was asked twice, both times for the pull request and nothing else.
    let calls = w.gh_calls();
    assert_eq!(calls.len(), 2, "{calls:?}");
    for call in &calls {
        assert!(
            call.contains("https://api.github.com/repos/octo/demo/pulls/7"),
            "{call}"
        );
        assert!(!call.contains("POST") && !call.contains("PATCH"), "{call}");
    }
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind IN ('card.note','card.criterion','card.pr','card.review','card.done')"),
        7
    );
}

#[test]
fn next_orders_by_priority_then_number_and_says_why_a_card_is_skipped() {
    let w = world(true);
    let low = w.ready_card("low/", 0, "p3");
    let urgent = w.ready_card("urgent/", 0, "p0");
    let held = w.ready_card("held/", 0, "p0");
    let clash = w.ready_card("held/deep/x.rs", 0, "p1");
    let also = w.ready_card("also/", 0, "p0");
    // A card without a scope: `take` refuses it, so `next` does not offer it.
    let bare = result(&w.card(None, &["add", "--title", "bare"]))["name"]
        .as_str()
        .unwrap()
        .to_string();
    result(&w.card(None, &["ready", &bare]));
    result(&w.card(Some(S_B), &["take", &held]));

    let before = w.home.snapshot();
    let next = w.next(Some(S_A));
    assert_eq!(w.home.snapshot(), before, "next writes nothing");
    assert_eq!(names(&next["ready"]), [urgent.clone(), also, low]);
    assert_eq!(next["next"], json!(urgent));
    assert_eq!(names(&next["skipped"]), [held.clone(), clash, bare.clone()]);
    let skipped = next["skipped"].as_array().unwrap();
    assert_eq!(skipped[0]["reason"], json!("held"));
    assert_eq!(skipped[0]["holder"], json!(S_B));
    assert_eq!(skipped[0]["generation"], json!(1));
    assert!(skipped[0]["lease_until"].is_i64());
    assert_eq!(skipped[1]["reason"], json!("scope_overlap"));
    assert_eq!(skipped[1]["other"], json!(held));
    assert_eq!(skipped[1]["mine"], json!("held/deep/x.rs"));
    assert_eq!(skipped[1]["theirs"], json!("held/"));
    assert_eq!(skipped[2]["reason"], json!("empty_scope"));
    let (kind, message) = refusal(&w.card(Some(S_A), &["take", &bare]));
    assert_eq!(kind, "invalid_input", "take agrees: {message}");

    // A caller without a session id holds nothing and still sees the queue.
    let anonymous = w.next(None);
    assert_eq!(anonymous["holder"], Value::Null);
    assert_eq!(anonymous["mine"], json!([]));
    assert_eq!(anonymous["ready"], next["ready"]);
}

#[test]
fn next_no_longer_offers_a_card_after_a_second_session_took_it() {
    let w = world(true);
    let card = w.ready_card("src/a/", 0, "p2");
    assert_eq!(w.next(Some(S_A))["next"], json!(card));
    assert_eq!(w.next(Some(S_B))["next"], json!(card));

    result(&w.card(Some(S_B), &["take", &card]));
    let a = w.next(Some(S_A));
    assert_eq!(a["next"], Value::Null);
    assert_eq!(a["ready"], json!([]));
    assert_eq!(a["mine"], json!([]));
    assert_eq!(a["skipped"][0]["name"], json!(card));
    assert_eq!(a["skipped"][0]["reason"], json!("held"));
    assert_eq!(a["skipped"][0]["holder"], json!(S_B));
    let b = w.next(Some(S_B));
    assert_eq!(b["next"], json!(card));
    assert_eq!(b["next_is"], json!("mine"));
    assert_eq!(b["mine"][0]["name"], json!(card));
    assert_eq!(b["skipped"], json!([]));
    // And `take` agrees with what `next` said.
    let (kind, _) = refusal(&w.card(Some(S_A), &["take", &card]));
    assert_eq!(kind, "conflict");
}

#[test]
fn without_a_ledger_next_has_no_cards_and_the_other_actions_find_nothing_to_write() {
    let w = world(false);
    let before = w.home.snapshot();
    let next = w.next(Some(S_A));
    assert_eq!(next["ledger"], json!("absent"));
    assert_eq!(next["next"], Value::Null);
    assert_eq!(next["mine"], json!([]));
    assert_eq!(next["ready"], json!([]));
    assert_eq!(next["skipped"], json!([]));
    assert_eq!(next["writes"], json!(false));

    w.host_pr(7, "closed", true);
    let g = ["--generation", "1"];
    let actions: Vec<Vec<&str>> = vec![
        vec![
            "criterion",
            "C-1",
            "1",
            g[0],
            g[1],
            "--met",
            "--evidence",
            "e",
        ],
        vec!["pr", "C-1", "7", g[0], g[1], "--repo", "octo/demo"],
        vec!["review", "C-1", g[0], g[1]],
        vec!["done", "C-1", g[0], g[1]],
        vec!["renew", "C-1", g[0], g[1]],
        vec!["check", "C-1", g[0], g[1], "--path", "src/a.rs"],
    ];
    for args in &actions {
        let (kind, message) = refusal(&w.card(Some(S_A), args));
        assert_eq!(kind, "not_found", "{args:?}: {message}");
        assert!(message.contains("ledger-register-apply"), "{message}");
    }
    let (kind, message) =
        refusal(&w.card_stdin(Some(S_A), &["note", "C-1", "--kind", "comment"], b"text"));
    assert_eq!(kind, "not_found", "{message}");
    assert_eq!(w.home.snapshot(), before, "nothing was created");
    assert!(w.gh_calls().is_empty(), "no ledger, no call to the host");
}

#[test]
fn a_stale_generation_and_another_session_are_refused_by_every_writer() {
    let w = world(true);
    let (card, g1) = w.taken(S_A, "src/a/", 1);
    w.host_pr(7, "closed", true);
    let writers = |g: &str| -> Vec<Vec<String>> {
        [
            vec![
                "criterion",
                &card,
                "1",
                "--generation",
                g,
                "--met",
                "--evidence",
                "e",
            ],
            vec!["pr", &card, "7", "--generation", g, "--repo", "octo/demo"],
            vec!["review", &card, "--generation", g],
            vec!["done", &card, "--generation", g],
            vec!["renew", &card, "--generation", g],
        ]
        .iter()
        .map(|a| a.iter().map(|s| s.to_string()).collect())
        .collect()
    };

    // Another session, with the right generation.
    for args in writers(&g1) {
        let args: Vec<&str> = args.iter().map(String::as_str).collect();
        let (kind, message) = refusal(&w.card(Some(S_B), &args));
        assert_eq!(kind, "conflict", "{args:?}: {message}");
        assert!(
            message.contains(S_A),
            "{args:?} names the holder: {message}"
        );
    }
    // The holder hands the card over: its own generation is stale from then on.
    let handed = result(&w.card(
        Some(S_A),
        &["handoff", &card, "--generation", &g1, "--to", S_B],
    ));
    let g2 = handed["reservation"]["generation"].to_string();
    for args in writers(&g1) {
        let args: Vec<&str> = args.iter().map(String::as_str).collect();
        let (kind, message) = refusal(&w.card(Some(S_A), &args));
        assert_eq!(kind, "conflict", "{args:?}: {message}");
        assert!(message.contains("is stale"), "{args:?}: {message}");
    }
    // A session without an id is nobody.
    for args in writers(&g2) {
        let args: Vec<&str> = args.iter().map(String::as_str).collect();
        let (kind, message) = refusal(&w.card(None, &args));
        assert_eq!(kind, "invalid_input", "{args:?}: {message}");
        assert!(
            message.contains("this session has no id"),
            "{args:?}: {message}"
        );
    }

    assert!(
        w.gh_calls().is_empty(),
        "a refused caller never reaches the host"
    );
    assert_eq!(w.count("SELECT count(*) FROM card_link"), 0);
    assert_eq!(
        w.count("SELECT count(*) FROM card_ac WHERE status <> 'open'"),
        0
    );
    assert_eq!(
        w.count("SELECT count(*) FROM card WHERE state <> 'active'"),
        0
    );
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'fence_reject'"),
        10,
        "five writers refused twice each with a record; no record for a caller without an id"
    );

    // The new holder at the new generation does all of it.
    for args in writers(&g2) {
        let args: Vec<&str> = args.iter().map(String::as_str).collect();
        if args[0] == "done" {
            continue;
        }
        result(&w.card(Some(S_B), &args));
    }
    result(&w.card(Some(S_B), &["done", &card, "--generation", &g2]));
    assert_eq!(w.gh_calls().len(), 1);
}

#[test]
fn notes_are_text_from_stdin_kept_in_order_and_named_once() {
    let w = world(true);
    let (card, _) = w.taken(S_A, "src/a/", 0);
    let args = ["note", card.as_str(), "--kind", "observation"];

    // The holder, another session and a person without a session can all write a note.
    let first = result(&w.card_stdin(
        Some(S_A),
        &args,
        b"LANE-BLOCKED kind=decision which way?\n\n",
    ));
    let answer = result(&w.card_stdin(
        Some(S_B),
        &[
            "note", &card, "--kind", "decision", "--ref", "answer-1", "--by", "leader",
        ],
        "\u{c655}\u{b3c4}: the first way".as_bytes(),
    ));
    let plain = result(&w.card_stdin(None, &args, b"seen"));
    assert_eq!(first["author"], json!(S_A));
    assert_eq!(answer["author"], json!("leader"));
    assert_eq!(answer["ref"], json!("answer-1"));
    assert_eq!(plain["author"], json!("user"));
    assert_ne!(first["ref"], plain["ref"]);

    let show = result(&w.card(None, &["show", &card]));
    let notes = show["notes"].as_array().unwrap();
    let bodies: Vec<&str> = notes.iter().map(|n| n["body"].as_str().unwrap()).collect();
    assert_eq!(
        bodies,
        [
            "LANE-BLOCKED kind=decision which way?",
            "\u{c655}\u{b3c4}: the first way",
            "seen"
        ],
        "in the order written, without the trailing line breaks of the pipe"
    );
    assert_eq!(notes[1]["kind"], json!("decision"));
    assert_eq!(notes[0]["origin"], json!("local"));

    // Refused, each leaving the three notes as they are.
    let (kind, message) = refusal(&w.card_stdin(
        Some(S_A),
        &["note", &card, "--kind", "comment", "--ref", "answer-1"],
        b"again",
    ));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("answer-1"), "{message}");
    for (args, text) in [
        (vec!["note", card.as_str(), "--kind", "comment"], &b""[..]),
        (
            vec!["note", card.as_str(), "--kind", "comment"],
            &b" \n\n"[..],
        ),
        (
            vec!["note", card.as_str(), "--kind", "comment"],
            &b"\xff\xfe"[..],
        ),
        (vec!["note", card.as_str(), "--kind", "plan"], &b"text"[..]),
        (vec!["note", "#7", "--kind", "comment"], &b"text"[..]),
    ] {
        let (kind, message) = refusal(&w.card_stdin(Some(S_A), &args, text));
        assert_eq!(kind, "invalid_input", "{args:?}: {message}");
        assert!(message.starts_with("card note refused: "), "{message}");
    }
    let (kind, _) =
        refusal(&w.card_stdin(Some(S_A), &["note", "C-99", "--kind", "comment"], b"text"));
    assert_eq!(kind, "not_found");
    assert_eq!(w.count("SELECT count(*) FROM card_note"), 3);
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.note'"),
        3
    );
}

#[test]
fn a_note_larger_than_one_mebibyte_is_refused_and_one_of_exactly_that_size_is_kept() {
    let w = world(true);
    let card = w.ready_card("src/a/", 0, "p2");
    let args = ["note", card.as_str(), "--kind", "report"];
    let edge = vec![b'a'; 1_048_576];
    assert_eq!(
        result(&w.card_stdin(None, &args, &edge))["bytes"],
        json!(1_048_576)
    );
    let over = vec![b'a'; 1_048_577];
    let (kind, message) = refusal(&w.card_stdin(None, &args, &over));
    assert_eq!(kind, "invalid_input", "{message}");
    assert_eq!(w.count("SELECT count(*) FROM card_note"), 1);
}

#[test]
fn show_prints_the_reservation_that_holds_the_card() {
    let w = world(true);
    let card = w.ready_card("src/a/", 0, "p2");
    assert_eq!(
        result(&w.card(None, &["show", &card]))["reservation"],
        Value::Null
    );
    let taken = result(&w.card(
        Some(S_A),
        &[
            "take",
            &card,
            "--scope",
            "src/a/x.rs",
            "--branch",
            "card/1-a",
        ],
    ));
    let show = result(&w.card(None, &["show", &card]));
    assert_eq!(show["reservation"], taken["reservation"]);
    assert_eq!(show["reservation"]["holder"], json!(S_A));
    assert_eq!(show["reservation"]["generation"], json!(1));
    assert_eq!(show["reservation"]["write_scope"], json!(["src/a/x.rs"]));
    assert!(show["reservation"]["lease_until"].is_i64());
    assert_eq!(show["writes"], json!(false));
    result(&w.card(Some(S_A), &["release", &card, "--generation", "1"]));
    assert_eq!(
        result(&w.card(None, &["show", &card]))["reservation"],
        Value::Null
    );
}

#[test]
fn renew_moves_the_lease_end_for_the_holder_within_the_lease_limits() {
    let w = world(true);
    let card = w.ready_card("src/a/", 0, "p2");
    let taken = result(&w.card(Some(S_A), &["take", &card, "--lease-secs", "60"]));
    let first = taken["reservation"]["lease_until"].as_i64().unwrap();
    let renewed = result(&w.card(
        Some(S_A),
        &["renew", &card, "--generation", "1", "--lease-secs", "86400"],
    ));
    assert_eq!(renewed["action"], json!("renew"));
    assert_eq!(
        renewed["reservation"]["generation"],
        json!(1),
        "the same generation"
    );
    let second = renewed["reservation"]["lease_until"].as_i64().unwrap();
    assert!(second >= first + 86_400 - 60, "{first} -> {second}");
    for secs in ["59", "86401", "0"] {
        let (kind, message) = refusal(&w.card(
            Some(S_A),
            &["renew", &card, "--generation", "1", "--lease-secs", secs],
        ));
        assert_eq!(kind, "invalid_input", "{secs}: {message}");
    }
    result(&w.card(
        Some(S_A),
        &["renew", &card, "--generation", "1", "--lease-secs", "60"],
    ));
    assert_eq!(
        result(&w.card(None, &["show", &card]))["reservation"]["lease_until"],
        json!(second),
        "a shorter renewal never moves the end back"
    );
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'renew'"),
        2
    );
}

#[test]
fn check_answers_whether_this_session_may_write_a_path() {
    let w = world(true);
    let card = w.ready_card("src/card/", 0, "p2");
    result(&w.card(
        Some(S_A),
        &[
            "take",
            &card,
            "--scope",
            "src/card/",
            "--scope",
            "tests/card_*",
        ],
    ));
    let before = w.home.snapshot();

    let ok = result(&w.card(
        Some(S_A),
        &[
            "check",
            &card,
            "--generation",
            "1",
            "--path",
            "src/card/work.rs",
            "--path",
            "tests/card_work_cli.rs",
        ],
    ));
    assert_eq!(ok["action"], json!("check"));
    assert_eq!(ok["allowed"], json!(true));
    assert_eq!(ok["holder"], json!(S_A));
    assert_eq!(
        ok["paths"],
        json!([
            {"path": "src/card/work.rs", "pattern": "src/card/"},
            {"path": "tests/card_work_cli.rs", "pattern": "tests/card_*"},
        ])
    );
    assert_eq!(ok["writes"], json!(false));

    // Outside the reservation: one path is enough to refuse.
    for path in [
        "src/ledger/workflow.rs",
        "src",
        "src/card/../main.rs",
        "../x",
    ] {
        let (kind, message) = refusal(&w.card(
            Some(S_A),
            &[
                "check",
                &card,
                "--generation",
                "1",
                "--path",
                "src/card/a.rs",
                "--path",
                path,
            ],
        ));
        assert_eq!(kind, "conflict", "{path}: {message}");
        assert!(message.contains("conflict in write scope"), "{message}");
    }
    for path in [
        "",
        "/etc/passwd",
        "a\nb",
        "src\\card\\work.rs",
        " src/card/work.rs",
        "src/card/work.rs ",
        "C:\\x",
    ] {
        let (kind, message) = refusal(&w.card(
            Some(S_A),
            &["check", &card, "--generation", "1", "--path", path],
        ));
        assert_eq!(kind, "invalid_input", "{path:?}: {message}");
    }
    // Another session, a stale or unknown generation, a card that is not held.
    let (kind, message) = refusal(&w.card(
        Some(S_B),
        &[
            "check",
            &card,
            "--generation",
            "1",
            "--path",
            "src/card/work.rs",
        ],
    ));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains(S_A), "{message}");
    let (kind, _) = refusal(&w.card(
        Some(S_A),
        &[
            "check",
            &card,
            "--generation",
            "2",
            "--path",
            "src/card/work.rs",
        ],
    ));
    assert_eq!(kind, "conflict");
    assert_eq!(w.home.snapshot(), before, "check reads the ledger only");

    result(&w.card(Some(S_A), &["release", &card, "--generation", "1"]));
    let (kind, message) = refusal(&w.card(
        Some(S_A),
        &[
            "check",
            &card,
            "--generation",
            "1",
            "--path",
            "src/card/work.rs",
        ],
    ));
    assert_eq!(kind, "conflict", "{message}");
}

#[test]
fn done_is_refused_at_every_edge_until_the_card_can_close() {
    let w = world(true);
    let done = |card: &str, g: &str| w.card(Some(S_A), &["done", card, "--generation", g]);

    // No criterion at all.
    let (bare, g) = w.taken(S_A, "bare/", 0);
    let (kind, message) = refusal(&done(&bare, &g));
    assert_eq!(kind, "conflict", "still active: {message}");
    assert!(
        message.contains("`active`") && message.contains("`review`"),
        "{message}"
    );
    result(&w.card(Some(S_A), &["review", &bare, "--generation", &g]));
    let (kind, message) = refusal(&done(&bare, &g));
    assert_eq!(kind, "invalid_input");
    assert!(
        message.ends_with("the card has no acceptance criterion that is met"),
        "{message}"
    );

    // Two criteria.
    let (card, g) = w.taken(S_A, "src/a/", 2);
    result(&w.card(Some(S_A), &["review", &card, "--generation", &g]));
    let met = |seq: &str| {
        result(&w.card(
            Some(S_A),
            &[
                "criterion",
                &card,
                seq,
                "--generation",
                &g,
                "--met",
                "--evidence",
                "e",
            ],
        ))
    };
    met("1");
    let (_, message) = refusal(&done(&card, &g));
    assert!(
        message.ends_with("an acceptance criterion is still open"),
        "{message}"
    );
    met("2");
    let (_, message) = refusal(&done(&card, &g));
    assert!(
        message.ends_with("no merged pull request is linked"),
        "{message}"
    );
    w.host_pr(9, "open", false);
    let link = || {
        result(&w.card(
            Some(S_A),
            &["pr", &card, "9", "--generation", &g, "--repo", "octo/demo"],
        ))
    };
    assert_eq!(link()["pr_state"], json!("open"));
    let (_, message) = refusal(&done(&card, &g));
    assert!(
        message.ends_with("no merged pull request is linked"),
        "{message}"
    );
    w.host_pr(9, "closed", false);
    assert_eq!(link()["pr_state"], json!("closed"));
    let (_, message) = refusal(&done(&card, &g));
    assert!(
        message.ends_with("no merged pull request is linked"),
        "{message}"
    );
    assert_eq!(w.count("SELECT count(*) FROM card WHERE state = 'done'"), 0);
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.done'"),
        0
    );

    w.host_pr(9, "closed", true);
    assert_eq!(link()["pr_state"], json!("merged"));
    assert_eq!(result(&done(&card, &g))["state"], json!("done"));
    assert_eq!(w.count("SELECT count(*) FROM card WHERE state = 'done'"), 1);
    assert_eq!(
        w.count("SELECT count(*) FROM reservation WHERE card_id = 2 AND status = 'released'"),
        1
    );
}

#[test]
fn a_pull_request_needs_a_project_and_a_host_that_answers() {
    let w = world(true);
    let (card, g) = w.taken(S_A, "src/a/", 0);
    let pr = |extra: &[&str]| {
        let mut args = vec!["pr", card.as_str(), "7", "--generation", g.as_str()];
        args.extend_from_slice(extra);
        w.card(Some(S_A), &args)
    };

    // No project named and the card did not come from one.
    let (kind, message) = refusal(&pr(&[]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("--repo"), "{message}");
    let (kind, message) = refusal(&pr(&["--repo", "octo"]));
    assert_eq!(kind, "invalid_input", "{message}");
    let (kind, message) = refusal(&pr(&["--repo", "octo/demo", "--from", "gitlab"]));
    assert_ne!(kind, "internal", "{message}");
    assert!(
        message.contains("GitLab") || message.contains("gitlab"),
        "{message}"
    );
    let (kind, message) = refusal(&w.card(
        Some(S_A),
        &["pr", &card, "0", "--generation", &g, "--repo", "octo/demo"],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(w.gh_calls().is_empty(), "nothing was asked yet");

    // The host does not know the pull request: nothing is linked.
    let out = pr(&["--repo", "octo/demo"]);
    assert_eq!(out.status.code(), Some(1));
    assert_eq!(envelope(&out)["ok"], json!(false));
    assert!(!w.gh_calls().is_empty());
    assert_eq!(w.count("SELECT count(*) FROM card_link"), 0);
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.pr'"),
        0
    );
}

#[test]
fn the_pull_request_state_comes_from_the_connector_and_nowhere_else() {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    let mut l = ledger::open_write(home.path(), "demo").unwrap();
    l.with_write_tx(|tx, now| {
        tx.execute(
            "INSERT INTO card (title, state, write_scope, created_by, created_at, updated_at, \
                               source_host, source_project, source_number) \
             VALUES ('t', 'ready', '[\"src/a/\"]', 'test', ?1, ?1, 'github', 'octo/demo', 3)",
            [now],
        )?;
        Ok(())
    })
    .unwrap();
    ledger::reservation::reserve(&mut l, 1, S_A, &["src/a/".to_string()], 3600, None, None)
        .unwrap();
    drop(l);

    let fake = FakeConnector::new(HostKind::Github).with_pr(PrState {
        number: 12,
        phase: PrPhase::Merged,
        head: HEAD.to_string(),
        merge_commit: Some(MERGE.to_string()),
    });
    let request = |holder: &'static str, number: u64| work::PrRequest {
        holder,
        generation: 1,
        number,
        from: None,
        repo: None,
    };
    // Another session: refused before the connector is asked.
    let refused = work::pr(&project, home.path(), "C-1", &request(S_B, 12), &fake).unwrap_err();
    assert!(refused.contains("conflict in"), "{refused}");
    assert!(fake.calls().is_empty());

    // The holder: the card's own project is used when none is named.
    let linked = work::pr(&project, home.path(), "C-1", &request(S_A, 12), &fake).unwrap();
    assert_eq!(linked["to"], json!("github:octo/demo#12"));
    assert_eq!(linked["pr_state"], json!("merged"));
    assert_eq!(linked["merge_ref"], json!(MERGE));
    assert_eq!(
        fake.calls(),
        [Call::PrState {
            project: "octo/demo".to_string(),
            number: 12
        }]
    );
    // A pull request the connector does not know leaves no link.
    assert!(work::pr(&project, home.path(), "C-1", &request(S_A, 13), &fake).is_err());
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    let links: i64 = l
        .conn()
        .query_row("SELECT count(*) FROM card_link", [], |r| r.get(0))
        .unwrap();
    assert_eq!(links, 1);
}

#[test]
fn a_merged_pull_request_without_a_merge_commit_is_not_recorded_as_merged() {
    let w = world(true);
    let (card, g) = w.taken(S_A, "src/a/", 0);
    let pull = json!({"number": 7, "state": "closed", "merged": true, "head": {"sha": HEAD}});
    fs::write(w.dir.join("pull-7.json"), pull.to_string()).unwrap();
    let out = w.card(
        Some(S_A),
        &["pr", &card, "7", "--generation", &g, "--repo", "octo/demo"],
    );
    let (kind, message) = refusal(&out);
    assert_ne!(kind, "internal", "{message}");
    assert!(message.contains("merge commit"), "{message}");
    assert_eq!(w.count("SELECT count(*) FROM card_link"), 0);
}

#[test]
fn criterion_takes_exactly_one_of_met_with_evidence_or_withdraw_with_a_reason() {
    let w = world(true);
    let (card, g) = w.taken(S_A, "src/a/", 1);
    let base = ["criterion", card.as_str(), "1", "--generation", g.as_str()];
    for extra in [
        vec![],
        vec!["--met"],
        vec!["--withdraw"],
        vec!["--evidence", "e"],
        vec!["--met", "--reason", "r"],
        vec!["--withdraw", "--evidence", "e"],
        vec!["--met", "--evidence", "e", "--withdraw", "--reason", "r"],
    ] {
        let mut args = base.to_vec();
        args.extend_from_slice(&extra);
        let out = w.card(Some(S_A), &args);
        assert_ne!(out.status.code(), Some(0), "{extra:?}");
    }
    let (kind, message) = refusal(&w.card(
        Some(S_A),
        &[
            "criterion",
            &card,
            "2",
            "--generation",
            &g,
            "--met",
            "--evidence",
            "e",
        ],
    ));
    assert_eq!(kind, "not_found", "{message}");
    assert_eq!(
        w.count("SELECT count(*) FROM card_ac WHERE status <> 'open'"),
        0
    );
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.criterion'"),
        0
    );
}

#[test]
fn every_new_action_says_first_what_it_does_to_the_ledger() {
    let home = TempHome::new();
    for (action, first) in [
        ("next", "Reads the ledger only."),
        ("check", "Reads the ledger only."),
        ("note", "Writes the ledger."),
        ("criterion", "Writes the ledger."),
        ("pr", "Writes the ledger."),
        ("review", "Writes the ledger."),
        ("done", "Writes the ledger."),
        ("renew", "Writes the ledger."),
        ("reopen", "Writes the ledger."),
    ] {
        let out = home.command(&["card", action, "--help"]).output().unwrap();
        assert!(out.status.success(), "card {action} --help");
        let text = String::from_utf8_lossy(&out.stdout).into_owned();
        let head = text.lines().find(|l| !l.trim().is_empty()).unwrap_or("");
        assert!(
            head.starts_with(first),
            "card {action} --help starts with `{head}`"
        );
    }
    let out = home.command(&["card", "--help"]).output().unwrap();
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    for action in [
        "next",
        "note",
        "criterion",
        "pr",
        "review",
        "done",
        "renew",
        "check",
    ] {
        assert!(
            text.lines()
                .any(|l| l.trim_start().starts_with(&format!("{action} "))),
            "card --help lists {action}"
        );
    }
}

// ---- repair of the first audit --------------------------------------------------------------------

/// A token-shaped word built at run time, so no literal in this file has the shape.
fn token_shaped() -> String {
    format!("{}{}", ["gh", "p_"].concat(), "a1B2".repeat(9))
}

#[test]
fn a_note_records_the_session_that_wrote_it_and_show_prints_it() {
    let w = world(true);
    let card = w.ready_card("src/a/", 0, "p2");
    let note = |session: Option<&str>, extra: &[&str]| {
        let mut args = vec!["note", card.as_str()];
        args.extend_from_slice(extra);
        w.card_stdin(session, &args, b"text")
    };
    let as_leader = result(&note(Some(S_B), &["--kind", "decision", "--by", "leader"]));
    assert_eq!(as_leader["author"], json!("leader"));
    assert_eq!(
        as_leader["session"],
        json!(S_B),
        "--by does not hide the session"
    );
    let own = result(&note(Some(S_A), &["--kind", "comment"]));
    assert_eq!(own["author"], json!(S_A));
    assert_eq!(own["session"], json!(S_A));
    let by_hand = result(&note(None, &["--kind", "comment"]));
    assert_eq!(by_hand["author"], json!("user"));
    assert_eq!(by_hand["session"], Value::Null);

    let show = result(&w.card(None, &["show", &card]));
    let seen: Vec<(Value, Value)> = show["notes"]
        .as_array()
        .unwrap()
        .iter()
        .map(|n| (n["author"].clone(), n["session"].clone()))
        .collect();
    assert_eq!(
        seen,
        [
            (json!("leader"), json!(S_B)),
            (json!(S_A), json!(S_A)),
            (json!("user"), Value::Null),
        ]
    );
}

#[test]
fn a_decision_note_needs_an_explicit_author() {
    let w = world(true);
    let card = w.ready_card("src/a/", 0, "p2");
    let args = ["note", card.as_str(), "--kind", "decision"];
    for session in [None, Some(S_A)] {
        let (kind, message) = refusal(&w.card_stdin(session, &args, b"go left"));
        assert_eq!(kind, "invalid_input", "{session:?}: {message}");
        assert!(message.starts_with("card note refused: "), "{message}");
        assert!(message.contains("--by"), "{message}");
    }
    assert_eq!(w.count("SELECT count(*) FROM card_note"), 0);
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.note'"),
        0
    );
    // The other kinds keep their default author; a decision with --by is stored.
    result(&w.card_stdin(None, &["note", &card, "--kind", "report"], b"x"));
    let decided = result(&w.card_stdin(
        None,
        &["note", &card, "--kind", "decision", "--by", "user"],
        b"go left",
    ));
    assert_eq!(decided["author"], json!("user"));
    assert_eq!(decided["session"], Value::Null);
}

#[test]
fn a_review_card_is_not_released_but_reopened() {
    let w = world(true);
    let (card, g) = w.taken(S_A, "src/a/", 1);
    result(&w.card(Some(S_A), &["review", &card, "--generation", &g]));

    // Giving a review card back would leave it with nobody who can close it.
    let events = w.count("SELECT count(*) FROM event");
    let (kind, message) = refusal(&w.card(Some(S_A), &["release", &card, "--generation", &g]));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("`review`"), "{message}");
    assert!(
        message.contains("card reopen"),
        "it says what to do: {message}"
    );
    assert_eq!(
        w.count("SELECT count(*) FROM event"),
        events,
        "nothing was written"
    );
    let show = result(&w.card(None, &["show", &card]));
    assert_eq!(show["state"], json!("review"));
    assert_eq!(show["reservation"]["status"], json!("held"));

    // Reopen: holder only, then the card is active again and can be given back.
    let (kind, message) = refusal(&w.card(Some(S_B), &["reopen", &card, "--generation", &g]));
    assert_eq!(kind, "conflict", "{message}");
    let (kind, _) = refusal(&w.card(None, &["reopen", &card, "--generation", &g]));
    assert_eq!(kind, "invalid_input");
    let reopened = result(&w.card(Some(S_A), &["reopen", &card, "--generation", &g]));
    assert_eq!(reopened["action"], json!("reopen"));
    assert_eq!(reopened["from"], json!("review"));
    assert_eq!(reopened["state"], json!("active"));
    assert_eq!(reopened["reservation"]["status"], json!("held"));
    assert_eq!(reopened["events"], json!(1));
    let (kind, message) = refusal(&w.card(Some(S_A), &["reopen", &card, "--generation", &g]));
    assert_eq!(kind, "conflict", "not twice: {message}");
    let released = result(&w.card(Some(S_A), &["release", &card, "--generation", &g]));
    assert_eq!(released["state"], json!("ready"));
}

#[test]
fn next_shows_a_review_card_nobody_holds() {
    let w = world(true);
    let (card, g) = w.taken(S_A, "src/a/", 0);
    result(&w.card(Some(S_A), &["review", &card, "--generation", &g]));
    // Such a card can only come from outside the commands (an older build, a repair by hand).
    let mut l = ledger::open_write(w.home.path(), "demo").unwrap();
    l.with_write_tx(|tx, _| {
        tx.execute("UPDATE reservation SET status = 'released'", [])?;
        Ok(())
    })
    .unwrap();
    drop(l);
    let next = w.next(Some(S_A));
    assert_eq!(next["next"], Value::Null);
    assert_eq!(next["mine"], json!([]));
    assert_eq!(next["skipped"][0]["name"], json!(card));
    assert_eq!(next["skipped"][0]["state"], json!("review"));
    assert_eq!(next["skipped"][0]["reason"], json!("review_no_holder"));
}

#[test]
fn handoff_reports_the_state_the_card_is_in() {
    let w = world(true);
    let (card, g) = w.taken(S_A, "src/a/", 0);
    result(&w.card(Some(S_A), &["review", &card, "--generation", &g]));
    let handed = result(&w.card(
        Some(S_A),
        &["handoff", &card, "--generation", &g, "--to", S_B],
    ));
    assert_eq!(handed["state"], json!("review"));
    assert_eq!(handed["reservation"]["holder"], json!(S_B));
    let back = result(&w.card(Some(S_B), &["reopen", &card, "--generation", "2"]));
    assert_eq!(back["state"], json!("active"));
    let handed = result(&w.card(
        Some(S_B),
        &["handoff", &card, "--generation", "2", "--to", S_A],
    ));
    assert_eq!(handed["state"], json!("active"));
}

#[test]
fn a_card_from_a_host_project_takes_pull_requests_of_that_project_only() {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    let mut l = ledger::open_write(home.path(), "demo").unwrap();
    l.with_write_tx(|tx, now| {
        tx.execute(
            "INSERT INTO card (title, state, write_scope, created_by, created_at, updated_at, \
                               source_host, source_project, source_number) \
             VALUES ('t', 'ready', '[\"src/a/\"]', 'test', ?1, ?1, 'github', 'octo/demo', 3)",
            [now],
        )?;
        tx.execute(
            "INSERT INTO card (title, state, write_scope, created_by, created_at, updated_at) \
             VALUES ('local', 'ready', '[\"src/b/\"]', 'test', ?1, ?1)",
            [now],
        )?;
        Ok(())
    })
    .unwrap();
    for (card, scope) in [(1, "src/a/"), (2, "src/b/")] {
        ledger::reservation::reserve(&mut l, card, S_A, &[scope.to_string()], 3600, None, None)
            .unwrap();
    }
    drop(l);
    let fake = FakeConnector::new(HostKind::Github).with_pr(PrState {
        number: 12,
        phase: PrPhase::Open,
        head: HEAD.to_string(),
        merge_commit: None,
    });
    let request = |repo: Option<&'static str>| work::PrRequest {
        holder: S_A,
        generation: 1,
        number: 12,
        from: None,
        repo,
    };
    for other in ["evil/demo", "octo/other", "octo/demo2"] {
        let refused =
            work::pr(&project, home.path(), "C-1", &request(Some(other)), &fake).unwrap_err();
        assert!(refused.starts_with("card pr refused: "), "{refused}");
        assert!(
            refused.contains("octo/demo"),
            "it names the card's project: {refused}"
        );
    }
    assert!(fake.calls().is_empty(), "a refused project is never asked");
    // The card's own project, however it is spelled, and no --repo at all.
    work::pr(
        &project,
        home.path(),
        "C-1",
        &request(Some("Octo/Demo")),
        &fake,
    )
    .unwrap();
    work::pr(&project, home.path(), "C-1", &request(None), &fake).unwrap();
    // A card that came from nowhere keeps --repo.
    let local = work::pr(
        &project,
        home.path(),
        "C-2",
        &request(Some("any/where")),
        &fake,
    )
    .unwrap();
    assert_eq!(local["to"], json!("github:any/where#12"));
    assert_eq!(fake.calls().len(), 3);
}

#[test]
fn a_note_is_measured_after_the_trailing_line_breaks_are_gone() {
    let w = world(true);
    let card = w.ready_card("src/a/", 0, "p2");
    let args = ["note", card.as_str(), "--kind", "report"];
    let mut edge = vec![b'a'; 1_048_576];
    edge.extend(vec![b'\n'; 5_000]);
    assert_eq!(
        result(&w.card_stdin(None, &args, &edge))["bytes"],
        json!(1_048_576),
        "exactly one mebibyte of text, whatever follows it"
    );
    let mut over = vec![b'a'; 1_048_577];
    over.extend(vec![b'\n'; 5_000]);
    let (kind, message) = refusal(&w.card_stdin(None, &args, &over));
    assert_eq!(kind, "invalid_input", "{message}");
    // Text after a long run of line breaks is still text: nothing is cut short in silence.
    let mut split = vec![b'a'; 1_048_000];
    split.extend(vec![b'\n'; 5_000]);
    split.extend_from_slice(b"tail");
    let (kind, message) = refusal(&w.card_stdin(None, &args, &split));
    assert_eq!(kind, "invalid_input", "{message}");
    let mut inside = vec![b'a'; 1_000];
    inside.extend(vec![b'\n'; 5_000]);
    inside.extend_from_slice(b"tail");
    assert_eq!(
        result(&w.card_stdin(None, &args, &inside))["bytes"],
        json!(6_004)
    );
    assert_eq!(w.count("SELECT count(*) FROM card_note"), 2);
}

#[test]
fn a_note_with_the_shape_of_a_secret_or_a_terminal_escape_is_refused() {
    let w = world(true);
    let card = w.ready_card("src/a/", 0, "p2");
    let args = ["note", card.as_str(), "--kind", "report"];
    let token = token_shaped();
    let key = format!("{}{}", ["s", "k-"].concat(), "a1B2".repeat(6));
    for text in [
        format!("the token is {token}"),
        format!("first line\n{key}\nlast line"),
        format!("{}=hunter2", ["pass", "word"].concat()),
    ] {
        let (kind, message) = refusal(&w.card_stdin(None, &args, text.as_bytes()));
        assert_eq!(kind, "invalid_input", "{message}");
        assert!(message.starts_with("card note refused: "), "{message}");
        assert!(message.contains("secret"), "{message}");
        assert!(
            !message.contains(&token) && !message.contains(&key) && !message.contains("hunter2"),
            "the refusal does not repeat the text: {message}"
        );
    }
    for text in [
        &b"red \x1b[31mtext"[..],
        &b"carriage\rreturn"[..],
        &b"bell\x07"[..],
    ] {
        let (kind, message) = refusal(&w.card_stdin(None, &args, text));
        assert_eq!(kind, "invalid_input", "{message}");
        assert!(message.contains("control"), "{message}");
        // C-68 AC-3: a lone carriage return is named, not folded into a nameless refusal.
        if text.starts_with(b"carriage") {
            assert!(message.contains("carriage"), "{message}");
        }
    }
    assert_eq!(w.count("SELECT count(*) FROM card_note"), 0);
    // Ordinary words that only contain a short shape are notes like any other.
    let plain = result(&w.card_stdin(
        None,
        &args,
        b"risk-based task-1: ask-first\n\tindented token count",
    ));
    assert_eq!(plain["note_id"], json!(1));
}

/// C-68 AC-3 (card criterion 3): the note screen passes the two false positives the audit
/// named (a hyphenated lower-case name after the sk prefix, a prose option assignment), still
/// refuses a real key run and a secret-shaped option value, refuses directional characters
/// naming them, screens the one-line fields the same way as the body, and normalises Windows
/// line endings while refusing a lone carriage return by name.
#[test]
fn directional_characters_are_named_prose_passes_and_fields_are_screened() {
    let w = world(true);
    let card = w.ready_card("src/a/", 0, "p2");
    let args = ["note", card.as_str(), "--kind", "report"];
    let key = format!("{}{}", ["s", "k-"].concat(), "a1B2".repeat(6));

    // The two listed false positives pass (C-53 S1 re-audit N-c): a lower-case hyphenated name
    // after `sk` and an option assignment whose name is not a secret word. Since C-122
    // (SPEC-20261119-card-secret-screen, user decision 2026-11-19) a secret-named option
    // assignment is refused whatever its value is, so the passing prose uses `--mode=VALUE`.
    let plain = result(&w.card_stdin(
        None,
        &args,
        b"migration: sk-learn-pipeline tuned; run with --mode=VALUE in dry mode\n",
    ));
    assert_eq!(plain["action"], json!("note"), "{}", plain);
    let (kind, message) = refusal(&w.card_stdin(
        None,
        &args,
        b"migration: run with --token=VALUE in dry mode\n",
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("secret"), "{message}");
    // A real key run still goes, alone or behind an option name.
    for text in [format!("key {key}\n"), format!("run --flag={key}\n")] {
        let (kind, message) = refusal(&w.card_stdin(None, &args, text.as_bytes()));
        assert_eq!(kind, "invalid_input", "{message}");
        assert!(message.contains("secret"), "{message}");
    }

    // Directional characters are refused and the refusal names them (not a bare "control").
    for text in [
        format!("bidi {}\u{202E}reversed", "x"),
        format!("isolate {}\u{2066}start", "x"),
        format!("sep {}\u{2028}line", "x"),
    ] {
        let (kind, message) = refusal(&w.card_stdin(None, &args, text.as_bytes()));
        assert_eq!(kind, "invalid_input", "{message}");
        assert!(message.contains("direction"), "{message}");
    }

    // Windows line endings are normalised into the stored note; a lone CR is refused by name.
    let crlf = result(&w.card_stdin(None, &args, b"windows line one\r\nwindows line two\r\n"));
    assert_eq!(crlf["action"], json!("note"), "{}", crlf);
    // The pipe's end-of-text trim takes the trailing CRLF; the inner pair became one line
    // feed: 16 + 1 + 16 bytes.
    assert_eq!(crlf["bytes"], json!(33), "{}", crlf);
    let (kind, message) = refusal(&w.card_stdin(None, &args, b"mid\rdle\n"));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("carriage"), "{message}");

    // The one-line fields pass the same screen: a secret shape and a directional in a ref or
    // an evidence value are refused; plain values are taken.
    let (kind, message) = refusal(&w.card_stdin(
        None,
        &[
            "note",
            card.as_str(),
            "--kind",
            "report",
            "--ref",
            key.as_str(),
        ],
        b"plain body\n",
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("secret"), "{message}");
    let (kind, message) = refusal(&w.card_stdin(
        None,
        &[
            "note",
            card.as_str(),
            "--kind",
            "report",
            "--ref",
            "note:1:\u{202E}x",
        ],
        b"plain body\n",
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("direction"), "{message}");
    let plain_ref = result(&w.card_stdin(
        None,
        &["note", card.as_str(), "--kind", "report", "--ref", "plan-1"],
        b"plain body\n",
    ));
    assert_eq!(plain_ref["action"], json!("note"), "{}", plain_ref);
    let (kind, message) = refusal(&w.card(
        Some(S_A),
        &[
            "criterion",
            card.as_str(),
            "1",
            "--generation",
            "1",
            "--met",
            "--evidence",
            &format!("cargo test {key}"),
        ],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("secret"), "{message}");
    let (kind, message) = refusal(&w.card(
        Some(S_A),
        &[
            "criterion",
            card.as_str(),
            "1",
            "--generation",
            "1",
            "--met",
            "--evidence",
            "cargo test \u{202E}hidden",
        ],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("direction"), "{message}");
}
