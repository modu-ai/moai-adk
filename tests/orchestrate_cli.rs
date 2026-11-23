//! `moai card orchestrate plan` (SPEC-20261124-orchestrate-plan): the coordination window's round
//! as one read-only plan over the ledger and the lane registry.
//!
//! Every command runs against a temporary HOME and a temporary git project; the real `~/.moai`
//! and the real registry are not read or written. Ledger rows are inserted with SQL so their
//! holder, state and time are exactly what a case needs. No `claude` or `orca` process is
//! started: the host state is a fixture file in the shape of `claude agents --json`.
#![cfg(unix)]

mod common;

use common::git::git;
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Output;

const S_LEADER: &str = "eeeeeeee-0000-4000-8000-00000000000e";
const S_ONE: &str = "11111111-0000-4000-8000-000000000001";
const S_TWO: &str = "22222222-0000-4000-8000-000000000002";
const S_THREE: &str = "33333333-0000-4000-8000-000000000003";
const S_AUDITOR: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const HEAD_A: &str = "0123456789abcdef0123456789abcdef01234567";
const HEAD_B: &str = "89abcdef0123456789abcdef0123456789abcdef";
const HEAD_C: &str = "fedcba9876543210fedcba9876543210fedcba98";
const HEAD_D: &str = "00112233445566778899aabbccddeeff00112233";
const BASE: &str = "ffffffffffffffffffffffffffffffffffffffff";

struct World {
    home: TempHome,
    project: PathBuf,
    events: std::cell::Cell<i64>,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    git(&project, &["init", "-q", "-b", "main"]);
    ledger::register(home.path(), "demo").unwrap();
    World {
        home,
        project,
        events: std::cell::Cell::new(0),
    }
}

fn text(out: &Output) -> String {
    format!(
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn result(out: &Output) -> Value {
    assert_eq!(out.status.code(), Some(0), "{}", text(out));
    let v: Value =
        serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {}", text(out)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

impl World {
    fn project_str(&self) -> String {
        self.project.to_str().unwrap().to_string()
    }

    /// `moai <args>` as the host session `session` (none: no session id at all).
    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env_remove("CLAUDE_CODE_SESSION_ID")
            .env_remove("CODEX_THREAD_ID");
        if let Some(s) = session {
            cmd.env("CLAUDE_CODE_SESSION_ID", s);
        }
        cmd.output().expect("run moai")
    }

    /// `moai card orchestrate plan <project> [extra]` as `session`.
    fn plan_as(&self, session: Option<&str>, extra: &[&str]) -> Output {
        let project = self.project_str();
        let mut args = vec!["card", "orchestrate", "plan", project.as_str()];
        args.extend_from_slice(extra);
        self.run(session, &args)
    }

    fn plan(&self) -> Value {
        result(&self.plan_as(None, &[]))
    }

    fn register(&self, session: Option<&str>, name: &str, mode: &str) {
        let project = self.project_str();
        let out = self.run(
            session,
            &["session-register", &project, "--name", name, "--mode", mode],
        );
        assert_eq!(out.status.code(), Some(0), "{}", text(&out));
    }

    fn lane(&self, name: &str, session: &str) {
        self.register(Some(session), name, "agent");
    }

    fn sessions_dir(&self) -> PathBuf {
        moai_cli::clone_root::path(&self.project.join(".git")).join("sessions")
    }

    /// Adds `previous_sessions` to a registered entry, as `session-register --continue` leaves it.
    fn previous(&self, name: &str, previous: &[&str]) {
        let file = self.sessions_dir().join(format!("{name}.json"));
        let mut v: Value = serde_json::from_slice(&fs::read(&file).unwrap()).unwrap();
        v["previous_sessions"] = json!(previous);
        fs::write(&file, serde_json::to_vec(&v).unwrap()).unwrap();
    }

    fn sql(&self, sql: &str, params: impl rusqlite::Params) {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn().execute(sql, params).unwrap();
    }

    /// A card with one open criterion; `kind` `research` needs no SPEC.
    fn card(&self, id: i64, kind: &str, state: &str, priority: &str, scope: &[&str]) {
        self.sql(
            "INSERT INTO card (card_id, title, kind, state, priority, write_scope, created_by, \
             created_at, updated_at, closed_at) \
             VALUES (?1, ?2, ?3, ?4, ?5, ?6, 'tester', 1, 2, CASE WHEN ?4 = 'done' THEN 3 END)",
            rusqlite::params![
                id,
                format!("card {id}"),
                kind,
                state,
                priority,
                serde_json::to_string(scope).unwrap()
            ],
        );
        self.sql(
            "INSERT INTO card_ac (card_id, seq, text, status, updated_at) \
             VALUES (?1, 1, 'one criterion', 'open', 2)",
            rusqlite::params![id],
        );
    }

    fn reserve(&self, card: i64, generation: i64, holder: &str, status: &str) {
        self.sql(
            "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
             lease_until, updated_at, worktree, branch) \
             VALUES (?1, ?2, ?3, ?4, 100, 99999999999, 100, NULL, NULL)",
            rusqlite::params![card, generation, holder, status],
        );
    }

    fn note(&self, card: i64, author: &str, body: &str) {
        self.sql(
            "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
             VALUES (?1, 'report', ?2, 'local', ?3, 1000)",
            rusqlite::params![card, author, body],
        );
    }

    fn verdict(&self, card: i64, head: &str, pass: bool, reviewer: &str) {
        let n = self.events.get() + 1;
        self.events.set(n);
        let payload = json!({
            "head": head, "base": BASE, "verdict": if pass { "pass" } else { "fail" },
            "tier": 1, "evidence": "read and ran", "receipt": null, "reviewer": reviewer,
        });
        self.sql(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, 500, ?2, ?3, 'card.verdict', ?4)",
            rusqlite::params![
                format!("test-verdict-{n}"),
                reviewer,
                card,
                payload.to_string()
            ],
        );
    }

    fn ready_note(&self, card: i64, head: &str) {
        self.note(
            card,
            "agent-1",
            &format!("LANE-READY-INTEGRATION C-{card} head={head}\nbase={BASE}\nchecks passed"),
        );
    }

    fn agents_file(&self, body: &Value) -> String {
        let path = self.home.path().join("agents.json");
        fs::write(&path, serde_json::to_vec(body).unwrap()).unwrap();
        path.to_str().unwrap().to_string()
    }
}

fn names(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|x| x["name"].as_str().unwrap().to_string())
        .collect()
}

fn actions_of<'a>(plan: &'a Value, kind: &str) -> Vec<&'a Value> {
    plan["actions"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|a| a["kind"] == json!(kind))
        .collect()
}

fn lane<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["lanes"]
        .as_array()
        .unwrap()
        .iter()
        .find(|l| l["name"] == json!(name))
        .unwrap_or_else(|| panic!("no lane {name} in {plan}"))
}

// ---- plan_ (AC-1: REQ-1, REQ-2, REQ-11) -------------------------------------------------------

#[test]
fn plan_prints_the_six_parts_and_writes_nothing() {
    let w = world();
    w.register(Some(S_LEADER), "leader", "leader");
    w.lane("agent-1", S_ONE);
    w.card(1, "research", "ready", "p1", &["docs/a.md"]);
    let before = w.home.snapshot();
    let plan = result(&w.plan_as(Some(S_LEADER), &[]));
    for part in [
        "window",
        "lanes",
        "actions",
        "land_queue",
        "attention",
        "host",
    ] {
        assert!(plan.get(part).is_some(), "missing {part}: {plan}");
    }
    assert_eq!(plan["writes"], json!(false));
    assert_eq!(plan["ledger"], json!("ready"));
    assert_eq!(plan["window"]["status"], json!("registered"));
    assert!(plan["window"]["lease_until"].is_string(), "{plan}");
    assert_eq!(
        before,
        w.home.snapshot(),
        "the ledger file, the registry folder and the state folder stay as they were"
    );
}

#[test]
fn plan_of_a_project_without_a_ledger_is_empty_and_exits_zero() {
    let home = TempHome::new();
    let project = home.project("bare");
    git(&project, &["init", "-q", "-b", "main"]);
    let before = home.snapshot();
    let out = home
        .command(&["card", "orchestrate", "plan", project.to_str().unwrap()])
        .output()
        .unwrap();
    let plan = result(&out);
    assert_eq!(plan["ledger"], json!("absent"));
    for part in ["lanes", "actions", "land_queue", "attention"] {
        assert_eq!(plan[part], json!([]), "{part}: {plan}");
    }
    assert_eq!(plan["writes"], json!(false));
    assert_eq!(before, home.snapshot());
}

#[test]
fn plan_window_reads_registered_other_and_none() {
    let w = world();
    assert_eq!(w.plan()["window"]["status"], json!("none"));
    w.register(Some(S_LEADER), "leader", "leader");
    let mine = result(&w.plan_as(Some(S_LEADER), &[]));
    assert_eq!(mine["window"]["status"], json!("registered"));
    let theirs = result(&w.plan_as(Some(S_ONE), &[]));
    assert_eq!(theirs["window"]["status"], json!("other"));
    // A plan never registers or renews: the entry is as the registration wrote it.
    let entry = fs::read(w.sessions_dir().join("leader.json")).unwrap();
    let _ = w.plan_as(Some(S_ONE), &[]);
    assert_eq!(
        entry,
        fs::read(w.sessions_dir().join("leader.json")).unwrap()
    );
}

// ---- lanes_ (AC-2: REQ-3, REQ-4) --------------------------------------------------------------

#[test]
fn lanes_are_busy_free_or_unknown_by_their_reservations() {
    let w = world();
    w.lane("agent-1", S_ONE);
    w.lane("agent-2", S_TWO);
    w.register(None, "agent-3", "agent"); // no host session id: cannot be matched
    w.register(Some(S_THREE), "agent-4", "agent");
    w.card(1, "research", "active", "p1", &["docs/a.md"]);
    w.reserve(1, 1, S_ONE, "held");
    // agent-4 was cleared once: its earlier session held the card.
    w.card(2, "research", "active", "p1", &["docs/b.md"]);
    w.reserve(2, 1, S_TWO, "held");
    w.previous("agent-4", &[S_TWO]);
    let plan = w.plan();
    assert_eq!(lane(&plan, "agent-1")["state"], json!("busy"));
    assert_eq!(lane(&plan, "agent-2")["state"], json!("busy"));
    assert_eq!(lane(&plan, "agent-3")["state"], json!("unknown"));
    assert!(lane(&plan, "agent-3")["reason"].is_string());
    assert_eq!(
        lane(&plan, "agent-4")["state"],
        json!("busy"),
        "a previous session of the lane that holds a card makes it busy: {plan}"
    );
    // A leader entry is no lane.
    w.register(Some(S_LEADER), "leader", "leader");
    let plan = w.plan();
    assert!(
        names(&plan["lanes"]).iter().all(|n| n != "leader"),
        "{plan}"
    );
}

#[test]
fn lanes_a_free_lane_gets_the_first_ranked_card_and_the_next_lane_the_next() {
    let w = world();
    w.lane("agent-1", S_ONE);
    w.lane("agent-2", S_TWO);
    w.lane("agent-3", S_THREE);
    w.card(5, "research", "ready", "p2", &["docs/five.md"]);
    w.card(6, "research", "ready", "p0", &["docs/six.md"]);
    w.card(7, "research", "ready", "p1", &["docs/seven.md"]);
    let plan = w.plan();
    let dispatches: Vec<(String, String)> = actions_of(&plan, "dispatch")
        .iter()
        .map(|a| {
            (
                a["lane"].as_str().unwrap().to_string(),
                a["card"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    assert_eq!(
        dispatches,
        vec![
            ("agent-1".to_string(), "C-6".to_string()),
            ("agent-2".to_string(), "C-7".to_string()),
            ("agent-3".to_string(), "C-5".to_string()),
        ],
        "{plan}"
    );
    // One dispatch per lane at most, and no card twice.
    let mut cards: Vec<&str> = actions_of(&plan, "dispatch")
        .iter()
        .map(|a| a["card"].as_str().unwrap())
        .collect();
    cards.sort_unstable();
    cards.dedup();
    assert_eq!(cards.len(), 3);
}

#[test]
fn lanes_lane_order_is_numeric_not_alphabetical() {
    let w = world();
    w.lane("agent-10", S_ONE);
    w.lane("agent-2", S_TWO);
    w.card(5, "research", "ready", "p0", &["docs/five.md"]);
    w.card(6, "research", "ready", "p1", &["docs/six.md"]);
    let plan = w.plan();
    let d = actions_of(&plan, "dispatch");
    assert_eq!(d[0]["lane"], json!("agent-2"), "{plan}");
    assert_eq!(d[1]["lane"], json!("agent-10"), "{plan}");
}

#[test]
fn lanes_cards_that_cannot_be_dispatched_are_blocked_with_a_reason() {
    let w = world();
    w.lane("agent-1", S_ONE);
    // An empty scope, a held card, and a repository change whose SPEC is not approved.
    w.card(8, "research", "ready", "p0", &[]);
    w.card(9, "research", "active", "p0", &["docs/nine.md"]);
    w.reserve(9, 1, S_TWO, "held");
    w.card(10, "feature", "ready", "p0", &["src/ten.rs"]);
    let plan = w.plan();
    assert!(actions_of(&plan, "dispatch").is_empty(), "{plan}");
    let blocked: Vec<(String, String)> = actions_of(&plan, "blocked")
        .iter()
        .map(|a| {
            (
                a["card"].as_str().unwrap().to_string(),
                a["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    assert!(
        blocked.contains(&("C-8".into(), "empty_scope".into())),
        "{blocked:?}"
    );
    assert!(
        blocked.contains(&("C-9".into(), "held".into())),
        "{blocked:?}"
    );
    assert!(
        blocked.contains(&("C-10".into(), "spec_not_approved".into())),
        "{blocked:?}"
    );
}

// ---- clear_ (AC-3: REQ-5) ---------------------------------------------------------------------

#[test]
fn clear_a_lane_whose_card_closed_gets_a_clear_with_no_channel() {
    let w = world();
    w.lane("agent-1", S_ONE);
    w.card(1, "research", "done", "p1", &["docs/a.md"]);
    w.reserve(1, 1, S_ONE, "released");
    let plan = w.plan();
    let clears = actions_of(&plan, "clear");
    assert_eq!(clears.len(), 1, "{plan}");
    assert_eq!(clears[0]["lane"], json!("agent-1"));
    assert_eq!(clears[0]["card"], json!("C-1"));
    assert_eq!(clears[0]["channel"], json!("none"));
    let reason = clears[0]["reason"].as_str().unwrap();
    assert!(reason.contains("/clear"), "{reason}");
    assert!(reason.contains("message"), "{reason}");
}

#[test]
fn clear_a_lane_that_holds_a_card_gets_none_even_with_a_done_note() {
    let w = world();
    w.lane("agent-1", S_ONE);
    w.lane("agent-2", S_TWO);
    // agent-1 finished C-1 and holds C-2; agent-2 wrote a LANE-DONE line the ledger does not back.
    w.card(1, "research", "done", "p1", &["docs/a.md"]);
    w.reserve(1, 1, S_ONE, "released");
    w.card(2, "research", "active", "p1", &["docs/b.md"]);
    w.reserve(2, 1, S_ONE, "held");
    w.card(3, "research", "active", "p1", &["docs/c.md"]);
    w.reserve(3, 1, S_TWO, "held");
    w.note(3, "agent-2", "LANE-DONE agent-2 C-3 head abc — done");
    let plan = w.plan();
    assert!(actions_of(&plan, "clear").is_empty(), "{plan}");
}

// ---- audit_ (AC-4: REQ-6) ---------------------------------------------------------------------

#[test]
fn audit_a_ready_head_without_a_verdict_gets_one_audit_that_names_the_author() {
    let w = world();
    w.card(1, "research", "review", "p1", &["docs/a.md"]);
    w.reserve(1, 1, S_ONE, "held_for_review");
    w.ready_note(1, HEAD_A);
    let plan = w.plan();
    let audits = actions_of(&plan, "audit");
    assert_eq!(audits.len(), 1, "{plan}");
    let a = audits[0];
    assert_eq!(a["card"], json!("C-1"));
    assert_eq!(a["head"], json!(HEAD_A));
    assert_eq!(a["base"], json!(BASE));
    assert_eq!(a["author_sessions"], json!([S_ONE]));
    let rule = a["runner_rule"].as_str().unwrap();
    assert!(rule.contains("differs"), "{rule}");
    // The printed rule never names the author as the runner.
    assert!(!rule.contains(S_ONE), "{rule}");
    assert!(a.get("runner").is_none(), "no runner is chosen: {a}");
}

#[test]
fn audit_a_head_with_a_verdict_gets_none_and_a_new_head_gets_one_again() {
    let w = world();
    w.card(1, "research", "review", "p1", &["docs/a.md"]);
    w.reserve(1, 1, S_ONE, "held_for_review");
    w.ready_note(1, HEAD_A);
    w.verdict(1, HEAD_A, false, S_AUDITOR);
    assert!(actions_of(&w.plan(), "audit").is_empty());
    // A repair moves the head: the new head has no verdict yet.
    w.ready_note(1, HEAD_B);
    let plan = w.plan();
    let audits = actions_of(&plan, "audit");
    assert_eq!(audits.len(), 1, "{plan}");
    assert_eq!(audits[0]["head"], json!(HEAD_B));
}

// ---- land_ (AC-5: REQ-7) ----------------------------------------------------------------------

#[test]
fn land_lists_independent_passes_by_priority_then_number() {
    let w = world();
    // C-4 (p1) and C-3 (p0) have independent passes; C-5 failed; C-6 has no verdict;
    // C-7's only pass was written by the card's own author session.
    for (id, prio) in [(3, "p0"), (4, "p1"), (5, "p0"), (6, "p0"), (7, "p0")] {
        w.card(id, "research", "review", prio, &["docs/x.md"]);
    }
    w.reserve(3, 1, S_ONE, "held_for_review");
    w.reserve(4, 1, S_TWO, "held_for_review");
    w.reserve(5, 1, S_ONE, "held_for_review");
    w.reserve(6, 1, S_ONE, "held_for_review");
    w.reserve(7, 1, S_ONE, "held_for_review");
    w.ready_note(3, HEAD_A);
    w.ready_note(4, HEAD_B);
    w.ready_note(5, HEAD_C);
    w.ready_note(6, HEAD_D);
    w.ready_note(7, HEAD_D);
    w.verdict(3, HEAD_A, true, S_AUDITOR);
    w.verdict(4, HEAD_B, true, S_AUDITOR);
    w.verdict(5, HEAD_C, false, S_AUDITOR);
    w.verdict(7, HEAD_D, true, S_ONE);
    let before = w.home.snapshot();
    let plan = w.plan();
    let queue = plan["land_queue"].as_array().unwrap();
    let cards: Vec<&str> = queue.iter().map(|q| q["card"].as_str().unwrap()).collect();
    assert_eq!(cards, vec!["C-3", "C-4"], "{plan}");
    assert_eq!(
        queue[0]["command"],
        json!(format!("moai card land C-3 --head {HEAD_A} --by leader"))
    );
    assert_eq!(queue[0]["reviewer"], json!(S_AUDITOR));
    assert_eq!(before, w.home.snapshot(), "the plan lands nothing");
}

#[test]
fn land_a_later_fail_removes_the_head_from_the_queue() {
    let w = world();
    w.card(1, "research", "review", "p1", &["docs/a.md"]);
    w.reserve(1, 1, S_ONE, "held_for_review");
    w.ready_note(1, HEAD_A);
    w.verdict(1, HEAD_A, true, S_AUDITOR);
    assert_eq!(w.plan()["land_queue"].as_array().unwrap().len(), 1);
    w.verdict(1, HEAD_A, false, S_AUDITOR);
    assert!(w.plan()["land_queue"].as_array().unwrap().is_empty());
}

// ---- attention_ (AC-6: REQ-8, REQ-10) ---------------------------------------------------------

#[test]
fn attention_lists_decisions_and_waiting_sessions_and_answers_none() {
    let w = world();
    w.lane("agent-1", S_ONE);
    w.lane("agent-2", S_TWO);
    w.card(1, "research", "active", "p1", &["docs/a.md"]);
    w.reserve(1, 1, S_ONE, "held");
    w.note(
        1,
        "agent-1",
        "LANE-BLOCKED kind=decision C-1 which option?\noptions a b",
    );
    w.card(2, "research", "active", "p1", &["docs/b.md"]);
    w.reserve(2, 1, S_TWO, "held");
    w.note(2, "agent-2", "LANE-BLOCKED kind=decision C-2 answered one");
    w.note(2, "leader", "answer for C-2");
    let agents = w.agents_file(&json!([
        {"id": "ab12cd34", "sessionId": S_ONE, "name": "agent-1", "kind": "background",
         "state": "blocked", "status": "waiting", "waitingFor": "permission prompt"},
        {"id": "ef56ab78", "sessionId": S_TWO, "name": "agent-2", "kind": "background",
         "state": "working", "status": "busy"}
    ]));
    let plan = result(&w.plan_as(None, &["--agents-json", &agents]));
    let items = plan["attention"].as_array().unwrap();
    let notes: Vec<&Value> = items
        .iter()
        .filter(|i| i["kind"] == json!("note"))
        .collect();
    assert_eq!(notes.len(), 1, "the answered one is gone: {plan}");
    assert_eq!(notes[0]["card"], json!("C-1"));
    assert_eq!(notes[0]["family"], json!("blocked"));
    let sessions: Vec<&Value> = items
        .iter()
        .filter(|i| i["kind"] == json!("session"))
        .collect();
    assert_eq!(sessions.len(), 1, "{plan}");
    assert_eq!(sessions[0]["waitingFor"], json!("permission prompt"));
    assert_eq!(sessions[0]["answered_at"], json!("claude attach ab12cd34"));
    for item in items {
        assert!(item["answer"].is_null(), "no answer is written: {item}");
        assert!(item["key"].is_string(), "{item}");
    }
    assert_eq!(plan["host"]["status"], json!("read"));
    assert_eq!(
        lane(&plan, "agent-1")["host"]["waitingFor"],
        json!("permission prompt")
    );
    assert_eq!(lane(&plan, "agent-2")["host"]["state"], json!("working"));
}

#[test]
fn attention_an_unreadable_agent_file_is_unread_and_no_file_is_not_read() {
    let w = world();
    assert_eq!(w.plan()["host"]["status"], json!("not_read"));
    let bad = w.home.path().join("bad.json");
    fs::write(&bad, "this is not json").unwrap();
    let plan = result(&w.plan_as(None, &["--agents-json", bad.to_str().unwrap()]));
    assert_eq!(plan["host"]["status"], json!("unread"));
    assert!(plan["host"]["reason"].is_string());
    let unread: Vec<&Value> = plan["attention"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["status"] == json!("unread"))
        .collect();
    assert_eq!(unread.len(), 1, "{plan}");
    // A missing file is unread too.
    let missing = w.home.path().join("missing.json");
    let plan = result(&w.plan_as(None, &["--agents-json", missing.to_str().unwrap()]));
    assert_eq!(plan["host"]["status"], json!("unread"));
}

// ---- keys_ (AC-7: REQ-9) ----------------------------------------------------------------------

#[test]
fn keys_two_plans_of_one_ledger_are_equal_and_every_action_has_a_key() {
    let w = world();
    w.register(Some(S_LEADER), "leader", "leader");
    w.lane("agent-1", S_ONE);
    w.lane("agent-2", S_TWO);
    w.card(1, "research", "done", "p1", &["docs/a.md"]);
    w.reserve(1, 1, S_ONE, "released");
    w.card(2, "research", "ready", "p1", &["docs/b.md"]);
    w.card(3, "research", "review", "p1", &["docs/c.md"]);
    w.reserve(3, 1, S_TWO, "held_for_review");
    w.ready_note(3, HEAD_A);
    let first = w.plan_as(Some(S_LEADER), &[]);
    let second = w.plan_as(Some(S_LEADER), &[]);
    assert_eq!(first.stdout, second.stdout, "byte for byte equal");
    let plan = result(&first);
    let actions = plan["actions"].as_array().unwrap();
    assert!(!actions.is_empty());
    let shape = |key: &str| {
        let parts: Vec<&str> = key.split(':').collect();
        parts.len() == 3
            && ["clear", "dispatch", "audit", "blocked"].contains(&parts[0])
            && parts.iter().all(|p| !p.is_empty())
    };
    for a in actions {
        let key = a["key"].as_str().unwrap_or_else(|| panic!("no key: {a}"));
        assert!(shape(key), "{key}");
    }
    let mut keys: Vec<&str> = actions.iter().map(|a| a["key"].as_str().unwrap()).collect();
    keys.sort_unstable();
    keys.dedup();
    assert_eq!(keys.len(), actions.len(), "keys are unique");
}

#[test]
fn keys_a_new_note_changes_only_the_keys_it_answers() {
    let w = world();
    w.lane("agent-1", S_ONE);
    w.lane("agent-2", S_TWO);
    w.card(1, "research", "review", "p1", &["docs/a.md"]);
    w.reserve(1, 1, S_ONE, "held_for_review");
    w.ready_note(1, HEAD_A);
    let before = w.plan();
    w.card(2, "research", "active", "p1", &["docs/b.md"]);
    w.reserve(2, 1, S_TWO, "held");
    w.note(2, "agent-2", "LANE-BLOCKED kind=decision C-2 new question");
    let after = w.plan();
    assert_eq!(
        before["actions"], after["actions"],
        "unrelated actions keep their keys"
    );
    assert_eq!(
        after["attention"].as_array().unwrap().len(),
        before["attention"].as_array().unwrap().len() + 1
    );
}

// ---- text_ (AC-8: REQ-12) ---------------------------------------------------------------------

#[test]
fn text_help_and_plan_hold_no_key_shaped_text_and_no_outside_path() {
    let w = world();
    w.register(Some(S_LEADER), "leader", "leader");
    w.lane("agent-1", S_ONE);
    w.card(1, "research", "ready", "p1", &["docs/a.md"]);
    let help = w
        .home
        .command(&["card", "orchestrate", "--help"])
        .output()
        .unwrap();
    assert_eq!(help.status.code(), Some(0), "{}", text(&help));
    let help_text = text(&help);
    assert!(help_text.contains("Reads the ledger only"), "{help_text}");
    let plan_text = text(&w.plan_as(Some(S_LEADER), &[]));
    let home = w.home.path().to_str().unwrap().to_string();
    for (name, body) in [("help", &help_text), ("plan", &plan_text)] {
        assert!(!body.contains(&home), "{name} names the temporary home");
        assert!(!body.contains("/Users/"), "{name} names a real home");
        for needle in ["sk-ant", "ghp_", "BEGIN PRIVATE KEY", "Bearer "] {
            assert!(!body.contains(needle), "{name} holds {needle}");
        }
    }
}
