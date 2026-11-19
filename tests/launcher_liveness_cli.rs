//! The launcher liveness proof, BOARD 03 slice (a) (SPEC-20261120-launcher-liveness): the
//! registry entry names the host session id, the named `.live` lock is an exclusive File lock,
//! and the read-only lookup answers alive, launcher-exited or unknown — without writing a
//! single file. Synthetic git repositories in a temporary folder with a temporary HOME; the
//! real `~/.moai` is not read or written.
#![cfg(unix)]
mod common;

use common::git::git;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
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

    /// `moai session-register <repo> --name <name>` as the host session `session`.
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
        let v: Value = serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|e| panic!("stdout is not JSON ({e})"));
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }

    fn sessions(&self) -> PathBuf {
        self.repo.join(".git/.moai-cli/sessions")
    }

    fn entry(&self, name: &str) -> Value {
        let text = fs::read_to_string(self.sessions().join(format!("{name}.json"))).unwrap();
        serde_json::from_str(&text).unwrap()
    }
}

/// Every file under the sessions folder with its bytes and mtime, for the reads-nothing check.
fn snapshot(dir: &Path) -> Vec<(String, u128, Vec<u8>)> {
    let mut out = Vec::new();
    if !dir.is_dir() {
        return out;
    }
    for e in fs::read_dir(dir).unwrap().flatten() {
        let p = e.path();
        let Ok(meta) = fs::symlink_metadata(&p) else {
            continue;
        };
        let mtime = meta
            .modified()
            .ok()
            .and_then(|t| t.duration_since(std::time::UNIX_EPOCH).ok())
            .map(|d| d.as_nanos())
            .unwrap_or(0);
        let bytes = if meta.file_type().is_file() {
            fs::read(&p).unwrap()
        } else {
            Vec::new()
        };
        out.push((
            p.file_name().unwrap().to_string_lossy().into_owned(),
            mtime,
            bytes,
        ));
    }
    out.sort();
    out
}

/// AC-1 (REQ-1): the entry carries the host session id under schema version 1, a renewal by
/// the same instance keeps it, and a host without one writes no field.
#[test]
fn session_field_is_recorded_under_schema_one() {
    let f = Fixture::new();
    let first = f.register("lane-a", Some(S_A), None);
    let entry = f.entry("lane-a");
    assert_eq!(entry["schema_version"], json!(1), "{entry}");
    assert_eq!(entry["session"], json!(S_A), "{entry}");

    let instance = first["instance"].as_str().unwrap().to_string();
    f.register("lane-a", Some(S_A), Some(&instance));
    assert_eq!(f.entry("lane-a")["session"], json!(S_A));

    f.register("lane-b", None, None);
    let entry = f.entry("lane-b");
    assert!(
        entry.get("session").is_none() || entry["session"].is_null(),
        "{entry}"
    );
    assert_eq!(entry["schema_version"], json!(1), "{entry}");
}

/// AC-2 (REQ-2): the three answers, each with its cause — alive (the lock is held),
/// launcher-exited (the lock is free or missing), unknown (no entry names the id, the name
/// moved to another session id, or the lock cannot be probed).
#[test]
fn three_answers_of_the_liveness_lookup() {
    use moai_cli::lane_registry::{liveness_of_session, Liveness};
    let f = Fixture::new();
    f.register("lane-a", Some(S_A), None);

    // Alive: the launcher's exclusive lock is held (the launcher wiring itself is the next
    // slice; the test holds the lock the way the launcher will).
    let _held = moai_cli::session::lock_named_live(&f.sessions(), "lane-a").unwrap();
    let (state, evidence) = liveness_of_session(&f.repo, S_A);
    assert_eq!(state, Liveness::Alive, "{evidence}");

    // Launcher-exited: the lock is free.
    drop(_held);
    let (state, _) = liveness_of_session(&f.repo, S_A);
    assert_eq!(state, Liveness::LauncherExited);

    // Unknown: no entry names this id.
    let (state, _) = liveness_of_session(&f.repo, S_B);
    assert_eq!(state, Liveness::Unknown);

    // Unknown: the name moved to another session id (the old id has no entry any more).
    fs::write(
        f.sessions().join("lane-a.json"),
        format!(
            "{{\"schema_version\":1,\"name\":\"lane-a\",\"instance\":\"{}\",\"lease_until\":{},\"registered_at\":{},\"session\":\"{S_B}\"}}\n",
            "a".repeat(32),
            4102444800i64,
            1790000000i64
        ),
    )
    .unwrap();
    let (state, _) = liveness_of_session(&f.repo, S_A);
    assert_eq!(state, Liveness::Unknown, "the moved-away id is unknown");

    // Unknown: the lock cannot be probed (a directory where the lock file would be).
    fs::write(
        f.sessions().join("lane-b.json"),
        format!(
            "{{\"schema_version\":1,\"name\":\"lane-b\",\"instance\":\"{}\",\"lease_until\":{},\"registered_at\":{},\"session\":\"{S_B}\"}}\n",
            "b".repeat(32),
            4102444801i64,
            1790000000i64
        ),
    )
    .unwrap();
    fs::create_dir_all(f.sessions().join("lane-b.live")).unwrap();
    let (state, evidence) = liveness_of_session(&f.repo, S_B);
    assert_eq!(state, Liveness::Unknown, "{evidence}");
}

/// AC-3 (REQ-3): the lookup changes no file — the folder's names, bytes and mtimes are the
/// same before and after, and a missing `.live` file is not created.
#[test]
fn the_lookup_reads_nothing_and_creates_no_lock_file() {
    use moai_cli::lane_registry::liveness_of_session;
    let f = Fixture::new();
    f.register("lane-a", Some(S_A), None);
    assert!(!f.sessions().join("lane-a.live").exists(), "no lock yet");

    let before = snapshot(&f.sessions());
    let _ = liveness_of_session(&f.repo, S_A);
    let _ = liveness_of_session(&f.repo, S_B);
    let after = snapshot(&f.sessions());
    assert_eq!(before, after, "the lookup wrote nothing");
    assert!(
        !f.sessions().join("lane-a.live").exists(),
        "still no lock file"
    );

    // The read-only CLI flag answers through the same lookup and writes nothing either.
    let out = Command::new(BIN)
        .args(["session-register", &f.repo.to_string_lossy(), "--name", "x"])
        .arg("--liveness-of")
        .arg(S_A)
        .env("HOME", &f.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["state"], json!("launcher-exited"), "{v}");
    assert_eq!(v["result"]["writes"], json!(false), "{v}");
    assert_eq!(snapshot(&f.sessions()), after, "the flag wrote nothing");
}

/// AC-4 (REQ-4): the named live lock is exclusive, refuses a symlink and an unsafe name.
#[test]
fn the_named_live_lock_is_exclusive_and_refuses_bad_paths() {
    let f = Fixture::new();
    // The sessions folder exists once anything has registered; the launcher runs after that.
    fs::create_dir_all(f.sessions()).unwrap();
    let held = moai_cli::session::lock_named_live(&f.sessions(), "lane-a").unwrap();
    // A second open of the same file is a second lock holder and must not get the lock
    // (flock is per open file description, so this also holds within one process).
    let second = std::fs::OpenOptions::new()
        .read(true)
        .write(true)
        .open(f.sessions().join("lane-a.live"))
        .unwrap();
    assert!(
        second.try_lock().is_err(),
        "the same file does not lock twice"
    );

    assert!(moai_cli::session::lock_named_live(&f.sessions(), "../escape").is_err());
    assert!(moai_cli::session::lock_named_live(&f.sessions(), "lane-b").is_ok());

    // A symlink at the lock path is refused.
    drop(held);
    let real = f.sessions().join("real-target");
    fs::write(&real, b"").unwrap();
    #[cfg(unix)]
    std::os::unix::fs::symlink(&real, f.sessions().join("lane-c.live")).unwrap();
    assert!(moai_cli::session::lock_named_live(&f.sessions(), "lane-c").is_err());
}
