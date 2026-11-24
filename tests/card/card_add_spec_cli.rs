//! SPEC-20261124-followup-per-spec: `moai card add --spec <SPEC-ID>` records the SPEC in the same
//! write as the card and refuses a second open card on one SPEC; `--follow-up` allows one open
//! follow-up card per SPEC and `--split-of <card>` names the card it was split from. The relation
//! lives in the `card.spec` event payload, never in a new column. Every command runs against a
//! temporary HOME and a temporary project; the real home state folder is not read or written.

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Output;

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
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

/// The refusal: exit 1, its `kind` and its message.
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

/// A refusal from the argument parser: a non-zero exit and a message, nothing on stdout.
fn parse_refusal(out: &Output) -> String {
    assert_ne!(
        out.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// A HOME with a project `demo`, its registered ledger and the project folder.
struct World {
    home: TempHome,
    project: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        World { home, project }
    }

    fn p(&self) -> &str {
        self.project.to_str().unwrap()
    }

    /// `moai card <args> --project P`.
    fn card(&self, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", self.p()]);
        self.home.command(&all).output().expect("run moai")
    }

    /// A draft SPEC `SPEC-20261124-<slug>` with every template placeholder filled, so `spec check`
    /// passes it.
    fn spec(&self, slug: &str) -> String {
        let (id, path) = self.new_spec(slug);
        fs::write(&path, filled(&fs::read_to_string(&path).unwrap())).unwrap();
        id
    }

    /// A SPEC with an error finding: the filled template with every `shall` turned into `will`,
    /// so its requirements are no requirement lines. Returns the id.
    fn broken_spec(&self, slug: &str) -> String {
        let (id, path) = self.new_spec(slug);
        let text = filled(&fs::read_to_string(&path).unwrap()).replace("shall", "will");
        fs::write(&path, text).unwrap();
        let checked = self
            .home
            .command(&["spec", "check", &id, "--project", self.p()])
            .output()
            .expect("run moai");
        assert_ne!(
            checked.status.code(),
            Some(0),
            "the fixture SPEC must fail `spec check`"
        );
        id
    }

    /// `moai spec new`: the SPEC's id and the file it wrote (the path is the one the answer gives).
    fn new_spec(&self, slug: &str) -> (String, PathBuf) {
        let out = self
            .home
            .command(&[
                "spec",
                "new",
                slug,
                "--title",
                "t",
                "--minimal",
                "--date",
                "2026-11-24",
                "--project",
                self.p(),
            ])
            .output()
            .expect("run moai");
        let answer = ok(&out);
        let id = answer["id"].as_str().unwrap().to_string();
        let written = PathBuf::from(answer["path"].as_str().unwrap());
        let path = if written.is_absolute() {
            written
        } else {
            self.project.join(written)
        };
        (id, path)
    }

    /// `card add --spec <spec> <extra…>` with one criterion.
    fn add(&self, spec: &str, extra: &[&str]) -> Output {
        let mut args = vec!["add", "--title", "t", "--ac", "fixture", "--spec", spec];
        args.extend_from_slice(extra);
        self.card(&args)
    }

    fn cards(&self) -> i64 {
        self.count("SELECT count(*) FROM card")
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    /// The payload of the newest `card.spec` event of `card_id`.
    fn link_payload(&self, card_id: i64) -> Value {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let text: String = l
            .conn()
            .query_row(
                "SELECT payload FROM event WHERE kind = 'card.spec' AND card_id = ?1 \
                 ORDER BY event_id DESC LIMIT 1",
                [card_id],
                |r| r.get(0),
            )
            .unwrap();
        serde_json::from_str(&text).unwrap()
    }

    /// Closes card `n` the way a landing leaves it: a met criterion, a `card.landed` event, then
    /// `review` and `done`.
    fn close(&self, n: i64) {
        let raw =
            rusqlite::Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap())
                .unwrap();
        raw.execute_batch(&format!(
            "UPDATE card_ac SET status = 'met', evidence_ref = 'test' WHERE card_id = {n}; \
             INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES ('card.landed:{n}:test', unixepoch(), 'test', {n}, 'card.landed', '{{}}'); \
             UPDATE card SET state = 'ready' WHERE card_id = {n}; \
             UPDATE card SET state = 'active' WHERE card_id = {n}; \
             UPDATE card SET state = 'review' WHERE card_id = {n}; \
             UPDATE card SET state = 'done', closed_at = unixepoch() WHERE card_id = {n};"
        ))
        .unwrap();
    }
}

/// A template's text with every `<…>` placeholder filled with the word `filled`.
fn filled(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(open) = rest.find('<') {
        let Some(close) = rest[open..].find('>') else {
            break;
        };
        out.push_str(&rest[..open]);
        out.push_str("filled");
        rest = &rest[open + close + 1..];
    }
    out.push_str(rest);
    out
}

/// REQ-1: the card, its criterion, the `card.add` event and the `card.spec` event are one write.
#[test]
fn add_with_spec_links_in_one_write() {
    let w = World::new();
    let spec = w.spec("login-fix");
    let added = ok(&w.add(&spec, &[]));
    assert_eq!(added["name"], json!("C-1"), "{added}");
    assert_eq!(added["spec"], json!(spec), "{added}");
    assert_eq!(added["writes"], json!(true));
    assert_eq!(ok(&w.card(&["show", "C-1"]))["spec"], json!(spec));
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.add'"),
        1
    );
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.spec'"),
        1
    );
    assert_eq!(w.link_payload(1)["to"], json!(spec));
    assert_eq!(w.link_payload(1)["from"], Value::Null);
}

/// REQ-2: a SPEC that is missing, not an id or broken refuses the add and writes nothing.
#[test]
fn a_missing_or_broken_spec_is_refused() {
    let w = World::new();
    let broken = w.broken_spec("broken");
    let broken = broken.as_str();
    // A SPEC that is not there is `not_found`; a bad id or a SPEC with an error is `invalid_input`.
    for (bad, expected) in [
        ("SPEC-20261124-absent", "not_found"),
        ("not-a-spec-id", "invalid_input"),
        (broken, "invalid_input"),
    ] {
        let (kind, message) = refusal(&w.add(bad, &[]));
        assert_eq!(kind, expected, "{bad}: {message}");
        assert_eq!(w.cards(), 0, "{bad}: a refused add writes no card");
        assert_eq!(w.count("SELECT count(*) FROM event"), 0, "{bad}");
    }
}

/// REQ-3: a second open card on one SPEC is refused and the way out is named.
#[test]
fn a_second_open_card_on_one_spec_is_refused() {
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &[]));
    let events = w.count("SELECT count(*) FROM event");
    let (kind, message) = refusal(&w.add(&spec, &[]));
    assert_eq!(kind, "conflict", "{message}");
    for word in ["C-1", "--follow-up", "--split-of", "card criterion-add"] {
        assert!(
            message.contains(word),
            "the refusal lacks {word}: {message}"
        );
    }
    assert_eq!(w.cards(), 1, "nothing was written");
    assert_eq!(w.count("SELECT count(*) FROM event"), events);
    // Another SPEC is not blocked by the first one.
    let other = w.spec("other-fix");
    ok(&w.add(&other, &[]));
    assert_eq!(w.cards(), 2);
}

/// REQ-3 and D-4: only a card that is not `done` counts.
#[test]
fn a_done_card_does_not_count() {
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &[]));
    w.close(1);
    let second = ok(&w.add(&spec, &[]));
    assert_eq!(second["name"], json!("C-2"), "{second}");
}

/// SPEC-20261124-card-help-gates AC-7 (REQ-5): a withdrawn card and a folded card are `hold`, not
/// open, so they leave room on their SPEC; a card that is still open still blocks it (AUDIT #444
/// O6).
#[test]
fn a_hold_card_is_not_an_open_card_of_its_spec() {
    let w = World::new();
    let spec = w.spec("login-fix");
    let other = w.spec("other-fix");
    ok(&w.add(&spec, &[])); // C-1: withdrawn below
                            // `card withdraw` takes the project folder as its last positional argument.
    let withdrawn = w
        .home
        .command(&[
            "card",
            "withdraw",
            "C-1",
            "--reason",
            "no longer needed",
            "--by",
            "leader",
            w.p(),
        ])
        .output()
        .expect("run moai");
    ok(&withdrawn);
    assert_eq!(w.count("SELECT count(*) FROM card WHERE state = 'hold'"), 1);

    let second = ok(&w.add(&spec, &[])); // C-2: folded below
    assert_eq!(
        second["name"],
        json!("C-2"),
        "withdrawn card made room: {second}"
    );
    ok(&w.add(&other, &[])); // C-3: the fold target
    ledger::close::fold(
        &mut ledger::open_write(w.home.path(), "demo").unwrap(),
        2,
        3,
        "same work as the target",
        "lane",
        None,
    )
    .unwrap();
    assert_eq!(w.count("SELECT count(*) FROM card WHERE state = 'hold'"), 2);

    let third = ok(&w.add(&spec, &[])); // C-4: only hold cards were on the SPEC
    assert_eq!(
        third["name"],
        json!("C-4"),
        "folded card made room: {third}"
    );
    // C-4 is open now: a further plain card is refused again and C-4 is the one named.
    let (kind, message) = refusal(&w.add(&spec, &[]));
    assert_eq!(kind, "conflict", "{message}");
    // The SPEC id holds `C-2` inside `SPEC-2026…`, so the open card is read as the sentence
    // names it: `has an open card (C-4)`.
    assert!(message.contains("has an open card (C-4)"), "{message}");
}

/// REQ-4, REQ-5: one open follow-up card per SPEC, whatever other cards the SPEC has.
#[test]
fn one_open_follow_up_per_spec() {
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &[]));
    let follow = ok(&w.add(&spec, &["--follow-up"]));
    assert_eq!(follow["name"], json!("C-2"), "{follow}");
    assert_eq!(follow["relation"], json!("follow-up"), "{follow}");
    assert_eq!(w.link_payload(2)["relation"], json!("follow-up"));
    assert_eq!(w.link_payload(2)["to"], json!(spec));
    let (kind, message) = refusal(&w.add(&spec, &["--follow-up"]));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("C-2"), "{message}");
    assert!(message.contains("card criterion-add C-2 --ac"), "{message}");
    assert_eq!(w.cards(), 2, "nothing was written");
    // A plain add is refused while the follow-up card is the only open card too.
    w.close(1);
    let (kind, message) = refusal(&w.add(&spec, &[]));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("C-2"), "{message}");
}

/// REQ-5: a follow-up card that is `done` makes room for the next one.
#[test]
fn a_closed_follow_up_makes_room() {
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &["--follow-up"]));
    w.close(1);
    let next = ok(&w.add(&spec, &["--follow-up"]));
    assert_eq!(next["name"], json!("C-2"), "{next}");
}

/// REQ-6: `--split-of` is accepted for a card that names the SPEC and records where it came from.
#[test]
fn split_of_names_its_origin() {
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &[]));
    let split = ok(&w.add(&spec, &["--split-of", "C-1"]));
    assert_eq!(split["name"], json!("C-2"), "{split}");
    assert_eq!(split["relation"], json!("split-of"), "{split}");
    let payload = w.link_payload(2);
    assert_eq!(payload["relation"], json!("split-of"), "{payload}");
    assert_eq!(payload["of"], json!("C-1"), "{payload}");
    assert_eq!(payload["to"], json!(spec), "{payload}");
}

/// REQ-2 and REQ-6: `--split-of` names a card that exists and names the same SPEC.
#[test]
fn split_of_needs_the_same_spec() {
    let w = World::new();
    let spec = w.spec("login-fix");
    let other = w.spec("other-fix");
    ok(&w.add(&other, &[]));
    ok(&w.card(&["add", "--title", "no spec", "--ac", "fixture"]));
    ok(&w.add(&spec, &[]));
    let cards = w.cards();
    for origin in ["C-99", "C-1", "C-2"] {
        let (kind, message) = refusal(&w.add(&spec, &["--split-of", origin]));
        assert!(
            kind == "not_found" || kind == "invalid_input",
            "{origin}: {kind} {message}"
        );
        assert!(message.contains(origin), "{origin}: {message}");
        assert_eq!(w.cards(), cards, "{origin}: nothing was written");
    }
}

/// REQ-7: the relation flags need `--spec` and exclude each other.
#[test]
fn relation_flags_need_a_spec_and_exclude_each_other() {
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &[]));
    let before = w.count("SELECT count(*) FROM event");
    let bare = ["add", "--title", "t", "--ac", "fixture"];
    let with = |extra: &[&str]| {
        let mut args = bare.to_vec();
        args.extend_from_slice(extra);
        w.card(&args)
    };
    for extra in [
        vec!["--follow-up"],
        vec!["--split-of", "C-1"],
        vec!["--spec", spec.as_str(), "--follow-up", "--split-of", "C-1"],
    ] {
        let message = parse_refusal(&with(&extra));
        assert!(
            message.contains("--spec") || message.contains("cannot be used"),
            "{extra:?}: {message}"
        );
    }
    assert_eq!(w.cards(), 1, "nothing was written");
    assert_eq!(w.count("SELECT count(*) FROM event"), before);
}

/// REQ-8: a call with none of the new flags is the add it was.
#[test]
fn an_add_without_spec_is_unchanged() {
    let w = World::new();
    let first = ok(&w.card(&["add", "--title", "t", "--ac", "fixture"]));
    let second = ok(&w.card(&["add", "--title", "t", "--ac", "fixture"]));
    assert_eq!(first["name"], json!("C-1"));
    assert_eq!(second["name"], json!("C-2"));
    assert_eq!(first["events"], json!(1), "{first}");
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.spec'"),
        0
    );
}
