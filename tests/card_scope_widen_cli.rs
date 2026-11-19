//! `moai card scope` on a card that is held (SPEC-20261119-card-scope-widen, card C-94): the
//! write scope of an active reservation grows without a release and a retake. The reservation
//! rules live in `src/ledger/reservation.rs` (unit tests there); here the command, its answers,
//! its refusal kinds and what it leaves in the ledger.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`); `card scope` itself
//! needs no session (the leader runs it, `--by` is recorded).
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
const S_C: &str = "cccccccc-0000-4000-8000-00000000000c";
const WHY: &str = "the shared type lives under apps/web and the task changes it";

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
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    /// A ready card with one open criterion and `scope`, taken by `session` (generation 1, with
    /// a worktree and a branch). Returns its name.
    fn held(&self, session: &str, title: &str, scope: &[&str]) -> String {
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
        result(&self.card(
            Some(session),
            &[
                "take",
                &name,
                "--worktree",
                "/work/tree",
                "--branch",
                "card/x",
            ],
        ));
        name
    }

    /// The reservation of a card as `card show` prints it.
    fn reservation(&self, card: &str) -> Value {
        result(&self.card(None, &["show", card]))["reservation"].clone()
    }

    /// The card's own scope (`card.write_scope`) as `card show` prints it.
    fn own_scope(&self, card: &str) -> Value {
        result(&self.card(None, &["show", card]))["write_scope"].clone()
    }

    /// `card scope <card> --scope… --by leader [--reason …]`.
    fn widen(&self, card: &str, scope: &[&str], reason: Option<&str>) -> Output {
        let mut args = vec!["scope", card];
        for entry in scope {
            args.push("--scope");
            args.push(entry);
        }
        args.extend_from_slice(&["--by", "leader"]);
        if let Some(reason) = reason {
            args.extend_from_slice(&["--reason", reason]);
        }
        self.card(None, &args)
    }

    fn check(&self, session: &str, card: &str, path: &str) -> Output {
        self.card(
            Some(session),
            &["check", card, "--generation", "1", "--path", path],
        )
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

    fn rows(&self, sql: &str) -> Vec<Vec<Value>> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l.conn().prepare(sql).unwrap();
        let n = st.column_count();
        st.query_map([], |r| {
            Ok((0..n)
                .map(|i| match r.get_ref(i).unwrap() {
                    rusqlite::types::ValueRef::Integer(v) => json!(v),
                    rusqlite::types::ValueRef::Text(t) => {
                        json!(String::from_utf8_lossy(t).to_string())
                    }
                    _ => Value::Null,
                })
                .collect())
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
    }

    /// The patterns of the reservation of `card_id`, sorted.
    fn patterns(&self, card_id: i64) -> Vec<String> {
        self.rows(&format!(
            "SELECT s.pattern FROM reservation_scope s JOIN reservation r \
             ON r.reservation_id = s.reservation_id WHERE r.card_id = {card_id} \
             ORDER BY s.pattern"
        ))
        .into_iter()
        .map(|row| row[0].as_str().unwrap().to_string())
        .collect()
    }

    /// The patterns of the card's current (highest-generation) reservation, sorted.
    fn current_patterns(&self, card_id: i64) -> Vec<String> {
        self.rows(&format!(
            "SELECT s.pattern FROM reservation_scope s JOIN reservation r \
             ON r.reservation_id = s.reservation_id WHERE r.card_id = {card_id} \
             AND r.generation = (SELECT MAX(generation) FROM reservation WHERE card_id = {card_id}) \
             ORDER BY s.pattern"
        ))
        .into_iter()
        .map(|row| row[0].as_str().unwrap().to_string())
        .collect()
    }

    /// The `card.scope` events as `(actor, payload)`.
    fn scope_events(&self) -> Vec<(String, Value)> {
        self.rows("SELECT actor, payload FROM event WHERE kind = 'card.scope' ORDER BY event_id")
            .into_iter()
            .map(|row| {
                (
                    row[0].as_str().unwrap().to_string(),
                    serde_json::from_str(row[1].as_str().unwrap()).unwrap(),
                )
            })
            .collect()
    }

    /// Runs one SQL batch against a write connection — test-side injection only (the lease
    /// aging of the follow-up tests, and a trigger that fails the event write).
    fn exec(&self, sql: &str) {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.with_write_tx(|tx, _| {
            tx.execute_batch(sql)?;
            Ok(())
        })
        .unwrap();
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
fn a_held_card_gets_wider_under_the_same_generation() {
    let w = world();
    let card = w.held(S_A, "wide", &["apps/bo/"]);
    let before = w.reservation(&card);
    let own_before = w.own_scope(&card);
    assert_eq!(before["generation"], json!(1), "{before}");
    assert_eq!(before["write_scope"], json!(["apps/bo/"]), "{before}");
    // Outside the reservation the fence refuses.
    let (kind, message) = refusal(&w.check(S_A, &card, "apps/web/page.ts"));
    assert_eq!(kind, "conflict", "{message}");

    let done = result(&w.widen(&card, &["apps/bo/", "apps/web/"], Some(WHY)));
    assert_eq!(done["action"], json!("scope"), "{done}");
    assert_eq!(done["held"], json!(true), "{done}");
    assert_eq!(done["generation"], json!(1), "{done}");
    assert_eq!(done["holder"], json!(S_A), "{done}");
    assert_eq!(done["from"], json!(["apps/bo/"]), "{done}");
    assert_eq!(
        done["write_scope"],
        json!(["apps/bo/", "apps/web/"]),
        "{done}"
    );
    assert_eq!(done["added"], json!(["apps/web/"]), "{done}");
    assert_eq!(done["by"], json!("leader"), "{done}");
    assert_eq!(done["reason"], json!(WHY), "{done}");
    assert_eq!(done["events"], json!(1), "{done}");

    // Same reservation: generation, holder, lease, worktree, branch, status as before.
    let after = w.reservation(&card);
    for key in [
        "generation",
        "holder",
        "status",
        "acquired_at",
        "lease_until",
        "worktree",
        "branch",
    ] {
        assert_eq!(after[key], before[key], "{key} changed: {after}");
    }
    assert_eq!(
        after["write_scope"],
        json!(["apps/bo/", "apps/web/"]),
        "{after}"
    );
    assert_eq!(w.patterns(1), vec!["apps/bo/", "apps/web/"]);
    assert_eq!(
        w.rows("SELECT COUNT(*) FROM reservation WHERE card_id = 1"),
        vec![vec![json!(1)]],
        "no second reservation row"
    );
    // The card's own scope is not changed (the help says so).
    assert_eq!(w.own_scope(&card), own_before);

    // One event, naming who, why, the generation, the holder and both scopes.
    let events = w.scope_events();
    assert_eq!(events.len(), 1, "{events:?}");
    let (actor, payload) = &events[0];
    assert_eq!(actor, "leader");
    assert_eq!(payload["reason"], json!(WHY), "{payload}");
    assert_eq!(payload["generation"], json!(1), "{payload}");
    assert_eq!(payload["holder"], json!(S_A), "{payload}");
    assert_eq!(payload["held"], json!(true), "{payload}");
    assert_eq!(payload["from"], json!(["apps/bo/"]), "{payload}");
    assert_eq!(payload["to"], json!(["apps/bo/", "apps/web/"]), "{payload}");

    // The holder's next check, at the same generation, allows the new path.
    let ok = result(&w.check(S_A, &card, "apps/web/page.ts"));
    assert_eq!(ok["allowed"], json!(true), "{ok}");
}

#[test]
fn a_card_in_review_can_be_widened_too() {
    let w = world();
    let card = w.held(S_A, "review", &["apps/bo/"]);
    result(&w.card(
        Some(S_A),
        &[
            "criterion",
            &card,
            "1",
            "--generation",
            "1",
            "--met",
            "--evidence",
            "test x",
        ],
    ));
    result(&w.card(Some(S_A), &["review", &card, "--generation", "1"]));
    let shown = result(&w.card(None, &["show", &card]));
    assert_eq!(shown["state"], json!("review"), "{shown}");
    assert_eq!(shown["reservation"]["status"], json!("held"), "{shown}");
    let done = result(&w.widen(&card, &["apps/bo/", "docs/"], Some(WHY)));
    assert_eq!(done["state"], json!("review"), "{done}");
    assert_eq!(done["write_scope"], json!(["apps/bo/", "docs/"]), "{done}");
    assert_eq!(w.patterns(1), vec!["apps/bo/", "docs/"]);
}

#[test]
fn a_pattern_another_card_holds_is_refused_by_name() {
    let w = world();
    let a = w.held(S_A, "a", &["apps/bo/"]);
    let b = w.held(S_B, "b", &["apps/web/"]);
    assert_eq!((a.as_str(), b.as_str()), ("C-1", "C-2"));
    let before = w.dump();
    for wanted in [
        &["apps/bo/", "apps/web/"][..],
        &["apps/bo/", "apps/web/page.ts"][..],
    ] {
        let (kind, message) = refusal(&w.widen(&a, wanted, Some(WHY)));
        assert_eq!(kind, "conflict", "{message}");
        assert!(message.contains("C-2"), "names the other card: {message}");
        assert_eq!(w.dump(), before, "{wanted:?}: nothing was written");
    }
    // Once the other card gives the path back, the same request goes through.
    result(&w.card(Some(S_B), &["release", &b, "--generation", "1"]));
    let done = result(&w.widen(&a, &["apps/bo/", "apps/web/"], Some(WHY)));
    assert_eq!(
        done["write_scope"],
        json!(["apps/bo/", "apps/web/"]),
        "{done}"
    );
}

#[test]
fn narrowing_a_held_card_is_refused() {
    let w = world();
    let card = w.held(S_A, "narrow", &["apps/bo/", "apps/web/"]);
    let before = w.dump();
    for wanted in [
        &["apps/bo/", "docs/"][..],
        &["docs/"][..],
        &["apps/bo/"][..],
    ] {
        let (kind, message) = refusal(&w.widen(&card, wanted, Some(WHY)));
        assert_eq!(kind, "conflict", "{wanted:?}: {message}");
        assert!(
            message.contains("only grows"),
            "says why: {wanted:?}: {message}"
        );
        assert_eq!(w.dump(), before, "{wanted:?}: nothing was written");
    }
    let (_, message) = refusal(&w.widen(&card, &["apps/bo/", "docs/"], Some(WHY)));
    assert!(
        message.contains("apps/web/"),
        "names what is left out: {message}"
    );
    assert!(!message.contains("docs/"), "and only that: {message}");
}

#[test]
fn a_widening_that_cannot_be_recorded_is_invalid_input() {
    let w = world();
    let card = w.held(S_A, "bad", &["apps/bo/"]);
    let before = w.dump();
    let key = format!("{}{}", ["s", "k-"].concat(), "a1B2".repeat(6));
    let secret = format!("key {key} is in the shell");
    let long = "x".repeat(501);
    let wanted = ["apps/bo/", "docs/"];
    let cases: Vec<(&str, Option<&str>, &[&str])> = vec![
        ("no reason", None, &wanted),
        ("blank reason", Some("   "), &wanted),
        ("two lines", Some("one\ntwo"), &wanted),
        ("501 characters", Some(&long), &wanted),
        ("secret shape", Some(&secret), &wanted),
        ("nothing new", Some(WHY), &["apps/bo/"]),
        ("absolute path", Some(WHY), &["apps/bo/", "/etc/passwd"]),
        ("parent folder", Some(WHY), &["apps/bo/", "../x/"]),
    ];
    for (what, reason, scope) in cases {
        let (kind, message) = refusal(&w.widen(&card, scope, reason));
        assert_eq!(kind, "invalid_input", "{what}: {message}");
        assert!(
            !message.contains(&key),
            "{what}: the secret is not repeated: {message}"
        );
        assert_eq!(w.dump(), before, "{what}: nothing was written");
    }
}

#[test]
fn the_widened_scope_has_at_most_64_patterns() {
    let w = world();
    let names: Vec<String> = (0..63).map(|n| format!("d{n:02}/")).collect();
    let refs: Vec<&str> = names.iter().map(String::as_str).collect();
    let card = w.held(S_A, "many", &refs);
    assert_eq!(w.patterns(1).len(), 63);
    // 64: accepted.
    let mut wanted: Vec<&str> = refs.clone();
    wanted.push("e64/");
    let done = result(&w.widen(&card, &wanted, Some(WHY)));
    assert_eq!(done["write_scope"].as_array().unwrap().len(), 64, "{done}");
    // 65: refused, nothing written.
    let before = w.dump();
    wanted.push("e65/");
    let (kind, message) = refusal(&w.widen(&card, &wanted, Some(WHY)));
    assert_eq!(kind, "invalid_input", "{message}");
    assert_eq!(w.dump(), before);
    assert_eq!(w.patterns(1).len(), 64);
}

#[test]
fn a_card_nobody_holds_keeps_the_old_scope_rules() {
    let w = world();
    let added = result(&w.card(
        None,
        &[
            "add", "--title", "plain", "--scope", "apps/bo/", "--ac", "one",
        ],
    ));
    let card = added["name"].as_str().unwrap().to_string();
    result(&w.card(None, &["ready", &card]));
    // Replaced as a whole, no reason needed, one event with `from` and `to` only.
    let done = result(&w.widen(&card, &["docs/"], None));
    assert_eq!(done["from"], json!(["apps/bo/"]), "{done}");
    assert_eq!(done["write_scope"], json!(["docs/"]), "{done}");
    assert!(
        done.get("held").is_none(),
        "an unheld card has no `held`: {done}"
    );
    assert_eq!(w.own_scope(&card), json!(["docs/"]));
    let events = w.scope_events();
    assert_eq!(events.len(), 1, "{events:?}");
    assert_eq!(events[0].1, json!({"from": ["apps/bo/"], "to": ["docs/"]}));
    // A reason, when one is given, is recorded.
    let done = result(&w.widen(&card, &["src/"], Some(WHY)));
    assert_eq!(done["write_scope"], json!(["src/"]), "{done}");
    assert_eq!(w.scope_events()[1].1["reason"], json!(WHY));
    // Held and then given back, it is replaceable again.
    result(&w.card(Some(S_A), &["take", &card]));
    result(&w.card(Some(S_A), &["release", &card, "--generation", "1"]));
    let done = result(&w.widen(&card, &["tests/"], None));
    assert_eq!(done["write_scope"], json!(["tests/"]), "{done}");
}

/// Two lanes widen toward the same path at the same moment: the ledger serializes the writes,
/// so exactly one wins and the other is refused by the winner's card name.
#[test]
fn two_widenings_toward_one_path_have_one_winner() {
    let w = world();
    let a = w.held(S_A, "a", &["apps/bo/"]);
    let b = w.held(S_B, "b", &["apps/web/"]);
    let spawn = |card: &str, keep: &str| {
        let mut args = vec!["card", "scope", card, "--scope", keep, "--scope", "shared/"];
        args.extend_from_slice(&["--by", "leader", "--reason", WHY, "--project", &w.project]);
        w.home
            .command(&args)
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .expect("run moai")
    };
    let first = spawn(&a, "apps/bo/");
    let second = spawn(&b, "apps/web/");
    let outcomes = [
        first.wait_with_output().unwrap(),
        second.wait_with_output().unwrap(),
    ];
    let won: Vec<&Output> = outcomes
        .iter()
        .filter(|o| o.status.code() == Some(0))
        .collect();
    let lost: Vec<&Output> = outcomes
        .iter()
        .filter(|o| o.status.code() == Some(1))
        .collect();
    assert_eq!((won.len(), lost.len()), (1, 1), "{outcomes:?}");
    let (kind, message) = refusal(lost[0]);
    assert_eq!(kind, "conflict", "{message}");
    assert!(
        message.contains("C-1") || message.contains("C-2"),
        "names the winner: {message}"
    );
    let holders = w.rows(
        "SELECT r.card_id FROM reservation_scope s JOIN reservation r \
         ON r.reservation_id = s.reservation_id WHERE s.pattern = 'shared/'",
    );
    assert_eq!(
        holders.len(),
        1,
        "one reservation holds shared/: {holders:?}"
    );
}

#[test]
fn the_help_and_the_ops_doc_name_the_rule() {
    let w = world();
    let out = w
        .home
        .command(&["card", "scope", "--help"])
        .output()
        .expect("run moai");
    let help = String::from_utf8_lossy(&out.stdout);
    // The help is wrapped at the terminal width: compare words, not line breaks.
    let help = help.split_whitespace().collect::<Vec<_>>().join(" ");
    for phrase in [
        "only grows",
        "`--reason`",
        "the same generation",
        "the card's own scope is not changed",
    ] {
        assert!(help.contains(phrase), "{phrase} missing: {help}");
    }
    let doc = std::fs::read_to_string(
        std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs/design-review/appendix/operations/card.md"),
    )
    .unwrap();
    let row = doc
        .lines()
        .find(|line| line.starts_with("| `card scope"))
        .expect("the card scope row");
    for phrase in ["넓히기만", "`--reason`", "같은 세대", "카드 자체의 범위"] {
        assert!(row.contains(phrase), "{phrase} missing in the row: {row}");
    }
}

// ---- C-94 follow-up (leader note 2860, audit N4·N5): the guarantees the first slice's tests
// ---- did not pin. Each test below is shown to fail under its own mutation (SPEC Result).

/// The reason limit counts characters, not bytes: 500 ASCII and 500 Korean characters (1500
/// bytes) are taken, 501 of either is refused naming the limit.
#[test]
fn a_reason_of_500_characters_is_taken_and_501_refused_in_ascii_and_korean() {
    let w = world();
    let card = w.held(S_A, "reason", &["apps/bo/"]);
    let ascii_ok = "x".repeat(500);
    let ascii_over = "x".repeat(501);
    let ko_ok = "가".repeat(500);
    let ko_over = "가".repeat(501);
    let done = result(&w.widen(&card, &["apps/bo/", "d1/"], Some(&ascii_ok)));
    assert_eq!(done["reason"], json!(ascii_ok), "{done}");
    // The next widening keeps the patterns the card already grew to.
    let done = result(&w.widen(&card, &["apps/bo/", "d1/", "d2/"], Some(&ko_ok)));
    assert_eq!(done["reason"], json!(ko_ok), "{done}");
    let before = w.dump();
    let keep = &["apps/bo/", "d1/", "d2/"][..];
    for (what, r) in [("ascii", &ascii_over), ("korean", &ko_over)] {
        let mut wanted = keep.to_vec();
        wanted.push("d3/");
        let (kind, message) = refusal(&w.widen(&card, &wanted, Some(r)));
        assert_eq!(kind, "invalid_input", "{what}: {message}");
        assert!(message.contains("500"), "{what}: {message}");
    }
    assert_eq!(w.dump(), before, "the refused widenings wrote nothing");
    assert_eq!(w.current_patterns(1), vec!["apps/bo/", "d1/", "d2/"]);
}

/// SPEC D-2: widening needs no holder check — whoever signs `--by` may widen, and the event
/// names that actor.
#[test]
fn someone_who_is_not_the_holder_can_widen_as_the_spec_decides() {
    let w = world();
    let card = w.held(S_A, "deputy", &["apps/bo/"]);
    let mut args = vec![
        "scope",
        &card[..],
        "--scope",
        "apps/bo/",
        "--scope",
        "docs/",
    ];
    args.extend_from_slice(&["--by", "another-leader", "--reason", WHY]);
    let done = result(&w.card(None, &args));
    assert_eq!(done["held"], json!(true), "{done}");
    assert_eq!(done["holder"], json!(S_A), "{done}");
    assert_eq!(done["by"], json!("another-leader"), "{done}");
    assert_eq!(w.patterns(1), vec!["apps/bo/", "docs/"]);
    let events = w.scope_events();
    assert_eq!(events.len(), 1, "{events:?}");
    assert_eq!(events[0].0, "another-leader");
}

/// The widening does not look at the lease: an aged-out lease still grows the same
/// reservation at the same generation.
#[test]
fn a_widen_on_an_ended_lease_still_grows_the_scope() {
    let w = world();
    let card = w.held(S_A, "aged", &["apps/bo/"]);
    w.exec("UPDATE reservation SET lease_until = 1");
    let shown = w.reservation(&card);
    assert!(
        shown["lease_until"].as_i64().unwrap() <= 1,
        "aged out: {shown}"
    );
    let done = result(&w.widen(&card, &["apps/bo/", "docs/"], Some(WHY)));
    assert_eq!(done["held"], json!(true), "{done}");
    assert_eq!(done["generation"], json!(1), "{done}");
    assert_eq!(w.patterns(1), vec!["apps/bo/", "docs/"]);
}

/// The widened row flows through handoff, reclaim and release: every step processes it and
/// the scope stays the widened one.
#[test]
fn the_widened_row_survives_handoff_reclaim_and_release() {
    let w = world();
    let card = w.held(S_A, "flow", &["apps/bo/"]);
    let widened = json!(["apps/bo/", "apps/web/"]);
    result(&w.widen(&card, &["apps/bo/", "apps/web/"], Some(WHY)));
    result(&w.card(
        Some(S_A),
        &["handoff", &card, "--generation", "1", "--to", S_B],
    ));
    let shown = w.reservation(&card);
    assert_eq!(shown["generation"], json!(2), "{shown}");
    assert_eq!(shown["holder"], json!(S_B), "{shown}");
    assert_eq!(shown["write_scope"], widened, "{shown}");
    assert_eq!(w.current_patterns(1), vec!["apps/bo/", "apps/web/"]);
    // The lease ends; a third session reclaims: generation 3, scope kept.
    w.exec("UPDATE reservation SET lease_until = 1");
    result(&w.card(
        Some(S_C),
        &[
            "reclaim",
            &card,
            "--generation",
            "2",
            "--by",
            "leader",
            "--reason",
            "the lease ended and the follow-up needs the wider scope",
        ],
    ));
    let shown = w.reservation(&card);
    assert_eq!(shown["generation"], json!(3), "{shown}");
    assert_eq!(shown["holder"], json!(S_C), "{shown}");
    assert_eq!(shown["write_scope"], widened, "{shown}");
    result(&w.card(Some(S_C), &["release", &card, "--generation", "3"]));
    let statuses: Vec<Vec<Value>> =
        w.rows("SELECT status FROM reservation WHERE card_id = 1 ORDER BY generation");
    assert_eq!(
        statuses,
        vec![
            vec![json!("released")],
            vec![json!("reclaimed")],
            vec![json!("released")],
        ],
        "each generation ended the way its step ends it"
    );
    assert_eq!(w.current_patterns(1), vec!["apps/bo/", "apps/web/"]);
}

/// The overlap refusal compares without letter case. The two spellings of one pattern can
/// never coexist — reserve itself refuses a second card over a case-variant of a held
/// pattern — so the pins are: a widen into the upper-case spelling of another card's
/// lower-case pattern, and a take of the upper-case spelling while the lower-case one is
/// held (both refused, each naming the card that holds the case-variant).
#[test]
fn an_overlap_refusal_ignores_letter_case() {
    let w = world();
    let up = w.held(S_A, "up", &["APPS/WEB/"]);
    // Widen direction: a card holding a disjoint scope adds the lower-case spelling.
    let added = result(&w.card(
        None,
        &["add", "--title", "low", "--scope", "docs/", "--ac", "one"],
    ));
    let low = added["name"].as_str().unwrap().to_string();
    result(&w.card(None, &["ready", &low]));
    result(&w.card(
        Some(S_B),
        &[
            "take",
            &low,
            "--worktree",
            "/work/tree",
            "--branch",
            "card/low",
        ],
    ));
    let before = w.dump();
    let (kind, message) = refusal(&w.widen(&low, &["docs/", "apps/web/"], Some(WHY)));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("C-1"), "{message}");
    assert!(message.contains("apps/web/"), "{message}");
    assert_eq!(w.dump(), before, "nothing was written");
    // Reserve direction: taking another card over the lower-case spelling of a held pattern.
    let other = result(&w.card(
        None,
        &[
            "add",
            "--title",
            "other",
            "--scope",
            "apps/web/",
            "--ac",
            "one",
        ],
    ));
    let other = other["name"].as_str().unwrap().to_string();
    result(&w.card(None, &["ready", &other]));
    let (kind, message) = refusal(&w.card(
        Some(S_B),
        &[
            "take",
            &other,
            "--worktree",
            "/work/tree",
            "--branch",
            "card/other",
        ],
    ));
    assert_eq!(kind, "conflict", "{message}");
    assert!(
        message.contains("C-1"),
        "names the card holding the case-variant: {message}"
    );
    let _ = up;
}

/// The widen is one transaction: when the card.scope event cannot be written, the whole
/// widening is refused and every row is as it was. The failure is injected test-side (a
/// trigger that aborts the event insert), so no src change is needed.
#[test]
fn a_widen_whose_event_cannot_be_written_leaves_the_row_as_it_was() {
    let w = world();
    let card = w.held(S_A, "atomic", &["apps/bo/"]);
    w.exec(
        "CREATE TRIGGER block_scope_event BEFORE INSERT ON event \
         WHEN new.kind = 'card.scope' BEGIN SELECT RAISE(ABORT, 'injected by the test'); END;",
    );
    let before = w.dump();
    let (_, message) = refusal(&w.widen(&card, &["apps/bo/", "docs/"], Some(WHY)));
    assert_eq!(w.dump(), before, "the failed widen wrote nothing");
    assert_eq!(w.patterns(1), vec!["apps/bo/"]);
    assert!(
        !message.contains(WHY),
        "the reason is not echoed: {message}"
    );
    // Without the injection the same widening goes through, so the refusal was the trigger.
    w.exec("DROP TRIGGER block_scope_event;");
    let done = result(&w.widen(&card, &["apps/bo/", "docs/"], Some(WHY)));
    assert_eq!(done["held"], json!(true), "{done}");
    assert_eq!(w.patterns(1), vec!["apps/bo/", "docs/"]);
}
