//! `moai card ready | take | release | handoff` (Issue #3070 K4a): the dispatch path of a card
//! in an isolated HOME: publish (`add`), `ready`, `take`, a second `take` refused, `handoff`,
//! a stale `release` refused, `release`. The reservation rules themselves are #3068 P3's
//! (`tests/ledger_reservation.rs`); here the commands, their answers and their refusal kinds.
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::path::Path;
use std::process::{Output, Stdio};

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

fn registered() -> (TempHome, String) {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    (home, p(&project))
}

/// Adds a card with one scope and makes it ready; returns its name.
fn ready_card(home: &TempHome, project: &str, scope: &str) -> String {
    let added = result(&run(
        home,
        &[
            "card",
            "add",
            "--title",
            "t",
            "--scope",
            scope,
            "--project",
            project,
        ],
    ));
    let name = added["name"].as_str().unwrap().to_string();
    result(&run(home, &["card", "ready", &name, "--project", project]));
    name
}

fn take(home: &TempHome, project: &str, card: &str, holder: &str) -> Output {
    run(
        home,
        &[
            "card",
            "take",
            card,
            "--holder",
            holder,
            "--project",
            project,
        ],
    )
}

fn count(home: &TempHome, sql: &str) -> i64 {
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
}

#[test]
fn a_card_goes_from_draft_to_taken_handed_over_and_released() {
    let (home, project) = registered();
    let added = result(&run(
        &home,
        &[
            "card",
            "add",
            "--title",
            "dispatch me",
            "--ac",
            "it ships",
            "--scope",
            "src/a.rs",
            "--project",
            &project,
        ],
    ));
    assert_eq!(added["state"], json!("draft"));

    // A draft card cannot be taken.
    let (kind, message) = refusal(&take(&home, &project, "C-1", "agent-a"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("`draft`"), "{message}");

    let ready = result(&run(
        &home,
        &[
            "card",
            "ready",
            "C-1",
            "--by",
            "leader",
            "--project",
            &project,
        ],
    ));
    assert_eq!(ready["from"], json!("draft"));
    assert_eq!(ready["state"], json!("ready"));
    assert_eq!(ready["events"], json!(1));

    // Take: the card's own scope is used; the answer carries what release and handoff need.
    let taken = result(&run(
        &home,
        &[
            "card",
            "take",
            "C-1",
            "--holder",
            "agent-a",
            "--lease-secs",
            "3600",
            "--worktree",
            "/w/a",
            "--branch",
            "issue/1-a",
            "--project",
            &project,
        ],
    ));
    let r = &taken["reservation"];
    assert_eq!(taken["state"], json!("active"));
    assert_eq!(r["generation"], json!(1));
    assert_eq!(r["holder"], json!("agent-a"));
    assert_eq!(r["status"], json!("held"));
    assert_eq!(r["write_scope"], json!(["src/a.rs"]));
    assert_eq!(r["worktree"], json!("/w/a"));
    assert_eq!(r["branch"], json!("issue/1-a"));
    assert_eq!(
        r["lease_until"].as_i64().unwrap() - r["acquired_at"].as_i64().unwrap(),
        3600,
        "the lease is counted from the database clock"
    );
    let shown = result(&run(&home, &["card", "show", "C-1", "--project", &project]));
    assert_eq!(shown["state"], json!("active"));

    // A second take of the same card is refused, even by the same holder.
    for holder in ["agent-b", "agent-a"] {
        let (kind, message) = refusal(&take(&home, &project, "C-1", holder));
        assert_eq!(kind, "conflict", "{message}");
        assert!(message.contains("already held"), "{message}");
    }

    // Hand over: generation 2 for the new holder, same scope, branch and lease end.
    let handed = result(&run(
        &home,
        &[
            "card",
            "handoff",
            "C-1",
            "--holder",
            "agent-a",
            "--generation",
            "1",
            "--to",
            "agent-b",
            "--project",
            &project,
        ],
    ));
    let h = &handed["reservation"];
    assert_eq!(h["generation"], json!(2));
    assert_eq!(h["holder"], json!("agent-b"));
    assert_eq!(h["write_scope"], json!(["src/a.rs"]));
    assert_eq!(h["branch"], json!("issue/1-a"));
    assert_eq!(h["lease_until"], r["lease_until"]);
    assert_eq!(handed["from_holder"], json!("agent-a"));

    // The old holder with the old generation is refused and the card stays taken.
    let (kind, message) = refusal(&run(
        &home,
        &[
            "card",
            "release",
            "C-1",
            "--holder",
            "agent-a",
            "--generation",
            "1",
            "--project",
            &project,
        ],
    ));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("stale"), "{message}");
    assert_eq!(
        count(
            &home,
            "SELECT count(*) FROM event WHERE kind = 'fence_reject'"
        ),
        1,
        "the refusal is recorded once"
    );
    let shown = result(&run(&home, &["card", "show", "C-1", "--project", &project]));
    assert_eq!(shown["state"], json!("active"));

    // The new holder gives it back: the card is ready again and can be taken at generation 3.
    let released = result(&run(
        &home,
        &[
            "card",
            "release",
            "C-1",
            "--holder",
            "agent-b",
            "--generation",
            "2",
            "--project",
            &project,
        ],
    ));
    assert_eq!(released["state"], json!("ready"));
    assert_eq!(released["reservation"]["status"], json!("released"));
    let again = result(&take(&home, &project, "C-1", "agent-c"));
    assert_eq!(again["reservation"]["generation"], json!(3));
}

#[test]
fn six_takes_at_once_give_the_card_to_exactly_one() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/race.rs");
    let children: Vec<_> = (0..6)
        .map(|i| {
            home.command(&[
                "card",
                "take",
                &card,
                "--holder",
                &format!("agent-{i}"),
                "--project",
                &project,
            ])
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap()
        })
        .collect();
    let (mut won, mut lost) = (Vec::new(), 0);
    for child in children {
        let out = child.wait_with_output().unwrap();
        if out.status.code() == Some(0) {
            won.push(
                result(&out)["reservation"]["holder"]
                    .as_str()
                    .unwrap()
                    .to_string(),
            );
        } else {
            let (kind, message) = refusal(&out);
            assert_eq!(kind, "conflict", "{message}");
            lost += 1;
        }
    }
    assert_eq!((won.len(), lost), (1, 5), "winners {won:?}");
    assert_eq!(
        count(
            &home,
            "SELECT count(*) FROM reservation WHERE status = 'held'"
        ),
        1
    );
}

#[test]
fn two_cards_with_overlapping_scopes_cannot_both_be_taken() {
    let (home, project) = registered();
    let a = ready_card(&home, &project, "src/shared/**");
    let b = ready_card(&home, &project, "src/shared/x.rs");
    result(&take(&home, &project, &a, "agent-a"));
    let (kind, message) = refusal(&take(&home, &project, &b, "agent-b"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("write scope"), "{message}");
    let shown = result(&run(&home, &["card", "show", &b, "--project", &project]));
    assert_eq!(
        shown["state"],
        json!("ready"),
        "the refused card is untouched"
    );
}

#[test]
fn take_uses_the_given_scope_over_the_cards_own_and_needs_one() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/own.rs");
    let t = result(&run(
        &home,
        &[
            "card",
            "take",
            &card,
            "--holder",
            "a",
            "--scope",
            "lib/x.rs",
            "--scope",
            "lib/y.rs",
            "--project",
            &project,
        ],
    ));
    assert_eq!(
        t["reservation"]["write_scope"],
        json!(["lib/x.rs", "lib/y.rs"])
    );
    // A card without any scope, taken without one: refused as input, nothing stored.
    result(&run(
        &home,
        &["card", "add", "--title", "no scope", "--project", &project],
    ));
    result(&run(
        &home,
        &["card", "ready", "C-2", "--project", &project],
    ));
    let (kind, message) = refusal(&take(&home, &project, "C-2", "a"));
    assert_eq!(kind, "invalid_input", "{message}");
    assert_eq!(
        count(&home, "SELECT count(*) FROM reservation WHERE card_id = 2"),
        0
    );
}

#[test]
fn the_lease_and_the_holder_have_hard_edges() {
    let (home, project) = registered();
    let mut n = 0;
    let mut try_take = |secs: &str, holder: &str| -> Output {
        n += 1;
        let card = ready_card(&home, &project, &format!("src/edge{n}.rs"));
        run(
            &home,
            &[
                "card",
                "take",
                &card,
                "--holder",
                holder,
                "--lease-secs",
                secs,
                "--project",
                &project,
            ],
        )
    };
    for (secs, ok) in [
        ("59", false),
        ("60", true),
        ("86400", true),
        ("86401", false),
    ] {
        let out = try_take(secs, "a");
        if ok {
            result(&out);
        } else {
            assert_eq!(refusal(&out).0, "invalid_input", "{secs} seconds");
        }
    }
    let h200 = "h".repeat(200);
    let h201 = "h".repeat(201);
    result(&try_take("3600", &h200));
    assert_eq!(refusal(&try_take("3600", &h201)).0, "invalid_input");
    assert_eq!(refusal(&try_take("3600", "")).0, "invalid_input");
    assert_eq!(refusal(&try_take("3600", "a\nb")).0, "invalid_input");
}

#[test]
fn ready_refuses_what_is_not_a_draft_or_hold_card() {
    let (home, project) = registered();
    result(&run(
        &home,
        &["card", "add", "--title", "t", "--project", &project],
    ));
    result(&run(
        &home,
        &["card", "ready", "C-1", "--project", &project],
    ));
    // Already ready.
    let (kind, message) = refusal(&run(
        &home,
        &["card", "ready", "C-1", "--project", &project],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.starts_with("card ready refused: "), "{message}");
    // Unknown card, host number, bad name, bad actor.
    assert_eq!(
        refusal(&run(
            &home,
            &["card", "ready", "C-9", "--project", &project]
        ))
        .0,
        "not_found"
    );
    for bad in ["#12", "12", "C-0"] {
        let (kind, message) = refusal(&run(&home, &["card", "ready", bad, "--project", &project]));
        assert_eq!(kind, "invalid_input", "{bad}: {message}");
    }
    result(&run(
        &home,
        &["card", "add", "--title", "u", "--project", &project],
    ));
    let (kind, _) = refusal(&run(
        &home,
        &["card", "ready", "C-2", "--by", "", "--project", &project],
    ));
    assert_eq!(kind, "invalid_input");
    // A hold card becomes ready again.
    let conn =
        rusqlite::Connection::open(ledger::ledger_path(home.path(), "demo").unwrap()).unwrap();
    conn.execute("UPDATE card SET state = 'hold' WHERE card_id = 2", [])
        .unwrap();
    drop(conn);
    let r = result(&run(
        &home,
        &["card", "ready", "C-2", "--project", &project],
    ));
    assert_eq!(r["from"], json!("hold"));
    // Each successful ready wrote exactly one event with the actor.
    assert_eq!(
        count(
            &home,
            "SELECT count(*) FROM event WHERE kind = 'card.ready'"
        ),
        2
    );
}

#[test]
fn a_ready_that_fails_leaves_the_card_and_the_events_alone() {
    let (home, project) = registered();
    result(&run(
        &home,
        &["card", "add", "--title", "t", "--project", &project],
    ));
    let conn =
        rusqlite::Connection::open(ledger::ledger_path(home.path(), "demo").unwrap()).unwrap();
    conn.execute_batch(
        "CREATE TRIGGER test_no_ready BEFORE INSERT ON event WHEN NEW.kind = 'card.ready' \
         BEGIN SELECT RAISE(ABORT, 'no ready events'); END;",
    )
    .unwrap();
    let (kind, message) = refusal(&run(
        &home,
        &["card", "ready", "C-1", "--project", &project],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    let shown = result(&run(&home, &["card", "show", "C-1", "--project", &project]));
    assert_eq!(
        shown["state"],
        json!("draft"),
        "the state change was rolled back"
    );
    assert_eq!(shown["rev"], json!(1));
}

#[test]
fn without_a_ledger_every_action_is_not_found_and_creates_nothing() {
    let home = TempHome::new();
    let project = p(&home.project("demo"));
    let before = home.snapshot();
    for args in [
        vec!["card", "ready", "C-1", "--project", project.as_str()],
        vec![
            "card",
            "take",
            "C-1",
            "--holder",
            "a",
            "--project",
            project.as_str(),
        ],
        vec![
            "card",
            "release",
            "C-1",
            "--holder",
            "a",
            "--generation",
            "1",
            "--project",
            project.as_str(),
        ],
        vec![
            "card",
            "handoff",
            "C-1",
            "--holder",
            "a",
            "--generation",
            "1",
            "--to",
            "b",
            "--project",
            project.as_str(),
        ],
    ] {
        let (kind, message) = refusal(&run(&home, &args));
        assert_eq!(kind, "not_found", "{args:?}: {message}");
        assert!(message.contains("ledger-register-plan"), "{message}");
    }
    assert_eq!(home.snapshot(), before);
}

#[test]
fn a_release_or_handoff_by_the_wrong_party_or_at_a_wrong_generation_is_refused() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/w.rs");
    result(&take(&home, &project, &card, "owner"));
    for (holder, generation) in [("intruder", "1"), ("owner", "2"), ("owner", "0")] {
        let (kind, message) = refusal(&run(
            &home,
            &[
                "card",
                "release",
                &card,
                "--holder",
                holder,
                "--generation",
                generation,
                "--project",
                &project,
            ],
        ));
        assert!(
            kind == "conflict" || kind == "invalid_input",
            "{holder} {generation}: {kind} {message}"
        );
    }
    let (kind, _) = refusal(&run(
        &home,
        &[
            "card",
            "handoff",
            &card,
            "--holder",
            "owner",
            "--generation",
            "1",
            "--to",
            "owner",
            "--project",
            &project,
        ],
    ));
    assert_eq!(kind, "invalid_input", "a hand-over to oneself");
    // The reservation is still the owner's at generation 1.
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    let cur = ledger::reservation::current(l.conn(), 1).unwrap().unwrap();
    assert_eq!((cur.holder.as_str(), cur.generation), ("owner", 1));
}
