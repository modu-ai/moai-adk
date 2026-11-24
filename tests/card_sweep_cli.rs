//! `moai card sweep` and the take-time sweep (SPEC-20261124-ended-lease-sweep): a lease that ended
//! long ago returns its card to `ready` without a human hand, and the paths of such a card are
//! free for a take and a line lease.
//!
//! Every command runs against a temporary HOME; the real home ledger is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const THREE_DAYS: i64 = 3 * 24 * 3600;

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
    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        // The project goes before a `--`: what follows it is the command `line-lease` runs.
        let cut = args.iter().position(|a| *a == "--").unwrap_or(args.len());
        all.extend_from_slice(&args[..cut]);
        all.extend_from_slice(&["--project", &self.project]);
        all.extend_from_slice(&args[cut..]);
        let mut cmd = self.home.command(&all);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    /// A `ready` card with one open criterion and `scope`.
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

    /// `card take` as `session`, the worktree and branch the way a lane gives them.
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

    fn scope_of(&self, card: &str) -> Vec<String> {
        result(&self.card(None, &["show", card]))["reservation"]["write_scope"]
            .as_array()
            .unwrap()
            .iter()
            .map(|v| v.as_str().unwrap().to_string())
            .collect()
    }

    /// The lease of the held reservation of `card` ended `secs_ago` seconds ago.
    fn end_lease(&self, card: &str, secs_ago: i64) {
        let id: i64 = card.trim_start_matches("C-").parse().unwrap();
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(&format!(
                "UPDATE reservation SET lease_until = unixepoch() - {secs_ago} \
                 WHERE card_id = {id} AND status = 'held'"
            ))?;
            Ok(())
        })
        .unwrap();
    }

    /// The held reservation of `card` becomes `status` (what `recover` leaves for a row it could
    /// not confirm), its lease ended `secs_ago` seconds ago.
    fn set_status(&self, card: &str, status: &str, secs_ago: i64) {
        let id: i64 = card.trim_start_matches("C-").parse().unwrap();
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(&format!(
                "UPDATE reservation SET status = '{status}', lease_until = unixepoch() - {secs_ago} \
                 WHERE card_id = {id} AND status = 'held'"
            ))?;
            Ok(())
        })
        .unwrap();
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

/// How many rows of `table` the temporary ledger holds.
fn count(w: &World, table: &str) -> i64 {
    let l = ledger::open_read_only(w.home.path(), "demo").unwrap();
    let conn: &Connection = l.conn();
    conn.query_row(&format!("SELECT count(*) FROM {table}"), [], |r| r.get(0))
        .unwrap()
}

const TWENTY_FIVE_HOURS: i64 = 25 * 3600;
const TWENTY_THREE_HOURS: i64 = 23 * 3600;

impl World {
    /// `moai card next <project>` as `session`.
    fn next(&self, session: &str) -> Value {
        let mut cmd = self.home.command(&["card", "next", &self.project]);
        cmd.env(SESSION_VAR, session);
        cmd.stdin(Stdio::null());
        result(&cmd.output().expect("run moai"))
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

    /// The text of every note on `card`, oldest first.
    fn notes_of(&self, card: &str) -> Vec<String> {
        result(&self.card(None, &["show", card]))["notes"]
            .as_array()
            .unwrap()
            .iter()
            .map(|n| n["body"].as_str().unwrap().to_string())
            .collect()
    }

    /// Writes a complete profile with `extra` spliced in, as the stale tests do.
    fn write_profile(&self, extra: &str) {
        let project = std::path::Path::new(&self.project);
        let cfg = moai_cli::state_root::path(project, "project.yaml");
        std::fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        std::fs::write(
            &cfg,
            format!(
                "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]{extra}}}"
            ),
        )
        .unwrap();
    }
}

// ---- AC-4: the dry run and the leader rule -------------------------------------------------------

#[test]
fn dry_run_lists_the_cards_and_writes_nothing() {
    let w = world();
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    let events = count(&w, "event");
    let notes = count(&w, "card_note");
    let done = result(&w.sweep(&["--by", "leader", "--dry-run"]));
    assert_eq!(done["dry_run"], json!(true), "{done}");
    assert_eq!(done["swept"][0]["name"], json!(a), "{done}");
    assert_eq!(done["writes"], json!(false), "{done}");
    assert_eq!(
        done["grace_source"],
        json!("default_unreadable_profile"),
        "the fixture profile is incomplete, so the default holds and the answer says why: {done}"
    );
    assert_eq!(w.state_of(&a), "active");
    assert_eq!(count(&w, "event"), events);
    assert_eq!(count(&w, "card_note"), notes);
}

#[test]
fn refuses_a_sweep_that_is_not_by_the_leader_and_writes_nothing() {
    let w = world();
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    let events = count(&w, "event");
    let (_, message) = refusal(&w.sweep(&["--by", "somebody"]));
    assert!(message.contains("--by leader"), "{message}");
    assert_eq!(w.state_of(&a), "active");
    assert_eq!(count(&w, "event"), events, "nothing is written");
}

#[test]
fn sweep_returns_the_cards_of_an_ended_lease_to_ready_with_one_note_each() {
    let w = world();
    let late = w.held(S_A, "zombie", &["src/a.rs"]);
    let inside = w.held(S_B, "recent", &["src/b.rs"]);
    w.end_lease(&late, TWENTY_FIVE_HOURS);
    w.end_lease(&inside, TWENTY_THREE_HOURS);
    let done = result(&w.sweep(&["--by", "leader"]));
    assert_eq!(done["swept"].as_array().unwrap().len(), 1, "{done}");
    assert_eq!(done["swept"][0]["name"], json!(late), "{done}");
    assert_eq!(done["skipped"][0]["name"], json!(inside), "{done}");
    assert_eq!(
        done["skipped"][0]["reason"],
        json!("within_grace"),
        "{done}"
    );
    assert_eq!(done["writes"], json!(true), "{done}");
    assert_eq!(w.state_of(&late), "ready");
    assert_eq!(w.state_of(&inside), "active");
    let notes = w.notes_of(&late);
    assert_eq!(notes.len(), 1, "{notes:?}");
    assert!(notes[0].starts_with("SWEPT "), "{notes:?}");
}

#[test]
fn sweep_of_one_named_card_leaves_the_others() {
    let w = world();
    let one = w.held(S_A, "zombie one", &["src/a.rs"]);
    let two = w.held(S_B, "zombie two", &["src/b.rs"]);
    w.end_lease(&one, TWENTY_FIVE_HOURS);
    w.end_lease(&two, TWENTY_FIVE_HOURS);
    let done = result(&w.sweep(&[&one, "--by", "leader"]));
    assert_eq!(done["swept"].as_array().unwrap().len(), 1, "{done}");
    assert_eq!(w.state_of(&one), "ready");
    assert_eq!(w.state_of(&two), "active");
    // A card with a live lease is named, not listed: the answer says it has no ended lease.
    let live = w.held(S_B, "live", &["src/c.rs"]);
    let done = result(&w.sweep(&[&live, "--by", "leader"]));
    assert_eq!(
        done["skipped"][0]["reason"],
        json!("no_ended_lease"),
        "{done}"
    );
    assert_eq!(w.state_of(&live), "active");
}

#[test]
fn the_grace_period_comes_from_the_profile_key() {
    let w = world();
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    w.write_profile(", \"lease_sweep_hours\": 48");
    let done = result(&w.sweep(&["--by", "leader", "--dry-run"]));
    assert_eq!(done["grace_hours"], json!(48), "{done}");
    assert_eq!(done["grace_source"], json!("profile"), "{done}");
    assert!(done["swept"].as_array().unwrap().is_empty(), "{done}");
    w.write_profile(", \"lease_sweep_hours\": 24");
    let done = result(&w.sweep(&["--by", "leader", "--dry-run"]));
    assert_eq!(done["swept"][0]["name"], json!(a), "{done}");
}

// ---- AC-5: the swept holder and the next lane ------------------------------------------------------

#[test]
fn swept_holder_is_refused_with_the_old_generation_and_a_second_lane_takes_the_card() {
    let w = world();
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    result(&w.sweep(&["--by", "leader"]));
    for args in [
        vec!["renew", a.as_str(), "--generation", "1"],
        vec!["release", a.as_str(), "--generation", "1"],
    ] {
        let (_, message) = refusal(&w.card(Some(S_A), &args));
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
    assert_eq!(taken["reservation"]["holder"], json!(S_B), "{taken}");
}

// ---- AC-6: the stale row names the sweep ------------------------------------------------------------

#[test]
fn stale_row_names_the_sweep_past_the_grace_and_the_withdraw_inside_it() {
    let w = world();
    let late = w.held(S_A, "zombie", &["src/a.rs"]);
    let inside = w.held(S_B, "recent", &["src/b.rs"]);
    w.end_lease(&late, TWENTY_FIVE_HOURS);
    w.end_lease(&inside, TWENTY_THREE_HOURS);
    let listing = w.next(S_B);
    let row = |card: &str| -> Value {
        listing["stale"]
            .as_array()
            .unwrap_or_else(|| panic!("`next` answers no `stale` list: {listing}"))
            .iter()
            .find(|e| e["name"] == json!(card) && e["kind"] == json!("lease_ended"))
            .unwrap_or_else(|| panic!("no lease_ended row for {card}: {listing}"))
            .clone()
    };
    assert_eq!(
        row(&late)["command"],
        json!(format!("moai card sweep {late} --by leader"))
    );
    assert_eq!(
        row(&inside)["command"],
        json!(format!(
            "moai card withdraw {inside} --reason <why> --by leader"
        ))
    );
}

#[test]
fn stale_row_keeps_the_withdraw_for_a_card_in_review_past_the_grace() {
    let w = world();
    let a = w.held(S_A, "zombie in review", &["src/a.rs"]);
    result(&w.card(
        Some(S_A),
        &["review", &a, "--generation", "1", "--keep", "src/a.rs"],
    ));
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    let listing = w.next(S_B);
    let row = listing["stale"]
        .as_array()
        .unwrap_or_else(|| panic!("`next` answers no `stale` list: {listing}"))
        .iter()
        .find(|e| e["name"] == json!(a) && e["kind"] == json!("lease_ended"))
        .unwrap_or_else(|| panic!("no lease_ended row for {a}: {listing}"))
        .clone();
    assert_eq!(
        row["command"],
        json!(format!("moai card withdraw {a} --reason <why> --by leader")),
        "a card in review is never swept, so its row keeps the withdraw"
    );
}

// ---- AC-7: the take that sweeps ---------------------------------------------------------------------

#[test]
fn take_sweeps_a_card_past_the_grace_and_takes_it_in_one_step() {
    let w = world();
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    let taken = result(&w.take(S_B, &a));
    assert_eq!(taken["reservation"]["generation"], json!(2), "{taken}");
    assert_eq!(taken["reservation"]["holder"], json!(S_B), "{taken}");
    assert_eq!(w.state_of(&a), "active");
    let sweeps = count_kind(&w, "card.sweep");
    assert_eq!(sweeps, 1, "one card.sweep event");
    assert_eq!(count_kind(&w, "reserve"), 2, "the first take and this one");
    let notes = w.notes_of(&a);
    assert!(notes.iter().any(|n| n.starts_with("SWEPT ")), "{notes:?}");
}

#[test]
fn take_refuses_a_card_inside_the_grace_period_as_before_and_writes_nothing() {
    let w = world();
    let a = w.held(S_A, "recent", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_THREE_HOURS);
    let events = count(&w, "event");
    let (kind, message) = refusal(&w.take(S_B, &a));
    assert_eq!(kind, "conflict", "{message}");
    assert_eq!(count(&w, "event"), events, "nothing is written");
    assert_eq!(count_kind(&w, "card.sweep"), 0);
}

#[test]
fn take_of_a_swept_card_refused_by_the_scope_check_leaves_the_card_ready() {
    let w = world();
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    // Another card holds the path the swept card's scope names, with a live lease.
    let blocker = w.held(S_B, "blocker", &["src/a.rs"]);
    assert_eq!(w.state_of(&blocker), "active");
    let _ = refusal(&w.take(S_B, &a));
    assert_eq!(
        w.state_of(&a),
        "ready",
        "the sweep stays: the card is ready, only the take was refused"
    );
}

fn count_kind(w: &World, kind: &str) -> i64 {
    let l = ledger::open_read_only(w.home.path(), "demo").unwrap();
    l.conn()
        .query_row("SELECT count(*) FROM event WHERE kind = ?1", [kind], |r| {
            r.get(0)
        })
        .unwrap()
}

// ---- AC-8: the paths of a card whose lease ended three days ago ----------------------------------

#[test]
fn ended_lease_paths_a_take_succeeds_over_a_held_reservation() {
    let w = world();
    let a = w.held(S_A, "zombie", &["src/lib.rs", "docs/a.md"]);
    w.end_lease(&a, THREE_DAYS);
    let b = w.ready("next", &["src/lib.rs"]);
    result(&w.take(S_B, &b));
    assert_eq!(w.scope_of(&b), ["src/lib.rs"]);
    assert_eq!(
        w.scope_of(&a),
        ["docs/a.md"],
        "the old reservation lost only the taken path"
    );
}

#[test]
fn ended_lease_paths_a_take_succeeds_over_a_card_in_review() {
    let w = world();
    let a = w.held(S_A, "zombie in review", &["src/lib.rs", "docs/a.md"]);
    result(&w.card(
        Some(S_A),
        &["review", &a, "--generation", "1", "--keep", "src/lib.rs"],
    ));
    w.end_lease(&a, THREE_DAYS);
    let b = w.ready("next", &["src/lib.rs"]);
    result(&w.take(S_B, &b));
    assert_eq!(w.scope_of(&b), ["src/lib.rs"]);
}

#[test]
fn ended_lease_paths_a_recovered_reservation_still_refuses() {
    let w = world();
    let a = w.held(S_A, "recovered", &["src/lib.rs"]);
    w.set_status(&a, "held_for_review", THREE_DAYS);
    let b = w.ready("next", &["src/lib.rs"]);
    let events = count(&w, "event");
    let (kind, message) = refusal(&w.take(S_B, &b));
    assert_eq!(kind, "conflict", "{message}");
    assert!(
        message.contains(&a) || message.contains("src/lib.rs"),
        "{message}"
    );
    assert_eq!(count(&w, "event"), events, "a refused take writes nothing");
}

#[test]
fn ended_lease_paths_a_line_lease_succeeds_over_a_held_reservation() {
    let w = world();
    let a = w.held(S_A, "zombie", &["src/lib.rs"]);
    w.end_lease(&a, THREE_DAYS);
    let b = w.held(S_B, "writer", &["src/b.rs"]);
    let done = result(&w.card(
        Some(S_B),
        &[
            "line-lease",
            &b,
            "--generation",
            "1",
            "--path",
            "src/lib.rs",
            "--wait-secs",
            "0",
            "--",
            "true",
        ],
    ));
    assert_eq!(done["added"], json!(["src/lib.rs"]), "{done}");
    assert_eq!(w.scope_of(&b), ["src/b.rs"], "the lease is given back");
    assert!(
        w.scope_of(&a).is_empty(),
        "the old reservation lost the path"
    );
}
