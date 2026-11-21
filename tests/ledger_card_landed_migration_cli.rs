//! An existing ledger gets the `card.landed` exception in its `card_done_guard` trigger
//! (SPEC-20261121-ledger-card-landed-migration). Each test builds an old ledger in its own
//! temporary HOME: a ledger registered by this build whose `card_done_guard` is then put back to
//! the text the schema had before the exception existed. None touches the real HOME.

mod common;

use common::home::TempHome;
use moai_cli::ledger::{self, migrate, LedgerError};
use rusqlite::{Connection, OpenFlags};
use std::fs;
use std::path::{Path, PathBuf};

/// `card_done_guard` as every ledger made before the M-Flow exception holds it.
const OLD_GUARD: &str = r"
CREATE TRIGGER card_done_guard BEFORE UPDATE OF state ON card
WHEN NEW.state = 'done' AND OLD.state <> 'done'
BEGIN
  SELECT RAISE(ABORT, 'cannot close: the card has no acceptance criterion that is met')
   WHERE NOT EXISTS (SELECT 1 FROM card_ac WHERE card_id = NEW.card_id AND status = 'met');
  SELECT RAISE(ABORT, 'cannot close: an acceptance criterion is still open')
   WHERE EXISTS (SELECT 1 FROM card_ac WHERE card_id = NEW.card_id AND status = 'open');
  SELECT RAISE(ABORT, 'cannot close: no merged pull request is linked')
   WHERE NOT EXISTS (SELECT 1 FROM card_link
                      WHERE from_card = NEW.card_id AND kind = 'pr'
                        AND state = 'active' AND pr_state = 'merged');
END;
";

/// Cards of the fixture: 1 has a landed event, 2 a merged pull request, 3 neither.
const ROWS: &str = r#"
INSERT INTO card (card_id, title, state, created_by, created_at, updated_at)
  VALUES (1, 'landed', 'review', 't', 1, 1), (2, 'merged', 'review', 't', 1, 1),
         (3, 'neither', 'review', 't', 1, 1), (4, 'ready', 'ready', 't', 1, 1);
INSERT INTO card_ac (card_id, seq, text, status, evidence_ref, updated_at)
  VALUES (1, 1, 'a', 'met', 'e1', 1), (2, 1, 'a', 'met', 'e2', 1), (3, 1, 'a', 'met', 'e3', 1);
INSERT INTO card_note (card_id, kind, author, body, created_at)
  VALUES (4, 'comment', 'someone', 'a note that must survive', 1);
INSERT INTO card_link (from_card, kind, target_host, target_project, target_number, pr_state,
                       merge_ref, created_at, updated_at)
  VALUES (2, 'pr', 'github', 'acme/repo', 7, 'merged', 'abc', 1, 1);
INSERT INTO event (event_key, at, actor, card_id, kind, payload)
  VALUES ('card.landed:1:k', 1, 'lane', 1, 'card.landed', '{"head":"abc"}');
INSERT INTO reservation (card_id, generation, holder, status, acquired_at, lease_until, updated_at)
  VALUES (4, 1, 'h', 'released', 1, 2, 2);
"#;

const TABLES: [&str; 8] = [
    "card",
    "card_ac",
    "card_note",
    "card_link",
    "card_transition",
    "reservation",
    "reservation_scope",
    "event",
];

fn path_of(home: &TempHome) -> PathBuf {
    ledger::ledger_path(home.path(), "demo").expect("ledger path")
}

fn raw(path: &Path) -> Connection {
    let c = Connection::open(path).expect("open raw");
    c.execute_batch("PRAGMA foreign_keys = ON; PRAGMA recursive_triggers = ON;")
        .unwrap();
    c
}

/// A ledger of this build, its done guard put back to the old text and the fixture rows in it.
fn old_ledger() -> (TempHome, PathBuf) {
    let home = TempHome::new();
    ledger::register(home.path(), "demo").expect("register");
    let path = path_of(&home);
    let c = raw(&path);
    c.execute_batch("DROP TRIGGER card_done_guard;").unwrap();
    c.execute_batch(OLD_GUARD).unwrap();
    c.execute_batch(ROWS).unwrap();
    drop(c);
    (home, path)
}

fn guard_sql(c: &Connection) -> Option<String> {
    c.query_row(
        "SELECT sql FROM sqlite_master WHERE type = 'trigger' AND name = 'card_done_guard'",
        [],
        |r| r.get(0),
    )
    .ok()
}

/// The `card_done_guard` text a freshly registered ledger holds.
fn fresh_guard() -> String {
    let home = TempHome::new();
    ledger::register(home.path(), "demo").expect("register");
    guard_sql(&raw(&path_of(&home))).expect("fresh ledger has the guard")
}

/// Every row of every data table, in rowid order, as text.
fn dump(path: &Path) -> String {
    let c = Connection::open_with_flags(path, OpenFlags::SQLITE_OPEN_READ_ONLY).unwrap();
    let mut out = String::new();
    for table in TABLES {
        let mut stmt = c
            .prepare(&format!("SELECT * FROM {table} ORDER BY rowid"))
            .unwrap();
        let n = stmt.column_count();
        let mut rows = stmt.query([]).unwrap();
        while let Some(row) = rows.next().unwrap() {
            out.push_str(table);
            for i in 0..n {
                out.push_str(&format!("|{:?}", row.get_ref(i).unwrap()));
            }
            out.push('\n');
        }
    }
    out
}

fn backups(path: &Path) -> Vec<PathBuf> {
    let dir = path.parent().unwrap();
    let mut found: Vec<PathBuf> = fs::read_dir(dir)
        .unwrap()
        .filter_map(|e| e.ok())
        .map(|e| e.path())
        .filter(|p| {
            p.file_name()
                .unwrap()
                .to_string_lossy()
                .starts_with("ledger.db.bak-")
        })
        .collect();
    found.sort();
    found
}

fn close(path: &Path, card: i64) -> Result<usize, rusqlite::Error> {
    raw(path).execute(
        &format!("UPDATE card SET state = 'done', closed_at = 9 WHERE card_id = {card}"),
        [],
    )
}

fn meta(path: &Path, key: &str) -> Option<String> {
    raw(path)
        .query_row("SELECT value FROM meta WHERE key = ?1", [key], |r| r.get(0))
        .ok()
}

#[test]
fn the_fixture_is_really_an_old_ledger() {
    let (_home, path) = old_ledger();
    let err = close(&path, 1).expect_err("old guard refuses a landed card");
    assert!(
        err.to_string().contains("no merged pull request"),
        "{err}"
    );
    assert!(!guard_sql(&raw(&path)).unwrap().contains("card.landed"));
}

/// AC-1.
#[test]
fn migrates_an_old_ledger() {
    let fresh = fresh_guard();
    let (home, path) = old_ledger();
    let before = dump(&path);

    let w = ledger::open_write(home.path(), "demo").expect("open_write migrates");
    let upgraded = w
        .done_guard_upgrade()
        .expect("the open reports the upgrade")
        .to_path_buf();
    drop(w);

    let found = backups(&path);
    assert_eq!(found.len(), 1, "exactly one backup: {found:?}");
    assert_eq!(found[0], upgraded);
    assert_eq!(
        guard_sql(&raw(&path)).unwrap(),
        fresh,
        "the trigger text equals a fresh ledger's"
    );
    let note = meta(&path, "done_guard_landed").expect("the meta row");
    let name = found[0].file_name().unwrap().to_string_lossy().into_owned();
    assert!(note.ends_with(&format!("|{name}")), "{note} vs {name}");
    assert!(note.split('|').next().unwrap().parse::<i64>().unwrap() > 1_000_000_000);

    // The backup is the old ledger: same rows, old guard, same identity pragmas.
    let backup = Connection::open_with_flags(&found[0], OpenFlags::SQLITE_OPEN_READ_ONLY).unwrap();
    assert_eq!(dump(&found[0]), before);
    assert!(!guard_sql(&backup).unwrap().contains("card.landed"));
    let id: i64 = backup.query_row("PRAGMA application_id", [], |r| r.get(0)).unwrap();
    assert_eq!(id, ledger::APPLICATION_ID);
    let ver: i64 = backup.query_row("PRAGMA user_version", [], |r| r.get(0)).unwrap();
    assert_eq!(ver, ledger::SCHEMA_VERSION);
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let mode = fs::metadata(&found[0]).unwrap().permissions().mode() & 0o777;
        assert_eq!(mode, 0o600, "the backup is private like the ledger");
    }
}

/// AC-2.
#[test]
fn rows_and_close_rules() {
    let (home, path) = old_ledger();
    let before = dump(&path);
    drop(ledger::open_write(home.path(), "demo").expect("open_write migrates"));

    assert_eq!(dump(&path), before, "every row of every table is unchanged");
    assert_eq!(close(&path, 1).expect("a landed card closes"), 1);
    assert_eq!(close(&path, 2).expect("a merged-PR card closes"), 1);
    let err = close(&path, 3).expect_err("neither is refused");
    assert!(err.to_string().contains("no merged pull request"), "{err}");
    let state: String = raw(&path)
        .query_row("SELECT state FROM card WHERE card_id = 3", [], |r| r.get(0))
        .unwrap();
    assert_eq!(state, "review");
}

/// AC-3.
#[test]
fn second_open_is_a_noop() {
    let (home, path) = old_ledger();
    drop(ledger::open_write(home.path(), "demo").expect("first open migrates"));
    let bytes = fs::read(&path).unwrap();
    let note = meta(&path, "done_guard_landed");

    let w = ledger::open_write(home.path(), "demo").expect("second open");
    assert!(w.done_guard_upgrade().is_none(), "nothing was upgraded");
    drop(w);

    assert_eq!(backups(&path).len(), 1, "no second backup");
    assert_eq!(fs::read(&path).unwrap(), bytes, "the file bytes are unchanged");
    assert_eq!(meta(&path, "done_guard_landed"), note);
}

/// AC-4.
#[test]
fn failure_rolls_back() {
    let (_home, path) = old_ledger();
    let before = dump(&path);
    let guard_before = guard_sql(&raw(&path)).unwrap();

    let mut conn = raw(&path);
    let result = migrate::upgrade_done_guard_with(&mut conn, &path, |tx| {
        // The trigger is already dropped here: prove it, then fail.
        let n: i64 = tx
            .query_row(
                "SELECT count(*) FROM sqlite_master WHERE name = 'card_done_guard'",
                [],
                |r| r.get(0),
            )
            .unwrap();
        assert_eq!(n, 0, "the hook runs after the drop");
        Err(LedgerError::Constraint("injected failure".into()))
    });
    assert_eq!(
        result,
        Err(LedgerError::Constraint("injected failure".into()))
    );
    drop(conn);

    assert_eq!(guard_sql(&raw(&path)).unwrap(), guard_before);
    assert_eq!(dump(&path), before);
    assert_eq!(meta(&path, "done_guard_landed"), None);
    assert!(backups(&path).is_empty(), "a failed run leaves no backup");
    let check: String = raw(&path)
        .query_row("PRAGMA integrity_check", [], |r| r.get(0))
        .unwrap();
    assert_eq!(check, "ok");
    // The file still opens for writing and migrates on the next try.
    let home_path = path.ancestors().nth(4).unwrap().to_path_buf();
    drop(ledger::open_write(&home_path, "demo").expect("a later open migrates"));
    assert!(guard_sql(&raw(&path)).unwrap().contains("card.landed"));
}

/// AC-5.
#[cfg(unix)]
#[test]
fn backup_failure_refuses() {
    use std::os::unix::fs::PermissionsExt;
    // A folder nobody can write to also stops root-less tests only; skip as root.
    if unsafe { libc::geteuid() } == 0 {
        return;
    }
    let (home, path) = old_ledger();
    let before = dump(&path);
    let guard_before = guard_sql(&raw(&path)).unwrap();
    let dir = path.parent().unwrap().to_path_buf();
    fs::set_permissions(&dir, fs::Permissions::from_mode(0o500)).unwrap();

    let result = ledger::open_write(home.path(), "demo");

    fs::set_permissions(&dir, fs::Permissions::from_mode(0o700)).unwrap();
    assert!(result.is_err(), "no backup, no migration");
    assert_eq!(guard_sql(&raw(&path)).unwrap(), guard_before);
    assert_eq!(dump(&path), before);
    assert_eq!(meta(&path, "done_guard_landed"), None);
    assert!(backups(&path).is_empty());
}

/// REQ-2: a backup that does not read back is not trusted.
#[test]
fn an_existing_backup_name_is_never_overwritten() {
    let (home, path) = old_ledger();
    let before = dump(&path);
    // Take the clock the migration will read, then occupy every second around it.
    let now: i64 = raw(&path)
        .query_row("SELECT unixepoch()", [], |r| r.get(0))
        .unwrap();
    let stamp = |t: i64| -> String {
        raw(&path)
            .query_row(
                "SELECT strftime('%Y%m%dT%H%M%SZ', ?1, 'unixepoch')",
                [t],
                |r| r.get(0),
            )
            .unwrap()
    };
    let dir = path.parent().unwrap();
    let mut planted = Vec::new();
    for t in (now - 1)..=(now + 30) {
        let file = dir.join(format!("ledger.db.bak-{}", stamp(t)));
        fs::write(&file, b"planted").unwrap();
        planted.push(file);
    }

    let result = ledger::open_write(home.path(), "demo");

    assert!(result.is_err(), "an existing backup name is refused");
    for file in &planted {
        assert_eq!(fs::read(file).unwrap(), b"planted", "{file:?} untouched");
    }
    assert!(!guard_sql(&raw(&path)).unwrap().contains("card.landed"));
    assert_eq!(dump(&path), before);
}

/// AC-6.
#[test]
fn status_is_read_only() {
    let (home, path) = old_ledger();
    let bytes = fs::read(&path).unwrap();

    let r = ledger::open_read_only(home.path(), "demo").expect("read-only open");
    assert_eq!(
        migrate::done_guard_status(r.conn()),
        Ok(migrate::DoneGuard::NeedsMigration)
    );
    drop(r);
    assert_eq!(fs::read(&path).unwrap(), bytes, "reading changed nothing");
    assert!(backups(&path).is_empty());

    drop(ledger::open_write(home.path(), "demo").unwrap());
    let r = ledger::open_read_only(home.path(), "demo").unwrap();
    assert_eq!(
        migrate::done_guard_status(r.conn()),
        Ok(migrate::DoneGuard::Current)
    );
}

/// A ledger with no done guard at all gets the current one (the same rule: no exception, replace).
#[test]
fn an_absent_guard_gets_the_current_one() {
    let fresh = fresh_guard();
    let (home, path) = old_ledger();
    raw(&path).execute_batch("DROP TRIGGER card_done_guard;").unwrap();

    drop(ledger::open_write(home.path(), "demo").expect("open_write"));

    assert_eq!(guard_sql(&raw(&path)).unwrap(), fresh);
    assert_eq!(backups(&path).len(), 1);
}

/// A fresh ledger needs nothing: no backup, no meta row, and the open reports no upgrade.
#[test]
fn a_fresh_ledger_is_left_alone() {
    let home = TempHome::new();
    ledger::register(home.path(), "demo").unwrap();
    let path = path_of(&home);
    let bytes = fs::read(&path).unwrap();

    let w = ledger::open_write(home.path(), "demo").unwrap();
    assert!(w.done_guard_upgrade().is_none());
    drop(w);

    assert!(backups(&path).is_empty());
    assert_eq!(meta(&path, "done_guard_landed"), None);
    assert_eq!(fs::read(&path).unwrap(), bytes);
}
