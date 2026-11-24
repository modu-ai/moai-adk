//! `moai flow-measure` (SPEC-20261125-flow-measure): the flow stage times, read from the ledger
//! without writing. Per named card: submit to verdict, pass to done, the failed verdicts, the
//! auditor, the head and the reservation. Per `--hours` window: the events counted apart, the
//! stage percentiles, the share of first verdicts that failed and the `LANE-BLOCKED` kinds.
//!
//! Every command runs against a temporary HOME and a temporary git project; the user's real state
//! folder is not read or written. Ledger rows are inserted with SQL so every time is exactly what
//! a case needs. No `claude` or `orca` process starts, and no Python runs.
#![cfg(unix)]

mod common;

use common::git::git;
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::cell::Cell;
use std::path::PathBuf;
use std::process::Output;

const AUDITOR: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const LANE: &str = "22222222-0000-4000-8000-000000000002";
/// The start of every fixture; times below are seconds after it.
const B: i64 = 10_000_000;

struct World {
    home: TempHome,
    project: PathBuf,
    events: Cell<i64>,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    git(&project, &["init", "-q", "-b", "main"]);
    ledger::register(home.path(), "demo").unwrap();
    World {
        home,
        project,
        events: Cell::new(0),
    }
}

fn text(out: &Output) -> String {
    format!(
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn result(out: &Output) -> Value {
    assert_eq!(out.status.code(), Some(0), "{}", text(out));
    let v: Value =
        serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {}", text(out)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// The 40-hex head of a fixture card.
fn head_of(n: i64, round: i64) -> String {
    format!("{n:020x}{round:020x}")
}

impl World {
    fn project_str(&self) -> String {
        self.project.to_str().unwrap().to_string()
    }

    /// `flow-measure` with `args` and the project of this world.
    fn run(&self, args: &[&str]) -> Output {
        let project = self.project_str();
        let mut all = vec!["flow-measure", "--project", project.as_str()];
        all.extend_from_slice(args);
        let mut cmd = self.home.command(&all);
        cmd.env_remove("CLAUDE_CODE_SESSION_ID");
        cmd.output().expect("run moai")
    }

    fn sql(&self, sql: &str, params: impl rusqlite::Params) {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn().execute(sql, params).unwrap();
    }

    fn card(&self, id: i64, state: &str) {
        self.sql(
            "INSERT INTO card (card_id, title, kind, state, priority, write_scope, created_by, \
             created_at, updated_at, closed_at) \
             VALUES (?1, ?2, 'task', ?3, 'p2', '[\"a.md\"]', 'tester', 1, 2, \
             CASE WHEN ?3 = 'done' THEN 3 END)",
            rusqlite::params![id, format!("card {id}"), state],
        );
    }

    fn reservation(&self, card: i64, generation: i64, status: &str, acquired: i64, updated: i64) {
        self.sql(
            "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
             lease_until, updated_at, worktree, branch) \
             VALUES (?1, ?2, ?3, ?4, ?5, 99999999999, ?6, NULL, NULL)",
            rusqlite::params![card, generation, LANE, status, acquired, updated],
        );
    }

    fn note(&self, card: i64, author: &str, kind: &str, body: &str, at: i64) {
        self.sql(
            "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
             VALUES (?1, ?2, ?3, 'local', ?4, ?5)",
            rusqlite::params![card, kind, author, body, at],
        );
    }

    fn event(&self, card: Option<i64>, kind: &str, actor: &str, at: i64, payload: &Value) {
        let n = self.events.get() + 1;
        self.events.set(n);
        self.sql(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
            rusqlite::params![
                format!("fm-{n}"),
                at,
                actor,
                card,
                kind,
                payload.to_string()
            ],
        );
    }

    fn verdict(&self, card: i64, pass: bool, head: &str, at: i64) {
        self.event(
            Some(card),
            "card.verdict",
            AUDITOR,
            at,
            &json!({
                "head": head, "base": head_of(0, 0),
                "verdict": if pass { "pass" } else { "fail" },
                "tier": 1, "evidence": "read", "receipt": null, "reviewer": AUDITOR,
            }),
        );
    }

    fn lane_note(&self, card: i64, body: &str, at: i64) {
        self.note(card, "agent-1", "report", body, at);
    }

    fn ready(&self, card: i64, round: i64, at: i64) {
        let head = head_of(card, round);
        self.lane_note(
            card,
            &format!("LANE-READY-INTEGRATION C-{card} head={head}\nbase=x"),
            at,
        );
    }

    fn repaired(&self, card: i64, round: i64, at: i64) {
        let head = head_of(card, round);
        self.lane_note(
            card,
            &format!("REPAIRED C-{card} head={head}\nleftovers: none"),
            at,
        );
    }

    fn finish(&self, card: i64, at: i64) {
        self.event(
            Some(card),
            "card.finish",
            "leader",
            at,
            &json!({"via": "land"}),
        );
    }

    fn landed(&self, card: i64, at: i64) {
        self.event(
            Some(card),
            "card.landed",
            "leader",
            at,
            &json!({"via": "land"}),
        );
    }

    fn added(&self, card: Option<i64>, at: i64) {
        self.event(card, "card.add", "user", at, &json!({}));
    }

    /// A done card whose pass came `p2d` seconds before its finish: ready, pass, finish.
    fn finished_card(&self, id: i64, p2d: i64) {
        self.card(id, "done");
        self.reservation(id, 1, "released", B, B + 100 + 200 + p2d);
        self.ready(id, 1, B + 100);
        self.verdict(id, true, &head_of(id, 1), B + 200);
        self.finish(id, B + 200 + p2d);
    }
}

// ---- AC-1 ----------------------------------------------------------------------------------

#[test]
fn a_project_without_a_ledger_answers_absent_and_a_ledger_is_not_written() {
    let home = TempHome::new();
    let project = home.project("bare");
    git(&project, &["init", "-q", "-b", "main"]);
    let before = home.snapshot();
    let out = home
        .command(&[
            "flow-measure",
            "--project",
            project.to_str().unwrap(),
            "--hours",
            "12",
        ])
        .output()
        .unwrap();
    let r = result(&out);
    assert_eq!(r["ledger"], json!("absent"), "{r}");
    assert_eq!(r["writes"], json!(false));
    assert_eq!(r["window"]["issued"], json!(0), "{r}");
    assert_eq!(
        home.snapshot(),
        before,
        "a project without a ledger must stay as it was"
    );

    // With a ledger and no `--now`: the window ends at the clock of the database, and the files
    // are the same bytes before and after the run.
    let w = world();
    w.card(7, "ready");
    w.added(Some(7), B);
    let before = w.home.snapshot();
    let r = result(&w.run(&["--hours", "12"]));
    let clock = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64;
    let now = r["now"].as_i64().expect("now");
    assert!((now - clock).abs() < 30, "now {now} against clock {clock}");
    assert_eq!(r["ledger"], json!("present"));
    assert_eq!(
        r["window"]["issued"],
        json!(0),
        "the fixture event is far outside: {r}"
    );
    assert_eq!(
        w.home.snapshot(),
        before,
        "the ledger files are changed by a read"
    );
}

// ---- AC-2 ----------------------------------------------------------------------------------

#[test]
fn a_card_line_reads_submit_verdict_and_done_by_the_definitions() {
    let w = world();
    w.card(7, "done");
    w.reservation(7, 1, "released", B, B + 2500);
    w.ready(7, 1, B + 100);
    w.verdict(7, false, &head_of(7, 1), B + 700);
    w.verdict(7, false, &head_of(7, 1), B + 800);
    w.repaired(7, 2, B + 1000);
    // A leader note and a decision note that start with a marker are not hand-back notes.
    w.note(
        7,
        "leader",
        "report",
        &format!("REPAIRED C-7 head={}\nquoted", head_of(7, 3)),
        B + 1100,
    );
    w.note(
        7,
        "agent-1",
        "decision",
        &format!("LANE-READY-INTEGRATION C-7 head={}\nquoted", head_of(7, 4)),
        B + 1200,
    );
    w.verdict(7, true, &head_of(7, 2), B + 1600);
    // A hand-back note after the pass is not the submit of that pass.
    w.repaired(7, 5, B + 1800);
    w.finish(7, B + 2500);

    let r = result(&w.run(&["C-7"]));
    let c = &r["cards"][0];
    assert_eq!(c["card"], json!("C-7"), "{r}");
    assert_eq!(c["submit_at"], json!(B + 1000), "{c}");
    assert_eq!(c["verdict_at"], json!(B + 1600), "{c}");
    assert_eq!(c["done_at"], json!(B + 2500), "{c}");
    assert_eq!(c["submit_to_verdict_sec"], json!(600), "{c}");
    assert_eq!(c["pass_to_done_sec"], json!(900), "{c}");
    assert_eq!(c["fail_count"], json!(2), "two fails and one pass: {c}");
    assert_eq!(c["auditor"], json!(AUDITOR), "{c}");
    assert_eq!(c["head"], json!(head_of(7, 2)), "{c}");
    assert_eq!(c["reservation"]["status"], json!("released"), "{c}");
    assert_eq!(c["reservation"]["left_held_at"], json!(B + 2500), "{c}");
    assert_eq!(r["writes"], json!(false));
}

// ---- AC-3 ----------------------------------------------------------------------------------

#[test]
fn a_span_with_a_missing_end_is_null_and_never_zero() {
    let w = world();
    // C-8: handed back, never judged. C-9: passed, never finished, still held. C-10: nothing.
    w.card(8, "ready");
    w.ready(8, 1, B + 100);
    w.card(9, "active");
    w.reservation(9, 1, "held", B, B);
    w.ready(9, 1, B + 100);
    w.verdict(9, true, &head_of(9, 1), B + 400);
    w.card(10, "ready");

    let r = result(&w.run(&["C-8", "C-9", "C-10"]));
    let c8 = &r["cards"][0];
    assert_eq!(c8["submit_to_verdict_sec"], Value::Null, "{c8}");
    assert_eq!(c8["pass_to_done_sec"], Value::Null, "{c8}");
    assert_eq!(c8["verdict_at"], Value::Null, "{c8}");
    assert_eq!(c8["head"], Value::Null, "{c8}");
    assert_eq!(c8["fail_count"], json!(0), "{c8}");
    assert_eq!(c8["reservation"], Value::Null, "{c8}");
    let c9 = &r["cards"][1];
    assert_eq!(c9["submit_to_verdict_sec"], json!(300), "{c9}");
    assert_eq!(c9["pass_to_done_sec"], Value::Null, "{c9}");
    assert_eq!(c9["done_at"], Value::Null, "{c9}");
    assert_eq!(c9["reservation"]["status"], json!("held"), "{c9}");
    assert_eq!(c9["reservation"]["left_held_at"], Value::Null, "{c9}");
    let c10 = &r["cards"][2];
    assert_eq!(c10["submit_at"], Value::Null, "{c10}");
    assert_eq!(c10["submit_to_verdict_sec"], Value::Null, "{c10}");
}

// ---- AC-4 ----------------------------------------------------------------------------------

#[test]
fn the_percentiles_are_nearest_rank_and_carry_the_raw_values() {
    let w = world();
    // Inserted out of order: the summary sorts.
    for (id, p2d) in [(21, 300), (22, 60), (23, 240), (24, 120), (25, 180)] {
        w.finished_card(id, p2d);
    }
    let r = result(&w.run(&["C-21", "C-22", "C-23", "C-24", "C-25"]));
    let s = &r["pass_to_done"];
    assert_eq!(s["n"], json!(5), "{s}");
    assert_eq!(s["p50_sec"], json!(180), "rank ceil(0.5 * 5) = 3: {s}");
    assert_eq!(s["p90_sec"], json!(300), "rank ceil(0.9 * 5) = 5: {s}");
    assert_eq!(s["values_sec"], json!([60, 120, 180, 240, 300]), "{s}");

    // One sample: both percentiles are that value.
    let r = result(&w.run(&["C-24"]));
    let s = &r["pass_to_done"];
    assert_eq!(
        (s["n"].clone(), s["p50_sec"].clone(), s["p90_sec"].clone()),
        (json!(1), json!(120), json!(120)),
        "{s}"
    );

    // No sample: `null`, never zero, and the count is zero.
    w.card(30, "ready");
    let r = result(&w.run(&["C-30"]));
    let s = &r["pass_to_done"];
    assert_eq!(s["n"], json!(0), "{s}");
    assert_eq!(s["p50_sec"], Value::Null, "{s}");
    assert_eq!(s["p90_sec"], Value::Null, "{s}");
    assert_eq!(s["values_sec"], json!([]), "{s}");
}

// ---- AC-5 ----------------------------------------------------------------------------------

#[test]
fn a_window_counts_issued_landed_and_finished_apart() {
    let w = world();
    let now = B + 100_000;
    w.card(1, "done");
    // issued: inside, on the edge (counted), one second outside, in the future.
    w.added(Some(1), now - 10);
    w.added(Some(1), now - 3600);
    w.added(Some(1), now - 3601);
    w.added(Some(1), now + 5);
    // landed: one inside, one outside.
    w.landed(1, now - 100);
    w.landed(1, now - 7200);
    // finished: three inside.
    w.finish(1, now - 300);
    w.finish(1, now - 200);
    w.finish(1, now - 100);
    // Another kind in the window is not counted.
    w.event(Some(1), "card.reclaim", "leader", now - 50, &json!({}));

    let r = result(&w.run(&["--hours", "1", "--now", &now.to_string()]));
    let win = &r["window"];
    assert_eq!(win["hours"], json!(1), "{win}");
    assert_eq!(win["since"], json!(now - 3600), "{win}");
    assert_eq!(win["until"], json!(now), "{win}");
    assert_eq!(win["issued"], json!(2), "{win}");
    assert_eq!(win["landed"], json!(1), "{win}");
    assert_eq!(win["finished"], json!(3), "{win}");
}

// ---- AC-6 ----------------------------------------------------------------------------------

#[test]
fn a_window_gives_stage_percentiles_first_verdict_fail_share_and_block_kinds() {
    let w = world();
    let now = B + 100_000;
    let t = now - 30_000;

    // C-31: taken, handed back, failed, repaired, passed, landed, finished.
    w.card(31, "done");
    w.reservation(31, 1, "released", t, t + 2700);
    w.ready(31, 1, t + 600);
    w.verdict(31, false, &head_of(31, 1), t + 1200);
    w.repaired(31, 2, t + 1500);
    w.verdict(31, true, &head_of(31, 2), t + 2100);
    w.landed(31, t + 2400);
    w.finish(31, t + 2700);
    // C-32: passed at the first verdict, landed, finished.
    w.card(32, "done");
    w.reservation(32, 1, "released", t, t + 1900);
    w.ready(32, 1, t + 100);
    w.verdict(32, true, &head_of(32, 1), t + 400);
    w.landed(32, t + 1000);
    w.finish(32, t + 1900);
    // C-33: handed back and failed, nothing after.
    w.card(33, "ready");
    w.reservation(33, 1, "released", t, t + 1300);
    w.ready(33, 1, t + 300);
    w.verdict(33, false, &head_of(33, 1), t + 1200);
    // C-34: judged long before the window; none of its spans may count.
    w.card(34, "done");
    w.reservation(34, 1, "released", now - 90_000, now - 80_000);
    w.ready(34, 1, now - 90_000 + 100);
    w.verdict(34, true, &head_of(34, 1), now - 90_000 + 5100);
    w.landed(34, now - 84_000);
    w.finish(34, now - 80_000);

    // LANE-BLOCKED notes: four in the window, one before it, one that is not a block.
    w.lane_note(31, "LANE-BLOCKED kind=decision C-31 which option", t + 10);
    w.lane_note(32, "LANE-BLOCKED kind=permission C-32 write denied", t + 20);
    w.lane_note(33, "LANE-BLOCKED kind=decision C-33 another choice", t + 30);
    w.lane_note(33, "LANE-BLOCKED C-33 no kind word here", t + 40);
    w.lane_note(33, "LANE-BLOCKED kind=decision C-33 long ago", now - 90_000);
    w.lane_note(33, "NOT-A-BLOCK kind=decision", t + 50);

    let r = result(&w.run(&["--hours", "10", "--now", &now.to_string()]));
    let win = &r["window"];
    let stage = |name: &str| {
        let s = &win["stages"][name];
        (s["n"].clone(), s["p50_sec"].clone(), s["p90_sec"].clone())
    };
    // take_to_submit: 600, 100, 300 -> sorted 100, 300, 600.
    assert_eq!(
        stage("take_to_submit"),
        (json!(3), json!(300), json!(600)),
        "{win}"
    );
    // submit_to_verdict: 600 and 600 (C-31), 300 (C-32), 900 (C-33) -> 300, 600, 600, 900.
    assert_eq!(
        stage("submit_to_verdict"),
        (json!(4), json!(600), json!(900)),
        "{win}"
    );
    // pass_to_landed: 300 (C-31), 600 (C-32).
    assert_eq!(
        stage("pass_to_landed"),
        (json!(2), json!(300), json!(600)),
        "{win}"
    );
    // pass_to_done: 600 (C-31), 1500 (C-32).
    assert_eq!(
        stage("pass_to_done"),
        (json!(2), json!(600), json!(1500)),
        "{win}"
    );

    // First verdicts in the window: C-31 fail, C-32 pass, C-33 fail.
    let fv = &win["first_verdict"];
    assert_eq!(fv["n"], json!(3), "{fv}");
    assert_eq!(fv["fail"], json!(2), "{fv}");
    assert!(
        (fv["fail_rate"].as_f64().unwrap() - 0.67).abs() < 1e-9,
        "{fv}"
    );

    let lb = &win["lane_blocked"];
    assert_eq!(lb["total"], json!(4), "{lb}");
    assert_eq!(
        lb["kinds"],
        json!({"decision": 2, "permission": 1, "unspecified": 1}),
        "{lb}"
    );

    // A window with nothing in it: counts zero, summaries and the share are `null`.
    let r = result(&w.run(&["--hours", "1", "--now", &(B + 5).to_string()]));
    let win = &r["window"];
    assert_eq!(win["issued"], json!(0), "{win}");
    assert_eq!(win["stages"]["pass_to_done"]["n"], json!(0), "{win}");
    assert_eq!(
        win["stages"]["pass_to_done"]["p50_sec"],
        Value::Null,
        "{win}"
    );
    assert_eq!(win["first_verdict"]["fail_rate"], Value::Null, "{win}");
    assert_eq!(win["lane_blocked"]["total"], json!(0), "{win}");
}

// ---- AC-7 ----------------------------------------------------------------------------------

#[test]
fn the_table_lines_carry_the_numbers_of_the_json() {
    let w = world();
    let now = B + 100_000;
    w.card(7, "done");
    w.reservation(7, 1, "released", B, B + 2500);
    w.ready(7, 1, B + 1000);
    w.verdict(7, true, &head_of(7, 1), B + 1600);
    w.finish(7, B + 2500);

    let r = result(&w.run(&["C-7"]));
    let lines: Vec<String> = r["table"]
        .as_array()
        .expect("table")
        .iter()
        .map(|l| l.as_str().expect("a line").to_string())
        .collect();
    let card_line = lines
        .iter()
        .find(|l| l.contains("C-7"))
        .unwrap_or_else(|| panic!("no line names C-7: {lines:?}"));
    assert!(card_line.contains("10.0"), "600 s is 10.0 min: {card_line}");
    assert!(card_line.contains("15.0"), "900 s is 15.0 min: {card_line}");
    assert!(card_line.contains("released"), "{card_line}");
    let summary = lines
        .iter()
        .find(|l| l.contains("pass_to_done"))
        .expect("summary line");
    assert!(
        summary.contains("n=1") && summary.contains("P50") && summary.contains("15.0"),
        "{summary}"
    );

    let r = result(&w.run(&["--hours", "100", "--now", &now.to_string()]));
    let joined = r["table"]
        .as_array()
        .expect("table")
        .iter()
        .filter_map(Value::as_str)
        .collect::<Vec<_>>()
        .join("\n");
    for needle in [
        "issued",
        "landed",
        "finished",
        "take_to_submit",
        "submit_to_verdict",
        "pass_to_landed",
        "pass_to_done",
        "first_verdict",
        "LANE-BLOCKED",
    ] {
        assert!(
            joined.contains(needle),
            "the window table lacks `{needle}`:\n{joined}"
        );
    }
}

// ---- AC-8 ----------------------------------------------------------------------------------

#[test]
fn a_bad_argument_is_refused_by_name_and_writes_nothing() {
    let w = world();
    w.card(7, "ready");
    let before = w.home.snapshot();
    let cases: [(&[&str], &str); 5] = [
        (&[], "--hours"),
        (&["C-x"], "C-x"),
        (&["C-999"], "C-999"),
        (&["--hours", "0"], "--hours"),
        (&["--hours", "abc"], "--hours"),
    ];
    for (args, names) in cases {
        let out = w.run(args);
        assert_ne!(
            out.status.code(),
            Some(0),
            "{args:?} must be refused: {}",
            text(&out)
        );
        assert!(
            text(&out).contains(names),
            "{args:?} must name `{names}`: {}",
            text(&out)
        );
    }
    assert_eq!(w.home.snapshot(), before, "a refusal wrote something");
}
