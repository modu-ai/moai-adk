//! `moai card offer` and `moai card accept` (SPEC-20261126-card-offer-accept): the holder offers a
//! card to one named session, nothing moves until that session accepts, and every situation in
//! which the offer is no longer good is refused with the ledger exactly as it was.
//!
//! Every command runs against a temporary HOME; the real home state is not read or written. The
//! session is the id in the environment (`CLAUDE_CODE_SESSION_ID`). No test sleeps: a lease is
//! ended by the same `UPDATE reservation` the reclaim tests use, and an offer whose validity has
//! run out is a second `card.offer` row copied from the real one with `valid_until` in the past.
#![cfg(unix)]

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
/// The holder, the target, a stranger and a second target.
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const S_C: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_D: &str = "dddddddd-0000-4000-8000-00000000000d";
const SCOPE: &str = "src/offer/";
const BRANCH: &str = "lane-a";
/// The tables a refusal must leave untouched.
const TABLES: [&str; 7] = [
    "card",
    "card_ac",
    "card_note",
    "card_link",
    "reservation",
    "reservation_scope",
    "event",
];

struct World {
    home: TempHome,
    _host: tempfile::TempDir,
    dir: PathBuf,
    project: String,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    let host = tempfile::Builder::new()
        .prefix("moai-test-host-")
        .tempdir()
        .unwrap();
    let dir = fs::canonicalize(host.path()).unwrap();
    World {
        home,
        _host: host,
        dir,
        project: project.to_str().unwrap().to_string(),
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

fn card_number(card: &str) -> i64 {
    card.trim_start_matches("C-").parse().unwrap()
}

impl World {
    /// `moai card <args> --project <project>` as the host session `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        // Not the cargo working folder: a take without `--worktree` records the git worktree of
        // the folder it runs in.
        cmd.current_dir(&self.dir);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    /// A ready card with one criterion taken by `holder` on the branch `lane-a`, with no worktree
    /// (a reclaim reads the recorded worktree with git, and the reclaim tests record none); its
    /// name.
    fn taken(&self, holder: &str) -> String {
        self.taken_with(holder, &[])
    }

    /// The same, with extra options of `card take`.
    fn taken_with(&self, holder: &str, extra: &[&str]) -> String {
        let added = result(&self.card(
            None,
            &["add", "--title", "t", "--scope", SCOPE, "--ac", "one"],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        let mut args = vec!["take", &name, "--branch", BRANCH];
        args.extend_from_slice(extra);
        result(&self.card(Some(holder), &args));
        name
    }

    fn offer(&self, session: &str, card: &str, generation: &str, to: &str) -> Output {
        self.card(
            Some(session),
            &["offer", card, "--generation", generation, "--to", to],
        )
    }

    fn offer_with(&self, session: &str, card: &str, generation: &str, extra: &[&str]) -> Output {
        let mut args = vec!["offer", card, "--generation", generation];
        args.extend_from_slice(extra);
        self.card(Some(session), &args)
    }

    fn cancel(&self, session: &str, card: &str, generation: &str) -> Output {
        self.offer_with(session, card, generation, &["--cancel"])
    }

    fn accept(&self, session: &str, card: &str, generation: &str) -> Output {
        self.card(Some(session), &["accept", card, "--generation", generation])
    }

    fn handoff(&self, session: &str, card: &str, generation: &str, to: &str) -> Output {
        self.card(
            Some(session),
            &["handoff", card, "--generation", generation, "--to", to],
        )
    }

    fn reclaim(&self, session: &str, card: &str, generation: &str) -> Output {
        self.card(
            Some(session),
            &[
                "reclaim",
                card,
                "--generation",
                generation,
                "--by",
                "leader",
                "--reason",
                "the holder is gone",
            ],
        )
    }

    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
    }

    fn end_lease(&self, card: &str) {
        self.sql(&format!(
            "UPDATE reservation SET lease_until = unixepoch() - 5 \
             WHERE card_id = {} AND status = 'held'",
            card_number(card)
        ));
    }

    /// Makes the newest `card.offer` of the card run out: a newer row copied from it with
    /// `valid_until` five seconds in the past (the event table is append-only).
    fn expire_newest_offer(&self, card: &str) {
        self.sql(&format!(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             SELECT event_key || '-ran-out', unixepoch(), actor, card_id, kind, \
                    json_set(payload, '$.valid_until', unixepoch() - 5) \
             FROM event WHERE card_id = {} AND kind = 'card.offer' \
             ORDER BY event_id DESC LIMIT 1",
            card_number(card)
        ));
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

    fn now(&self) -> i64 {
        self.rows("SELECT unixepoch()")[0][0].as_i64().unwrap()
    }

    fn dump(&self) -> Vec<String> {
        self.dump_of(&TABLES)
    }

    fn dump_of(&self, tables: &[&str]) -> Vec<String> {
        let mut out = Vec::new();
        for table in tables {
            for r in self.rows(&format!("SELECT * FROM {table} ORDER BY 1, 2")) {
                out.push(format!("{table}:{r}"));
            }
        }
        out
    }

    /// The payloads of the card's events of `kind`, oldest first.
    fn events_of(&self, kind: &str) -> Vec<Value> {
        self.rows(&format!(
            "SELECT payload FROM event WHERE kind = '{kind}' ORDER BY event_id"
        ))
        .iter()
        .map(|r| serde_json::from_str(r[0].as_str().unwrap()).unwrap())
        .collect()
    }

    fn count(&self, kind: &str) -> usize {
        self.rows(&format!("SELECT 1 FROM event WHERE kind = '{kind}'"))
            .len()
    }

    /// The invariant of one owner: at most one `held` reservation per card, and when there is
    /// one its generation is the card's current generation. Returns the holder of the `held`
    /// reservation, if any.
    fn one_owner(&self, card: &str) -> Option<String> {
        let n = card_number(card);
        let held = self.rows(&format!(
            "SELECT holder, generation FROM reservation WHERE card_id = {n} AND status = 'held'"
        ));
        assert!(
            held.len() <= 1,
            "{card} has {} held reservations",
            held.len()
        );
        let current = self.rows(&format!("SELECT generation FROM card WHERE card_id = {n}"))[0][0]
            .as_i64()
            .unwrap();
        held.first().map(|r| {
            assert_eq!(r[1].as_i64().unwrap(), current, "{card}: a stale held row");
            r[0].as_str().unwrap().to_string()
        })
    }

    /// A refusal of kind `conflict` with the word `code` in the message and the ledger exactly
    /// as `before`.
    fn refused(&self, label: &str, out: &Output, code: &str, before: &[String]) -> String {
        let (kind, message) = refusal(out);
        assert_eq!(kind, "conflict", "{label}: {message}");
        assert!(message.contains(code), "{label}: {code}: {message}");
        assert_eq!(&self.dump(), before, "{label}: a refusal wrote something");
        message
    }
}

/// AC-1 (REQ-1, REQ-5, REQ-7): an offer moves nothing; the target's accept in time releases the
/// old reservation and gives the target the next generation with the same scope, worktree and
/// branch and a new lease; one event of each kind, no reclaim.
#[test]
fn an_offer_in_time_moves_the_card_to_the_target_on_accept() {
    let w = world();
    let wt = w.dir.join("wt");
    fs::create_dir_all(&wt).unwrap();
    let card = w.taken_with(S_A, &["--worktree", wt.to_str().unwrap()]);
    let n = card_number(&card);
    let moving = ["card", "reservation", "reservation_scope"];
    let before = w.dump_of(&moving);
    let now = w.now();

    let offered = result(&w.offer_with(
        S_A,
        &card,
        "1",
        &["--to", S_B, "--handover", "pushed; next: criterion 2"],
    ));
    assert_eq!(offered["action"], json!("offer"));
    assert_eq!(offered["generation"], json!(1));
    assert_eq!(offered["to"], json!(S_B));
    let valid_until = offered["valid_until"].as_i64().unwrap();
    assert!(
        (now + 1_790..=now + 1_810).contains(&valid_until),
        "1800 seconds by default, got {}",
        valid_until - now
    );
    assert_eq!(w.dump_of(&moving), before, "an offer moves nothing");
    let events = w.events_of("card.offer");
    assert_eq!(events.len(), 1);
    assert_eq!(events[0]["holder"], json!(S_A));
    assert_eq!(events[0]["generation"], json!(1));
    assert_eq!(events[0]["to"], json!(S_B));
    assert_eq!(events[0]["valid_until"], json!(valid_until));
    assert_eq!(events[0]["handover"], json!("pushed; next: criterion 2"));
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_A));

    let got = result(&w.accept(S_B, &card, "1"));
    assert_eq!(got["action"], json!("accept"));
    let new = &got["reservation"];
    assert_eq!(new["generation"], json!(2));
    assert_eq!(new["holder"], json!(S_B));
    assert_eq!(new["status"], json!("held"));
    assert_eq!(new["write_scope"], json!([SCOPE]));
    assert_eq!(new["branch"], json!(BRANCH));
    assert_eq!(new["worktree"], json!(wt.to_str().unwrap()));
    let lease = new["lease_until"].as_i64().unwrap() - now;
    assert!(
        (14_390..=14_420).contains(&lease),
        "a new four hour lease, got {lease}"
    );
    assert_eq!(got["previous"]["generation"], json!(1));
    assert_eq!(got["previous"]["status"], json!("released"));
    assert_eq!(got["events"], json!(1));
    assert_eq!(got["writes"], json!(true));
    let accepted = w.events_of("card.accept");
    assert_eq!(accepted.len(), 1);
    assert_eq!(accepted[0]["from_holder"], json!(S_A));
    assert_eq!(accepted[0]["from_generation"], json!(1));
    assert_eq!(accepted[0]["holder"], json!(S_B));
    assert_eq!(accepted[0]["generation"], json!(2));
    assert_eq!(w.count("card.reclaim"), 0);
    assert_eq!(
        w.rows(&format!("SELECT state FROM card WHERE card_id = {n}")),
        vec![json!(["active"])]
    );
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_B));

    // The old holder is fenced out, the new one writes.
    let old = w.card(
        Some(S_A),
        &[
            "check",
            &card,
            "--generation",
            "1",
            "--path",
            "src/offer/x.rs",
        ],
    );
    let (kind, message) = refusal(&old);
    assert_eq!(kind, "conflict", "{message}");
    result(&w.card(
        Some(S_B),
        &[
            "check",
            &card,
            "--generation",
            "2",
            "--path",
            "src/offer/x.rs",
        ],
    ));
}

/// AC-2 (REQ-6): an offer whose validity ran out is refused on accept and the holder keeps the
/// card.
#[test]
fn an_expired_offer_is_refused_and_the_holder_keeps_the_card() {
    let w = world();
    let card = w.taken(S_A);
    result(&w.offer(S_A, &card, "1", S_B));
    w.expire_newest_offer(&card);
    let before = w.dump();
    let out = w.accept(S_B, &card, "1");
    w.refused("expired", &out, "offer_expired", &before);
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_A));
    assert_eq!(w.count("card.accept"), 0);
}

/// AC-3 (REQ-6): the holder's lease ended after the offer: the accept is refused and the reclaim
/// road is open for the same card.
#[test]
fn an_accept_after_the_holders_lease_ended_is_refused_and_reclaim_still_works() {
    let w = world();
    let card = w.taken(S_A);
    result(&w.offer(S_A, &card, "1", S_B));
    w.end_lease(&card);
    let before = w.dump();
    let out = w.accept(S_B, &card, "1");
    w.refused("lease ended", &out, "lease_expired", &before);
    let got = result(&w.reclaim(S_C, &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
    assert_eq!(got["reservation"]["holder"], json!(S_C));
    assert_eq!(w.count("card.accept"), 0);
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_C));
}

/// AC-4 (REQ-4, REQ-6, REQ-9): a cancelled offer is refused on accept; a cancel with nothing to
/// cancel is refused with `no_offer`.
#[test]
fn a_cancelled_offer_is_refused_on_accept() {
    let w = world();
    let card = w.taken(S_A);
    let before = w.dump();
    let out = w.cancel(S_A, &card, "1");
    w.refused("cancel without an offer", &out, "no_offer", &before);

    result(&w.offer(S_A, &card, "1", S_B));
    let cancelled = result(&w.cancel(S_A, &card, "1"));
    assert_eq!(cancelled["action"], json!("offer-cancel"));
    assert_eq!(w.events_of("card.offer_cancel").len(), 1);
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_A));

    let before = w.dump();
    let out = w.accept(S_B, &card, "1");
    w.refused("cancelled", &out, "offer_cancelled", &before);
    let out = w.cancel(S_A, &card, "1");
    w.refused("second cancel", &out, "no_offer", &before);
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_A));
}

/// AC-5 (REQ-6): the generation moved after the offer: the old generation is stale, the new one
/// has no offer, and the new holder stays.
#[test]
fn an_offer_does_not_survive_a_generation_change() {
    let w = world();
    let card = w.taken(S_A);
    result(&w.offer(S_A, &card, "1", S_B));
    result(&w.handoff(S_A, &card, "1", S_C));
    let before = w.dump();
    let out = w.accept(S_B, &card, "1");
    let message = w.refused("old generation", &out, "stale", &before);
    assert!(message.contains("generation 1 is stale"), "{message}");
    let out = w.accept(S_B, &card, "2");
    w.refused("new generation", &out, "no_offer", &before);
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_C));
}

/// AC-6 (REQ-5, REQ-6): the same target accepting twice: the second is stale and the first
/// result stays.
#[test]
fn a_second_accept_by_the_same_target_is_refused_as_stale() {
    let w = world();
    let card = w.taken(S_A);
    result(&w.offer(S_A, &card, "1", S_B));
    result(&w.accept(S_B, &card, "1"));
    let before = w.dump();
    let out = w.accept(S_B, &card, "1");
    let message = w.refused("second accept", &out, "stale", &before);
    assert!(message.contains("generation 1 is stale"), "{message}");
    let out = w.accept(S_B, &card, "2");
    w.refused("accept at the new generation", &out, "no_offer", &before);
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_B));
    assert_eq!(w.count("card.accept"), 1);
}

/// AC-7 (REQ-6): a session that is not the target is refused.
#[test]
fn a_session_that_is_not_the_target_is_refused() {
    let w = world();
    let card = w.taken(S_A);
    result(&w.offer(S_A, &card, "1", S_B));
    let before = w.dump();
    let out = w.accept(S_C, &card, "1");
    w.refused("stranger", &out, "not_target", &before);
    // Not even the holder accepts its own offer.
    let out = w.accept(S_A, &card, "1");
    w.refused("the holder itself", &out, "not_target", &before);
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_A));
}

/// The leader's condition: with no offer no session takes another session's reservation by
/// accepting.
#[test]
fn a_session_cannot_take_a_card_by_accepting_without_an_offer() {
    let w = world();
    let card = w.taken(S_A);
    let before = w.dump();
    for session in [S_A, S_B, S_C] {
        let out = w.accept(session, &card, "1");
        w.refused("no offer", &out, "no_offer", &before);
    }
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_A));
    assert_eq!(w.count("card.accept"), 0);
}

/// AC-8 (REQ-3, REQ-6): offering again makes the newest offer the only valid one.
#[test]
fn a_new_offer_replaces_the_earlier_one() {
    let w = world();
    let card = w.taken(S_A);
    result(&w.offer(S_A, &card, "1", S_B));
    result(&w.offer(S_A, &card, "1", S_D));
    assert_eq!(w.events_of("card.offer").len(), 2);
    let before = w.dump();
    let out = w.accept(S_B, &card, "1");
    w.refused("the first target", &out, "not_target", &before);
    let got = result(&w.accept(S_D, &card, "1"));
    assert_eq!(got["reservation"]["holder"], json!(S_D));
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_D));
}

/// AC-9 (REQ-8): while an offer waits, take, reclaim and adopt answer for the card as they do
/// without an offer, and the offer is still good afterwards.
#[test]
fn an_open_offer_does_not_change_take_reclaim_or_adopt() {
    let w = world();
    let card = w.taken(S_A);
    let answers = |w: &World| -> Vec<(String, String)> {
        let take = refusal(&w.card(Some(S_C), &["take", &card]));
        let reclaim = refusal(&w.reclaim(S_C, &card, "1"));
        let adopt = refusal(&w.card(Some(S_C), &["adopt", &card, "--generation", "1"]));
        vec![take, reclaim, adopt]
    };
    let without = answers(&w);
    let before = w.dump();
    result(&w.offer(S_A, &card, "1", S_B));
    let with = answers(&w);
    assert_eq!(with, without, "an offer changes the answers");
    // The three refusals wrote nothing except what the offer itself wrote.
    let offered = w.dump();
    assert_eq!(
        offered.iter().filter(|r| !before.contains(r)).count(),
        offered.len() - before.len(),
        "a refusal changed an old row"
    );
    assert_eq!(w.count("card.offer"), 1);
    let got = result(&w.accept(S_B, &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
    assert_eq!(w.one_owner(&card).as_deref(), Some(S_B));
}

/// AC-10 (REQ-2): the fence refuses the offers it should: the reservation, the generation and
/// the offer count stay; a refused holder call is recorded once as `fence_reject`, an input
/// that is not a session id is refused before anything is read.
#[test]
fn an_offer_that_the_fence_refuses_changes_no_reservation() {
    let w = world();
    let card = w.taken(S_A);
    let moving = ["card", "reservation", "reservation_scope"];
    let before = w.dump_of(&moving);

    // A stranger, a generation never issued, itself, and text that is not a session id.
    let out = w.offer(S_C, &card, "1", S_B);
    assert_eq!(refusal(&out).0, "conflict");
    assert_eq!(
        w.count("fence_reject"),
        1,
        "the one record of a refused holder call"
    );
    let out = w.offer(S_A, &card, "7", S_B);
    assert_eq!(refusal(&out).0, "conflict");
    let out = w.offer(S_A, &card, "1", S_A);
    let (kind, message) = refusal(&out);
    assert_eq!(kind, "invalid_input", "{message}");
    let out = w.offer(S_A, &card, "1", "not-a-session!");
    assert_eq!(refusal(&out).0, "invalid_input");
    assert_eq!(w.dump_of(&moving), before);
    assert_eq!(w.count("card.offer"), 0);

    // An ended lease: the holder can no longer offer.
    w.end_lease(&card);
    let ended = w.dump_of(&moving);
    let out = w.offer(S_A, &card, "1", S_B);
    assert_eq!(refusal(&out).0, "conflict");
    assert_eq!(w.dump_of(&moving), ended);
    assert_eq!(w.count("card.offer"), 0);

    // After a handoff the old generation is stale for an offer too.
    let w2 = world();
    let card2 = w2.taken(S_A);
    result(&w2.handoff(S_A, &card2, "1", S_C));
    let out = w2.offer(S_A, &card2, "1", S_B);
    assert_eq!(refusal(&out).0, "conflict");
    assert_eq!(w2.count("card.offer"), 0);
    assert_eq!(w2.one_owner(&card2).as_deref(), Some(S_C));
}

/// AC-11 (REQ-7): offers, accepts and cancels write no reclaim and use none of the allowance:
/// three reclaims of a dead holder still go through and the fourth is the limit.
#[test]
fn offer_and_accept_leave_the_reclaim_allowance_whole() {
    let w = world();
    let card = w.taken(S_A);
    result(&w.offer(S_A, &card, "1", S_B));
    result(&w.accept(S_B, &card, "1"));
    result(&w.offer(S_B, &card, "2", S_C));
    result(&w.cancel(S_B, &card, "2"));
    result(&w.offer(S_B, &card, "2", S_A));
    result(&w.accept(S_A, &card, "2"));
    assert_eq!(w.count("card.reclaim"), 0);

    let mut generation = 3;
    for session in [S_B, S_C, S_D] {
        w.end_lease(&card);
        let got = result(&w.reclaim(session, &card, &generation.to_string()));
        generation += 1;
        assert_eq!(got["reservation"]["generation"], json!(generation));
    }
    w.end_lease(&card);
    let out = w.reclaim(S_A, &card, &generation.to_string());
    let (kind, message) = refusal(&out);
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("reclaim limit"), "{message}");
    assert_eq!(w.count("card.reclaim"), 3);
}

/// AC-12 (REQ-5, REQ-8): a long mixed run of every road — offer, cancel, offer again, accept,
/// handoff, an offer that ran out, an ended lease, a reclaim — keeps one owner at every step.
#[test]
fn every_offer_scenario_ends_with_one_owner() {
    let w = world();
    let card = w.taken(S_A);
    let owner = |w: &World, expect: &str| {
        assert_eq!(w.one_owner(&card).as_deref(), Some(expect));
    };
    owner(&w, S_A);
    result(&w.offer(S_A, &card, "1", S_B));
    owner(&w, S_A);
    result(&w.cancel(S_A, &card, "1"));
    owner(&w, S_A);
    result(&w.offer(S_A, &card, "1", S_B));
    result(&w.accept(S_B, &card, "1"));
    owner(&w, S_B);
    result(&w.handoff(S_B, &card, "2", S_C));
    owner(&w, S_C);
    result(&w.offer(S_C, &card, "3", S_D));
    w.expire_newest_offer(&card);
    let _ = w.accept(S_D, &card, "3");
    owner(&w, S_C);
    result(&w.offer(S_C, &card, "3", S_D));
    w.end_lease(&card);
    let _ = w.accept(S_D, &card, "3");
    // A lapsed lease keeps its row `held` until somebody reclaims it, and it is still one owner.
    owner(&w, S_C);
    result(&w.reclaim(S_A, &card, "3"));
    owner(&w, S_A);
}
