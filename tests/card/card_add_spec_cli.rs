//! SPEC-20261124-followup-per-spec: `moai card add --spec <SPEC-ID>` records the SPEC in the same
//! write as the card and refuses a second open card on one SPEC; `--follow-up` adds a follow-up
//! card and is never refused for open cards, so a SPEC may have several open follow-up cards;
//! `--follow-up-of <card>` does the same and names the card it follows, and `--split-of <card>`
//! names the card it was split from (each origin must name the same SPEC). The relation lives in
//! the `card.spec` event payload, never in a new column, and `card show` prints it as
//! `spec_link`. Every command runs against a temporary HOME and a temporary project; the real
//! home state folder is not read or written.

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

/// REQ-4, REQ-5: a SPEC may have several open follow-up cards at once, whatever other cards it
/// has; each carries `relation: follow-up`, and every one is listed by the SPEC's open cards.
#[test]
fn several_open_follow_ups_per_spec() {
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &[]));
    for n in 2..=4 {
        let follow = ok(&w.add(&spec, &["--follow-up"]));
        assert_eq!(follow["name"], json!(format!("C-{n}")), "{follow}");
        assert_eq!(follow["relation"], json!("follow-up"), "{follow}");
        assert_eq!(follow.get("of"), None, "{follow}");
        assert_eq!(w.link_payload(n)["relation"], json!("follow-up"));
        assert_eq!(w.link_payload(n)["to"], json!(spec));
        assert_eq!(w.link_payload(n).get("of"), None);
    }
    assert_eq!(w.cards(), 4);
    // The three follow-ups are open together: `card show` gives each its link, and the ledger
    // counts all of them as cards of the SPEC.
    for n in 2..=4 {
        let shown = ok(&w.card(&["show", &format!("C-{n}")]));
        assert_eq!(shown["spec"], json!(spec), "{shown}");
        assert_eq!(
            shown["spec_link"],
            json!({"relation": "follow-up", "of": null}),
            "{shown}"
        );
        assert_eq!(shown["state"], json!("draft"), "{shown}");
    }
    let linked = format!(
        "SELECT count(*) FROM card WHERE state NOT IN ('done', 'hold') AND {} = '{spec}'",
        moai_cli::card::spec_link::LINKED_COLUMN
    );
    assert_eq!(w.count(&linked), 4);
    // A plain add is still refused, and its refusal names every open card, the follow-ups too.
    let (kind, message) = refusal(&w.add(&spec, &[]));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("C-1, C-2, C-3, C-4"), "{message}");
    assert_eq!(w.cards(), 4, "nothing was written");
}

/// REQ-4: a plain add on a SPEC whose only open cards are follow-ups is refused and names all
/// of them and the ways out.
#[test]
fn a_plain_add_names_every_open_follow_up() {
    let w = World::new();
    let spec = w.spec("login-fix");
    for _ in 0..3 {
        ok(&w.add(&spec, &["--follow-up"]));
    }
    let (kind, message) = refusal(&w.add(&spec, &[]));
    assert_eq!(kind, "conflict", "{message}");
    for word in [
        "C-1, C-2, C-3",
        "--follow-up",
        "--follow-up-of <card>",
        "--split-of",
        "card criterion-add C-1 --ac",
    ] {
        assert!(
            message.contains(word),
            "the refusal lacks {word}: {message}"
        );
    }
    assert!(!message.contains("one follow-up"), "{message}");
    assert_eq!(w.cards(), 3, "nothing was written");
    // Closing all three makes room again.
    for n in 1..=3 {
        w.close(n);
    }
    ok(&w.add(&spec, &[]));
}

/// REQ-5: a follow-up card that is `done` leaves the SPEC's open cards; the next follow-up and
/// the one still open stay side by side.
#[test]
fn a_closed_follow_up_leaves_the_open_cards() {
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &["--follow-up"]));
    ok(&w.add(&spec, &["--follow-up"]));
    w.close(1);
    let next = ok(&w.add(&spec, &["--follow-up"]));
    assert_eq!(next["name"], json!("C-3"), "{next}");
    // C-2 and C-3 are open; a plain add names both and not the closed C-1.
    let (kind, message) = refusal(&w.add(&spec, &[]));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("has an open card (C-2, C-3)"), "{message}");
}

/// A card that names no SPEC link shows `spec_link: null`; a link made by `card spec` shows both
/// fields null, and an unlink shows `null` again.
#[test]
fn card_show_spec_link_follows_the_newest_link() {
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.card(&["add", "--title", "t", "--ac", "fixture"]));
    let shown = ok(&w.card(&["show", "C-1"]));
    assert_eq!(shown["spec"], Value::Null, "{shown}");
    assert_eq!(shown["spec_link"], Value::Null, "{shown}");
    ok(&w.card(&["spec", "C-1", &spec]));
    let shown = ok(&w.card(&["show", "C-1"]));
    assert_eq!(
        shown["spec_link"],
        json!({"relation": null, "of": null}),
        "{shown}"
    );
    ok(&w.card(&["spec", "C-1", "--unlink"]));
    let shown = ok(&w.card(&["show", "C-1"]));
    assert_eq!(shown["spec_link"], Value::Null, "{shown}");
    // A plain add with `--spec` has a link with no relation.
    ok(&w.add(&spec, &[]));
    let shown = ok(&w.card(&["show", "C-2"]));
    assert_eq!(
        shown["spec_link"],
        json!({"relation": null, "of": null}),
        "{shown}"
    );
}

/// `--follow-up-of <card>` is a follow-up card that records the card it follows, and it is never
/// refused for open cards. Needs the `--follow-up-of` flag of `src/ops/card.rs`.
#[test]
fn follow_up_of_names_its_origin() {
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &[]));
    let follow = ok(&w.add(&spec, &["--follow-up-of", "C-1"]));
    assert_eq!(follow["name"], json!("C-2"), "{follow}");
    assert_eq!(follow["relation"], json!("follow-up"), "{follow}");
    assert_eq!(follow["of"], json!("C-1"), "{follow}");
    let payload = w.link_payload(2);
    assert_eq!(payload["relation"], json!("follow-up"), "{payload}");
    assert_eq!(payload["of"], json!("C-1"), "{payload}");
    assert_eq!(payload["to"], json!(spec), "{payload}");
    let shown = ok(&w.card(&["show", "C-2"]));
    assert_eq!(
        shown["spec_link"],
        json!({"relation": "follow-up", "of": "C-1"}),
        "{shown}"
    );
    // A follow-up of the follow-up and a second follow-up of C-1 are both accepted, next to the
    // two cards that are already open.
    let second = ok(&w.add(&spec, &["--follow-up-of", "C-2"]));
    assert_eq!(second["of"], json!("C-2"), "{second}");
    let third = ok(&w.add(&spec, &["--follow-up-of", "C-1"]));
    assert_eq!(third["name"], json!("C-4"), "{third}");
    assert_eq!(w.cards(), 4);
    // The origin may be a closed card.
    w.close(1);
    ok(&w.add(&spec, &["--follow-up-of", "C-1"]));
    assert_eq!(w.cards(), 5);
}

/// `--follow-up-of` names a card that exists and names the same SPEC; a refused add writes
/// nothing. Needs the `--follow-up-of` flag of `src/ops/card.rs`.
#[test]
fn follow_up_of_needs_the_same_spec() {
    let w = World::new();
    let spec = w.spec("login-fix");
    let other = w.spec("other-fix");
    ok(&w.add(&other, &[])); // C-1: another SPEC
    ok(&w.card(&["add", "--title", "no spec", "--ac", "fixture"])); // C-2: no SPEC
    ok(&w.add(&spec, &["--follow-up"])); // C-3: the SPEC's own card
    let cards = w.cards();
    let events = w.count("SELECT count(*) FROM event");
    // C-99 is missing, C-1 names another SPEC, C-2 none; `not-a-card` is no card name.
    for origin in ["C-99", "C-1", "C-2", "not-a-card"] {
        let (kind, message) = refusal(&w.add(&spec, &["--follow-up-of", origin]));
        assert!(
            kind == "not_found" || kind == "invalid_input",
            "{origin}: {kind} {message}"
        );
        assert!(message.contains(origin), "{origin}: {message}");
        assert!(message.contains("--follow-up-of"), "{origin}: {message}");
        assert_eq!(w.cards(), cards, "{origin}: nothing was written");
        assert_eq!(
            w.count("SELECT count(*) FROM event"),
            events,
            "{origin}: no event was written"
        );
    }
    // The card of the SPEC itself is accepted.
    ok(&w.add(&spec, &["--follow-up-of", "C-3"]));
    assert_eq!(w.cards(), cards + 1);
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
    // The `--follow-up-of` rows pass once `src/ops/card.rs` has the flag (`requires = "spec"`,
    // `conflicts_with` the other two); until then clap refuses the unknown flag, which is also a
    // refusal, so the rows pin the three flags' exclusion once the flag exists.
    for extra in [
        vec!["--follow-up"],
        vec!["--split-of", "C-1"],
        vec!["--follow-up-of", "C-1"],
        vec!["--spec", spec.as_str(), "--follow-up", "--split-of", "C-1"],
        vec![
            "--spec",
            spec.as_str(),
            "--follow-up-of",
            "C-1",
            "--follow-up",
        ],
        vec![
            "--spec",
            spec.as_str(),
            "--follow-up-of",
            "C-1",
            "--split-of",
            "C-1",
        ],
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

// ---- SPEC-20261124-card-help-gates (revision 2): one open card per SPEC on every path --------

impl World {
    /// `card ready C-n` by the leader.
    fn ready(&self, n: i64) -> Output {
        self.card(&["ready", &format!("C-{n}"), "--by", "leader"])
    }

    /// `card spec C-n <spec>` (or `--unlink` without a SPEC).
    fn link(&self, n: i64, spec: Option<&str>) -> Output {
        let name = format!("C-{n}");
        match spec {
            Some(spec) => self.card(&["spec", &name, spec, "--by", "leader"]),
            None => self.card(&["spec", &name, "--unlink", "--by", "leader"]),
        }
    }

    /// `card withdraw C-n`, which puts the card on `hold`.
    fn withdraw(&self, n: i64) {
        let out = self
            .home
            .command(&[
                "card",
                "withdraw",
                &format!("C-{n}"),
                "--reason",
                "fixture",
                "--by",
                "leader",
                self.p(),
            ])
            .output()
            .expect("run moai");
        ok(&out);
    }

    /// A link that an earlier build allowed, written as the ledger would hold it: a `card.spec`
    /// event with no relation, made without the refusal of this revision.
    fn legacy_link(&self, n: i64, spec: &str) {
        let raw =
            rusqlite::Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap())
                .unwrap();
        raw.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), 'test', ?2, 'card.spec', ?3)",
            rusqlite::params![
                format!("card.spec:{n}:legacy:{spec}"),
                n,
                json!({"from": Value::Null, "to": spec}).to_string()
            ],
        )
        .unwrap();
    }

    /// The `state` column of card `n`.
    fn state(&self, n: i64) -> String {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn()
            .query_row("SELECT state FROM card WHERE card_id = ?1", [n], |r| {
                r.get(0)
            })
            .unwrap()
    }
}

/// The words every refusal of the open-card rule carries: the kind, the SPEC, the open card, and
/// the two ways out the rule names.
fn assert_open_card_refusal(out: &Output, spec: &str, open_card: &str, verb: &str) {
    let (kind, message) = refusal(out);
    assert_eq!(kind, "conflict", "{message}");
    for word in [
        verb,
        "conflict in spec",
        spec,
        open_card,
        "--follow-up",
        "card criterion-add",
    ] {
        assert!(
            message.contains(word),
            "the refusal lacks {word}: {message}"
        );
    }
}

/// AC-10 (REQ-8): a `hold` card that names a SPEC without a relation does not come back to
/// `ready` beside another open card of that SPEC. Nothing is written by the refusal.
#[test]
fn a_hold_card_does_not_become_ready_beside_an_open_card_of_its_spec() {
    // C-1 is withdrawn, then C-2 takes the SPEC and stays open.
    let setup = || {
        let w = World::new();
        let spec = w.spec("login-fix");
        ok(&w.add(&spec, &[]));
        w.withdraw(1);
        ok(&w.add(&spec, &[]));
        (w, spec)
    };
    let (w, spec) = setup();
    let events = w.count("SELECT count(*) FROM event");
    assert_open_card_refusal(&w.ready(1), &spec, "C-2", "card ready");
    assert_eq!(w.state(1), "hold", "the refusal changed the card");
    assert_eq!(w.count("SELECT count(*) FROM event"), events);
    // The open card is done: the same call is accepted.
    w.close(2);
    assert_eq!(ok(&w.ready(1))["state"], json!("ready"));

    // The open card is on hold instead: accepted as well.
    let (w, _) = setup();
    w.withdraw(2);
    assert_eq!(ok(&w.ready(1))["state"], json!("ready"));

    // A follow-up card on hold carries a relation and is never refused by this rule.
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &[]));
    ok(&w.add(&spec, &["--follow-up"]));
    w.withdraw(2);
    assert_eq!(w.state(2), "hold");
    assert_eq!(ok(&w.ready(2))["state"], json!("ready"));

    // A hold card that names no SPEC is never refused by this rule.
    let w = World::new();
    let spec = w.spec("login-fix");
    ok(&w.add(&spec, &[]));
    ok(&w.card(&["add", "--title", "t", "--ac", "fixture"]));
    w.withdraw(2);
    assert_eq!(ok(&w.ready(2))["state"], json!("ready"));
}

/// AC-10 (REQ-8) and the leader's note on the refusal: a `draft` card is never refused, and two
/// open cards that an earlier build left on one SPEC are not touched by the rule.
#[test]
fn a_draft_card_and_a_legacy_pair_of_open_cards_still_go_ready() {
    let w = World::new();
    let spec = w.spec("login-fix");
    let other = w.spec("other-fix");
    ok(&w.add(&spec, &[])); // C-1
    ok(&w.add(&other, &[])); // C-2
    w.legacy_link(2, &spec); // C-1 and C-2 are both open on one SPEC now
    assert_eq!(ok(&w.ready(2))["state"], json!("ready"));
    assert_eq!(ok(&w.ready(1))["state"], json!("ready"));
}

/// AC-11 (REQ-9): `card spec` does not add a second open card to a SPEC; the link of a `done` or
/// `hold` card, a link to the SPEC the card already names, `--unlink` and a link to a SPEC with no
/// open card are accepted. A refusal writes nothing.
#[test]
fn card_spec_does_not_add_a_second_open_card_to_a_spec() {
    let w = World::new();
    let spec = w.spec("login-fix");
    let other = w.spec("other-fix");
    ok(&w.add(&spec, &[])); // C-1, open on `spec`
    for _ in 0..3 {
        ok(&w.card(&["add", "--title", "t", "--ac", "fixture"])); // C-2, C-3, C-4: no SPEC
    }
    w.close(3);
    w.withdraw(4);

    let events = w.count("SELECT count(*) FROM event");
    assert_open_card_refusal(&w.link(2, Some(&spec)), &spec, "C-1", "card spec");
    assert_eq!(w.count("SELECT count(*) FROM event"), events);
    assert_eq!(ok(&w.card(&["show", "C-2"]))["spec"], Value::Null);

    // A done card and a hold card link to it, and neither one makes a second open card.
    assert_eq!(ok(&w.link(3, Some(&spec)))["changed"], json!(true));
    assert_eq!(ok(&w.link(4, Some(&spec)))["changed"], json!(true));
    // The same SPEC again writes nothing.
    assert_eq!(ok(&w.link(1, Some(&spec)))["changed"], json!(false));
    // A SPEC with no open card takes an open card.
    assert_eq!(ok(&w.link(2, Some(&other)))["changed"], json!(true));
    // That card moves to `spec` only when `spec` has no other open card: it has C-1.
    assert_open_card_refusal(&w.link(2, Some(&spec)), &spec, "C-1", "card spec");

    // A pair that an earlier build left on one SPEC: `--unlink` and the same link still work.
    w.legacy_link(2, &spec);
    assert_eq!(ok(&w.link(2, Some(&spec)))["changed"], json!(false));
    assert_eq!(ok(&w.link(2, None))["changed"], json!(true));
}

/// AC-12 (REQ-10): the help of `card ready` and `card spec` names the refusal and the way out, in
/// words the refusal text carries.
#[test]
fn the_ready_and_spec_help_name_the_open_card_refusal() {
    let w = World::new();
    for verb in ["ready", "spec"] {
        let out = w
            .home
            .command(&["card", verb, "--help"])
            .output()
            .expect("run moai");
        let help = String::from_utf8_lossy(&out.stdout)
            .split_whitespace()
            .collect::<Vec<_>>()
            .join(" ");
        for word in [
            "conflict in spec",
            "another open card",
            "--follow-up",
            "card criterion-add",
        ] {
            assert!(
                help.contains(word),
                "`card {verb} --help` lacks {word}: {help}"
            );
        }
    }
}
