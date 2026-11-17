//! `moai card list | show | add` (Issue #3070 C1): what they print, what they store, what they
//! refuse and that no action creates the ledger. Every command runs against a temporary HOME;
//! the real `~/.moai` is not read or written.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::path::Path;
use std::process::Output;

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().expect("run moai")
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
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("card"), "{v}");
    v["result"].clone()
}

/// The refusal: exit 1, its `kind` and its message.
fn refusal(out: &Output) -> (String, String) {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

fn p(project: &Path) -> String {
    project.to_str().unwrap().to_string()
}

/// A HOME with a project `demo` and its registered ledger.
fn registered() -> (TempHome, String) {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    (home, p(&project))
}

fn count(home: &TempHome, table: &str) -> i64 {
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    l.conn()
        .query_row(&format!("SELECT count(*) FROM {table}"), [], |r| r.get(0))
        .unwrap()
}

fn raw(home: &TempHome) -> rusqlite::Connection {
    rusqlite::Connection::open(ledger::ledger_path(home.path(), "demo").unwrap()).unwrap()
}

#[test]
fn list_without_a_ledger_is_empty_and_creates_nothing() {
    let home = TempHome::new();
    let project = home.project("demo");
    let before = home.snapshot();
    let r = result(&run(&home, &["card", "list", &p(&project)]));
    assert_eq!(r["action"], json!("list"));
    assert_eq!(r["ledger"], json!("absent"));
    assert_eq!(r["count"], json!(0));
    assert_eq!(r["cards"], json!([]));
    assert_eq!(r["writes"], json!(false));
    assert_eq!(
        home.snapshot(),
        before,
        "no folder, no file, no mtime change"
    );
}

#[test]
fn add_and_show_without_a_ledger_are_not_found_and_create_nothing() {
    let home = TempHome::new();
    let project = home.project("demo");
    let pr = p(&project);
    let before = home.snapshot();
    for args in [
        vec!["card", "add", "--title", "T", "--project", pr.as_str()],
        vec!["card", "show", "C-1", "--project", pr.as_str()],
    ] {
        let (kind, message) = refusal(&run(&home, &args));
        assert_eq!(kind, "not_found", "{message}");
        assert!(
            message.contains("ledger-register-plan") && message.contains("ledger-register-apply"),
            "{message}"
        );
    }
    assert_eq!(home.snapshot(), before, "add must not register the ledger");
}

#[test]
fn a_card_goes_in_and_comes_back_out() {
    let (home, project) = registered();
    let added = result(&run(
        &home,
        &[
            "card",
            "add",
            "--title",
            "First card",
            "--body",
            "why",
            "--kind",
            "fix",
            "--priority",
            "p1",
            "--scope",
            "src/a.rs",
            "--scope",
            "docs",
            "--ac",
            "it works",
            "--ac",
            "it is tested",
            "--by",
            "agent-7",
            "--project",
            &project,
        ],
    ));
    assert_eq!(added["name"], json!("C-1"));
    assert_eq!(added["state"], json!("draft"));
    assert_eq!(added["criteria"], json!(2));
    assert_eq!(added["events"], json!(1));
    assert_eq!(added["writes"], json!(true));
    // One revision for the card and one for each criterion (the schema's own trigger).
    assert_eq!(added["rev"], json!(3));
    result(&run(
        &home,
        &["card", "add", "--title", "Second", "--project", &project],
    ));

    let list = result(&run(&home, &["card", "list", &project]));
    assert_eq!(list["ledger"], json!("ready"));
    assert_eq!(list["count"], json!(2));
    let names: Vec<&str> = list["cards"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap())
        .collect();
    assert_eq!(names, ["C-1", "C-2"], "oldest first");
    assert_eq!(list["cards"][0]["kind"], json!("fix"));
    assert_eq!(list["cards"][1]["kind"], json!("task"), "the default kind");
    assert_eq!(
        list["cards"][1]["priority"],
        json!("p2"),
        "the default priority"
    );

    let show = result(&run(&home, &["card", "show", "C-1", "--project", &project]));
    assert_eq!(show["title"], json!("First card"));
    assert_eq!(show["body"], json!("why"));
    assert_eq!(show["priority"], json!("p1"));
    assert_eq!(show["write_scope"], json!(["src/a.rs", "docs"]));
    assert_eq!(show["created_by"], json!("agent-7"));
    assert_eq!(show["rev"], json!(3));
    assert_eq!(show["source"], Value::Null);
    let seqs: Vec<i64> = show["criteria"]
        .as_array()
        .unwrap()
        .iter()
        .map(|a| a["seq"].as_i64().unwrap())
        .collect();
    assert_eq!(seqs, [1, 2], "numbered from 1 in the order given");
    assert_eq!(show["criteria"][1]["text"], json!("it is tested"));
    assert_eq!(show["criteria"][0]["status"], json!("open"));
    assert_eq!(show["notes"], json!([]));
    assert_eq!(show["links"], json!([]));
    assert_eq!(show["writes"], json!(false));
    // Without `--by` the actor is `user`, in the card and in the add's answer.
    let second = result(&run(&home, &["card", "show", "C-2", "--project", &project]));
    assert_eq!(second["created_by"], json!("user"));
    assert_eq!(second["rev"], json!(1), "no criterion, no extra revision");

    // One event per add, with the actor and the card.
    assert_eq!(count(&home, "event"), 2);
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    let (kind, actor, card): (String, String, i64) = l
        .conn()
        .query_row(
            "SELECT kind, actor, card_id FROM event ORDER BY event_id LIMIT 1",
            [],
            |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?)),
        )
        .unwrap();
    assert_eq!(
        (kind.as_str(), actor.as_str(), card),
        ("card.add", "agent-7", 1)
    );
}

#[test]
fn list_and_show_leave_the_ledger_file_alone() {
    let (home, project) = registered();
    result(&run(
        &home,
        &["card", "add", "--title", "T", "--project", &project],
    ));
    let before = home.snapshot();
    result(&run(&home, &["card", "list", &project]));
    result(&run(&home, &["card", "show", "C-1", "--project", &project]));
    assert_eq!(
        home.snapshot(),
        before,
        "reading writes nothing, not even a journal"
    );
}

#[test]
fn an_add_that_the_ledger_refuses_leaves_no_card_no_criterion_and_no_event() {
    let (home, project) = registered();
    // The event is the last row the add stores; a ledger that refuses it must take back the
    // card and its criteria too.
    raw(&home)
        .execute_batch(
            "CREATE TRIGGER test_no_events BEFORE INSERT ON event \
             BEGIN SELECT RAISE(ABORT, 'no events today'); END;",
        )
        .unwrap();
    let (kind, message) = refusal(&run(
        &home,
        &[
            "card",
            "add",
            "--title",
            "T",
            "--ac",
            "a",
            "--ac",
            "b",
            "--project",
            &project,
        ],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("ledger write refused"), "{message}");
    assert_eq!(count(&home, "card"), 0);
    assert_eq!(count(&home, "card_ac"), 0);
    assert_eq!(count(&home, "event"), 0);
    raw(&home)
        .execute_batch("DROP TRIGGER test_no_events;")
        .unwrap();
    result(&run(
        &home,
        &["card", "add", "--title", "T", "--project", &project],
    ));
    assert_eq!(count(&home, "card"), 1);
    // The refused add used up no card number visible to the user: the next card is C-1.
    let show = result(&run(&home, &["card", "show", "C-1", "--project", &project]));
    assert_eq!(show["title"], json!("T"));
}

#[test]
fn a_refused_input_is_invalid_input_and_stores_nothing() {
    let (home, project) = registered();
    let long_ok = "x".repeat(300);
    let long_bad = "x".repeat(301);
    result(&run(
        &home,
        &["card", "add", "--title", &long_ok, "--project", &project],
    ));
    assert_eq!(count(&home, "card"), 1, "300 characters fit");
    for args in [
        vec![
            "card",
            "add",
            "--title",
            long_bad.as_str(),
            "--project",
            project.as_str(),
        ],
        vec!["card", "add", "--title", "", "--project", project.as_str()],
        vec![
            "card",
            "add",
            "--title",
            "a\nb",
            "--project",
            project.as_str(),
        ],
        vec![
            "card",
            "add",
            "--title",
            "t",
            "--kind",
            "nope",
            "--project",
            project.as_str(),
        ],
        vec![
            "card",
            "add",
            "--title",
            "t",
            "--priority",
            "p9",
            "--project",
            project.as_str(),
        ],
        vec![
            "card",
            "add",
            "--title",
            "t",
            "--scope",
            "../x",
            "--project",
            project.as_str(),
        ],
        vec![
            "card",
            "add",
            "--title",
            "t",
            "--ac",
            "",
            "--project",
            project.as_str(),
        ],
        vec![
            "card",
            "add",
            "--title",
            "t",
            "--by",
            "",
            "--project",
            project.as_str(),
        ],
    ] {
        let (kind, message) = refusal(&run(&home, &args));
        assert_eq!(kind, "invalid_input", "{args:?}: {message}");
        assert!(message.starts_with("card add refused: "), "{message}");
    }
    assert_eq!(count(&home, "card"), 1, "nothing else was stored");
    assert_eq!(count(&home, "event"), 1);
}

#[test]
fn a_by_of_eighty_characters_fits_and_eighty_one_do_not() {
    let (home, project) = registered();
    let ok = "b".repeat(80);
    let bad = "b".repeat(81);
    result(&run(
        &home,
        &[
            "card",
            "add",
            "--title",
            "t",
            "--by",
            &ok,
            "--project",
            &project,
        ],
    ));
    let (kind, _) = refusal(&run(
        &home,
        &[
            "card",
            "add",
            "--title",
            "t",
            "--by",
            &bad,
            "--project",
            &project,
        ],
    ));
    assert_eq!(kind, "invalid_input");
    assert_eq!(count(&home, "card"), 1);
}

#[test]
fn list_filters_by_state_and_refuses_an_unknown_one() {
    let (home, project) = registered();
    result(&run(
        &home,
        &["card", "add", "--title", "a", "--project", &project],
    ));
    result(&run(
        &home,
        &["card", "add", "--title", "b", "--project", &project],
    ));
    raw(&home)
        .execute_batch("UPDATE card SET state = 'ready' WHERE card_id = 2;")
        .unwrap();
    let ready = result(&run(&home, &["card", "list", &project, "--state", "ready"]));
    assert_eq!(ready["count"], json!(1));
    assert_eq!(ready["cards"][0]["name"], json!("C-2"));
    let draft = result(&run(&home, &["card", "list", &project, "--state", "draft"]));
    assert_eq!(draft["cards"][0]["name"], json!("C-1"));
    let done = result(&run(&home, &["card", "list", &project, "--state", "done"]));
    assert_eq!(done["count"], json!(0));
    let (kind, message) = refusal(&run(&home, &["card", "list", &project, "--state", "nope"]));
    assert_eq!(kind, "invalid_input", "{message}");
}

#[test]
fn a_host_number_names_a_card_only_when_it_names_exactly_one() {
    let (home, project) = registered();
    raw(&home)
        .execute_batch(
            "INSERT INTO card (title, created_by, created_at, updated_at, source_host, \
                               source_project, source_number) \
             VALUES ('one', 'x', 1, 1, 'github', 'o/a', 12), \
                    ('two', 'x', 1, 1, 'github', 'o/a', 13), \
                    ('three', 'x', 1, 1, 'github', 'o/b', 13);",
        )
        .unwrap();
    let one = result(&run(&home, &["card", "show", "#12", "--project", &project]));
    assert_eq!(one["name"], json!("C-1"));
    assert_eq!(one["source"], json!("github:o/a#12"));
    let (kind, message) = refusal(&run(&home, &["card", "show", "#13", "--project", &project]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(
        message.contains("C-2") && message.contains("C-3"),
        "{message}"
    );
    let (kind, _) = refusal(&run(&home, &["card", "show", "#99", "--project", &project]));
    assert_eq!(kind, "not_found");
    let (kind, _) = refusal(&run(
        &home,
        &["card", "show", "C-99", "--project", &project],
    ));
    assert_eq!(kind, "not_found");
    for bad in ["12", "C-0", "C-01", "#0", "C-", "x"] {
        let (kind, message) = refusal(&run(&home, &["card", "show", bad, "--project", &project]));
        assert_eq!(kind, "invalid_input", "{bad}: {message}");
    }
}

#[test]
fn show_lists_notes_and_links_in_their_order() {
    let (home, project) = registered();
    result(&run(
        &home,
        &["card", "add", "--title", "a", "--project", &project],
    ));
    result(&run(
        &home,
        &["card", "add", "--title", "b", "--project", &project],
    ));
    raw(&home)
        .execute_batch(
            "INSERT INTO card_note (card_id, kind, author, body, created_at) \
               VALUES (1, 'comment', 'x', 'first', 1), (1, 'decision', 'y', 'second', 2); \
             INSERT INTO card_link (from_card, kind, to_card, created_at, updated_at) \
               VALUES (1, 'relates', 2, 1, 1); \
             INSERT INTO card_link (from_card, kind, target_host, target_project, target_number, \
                                    pr_state, merge_ref, created_at, updated_at) \
               VALUES (1, 'pr', 'github', 'o/a', 7, 'merged', 'abc', 1, 1);",
        )
        .unwrap();
    let show = result(&run(&home, &["card", "show", "C-1", "--project", &project]));
    let bodies: Vec<&str> = show["notes"]
        .as_array()
        .unwrap()
        .iter()
        .map(|n| n["body"].as_str().unwrap())
        .collect();
    assert_eq!(bodies, ["first", "second"]);
    assert_eq!(show["links"][0]["to"], json!("C-2"));
    assert_eq!(show["links"][1]["to"], json!("github:o/a#7"));
    assert_eq!(show["links"][1]["pr_state"], json!("merged"));
}

#[test]
fn six_adds_at_once_all_land_with_six_different_numbers() {
    let (home, project) = registered();
    let children: Vec<_> = (0..6)
        .map(|i| {
            home.command(&[
                "card",
                "add",
                "--title",
                &format!("card {i}"),
                "--project",
                &project,
            ])
            .stdout(std::process::Stdio::piped())
            .stderr(std::process::Stdio::piped())
            .spawn()
            .unwrap()
        })
        .collect();
    let mut ids = Vec::new();
    for child in children {
        let out = child.wait_with_output().unwrap();
        ids.push(result(&out)["card_id"].as_i64().unwrap());
    }
    ids.sort_unstable();
    assert_eq!(ids, [1, 2, 3, 4, 5, 6], "no number twice, none lost");
    assert_eq!(count(&home, "card"), 6);
    assert_eq!(count(&home, "event"), 6);
}

#[test]
fn every_action_says_first_what_it_does_to_the_ledger() {
    let home = TempHome::new();
    for (action, first) in [
        ("list", "Reads the ledger only."),
        ("show", "Reads the ledger only."),
        ("add", "Writes the ledger."),
        ("ready", "Writes the ledger."),
        ("take", "Writes the ledger."),
        ("release", "Writes the ledger."),
        ("handoff", "Writes the ledger."),
        ("spec", "Writes the ledger."),
    ] {
        let out = run(&home, &["card", action, "--help"]);
        assert!(out.status.success());
        let text = String::from_utf8_lossy(&out.stdout).into_owned();
        let head = text.lines().find(|l| !l.trim().is_empty()).unwrap_or("");
        assert!(
            head.starts_with(first),
            "card {action} --help starts with `{head}`"
        );
        let other = if first.starts_with("Reads") {
            "Writes the ledger."
        } else {
            "Reads the ledger only."
        };
        assert!(!head.contains(other), "card {action}: `{head}`");
    }
}

#[test]
fn a_ledger_that_is_not_usable_is_blocked_for_every_action_and_left_alone() {
    let (home, project) = registered();
    let path = ledger::ledger_path(home.path(), "demo").unwrap();
    std::fs::write(&path, b"this is not a database").unwrap();
    let before = home.snapshot();
    for args in [
        vec!["card", "list", project.as_str()],
        vec!["card", "show", "C-1", "--project", project.as_str()],
        vec!["card", "add", "--title", "t", "--project", project.as_str()],
    ] {
        let (kind, message) = refusal(&run(&home, &args));
        assert_eq!(kind, "blocked", "{args:?}: {message}");
        assert!(message.contains("ledger_unavailable"), "{message}");
    }
    assert_eq!(
        home.snapshot(),
        before,
        "a file that is not the ledger is never repaired"
    );
}

/// Card C-68 N-d: the examples under `moai card --help` run as typed, top to bottom, in a
/// project with a registered ledger. `--project .` is the folder the commands run in, and
/// ` < file` is that file on standard input.
#[test]
fn the_help_examples_of_card_run_as_typed() {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    let (_, examples) = moai_cli::help::EXAMPLES
        .iter()
        .find(|(word, _)| *word == "card")
        .expect("card has examples");
    assert!((1..=3).contains(&examples.len()), "{examples:?}");
    let mut actions = Vec::new();
    for example in *examples {
        let (args, input) = match example.split_once(" < ") {
            Some((args, file)) => (args, Some(file)),
            None => (*example, None),
        };
        let argv: Vec<&str> = args.split_whitespace().collect();
        let mut cmd = home.command(&argv);
        cmd.current_dir(&project);
        match input {
            Some(file) => {
                let path = project.join(file);
                std::fs::write(&path, "the login test passes\n").unwrap();
                cmd.stdin(std::fs::File::open(&path).unwrap());
            }
            None => {
                cmd.stdin(std::process::Stdio::null());
            }
        }
        let out = cmd.output().expect("run moai");
        assert_eq!(
            out.status.code(),
            Some(0),
            "`moai {example}` does not run as typed: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        actions.push(envelope(&out)["result"]["action"].clone());
    }
    // The steps a new card goes through before a session takes it.
    assert_eq!(actions, vec![json!("add"), json!("ready"), json!("note")]);
    let shown = result(&run(
        &home,
        &["card", "show", "C-1", "--project", &p(&project)],
    ));
    assert_eq!(shown["state"], json!("ready"));
    assert_eq!(shown["notes"].as_array().unwrap().len(), 1);
}
