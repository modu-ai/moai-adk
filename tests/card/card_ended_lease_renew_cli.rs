//! SPEC-20261127-waiting-card-lease: a holder whose lease ended while it waited keeps its card with
//! `card renew`, and the refusal says so. Every command runs the built binary against a temporary
//! HOME and a temporary repository; the real home state folder is not read or written. The clock
//! cannot be set, so a test that needs a lease to end moves it in the ledger directly (the way
//! `card_review_lease_hold_cli` and `card_reclaim_cli` do).
#![cfg(unix)]

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const DAY: i64 = 86_400;
/// Seconds a measured lease may differ from its nominal length: the commands run in between.
const SLACK: i64 = 30;

struct World {
    home: TempHome,
    p: String,
    /// The folder the commands run in: not a git worktree of the repository under test.
    host: tempfile::TempDir,
}

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
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
    v["result"].clone()
}

/// A refusal: exit 1, `ok` false; the kind and the message.
fn refusal(out: &Output) -> (String, String) {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        git(&project, &["add", "a.txt"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            host: tempfile::Builder::new()
                .prefix("moai-test-host-")
                .tempdir()
                .unwrap(),
        }
    }

    fn project(&self) -> PathBuf {
        PathBuf::from(&self.p)
    }

    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let mut cmd: Command = self.home.command(&all);
        cmd.current_dir(self.host.path());
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    /// A ready card with the scope entries given, taken by `S_A` at generation 1; returns its name.
    fn taken_card(&self, scope: &[&str]) -> String {
        let mut args = vec!["add", "--title", "waits for a landing", "--ac", "fixture"];
        for entry in scope {
            args.push("--scope");
            args.push(entry);
        }
        let name = result(&self.card(None, &args))["name"]
            .as_str()
            .unwrap()
            .to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(
            Some(S_A),
            &[
                "take",
                &name,
                "--worktree",
                "/work/tree",
                "--branch",
                "card/ended-lease-renew",
            ],
        ));
        name
    }

    /// Runs `statement` against the ledger directly.
    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
    }

    /// Moves the lease of the card's `held` reservation to `secs` from now (negative: ended).
    fn set_lease(&self, card: i64, secs: i64) {
        self.sql(&format!(
            "UPDATE reservation SET lease_until = unixepoch() + ({secs}) \
             WHERE card_id = {card} AND status = 'held'"
        ));
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }
}

/// REQ-4, REQ-5: the refusal of an ended own lease keeps its kind and its first words and names
/// `card renew` with the card and the generation; the renew then keeps generation 1, no reclaim is
/// counted and the same `card check` answers `allowed: true`.
#[test]
fn ac1_an_ended_own_lease_is_refused_with_the_renew_way_and_renew_keeps_the_generation() {
    let w = World::new();
    let card = w.taken_card(&["apps/bo/"]);
    w.set_lease(1, -5);
    let check = |w: &World| {
        w.card(
            Some(S_A),
            &[
                "check",
                &card,
                "--generation",
                "1",
                "--path",
                "apps/bo/x.rs",
            ],
        )
    };
    let (kind, message) = refusal(&check(&w));
    assert_eq!(kind, "conflict", "{message}");
    assert!(
        message.starts_with("reservation conflict in card: the lease ended at "),
        "the refusal lost its first words: {message}"
    );
    assert!(
        message.contains("moai card renew C-1 --generation 1"),
        "the refusal does not name the renew: {message}"
    );
    let renewed = result(&w.card(
        Some(S_A),
        &["renew", &card, "--generation", "1", "--lease-secs", "86400"],
    ));
    assert_eq!(renewed["reservation"]["generation"], json!(1), "{renewed}");
    assert_eq!(renewed["reservation"]["status"], json!("held"), "{renewed}");
    let left =
        renewed["reservation"]["lease_until"].as_i64().unwrap() - w.count("SELECT unixepoch()");
    assert!(left >= DAY - SLACK, "the renewed lease ends in {left} s");
    let allowed = result(&check(&w));
    assert_eq!(allowed["allowed"], json!(true), "{allowed}");
    assert_eq!(
        w.count("SELECT COUNT(*) FROM event WHERE kind = 'card.reclaim'"),
        0,
        "the renew counted a reclaim"
    );
    assert_eq!(
        w.count("SELECT COUNT(*) FROM reservation WHERE card_id = 1"),
        1,
        "the renew made a second reservation"
    );
}

/// REQ-1: the help of `card renew` says that an ended lease is renewed by its holder, and no longer
/// says that a live lease is needed.
#[test]
fn ac2_the_renew_help_says_an_ended_lease_is_renewed() {
    let w = World::new();
    let out = w
        .home
        .command(&["card", "renew", "--help"])
        .current_dir(w.host.path())
        .stdin(Stdio::null())
        .output()
        .expect("run moai");
    assert_eq!(out.status.code(), Some(0));
    let text = String::from_utf8_lossy(&out.stdout)
        .split_whitespace()
        .collect::<Vec<_>>()
        .join(" ");
    assert!(
        text.contains("whether or not the lease has ended"),
        "the help does not say an ended lease is renewed: {text}"
    );
    assert!(
        !text.contains("Needs a lease that has not ended"),
        "the help still says a live lease is needed: {text}"
    );
}
