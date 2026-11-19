//! `moai session-register` liveness anchors, and (slice 2b, later) the `card reclaim`
//! liveness gate (card C-135, SPEC-20261120-reclaim-liveness): the registry entry records the
//! host session id and the registering process's id and start time, a renewal by the same
//! instance keeps the first anchor, and the registry answers the anchor of a session id, so a
//! reclaim can tell an ended lease whose holder still runs from one whose holder is gone.
//!
//! Synthetic git repositories in a temporary folder with a temporary HOME; the real `~/.moai`
//! is not read or written.
#![cfg(unix)]
mod common;

use common::git::git;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Command;

const BIN: &str = crate::common::MOAI_BIN;
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

struct Fixture {
    _dir: tempfile::TempDir,
    repo: PathBuf,
    home: PathBuf,
}

impl Fixture {
    fn new() -> Self {
        let dir = tempfile::tempdir().unwrap();
        let base = fs::canonicalize(dir.path()).unwrap();
        let repo = base.join("repo");
        let home = base.join("home");
        fs::create_dir_all(&repo).unwrap();
        fs::create_dir_all(&home).unwrap();
        git(&repo, &["init", "-q", "-b", "main"]);
        Fixture {
            _dir: dir,
            repo,
            home,
        }
    }

    /// `moai session-register <repo> --name <name>` as the session `session` (a host id), with
    /// `--instance` for a renewal. The answer's `result`.
    fn register(&self, name: &str, session: Option<&str>, instance: Option<&str>) -> Value {
        let mut cmd = Command::new(BIN);
        cmd.arg("session-register")
            .arg(&self.repo)
            .args(["--name", name]);
        if let Some(i) = instance {
            cmd.args(["--instance", i]);
        }
        let out = cmd
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            // The test process may itself run under a host session id; only the case under
            // test may pass one down.
            .env_remove("CLAUDE_CODE_SESSION_ID")
            .env_remove("CODEX_THREAD_ID")
            .envs(session.map(|s| ("CLAUDE_CODE_SESSION_ID", s)))
            .output()
            .unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
            panic!(
                "stdout is not JSON ({e}): {}",
                String::from_utf8_lossy(&out.stdout)
            )
        });
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }

    /// The registry entry file of `name`.
    fn entry(&self, name: &str) -> Value {
        let path = self
            .repo
            .join(".git/.moai-cli/sessions")
            .join(format!("{name}.json"));
        let text = fs::read_to_string(&path).unwrap_or_else(|e| panic!("entry of {name}: {e}"));
        serde_json::from_str(&text).unwrap()
    }
}

/// AC-1 (REQ-1, REQ-2): the entry records the session id, the registering process id and its
/// start time, and a renewal by the same instance keeps the first anchor — the renewing
/// process is not the session's process.
#[test]
fn registry_anchor_is_recorded_and_kept_over_a_renewal() {
    let f = Fixture::new();
    let first = f.register("lane-a", Some(S_A), None);
    let entry = f.entry("lane-a");
    assert_eq!(entry["schema_version"], json!(2), "{entry}");
    assert_eq!(entry["session"], json!(S_A), "{entry}");
    let pid = entry["pid"].as_u64().expect("pid is a number");
    assert!(pid > 0, "{entry}");
    let start = entry["process_start"].as_str().expect("start time text");
    assert!(!start.is_empty(), "{entry}");

    // A renewal by the same instance runs in another process; the anchor stays the first one's.
    let instance = first["instance"].as_str().unwrap().to_string();
    f.register("lane-a", Some(S_A), Some(&instance));
    let renewed = f.entry("lane-a");
    assert_eq!(renewed["session"], json!(S_A), "{renewed}");
    assert_eq!(
        renewed["pid"],
        json!(pid),
        "the first anchor is kept: {renewed}"
    );
    assert_eq!(renewed["process_start"], json!(start), "{renewed}");
}

/// AC-1, REQ-1: with no host session id the entry carries no session and no anchor look-up
/// will match it; the registration itself still succeeds.
#[test]
fn registry_entry_without_a_session_id_has_no_anchor_fields_for_lookup() {
    let f = Fixture::new();
    f.register("lane-b", None, None);
    let entry = f.entry("lane-b");
    assert!(
        entry.get("session").is_none() || entry["session"].is_null(),
        "{entry}"
    );
    assert!(moai_cli::lane_registry::anchor_of_session(&f.repo, S_A).is_none());
}

/// AC-2 (REQ-3): the registry answers the anchor of the entry that names a session id, and
/// none for another id or an old schema-1 entry without a session.
#[test]
fn anchor_lookup_answers_the_entry_that_names_the_session() {
    let f = Fixture::new();
    f.register("lane-c", Some(S_A), None);
    let entry = f.entry("lane-c");
    let pid = entry["pid"].as_u64().unwrap() as u32;
    let start = entry["process_start"].as_str().map(str::to_string);

    let anchor = moai_cli::lane_registry::anchor_of_session(&f.repo, S_A)
        .expect("the entry names this session");
    assert_eq!(anchor.pid, Some(pid));
    assert_eq!(anchor.process_start, start);

    assert!(moai_cli::lane_registry::anchor_of_session(&f.repo, S_B).is_none());

    // An old schema-1 entry names no session: it is not an anchor for any id.
    let dir = f.repo.join(".git/.moai-cli/sessions");
    fs::write(
        dir.join("lane-old.json"),
        format!(
            "{{\"schema_version\":1,\"name\":\"lane-old\",\"instance\":\"{}\",\"lease_until\":{},\"registered_at\":{}}}\n",
            "f".repeat(32),
            4102444800i64,
            1790000000i64
        ),
    )
    .unwrap();
    assert!(moai_cli::lane_registry::anchor_of_session(&f.repo, S_B).is_none());
    let anchor = moai_cli::lane_registry::anchor_of_session(&f.repo, S_A)
        .expect("the scan skips the old entry and still finds lane-c");
    assert_eq!(anchor.pid, Some(pid));
}
