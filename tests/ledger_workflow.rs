//! The finish path of a card in the ledger library (card C-53 S1): next-card selection, notes,
//! criteria, the pull request link, `review` and `done`, and the write-boundary predicate. Every
//! test runs in its own temporary HOME; none touches the real `~/.moai`, and nothing here reads
//! a code host: the state of a pull request is a value the test passes in.

mod common;

use common::home::TempHome;
use moai_cli::ledger::reservation::{self, FenceError};
use moai_cli::ledger::workflow::{
    self, CriterionChange, PrTarget, SkipReason, WorkflowError, MAX_LINE_CHARS, MAX_NOTE_BYTES,
    NOTE_KINDS,
};
use moai_cli::ledger::{self, WriteLedger};
use moai_cli::scope;
use rusqlite::Connection;

const PROJECT: &str = "demo";
const A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

fn setup() -> (TempHome, WriteLedger) {
    let home = TempHome::new();
    ledger::register(home.path(), PROJECT).unwrap();
    let w = ledger::open_write(home.path(), PROJECT).unwrap();
    (home, w)
}

/// Inserts a card directly (an `INSERT` has no state check) with `criteria` open criteria.
fn add_card(w: &mut WriteLedger, state: &str, priority: &str, scope: &str, criteria: i64) -> i64 {
    w.with_write_tx(|tx, now| {
        tx.execute(
            "INSERT INTO card (title, state, priority, write_scope, created_by, created_at, updated_at) \
             VALUES ('card', ?1, ?2, ?3, 'test', ?4, ?4)",
            rusqlite::params![state, priority, scope, now],
        )?;
        let id = tx.last_insert_rowid();
        for seq in 1..=criteria {
            tx.execute(
                "INSERT INTO card_ac (card_id, seq, text, updated_at) VALUES (?1, ?2, 'ac', ?3)",
                rusqlite::params![id, seq, now],
            )?;
        }
        Ok(id)
    })
    .unwrap()
}

/// A ready card with its own scope (`zone<n>/`) and `criteria` open criteria, taken by `holder`.
/// Returns `(card, generation)`.
fn taken(w: &mut WriteLedger, holder: &str, criteria: i64) -> (i64, i64) {
    let card = add_card(w, "ready", "p2", "[]", criteria);
    let r =
        reservation::reserve(w, card, holder, &[format!("zone{card}/")], 3600, None, None).unwrap();
    (card, r.generation)
}

fn sql(w: &mut WriteLedger, statement: &str) {
    w.with_write_tx(|tx, _| {
        tx.execute_batch(statement)?;
        Ok(())
    })
    .unwrap();
}

fn one<T: rusqlite::types::FromSql>(conn: &Connection, statement: &str) -> T {
    conn.query_row(statement, [], |r| r.get(0)).unwrap()
}

fn merged(number: i64) -> PrTarget<'static> {
    PrTarget {
        host: "github",
        project: "octo/demo",
        number,
        pr_state: "merged",
        merge_ref: Some("0123456789abcdef0123456789abcdef01234567"),
    }
}

fn open_pr(number: i64) -> PrTarget<'static> {
    PrTarget {
        host: "github",
        project: "octo/demo",
        number,
        pr_state: "open",
        merge_ref: None,
    }
}

fn met(w: &mut WriteLedger, card: i64, holder: &str, generation: i64, seq: i64) {
    workflow::set_criterion(
        w,
        card,
        holder,
        generation,
        seq,
        CriterionChange::Met { evidence: "run 1" },
    )
    .unwrap();
}

fn fence_rejects(conn: &Connection, op: &str, reason: &str) -> i64 {
    conn.query_row(
        "SELECT count(*) FROM event WHERE kind = 'fence_reject' \
         AND json_extract(payload, '$.op') = ?1 AND json_extract(payload, '$.reason') = ?2",
        [op, reason],
        |r| r.get(0),
    )
    .unwrap()
}

// ---- next --------------------------------------------------------------------------------------

#[test]
fn next_gives_my_cards_then_ready_cards_by_priority_then_number_and_a_reason_for_every_skip() {
    let (_home, mut w) = setup();
    let low = add_card(&mut w, "ready", "p3", r#"["low/"]"#, 0); // 1
    let mine = add_card(&mut w, "ready", "p2", r#"["mine/"]"#, 0); // 2
    let late = add_card(&mut w, "ready", "p1", r#"["late/"]"#, 0); // 3
    let top = add_card(&mut w, "ready", "p0", r#"["top/"]"#, 0); // 4
    let early = add_card(&mut w, "ready", "p1", r#"["early/"]"#, 0); // 5
    let theirs = add_card(&mut w, "ready", "p0", r#"["theirs/"]"#, 0); // 6
    let clash = add_card(&mut w, "ready", "p0", r#"["theirs/deep/x.rs"]"#, 0); // 7
    let bare = add_card(&mut w, "ready", "p0", "[]", 0); // 8
    let draft = add_card(&mut w, "draft", "p0", r#"["draft/"]"#, 0); // 9
    let mine_r = reservation::reserve(
        &mut w,
        mine,
        A,
        &["mine/".to_string()],
        3600,
        Some("/w/a"),
        Some("card/2-x"),
    )
    .unwrap();
    let theirs_r = reservation::reserve(
        &mut w,
        theirs,
        B,
        &["theirs/".to_string()],
        3600,
        None,
        None,
    )
    .unwrap();
    let before: i64 = one(w.conn(), "SELECT count(*) FROM event");

    let n = workflow::next(w.conn(), Some(A)).unwrap();
    assert_eq!(n.mine.len(), 1);
    assert_eq!(n.mine[0].reservation, mine_r);
    assert_eq!(n.mine[0].state, "active");
    assert!(n.mine[0].lease_live);
    assert_eq!(n.mine[0].reservation.worktree.as_deref(), Some("/w/a"));
    assert_eq!(n.mine[0].reservation.branch.as_deref(), Some("card/2-x"));
    let ready: Vec<(i64, &str)> = n
        .ready
        .iter()
        .map(|c| (c.card_id, c.priority.as_str()))
        .collect();
    assert_eq!(
        ready,
        [(top, "p0"), (late, "p1"), (early, "p1"), (low, "p3")],
        "priority first, then the lower number"
    );
    assert_eq!(n.ready[0].write_scope, ["top/"]);
    let skipped: Vec<(i64, &str)> = n
        .skipped
        .iter()
        .map(|s| (s.card_id, s.reason.code()))
        .collect();
    assert_eq!(
        skipped,
        [
            (theirs, "held"),
            (clash, "scope_overlap"),
            (bare, "empty_scope")
        ]
    );
    assert_eq!(
        n.skipped[0].reason,
        SkipReason::Held {
            holder: B.to_string(),
            generation: theirs_r.generation,
            lease_until: theirs_r.lease_until,
        }
    );
    assert_eq!(
        n.skipped[1].reason,
        SkipReason::ScopeOverlap {
            other_card: theirs,
            mine: "theirs/deep/x.rs".to_string(),
            theirs: "theirs/".to_string(),
        }
    );
    assert!(
        n.ready.iter().all(|c| c.card_id != draft) && n.skipped.iter().all(|s| s.card_id != draft),
        "a draft card is neither offered nor skipped"
    );
    assert_eq!(
        one::<i64>(w.conn(), "SELECT count(*) FROM event"),
        before,
        "next writes nothing"
    );

    // The other holder sees the mirror image; a caller without a session holds nothing.
    let other = workflow::next(w.conn(), Some(B)).unwrap();
    assert_eq!(other.mine.len(), 1);
    assert_eq!(other.mine[0].reservation.card_id, theirs);
    let held: Vec<i64> = other
        .skipped
        .iter()
        .filter(|s| s.reason.code() == "held")
        .map(|s| s.card_id)
        .collect();
    assert_eq!(held, [mine]);
    assert!(
        other
            .skipped
            .iter()
            .any(|s| s.card_id == clash && s.reason.code() == "scope_overlap"),
        "a card that overlaps the caller's own reservation cannot be taken either"
    );
    let nobody = workflow::next(w.conn(), None).unwrap();
    assert!(nobody.mine.is_empty());
    assert_eq!(
        nobody
            .skipped
            .iter()
            .filter(|s| s.reason.code() == "held")
            .count(),
        2
    );
}

#[test]
fn next_stops_offering_a_card_the_moment_another_holder_takes_it() {
    let (_home, mut w) = setup();
    let card = add_card(&mut w, "ready", "p2", r#"["src/a/"]"#, 0);
    let n = workflow::next(w.conn(), Some(A)).unwrap();
    assert_eq!(n.ready.len(), 1);
    assert_eq!(n.ready[0].card_id, card);
    assert!(n.skipped.is_empty());

    let r =
        reservation::reserve(&mut w, card, B, &["src/a/".to_string()], 3600, None, None).unwrap();
    let n = workflow::next(w.conn(), Some(A)).unwrap();
    assert!(n.ready.is_empty() && n.mine.is_empty());
    assert_eq!(n.skipped.len(), 1);
    assert_eq!(n.skipped[0].state, "active");
    assert_eq!(
        n.skipped[0].reason,
        SkipReason::Held {
            holder: B.to_string(),
            generation: 1,
            lease_until: r.lease_until
        }
    );

    // Given back: offered again.
    reservation::release(&mut w, card, B, 1).unwrap();
    let n = workflow::next(w.conn(), Some(A)).unwrap();
    assert_eq!(n.ready.len(), 1);
    assert!(n.skipped.is_empty());
}

#[test]
fn next_marks_an_ended_lease_and_leaves_closed_cards_out() {
    let (_home, mut w) = setup();
    let (card, _) = taken(&mut w, A, 0);
    add_card(&mut w, "hold", "p0", r#"["h/"]"#, 0);
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() WHERE card_id = 1",
    );
    let n = workflow::next(w.conn(), Some(A)).unwrap();
    assert_eq!(n.mine[0].reservation.card_id, card);
    assert!(
        !n.mine[0].lease_live,
        "the lease-end second no longer holds"
    );
    assert!(n.ready.is_empty() && n.skipped.is_empty());
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() + 1 WHERE card_id = 1",
    );
    assert!(
        workflow::next(w.conn(), Some(A)).unwrap().mine[0].lease_live,
        "one second before the end it still holds"
    );
}

// ---- notes -------------------------------------------------------------------------------------

#[test]
fn a_note_is_appended_with_a_ref_of_its_own_and_one_event() {
    let (_home, mut w) = setup();
    let card = add_card(&mut w, "draft", "p2", "[]", 0);
    let first =
        workflow::add_note(&mut w, card, "report", A, None, "PLAN C-1\nstep one", None).unwrap();
    let second =
        workflow::add_note(&mut w, card, "report", A, None, "PLAN C-1\nstep one", None).unwrap();
    assert_eq!((first.note_id, second.note_id), (1, 2));
    assert_ne!(
        first.reference, second.reference,
        "each note has its own ref"
    );
    assert!(
        first.reference.starts_with("note:1:"),
        "{}",
        first.reference
    );
    assert_eq!(first.reference.len(), "note:1:".len() + 32);
    assert_eq!(first.bytes, "PLAN C-1\nstep one".len());
    let named = workflow::add_note(
        &mut w,
        card,
        "decision",
        "user",
        None,
        "answer",
        Some("decision-7"),
    )
    .unwrap();
    assert_eq!(named.reference, "decision-7");
    assert_eq!(
        one::<String>(
            w.conn(),
            "SELECT group_concat(kind || '|' || author || '|' || origin || '|' || body, ';') \
             FROM (SELECT * FROM card_note ORDER BY note_id)"
        ),
        format!(
            "report|{A}|local|PLAN C-1\nstep one;report|{A}|local|PLAN C-1\nstep one;decision|user|local|answer"
        )
    );
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.note'"
        ),
        3
    );
    assert_eq!(
        one::<String>(
            w.conn(),
            "SELECT json_extract(payload, '$.ref') FROM event WHERE kind = 'card.note' \
             ORDER BY event_id DESC LIMIT 1"
        ),
        "decision-7"
    );

    // The same ref on the same card is refused and nothing is added; on another card it is free.
    let again = workflow::add_note(
        &mut w,
        card,
        "comment",
        "user",
        None,
        "x",
        Some("decision-7"),
    );
    assert_eq!(
        again,
        Err(WorkflowError::DuplicateRef {
            card_id: card,
            reference: "decision-7".to_string()
        })
    );
    assert_eq!(one::<i64>(w.conn(), "SELECT count(*) FROM card_note"), 3);
    let other = add_card(&mut w, "draft", "p2", "[]", 0);
    workflow::add_note(
        &mut w,
        other,
        "comment",
        "user",
        None,
        "x",
        Some("decision-7"),
    )
    .unwrap();
}

#[test]
fn notes_are_append_only() {
    let (_home, mut w) = setup();
    let card = add_card(&mut w, "draft", "p2", "[]", 0);
    workflow::add_note(&mut w, card, "comment", "user", None, "first", None).unwrap();
    for statement in [
        "UPDATE card_note SET body = 'changed' WHERE note_id = 1",
        "DELETE FROM card_note WHERE note_id = 1",
    ] {
        let refused = w.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        });
        assert!(refused.is_err(), "{statement}");
    }
    workflow::add_note(&mut w, card, "comment", "user", None, "second", None).unwrap();
    assert_eq!(
        one::<String>(
            w.conn(),
            "SELECT group_concat(body, ',') FROM (SELECT body FROM card_note ORDER BY note_id)"
        ),
        "first,second"
    );
}

#[test]
fn a_note_has_one_of_four_kinds_a_body_within_its_limits_and_a_one_line_ref() {
    let (_home, mut w) = setup();
    let card = add_card(&mut w, "draft", "p2", "[]", 0);
    for kind in NOTE_KINDS {
        workflow::add_note(&mut w, card, kind, "user", None, "x", None).unwrap();
    }
    let bad = |r: Result<workflow::Note, WorkflowError>| match r {
        Err(WorkflowError::InvalidInput(why)) => why,
        other => panic!("{other:?}"),
    };
    assert!(bad(workflow::add_note(
        &mut w, card, "plan", "user", None, "x", None
    ))
    .contains("kind"));
    assert!(bad(workflow::add_note(
        &mut w, card, "", "user", None, "x", None
    ))
    .contains("kind"));
    assert!(bad(workflow::add_note(
        &mut w, card, "comment", "user", None, "", None
    ))
    .contains("empty"));
    assert!(bad(workflow::add_note(
        &mut w, card, "comment", "user", None, "a\0b", None
    ))
    .contains("NUL"));
    let edge = "a".repeat(MAX_NOTE_BYTES);
    assert_eq!(
        workflow::add_note(&mut w, card, "comment", "user", None, &edge, None)
            .unwrap()
            .bytes,
        MAX_NOTE_BYTES
    );
    let over = "a".repeat(MAX_NOTE_BYTES + 1);
    assert!(bad(workflow::add_note(
        &mut w, card, "comment", "user", None, &over, None
    ))
    .contains("1048577"));
    let long = "r".repeat(MAX_LINE_CHARS);
    workflow::add_note(&mut w, card, "comment", "user", None, "x", Some(&long)).unwrap();
    let longer = "r".repeat(MAX_LINE_CHARS + 1);
    for reference in [longer.as_str(), "", " ", "two\nlines"] {
        bad(workflow::add_note(
            &mut w,
            card,
            "comment",
            "user",
            None,
            "x",
            Some(reference),
        ));
    }
    for author in ["", " ", "two\nlines"] {
        bad(workflow::add_note(
            &mut w, card, "comment", author, None, "x", None,
        ));
    }
    assert_eq!(
        workflow::add_note(&mut w, 99, "comment", "user", None, "x", None),
        Err(WorkflowError::CardNotFound(99))
    );
    assert_eq!(
        one::<i64>(w.conn(), "SELECT count(*) FROM card_note"),
        6,
        "four kinds, the largest body and the longest ref; no refused note left a row"
    );
}

// ---- criteria ----------------------------------------------------------------------------------

#[test]
fn the_holder_marks_a_criterion_met_with_evidence_or_withdrawn_with_a_reason() {
    let (_home, mut w) = setup();
    let (card, g) = taken(&mut w, A, 2);
    let rev: i64 = one(w.conn(), "SELECT rev FROM card WHERE card_id = 1");
    let c = workflow::set_criterion(
        &mut w,
        card,
        A,
        g,
        1,
        CriterionChange::Met {
            evidence: "ci run 42",
        },
    )
    .unwrap();
    assert_eq!((c.from.as_str(), c.status.as_str()), ("open", "met"));
    assert_eq!(c.evidence_ref.as_deref(), Some("ci run 42"));
    let c = workflow::set_criterion(
        &mut w,
        card,
        A,
        g,
        2,
        CriterionChange::Withdrawn {
            reason: "moved to C-9",
        },
    )
    .unwrap();
    assert_eq!((c.from.as_str(), c.status.as_str()), ("open", "withdrawn"));
    assert_eq!(c.withdrawn_reason.as_deref(), Some("moved to C-9"));
    assert_eq!(
        one::<String>(
            w.conn(),
            "SELECT group_concat(seq || status || coalesce(evidence_ref, '-') || coalesce(withdrawn_reason, '-'), ';') \
             FROM (SELECT * FROM card_ac ORDER BY seq)"
        ),
        "1metci run 42-;2withdrawn-moved to C-9"
    );
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.criterion'"
        ),
        2
    );
    assert!(one::<i64>(w.conn(), "SELECT rev FROM card WHERE card_id = 1") > rev);

    // A met criterion can still be withdrawn; it keeps its evidence.
    let c = workflow::set_criterion(
        &mut w,
        card,
        A,
        g,
        1,
        CriterionChange::Withdrawn { reason: "wrong" },
    )
    .unwrap();
    assert_eq!((c.from.as_str(), c.status.as_str()), ("met", "withdrawn"));
    assert_eq!(c.evidence_ref.as_deref(), Some("ci run 42"));
}

#[test]
fn a_criterion_change_needs_a_real_criterion_and_a_one_line_text() {
    let (_home, mut w) = setup();
    let (card, g) = taken(&mut w, A, 1);
    for seq in [0, 2, -1] {
        let r = workflow::set_criterion(
            &mut w,
            card,
            A,
            g,
            seq,
            CriterionChange::Met { evidence: "e" },
        );
        match (seq, r) {
            (2, Err(WorkflowError::CriterionNotFound { card_id, seq: 2 })) => {
                assert_eq!(card_id, card)
            }
            (0 | -1, Err(WorkflowError::InvalidInput(_))) => {}
            other => panic!("{other:?}"),
        }
    }
    let edge = "e".repeat(MAX_LINE_CHARS);
    let over = "e".repeat(MAX_LINE_CHARS + 1);
    for text in ["", "  ", "two\nlines", over.as_str()] {
        assert!(matches!(
            workflow::set_criterion(
                &mut w,
                card,
                A,
                g,
                1,
                CriterionChange::Met { evidence: text }
            ),
            Err(WorkflowError::InvalidInput(_))
        ));
        assert!(matches!(
            workflow::set_criterion(
                &mut w,
                card,
                A,
                g,
                1,
                CriterionChange::Withdrawn { reason: text }
            ),
            Err(WorkflowError::InvalidInput(_))
        ));
    }
    assert_eq!(
        one::<String>(w.conn(), "SELECT status FROM card_ac WHERE card_id = 1"),
        "open"
    );
    workflow::set_criterion(
        &mut w,
        card,
        A,
        g,
        1,
        CriterionChange::Met { evidence: &edge },
    )
    .unwrap();
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.criterion'"
        ),
        1
    );
}

#[test]
fn a_stale_generation_another_holder_and_an_ended_lease_cannot_change_a_criterion() {
    let (_home, mut w) = setup();
    let (card, g1) = taken(&mut w, A, 1);
    let change = CriterionChange::Met { evidence: "e" };
    assert_eq!(
        workflow::set_criterion(&mut w, card, B, g1, 1, change),
        Err(WorkflowError::Fence(FenceError::WrongHolder {
            held_by: A.to_string()
        }))
    );
    let g2 = reservation::handoff(&mut w, card, A, g1, B)
        .unwrap()
        .generation;
    assert_eq!(
        workflow::set_criterion(&mut w, card, A, g1, 1, change),
        Err(WorkflowError::Fence(FenceError::StaleGeneration {
            offered: g1,
            current: g2
        }))
    );
    assert!(matches!(
        workflow::set_criterion(&mut w, card, B, g2 + 1, 1, change),
        Err(WorkflowError::Fence(FenceError::UnknownGeneration { .. }))
    ));
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() WHERE status = 'held'",
    );
    assert!(matches!(
        workflow::set_criterion(&mut w, card, B, g2, 1, change),
        Err(WorkflowError::Fence(FenceError::LeaseExpired { .. }))
    ));
    assert_eq!(
        one::<String>(w.conn(), "SELECT status FROM card_ac WHERE card_id = 1"),
        "open",
        "no refused call changed the criterion"
    );
    assert_eq!(fence_rejects(w.conn(), "criterion", "wrong_holder"), 1);
    assert_eq!(fence_rejects(w.conn(), "criterion", "stale_generation"), 1);
    assert_eq!(
        fence_rejects(w.conn(), "criterion", "unknown_generation"),
        1
    );
    assert_eq!(fence_rejects(w.conn(), "criterion", "lease_expired"), 1);
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() + 60 WHERE status = 'held'",
    );
    workflow::set_criterion(&mut w, card, B, g2, 1, change).unwrap();
}

// ---- pull request link -------------------------------------------------------------------------

#[test]
fn a_pull_request_is_linked_once_and_its_state_follows_the_host() {
    let (_home, mut w) = setup();
    let (card, g) = taken(&mut w, A, 0);
    let link = workflow::link_pr(&mut w, card, A, g, &open_pr(7)).unwrap();
    assert!(link.created);
    assert_eq!(
        (link.host.as_str(), link.project.as_str(), link.number),
        ("github", "octo/demo", 7)
    );
    assert_eq!((link.pr_state.as_str(), link.merge_ref), ("open", None));
    let again = workflow::link_pr(&mut w, card, A, g, &merged(7)).unwrap();
    assert!(!again.created, "the same pull request keeps its one link");
    assert_eq!(again.link_id, link.link_id);
    assert_eq!(again.pr_state, "merged");
    assert_eq!(
        again.merge_ref.as_deref(),
        Some("0123456789abcdef0123456789abcdef01234567")
    );
    workflow::link_pr(&mut w, card, A, g, &open_pr(8)).unwrap();
    assert_eq!(
        one::<String>(
            w.conn(),
            "SELECT group_concat(target_number || pr_state || state, ';') \
             FROM (SELECT * FROM card_link ORDER BY link_id)"
        ),
        "7mergedactive;8openactive"
    );
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.pr'"
        ),
        3
    );

    // A retired link stays retired: linking its pull request again is refused and changes
    // nothing (N8).
    sql(
        &mut w,
        "UPDATE card_link SET state = 'retired' WHERE target_number = 8",
    );
    assert_eq!(
        workflow::link_pr(&mut w, card, A, g, &merged(8)),
        Err(WorkflowError::LinkRetired {
            card_id: card,
            target: "github:octo/demo#8".to_string()
        })
    );
    assert_eq!(
        one::<String>(
            w.conn(),
            "SELECT state || '|' || pr_state || '|' || coalesce(merge_ref, '-') FROM card_link \
             WHERE target_number = 8"
        ),
        "retired|open|-"
    );
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.pr'"
        ),
        3,
        "a refused link leaves no event"
    );
}

#[test]
fn a_pull_request_link_checks_its_values_before_it_writes() {
    let (_home, mut w) = setup();
    let (card, g) = taken(&mut w, A, 0);
    let ok = merged(7);
    for bad in [
        PrTarget {
            host: "bitbucket",
            ..ok
        },
        PrTarget {
            project: "Octo/Demo",
            ..ok
        },
        PrTarget { project: "", ..ok },
        PrTarget {
            project: "octo demo",
            ..ok
        },
        PrTarget { number: 0, ..ok },
        PrTarget { number: -1, ..ok },
        PrTarget {
            pr_state: "draft",
            ..ok
        },
        PrTarget {
            merge_ref: None,
            ..ok
        },
        PrTarget {
            merge_ref: Some(""),
            ..ok
        },
        PrTarget {
            merge_ref: Some("a\nb"),
            ..ok
        },
    ] {
        assert!(
            matches!(
                workflow::link_pr(&mut w, card, A, g, &bad),
                Err(WorkflowError::InvalidInput(_))
            ),
            "{bad:?}"
        );
    }
    assert_eq!(one::<i64>(w.conn(), "SELECT count(*) FROM card_link"), 0);
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.pr'"
        ),
        0
    );
    workflow::link_pr(&mut w, card, A, g, &PrTarget { number: 1, ..ok }).unwrap();
}

#[test]
fn only_the_holder_at_the_current_generation_links_a_pull_request_also_after_the_lease() {
    let (_home, mut w) = setup();
    let (card, g1) = taken(&mut w, A, 0);
    assert!(matches!(
        workflow::link_pr(&mut w, card, B, g1, &open_pr(7)),
        Err(WorkflowError::Fence(FenceError::WrongHolder { .. }))
    ));
    assert!(matches!(
        workflow::holder_check(&mut w, "pr", card, B, g1, false),
        Err(WorkflowError::Fence(FenceError::WrongHolder { .. }))
    ));
    let g2 = reservation::handoff(&mut w, card, A, g1, B)
        .unwrap()
        .generation;
    assert!(matches!(
        workflow::link_pr(&mut w, card, A, g1, &open_pr(7)),
        Err(WorkflowError::Fence(FenceError::StaleGeneration { .. }))
    ));
    assert_eq!(one::<i64>(w.conn(), "SELECT count(*) FROM card_link"), 0);
    assert_eq!(fence_rejects(w.conn(), "pr", "wrong_holder"), 2);
    assert_eq!(fence_rejects(w.conn(), "pr", "stale_generation"), 1);

    // The lease is over: the holder can still record what the host says (as with `release`),
    // while a check that asks for a live lease refuses.
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() - 5 WHERE status = 'held'",
    );
    let events: i64 = one(w.conn(), "SELECT count(*) FROM event");
    workflow::holder_check(&mut w, "pr", card, B, g2, false).unwrap();
    assert_eq!(
        one::<i64>(w.conn(), "SELECT count(*) FROM event"),
        events,
        "a passed check writes nothing"
    );
    assert!(matches!(
        workflow::holder_check(&mut w, "pr", card, B, g2, true),
        Err(WorkflowError::Fence(FenceError::LeaseExpired { .. }))
    ));
    workflow::link_pr(&mut w, card, B, g2, &merged(7)).unwrap();
    assert_eq!(
        workflow::link_pr(&mut w, 99, B, g2, &merged(7)),
        Err(WorkflowError::CardNotFound(99))
    );
}

// ---- review and done ---------------------------------------------------------------------------

#[test]
fn review_moves_an_active_card_and_keeps_the_reservation() {
    let (_home, mut w) = setup();
    let (card, g) = taken(&mut w, A, 1);
    assert!(matches!(
        workflow::review(&mut w, card, B, g),
        Err(WorkflowError::Fence(FenceError::WrongHolder { .. }))
    ));
    assert!(matches!(
        workflow::review(&mut w, card, A, g + 1),
        Err(WorkflowError::Fence(FenceError::UnknownGeneration { .. }))
    ));
    assert_eq!(one::<String>(w.conn(), "SELECT state FROM card"), "active");
    let moved = workflow::review(&mut w, card, A, g).unwrap();
    assert_eq!(
        (moved.from.as_str(), moved.state.as_str()),
        ("active", "review")
    );
    assert_eq!(moved.closed_at, None);
    assert_eq!(moved.reservation.status, "held");
    assert_eq!(one::<String>(w.conn(), "SELECT state FROM card"), "review");
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.review'"
        ),
        1
    );
    // Not twice: the card is no longer `active`.
    assert_eq!(
        workflow::review(&mut w, card, A, g),
        Err(WorkflowError::WrongState {
            card_id: card,
            state: "review".to_string(),
            needs: "active",
            action: "sent to review"
        })
    );
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.review'"
        ),
        1
    );
    assert_eq!(fence_rejects(w.conn(), "review", "wrong_holder"), 1);

    // An ended lease cannot send a card to review.
    let (other, g) = taken(&mut w, B, 0);
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() WHERE card_id = 2",
    );
    assert!(matches!(
        workflow::review(&mut w, other, B, g),
        Err(WorkflowError::Fence(FenceError::LeaseExpired { .. }))
    ));
}

fn refused_close(w: &mut WriteLedger, card: i64, g: i64) -> String {
    match workflow::done(w, card, A, g) {
        Err(WorkflowError::NotClosable(why)) => {
            assert_eq!(
                one::<String>(
                    w.conn(),
                    &format!("SELECT state FROM card WHERE card_id = {card}")
                ),
                "review"
            );
            why
        }
        other => panic!("{other:?}"),
    }
}

#[test]
fn done_needs_a_met_criterion_no_open_one_and_a_merged_pull_request() {
    let (_home, mut w) = setup();

    // No criterion at all, even with a merged pull request.
    let (bare, g) = taken(&mut w, A, 0);
    workflow::link_pr(&mut w, bare, A, g, &merged(1)).unwrap();
    workflow::review(&mut w, bare, A, g).unwrap();
    assert_eq!(
        refused_close(&mut w, bare, g),
        "cannot close: the card has no acceptance criterion that is met"
    );

    // Two criteria: the edges one after another on the same card.
    let (card, g) = taken(&mut w, A, 2);
    workflow::review(&mut w, card, A, g).unwrap();
    assert_eq!(
        refused_close(&mut w, card, g),
        "cannot close: the card has no acceptance criterion that is met"
    );
    met(&mut w, card, A, g, 1);
    assert_eq!(
        refused_close(&mut w, card, g),
        "cannot close: an acceptance criterion is still open",
        "one met, one open"
    );
    met(&mut w, card, A, g, 2);
    assert_eq!(
        refused_close(&mut w, card, g),
        "cannot close: no merged pull request is linked",
        "all met, no pull request"
    );
    workflow::link_pr(&mut w, card, A, g, &open_pr(7)).unwrap();
    assert_eq!(
        refused_close(&mut w, card, g),
        "cannot close: no merged pull request is linked",
        "the pull request is open"
    );
    workflow::link_pr(
        &mut w,
        card,
        A,
        g,
        &PrTarget {
            pr_state: "closed",
            ..open_pr(7)
        },
    )
    .unwrap();
    assert_eq!(
        refused_close(&mut w, card, g),
        "cannot close: no merged pull request is linked",
        "the pull request was closed without a merge"
    );
    workflow::link_pr(&mut w, card, A, g, &merged(7)).unwrap();
    sql(
        &mut w,
        "UPDATE card_link SET state = 'retired' WHERE from_card = 2",
    );
    assert_eq!(
        refused_close(&mut w, card, g),
        "cannot close: no merged pull request is linked",
        "a retired link does not count"
    );
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.done'"
        ),
        0
    );

    // All met and merged (the link is live again only because the test says so by hand).
    sql(
        &mut w,
        "UPDATE card_link SET state = 'active' WHERE from_card = 2",
    );
    let closed = workflow::done(&mut w, card, A, g).unwrap();
    assert_eq!(
        (closed.from.as_str(), closed.state.as_str()),
        ("review", "done")
    );
    assert_eq!(closed.reservation.status, "released");
    let (state, closed_at): (String, Option<i64>) = w
        .conn()
        .query_row(
            "SELECT state, closed_at FROM card WHERE card_id = 2",
            [],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .unwrap();
    assert_eq!(state, "done");
    assert!(closed_at.is_some());
    assert_eq!(closed.closed_at, closed_at);
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.done'"
        ),
        1
    );
    assert_eq!(
        reservation::current(w.conn(), card).unwrap(),
        None,
        "a closed card is held by nobody"
    );
    // Closed is closed: the old holder's generation is over.
    assert!(matches!(
        workflow::done(&mut w, card, A, g),
        Err(WorkflowError::Fence(FenceError::NotHeld { .. }))
    ));
    let n = workflow::next(w.conn(), Some(A)).unwrap();
    assert!(n.mine.iter().all(|m| m.reservation.card_id != card));
    assert!(n.ready.is_empty());
}

#[test]
fn a_withdrawn_criterion_does_not_block_done_but_cannot_stand_in_for_a_met_one() {
    let (_home, mut w) = setup();
    let (card, g) = taken(&mut w, A, 2);
    workflow::link_pr(&mut w, card, A, g, &merged(7)).unwrap();
    workflow::review(&mut w, card, A, g).unwrap();
    for seq in [1, 2] {
        workflow::set_criterion(
            &mut w,
            card,
            A,
            g,
            seq,
            CriterionChange::Withdrawn { reason: "r" },
        )
        .unwrap();
    }
    assert_eq!(
        refused_close(&mut w, card, g),
        "cannot close: the card has no acceptance criterion that is met",
        "every criterion withdrawn"
    );
    met(&mut w, card, A, g, 1);
    workflow::done(&mut w, card, A, g).unwrap();
}

#[test]
fn done_starts_from_review_only_and_only_for_the_holder_at_the_current_generation() {
    let (_home, mut w) = setup();
    let (card, g1) = taken(&mut w, A, 1);
    met(&mut w, card, A, g1, 1);
    workflow::link_pr(&mut w, card, A, g1, &merged(7)).unwrap();
    assert_eq!(
        workflow::done(&mut w, card, A, g1),
        Err(WorkflowError::WrongState {
            card_id: card,
            state: "active".to_string(),
            needs: "review",
            action: "closed"
        }),
        "the schema has no active -> done"
    );
    workflow::review(&mut w, card, A, g1).unwrap();
    assert!(matches!(
        workflow::done(&mut w, card, B, g1),
        Err(WorkflowError::Fence(FenceError::WrongHolder { .. }))
    ));
    let g2 = reservation::handoff(&mut w, card, A, g1, B)
        .unwrap()
        .generation;
    assert_eq!(
        workflow::done(&mut w, card, A, g1),
        Err(WorkflowError::Fence(FenceError::StaleGeneration {
            offered: g1,
            current: g2
        }))
    );
    assert_eq!(one::<String>(w.conn(), "SELECT state FROM card"), "review");
    assert_eq!(fence_rejects(w.conn(), "done", "wrong_holder"), 1);
    assert_eq!(fence_rejects(w.conn(), "done", "stale_generation"), 1);

    // The merge came after the lease ended: the holder still closes the card.
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() - 5 WHERE status = 'held'",
    );
    let closed = workflow::done(&mut w, card, B, g2).unwrap();
    assert_eq!(closed.state, "done");
    assert_eq!(closed.reservation.holder, B);
}

// ---- write boundary ----------------------------------------------------------------------------

#[test]
fn a_pattern_covers_itself_and_what_lies_below_it_never_what_lies_above_or_beside() {
    let inside = [
        ("src/card/", "src/card/work.rs"),
        ("src/card", "src/card/work.rs"),
        ("src/card", "src/card"),
        ("src/card/work.rs", "src/card/work.rs"),
        ("src/card/", "SRC/Card/Work.rs"),
        ("src/card/", "src/./card/x/../work.rs"),
        ("src/**", "src/a/b/c.rs"),
        ("src/**", "src"),
        ("src/**/mod.rs", "src/mod.rs"),
        ("src/**/mod.rs", "src/a/b/mod.rs"),
        ("src/*.rs", "src/main.rs"),
        ("src/*/mod.rs", "src/card/mod.rs"),
        ("tests/card_*", "tests/card_work_cli.rs"),
        ("**", "anything/at/all"),
    ];
    for (pattern, path) in inside {
        assert!(workflow::covers(pattern, path), "{pattern} covers {path}");
        assert!(
            scope::overlap(pattern, path),
            "{pattern} / {path}: a covered path always overlaps its pattern"
        );
    }
    let outside = [
        ("src/card/", "src/ledger/workflow.rs"),
        ("src/card/", "src/card_extra/x.rs"),
        ("src/card/", "src"),
        ("src/card/work.rs", "src/card"),
        ("src/card/work.rs", "src/card/work.rs.bak"),
        ("src/*.rs", "src/card/work.rs"),
        ("src/*.rs", "src/main.txt"),
        ("src/**/mod.rs", "src/a/lib.rs"),
        ("src/card/", "src/card/../ledger/x.rs"),
        ("src/card/", "../outside"),
        ("src/card/", "../src/card/x"),
        ("e/", "q/../../e/x"),
        ("src/card/", "src/../../src/card/x"),
        ("src/card/", "src/card/../../../etc/passwd"),
        ("../up/", "anything"),
        ("src/card/", ""),
        ("src/card/", "."),
        ("", "src/card/work.rs"),
    ];
    for (pattern, path) in outside {
        assert!(
            !workflow::covers(pattern, path),
            "{pattern} does not cover {path}"
        );
    }
}

#[test]
fn the_first_covering_pattern_of_a_reservation_is_named() {
    let (_home, mut w) = setup();
    let card = add_card(&mut w, "ready", "p2", "[]", 0);
    let scope = ["src/card/".to_string(), "tests/card_*".to_string()];
    let r = reservation::reserve(&mut w, card, A, &scope, 3600, None, None).unwrap();
    assert_eq!(
        workflow::covering_pattern(&r, "src/card/work.rs"),
        Some("src/card/")
    );
    assert_eq!(
        workflow::covering_pattern(&r, "tests/card_work_cli.rs"),
        Some("tests/card_*")
    );
    assert_eq!(workflow::covering_pattern(&r, "src/main.rs"), None);
}

#[test]
fn every_refusal_reads_as_the_kind_the_envelope_needs() {
    use moai_cli::contract::{classify, Kind};
    let cases = [
        (
            WorkflowError::InvalidInput("x".into()),
            "invalid_input",
            Kind::InvalidInput,
        ),
        (
            WorkflowError::CardNotFound(3),
            "card_not_found",
            Kind::NotFound,
        ),
        (
            WorkflowError::Fence(FenceError::WrongHolder {
                held_by: "b".into(),
            }),
            "wrong_holder",
            Kind::Conflict,
        ),
        (
            WorkflowError::CriterionNotFound { card_id: 3, seq: 9 },
            "criterion_not_found",
            Kind::NotFound,
        ),
        (
            WorkflowError::DuplicateRef {
                card_id: 3,
                reference: "r".into(),
            },
            "duplicate_ref",
            Kind::Conflict,
        ),
        (
            WorkflowError::WrongState {
                card_id: 3,
                state: "active".into(),
                needs: "review",
                action: "closed",
            },
            "wrong_state",
            Kind::Conflict,
        ),
        (
            WorkflowError::NotClosable(
                "cannot close: an acceptance criterion is still open".into(),
            ),
            "not_closable",
            Kind::InvalidInput,
        ),
        (
            WorkflowError::LinkRetired {
                card_id: 3,
                target: "github:o/r#7".into(),
            },
            "link_retired",
            Kind::Conflict,
        ),
    ];
    for (error, code, kind) in cases {
        assert_eq!(error.code(), code);
        assert_eq!(classify(&error.to_string()), kind, "{error}");
    }
}

// ---- repair of the first audit (N1, N2, N9) ------------------------------------------------------

#[test]
fn every_note_event_carries_the_session_that_wrote_it_whatever_the_author_says() {
    let (_home, mut w) = setup();
    let card = add_card(&mut w, "draft", "p2", "[]", 0);
    workflow::add_note(&mut w, card, "decision", "leader", Some(B), "go left", None).unwrap();
    workflow::add_note(&mut w, card, "comment", A, Some(A), "seen", None).unwrap();
    workflow::add_note(&mut w, card, "comment", "user", None, "by hand", None).unwrap();
    let sessions: Vec<Option<String>> = w
        .conn()
        .prepare(
            "SELECT json_extract(payload, '$.session') FROM event WHERE kind = 'card.note' \
             ORDER BY event_id",
        )
        .unwrap()
        .query_map([], |r| r.get(0))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    assert_eq!(
        sessions,
        [Some(B.to_string()), Some(A.to_string()), None],
        "the author is what was claimed; the session is who wrote"
    );
    assert_eq!(
        one::<String>(
            w.conn(),
            "SELECT json_type(payload, '$.session') FROM event WHERE kind = 'card.note' \
             ORDER BY event_id DESC LIMIT 1"
        ),
        "null",
        "no session is written as null, not left out"
    );
    assert_eq!(
        one::<String>(w.conn(), "SELECT author FROM card_note WHERE note_id = 1"),
        "leader"
    );
    // A session id has the shape of one.
    for bad in ["", " ", "two\nlines"] {
        assert!(matches!(
            workflow::add_note(&mut w, card, "comment", "user", Some(bad), "x", None),
            Err(WorkflowError::InvalidInput(_))
        ));
    }
    assert_eq!(one::<i64>(w.conn(), "SELECT count(*) FROM card_note"), 3);
}

#[test]
fn a_note_keeps_tabs_and_line_breaks_and_refuses_every_other_control_character() {
    let (_home, mut w) = setup();
    let card = add_card(&mut w, "draft", "p2", "[]", 0);
    workflow::add_note(&mut w, card, "comment", "user", None, "a\tb\nc", None).unwrap();
    for bad in [
        "a\u{1b}[31mred",
        "a\rb",
        "bell\u{7}",
        "a\u{7f}b",
        "a\u{9b}b",
        "a\0b",
    ] {
        match workflow::add_note(&mut w, card, "comment", "user", None, bad, None) {
            Err(WorkflowError::InvalidInput(why)) => {
                assert!(why.contains("control") || why.contains("NUL"), "{why}")
            }
            other => panic!("{bad:?}: {other:?}"),
        }
    }
    assert_eq!(one::<i64>(w.conn(), "SELECT count(*) FROM card_note"), 1);
}

#[test]
fn reopen_takes_a_review_card_back_to_active_for_its_holder() {
    let (_home, mut w) = setup();
    let (card, g) = taken(&mut w, A, 1);
    assert_eq!(
        workflow::reopen(&mut w, card, A, g),
        Err(WorkflowError::WrongState {
            card_id: card,
            state: "active".to_string(),
            needs: "review",
            action: "reopened"
        })
    );
    workflow::review(&mut w, card, A, g).unwrap();
    assert!(matches!(
        workflow::reopen(&mut w, card, B, g),
        Err(WorkflowError::Fence(FenceError::WrongHolder { .. }))
    ));
    assert!(matches!(
        workflow::reopen(&mut w, card, A, g + 1),
        Err(WorkflowError::Fence(FenceError::UnknownGeneration { .. }))
    ));
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() WHERE card_id = 1",
    );
    assert!(matches!(
        workflow::reopen(&mut w, card, A, g),
        Err(WorkflowError::Fence(FenceError::LeaseExpired { .. }))
    ));
    assert_eq!(one::<String>(w.conn(), "SELECT state FROM card"), "review");
    assert_eq!(fence_rejects(w.conn(), "reopen", "wrong_holder"), 1);
    assert_eq!(fence_rejects(w.conn(), "reopen", "lease_expired"), 1);

    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() + 600 WHERE card_id = 1",
    );
    let moved = workflow::reopen(&mut w, card, A, g).unwrap();
    assert_eq!(
        (moved.from.as_str(), moved.state.as_str()),
        ("review", "active")
    );
    assert_eq!(moved.reservation.status, "held");
    assert_eq!(moved.reservation.generation, g, "the same generation");
    assert_eq!(moved.closed_at, None);
    assert_eq!(one::<String>(w.conn(), "SELECT state FROM card"), "active");
    assert_eq!(
        one::<i64>(
            w.conn(),
            "SELECT count(*) FROM event WHERE kind = 'card.reopen'"
        ),
        1
    );
    // The repaired card goes to review again and closes as usual.
    met(&mut w, card, A, g, 1);
    workflow::link_pr(&mut w, card, A, g, &merged(7)).unwrap();
    workflow::review(&mut w, card, A, g).unwrap();
    workflow::done(&mut w, card, A, g).unwrap();
}

#[test]
fn next_lists_a_review_card_nobody_holds_under_its_own_reason() {
    let (_home, mut w) = setup();
    let (card, g) = taken(&mut w, A, 0);
    workflow::review(&mut w, card, A, g).unwrap();
    let n = workflow::next(w.conn(), Some(B)).unwrap();
    assert_eq!(n.skipped[0].reason.code(), "held", "held while in review");

    // `release` refuses a review card (card C-68 N-a), so a ledger from an older build or one
    // edited by hand is the only way to get a review card with no holder.
    assert!(matches!(
        reservation::release(&mut w, card, A, g),
        Err(FenceError::CardInReview { .. })
    ));
    sql(
        &mut w,
        &format!("UPDATE reservation SET status = 'released' WHERE card_id = {card}"),
    );
    for holder in [Some(A), Some(B), None] {
        let n = workflow::next(w.conn(), holder).unwrap();
        assert!(n.mine.is_empty() && n.ready.is_empty());
        assert_eq!(n.skipped.len(), 1);
        assert_eq!(n.skipped[0].card_id, card);
        assert_eq!(n.skipped[0].state, "review");
        assert_eq!(n.skipped[0].reason, SkipReason::ReviewNoHolder);
        assert_eq!(n.skipped[0].reason.code(), "review_no_holder");
    }
}

/// SPEC-20261119-criterion-add-session REQ-5, REQ-6 (card C-125, C-108 audit N3): the total is the
/// number of rows the card has, not its highest number. Rows 1 and 5 make a gap (the table takes
/// any `seq`) and row 5 is withdrawn, so the count includes a withdrawn row and the new numbers
/// still go after it (REQ-5, REQ-6; C-108 audit N5); two adds are numbered after 5 and the total
/// is 4.
#[test]
fn criteria_total_is_the_row_count() {
    let (_home, mut w) = setup();
    let card = add_card(&mut w, "ready", "p2", "[]", 1);
    sql(
        &mut w,
        &format!(
            "INSERT INTO card_ac (card_id, seq, text, status, withdrawn_reason, updated_at) \
             VALUES ({card}, 5, 'five', 'withdrawn', 'gone', 1)"
        ),
    );
    let texts = vec!["a".to_string(), "b".to_string()];
    let added = workflow::add_criteria(&mut w, card, "leader", &texts).unwrap();
    assert_eq!(added.added, vec![6, 7], "numbered after the highest");
    assert_eq!(added.total, 4, "the row count, not the highest number");
    assert_eq!(
        one::<i64>(
            w.conn(),
            &format!("SELECT count(*) FROM card_ac WHERE card_id = {card}")
        ),
        4
    );
    // One more add on the same card: still the row count (rows 1, 5, 6, 7 and the new one).
    let again = workflow::add_criteria(&mut w, card, "leader", &["c".to_string()]).unwrap();
    assert_eq!((again.added, again.total), (vec![8], 5));
}

/// REQ-1, REQ-2, REQ-3 (card C-125, C-108 audit N1): the event records the adding session next to
/// the author; `null` when the caller has no session id; the session is one line within the
/// author limit and a refusal writes nothing.
#[test]
fn criteria_event_records_the_adding_session() {
    let (_home, mut w) = setup();
    let card = add_card(&mut w, "ready", "p2", "[]", 0);
    let session = "8659f17c-8040-48ff-a194-69b556239ea0";
    let one_text = vec!["a".to_string()];
    workflow::add_criteria_as(&mut w, card, "leader", Some(session), &one_text).unwrap();
    workflow::add_criteria_as(&mut w, card, "someone-else", None, &one_text).unwrap();
    let rows: Vec<(String, String)> = w
        .conn()
        .prepare(
            "SELECT actor, payload FROM event WHERE kind = 'card.criterion_add' ORDER BY event_id",
        )
        .unwrap()
        .query_map([], |r| Ok((r.get(0)?, r.get(1)?)))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    assert_eq!(rows.len(), 2, "{rows:?}");
    let payload = |i: usize| serde_json::from_str::<serde_json::Value>(&rows[i].1).unwrap();
    // The author is what `--by` said; the session is its own field.
    assert_eq!(rows[0].0, "leader");
    assert_eq!(
        payload(0),
        serde_json::json!({"first": 1, "count": 1, "session": session})
    );
    assert_eq!(rows[1].0, "someone-else");
    assert_eq!(
        payload(1),
        serde_json::json!({"first": 2, "count": 1, "session": null})
    );
    // A session that is not one line, or longer than the author limit, is refused before any write.
    let events: i64 = one(w.conn(), "SELECT count(*) FROM event");
    let rows_before: i64 = one(w.conn(), "SELECT count(*) FROM card_ac");
    for bad in [
        "two\nlines".to_string(),
        "x".repeat(workflow::MAX_AUTHOR_CHARS + 1),
    ] {
        assert!(matches!(
            workflow::add_criteria_as(&mut w, card, "leader", Some(&bad), &one_text),
            Err(WorkflowError::InvalidInput(_))
        ));
    }
    assert_eq!(one::<i64>(w.conn(), "SELECT count(*) FROM event"), events);
    assert_eq!(
        one::<i64>(w.conn(), "SELECT count(*) FROM card_ac"),
        rows_before
    );
    // The limit itself passes (N and N+1).
    let at_limit = "y".repeat(workflow::MAX_AUTHOR_CHARS);
    workflow::add_criteria_as(&mut w, card, "leader", Some(&at_limit), &one_text).unwrap();
}

#[test]
fn criteria_are_added_to_any_card_but_a_closed_one_and_each_gets_its_own_number() {
    let (_home, mut w) = setup();
    // A card nobody holds, with one criterion (number 1) and one withdrawn (number 2).
    let card = add_card(&mut w, "ready", "p2", "[]", 2);
    sql(
        &mut w,
        &format!(
            "UPDATE card_ac SET status = 'withdrawn', withdrawn_reason = 'x' \
             WHERE card_id = {card} AND seq = 2"
        ),
    );
    let texts = vec!["a".to_string(), "b".to_string()];
    let added = workflow::add_criteria(&mut w, card, "leader", &texts).unwrap();
    assert_eq!(added.added, vec![3, 4]);
    assert_eq!(added.total, 4);
    let conn = w.conn();
    let rows: Vec<(i64, String, String)> = conn
        .prepare("SELECT seq, text, status FROM card_ac WHERE card_id = ?1 ORDER BY seq")
        .unwrap()
        .query_map([card], |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?)))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    assert_eq!(
        rows[2..],
        [
            (3, "a".to_string(), "open".to_string()),
            (4, "b".to_string(), "open".to_string())
        ]
    );
    let event: (String, String) = conn
        .query_row(
            "SELECT actor, payload FROM event WHERE kind = 'card.criterion_add'",
            [],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .unwrap();
    assert_eq!(event.0, "leader");
    assert_eq!(
        serde_json::from_str::<serde_json::Value>(&event.1).unwrap(),
        serde_json::json!({"first": 3, "count": 2, "session": null})
    );

    // Nothing to add, a missing card and a closed card are refused and write nothing.
    let events: i64 = one(w.conn(), "SELECT count(*) FROM event");
    assert_eq!(
        workflow::add_criteria(&mut w, card, "leader", &[]).unwrap_err(),
        WorkflowError::InvalidInput("there is no criterion to add".into())
    );
    assert_eq!(
        workflow::add_criteria(&mut w, 999, "leader", &texts).unwrap_err(),
        WorkflowError::CardNotFound(999)
    );
    sql(
        &mut w,
        "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
         VALUES ('closed', 'done', 'test', 1, 1, 1)",
    );
    let closed: i64 = one(w.conn(), "SELECT card_id FROM card WHERE state = 'done'");
    assert_eq!(
        workflow::add_criteria(&mut w, closed, "leader", &texts).unwrap_err(),
        WorkflowError::Closed { card_id: closed }
    );
    assert_eq!(one::<i64>(w.conn(), "SELECT count(*) FROM event"), events);
    assert_eq!(
        one::<i64>(
            w.conn(),
            &format!("SELECT count(*) FROM card_ac WHERE card_id = {closed}")
        ),
        0
    );
    // The author is one line, within the limit.
    assert!(matches!(
        workflow::add_criteria(&mut w, card, " ", &texts),
        Err(WorkflowError::InvalidInput(_))
    ));
}

// ---- card withdraw: a withdrawn card cannot be taken (no schema change) ----

fn state_of(w: &WriteLedger, card: i64) -> String {
    w.conn()
        .query_row("SELECT state FROM card WHERE card_id = ?1", [card], |r| {
            r.get(0)
        })
        .unwrap()
}

fn withdraw_events(w: &WriteLedger, card: i64) -> i64 {
    w.conn()
        .query_row(
            "SELECT COUNT(*) FROM event WHERE card_id = ?1 AND kind = 'card.withdraw'",
            [card],
            |r| r.get(0),
        )
        .unwrap()
}

#[test]
fn withdraw_moves_a_draft_or_ready_card_to_hold_and_records_the_reason() {
    let (_home, mut w) = setup();
    let draft = add_card(&mut w, "draft", "p2", r#"["demo/"]"#, 0);
    workflow::withdraw(
        &mut w,
        draft,
        "superseded by another card",
        "leader",
        Some(A),
    )
    .unwrap();
    assert_eq!(state_of(&w, draft), "hold");
    let (reason, by): (String, String) = w
        .conn()
        .query_row(
            "SELECT json_extract(payload, '$.reason'), json_extract(payload, '$.by') \
             FROM event WHERE card_id = ?1 AND kind = 'card.withdraw'",
            [draft],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .unwrap();
    assert_eq!(reason, "superseded by another card");
    assert_eq!(by, "leader");

    let ready = add_card(&mut w, "ready", "p1", r#"["other/"]"#, 0);
    workflow::withdraw(&mut w, ready, "no longer needed", "leader", Some(B)).unwrap();
    assert_eq!(state_of(&w, ready), "hold");
}

#[test]
fn withdraw_refuses_a_live_reservation_another_state_and_an_unknown_card() {
    let (_home, mut w) = setup();
    let (held, _g) = taken(&mut w, A, 1);
    assert!(matches!(
        workflow::withdraw(&mut w, held, "gone", "leader", Some(A)),
        Err(workflow::WithdrawError::Held { .. })
    ));
    for state in ["hold", "review"] {
        let card = add_card(&mut w, state, "p2", r#"["x/"]"#, 0);
        assert!(matches!(
            workflow::withdraw(&mut w, card, "gone", "leader", Some(A)),
            Err(workflow::WithdrawError::WrongState { .. })
        ));
    }
    // `done` cannot be withdrawn: close one through the real path.
    let (done, g) = taken(&mut w, A, 1);
    met(&mut w, done, A, g, 1);
    workflow::link_pr(&mut w, done, A, g, &merged(1)).unwrap();
    workflow::review(&mut w, done, A, g).unwrap();
    workflow::done(&mut w, done, A, g).unwrap();
    assert!(matches!(
        workflow::withdraw(&mut w, done, "gone", "leader", Some(A)),
        Err(workflow::WithdrawError::WrongState { .. })
    ));
    assert!(matches!(
        workflow::withdraw(&mut w, 99_999, "gone", "leader", Some(A)),
        Err(workflow::WithdrawError::CardNotFound(99_999))
    ));
    // Every refusal changed nothing: no withdraw event anywhere, the held card is still active.
    assert_eq!(withdraw_events(&w, held), 0);
    for state in ["hold", "review"] {
        let card = add_card(&mut w, state, "p2", r#"["x/"]"#, 0);
        assert_eq!(withdraw_events(&w, card), 0);
    }
    assert_eq!(
        withdraw_events(&w, done),
        0,
        "the refused withdraw wrote nothing"
    );
    let still: String = w
        .conn()
        .query_row("SELECT state FROM card WHERE card_id = ?1", [held], |r| {
            r.get(0)
        })
        .unwrap();
    assert_eq!(still, "active");
}

#[test]
fn withdraw_input_limits_accept_n_and_refuse_n_plus_one() {
    let (_home, mut w) = setup();
    let card = add_card(&mut w, "ready", "p2", r#"["demo/"]"#, 0);
    let reason_500 = "r".repeat(MAX_LINE_CHARS);
    let reason_501 = "r".repeat(MAX_LINE_CHARS + 1);
    let by_200 = "b".repeat(200);
    let by_201 = "b".repeat(201);
    workflow::withdraw(&mut w, card, &reason_500, &by_200, Some(A)).unwrap();
    // The card is hold now, but input bounds are checked first: each bad input is still an
    // InvalidInput, not a WrongState.
    assert!(matches!(
        workflow::withdraw(&mut w, card, &reason_501, "leader", Some(A)),
        Err(workflow::WithdrawError::InvalidInput(_))
    ));
    assert!(matches!(
        workflow::withdraw(&mut w, card, "gone", &by_201, Some(A)),
        Err(workflow::WithdrawError::InvalidInput(_))
    ));
    assert!(matches!(
        workflow::withdraw(&mut w, card, "", "leader", Some(A)),
        Err(workflow::WithdrawError::InvalidInput(_))
    ));
    assert!(matches!(
        workflow::withdraw(&mut w, card, "gone\ta tab", "leader", Some(A)),
        Err(workflow::WithdrawError::InvalidInput(_))
    ));
}
