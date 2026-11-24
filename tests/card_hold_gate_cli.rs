//! A structured HOLD and a prerequisite on a card are read by `card next` and `card take` before
//! any reservation (SPEC-20261121-card-hold-gate). The HOLD is a `card.hold` event, lifted only by
//! a `card.unhold` event; a prerequisite is an active `blocks` link from the card that has to be
//! done first. Free text in a note or a body is never read. Every command runs against a
//! temporary HOME; the real home state is not read or written.

mod common;

use common::home::TempHome;
use moai_cli::ledger::{self, workflow};
use serde_json::{json, Value};
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const CATEGORIES: [&str; 6] = [
    "money",
    "legal",
    "external_contract",
    "secret",
    "real_data",
    "other",
];

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().expect("run moai")
}

fn run_with_stdin(home: &TempHome, args: &[&str], stdin: &str) -> Output {
    use std::io::Write;
    let mut child = home
        .command(args)
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()
        .expect("spawn moai");
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.as_bytes())
        .unwrap();
    child.wait_with_output().expect("wait for moai")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {}",
            String::from_utf8_lossy(&out.stdout)
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
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
}

struct Fixture {
    home: TempHome,
    project: String,
}

impl Fixture {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        Fixture {
            home,
            project: project.to_str().unwrap().to_string(),
        }
    }

    /// A `ready` card with one criterion and its own scope; returns its number.
    fn card(&self, scope: &str) -> i64 {
        let added = result(&run(
            &self.home,
            &[
                "card",
                "add",
                "--title",
                "t",
                "--scope",
                scope,
                "--ac",
                "it works",
                "--project",
                &self.project,
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&run(
            &self.home,
            &["card", "ready", &name, "--project", &self.project],
        ));
        added["card_id"].as_i64().unwrap()
    }

    fn next(&self) -> Value {
        result(&as_session(&self.home, &["card", "next", &self.project]))
    }

    fn take(&self, card_id: i64) -> Output {
        as_session(
            &self.home,
            &[
                "card",
                "take",
                &format!("C-{card_id}"),
                "--project",
                &self.project,
            ],
        )
    }

    fn show(&self, card_id: i64) -> Value {
        result(&run(
            &self.home,
            &[
                "card",
                "show",
                &format!("C-{card_id}"),
                "--project",
                &self.project,
            ],
        ))
    }

    fn hold(&self, card_id: i64, category: &str, reason: &str) -> Result<(), workflow::HoldError> {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        workflow::hold(&mut w, card_id, category, reason, "leader", Some(S_A))
    }

    fn unhold(&self, card_id: i64) -> Result<(), workflow::HoldError> {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        workflow::unhold(&mut w, card_id, "decided", "leader", Some(S_A))
    }

    fn needs(&self, card_id: i64, needs: i64) -> Result<bool, workflow::PrereqError> {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        workflow::add_prerequisite(&mut w, card_id, needs, "leader", Some(S_A))
    }

    fn count(&self, query: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(query, [], |r| r.get(0)).unwrap()
    }

    /// Closes a card the way a landing would, without the guards a test ledger does not need.
    fn close(&self, card_id: i64) {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.with_write_tx(|tx, _| {
            tx.execute_batch(&format!(
                "DROP TRIGGER IF EXISTS card_state_transition; \
                 DROP TRIGGER IF EXISTS card_done_guard; \
                 UPDATE card SET state = 'done', closed_at = unixepoch() WHERE card_id = {card_id}"
            ))?;
            Ok(())
        })
        .unwrap();
    }
}

fn as_session(home: &TempHome, args: &[&str]) -> Output {
    home.command(args)
        .env(SESSION_VAR, S_A)
        .output()
        .expect("run moai")
}

fn names(list: &Value) -> Vec<String> {
    list.as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap().to_string())
        .collect()
}

fn skipped<'a>(next: &'a Value, card_id: i64) -> &'a Value {
    next["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["card_id"] == json!(card_id))
        .unwrap_or_else(|| panic!("C-{card_id} is not in skipped: {next}"))
}

#[test]
fn a_ready_card_takes_a_hold_of_each_category_and_the_event_keeps_the_fields() {
    let f = Fixture::new();
    for (i, category) in CATEGORIES.iter().enumerate() {
        let id = f.card(&format!("src/c{i}.rs"));
        f.hold(id, category, &format!("because {category}"))
            .unwrap();
        let payload: String = {
            let l = ledger::open_read_only(f.home.path(), "demo").unwrap();
            l.conn()
                .query_row(
                    "SELECT payload FROM event WHERE card_id = ?1 AND kind = 'card.hold'",
                    [id],
                    |r| r.get(0),
                )
                .unwrap()
        };
        let p: Value = serde_json::from_str(&payload).unwrap();
        assert_eq!(p["category"], json!(category), "{p}");
        assert_eq!(p["reason"], json!(format!("because {category}")), "{p}");
        assert_eq!(p["by"], json!("leader"), "{p}");
    }
    assert_eq!(
        f.count("SELECT count(*) FROM event WHERE kind = 'card.hold'"),
        6
    );
}

#[test]
fn a_hold_is_refused_for_a_bad_category_a_taken_card_and_a_card_already_on_hold() {
    let f = Fixture::new();
    let a = f.card("src/a.rs");
    let taken = f.card("src/b.rs");
    result(&f.take(taken));
    let on_hold = f.card("src/c.rs");
    f.hold(on_hold, "money", "first").unwrap();
    let events = f.count("SELECT count(*) FROM event");

    assert!(matches!(
        f.hold(a, "feelings", "x"),
        Err(workflow::HoldError::InvalidInput(_))
    ));
    assert!(matches!(
        f.hold(taken, "money", "x"),
        Err(workflow::HoldError::Held { .. })
    ));
    assert!(matches!(
        f.hold(on_hold, "legal", "second"),
        Err(workflow::HoldError::AlreadyOnHold { .. })
    ));
    assert!(matches!(
        f.hold(999, "money", "x"),
        Err(workflow::HoldError::CardNotFound(999))
    ));
    f.close(a);
    assert!(matches!(
        f.hold(a, "money", "x"),
        Err(workflow::HoldError::WrongState { .. })
    ));
    assert_eq!(f.count("SELECT count(*) FROM event"), events);
}

#[test]
fn unhold_appends_its_event_and_the_card_is_offered_again() {
    let f = Fixture::new();
    let id = f.card("src/a.rs");
    assert!(matches!(
        f.unhold(id),
        Err(workflow::HoldError::NotOnHold { .. })
    ));
    f.hold(id, "legal", "wait for counsel").unwrap();
    assert_eq!(f.next()["next"], Value::Null);

    f.unhold(id).unwrap();
    assert_eq!(
        f.count("SELECT count(*) FROM event WHERE kind = 'card.unhold'"),
        1
    );
    assert_eq!(f.next()["next"], json!(format!("C-{id}")));
    // The hold can be set again after it was lifted.
    f.hold(id, "money", "again").unwrap();
    assert_eq!(f.next()["next"], Value::Null);
}

#[test]
fn nothing_but_unhold_lifts_a_hold_and_free_text_changes_nothing() {
    let f = Fixture::new();
    let id = f.card("src/a.rs");
    // Free text that says HOLD never makes a card held ...
    let quiet = f.card("src/b.rs");
    result(&run_with_stdin(
        &f.home,
        &[
            "card",
            "note",
            &format!("C-{quiet}"),
            "--kind",
            "comment",
            "--project",
            &f.project,
        ],
        "HOLD: wait for the budget\n",
    ));
    let next = f.next();
    assert!(
        names(&next["ready"]).contains(&format!("C-{quiet}")),
        "{next}"
    );
    result(&f.take(quiet));

    // ... and free text that says the HOLD is over never lifts a structured one.
    f.hold(id, "money", "budget").unwrap();
    result(&run_with_stdin(
        &f.home,
        &[
            "card",
            "note",
            &format!("C-{id}"),
            "--kind",
            "comment",
            "--project",
            &f.project,
        ],
        "the HOLD is lifted, go on\n",
    ));
    let next = f.next();
    assert_eq!(skipped(&next, id)["reason"], json!("hold"), "{next}");
    refusal(&f.take(id));
    assert_eq!(
        f.count("SELECT count(*) FROM event WHERE kind = 'card.unhold'"),
        0
    );
}

#[test]
fn next_skips_a_held_card_with_its_category_and_reason() {
    let f = Fixture::new();
    let held = f.card("src/a.rs");
    let free = f.card("src/b.rs");
    f.hold(held, "external_contract", "vendor has not signed")
        .unwrap();

    let next = f.next();
    assert_eq!(names(&next["ready"]), vec![format!("C-{free}")], "{next}");
    assert_eq!(next["next"], json!(format!("C-{free}")), "{next}");
    let s = skipped(&next, held);
    assert_eq!(s["reason"], json!("hold"), "{s}");
    assert_eq!(s["category"], json!("external_contract"), "{s}");
    assert_eq!(s["hold_reason"], json!("vendor has not signed"), "{s}");
}

#[test]
fn a_prerequisite_is_added_once_and_refuses_a_self_need_and_a_two_card_cycle() {
    let f = Fixture::new();
    let a = f.card("src/a.rs");
    let b = f.card("src/b.rs");
    assert_eq!(f.needs(a, b), Ok(true));
    let links = f.count("SELECT count(*) FROM card_link WHERE kind = 'blocks'");
    let events = f.count("SELECT count(*) FROM event");
    // The link runs from the card that must be done first to the card that waits.
    assert_eq!(
        f.count(&format!(
            "SELECT count(*) FROM card_link WHERE kind = 'blocks' AND state = 'active' \
             AND from_card = {b} AND to_card = {a}"
        )),
        1
    );

    assert_eq!(f.needs(a, b), Ok(false));
    assert!(matches!(
        f.needs(a, a),
        Err(workflow::PrereqError::SelfNeed(_))
    ));
    assert!(matches!(
        f.needs(b, a),
        Err(workflow::PrereqError::Cycle { .. })
    ));
    assert!(matches!(
        f.needs(a, 999),
        Err(workflow::PrereqError::CardNotFound(999))
    ));
    assert_eq!(
        f.count("SELECT count(*) FROM card_link WHERE kind = 'blocks'"),
        links
    );
    assert_eq!(f.count("SELECT count(*) FROM event"), events);
}

#[test]
fn next_skips_a_card_whose_prerequisite_is_not_done_and_offers_it_when_it_is() {
    let f = Fixture::new();
    let waits = f.card("src/a.rs");
    let first = f.card("src/b.rs");
    let second = f.card("src/c.rs");
    f.needs(waits, first).unwrap();
    f.needs(waits, second).unwrap();

    let next = f.next();
    assert!(
        !names(&next["ready"]).contains(&format!("C-{waits}")),
        "{next}"
    );
    let s = skipped(&next, waits);
    assert_eq!(s["reason"], json!("prerequisite"), "{s}");
    assert_eq!(
        s["needs"],
        json!([format!("C-{first}"), format!("C-{second}")]),
        "{s}"
    );

    f.close(first);
    let s = skipped(&f.next(), waits).clone();
    assert_eq!(s["needs"], json!([format!("C-{second}")]), "{s}");

    f.close(second);
    let next = f.next();
    assert!(
        names(&next["ready"]).contains(&format!("C-{waits}")),
        "{next}"
    );
}

#[test]
fn take_refuses_a_held_card_and_a_card_with_an_open_prerequisite_and_writes_nothing() {
    let f = Fixture::new();
    let held = f.card("src/a.rs");
    let waits = f.card("src/b.rs");
    let first = f.card("src/c.rs");
    f.hold(held, "secret", "token not issued").unwrap();
    f.needs(waits, first).unwrap();
    let before = f.home.snapshot();

    let message = refusal(&f.take(held));
    assert!(message.starts_with("card take refused: "), "{message}");
    assert!(message.contains("secret"), "{message}");
    assert!(message.contains("token not issued"), "{message}");
    assert!(message.contains("nothing was taken"), "{message}");
    let message = refusal(&f.take(waits));
    assert!(message.starts_with("card take refused: "), "{message}");
    assert!(message.contains(&format!("C-{first}")), "{message}");
    assert!(message.contains("nothing was taken"), "{message}");

    assert_eq!(f.home.snapshot(), before, "the ledger changed");
    assert_eq!(f.count("SELECT count(*) FROM reservation"), 0);

    // Once the hold is lifted and the prerequisite is done, the same takes go through.
    f.unhold(held).unwrap();
    f.close(first);
    result(&f.take(held));
    result(&f.take(waits));
}

#[test]
fn show_carries_the_hold_and_the_unfinished_prerequisites() {
    let f = Fixture::new();
    let id = f.card("src/a.rs");
    let first = f.card("src/b.rs");
    let shown = f.show(id);
    assert_eq!(shown["hold"], Value::Null, "{shown}");
    assert_eq!(shown["prerequisites"], json!([]), "{shown}");

    f.hold(id, "real_data", "needs the real export").unwrap();
    f.needs(id, first).unwrap();
    let shown = f.show(id);
    assert_eq!(shown["hold"]["category"], json!("real_data"), "{shown}");
    assert_eq!(
        shown["hold"]["reason"],
        json!("needs the real export"),
        "{shown}"
    );
    assert_eq!(shown["hold"]["by"], json!("leader"), "{shown}");
    assert!(shown["hold"]["at"].as_i64().unwrap() > 0, "{shown}");
    assert_eq!(
        shown["prerequisites"],
        json!([format!("C-{first}")]),
        "{shown}"
    );

    f.close(first);
    assert_eq!(f.show(id)["prerequisites"], json!([]));
}
