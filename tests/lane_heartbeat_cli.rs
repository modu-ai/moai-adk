//! The silence sweep (SPEC-20261119-live-lease-takeover): a card whose lease ended and whose
//! holder wrote nothing to the ledger for the silence window returns to `ready` before the
//! 24-hour grace of `SPEC-20261124-ended-lease-sweep` has run out, and a card whose lease is
//! live, or whose holder wrote recently, is never touched.
//!
//! Every command runs against a temporary HOME; the real home ledger is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`). A test ages a card with
//! SQL on its own temporary ledger: the event table is append-only in a real ledger, so the guard
//! is lifted in the throwaway one, as `tests/card_cleanup_outbox_cli.rs` does.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger::{self, sweep};
use rusqlite::Connection;
use serde_json::{json, Value};
use std::io::Write;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const HOUR: i64 = 3600;
const THREE_HOURS: i64 = 3 * HOUR;
const FIVE_MINUTES: i64 = 300;
const TEN_HOURS: i64 = 10 * HOUR;
const TWENTY_FIVE_HOURS: i64 = 25 * HOUR;
const GRACE_SECS: i64 = 24 * HOUR;

struct World {
    home: TempHome,
    project: String,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    World {
        home,
        project: project.to_str().unwrap().to_string(),
    }
}

impl World {
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    /// `card note` as `session`, the text on stdin.
    fn note_as(&self, session: &str, card: &str, text: &str) {
        let mut cmd = self.home.command(&[
            "card",
            "note",
            card,
            "--kind",
            "report",
            "--project",
            &self.project,
        ]);
        cmd.env(SESSION_VAR, session);
        cmd.stdin(Stdio::piped()).stdout(Stdio::piped());
        let mut child = cmd.spawn().expect("run moai");
        child
            .stdin
            .take()
            .unwrap()
            .write_all(text.as_bytes())
            .unwrap();
        result(&child.wait_with_output().unwrap());
    }

    fn ready(&self, title: &str, scope: &[&str]) -> String {
        let mut args = vec!["add", "--title", title, "--ac", "one"];
        for entry in scope {
            args.push("--scope");
            args.push(entry);
        }
        let name = result(&self.card(None, &args))["name"]
            .as_str()
            .unwrap()
            .to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    fn take(&self, session: &str, name: &str) -> Output {
        self.card(
            Some(session),
            &[
                "take",
                name,
                "--worktree",
                "/work/tree",
                "--branch",
                "card/x",
            ],
        )
    }

    /// A ready card taken by `session` at generation 1.
    fn held(&self, session: &str, title: &str, scope: &[&str]) -> String {
        let name = self.ready(title, scope);
        result(&self.take(session, &name));
        name
    }

    fn id(card: &str) -> i64 {
        card.trim_start_matches("C-").parse().unwrap()
    }

    /// The lease of the held reservation of `card` ended `secs_ago` seconds ago.
    fn end_lease(&self, card: &str, secs_ago: i64) {
        self.sql(&format!(
            "UPDATE reservation SET lease_until = unixepoch() - {secs_ago} \
             WHERE card_id = {} AND status = 'held'",
            Self::id(card)
        ));
    }

    /// The lease of the held reservation of `card` runs for `secs_ahead` more seconds.
    fn live_lease(&self, card: &str, secs_ahead: i64) {
        self.sql(&format!(
            "UPDATE reservation SET lease_until = unixepoch() + {secs_ahead} \
             WHERE card_id = {} AND status = 'held'",
            Self::id(card)
        ));
    }

    /// The holder of `card` last wrote `secs_ago` seconds ago: the reservation and every event
    /// of the card move back, so nothing newer remains.
    fn go_silent(&self, card: &str, secs_ago: i64) {
        self.sql(&format!(
            "DROP TRIGGER IF EXISTS event_no_update; \
             UPDATE reservation SET updated_at = unixepoch() - {secs_ago} \
             WHERE card_id = {id}; \
             UPDATE event SET at = unixepoch() - {secs_ago} WHERE card_id = {id};",
            id = Self::id(card)
        ));
    }

    /// The reservation of `card` changed `secs_ago` seconds ago (a take, a renew).
    fn touched(&self, card: &str, secs_ago: i64) {
        self.sql(&format!(
            "UPDATE reservation SET updated_at = unixepoch() - {secs_ago} WHERE card_id = {}",
            Self::id(card)
        ));
    }

    /// The held reservation of `card` becomes `status`, its lease ended `secs_ago` seconds ago.
    fn set_status(&self, card: &str, status: &str, secs_ago: i64) {
        self.sql(&format!(
            "UPDATE reservation SET status = '{status}', lease_until = unixepoch() - {secs_ago} \
             WHERE card_id = {} AND status = 'held'",
            Self::id(card)
        ));
    }

    fn sql(&self, script: &str) {
        let file = ledger::ledger_path(self.home.path(), "demo").unwrap();
        let conn = Connection::open(&file).unwrap();
        conn.execute_batch(script).unwrap();
    }

    fn sweep(&self, args: &[&str]) -> Output {
        let mut all = vec!["sweep"];
        all.extend_from_slice(args);
        self.card(None, &all)
    }

    fn state_of(&self, card: &str) -> String {
        result(&self.card(None, &["show", card]))["state"]
            .as_str()
            .unwrap()
            .to_string()
    }

    fn notes_of(&self, card: &str) -> Vec<String> {
        result(&self.card(None, &["show", card]))["notes"]
            .as_array()
            .unwrap()
            .iter()
            .map(|n| n["body"].as_str().unwrap().to_string())
            .collect()
    }

    /// The payload of the one event of `kind` that `card` has.
    fn event_payload(&self, card: &str, kind: &str) -> Value {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let text: String = l
            .conn()
            .query_row(
                "SELECT payload FROM event WHERE card_id = ?1 AND kind = ?2",
                rusqlite::params![Self::id(card), kind],
                |r| r.get(0),
            )
            .unwrap();
        serde_json::from_str(&text).unwrap()
    }

    fn count(&self, table: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn()
            .query_row(&format!("SELECT count(*) FROM {table}"), [], |r| r.get(0))
            .unwrap()
    }

    fn count_kind(&self, kind: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn()
            .query_row("SELECT count(*) FROM event WHERE kind = ?1", [kind], |r| {
                r.get(0)
            })
            .unwrap()
    }

    /// Every row count that a refused or skipped call must leave as it was.
    fn rows(&self) -> (i64, i64, i64) {
        (
            self.count("event"),
            self.count("card_note"),
            self.count("reservation"),
        )
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
    v["result"].clone()
}

fn refusal(out: &Output) -> String {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
}

// ---- AC-1: a silent holder with an ended lease ----------------------------------------------------

#[test]
fn silent_holder_with_an_ended_lease_is_swept_before_the_grace_runs_out() {
    let w = world();
    let a = w.held(S_A, "stopped lane", &["src/a.rs"]);
    w.end_lease(&a, THREE_HOURS);
    w.go_silent(&a, THREE_HOURS);
    let done = result(&w.sweep(&["--by", "leader"]));
    assert_eq!(done["swept"].as_array().unwrap().len(), 1, "{done}");
    let row = &done["swept"][0];
    assert_eq!(row["name"], json!(a), "{done}");
    assert_eq!(row["via"], json!("heartbeat"), "{done}");
    let silent = row["silent_secs"].as_i64().unwrap();
    assert!(
        (THREE_HOURS..THREE_HOURS + 60).contains(&silent),
        "silent_secs {silent}: {done}"
    );
    assert!(row["last_seen_at"].as_i64().unwrap() > 0, "{done}");
    assert_eq!(w.state_of(&a), "ready");
    let event = w.event_payload(&a, "card.sweep");
    assert_eq!(event["via"], json!("heartbeat"), "{event}");
    assert_eq!(event["silent_secs"], row["silent_secs"], "{event}");
    assert_eq!(event["last_seen_at"], row["last_seen_at"], "{event}");
    let notes = w.notes_of(&a);
    assert_eq!(notes.len(), 1, "{notes:?}");
    assert!(notes[0].starts_with("SWEPT "), "{notes:?}");
    assert!(notes[0].contains("180 minute(s)"), "{notes:?}");
}

#[test]
fn silent_past_the_grace_keeps_the_grace_as_its_reason() {
    let w = world();
    let a = w.held(S_A, "stopped long ago", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    w.go_silent(&a, TWENTY_FIVE_HOURS);
    let done = result(&w.sweep(&["--by", "leader"]));
    assert_eq!(done["swept"][0]["via"], json!("grace"), "{done}");
    assert_eq!(
        w.event_payload(&a, "card.sweep")["via"],
        json!("grace"),
        "the grace ran out, so the grace is the reason whatever the heartbeat says"
    );
}

#[test]
fn silent_holder_is_listed_by_a_dry_run_and_nothing_is_written() {
    let w = world();
    let a = w.held(S_A, "stopped lane", &["src/a.rs"]);
    w.end_lease(&a, THREE_HOURS);
    w.go_silent(&a, THREE_HOURS);
    let before = w.rows();
    let done = result(&w.sweep(&["--by", "leader", "--dry-run"]));
    assert_eq!(done["swept"][0]["name"], json!(a), "{done}");
    assert_eq!(done["swept"][0]["via"], json!("heartbeat"), "{done}");
    assert_eq!(done["writes"], json!(false), "{done}");
    assert_eq!(w.rows(), before);
    assert_eq!(w.state_of(&a), "active");
}

// ---- REQ-12: the silence window is two hours ------------------------------------------------------

#[test]
fn silence_window_one_hour_fifty_nine_keeps_the_card_and_two_hours_sweeps_it() {
    let w = world();
    let short = w.held(S_A, "silent 1h59", &["src/a.rs"]);
    let full = w.held(S_B, "silent 2h", &["src/b.rs"]);
    for card in [&short, &full] {
        w.end_lease(card, THREE_HOURS);
    }
    w.go_silent(&short, 2 * HOUR - 60);
    w.go_silent(&full, 2 * HOUR);
    let done = result(&w.sweep(&["--by", "leader"]));
    let names: Vec<&Value> = done["swept"]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| &r["name"])
        .collect();
    assert_eq!(names, vec![&json!(full)], "{done}");
    assert_eq!(done["swept"][0]["via"], json!("heartbeat"), "{done}");
    assert_eq!(
        w.state_of(&short),
        "active",
        "1 hour 59 minutes keeps the card"
    );
    assert_eq!(w.state_of(&full), "ready", "two hours sweeps it");
}

// ---- AC-2: a live lease is never swept ------------------------------------------------------------

#[test]
fn live_lease_with_a_silent_holder_is_never_swept_nor_taken() {
    let w = world();
    let a = w.held(S_A, "slow lane", &["src/a.rs"]);
    w.live_lease(&a, HOUR);
    w.go_silent(&a, TEN_HOURS);
    let before = w.rows();
    let done = result(&w.sweep(&["--by", "leader"]));
    assert!(done["swept"].as_array().unwrap().is_empty(), "{done}");
    let done = result(&w.sweep(&["--by", "leader", "--dry-run"]));
    assert!(done["swept"].as_array().unwrap().is_empty(), "{done}");
    let named = result(&w.sweep(&[&a, "--by", "leader"]));
    assert!(named["swept"].as_array().unwrap().is_empty(), "{named}");
    assert_eq!(named["skipped"][0]["reason"], json!("no_ended_lease"));
    let message = refusal(&w.take(S_B, &a));
    assert!(!message.is_empty());
    assert_eq!(w.state_of(&a), "active");
    assert_eq!(w.rows(), before, "no row was written");
    assert_eq!(w.count_kind("card.sweep"), 0);
}

// ---- AC-3: a recent sign of life keeps the card ---------------------------------------------------

#[test]
fn heartbeat_a_recent_reservation_update_keeps_a_card_with_an_ended_lease() {
    let w = world();
    let a = w.held(S_A, "working lane", &["src/a.rs"]);
    w.end_lease(&a, THREE_HOURS);
    w.go_silent(&a, THREE_HOURS);
    w.touched(&a, FIVE_MINUTES);
    let before = w.rows();
    let done = result(&w.sweep(&["--by", "leader"]));
    assert!(done["swept"].as_array().unwrap().is_empty(), "{done}");
    assert_eq!(w.state_of(&a), "active");
    assert_eq!(w.rows(), before);
}

#[test]
fn heartbeat_a_recent_note_of_the_holder_keeps_the_card() {
    let w = world();
    let a = w.held(S_A, "working lane", &["src/a.rs"]);
    w.end_lease(&a, THREE_HOURS);
    w.go_silent(&a, THREE_HOURS);
    w.note_as(S_A, &a, "RESUME still working");
    let listed = result(&w.sweep(&["--by", "leader", "--dry-run"]));
    assert!(
        listed["swept"].as_array().unwrap().is_empty(),
        "a dry run agrees with the real run: {listed}"
    );
    let done = result(&w.sweep(&["--by", "leader"]));
    assert!(done["swept"].as_array().unwrap().is_empty(), "{done}");
    assert_eq!(
        done["skipped"][0]["reason"],
        json!("within_grace"),
        "{done}"
    );
    assert_eq!(w.state_of(&a), "active");
}

#[test]
fn heartbeat_a_note_of_another_session_does_not_count() {
    let w = world();
    let a = w.held(S_A, "stopped lane", &["src/a.rs"]);
    w.end_lease(&a, THREE_HOURS);
    w.go_silent(&a, THREE_HOURS);
    w.note_as(S_B, &a, "a question for the holder");
    let done = result(&w.sweep(&["--by", "leader"]));
    assert_eq!(done["swept"][0]["name"], json!(a), "{done}");
    assert_eq!(w.state_of(&a), "ready");
}

#[test]
fn listed_then_the_holder_writes_keeps_the_card() {
    let w = world();
    let a = w.held(S_A, "stopped then back", &["src/a.rs"]);
    w.end_lease(&a, THREE_HOURS);
    w.go_silent(&a, THREE_HOURS);
    let mut l = ledger::open_write(w.home.path(), "demo").unwrap();
    let listed = sweep::sweep_all(&mut l, GRACE_SECS, "leader", None, true).unwrap();
    assert_eq!(listed.swept.len(), 1, "{listed:?}");
    let lease_until = listed.swept[0].lease_until;
    drop(l);
    // The holder writes between the listing and the sweep.
    w.note_as(S_A, &a, "RESUME back at work");
    let mut l = ledger::open_write(w.home.path(), "demo").unwrap();
    let out = sweep::sweep_card(
        &mut l,
        World::id(&a),
        Some(lease_until),
        GRACE_SECS,
        "leader",
        None,
    )
    .unwrap();
    assert!(
        matches!(out, Err(sweep::SkipReason::WithinGrace { .. })),
        "{out:?}"
    );
    drop(l);
    assert_eq!(w.state_of(&a), "active");
    assert_eq!(w.count_kind("card.sweep"), 0);
}

// ---- AC-5: the take that sweeps a silent card -----------------------------------------------------

#[test]
fn take_sweeps_a_silent_card_with_an_ended_lease_and_takes_it() {
    let w = world();
    let a = w.held(S_A, "stopped lane", &["src/a.rs"]);
    w.end_lease(&a, THREE_HOURS);
    w.go_silent(&a, THREE_HOURS);
    let taken = result(&w.take(S_B, &a));
    assert_eq!(taken["reservation"]["generation"], json!(2), "{taken}");
    assert_eq!(taken["reservation"]["holder"], json!(S_B), "{taken}");
    assert_eq!(w.state_of(&a), "active");
    assert_eq!(w.count_kind("card.sweep"), 1);
    assert_eq!(w.count_kind("reserve"), 2);
    assert_eq!(w.event_payload(&a, "card.sweep")["via"], json!("heartbeat"));
}

#[test]
fn take_refuses_an_ended_lease_with_a_fresh_heartbeat_and_writes_nothing() {
    let w = world();
    let a = w.held(S_A, "working lane", &["src/a.rs"]);
    w.end_lease(&a, THREE_HOURS);
    let before = w.rows();
    let message = refusal(&w.take(S_B, &a));
    assert!(!message.is_empty());
    assert_eq!(w.rows(), before, "nothing is written");
    assert_eq!(w.state_of(&a), "active");
    assert_eq!(w.count_kind("card.sweep"), 0);
}

// ---- AC-6: the swept holder -----------------------------------------------------------------------

#[test]
fn swept_holder_is_refused_with_the_old_generation_after_a_silence_sweep() {
    let w = world();
    let a = w.held(S_A, "stopped lane", &["src/a.rs"]);
    w.end_lease(&a, THREE_HOURS);
    w.go_silent(&a, THREE_HOURS);
    result(&w.sweep(&["--by", "leader"]));
    for args in [
        vec!["renew", a.as_str(), "--generation", "1"],
        vec!["release", a.as_str(), "--generation", "1"],
    ] {
        let message = refusal(&w.card(Some(S_A), &args));
        assert!(!message.is_empty(), "{args:?} must be refused");
    }
    let checked = w.card(
        Some(S_A),
        &["check", &a, "--generation", "1", "--path", "src/a.rs"],
    );
    let allowed =
        checked.status.code() == Some(0) && envelope(&checked)["result"]["allowed"] == json!(true);
    assert!(!allowed, "the old holder may not write after a sweep");
    let taken = result(&w.take(S_B, &a));
    assert_eq!(taken["reservation"]["generation"], json!(2), "{taken}");
}

// ---- AC-7: what the silence sweep leaves alone ----------------------------------------------------

#[test]
fn not_swept_a_card_in_review_nor_an_unconfirmed_reservation() {
    let w = world();
    let review = w.held(S_A, "in review", &["src/a.rs"]);
    result(&w.card(
        Some(S_A),
        &["review", &review, "--generation", "1", "--keep", "src/a.rs"],
    ));
    let unconfirmed = w.held(S_B, "unconfirmed", &["src/b.rs"]);
    let candidate = w.held(S_B, "candidate", &["src/c.rs"]);
    for card in [&review, &unconfirmed, &candidate] {
        w.go_silent(card, THREE_HOURS);
    }
    w.end_lease(&review, THREE_HOURS);
    w.set_status(&unconfirmed, "held_for_review", THREE_HOURS);
    w.set_status(&candidate, "expired_candidate", THREE_HOURS);
    let before = w.rows();
    let done = result(&w.sweep(&["--by", "leader"]));
    assert!(done["swept"].as_array().unwrap().is_empty(), "{done}");
    let reasons: Vec<(String, String)> = done["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| {
            (
                s["name"].as_str().unwrap().to_string(),
                s["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    assert!(
        reasons.contains(&(review.clone(), "in_review".into())),
        "{reasons:?}"
    );
    assert!(
        reasons.contains(&(unconfirmed.clone(), "not_held".into())),
        "{reasons:?}"
    );
    assert!(
        reasons.contains(&(candidate.clone(), "not_held".into())),
        "{reasons:?}"
    );
    assert_eq!(w.rows(), before, "no row was written");
}
