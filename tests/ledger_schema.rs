//! Ledger schema version 1 (Issue #3068 P1, AC10; contract comment 6896940106 and its amendment
//! 1, comment 6897392517): every CHECK, UNIQUE, partial index and trigger at its edge, in the
//! direction the schema lets through and the direction it refuses. Each test builds its own
//! ledger in its own temporary HOME through the library; none touches the real HOME.

mod common;

use common::home::TempHome;
use moai_cli::ledger::{self, LedgerError, WriteLedger};

struct Fx {
    _home: TempHome,
    w: WriteLedger,
}

fn fx() -> Fx {
    let home = TempHome::new();
    ledger::register(home.path(), "demo").expect("register");
    let w = ledger::open_write(home.path(), "demo").expect("open for writing");
    Fx { _home: home, w }
}

/// Runs `sql` (one or more statements) in one write transaction.
fn exec(w: &mut WriteLedger, sql: &str) -> Result<(), LedgerError> {
    w.with_write_tx(|tx, _| {
        tx.execute_batch(sql)?;
        Ok(())
    })
}

/// Runs one `INSERT` and returns its row id.
fn insert(w: &mut WriteLedger, sql: &str) -> Result<i64, LedgerError> {
    w.with_write_tx(|tx, _| {
        tx.execute(sql, [])?;
        Ok(tx.last_insert_rowid())
    })
}

fn count(w: &WriteLedger, sql: &str) -> i64 {
    w.conn().query_row(sql, [], |r| r.get(0)).unwrap()
}

fn text(w: &WriteLedger, sql: &str) -> String {
    w.conn().query_row(sql, [], |r| r.get(0)).unwrap()
}

/// The write was refused by the schema; returns the message.
fn refused<T: std::fmt::Debug>(r: Result<T, LedgerError>) -> String {
    match r {
        Err(LedgerError::Constraint(m)) => m,
        other => panic!("expected the schema to refuse, got {other:?}"),
    }
}

fn refused_with<T: std::fmt::Debug>(r: Result<T, LedgerError>, needle: &str) {
    let m = refused(r);
    assert!(m.contains(needle), "{needle:?} not in {m:?}");
}

const NEW_CARD: &str =
    "INSERT INTO card (title, created_by, created_at, updated_at) VALUES ('t', 'test', 1, 1)";

fn card_in_state(w: &mut WriteLedger, state: &str) -> i64 {
    let closed = if state == "done" { "5" } else { "NULL" };
    insert(
        w,
        &format!(
            "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
             VALUES ('t', '{state}', 'test', 1, 1, {closed})"
        ),
    )
    .unwrap()
}

fn ac(w: &mut WriteLedger, card: i64, seq: i64, status: &str) -> Result<i64, LedgerError> {
    let (evidence, reason) = match status {
        "met" => ("'pr:x#1'", "NULL"),
        "withdrawn" => ("NULL", "'dropped'"),
        _ => ("NULL", "NULL"),
    };
    insert(
        w,
        &format!(
            "INSERT INTO card_ac (card_id, seq, text, status, evidence_ref, withdrawn_reason, updated_at) \
             VALUES ({card}, {seq}, 'criterion', '{status}', {evidence}, {reason}, 1)"
        ),
    )
}

fn merged_pr(w: &mut WriteLedger, card: i64) -> i64 {
    insert(
        w,
        &format!(
            "INSERT INTO card_link (from_card, kind, target_host, target_project, target_number, \
             pr_state, merge_ref, created_at, updated_at) \
             VALUES ({card}, 'pr', 'github', 'acme/repo', 7, 'merged', 'abc123', 1, 1)"
        ),
    )
    .unwrap()
}

fn close(w: &mut WriteLedger, card: i64) -> Result<(), LedgerError> {
    exec(
        w,
        &format!("UPDATE card SET state = 'done', closed_at = 9 WHERE card_id = {card}"),
    )
}

// ---- identity -------------------------------------------------------------------------------

#[test]
fn the_file_carries_its_identity_and_the_meta_rows() {
    let f = fx();
    assert_eq!(count(&f.w, "PRAGMA application_id"), 0x4D4F_4149);
    assert_eq!(count(&f.w, "PRAGMA user_version"), 1);
    assert_eq!(text(&f.w, "PRAGMA journal_mode"), "delete");
    assert_eq!(count(&f.w, "PRAGMA foreign_keys"), 1);
    assert_eq!(
        text(&f.w, "SELECT value FROM meta WHERE key = 'schema_version'"),
        "1"
    );
    assert_eq!(
        text(&f.w, "SELECT value FROM meta WHERE key = 'project_id'"),
        "demo"
    );
    assert_eq!(
        text(&f.w, "SELECT value FROM meta WHERE key = 'created_by'"),
        moai_cli::BUILD_VERSION
    );
    assert!(
        text(&f.w, "SELECT value FROM meta WHERE key = 'created_at'")
            .parse::<i64>()
            .unwrap()
            > 1_000_000_000
    );
    assert_eq!(count(&f.w, "SELECT count(*) FROM meta"), 4);
}

#[test]
fn every_table_of_the_contract_exists_and_is_strict() {
    let f = fx();
    for table in [
        "meta",
        "card",
        "card_ac",
        "card_note",
        "card_link",
        "card_transition",
        "reservation",
        "reservation_scope",
        "event",
    ] {
        let strict = count(
            &f.w,
            &format!("SELECT strict FROM pragma_table_list WHERE name = '{table}'"),
        );
        assert_eq!(strict, 1, "{table} is STRICT");
    }
    // The internal tables are not part of P1.
    for table in [
        "fence_log",
        "slot_hold",
        "reclaim",
        "recovery_run",
        "recovery_item",
    ] {
        assert_eq!(
            count(
                &f.w,
                &format!("SELECT count(*) FROM sqlite_master WHERE name = '{table}'")
            ),
            0,
            "{table}"
        );
    }
}

#[test]
fn a_strict_column_refuses_text_in_an_integer_column() {
    let mut f = fx();
    refused(exec(
        &mut f.w,
        "INSERT INTO card (title, created_by, created_at, updated_at) VALUES ('t', 'x', 'soon', 1)",
    ));
}

#[test]
fn foreign_keys_are_on_for_every_connection() {
    let mut f = fx();
    refused_with(ac(&mut f.w, 999, 1, "open"), "FOREIGN KEY");
    refused_with(
        exec(
            &mut f.w,
            "INSERT INTO reservation_scope (reservation_id, pattern) VALUES (999, 'src/')",
        ),
        "FOREIGN KEY",
    );
}

// ---- card -----------------------------------------------------------------------------------

#[test]
fn card_state_kind_and_priority_accept_their_values_and_refuse_others() {
    let mut f = fx();
    for state in ["draft", "ready", "active", "review", "done", "hold"] {
        card_in_state(&mut f.w, state);
    }
    for kind in [
        "task", "feature", "fix", "docs", "research", "audit", "chore",
    ] {
        insert(
            &mut f.w,
            &format!(
                "INSERT INTO card (title, kind, created_by, created_at, updated_at) VALUES ('t', '{kind}', 'x', 1, 1)"
            ),
        )
        .unwrap();
    }
    for priority in ["p0", "p1", "p2", "p3"] {
        insert(
            &mut f.w,
            &format!(
                "INSERT INTO card (title, priority, created_by, created_at, updated_at) VALUES ('t', '{priority}', 'x', 1, 1)"
            ),
        )
        .unwrap();
    }
    assert_eq!(count(&f.w, "SELECT count(*) FROM card"), 17);
    for (col, bad) in [("state", "open"), ("kind", "bug"), ("priority", "p4")] {
        refused_with(
            exec(
                &mut f.w,
                &format!(
                    "INSERT INTO card (title, {col}, created_by, created_at, updated_at) VALUES ('t', '{bad}', 'x', 1, 1)"
                ),
            ),
            "CHECK constraint failed",
        );
    }
    assert_eq!(count(&f.w, "SELECT count(*) FROM card"), 17);
}

#[test]
fn a_title_is_one_to_three_hundred_characters() {
    let mut f = fx();
    for len in [1usize, 300] {
        insert(
            &mut f.w,
            &format!(
                "INSERT INTO card (title, created_by, created_at, updated_at) VALUES ('{}', 'x', 1, 1)",
                "a".repeat(len)
            ),
        )
        .unwrap();
    }
    for len in [0usize, 301] {
        refused(exec(
            &mut f.w,
            &format!(
                "INSERT INTO card (title, created_by, created_at, updated_at) VALUES ('{}', 'x', 1, 1)",
                "a".repeat(len)
            ),
        ));
    }
}

#[test]
fn the_write_scope_is_a_json_array() {
    let mut f = fx();
    for ok in ["[]", "[\"src/\"]"] {
        insert(
            &mut f.w,
            &format!(
                "INSERT INTO card (title, write_scope, created_by, created_at, updated_at) VALUES ('t', '{ok}', 'x', 1, 1)"
            ),
        )
        .unwrap();
    }
    for bad in ["{}", "\"src/\"", "not json", "[1,"] {
        refused(exec(
            &mut f.w,
            &format!(
                "INSERT INTO card (title, write_scope, created_by, created_at, updated_at) VALUES ('t', '{bad}', 'x', 1, 1)"
            ),
        ));
    }
}

fn card_with_source(host: &str, project: &str, number: &str) -> String {
    format!(
        "INSERT INTO card (title, created_by, created_at, updated_at, source_host, source_project, source_number) \
         VALUES ('t', 'x', 1, 1, {host}, {project}, {number})"
    )
}

#[test]
fn the_source_is_all_null_or_all_set() {
    let mut f = fx();
    insert(&mut f.w, &card_with_source("NULL", "NULL", "NULL")).unwrap();
    insert(&mut f.w, &card_with_source("'github'", "'acme/repo'", "5")).unwrap();
    // Each of the six partial combinations is refused.
    for (h, p, n) in [
        ("'github'", "NULL", "NULL"),
        ("NULL", "'acme/repo'", "NULL"),
        ("NULL", "NULL", "5"),
        ("'github'", "'acme/repo'", "NULL"),
        ("'github'", "NULL", "5"),
        ("NULL", "'acme/repo'", "5"),
    ] {
        refused_with(
            exec(&mut f.w, &card_with_source(h, p, n)),
            "CHECK constraint failed",
        );
    }
    assert_eq!(count(&f.w, "SELECT count(*) FROM card"), 2);
}

#[test]
fn the_source_host_number_and_project_shape_have_edges() {
    let mut f = fx();
    refused(exec(
        &mut f.w,
        &card_with_source("'bitbucket'", "'acme/repo'", "5"),
    ));
    insert(&mut f.w, &card_with_source("'gitlab'", "'g/sub/p'", "1")).unwrap();
    refused(exec(&mut f.w, &card_with_source("'github'", "'a/b'", "0")));
    refused(exec(&mut f.w, &card_with_source("'github'", "'a/b'", "-1")));
    // Amendment 1: the project is normalised by the connector and checked here.
    for bad in [
        "''",
        "'Acme/Repo'",
        "'acme/Repo'",
        "'acme /repo'",
        "'acme/repo '",
    ] {
        refused_with(
            exec(&mut f.w, &card_with_source("'github'", bad, "9")),
            "CHECK constraint failed",
        );
    }
}

#[test]
fn two_cards_without_a_source_are_allowed_and_the_same_source_twice_is_not() {
    let mut f = fx();
    insert(&mut f.w, NEW_CARD).unwrap();
    insert(&mut f.w, NEW_CARD).unwrap();
    insert(&mut f.w, &card_with_source("'github'", "'acme/repo'", "5")).unwrap();
    refused_with(
        exec(&mut f.w, &card_with_source("'github'", "'acme/repo'", "5")),
        "UNIQUE constraint failed",
    );
    // A different number, project or host is another source.
    insert(&mut f.w, &card_with_source("'github'", "'acme/repo'", "6")).unwrap();
    insert(&mut f.w, &card_with_source("'github'", "'acme/other'", "5")).unwrap();
    insert(&mut f.w, &card_with_source("'gitlab'", "'acme/repo'", "5")).unwrap();
    assert_eq!(count(&f.w, "SELECT count(*) FROM card"), 6);
}

#[test]
fn a_source_that_differs_only_in_case_cannot_slip_past_the_unique_key() {
    let mut f = fx();
    insert(&mut f.w, &card_with_source("'github'", "'acme/repo'", "5")).unwrap();
    refused_with(
        exec(&mut f.w, &card_with_source("'github'", "'ACME/REPO'", "5")),
        "CHECK constraint failed",
    );
}

#[test]
fn a_card_is_done_exactly_when_it_has_a_closing_time() {
    let mut f = fx();
    insert(
        &mut f.w,
        "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) VALUES ('t', 'done', 'x', 1, 1, 5)",
    )
    .unwrap();
    insert(&mut f.w, NEW_CARD).unwrap();
    refused_with(
        exec(
            &mut f.w,
            "INSERT INTO card (title, state, created_by, created_at, updated_at) VALUES ('t', 'done', 'x', 1, 1)",
        ),
        "CHECK constraint failed",
    );
    refused_with(
        exec(
            &mut f.w,
            "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) VALUES ('t', 'draft', 'x', 1, 1, 5)",
        ),
        "CHECK constraint failed",
    );
}

// ---- card_ac --------------------------------------------------------------------------------

#[test]
fn a_met_criterion_needs_evidence_and_a_withdrawn_one_needs_a_reason() {
    let mut f = fx();
    let c = insert(&mut f.w, NEW_CARD).unwrap();
    ac(&mut f.w, c, 1, "met").unwrap();
    ac(&mut f.w, c, 2, "withdrawn").unwrap();
    ac(&mut f.w, c, 3, "open").unwrap();
    refused_with(
        exec(
            &mut f.w,
            &format!("INSERT INTO card_ac (card_id, seq, text, status, updated_at) VALUES ({c}, 4, 'x', 'met', 1)"),
        ),
        "CHECK constraint failed",
    );
    refused_with(
        exec(
            &mut f.w,
            &format!("INSERT INTO card_ac (card_id, seq, text, status, updated_at) VALUES ({c}, 4, 'x', 'withdrawn', 1)"),
        ),
        "CHECK constraint failed",
    );
    refused(exec(
        &mut f.w,
        &format!("INSERT INTO card_ac (card_id, seq, text, status, updated_at) VALUES ({c}, 4, 'x', 'closed', 1)"),
    ));
    // Changing an open criterion to met without evidence is refused the same way.
    refused_with(
        exec(&mut f.w, "UPDATE card_ac SET status = 'met' WHERE seq = 3"),
        "CHECK constraint failed",
    );
}

#[test]
fn a_criterion_number_starts_at_one_and_is_unique_per_card() {
    let mut f = fx();
    let a = insert(&mut f.w, NEW_CARD).unwrap();
    let b = insert(&mut f.w, NEW_CARD).unwrap();
    ac(&mut f.w, a, 1, "open").unwrap();
    ac(&mut f.w, b, 1, "open").unwrap();
    refused_with(ac(&mut f.w, a, 1, "open"), "UNIQUE constraint failed");
    refused_with(ac(&mut f.w, a, 0, "open"), "CHECK constraint failed");
    refused(exec(
        &mut f.w,
        &format!("INSERT INTO card_ac (card_id, seq, text, updated_at) VALUES ({a}, 2, '', 1)"),
    ));
}

#[test]
fn a_criterion_keeps_its_card_and_number_but_may_change_otherwise() {
    let mut f = fx();
    let a = insert(&mut f.w, NEW_CARD).unwrap();
    let b = insert(&mut f.w, NEW_CARD).unwrap();
    ac(&mut f.w, a, 1, "open").unwrap();
    refused_with(
        exec(
            &mut f.w,
            &format!("UPDATE card_ac SET card_id = {b} WHERE seq = 1"),
        ),
        "keeps its card and number",
    );
    refused_with(
        exec(&mut f.w, "UPDATE card_ac SET seq = 2 WHERE seq = 1"),
        "keeps its card and number",
    );
    // Writing the same value back, and changing other columns, is allowed.
    exec(&mut f.w, "UPDATE card_ac SET seq = 1 WHERE seq = 1").unwrap();
    exec(
        &mut f.w,
        "UPDATE card_ac SET text = 'new', status = 'withdrawn', withdrawn_reason = 'no longer needed' WHERE seq = 1",
    )
    .unwrap();
    assert_eq!(
        text(&f.w, "SELECT status FROM card_ac WHERE seq = 1"),
        "withdrawn"
    );
}

// ---- card_note ------------------------------------------------------------------------------

fn note(card: i64, kind: &str, body: &str, reference: &str) -> String {
    format!(
        "INSERT INTO card_note (card_id, kind, author, origin, body, ref, created_at) \
         VALUES ({card}, '{kind}', 'a', 'import', '{body}', {reference}, 1)"
    )
}

#[test]
fn a_note_has_a_kind_origin_and_body_and_an_import_reference_counts_once() {
    let mut f = fx();
    let a = insert(&mut f.w, NEW_CARD).unwrap();
    let b = insert(&mut f.w, NEW_CARD).unwrap();
    for kind in ["decision", "report", "observation", "comment"] {
        insert(&mut f.w, &note(a, kind, "text", "NULL")).unwrap();
    }
    // Two notes without a reference are fine; the same (card, ref) twice is not.
    insert(&mut f.w, &note(a, "comment", "again", "NULL")).unwrap();
    insert(&mut f.w, &note(a, "comment", "x", "'c-1'")).unwrap();
    refused_with(
        exec(&mut f.w, &note(a, "comment", "x", "'c-1'")),
        "UNIQUE constraint failed",
    );
    insert(&mut f.w, &note(a, "comment", "x", "'c-2'")).unwrap();
    insert(&mut f.w, &note(b, "comment", "x", "'c-1'")).unwrap();
    refused(exec(&mut f.w, &note(a, "memo", "x", "NULL")));
    refused(exec(&mut f.w, &note(a, "comment", "", "NULL")));
    refused(exec(
        &mut f.w,
        &format!("INSERT INTO card_note (card_id, kind, author, origin, body, created_at) VALUES ({a}, 'comment', 'a', 'remote', 'x', 1)"),
    ));
}

// ---- card_link ------------------------------------------------------------------------------

const PR_COLS: &str = "INSERT INTO card_link (from_card, kind, to_card, target_host, target_project, target_number, pr_state, merge_ref, created_at, updated_at)";

#[test]
fn a_pr_link_has_a_target_and_no_card_and_a_card_link_the_opposite() {
    let mut f = fx();
    let a = insert(&mut f.w, NEW_CARD).unwrap();
    let b = insert(&mut f.w, NEW_CARD).unwrap();
    for kind in ["blocks", "split_of", "relates"] {
        insert(
            &mut f.w,
            &format!("{PR_COLS} VALUES ({a}, '{kind}', {b}, NULL, NULL, NULL, NULL, NULL, 1, 1)"),
        )
        .unwrap();
    }
    insert(
        &mut f.w,
        &format!(
            "{PR_COLS} VALUES ({a}, 'pr', NULL, 'github', 'acme/repo', 3, 'open', NULL, 1, 1)"
        ),
    )
    .unwrap();
    for bad in [
        // a pr link with a card
        format!("VALUES ({a}, 'pr', {b}, 'github', 'acme/repo', 4, NULL, NULL, 1, 1)"),
        // a pr link without any target column, then with each one missing
        format!("VALUES ({a}, 'pr', NULL, NULL, NULL, NULL, NULL, NULL, 1, 1)"),
        format!("VALUES ({a}, 'pr', NULL, NULL, 'acme/repo', 4, NULL, NULL, 1, 1)"),
        format!("VALUES ({a}, 'pr', NULL, 'github', NULL, 4, NULL, NULL, 1, 1)"),
        format!("VALUES ({a}, 'pr', NULL, 'github', 'acme/repo', NULL, NULL, NULL, 1, 1)"),
        // a card link without a card, or with a pull request target, or with a pr state
        format!("VALUES ({a}, 'blocks', NULL, NULL, NULL, NULL, NULL, NULL, 1, 1)"),
        format!("VALUES ({a}, 'relates', {b}, 'github', 'acme/repo', 4, NULL, NULL, 1, 1)"),
        format!("VALUES ({a}, 'relates', {b}, NULL, NULL, NULL, 'open', NULL, 1, 1)"),
    ] {
        refused_with(
            exec(&mut f.w, &format!("{PR_COLS} {bad}")),
            "CHECK constraint failed",
        );
    }
    assert_eq!(count(&f.w, "SELECT count(*) FROM card_link"), 4);
}

#[test]
fn a_merged_pr_needs_a_merge_reference_and_a_card_cannot_link_to_itself() {
    let mut f = fx();
    let a = insert(&mut f.w, NEW_CARD).unwrap();
    insert(
        &mut f.w,
        &format!(
            "{PR_COLS} VALUES ({a}, 'pr', NULL, 'github', 'acme/repo', 3, 'merged', 'abc', 1, 1)"
        ),
    )
    .unwrap();
    refused_with(
        exec(
            &mut f.w,
            &format!("{PR_COLS} VALUES ({a}, 'pr', NULL, 'github', 'acme/repo', 4, 'merged', NULL, 1, 1)"),
        ),
        "CHECK constraint failed",
    );
    for (number, state) in [(10, "open"), (12, "closed")] {
        insert(
            &mut f.w,
            &format!("{PR_COLS} VALUES ({a}, 'pr', NULL, 'github', 'acme/repo', {number}, '{state}', NULL, 1, 1)"),
        )
        .unwrap();
    }
    refused(exec(
        &mut f.w,
        &format!(
            "{PR_COLS} VALUES ({a}, 'pr', NULL, 'github', 'acme/repo', 11, 'unknown', NULL, 1, 1)"
        ),
    ));
    refused_with(
        exec(
            &mut f.w,
            &format!("{PR_COLS} VALUES ({a}, 'blocks', {a}, NULL, NULL, NULL, NULL, NULL, 1, 1)"),
        ),
        "CHECK constraint failed",
    );
}

#[test]
fn a_link_is_unique_per_card_pair_and_per_pr_and_the_pr_project_shape_is_checked() {
    let mut f = fx();
    let a = insert(&mut f.w, NEW_CARD).unwrap();
    let b = insert(&mut f.w, NEW_CARD).unwrap();
    let card_link =
        format!("{PR_COLS} VALUES ({a}, 'blocks', {b}, NULL, NULL, NULL, NULL, NULL, 1, 1)");
    insert(&mut f.w, &card_link).unwrap();
    refused_with(exec(&mut f.w, &card_link), "UNIQUE constraint failed");
    // Another kind between the same cards, and the reverse direction, are other links.
    insert(
        &mut f.w,
        &format!("{PR_COLS} VALUES ({a}, 'relates', {b}, NULL, NULL, NULL, NULL, NULL, 1, 1)"),
    )
    .unwrap();
    insert(
        &mut f.w,
        &format!("{PR_COLS} VALUES ({b}, 'blocks', {a}, NULL, NULL, NULL, NULL, NULL, 1, 1)"),
    )
    .unwrap();
    let pr =
        format!("{PR_COLS} VALUES ({a}, 'pr', NULL, 'github', 'acme/repo', 3, 'open', NULL, 1, 1)");
    insert(&mut f.w, &pr).unwrap();
    refused_with(exec(&mut f.w, &pr), "UNIQUE constraint failed");
    // The same pull request on another card is another link.
    insert(
        &mut f.w,
        &format!(
            "{PR_COLS} VALUES ({b}, 'pr', NULL, 'github', 'acme/repo', 3, 'open', NULL, 1, 1)"
        ),
    )
    .unwrap();
    // Amendment 1: the target project is normalised.
    for bad in ["'Acme/Repo'", "'acme/ repo'", "''"] {
        refused_with(
            exec(
                &mut f.w,
                &format!(
                    "{PR_COLS} VALUES ({a}, 'pr', NULL, 'github', {bad}, 8, NULL, NULL, 1, 1)"
                ),
            ),
            "CHECK constraint failed",
        );
    }
}

// ---- rows are never deleted, notes and events never change ---------------------------------

#[test]
fn rows_of_the_six_tables_are_never_deleted_and_a_replace_cannot_remove_them() {
    let mut f = fx();
    let c = insert(&mut f.w, NEW_CARD).unwrap();
    ac(&mut f.w, c, 1, "open").unwrap();
    insert(&mut f.w, &note(c, "comment", "x", "'ref-1'")).unwrap();
    merged_pr(&mut f.w, c);
    exec(
        &mut f.w,
        &format!(
            "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, lease_until, updated_at) \
             VALUES ({c}, 1, 'h', 'held', 1, 2, 1)"
        ),
    )
    .unwrap();
    exec(
        &mut f.w,
        "INSERT INTO event (event_key, at, actor, kind) VALUES ('k1', 1, 'a', 'card_add')",
    )
    .unwrap();
    for table in [
        "card",
        "card_ac",
        "card_note",
        "card_link",
        "reservation",
        "event",
    ] {
        refused_with(
            exec(&mut f.w, &format!("DELETE FROM {table}")),
            "rows are never deleted",
        );
        assert_eq!(
            count(&f.w, &format!("SELECT count(*) FROM {table}")),
            1,
            "{table} kept its row"
        );
    }
    // REPLACE removes the old row first; the delete trigger still runs (recursive triggers).
    refused_with(
        exec(
            &mut f.w,
            &format!(
                "INSERT OR REPLACE INTO card_note (card_id, kind, author, origin, body, ref, created_at) \
                 VALUES ({c}, 'comment', 'a', 'import', 'rewritten', 'ref-1', 2)"
            ),
        ),
        "rows are never deleted",
    );
    refused_with(
        exec(
            &mut f.w,
            &format!(
                "INSERT OR REPLACE INTO card (card_id, title, created_by, created_at, updated_at) \
                 VALUES ({c}, 'rewritten', 'x', 1, 1)"
            ),
        ),
        "rows are never deleted",
    );
    assert_eq!(text(&f.w, "SELECT body FROM card_note"), "x");
}

#[test]
fn a_note_and_an_event_can_never_be_changed() {
    let mut f = fx();
    let c = insert(&mut f.w, NEW_CARD).unwrap();
    insert(&mut f.w, &note(c, "comment", "original", "NULL")).unwrap();
    exec(
        &mut f.w,
        "INSERT INTO event (event_key, at, actor, kind) VALUES ('k1', 1, 'a', 'card_add')",
    )
    .unwrap();
    refused_with(
        exec(&mut f.w, "UPDATE card_note SET body = 'edited'"),
        "append-only",
    );
    refused_with(
        exec(&mut f.w, "UPDATE event SET actor = 'someone else'"),
        "append-only",
    );
    assert_eq!(text(&f.w, "SELECT body FROM card_note"), "original");
    assert_eq!(text(&f.w, "SELECT actor FROM event"), "a");
}

// ---- state transitions ----------------------------------------------------------------------

const STATES: [&str; 6] = ["draft", "ready", "active", "review", "done", "hold"];
const ALLOWED: [(&str, &str); 14] = [
    ("draft", "ready"),
    ("ready", "draft"),
    ("ready", "active"),
    ("active", "ready"),
    ("active", "review"),
    ("review", "active"),
    ("review", "done"),
    ("done", "ready"),
    ("draft", "hold"),
    ("ready", "hold"),
    ("active", "hold"),
    ("review", "hold"),
    ("hold", "draft"),
    ("hold", "ready"),
];

#[test]
fn the_transition_table_holds_exactly_the_contract_rows() {
    let f = fx();
    let mut rows: Vec<(String, String)> =
        f.w.conn()
            .prepare("SELECT from_state, to_state FROM card_transition ORDER BY 1, 2")
            .unwrap()
            .query_map([], |r| Ok((r.get(0)?, r.get(1)?)))
            .unwrap()
            .map(Result::unwrap)
            .collect();
    let mut want: Vec<(String, String)> = ALLOWED
        .iter()
        .map(|(a, b)| (a.to_string(), b.to_string()))
        .collect();
    rows.sort();
    want.sort();
    assert_eq!(rows, want);
}

#[test]
fn a_card_moves_between_states_only_along_the_transition_table() {
    let mut f = fx();
    for from in STATES {
        for to in STATES {
            if from == to {
                continue;
            }
            let c = card_in_state(&mut f.w, from);
            // Everything a close needs is in place, so only the transition table can refuse.
            ac(&mut f.w, c, 1, "met").unwrap();
            merged_pr(&mut f.w, c);
            let closed = if to == "done" { "9" } else { "NULL" };
            let result = exec(
                &mut f.w,
                &format!(
                    "UPDATE card SET state = '{to}', closed_at = {closed} WHERE card_id = {c}"
                ),
            );
            let allowed = ALLOWED.contains(&(from, to));
            match (allowed, result) {
                (true, Ok(())) => {
                    assert_eq!(
                        text(&f.w, &format!("SELECT state FROM card WHERE card_id = {c}")),
                        to
                    );
                }
                (false, r) => {
                    refused_with(r, "illegal card state transition");
                    assert_eq!(
                        text(&f.w, &format!("SELECT state FROM card WHERE card_id = {c}")),
                        from
                    );
                }
                (true, Err(e)) => panic!("{from} -> {to} should be allowed: {e}"),
            }
        }
    }
}

#[test]
fn writing_the_same_state_back_is_not_a_transition() {
    let mut f = fx();
    for state in STATES {
        let c = card_in_state(&mut f.w, state);
        exec(
            &mut f.w,
            &format!("UPDATE card SET state = '{state}' WHERE card_id = {c}"),
        )
        .unwrap();
    }
}

// ---- the done guard ---------------------------------------------------------------------------

#[test]
fn a_card_with_no_criterion_cannot_be_closed() {
    let mut f = fx();
    let c = card_in_state(&mut f.w, "review");
    merged_pr(&mut f.w, c);
    refused_with(close(&mut f.w, c), "no acceptance criterion that is met");
    assert_eq!(text(&f.w, "SELECT state FROM card"), "review");
}

#[test]
fn a_card_with_one_open_criterion_cannot_be_closed() {
    let mut f = fx();
    let c = card_in_state(&mut f.w, "review");
    ac(&mut f.w, c, 1, "met").unwrap();
    ac(&mut f.w, c, 2, "open").unwrap();
    merged_pr(&mut f.w, c);
    refused_with(close(&mut f.w, c), "still open");
}

#[test]
fn a_card_without_a_pr_link_cannot_be_closed() {
    let mut f = fx();
    let c = card_in_state(&mut f.w, "review");
    ac(&mut f.w, c, 1, "met").unwrap();
    refused_with(close(&mut f.w, c), "no merged pull request");
}

#[test]
fn only_an_active_merged_pr_link_of_this_card_lets_it_close() {
    let mut f = fx();
    let c = card_in_state(&mut f.w, "review");
    let other = card_in_state(&mut f.w, "review");
    ac(&mut f.w, c, 1, "met").unwrap();
    // Not merged, a retired merged link, a card link and another card's merged link: all refused.
    exec(
        &mut f.w,
        &format!(
            "{PR_COLS} VALUES ({c}, 'pr', NULL, 'github', 'acme/repo', 1, 'open', NULL, 1, 1)"
        ),
    )
    .unwrap();
    refused_with(close(&mut f.w, c), "no merged pull request");
    exec(
        &mut f.w,
        &format!(
            "{PR_COLS} VALUES ({c}, 'pr', NULL, 'github', 'acme/repo', 2, 'closed', NULL, 1, 1)"
        ),
    )
    .unwrap();
    refused_with(close(&mut f.w, c), "no merged pull request");
    let retired = merged_pr(&mut f.w, c);
    exec(
        &mut f.w,
        &format!("UPDATE card_link SET state = 'retired' WHERE link_id = {retired}"),
    )
    .unwrap();
    refused_with(close(&mut f.w, c), "no merged pull request");
    exec(
        &mut f.w,
        &format!("{PR_COLS} VALUES ({c}, 'relates', {other}, NULL, NULL, NULL, NULL, NULL, 1, 1)"),
    )
    .unwrap();
    refused_with(close(&mut f.w, c), "no merged pull request");
    merged_pr(&mut f.w, other);
    refused_with(close(&mut f.w, c), "no merged pull request");
    assert_eq!(
        text(&f.w, &format!("SELECT state FROM card WHERE card_id = {c}")),
        "review"
    );
}

#[test]
fn a_card_with_every_criterion_met_and_a_merged_pr_closes() {
    let mut f = fx();
    let c = card_in_state(&mut f.w, "review");
    ac(&mut f.w, c, 1, "met").unwrap();
    ac(&mut f.w, c, 2, "met").unwrap();
    merged_pr(&mut f.w, c);
    close(&mut f.w, c).unwrap();
    assert_eq!(text(&f.w, "SELECT state FROM card"), "done");
    assert_eq!(count(&f.w, "SELECT closed_at FROM card"), 9);
}

#[test]
fn a_withdrawn_criterion_does_not_block_a_close_but_cannot_be_the_only_one() {
    let mut f = fx();
    let c = card_in_state(&mut f.w, "review");
    ac(&mut f.w, c, 1, "met").unwrap();
    ac(&mut f.w, c, 2, "withdrawn").unwrap();
    merged_pr(&mut f.w, c);
    close(&mut f.w, c).unwrap();

    let only = card_in_state(&mut f.w, "review");
    ac(&mut f.w, only, 1, "withdrawn").unwrap();
    merged_pr(&mut f.w, only);
    refused_with(close(&mut f.w, only), "no acceptance criterion that is met");
}

#[test]
fn a_card_that_is_already_done_is_not_checked_again() {
    let mut f = fx();
    let c = card_in_state(&mut f.w, "done");
    // Writing the state and closing time back (what a recovery or a title edit may do).
    exec(
        &mut f.w,
        &format!("UPDATE card SET state = 'done', closed_at = 5 WHERE card_id = {c}"),
    )
    .unwrap();
}

// ---- no state check on INSERT (amendment 1) ------------------------------------------------

#[test]
fn a_recovery_can_insert_a_done_card_with_explicit_numbers_and_the_next_number_follows() {
    let mut f = fx();
    exec(
        &mut f.w,
        "INSERT INTO card (card_id, title, state, rev, created_by, created_at, updated_at, closed_at) \
         VALUES (7, 'restored', 'done', 5, 'recover', 100, 200, 300)",
    )
    .unwrap();
    assert_eq!(count(&f.w, "SELECT rev FROM card WHERE card_id = 7"), 5);
    assert_eq!(
        text(&f.w, "SELECT state FROM card WHERE card_id = 7"),
        "done"
    );
    // Its criteria and links follow the card.
    exec(
        &mut f.w,
        "INSERT INTO card_ac (ac_id, card_id, seq, text, status, evidence_ref, updated_at) \
         VALUES (40, 7, 1, 'c', 'met', 'pr:x#1', 1)",
    )
    .unwrap();
    // Inserting criteria touched the card; a recovery that wants the old rev writes it back.
    exec(&mut f.w, "UPDATE card SET rev = 5 WHERE card_id = 7").unwrap();
    assert_eq!(count(&f.w, "SELECT rev FROM card WHERE card_id = 7"), 5);
    let next = insert(&mut f.w, NEW_CARD).unwrap();
    assert_eq!(next, 8, "the next number is max + 1");
    let next_ac = ac(&mut f.w, next, 1, "open").unwrap();
    assert_eq!(next_ac, 41);
}

// ---- touch ------------------------------------------------------------------------------------

#[test]
fn a_card_change_bumps_rev_once_and_sets_updated_at_from_the_db_clock() {
    let mut f = fx();
    let c = insert(&mut f.w, NEW_CARD).unwrap();
    assert_eq!(count(&f.w, "SELECT rev FROM card"), 1);
    assert_eq!(count(&f.w, "SELECT updated_at FROM card"), 1);
    exec(
        &mut f.w,
        &format!("UPDATE card SET title = 'renamed' WHERE card_id = {c}"),
    )
    .unwrap();
    assert_eq!(
        count(&f.w, "SELECT rev FROM card"),
        2,
        "one bump per change"
    );
    assert!(
        count(&f.w, "SELECT updated_at FROM card") > 1_000_000_000,
        "updated_at is the DB clock"
    );
    exec(
        &mut f.w,
        &format!("UPDATE card SET state = 'ready' WHERE card_id = {c}"),
    )
    .unwrap();
    assert_eq!(count(&f.w, "SELECT rev FROM card"), 3);
}

#[test]
fn a_criterion_insert_or_update_touches_its_card_and_an_explicit_rev_is_kept() {
    let mut f = fx();
    let c = insert(&mut f.w, NEW_CARD).unwrap();
    ac(&mut f.w, c, 1, "open").unwrap();
    assert_eq!(count(&f.w, "SELECT rev FROM card"), 2);
    exec(
        &mut f.w,
        "UPDATE card_ac SET status = 'met', evidence_ref = 'pr:x#1' WHERE seq = 1",
    )
    .unwrap();
    assert_eq!(count(&f.w, "SELECT rev FROM card"), 3);
    assert!(count(&f.w, "SELECT updated_at FROM card") > 1_000_000_000);
    // A writer that sets rev itself (a recovery) is not bumped on top.
    exec(
        &mut f.w,
        &format!("UPDATE card SET rev = 40, title = 'restored' WHERE card_id = {c}"),
    )
    .unwrap();
    assert_eq!(count(&f.w, "SELECT rev FROM card"), 40);
}

// ---- reservation ------------------------------------------------------------------------------

fn reservation(
    w: &mut WriteLedger,
    card: i64,
    generation: i64,
    status: &str,
) -> Result<(), LedgerError> {
    exec(
        w,
        &format!(
            "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, lease_until, updated_at) \
             VALUES ({card}, {generation}, 'h', '{status}', 1, 2, 1)"
        ),
    )
}

#[test]
fn a_card_has_at_most_one_active_reservation() {
    let mut f = fx();
    let a = insert(&mut f.w, NEW_CARD).unwrap();
    let b = insert(&mut f.w, NEW_CARD).unwrap();
    reservation(&mut f.w, a, 1, "held").unwrap();
    // Each active status blocks another active row of the same card.
    for (generation, status) in [
        (2, "held"),
        (3, "expired_candidate"),
        (4, "held_for_review"),
    ] {
        refused_with(
            reservation(&mut f.w, a, generation, status),
            "UNIQUE constraint failed",
        );
    }
    // Rows that are no longer active, and other cards, are not blocked.
    reservation(&mut f.w, a, 5, "released").unwrap();
    reservation(&mut f.w, a, 6, "released").unwrap();
    reservation(&mut f.w, a, 7, "reclaimed").unwrap();
    reservation(&mut f.w, b, 1, "held").unwrap();
    // The active row going inactive frees the card.
    exec(
        &mut f.w,
        &format!(
            "UPDATE reservation SET status = 'released' WHERE card_id = {a} AND generation = 1"
        ),
    )
    .unwrap();
    reservation(&mut f.w, a, 8, "expired_candidate").unwrap();
    refused_with(
        reservation(&mut f.w, a, 9, "held"),
        "UNIQUE constraint failed",
    );
}

#[test]
fn a_reservation_generation_starts_at_one_and_is_unique_per_card() {
    let mut f = fx();
    let a = insert(&mut f.w, NEW_CARD).unwrap();
    refused_with(
        reservation(&mut f.w, a, 0, "held"),
        "CHECK constraint failed",
    );
    reservation(&mut f.w, a, 1, "released").unwrap();
    refused_with(
        reservation(&mut f.w, a, 1, "released"),
        "UNIQUE constraint failed",
    );
    refused(reservation(&mut f.w, a, 2, "lost"));
}

// ---- event ------------------------------------------------------------------------------------

#[test]
fn an_event_key_counts_once_and_a_replay_adds_nothing() {
    let mut f = fx();
    let insert_event = "INSERT INTO event (event_key, at, actor, kind) VALUES ('import:github:acme/repo:5:abc', 1, 'a', 'import')";
    exec(&mut f.w, insert_event).unwrap();
    refused_with(exec(&mut f.w, insert_event), "UNIQUE constraint failed");
    let n = f
        .w
        .with_write_tx(|tx, _| {
            Ok(tx.execute(
                "INSERT INTO event (event_key, at, actor, kind) VALUES ('import:github:acme/repo:5:abc', 1, 'a', 'import') \
                 ON CONFLICT (event_key) DO NOTHING",
                [],
            )?)
        })
        .unwrap();
    assert_eq!(n, 0, "a replay of the same key changes nothing");
    assert_eq!(count(&f.w, "SELECT count(*) FROM event"), 1);
    // The kind list is open: a kind a later slice adds needs no schema change.
    exec(
        &mut f.w,
        "INSERT INTO event (event_key, at, actor, kind) VALUES ('k2', 1, 'a', 'host_export')",
    )
    .unwrap();
    // The payload must be JSON.
    refused_with(
        exec(
            &mut f.w,
            "INSERT INTO event (event_key, at, actor, kind, payload) VALUES ('k3', 1, 'a', 'x', 'not json')",
        ),
        "CHECK constraint failed",
    );
}

#[test]
fn a_local_event_key_is_kind_card_and_32_random_hex_digits() {
    let f = fx();
    let key = ledger::new_event_key(f.w.conn(), "card_add", Some(12)).unwrap();
    let parts: Vec<&str> = key.split(':').collect();
    assert_eq!(&parts[..2], ["card_add", "12"], "{key}");
    assert_eq!(parts.len(), 3);
    assert_eq!(parts[2].len(), 32, "{key}");
    assert!(parts[2]
        .bytes()
        .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b)));
    let none = ledger::new_event_key(f.w.conn(), "recover", None).unwrap();
    assert!(none.starts_with("recover:0:"), "{none}");
    let again = ledger::new_event_key(f.w.conn(), "card_add", Some(12)).unwrap();
    assert_ne!(key, again, "the random part differs on every call");
}
