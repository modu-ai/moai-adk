//! `moai leader-round` and the lease sweep (SPEC-20261124-leader-round-lease-sweep): each round
//! calls `moai card sweep` once before it builds its answer, names the cards it returned, and
//! only lists them when the profile key `lease_sweep_in_round` is false.
//!
//! Every command runs against a temporary HOME; the real home ledger is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const TWENTY_FIVE_HOURS: i64 = 25 * 3600;
const TWENTY_THREE_HOURS: i64 = 23 * 3600;
const NOW: &str = "2026-11-24T12:00:00Z";

struct World {
    home: TempHome,
    project: String,
}

/// A temporary HOME holding a registered ledger `demo` and a project with a complete profile that
/// carries `extra` (a JSON fragment such as `, "lease_sweep_in_round": false`).
fn world(extra: &str) -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    let w = World {
        home,
        project: project.to_str().unwrap().to_string(),
    };
    w.write_profile(extra);
    w
}

impl World {
    /// Writes a complete profile with `extra` spliced in.
    fn write_profile(&self, extra: &str) {
        let project = std::path::Path::new(&self.project);
        let cfg = moai_cli::state_root::path(project, "project.yaml");
        std::fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        std::fs::write(
            &cfg,
            format!(
                "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]{extra}}}"
            ),
        )
        .unwrap();
    }

    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    /// A card taken by `session` at generation 1.
    fn held(&self, session: &str, title: &str, scope: &[&str]) -> String {
        let mut args = vec!["add", "--title", title, "--ac", "one"];
        for entry in scope {
            args.push("--scope");
            args.push(entry);
        }
        let name = card_result(&self.card(None, &args))["name"]
            .as_str()
            .unwrap()
            .to_string();
        card_result(&self.card(None, &["ready", &name]));
        card_result(&self.card(
            Some(session),
            &[
                "take",
                &name,
                "--worktree",
                "/work/tree",
                "--branch",
                "card/x",
            ],
        ));
        name
    }

    /// The lease of the held reservation of `card` ended `secs_ago` seconds ago.
    fn end_lease(&self, card: &str, secs_ago: i64) {
        let id: i64 = card.trim_start_matches("C-").parse().unwrap();
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(&format!(
                "UPDATE reservation SET lease_until = unixepoch() - {secs_ago} \
                 WHERE card_id = {id} AND status = 'held'"
            ))?;
            Ok(())
        })
        .unwrap();
    }

    fn state_of(&self, card: &str) -> String {
        card_result(&self.card(None, &["show", card]))["state"]
            .as_str()
            .unwrap()
            .to_string()
    }

    /// How many rows of `table` the temporary ledger holds.
    fn count(&self, table: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let conn: &Connection = l.conn();
        conn.query_row(&format!("SELECT count(*) FROM {table}"), [], |r| r.get(0))
            .unwrap()
    }

    /// `moai leader-round` over an observations file with no lane. `without_home` runs it with no
    /// HOME in the environment.
    fn round_output(&self, without_home: bool) -> Output {
        let obs = self.home.path().join("observations.json");
        std::fs::write(&obs, json!({"schema_version": 1, "lanes": []}).to_string()).unwrap();
        let mut cmd = self.home.command(&[
            "leader-round",
            &self.project,
            "--observations",
            obs.to_str().unwrap(),
            "--now",
            NOW,
        ]);
        if without_home {
            cmd.env_remove("HOME");
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    fn round(&self) -> Value {
        round_result(&self.round_output(false))
    }
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

fn card_result(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn round_result(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("leader.round"), "{v}");
    v["result"].clone()
}

fn names(part: &Value, key: &str) -> Vec<String> {
    part[key]
        .as_array()
        .unwrap_or_else(|| panic!("`{key}` is not a list: {part}"))
        .iter()
        .map(|c| c["name"].as_str().unwrap().to_string())
        .collect()
}

// ---- AC-1: the round calls the sweep once --------------------------------------------------------

#[test]
fn round_sweeps_an_ended_lease_and_names_the_card() {
    let w = world("");
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    let r = w.round();
    let part = &r["lease_sweep"];
    assert_eq!(part["status"], json!("swept"), "{r}");
    assert_eq!(names(part, "cards"), vec![a.clone()], "{r}");
    assert_eq!(part["cards"][0]["generation"], json!(1), "{r}");
    assert_eq!(part["cards"][0]["holder"], json!(S_A), "{r}");
    assert!(part["cards"][0]["lease_until"].is_i64(), "{r}");
    assert_eq!(r["writes"], json!(true), "{r}");
    assert_eq!(w.state_of(&a), "ready");
}

#[test]
fn round_sweeps_nothing_on_the_second_round() {
    let w = world("");
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    w.round();
    let events = w.count("event");
    let again = w.round();
    assert_eq!(again["lease_sweep"]["status"], json!("swept"), "{again}");
    assert_eq!(names(&again["lease_sweep"], "cards"), Vec::<String>::new());
    assert_eq!(again["writes"], json!(false), "{again}");
    assert_eq!(w.count("event"), events, "a second round wrote an event");
}

#[test]
fn round_sweeps_before_it_builds_the_rest_of_the_answer() {
    let w = world("");
    let r = w.round();
    let keys: Vec<&str> = r.as_object().unwrap().keys().map(String::as_str).collect();
    assert!(keys.contains(&"lease_sweep"), "{keys:?}");
    assert!(
        keys.contains(&"lanes") && keys.contains(&"round"),
        "{keys:?}"
    );
}

// ---- AC-2: the profile key ----------------------------------------------------------------------

#[test]
fn profile_off_lists_the_card_and_writes_nothing() {
    let w = world(", \"lease_sweep_in_round\": false");
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    let (events, notes) = (w.count("event"), w.count("card_note"));
    let r = w.round();
    let part = &r["lease_sweep"];
    assert_eq!(part["status"], json!("listed"), "{r}");
    assert_eq!(names(part, "cards"), vec![a.clone()], "{r}");
    assert_eq!(r["writes"], json!(false), "{r}");
    assert_eq!(w.state_of(&a), "active");
    assert_eq!(w.count("event"), events, "the round wrote an event");
    assert_eq!(w.count("card_note"), notes, "the round wrote a note");
}

#[test]
fn profile_on_and_absent_key_sweep_alike() {
    for extra in [", \"lease_sweep_in_round\": true", ""] {
        let w = world(extra);
        let a = w.held(S_A, "zombie", &["src/a.rs"]);
        w.end_lease(&a, TWENTY_FIVE_HOURS);
        let r = w.round();
        assert_eq!(r["lease_sweep"]["status"], json!("swept"), "{extra}: {r}");
        assert_eq!(w.state_of(&a), "ready", "{extra}");
    }
}

/// A profile that cannot be read cannot say "on": the round only lists and names where the answer
/// came from.
#[test]
fn profile_unreadable_lists_only() {
    // `lease_sweep_hours: 0` is refused when the profile is read.
    let w = world(", \"lease_sweep_hours\": 0");
    assert!(moai_cli::profile::read_profile(std::path::Path::new(&w.project)).is_err());
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    let r = w.round();
    let part = &r["lease_sweep"];
    assert_eq!(part["status"], json!("listed"), "{r}");
    assert_eq!(part["switch_source"], json!("unreadable_profile"), "{r}");
    assert_eq!(names(part, "cards"), vec![a.clone()], "{r}");
    assert_eq!(w.state_of(&a), "active");
}

#[test]
fn profile_key_must_be_a_bool() {
    let w = world(", \"lease_sweep_in_round\": \"no\"");
    assert!(
        moai_cli::profile::read_profile(std::path::Path::new(&w.project)).is_err(),
        "a string value must not read"
    );
    w.write_profile(", \"lease_sweep_in_round\": false");
    let (p, _) = moai_cli::profile::read_profile(std::path::Path::new(&w.project)).unwrap();
    assert_eq!(p.lease_sweep_in_round, Some(false));
}

// ---- AC-3: a missing piece never fails the round ---------------------------------------------------

#[test]
fn not_run_without_a_profile() {
    let home = TempHome::new();
    let project = home.path().join("work").join("bare");
    std::fs::create_dir_all(&project).unwrap();
    let obs = home.path().join("observations.json");
    std::fs::write(&obs, json!({"schema_version": 1, "lanes": []}).to_string()).unwrap();
    let out = home
        .command(&[
            "leader-round",
            project.to_str().unwrap(),
            "--observations",
            obs.to_str().unwrap(),
            "--now",
            NOW,
        ])
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let r = round_result(&out);
    assert_eq!(r["lease_sweep"]["status"], json!("not_run"), "{r}");
    assert!(r["lease_sweep"]["reason"].is_string(), "{r}");
    assert_eq!(r["writes"], json!(false), "{r}");
    assert!(r["lanes"].is_array() && r["round"].is_object(), "{r}");
}

#[test]
fn not_run_without_a_ledger() {
    let home = TempHome::new();
    let project = home.project("demo");
    let w = World {
        home,
        project: project.to_str().unwrap().to_string(),
    };
    w.write_profile("");
    let r = w.round();
    assert_eq!(r["lease_sweep"]["status"], json!("not_run"), "{r}");
    assert!(r["lease_sweep"]["reason"].is_string(), "{r}");
    assert!(r["lanes"].is_array() && r["round"].is_object(), "{r}");
    assert!(
        !w.home.path().join(".moai").exists(),
        "a round with no ledger must not create one"
    );
}

#[test]
fn not_run_without_home() {
    let w = world("");
    let a = w.held(S_A, "zombie", &["src/a.rs"]);
    w.end_lease(&a, TWENTY_FIVE_HOURS);
    let r = round_result(&w.round_output(true));
    assert_eq!(r["lease_sweep"]["status"], json!("not_run"), "{r}");
    assert!(
        r["lease_sweep"]["reason"]
            .as_str()
            .unwrap()
            .contains("HOME"),
        "{r}"
    );
    assert_eq!(w.state_of(&a), "active", "no HOME, no sweep");
    assert!(r["lanes"].is_array() && r["round"].is_object(), "{r}");
}

#[test]
fn error_when_the_ledger_file_is_not_a_ledger_and_the_round_goes_on() {
    let w = world("");
    let path = ledger::location::ledger_path(w.home.path(), "demo").unwrap();
    std::fs::write(&path, b"not a database").unwrap();
    let r = w.round();
    assert_eq!(r["lease_sweep"]["status"], json!("error"), "{r}");
    assert!(r["lease_sweep"]["reason"].is_string(), "{r}");
    assert_eq!(r["writes"], json!(false), "{r}");
    assert!(r["lanes"].is_array() && r["round"].is_object(), "{r}");
}

// ---- AC-4: the round calls the sweep and does not widen it ---------------------------------------

#[test]
fn skipped_cards_stay_as_they_are_and_are_named_with_a_reason() {
    let w = world("");
    let late = w.held(S_A, "late", &["src/a.rs"]);
    let inside = w.held(S_A, "inside the grace", &["src/b.rs"]);
    let review = w.held(S_A, "in review", &["src/c.rs"]);
    card_result(&w.card(
        Some(S_A),
        &["review", &review, "--generation", "1", "--keep", "src/c.rs"],
    ));
    w.end_lease(&late, TWENTY_FIVE_HOURS);
    w.end_lease(&inside, TWENTY_THREE_HOURS);
    w.end_lease(&review, TWENTY_FIVE_HOURS);
    let r = w.round();
    let part = &r["lease_sweep"];
    assert_eq!(names(part, "cards"), vec![late.clone()], "{r}");
    let skipped: Vec<(String, String)> = part["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| {
            (
                s["name"].as_str().unwrap().to_string(),
                s["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    assert!(skipped.iter().any(|(n, _)| *n == inside), "{skipped:?}");
    assert!(skipped.iter().any(|(n, _)| *n == review), "{skipped:?}");
    assert_eq!(w.state_of(&inside), "active");
    assert_eq!(w.state_of(&review), "review");
    assert_eq!(w.state_of(&late), "ready");
}
