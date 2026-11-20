//! `moai audit-queue` (SPEC-20261120-audit-request-notes): the open audit requests of a ledger,
//! read from card notes of four fixed first lines, with their age, state and stage times.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. Notes
//! are inserted with SQL because a case needs a chosen note time that `card note` cannot give;
//! the ledger clock of the command is the unix time of that moment, so ages are read as ranges.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::Value;
use std::path::Path;
use std::process::Output;

const H1: &str = "1111111111111111111111111111111111111111";
const H2: &str = "2222222222222222222222222222222222222222";
const H3: &str = "3333333333333333333333333333333333333333";

struct World {
    home: TempHome,
    project: String,
    now: i64,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo").to_str().unwrap().to_string();
    ledger::register(home.path(), "demo").unwrap();
    let mut w = ledger::open_write(home.path(), "demo").unwrap();
    let now = w
        .with_write_tx(|tx, now| {
            for title in ["first", "second"] {
                tx.execute(
                    "INSERT INTO card (title, state, priority, write_scope, created_by, created_at, updated_at) \
                     VALUES (?1, 'active', 'p2', '[]', 'test', ?2, ?2)",
                    rusqlite::params![title, now],
                )?;
            }
            Ok(now)
        })
        .unwrap();
    World { home, project, now }
}

impl World {
    /// A local note of `card` (1 or 2) written `ago` seconds before the ledger's now.
    fn note(&self, card: i64, author: &str, ago: i64, text: &str) {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        let at = self.now - ago;
        w.with_write_tx(|tx, _| {
            tx.execute(
                "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
                 VALUES (?1, 'report', ?2, 'local', ?3, ?4)",
                rusqlite::params![card, author, text, at],
            )?;
            Ok(())
        })
        .unwrap();
    }

    /// A note of `card` that came from an import, not written here.
    fn note_import(&self, card: i64, ago: i64, text: &str) {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        let at = self.now - ago;
        w.with_write_tx(|tx, _| {
            tx.execute(
                "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
                 VALUES (?1, 'report', 'importer', 'import', ?2, ?3)",
                rusqlite::params![card, text, at],
            )?;
            Ok(())
        })
        .unwrap();
    }

    /// A third card that is already done (the ledger refuses to close a card that has no met
    /// criterion, so the row is inserted done); returns its id.
    fn done_card(&self) -> i64 {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.with_write_tx(|tx, now| {
            tx.execute(
                "INSERT INTO card (title, state, priority, write_scope, created_by, created_at, updated_at, closed_at) \
                 VALUES ('third', 'done', 'p2', '[]', 'test', ?1, ?1, ?1)",
                [now],
            )?;
            Ok(tx.last_insert_rowid())
        })
        .unwrap()
    }

    fn queue_raw(&self, extra: &[&str]) -> Output {
        let mut args = vec!["audit-queue", "--project", &self.project];
        args.extend_from_slice(extra);
        self.home.command(&args).output().expect("run moai")
    }

    fn queue(&self, extra: &[&str]) -> Value {
        let out = self.queue_raw(extra);
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stdout)
        );
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(v["ok"], true, "{v}");
        v["result"].clone()
    }
}

fn item<'a>(result: &'a Value, card: &str, head: &str) -> &'a Value {
    result["requests"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["card"] == card && r["head"] == head)
        .unwrap_or_else(|| panic!("no request {card} {head} in {result}"))
}

fn request(card: &str, head: &str) -> String {
    format!("AUDIT-REQUEST {card} head={head} tier=1 scope=src/a.rs and tests/a.rs\nmore text")
}

#[test]
fn lists_a_request_with_its_head_tier_scope_and_age() {
    let w = world();
    w.note(1, "lane-a", 120, &request("C-1", H1));
    let r = w.queue(&[]);
    assert_eq!(r["writes"], false);
    assert_eq!(r["count"], 1, "{r}");
    let it = item(&r, "C-1", H1);
    assert_eq!(it["tier"], 1);
    assert_eq!(it["scope"], "src/a.rs and tests/a.rs");
    assert_eq!(it["state"], "waiting");
    let age = it["age_secs"].as_i64().unwrap();
    assert!((120..=180).contains(&age), "{age}");
}

#[test]
fn lists_nothing_for_a_head_that_has_a_verdict_and_nothing_for_an_absent_ledger() {
    let w = world();
    w.note(1, "lane-a", 300, &request("C-1", H1));
    w.note(
        1,
        "leader",
        200,
        &format!("AUDIT-CLAIM C-1 head={H1} by=audit-x"),
    );
    w.note(
        1,
        "leader",
        100,
        &format!("AUDIT-VERDICT C-1 head={H1} verdict=PASS"),
    );
    let r = w.queue(&[]);
    assert_eq!(r["count"], 0, "{r}");
    let empty = TempHome::new();
    let project = empty.project("demo").to_str().unwrap().to_string();
    let out = empty
        .command(&["audit-queue", "--project", &project])
        .output()
        .unwrap();
    assert!(out.status.success());
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["result"]["ledger"], "absent");
    assert_eq!(v["result"]["count"], 0);
    assert!(
        !Path::new(&empty.path().join(".moai")).exists(),
        "nothing is created"
    );
}

#[test]
fn state_tells_waiting_running_repair_and_stale_apart() {
    let w = world();
    // C-1: request only -> waiting. C-2: request and claim -> running.
    w.note(1, "lane-a", 100, &request("C-1", H1));
    w.note(2, "lane-b", 400, &request("C-2", H2));
    w.note(
        2,
        "leader",
        300,
        &format!("AUDIT-CLAIM C-2 head={H2} by=audit-y"),
    );
    let r = w.queue(&[]);
    assert_eq!(item(&r, "C-1", H1)["state"], "waiting");
    assert_eq!(item(&r, "C-2", H2)["state"], "running");
    assert_eq!(item(&r, "C-2", H2)["auditor"], "audit-y");
    // A FAIL with no repair is `repair`; with a repair it is gone and the repair head is a
    // request of its own that goes to the first auditor.
    w.note(
        2,
        "leader",
        200,
        &format!("AUDIT-VERDICT C-2 head={H2} verdict=FAIL"),
    );
    let r = w.queue(&[]);
    assert_eq!(item(&r, "C-2", H2)["state"], "repair");
    w.note(2, "lane-b", 150, &format!("REPAIR C-2 head={H3} of={H2}"));
    let r = w.queue(&[]);
    let again = item(&r, "C-2", H3);
    assert_eq!(again["state"], "waiting");
    assert_eq!(again["of"], H2);
    assert_eq!(
        again["auditor"], "audit-y",
        "the repair goes to the first auditor"
    );
}

#[test]
fn state_stale_starts_after_the_limit_and_not_before() {
    let w = world();
    w.note(1, "lane-a", 1000, &request("C-1", H1));
    // The age is the time since the note; the limit is compared with `>` after the note's own
    // second, so 5 s under and 5 s over the age cover the edge without racing the clock.
    let under = w.queue(&["--stale-secs", "1100"]);
    assert_eq!(item(&under, "C-1", H1)["state"], "waiting");
    let over = w.queue(&["--stale-secs", "900"]);
    let it = item(&over, "C-1", H1);
    assert_eq!(it["state"], "stale");
    assert_eq!(it["was"], "waiting");
    let bad = w.queue_raw(&["--stale-secs", "0"]);
    assert!(!bad.status.success(), "a limit of 0 is refused");
}

#[test]
fn times_are_computed_from_the_note_times_and_leave_out_a_span_without_an_end() {
    let w = world();
    w.note(1, "lane-a", 1000, &request("C-1", H1));
    w.note(
        1,
        "leader",
        940,
        &format!("AUDIT-CLAIM C-1 head={H1} by=audit-x"),
    );
    w.note(
        1,
        "leader",
        700,
        &format!("AUDIT-VERDICT C-1 head={H1} verdict=FAIL"),
    );
    w.note(1, "lane-a", 600, &format!("REPAIR C-1 head={H2} of={H1}"));
    w.note(
        1,
        "leader",
        500,
        &format!("AUDIT-CLAIM C-1 head={H2} by=audit-x"),
    );
    let r = w.queue(&["--all"]);
    let first = item(&r, "C-1", H1);
    assert_eq!(first["times"]["request_to_claim_secs"], 60);
    assert_eq!(first["times"]["claim_to_verdict_secs"], 240);
    assert_eq!(first["times"]["verdict_to_repair_secs"], 100);
    assert_eq!(first["times"]["repair_to_claim_secs"], 100);
    // The repair head has a claim but no verdict: the note of its repair is its request, and its
    // audit span is absent, not zero.
    let second = item(&r, "C-1", H2);
    assert_eq!(second["times"]["request_to_claim_secs"], 100);
    assert_eq!(second["times"]["claim_to_verdict_secs"], Value::Null);
    assert_eq!(second["state"], "running");
}

#[test]
fn rules_a_second_request_for_a_passed_head_is_reuse() {
    let w = world();
    w.note(1, "lane-a", 500, &request("C-1", H1));
    w.note(
        1,
        "leader",
        400,
        &format!("AUDIT-CLAIM C-1 head={H1} by=audit-x"),
    );
    w.note(
        1,
        "leader",
        300,
        &format!("AUDIT-VERDICT C-1 head={H1} verdict=PASS"),
    );
    w.note(1, "lane-a", 100, &request("C-1", H1));
    let r = w.queue(&[]);
    let it = item(&r, "C-1", H1);
    assert_eq!(it["state"], "reuse", "{it}");
    assert_eq!(it["verdict"], "PASS");
}

#[test]
fn rules_a_broken_line_is_listed_as_malformed_and_not_counted() {
    let w = world();
    w.note(
        1,
        "lane-a",
        100,
        "AUDIT-REQUEST C-1 head=abc tier=1 scope=x",
    );
    w.note(
        1,
        "lane-a",
        90,
        &format!("AUDIT-VERDICT C-1 head={H1} verdict=MAYBE"),
    );
    w.note(1, "lane-a", 80, "REPAIR C-1 head=zz");
    w.note(1, "lane-a", 70, "audit-request in lower case is just text");
    let r = w.queue(&[]);
    assert_eq!(r["count"], 0, "{r}");
    let bad = r["malformed"].as_array().unwrap();
    assert_eq!(bad.len(), 3, "{r}");
    assert!(bad
        .iter()
        .all(|b| b["card"] == "C-1" && b["note_id"].is_i64()));
}

/// A card whose first audit failed, repaired at `H3`, whose second audit failed too, repaired at
/// `H2` (a head that sorts before `H3`): `FAIL` verdicts and `REPAIR` notes in note order.
fn notes_of_a_repair_of_a_repair(w: &World) {
    w.note(1, "lane-a", 1000, &request("C-1", H1));
    w.note(
        1,
        "leader",
        950,
        &format!("AUDIT-CLAIM C-1 head={H1} by=audit-x"),
    );
    w.note(
        1,
        "leader",
        900,
        &format!("AUDIT-VERDICT C-1 head={H1} verdict=FAIL"),
    );
    w.note(1, "lane-a", 800, &format!("REPAIR C-1 head={H3} of={H1}"));
    w.note(
        1,
        "leader",
        750,
        &format!("AUDIT-CLAIM C-1 head={H3} by=audit-x"),
    );
    w.note(
        1,
        "leader",
        700,
        &format!("AUDIT-VERDICT C-1 head={H3} verdict=FAIL"),
    );
    w.note(1, "lane-a", 600, &format!("REPAIR C-1 head={H2} of={H3}"));
}

#[test]
fn state_a_repair_of_a_repair_keeps_the_tier_and_scope_whatever_the_head_order() {
    let w = world();
    notes_of_a_repair_of_a_repair(&w);
    let r = w.queue(&["--all"]);
    for head in [H1, H3, H2] {
        let it = item(&r, "C-1", head);
        assert_eq!(it["tier"], 1, "{head}: {it}");
        assert_eq!(it["scope"], "src/a.rs and tests/a.rs", "{head}: {it}");
    }
    assert_eq!(item(&r, "C-1", H2)["of"], H3);
}

#[test]
fn state_the_auditor_of_a_card_is_its_first_claimer_when_another_claims_later() {
    let w = world();
    w.note(1, "lane-a", 1000, &request("C-1", H1));
    w.note(
        1,
        "leader",
        950,
        &format!("AUDIT-CLAIM C-1 head={H1} by=audit-x"),
    );
    w.note(
        1,
        "leader",
        900,
        &format!("AUDIT-VERDICT C-1 head={H1} verdict=FAIL"),
    );
    w.note(1, "lane-a", 800, &format!("REPAIR C-1 head={H2} of={H1}"));
    w.note(
        1,
        "leader",
        750,
        &format!("AUDIT-CLAIM C-1 head={H2} by=audit-z"),
    );
    let r = w.queue(&["--all"]);
    let again = item(&r, "C-1", H2);
    assert_eq!(again["claim_by"], "audit-z");
    assert_eq!(again["auditor"], "audit-x", "{again}");
    assert_eq!(item(&r, "C-1", H1)["auditor"], "audit-x");
}

#[test]
fn times_two_repairs_of_one_head_measure_the_span_to_the_claim_of_the_first() {
    let w = world();
    w.note(1, "lane-a", 1000, &request("C-1", H1));
    w.note(
        1,
        "leader",
        950,
        &format!("AUDIT-CLAIM C-1 head={H1} by=audit-x"),
    );
    w.note(
        1,
        "leader",
        900,
        &format!("AUDIT-VERDICT C-1 head={H1} verdict=FAIL"),
    );
    w.note(1, "lane-a", 700, &format!("REPAIR C-1 head={H2} of={H1}"));
    w.note(
        1,
        "leader",
        640,
        &format!("AUDIT-CLAIM C-1 head={H2} by=audit-x"),
    );
    w.note(1, "lane-a", 600, &format!("REPAIR C-1 head={H3} of={H1}"));
    w.note(
        1,
        "leader",
        590,
        &format!("AUDIT-CLAIM C-1 head={H3} by=audit-x"),
    );
    let r = w.queue(&["--all"]);
    let first = item(&r, "C-1", H1);
    assert_eq!(first["times"]["verdict_to_repair_secs"], 200);
    assert_eq!(first["times"]["repair_to_claim_secs"], 60, "{first}");
}

#[test]
fn rules_a_done_card_and_an_import_note_are_not_listed() {
    let w = world();
    let done = w.done_card();
    assert_eq!(done, 3);
    w.note(1, "lane-a", 300, &request("C-1", H1));
    w.note(done, "lane-b", 300, &request("C-3", H2));
    w.note_import(1, 200, &request("C-1", H3));
    let r = w.queue(&[]);
    assert_eq!(r["count"], 1, "{r}");
    assert_eq!(r["requests"][0]["card"], "C-1");
    assert_eq!(r["requests"][0]["head"], H1);
}

#[test]
fn rules_the_stale_limit_stops_at_604800_and_a_note_from_the_future_has_age_zero() {
    let w = world();
    w.note(1, "lane-a", -100, &request("C-1", H1));
    let ok = w.queue(&["--stale-secs", "604800"]);
    assert_eq!(item(&ok, "C-1", H1)["age_secs"], 0, "{ok}");
    assert_eq!(item(&ok, "C-1", H1)["state_age_secs"], 0, "{ok}");
    let bad = w.queue_raw(&["--stale-secs", "604801"]);
    assert!(!bad.status.success(), "604801 is over the limit");
}

#[test]
fn rules_a_request_for_a_head_with_a_fail_verdict_is_reuse_and_the_repair_state_stays() {
    let w = world();
    w.note(1, "lane-a", 500, &request("C-1", H1));
    w.note(
        1,
        "leader",
        450,
        &format!("AUDIT-CLAIM C-1 head={H1} by=audit-x"),
    );
    w.note(
        1,
        "leader",
        400,
        &format!("AUDIT-VERDICT C-1 head={H1} verdict=FAIL"),
    );
    w.note(1, "lane-a", 100, &request("C-1", H1));
    let r = w.queue(&[]);
    let mut states: Vec<&str> = r["requests"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["head"] == H1)
        .map(|i| i["state"].as_str().unwrap())
        .collect();
    states.sort_unstable();
    assert_eq!(states, ["repair", "reuse"], "{r}");
}

#[test]
fn rules_the_command_does_not_change_the_ledger() {
    let w = world();
    w.note(1, "lane-a", 100, &request("C-1", H1));
    let before = w.home.snapshot();
    w.queue(&[]);
    w.queue(&["--all", "--stale-secs", "60"]);
    assert_eq!(w.home.snapshot(), before, "the ledger bytes are the same");
}

#[test]
fn help_names_the_shapes_the_states_and_the_reuse_rule() {
    let home = TempHome::new();
    let out = home
        .command(&["audit-queue", "--help"])
        .output()
        .expect("run moai");
    // The help wraps lines, so a phrase may break across a newline: compare with one space.
    let help = String::from_utf8_lossy(&out.stdout)
        .split_whitespace()
        .collect::<Vec<_>>()
        .join(" ");
    for phrase in [
        "AUDIT-REQUEST",
        "AUDIT-CLAIM",
        "AUDIT-VERDICT",
        "REPAIR",
        "waiting",
        "running",
        "repair",
        "stale",
        "reuse",
        "--stale-secs",
        "first claim",
    ] {
        assert!(help.contains(phrase), "help must say `{phrase}`:\n{help}");
    }
    let doc = std::fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs/design-review/appendix/operations/audit_queue.md"),
    )
    .expect("the operations fragment exists");
    assert!(doc.starts_with("# `audit-queue`"), "{doc}");
    for phrase in ["AUDIT-REQUEST", "AUDIT-VERDICT", "stale", "reuse"] {
        assert!(doc.contains(phrase), "the fragment must say `{phrase}`");
    }
}
