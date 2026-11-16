//! Recovering a damaged ledger (Issue #3068 P6; AC7 and a reinforcement of AC6). Every test
//! runs in its own temporary HOME; none touches the real `~/.moai`. The kill tests start this
//! same test binary again (`--exact` on a child test that does nothing unless its environment
//! variable is set) and use the library's small steps, so no kill hook exists in production code.

mod common;

use common::home::TempHome;
use common::CleanEnv;
use moai_cli::ledger::recover::{
    self, ApplyState, JournalFact, Observations, RemoteFact, WorktreeFact, HELD_FOR_REVIEW,
    RESTORE_HELD,
};
use moai_cli::ledger::reservation::{self, FenceError, ReserveError};
use moai_cli::ledger::{self, LedgerError, WriteLedger};
use rusqlite::Connection;
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::time::{Duration, Instant};

const PROJECT: &str = "demo";

fn ledger_file(home: &TempHome) -> PathBuf {
    ledger::ledger_path(home.path(), PROJECT).unwrap()
}

fn sibling(path: &Path, suffix: &str) -> PathBuf {
    let mut os = path.as_os_str().to_os_string();
    os.push(suffix);
    PathBuf::from(os)
}

fn sql(w: &mut WriteLedger, statement: &str) {
    w.with_write_tx(|tx, _| {
        tx.execute_batch(statement)?;
        Ok(())
    })
    .unwrap();
}

fn add_card(w: &mut WriteLedger, state: &str) -> i64 {
    w.with_write_tx(|tx, now| {
        tx.execute(
            "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
             VALUES ('card', ?1, 'test', ?2, ?2, CASE WHEN ?1 = 'done' THEN ?2 END)",
            rusqlite::params![state, now],
        )?;
        Ok(tx.last_insert_rowid())
    })
    .unwrap()
}

fn s(patterns: &[&str]) -> Vec<String> {
    patterns.iter().map(|p| p.to_string()).collect()
}

/// A project whose ledger holds every kind of row:
/// - card 1: alice holds generation 1 (worktree `/wt/alice`, branch `issue/1`)
/// - card 2: bob holds generation 1 (worktree `/wt/bob`, branch `issue/2`)
/// - card 3: carol's generation 1 was handed to dave: generation 2 `held`, generation 1 `released`
/// - card 4: reserved and released (`released`, card `ready`)
/// - card 5: `draft` with a criterion, a note and a pull request link
/// - card 6: `done`
struct World {
    home: TempHome,
}

impl World {
    fn new() -> World {
        let home = TempHome::new();
        ledger::register(home.path(), PROJECT).unwrap();
        let mut w = ledger::open_write(home.path(), PROJECT).unwrap();
        let (c1, c2, c3, c4) = (
            add_card(&mut w, "ready"),
            add_card(&mut w, "ready"),
            add_card(&mut w, "ready"),
            add_card(&mut w, "ready"),
        );
        reservation::reserve(
            &mut w,
            c1,
            "alice",
            &s(&["zone1/"]),
            3600,
            Some("/wt/alice"),
            Some("issue/1"),
        )
        .unwrap();
        reservation::reserve(
            &mut w,
            c2,
            "bob",
            &s(&["zone2/"]),
            3600,
            Some("/wt/bob"),
            Some("issue/2"),
        )
        .unwrap();
        reservation::reserve(
            &mut w,
            c3,
            "carol",
            &s(&["zone3/", "zone3b/"]),
            3600,
            Some("/wt/carol"),
            Some("issue/3"),
        )
        .unwrap();
        reservation::handoff(&mut w, c3, "carol", 1, "dave").unwrap();
        reservation::reserve(&mut w, c4, "erin", &s(&["zone4/"]), 3600, None, None).unwrap();
        reservation::release(&mut w, c4, "erin", 1).unwrap();
        let c5 = add_card(&mut w, "draft");
        let _c6 = add_card(&mut w, "done");
        sql(
            &mut w,
            &format!(
                "INSERT INTO card_ac (card_id, seq, text, updated_at) VALUES ({c5}, 1, 'works', 1);
                 INSERT INTO card_note (card_id, kind, author, body, created_at) VALUES ({c5}, 'decision', 'me', 'chose X', 1);
                 INSERT INTO card_link (from_card, kind, target_host, target_project, target_number, pr_state, merge_ref, created_at, updated_at)
                   VALUES ({c5}, 'pr', 'github', 'o/r', 12, 'merged', 'abc', 1, 1);
                 INSERT INTO event (event_key, at, actor, card_id, kind, payload) VALUES ('import:github:o/r:12:x', 5, 'me', {c5}, 'import', '{{}}');"
            ),
        );
        drop(w);
        World { home }
    }

    fn ledger(&self) -> PathBuf {
        ledger_file(&self.home)
    }

    /// A byte copy of the ledger outside the project folder: a source that stays as it is.
    fn copy_as(&self, name: &str) -> PathBuf {
        let to = self.home.path().join("sources").join(name);
        fs::create_dir_all(to.parent().unwrap()).unwrap();
        fs::copy(self.ledger(), &to).unwrap();
        to
    }

    fn reader(&self) -> ledger::ReadLedger {
        ledger::open_read_only(self.home.path(), PROJECT).unwrap()
    }
}

fn plan_of(w: &World, source: &Path, obs: &Observations) -> recover::RecoveryPlan {
    recover::plan(w.home.path(), PROJECT, source, obs).unwrap()
}

/// Journal, worktree and remote facts that confirm every active reservation of `source`.
fn confirm_all(w: &World, source: &Path) -> Observations {
    let mut obs = Observations::default();
    for r in recover::queries(w.home.path(), PROJECT, source).unwrap() {
        obs.journal.push(JournalFact {
            holder: r.holder.clone(),
            generation: r.generation,
            live: true,
        });
        let (path, branch) = (r.worktree.clone().unwrap(), r.branch.clone().unwrap());
        obs.worktrees.push(WorktreeFact {
            path,
            exists: true,
            branch: Some(branch.clone()),
            clean: true,
        });
        obs.remotes.push(RemoteFact {
            branch,
            card_id: Some(r.card_id),
            holder: Some(r.holder),
            generation: Some(r.generation),
            exists: true,
        });
    }
    obs
}

/// Every row of `table`, as one JSON array text per row, in key order.
fn dump(conn: &Connection, table: &str, cols: &str, order: &str) -> Vec<String> {
    conn.prepare(&format!(
        "SELECT json_array({cols}) FROM {table} ORDER BY {order}"
    ))
    .unwrap()
    .query_map([], |r| r.get::<_, String>(0))
    .unwrap()
    .map(Result::unwrap)
    .collect()
}

const CARD_COLS: &str = "card_id, title, body, kind, state, priority, write_scope, generation, rev, created_by, created_at, updated_at, closed_at, source_host, source_project, source_number, source_url";
const AC_COLS: &str =
    "ac_id, card_id, seq, text, verify, status, evidence_ref, withdrawn_reason, updated_at";
const NOTE_COLS: &str = "note_id, card_id, kind, author, origin, body, ref, created_at";
const LINK_COLS: &str = "link_id, from_card, kind, to_card, target_host, target_project, target_number, pr_state, merge_ref, state, created_at, updated_at";
const EVENT_COLS: &str = "event_id, event_key, at, actor, card_id, kind, payload";
const RES_COLS: &str = "reservation_id, card_id, generation, holder, status, acquired_at, lease_until, updated_at, worktree, branch";

fn count(conn: &Connection, table: &str) -> i64 {
    conn.query_row(&format!("SELECT count(*) FROM {table}"), [], |r| r.get(0))
        .unwrap()
}

fn sequence(conn: &Connection, table: &str) -> Option<i64> {
    conn.query_row(
        "SELECT seq FROM sqlite_sequence WHERE name = ?1",
        [table],
        |r| r.get(0),
    )
    .ok()
}

fn statuses(conn: &Connection) -> Vec<(i64, i64, String)> {
    conn.prepare("SELECT card_id, generation, status FROM reservation ORDER BY reservation_id")
        .unwrap()
        .query_map([], |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?)))
        .unwrap()
        .map(Result::unwrap)
        .collect()
}

fn recover_events(conn: &Connection) -> Vec<Value> {
    conn.prepare("SELECT payload FROM event WHERE kind = 'recover' ORDER BY event_id")
        .unwrap()
        .query_map([], |r| r.get::<_, String>(0))
        .unwrap()
        .map(|p| serde_json::from_str(&p.unwrap()).unwrap())
        .collect()
}

fn names_in_project(home: &TempHome) -> Vec<String> {
    TempHome::names_in(ledger_file(home).parent().unwrap())
}

fn kept_files(home: &TempHome) -> Vec<String> {
    names_in_project(home)
        .into_iter()
        .filter(|n| n.contains(".damaged-"))
        .collect()
}

const SUFFIX: &str = "20261116T010203Z";

/// Damages `path` by overwriting `len` bytes at `at`.
fn overwrite(path: &Path, at: u64, bytes: &[u8]) {
    use std::io::{Seek, SeekFrom, Write};
    let mut f = fs::OpenOptions::new().write(true).open(path).unwrap();
    f.seek(SeekFrom::Start(at)).unwrap();
    f.write_all(bytes).unwrap();
}

// ---- the plan is read-only ----------------------------------------------------------------------

#[test]
fn a_plan_writes_nothing_even_to_the_source() {
    let w = World::new();
    let source = w.copy_as("damaged.db");
    let obs = confirm_all(&w, &source);
    let before = w.home.snapshot();
    let plan = plan_of(&w, &source, &obs);
    assert!(plan.plan_digest.starts_with("sha256:"));
    assert_eq!(plan.target, sibling(&w.ledger(), ".recovering"));
    // The live ledger as the source too.
    let live = recover::plan(w.home.path(), PROJECT, &w.ledger(), &obs).unwrap();
    assert_eq!(live.plan_digest, plan.plan_digest, "same bytes, same plan");
    assert_eq!(
        w.home.snapshot(),
        before,
        "no new entry, no changed byte, no changed modification time"
    );
}

#[test]
fn the_plan_counts_every_table_and_states_the_decisions() {
    let w = World::new();
    let source = w.copy_as("damaged.db");
    let r = w.reader();
    let plan = plan_of(&w, &source, &confirm_all(&w, &source));
    let counts: Vec<(String, u64, u64)> = plan
        .tables
        .iter()
        .map(|t| (t.table.clone(), t.rows_read, t.copied))
        .collect();
    for (table, rows_read, copied) in &counts {
        assert_eq!(*rows_read as i64, count(r.conn(), table), "{table}");
        assert_eq!(rows_read, copied, "{table}: nothing is lost or refused");
    }
    assert!(plan.tables.iter().all(|t| t.state == "ok"));
    assert_eq!(plan.source_check, "ok");
    assert!(plan.unreadable_tables.is_empty());
    assert_eq!(plan.rejected_total, 0);
    // Three active reservations (alice, bob, dave); two `released` rows are copied as they are.
    assert_eq!(plan.decisions.len(), 3);
    assert_eq!(plan.copied_as_is, 2);
    assert_eq!((plan.restore_held, plan.held_for_review), (3, 0));
    assert!(plan.decisions.iter().all(|d| d.decision == RESTORE_HELD));
    assert_eq!(plan.card_generations_raised, 0);
    assert!(plan.active_cards_without_holder.is_empty());
    assert_eq!(plan.target_state, recover::TargetState::Healthy);
}

#[test]
fn a_plan_in_a_home_that_is_not_ready_is_refused() {
    let w = World::new();
    let source = w.copy_as("damaged.db");
    w.home.make_go_generation();
    let err = recover::plan(w.home.path(), PROJECT, &source, &Observations::default()).unwrap_err();
    assert!(matches!(err, LedgerError::HomeNotReady(_)), "{err:?}");
    let err = recover::apply(
        w.home.path(),
        PROJECT,
        &source,
        &Observations::default(),
        "sha256:x",
    )
    .unwrap_err();
    assert!(matches!(err, LedgerError::HomeNotReady(_)), "{err:?}");
}

// ---- damaged sources ----------------------------------------------------------------------------

/// One damaged source: what the plan answers, and that it wrote nothing.
fn plan_error(w: &World, source: &Path) -> LedgerError {
    let before = w.home.snapshot();
    let err = recover::plan(w.home.path(), PROJECT, source, &Observations::default()).unwrap_err();
    assert_eq!(
        w.home.snapshot(),
        before,
        "a refused plan writes nothing: {err}"
    );
    let queried = recover::queries(w.home.path(), PROJECT, source).unwrap_err();
    assert_eq!(queried, err, "the reservation query refuses the same way");
    err
}

#[test]
fn a_source_that_is_not_a_database_is_refused() {
    let w = World::new();
    let source = w.copy_as("notadb.db");
    overwrite(&source, 0, b"NOT A DATABASE!!");
    let err = plan_error(&w, &source);
    assert!(
        matches!(&err, LedgerError::Unavailable(m) if m.contains("not a SQLite database")),
        "{err:?}"
    );
    let empty = w.home.path().join("sources/empty.db");
    fs::write(&empty, b"").unwrap();
    assert!(matches!(
        plan_error(&w, &empty),
        LedgerError::Unavailable(_)
    ));
    let short = w.home.path().join("sources/short.db");
    fs::write(&short, &fs::read(w.ledger()).unwrap()[..99]).unwrap();
    assert!(matches!(
        plan_error(&w, &short),
        LedgerError::Unavailable(_)
    ));
}

#[test]
fn a_source_of_another_application_or_project_or_version_is_refused() {
    let w = World::new();
    // another application_id (header bytes 68..72)
    let other_app = w.copy_as("other_app.db");
    overwrite(&other_app, 68, &0x1234_5678_i32.to_be_bytes());
    assert!(matches!(
        plan_error(&w, &other_app),
        LedgerError::Unavailable(m) if m.contains("application_id")
    ));
    // a newer schema version (header bytes 60..64)
    let newer = w.copy_as("newer.db");
    overwrite(&newer, 60, &2_i32.to_be_bytes());
    assert_eq!(
        plan_error(&w, &newer),
        LedgerError::SchemaNewer {
            found: 2,
            supported: 1
        }
    );
    // another project's ledger
    let other = TempHome::new();
    ledger::register(other.path(), "other").unwrap();
    let theirs = ledger::ledger_path(other.path(), "other").unwrap();
    let source = w.home.path().join("sources/other.db");
    fs::create_dir_all(source.parent().unwrap()).unwrap();
    fs::copy(theirs, &source).unwrap();
    assert_eq!(
        plan_error(&w, &source),
        LedgerError::IdCollision {
            expected: PROJECT.into(),
            found: "other".into()
        }
    );
}

#[test]
fn a_source_in_wal_mode_is_refused() {
    let w = World::new();
    let wal = w.copy_as("wal.db");
    overwrite(&wal, 18, &[2, 2]);
    assert_eq!(plan_error(&w, &wal), LedgerError::WalUnexpected);
    // one byte is enough, either of the two
    let one = w.copy_as("wal_one.db");
    overwrite(&one, 19, &[2]);
    assert_eq!(plan_error(&w, &one), LedgerError::WalUnexpected);
}

#[test]
fn a_link_a_folder_or_a_missing_file_is_not_a_source() {
    let w = World::new();
    let real = w.copy_as("real.db");
    let link = w.home.path().join("sources/link.db");
    std::os::unix::fs::symlink(&real, &link).unwrap();
    assert!(matches!(
        plan_error(&w, &link),
        LedgerError::Unavailable(m) if m.contains("link or not a regular file")
    ));
    let folder = w.home.path().join("sources/folder.db");
    fs::create_dir_all(&folder).unwrap();
    assert!(matches!(
        plan_error(&w, &folder),
        LedgerError::Unavailable(m) if m.contains("link or not a regular file")
    ));
    assert!(matches!(
        plan_error(&w, &w.home.path().join("sources/none.db")),
        LedgerError::Unavailable(m) if m.contains("does not exist")
    ));
}

#[test]
fn a_source_with_a_leftover_journal_is_refused_until_a_writer_rolled_it_back() {
    let w = World::new();
    let source = w.copy_as("hot.db");
    fs::write(sibling(&source, "-journal"), b"left behind").unwrap();
    assert!(matches!(
        plan_error(&w, &source),
        LedgerError::Unavailable(m) if m.contains("needs_recovery")
    ));
}

#[test]
fn a_truncated_source_gives_what_the_remaining_pages_hold() {
    let w = World::new();
    // Many events, so the file spans many pages; then cut its tail off.
    let mut writer = ledger::open_write(w.home.path(), PROJECT).unwrap();
    sql(
        &mut writer,
        "WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n WHERE i < 400) \
         INSERT INTO event (event_key, at, actor, kind, payload) \
         SELECT 'fill:' || i, i, 'x', 'filler', json_object('pad', printf('%0300d', i)) FROM n;",
    );
    drop(writer);
    let total_events = count(w.reader().conn(), "event");
    let source = w.copy_as("truncated.db");
    let len = fs::metadata(&source).unwrap().len();
    let pages = len / 4096;
    assert!(pages > 30, "the fixture spans many pages: {pages}");
    // The last third of the file is gone: the events, which were added last.
    fs::OpenOptions::new()
        .write(true)
        .open(&source)
        .unwrap()
        .set_len(pages * 2 / 3 * 4096)
        .unwrap();
    let before = w.home.snapshot();
    let plan = plan_of(&w, &source, &Observations::default());
    assert_eq!(w.home.snapshot(), before);
    let event = plan.tables.iter().find(|t| t.table == "event").unwrap();
    assert_ne!(event.state, "ok", "the cut lost event pages: {event:?}");
    assert!(
        (event.rows_read as i64) < total_events,
        "{} of {total_events}",
        event.rows_read
    );
    assert_eq!(
        event.last_good_id,
        (event.rows_read > 0).then_some(event.rows_read as i64)
    );
    assert_eq!(
        event.lost_estimate,
        Some(total_events - event.rows_read as i64),
        "the source's own id counter says how many rows are gone"
    );
    assert!(event.error.is_some());
    assert_ne!(plan.source_check, "ok");
    // Every table reports; none aborts the plan.
    assert_eq!(plan.tables.len(), 7);
}

#[test]
fn a_source_cut_inside_its_schema_is_refused_because_its_project_cannot_be_confirmed() {
    let w = World::new();
    let source = w.copy_as("cut_short.db");
    fs::OpenOptions::new()
        .write(true)
        .open(&source)
        .unwrap()
        .set_len(4096)
        .unwrap();
    let err = plan_error(&w, &source);
    assert!(
        matches!(&err, LedgerError::Unavailable(m) if m.contains("project id inside the source cannot be read")),
        "{err:?}"
    );
}

/// The leaf pages of `table` in the file at `path`, by the `dbstat` view.
fn leaf_pages(path: &Path, table: &str) -> Vec<u64> {
    let conn =
        Connection::open_with_flags(path, rusqlite::OpenFlags::SQLITE_OPEN_READ_ONLY).unwrap();
    let mut stmt = conn
        .prepare("SELECT pageno FROM dbstat WHERE name = ?1 AND pagetype = 'leaf' ORDER BY pageno")
        .unwrap();
    let pages: Vec<u64> = stmt
        .query_map([table], |r| r.get::<_, i64>(0).map(|n| n as u64))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    pages
}

/// A world whose event table spans many pages, and a copy with one middle leaf page of `event`
/// overwritten. Returns the world, the source, the number of events the ledger held, and the
/// highest id the event counter had reached.
fn world_with_a_broken_event_page() -> (World, PathBuf, i64, i64) {
    let w = World::new();
    let mut writer = ledger::open_write(w.home.path(), PROJECT).unwrap();
    sql(
        &mut writer,
        "WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n WHERE i < 600) \
         INSERT INTO event (event_key, at, actor, kind, payload) \
         SELECT 'fill:' || i, i, 'x', 'filler', json_object('pad', printf('%0300d', i)) FROM n;",
    );
    drop(writer);
    let total = count(w.reader().conn(), "event");
    let counter = sequence(w.reader().conn(), "event").unwrap();
    let source = w.copy_as("broken_page.db");
    let leaves = leaf_pages(&source, "event");
    assert!(leaves.len() > 20, "the fixture has {} leaves", leaves.len());
    let middle = leaves[leaves.len() / 2];
    overwrite(&source, (middle - 1) * 4096, &vec![0xFF; 4096]);
    (w, source, total, counter)
}

#[test]
fn a_broken_page_in_the_middle_of_event_stops_that_table_at_its_last_good_row() {
    let (w, source, total, counter) = world_with_a_broken_event_page();
    let before = w.home.snapshot();
    let plan = plan_of(&w, &source, &Observations::default());
    assert_eq!(w.home.snapshot(), before, "the plan wrote nothing");
    let event = plan.tables.iter().find(|t| t.table == "event").unwrap();
    assert_eq!(event.state, "partial");
    assert!(event.rows_read > 0 && (event.rows_read as i64) < total);
    assert_eq!(event.last_good_id, Some(event.rows_read as i64));
    assert_eq!(event.lost_estimate, Some(counter - event.rows_read as i64));
    assert!(event.error.as_deref().unwrap().contains("malformed"));
    assert_eq!(event.copied, event.rows_read);
    // The other tables are whole.
    for t in plan.tables.iter().filter(|t| t.table != "event") {
        assert_eq!(t.state, "ok", "{}", t.table);
    }
    assert!(plan.unreadable_tables.is_empty());
    assert_ne!(plan.source_check, "ok");
}

#[test]
fn a_recovery_from_a_partial_event_table_never_reissues_a_lost_id() {
    let (w, source, total, counter) = world_with_a_broken_event_page();
    let obs = Observations::default();
    let plan = plan_of(&w, &source, &obs);
    let read = plan
        .tables
        .iter()
        .find(|t| t.table == "event")
        .unwrap()
        .rows_read as i64;
    assert!(read < total);
    fs::remove_file(w.ledger()).unwrap();
    let result = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &source,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap();
    assert_eq!(result.state, ApplyState::Applied);
    let r = w.reader();
    // the rows read plus the one `recover` event; the counter stays past the highest id issued
    assert_eq!(count(r.conn(), "event"), read + 1);
    let recover_id: i64 = r
        .conn()
        .query_row(
            "SELECT event_id FROM event WHERE kind = 'recover'",
            [],
            |row| row.get(0),
        )
        .unwrap();
    assert_eq!(
        recover_id,
        counter + 1,
        "the first new id is past every id the source issued"
    );
    assert_eq!(sequence(r.conn(), "event"), Some(counter + 1));
}

// ---- apply --------------------------------------------------------------------------------------

/// Damages the ledger beyond use: a header that is not SQLite's. The good copy is the `.bak`.
fn break_ledger(w: &World) -> PathBuf {
    let bak = sibling(&w.ledger(), ".bak");
    fs::copy(w.ledger(), &bak).unwrap();
    overwrite(&w.ledger(), 0, b"NOT A DATABASE!!");
    bak
}

#[test]
fn apply_keeps_every_id_and_row_and_sets_the_damaged_file_aside() {
    let w = World::new();
    let reference = w.copy_as("reference.db");
    let ref_conn =
        Connection::open_with_flags(&reference, rusqlite::OpenFlags::SQLITE_OPEN_READ_ONLY)
            .unwrap();
    let bak = break_ledger(&w);
    assert_eq!(bak, sibling(&w.ledger(), ".bak"));
    let damaged_bytes_before = fs::read(w.ledger()).unwrap();
    // The damaged ledger refuses to open.
    assert!(matches!(
        ledger::open_read_only(w.home.path(), PROJECT),
        Err(LedgerError::Unavailable(_))
    ));
    let obs = confirm_all(&w, &bak);
    let plan = plan_of(&w, &bak, &obs);
    assert_eq!(plan.target_state, recover::TargetState::Unavailable);
    let result = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap();
    assert_eq!(result.state, ApplyState::Applied);
    let built = result.built.as_ref().unwrap();
    assert_eq!(built.integrity_check, "ok");
    assert_eq!(built.foreign_key_violations, 0);
    assert!(!built.replaced_unfinished);
    let set_aside = result.set_aside.as_ref().unwrap();
    let kept = sibling(&w.ledger(), &format!(".damaged-{SUFFIX}"));
    assert_eq!(set_aside.damaged.as_ref(), Some(&kept));
    assert!(!set_aside.reused);
    assert_eq!(
        fs::read(&kept).unwrap(),
        damaged_bytes_before,
        "the damaged file is kept byte for byte"
    );
    assert_eq!(
        TempHome::names_in(w.ledger().parent().unwrap()),
        vec![
            "ledger.db".to_string(),
            "ledger.db.bak".to_string(),
            format!("ledger.db.damaged-{SUFFIX}"),
        ],
        "no scratch file is left"
    );

    // Ids and rows are the source's, table by table.
    let new = w.reader();
    for (table, cols, order) in [
        ("card", CARD_COLS, "card_id"),
        ("card_ac", AC_COLS, "ac_id"),
        ("card_note", NOTE_COLS, "note_id"),
        ("card_link", LINK_COLS, "link_id"),
    ] {
        assert_eq!(
            dump(new.conn(), table, cols, order),
            dump(&ref_conn, table, cols, order),
            "{table}"
        );
    }
    let old_events = dump(&ref_conn, "event", EVENT_COLS, "event_id");
    let new_events = dump(new.conn(), "event", EVENT_COLS, "event_id");
    assert_eq!(new_events.len(), old_events.len() + 1);
    assert_eq!(&new_events[..old_events.len()], &old_events[..]);
    assert_eq!(
        dump(new.conn(), "reservation", RES_COLS, "reservation_id"),
        dump(&ref_conn, "reservation", RES_COLS, "reservation_id"),
        "all three were confirmed, so every reservation row is as it was"
    );
    assert_eq!(
        dump(
            new.conn(),
            "reservation_scope",
            "reservation_id, pattern",
            "reservation_id, pattern"
        ),
        dump(
            &ref_conn,
            "reservation_scope",
            "reservation_id, pattern",
            "reservation_id, pattern"
        ),
    );
    // The id counters: the next id is the highest plus one.
    for table in [
        "card",
        "card_ac",
        "card_note",
        "card_link",
        "reservation",
        "event",
    ] {
        assert_eq!(
            sequence(new.conn(), table),
            sequence(&ref_conn, table).map(|n| if table == "event" { n + 1 } else { n }),
            "{table}"
        );
    }
    let check: String = new
        .conn()
        .query_row("PRAGMA integrity_check", [], |r| r.get(0))
        .unwrap();
    assert_eq!(check, "ok");
    // The new file is a working ledger: the triggers are back.
    let mut writer = ledger::open_write(w.home.path(), PROJECT).unwrap();
    let blocked = writer.with_write_tx(|tx, _| {
        tx.execute("DELETE FROM card WHERE card_id = 1", [])?;
        Ok(())
    });
    assert!(
        matches!(blocked, Err(LedgerError::Constraint(_))),
        "{blocked:?}"
    );
    let next_card = add_card(&mut writer, "draft");
    assert_eq!(next_card, 7, "the next card number is the highest plus one");
}

#[test]
fn the_recover_event_is_written_once_and_a_second_apply_changes_nothing() {
    let w = World::new();
    let bak = break_ledger(&w);
    let obs = confirm_all(&w, &bak);
    let plan = plan_of(&w, &bak, &obs);
    let first = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap();
    assert_eq!(first.state, ApplyState::Applied);
    let after_first = w.home.snapshot();

    // Applying the same plan again, with the same or another suffix, answers `unchanged`.
    for suffix in [SUFFIX, "20261116T999999Z"] {
        let again = recover::apply_with_suffix(
            w.home.path(),
            PROJECT,
            &bak,
            &obs,
            &plan.plan_digest,
            suffix,
        )
        .unwrap();
        assert_eq!(again.state, ApplyState::Unchanged);
        assert!(again.built.is_none() && again.set_aside.is_none());
        assert_eq!(w.home.snapshot(), after_first, "nothing was written");
    }
    // With the source named as the ledger path (now the new ledger) the digest still matches
    // the recover event.
    let via_ledger = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &w.ledger(),
        &obs,
        &plan.plan_digest,
        "20261116T888888Z",
    )
    .unwrap();
    assert_eq!(via_ledger.state, ApplyState::Unchanged);
    assert_eq!(w.home.snapshot(), after_first);

    let r = w.reader();
    let events = recover_events(r.conn());
    assert_eq!(events.len(), 1, "one `recover` event");
    assert_eq!(events[0]["plan_digest"], plan.plan_digest);
    assert_eq!(events[0]["source"], bak.display().to_string());
    assert_eq!(events[0]["restore_held"], 3);
    assert_eq!(events[0]["decisions"].as_array().unwrap().len(), 3);
    let keys: Vec<String> = r
        .conn()
        .prepare("SELECT event_key FROM event")
        .unwrap()
        .query_map([], |row| row.get(0))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    let distinct: std::collections::BTreeSet<_> = keys.iter().collect();
    assert_eq!(keys.len(), distinct.len(), "event keys are unique");
    assert!(keys.contains(&format!("recover:{}", plan.plan_digest)));
    assert_eq!(
        first.built.unwrap().recover_event_id,
        Some(count(r.conn(), "event"))
    );
}

#[test]
fn event_keys_survive_the_copy_so_a_replay_adds_nothing() {
    let w = World::new();
    let reference = w.copy_as("reference.db");
    let bak = break_ledger(&w);
    let plan = plan_of(&w, &bak, &Observations::default());
    recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &Observations::default(),
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap();
    let ref_conn = Connection::open(&reference).unwrap();
    let old: Vec<String> = ref_conn
        .prepare("SELECT event_key FROM event ORDER BY event_id")
        .unwrap()
        .query_map([], |r| r.get(0))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    let r = w.reader();
    let new: Vec<String> = r
        .conn()
        .prepare("SELECT event_key FROM event ORDER BY event_id")
        .unwrap()
        .query_map([], |row| row.get(0))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    assert_eq!(&new[..old.len()], &old[..], "the keys are the source's");
    assert!(old.contains(&"import:github:o/r:12:x".to_string()));
    // Replaying an event whose key is held adds nothing.
    let mut writer = ledger::open_write(w.home.path(), PROJECT).unwrap();
    let before = count(writer.conn(), "event");
    sql(
        &mut writer,
        "INSERT INTO event (event_key, at, actor, kind) VALUES ('import:github:o/r:12:x', 9, 'me', 'import') \
         ON CONFLICT(event_key) DO NOTHING;",
    );
    assert_eq!(count(writer.conn(), "event"), before);
}

#[test]
fn apply_refuses_a_stale_digest_a_healthy_ledger_and_a_used_suffix() {
    let w = World::new();
    let bak = break_ledger(&w);
    let obs = confirm_all(&w, &bak);
    let plan = plan_of(&w, &bak, &obs);

    // A digest that is not the recomputed one.
    let err = recover::apply_with_suffix(w.home.path(), PROJECT, &bak, &obs, "sha256:0", SUFFIX)
        .unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("stale_plan")),
        "{err:?}"
    );
    // The same source with other observations is another plan.
    let err = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &Observations::default(),
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("stale_plan")),
        "{err:?}"
    );
    // A source that changed after the preview.
    let moved = w.home.path().join("sources/moved.db");
    fs::create_dir_all(moved.parent().unwrap()).unwrap();
    fs::copy(&bak, &moved).unwrap();
    let p2 = plan_of(&w, &moved, &obs);
    assert_eq!(
        p2.plan_digest, plan.plan_digest,
        "the digest follows the content, not the path"
    );
    let writer = Connection::open(&moved).unwrap();
    writer
        .execute_batch(
            "PRAGMA foreign_keys = ON; UPDATE card SET title = 'changed' WHERE card_id = 5;",
        )
        .unwrap();
    drop(writer);
    let err = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &moved,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("stale_plan")),
        "{err:?}"
    );
    let before = w.home.snapshot();
    // A bad suffix.
    let long = "a".repeat(41);
    for bad in ["", "a-b", "../x", long.as_str()] {
        let err =
            recover::apply_with_suffix(w.home.path(), PROJECT, &bak, &obs, &plan.plan_digest, bad)
                .unwrap_err();
        assert!(
            matches!(&err, LedgerError::Constraint(m) if m.starts_with("invalid_suffix")),
            "{bad:?}: {err:?}"
        );
    }
    assert_eq!(
        w.home.snapshot(),
        before,
        "every refusal left the HOME as it was"
    );

    // A suffix whose kept name is taken by another file.
    let taken = sibling(&w.ledger(), &format!(".damaged-{SUFFIX}"));
    fs::write(&taken, b"somebody else's file").unwrap();
    let err = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("suffix_in_use")),
        "{err:?}"
    );
    assert_eq!(fs::read(&taken).unwrap(), b"somebody else's file");
    assert!(
        std::fs::symlink_metadata(sibling(&w.ledger(), ".recovering")).is_ok(),
        "the finished new ledger waits; only the install was refused"
    );
    // Another suffix goes through; the file that held the name stays.
    let ok = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &plan.plan_digest,
        "20261116T000001Z",
    )
    .unwrap();
    assert_eq!(ok.state, ApplyState::Applied);
    assert_eq!(fs::read(&taken).unwrap(), b"somebody else's file");

    // The ledger at the real name is now healthy and a different plan is refused outright.
    let other = recover::plan(w.home.path(), PROJECT, &bak, &Observations::default()).unwrap();
    assert_eq!(other.target_state, recover::TargetState::Healthy);
    let err = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &Observations::default(),
        &other.plan_digest,
        "20261116T000002Z",
    )
    .unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("target_healthy")),
        "{err:?}"
    );
}

#[test]
fn a_link_or_a_folder_at_the_ledger_path_or_the_scratch_name_is_never_replaced() {
    let w = World::new();
    let bak = break_ledger(&w);
    let obs = Observations::default();
    let plan = plan_of(&w, &bak, &obs);
    // A folder where the scratch file goes.
    let scratch = sibling(&w.ledger(), ".recovering");
    fs::create_dir_all(&scratch).unwrap();
    let err = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("recovering_not_regular")),
        "{err:?}"
    );
    fs::remove_dir(&scratch).unwrap();
    // A link at the ledger path.
    let target = w.home.path().join("elsewhere.db");
    fs::copy(&bak, &target).unwrap();
    fs::remove_file(w.ledger()).unwrap();
    std::os::unix::fs::symlink(&target, w.ledger()).unwrap();
    let p = plan_of(&w, &bak, &obs);
    assert_eq!(p.target_state, recover::TargetState::NotRegular);
    let err =
        recover::apply_with_suffix(w.home.path(), PROJECT, &bak, &obs, &p.plan_digest, SUFFIX)
            .unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("target_not_regular")),
        "{err:?}"
    );
    assert!(std::fs::symlink_metadata(w.ledger())
        .unwrap()
        .file_type()
        .is_symlink());
}

#[test]
fn an_unfinished_scratch_file_and_a_leftover_journal_of_the_damaged_file_are_handled() {
    let w = World::new();
    let bak = break_ledger(&w);
    let obs = confirm_all(&w, &bak);
    let plan = plan_of(&w, &bak, &obs);
    // A killed build left a half file and its journal at the scratch name.
    let scratch = sibling(&w.ledger(), ".recovering");
    fs::write(&scratch, vec![7u8; 5000]).unwrap();
    fs::write(sibling(&scratch, "-journal"), b"half").unwrap();
    // The damaged file has a journal of its own.
    fs::write(sibling(&w.ledger(), "-journal"), b"old journal").unwrap();
    let result = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap();
    assert!(result.built.as_ref().unwrap().replaced_unfinished);
    let set_aside = result.set_aside.unwrap();
    let kept_journal = sibling(&w.ledger(), &format!("-journal.damaged-{SUFFIX}"));
    assert_eq!(set_aside.journal.as_ref(), Some(&kept_journal));
    assert_eq!(fs::read(&kept_journal).unwrap(), b"old journal");
    assert!(
        !sibling(&w.ledger(), "-journal").exists(),
        "no journal of the old file is left to be rolled into the new one"
    );
    assert!(!scratch.exists() && !sibling(&scratch, "-journal").exists());
    let r = w.reader();
    assert_eq!(count(r.conn(), "card"), 6);
}

#[test]
fn a_recovery_into_a_folder_with_no_ledger_works_from_the_bak_copy() {
    let w = World::new();
    let bak = break_ledger(&w);
    fs::remove_file(w.ledger()).unwrap();
    let obs = Observations::default();
    let plan = plan_of(&w, &bak, &obs);
    assert_eq!(plan.target_state, recover::TargetState::Absent);
    let result = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap();
    assert_eq!(result.set_aside.unwrap(), recover::SetAside::default());
    assert!(kept_files(&w.home).is_empty(), "there was nothing to keep");
    assert_eq!(count(w.reader().conn(), "card"), 6);
}

// ---- the decision rule, end to end --------------------------------------------------------------

#[test]
fn only_a_reservation_confirmed_three_ways_comes_back_held() {
    let w = World::new();
    let bak = break_ledger(&w);
    let mut obs = confirm_all(&w, &bak);
    // alice: the journal lease ended. bob: the worktree is on another branch.
    // dave: the remote has no fact for his branch.
    obs.journal
        .iter_mut()
        .find(|j| j.holder == "alice")
        .unwrap()
        .live = false;
    obs.worktrees
        .iter_mut()
        .find(|t| t.path == "/wt/bob")
        .unwrap()
        .branch = Some("other".into());
    obs.remotes.retain(|r| r.holder.as_deref() != Some("dave"));
    let plan = plan_of(&w, &bak, &obs);
    assert_eq!((plan.restore_held, plan.held_for_review), (0, 3));
    let by_holder = |h: &str| plan.decisions.iter().find(|d| d.holder == h).unwrap();
    assert_eq!(
        by_holder("alice").reasons,
        vec![recover::REASON_JOURNAL_NOT_LIVE]
    );
    assert_eq!(
        by_holder("bob").reasons,
        vec![recover::REASON_BRANCH_MISMATCH]
    );
    assert_eq!(
        by_holder("dave").reasons,
        vec![recover::REASON_REMOTE_ABSENT]
    );
    recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap();
    let r = w.reader();
    assert_eq!(
        statuses(r.conn()),
        vec![
            (1, 1, HELD_FOR_REVIEW.to_string()),
            (2, 1, HELD_FOR_REVIEW.to_string()),
            (3, 1, "released".to_string()),
            (3, 2, HELD_FOR_REVIEW.to_string()),
            (4, 1, "released".to_string()),
        ],
        "released rows are copied as they are; the three active ones are held for review"
    );
    let events = recover_events(r.conn());
    assert_eq!(events[0]["held_for_review"], 3);
    assert_eq!(events[0]["restore_held"], 0);
}

#[test]
fn no_observation_at_all_holds_everything_active_for_review() {
    let w = World::new();
    let bak = break_ledger(&w);
    let plan = plan_of(&w, &bak, &Observations::default());
    assert_eq!((plan.restore_held, plan.held_for_review), (0, 3));
    for d in &plan.decisions {
        assert_eq!(
            d.reasons,
            vec!["journal_missing", "worktree_not_observed", "remote_absent"]
        );
    }
}

#[test]
fn after_a_recovery_a_held_for_review_card_refuses_a_new_reservation_and_a_restored_holder_passes_the_fence(
) {
    let w = World::new();
    let bak = break_ledger(&w);
    let mut obs = confirm_all(&w, &bak);
    // bob stays unconfirmed; alice and dave are confirmed.
    obs.journal.retain(|j| j.holder != "bob");
    let plan = plan_of(&w, &bak, &obs);
    assert_eq!((plan.restore_held, plan.held_for_review), (2, 1));
    recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap();
    let mut writer = ledger::open_write(w.home.path(), PROJECT).unwrap();

    // bob's card (2) is `active` with a `held_for_review` holder: nobody can take it.
    let err = reservation::reserve(&mut writer, 2, "mallory", &s(&["zone9/"]), 3600, None, None)
        .unwrap_err();
    // `reserve` wants a ready card first; the card is active, which is also a refusal.
    assert!(
        matches!(
            err,
            ReserveError::CardNotReady { .. } | ReserveError::AlreadyHeld { .. }
        ),
        "{err:?}"
    );
    // Even a card forced to `ready` cannot be taken: the review row occupies the holder slot.
    sql(
        &mut writer,
        "UPDATE card SET state = 'ready' WHERE card_id = 2",
    );
    let err = reservation::reserve(&mut writer, 2, "mallory", &s(&["zone9/"]), 3600, None, None)
        .unwrap_err();
    assert!(
        matches!(err, ReserveError::AlreadyHeld { ref holder, generation: 1 } if holder == "bob"),
        "{err:?}"
    );
    assert_eq!(
        reservation::current(writer.conn(), 2)
            .unwrap()
            .unwrap()
            .status,
        HELD_FOR_REVIEW
    );

    // alice (card 1, generation 1) was restored as `held`: the fence passes at the right
    // generation and refuses a wrong one, and the refusal is recorded.
    reservation::fence(&mut writer, 1, "alice", 1).unwrap();
    let before = count(writer.conn(), "event");
    let err = reservation::fence(&mut writer, 1, "alice", 2).unwrap_err();
    assert!(
        matches!(
            err,
            FenceError::UnknownGeneration {
                offered: 2,
                current: 1
            }
        ),
        "{err:?}"
    );
    assert_eq!(count(writer.conn(), "event"), before + 1);
    // dave holds card 3 at generation 2; generation 1 (carol's) is stale.
    reservation::fence(&mut writer, 3, "dave", 2).unwrap();
    let err = reservation::fence(&mut writer, 3, "carol", 1).unwrap_err();
    assert!(
        matches!(
            err,
            FenceError::StaleGeneration {
                offered: 1,
                current: 2
            }
        ),
        "{err:?}"
    );
    // bob's review row is not `held`: the fence says so.
    let err = reservation::fence(&mut writer, 2, "bob", 1).unwrap_err();
    assert!(matches!(err, FenceError::NotHeld { .. }), "{err:?}");
    // A new generation is never issued twice: the card kept its counter.
    let card_generation: i64 = writer
        .conn()
        .query_row("SELECT generation FROM card WHERE card_id = 3", [], |r| {
            r.get(0)
        })
        .unwrap();
    assert_eq!(card_generation, 2);
}

#[test]
fn a_cards_generation_is_raised_to_the_highest_reservation_generation() {
    let w = World::new();
    // The card row says 0 but the ledger holds a generation-2 reservation: the card row was
    // rolled back by damage. The restored card must not issue generation 2 again.
    let lowered = w.copy_as("lowered.db");
    let conn = Connection::open(&lowered).unwrap();
    conn.execute_batch("UPDATE card SET generation = 0 WHERE card_id = 3;")
        .unwrap();
    drop(conn);
    let obs = confirm_all(&w, &lowered);
    let plan = plan_of(&w, &lowered, &obs);
    assert_eq!(plan.card_generations_raised, 1);
    let dave = plan.decisions.iter().find(|d| d.holder == "dave").unwrap();
    assert_eq!(
        dave.decision, RESTORE_HELD,
        "equal to the restored generation is not below it"
    );
    fs::remove_file(w.ledger()).unwrap();
    recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &lowered,
        &obs,
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap();
    let generation: i64 = w
        .reader()
        .conn()
        .query_row("SELECT generation FROM card WHERE card_id = 3", [], |r| {
            r.get(0)
        })
        .unwrap();
    assert_eq!(generation, 2);
}

#[test]
fn a_reservation_below_its_cards_restored_generation_is_held_for_review() {
    let w = World::new();
    let source = w.copy_as("stale.db");
    // Card 1 has a generation-5 reservation row `released` and an old active one at generation 1:
    // the active one was superseded.
    let conn = Connection::open(&source).unwrap();
    conn.execute_batch("UPDATE card SET generation = 5 WHERE card_id = 1;")
        .unwrap();
    drop(conn);
    let obs = confirm_all(&w, &source);
    let plan = plan_of(&w, &source, &obs);
    let alice = plan.decisions.iter().find(|d| d.holder == "alice").unwrap();
    assert_eq!(alice.decision, HELD_FOR_REVIEW);
    assert_eq!(alice.reasons, vec![recover::REASON_GENERATION_BELOW_CARD]);
    let bob = plan.decisions.iter().find(|d| d.holder == "bob").unwrap();
    assert_eq!(bob.decision, RESTORE_HELD, "other cards are not affected");
}

// ---- rows the schema refuses --------------------------------------------------------------------

#[test]
fn a_row_the_schema_refuses_is_named_and_left_in_the_source() {
    let w = World::new();
    let source = w.copy_as("refused.db");
    let conn = Connection::open(&source).unwrap();
    // Remove the delete trigger on the copy only, then lose card 5: its criterion, note and link
    // lose their parent.
    conn.execute_batch(
        "DROP TRIGGER card_no_delete; PRAGMA foreign_keys = OFF; DELETE FROM card WHERE card_id = 5;",
    )
    .unwrap();
    drop(conn);
    let before = w.home.snapshot();
    let plan = plan_of(&w, &source, &Observations::default());
    assert_eq!(w.home.snapshot(), before);
    let refused: Vec<(&str, &str)> = plan
        .rejected
        .iter()
        .map(|r| (r.table.as_str(), r.id.as_str()))
        .collect();
    assert!(refused.contains(&("card_ac", "1")), "{refused:?}");
    assert!(refused.contains(&("card_note", "1")), "{refused:?}");
    assert!(refused.contains(&("card_link", "1")), "{refused:?}");
    assert!(
        refused.iter().any(|(t, _)| *t == "event"),
        "the import event of card 5: {refused:?}"
    );
    assert_eq!(plan.rejected_total as usize, plan.rejected.len());
    assert!(plan
        .rejected
        .iter()
        .all(|r| r.reason.contains("FOREIGN KEY")));
    let ac = plan.tables.iter().find(|t| t.table == "card_ac").unwrap();
    assert_eq!((ac.rows_read, ac.copied, ac.rejected), (1, 0, 1));
    // The same refusals in the real build, and the result still passes its checks.
    fs::remove_file(w.ledger()).unwrap();
    let result = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &source,
        &Observations::default(),
        &plan.plan_digest,
        SUFFIX,
    )
    .unwrap();
    assert_eq!(result.built.unwrap().integrity_check, "ok");
    let r = w.reader();
    assert_eq!(count(r.conn(), "card_ac"), 0);
    assert_eq!(count(r.conn(), "card"), 5);
}

#[test]
fn the_rejected_list_is_capped_but_its_count_is_not() {
    for n in [
        recover::MAX_LISTED_REJECTED,
        recover::MAX_LISTED_REJECTED + 1,
    ] {
        let w = World::new();
        let source = w.copy_as("many.db");
        let conn = Connection::open(&source).unwrap();
        // Notes whose card does not exist.
        conn.execute_batch(&format!(
            "PRAGMA foreign_keys = OFF;
             WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n WHERE i < {n})
             INSERT INTO card_note (card_id, kind, author, body, created_at) SELECT 999, 'comment', 'x', 'b', i FROM n;"
        ))
        .unwrap();
        drop(conn);
        let plan = plan_of(&w, &source, &Observations::default());
        assert_eq!(
            plan.rejected_total as usize, n,
            "every refused row is counted"
        );
        assert_eq!(
            plan.rejected.len(),
            n.min(recover::MAX_LISTED_REJECTED),
            "at most {} are listed",
            recover::MAX_LISTED_REJECTED
        );
    }
}

// ---- leftovers of register ----------------------------------------------------------------------

/// A pid that belonged to a process that has since exited.
fn dead_pid() -> u32 {
    let mut child = std::process::Command::new(std::env::current_exe().unwrap())
        .args(["--exact", "no_such_test_exists", "--nocapture"])
        .clean_env()
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()
        .unwrap();
    let pid = child.id();
    child.wait().unwrap();
    pid
}

#[test]
fn stale_register_leftovers_are_reported_and_never_removed() {
    let w = World::new();
    let dir = w.ledger().parent().unwrap().to_path_buf();
    let dead = dead_pid();
    let own = std::process::id();
    let stale_name = format!(".creating-{dead}-0");
    for name in [
        stale_name.clone(),
        format!(".creating-{own}-0"),
        ".creating-x-1".to_string(),
        ".creating-1".to_string(),
        ".creating-99999999999-1".to_string(),
    ] {
        fs::write(dir.join(name), b"half a ledger").unwrap();
    }
    let source = w.copy_as("damaged.db");
    let before = w.home.snapshot();
    let stale = recover::stale_temp(w.home.path(), PROJECT).unwrap();
    assert_eq!(
        stale,
        vec![recover::StaleTemp {
            name: stale_name.clone(),
            pid: i64::from(dead)
        }],
        "only a name with a dead pid; the live pid, the malformed names and the oversized pid are not stale"
    );
    // The plan reports them too, and removes nothing.
    let plan = recover::plan(w.home.path(), PROJECT, &source, &Observations::default()).unwrap();
    assert_eq!(plan.stale_temp, stale);
    assert!(dir.join(&stale_name).exists());
    // A project with no folder has none.
    assert_eq!(
        recover::stale_temp(w.home.path(), "nowhere").unwrap(),
        vec![]
    );
    assert_eq!(
        recover::stale_temp(w.home.path(), "../x"),
        Err(LedgerError::InvalidProjectId)
    );
    assert_eq!(w.home.snapshot(), before, "reporting changed nothing");
}

// ---- a writer killed mid-transaction ------------------------------------------------------------

const KILL_ENV: &str = "MOAI_RECOVER_KILL_DIR";
const STEP_ENV: &str = "MOAI_RECOVER_STEP";

fn wait_for(path: &Path, what: &str) {
    let deadline = Instant::now() + Duration::from_secs(60);
    while !path.exists() {
        assert!(Instant::now() < deadline, "timed out waiting for {what}");
        std::thread::sleep(Duration::from_millis(1));
    }
}

fn spawn_self(
    test: &str,
    home: &TempHome,
    env: &[(&str, &Path)],
    extra: &[(&str, &str)],
) -> std::process::Child {
    let mut cmd = std::process::Command::new(std::env::current_exe().unwrap());
    cmd.args(["--exact", test, "--nocapture", "--test-threads=1"])
        .clean_env()
        .env("HOME", home.path())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped());
    for (k, v) in env {
        cmd.env(k, v);
    }
    for (k, v) in extra {
        cmd.env(k, v);
    }
    cmd.spawn().unwrap()
}

/// The child of the hot-journal measurement: starts a write transaction, writes enough (with a
/// tiny page cache) that the changes and the journal reach the files, and waits to be killed.
#[test]
fn hot_journal_child() {
    let Some(dir) = std::env::var_os(KILL_ENV).map(PathBuf::from) else {
        return;
    };
    if std::env::var(STEP_ENV).as_deref() != Ok("hot_journal") {
        return;
    }
    let home = PathBuf::from(std::env::var_os("HOME").unwrap());
    let mut w = ledger::open_write(&home, PROJECT).unwrap();
    let _: Result<(), LedgerError> = w.with_write_tx(|tx, _| {
        tx.execute_batch("PRAGMA cache_size = 10;")?;
        tx.execute_batch(
            "WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n WHERE i < 4000) \
             INSERT INTO event (event_key, at, actor, kind, payload) \
             SELECT 'doomed:' || i, 1, 'doomed', 'filler', json_object('pad', printf('%0400d', i)) FROM n;",
        )?;
        fs::write(dir.join("in_tx"), "").unwrap();
        loop {
            std::thread::sleep(Duration::from_millis(50));
        }
    });
}

/// P1 leftover (ii), measured: a writer killed inside its transaction leaves a hot journal. A
/// read-only open then reports `ledger_unavailable` with `needs_recovery`; the next writer rolls
/// the journal back and the ledger is whole again with no recovery needed.
#[test]
fn a_writer_killed_mid_transaction_leaves_a_hot_journal_that_read_only_reports_as_needs_recovery() {
    let w = World::new();
    let events_before = count(w.reader().conn(), "event");
    let signals = tempfile::tempdir().unwrap();
    let mut child = spawn_self(
        "hot_journal_child",
        &w.home,
        &[(KILL_ENV, signals.path())],
        &[(STEP_ENV, "hot_journal")],
    );
    wait_for(
        &signals.path().join("in_tx"),
        "the child inside its transaction",
    );
    child.kill().unwrap();
    assert!(!child.wait().unwrap().success());
    let journal = sibling(&w.ledger(), "-journal");
    assert!(journal.exists(), "the killed writer left its journal");

    // What P1 returns, pinned.
    let err = ledger::open_read_only(w.home.path(), PROJECT)
        .err()
        .unwrap();
    assert!(
        matches!(&err, LedgerError::Unavailable(m) if m.starts_with("needs_recovery")),
        "{err:?}"
    );
    // The plan refuses the same file as a source with the same word, and writes nothing.
    let before = w.home.snapshot();
    let err = recover::plan(
        w.home.path(),
        PROJECT,
        &w.ledger(),
        &Observations::default(),
    )
    .unwrap_err();
    assert!(
        matches!(&err, LedgerError::Unavailable(m) if m.contains("needs_recovery")),
        "{err:?}"
    );
    assert_eq!(w.home.snapshot(), before);

    // The next writer rolls it back; no event of the dead writer remains.
    let writer = ledger::open_write(w.home.path(), PROJECT).unwrap();
    assert_eq!(count(writer.conn(), "event"), events_before);
    drop(writer);
    assert!(!journal.exists());
    assert_eq!(w.reader().schema_version(), 1);
}

// ---- killing the applying process ---------------------------------------------------------------

/// The child of the kill-point tests: runs the apply's steps one by one through the library's
/// public functions, up to the step named by `STEP_ENV`, announces it, and waits to be killed.
/// The kill points live here, in the test binary, not in the library.
#[test]
fn apply_child() {
    let Some(dir) = std::env::var_os(KILL_ENV).map(PathBuf::from) else {
        return;
    };
    let Ok(stop_after) = std::env::var(STEP_ENV) else {
        return;
    };
    if stop_after == "hot_journal" {
        return;
    }
    let home = PathBuf::from(std::env::var_os("HOME").unwrap());
    let source = PathBuf::from(fs::read_to_string(dir.join("source")).unwrap());
    let digest = fs::read_to_string(dir.join("digest")).unwrap();
    let obs: Observations =
        serde_json::from_str(&fs::read_to_string(dir.join("obs")).unwrap()).unwrap();
    let park = |step: &str| {
        if stop_after == step {
            fs::write(dir.join("parked"), step).unwrap();
            loop {
                std::thread::sleep(Duration::from_millis(50));
            }
        }
    };
    park("before_build");
    recover::build_recovering(&home, PROJECT, &source, &obs, &digest).unwrap();
    park("after_build");
    recover::set_aside_damaged(&home, PROJECT, "20261116T000001Z").unwrap();
    park("after_set_aside");
    recover::install_recovered(&home, PROJECT).unwrap();
    park("after_install");
}

/// What a normal open sees after a kill: the old state or the new state, never an absent ledger.
fn old_or_new(w: &World) -> &'static str {
    match ledger::open_read_only(w.home.path(), PROJECT) {
        Ok(_) => "new",
        Err(LedgerError::Unavailable(_)) => "old",
        Err(other) => panic!("neither the old nor the new state: {other:?}"),
    }
}

fn kill_at(step: &str) -> (World, PathBuf, Observations, String, Vec<u8>) {
    let w = World::new();
    let bak = break_ledger(&w);
    let obs = confirm_all(&w, &bak);
    let plan = plan_of(&w, &bak, &obs);
    let old_bytes = fs::read(w.ledger()).unwrap();
    let signals = tempfile::tempdir().unwrap();
    fs::write(signals.path().join("source"), bak.display().to_string()).unwrap();
    fs::write(signals.path().join("digest"), &plan.plan_digest).unwrap();
    fs::write(
        signals.path().join("obs"),
        serde_json::to_string(&obs).unwrap(),
    )
    .unwrap();
    let mut child = spawn_self(
        "apply_child",
        &w.home,
        &[(KILL_ENV, signals.path())],
        &[(STEP_ENV, step)],
    );
    wait_for(
        &signals.path().join("parked"),
        "the child at its kill point",
    );
    child.kill().unwrap();
    assert!(!child.wait().unwrap().success(), "killed, not finished");
    (w, bak, obs, plan.plan_digest, old_bytes)
}

#[test]
fn a_kill_before_the_build_changes_nothing_and_a_rerun_converges() {
    let (w, bak, obs, digest, old) = kill_at("before_build");
    assert_eq!(old_or_new(&w), "old");
    assert_eq!(fs::read(w.ledger()).unwrap(), old);
    assert!(!sibling(&w.ledger(), ".recovering").exists());
    assert!(kept_files(&w.home).is_empty());
    let r =
        recover::apply_with_suffix(w.home.path(), PROJECT, &bak, &obs, &digest, SUFFIX).unwrap();
    assert_eq!(r.state, ApplyState::Applied);
    assert_eq!(old_or_new(&w), "new");
}

#[test]
fn a_kill_after_the_build_leaves_the_old_ledger_and_a_rerun_converges() {
    let (w, bak, obs, digest, old) = kill_at("after_build");
    assert_eq!(
        old_or_new(&w),
        "old",
        "the real name still holds the damaged file"
    );
    assert_eq!(fs::read(w.ledger()).unwrap(), old);
    let scratch = sibling(&w.ledger(), ".recovering");
    assert!(scratch.exists(), "the finished new file waits");
    assert!(kept_files(&w.home).is_empty());
    let r =
        recover::apply_with_suffix(w.home.path(), PROJECT, &bak, &obs, &digest, SUFFIX).unwrap();
    assert_eq!(r.state, ApplyState::Applied);
    assert!(
        r.built.unwrap().replaced_unfinished,
        "the earlier scratch copy was replaced"
    );
    assert_eq!(old_or_new(&w), "new");
    assert_eq!(
        fs::read(sibling(&w.ledger(), &format!(".damaged-{SUFFIX}"))).unwrap(),
        old
    );
}

#[test]
fn a_kill_between_the_set_aside_and_the_install_never_leaves_the_name_empty() {
    let (w, bak, obs, digest, old) = kill_at("after_set_aside");
    // The damaged file has two names; the real name still points at it.
    assert_eq!(old_or_new(&w), "old");
    assert_eq!(fs::read(w.ledger()).unwrap(), old);
    let kept = sibling(&w.ledger(), ".damaged-20261116T000001Z");
    assert_eq!(fs::read(&kept).unwrap(), old);
    assert!(sibling(&w.ledger(), ".recovering").exists());
    // A rerun with another suffix reuses the existing kept name instead of making a second one.
    let r = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &digest,
        "20261116T000002Z",
    )
    .unwrap();
    assert_eq!(r.state, ApplyState::Applied);
    let set_aside = r.set_aside.unwrap();
    assert!(set_aside.reused);
    assert_eq!(set_aside.damaged.as_ref(), Some(&kept));
    assert_eq!(kept_files(&w.home).len(), 1, "one kept file, not two");
    assert_eq!(old_or_new(&w), "new");
    assert_eq!(fs::read(&kept).unwrap(), old);
}

#[test]
fn a_kill_after_the_install_leaves_the_new_ledger_and_a_rerun_changes_nothing() {
    let (w, bak, obs, digest, old) = kill_at("after_install");
    assert_eq!(old_or_new(&w), "new");
    assert_eq!(
        fs::read(sibling(&w.ledger(), ".damaged-20261116T000001Z")).unwrap(),
        old
    );
    let before = w.home.snapshot();
    let r = recover::apply_with_suffix(
        w.home.path(),
        PROJECT,
        &bak,
        &obs,
        &digest,
        "20261116T000002Z",
    )
    .unwrap();
    assert_eq!(r.state, ApplyState::Unchanged);
    assert_eq!(w.home.snapshot(), before);
    assert_eq!(recover_events(w.reader().conn()).len(), 1);
}

#[test]
fn installing_never_loses_a_file_that_has_no_other_name() {
    let w = World::new();
    let bak = break_ledger(&w);
    let obs = Observations::default();
    let plan = plan_of(&w, &bak, &obs);
    recover::build_recovering(w.home.path(), PROJECT, &bak, &obs, &plan.plan_digest).unwrap();
    let err = recover::install_recovered(w.home.path(), PROJECT).unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("damaged_not_set_aside")),
        "{err:?}"
    );
    assert_eq!(old_or_new(&w), "old");
    // Nothing to install.
    fs::remove_file(sibling(&w.ledger(), ".recovering")).unwrap();
    let err = recover::install_recovered(w.home.path(), PROJECT).unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("nothing_to_install")),
        "{err:?}"
    );
}

// ---- a built file is checked --------------------------------------------------------------------

#[test]
fn a_finished_file_that_fails_its_checks_is_not_accepted() {
    let w = World::new();
    let copy = w.copy_as("check.db");
    assert_eq!(
        recover::check_built(&copy, PROJECT).unwrap(),
        ("ok".to_string(), 0)
    );
    // A foreign key violation made with enforcement off.
    let conn = Connection::open(&copy).unwrap();
    conn.execute_batch("PRAGMA foreign_keys = OFF; INSERT INTO card_note (card_id, kind, author, body, created_at) VALUES (4242, 'comment', 'x', 'b', 1);").unwrap();
    drop(conn);
    let err = recover::check_built(&copy, PROJECT).unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("check_failed") && m.contains("1 foreign key")),
        "{err:?}"
    );
}

// ---- the library keeps to its place -------------------------------------------------------------

/// The same names the repository scan of the isolation slice forbids in `src/ledger/`.
#[test]
fn the_recovery_library_spawns_nothing_and_reaches_no_network_or_forge() {
    let text =
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("src/ledger/recover.rs"))
            .unwrap();
    let code: String = text
        .lines()
        .filter(|l| !l.trim_start().starts_with("//"))
        .collect::<Vec<_>>()
        .join("\n");
    for name in [
        "Command::new(",
        "std::net",
        "\"git\"",
        "\"gh\"",
        "http://",
        "https://",
        "crate::forge",
        "moai_cli::forge",
    ] {
        assert!(!code.contains(name), "src/ledger/recover.rs holds `{name}`");
    }
}
