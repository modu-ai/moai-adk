//! leader-round slice 1: the per-lane summary joins the registry entry's `session` with the
//! ledger reservation's `holder`. Every scene builds a temporary HOME with a temporary
//! ledger and a temporary registry directory; nothing here reads or writes the real HOME,
//! and the summary itself writes nothing (the ledger file's digest is compared).

mod common;

use common::home::TempHome;
use moai_cli::lane_ledger::lane_summaries;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    path::{Path, PathBuf},
    process::Output,
    time::{SystemTime, UNIX_EPOCH},
};

struct World {
    home: TempHome,
    project: String,
    entries: PathBuf,
}

fn now_secs() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo").to_str().unwrap().to_string();
    moai_cli::ledger::register(home.path(), "demo").unwrap();
    let entries = home.path().join("registry-entries");
    fs::create_dir_all(&entries).unwrap();
    World {
        home,
        project,
        entries,
    }
}

impl World {
    fn entry(&self, name: &str, session: &str) {
        let now = now_secs();
        fs::write(
            self.entries.join(format!("{name}.json")),
            json!({
                "schema_version": 1, "name": name, "instance": "0".repeat(32),
                "session": session, "lease_until": now + 3600,
            })
            .to_string(),
        )
        .unwrap();
    }

    /// One `moai` run as the named session.
    fn run(&self, sess: &str, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env("CLAUDE_CODE_SESSION_ID", sess);
        cmd.output().expect("run moai")
    }

    fn run_ok(&self, sess: &str, args: &[&str]) -> Value {
        let out = self.run(sess, args);
        assert_eq!(
            out.status.code(),
            Some(0),
            "{args:?}: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    /// Takes `card` for `sess` and returns the generation the take answered. A fresh card is
    /// `draft`, so the take is preceded by its `ready` (ignored when already ready).
    fn take(&self, sess: &str, card: &str) -> i64 {
        let _ = self
            .run(sess, &["card", "ready", card, "--project", &self.project])
            .status;
        let v = self.run_ok(
            sess,
            &[
                "card",
                "take",
                card,
                "--project",
                &self.project,
                "--lease-secs",
                "3600",
                "--scope",
                "src",
            ],
        );
        v["result"]["reservation"]["generation"]
            .as_i64()
            .expect("take answers a generation")
    }

    fn ledger_digest(&self) -> String {
        let bytes = fs::read(self.home.path().join(".moai/projects/demo/ledger.db"))
            .expect("the ledger file exists");
        format!("{:x}", Sha256::digest(&bytes))
    }

    fn summary(&self) -> Value {
        lane_summaries(
            &self.entries,
            self.home.path(),
            Path::new(&self.project),
            now_secs(),
        )
    }

    fn lane<'v>(&'v self, v: &'v Value, name: &str) -> &'v Value {
        v["lanes"]["items"]
            .as_array()
            .expect("lanes items")
            .iter()
            .find(|l| l["lane"] == json!(name))
            .unwrap_or_else(|| panic!("no lane {name}: {v}"))
    }
}

/// Scene 1 (REQ-5): a lane whose session holds a card whose newest note opens MERGE-READY.
#[test]
fn a_lane_with_a_merge_ready_card_lists_it() {
    let w = world();
    w.entry("agent-1", "aaaa0001");
    w.run_ok(
        "aaaa0001",
        &[
            "card",
            "add",
            "--title",
            "the merge ready card",
            "--project",
            &w.project,
        ],
    );
    w.take("aaaa0001", "C-1");
    let w2 = moai_cli::ledger::open_write(w.home.path(), "demo").unwrap();
    w2.conn()
        .execute(
            "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
             VALUES (1, 'report', 'agent-1', 'local', 'MERGE-READY C-1 audit PASS', 1)",
            [],
        )
        .unwrap();
    drop(w2);
    let digest = w.ledger_digest();
    let v = w.summary();
    assert_eq!(v["lanes"]["state"], json!("run"), "{v}");
    let lane = w.lane(&v, "agent-1");
    assert_eq!(lane["idle"], json!(false));
    let cards = lane["cards"].as_array().unwrap();
    assert_eq!(cards.len(), 1, "{v}");
    assert_eq!(cards[0]["card"], json!("C-1"));
    assert_eq!(cards[0]["status"], json!("held"));
    assert_eq!(cards[0]["newest_note_kind"], json!("MERGE-READY"));
    assert!(cards[0]["lease_left"].as_i64().unwrap() > 0, "{cards:?}");
    assert_eq!(cards[0]["landed_waiting_finish"], json!(false));
    assert_eq!(w.ledger_digest(), digest, "the summary wrote nothing");
}

/// Scene 2 (REQ-5): a lane with SWITCH and a released card is not idle — the SWITCH asks the
/// leader.
#[test]
fn a_lane_with_switch_and_a_released_card_is_not_idle() {
    let w = world();
    w.entry("agent-2", "aaaa0002");
    w.run_ok(
        "aaaa0002",
        &[
            "card",
            "add",
            "--title",
            "the switched card",
            "--project",
            &w.project,
        ],
    );
    let gen = w.take("aaaa0002", "C-1");
    let done = w.run(
        "aaaa0002",
        &[
            "card",
            "release",
            "C-1",
            "--generation",
            &gen.to_string(),
            "--project",
            &w.project,
        ],
    );
    assert_eq!(
        done.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&done.stderr)
    );
    let w2 = moai_cli::ledger::open_write(w.home.path(), "demo").unwrap();
    w2.conn()
        .execute(
            "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
             VALUES (1, 'comment', 'agent-2', 'local', 'SWITCH C-1 handed to the leader', 1)",
            [],
        )
        .unwrap();
    drop(w2);
    let v = w.summary();
    let lane = w.lane(&v, "agent-2");
    let cards = lane["cards"].as_array().unwrap();
    assert_eq!(cards.len(), 1, "{v}");
    assert_eq!(cards[0]["status"], json!("released"));
    assert_eq!(cards[0]["newest_note_kind"], json!("SWITCH"));
    assert_eq!(lane["idle"], json!(false));
}

/// Scene 3 (REQ-5): a held card recorded landed while the card is not closed waits for the
/// finish, and the line says so.
#[test]
fn a_landed_open_card_waits_for_the_finish() {
    let w = world();
    w.entry("agent-3", "aaaa0003");
    w.run_ok(
        "aaaa0003",
        &[
            "card",
            "add",
            "--title",
            "the landed card",
            "--project",
            &w.project,
        ],
    );
    let gen = w.take("aaaa0003", "C-1");
    let head = "a".repeat(40);
    let review = w.run(
        "aaaa0003",
        &[
            "card",
            "review",
            "C-1",
            "--generation",
            &gen.to_string(),
            "--project",
            &w.project,
        ],
    );
    assert_eq!(
        review.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&review.stderr)
    );
    let landed = w.run(
        "aaaa0003",
        &[
            "card",
            "landed",
            "C-1",
            "--head",
            &head,
            "--by",
            "agent-3",
            "--project",
            &w.project,
        ],
    );
    assert_eq!(
        landed.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&landed.stderr)
    );
    let v = w.summary();
    let lane = w.lane(&v, "agent-3");
    let cards = lane["cards"].as_array().unwrap();
    assert_eq!(cards.len(), 1, "{v}");
    assert_eq!(cards[0]["status"], json!("held"));
    assert_eq!(cards[0]["landed_waiting_finish"], json!(true), "{cards:?}");
}

/// Scene 4 (REQ-6): a registered lane whose session holds nothing is idle.
#[test]
fn a_lane_holding_nothing_is_idle() {
    let w = world();
    w.entry("agent-4", "aaaa0004");
    let v = w.summary();
    let lane = w.lane(&v, "agent-4");
    assert_eq!(lane["cards"], json!([]));
    assert_eq!(lane["idle"], json!(true));
}

/// Scene 5 (REQ-7): after a clear the registry names the new session, the old session's
/// cards belong to no listed lane, and nothing is reported stale.
#[test]
fn a_session_that_changed_after_a_clear_holds_nothing() {
    let w = world();
    w.entry("agent-5", "aaaa0005");
    w.run_ok(
        "bbbb0001",
        &[
            "card",
            "add",
            "--title",
            "the cleared card",
            "--project",
            &w.project,
        ],
    );
    w.take("bbbb0001", "C-1");
    let v = w.summary();
    let lane = w.lane(&v, "agent-5");
    assert_eq!(lane["session"], json!("aaaa0005"));
    assert_eq!(lane["cards"], json!([]));
    assert_eq!(lane["idle"], json!(true));
    assert!(
        v.to_string().contains("bbbb0001") == false,
        "the old session id is not reported: {v}"
    );
}
