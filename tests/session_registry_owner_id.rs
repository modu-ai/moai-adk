//! SPEC-20261125-session-registry-session-key, the limit side (REQ-6, AC-2): the work-in-progress
//! limit counts each live registered session once, whatever the form of its entry file, and the
//! multiple per lane stays 2. The calls run in the library on a temporary clone with hand-written
//! entries; a temporary HOME stands beside them as the only home a run of the binary would see,
//! and no real registry of this machine is read or written.
//!
//! The `ac1_*` tests below are the entry side (AC-1): the entry file of a session with a host
//! session id is named by that id, a name is held by one live session at a time, and every
//! lookup reads an entry of either file form.
mod common;

use moai_cli::card::wip_limit::{live_lanes, per_lane, DEFAULT_PER_LANE};
use moai_cli::lane_registry::{
    launcher_trace_of_session, live_session_count, liveness_of_session, mode_of, session_entries,
    session_live, Liveness,
};
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::{Command, Output};

/// A lease that has not ended, whatever the clock says.
const LIVE: i64 = 4_102_444_800;
/// A lease that ended long ago.
const ENDED: i64 = 1;

struct Fixture {
    _dir: tempfile::TempDir,
    home: common::home::TempHome,
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
            home: common::home::TempHome::new(),
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
        live_lanes(&self.repo).live
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

/// REQ-6 with the lock lanes (SPEC-20261125-wip-limit-live-lock-lanes): an entry file named by its
/// session id is read by the lane scan, not left out as unreadable.
#[test]
fn ac2_a_session_keyed_entry_is_counted_not_unreadable() {
    let fx = Fixture::new();
    fx.entry(
        "sess-one",
        "demo-agent-ben",
        Some("sess-one"),
        "agent",
        LIVE,
    );
    let lanes = live_lanes(&fx.repo);
    assert_eq!(lanes.live, 1, "{lanes:?}");
    assert!(lanes.unreadable.is_empty(), "{lanes:?}");
}

/// REQ-6 with the lock lanes: a held launcher lock and the session-keyed entry of the same lane
/// are one lane, whichever of the two name forms the lock carries and whether the lease ended.
#[test]
fn ac2_a_held_lock_and_a_session_keyed_entry_are_one_lane() {
    let fx = Fixture::new();
    // The lock carries the short form, the entry the project-prefixed form.
    fx.entry(
        "sess-one",
        "demo-agent-ben",
        Some("sess-one"),
        "agent",
        LIVE,
    );
    let _short = moai_cli::session::lock_named_live(&fx.sessions, "agent-ben").unwrap();
    assert_eq!(fx.lanes(), 1);
    // Its lease ends while the launcher still holds the lock: still one lane.
    fx.entry(
        "sess-one",
        "demo-agent-ben",
        Some("sess-one"),
        "agent",
        ENDED,
    );
    assert_eq!(fx.lanes(), 1);
    // An older entry of the same lane name under another session does not make a second lane.
    fx.entry(
        "demo-agent-ben",
        "demo-agent-ben",
        Some("sess-old"),
        "agent",
        ENDED,
    );
    assert_eq!(fx.lanes(), 1);
    // A different lane with its own lock is a second lane.
    fx.entry(
        "sess-two",
        "demo-agent-cara",
        Some("sess-two"),
        "agent",
        ENDED,
    );
    let _cara = moai_cli::session::lock_named_live(&fx.sessions, "demo-agent-cara").unwrap();
    assert_eq!(fx.lanes(), 2);
}

/// REQ-6 with the lock lanes: a lock with no entry counts once whichever name form it carries,
/// and not again beside the entry of the same lane.
#[test]
fn ac2_a_lock_alone_counts_once_for_either_name_form() {
    let fx = Fixture::new();
    let _short = moai_cli::session::lock_named_live(&fx.sessions, "agent-dana").unwrap();
    let _long = moai_cli::session::lock_named_live(&fx.sessions, "demo-agent-dana").unwrap();
    assert_eq!(fx.lanes(), 1);
}

/// REQ-6: the multiple per lane is still 2, with no profile in the clone to change it.
#[test]
fn ac2_the_multiple_per_lane_is_still_two() {
    let fx = Fixture::new();
    assert_eq!(DEFAULT_PER_LANE, 2);
    assert_eq!(per_lane(&fx.repo), 2);
}

// ---- AC-1: the entry file is keyed by the host session id (REQ-1 to REQ-5) ----------------------
//
// Every write goes through the binary, built by `registrar`: a temporary HOME, no git
// configuration of the machine, and the host session variables set per case, never inherited.
// The read-only lookups also run in the library on the same temporary clone.

/// `--now` of every call through the binary.
const T0: &str = "2026-11-13T10:00:00Z";
/// A clock for the library calls: after `ENDED`, before `LIVE`.
const NOW: i64 = 1_000_000;

/// A well-formed host session id for number `n`.
fn sid(n: u64) -> String {
    format!("{n:08x}-0000-4000-8000-{n:012x}")
}

/// The entry file name of the host session `n`.
fn sfile(n: u64) -> String {
    format!("{}.json", sid(n))
}

/// A 32-digit instance of one hex digit.
fn inst(c: char) -> String {
    c.to_string().repeat(32)
}

/// The one place that builds a command for the binary under test.
fn registrar(fx: &Fixture) -> Command {
    let mut c = Command::new(common::MOAI_BIN);
    c.env("HOME", fx.home.path())
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env_remove("CLAUDE_CODE_SESSION_ID")
        .env_remove("CODEX_THREAD_ID");
    c
}

/// `session-register <repo> <args> --now T0` run in the clone as the host session `session`
/// (none: the host gave no id).
fn reg(fx: &Fixture, session: Option<&str>, args: &[&str]) -> Output {
    let repo = fx.repo.to_str().unwrap();
    let mut cmd = registrar(fx);
    cmd.arg("session-register")
        .arg(repo)
        .args(args)
        .args(["--now", T0])
        .current_dir(&fx.repo);
    if let Some(s) = session {
        cmd.env("CLAUDE_CODE_SESSION_ID", s);
    }
    cmd.output().unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope on stdout; stderr {}",
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

/// The error kind and message of a call that must exit 1.
fn refused(out: &Output) -> (String, String) {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    assert_eq!(v["ok"], false, "{v}");
    (
        v["error"]["kind"].as_str().unwrap_or_default().to_string(),
        v["error"]["message"]
            .as_str()
            .unwrap_or_default()
            .to_string(),
    )
}

impl Fixture {
    /// Like [`Fixture::entry`] with the instance given.
    fn entry_with(
        &self,
        file: &str,
        name: &str,
        session: Option<&str>,
        mode: &str,
        lease: i64,
        instance: &str,
    ) {
        let mut v: Value = json!({
            "schema_version": 1,
            "name": name,
            "instance": instance,
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

    /// The sorted names of the `*.json` files in the registry folder.
    fn files(&self) -> Vec<String> {
        let mut out: Vec<String> = fs::read_dir(&self.sessions)
            .unwrap()
            .flatten()
            .map(|e| e.file_name().to_string_lossy().to_string())
            .filter(|n| n.ends_with(".json"))
            .collect();
        out.sort();
        out
    }

    fn bytes(&self, file: &str) -> Vec<u8> {
        fs::read(self.sessions.join(file)).unwrap()
    }

    fn read(&self, file: &str) -> Value {
        serde_json::from_slice(&self.bytes(file)).unwrap()
    }
}

/// A sorted list of file names.
fn sorted(names: &[&str]) -> Vec<String> {
    let mut v: Vec<String> = names.iter().map(|s| s.to_string()).collect();
    v.sort();
    v
}

/// REQ-1: a session with a host session id registers into the file named by that id, with the
/// lane name kept inside the entry.
#[test]
fn ac1_a_session_with_an_id_gets_an_entry_file_named_by_that_id() {
    let fx = Fixture::new();
    let got = ok(&reg(&fx, Some(&sid(1)), &["--name", "agent-1"]));
    assert_eq!(got["status"], "registered", "{got}");
    assert_eq!(fx.files(), sorted(&[&sfile(1)]));
    let e = fx.read(&sfile(1));
    assert_eq!(e["name"], "agent-1");
    assert_eq!(e["session"], sid(1).as_str());
    assert!(
        got["entry"].as_str().unwrap().ends_with(&sfile(1)),
        "the answer names the file: {got}"
    );
}

/// REQ-1: with no usable id the entry is keyed by the name, as before: no id at all, an id that
/// is not well formed, and an id that starts with `-` (it is no file name).
#[test]
fn ac1_a_session_without_a_usable_id_keeps_a_name_keyed_file() {
    for (label, id) in [
        ("no id", None),
        ("not well formed", Some("not-a-host-id!".to_string())),
        ("starts with a dash", Some("-abcdef01-2345".to_string())),
    ] {
        let fx = Fixture::new();
        let got = ok(&reg(&fx, id.as_deref(), &["--name", "agent-1"]));
        assert_eq!(got["status"], "registered", "{label}: {got}");
        assert_eq!(fx.files(), sorted(&["agent-1.json"]), "{label}");
        let e = fx.read("agent-1.json");
        assert_eq!(e["name"], "agent-1", "{label}");
        // The `-` id is a host id, so the entry still records it; the others record none.
        match label {
            "starts with a dash" => assert_eq!(e["session"], "-abcdef01-2345", "{label}"),
            _ => assert!(e.get("session").is_none(), "{label}: {e}"),
        }
    }
    // A registration that has no id does not touch the entries of sessions that have none.
    let fx = Fixture::new();
    fx.entry_with("agent-8", "agent-8", None, "agent", LIVE, &inst('8'));
    ok(&reg(&fx, None, &["--name", "agent-3"]));
    assert_eq!(fx.files(), sorted(&["agent-3.json", "agent-8.json"]));
}

/// REQ-2: a name another live session holds is refused with the conflict text, and both entries
/// stay as they were.
#[test]
fn ac1_a_name_held_by_another_live_session_is_refused_and_both_entries_stay() {
    let fx = Fixture::new();
    ok(&reg(&fx, Some(&sid(1)), &["--name", "agent-1"]));
    assert_eq!(fx.files(), sorted(&[&sfile(1)]));
    let first = fx.bytes(&sfile(1));
    let (kind, message) = refused(&reg(&fx, Some(&sid(2)), &["--name", "agent-1"]));
    assert_eq!(kind, "conflict", "{message}");
    assert!(
        message.contains(
            "conflict in session registry: name `agent-1` is in use by another instance until"
        ),
        "{message}"
    );
    assert_eq!(fx.files(), sorted(&[&sfile(1)]));
    assert_eq!(fx.bytes(&sfile(1)), first);

    // The second session has its own entry under another name; the refusal leaves both alone.
    ok(&reg(&fx, Some(&sid(2)), &["--name", "agent-2"]));
    assert_eq!(fx.files(), sorted(&[&sfile(1), &sfile(2)]));
    let second = fx.bytes(&sfile(2));
    assert_eq!(
        refused(&reg(&fx, Some(&sid(2)), &["--name", "agent-1"])).0,
        "conflict"
    );
    assert_eq!(fx.files(), sorted(&[&sfile(1), &sfile(2)]));
    assert_eq!(fx.bytes(&sfile(1)), first);
    assert_eq!(fx.bytes(&sfile(2)), second);
}

/// REQ-2, REQ-4: a name held in the old, name-keyed form is as taken as one held in the new form.
#[test]
fn ac1_a_name_held_in_the_name_keyed_form_is_refused_to_a_session_with_an_id() {
    let fx = Fixture::new();
    fx.entry_with("agent-1", "agent-1", None, "agent", LIVE, &inst('a'));
    let before = fx.bytes("agent-1.json");
    let (kind, message) = refused(&reg(&fx, Some(&sid(1)), &["--name", "agent-1"]));
    assert_eq!(kind, "conflict", "{message}");
    assert_eq!(fx.files(), sorted(&["agent-1.json"]));
    assert_eq!(fx.bytes("agent-1.json"), before);
}

/// REQ-1, REQ-2: an entry of the name whose lease has ended is replaced by the new registration,
/// whatever file form it was in, and no second file for the name is left.
#[test]
fn ac1_an_ended_entry_of_the_name_is_replaced_whatever_its_form() {
    // The ended entry is name keyed.
    let fx = Fixture::new();
    fx.entry_with("agent-1", "agent-1", None, "agent", ENDED, &inst('a'));
    let got = ok(&reg(&fx, Some(&sid(1)), &["--name", "agent-1"]));
    assert_eq!(got["previous"], "expired", "{got}");
    assert_eq!(fx.files(), sorted(&[&sfile(1)]), "name-keyed ended entry");

    // The ended entry is keyed by another session's id.
    let fx = Fixture::new();
    fx.entry_with(
        &sid(2),
        "agent-1",
        Some(&sid(2)),
        "agent",
        ENDED,
        &inst('a'),
    );
    let got = ok(&reg(&fx, Some(&sid(1)), &["--name", "agent-1"]));
    assert_eq!(got["previous"], "expired", "{got}");
    assert_eq!(
        fx.files(),
        sorted(&[&sfile(1)]),
        "session-keyed ended entry"
    );

    // A registration with no usable id takes the name over into a name-keyed file.
    let fx = Fixture::new();
    fx.entry_with(
        &sid(2),
        "agent-1",
        Some(&sid(2)),
        "agent",
        ENDED,
        &inst('a'),
    );
    let got = ok(&reg(&fx, None, &["--name", "agent-1"]));
    assert_eq!(got["previous"], "expired", "{got}");
    assert_eq!(fx.files(), sorted(&["agent-1.json"]), "no id");
}

/// REQ-3: a session that registers again under another name leaves one entry; the entry of
/// another session is not touched.
#[test]
fn ac1_a_rename_leaves_one_entry_for_the_session() {
    let fx = Fixture::new();
    ok(&reg(&fx, Some(&sid(1)), &["--name", "agent-1"]));
    ok(&reg(&fx, Some(&sid(2)), &["--name", "agent-2"]));
    assert_eq!(fx.files(), sorted(&[&sfile(1), &sfile(2)]));
    let other = fx.bytes(&sfile(2));
    let got = ok(&reg(&fx, Some(&sid(1)), &["--name", "agent-5"]));
    assert_eq!(got["status"], "registered", "{got}");
    assert_eq!(fx.files(), sorted(&[&sfile(1), &sfile(2)]));
    assert_eq!(fx.read(&sfile(1))["name"], "agent-5");
    assert_eq!(fx.bytes(&sfile(2)), other, "the other session's entry");
    // The old name is free for anyone again.
    let got = ok(&reg(&fx, Some(&sid(3)), &["--name", "agent-1"]));
    assert_eq!(got["previous"], "none", "{got}");
}

/// REQ-3, REQ-4: a rename also replaces the session's own entry when that is in the name-keyed
/// form; an entry of another session, or one that names no session, is left alone.
#[test]
fn ac1_a_rename_replaces_an_own_entry_of_the_name_keyed_form() {
    let fx = Fixture::new();
    fx.entry_with(
        "agent-1",
        "agent-1",
        Some(&sid(1)),
        "agent",
        LIVE,
        &inst('a'),
    );
    fx.entry_with(
        "agent-9",
        "agent-9",
        Some(&sid(3)),
        "agent",
        LIVE,
        &inst('b'),
    );
    fx.entry_with("agent-8", "agent-8", None, "agent", LIVE, &inst('c'));
    let got = ok(&reg(&fx, Some(&sid(1)), &["--name", "agent-2"]));
    assert_eq!(got["status"], "registered", "{got}");
    assert_eq!(
        fx.files(),
        sorted(&["agent-8.json", "agent-9.json", &sfile(1)])
    );
    assert_eq!(fx.read(&sfile(1))["name"], "agent-2");
}

/// REQ-1: a renewal with the instance rewrites the entry where it is and keeps the session it
/// stores, whoever renews: the same session, a call with no id, or another session.
#[test]
fn ac1_a_renewal_with_the_instance_keeps_the_entry_in_its_file() {
    // From the session that registered (session-keyed file).
    let fx = Fixture::new();
    let first = ok(&reg(&fx, Some(&sid(1)), &["--name", "agent-1"]));
    let instance = first["instance"].as_str().unwrap().to_string();
    let renewed = ok(&reg(
        &fx,
        Some(&sid(1)),
        &["--name", "agent-1", "--instance", &instance],
    ));
    assert_eq!(renewed["status"], "renewed", "{renewed}");
    assert_eq!(fx.files(), sorted(&[&sfile(1)]));
    assert_eq!(fx.read(&sfile(1))["session"], sid(1).as_str());

    // From a call with no id: the session-keyed entry keeps its file and its session.
    let renewed = ok(&reg(
        &fx,
        None,
        &["--name", "agent-1", "--instance", &instance],
    ));
    assert_eq!(renewed["status"], "renewed", "{renewed}");
    assert_eq!(fx.files(), sorted(&[&sfile(1)]));
    assert_eq!(fx.read(&sfile(1))["session"], sid(1).as_str());

    // From another session: the same file, the stored session, no file of the caller's.
    let renewed = ok(&reg(
        &fx,
        Some(&sid(2)),
        &["--name", "agent-1", "--instance", &instance],
    ));
    assert_eq!(renewed["status"], "renewed", "{renewed}");
    assert_eq!(fx.files(), sorted(&[&sfile(1)]));
    assert_eq!(fx.read(&sfile(1))["session"], sid(1).as_str());

    // A name-keyed entry that stores no session stays name keyed when a session with an id
    // renews it, and still stores no session.
    let fx = Fixture::new();
    let first = ok(&reg(&fx, None, &["--name", "agent-2"]));
    let instance = first["instance"].as_str().unwrap().to_string();
    assert_eq!(fx.files(), sorted(&["agent-2.json"]));
    let renewed = ok(&reg(
        &fx,
        Some(&sid(1)),
        &["--name", "agent-2", "--instance", &instance],
    ));
    assert_eq!(renewed["status"], "renewed", "{renewed}");
    assert_eq!(fx.files(), sorted(&["agent-2.json"]));
    assert!(fx.read("agent-2.json").get("session").is_none());
}

/// REQ-1: `--continue` moves the entry of the name to the file of the continuing session and
/// removes the predecessor's file.
#[test]
fn ac1_continue_moves_the_entry_to_the_new_session_file() {
    let fx = Fixture::new();
    ok(&reg(&fx, Some(&sid(1)), &["--name", "agent-9"]));
    assert_eq!(fx.files(), sorted(&[&sfile(1)]));
    let got = ok(&reg(
        &fx,
        Some(&sid(2)),
        &["--name", "agent-9", "--continue"],
    ));
    assert_eq!(got["status"], "continued", "{got}");
    assert_eq!(fx.files(), sorted(&[&sfile(2)]));
    let e = fx.read(&sfile(2));
    assert_eq!(e["name"], "agent-9");
    assert_eq!(e["session"], sid(2).as_str());
    assert_eq!(e["previous_sessions"], json!([sid(1)]));
}

/// REQ-5: `--show` finds the entry of a name in either file form, and a name with no entry, or
/// only an ended one, is `free`.
#[test]
fn ac1_show_reads_an_entry_of_either_form() {
    let fx = Fixture::new();
    fx.entry_with(&sid(1), "agent-1", Some(&sid(1)), "agent", LIVE, &inst('a'));
    fx.entry_with("agent-2", "agent-2", None, "agent", LIVE, &inst('b'));
    fx.entry_with(
        &sid(4),
        "agent-4",
        Some(&sid(4)),
        "agent",
        ENDED,
        &inst('d'),
    );
    let show = |name: &str| ok(&reg(&fx, None, &["--show", "--name", name]));
    let s1 = show("agent-1");
    assert_eq!(s1["status"], "held", "{s1}");
    assert_eq!(s1["instance"], inst('a').as_str());
    assert_eq!(s1["session"], sid(1).as_str());
    assert_eq!(s1["mode"], "agent");
    let s2 = show("agent-2");
    assert_eq!(s2["status"], "held", "{s2}");
    assert_eq!(s2["instance"], inst('b').as_str());
    assert_eq!(show("agent-3")["status"], "free");
    assert_eq!(show("agent-4")["status"], "free");
}

/// REQ-5: `--release` removes the entry of a name in either file form, by the same
/// compare-and-delete, and removes nothing else.
#[test]
fn ac1_release_removes_an_entry_of_either_form() {
    let fx = Fixture::new();
    fx.entry_with(&sid(1), "agent-1", Some(&sid(1)), "agent", LIVE, &inst('a'));
    fx.entry_with("agent-2", "agent-2", None, "agent", LIVE, &inst('b'));
    fx.entry_with(
        &sid(3),
        "agent-3",
        Some(&sid(3)),
        "agent",
        ENDED,
        &inst('c'),
    );
    let release = |name: &str, instance: &str| {
        reg(
            &fx,
            None,
            &["--release", "--name", name, "--instance", instance],
        )
    };
    // Another instance cannot release it, and the file stays.
    let (kind, _) = refused(&release("agent-1", &inst('f')));
    assert_eq!(kind, "conflict");
    assert_eq!(fx.files().len(), 3);
    // The holder does, in the session-keyed form and in the name-keyed form.
    assert_eq!(ok(&release("agent-1", &inst('a')))["status"], "released");
    assert_eq!(fx.files(), sorted(&["agent-2.json", &sfile(3)]));
    assert_eq!(ok(&release("agent-2", &inst('b')))["status"], "released");
    assert_eq!(fx.files(), sorted(&[&sfile(3)]));
    // An ended entry is already free and is not removed by a release.
    assert_eq!(
        ok(&release("agent-3", &inst('c')))["status"],
        "already_free"
    );
    assert_eq!(fx.files(), sorted(&[&sfile(3)]));
}

/// REQ-5: `--lane-check` answers `lane` for a name whose entry runs, in either file form.
#[test]
fn ac1_lane_check_reads_an_entry_of_either_form() {
    let fx = Fixture::new();
    fx.entry_with(&sid(1), "agent-1", Some(&sid(1)), "agent", LIVE, &inst('a'));
    fx.entry_with("agent-2", "agent-2", None, "agent", LIVE, &inst('b'));
    fx.entry_with(
        &sid(3),
        "agent-3",
        Some(&sid(3)),
        "agent",
        ENDED,
        &inst('c'),
    );
    let check = |name: &str| ok(&reg(&fx, None, &["--lane-check", name]));
    for lane in ["agent-1", "agent-2"] {
        let got = check(lane);
        assert_eq!(got["status"], "lane", "{lane}: {got}");
        assert_eq!(got["basis"], "registry_lease", "{lane}");
    }
    for other in ["agent-3", "agent-4"] {
        assert_eq!(check(other)["status"], "not_lane", "{other}");
    }
}

/// REQ-5: the library's liveness lookups find an entry in either file form: the liveness and the
/// launcher trace of a session id, the live lease of a name and its mode.
#[test]
fn ac1_liveness_lookups_read_an_entry_of_either_form() {
    let fx = Fixture::new();
    fx.entry_with(&sid(1), "agent-1", Some(&sid(1)), "agent", LIVE, &inst('a'));
    fx.entry_with(
        "agent-3",
        "agent-3",
        Some(&sid(3)),
        "leader",
        LIVE,
        &inst('c'),
    );
    fx.entry_with(
        &sid(2),
        "agent-2",
        Some(&sid(2)),
        "agent",
        ENDED,
        &inst('b'),
    );
    let common = fx.repo.join(".git");

    // The session id of the session-keyed entry names its lane; no `.live` file yet.
    let trace = launcher_trace_of_session(&fx.repo, &sid(1)).unwrap();
    assert_eq!(
        (trace.name.as_str(), trace.lease_until, trace.live_file),
        ("agent-1", LIVE, false)
    );
    assert_eq!(
        liveness_of_session(&fx.repo, &sid(1)).0,
        Liveness::LauncherExited
    );
    // The name-keyed entry that records a session answers the same way.
    let trace = launcher_trace_of_session(&fx.repo, &sid(3)).unwrap();
    assert_eq!(trace.name, "agent-3");
    assert_eq!(
        liveness_of_session(&fx.repo, &sid(3)).0,
        Liveness::LauncherExited
    );

    // A launcher that holds the name's `.live` lock makes the session alive.
    let live = fs::File::create(fx.sessions.join("agent-1.live")).unwrap();
    live.lock().unwrap();
    assert_eq!(liveness_of_session(&fx.repo, &sid(1)).0, Liveness::Alive);
    assert!(
        launcher_trace_of_session(&fx.repo, &sid(1))
            .unwrap()
            .live_file
    );

    // The lease of a name and its mode, whatever the form of the file.
    assert_eq!(session_live(&fx.repo, "agent-1", NOW), Some(true));
    assert_eq!(session_live(&fx.repo, "agent-3", NOW), Some(true));
    assert_eq!(session_live(&fx.repo, "agent-2", NOW), Some(false), "ended");
    assert_eq!(session_live(&fx.repo, "agent-9", NOW), Some(false), "none");
    assert_eq!(mode_of(&common, "agent-1", NOW), Some("agent"));
    assert_eq!(mode_of(&common, "agent-3", NOW), Some("leader"));
    assert_eq!(mode_of(&common, "agent-2", NOW), None, "ended");
    drop(live);
}

/// REQ-4: an entry whose file is named by the lane name stays readable and writable as before,
/// including one that records a host session id.
#[test]
fn ac1_the_name_keyed_entry_stays_readable() {
    let fx = Fixture::new();
    fx.entry_with(
        "agent-1",
        "agent-1",
        Some(&sid(1)),
        "agent",
        LIVE,
        &inst('a'),
    );
    let held = ok(&reg(&fx, None, &["--show", "--name", "agent-1"]));
    assert_eq!(held["status"], "held", "{held}");
    assert_eq!(held["instance"], inst('a').as_str());
    // Its own session renews it in the file it is in.
    let renewed = ok(&reg(
        &fx,
        Some(&sid(1)),
        &["--name", "agent-1", "--instance", &inst('a')],
    ));
    assert_eq!(renewed["status"], "renewed", "{renewed}");
    assert_eq!(fx.files(), sorted(&["agent-1.json"]));
    assert_eq!(fx.read("agent-1.json")["session"], sid(1).as_str());
    // Another session is refused.
    assert_eq!(
        refused(&reg(&fx, Some(&sid(2)), &["--name", "agent-1"])).0,
        "conflict"
    );
    assert_eq!(fx.files(), sorted(&["agent-1.json"]));
    // The listing reads it.
    let names: Vec<String> = session_entries(&fx.repo)
        .unwrap()
        .into_iter()
        .map(|e| e.name)
        .collect();
    assert_eq!(names, ["agent-1"]);
}

/// REQ-5: the listing and the live-session count read both forms; two entries that name one
/// lane are both listed and count once.
#[test]
fn ac1_session_entries_and_live_session_count_read_both_forms() {
    let fx = Fixture::new();
    fx.entry_with("agent-1", "agent-1", None, "agent", LIVE, &inst('a'));
    fx.entry_with(&sid(1), "agent-2", Some(&sid(1)), "agent", LIVE, &inst('b'));
    fx.entry_with("agent-2", "agent-2", None, "agent", LIVE, &inst('c'));
    fx.entry_with(
        &sid(3),
        "agent-3",
        Some(&sid(3)),
        "agent",
        ENDED,
        &inst('d'),
    );
    let names: Vec<String> = session_entries(&fx.repo)
        .unwrap()
        .into_iter()
        .map(|e| e.name)
        .collect();
    assert_eq!(names, ["agent-1", "agent-2", "agent-2", "agent-3"]);
    assert_eq!(live_session_count(&fx.sessions, NOW), Ok(2));
    fx.entry_with(&sid(4), "agent-4", Some(&sid(4)), "agent", LIVE, &inst('e'));
    assert_eq!(live_session_count(&fx.sessions, NOW), Ok(3));
    assert_eq!(live_session_count(&fx.sessions, LIVE), Ok(0));
}

/// REQ-4, REQ-5: what cannot be read stays an error and the file is kept: a symlink, a malformed
/// file, one over 4096 bytes, and an entry whose file name is neither its name nor its session id.
#[test]
fn ac1_a_broken_or_mismatched_entry_is_still_an_error_and_is_kept() {
    // The file of the name itself: a symlink.
    let fx = Fixture::new();
    let target = fx.sessions.parent().unwrap().join("elsewhere.json");
    fs::write(&target, "keep").unwrap();
    std::os::unix::fs::symlink(&target, fx.sessions.join("agent-1.json")).unwrap();
    let (kind, message) = refused(&reg(&fx, Some(&sid(1)), &["--name", "agent-1"]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("symlink"), "{message}");
    assert_eq!(fs::read_to_string(&target).unwrap(), "keep");
    assert!(fs::symlink_metadata(fx.sessions.join("agent-1.json"))
        .unwrap()
        .file_type()
        .is_symlink());

    let over = {
        let mut s = json!({
            "schema_version": 1, "name": "agent-1", "instance": inst('a'), "lease_until": LIVE
        })
        .to_string();
        s.push_str(&" ".repeat(4097 - s.len()));
        s
    };
    let strange = json!({
        "schema_version": 1, "name": "agent-6", "instance": inst('a'), "lease_until": LIVE
    })
    .to_string();
    for (label, text) in [
        ("malformed", "{not json".to_string()),
        ("over 4096 bytes", over.clone()),
        ("another name", strange.clone()),
    ] {
        let fx = Fixture::new();
        fs::write(fx.sessions.join("agent-1.json"), &text).unwrap();
        let (kind, message) = refused(&reg(&fx, Some(&sid(1)), &["--name", "agent-1"]));
        assert_eq!(kind, "invalid_input", "{label}: {message}");
        assert_eq!(
            fs::read_to_string(fx.sessions.join("agent-1.json")).unwrap(),
            text,
            "{label}"
        );
    }

    // The file of a session id: a lookup of any name fails with the file named, never "free".
    let file = sfile(9);
    let neither = json!({
        "schema_version": 1, "name": "agent-1", "instance": inst('a'),
        "lease_until": LIVE, "session": sid(8)
    })
    .to_string();
    for (label, text) in [
        ("malformed", Some("{not json".to_string())),
        ("over 4096 bytes", Some(over.clone())),
        // Its name is the one looked up, but its file name is neither its name nor its session.
        ("neither name nor session", Some(neither)),
        ("symlink", None),
    ] {
        let fx = Fixture::new();
        match &text {
            Some(t) => fs::write(fx.sessions.join(&file), t).unwrap(),
            None => {
                let other = fx.sessions.parent().unwrap().join("other.json");
                fs::write(&other, "keep").unwrap();
                std::os::unix::fs::symlink(&other, fx.sessions.join(&file)).unwrap();
            }
        }
        for args in [
            vec!["--show", "--name", "agent-1"],
            vec!["--name", "agent-1"],
            vec!["--show", "--name", "agent-2"],
        ] {
            let (kind, message) = refused(&reg(&fx, Some(&sid(1)), &args));
            assert_eq!(kind, "invalid_input", "{label} {args:?}: {message}");
            assert!(message.contains(&file), "{label} {args:?}: {message}");
        }
        assert!(
            fs::symlink_metadata(fx.sessions.join(&file)).is_ok(),
            "{label}: the file is kept"
        );
        assert!(!fx.sessions.join("agent-1.json").exists(), "{label}");
    }

    // A file name that is neither the name nor the session id of its entry is skipped by the
    // lookup of a session id.
    let fx = Fixture::new();
    fx.entry_with(
        "agent-7",
        "agent-9",
        Some(&sid(8)),
        "agent",
        LIVE,
        &inst('a'),
    );
    assert!(launcher_trace_of_session(&fx.repo, &sid(8)).is_err());
    assert_eq!(liveness_of_session(&fx.repo, &sid(8)).0, Liveness::Unknown);
}

/// REQ-4, REQ-5: a leader entry in the session-keyed form is read by the leader registration:
/// its session is `alive` by the lease, and a second leader is refused as a leader conflict, not
/// with an entry error.
#[test]
fn ac1_a_session_keyed_leader_entry_is_read_by_the_leader_registration() {
    let fx = Fixture::new();
    fx.entry_with(
        &sid(1),
        "demo-leader",
        Some(&sid(1)),
        "leader",
        LIVE,
        &inst('a'),
    );
    let got = ok(&reg(&fx, None, &["--liveness-of", &sid(1)]));
    assert_eq!(got["state"], "alive", "{got}");
    assert_eq!(got["evidence"]["basis"], "registry_lease", "{got}");
    assert_eq!(got["evidence"]["name"], "demo-leader", "{got}");

    let before = fx.bytes(&sfile(1));
    let (kind, message) = refused(&reg(
        &fx,
        Some(&sid(2)),
        &["--name", "other-leader", "--mode", "leader"],
    ));
    assert_eq!(kind, "conflict", "{message}");
    assert!(
        message.contains("the leader entry `demo-leader` is live until"),
        "{message}"
    );
    assert_eq!(fx.files(), sorted(&[&sfile(1)]));
    assert_eq!(fx.bytes(&sfile(1)), before);
}
