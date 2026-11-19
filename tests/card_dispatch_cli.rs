//! `moai card ready | take | release | handoff` (Issue #3070 K4a): the dispatch path of a card
//! in an isolated HOME: publish (`add`), `ready`, `take`, a second `take` refused, `handoff`,
//! a stale `release` refused, `release`. The reservation rules themselves are #3068 P3's
//! (`tests/ledger_reservation.rs`); here the commands, their answers and their refusal kinds.
//!
//! The holder is never typed: it is the session id the host put in the environment
//! (`CLAUDE_CODE_SESSION_ID`), so a caller that only knows another holder's name cannot act for it
//! (#3115 verdict N1). `--home` is exercised from a second, empty HOME so a command that ignored
//! it would find no ledger. Every command runs against a temporary HOME; the real `~/.moai` is
//! not read or written.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::path::Path;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const S_C: &str = "cccccccc-0000-4000-8000-00000000000c";

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().expect("run moai")
}

/// A command run as the session `session`.
fn as_session(home: &TempHome, session: &str, args: &[&str]) -> Output {
    home.command(args)
        .env(SESSION_VAR, session)
        .output()
        .expect("run moai")
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

fn take(home: &TempHome, project: &str, card: &str, session: &str) -> Output {
    as_session(home, session, &["card", "take", card, "--project", project])
}

fn release(home: &TempHome, project: &str, card: &str, session: &str, generation: &str) -> Output {
    as_session(
        home,
        session,
        &[
            "card",
            "release",
            card,
            "--generation",
            generation,
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
    let (kind, message) = refusal(&take(&home, &project, "C-1", S_A));
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

    // Take: the holder is the session; the card's own scope is used; the answer carries what
    // release and handoff need.
    let taken = result(&as_session(
        &home,
        S_A,
        &[
            "card",
            "take",
            "C-1",
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
    assert_eq!(r["holder"], json!(S_A));
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

    // A second take of the same card is refused, even by the same session.
    for session in [S_B, S_A] {
        let (kind, message) = refusal(&take(&home, &project, "C-1", session));
        assert_eq!(kind, "conflict", "{message}");
        assert!(message.contains("already held"), "{message}");
    }

    // Hand over: generation 2 for the receiver, same scope, branch and lease end.
    let handed = result(&as_session(
        &home,
        S_A,
        &[
            "card",
            "handoff",
            "C-1",
            "--generation",
            "1",
            "--to",
            S_B,
            "--project",
            &project,
        ],
    ));
    let h = &handed["reservation"];
    assert_eq!(h["generation"], json!(2));
    assert_eq!(h["holder"], json!(S_B));
    assert_eq!(h["write_scope"], json!(["src/a.rs"]));
    assert_eq!(h["branch"], json!("issue/1-a"));
    assert_eq!(h["lease_until"], r["lease_until"]);
    assert_eq!(handed["from_holder"], json!(S_A));

    // The old session with the old generation is refused and the card stays taken.
    let (kind, message) = refusal(&release(&home, &project, "C-1", S_A, "1"));
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
    let released = result(&release(&home, &project, "C-1", S_B, "2"));
    assert_eq!(released["state"], json!("ready"));
    assert_eq!(released["reservation"]["status"], json!("released"));
    let again = result(&take(&home, &project, "C-1", S_C));
    assert_eq!(again["reservation"]["generation"], json!(3));
    assert_eq!(again["reservation"]["holder"], json!(S_C));
    let a = &again["reservation"];
    assert_eq!(
        a["lease_until"].as_i64().unwrap() - a["acquired_at"].as_i64().unwrap(),
        14_400,
        "the default lease is four hours"
    );
    // The ready event carries the actor that was given.
    assert_eq!(
        count(
            &home,
            "SELECT count(*) FROM event WHERE kind = 'card.ready' AND actor = 'leader'"
        ),
        1
    );
}

#[test]
fn another_session_cannot_act_for_the_holder_and_a_typed_holder_is_not_accepted() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/own.rs");
    result(&take(&home, &project, &card, S_A));

    // Another session that knows the generation can neither release nor hand over the card.
    let (kind, message) = refusal(&release(&home, &project, &card, S_B, "1"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("held by"), "{message}");
    let (kind, message) = refusal(&as_session(
        &home,
        S_B,
        &[
            "card",
            "handoff",
            &card,
            "--generation",
            "1",
            "--to",
            S_C,
            "--project",
            &project,
        ],
    ));
    assert_eq!(kind, "conflict", "{message}");

    // `--holder` no longer exists: naming the holder does not work, and nothing changes.
    for args in [
        vec![
            "card",
            "release",
            card.as_str(),
            "--holder",
            S_A,
            "--generation",
            "1",
            "--project",
            project.as_str(),
        ],
        vec![
            "card",
            "take",
            card.as_str(),
            "--holder",
            S_B,
            "--project",
            project.as_str(),
        ],
    ] {
        let out = as_session(&home, S_B, &args);
        assert!(!out.status.success(), "{args:?}");
    }
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    let cur = ledger::reservation::current(l.conn(), 1).unwrap().unwrap();
    assert_eq!(
        (cur.holder.as_str(), cur.generation, cur.status.as_str()),
        (S_A, 1, "held")
    );
}

#[test]
fn a_session_without_a_usable_id_holds_nothing_and_changes_nothing() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/none.rs");
    let before = count(&home, "SELECT count(*) FROM event");
    // No variable at all, then values that are not session ids.
    let mut outs = vec![run(&home, &["card", "take", &card, "--project", &project])];
    for bad in ["", "agent-7", "short", "not a session id"] {
        outs.push(as_session(
            &home,
            bad,
            &["card", "take", &card, "--project", &project],
        ));
    }
    for out in outs {
        let (kind, message) = refusal(&out);
        assert_eq!(kind, "invalid_input", "{message}");
        assert!(message.starts_with("card take refused: "), "{message}");
        assert!(message.contains(SESSION_VAR), "{message}");
    }
    for action in ["release", "handoff"] {
        let mut args = vec!["card", action, card.as_str(), "--generation", "1"];
        if action == "handoff" {
            args.extend(["--to", S_B]);
        }
        args.extend(["--project", project.as_str()]);
        let (kind, message) = refusal(&run(&home, &args));
        assert_eq!(kind, "invalid_input", "{action}: {message}");
    }
    assert_eq!(count(&home, "SELECT count(*) FROM reservation"), 0);
    assert_eq!(count(&home, "SELECT count(*) FROM event"), before);
}

#[test]
fn the_receiver_of_a_handoff_must_be_a_session_id() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/recv.rs");
    result(&take(&home, &project, &card, S_A));
    for bad in ["agent-b", "", "abc", S_A] {
        let (kind, message) = refusal(&as_session(
            &home,
            S_A,
            &[
                "card",
                "handoff",
                &card,
                "--generation",
                "1",
                "--to",
                bad,
                "--project",
                &project,
            ],
        ));
        assert_eq!(kind, "invalid_input", "{bad:?}: {message}");
    }
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    let cur = ledger::reservation::current(l.conn(), 1).unwrap().unwrap();
    assert_eq!((cur.holder.as_str(), cur.generation), (S_A, 1));
}

#[test]
fn six_takes_at_once_give_the_card_to_exactly_one() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/race.rs");
    let children: Vec<_> = (0..6)
        .map(|i| {
            home.command(&["card", "take", &card, "--project", &project])
                .env(
                    SESSION_VAR,
                    format!("dddddddd-0000-4000-8000-00000000000{i}"),
                )
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
    result(&take(&home, &project, &a, S_A));
    let (kind, message) = refusal(&take(&home, &project, &b, S_B));
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
    let t = result(&as_session(
        &home,
        S_A,
        &[
            "card",
            "take",
            &card,
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
    let (kind, message) = refusal(&take(&home, &project, "C-2", S_B));
    assert_eq!(kind, "invalid_input", "{message}");
    assert_eq!(
        count(&home, "SELECT count(*) FROM reservation WHERE card_id = 2"),
        0
    );
}

#[test]
fn the_lease_has_hard_edges() {
    let (home, project) = registered();
    let mut n = 0;
    let mut try_take = |secs: &str| -> Output {
        n += 1;
        let card = ready_card(&home, &project, &format!("src/edge{n}.rs"));
        as_session(
            &home,
            S_A,
            &[
                "card",
                "take",
                &card,
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
        let out = try_take(secs);
        if ok {
            result(&out);
        } else {
            assert_eq!(refusal(&out).0, "invalid_input", "{secs} seconds");
        }
    }
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
    // Each successful ready wrote exactly one event.
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
        vec!["card", "take", "C-1", "--project", project.as_str()],
        vec![
            "card",
            "release",
            "C-1",
            "--generation",
            "1",
            "--project",
            project.as_str(),
        ],
        vec![
            "card",
            "handoff",
            "C-1",
            "--generation",
            "1",
            "--to",
            S_B,
            "--project",
            project.as_str(),
        ],
    ] {
        let (kind, message) = refusal(&as_session(&home, S_A, &args));
        assert_eq!(kind, "not_found", "{args:?}: {message}");
        assert!(message.contains("ledger-register-plan"), "{message}");
    }
    assert_eq!(home.snapshot(), before);
}

/// `--home` names the HOME that holds the ledger. Each command runs from a second, empty HOME:
/// with `--home` it finds the ledger, without it it finds none. A command that ignored `--home`
/// would fail the first half.
#[test]
fn the_home_option_is_the_home_that_is_used() {
    let (real, project) = registered();
    let decoy = TempHome::new();
    let real_home = real.path().to_str().unwrap().to_string();
    let h = |args: &[&str]| -> Vec<String> {
        let mut v: Vec<String> = args.iter().map(|s| s.to_string()).collect();
        v.extend(["--home".into(), real_home.clone()]);
        v
    };
    let from_decoy = |args: Vec<String>| -> Output {
        let refs: Vec<&str> = args.iter().map(String::as_str).collect();
        decoy
            .command(&refs)
            .env(SESSION_VAR, S_A)
            .output()
            .expect("run moai")
    };
    let r = result(&from_decoy(h(&[
        "card",
        "add",
        "--title",
        "far away",
        "--scope",
        "src/far.rs",
        "--project",
        &project,
    ])));
    assert_eq!(r["name"], json!("C-1"));
    assert_eq!(count(&real, "SELECT count(*) FROM card"), 1);
    result(&from_decoy(h(&[
        "card",
        "ready",
        "C-1",
        "--project",
        &project,
    ])));
    let t = result(&from_decoy(h(&[
        "card",
        "take",
        "C-1",
        "--project",
        &project,
    ])));
    assert_eq!(t["reservation"]["generation"], json!(1));
    let list = result(&from_decoy(h(&["card", "list", &project])));
    assert_eq!(list["count"], json!(1));
    assert_eq!(list["cards"][0]["state"], json!("active"));
    let show = result(&from_decoy(h(&[
        "card",
        "show",
        "C-1",
        "--project",
        &project,
    ])));
    assert_eq!(show["title"], json!("far away"));
    let hand = result(&from_decoy(h(&[
        "card",
        "handoff",
        "C-1",
        "--generation",
        "1",
        "--to",
        S_B,
        "--project",
        &project,
    ])));
    assert_eq!(hand["reservation"]["generation"], json!(2));
    let args = h(&[
        "card",
        "release",
        "C-1",
        "--generation",
        "2",
        "--project",
        &project,
    ]);
    let refs: Vec<&str> = args.iter().map(String::as_str).collect();
    let released = result(
        &decoy
            .command(&refs)
            .env(SESSION_VAR, S_B)
            .output()
            .expect("run moai"),
    );
    assert_eq!(released["state"], json!("ready"));

    // Without `--home` the decoy HOME holds no ledger and nothing is created there.
    let before = decoy.snapshot();
    let list = result(&decoy.command(&["card", "list", &project]).output().unwrap());
    assert_eq!(list["ledger"], json!("absent"));
    let (kind, _) = refusal(
        &decoy
            .command(&["card", "add", "--title", "t", "--project", &project])
            .output()
            .unwrap(),
    );
    assert_eq!(kind, "not_found");
    assert_eq!(decoy.snapshot(), before);
}

#[test]
fn a_handoff_with_an_old_generation_is_refused_even_by_the_same_session() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/gen.rs");
    // The same session takes (g1), gives back, takes again (g2).
    result(&take(&home, &project, &card, S_A));
    result(&release(&home, &project, &card, S_A, "1"));
    let again = result(&take(&home, &project, &card, S_A));
    assert_eq!(again["reservation"]["generation"], json!(2));
    let handoff = |generation: &str| -> Output {
        as_session(
            &home,
            S_A,
            &[
                "card",
                "handoff",
                &card,
                "--generation",
                generation,
                "--to",
                S_B,
                "--project",
                &project,
            ],
        )
    };
    // The old generation is stale although the holder is the same; nothing moves.
    let (kind, message) = refusal(&handoff("1"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("stale"), "{message}");
    // A generation that was never issued is refused too.
    let (kind, message) = refusal(&handoff("9"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("never issued"), "{message}");
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    let cur = ledger::reservation::current(l.conn(), 1).unwrap().unwrap();
    assert_eq!((cur.holder.as_str(), cur.generation), (S_A, 2));
    drop(l);
    // The current generation works and the receiver gets the next one.
    let ok = result(&handoff("2"));
    assert_eq!(ok["reservation"]["generation"], json!(3));
    assert_eq!(ok["reservation"]["holder"], json!(S_B));
}

#[test]
fn the_holder_comes_from_the_two_host_variables_and_from_nothing_else() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/env.rs");
    let with_env = |pairs: &[(&str, &str)]| -> Output {
        let mut cmd = home.command(&["card", "take", &card, "--project", &project]);
        for (k, v) in pairs {
            cmd.env(k, v);
        }
        cmd.output().expect("run moai")
    };
    // Other variables a caller might set are not read: no session id, no reservation.
    for name in [
        "MOAI_CARD_HOLDER",
        "MOAI_SESSION_ID",
        "SESSION_ID",
        "USER",
        "LOGNAME",
    ] {
        let (kind, message) = refusal(&with_env(&[(name, S_B)]));
        assert_eq!(kind, "invalid_input", "{name}: {message}");
    }
    assert_eq!(count(&home, "SELECT count(*) FROM reservation"), 0);
    // The Claude variable wins over the Codex one; the Codex one works alone.
    let t = result(&with_env(&[("CODEX_THREAD_ID", S_B), (SESSION_VAR, S_A)]));
    assert_eq!(t["reservation"]["holder"], json!(S_A));
    result(&release(&home, &project, &card, S_A, "1"));
    let t = result(&with_env(&[
        ("CODEX_THREAD_ID", S_B),
        ("MOAI_CARD_HOLDER", S_C),
    ]));
    assert_eq!(t["reservation"]["holder"], json!(S_B));
}

#[test]
fn a_card_with_a_live_reservation_cannot_be_made_ready() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/live.rs");
    result(&take(&home, &project, &card, S_A));
    // A `hold` card that still has a reservation (made by hand: no command holds a card yet).
    let conn =
        rusqlite::Connection::open(ledger::ledger_path(home.path(), "demo").unwrap()).unwrap();
    conn.execute("UPDATE card SET state = 'hold' WHERE card_id = 1", [])
        .unwrap();
    drop(conn);
    let (kind, message) = refusal(&run(
        &home,
        &["card", "ready", &card, "--project", &project],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("live reservation"), "{message}");
    let shown = result(&run(&home, &["card", "show", &card, "--project", &project]));
    assert_eq!(shown["state"], json!("hold"), "the card is untouched");
    // Once the holder gives it back, the hold card becomes ready.
    result(&release(&home, &project, &card, S_A, "1"));
    result(&run(
        &home,
        &["card", "ready", &card, "--project", &project],
    ));
}

#[test]
fn ready_on_a_taken_card_names_its_state_and_changes_nothing() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/taken.rs");
    result(&take(&home, &project, &card, S_A));
    let shown = result(&run(&home, &["card", "show", &card, "--project", &project]));
    assert_eq!(shown["state"], json!("active"));
    // The state is the reason, not the reservation that goes with it.
    let (kind, message) = refusal(&run(
        &home,
        &["card", "ready", &card, "--project", &project],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("is `active`"), "{message}");
    assert!(!message.contains("live reservation"), "{message}");
    // Nothing was written: one `card.ready` event (from `ready_card`), one held reservation.
    assert_eq!(
        count(
            &home,
            "SELECT count(*) FROM event WHERE kind = 'card.ready'"
        ),
        1
    );
    assert_eq!(
        count(
            &home,
            "SELECT count(*) FROM reservation WHERE status = 'held'"
        ),
        1
    );
}

// ---- release and review (card C-68 N-a) -----------------------------------------------------------

fn generation_args<'a>(action: &'a str, card: &'a str, project: &'a str) -> [&'a str; 7] {
    [
        "card",
        action,
        card,
        "--generation",
        "1",
        "--project",
        project,
    ]
}

#[test]
fn a_release_of_a_review_card_by_another_session_is_a_fence_refusal() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/rev.rs");
    result(&take(&home, &project, &card, S_A));
    result(&as_session(
        &home,
        S_A,
        &generation_args("review", &card, &project),
    ));
    // Not the holder: the fence answers (and records it), not the sentence for the holder.
    let (kind, message) = refusal(&release(&home, &project, &card, S_B, "1"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("is held by"), "{message}");
    assert!(!message.contains("card reopen"), "{message}");
    assert_eq!(
        count(
            &home,
            "SELECT count(*) FROM event WHERE kind = 'fence_reject' \
             AND json_extract(payload, '$.op') = 'release' \
             AND json_extract(payload, '$.reason') = 'wrong_holder'"
        ),
        1
    );
    // The holder with a generation that was never issued: the fence again.
    let (kind, message) = refusal(&release(&home, &project, &card, S_A, "7"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("never issued"), "{message}");
    // The holder itself is told what to do, and nothing is written for it.
    let events = count(&home, "SELECT count(*) FROM event");
    let (kind, message) = refusal(&release(&home, &project, &card, S_A, "1"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(
        message.contains(&format!("card reopen {card} --generation 1")),
        "{message}"
    );
    assert_eq!(count(&home, "SELECT count(*) FROM event"), events);
    assert_eq!(
        count(
            &home,
            "SELECT count(*) FROM reservation WHERE status = 'held'"
        ),
        1
    );
}

/// The holder starts `card release` and `card review` at the same moment, 100 times (two
/// processes per round, the one started first alternates). The state check of `release` runs
/// inside its transaction, so exactly one of the two succeeds and no card is left in `review`
/// with a reservation that was given back.
#[test]
fn a_hundred_release_and_review_races_never_strand_a_review_card() {
    const ROUNDS: usize = 100;
    let (home, project) = registered();
    let mut stranded = Vec::new();
    for round in 0..ROUNDS {
        let card = ready_card(&home, &project, &format!("race/{round}/"));
        result(&take(&home, &project, &card, S_A));
        let start = |action: &str| {
            home.command(&generation_args(action, &card, &project))
                .env(SESSION_VAR, S_A)
                .stdout(Stdio::piped())
                .stderr(Stdio::piped())
                .spawn()
                .expect("start moai")
        };
        let order = if round % 2 == 0 {
            ["release", "review"]
        } else {
            ["review", "release"]
        };
        let first = start(order[0]);
        let second = start(order[1]);
        let outs = [
            first.wait_with_output().unwrap(),
            second.wait_with_output().unwrap(),
        ];
        let ok: Vec<&str> = order
            .iter()
            .zip(&outs)
            .filter(|(_, out)| out.status.code() == Some(0))
            .map(|(action, _)| *action)
            .collect();
        let l = ledger::open_read_only(home.path(), "demo").unwrap();
        let id = (round + 1) as i64;
        let state: String = l
            .conn()
            .query_row("SELECT state FROM card WHERE card_id = ?1", [id], |r| {
                r.get(0)
            })
            .unwrap();
        let held = ledger::reservation::current(l.conn(), id)
            .unwrap()
            .is_some_and(|r| r.status == "held");
        if state == "review" && !held {
            // Counted, not asserted here: the last line reports how many rounds went wrong.
            stranded.push(card.clone());
            continue;
        }
        match ok.as_slice() {
            ["release"] => assert_eq!((state.as_str(), held), ("ready", false), "{card}"),
            ["review"] => assert_eq!((state.as_str(), held), ("review", true), "{card}"),
            other => panic!("{card}: exactly one of the two succeeds, got {other:?}"),
        }
    }
    assert!(
        stranded.is_empty(),
        "{} of {ROUNDS} rounds left a `review` card without a live reservation: {stranded:?}",
        stranded.len()
    );
}

#[test]
fn a_bad_card_name_is_not_repeated_raw_on_stderr_or_stdout() {
    let (home, project) = registered();
    let nasty = "\u{1b}[31mC-\u{1b}[0m".to_string() + &"z".repeat(100);
    for action in ["show", "ready"] {
        let out = run(&home, &["card", action, &nasty, "--project", &project]);
        assert_eq!(out.status.code(), Some(1));
        for bytes in [&out.stdout, &out.stderr] {
            assert!(!bytes.contains(&0x1b), "{action}: raw escape in output");
            assert!(
                !String::from_utf8_lossy(bytes).contains(&"z".repeat(41)),
                "{action}: the name is repeated whole"
            );
        }
        assert_eq!(refusal(&out).0, "invalid_input");
    }
}

// ---- scope (card C-53: a card added without --scope is skipped by `next`) ------------------------

fn sql(home: &TempHome, statement: &str) {
    let mut w = ledger::open_write(home.path(), "demo").unwrap();
    w.with_write_tx(|tx, _| {
        tx.execute_batch(statement)?;
        Ok(())
    })
    .unwrap();
}

fn set_scope(home: &TempHome, project: &str, card: &str, scope: &[&str]) -> Output {
    let mut args = vec!["card", "scope", card];
    for s in scope {
        args.extend_from_slice(&["--scope", s]);
    }
    args.extend_from_slice(&["--by", "leader", "--project", project]);
    run(home, &args)
}

fn write_scope(home: &TempHome, card_id: i64) -> Value {
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    let text: String = l
        .conn()
        .query_row(
            "SELECT write_scope FROM card WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )
        .unwrap();
    serde_json::from_str(&text).unwrap()
}

/// Adds a card without a scope; returns its name.
fn bare_card(home: &TempHome, project: &str) -> String {
    let added = result(&run(
        home,
        &["card", "add", "--title", "no scope", "--project", project],
    ));
    added["name"].as_str().unwrap().to_string()
}

#[test]
fn scope_lets_next_offer_a_card_that_was_added_without_one() {
    let (home, project) = registered();
    let card = bare_card(&home, &project);
    result(&run(
        &home,
        &["card", "ready", &card, "--project", &project],
    ));
    // `next` takes the project folder as its last word, not as --project.
    let before = result(&as_session(&home, S_A, &["card", "next", &project]));
    assert_eq!(before["next"], Value::Null, "{before}");
    assert_eq!(before["skipped"][0]["reason"], json!("empty_scope"));

    let set = result(&set_scope(
        &home,
        &project,
        &card,
        &["src/login/", "docs/a.md"],
    ));
    assert_eq!(set["action"], json!("scope"));
    assert_eq!(set["state"], json!("ready"));
    assert_eq!(set["from"], json!([]));
    assert_eq!(set["write_scope"], json!(["src/login/", "docs/a.md"]));
    assert_eq!(write_scope(&home, 1), json!(["src/login/", "docs/a.md"]));
    // One event with the old scope, the new one and who set it.
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    let (actor, payload): (String, String) = l
        .conn()
        .query_row(
            "SELECT actor, payload FROM event WHERE kind = 'card.scope'",
            [],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .unwrap();
    assert_eq!(actor, "leader");
    assert_eq!(
        serde_json::from_str::<Value>(&payload).unwrap(),
        json!({"from": [], "to": ["src/login/", "docs/a.md"]})
    );

    let after = result(&as_session(&home, S_A, &["card", "next", &project]));
    assert_eq!(after["next"], json!(card), "{after}");
    assert_eq!(after["skipped"], json!([]));
    // A second call replaces the scope, and the event keeps what it replaced.
    let again = result(&set_scope(&home, &project, &card, &["src/other.rs"]));
    assert_eq!(again["from"], json!(["src/login/", "docs/a.md"]));
    assert_eq!(write_scope(&home, 1), json!(["src/other.rs"]));
    // `take` without --scope reserves the card's new scope.
    let taken = result(&take(&home, &project, &card, S_A));
    assert_eq!(taken["reservation"]["write_scope"], json!(["src/other.rs"]));
}

#[test]
fn scope_is_set_only_on_a_draft_or_ready_card_that_nobody_holds() {
    let (home, project) = registered();
    // draft (C-1): allowed.
    let draft = bare_card(&home, &project);
    assert_eq!(
        result(&set_scope(&home, &project, &draft, &["src/d.rs"]))["state"],
        json!("draft")
    );
    // active (C-2, taken): refused with the state and the holder.
    let active = ready_card(&home, &project, "src/a.rs");
    result(&take(&home, &project, &active, S_A));
    // review (C-3): refused the same way.
    let review = ready_card(&home, &project, "src/r.rs");
    result(&take(&home, &project, &review, S_A));
    result(&as_session(
        &home,
        S_A,
        &generation_args("review", &review, &project),
    ));
    // hold (C-4): refused, nobody holds it.
    let hold = ready_card(&home, &project, "src/h.rs");
    sql(&home, "UPDATE card SET state = 'hold' WHERE card_id = 4");
    // done (C-5): refused.
    sql(
        &home,
        "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
         VALUES ('closed', 'done', 'user', 1, 1, 1)",
    );
    let done = "C-5".to_string();
    // ready (C-6), but a live reservation is left on it: refused for the holder.
    let held = ready_card(&home, &project, "src/x.rs");
    result(&take(&home, &project, &held, S_B));
    sql(&home, "UPDATE card SET state = 'ready' WHERE card_id = 6");

    let events = count(&home, "SELECT count(*) FROM event");
    for (card, state, holder) in [
        (&active, "active", Some(S_A)),
        (&review, "review", Some(S_A)),
        (&hold, "hold", None),
        (&done, "done", None),
        (&held, "ready", Some(S_B)),
    ] {
        let (kind, message) = refusal(&set_scope(&home, &project, card, &["src/new.rs"]));
        // A card with an active reservation is widened, not refused for being held (card C-94,
        // `tests/card_scope_widen_cli.rs`): without `--reason` it is a missing input, and it
        // still names the state and the holder. Every other card keeps its conflict.
        let widenable = state == "active" || state == "review";
        let want = if widenable {
            "invalid_input"
        } else {
            "conflict"
        };
        assert_eq!(kind, want, "{card}: {message}");
        if widenable {
            assert!(message.contains("needs --reason"), "{card}: {message}");
        }
        assert!(
            message.contains(&format!("{card} is `{state}`")),
            "{card}: {message}"
        );
        match holder {
            Some(h) => assert!(
                message.contains(&format!("held by {h} at generation 1")),
                "{card}: {message}"
            ),
            None => assert!(!message.contains("held by"), "{card}: {message}"),
        }
        // `take --scope` is no way to change a card's scope (audit N3): never suggested.
        assert!(!message.contains("take --scope"), "{card}: {message}");
    }
    assert_eq!(count(&home, "SELECT count(*) FROM event"), events);
    assert_eq!(
        count(
            &home,
            "SELECT count(*) FROM card WHERE write_scope LIKE '%src/new.rs%'"
        ),
        0
    );
    let (kind, _) = refusal(&set_scope(&home, &project, "C-99", &["src/a.rs"]));
    assert_eq!(kind, "not_found");
}

#[test]
fn scope_entries_are_checked_like_add_and_within_the_reservation_limits() {
    let (home, project) = registered();
    let card = bare_card(&home, &project);
    for (scope, says) in [
        (vec!["/abs/dir"], "not a relative path"),
        (vec!["."], "names no place"),
        (vec!["src/../etc"], "leaves the repository"),
        (vec![""], "is empty"),
        (vec!["src\\a"], "backslash"),
        (vec!["src/a.rs", "src/a.rs"], "repeats entry 1"),
    ] {
        let (kind, message) = refusal(&set_scope(&home, &project, &card, &scope));
        assert_eq!(kind, "invalid_input", "{scope:?}: {message}");
        assert!(message.contains(says), "{scope:?}: {message}");
    }
    // At most 64 entries (64 passes, 65 is refused), each at most 512 characters.
    let many: Vec<String> = (0..65).map(|i| format!("src/f{i}.rs")).collect();
    let many: Vec<&str> = many.iter().map(String::as_str).collect();
    let (kind, message) = refusal(&set_scope(&home, &project, &card, &many));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("65 entries, at most 64"), "{message}");
    result(&set_scope(&home, &project, &card, &many[..64]));
    let long = format!("src/{}", "a".repeat(508));
    assert_eq!(long.chars().count(), 512);
    result(&set_scope(&home, &project, &card, &[&long]));
    let longer = format!("{long}b");
    let (kind, message) = refusal(&set_scope(&home, &project, &card, &[&longer]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("longer than 512"), "{message}");
    // No --scope, or no --by, is a usage error before the ledger is opened.
    for args in [
        vec![
            "card",
            "scope",
            card.as_str(),
            "--by",
            "leader",
            "--project",
            &project,
        ],
        vec![
            "card",
            "scope",
            card.as_str(),
            "--scope",
            "src/a.rs",
            "--project",
            &project,
        ],
    ] {
        let out = run(&home, &args);
        assert_eq!(out.status.code(), Some(2), "{args:?}");
    }
    assert_eq!(
        count(
            &home,
            "SELECT count(*) FROM event WHERE kind = 'card.scope'"
        ),
        2
    );
}

/// Card C-68 N-e: `take --scope` is checked as `add --scope` is: no absolute path, no `.`.
#[test]
fn take_refuses_a_scope_that_add_would_refuse() {
    let (home, project) = registered();
    let card = ready_card(&home, &project, "src/t.rs");
    for (scope, says) in [
        ("/abs/dir", "not a relative path"),
        (".", "names no place"),
        ("../up", "leaves the repository"),
    ] {
        let (kind, message) = refusal(&as_session(
            &home,
            S_A,
            &[
                "card",
                "take",
                &card,
                "--scope",
                scope,
                "--project",
                &project,
            ],
        ));
        assert_eq!(kind, "invalid_input", "{scope}: {message}");
        assert!(message.starts_with("card take refused: "), "{message}");
        assert!(message.contains(says), "{scope}: {message}");
    }
    assert_eq!(count(&home, "SELECT count(*) FROM reservation"), 0);
    let taken = result(&as_session(
        &home,
        S_A,
        &[
            "card",
            "take",
            &card,
            "--scope",
            "src/t/",
            "--project",
            &project,
        ],
    ));
    assert_eq!(taken["reservation"]["write_scope"], json!(["src/t/"]));
}

/// Card C-53 audit N4: `add --scope` has the reservation limits of `scope` and `take`: 64
/// entries pass and 65 are refused, 512 characters pass and 513 are refused; nothing is written.
#[test]
fn add_refuses_a_scope_the_reservation_could_not_hold() {
    let (home, project) = registered();
    let add = |scope: &[String]| {
        let mut args = vec!["card", "add", "--title", "t"];
        for s in scope {
            args.extend_from_slice(&["--scope", s.as_str()]);
        }
        args.extend_from_slice(&["--project", &project]);
        run(&home, &args)
    };
    let many: Vec<String> = (0..65).map(|i| format!("src/f{i}.rs")).collect();
    let (kind, message) = refusal(&add(&many));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(
        message.starts_with("card add refused: the write scope has 65 entries, at most 64"),
        "{message}"
    );
    let long = format!("src/{}", "a".repeat(508));
    let (kind, message) = refusal(&add(&[format!("{long}b")]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("longer than 512"), "{message}");
    assert_eq!(count(&home, "SELECT count(*) FROM card"), 0);
    result(&add(&many[..64]));
    result(&add(&[long]));
    assert_eq!(count(&home, "SELECT count(*) FROM card"), 2);
}

/// The patterns of the card's current reservation, sorted (as `reserve` stores them).
fn reserved_scope(home: &TempHome, card_id: i64) -> Vec<String> {
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    let mut stmt = l
        .conn()
        .prepare(
            "SELECT s.pattern FROM reservation r \
             JOIN reservation_scope s ON s.reservation_id = r.reservation_id \
             WHERE r.card_id = ?1 AND r.status = 'held' ORDER BY s.pattern",
        )
        .unwrap();
    let rows = stmt
        .query_map([card_id], |r| r.get::<_, String>(0))
        .unwrap()
        .collect::<Result<Vec<_>, _>>()
        .unwrap();
    rows
}

/// Audit B1 of card C-53: `card take` without `--scope` and `card scope` start at the same
/// moment, 100 times (two processes per round, the one started first alternates). Whenever
/// both succeed, `scope` committed first, so the reservation holds the card's new scope; a
/// reservation that keeps the old scope while the card row holds the new one is counted.
#[test]
fn a_hundred_take_and_scope_races_never_reserve_a_scope_the_card_no_longer_has() {
    const ROUNDS: usize = 100;
    let (home, project) = registered();
    let mut split = Vec::new();
    let (mut both, mut take_only) = (0, 0);
    for round in 0..ROUNDS {
        let card = ready_card(&home, &project, &format!("old/{round}/"));
        let new = format!("new/{round}/");
        let start = |action: &str| {
            let args: Vec<&str> = if action == "take" {
                vec!["card", "take", &card, "--project", &project]
            } else {
                vec![
                    "card",
                    "scope",
                    &card,
                    "--scope",
                    &new,
                    "--by",
                    "leader",
                    "--project",
                    &project,
                ]
            };
            home.command(&args)
                .env(SESSION_VAR, S_A)
                .stdout(Stdio::piped())
                .stderr(Stdio::piped())
                .spawn()
                .expect("start moai")
        };
        let order = if round % 2 == 0 {
            ["take", "scope"]
        } else {
            ["scope", "take"]
        };
        let first = start(order[0]);
        let second = start(order[1]);
        let outs = [
            first.wait_with_output().unwrap(),
            second.wait_with_output().unwrap(),
        ];
        let ok = |action: &str| {
            order
                .iter()
                .zip(&outs)
                .any(|(a, out)| *a == action && out.status.code() == Some(0))
        };
        let id = (round + 1) as i64;
        let card_scope: Vec<String> = serde_json::from_value(write_scope(&home, id)).unwrap();
        let reserved = reserved_scope(&home, id);
        match (ok("take"), ok("scope")) {
            (true, true) => {
                both += 1;
                if reserved != card_scope || card_scope != [new.clone()] {
                    split.push(format!(
                        "{card}: card {card_scope:?}, reserved {reserved:?}"
                    ));
                }
            }
            (true, false) => {
                take_only += 1;
                assert_eq!(card_scope, [format!("old/{round}/")], "{card}");
                assert_eq!(reserved, card_scope, "{card}");
            }
            other => panic!("{card}: `take` always succeeds here, got (take, scope) = {other:?}"),
        }
    }
    assert!(
        split.is_empty(),
        "{} of {both} rounds where both succeeded reserved another scope than the card's: {split:?}",
        split.len()
    );
    assert_eq!(both + take_only, ROUNDS);
    eprintln!("take and scope both succeeded in {both} rounds, take alone in {take_only}");
}
