//! SPEC-20261121-leader-run-lease: one supervising leader run per project. `leader-run begin`
//! takes a lease (a file in the project's state folder, written under a file lock), `skipped` tells a
//! second tick who holds it, an ended lease is taken over, nothing ready or a HOLD gives `idle`
//! without a lease, and `end` releases only the holder's lease. Every command runs in a clean
//! environment with a temporary HOME and a temporary project; times are given with `--now`.

mod common;

use common::home::TempHome;
use serde_json::Value;
use std::fs;
use std::path::PathBuf;
use std::process::Output;

struct World {
    home: TempHome,
    project: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("lease-fixture");
        World { home, project }
    }

    fn lease_file(&self) -> PathBuf {
        moai_cli::state_root::path(&self.project, "state").join("leader-run.lease")
    }

    fn run(&self, args: &[&str]) -> Output {
        let project = self.project.to_str().unwrap();
        let mut full = vec!["leader-run", args[0], project];
        full.extend_from_slice(&args[1..]);
        self.home.command(&full).output().expect("run moai")
    }

    fn begin(&self, session: &str, ready: u32, now: u64, extra: &[&str]) -> Value {
        let (ready, now) = (ready.to_string(), now.to_string());
        let mut a = vec![
            "begin",
            "--session",
            session,
            "--ready",
            &ready,
            "--lease-secs",
            "60",
            "--now",
            &now,
        ];
        a.extend_from_slice(extra);
        ok(&self.run(&a))
    }
}

fn parse(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = parse(out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["ok"], Value::Bool(true), "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> (String, String) {
    let v = parse(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    assert_eq!(v["ok"], Value::Bool(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

/// REQ-1: nothing held and a card ready: `taken`, and the lease names the session and its end.
#[test]
fn begin_takes_the_lease_when_a_card_is_ready() {
    let w = World::new();
    let r = w.begin("lead-a", 2, 1000, &[]);
    assert_eq!(r["result"], "taken", "{r}");
    assert_eq!(r["renewed"], Value::Bool(false), "{r}");
    assert_eq!(r["lease_until"], 1060, "{r}");
    let file: Value = serde_json::from_str(&fs::read_to_string(w.lease_file()).unwrap()).unwrap();
    assert_eq!(file["session"], "lead-a");
    assert_eq!(file["started"], 1000);
    assert_eq!(file["lease_until"], 1060);
}

/// REQ-2: a second session while the lease runs: `skipped` with the holder, the age and the
/// lease end, and the lease file is byte for byte what it was.
#[test]
fn a_second_session_is_skipped_and_nothing_changes() {
    let w = World::new();
    w.begin("lead-a", 1, 1000, &[]);
    let before = fs::read(w.lease_file()).unwrap();
    let r = w.begin("lead-b", 1, 1030, &[]);
    assert_eq!(r["result"], "skipped", "{r}");
    assert_eq!(r["holder"], "lead-a", "{r}");
    assert_eq!(r["age_secs"], 30, "{r}");
    assert_eq!(r["lease_until"], 1060, "{r}");
    assert_eq!(fs::read(w.lease_file()).unwrap(), before);
}

/// REQ-2, the overlap: two ticks that start at the same instant give one `taken` and one
/// `skipped`, whoever is first; repeated in fresh projects.
#[test]
fn two_overlapping_begins_give_one_taken_and_one_skipped() {
    for round in 0..20 {
        let w = World::new();
        let project = w.project.to_str().unwrap().to_string();
        let mut kids = Vec::new();
        for who in ["lead-a", "lead-b"] {
            let mut cmd = w.home.command(&[
                "leader-run",
                "begin",
                &project,
                "--session",
                who,
                "--ready",
                "1",
                "--lease-secs",
                "60",
                "--now",
                "5000",
            ]);
            kids.push(cmd.spawn_piped());
        }
        let results: Vec<String> = kids
            .into_iter()
            .map(|k| {
                ok(&k.wait_with_output().unwrap())["result"]
                    .as_str()
                    .unwrap()
                    .to_string()
            })
            .collect();
        let taken = results.iter().filter(|r| *r == "taken").count();
        let skipped = results.iter().filter(|r| *r == "skipped").count();
        assert_eq!((taken, skipped), (1, 1), "round {round}: {results:?}");
    }
}

trait SpawnPiped {
    fn spawn_piped(&mut self) -> std::process::Child;
}

impl SpawnPiped for std::process::Command {
    fn spawn_piped(&mut self) -> std::process::Child {
        self.stdin(std::process::Stdio::null())
            .stdout(std::process::Stdio::piped())
            .stderr(std::process::Stdio::piped())
            .spawn()
            .expect("run moai")
    }
}

/// REQ-3: an ended lease (a restart, a leader that never ended) is taken over, and the answer
/// says whose lease it replaced.
#[test]
fn an_ended_lease_is_taken_over() {
    let w = World::new();
    w.begin("lead-a", 1, 1000, &[]);
    let r = w.begin("lead-b", 1, 1061, &[]);
    assert_eq!(r["result"], "taken", "{r}");
    assert_eq!(r["replaced"]["session"], "lead-a", "{r}");
    assert_eq!(r["replaced"]["lease_until"], 1060, "{r}");
    let file: Value = serde_json::from_str(&fs::read_to_string(w.lease_file()).unwrap()).unwrap();
    assert_eq!(file["session"], "lead-b");
    // The edge: at the lease end itself the lease still holds.
    let w = World::new();
    w.begin("lead-a", 1, 1000, &[]);
    let r = w.begin("lead-b", 1, 1060, &[]);
    assert_eq!(r["result"], "skipped", "{r}");
}

/// REQ-4: the holder ends its lease; another session's end is refused and changes nothing; an
/// end with no lease is not an error and says so.
#[test]
fn end_releases_only_the_holders_lease() {
    let w = World::new();
    w.begin("lead-a", 1, 1000, &[]);
    let before = fs::read(w.lease_file()).unwrap();
    let (kind, message) = refused(&w.run(&["end", "--session", "lead-b"]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("lead-a"), "{message}");
    assert_eq!(fs::read(w.lease_file()).unwrap(), before);
    let r = ok(&w.run(&["end", "--session", "lead-a"]));
    assert_eq!(r["released"], Value::Bool(true), "{r}");
    assert!(!w.lease_file().exists());
    let r = ok(&w.run(&["end", "--session", "lead-a"]));
    assert_eq!(r["released"], Value::Bool(false), "{r}");
}

/// REQ-5: no ready card, or a HOLD: `idle`, no lease, and what wakes the run; the first call
/// that reports a ready card (0 to 1) takes the lease.
#[test]
fn nothing_ready_or_a_hold_is_idle_without_a_lease_until_a_card_is_ready() {
    let w = World::new();
    let r = w.begin("lead-a", 0, 1000, &[]);
    assert_eq!(r["result"], "idle", "{r}");
    assert_eq!(r["wake_on"], "card_ready", "{r}");
    assert!(!w.lease_file().exists());
    let r = w.begin("lead-a", 3, 1010, &["--hold"]);
    assert_eq!(r["result"], "idle", "{r}");
    assert_eq!(r["wake_on"], "hold_lifted", "{r}");
    assert!(!w.lease_file().exists());
    let r = w.begin("lead-a", 1, 1020, &[]);
    assert_eq!(r["result"], "taken", "{r}");
    assert!(w.lease_file().is_file());
}

/// REQ-5: idle never touches a lease that another session holds, and says so.
#[test]
fn idle_leaves_another_sessions_lease_alone() {
    let w = World::new();
    w.begin("lead-a", 1, 1000, &[]);
    let before = fs::read(w.lease_file()).unwrap();
    let r = w.begin("lead-b", 0, 1010, &[]);
    assert_eq!(r["result"], "idle", "{r}");
    assert_eq!(fs::read(w.lease_file()).unwrap(), before);
}

/// REQ-6: the holder's own next tick renews the lease (a tick that never ended).
#[test]
fn the_same_session_renews_its_unexpired_lease() {
    let w = World::new();
    w.begin("lead-a", 1, 1000, &[]);
    let r = w.begin("lead-a", 1, 1030, &[]);
    assert_eq!(r["result"], "taken", "{r}");
    assert_eq!(r["renewed"], Value::Bool(true), "{r}");
    assert_eq!(r["lease_until"], 1090, "{r}");
    let file: Value = serde_json::from_str(&fs::read_to_string(w.lease_file()).unwrap()).unwrap();
    assert_eq!(file["started"], 1000, "a renewal keeps the start");
    assert_eq!(file["lease_until"], 1090);
}

/// `status` reads only: it reports the holder, the age and whether the lease has ended, and
/// writes nothing.
#[test]
fn status_reads_and_writes_nothing() {
    let w = World::new();
    // No lock file and no folder either: the list of the project's files is the same before
    // and after the very first status on a project that never had a lease.
    let empty_before = project_files(&w.project);
    let r = ok(&w.run(&["status", "--now", "1000"]));
    assert_eq!(r["lease"], Value::Null, "{r}");
    assert!(!w.lease_file().exists());
    assert_eq!(
        project_files(&w.project),
        empty_before,
        "status created a file"
    );
    w.begin("lead-a", 1, 1000, &[]);
    let before = fs::read(w.lease_file()).unwrap();
    let files_before = project_files(&w.project);
    let r = ok(&w.run(&["status", "--now", "1100"]));
    assert_eq!(r["lease"]["session"], "lead-a", "{r}");
    assert_eq!(r["lease"]["age_secs"], 100, "{r}");
    assert_eq!(r["lease"]["ended"], Value::Bool(true), "{r}");
    assert_eq!(fs::read(w.lease_file()).unwrap(), before);
    assert_eq!(
        project_files(&w.project),
        files_before,
        "status changed a file"
    );
}

/// Limits: a session name that is empty, longer than 80 characters or holds a control character,
/// a lease of 59 or 86401 seconds, and a begin without `--ready` are refused before anything is
/// written; 60 and 86400 are accepted.
#[test]
fn bad_input_is_refused_before_anything_is_written() {
    let w = World::new();
    let long = "x".repeat(81);
    for session in ["", long.as_str(), "a\nb"] {
        let (kind, _) = refused(&w.run(&[
            "begin",
            "--session",
            session,
            "--ready",
            "1",
            "--lease-secs",
            "60",
            "--now",
            "1000",
        ]));
        assert_eq!(kind, "invalid_input", "session {session:?}");
    }
    for secs in ["59", "86401"] {
        let (kind, _) = refused(&w.run(&[
            "begin",
            "--session",
            "lead-a",
            "--ready",
            "1",
            "--lease-secs",
            secs,
            "--now",
            "1000",
        ]));
        assert_eq!(kind, "invalid_input", "lease {secs}");
    }
    assert!(!w.lease_file().exists(), "nothing was written");
    for secs in ["60", "86400"] {
        let w = World::new();
        let r = ok(&w.run(&[
            "begin",
            "--session",
            "lead-a",
            "--ready",
            "1",
            "--lease-secs",
            secs,
            "--now",
            "1000",
        ]));
        assert_eq!(r["result"], "taken", "lease {secs}: {r}");
    }
    // The session name has the same edge: 80 characters are accepted, 81 are refused above.
    let w = World::new();
    let edge = "x".repeat(80);
    let r = w.begin(&edge, 1, 1000, &[]);
    assert_eq!(r["result"], "taken", "an 80 character session name: {r}");
    let out = w.run(&["begin", "--session", "lead-a", "--now", "1000"]);
    assert_eq!(
        out.status.code(),
        Some(2),
        "a begin without --ready is a usage error"
    );
}

/// A lease file that cannot be read as a lease (not JSON) is not taken over silently and not
/// overwritten: the answer is a refusal that names the file.
#[test]
fn a_damaged_lease_file_is_refused_and_kept() {
    let w = World::new();
    fs::create_dir_all(w.lease_file().parent().unwrap()).unwrap();
    fs::write(w.lease_file(), "not json").unwrap();
    let out = w.run(&[
        "begin",
        "--session",
        "lead-a",
        "--ready",
        "1",
        "--lease-secs",
        "60",
        "--now",
        "1000",
    ]);
    let (kind, message) = refused(&out);
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("leader-run.lease"), "{message}");
    assert_eq!(fs::read_to_string(w.lease_file()).unwrap(), "not json");
}

/// The command writes only the lease file and its lock file (`leader-run.guard`, held while a
/// call decides) inside the project's state folder: nothing under HOME and nothing else in the
/// project.
#[test]
fn only_the_lease_file_is_written() {
    let w = World::new();
    let home_before = w.home.snapshot();
    let project_before = project_files(&w.project);
    w.begin("lead-a", 1, 1000, &[]);
    let after = project_files(&w.project);
    let added: Vec<&String> = after
        .iter()
        .filter(|f| !project_before.contains(f))
        .collect();
    let mut names: Vec<&str> = added
        .iter()
        .map(|f| f.rsplit('/').next().unwrap())
        .collect();
    names.sort();
    assert_eq!(names, ["leader-run.guard", "leader-run.lease"], "{added:?}");
    let home_after = w.home.snapshot();
    let new_outside: Vec<_> = home_after
        .iter()
        .filter(|e| !e.path.starts_with("work/") && !home_before.contains(e))
        .collect();
    assert!(new_outside.is_empty(), "{new_outside:?}");
}

fn project_files(dir: &std::path::Path) -> Vec<String> {
    fn walk(dir: &std::path::Path, base: &std::path::Path, out: &mut Vec<String>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                walk(&p, base, out);
            } else {
                out.push(p.strip_prefix(base).unwrap().to_string_lossy().to_string());
            }
        }
    }
    let mut out = Vec::new();
    walk(dir, dir, &mut out);
    out.sort();
    out
}
