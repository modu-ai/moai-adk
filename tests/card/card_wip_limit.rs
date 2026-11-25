//! The limit of cards in work and in review (SPEC-20261125-landing-moves-checkout REQ-1 to
//! REQ-4): `card take` and `card add` refuse once the `active` and `review` cards number at least
//! `wip_limit` (default 2) times the live registered lanes (at least 1), `--by user` adds pass,
//! `wip_limit` 0 turns the limit off, and `card next` at the limit names the cards that wait for
//! an audit.
//!
//! Every command runs against a temporary HOME and a temporary git repository; the real home
//! folder and the real session registry are never read or written.
#![cfg(unix)]

use crate::common;

use common::git::git_in;
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";

/// A well-formed host session id for number `n`.
fn sid(n: u64) -> String {
    format!("{n:08x}-0000-4000-8000-{n:012x}")
}

fn profile(wip_limit: Option<u32>) -> Vec<u8> {
    let limit = wip_limit.map_or(String::new(), |n| format!(", \"wip_limit\": {n}"));
    format!(
        "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {{\"flow\": \"m-flow\", \"integration_branch\": \"integ\"{limit}}}}}"
    )
    .into_bytes()
}

struct World {
    home: TempHome,
    project: PathBuf,
}

fn world(wip_limit: Option<u32>) -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    git_in(&project, &["init", "-q", "-b", "main"]);
    fs::write(project.join("a.txt"), "a\n").unwrap();
    git_in(&project, &["add", "-A"]);
    git_in(&project, &["commit", "-q", "-m", "first"]);
    let w = World { home, project };
    w.set_limit(wip_limit);
    w
}

impl World {
    fn set_limit(&self, wip_limit: Option<u32>) {
        let cfg = self.project.join(".moai-cli/project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, profile(wip_limit)).unwrap();
    }

    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.current_dir(&self.project);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", self.project.to_str().unwrap()]);
        self.run(session, &all)
    }

    /// `moai card next <project>` as `session`.
    fn next(&self, session: &str) -> Output {
        self.run(
            Some(session),
            &["card", "next", self.project.to_str().unwrap()],
        )
    }

    /// A lane entry of the session registry, as `session-register` writes it.
    fn lane(&self, name: &str, session: &str) {
        let out = self.run(
            Some(session),
            &[
                "session-register",
                self.project.to_str().unwrap(),
                "--name",
                name,
            ],
        );
        result(&out);
    }

    /// A leader entry of the session registry (`session-register --mode leader`).
    fn leader(&self, name: &str, session: &str) {
        let out = self.run(
            Some(session),
            &[
                "session-register",
                self.project.to_str().unwrap(),
                "--name",
                name,
                "--mode",
                "leader",
            ],
        );
        result(&out);
    }

    /// Ends the lease of the registry entry `name`: its `lease_until` becomes 1.
    fn expire(&self, name: &str) {
        let dir = moai_cli::lane_registry::registry_dir_unlinked(&self.project).unwrap();
        let file = dir.join(format!("{name}.json"));
        let mut entry: Value = serde_json::from_slice(&fs::read(&file).unwrap()).unwrap();
        entry["lease_until"] = json!(1);
        fs::write(&file, entry.to_string()).unwrap();
    }

    /// A ready card with one criterion and its own write scope; returns its name.
    fn ready_card(&self, scope: &str) -> String {
        let added = result(&self.card(
            None,
            &["add", "--title", "t", "--scope", scope, "--ac", "one"],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    fn take(&self, session: &str, card: &str) -> Output {
        self.card(Some(session), &["take", card])
    }

    /// `n` cards taken by `session`, each in `active`.
    fn working(&self, session: &str, n: usize) -> Vec<String> {
        (0..n)
            .map(|i| {
                let card = self.ready_card(&format!("src/w{i}/"));
                result(&self.take(session, &card));
                card
            })
            .collect()
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

/// REQ-7 (AC-6): with one lane the limit is two: a third take refuses and names the cards that
/// work, a second registered lane doubles the limit, and the ledger is unchanged by the refusal.
#[test]
fn take_stops_at_the_limit_and_a_second_lane_raises_it() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let held = w.working(&sid(1), 2);
    let third = w.ready_card("src/third/");

    let (kind, message) = refusal(&w.take(&sid(1), &third));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("work-in-progress limit"), "{message}");
    for card in &held {
        assert!(
            message.contains(card.as_str()),
            "{card} is named: {message}"
        );
    }
    let states = w.run(None, &["card", "list", w.project.to_str().unwrap()]);
    let listed = result(&states).to_string();
    assert!(
        listed.contains("ready"),
        "the third card is still ready: {listed}"
    );

    w.lane("agent-2", &sid(2));
    result(&w.take(&sid(2), &third));
}

/// REQ-8 (AC-6): at the limit `card add` refuses for a lane or the leader and passes for the
/// user.
#[test]
fn add_stops_at_the_limit_unless_the_user_adds() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.working(&sid(1), 2);

    let out = w.card(
        None,
        &[
            "add",
            "--title",
            "more",
            "--scope",
            "src/more/",
            "--ac",
            "one",
            "--by",
            "leader",
        ],
    );
    let (kind, message) = refusal(&out);
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("work-in-progress limit"), "{message}");

    let out = w.card(
        None,
        &[
            "add",
            "--title",
            "more",
            "--scope",
            "src/more/",
            "--ac",
            "one",
            "--by",
            "user",
        ],
    );
    result(&out);
}

/// REQ-9 (AC-6): `wip_limit` 0 turns the limit off; another number is the multiple per lane.
#[test]
fn a_wip_limit_of_zero_turns_the_limit_off() {
    let w = world(Some(0));
    w.lane("agent-1", &sid(1));
    w.working(&sid(1), 3);

    w.set_limit(Some(1));
    let fourth = w.ready_card("src/fourth/");
    let (kind, message) = refusal(&w.take(&sid(1), &fourth));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("work-in-progress limit"), "{message}");
}

/// REQ-10 (AC-7): at the limit `card next` names the review cards that no one audited and the
/// asking session did not write, in place of a card to take.
#[test]
fn next_at_the_limit_names_the_cards_that_wait_for_an_audit() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let held = w.working(&sid(1), 2);
    result(&w.card(Some(&sid(1)), &["review", &held[1], "--generation", "1"]));
    let waiting = w.ready_card("src/waiting/");

    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["at_limit"], json!(true), "{got}");
    assert_eq!(got["next_is"], json!("audit"), "{got}");
    assert_eq!(got["next"], json!(held[1]), "{got}");
    let names: Vec<&str> = got["audit_needed"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap())
        .collect();
    assert_eq!(names, vec![held[1].as_str()], "{got}");
    assert_ne!(got["next"], json!(waiting), "no card to take at the limit");

    // The session that wrote the card in review is not asked to audit it.
    let mine = result(&w.next(&sid(1)));
    assert_eq!(mine["audit_needed"], json!([]), "{mine}");
}

/// REQ-7 (AC-6): with no lane registered the limit is still one lane's, two: the first two takes
/// pass and the third refuses.
#[test]
fn with_no_lane_registered_the_limit_is_two() {
    let w = world(None);
    w.working(&sid(1), 2);
    let third = w.ready_card("src/third/");
    let (kind, message) = refusal(&w.take(&sid(1), &third));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("1 lane(s)"), "{message}");
}

/// REQ-7 (AC-6): a lane whose lease ended and a leader entry are no lanes; only the one live lane
/// counts, so the limit stays two and the third take refuses.
#[test]
fn a_lane_with_an_ended_lease_and_a_leader_entry_do_not_count() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.lane("agent-2", &sid(2));
    w.expire("agent-2");
    w.leader("leader-1", &sid(9));
    w.working(&sid(1), 2);
    let third = w.ready_card("src/third/");
    let (kind, message) = refusal(&w.take(&sid(1), &third));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("1 lane(s)"), "{message}");
}

/// REQ-10 (AC-7): at the limit with no review card to audit, `card next` names the limit and no
/// card.
#[test]
fn next_at_the_limit_with_nothing_to_audit_names_the_limit() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.working(&sid(1), 2);
    w.ready_card("src/waiting/");

    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["at_limit"], json!(true), "{got}");
    assert_eq!(got["next_is"], json!("limit"), "{got}");
    assert_eq!(got["next"], Value::Null, "{got}");
    assert_eq!(got["audit_needed"], json!([]), "{got}");
}
