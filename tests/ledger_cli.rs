//! The ledger commands `ledger-status`, `ledger-register-plan` and `ledger-register-apply`
//! (Issue #3068 P1; AC6 and AC8 basics): what they create, what they refuse, and that a report
//! never writes. Every command runs against a temporary HOME; the real `~/.moai` is not read.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::Output;

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().expect("run moai")
}

fn project_arg(project: &Path) -> String {
    project.to_str().unwrap().to_string()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn error_kind(out: &Output) -> String {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["kind"].as_str().unwrap().to_string()
}

fn apply_args<'a>(project: &'a str, digest: &'a str, permission: &'a str) -> Vec<&'a str> {
    vec![
        "ledger-register-apply",
        project,
        "--plan-digest",
        digest,
        "--permission-reference",
        permission,
    ]
}

#[test]
fn status_of_a_project_without_a_ledger_is_absent_and_creates_nothing() {
    let home = TempHome::new();
    let project = home.project("demo");
    let p = project_arg(&project);
    let before = home.snapshot();
    // $HOME is the temporary folder (no --home), then the same with --home.
    for args in [
        vec!["ledger-status", p.as_str()],
        vec![
            "ledger-status",
            p.as_str(),
            "--home",
            home.path().to_str().unwrap(),
        ],
    ] {
        let r = result(&run(&home, &args));
        assert_eq!(r["exists"], json!(false));
        assert_eq!(r["state"], json!("absent"));
        assert_eq!(r["project_id"], json!("demo"));
        assert_eq!(r["schema_version"], Value::Null);
        assert_eq!(r["wal"], json!(false));
        assert_eq!(r["cards"], Value::Null);
        assert_eq!(r["writes"], json!(false));
    }
    assert_eq!(
        home.snapshot(),
        before,
        "no folder, no file, no mtime change"
    );
    assert!(!home.path().join(".moai").exists());
}

#[test]
fn plan_names_what_registering_would_make_and_writes_nothing() {
    let home = TempHome::new();
    let project = home.project("demo");
    let p = project_arg(&project);
    let before = home.snapshot();
    let r = result(&run(&home, &["ledger-register-plan", &p]));
    assert_eq!(r["state"], json!("absent"));
    assert_eq!(r["action"], json!("register"));
    assert_eq!(r["writes"], json!(false));
    let file = ledger::ledger_path(home.path(), "demo").unwrap();
    let dir = file.parent().unwrap();
    assert_eq!(
        r["would_create"],
        json!([dir.display().to_string(), file.display().to_string()])
    );
    assert!(r["plan_digest"].as_str().unwrap().starts_with("sha256:"));
    assert_eq!(home.snapshot(), before);
}

#[test]
fn register_apply_needs_a_permission_reference_and_a_current_digest() {
    let home = TempHome::new();
    let project = home.project("demo");
    let p = project_arg(&project);
    let plan = result(&run(&home, &["ledger-register-plan", &p]));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let before = home.snapshot();
    for bad in ["", "null", "placeholder", "   "] {
        let out = run(&home, &apply_args(&p, &digest, bad));
        assert_eq!(error_kind(&out), "permission_reference_missing", "{bad:?}");
    }
    let stale = format!("sha256:{}", "0".repeat(64));
    let out = run(&home, &apply_args(&p, &stale, "approved-by-test"));
    assert_eq!(error_kind(&out), "stale_plan");
    assert_eq!(
        home.snapshot(),
        before,
        "nothing was written by the refusals"
    );
}

#[test]
fn register_apply_creates_the_ledger_and_status_then_reports_it() {
    let home = TempHome::new();
    let project = home.project("demo");
    let p = project_arg(&project);
    let plan = result(&run(&home, &["ledger-register-plan", &p]));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();

    let applied = result(&run(&home, &apply_args(&p, &digest, "approved-by-test")));
    assert_eq!(applied["state"], json!("applied"));
    let file = ledger::ledger_path(home.path(), "demo").unwrap();
    assert!(file.is_file());
    assert_eq!(TempHome::names_in(file.parent().unwrap()).len(), 1);

    let status_before = home.snapshot();
    let s = result(&run(&home, &["ledger-status", &p]));
    assert_eq!(s["exists"], json!(true));
    assert_eq!(s["state"], json!("ready"));
    assert_eq!(s["schema_version"], json!(1));
    assert_eq!(s["wal"], json!(false));
    assert_eq!(s["cards"], json!({"total": 0, "by_state": {}}));
    assert_eq!(s["problem"], Value::Null);
    assert_eq!(
        home.snapshot(),
        status_before,
        "a status report writes nothing"
    );
}

#[test]
fn status_counts_cards_by_state() {
    let home = TempHome::new();
    let project = home.project("demo");
    let p = project_arg(&project);
    ledger::register(home.path(), "demo").unwrap();
    let mut w = ledger::open_write(home.path(), "demo").unwrap();
    w.with_write_tx(|tx, now| {
        for state in ["draft", "draft", "ready"] {
            tx.execute(
                "INSERT INTO card (title, state, created_by, created_at, updated_at) VALUES ('t', ?1, 'x', ?2, ?2)",
                rusqlite::params![state, now],
            )?;
        }
        Ok(())
    })
    .unwrap();
    drop(w);
    let s = result(&run(&home, &["ledger-status", &p]));
    assert_eq!(
        s["cards"],
        json!({"total": 3, "by_state": {"draft": 2, "ready": 1}})
    );
}

#[test]
fn registering_twice_is_idempotent_and_an_old_digest_goes_stale() {
    let home = TempHome::new();
    let project = home.project("demo");
    let p = project_arg(&project);
    let first_plan = result(&run(&home, &["ledger-register-plan", &p]));
    let first_digest = first_plan["plan_digest"].as_str().unwrap().to_string();
    result(&run(
        &home,
        &apply_args(&p, &first_digest, "approved-by-test"),
    ));
    let after_first = home.snapshot();

    // The plan of a registered project says there is nothing to do; applying it changes nothing.
    let second_plan = result(&run(&home, &["ledger-register-plan", &p]));
    assert_eq!(second_plan["state"], json!("registered"));
    assert_eq!(second_plan["action"], json!("none"));
    assert_eq!(second_plan["would_create"], json!([]));
    assert_ne!(second_plan["plan_digest"], first_plan["plan_digest"]);
    let second_digest = second_plan["plan_digest"].as_str().unwrap().to_string();
    let again = result(&run(
        &home,
        &apply_args(&p, &second_digest, "approved-by-test"),
    ));
    assert_eq!(again["state"], json!("unchanged"));
    assert_eq!(again["created"], json!([]));
    assert_eq!(
        home.snapshot(),
        after_first,
        "no second file, no change at all"
    );

    // The first digest describes a state that is gone.
    let out = run(&home, &apply_args(&p, &first_digest, "approved-by-test"));
    assert_eq!(error_kind(&out), "stale_plan");
    assert_eq!(home.snapshot(), after_first);
}

#[test]
fn the_go_generations_home_blocks_the_plan_and_the_apply() {
    let home = TempHome::new();
    let project = home.project("demo");
    let p = project_arg(&project);
    home.make_go_generation();
    let before = home.snapshot();
    let plan = result(&run(&home, &["ledger-register-plan", &p]));
    assert_eq!(plan["action"], json!("blocked"));
    assert_eq!(plan["problem"], json!("home_not_ready"));
    assert_eq!(plan["would_create"], json!([]));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let out = run(&home, &apply_args(&p, &digest, "approved-by-test"));
    assert_eq!(error_kind(&out), "blocked");
    assert_eq!(home.snapshot(), before);
}

#[test]
fn a_project_without_a_profile_is_not_found_for_every_command() {
    let home = TempHome::new();
    let bare = home.path().join("work/bare");
    fs::create_dir_all(&bare).unwrap();
    let p = project_arg(&bare);
    let before = home.snapshot();
    for args in [
        vec!["ledger-status", p.as_str()],
        vec!["ledger-register-plan", p.as_str()],
    ] {
        assert_eq!(error_kind(&run(&home, &args)), "not_found");
    }
    let out = run(&home, &apply_args(&p, "sha256:00", "approved-by-test"));
    assert_eq!(error_kind(&out), "not_found");
    assert_eq!(home.snapshot(), before);
}

#[test]
fn status_reports_an_unusable_file_and_never_repairs_it() {
    let home = TempHome::new();
    let project = home.project("demo");
    let p = project_arg(&project);
    ledger::register(home.path(), "demo").unwrap();
    let file = ledger::ledger_path(home.path(), "demo").unwrap();

    // Garbage at the ledger path.
    fs::write(&file, b"definitely not a database").unwrap();
    let before = home.snapshot();
    let s = result(&run(&home, &["ledger-status", &p]));
    assert_eq!(s["exists"], json!(true));
    assert_eq!(s["state"], json!("unavailable"));
    assert_eq!(s["problem"], json!("ledger_unavailable"));
    assert_eq!(home.snapshot(), before);

    // A WAL-mode file: reported with wal true, bytes unchanged.
    fs::remove_file(&file).unwrap();
    ledger::register(home.path(), "demo").unwrap();
    {
        let c = rusqlite::Connection::open(&file).unwrap();
        let _: String = c
            .query_row("PRAGMA journal_mode = WAL", [], |r| r.get(0))
            .unwrap();
    }
    let before = home.snapshot();
    let s = result(&run(&home, &["ledger-status", &p]));
    assert_eq!(s["state"], json!("unavailable"));
    assert_eq!(s["problem"], json!("wal_unexpected"));
    assert_eq!(s["wal"], json!(true));
    assert_eq!(home.snapshot(), before);

    // Another project's ledger under this project's name.
    fs::remove_file(&file).unwrap();
    let other = TempHome::new();
    ledger::register(other.path(), "other").unwrap();
    fs::copy(ledger::ledger_path(other.path(), "other").unwrap(), &file).unwrap();
    let s = result(&run(&home, &["ledger-status", &p]));
    assert_eq!(s["problem"], json!("id_collision"));
}

#[test]
fn register_apply_will_not_replace_a_file_that_is_not_this_ledger() {
    let home = TempHome::new();
    let project = home.project("demo");
    let p = project_arg(&project);
    let file = ledger::ledger_path(home.path(), "demo").unwrap();
    fs::create_dir_all(file.parent().unwrap()).unwrap();
    fs::write(&file, b"someone else's file").unwrap();
    let plan = result(&run(&home, &["ledger-register-plan", &p]));
    assert_eq!(plan["state"], json!("unavailable"));
    assert_eq!(plan["action"], json!("none"));
    assert_eq!(plan["problem"], json!("ledger_unavailable"));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let before = home.snapshot();
    let out = run(&home, &apply_args(&p, &digest, "approved-by-test"));
    assert_eq!(error_kind(&out), "blocked");
    assert_eq!(fs::read(&file).unwrap(), b"someone else's file");
    assert_eq!(home.snapshot(), before);
}
