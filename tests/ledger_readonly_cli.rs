//! C-43 AC6, piece B: the card ledger's query commands — `card show`, `card list`, `card next`
//! — leave the whole HOME exactly as they found it: the SQLite ledger and any sidecar file, the
//! records the older system keeps beside the project (a claim journal, a session entry, a
//! proposal), and the git worktree recorded on the reservation of a real repository the test
//! creates. The proof compares a before/after snapshot of every entry under the HOME, folder
//! modification times included (`TempHome::snapshot`), so anything a query creates, rewrites or
//! merely touches fails. The library half of the criterion — a write on a read-only connection
//! fails and changes nothing — is
//! `tests/ledger_open.rs::a_write_on_a_read_only_connection_fails_and_changes_nothing`.
//! Every case runs in a temporary HOME with a temporary repository; the real `~/.moai` is not
//! read or written.

mod common;

use common::git;
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::io::Write as _;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const HOLDER: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const BYSTANDER: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

fn run(home: &TempHome, session: &str, args: &[&str]) -> Output {
    home.command(args)
        .env(SESSION_VAR, session)
        .output()
        .expect("run moai")
}

/// A command whose success envelope is unwrapped, with both pipes shown on failure.
fn ok(home: &TempHome, session: &str, args: &[&str]) -> Value {
    let out = run(home, session, args);
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// `card note`, whose text arrives on stdin.
fn note(home: &TempHome, session: &str, card: &str, project: &str, text: &str) -> Output {
    let mut child = home
        .command(&["card", "note", card, "--kind", "comment", "--project", project])
        .env(SESSION_VAR, session)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .expect("spawn card note");
    let mut pipe = child.stdin.take().unwrap();
    // The child may exit before reading all of stdin; its exit status is what is asserted,
    // never that the write completed.
    if let Err(e) = pipe.write_all(text.as_bytes()) {
        assert_eq!(
            e.kind(),
            std::io::ErrorKind::BrokenPipe,
            "card note stdin write failed"
        );
    }
    drop(pipe);
    child.wait_with_output().expect("card note answer")
}

/// The world criterion 6 names, under one temporary HOME: a real git repository with a worktree
/// the reservation records, a registered ledger holding one taken card with a note, and the
/// older system's records beside the project (a claim journal, a session entry, a proposal).
/// Returns the HOME, the project path and the card's name.
fn world() -> (TempHome, String, String) {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();

    // A real repository with one commit, and a worktree of it the reservation will record.
    let repo = project.join("repo");
    fs::create_dir_all(&repo).unwrap();
    fs::write(repo.join("lib.rs"), "pub fn a() {}\n").unwrap();
    git::git_ok(&repo, &["init", "--quiet"]);
    git::git_in(&repo, &["add", "."]);
    git::git_in(&repo, &["commit", "--quiet", "-m", "first"]);
    let wt = home.path().join("wt");
    git::git_ok(
        &repo,
        &[
            "worktree",
            "add",
            "--quiet",
            wt.to_str().unwrap(),
            "-b",
            "card/demo-t",
        ],
    );
    // Work waiting in the worktree, like a lane's uncommitted change.
    fs::write(wt.join("note.md"), "work in the worktree\n").unwrap();

    // The older system's records: a claim journal and proposals beside the project state, a
    // session entry where a real checkout's clone-common area would keep it.
    let state = project.join(".moai-cli");
    fs::write(state.join("journal.log"), "claim session=a lane=agent-9\n").unwrap();
    fs::create_dir_all(state.join("proposals")).unwrap();
    fs::write(state.join("proposals").join("p-1.md"), "# proposal\nkeep\n").unwrap();
    let sessions = project.join(".git").join(".moai-cli").join("sessions");
    fs::create_dir_all(&sessions).unwrap();
    fs::write(
        sessions.join("agent-9.json"),
        "{\"schema_version\":1,\"name\":\"agent-9\"}\n",
    )
    .unwrap();

    let p = project.to_str().unwrap().to_string();
    let added = ok(
        &home,
        HOLDER,
        &[
            "card",
            "add",
            "--title",
            "t",
            "--scope",
            "apps/demo/",
            "--project",
            &p,
        ],
    );
    let name = added["name"].as_str().unwrap().to_string();
    ok(&home, HOLDER, &["card", "ready", &name, "--project", &p]);
    let taken = ok(
        &home,
        HOLDER,
        &[
            "card",
            "take",
            &name,
            "--project",
            &p,
            "--worktree",
            wt.to_str().unwrap(),
            "--branch",
            "card/demo-t",
        ],
    );
    assert_eq!(taken["state"], json!("active"), "{taken}");
    let noted = note(&home, HOLDER, &name, &p, "a note the queries must not disturb\n");
    assert_eq!(
        noted.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&noted.stdout)
    );
    (home, p, name)
}

#[test]
fn show_and_list_leave_the_home_byte_for_byte_as_it_was() {
    let (home, project, name) = world();
    let before = home.snapshot();

    let shown = ok(
        &home,
        BYSTANDER,
        &["card", "show", &name, "--project", &project],
    );
    assert_eq!(shown["name"], json!(name), "{shown}");
    let notes = shown["notes"].as_array().unwrap();
    assert!(!notes.is_empty(), "the card carries its note: {shown}");

    let listed = ok(&home, BYSTANDER, &["card", "list", &project]);
    assert_eq!(listed["count"], json!(1), "{listed}");
    assert_eq!(listed["cards"][0]["name"], json!(name), "{listed}");

    assert_eq!(
        home.snapshot(),
        before,
        "card show or card list changed the HOME"
    );
}

#[test]
fn next_answers_without_touching_the_home() {
    let (home, project, _name) = world();
    let before = home.snapshot();

    let answered = ok(&home, BYSTANDER, &["card", "next", &project]);
    assert!(
        answered["next"].is_null(),
        "nothing is offered to the bystander: {answered}"
    );

    assert_eq!(home.snapshot(), before, "card next changed the HOME");
}
