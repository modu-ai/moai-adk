//! Registering, opening and writing the project ledger (Issue #3068 P1, AC6/AC8 basics): what
//! creates a folder or a file and what must not, the refusals for a file that is not this
//! project's ledger, and a two-process race on the write lock. Every test runs in its own
//! temporary HOME; none touches the real `~/.moai`.

mod common;

use common::home::TempHome;
use common::CleanEnv;
use moai_cli::ledger::{self, LedgerError, Registered};
use std::fs;
use std::path::{Path, PathBuf};

fn path_of(home: &TempHome, id: &str) -> PathBuf {
    ledger::ledger_path(home.path(), id).unwrap()
}

fn registered(id: &str) -> (TempHome, PathBuf) {
    let home = TempHome::new();
    assert_eq!(
        ledger::register(home.path(), id).unwrap(),
        Registered::Created
    );
    let file = path_of(&home, id);
    (home, file)
}

/// The names in the ledger's folder.
fn names(file: &Path) -> Vec<String> {
    TempHome::names_in(file.parent().unwrap())
}

fn open_errors(home: &TempHome, id: &str) -> Vec<LedgerError> {
    vec![
        ledger::open_read_only(home.path(), id).err().unwrap(),
        ledger::open_write(home.path(), id).err().unwrap(),
    ]
}

// ---- register ---------------------------------------------------------------------------------

#[test]
fn register_creates_exactly_the_folders_and_one_file() {
    let home = TempHome::new();
    assert!(home.snapshot().is_empty());
    assert_eq!(
        ledger::register(home.path(), "demo").unwrap(),
        Registered::Created
    );
    let file = path_of(&home, "demo");
    let paths: Vec<String> = home.snapshot().into_iter().map(|e| e.path).collect();
    let file_rel = file.strip_prefix(home.path()).unwrap().to_string_lossy();
    assert_eq!(
        paths,
        vec![
            ".moai".to_string(),
            ".moai/projects".to_string(),
            ".moai/projects/demo".to_string(),
            file_rel.into_owned(),
        ]
    );
    // No journal, WAL or temporary file is left beside it.
    assert_eq!(names(&file).len(), 1);
    // And it opens, both ways.
    ledger::open_read_only(home.path(), "demo").unwrap();
    ledger::open_write(home.path(), "demo").unwrap();
}

#[test]
fn registering_twice_changes_nothing() {
    let (home, _) = registered("demo");
    let before = home.snapshot();
    assert_eq!(
        ledger::register(home.path(), "demo").unwrap(),
        Registered::AlreadyExists
    );
    assert_eq!(home.snapshot(), before, "bytes, entries and mtimes");
}

#[test]
fn two_projects_get_two_folders_in_one_projects_folder() {
    let (home, a) = registered("alpha");
    ledger::register(home.path(), "beta").unwrap();
    let projects = a.parent().unwrap().parent().unwrap();
    assert_eq!(TempHome::names_in(projects), vec!["alpha", "beta"]);
}

#[test]
fn concurrent_registrations_make_one_file_and_one_creator() {
    let home = TempHome::new();
    let results: Vec<Registered> = std::thread::scope(|s| {
        let handles: Vec<_> = (0..4)
            .map(|_| s.spawn(|| ledger::register(home.path(), "demo")))
            .collect();
        handles
            .into_iter()
            .map(|h| h.join().unwrap().unwrap())
            .collect()
    });
    assert_eq!(
        results
            .iter()
            .filter(|r| **r == Registered::Created)
            .count(),
        1,
        "{results:?}"
    );
    assert_eq!(
        names(&path_of(&home, "demo")).len(),
        1,
        "no stray temporary file"
    );
}

#[test]
fn register_refuses_a_project_id_that_could_leave_its_folder_and_creates_nothing() {
    let home = TempHome::new();
    for bad in ["", "..", "a/b", "a b", "a.b"] {
        assert_eq!(
            ledger::register(home.path(), bad),
            Err(LedgerError::InvalidProjectId),
            "{bad:?}"
        );
    }
    assert!(home.snapshot().is_empty());
}

#[test]
fn register_waits_for_the_go_generations_home_and_a_missing_home_is_not_created() {
    let home = TempHome::new();
    home.make_go_generation();
    let before = home.snapshot();
    assert!(matches!(
        ledger::register(home.path(), "demo"),
        Err(LedgerError::HomeNotReady(_))
    ));
    assert_eq!(home.snapshot(), before);

    let parent = TempHome::new();
    let missing = parent.path().join("no-such-home");
    assert!(matches!(
        ledger::register(&missing, "demo"),
        Err(LedgerError::HomeNotReady(_))
    ));
    assert!(!missing.exists());
}

// ---- open never creates --------------------------------------------------------------------------

#[test]
fn opening_an_absent_ledger_creates_no_folder_and_no_file() {
    let fresh = TempHome::new();
    // A HOME with the folders but no ledger, and one with the project folder but no file.
    let folders = TempHome::new();
    fs::create_dir_all(folders.path().join(".moai/projects")).unwrap();
    let project_folder = TempHome::new();
    fs::create_dir_all(project_folder.path().join(".moai/projects/demo")).unwrap();
    for home in [&fresh, &folders, &project_folder] {
        let before = home.snapshot();
        for error in open_errors(home, "demo") {
            assert_eq!(error, LedgerError::Absent);
        }
        assert_eq!(
            home.snapshot(),
            before,
            "entries, bytes and folder modification times"
        );
    }
}

#[test]
fn a_read_only_open_and_read_leaves_the_folder_and_the_file_unchanged() {
    let (home, file) = registered("demo");
    let before = home.snapshot();
    {
        let r = ledger::open_read_only(home.path(), "demo").unwrap();
        assert_eq!(r.schema_version(), 1);
        assert_eq!(r.project_id(), "demo");
        let n: i64 = r
            .conn()
            .query_row("SELECT count(*) FROM card", [], |row| row.get(0))
            .unwrap();
        assert_eq!(n, 0);
        assert_eq!(
            names(&file).len(),
            1,
            "no -journal, -wal or -shm while open"
        );
    }
    assert_eq!(home.snapshot(), before);
}

#[test]
fn a_write_on_a_read_only_connection_fails_and_changes_nothing() {
    let (home, file) = registered("demo");
    let before = home.snapshot();
    let r = ledger::open_read_only(home.path(), "demo").unwrap();
    let error = r
        .conn()
        .execute(
            "INSERT INTO card (title, created_by, created_at, updated_at) VALUES ('t', 'x', 1, 1)",
            [],
        )
        .unwrap_err();
    assert_eq!(LedgerError::from(error), LedgerError::ReadOnly);
    for sql in [
        "DELETE FROM meta",
        "PRAGMA user_version = 9",
        "CREATE TABLE sneaky (a)",
    ] {
        assert!(r.conn().execute_batch(sql).is_err(), "{sql}");
    }
    drop(r);
    assert_eq!(names(&file).len(), 1);
    assert_eq!(home.snapshot(), before);
}

// ---- WAL ---------------------------------------------------------------------------------------

fn make_wal(file: &Path) {
    let c = rusqlite::Connection::open(file).unwrap();
    let mode: String = c
        .query_row("PRAGMA journal_mode = WAL", [], |r| r.get(0))
        .unwrap();
    assert_eq!(mode, "wal");
}

#[test]
fn a_wal_mode_file_is_refused_before_sqlite_touches_it() {
    let (home, file) = registered("demo");
    make_wal(&file);
    let bytes = fs::read(&file).unwrap();
    assert_eq!((bytes[18], bytes[19]), (2, 2));
    let before = home.snapshot();
    for error in open_errors(&home, "demo") {
        assert_eq!(error, LedgerError::WalUnexpected);
    }
    assert_eq!(
        ledger::register(home.path(), "demo"),
        Err(LedgerError::WalUnexpected)
    );
    assert_eq!(fs::read(&file).unwrap(), bytes, "file bytes unchanged");
    assert_eq!(home.snapshot(), before, "no -wal or -shm was made");
}

#[test]
fn either_header_byte_of_a_wal_file_is_enough_to_refuse_it() {
    for at in [18usize, 19] {
        let (home, file) = registered("demo");
        let mut bytes = fs::read(&file).unwrap();
        assert_eq!((bytes[at], bytes[18], bytes[19]), (1, 1, 1));
        bytes[at] = 2;
        fs::write(&file, &bytes).unwrap();
        let before = home.snapshot();
        for error in open_errors(&home, "demo") {
            assert_eq!(error, LedgerError::WalUnexpected, "byte {at}");
        }
        assert_eq!(home.snapshot(), before);
    }
}

// ---- files that are not this project's ledger -------------------------------------------------

fn assert_unavailable_and_untouched(home: &TempHome, file: &Path) {
    let bytes = fs::read(file).unwrap();
    let before = home.snapshot();
    for error in open_errors(home, "demo") {
        assert!(
            matches!(error, LedgerError::Unavailable(_)),
            "{error:?} for {} bytes",
            bytes.len()
        );
        assert_eq!(error.code(), "ledger_unavailable");
    }
    assert!(matches!(
        ledger::register(home.path(), "demo"),
        Err(LedgerError::Unavailable(_))
    ));
    assert_eq!(fs::read(file).unwrap(), bytes, "bytes unchanged");
    assert_eq!(names(file).len(), 1, "no replacement and no side file");
    assert_eq!(home.snapshot(), before);
}

fn put(home: &TempHome, bytes: &[u8]) -> PathBuf {
    let file = path_of(home, "demo");
    fs::create_dir_all(file.parent().unwrap()).unwrap();
    fs::write(&file, bytes).unwrap();
    file
}

#[test]
fn a_corrupt_or_foreign_file_is_unavailable_and_is_never_replaced() {
    let garbage = vec![0x5Au8; 8192];
    let mut short_header = b"SQLite format 3\0".to_vec();
    short_header.resize(99, 0);
    for (name, bytes) in [
        ("empty file", Vec::new()),
        ("text", b"this is not a database".to_vec()),
        ("garbage", garbage),
        ("header cut short", short_header),
    ] {
        let home = TempHome::new();
        let file = put(&home, &bytes);
        assert_unavailable_and_untouched(&home, &file);
        eprintln!("{name}: ok");
    }
}

#[test]
fn a_ledger_whose_pages_were_overwritten_is_unavailable_and_unchanged() {
    let (home, file) = registered("demo");
    let mut bytes = fs::read(&file).unwrap();
    for b in bytes[100..].iter_mut() {
        *b = 0xFF;
    }
    fs::write(&file, &bytes).unwrap();
    assert_unavailable_and_untouched(&home, &file);
}

#[test]
fn another_application_id_or_a_missing_schema_is_unavailable() {
    // A real database that is not a MoAI ledger.
    let home = TempHome::new();
    let file = put(&home, b"");
    {
        let c = rusqlite::Connection::open(&file).unwrap();
        c.execute_batch("CREATE TABLE other (a); PRAGMA user_version = 1;")
            .unwrap();
    }
    assert_unavailable_and_untouched(&home, &file);

    // The right application id but no schema (user_version 0).
    let home = TempHome::new();
    let file = put(&home, b"");
    {
        let c = rusqlite::Connection::open(&file).unwrap();
        c.execute_batch(
            "CREATE TABLE meta (key TEXT, value TEXT); PRAGMA application_id = 1296125257;",
        )
        .unwrap();
    }
    assert_unavailable_and_untouched(&home, &file);

    // The right application id and version but no project id row.
    let home = TempHome::new();
    let file = put(&home, b"");
    {
        let c = rusqlite::Connection::open(&file).unwrap();
        c.execute_batch(
            "CREATE TABLE meta (key TEXT, value TEXT); PRAGMA application_id = 1296125257; PRAGMA user_version = 1;",
        )
        .unwrap();
    }
    assert_unavailable_and_untouched(&home, &file);
}

#[test]
fn a_link_or_a_folder_at_the_ledger_path_is_unavailable() {
    let (home, file) = registered("demo");
    let real = file.with_extension("real");
    fs::rename(&file, &real).unwrap();
    #[cfg(unix)]
    {
        std::os::unix::fs::symlink(&real, &file).unwrap();
        for error in open_errors(&home, "demo") {
            assert!(matches!(error, LedgerError::Unavailable(_)), "{error:?}");
        }
        assert!(matches!(
            ledger::register(home.path(), "demo"),
            Err(LedgerError::Unavailable(_))
        ));
        fs::remove_file(&file).unwrap();
    }
    fs::create_dir(&file).unwrap();
    for error in open_errors(&home, "demo") {
        assert!(matches!(error, LedgerError::Unavailable(_)), "{error:?}");
    }
}

#[test]
fn a_ledger_of_another_project_is_an_id_collision_and_stays_as_it_was() {
    let (home, alpha) = registered("alpha");
    let beta = path_of(&home, "beta");
    fs::create_dir_all(beta.parent().unwrap()).unwrap();
    fs::copy(&alpha, &beta).unwrap();
    let bytes = fs::read(&beta).unwrap();
    let before = home.snapshot();
    let want = LedgerError::IdCollision {
        expected: "beta".into(),
        found: "alpha".into(),
    };
    for error in open_errors(&home, "beta") {
        assert_eq!(error, want);
    }
    assert_eq!(ledger::register(home.path(), "beta"), Err(want));
    assert_eq!(fs::read(&beta).unwrap(), bytes);
    assert_eq!(home.snapshot(), before);
    // The original still opens for its own project.
    ledger::open_read_only(home.path(), "alpha").unwrap();
}

#[test]
fn a_newer_schema_version_is_refused_and_the_current_one_opens() {
    let (home, file) = registered("demo");
    ledger::open_read_only(home.path(), "demo").unwrap();
    {
        let c = rusqlite::Connection::open(&file).unwrap();
        c.execute_batch("PRAGMA user_version = 2;").unwrap();
    }
    let bytes = fs::read(&file).unwrap();
    let before = home.snapshot();
    for error in open_errors(&home, "demo") {
        assert_eq!(
            error,
            LedgerError::SchemaNewer {
                found: 2,
                supported: 1
            }
        );
    }
    assert_eq!(
        ledger::register(home.path(), "demo"),
        Err(LedgerError::SchemaNewer {
            found: 2,
            supported: 1
        })
    );
    assert_eq!(fs::read(&file).unwrap(), bytes);
    assert_eq!(home.snapshot(), before);
}

// ---- writing ------------------------------------------------------------------------------------

#[test]
fn a_committed_write_leaves_only_the_database_file_and_an_error_rolls_everything_back() {
    let (home, file) = registered("demo");
    let mut w = ledger::open_write(home.path(), "demo").unwrap();
    let id = w
        .with_write_tx(|tx, _| {
            tx.execute(
                "INSERT INTO card (title, created_by, created_at, updated_at) VALUES ('kept', 'x', 1, 1)",
                [],
            )?;
            Ok(tx.last_insert_rowid())
        })
        .unwrap();
    assert_eq!(id, 1);
    assert_eq!(
        names(&file).len(),
        1,
        "the rollback journal is gone after commit"
    );

    let failed: Result<(), LedgerError> = w.with_write_tx(|tx, _| {
        tx.execute(
            "INSERT INTO card (title, created_by, created_at, updated_at) VALUES ('lost', 'x', 1, 1)",
            [],
        )?;
        Err(LedgerError::Io("stop here".into()))
    });
    assert_eq!(failed, Err(LedgerError::Io("stop here".into())));
    let titles: Vec<String> = w
        .conn()
        .prepare("SELECT title FROM card")
        .unwrap()
        .query_map([], |r| r.get(0))
        .unwrap()
        .map(Result::unwrap)
        .collect();
    assert_eq!(titles, vec!["kept"]);
    // A second connection sees the committed row and not the rolled-back one.
    let r = ledger::open_read_only(home.path(), "demo").unwrap();
    let n: i64 = r
        .conn()
        .query_row("SELECT count(*) FROM card", [], |row| row.get(0))
        .unwrap();
    assert_eq!(n, 1);
}

#[test]
fn the_transaction_clock_is_the_db_clock() {
    let (home, _) = registered("demo");
    let mut w = ledger::open_write(home.path(), "demo").unwrap();
    let (now, in_tx): (i64, i64) = w
        .with_write_tx(|tx, now| Ok((now, tx.query_row("SELECT unixepoch()", [], |r| r.get(0))?)))
        .unwrap();
    // The same clock SQLite uses for unixepoch(); not zero, not a counter.
    assert!((in_tx - now).abs() <= 2, "{now} vs {in_tx}");
    let system = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64;
    assert!((system - now).abs() < 120, "{now} vs the system {system}");
}

// ---- two processes ------------------------------------------------------------------------------

/// Marker for the re-executed test binary: the child half of the race below.
const RACE_ENV: &str = "MOAI_LEDGER_RACE_DIR";

fn wait_for(path: &Path, what: &str) {
    let deadline = std::time::Instant::now() + std::time::Duration::from_secs(30);
    while !path.exists() {
        assert!(
            std::time::Instant::now() < deadline,
            "timed out waiting for {what}"
        );
        std::thread::sleep(std::time::Duration::from_millis(10));
    }
}

/// The child process of `two_processes_writing_at_once_do_not_lose_an_update`. Run alone it
/// does nothing; the race test starts the same test binary with `RACE_ENV` set.
#[test]
fn ledger_race_child() {
    let Some(dir) = std::env::var_os(RACE_ENV).map(PathBuf::from) else {
        return;
    };
    let role = std::env::var("MOAI_LEDGER_RACE_ROLE").unwrap();
    let home = PathBuf::from(std::env::var_os("HOME").unwrap());
    fs::write(dir.join(format!("{role}.start")), "").unwrap();
    let mut w = ledger::open_write(&home, "demo").unwrap();
    w.with_write_tx(|tx, now| {
        // Only a process that holds the write lock gets here.
        fs::write(dir.join(format!("{role}.in_tx")), "").unwrap();
        if role == "a" {
            wait_for(&dir.join("a.go"), "the parent's go signal");
        }
        let seen: i64 = tx.query_row(
            "SELECT CAST(value AS INTEGER) FROM meta WHERE key = 'race_counter'",
            [],
            |r| r.get(0),
        )?;
        tx.execute(
            "UPDATE meta SET value = CAST(?1 AS TEXT) WHERE key = 'race_counter'",
            [seen + 1],
        )?;
        tx.execute(
            "INSERT INTO event (event_key, at, actor, kind, payload) VALUES (?1, ?2, ?3, 'race', ?4)",
            rusqlite::params![
                format!("race:{role}"),
                now,
                role,
                format!("{{\"seen\":{seen}}}")
            ],
        )?;
        Ok(())
    })
    .unwrap();
}

#[test]
fn two_processes_writing_at_once_do_not_lose_an_update() {
    let (home, _) = registered("demo");
    let mut w = ledger::open_write(home.path(), "demo").unwrap();
    w.with_write_tx(|tx, _| {
        tx.execute(
            "INSERT INTO meta (key, value) VALUES ('race_counter', '0')",
            [],
        )?;
        Ok(())
    })
    .unwrap();
    drop(w);
    let signals = tempfile::tempdir().unwrap();
    let spawn = |role: &str| {
        std::process::Command::new(std::env::current_exe().unwrap())
            .args([
                "--exact",
                "ledger_race_child",
                "--nocapture",
                "--test-threads=1",
            ])
            .clean_env()
            .env("HOME", home.path())
            .env(RACE_ENV, signals.path())
            .env("MOAI_LEDGER_RACE_ROLE", role)
            .stdout(std::process::Stdio::piped())
            .stderr(std::process::Stdio::piped())
            .spawn()
            .unwrap()
    };
    let a = spawn("a");
    // A holds the write lock inside its transaction...
    wait_for(&signals.path().join("a.in_tx"), "A inside its transaction");
    // ... when B starts and tries to write.
    let b = spawn("b");
    wait_for(&signals.path().join("b.start"), "B to start");
    std::thread::sleep(std::time::Duration::from_millis(600));
    assert!(
        !signals.path().join("b.in_tx").exists(),
        "B must wait for A's write lock before its closure runs (BEGIN IMMEDIATE)"
    );
    fs::write(signals.path().join("a.go"), "").unwrap();
    for (role, child) in [("a", a), ("b", b)] {
        let out = child.wait_with_output().unwrap();
        assert!(
            out.status.success(),
            "{role} failed:\n{}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
    }
    let r = ledger::open_read_only(home.path(), "demo").unwrap();
    let counter: String = r
        .conn()
        .query_row(
            "SELECT value FROM meta WHERE key = 'race_counter'",
            [],
            |row| row.get(0),
        )
        .unwrap();
    assert_eq!(counter, "2", "no lost update");
    let seen = |role: &str| -> String {
        r.conn()
            .query_row(
                "SELECT payload FROM event WHERE event_key = ?1",
                [format!("race:{role}")],
                |row| row.get(0),
            )
            .unwrap()
    };
    assert_eq!(seen("a"), "{\"seen\":0}");
    assert_eq!(seen("b"), "{\"seen\":1}", "B read after A committed");
}
