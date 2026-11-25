//! SPEC-20261125-session-registry-session-key, the limit side (REQ-6, AC-2): the work-in-progress
//! limit counts each live registered session once, whatever the form of its entry file, and the
//! multiple per lane stays 2. The calls run in the library on a temporary clone with hand-written
//! entries; a temporary HOME stands beside them as the only home a run of the binary would see,
//! and no real registry of this machine is read or written.
mod common;

use moai_cli::card::wip_limit::{live_lanes, per_lane, DEFAULT_PER_LANE};
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;

/// A lease that has not ended, whatever the clock says.
const LIVE: i64 = 4_102_444_800;
/// A lease that ended long ago.
const ENDED: i64 = 1;

struct Fixture {
    _dir: tempfile::TempDir,
    _home: common::home::TempHome,
    repo: PathBuf,
    sessions: PathBuf,
}

impl Fixture {
    fn new() -> Self {
        let dir = tempfile::tempdir().unwrap();
        let repo = fs::canonicalize(dir.path()).unwrap().join("demo");
        fs::create_dir_all(&repo).unwrap();
        common::git::git_ok(&repo, &["init", "-q", "-b", "main"]);
        common::git::git_ok(
            &repo,
            &[
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "commit",
                "-q",
                "--allow-empty",
                "-m",
                "base",
            ],
        );
        let sessions = moai_cli::clone_root::path(&repo.join(".git")).join("sessions");
        fs::create_dir_all(&sessions).unwrap();
        Fixture {
            _dir: dir,
            _home: common::home::TempHome::new(),
            repo,
            sessions,
        }
    }

    /// Writes the entry file `<file>.json` for the session `name`, with the optional host session
    /// id and mode.
    fn entry(&self, file: &str, name: &str, session: Option<&str>, mode: &str, lease: i64) {
        let mut v: Value = json!({
            "schema_version": 1,
            "name": name,
            "instance": "0123456789abcdef0123456789abcdef",
            "lease_until": lease,
            "mode": mode,
        });
        if let Some(s) = session {
            v["session"] = json!(s);
        }
        fs::write(
            self.sessions.join(format!("{file}.json")),
            serde_json::to_vec(&v).unwrap(),
        )
        .unwrap();
    }

    fn lanes(&self) -> usize {
        live_lanes(&self.repo)
    }
}

/// REQ-6: two entries that carry one host session id are one live session, so the limit counts
/// one lane.
#[test]
fn ac2_live_lanes_counts_a_session_once_for_two_entry_forms() {
    let fx = Fixture::new();
    fx.entry("agent-ben", "agent-ben", Some("sess-one"), "agent", LIVE);
    fx.entry(
        "demo-agent-ben",
        "demo-agent-ben",
        Some("sess-one"),
        "agent",
        LIVE,
    );
    assert_eq!(fx.lanes(), 1);
    // A second session is a second lane.
    fx.entry(
        "demo-agent-cara",
        "demo-agent-cara",
        Some("sess-two"),
        "agent",
        LIVE,
    );
    assert_eq!(fx.lanes(), 2);
}

/// REQ-6: a leader entry and an entry whose lease has ended are not lanes, and an ended entry of
/// a session does not hide its live one.
#[test]
fn ac2_live_lanes_leaves_out_leaders_and_ended_leases() {
    let fx = Fixture::new();
    fx.entry(
        "demo-leader",
        "demo-leader",
        Some("sess-lead"),
        "leader",
        LIVE,
    );
    fx.entry("agent-ben", "agent-ben", Some("sess-old"), "agent", ENDED);
    assert_eq!(fx.lanes(), 0);
    fx.entry("agent-cara", "agent-cara", Some("sess-one"), "agent", LIVE);
    fx.entry(
        "demo-agent-cara",
        "demo-agent-cara",
        Some("sess-one"),
        "agent",
        ENDED,
    );
    assert_eq!(fx.lanes(), 1);
}

/// REQ-6: an entry with no host session id counts by its name, one lane per name; a session id
/// that spells another entry's name is a different key.
#[test]
fn ac2_an_entry_without_a_session_counts_by_name() {
    let fx = Fixture::new();
    fx.entry("agent-ben", "agent-ben", None, "agent", LIVE);
    fx.entry("agent-cara", "agent-cara", None, "agent", LIVE);
    assert_eq!(fx.lanes(), 2);
    // A session id equal to the name of a session-less entry is still another lane.
    fx.entry(
        "demo-agent-dana",
        "demo-agent-dana",
        Some("agent-ben"),
        "agent",
        LIVE,
    );
    assert_eq!(fx.lanes(), 3);
}

/// REQ-6: the multiple per lane is still 2, with no profile in the clone to change it.
#[test]
fn ac2_the_multiple_per_lane_is_still_two() {
    let fx = Fixture::new();
    assert_eq!(DEFAULT_PER_LANE, 2);
    assert_eq!(per_lane(&fx.repo), 2);
}
