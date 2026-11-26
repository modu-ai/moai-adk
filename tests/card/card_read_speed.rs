//! `moai card next` and `moai card list` on a ledger of the size a real project reaches
//! (SPEC-20261126-card-next-speed): about 500 cards, 15,000 events and 9,000 notes.
//!
//! - The two commands answer within a generous time cap (the best of three runs, so a loaded
//!   runner does not fail the test), and `card next` starts a number of git processes that does
//!   not grow with the number of cards.
//! - The answer is the same with and without the `event_card_kind` index (an earlier ledger file).
//! - The batched head check gives the verdict of the per-head check.
//!
//! Every command runs against a temporary HOME and a temporary git repository; the real home
//! state is not read or written.
#![cfg(unix)]

use crate::common;

use common::home::TempHome;
use moai_cli::card::landed::{contained, contained_many, Containment};
use moai_cli::ledger;
use rusqlite::{params, Connection};
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Stdio;
use std::time::{Duration, Instant};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const SESSION: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const CARDS: i64 = 500;
const EVENTS: i64 = 15_000;
const NOTES: i64 = 9_000;
/// The time cap of one command, in seconds, on a ledger of this size: the best of three runs of
/// the unoptimized test build (its SQLite is built without optimization) on a loaded runner,
/// where a git process alone takes 20 to 60 ms. A release build on the same ledger needs a
/// fraction of it; the old shape (a table scan per card, a git process per card) needed several
/// times more. The number of git processes below is the load-independent guard.
const CAP_SECONDS: f64 = 1.0;
/// The most git processes `card next` may start on this ledger. The count is the fixed cost of
/// the identity, the tip and the batched head check; one process per card would be about 400.
const GIT_PROCESS_CAP: usize = 24;

const MFLOW_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";

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

/// An absolute `git` on this runner's `PATH`, for the child that has none.
fn git_executable() -> PathBuf {
    let path = std::env::var_os("PATH").expect("PATH");
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
        .expect("a git program on PATH")
}

/// A program that appends one line to `$GIT_COUNT_LOG` and runs the real git, so a test can count
/// how many git processes a command starts.
fn counting_git(dir: &Path) -> PathBuf {
    let script = dir.join("counting-git");
    fs::write(
        &script,
        "#!/bin/sh\necho x >> \"$GIT_COUNT_LOG\"\nexec \"$GIT_COUNT_REAL\" \"$@\"\n",
    )
    .unwrap();
    fs::set_permissions(&script, fs::Permissions::from_mode(0o755)).unwrap();
    script
}

struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    git_bin: PathBuf,
    counter: PathBuf,
    log: PathBuf,
    /// Commits of the integration branch (inside), of a side branch (outside), and one that
    /// names nothing.
    inside: Vec<String>,
    outside: String,
    missing: String,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        let mut inside = Vec::new();
        for n in 0..3 {
            fs::write(project.join(format!("f{n}.txt")), format!("{n}\n")).unwrap();
            git(&project, &["add", &format!("f{n}.txt")]);
            git(&project, &["commit", "-q", "-m", &format!("commit {n}")]);
            inside.push(git(&project, &["rev-parse", "HEAD"]));
        }
        git(&project, &["branch", "integ"]);
        git(&project, &["checkout", "-q", "-b", "side"]);
        fs::write(project.join("side.txt"), "side\n").unwrap();
        git(&project, &["add", "side.txt"]);
        git(&project, &["commit", "-q", "-m", "not integrated"]);
        let outside = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["checkout", "-q", "main"]);
        let cfg = moai_cli::state_root::path(&project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, MFLOW_PROFILE).unwrap();
        let counter = counting_git(home.path());
        let log = home.path().join("git-count.log");
        World {
            p: project.to_str().unwrap().to_string(),
            git_bin: git_executable(),
            home,
            project,
            counter,
            log,
            inside,
            outside,
            missing: "f".repeat(40),
        }
    }

    fn ledger_path(&self) -> PathBuf {
        ledger::ledger_path(self.home.path(), "demo").unwrap()
    }

    /// 500 cards (a `ready` card with a scope and an open criterion each), 15,000 events
    /// (every card has one `card.landed` head: inside the tip, outside it or no commit), and
    /// 9,000 notes. Cards 1..=CARDS get the ids 1..=CARDS.
    fn seed(&self) {
        let mut conn = Connection::open(self.ledger_path()).unwrap();
        let tx = conn.transaction().unwrap();
        let heads = [
            self.inside[0].as_str(),
            self.inside[1].as_str(),
            self.inside[2].as_str(),
            self.outside.as_str(),
            self.missing.as_str(),
        ];
        for id in 1..=CARDS {
            tx.execute(
                "INSERT INTO card (title, state, write_scope, created_by, created_at, updated_at) \
                 VALUES (?1, 'ready', ?2, 'fixture', 1, 1)",
                params![format!("card {id}"), format!("[\"src/area{id}/**\"]")],
            )
            .unwrap();
            tx.execute(
                "INSERT INTO card_ac (card_id, seq, text, updated_at) VALUES (?1, 1, 'one', 1)",
                params![id],
            )
            .unwrap();
            // One landed head per card; a fifth of the heads are real commits of the tip.
            let head = heads[(id as usize) % heads.len()];
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, 2, 'fixture', ?2, 'card.landed', ?3)",
                params![
                    format!("landed-{id}"),
                    id,
                    json!({ "head": head }).to_string()
                ],
            )
            .unwrap();
        }
        for n in 0..(EVENTS - CARDS) {
            let card = n % CARDS + 1;
            let kind = if n % 3 == 0 { "card.spec" } else { "card.note" };
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, 3, 'fixture', ?2, ?3, ?4)",
                params![
                    format!("e-{n}"),
                    card,
                    kind,
                    json!({ "to": format!("SPEC-20261126-fixture-{}", n % 40) }).to_string()
                ],
            )
            .unwrap();
        }
        for n in 0..NOTES {
            tx.execute(
                "INSERT INTO card_note (card_id, kind, author, body, created_at) \
                 VALUES (?1, 'comment', 'fixture', ?2, 4)",
                params![n % CARDS + 1, format!("note {n}")],
            )
            .unwrap();
        }
        tx.commit().unwrap();
    }

    fn drop_event_index(&self) {
        Connection::open(self.ledger_path())
            .unwrap()
            .execute_batch("DROP INDEX IF EXISTS event_card_kind;")
            .unwrap();
    }

    fn has_event_index(&self) -> bool {
        Connection::open(self.ledger_path())
            .unwrap()
            .query_row(
                "SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE name = 'event_card_kind')",
                [],
                |r| r.get(0),
            )
            .unwrap()
    }

    /// One `moai card <args> <project>` run: the answer and how long it took.
    fn run(&self, args: &[&str]) -> (Value, Duration) {
        self.run_with(args, true)
    }

    /// [`World::run`]; with `counted` false git runs directly, without the counting wrapper
    /// (one more process per git command), so a timing is not slowed by the count.
    fn run_with(&self, args: &[&str], counted: bool) -> (Value, Duration) {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.push(&self.p);
        let mut cmd = self.home.command(&all);
        if counted {
            cmd.env("MOAI_GIT_BIN", &self.counter)
                .env("GIT_COUNT_REAL", &self.git_bin)
                .env("GIT_COUNT_LOG", &self.log);
        } else {
            cmd.env("MOAI_GIT_BIN", &self.git_bin);
        }
        cmd.env(SESSION_VAR, SESSION).stdin(Stdio::null());
        let _ = fs::remove_file(&self.log);
        let start = Instant::now();
        let out = cmd.output().expect("run moai");
        let took = start.elapsed();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
            panic!(
                "stdout is not JSON ({e}): {} {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        assert_eq!(out.status.code(), Some(0), "{v}");
        (v["result"].clone(), took)
    }

    /// The git processes the last [`World::run`] started.
    fn git_processes(&self) -> usize {
        fs::read_to_string(&self.log)
            .map(|s| s.lines().count())
            .unwrap_or(0)
    }

    /// The best time of three runs, and the answer of the last.
    fn best_of_three(&self, args: &[&str]) -> (Value, f64, usize) {
        let (answer, _) = self.run(args);
        let processes = self.git_processes();
        let best = (0..3)
            .map(|_| self.run_with(args, false).1.as_secs_f64())
            .fold(f64::MAX, f64::min);
        (answer, best, processes)
    }
}

/// `value` without the fields that change from run to run.
fn stable(value: &Value) -> Value {
    match value {
        Value::Object(map) => Value::Object(
            map.iter()
                .filter(|(k, _)| !matches!(k.as_str(), "now" | "lease_until" | "at"))
                .map(|(k, v)| (k.clone(), stable(v)))
                .collect(),
        ),
        Value::Array(items) => Value::Array(items.iter().map(stable).collect()),
        other => other.clone(),
    }
}

#[test]
fn card_next_and_card_list_answer_a_real_sized_ledger_within_the_cap() {
    let w = World::new();
    w.seed();
    assert!(w.has_event_index(), "a new ledger has the index");
    let (next, next_s, git_processes) = w.best_of_three(&["next"]);
    let (list, list_s, _) = w.best_of_three(&["list"]);
    assert!(
        next_s <= CAP_SECONDS,
        "card next took {next_s:.3} s, cap {CAP_SECONDS} s"
    );
    assert!(
        list_s <= CAP_SECONDS,
        "card list took {list_s:.3} s, cap {CAP_SECONDS} s"
    );
    assert!(
        git_processes <= GIT_PROCESS_CAP,
        "card next started {git_processes} git processes for {CARDS} cards, cap {GIT_PROCESS_CAP}"
    );
    // The answers carry the cards: the proofs of landing took the inside heads out of `ready`.
    assert!(
        next["landing"]["checked"] == json!(true),
        "{}",
        next["landing"]
    );
    let listed = list["cards"].as_array().map(Vec::len).unwrap_or(0);
    assert!(listed > 0, "card list answered no card: {list}");
}

#[test]
fn the_answers_do_not_depend_on_the_event_index() {
    let w = World::new();
    w.seed();
    let (next_with, _) = w.run(&["next"]);
    let (list_with, _) = w.run(&["list"]);
    w.drop_event_index();
    assert!(!w.has_event_index());
    let (next_without, _) = w.run(&["next"]);
    let (list_without, _) = w.run(&["list"]);
    assert_eq!(stable(&next_with), stable(&next_without), "card next");
    assert_eq!(stable(&list_with), stable(&list_without), "card list");
    assert!(
        !w.has_event_index(),
        "a read-only command never makes the index"
    );
}

#[test]
fn the_batched_head_check_gives_the_verdict_of_the_per_head_check() {
    let w = World::new();
    let tip = git(&w.project, &["rev-parse", "integ"]);
    let heads: Vec<&str> = vec![
        w.inside[0].as_str(),
        w.outside.as_str(),
        w.missing.as_str(),
        w.inside[2].as_str(),
        w.outside.as_str(), // a repeated head
        "not-a-sha",
        w.inside[1].as_str(),
    ];
    let many = contained_many(&w.project, &heads, &tip).unwrap();
    for head in &heads {
        let one = contained(&w.project, head, &tip).unwrap();
        assert_eq!(many[*head], one, "head {head}");
    }
    assert_eq!(many[&w.inside[0]], Containment::Inside);
    assert_eq!(many[&w.outside], Containment::Outside);
    assert_eq!(many[&w.missing], Containment::NotACommit);
    assert_eq!(many["not-a-sha"], Containment::NotACommit);
    assert_eq!(
        many.len(),
        heads
            .iter()
            .collect::<std::collections::BTreeSet<_>>()
            .len(),
        "one entry per distinct head"
    );
    // No head, or only heads that name nothing, starts no failing git command.
    assert!(contained_many(&w.project, &[], &tip).unwrap().is_empty());
    let none = contained_many(&w.project, &[w.missing.as_str()], &tip).unwrap();
    assert_eq!(none[&w.missing], Containment::NotACommit);
    // A tip that is no commit is an error, as for the per-head check.
    assert!(contained_many(&w.project, &[w.inside[0].as_str()], &w.missing).is_err());
    assert!(contained(&w.project, &w.inside[0], &w.missing).is_err());
}

/// `card next` builds the leader name and the lane prefix from one project name; the session
/// module's own functions must keep the same two forms.
#[test]
fn the_leader_name_and_lane_prefix_keep_their_forms_of_the_project_name() {
    let w = World::new();
    let name = moai_cli::session::project_name(&w.project);
    assert_eq!(
        moai_cli::session::leader_session_name(&w.project),
        format!("{name}-leader")
    );
    assert_eq!(
        moai_cli::session::agent_prefix(&w.project),
        format!("{name}-agent-")
    );
}
