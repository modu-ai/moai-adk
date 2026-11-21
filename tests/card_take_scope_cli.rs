//! SPEC-20261121-take-scope-stored: `moai card take` keeps the scope it was given on the card,
//! fills an empty scope from the SPEC the card names or from the card's last reservation, names
//! the two ways out when it has no source, and `moai card next` lists what a take could fill.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

struct World {
    home: TempHome,
    root: PathBuf,
    project: String,
}

fn world() -> World {
    let home = TempHome::new();
    let root = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    World {
        project: root.to_str().unwrap().to_string(),
        home,
        root,
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

/// The message of a refused command (exit 1, `ok: false`).
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

impl World {
    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        self.run(session, &all)
    }

    /// `moai card next <project>` (the project is a positional argument there).
    fn next(&self, session: Option<&str>) -> Value {
        result(&self.run(session, &["card", "next", &self.project]))
    }

    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    /// Runs `statement` against the ledger directly.
    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    /// A `ready` card without a write scope (`--scope` is left out of `card add`).
    fn empty_card(&self, title: &str) -> String {
        let added = result(&self.card(None, &["add", "--title", title]));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    /// A SPEC in the project with the `write:` paths given, light form; returns its id.
    fn spec(&self, slug: &str, paths: &[String]) -> String {
        let id = format!("SPEC-20261121-{slug}");
        let list = paths
            .iter()
            .map(|p| format!("`{p}`"))
            .collect::<Vec<_>>()
            .join(", ");
        let body = format!(
            "---\nid: {id}\ntitle: \"a demo\"\nstatus: draft\nrevision: 1\ncreated: 2026-11-21\n---\n\n\
             # a demo\n\n## Intent\n\nA demo.\n\n## Requirements\n\n- REQ-1: The tool shall do a thing.\n\n\
             ## Acceptance\n\n- AC-1 (REQ-1): it does the thing.\n  verify: `true`\n\n\
             ## Scope\n\n- write: {list}\n- read-only: every other file\n\n## Plan\n\n1. Do it.\n"
        );
        let dir = self.root.join(".moai/specs").join(&id);
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("spec.md"), body).unwrap();
        id
    }

    fn link(&self, card: &str, spec: &str) {
        result(&self.card(None, &["spec", card, spec]));
    }

    fn show(&self, card: &str) -> Value {
        result(&self.card(None, &["show", card]))
    }

    /// The card's own scope as the `card show` answer reads it.
    fn own_scope(&self, card: &str) -> Value {
        self.show(card)["write_scope"].clone()
    }
}

fn paths(n: usize) -> Vec<String> {
    (0..n).map(|i| format!("src/part_{i:03}.rs")).collect()
}

/// AC-1 (REQ-1, REQ-5): the scope given to a take is on the card after the release, and the card
/// is on the list `card next` makes.
#[test]
fn a_given_scope_is_kept_on_the_card() {
    let w = world();
    let name = w.empty_card("given");
    let took = result(&w.card(
        Some(S_A),
        &["take", &name, "--scope", "src/a.rs", "--scope", "docs/"],
    ));
    assert_eq!(took["scope_source"], json!("given"), "{took}");
    assert_eq!(
        took["reservation"]["write_scope"],
        json!(["docs/", "src/a.rs"]),
        "{took}"
    );
    result(&w.card(Some(S_A), &["release", &name, "--generation", "1"]));
    assert_eq!(w.own_scope(&name), json!(["docs/", "src/a.rs"]));
    let next = w.next(Some(S_B));
    let ready: Vec<&str> = next["ready"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap())
        .collect();
    assert_eq!(ready, [name.as_str()], "{next}");
}

/// AC-2 (REQ-2): a card that has its own scope keeps it; only the reservation takes the given one.
#[test]
fn an_own_scope_is_never_changed_by_take() {
    let w = world();
    let added = result(&w.card(None, &["add", "--title", "own", "--scope", "src/a.rs"]));
    let name = added["name"].as_str().unwrap().to_string();
    result(&w.card(None, &["ready", &name]));
    let took = result(&w.card(Some(S_A), &["take", &name, "--scope", "src/b.rs"]));
    assert_eq!(took["reservation"]["write_scope"], json!(["src/b.rs"]));
    assert_eq!(took["scope_source"], json!("given"), "{took}");
    assert_eq!(w.own_scope(&name), json!(["src/a.rs"]));
}

/// AC-3 (REQ-3, REQ-5): an empty card that names a SPEC with a `write:` list is filled from it.
#[test]
fn an_empty_card_is_filled_from_its_spec() {
    let w = world();
    let name = w.empty_card("spec");
    let spec = w.spec("fill-demo", &["src/x.rs".into(), "tests/x.rs".into()]);
    w.link(&name, &spec);
    let took = result(&w.card(Some(S_A), &["take", &name]));
    assert_eq!(took["scope_source"], json!("spec"), "{took}");
    assert_eq!(
        took["reservation"]["write_scope"],
        json!(["src/x.rs", "tests/x.rs"]),
        "{took}"
    );
    assert_eq!(w.own_scope(&name), json!(["src/x.rs", "tests/x.rs"]));
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.scope'"),
        1
    );
}

/// AC-4 (REQ-4, REQ-5): an empty card with no SPEC list gets the scope of its last reservation.
#[test]
fn an_empty_card_is_filled_from_its_last_reservation() {
    let w = world();
    let name = w.empty_card("again");
    result(&w.card(Some(S_A), &["take", &name, "--scope", "src/old.rs"]));
    result(&w.card(Some(S_A), &["release", &name, "--generation", "1"]));
    // A card released before this change has no scope of its own: the shape of the ready cards
    // that `card next` skips as `empty_scope`.
    w.sql("UPDATE card SET write_scope = '[]' WHERE card_id = 1");
    assert_eq!(w.own_scope(&name), json!([]));
    let took = result(&w.card(Some(S_B), &["take", &name]));
    assert_eq!(took["scope_source"], json!("last_reservation"), "{took}");
    assert_eq!(took["reservation"]["write_scope"], json!(["src/old.rs"]));
    assert_eq!(w.own_scope(&name), json!(["src/old.rs"]));
}

/// AC-5 (REQ-6): with no source the refusal names both ways out and nothing is written.
#[test]
fn no_source_names_the_two_ways_out() {
    let w = world();
    let name = w.empty_card("nothing");
    let events = w.count("SELECT count(*) FROM event");
    let message = refusal(&w.card(Some(S_A), &["take", &name]));
    assert!(message.contains("--scope"), "{message}");
    assert!(message.contains("card scope"), "{message}");
    assert_eq!(w.count("SELECT count(*) FROM event"), events);
    assert_eq!(w.count("SELECT count(*) FROM reservation"), 0);
    assert_eq!(w.own_scope(&name), json!([]));
}

/// AC-6 (REQ-7): the fill meets the overlap check in the same transaction; a refusal leaves the
/// card's scope empty and writes no `card.scope` event.
#[test]
fn a_fill_keeps_the_overlap_check() {
    let w = world();
    let holder = w.empty_card("holder");
    result(&w.card(Some(S_A), &["take", &holder, "--scope", "src/x.rs"]));
    let name = w.empty_card("filled");
    let spec = w.spec("fill-overlap", &["src/x.rs".into()]);
    w.link(&name, &spec);
    let events = w.count("SELECT count(*) FROM event WHERE kind = 'card.scope'");
    let message = refusal(&w.card(Some(S_B), &["take", &name]));
    // The reservation's own refusal names the write scope in conflict and the card that holds it.
    assert!(message.contains("conflict in write scope"), "{message}");
    assert!(message.contains(&holder), "{message}");
    assert_eq!(w.own_scope(&name), json!([]));
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.scope'"),
        events
    );
}

/// AC-7 (REQ-8): a SPEC list of 64 paths fills the scope, a list of 65 is refused with the limit.
#[test]
fn the_spec_list_has_the_same_limit() {
    let w = world();
    let fits = w.empty_card("sixty four");
    let spec = w.spec("fill-sixty-four", &paths(64));
    w.link(&fits, &spec);
    let took = result(&w.card(Some(S_A), &["take", &fits]));
    assert_eq!(took["scope_source"], json!("spec"), "{took}");
    assert_eq!(
        took["reservation"]["write_scope"].as_array().unwrap().len(),
        64
    );

    let too_many = w.empty_card("sixty five");
    let spec = w.spec("fill-sixty-five", &paths(65));
    w.link(&too_many, &spec);
    let reservations = w.count("SELECT count(*) FROM reservation");
    let message = refusal(&w.card(Some(S_A), &["take", &too_many]));
    assert!(message.contains("64"), "{message}");
    assert_eq!(w.count("SELECT count(*) FROM reservation"), reservations);
    assert_eq!(w.own_scope(&too_many), json!([]));
}

/// AC-8 (REQ-9): `card next` lists the cards a take could fill, with the scope and its source,
/// and still skips a card with no source as `empty_scope`.
#[test]
fn next_lists_what_take_can_fill() {
    let w = world();
    let from_spec = w.empty_card("from spec");
    let spec = w.spec("fill-next", &["src/n.rs".into()]);
    w.link(&from_spec, &spec);
    let from_reservation = w.empty_card("from reservation");
    result(&w.card(
        Some(S_A),
        &["take", &from_reservation, "--scope", "src/r.rs"],
    ));
    result(&w.card(
        Some(S_A),
        &["release", &from_reservation, "--generation", "1"],
    ));
    w.sql("UPDATE card SET write_scope = '[]' WHERE title = 'from reservation'");
    let nothing = w.empty_card("nothing");

    let next = w.next(Some(S_B));
    let ready = next["ready"].as_array().unwrap();
    let find = |name: &str| ready.iter().find(|c| c["name"] == json!(name)).cloned();
    let a = find(&from_spec).unwrap_or_else(|| panic!("no {from_spec} in {next}"));
    assert_eq!(a["write_scope"], json!(["src/n.rs"]), "{a}");
    assert_eq!(a["scope_source"], json!("spec"), "{a}");
    let b = find(&from_reservation).unwrap_or_else(|| panic!("no {from_reservation} in {next}"));
    assert_eq!(b["write_scope"], json!(["src/r.rs"]), "{b}");
    assert_eq!(b["scope_source"], json!("last_reservation"), "{b}");
    let skipped = next["skipped"].as_array().unwrap();
    let empty = skipped
        .iter()
        .find(|c| c["name"] == json!(nothing))
        .unwrap_or_else(|| panic!("no {nothing} in {next}"));
    assert_eq!(empty["reason"], json!("empty_scope"), "{empty}");
    // `next` only reads: the cards still have no scope of their own.
    assert_eq!(w.own_scope(&from_spec), json!([]));
}
