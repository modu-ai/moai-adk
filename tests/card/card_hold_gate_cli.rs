//! A structured HOLD and a prerequisite on a card are read by `card next` and `card take` before
//! any reservation (SPEC-20261121-card-hold-gate). The HOLD is a `card.hold` event, lifted only by
//! a `card.unhold` event; a prerequisite is an active `blocks` link from the card that has to be
//! done first. Free text in a note or a body is never read. Every command runs against a
//! temporary HOME; the real home state is not read or written.

use crate::common;

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

fn skipped(next: &Value, card_id: i64) -> &Value {
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

// ---- the commands ------------------------------------------------------------------------------

impl Fixture {
    fn cli(&self, args: &[&str]) -> Output {
        let mut full: Vec<&str> = args.to_vec();
        full.extend_from_slice(&["--project", &self.project]);
        as_session(&self.home, &full)
    }

    fn note(&self, card_id: i64, text: &str) -> Value {
        result(&run_with_stdin(
            &self.home,
            &[
                "card",
                "note",
                &format!("C-{card_id}"),
                "--kind",
                "comment",
                "--project",
                &self.project,
            ],
            text,
        ))
    }
}

#[test]
fn card_hold_takes_each_category_and_refuses_a_bad_one_writing_nothing() {
    let f = Fixture::new();
    for (i, category) in CATEGORIES.iter().enumerate() {
        let id = f.card(&format!("src/c{i}.rs"));
        let answer = result(&f.cli(&[
            "card",
            "hold",
            &format!("C-{id}"),
            "--category",
            category,
            "--reason",
            "why",
            "--by",
            "leader",
        ]));
        assert_eq!(answer["category"], json!(category), "{answer}");
        assert_eq!(f.show(id)["hold"]["category"], json!(category));
    }
    let id = f.card("src/z.rs");
    let events = f.count("SELECT count(*) FROM event");
    let message = refusal(&f.cli(&[
        "card",
        "hold",
        &format!("C-{id}"),
        "--category",
        "mood",
        "--reason",
        "why",
        "--by",
        "leader",
    ]));
    assert!(message.starts_with("card hold refused: "), "{message}");
    assert!(message.contains("external_contract"), "{message}");
    assert_eq!(f.count("SELECT count(*) FROM event"), events);
}

#[test]
fn card_unhold_lifts_only_a_hold_in_force_and_the_card_is_offered_again() {
    let f = Fixture::new();
    let id = f.card("src/a.rs");
    let name = format!("C-{id}");
    let unhold = [
        "card",
        "unhold",
        name.as_str(),
        "--reason",
        "decided",
        "--by",
        "leader",
    ];
    let message = refusal(&f.cli(&unhold));
    assert!(message.starts_with("card unhold refused: "), "{message}");

    result(&f.cli(&[
        "card",
        "hold",
        &name,
        "--category",
        "money",
        "--reason",
        "budget",
        "--by",
        "leader",
    ]));
    assert_eq!(f.next()["next"], Value::Null);
    result(&f.cli(&unhold));
    assert_eq!(f.next()["next"], json!(name));
    assert_eq!(f.show(id)["hold"], Value::Null);
}

#[test]
fn card_prereq_adds_one_link_once_and_refuses_a_self_need_and_a_cycle() {
    let f = Fixture::new();
    let a = f.card("src/a.rs");
    let b = f.card("src/b.rs");
    let (na, nb) = (format!("C-{a}"), format!("C-{b}"));
    let prereq = |card: &str, needs: &str| {
        f.cli(&["card", "prereq", card, "--needs", needs, "--by", "leader"])
    };
    let first = result(&prereq(&na, &nb));
    assert_eq!(first["created"], json!(true), "{first}");
    let again = result(&prereq(&na, &nb));
    assert_eq!(again["created"], json!(false), "{again}");
    assert_eq!(again["writes"], json!(false), "{again}");
    assert_eq!(
        f.count("SELECT count(*) FROM card_link WHERE kind = 'blocks'"),
        1
    );

    assert!(refusal(&prereq(&na, &na)).contains("cannot need itself"));
    assert!(refusal(&prereq(&nb, &na)).contains("wait for each other"));
    assert_eq!(
        f.count("SELECT count(*) FROM card_link WHERE kind = 'blocks'"),
        1
    );
    assert_eq!(f.show(a)["prerequisites"], json!([nb]));
}

#[test]
fn hold_candidates_lists_free_text_holds_read_only_until_a_person_moves_one() {
    let f = Fixture::new();
    let in_body = {
        let added = result(&run(
            &f.home,
            &[
                "card",
                "add",
                "--title",
                "t",
                "--scope",
                "src/a.rs",
                "--ac",
                "ok",
                "--body",
                "line one\nHOLD: waiting for the contract",
                "--project",
                &f.project,
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&run(
            &f.home,
            &["card", "ready", &name, "--project", &f.project],
        ));
        added["card_id"].as_i64().unwrap()
    };
    let in_note = f.card("src/b.rs");
    let clean = f.card("src/c.rs");
    let lowercase = f.card("src/d.rs");
    f.note(in_note, "earlier note\n");
    let note = f.note(in_note, "status\nHOLD: user said wait\n");
    let note_id = note["note_id"].as_i64().unwrap();
    f.note(
        lowercase,
        "we may hold this and hold it again; HOLDS and NO_HOLD are not it\n",
    );
    let before = f.home.snapshot();

    let listed = result(&f.cli(&["card", "hold-candidates"]));
    assert_eq!(f.home.snapshot(), before, "the command wrote something");
    assert_eq!(listed["writes"], json!(false), "{listed}");
    assert_eq!(listed["count"], json!(2), "{listed}");
    let rows = listed["candidates"].as_array().unwrap();
    assert_eq!(rows[0]["card_id"], json!(in_body), "{listed}");
    assert_eq!(rows[0]["source"], json!("body"), "{listed}");
    assert_eq!(
        rows[0]["quote"],
        json!("HOLD: waiting for the contract"),
        "{listed}"
    );
    assert_eq!(rows[1]["card_id"], json!(in_note), "{listed}");
    assert_eq!(rows[1]["note_id"], json!(note_id), "{listed}");
    assert_eq!(rows[1]["quote"], json!("HOLD: user said wait"), "{listed}");
    assert!(
        rows.iter().all(|r| r["card_id"] != json!(clean)),
        "{listed}"
    );

    // Until a person moves one, next and take treat both as before.
    let next = f.next();
    let ready = names(&next["ready"]);
    assert!(ready.contains(&format!("C-{in_body}")), "{next}");
    assert!(ready.contains(&format!("C-{in_note}")), "{next}");

    result(&f.cli(&[
        "card",
        "hold",
        &format!("C-{in_body}"),
        "--category",
        "external_contract",
        "--reason",
        "from the candidate list",
        "--by",
        "leader",
    ]));
    let listed = result(&f.cli(&["card", "hold-candidates"]));
    assert_eq!(listed["count"], json!(1), "{listed}");
    assert_eq!(
        listed["candidates"][0]["card_id"],
        json!(in_note),
        "{listed}"
    );
    let next = f.next();
    assert!(
        !names(&next["ready"]).contains(&format!("C-{in_body}")),
        "{next}"
    );
    assert_eq!(skipped(&next, in_body)["reason"], json!("hold"), "{next}");
}

#[test]
fn guidance_tells_a_lane_to_read_hold_and_prerequisites_before_it_creates_a_worktree() {
    // Compiled in, not read at run time: the repository scan table counts a run-time read.
    let work = include_str!("../../skills/moai/workflows/work.md");
    for phrase in [
        "Read the HOLD and the prerequisites of a card before you create a worktree for it",
        "the reason `hold` or `prerequisite`",
        "whose shown record carries a `hold`",
        "gets no worktree",
        "only a person lifts it with the unhold command",
    ] {
        assert!(work.contains(phrase), "{phrase} missing in work.md");
    }
}

// ---- removable prerequisites, loops of any length, withdrawn prerequisites ---------------------
// (SPEC-20261127-card-prereq-remove)

impl Fixture {
    /// `card prereq C-n --needs C-m --remove --by leader`.
    fn remove_prereq(&self, card_id: i64, needs: i64) -> Output {
        self.cli(&[
            "card",
            "prereq",
            &format!("C-{card_id}"),
            "--needs",
            &format!("C-{needs}"),
            "--remove",
            "--by",
            "leader",
        ])
    }

    /// `card withdraw` and `card fold` take the project as a positional argument.
    fn withdraw(&self, card_id: i64) -> Value {
        result(&run(
            &self.home,
            &[
                "card",
                "withdraw",
                &format!("C-{card_id}"),
                &self.project,
                "--reason",
                "not needed any more",
                "--by",
                "leader",
            ],
        ))
    }

    fn fold(&self, card_id: i64, into: i64) -> Value {
        result(&run(
            &self.home,
            &[
                "card",
                "fold",
                &format!("C-{card_id}"),
                &self.project,
                "--into",
                &format!("C-{into}"),
                "--reason",
                "the work moved",
                "--by",
                "leader",
            ],
        ))
    }

    /// Puts cards in a state without the transition guards a test ledger does not need (a card in
    /// `hold` this way has no `card.withdraw` or `card.fold` event). One write for all of them:
    /// reopening a ledger whose triggers were dropped makes a backup named by the second.
    fn set_states(&self, states: &[(i64, &str)]) {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.with_write_tx(|tx, _| {
            tx.execute_batch(
                "DROP TRIGGER IF EXISTS card_state_transition; \
                 DROP TRIGGER IF EXISTS card_done_guard;",
            )?;
            for (card_id, state) in states {
                tx.execute_batch(&format!(
                    "UPDATE card SET state = '{state}', \
                     closed_at = CASE WHEN '{state}' = 'done' THEN unixepoch() ELSE closed_at END \
                     WHERE card_id = {card_id}"
                ))?;
            }
            Ok(())
        })
        .unwrap();
    }

    fn unfinished(&self, card_id: i64) -> Vec<i64> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        workflow::unfinished_prerequisites(l.conn(), card_id).unwrap()
    }

    fn link_rows(&self) -> i64 {
        self.count("SELECT count(*) FROM card_link WHERE kind = 'blocks'")
    }
}

#[test]
fn prereq_ac1_a_loop_of_any_length_is_refused_naming_every_card_and_writing_nothing() {
    let f = Fixture::new();
    let a = f.card("src/a.rs");
    let b = f.card("src/b.rs");
    let c = f.card("src/c.rs");
    let d = f.card("src/d.rs");
    let e = f.card("src/e.rs");
    // a needs b, b needs c: `c needs a` would make c wait for a, a for b and b for c.
    assert_eq!(f.needs(a, b), Ok(true));
    assert_eq!(f.needs(b, c), Ok(true));
    let (links, events) = (f.link_rows(), f.count("SELECT count(*) FROM event"));
    let err = f.needs(c, a).unwrap_err();
    assert!(
        matches!(err, workflow::PrereqError::Cycle { .. }),
        "{err:?}"
    );
    let words = err.to_string();
    assert!(
        words.contains(&format!("C-{c} -> C-{a} -> C-{b} -> C-{c}")),
        "{words}"
    );
    assert_eq!(f.link_rows(), links);
    assert_eq!(f.count("SELECT count(*) FROM event"), events);

    // Four cards: c needs d, so `d needs a` closes d -> a -> b -> c -> d. An unrelated card is not
    // named and is not refused.
    assert_eq!(f.needs(c, d), Ok(true));
    let (links, events) = (f.link_rows(), f.count("SELECT count(*) FROM event"));
    let words = f.needs(d, a).unwrap_err().to_string();
    assert!(
        words.contains(&format!("C-{d} -> C-{a} -> C-{b} -> C-{c} -> C-{d}")),
        "{words}"
    );
    assert!(!words.contains(&format!("C-{e}")), "{words}");
    assert_eq!(f.link_rows(), links);
    assert_eq!(f.count("SELECT count(*) FROM event"), events);
    assert_eq!(f.needs(e, a), Ok(true));

    // The two-card refusal keeps its words.
    let two = f.needs(b, a).unwrap_err().to_string();
    assert_eq!(
        two,
        format!("C-{a} already needs C-{b}, so they would wait for each other")
    );
}

#[test]
fn prereq_ac1_a_retired_link_does_not_close_a_loop() {
    let f = Fixture::new();
    let x = f.card("src/x.rs");
    let y = f.card("src/y.rs");
    let z = f.card("src/z.rs");
    assert_eq!(f.needs(x, y), Ok(true));
    assert_eq!(f.needs(y, z), Ok(true));
    result(&f.remove_prereq(y, z));
    // y no longer needs z, so z may need x.
    assert_eq!(f.needs(z, x), Ok(true));
}

#[test]
fn prereq_ac2_remove_retires_the_link_and_writes_one_event() {
    let f = Fixture::new();
    let a = f.card("src/a.rs");
    let b = f.card("src/b.rs");
    assert_eq!(f.needs(a, b), Ok(true));
    let rows = f.link_rows();

    let removed = result(&f.remove_prereq(a, b));
    assert_eq!(removed["removed"], json!(true), "{removed}");
    assert_eq!(removed["writes"], json!(true), "{removed}");
    // The row stays; only its state changes.
    assert_eq!(f.link_rows(), rows);
    assert_eq!(
        f.count(&format!(
            "SELECT count(*) FROM card_link WHERE kind = 'blocks' AND state = 'retired' \
             AND from_card = {b} AND to_card = {a}"
        )),
        1
    );
    let payload: String = {
        let l = ledger::open_read_only(f.home.path(), "demo").unwrap();
        l.conn()
            .query_row(
                "SELECT payload FROM event WHERE card_id = ?1 AND kind = 'card.prereq_remove'",
                [a],
                |r| r.get(0),
            )
            .unwrap()
    };
    let p: Value = serde_json::from_str(&payload).unwrap();
    assert_eq!(p["needs"], json!(b), "{p}");
    assert_eq!(p["by"], json!("leader"), "{p}");

    // A second call finds no active link and writes nothing.
    let events = f.count("SELECT count(*) FROM event");
    let again = result(&f.remove_prereq(a, b));
    assert_eq!(again["removed"], json!(false), "{again}");
    assert_eq!(again["writes"], json!(false), "{again}");
    assert_eq!(f.count("SELECT count(*) FROM event"), events);

    // An unknown card and a missing `--by` are refused and write nothing.
    let message = refusal(&f.remove_prereq(a, 999));
    assert!(message.starts_with("card prereq refused: "), "{message}");
    let no_by = f.cli(&[
        "card",
        "prereq",
        &format!("C-{a}"),
        "--needs",
        &format!("C-{b}"),
        "--remove",
    ]);
    assert!(!no_by.status.success());
    assert_eq!(f.count("SELECT count(*) FROM event"), events);
    assert_eq!(f.link_rows(), rows);
}

#[test]
fn prereq_ac2_a_retired_link_is_added_again_as_the_same_row_after_the_loop_check() {
    let f = Fixture::new();
    let a = f.card("src/a.rs");
    let b = f.card("src/b.rs");
    assert_eq!(f.needs(a, b), Ok(true));
    result(&f.remove_prereq(a, b));
    let rows = f.link_rows();

    // While the first link is retired the opposite one may be added ...
    assert_eq!(f.needs(b, a), Ok(true));
    // ... and then the first one would close a loop, so adding it again is refused.
    assert!(matches!(
        f.needs(a, b),
        Err(workflow::PrereqError::Cycle { .. })
    ));
    result(&f.remove_prereq(b, a));

    // With nothing in the way the same row becomes active again and no row is added.
    assert_eq!(f.needs(a, b), Ok(true));
    assert_eq!(f.link_rows(), rows + 1, "only the opposite pair is new");
    assert_eq!(
        f.count(&format!(
            "SELECT count(*) FROM card_link WHERE kind = 'blocks' AND state = 'active' \
             AND from_card = {b} AND to_card = {a}"
        )),
        1
    );
}

#[test]
fn prereq_ac3_removing_the_last_prerequisite_offers_the_card_and_take_goes_through() {
    let f = Fixture::new();
    let waits = f.card("src/a.rs");
    let first = f.card("src/b.rs");
    f.needs(waits, first).unwrap();
    assert_eq!(skipped(&f.next(), waits)["reason"], json!("prerequisite"));
    let message = refusal(&f.take(waits));
    assert!(message.contains(&format!("C-{first}")), "{message}");

    result(&f.remove_prereq(waits, first));
    let next = f.next();
    assert!(
        names(&next["ready"]).contains(&format!("C-{waits}")),
        "{next}"
    );
    assert_eq!(f.show(waits)["prerequisites"], json!([]));
    result(&f.take(waits));
}

#[test]
fn prereq_ac5_the_hold_read_goes_through_the_event_card_kind_index() {
    let f = Fixture::new();
    let id = f.card("src/a.rs");
    f.hold(id, "money", "budget").unwrap();
    let l = ledger::open_read_only(f.home.path(), "demo").unwrap();
    let plan: Vec<String> = l
        .conn()
        .prepare(&format!("EXPLAIN QUERY PLAN {}", workflow::HOLD_READ_SQL))
        .unwrap()
        .query_map([id], |r| r.get::<_, String>(3))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    let plan = plan.join(" | ");
    assert!(plan.contains("event_card_kind"), "{plan}");
    assert!(
        !plan.contains("SCAN event"),
        "the hold read must not scan the whole event table: {plan}"
    );
}

#[test]
fn prereq_ac3_only_a_withdrawn_prerequisite_stops_blocking() {
    let f = Fixture::new();
    let waits = f.card("src/w.rs");
    let withdrawn = f.card("src/p1.rs");
    let folded = f.card("src/p2.rs");
    let plain_hold = f.card("src/p3.rs");
    let active = f.card("src/p4.rs");
    let review = f.card("src/p5.rs");
    let done = f.card("src/p6.rs");
    let target = f.card("src/t.rs");
    for needs in [withdrawn, folded, plain_hold, active, review, done] {
        f.needs(waits, needs).unwrap();
    }
    f.withdraw(withdrawn);
    f.fold(folded, target);
    result(&f.take(active));
    f.set_states(&[(plain_hold, "hold"), (review, "review"), (done, "done")]);

    // The withdrawn and the done card do not block; the other four still do.
    assert_eq!(
        f.unfinished(waits),
        vec![folded, plain_hold, active, review],
        "the withdrawn card must not block, the folded, plain hold, active and review ones must"
    );
    let next = f.next();
    assert_eq!(
        skipped(&next, waits)["needs"],
        json!([
            format!("C-{folded}"),
            format!("C-{plain_hold}"),
            format!("C-{active}"),
            format!("C-{review}")
        ]),
        "{next}"
    );

    // A withdrawn card that is made ready again is work again, so it blocks again.
    result(&f.cli(&["card", "ready", &format!("C-{withdrawn}")]));
    assert_eq!(
        f.unfinished(waits),
        vec![withdrawn, folded, plain_hold, active, review]
    );
}
