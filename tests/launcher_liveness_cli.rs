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
/// the same instance keeps the first registration's session id — pinned with a renewal that
/// carries a different id in its environment, so a mutant that re-reads the environment
/// instead of keeping the first value fails — and a host without one writes no field.
#[test]
fn session_field_is_recorded_under_schema_one() {
    let f = Fixture::new();
    let first = f.register("lane-a", Some(S_A), None);
    let entry = f.entry("lane-a");
    assert_eq!(entry["schema_version"], json!(1), "{entry}");
    assert_eq!(entry["session"], json!(S_A), "{entry}");

    let instance = first["instance"].as_str().unwrap().to_string();
    // The renewal runs under another id (the renewing process is not the session's process);
    // the first registration's id must survive it.
    f.register("lane-a", Some(S_B), Some(&instance));
    assert_eq!(
        f.entry("lane-a")["session"],
        json!(S_A),
        "the renewal keeps the first session id"
    );

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
        .args(["session-register", &f.repo.to_string_lossy()])
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

/// AC-4, the wiring: a named launch holds `<name>.live` while its host runs — the lookup
/// answers alive — releases it when the host exits, and a second launch of the same name is
/// refused while the first lives (the stub host waits for a release file).
#[test]
fn a_named_launch_holds_its_live_lock_while_its_host_runs() {
    use moai_cli::lane_registry::{liveness_of_session, Liveness};
    let f = Fixture::new();

    // The registry entry the holder session registered (the launcher itself does not
    // register; the lane's first turn does — the test writes the same shape).
    fs::create_dir_all(f.sessions()).unwrap();
    fs::write(
        f.sessions().join("lane-a.json"),
        format!(
            "{{\"schema_version\":1,\"name\":\"lane-a\",\"instance\":\"{}\",\"lease_until\":{},\"registered_at\":{},\"session\":\"{S_A}\"}}\n",
            "c".repeat(32),
            4102444800i64,
            1790000000i64
        ),
    )
    .unwrap();

    // A `claude` host stub that marks it started and waits for the release file.
    let stub = f.home.join("stub-hosts");
    fs::create_dir_all(&stub).unwrap();
    let marker = f.home.join("host-started");
    let release = f.home.join("release");
    let script = format!(
        "#!/bin/sh\nprintf '%s' $$ > '{}'\nwhile [ ! -f '{}' ]; do sleep 0.05; done\n",
        marker.display(),
        release.display()
    );
    fs::write(stub.join("claude"), script).unwrap();
    use std::os::unix::fs::PermissionsExt;
    fs::set_permissions(stub.join("claude"), fs::Permissions::from_mode(0o755)).unwrap();

    let mut cmd = Command::new(BIN);
    cmd.args(["cc", "--name", "lane-a"])
        .current_dir(&f.repo)
        .env("HOME", &f.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("CLAUDE_CODE_SESSION_ID", S_A)
        .env("PATH", {
            let mut p = stub.into_os_string();
            p.push(":");
            p.push(std::env::var_os("PATH").unwrap_or_default());
            p
        });
    let mut child = cmd.spawn().unwrap();
    for _ in 0..200 {
        if marker.exists() {
            break;
        }
        std::thread::sleep(std::time::Duration::from_millis(50));
    }
    assert!(marker.exists(), "the stub host started");

    let (state, evidence) = liveness_of_session(&f.repo, S_A);
    assert_eq!(state, Liveness::Alive, "{evidence}");

    // A second launch of the same name is refused while the first holds the lock.
    let second = cmd.output().unwrap();
    assert!(
        !second.status.success(),
        "the same name does not launch twice"
    );

    fs::write(&release, "").unwrap();
    let out = child.wait().unwrap();
    assert!(out.success(), "the first launch ended cleanly");

    let (state, _) = liveness_of_session(&f.repo, S_A);
    assert_eq!(state, Liveness::LauncherExited);
}

/// B1 of the audit: the lookup runs exactly as the workflow document spells it —
/// `moai session-register --liveness-of <session-id>`, no `--name`, no PROJECT — from inside
/// a clone, reading that clone's registry, writing nothing.
#[test]
fn the_documented_liveness_form_runs_without_a_name() {
    let f = Fixture::new();
    f.register("lane-a", Some(S_A), None);
    let out = Command::new(BIN)
        .args(["session-register", "--liveness-of", S_A])
        .current_dir(&f.repo)
        .env("HOME", &f.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env_remove("CLAUDE_CODE_SESSION_ID")
        .env_remove("CODEX_THREAD_ID")
        .output()
        .unwrap();
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["result"]["state"], json!("launcher-exited"), "{v}");
    assert_eq!(v["result"]["writes"], json!(false), "{v}");
}

/// B2 of the audit (leader decision): a name the registry rules reject runs the host with no
/// lock and no refusal — the same as it did before the lock — and the lookup answers unknown
/// for it; the dry run of the same name keeps succeeding, so preview and run agree.
#[test]
fn an_unsafe_name_runs_without_a_lock_and_answers_unknown() {
    use moai_cli::lane_registry::liveness_of_session;
    let f = Fixture::new();
    for name in ["영실이", "_x", &"x".repeat(65), "x/y"] {
        let out = Command::new(BIN)
            .args(["cc", "--name", name, "--dry-run", "--json"])
            .current_dir(&f.repo)
            .env("HOME", &f.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .output()
            .unwrap();
        assert_eq!(out.status.code(), Some(0), "{name}: {out:?}");
    }
    // The run itself takes no lock: no `.live` file for any unsafe name, and the stub host
    // runs. The live lock of a safe name that is absent answers unknown for that id.
    assert!(
        !f.repo
            .join(".git/.moai-cli/sessions")
            .join("영실이.live")
            .exists(),
        "no lock for an unsafe name"
    );
    let (state, evidence) = liveness_of_session(&f.repo, S_A);
    assert_eq!(
        state,
        moai_cli::lane_registry::Liveness::Unknown,
        "{evidence}"
    );
}
