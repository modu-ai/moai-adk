//! The folders above a ledger (Issue #3068 hardening: C3 owner-only modes, C4 no symbolic
//! links) and transient errors that are not damage (R2). Every test runs in a temporary HOME;
//! none touches the real `~/.moai`. Unix only: modes, owners, links and pipes are what it tests
//! (the Windows build compiles this file to nothing).

#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger::recover::{self, Observations, TargetState};
use moai_cli::ledger::{self, LedgerError, Registered};
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::{symlink, FileTypeExt, PermissionsExt};
use std::path::{Path, PathBuf};
use std::process::Output;

const ID: &str = "demo";

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().expect("run moai")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("not JSON ({e}): {}", String::from_utf8_lossy(&out.stdout)))
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{}\n{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    envelope(out)["result"].clone()
}

fn error_message(out: &Output) -> String {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    envelope(out)["error"]["message"]
        .as_str()
        .unwrap()
        .to_string()
}

fn mode(path: &Path) -> u32 {
    fs::symlink_metadata(path).unwrap().permissions().mode() & 0o777
}

fn dot_moai(h: &TempHome) -> PathBuf {
    h.path().join(".moai")
}
fn projects(h: &TempHome) -> PathBuf {
    dot_moai(h).join("projects")
}
fn project(h: &TempHome) -> PathBuf {
    projects(h).join(ID)
}

/// `moai ledger-register-plan` then `-apply`, with the child's umask set first.
fn register_with_umask(h: &TempHome, profile: &Path, umask: u32) -> Value {
    use std::os::unix::process::CommandExt;
    let arg = profile.to_str().unwrap();
    let mut plan_cmd = h.command(&["ledger-register-plan", arg]);
    // SAFETY: umask is async-signal-safe; it runs in the forked child before exec.
    unsafe {
        plan_cmd.pre_exec(move || {
            libc::umask(umask as libc::mode_t);
            Ok(())
        });
    }
    let plan = result(&plan_cmd.output().unwrap());
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let mut apply_cmd = h.command(&[
        "ledger-register-apply",
        arg,
        "--plan-digest",
        &digest,
        "--permission-reference",
        "ref-1",
    ]);
    unsafe {
        apply_cmd.pre_exec(move || {
            libc::umask(umask as libc::mode_t);
            Ok(())
        });
    }
    result(&apply_cmd.output().unwrap())
}

// ---- C3: owner-only folders ---------------------------------------------------------------------

/// Every folder the register makes is 0700, also when the umask would open it (000) and when it
/// would close it further (077; 177 would even cut the owner's search bit), and the ledger file is 0600.
#[test]
fn register_makes_every_new_folder_0700_whatever_the_umask() {
    for umask in [0o000, 0o022, 0o027, 0o077, 0o177] {
        let h = TempHome::new();
        let profile = h.project(ID);
        let applied = register_with_umask(&h, &profile, umask);
        assert_eq!(applied["state"], json!("applied"), "umask {umask:o}");
        assert_eq!(applied["tightened"], json!([]));
        for folder in [dot_moai(&h), projects(&h), project(&h)] {
            assert_eq!(
                mode(&folder),
                0o700,
                "{} under umask {umask:o}",
                folder.display()
            );
        }
        let file = ledger::ledger_path(h.path(), ID).unwrap();
        assert_eq!(mode(&file), 0o600, "umask {umask:o}");
    }
}

/// The library call alone (no binary) gives the same modes under the test process's umask, and
/// reports the folders it made.
#[test]
fn the_library_register_reports_what_it_made() {
    let h = TempHome::new();
    let (r, report) = ledger::open::register_with_report(h.path(), ID).unwrap();
    assert_eq!(r, Registered::Created);
    assert_eq!(
        report.created,
        vec![dot_moai(&h), projects(&h), project(&h)]
    );
    assert!(report.tightened.is_empty());
    for folder in [dot_moai(&h), projects(&h), project(&h)] {
        assert_eq!(mode(&folder), 0o700);
    }
}

/// Existing loose folders of this user: the plan lists them, the apply tightens them and says so,
/// `.moai` is left alone, and a second plan has nothing left.
#[test]
fn existing_loose_folders_are_listed_then_tightened_but_dot_moai_is_left_alone() {
    let h = TempHome::new();
    let profile = h.project(ID);
    fs::create_dir_all(project(&h)).unwrap();
    for (folder, m) in [
        (dot_moai(&h), 0o755),
        (projects(&h), 0o755),
        (project(&h), 0o750),
    ] {
        fs::set_permissions(&folder, fs::Permissions::from_mode(m)).unwrap();
    }
    let arg = profile.to_str().unwrap();
    let before_status = h.snapshot();
    let status = result(&run(&h, &["ledger-status", arg]));
    assert_eq!(
        status["folder_modes"],
        json!({"projects": "0755", "project": "0750"})
    );
    assert_eq!(h.snapshot(), before_status, "status changes nothing");

    let plan = result(&run(&h, &["ledger-register-plan", arg]));
    assert_eq!(plan["action"], json!("register"));
    assert_eq!(
        plan["would_tighten"],
        json!([
            projects(&h).display().to_string(),
            project(&h).display().to_string()
        ])
    );
    assert_eq!(h.snapshot(), before_status, "the plan changes nothing");
    let digest = plan["plan_digest"].as_str().unwrap();
    let applied = result(&run(
        &h,
        &[
            "ledger-register-apply",
            arg,
            "--plan-digest",
            digest,
            "--permission-reference",
            "ref-1",
        ],
    ));
    assert_eq!(applied["state"], json!("applied"));
    assert_eq!(applied["tightened"], plan["would_tighten"]);
    assert_eq!(mode(&dot_moai(&h)), 0o755, ".moai belongs to other tools");
    assert_eq!((mode(&projects(&h)), mode(&project(&h))), (0o700, 0o700));
    let status = result(&run(&h, &["ledger-status", arg]));
    assert_eq!(
        status["folder_modes"],
        json!({"projects": "0700", "project": "0700"})
    );
    let plan = result(&run(&h, &["ledger-register-plan", arg]));
    assert_eq!(plan["would_tighten"], json!([]));
}

/// An already registered project with loose folders is tightened by the register apply too (the
/// plan shows it), and the digest the plan gave before differs from the one after.
#[test]
fn a_registered_project_with_loose_folders_is_tightened_by_the_apply() {
    let h = TempHome::new();
    let profile = h.project(ID);
    ledger::register(h.path(), ID).unwrap();
    for folder in [projects(&h), project(&h)] {
        fs::set_permissions(&folder, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let arg = profile.to_str().unwrap();
    let plan = result(&run(&h, &["ledger-register-plan", arg]));
    assert_eq!(plan["action"], json!("none"));
    assert_eq!(plan["would_tighten"].as_array().unwrap().len(), 2);
    let applied = result(&run(
        &h,
        &[
            "ledger-register-apply",
            arg,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "ref-1",
        ],
    ));
    assert_eq!(applied["state"], json!("unchanged"));
    assert_eq!(applied["tightened"], plan["would_tighten"]);
    assert_eq!((mode(&projects(&h)), mode(&project(&h))), (0o700, 0o700));
    let again = result(&run(&h, &["ledger-register-plan", arg]));
    assert_ne!(again["plan_digest"], plan["plan_digest"]);
}

/// An absent ledger folder reports null modes.
#[test]
fn status_reports_null_modes_for_absent_folders() {
    let h = TempHome::new();
    let profile = h.project(ID);
    let status = result(&run(&h, &["ledger-status", profile.to_str().unwrap()]));
    assert_eq!(
        status["folder_modes"],
        json!({"projects": null, "project": null})
    );
}

/// Recovery makes no folder; its files are private.
#[test]
fn recovery_files_are_private_and_no_folder_is_made() {
    let h = TempHome::new();
    ledger::register(h.path(), ID).unwrap();
    let file = ledger::ledger_path(h.path(), ID).unwrap();
    let bak = h.path().join("bak.db");
    fs::copy(&file, &bak).unwrap();
    fs::write(&file, b"NOT A DATABASE!! xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx").unwrap();
    let dirs_before: Vec<_> = TempHome::names_in(&projects(&h));
    let plan = recover::plan(h.path(), ID, &bak, &Observations::default()).unwrap();
    recover::apply_with_suffix(
        h.path(),
        ID,
        &bak,
        &Observations::default(),
        &plan.plan_digest,
        "20261116T000001Z",
    )
    .unwrap();
    assert_eq!(TempHome::names_in(&projects(&h)), dirs_before);
    assert_eq!(mode(&project(&h)), 0o700);
    assert_eq!(mode(&file), 0o600);
    assert_eq!(mode(&project(&h).join(".recover.lock")), 0o600);
}

// ---- C4: no symbolic link at the three components ------------------------------------------------

/// A second HOME holding a real registered tree; the first HOME's component links to its
/// counterpart there. Nothing may be written there through the link.
struct Linked {
    home: TempHome,
    profile: PathBuf,
    real: TempHome,
    /// The path that is the link.
    link: PathBuf,
}

fn linked(which: &'static str, dangling: bool) -> Linked {
    let home = TempHome::new();
    let profile = home.project(ID);
    let real = TempHome::new();
    ledger::register(real.path(), ID).unwrap();
    let (link, target) = match which {
        "dot_moai" => (dot_moai(&home), dot_moai(&real)),
        "projects" => {
            fs::create_dir_all(dot_moai(&home)).unwrap();
            (projects(&home), projects(&real))
        }
        _ => {
            fs::create_dir_all(projects(&home)).unwrap();
            (project(&home), project(&real))
        }
    };
    let target = if dangling {
        real.path().join("does-not-exist")
    } else {
        target
    };
    symlink(&target, &link).unwrap();
    Linked {
        home,
        profile,
        real,
        link,
    }
}

fn assert_path_link(err: &LedgerError, what: &str) {
    assert!(matches!(err, LedgerError::PathLink(_)), "{what}: {err:?}");
    assert_eq!(err.code(), "path_link");
}

#[test]
fn a_link_at_any_of_the_three_components_is_refused_everywhere_and_nothing_is_written_through_it() {
    for dangling in [false, true] {
        for which in ["dot_moai", "projects", "project"] {
            let l = linked(which, dangling);
            let name = format!("{which} (dangling {dangling})");
            let real_before = l.real.snapshot();
            let h = l.home.path();

            assert_path_link(&ledger::register(h, ID).unwrap_err(), &name);
            assert_path_link(&ledger::open_read_only(h, ID).err().unwrap(), &name);
            assert_path_link(&ledger::open_write(h, ID).err().unwrap(), &name);
            let source = l.home.path().join("source.db");
            fs::copy(ledger::ledger_path(l.real.path(), ID).unwrap(), &source).unwrap();
            assert_path_link(
                &recover::plan(h, ID, &source, &Observations::default()).unwrap_err(),
                &name,
            );
            assert_path_link(
                &recover::apply_with_suffix(
                    h,
                    ID,
                    &source,
                    &Observations::default(),
                    "sha256:0",
                    "20261116T000001Z",
                )
                .unwrap_err(),
                &name,
            );
            assert_path_link(&recover::lock_recovery(h, ID).unwrap_err(), &name);
            assert_path_link(&recover::stale_temp(h, ID).unwrap_err(), &name);

            // The binary: the plan is blocked, the apply and the recovery refuse, the status
            // reports the link.
            let arg = l.profile.to_str().unwrap();
            let plan = result(&run(&l.home, &["ledger-register-plan", arg]));
            assert_eq!(plan["action"], json!("blocked"), "{name}");
            assert_eq!(plan["problem"], json!("path_link"), "{name}");
            assert_eq!(plan["writes"], json!(false));
            let digest = plan["plan_digest"].as_str().unwrap();
            let msg = error_message(&run(
                &l.home,
                &[
                    "ledger-register-apply",
                    arg,
                    "--plan-digest",
                    digest,
                    "--permission-reference",
                    "ref-1",
                ],
            ));
            assert!(msg.contains("path_link"), "{name}: {msg}");
            let status = result(&run(&l.home, &["ledger-status", arg]));
            assert_eq!(status["problem"], json!("path_link"), "{name}");
            assert_eq!(status["state"], json!("unavailable"));
            assert_eq!(status["writes"], json!(false));
            for verb in ["ledger-recover-plan", "ledger-recover-apply"] {
                let mut args = vec![verb, arg, "--source", source.to_str().unwrap()];
                if verb == "ledger-recover-apply" {
                    args.extend(["--plan-digest", "sha256:0", "--permission-reference", "r1"]);
                }
                let msg = error_message(&run(&l.home, &args));
                assert!(msg.contains("path_link"), "{name} {verb}: {msg}");
            }

            // Nothing was written through the link, and the link itself is still a link.
            assert_eq!(
                l.real.snapshot(),
                real_before,
                "{name}: the link target is unchanged"
            );
            assert!(
                fs::symlink_metadata(&l.link)
                    .unwrap()
                    .file_type()
                    .is_symlink(),
                "{name}"
            );
        }
    }
}

/// The HOME itself (and so its ancestors) may be reached through a link.
#[test]
fn a_home_reached_through_a_link_still_works() {
    let real = TempHome::new();
    let holder = TempHome::new();
    let via = holder.path().join("home-link");
    symlink(real.path(), &via).unwrap();
    assert_eq!(ledger::register(&via, ID).unwrap(), Registered::Created);
    ledger::open_read_only(&via, ID).unwrap();
    ledger::open_write(&via, ID).unwrap();
    assert_eq!(
        ledger::register(&via, ID).unwrap(),
        Registered::AlreadyExists
    );
    assert_eq!(mode(&real.path().join(".moai/projects").join(ID)), 0o700);
}

// ---- R2: a transient error is not damage ----------------------------------------------------------

fn is_root() -> bool {
    unsafe { libc::geteuid() == 0 }
}

/// A file this process cannot read is `io`, not `ledger_unavailable`; the recovery classifies
/// it as `unconfirmed` and refuses to replace it. A header cut short is damage.
#[test]
fn an_unreadable_header_is_io_and_a_short_header_is_damage() {
    let h = TempHome::new();
    ledger::register(h.path(), ID).unwrap();
    let file = ledger::ledger_path(h.path(), ID).unwrap();
    let source = h.path().join("good.db");
    fs::copy(&file, &source).unwrap();

    // Cut short: fewer than 100 bytes. Damage, and replaceable.
    fs::write(&file, vec![b'x'; 50]).unwrap();
    assert!(matches!(
        ledger::open_read_only(h.path(), ID).err().unwrap(),
        LedgerError::Unavailable(_)
    ));
    let plan = recover::plan(h.path(), ID, &source, &Observations::default()).unwrap();
    assert_eq!(plan.target_state, TargetState::Unavailable);

    if is_root() {
        return; // root reads a 000 file
    }
    fs::copy(&source, &file).unwrap();
    fs::set_permissions(&file, fs::Permissions::from_mode(0o000)).unwrap();
    let err = ledger::open_read_only(h.path(), ID).err().unwrap();
    assert!(matches!(err, LedgerError::Io(_)), "{err:?}");
    let plan = recover::plan(h.path(), ID, &source, &Observations::default()).unwrap();
    assert_eq!(plan.target_state, TargetState::Unconfirmed);
    let before = fs::metadata(&file).unwrap().len();
    let err = recover::apply_with_suffix(
        h.path(),
        ID,
        &source,
        &Observations::default(),
        &plan.plan_digest,
        "20261116T000001Z",
    )
    .unwrap_err();
    fs::set_permissions(&file, fs::Permissions::from_mode(0o600)).unwrap();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("target_not_damaged")),
        "{err:?}"
    );
    assert_eq!(fs::metadata(&file).unwrap().len(), before);
    assert!(TempHome::names_in(file.parent().unwrap())
        .iter()
        .all(|n| !n.contains("damaged") && !n.contains("recovering")));
}

/// The `meta` lookup: a missing table or row is damage; an error that is neither damage nor a
/// lock (here a view that fails with an integer overflow when read) is `io`, and the recovery
/// does not replace such a ledger. The failure is injected by replacing the `meta` table in a
/// copy of the ledger with that view: deterministic, no timing.
#[test]
fn a_failing_meta_lookup_is_io_but_a_missing_table_or_row_is_damage() {
    let h = TempHome::new();
    ledger::register(h.path(), ID).unwrap();
    let file = ledger::ledger_path(h.path(), ID).unwrap();
    let source = h.path().join("good.db");
    fs::copy(&file, &source).unwrap();

    let edit = |sql: &str| {
        fs::copy(&source, &file).unwrap();
        Connection::open(&file).unwrap().execute_batch(sql).unwrap();
    };

    edit("DROP TABLE meta;");
    assert!(matches!(
        ledger::open_read_only(h.path(), ID).err().unwrap(),
        LedgerError::Unavailable(_)
    ));
    let plan = recover::plan(h.path(), ID, &source, &Observations::default()).unwrap();
    assert_eq!(plan.target_state, TargetState::Unavailable);

    edit("DELETE FROM meta WHERE key = 'project_id';");
    assert!(matches!(
        ledger::open_read_only(h.path(), ID).err().unwrap(),
        LedgerError::Unavailable(_)
    ));

    edit(
        "DROP TABLE meta; CREATE VIEW meta AS \
         SELECT 'project_id' AS key, abs(-9223372036854775808) AS value;",
    );
    let err = ledger::open_read_only(h.path(), ID).err().unwrap();
    assert!(matches!(err, LedgerError::Io(_)), "{err:?}");
    let bytes = fs::read(&file).unwrap();
    let plan = recover::plan(h.path(), ID, &source, &Observations::default()).unwrap();
    assert_eq!(plan.target_state, TargetState::Unconfirmed);
    let err = recover::apply_with_suffix(
        h.path(),
        ID,
        &source,
        &Observations::default(),
        &plan.plan_digest,
        "20261116T000001Z",
    )
    .unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("target_not_damaged")),
        "{err:?}"
    );
    assert_eq!(
        fs::read(&file).unwrap(),
        bytes,
        "the ledger was not replaced"
    );
}

// ---- R4: the lock path must be a regular file -----------------------------------------------------

fn lock_path(h: &TempHome) -> PathBuf {
    project(h).join(recover::LOCK_NAME)
}

/// Runs `lock_recovery` on a thread and gives it five seconds: a pipe that is opened for writing
/// without a reader would block it for ever.
fn lock_promptly(h: &TempHome) -> Result<(), LedgerError> {
    let (tx, rx) = std::sync::mpsc::channel();
    let home = h.path().to_path_buf();
    std::thread::spawn(move || {
        let _ = tx.send(recover::lock_recovery(&home, ID).map(|_| ()));
    });
    rx.recv_timeout(std::time::Duration::from_secs(5))
        .expect("lock_recovery returned promptly")
}

fn assert_lock_not_regular(result: Result<(), LedgerError>, what: &str) {
    assert!(
        matches!(&result, Err(LedgerError::Constraint(m)) if m.starts_with("lock_not_regular")),
        "{what}: {result:?}"
    );
}

#[test]
fn a_fifo_a_folder_and_a_link_at_the_lock_path_are_refused_without_being_opened() {
    // A pipe: opening it for writing would block; the lstat refuses first.
    let h = TempHome::new();
    ledger::register(h.path(), ID).unwrap();
    let fifo = std::ffi::CString::new(lock_path(&h).to_str().unwrap()).unwrap();
    assert_eq!(unsafe { libc::mkfifo(fifo.as_ptr(), 0o600) }, 0);
    // (`snapshot` would read the pipe and block; compare the folder listing instead.)
    let before = TempHome::names_in(&project(&h));
    assert_lock_not_regular(lock_promptly(&h), "fifo");
    assert_eq!(TempHome::names_in(&project(&h)), before, "nothing changed");
    assert!(fs::symlink_metadata(lock_path(&h))
        .unwrap()
        .file_type()
        .is_fifo());

    // A folder.
    let h = TempHome::new();
    ledger::register(h.path(), ID).unwrap();
    fs::create_dir(lock_path(&h)).unwrap();
    let before = h.snapshot();
    assert_lock_not_regular(lock_promptly(&h), "folder");
    assert_eq!(h.snapshot(), before);

    // A link to a regular file elsewhere: not followed, the target is not locked or changed.
    let h = TempHome::new();
    ledger::register(h.path(), ID).unwrap();
    let elsewhere = h.path().join("elsewhere.txt");
    fs::write(&elsewhere, b"keep").unwrap();
    symlink(&elsewhere, lock_path(&h)).unwrap();
    let before = h.snapshot();
    assert_lock_not_regular(lock_promptly(&h), "link");
    assert_eq!(h.snapshot(), before);
    assert_eq!(fs::read(&elsewhere).unwrap(), b"keep");

    // A dangling link, too.
    let h = TempHome::new();
    ledger::register(h.path(), ID).unwrap();
    symlink(h.path().join("nowhere"), lock_path(&h)).unwrap();
    assert_lock_not_regular(lock_promptly(&h), "dangling link");

    // A regular file is the normal case and stays 0600.
    let h = TempHome::new();
    ledger::register(h.path(), ID).unwrap();
    lock_promptly(&h).unwrap();
    assert_eq!(mode(&lock_path(&h)), 0o600);
}

// ---- R3: a refused apply leaves nothing behind -----------------------------------------------------

/// A stale digest, a healthy target and a target that is not damaged are all answered before the
/// lock file is made: the HOME is byte-for-byte what it was.
#[test]
fn a_refused_apply_makes_no_lock_file() {
    let h = TempHome::new();
    ledger::register(h.path(), ID).unwrap();
    let file = ledger::ledger_path(h.path(), ID).unwrap();
    let source = h.path().join("good.db");
    fs::copy(&file, &source).unwrap();
    let obs = Observations::default();
    let apply = |digest: &str| {
        recover::apply_with_suffix(h.path(), ID, &source, &obs, digest, "20261116T000001Z")
    };

    // Healthy target: refused as healthy.
    let before = h.snapshot();
    let plan = recover::plan(h.path(), ID, &source, &obs).unwrap();
    let err = apply(&plan.plan_digest).unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("target_healthy")),
        "{err:?}"
    );
    assert_eq!(h.snapshot(), before);

    // Stale digest on a damaged target.
    fs::write(&file, vec![b'x'; 200]).unwrap();
    let before = h.snapshot();
    let err = apply("sha256:0").unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("stale_plan")),
        "{err:?}"
    );
    assert_eq!(h.snapshot(), before);
    assert!(!lock_path(&h).exists());

    // A busy (not damaged) target.
    fs::copy(&source, &file).unwrap();
    let holder = Connection::open(&file).unwrap();
    holder.execute_batch("BEGIN EXCLUSIVE;").unwrap();
    let before = h.snapshot();
    let plan = recover::plan(h.path(), ID, &source, &obs).unwrap();
    let err = apply(&plan.plan_digest).unwrap_err();
    assert!(
        matches!(&err, LedgerError::Constraint(m) if m.starts_with("target_not_damaged")),
        "{err:?}"
    );
    assert_eq!(h.snapshot(), before);
    holder.execute_batch("ROLLBACK;").unwrap();
    assert!(!lock_path(&h).exists());
}
