//! The coordination window's registration (SPEC-20261123-leader-registration-lease, AC-1 to
//! AC-4): a leader entry carries the host session id, a second live leader is refused, a
//! renewal comes from the same session, and `--show` and `--liveness-of` let a lane read the
//! live window's session id. Synthetic git repositories in a temporary folder with a temporary
//! HOME; the real state folder of this machine is not read or written.
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
const T0: &str = "2026-11-23T00:00:00Z";
/// Five hours after `T0`: past the four hour default lease.
const T5H: &str = "2026-11-23T05:00:00Z";

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

    /// `moai session-register <args>` as the host session `session`; the whole envelope.
    fn run(&self, session: Option<&str>, args: &[&str]) -> Value {
        let mut cmd = Command::new(BIN);
        cmd.arg("session-register").arg(&self.repo).args(args);
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
        serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
            panic!(
                "stdout is not JSON ({e}): {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    /// A leader registration of `name` as `session` at `now`, with `--instance` when given.
    fn leader(
        &self,
        session: Option<&str>,
        name: &str,
        instance: Option<&str>,
        now: &str,
    ) -> Value {
        let mut args = vec!["--name", name, "--mode", "leader", "--now", now];
        if let Some(i) = instance {
            args.extend(["--instance", i]);
        }
        self.run(session, &args)
    }

    fn sessions(&self) -> PathBuf {
        moai_cli::clone_root::path(&self.repo.join(".git")).join("sessions")
    }

    fn entry_bytes(&self, name: &str) -> Vec<u8> {
        fs::read(self.sessions().join(format!("{name}.json"))).unwrap()
    }

    fn entry(&self, name: &str) -> Value {
        serde_json::from_slice(&self.entry_bytes(name)).unwrap()
    }
}

/// Every file under `dir` with its bytes, sorted; an empty list when the folder is missing.
fn snapshot(dir: &Path) -> Vec<(String, Vec<u8>)> {
    let mut out = Vec::new();
    if let Ok(entries) = fs::read_dir(dir) {
        for e in entries.flatten() {
            let p = e.path();
            if p.is_file() {
                out.push((
                    p.file_name().unwrap().to_string_lossy().into_owned(),
                    fs::read(&p).unwrap(),
                ));
            }
        }
    }
    out.sort();
    out
}

fn message(v: &Value) -> String {
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

/// AC-1 (REQ-1): the leader entry carries the running session's id and its mode.
#[test]
fn ac1_a_leader_registration_writes_the_session_id() {
    let f = Fixture::new();
    let v = f.leader(Some(S_A), "leader", None, T0);
    assert_eq!(v["ok"], json!(true), "{v}");
    let entry = f.entry("leader");
    assert_eq!(entry["session"], json!(S_A), "{entry}");
    assert_eq!(entry["mode"], json!("leader"), "{entry}");
    assert_eq!(entry["schema_version"], json!(1), "{entry}");
}

/// AC-1 (REQ-2): without a host session id a leader registration is refused naming the id, and
/// the registry folder is exactly as it was; an `agent` registration needs none.
#[test]
fn ac1_a_leader_registration_without_a_session_id_is_refused() {
    let f = Fixture::new();
    let before = snapshot(&f.sessions());
    let v = f.leader(None, "leader", None, T0);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(message(&v).contains("session id"), "{v}");
    assert_eq!(snapshot(&f.sessions()), before);
    assert!(!f.sessions().join("leader.json").exists());
    let agent = f.run(None, &["--name", "agent-4", "--mode", "agent", "--now", T0]);
    assert_eq!(agent["ok"], json!(true), "{agent}");
}

/// AC-2 (REQ-3): a live leader entry refuses a second leader under another name and one under
/// the same name by another instance, naming the entry and its lease end but never its
/// instance; an `agent` registration is not refused; after the lease ended the name is free.
#[test]
fn ac2_a_second_live_leader_is_refused() {
    let f = Fixture::new();
    let first = f.leader(Some(S_A), "leader", None, T0);
    assert_eq!(first["ok"], json!(true), "{first}");
    let instance = first["result"]["instance"].as_str().unwrap().to_string();
    let before = snapshot(&f.sessions());

    let other_name = f.leader(Some(S_B), "boss", None, T0);
    assert_eq!(other_name["ok"], json!(false), "{other_name}");
    let text = message(&other_name);
    assert!(text.contains("leader"), "{text}");
    assert!(text.contains("2026-11-23T04:00:00Z"), "{text}");
    assert!(!text.contains(&instance), "{text}");
    assert!(!f.sessions().join("boss.json").exists());

    let same_name = f.leader(Some(S_B), "leader", None, T0);
    assert_eq!(same_name["ok"], json!(false), "{same_name}");
    assert!(!message(&same_name).contains(&instance), "{same_name}");
    assert_eq!(
        snapshot(&f.sessions()),
        before,
        "the refusals write nothing"
    );

    let agent = f.run(
        Some(S_B),
        &["--name", "agent-4", "--mode", "agent", "--now", T0],
    );
    assert_eq!(agent["ok"], json!(true), "{agent}");

    let later = f.leader(Some(S_B), "boss", None, T5H);
    assert_eq!(later["ok"], json!(true), "{later}");
    assert_eq!(f.entry("boss")["session"], json!(S_B));
}

/// AC-3 (REQ-4): the holding instance renews from its own session and keeps `session`; a
/// renewal from another session is refused and the entry is byte for byte as it was.
#[test]
fn ac3_a_renewal_comes_from_the_same_session() {
    let f = Fixture::new();
    let first = f.leader(Some(S_A), "leader", None, T0);
    let instance = first["result"]["instance"].as_str().unwrap().to_string();

    let one_hour = "2026-11-23T01:00:00Z";
    let renewed = f.leader(Some(S_A), "leader", Some(&instance), one_hour);
    assert_eq!(renewed["ok"], json!(true), "{renewed}");
    assert_eq!(renewed["result"]["status"], json!("renewed"), "{renewed}");
    let entry = f.entry("leader");
    assert_eq!(entry["session"], json!(S_A), "{entry}");
    let at = moai_cli::forge::policy::parse_utc(one_hour).unwrap();
    assert_eq!(entry["lease_until"], json!(at + 14_400), "{entry}");

    let before = f.entry_bytes("leader");
    let foreign = f.leader(Some(S_B), "leader", Some(&instance), "2026-11-23T02:00:00Z");
    assert_eq!(foreign["ok"], json!(false), "{foreign}");
    assert!(message(&foreign).contains("session"), "{foreign}");
    assert_eq!(f.entry_bytes("leader"), before);
    let none = f.leader(None, "leader", Some(&instance), "2026-11-23T02:00:00Z");
    assert_eq!(none["ok"], json!(false), "{none}");
    assert_eq!(f.entry_bytes("leader"), before);
}

/// AC-3 (REQ-4): an entry from before this rule has no session id; the renewal by its own
/// instance writes the running session's id into it.
#[test]
fn ac3_a_renewal_fills_in_a_missing_session_id() {
    let f = Fixture::new();
    let instance = "0123456789abcdef0123456789abcdef";
    fs::create_dir_all(f.sessions()).unwrap();
    let at = moai_cli::forge::policy::parse_utc(T0).unwrap();
    fs::write(
        f.sessions().join("leader.json"),
        json!({
            "schema_version": 1, "name": "leader", "instance": instance,
            "lease_until": at - 38 * 3600, "registered_at": at - 40 * 3600, "mode": "leader",
        })
        .to_string(),
    )
    .unwrap();
    let v = f.leader(Some(S_A), "leader", Some(instance), T0);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["previous"], json!("expired"), "{v}");
    assert_eq!(f.entry("leader")["session"], json!(S_A));
    assert_eq!(f.entry("leader")["instance"], json!(instance));
}

/// AC-4 (REQ-5, REQ-6): a lane reads the live window's session id with `--show`, and
/// `--liveness-of` answers `alive` from the lease while no `.live` file exists; both write
/// nothing; after the lease the id is no longer alive.
#[test]
fn ac4_a_lane_reads_the_live_leader_session() {
    let f = Fixture::new();
    let first = f.leader(Some(S_A), "leader", None, T0);
    assert_eq!(first["ok"], json!(true), "{first}");
    let before = snapshot(&f.sessions());

    let show = f.run(None, &["--show", "--name", "leader", "--now", T0]);
    assert_eq!(show["ok"], json!(true), "{show}");
    assert_eq!(show["result"]["status"], json!("held"), "{show}");
    assert_eq!(show["result"]["session"], json!(S_A), "{show}");
    assert_eq!(show["result"]["mode"], json!("leader"), "{show}");

    let alive = f.run(None, &["--liveness-of", S_A, "--now", T0]);
    assert_eq!(alive["ok"], json!(true), "{alive}");
    assert_eq!(alive["result"]["state"], json!("alive"), "{alive}");
    assert_eq!(alive["result"]["writes"], json!(false), "{alive}");
    assert!(!f.sessions().join("leader.live").exists());
    assert_eq!(snapshot(&f.sessions()), before, "the reads write nothing");

    let ended = f.run(None, &["--liveness-of", S_A, "--now", T5H]);
    assert_eq!(ended["ok"], json!(true), "{ended}");
    assert_ne!(ended["result"]["state"], json!("alive"), "{ended}");
    let gone = f.run(None, &["--show", "--name", "leader", "--now", T5H]);
    assert_eq!(gone["result"]["status"], json!("free"), "{gone}");
}

/// AC-4 (REQ-6): a lane entry (mode `agent` or none) is not made alive by its lease; only a
/// leader entry is read that way.
#[test]
fn ac4_only_a_leader_entry_is_alive_by_its_lease() {
    let f = Fixture::new();
    let lane = f.run(Some(S_B), &["--name", "agent-4", "--now", T0]);
    assert_eq!(lane["ok"], json!(true), "{lane}");
    let v = f.run(None, &["--liveness-of", S_B, "--now", T0]);
    assert_ne!(v["result"]["state"], json!("alive"), "{v}");
}

/// The registered instance of a `registered` or `renewed` answer.
fn instance_of(v: &Value) -> String {
    v["result"]["instance"].as_str().unwrap().to_string()
}

/// AC-2 (REQ-3), the second repair: a renewal by the instance of a lane entry that asks for
/// mode `leader` is a leader registration like any other, so a live leader entry refuses it —
/// for an entry of mode `agent` and for one that has no mode and no session — and the entry
/// is byte for byte as it was.
#[test]
fn ac2_a_renewal_into_mode_leader_is_refused_while_another_leader_lives() {
    let f = Fixture::new();
    let first = f.leader(Some(S_A), "leader", None, T0);
    assert_eq!(first["ok"], json!(true), "{first}");

    let lane = f.run(
        Some(S_B),
        &["--name", "agent-4", "--mode", "agent", "--now", T0],
    );
    let bare = f.run(None, &["--name", "agent-5", "--now", T0]);
    for (name, v) in [("agent-4", &lane), ("agent-5", &bare)] {
        assert_eq!(v["ok"], json!(true), "{v}");
        let before = f.entry_bytes(name);
        let before_leader = f.entry_bytes("leader");
        let renewed = f.leader(
            Some(S_B),
            name,
            Some(&instance_of(v)),
            "2026-11-23T00:30:00Z",
        );
        assert_eq!(renewed["ok"], json!(false), "{name}: {renewed}");
        assert!(message(&renewed).contains("conflict"), "{name}: {renewed}");
        assert_eq!(f.entry_bytes(name), before, "{name} is as it was");
        assert_eq!(f.entry_bytes("leader"), before_leader);
    }

    // The same renewal is accepted once the first leader's lease has ended.
    let later = f.leader(Some(S_B), "agent-4", Some(&instance_of(&lane)), T5H);
    assert_eq!(later["ok"], json!(true), "{later}");
    assert_eq!(f.entry("agent-4")["mode"], json!("leader"));
}

/// AC-3 (REQ-4), the second repair: a leader entry held under a lease answers only to its own
/// session whatever `--mode` the caller gives, so another session that read the instance
/// from `--show` cannot renew it with no mode or with mode `agent`; the entry keeps its mode
/// and its session. Its own session renews it with no mode and it stays a leader entry.
#[test]
fn ac3_a_leader_entry_answers_only_to_its_own_session_whatever_the_mode() {
    let f = Fixture::new();
    let first = f.leader(Some(S_A), "leader", None, T0);
    let instance = instance_of(&first);
    let at = "2026-11-23T01:00:00Z";
    let before = f.entry_bytes("leader");

    for extra in [&[][..], &["--mode", "agent"][..]] {
        let mut args = vec!["--name", "leader", "--instance", &instance, "--now", at];
        args.extend(extra);
        let foreign = f.run(Some(S_B), &args);
        assert_eq!(foreign["ok"], json!(false), "{extra:?}: {foreign}");
        assert!(message(&foreign).contains("session"), "{foreign}");
        assert_eq!(f.entry_bytes("leader"), before, "{extra:?}");
        let none = f.run(None, &args);
        assert_eq!(none["ok"], json!(false), "{extra:?}: {none}");
        assert_eq!(f.entry_bytes("leader"), before, "{extra:?}");
    }

    let own = f.run(
        Some(S_A),
        &["--name", "leader", "--instance", &instance, "--now", at],
    );
    assert_eq!(own["ok"], json!(true), "{own}");
    assert_eq!(own["result"]["status"], json!("renewed"), "{own}");
    let entry = f.entry("leader");
    assert_eq!(entry["mode"], json!("leader"), "{entry}");
    assert_eq!(entry["session"], json!(S_A), "{entry}");
    let shown = f.run(None, &["--show", "--name", "leader", "--now", at]);
    assert_eq!(shown["result"]["mode"], json!("leader"), "{shown}");
}

/// AC-4 (REQ-6): a `.live` file is the launcher's word and the lease never overrides it: with
/// the file present a leader entry whose lease runs is not `alive` by its lease.
#[test]
fn ac4_a_launcher_lock_file_is_never_overridden_by_the_lease() {
    let f = Fixture::new();
    let first = f.leader(Some(S_A), "leader", None, T0);
    assert_eq!(first["ok"], json!(true), "{first}");
    fs::write(f.sessions().join("leader.live"), b"").unwrap();
    let v = f.run(None, &["--liveness-of", S_A, "--now", T0]);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_ne!(v["result"]["state"], json!("alive"), "{v}");
    assert_ne!(
        v["result"]["evidence"]["basis"],
        json!("registry_lease"),
        "{v}"
    );
}
