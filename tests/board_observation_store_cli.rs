//! BOARD 02 (SPEC-20261120-board-observation-store): the observation store beside the ledger —
//! one file in the project folder with the ledger folder's rules (owner-only, links refused),
//! idempotent rows unique per (source, source event id), the six fixed kinds, 30-day retention,
//! cursor advance and appends in one transaction, and an opaque read cursor returned verbatim.
//! The module is included by path: its product declaration in `src/lib.rs` waits for that
//! file's reservation to free (decision note `decision-lib-decl-deferred`), so the module is
//! written with `moai_cli::` absolute paths only. Every case runs in a temporary HOME.

#[path = "../src/board/observation.rs"]
mod observation;

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use observation::{Observation, Store};

const NOW: i64 = 1_800_000_000;

fn row(source: &str, event_id: &str, kind: &str, content: &str) -> Observation {
    Observation {
        source: source.into(),
        source_event_id: event_id.into(),
        scope: "project".into(),
        observed_at: NOW,
        kind: kind.into(),
        content: content.into(),
    }
}

/// A HOME with a registered ledger and one stored observation row.
fn store_with_one_row() -> (TempHome, Store) {
    let h = TempHome::new();
    ledger::register(h.path(), "demo").unwrap();
    let mut s = Store::open(h.path(), "demo").unwrap();
    let a = s
        .commit(
            &[row(
                "leader-round",
                "e1",
                "lane-observation@1",
                "{\"lane\":\"a\"}",
            )],
            "c-1",
        )
        .unwrap();
    assert_eq!(a.added, 1);
    (h, s)
}

#[test]
fn the_store_is_one_private_file_beside_the_ledger_and_links_are_refused() {
    let h = TempHome::new();
    ledger::register(h.path(), "demo").unwrap();
    let mut s = Store::open(h.path(), "demo").unwrap();
    s.commit(&[row("x", "e1", "wake@1", "{}")], "c").unwrap();
    let store_path = s.path().to_path_buf();
    let project = store_path.parent().unwrap().to_path_buf();
    assert!(
        project.ends_with(".moai/projects/demo"),
        "the store lives in the project folder: {}",
        project.display()
    );
    assert!(
        project.join("ledger.db").exists(),
        "the ledger is beside it"
    );
    assert_eq!(store_path.file_name().unwrap(), "observation.db");
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let mode = std::fs::metadata(&store_path).unwrap().permissions().mode();
        assert_eq!(mode & 0o777, 0o600, "the store file is owner-only");
    }
    // A link where the store would go is refused and nothing is followed.
    drop(s);
    std::fs::remove_file(&store_path).unwrap();
    let elsewhere = h.path().join("elsewhere");
    std::fs::write(&elsewhere, "not a store").unwrap();
    std::os::unix::fs::symlink(&elsewhere, &store_path).unwrap();
    assert!(Store::open(h.path(), "demo").is_err());
}

#[test]
fn feeding_the_same_observation_file_twice_adds_no_rows() {
    let (_h, mut s) = store_with_one_row();
    let again = s
        .commit(
            &[row(
                "leader-round",
                "e1",
                "lane-observation@1",
                "{\"lane\":\"a\"}",
            )],
            "c-2",
        )
        .unwrap();
    assert_eq!(again.added, 0, "the second feed adds nothing");
    assert_eq!(again.unchanged, 1);
    assert!(again.conflicts.is_empty());
    assert_eq!(s.count(), 1);
}

#[test]
fn the_same_id_with_different_content_is_kept_apart_as_a_conflict() {
    let (_h, mut s) = store_with_one_row();
    let again = s
        .commit(
            &[row(
                "leader-round",
                "e1",
                "lane-observation@1",
                "{\"lane\":\"B\"}",
            )],
            "c-2",
        )
        .unwrap();
    assert_eq!(again.added, 0);
    assert_eq!(again.unchanged, 0);
    assert_eq!(again.conflicts.len(), 1, "the differing row is named");
    assert_eq!(s.count(), 1, "the original row stays");
}

#[test]
fn the_six_kinds_are_fixed_and_an_unknown_kind_is_refused() {
    let (_h, mut s) = store_with_one_row();
    for kind in [
        "lane-observation@1",
        "pr-state@1",
        "local-check@1",
        "bundle-run@1",
        "gap@1",
        "wake@1",
    ] {
        let a = s
            .commit(&[row("k", &format!("e-{kind}"), kind, "{}")], "c")
            .unwrap();
        assert_eq!(a.added, 1, "{kind}");
    }
    assert_eq!(s.count(), 7);
    let refused = s.commit(&[row("k", "e-bad", "run-start@1", "{}")], "c");
    assert!(refused.is_err(), "an unknown kind is refused");
    assert_eq!(s.count(), 7, "the failed batch wrote nothing");
}

#[test]
fn a_failed_batch_leaves_the_cursor_where_it_was() {
    let (_h, mut s) = store_with_one_row();
    assert_eq!(s.cursor().as_deref(), Some("c-1"));
    assert!(s
        .commit(
            &[
                row("k", "e-ok", "gap@1", "{}"),
                row("k", "e-bad", "nope@9", "{}")
            ],
            "c-2"
        )
        .is_err());
    assert_eq!(
        s.cursor().as_deref(),
        Some("c-1"),
        "one transaction: no cursor advance"
    );
    assert_eq!(s.count(), 1, "and no rows");
}

#[test]
fn observations_older_than_thirty_days_are_pruned_and_the_ledger_stays() {
    let h = TempHome::new();
    ledger::register(h.path(), "demo").unwrap();
    let ledger_path = h.path().join(".moai/projects/demo/ledger.db");
    let mut s = Store::open(h.path(), "demo").unwrap();
    let old = Observation {
        observed_at: NOW - 31 * 24 * 3600,
        ..row("old", "e-old", "pr-state@1", "{}")
    };
    s.commit(&[old, row("new", "e-new", "pr-state@1", "{}")], "c")
        .unwrap();
    let before = std::fs::read(&ledger_path).unwrap();
    let pruned = s.prune(NOW).unwrap();
    assert_eq!(pruned, 1, "only the 31-day-old row goes");
    assert_eq!(s.count(), 1);
    assert_eq!(
        std::fs::read(&ledger_path).unwrap(),
        before,
        "the ledger file is untouched"
    );
}

#[test]
fn deleting_the_cursor_and_committing_again_gives_the_same_result() {
    let (_h, mut s) = store_with_one_row();
    s.clear_cursor().unwrap();
    assert_eq!(s.cursor(), None);
    let again = s
        .commit(
            &[row(
                "leader-round",
                "e1",
                "lane-observation@1",
                "{\"lane\":\"a\"}",
            )],
            "c-1",
        )
        .unwrap();
    assert_eq!(again.added, 0, "no cursor, same rows: idempotent");
    assert_eq!(s.count(), 1);
}

#[test]
fn the_read_cursor_is_opaque_and_returned_verbatim() {
    let (h, mut s) = store_with_one_row();
    for i in 2..=5 {
        s.commit(
            &[row("r", &format!("e{i}"), "wake@1", "{}")],
            &format!("c-{i}"),
        )
        .unwrap();
    }
    let r = observation::open_read(h.path(), "demo").unwrap();
    let (page, next) = r.read_after(None, 2).unwrap();
    assert_eq!(page.len(), 2);
    let next = next.unwrap();
    let (rest, end) = r.read_after(Some(&next), 10).unwrap();
    assert_eq!(rest.len(), 3);
    assert!(end.is_none());
    // The cursor the caller hands back is used as it is: a foreign string is refused, not
    // recomputed.
    assert!(r.read_after(Some("not-a-cursor"), 10).is_err());
    assert_eq!(
        s.cursor().as_deref(),
        Some("c-5"),
        "the stored cursor is verbatim too"
    );
}

#[test]
fn reading_adds_no_rows_and_changes_no_file() {
    let h = TempHome::new();
    ledger::register(h.path(), "demo").unwrap();
    let mut s = Store::open(h.path(), "demo").unwrap();
    s.commit(&[row("r", "e1", "wake@1", "{}")], "c").unwrap();
    let before = h.snapshot();
    let r = observation::open_read(h.path(), "demo").unwrap();
    let (page, _) = r.read_after(None, 10).unwrap();
    assert_eq!(page.len(), 1);
    drop(r);
    drop(s);
    assert_eq!(h.snapshot(), before, "reading wrote nothing");
}
