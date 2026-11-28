//! SPEC-20261126-leader-slot-front (AC-1 to AC-4): the `--full` run of `check affected` made by the
//! session that the live registry entry `leader` (mode leader) names takes a place at the front of
//! the heavy slot waiting queue; every other caller queues in arrival order. Only the order of
//! waiters changes: the slot count, the limits and the running commands stay as they are. Every
//! case uses a temporary git repository, a temporary HOME, a temporary slot folder with one slot
//! and fake `cargo`, `python3` and `rustfmt` scripts; the real tools, the machine's slots and the
//! real registry are never used. Waits are bounded by log markers plus a hang guard.
#![cfg(unix)]
mod common;

use common::git::{git, git_ok};
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::os::unix::process::CommandExt;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Output, Stdio};
use std::time::{Duration, Instant};

const BIN: &str = crate::common::MOAI_BIN;
const HANG: Duration = Duration::from_secs(120);
/// Roomy machine numbers (debug-only seam), so the real disk and memory never decide.
const ROOMY: &str = "free_gb=1000,swap_pct=0";
/// Host session ids in the shape the registry accepts.
const LEADER: &str = "11111111-1111-4111-8111-111111111111";
const OTHER: &str = "22222222-2222-4222-8222-222222222222";
const AGENT: &str = "33333333-3333-4333-8333-333333333333";

/// A wrapper in a process group of its own, killed with its command on drop.
struct Spawned {
    child: Option<Child>,
    group: i32,
}

impl Spawned {
    fn finish(mut self) -> Output {
        let mut child = self.child.take().unwrap();
        let start = Instant::now();
        while child.try_wait().unwrap().is_none() {
            if start.elapsed() > HANG {
                kill_group(self.group);
                let _ = child.wait();
                panic!("hang guard: the wrapper did not end");
            }
            std::thread::sleep(Duration::from_millis(20));
        }
        child.wait_with_output().unwrap()
    }
}

impl Drop for Spawned {
    fn drop(&mut self) {
        kill_group(self.group);
        if let Some(mut c) = self.child.take() {
            let _ = c.wait();
        }
    }
}

fn kill_group(group: i32) {
    // SAFETY: kill(2) on a process group this test created; ESRCH when it is already gone.
    unsafe { libc::kill(-group, libc::SIGKILL) };
}

fn spawn(mut c: Command, stdout: &Path, stderr: &Path) -> Spawned {
    c.stdin(Stdio::null())
        .stdout(fs::File::create(stdout).unwrap())
        .stderr(fs::File::create(stderr).unwrap())
        .process_group(0);
    let child = c.spawn().unwrap();
    let group = child.id() as i32;
    Spawned {
        child: Some(child),
        group,
    }
}

fn wait_for(what: &str, mut cond: impl FnMut() -> bool) {
    let start = Instant::now();
    while !cond() {
        assert!(start.elapsed() < HANG, "hang guard: {what}");
        std::thread::sleep(Duration::from_millis(20));
    }
}

fn text(p: &Path) -> String {
    fs::read_to_string(p).unwrap_or_default()
}

fn put(root: &Path, rel: &str, body: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, body).unwrap();
}

/// A repository with one commit on top of the base that changes a file one test target covers.
struct Fixture {
    dir: tempfile::TempDir,
    base: String,
    head: String,
}

impl Fixture {
    fn new() -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        put(p, "src/update_models.rs", "// a\n");
        put(p, "tests/update_command_cli.rs", "fn main() {}\n");
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        put(p, "src/update_models.rs", "// b\n");
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "change"]);
        let head = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        Fixture { dir, base, head }
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    fn range(&self) -> String {
        format!("{}..{}", self.base, self.head)
    }
}

/// Which session, flags and machine one `check affected` call is made with.
#[derive(Clone)]
struct Call {
    sid: Option<&'static str>,
    full: bool,
    max: &'static str,
    machine: String,
    /// Another checkout of the same repository to run in (a linked worktree).
    dir: Option<PathBuf>,
}

impl Call {
    /// The registered leader session's `--full` run.
    fn leader() -> Call {
        Call {
            sid: Some(LEADER),
            full: true,
            max: "1",
            machine: ROOMY.to_string(),
            dir: None,
        }
    }
}

/// The fixture, the fakes, a slot folder and a HOME of one case.
struct Lab {
    fx: Fixture,
    tools: tempfile::TempDir,
    slots: tempfile::TempDir,
    home: tempfile::TempDir,
    work: tempfile::TempDir,
}

impl Lab {
    /// The limits file allows one heavy run and, when `cpu` is more than 0, holds a run back
    /// while the CPU free percentage is below it.
    fn new(cpu: u32) -> Lab {
        let tools = tempfile::tempdir().unwrap();
        for n in ["cargo", "python3", "rustfmt"] {
            let script =
                format!("#!/bin/sh\necho \"$FAKE_TAG {n} $*\" >> \"$FAKE_CALLS\"\nexit 0\n");
            let path = tools.path().join(n);
            fs::write(&path, script).unwrap();
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
        let slots = tempfile::tempdir().unwrap();
        fs::write(
            slots.path().join("limits.json"),
            format!(
                r#"{{"max":1,"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0,"min_cpu_free_pct":{cpu}}}"#
            ),
        )
        .unwrap();
        Lab {
            fx: Fixture::new(),
            tools,
            slots,
            home: tempfile::tempdir().unwrap(),
            work: tempfile::tempdir().unwrap(),
        }
    }

    fn calls(&self) -> PathBuf {
        self.tools.path().join("calls.log")
    }

    /// The lines of the call log, in the order the commands ran.
    fn order(&self) -> Vec<String> {
        text(&self.calls()).lines().map(String::from).collect()
    }

    /// The place of the first call line that starts with `prefix`; the case fails without one.
    fn at(&self, prefix: &str) -> usize {
        let lines = self.order();
        lines
            .iter()
            .position(|l| l.starts_with(prefix))
            .unwrap_or_else(|| panic!("no `{prefix}` call in {lines:?}"))
    }

    fn path(&self) -> String {
        let git_dir = Command::new("sh")
            .args(["-c", "dirname \"$(command -v git)\""])
            .output()
            .unwrap();
        format!(
            "{}:{}:/usr/bin:/bin",
            self.tools.path().display(),
            String::from_utf8_lossy(&git_dir.stdout).trim()
        )
    }

    /// The environment every command of a case shares; no session id until a caller sets one.
    fn env(&self, c: &mut Command, tag: &str) {
        c.env("HOME", self.home.path())
            .env_remove("MOAI_HOME")
            .env_remove("XDG_CACHE_HOME")
            .env_remove("CARGO_TARGET_DIR")
            .env_remove("MOAI_HEAVY_SLOT_FD")
            .env_remove("RUST_TEST_THREADS")
            .env_remove("NEXTEST_TEST_THREADS")
            .env_remove("CLAUDE_CODE_SESSION_ID")
            .env_remove("CODEX_THREAD_ID")
            .env("MOAI_TEST_MACHINE", ROOMY)
            .env("PATH", self.path())
            .env("FAKE_CALLS", self.calls())
            .env("FAKE_TAG", tag);
    }

    /// `moai session-register --name <name> <extra…> <repository>` as the session `sid`.
    fn register(&self, sid: &str, name: &str, extra: &[&str]) -> Output {
        let mut c = Command::new(BIN);
        self.env(&mut c, "register");
        c.env("CLAUDE_CODE_SESSION_ID", sid)
            .args(["session-register", "--name", name])
            .args(extra)
            .arg(self.fx.root());
        c.stdin(Stdio::null()).output().unwrap()
    }

    fn register_leader(&self) {
        let o = self.register(LEADER, "leader", &["--mode", "leader"]);
        assert!(o.status.success(), "{o:?}");
    }

    /// A lane run (`heavy-slot-acquire --name <name>`) whose command writes `tag` to the call log;
    /// returned once it waits in the queue.
    fn lane(&self, tag: &str, name: &str, sid: Option<&str>) -> Spawned {
        let out = self.work.path().join(format!("{tag}.out"));
        let mut c = Command::new(BIN);
        self.env(&mut c, tag);
        if let Some(sid) = sid {
            c.env("CLAUDE_CODE_SESSION_ID", sid);
        }
        c.args(["heavy-slot-acquire", "--max", "1", "--name", name, "--dir"])
            .arg(self.slots.path())
            .args(["--", "sh", "-c", "echo \"$FAKE_TAG\" >> \"$FAKE_CALLS\""]);
        let s = spawn(c, &out, &self.work.path().join(format!("{tag}.err")));
        wait_for(&format!("{tag} waiting"), || {
            text(&out).contains("heavy-slot: waiting")
        });
        s
    }

    /// A lane run that holds the slot until `go` exists, then writes `tag`; returned once it runs.
    fn running(&self, tag: &str, go: &Path) -> Spawned {
        let out = self.work.path().join(format!("{tag}.out"));
        let started = self.work.path().join(format!("{tag}.started"));
        let mut c = Command::new(BIN);
        self.env(&mut c, tag);
        let script = format!(
            "touch '{}'; until [ -f '{}' ]; do sleep 0.1; done; echo \"$FAKE_TAG\" >> \"$FAKE_CALLS\"",
            started.display(),
            go.display()
        );
        c.args([
            "heavy-slot-acquire",
            "--max",
            "1",
            "--name",
            "running",
            "--dir",
        ])
        .arg(self.slots.path())
        .args(["--", "sh", "-c", &script]);
        let s = spawn(c, &out, &self.work.path().join(format!("{tag}.err")));
        wait_for(&format!("{tag} running"), || started.exists());
        s
    }

    /// `check affected` of the fixture range as `call`; returned with its log folder.
    fn check(&self, tag: &str, call: &Call) -> (Spawned, PathBuf) {
        let (child, out_dir) = self.check_unwaited(tag, call);
        wait_for(&format!("{tag} in the queue"), || {
            text(&out_dir.join("heavy-slot.log")).contains("waiting")
        });
        (child, out_dir)
    }

    fn check_unwaited(&self, tag: &str, call: &Call) -> (Spawned, PathBuf) {
        let out_dir = self.work.path().join(format!("{tag}.logs"));
        fs::create_dir_all(&out_dir).unwrap();
        let spawned = spawn(
            self.check_command(tag, call, &out_dir),
            &self.work.path().join(format!("{tag}.json")),
            &self.work.path().join(format!("{tag}.err")),
        );
        (spawned, out_dir)
    }

    fn check_command(&self, tag: &str, call: &Call, out_dir: &Path) -> Command {
        let mut c = Command::new(BIN);
        self.env(&mut c, tag);
        c.env("MOAI_TEST_MACHINE", &call.machine);
        if let Some(sid) = call.sid {
            c.env("CLAUDE_CODE_SESSION_ID", sid);
        }
        c.args(["check", "affected", &self.fx.range(), "--dir"])
            .arg(call.dir.as_deref().unwrap_or(self.fx.root()))
            .arg("--slot-dir")
            .arg(self.slots.path())
            .arg("--out")
            .arg(out_dir)
            .args(["--max", call.max]);
        if call.full {
            c.arg("--full");
        }
        c
    }

    /// Slot `i` of the slot folder held by this test, a process apart from every command.
    fn hold(&self, i: usize) -> fs::File {
        let f = fs::OpenOptions::new()
            .read(true)
            .write(true)
            .create(true)
            .truncate(false)
            .open(self.slots.path().join(format!("slot-{i}")))
            .unwrap();
        f.lock().unwrap();
        f
    }
}

fn ok(s: Spawned) {
    let o = s.finish();
    assert!(o.status.success(), "{o:?}");
}

/// One lane run waits behind a held slot; `call` arrives after it and waits too; the slot is let
/// go. The lane run starts first and the call's log folder is returned for its queue line.
fn arrives_behind_a_lane(lab: &Lab, call: &Call) -> String {
    let slot = lab.hold(0);
    let l1 = lab.lane("L1", "lane-1", None);
    let (c, logs) = lab.check("C", call);
    drop(slot);
    ok(l1);
    ok(c);
    assert!(
        lab.at("L1") < lab.at("C cargo clippy"),
        "a run that is not the front run passed the lane: {:?}",
        lab.order()
    );
    text(&logs.join("heavy-slot.log"))
}

// ---- AC-1 (REQ-1, REQ-2, REQ-3) --------------------------------------------------------------

/// The registered leader session's `--full` run that arrives after two waiting lane runs starts
/// before both.
#[test]
fn front_ac1_a_leader_full_run_starts_before_two_waiting_lane_runs() {
    let lab = Lab::new(0);
    lab.register_leader();
    let slot = lab.hold(0);
    let l1 = lab.lane("L1", "lane-1", None);
    let l2 = lab.lane("L2", "lane-2", None);
    let (a, logs) = lab.check("A", &Call::leader());
    let log = text(&logs.join("heavy-slot.log"));
    assert!(
        log.contains("waiting 1 of 3 for 1 slot(s)"),
        "the leader run is not first of three: {log}"
    );
    drop(slot);
    ok(a);
    ok(l1);
    ok(l2);
    let (a_at, l1_at, l2_at) = (lab.at("A cargo clippy"), lab.at("L1"), lab.at("L2"));
    assert!(a_at < l1_at && l1_at < l2_at, "order was {:?}", lab.order());
}

/// Two leader runs keep their arrival order ahead of the lane runs.
#[test]
fn front_ac1_two_leader_runs_keep_their_arrival_order_ahead_of_the_lanes() {
    let lab = Lab::new(0);
    lab.register_leader();
    let slot = lab.hold(0);
    let l1 = lab.lane("L1", "lane-1", None);
    let (a, a_logs) = lab.check("A", &Call::leader());
    let (b, b_logs) = lab.check("B", &Call::leader());
    assert!(
        text(&a_logs.join("heavy-slot.log")).contains("waiting 1 of 2 for 1 slot(s)"),
        "{}",
        text(&a_logs.join("heavy-slot.log"))
    );
    assert!(
        text(&b_logs.join("heavy-slot.log")).contains("waiting 2 of 3 for 1 slot(s)"),
        "{}",
        text(&b_logs.join("heavy-slot.log"))
    );
    drop(slot);
    ok(a);
    ok(b);
    ok(l1);
    let (a_at, b_at, l1_at) = (
        lab.at("A cargo clippy"),
        lab.at("B cargo clippy"),
        lab.at("L1"),
    );
    assert!(a_at < b_at && b_at < l1_at, "order was {:?}", lab.order());
}

/// A run in a linked worktree reads the leader entry of its repository.
#[test]
fn front_ac1_a_run_in_a_linked_worktree_reads_the_repository_leader_entry() {
    let lab = Lab::new(0);
    lab.register_leader();
    let linked = lab.work.path().join("linked");
    git_ok(
        lab.fx.root(),
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            linked.to_str().unwrap(),
            &lab.fx.head,
        ],
    );
    let slot = lab.hold(0);
    let l1 = lab.lane("L1", "lane-1", None);
    let call = Call {
        dir: Some(linked),
        ..Call::leader()
    };
    let (a, logs) = lab.check("A", &call);
    assert!(
        text(&logs.join("heavy-slot.log")).contains("waiting 1 of 2 for 1 slot(s)"),
        "{}",
        text(&logs.join("heavy-slot.log"))
    );
    drop(slot);
    ok(a);
    ok(l1);
    assert!(lab.at("A cargo clippy") < lab.at("L1"), "{:?}", lab.order());
}

// ---- AC-2 (REQ-6) ----------------------------------------------------------------------------

/// The running command finishes untouched before the leader run starts, `--max 4` does not raise
/// the file's `max` of 1 and `limits.json` keeps its bytes.
#[test]
fn front_ac2_the_running_command_finishes_and_the_limits_stay() {
    let lab = Lab::new(0);
    lab.register_leader();
    let limits = fs::read(lab.slots.path().join("limits.json")).unwrap();
    let go = lab.work.path().join("go");
    let running = lab.running("R", &go);
    let call = Call {
        max: "4",
        ..Call::leader()
    };
    let (a, logs) = lab.check("A", &call);
    assert!(
        text(&logs.join("heavy-slot.log")).contains("waiting 1 of 1 for 1 slot(s)"),
        "--max 4 raised the file limit of 1: {}",
        text(&logs.join("heavy-slot.log"))
    );
    // The front place does not start the leader run while the slot is held.
    std::thread::sleep(Duration::from_millis(800));
    assert!(
        !lab.order().iter().any(|l| l.starts_with("A cargo clippy")),
        "the leader run started over a held slot: {:?}",
        lab.order()
    );
    fs::write(&go, "").unwrap();
    ok(running);
    ok(a);
    assert!(lab.at("R") < lab.at("A cargo clippy"), "{:?}", lab.order());
    assert!(
        !lab.slots.path().join("slot-1").exists(),
        "a second slot appeared"
    );
    assert_eq!(
        fs::read(lab.slots.path().join("limits.json")).unwrap(),
        limits,
        "limits.json changed"
    );
}

/// A busy machine answers the leader run with exit 75 on every cargo step, front place or not.
#[test]
fn front_ac2_a_busy_machine_answers_the_leader_run_with_75() {
    let lab = Lab::new(0);
    lab.register_leader();
    let out_dir = lab.work.path().join("busy.logs");
    fs::create_dir_all(&out_dir).unwrap();
    let call = Call {
        machine: "free_gb=1,swap_pct=0".to_string(),
        ..Call::leader()
    };
    let o = lab.check_command("A", &call, &out_dir).output().unwrap();
    let stdout = String::from_utf8_lossy(&o.stdout);
    assert!(
        stdout.contains(" clippy=75 test=75 result=FAIL"),
        "{stdout} / {}",
        String::from_utf8_lossy(&o.stderr)
    );
    assert!(
        !lab.order().iter().any(|l| l.contains("cargo clippy")),
        "{:?}",
        lab.order()
    );
}

/// A low CPU reading holds the leader run too: the front place is an order, not a pass.
#[test]
fn front_ac2_a_low_cpu_reading_holds_the_leader_run_too() {
    let lab = Lab::new(50);
    lab.register_leader();
    let call = Call {
        machine: format!("{ROOMY},cpu_free_pct=10,cpu_cores=8"),
        ..Call::leader()
    };
    let (a, logs) = lab.check_unwaited("A", &call);
    wait_for("the CPU wait line", || {
        text(&logs.join("heavy-slot.log")).contains("waiting for CPU headroom")
    });
    std::thread::sleep(Duration::from_millis(800));
    assert!(
        !lab.order().iter().any(|l| l.starts_with("A cargo clippy")),
        "the leader run started at 10% free CPU: {:?}",
        lab.order()
    );
    drop(a);
}

// ---- AC-3 (REQ-4, REQ-5) ---------------------------------------------------------------------

/// A session that is not the registered leader's queues in arrival order, with a reason.
#[test]
fn front_ac3_another_session_queues_in_arrival_order() {
    let lab = Lab::new(0);
    lab.register_leader();
    let call = Call {
        sid: Some(OTHER),
        ..Call::leader()
    };
    let log = arrives_behind_a_lane(&lab, &call);
    assert!(log.contains("arrival order"), "{log}");
}

/// `heavy-slot-acquire --name leader` is a lane run whatever its name and session id.
#[test]
fn front_ac3_a_caller_named_leader_queues_in_arrival_order() {
    let lab = Lab::new(0);
    lab.register_leader();
    let slot = lab.hold(0);
    let l1 = lab.lane("L1", "lane-1", None);
    let ld = lab.lane("LD", "leader", Some(LEADER));
    drop(slot);
    ok(l1);
    ok(ld);
    assert!(lab.at("L1") < lab.at("LD"), "{:?}", lab.order());
}

/// An entry that is not in leader mode confers nothing.
#[test]
fn front_ac3_an_entry_without_leader_mode_queues_in_arrival_order() {
    let lab = Lab::new(0);
    let o = lab.register(AGENT, "leader", &["--mode", "agent"]);
    assert!(o.status.success(), "{o:?}");
    let call = Call {
        sid: Some(AGENT),
        ..Call::leader()
    };
    let log = arrives_behind_a_lane(&lab, &call);
    assert!(log.contains("arrival order"), "{log}");
}

/// An entry whose lease has passed confers nothing.
#[test]
fn front_ac3_an_entry_past_its_lease_queues_in_arrival_order() {
    let lab = Lab::new(0);
    let o = lab.register(
        LEADER,
        "leader",
        &[
            "--mode",
            "leader",
            "--lease-secs",
            "60",
            "--now",
            "2020-01-01T00:00:00Z",
        ],
    );
    assert!(o.status.success(), "{o:?}");
    let log = arrives_behind_a_lane(&lab, &Call::leader());
    assert!(log.contains("arrival order"), "{log}");
}

/// A call without a session id queues in arrival order even with a live leader entry.
#[test]
fn front_ac3_a_call_without_a_session_id_queues_in_arrival_order() {
    let lab = Lab::new(0);
    lab.register_leader();
    let call = Call {
        sid: None,
        ..Call::leader()
    };
    let log = arrives_behind_a_lane(&lab, &call);
    assert!(log.contains("arrival order (no session id)"), "{log}");
}

/// No registered leader at all: arrival order, and the run is not refused.
#[test]
fn front_ac3_no_registered_leader_queues_in_arrival_order() {
    let lab = Lab::new(0);
    let log = arrives_behind_a_lane(&lab, &Call::leader());
    assert!(log.contains("arrival order"), "{log}");
}

/// The leader session's run without `--full` is a lane run.
#[test]
fn front_ac3_a_leader_run_without_full_queues_in_arrival_order() {
    let lab = Lab::new(0);
    lab.register_leader();
    let call = Call {
        full: false,
        ..Call::leader()
    };
    let log = arrives_behind_a_lane(&lab, &call);
    assert!(
        !log.contains("front place"),
        "a run without --full took the front: {log}"
    );
}

/// A second `leader` registration over a live entry is refused, so no lane takes the name.
#[test]
fn front_ac3_a_second_leader_registration_over_a_live_entry_is_refused() {
    let lab = Lab::new(0);
    lab.register_leader();
    let o = lab.register(OTHER, "leader", &["--mode", "leader"]);
    assert!(!o.status.success(), "{o:?}");
    // The entry still names the first session: its run still takes the front.
    let slot = lab.hold(0);
    let l1 = lab.lane("L1", "lane-1", None);
    let (a, logs) = lab.check("A", &Call::leader());
    assert!(
        text(&logs.join("heavy-slot.log")).contains("waiting 1 of 2 for 1 slot(s)"),
        "{}",
        text(&logs.join("heavy-slot.log"))
    );
    drop(slot);
    ok(a);
    ok(l1);
}

// ---- AC-4 (REQ-7) ----------------------------------------------------------------------------

/// The log of a `--full` run names the front place, or arrival order with its reason; a run
/// without `--full` writes neither; the one slot record is still the only `heavy-slot: slot ` line.
#[test]
fn front_ac4_the_log_names_the_front_place_or_the_arrival_order_and_why() {
    let lab = Lab::new(0);
    lab.register_leader();
    let run_log = |tag: &str, call: &Call| -> String {
        let out_dir = lab.work.path().join(format!("{tag}.logs"));
        fs::create_dir_all(&out_dir).unwrap();
        let o = lab.check_command(tag, call, &out_dir).output().unwrap();
        assert!(o.status.success(), "{o:?}");
        text(&out_dir.join("heavy-slot.log"))
    };
    let front = run_log("F", &Call::leader());
    assert!(front.contains("front place"), "{front}");
    let no_id = run_log(
        "N",
        &Call {
            sid: None,
            ..Call::leader()
        },
    );
    assert!(no_id.contains("arrival order (no session id)"), "{no_id}");
    let other = run_log(
        "O",
        &Call {
            sid: Some(OTHER),
            ..Call::leader()
        },
    );
    assert!(other.contains("arrival order ("), "{other}");
    assert!(
        !other.contains("refused:"),
        "the registry refusal text was copied: {other}"
    );
    let bare = run_log(
        "B",
        &Call {
            full: false,
            ..Call::leader()
        },
    );
    assert!(
        !bare.contains("front place") && !bare.contains("arrival order"),
        "{bare}"
    );
    for log in [&front, &no_id, &other, &bare] {
        assert_eq!(log.matches("heavy-slot: slot ").count(), 1, "{log}");
    }
}
