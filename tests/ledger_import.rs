//! The SQLite side of a host import (#3071 H3a): what one batch writes, what a second run adds
//! (nothing), and that a batch that fails writes none of itself. Every test runs in its own
//! temporary HOME; none touches the real `~/.moai`.

mod common;

use common::home::TempHome;
use moai_cli::forge::import_apply::{
    BatchAc, BatchCard, BatchNote, CardOutcome, ImportBatch, ImportLedger,
};
use moai_cli::ledger::{self, import::SqliteImport};

const ID: &str = "demo";

fn open() -> (TempHome, ledger::open::WriteLedger) {
    let home = TempHome::new();
    ledger::register(home.path(), ID).unwrap();
    let write = ledger::open_write(home.path(), ID).unwrap();
    (home, write)
}

fn card(number: u64, title: &str) -> BatchCard {
    BatchCard {
        source_host: "github",
        source_project: "owner/repo".into(),
        source_number: number,
        title: title.into(),
        kind: "task",
        body: format!("body {number}"),
        acs: Vec::new(),
        notes: Vec::new(),
        event_key: format!("import:github:owner/repo:{number}:h"),
        event_payload: "{\"digest\":\"d\"}".into(),
    }
}

fn with_rows(mut c: BatchCard) -> BatchCard {
    c.acs = vec![
        BatchAc {
            text: "first".into(),
            verify: "run it".into(),
        },
        BatchAc {
            text: "second".into(),
            verify: String::new(),
        },
    ];
    c.notes = vec![
        BatchNote {
            reference: "101".into(),
            kind: "comment",
            author: "alice".into(),
            body: "a comment".into(),
        },
        BatchNote {
            reference: "102".into(),
            kind: "observation",
            author: "bob".into(),
            body: "a reservation line, kept as text".into(),
        },
    ];
    c
}

fn batch(cards: Vec<BatchCard>) -> ImportBatch {
    ImportBatch {
        digest: "sha256:d".into(),
        cards,
    }
}

fn count(w: &ledger::open::WriteLedger, table: &str) -> i64 {
    w.conn()
        .query_row(&format!("SELECT count(*) FROM {table}"), [], |r| r.get(0))
        .unwrap()
}

fn counts(w: &ledger::open::WriteLedger) -> [i64; 4] {
    ["card", "card_ac", "card_note", "event"].map(|t| count(w, t))
}

#[test]
fn a_batch_writes_cards_criteria_notes_and_one_event_each() {
    let (_home, mut w) = open();
    let out = SqliteImport::new(&mut w)
        .commit(&batch(vec![with_rows(card(7, "Seven")), {
            let mut eight = card(8, "Eight");
            eight.kind = "fix";
            eight
        }]))
        .unwrap();
    assert_eq!(
        out.outcomes,
        vec![CardOutcome::Created, CardOutcome::Created]
    );
    assert_eq!(counts(&w), [2, 2, 2, 2]);
    let kinds: Vec<(i64, String, String, String)> = w
        .conn()
        .prepare("SELECT source_number, kind, title, body FROM card ORDER BY source_number")
        .unwrap()
        .query_map([], |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?, r.get(3)?)))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    assert_eq!(
        kinds,
        vec![
            (7, "task".into(), "Seven".into(), "body 7".into()),
            (8, "fix".into(), "Eight".into(), "body 8".into())
        ]
    );

    let (state, by, host, project, number, rev, created, updated): (
        String,
        String,
        String,
        String,
        i64,
        i64,
        i64,
        i64,
    ) = w
        .conn()
        .query_row(
            "SELECT state, created_by, source_host, source_project, source_number, rev, \
                    created_at, updated_at FROM card WHERE source_number = 7",
            [],
            |r| {
                Ok((
                    r.get(0)?,
                    r.get(1)?,
                    r.get(2)?,
                    r.get(3)?,
                    r.get(4)?,
                    r.get(5)?,
                    r.get(6)?,
                    r.get(7)?,
                ))
            },
        )
        .unwrap();
    assert_eq!((state.as_str(), by.as_str()), ("draft", "import"));
    assert_eq!(
        (host.as_str(), project.as_str(), number),
        ("github", "owner/repo", 7)
    );
    assert_eq!(
        rev, 3,
        "one revision per criterion row, from the schema's trigger"
    );
    // Times come from the transaction's DB clock: a real time, the same for the card and its event.
    assert!(
        created > 1_700_000_000 && updated >= created,
        "{created} {updated}"
    );
    let event_at: i64 = w
        .conn()
        .query_row(
            "SELECT at FROM event WHERE event_key LIKE '%:7:%'",
            [],
            |r| r.get(0),
        )
        .unwrap();
    assert_eq!(created, event_at);

    let seqs: Vec<(i64, String, String)> = w
        .conn()
        .prepare("SELECT seq, text, status FROM card_ac ORDER BY seq")
        .unwrap()
        .query_map([], |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?)))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    assert_eq!(
        seqs,
        vec![
            (1, "first".to_string(), "open".to_string()),
            (2, "second".to_string(), "open".to_string())
        ]
    );
    let notes: Vec<(String, String, String, String)> = w
        .conn()
        .prepare("SELECT ref, kind, author, origin FROM card_note ORDER BY ref")
        .unwrap()
        .query_map([], |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?, r.get(3)?)))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    assert_eq!(
        notes,
        vec![
            (
                "101".into(),
                "comment".into(),
                "alice".into(),
                "import".into()
            ),
            (
                "102".into(),
                "observation".into(),
                "bob".into(),
                "import".into()
            )
        ]
    );
    let (key, actor, kind): (String, String, String) = w
        .conn()
        .query_row(
            "SELECT event_key, actor, kind FROM event WHERE event_key LIKE '%:7:%'",
            [],
            |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?)),
        )
        .unwrap();
    assert_eq!(
        (key.as_str(), actor.as_str(), kind.as_str()),
        ("import:github:owner/repo:7:h", "import", "import")
    );
    let (event_card, payload): (i64, String) = w
        .conn()
        .query_row(
            "SELECT card_id, payload FROM event WHERE event_key LIKE '%:7:%'",
            [],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .unwrap();
    let card_seven: i64 = w
        .conn()
        .query_row(
            "SELECT card_id FROM card WHERE source_number = 7",
            [],
            |r| r.get(0),
        )
        .unwrap();
    assert_eq!(event_card, card_seven, "the event belongs to its card");
    // Every card's own event is on that card, not only the first one's.
    let joined: i64 = w
        .conn()
        .query_row(
            "SELECT count(*) FROM event e JOIN card c ON c.card_id = e.card_id \
             WHERE e.event_key = 'import:github:owner/repo:' || c.source_number || ':h'",
            [],
            |r| r.get(0),
        )
        .unwrap();
    assert_eq!(joined, 2);
    assert_eq!(payload, "{\"digest\":\"d\"}");
    // Nothing but text: no reservation, no link, no state beyond draft.
    assert_eq!((count(&w, "reservation"), count(&w, "card_link")), (0, 0));
    assert_eq!(
        w.conn()
            .query_row(
                "SELECT count(*) FROM card WHERE state <> 'draft'",
                [],
                |r| r.get::<_, i64>(0)
            )
            .unwrap(),
        0
    );
}

#[test]
fn a_second_run_adds_nothing_and_leaves_the_first_card_as_it_was() {
    let (_home, mut w) = open();
    let first = batch(vec![with_rows(card(7, "Seven"))]);
    SqliteImport::new(&mut w).commit(&first).unwrap();
    let before = counts(&w);
    let rev_before: i64 = w
        .conn()
        .query_row("SELECT rev FROM card", [], |r| r.get(0))
        .unwrap();

    // The same batch.
    let again = SqliteImport::new(&mut w).commit(&first).unwrap();
    assert_eq!(again.outcomes, vec![CardOutcome::AlreadyThere]);
    assert_eq!(counts(&w), before);

    // The host changed the issue meanwhile: another title, another body, a new comment.
    let mut changed = with_rows(card(7, "Seven, renamed"));
    changed.body = "new body".into();
    changed.event_key = "import:github:owner/repo:7:h2".into();
    changed.notes.push(BatchNote {
        reference: "103".into(),
        kind: "comment",
        author: "carol".into(),
        body: "late".into(),
    });
    let out = SqliteImport::new(&mut w)
        .commit(&batch(vec![changed]))
        .unwrap();
    assert_eq!(out.outcomes, vec![CardOutcome::AlreadyThere]);
    assert_eq!(counts(&w), before, "no note, event or criterion is added");
    let (title, body, rev): (String, String, i64) = w
        .conn()
        .query_row("SELECT title, body, rev FROM card", [], |r| {
            Ok((r.get(0)?, r.get(1)?, r.get(2)?))
        })
        .unwrap();
    assert_eq!((title.as_str(), body.as_str()), ("Seven", "body 7"));
    assert_eq!(rev, rev_before);
}

#[test]
fn outcomes_follow_the_order_of_the_batch_when_some_cards_are_there() {
    let (_home, mut w) = open();
    SqliteImport::new(&mut w)
        .commit(&batch(vec![card(8, "Eight")]))
        .unwrap();
    let out = SqliteImport::new(&mut w)
        .commit(&batch(vec![
            card(7, "Seven"),
            card(8, "Eight"),
            card(9, "Nine"),
        ]))
        .unwrap();
    assert_eq!(
        out.outcomes,
        vec![
            CardOutcome::Created,
            CardOutcome::AlreadyThere,
            CardOutcome::Created
        ]
    );
    assert_eq!(count(&w, "card"), 3);
}

#[test]
fn one_source_twice_in_a_batch_makes_one_card() {
    let (_home, mut w) = open();
    let mut dup = card(7, "Seven again");
    dup.event_key = "import:github:owner/repo:7:other".into();
    let out = SqliteImport::new(&mut w)
        .commit(&batch(vec![card(7, "Seven"), dup]))
        .unwrap();
    assert_eq!(
        out.outcomes,
        vec![CardOutcome::Created, CardOutcome::AlreadyThere]
    );
    assert_eq!(count(&w, "card"), 1);
}

#[test]
fn the_same_number_in_another_project_or_host_is_another_card() {
    let (_home, mut w) = open();
    let mut other_project = card(7, "Seven");
    other_project.source_project = "owner/other".into();
    other_project.event_key = "import:github:owner/other:7:h".into();
    let mut other_host = card(7, "Seven");
    other_host.source_host = "gitlab";
    other_host.event_key = "import:gitlab:owner/repo:7:h".into();
    let out = SqliteImport::new(&mut w)
        .commit(&batch(vec![card(7, "Seven"), other_project, other_host]))
        .unwrap();
    assert_eq!(out.outcomes, vec![CardOutcome::Created; 3]);
    assert_eq!(count(&w, "card"), 3);
}

#[test]
fn a_batch_that_fails_halfway_writes_none_of_itself() {
    let (_home, mut w) = open();
    SqliteImport::new(&mut w)
        .commit(&batch(vec![card(1, "Already there")]))
        .unwrap();
    let before = counts(&w);
    // The third card has an empty criterion text, which the schema refuses.
    let mut bad = with_rows(card(9, "Bad"));
    bad.acs[1].text = String::new();
    let result = SqliteImport::new(&mut w).commit(&batch(vec![
        with_rows(card(7, "Seven")),
        card(8, "Eight"),
        bad,
    ]));
    assert!(result.is_err(), "{result:?}");
    assert_eq!(counts(&w), before, "the two good cards are not kept");
    // The ledger is still usable afterwards.
    let ok = SqliteImport::new(&mut w)
        .commit(&batch(vec![card(7, "Seven")]))
        .unwrap();
    assert_eq!(ok.outcomes, vec![CardOutcome::Created]);
}

#[test]
fn a_title_may_be_300_characters_and_not_301_or_none() {
    let (_home, mut w) = open();
    let ok = SqliteImport::new(&mut w)
        .commit(&batch(vec![card(1, &"t".repeat(300))]))
        .unwrap();
    assert_eq!(ok.outcomes, vec![CardOutcome::Created]);
    for (n, title) in [(2, "t".repeat(301)), (3, String::new())] {
        let before = counts(&w);
        let err = SqliteImport::new(&mut w).commit(&batch(vec![card(n, &title)]));
        assert!(
            err.is_err(),
            "title of {} characters",
            title.chars().count()
        );
        assert_eq!(counts(&w), before);
    }
}

#[test]
fn an_event_key_that_is_already_used_stops_the_whole_batch() {
    let (_home, mut w) = open();
    SqliteImport::new(&mut w)
        .commit(&batch(vec![card(1, "One")]))
        .unwrap();
    let before = counts(&w);
    let mut clash = card(2, "Two");
    clash.event_key = "import:github:owner/repo:1:h".into();
    let result = SqliteImport::new(&mut w).commit(&batch(vec![card(3, "Three"), clash]));
    assert!(result.is_err());
    assert_eq!(counts(&w), before);
}

#[test]
fn an_empty_batch_writes_nothing_and_succeeds() {
    let (_home, mut w) = open();
    let out = SqliteImport::new(&mut w)
        .commit(&batch(Vec::new()))
        .unwrap();
    assert!(out.outcomes.is_empty());
    assert_eq!(counts(&w), [0, 0, 0, 0]);
}

#[test]
fn a_fault_names_the_reason_without_the_issue_text() {
    let (_home, mut w) = open();
    let mut bad = card(1, "A very private title");
    bad.body = "a very private body".into();
    bad.acs = vec![BatchAc {
        text: String::new(),
        verify: "a very private verify".into(),
    }];
    let fault = SqliteImport::new(&mut w)
        .commit(&batch(vec![bad]))
        .unwrap_err();
    assert!(!fault.0.is_empty());
    assert!(!fault.0.contains("private"), "{}", fault.0);
}
