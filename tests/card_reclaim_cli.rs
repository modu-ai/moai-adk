//! `moai card reclaim` (SPEC-20261119-card-reclaim): a session takes over a card whose
//! reservation has ended, so the card can be closed. The stuck shape is a `review` card whose
//! holder is gone, whose pull request is merged and whose lease ended: `card renew`, `criterion`
//! and `release` refuse it. The reservation rules are `tests/ledger_reservation.rs`; here the
//! command, its answers, its refusal kinds and what it leaves in the ledger.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`). `card pr` reads the
//! pull request through a stub `gh` (`MOAI_FORGE_GH_BIN`); no test reaches the network.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const MERGE: &str = "0123456789abcdef0123456789abcdef01234567";
const HEAD: &str = "89abcdef0123456789abcdef0123456789abcdef";
const REASON: &str = "S_A is gone (process ended 2026-11-18); PR 7 merged; criterion 3 not done";

struct World {
    home: TempHome,
    _host: tempfile::TempDir,
    dir: PathBuf,
    project: String,
    gh: PathBuf,
    /// The absolute git path the reclaim gate hands its child: the runner's clean
    /// environment carries no `PATH`, so the child cannot find `git` itself.
    git: Option<PathBuf>,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
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
        home,
        _host: host,
        dir,
        project: project.to_str().unwrap().to_string(),
        gh,
        git: git_executable(),
    }
}

/// An absolute `git` on this runner's `PATH`, for the child that has none.
fn git_executable() -> Option<PathBuf> {
    let path = std::env::var_os("PATH")?;
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
}

impl World {
    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        cmd.env("MOAI_FORGE_GH_BIN", &self.gh);
        if let Some(git) = &self.git {
            cmd.env("MOAI_GIT_BIN", git);
        }
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

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

    /// Runs `statement` against the ledger directly (the only way to end a lease: its length
    /// is at least 60 seconds).
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

    fn rows(&self, sql: &str) -> Vec<Value> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l.conn().prepare(sql).unwrap();
        let n = st.column_count();
        st.query_map([], |r| {
            Ok(Value::Array(
                (0..n)
                    .map(|i| match r.get_ref(i).unwrap() {
                        rusqlite::types::ValueRef::Integer(v) => json!(v),
                        rusqlite::types::ValueRef::Text(t) => {
                            json!(String::from_utf8_lossy(t).to_string())
                        }
                        _ => Value::Null,
                    })
                    .collect(),
            ))
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
    }

    /// A card with three criteria taken by `S_A` (generation 1, with a worktree and a branch),
    /// two criteria met, a note, an open pull request linked and the card in `review`; the lease
    /// then ended. This is the shape of the cards that were stuck.
    fn stuck_review_card(&self) -> String {
        let added = result(&self.card(
            None,
            &[
                "add", "--title", "stuck", "--scope", "apps/bo/", "--ac", "one", "--ac", "two",
                "--ac", "three",
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(
            Some(S_A),
            &[
                "take",
                &name,
                "--worktree",
                "/work/tree",
                "--branch",
                "card/1-stuck",
            ],
        ));
        for n in ["1", "2"] {
            result(&self.card(
                Some(S_A),
                &[
                    "criterion",
                    &name,
                    n,
                    "--generation",
                    "1",
                    "--met",
                    "--evidence",
                    "test x",
                ],
            ));
        }
        result(&self.card_note(S_A, &name, "PLAN the work"));
        self.host_pr(7, "closed", true);
        result(&self.card(
            Some(S_A),
            &["pr", &name, "7", "--generation", "1", "--repo", "octo/demo"],
        ));
        result(&self.card(Some(S_A), &["review", &name, "--generation", "1"]));
        self.end_lease(1);
        name
    }

    fn card_note(&self, session: &str, card: &str, text: &str) -> Output {
        let mut cmd = self.home.command(&[
            "card",
            "note",
            card,
            "--kind",
            "comment",
            "--project",
            &self.project,
        ]);
        cmd.env(SESSION_VAR, session)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        let mut child = cmd.spawn().expect("run moai");
        common::feed_stdin(&mut child, text.as_bytes());
        child.wait_with_output().unwrap()
    }

    fn reclaim(&self, session: Option<&str>, card: &str, generation: &str) -> Output {
        self.card(
            session,
            &[
                "reclaim",
                card,
                "--generation",
                generation,
                "--by",
                "leader",
                "--reason",
                REASON,
            ],
        )
    }
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

#[test]
fn reclaim_lets_the_new_holder_close_a_review_card() {
    let w = world();
    let card = w.stuck_review_card();

    // The commands that were blocked for the stuck cards still are, for the old holder.
    let (kind, message) = refusal(&w.card(Some(S_A), &["renew", &card, "--generation", "1"]));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("the lease ended"), "{message}");
    let (kind, message) = refusal(&w.card(Some(S_A), &["release", &card, "--generation", "1"]));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("is `review`"), "{message}");

    let kept = ["card_ac", "card_note", "card_link"];
    let before = {
        let l = ledger::open_read_only(w.home.path(), "demo").unwrap();
        dump_tables(l.conn(), &kept)
    };

    let got = result(&w.reclaim(Some(S_B), &card, "1"));
    assert_eq!(got["action"], json!("reclaim"));
    assert_eq!(got["state"], json!("review"), "the card keeps its state");
    assert_eq!(got["by"], json!("leader"));
    assert_eq!(got["reason"], json!(REASON));
    assert_eq!(got["events"], json!(1));
    assert_eq!(got["writes"], json!(true));
    let new = &got["reservation"];
    assert_eq!(new["generation"], json!(2));
    assert_eq!(new["holder"], json!(S_B));
    assert_eq!(new["status"], json!("held"));
    assert_eq!(new["worktree"], json!("/work/tree"));
    assert_eq!(new["branch"], json!("card/1-stuck"));
    assert_eq!(new["write_scope"], json!(["apps/bo/"]));
    let old = &got["previous"];
    assert_eq!(old["generation"], json!(1));
    assert_eq!(old["holder"], json!(S_A));
    assert_eq!(old["status"], json!("reclaimed"));

    // Notes, criteria and links are the bytes they were.
    let after = {
        let l = ledger::open_read_only(w.home.path(), "demo").unwrap();
        dump_tables(l.conn(), &kept)
    };
    assert_eq!(before, after);

    // The reclaimer finishes the card: the open criterion, the merged pull request, `done`.
    let shown = result(&w.card(None, &["show", &card]));
    assert_eq!(shown["reservation"]["holder"], json!(S_B));
    assert_eq!(shown["reservation"]["generation"], json!(2));
    result(&w.card(
        Some(S_B),
        &[
            "criterion",
            &card,
            "3",
            "--generation",
            "2",
            "--met",
            "--evidence",
            "done by S_B",
        ],
    ));
    result(&w.card(
        Some(S_B),
        &["pr", &card, "7", "--generation", "2", "--repo", "octo/demo"],
    ));
    let done = result(&w.card(Some(S_B), &["done", &card, "--generation", "2"]));
    assert_eq!(done["state"], json!("done"));
    assert_eq!(done["reservation"]["status"], json!("released"));
    assert_eq!(
        w.rows("SELECT generation, status FROM reservation ORDER BY generation"),
        vec![json!([1, "reclaimed"]), json!([2, "released"])]
    );
}

#[test]
fn reclaim_can_hand_an_active_card_back_to_the_ready_queue() {
    let w = world();
    let added = result(&w.card(None, &["add", "--title", "t", "--scope", "src/a/"]));
    let card = added["name"].as_str().unwrap().to_string();
    result(&w.card(None, &["ready", &card]));
    result(&w.card(Some(S_A), &["take", &card]));
    w.end_lease(1);
    let got = result(&w.reclaim(Some(S_B), &card, "1"));
    assert_eq!(got["state"], json!("active"));
    let released = result(&w.card(Some(S_B), &["release", &card, "--generation", "2"]));
    assert_eq!(released["state"], json!("ready"));
}

#[test]
fn reclaim_refusals_change_nothing() {
    let w = world();
    let card = w.stuck_review_card();
    // A card that is not stuck at all: its lease is live.
    let live = result(&w.card(None, &["add", "--title", "live", "--scope", "src/live/"]));
    let live = live["name"].as_str().unwrap().to_string();
    result(&w.card(None, &["ready", &live]));
    result(&w.card(Some(S_A), &["take", &live]));

    let check = |name: &str, out: Output, kind: &str, needle: &str, before: &[String]| {
        let (k, message) = refusal(&out);
        assert_eq!(k, kind, "{name}: {message}");
        assert!(message.contains(needle), "{name}: {message}");
        assert_eq!(&w.dump(), before, "{name} wrote something");
    };
    let before = w.dump();
    check(
        "live lease",
        w.reclaim(Some(S_B), &live, "1"),
        "conflict",
        "only an ended lease is reclaimed",
        &before,
    );
    check(
        "never issued generation",
        w.reclaim(Some(S_B), &card, "2"),
        "conflict",
        "was never issued",
        &before,
    );
    check(
        "missing card",
        w.reclaim(Some(S_B), "C-99", "1"),
        "not_found",
        "card not found",
        &before,
    );
    check(
        "host number",
        w.reclaim(Some(S_B), "#7", "1"),
        "invalid_input",
        "name the card by its C-number",
        &before,
    );
    check(
        "no session id",
        w.reclaim(None, &card, "1"),
        "invalid_input",
        "this session has no id",
        &before,
    );
    for (name, by, reason, needle) in [
        ("empty by", "", REASON, "--by must be one line"),
        (
            "by too long",
            &"b".repeat(81),
            REASON,
            "--by must be one line",
        ),
        ("empty reason", "leader", "", "--reason must be"),
        ("blank reason", "leader", "   ", "--reason must be"),
        ("two lines", "leader", "a\nb", "--reason must be"),
        (
            "reason too long",
            "leader",
            &"r".repeat(501),
            "--reason must be",
        ),
    ] {
        check(
            name,
            w.card(
                Some(S_B),
                &[
                    "reclaim",
                    &card,
                    "--generation",
                    "1",
                    "--by",
                    by,
                    "--reason",
                    reason,
                ],
            ),
            "invalid_input",
            needle,
            &before,
        );
    }
    for (name, lease) in [("lease 59", "59"), ("lease 86401", "86401")] {
        check(
            name,
            w.card(
                Some(S_B),
                &[
                    "reclaim",
                    &card,
                    "--generation",
                    "1",
                    "--by",
                    "leader",
                    "--reason",
                    REASON,
                    "--lease-secs",
                    lease,
                ],
            ),
            "invalid_input",
            "lease_secs",
            &before,
        );
    }
    // Missing options are a usage error, and nothing is written either.
    let missing = w.card(Some(S_B), &["reclaim", &card, "--generation", "1"]);
    assert_ne!(missing.status.code(), Some(0));
    assert_eq!(w.dump(), before);

    // The limits are accepted: 80 characters of `--by`, 500 of the reason, a 60 second lease.
    let got = result(&w.card(
        Some(S_B),
        &[
            "reclaim",
            &card,
            "--generation",
            "1",
            "--by",
            &"b".repeat(80),
            "--reason",
            &"r".repeat(500),
            "--lease-secs",
            "60",
        ],
    ));
    assert_eq!(got["reservation"]["generation"], json!(2));

    // A reclaim at the old generation is stale once it has happened, and a done card is no
    // reclaim target: close the card, then try both.
    let before = w.dump();
    check(
        "stale generation",
        w.reclaim(Some(S_A), &card, "1"),
        "conflict",
        "is stale",
        &before,
    );
    w.host_pr(7, "closed", true);
    result(&w.card(
        Some(S_B),
        &[
            "criterion",
            &card,
            "3",
            "--generation",
            "2",
            "--met",
            "--evidence",
            "x",
        ],
    ));
    result(&w.card(
        Some(S_B),
        &["pr", &card, "7", "--generation", "2", "--repo", "octo/demo"],
    ));
    result(&w.card(Some(S_B), &["done", &card, "--generation", "2"]));
    let before = w.dump();
    check(
        "done card",
        w.reclaim(Some(S_A), &card, "2"),
        "conflict",
        "is `done`",
        &before,
    );
}

#[test]
fn the_reclaim_event_names_everything_and_the_old_generation_is_stale() {
    let w = world();
    let card = w.stuck_review_card();
    let ended = w.rows("SELECT lease_until FROM reservation WHERE generation = 1")[0][0]
        .as_i64()
        .unwrap();
    result(&w.reclaim(Some(S_B), &card, "1"));

    let events = w.rows(
        "SELECT actor, card_id, payload FROM event WHERE kind = 'card.reclaim' ORDER BY event_id",
    );
    assert_eq!(events.len(), 1);
    assert_eq!(events[0][0], json!("leader"), "the actor is `--by`");
    assert_eq!(events[0][1], json!(1));
    let p: Value = serde_json::from_str(events[0][2].as_str().unwrap()).unwrap();
    assert_eq!(p["from_holder"], json!(S_A));
    assert_eq!(p["from_generation"], json!(1));
    assert_eq!(p["from_lease_until"], json!(ended));
    assert_eq!(p["holder"], json!(S_B));
    assert_eq!(p["generation"], json!(2));
    assert_eq!(p["by"], json!("leader"));
    assert_eq!(p["reason"], json!(REASON));
    assert_eq!(p["card_state"], json!("review"));

    // The old generation is refused by every fence-checked call; the new one is not.
    for args in [
        vec!["renew", &card[..], "--generation", "1"],
        vec!["release", &card[..], "--generation", "1"],
        vec![
            "handoff",
            &card[..],
            "--generation",
            "1",
            "--to",
            "cccccccc-0000-4000-8000-00000000000c",
        ],
        vec![
            "criterion",
            &card[..],
            "3",
            "--generation",
            "1",
            "--met",
            "--evidence",
            "x",
        ],
        vec![
            "check",
            &card[..],
            "--generation",
            "1",
            "--path",
            "apps/bo/a.rs",
        ],
    ] {
        let (kind, message) = refusal(&w.card(Some(S_A), &args));
        assert_eq!(kind, "conflict", "{args:?}: {message}");
        assert!(message.contains("is stale"), "{args:?}: {message}");
    }
    let allowed = result(&w.card(
        Some(S_B),
        &[
            "check",
            &card,
            "--generation",
            "2",
            "--path",
            "apps/bo/a.rs",
        ],
    ));
    assert_eq!(allowed["allowed"], json!(true));
}

#[test]
fn the_help_says_the_command_writes_the_ledger() {
    let w = world();
    let out = w
        .home
        .command(&["card", "reclaim", "--help"])
        .output()
        .expect("run moai");
    let help = String::from_utf8_lossy(&out.stdout);
    assert!(help.starts_with("Writes the ledger."), "{help}");
    for option in ["--generation", "--by", "--reason", "--lease-secs"] {
        assert!(help.contains(option), "{option} missing: {help}");
    }
}

/// SPEC-20261119-card-reclaim-followups AC-1: the help says what a reader needs to act on it.
#[test]
fn reclaim_help_names_reopen_and_the_refusal_kinds() {
    let w = world();
    let out = w
        .home
        .command(&["card", "reclaim", "--help"])
        .output()
        .expect("run moai");
    let help = String::from_utf8_lossy(&out.stdout);
    // The help is wrapped at the terminal width: compare words, not line breaks.
    let help = help.split_whitespace().collect::<Vec<_>>().join(" ");
    assert!(
        help.contains("`card reopen`") && help.contains("`card release`"),
        "a `review` card is reopened before it is given back: {help}"
    );
    for kind in ["`conflict`", "`not_found`", "`usage`", "`invalid_input`"] {
        assert!(help.contains(kind), "{kind} missing: {help}");
    }
    // An option that is missing or cannot be read is a usage error (exit 2), not an
    // `invalid_input` (audit of #3223, B1).
    assert!(
        help.contains("`usage` (exit 2"),
        "the help does not say a missing option is `usage`: {help}"
    );
}

/// Audit of #3223 B1: a missing or unreadable option is a `usage` error (exit 2, the parser's
/// refusal before any ledger is read); a value that reads but is out of range is `invalid_input`
/// (exit 1). The help says the same; nothing is written either way.
#[test]
fn a_missing_or_unreadable_option_is_usage_and_an_out_of_range_value_is_invalid_input() {
    let w = world();
    let card = w.stuck_review_card();
    let before = w.dump();
    let usage = |args: &[&str]| {
        let mut all = vec!["reclaim", card.as_str()];
        all.extend_from_slice(args);
        let out = w.card(Some(S_B), &all);
        assert_eq!(
            out.status.code(),
            Some(2),
            "{args:?}: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        let v = envelope(&out);
        assert_eq!(v["ok"], json!(false), "{args:?}: {v}");
        assert_eq!(v["error"]["kind"], json!("usage"), "{args:?}: {v}");
    };
    usage(&["--by", "leader", "--reason", REASON]); // no --generation
    usage(&["--generation", "1", "--reason", REASON]); // no --by
    usage(&["--generation", "1", "--by", "leader"]); // no --reason
    usage(&["--generation", "x", "--by", "leader", "--reason", REASON]);
    usage(&["--generation", "-1", "--by", "leader", "--reason", REASON]);

    // Out of range: the option reads, the value is refused by the command.
    let (kind, _) = refusal(&w.card(
        Some(S_B),
        &[
            "reclaim",
            &card,
            "--generation",
            "1",
            "--by",
            "leader",
            "--reason",
            " ",
        ],
    ));
    assert_eq!(kind, "invalid_input");
    let (kind, _) = refusal(&w.card(
        Some(S_B),
        &[
            "reclaim",
            &card,
            "--generation",
            "1",
            "--by",
            "leader",
            "--reason",
            REASON,
            "--lease-secs",
            "59",
        ],
    ));
    assert_eq!(kind, "invalid_input");
    assert_eq!(w.dump(), before, "a refused reclaim writes nothing");
}

/// AC-3: `--reason` is kept for good and `card export` writes it out, so it gets the filter
/// `card note` uses. The refusal does not repeat the word and nothing is written.
#[test]
fn reclaim_refuses_a_secret_shaped_reason() {
    let w = world();
    let card = w.stuck_review_card();
    let key = format!("{}{}", ["s", "k-"].concat(), "a1B2".repeat(6));
    let before = w.dump();
    for reason in [
        format!("old holder gone, key {key} left in its shell"),
        format!("{}=hunter2 was in the log", ["pass", "word"].concat()),
    ] {
        let (kind, message) = refusal(&w.card(
            Some(S_B),
            &[
                "reclaim",
                &card,
                "--generation",
                "1",
                "--by",
                "leader",
                "--reason",
                &reason,
            ],
        ));
        assert_eq!(kind, "invalid_input", "{message}");
        assert!(
            message.starts_with("card reclaim refused: ") && message.contains("secret"),
            "{message}"
        );
        assert!(
            !message.contains(&key) && !message.contains("hunter2"),
            "the refusal does not repeat the text: {message}"
        );
        assert_eq!(w.dump(), before, "a refused reclaim writes nothing");
    }
    // An ordinary reason is taken (the card was not touched above).
    result(&w.reclaim(Some(S_B), &card, "1"));
}

/// AC-4: after a reclaim the old holder's `pr`, `done`, `reopen` and `review` are refused as
/// stale. They leave one `fence_reject` event each and change no other table.
#[test]
fn the_old_generation_cannot_move_or_close_the_card() {
    let w = world();
    let card = w.stuck_review_card();
    result(&w.reclaim(Some(S_B), &card, "1"));
    let tables = |w: &World| -> Vec<String> {
        w.dump()
            .into_iter()
            .filter(|line| !line.starts_with("event:"))
            .collect()
    };
    let before = tables(&w);
    let events_before = w.rows("SELECT COUNT(*) FROM event")[0][0].as_i64().unwrap();

    let attempts: [&[&str]; 4] = [
        &["pr", &card, "7", "--generation", "1", "--repo", "octo/demo"],
        &["done", &card, "--generation", "1"],
        &["reopen", &card, "--generation", "1"],
        &["review", &card, "--generation", "1"],
    ];
    for args in attempts {
        let (kind, message) = refusal(&w.card(Some(S_A), args));
        assert_eq!(kind, "conflict", "{args:?}: {message}");
        assert!(message.contains("is stale"), "{args:?}: {message}");
    }

    assert_eq!(tables(&w), before, "no card, reservation or link changed");
    let rejects = w.rows("SELECT kind FROM event ORDER BY event_id DESC LIMIT 4");
    assert!(
        rejects.iter().all(|r| r[0] == json!("fence_reject")),
        "{rejects:?}"
    );
    assert_eq!(
        w.rows("SELECT COUNT(*) FROM event")[0][0].as_i64().unwrap(),
        events_before + 4
    );
    // The new holder still closes the card (criterion 3 was open; the pull request is merged).
    result(&w.card(
        Some(S_B),
        &[
            "criterion",
            &card,
            "3",
            "--generation",
            "2",
            "--met",
            "--evidence",
            "x",
        ],
    ));
    let done = result(&w.card(Some(S_B), &["done", &card, "--generation", "2"]));
    assert_eq!(done["state"], json!("done"), "{done}");
}

// ---- takeover gates (SPEC-20261119-card-reclaim-guard) ----

impl World {
    /// A git repo with one commit; `pushed` adds a bare remote and pushes; `dirty`
    /// leaves an untracked file behind. Identity and guards come from `common::git_in`.
    fn git_worktree(&self, pushed: bool, dirty: bool) -> String {
        // A folder of its own on every call: a second repo must not inherit the first one's
        // state (an untracked file, a remote).
        let k = (0..)
            .find(|i| !self.dir.join(format!("wt{i}")).exists())
            .unwrap();
        let wt = self.dir.join(format!("wt{k}"));
        fs::create_dir_all(&wt).unwrap();
        common::git::git_in(&wt, &["init"]);
        fs::write(wt.join("f.txt"), "one\n").unwrap();
        common::git::git_in(&wt, &["add", "."]);
        common::git::git_in(&wt, &["commit", "-m", "one"]);
        if pushed {
            let bare = self.dir.join(format!("bare{k}.git"));
            fs::create_dir_all(&bare).unwrap();
            common::git::git_in(&bare, &["init", "--bare"]);
            common::git::git_in(&wt, &["remote", "add", "origin", bare.to_str().unwrap()]);
            common::git::git_in(&wt, &["push", "origin", "HEAD"]);
        }
        if dirty {
            fs::write(wt.join("g.txt"), "two\n").unwrap();
        }
        wt.to_str().unwrap().to_string()
    }

    /// A card taken by S_A whose reservation names `worktree`; the lease then ended.
    fn ended_card_at(&self, worktree: &str) -> String {
        // A write scope of its own (named after the worktree folder), so two cards of one test
        // never meet the scope-overlap refusal.
        let scope = format!(
            "apps/{}/",
            Path::new(worktree)
                .file_name()
                .and_then(|n| n.to_str())
                .unwrap_or("wt")
        );
        let added = result(&self.card(
            None,
            &["add", "--title", "gate", "--scope", &scope, "--ac", "one"],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        let n: i64 = name.trim_start_matches("C-").parse().unwrap();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(
            Some(S_A),
            &[
                "take",
                &name,
                "--worktree",
                worktree,
                "--branch",
                &format!("card/{n}-gate"),
            ],
        ));
        self.end_lease(n);
        name
    }
}

#[test]
fn reclaim_gates_refuse_a_worktree_with_uncommitted_or_unpushed_work() {
    let w = world();
    // Pushed but dirty: one untracked file.
    let wt = w.git_worktree(true, true);
    let name = w.ended_card_at(&wt);
    let before = w.dump();
    let (kind, message) = refusal(&w.reclaim(Some(S_B), &name, "1"));
    assert_eq!(kind, "conflict", "{message}");
    // The message names both counts (REQ-1), not only the two words.
    assert!(
        message.contains("holds 1 uncommitted change(s) and 0 unpushed commit(s)"),
        "{message}"
    );
    assert!(message.contains(&wt), "{message}");
    assert_eq!(w.dump(), before, "the refusal leaves every row as it was");

    // Clean but unpushed: one commit with no remote-tracking ref.
    let wt2 = w.git_worktree(false, false);
    let name2 = w.ended_card_at(&wt2);
    let before2 = w.dump();
    let (kind2, message2) = refusal(&w.reclaim(Some(S_B), &name2, "1"));
    assert_eq!(kind2, "conflict", "{message2}");
    assert!(
        message2.contains("holds 0 uncommitted change(s) and 1 unpushed commit(s)"),
        "{message2}"
    );
    assert_eq!(w.dump(), before2, "the refusal leaves every row as it was");
}

#[test]
fn reclaim_gates_proceed_with_an_absent_or_clean_worktree() {
    let w = world();
    let name = w.ended_card_at("/work/absent");
    let done = result(&w.reclaim(Some(S_B), &name, "1"));
    assert_eq!(done["reservation"]["status"], json!("held"), "{done}");
    // Give the scope back, or the next card's take hits the overlap refusal.
    result(&w.card(Some(S_B), &["release", &name, "--generation", "2"]));

    let wt = w.git_worktree(true, false);
    let name2 = w.ended_card_at(&wt);
    let done2 = result(&w.reclaim(Some(S_B), &name2, "1"));
    assert_eq!(done2["reservation"]["status"], json!("held"), "{done2}");
}

#[test]
fn reclaim_gates_refuse_when_git_cannot_be_read() {
    let w = world();
    let not_repo = w.dir.join("not-repo");
    fs::create_dir_all(&not_repo).unwrap();
    let name = w.ended_card_at(not_repo.to_str().unwrap());
    let before = w.dump();
    let (kind, message) = refusal(&w.reclaim(Some(S_B), &name, "1"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("could not be read"), "{message}");
    assert_eq!(w.dump(), before, "the refusal leaves every row as it was");
}

#[test]
fn reclaim_gates_refuse_a_recorded_path_that_is_not_absolute() {
    // A relative path names another folder under another working directory: "not found here"
    // proves nothing, so the gate refuses instead of reading it as "gone" (audit-c43 B1).
    let w = world();
    let name = w.ended_card_at("wt-rel");
    let before = w.dump();
    let (kind, message) = refusal(&w.reclaim(Some(S_B), &name, "1"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("not an absolute path"), "{message}");
    assert_eq!(w.dump(), before, "the refusal leaves every row as it was");
}

/// Gives a folder's permissions back when the test ends, passed or not.
#[cfg(unix)]
struct Unlock(PathBuf);

#[cfg(unix)]
impl Drop for Unlock {
    fn drop(&mut self) {
        use std::os::unix::fs::PermissionsExt;
        let _ = fs::set_permissions(&self.0, fs::Permissions::from_mode(0o755));
    }
}

#[cfg(unix)]
#[test]
fn reclaim_gates_refuse_a_path_that_cannot_be_stat_ed() {
    // `Path::exists` reads a permission error as "gone"; the gate must tell them apart.
    use std::os::unix::fs::PermissionsExt;
    let w = world();
    let parent = w.dir.join("locked");
    let inside = parent.join("wt-locked");
    fs::create_dir_all(&inside).unwrap();
    let name = w.ended_card_at(inside.to_str().unwrap());
    let _unlock = Unlock(parent.clone());
    fs::set_permissions(&parent, fs::Permissions::from_mode(0o000)).unwrap();
    if fs::metadata(&inside).is_ok() {
        eprintln!("skipped: this user can stat a mode-000 folder (root)");
        return;
    }
    let before = w.dump();
    let (kind, message) = refusal(&w.reclaim(Some(S_B), &name, "1"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("could not be read"), "{message}");
    assert_eq!(w.dump(), before, "the refusal leaves every row as it was");
}

#[test]
fn reclaim_gates_self_reclaim_skips_the_worktree_gate() {
    let w = world();
    let wt = w.git_worktree(true, true);
    let name = w.ended_card_at(&wt);
    let done = result(&w.reclaim(Some(S_A), &name, "1"));
    assert_eq!(done["reservation"]["holder"], json!(S_A), "{done}");
    assert_eq!(done["previous"]["status"], json!("reclaimed"), "{done}");
}

#[test]
fn reclaim_limit_refuses_the_fourth_reclaim_on_the_command_line() {
    let w = world();
    let name = w.ended_card_at("/work/absent");
    result(&w.reclaim(Some(S_B), &name, "1"));
    w.end_lease(1);
    result(&w.reclaim(Some(S_A), &name, "2"));
    w.end_lease(1);
    result(&w.reclaim(Some(S_B), &name, "3"));
    w.end_lease(1);
    let before = w.dump();
    let (kind, message) = refusal(&w.reclaim(Some(S_A), &name, "4"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("reclaim limit"), "{message}");
    assert_eq!(w.dump(), before, "the refusal leaves every row as it was");

    // The limit is per card: another card's first reclaim still goes through (audit-c43 NB-1).
    let other = w.ended_card_at("/work/other");
    let done = result(&w.reclaim(Some(S_B), &other, "1"));
    assert_eq!(done["reservation"]["status"], json!("held"), "{done}");
}
