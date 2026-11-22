//! `moai card reclaim` before the lease ends (SPEC-20261120-reclaim-dead-holder): a reservation
//! whose holder session is proven gone is taken over while its lease still runs, and only then.
//! The proof has three parts: the liveness lookup of the holder session answers launcher-exited,
//! the recorded worktree is gone or clean, and no process has it as its working folder.
//!
//! Every command runs against a temporary HOME and a temporary git repository that holds the
//! session registry; the real `~/.moai` is not read or written. The holder is the session id in
//! the environment (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

mod common;

use common::git::git_in;
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const REASON: &str = "the holder's terminal is gone";
const NAME: &str = "agent-1";

/// A well-formed session id for number `n`.
fn sid(n: u64) -> String {
    format!("{n:08x}-0000-4000-8000-{n:012x}")
}

struct World {
    home: TempHome,
    host: tempfile::TempDir,
    project: PathBuf,
    /// A worktree with its work pushed: clean.
    wt: PathBuf,
    path_env: std::ffi::OsString,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    // The project is a git repository: it holds the session registry.
    git_in(&project, &["init", "-q", "-b", "main"]);
    let host = tempfile::Builder::new()
        .prefix("moai-test-host-")
        .tempdir()
        .unwrap();
    let base = host.path();
    let bare = base.join("remote.git");
    fs::create_dir_all(&bare).unwrap();
    git_in(&bare, &["init", "-q", "--bare", "-b", "main"]);
    let wt = base.join("wt");
    fs::create_dir_all(&wt).unwrap();
    git_in(&wt, &["init", "-q", "-b", "main"]);
    fs::write(wt.join("a.txt"), "a\n").unwrap();
    git_in(&wt, &["add", "a.txt"]);
    git_in(&wt, &["commit", "-q", "-m", "first"]);
    git_in(&wt, &["remote", "add", "origin", bare.to_str().unwrap()]);
    git_in(&wt, &["push", "-q", "origin", "main"]);
    // The command runs with a clean environment; the registry lookup spawns `git`, so the
    // directory of this runner's `git` goes back into a minimal PATH.
    let git_dir = std::env::var_os("PATH")
        .and_then(|p| {
            std::env::split_paths(&p)
                .find(|d| d.join("git").is_file())
                .map(|d| d.into_os_string())
        })
        .unwrap_or_default();
    let mut path_env = git_dir;
    path_env.push(":/usr/bin:/bin");
    World {
        home,
        host,
        project,
        wt,
        path_env,
    }
}

impl World {
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        let project = self.project.to_str().unwrap();
        all.extend_from_slice(&["--project", project]);
        let mut cmd = self.home.command(&all);
        cmd.env("PATH", &self.path_env);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    fn sessions_dir(&self) -> PathBuf {
        self.project.join(".git/.moai-cli/sessions")
    }

    /// `moai session-register` for `name`, as the host session `session` (none: no id).
    fn register(&self, name: &str, session: Option<&str>) {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.arg("session-register")
            .arg(&self.project)
            .args(["--name", name])
            .env("HOME", self.home.path())
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env_remove(SESSION_VAR)
            .env_remove("CODEX_THREAD_ID");
        if let Some(s) = session {
            cmd.env(SESSION_VAR, s);
        }
        let out = cmd.output().unwrap();
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stdout)
        );
    }

    /// The state of a session whose launcher really ran and has exited: registered under
    /// `name`, a `.live` file nobody locks, and a registry lease that has ended. Nothing but
    /// this state is a proof that a session is gone (audit of the early reclaim, B1).
    fn register_gone(&self, name: &str, session: &str) {
        self.register(name, Some(session));
        fs::write(self.sessions_dir().join(format!("{name}.live")), b"").unwrap();
        self.set_registry_lease(name, -60);
    }

    /// Moves the registry lease of `name` to `offset` seconds from now (negative: ended).
    fn set_registry_lease(&self, name: &str, offset: i64) {
        let file = self.sessions_dir().join(format!("{name}.json"));
        let mut entry: Value = serde_json::from_str(&fs::read_to_string(&file).unwrap()).unwrap();
        let now = self.rows("SELECT unixepoch()")[0][0].as_i64().unwrap();
        entry["lease_until"] = json!(now + offset);
        fs::write(&file, serde_json::to_vec(&entry).unwrap()).unwrap();
    }

    /// A ready card taken by `holder` with the worktree `worktree` (none: no worktree).
    fn taken(&self, holder: &str, worktree: Option<&Path>, scope: &str) -> String {
        self.taken_on(holder, worktree, scope, "main")
    }

    /// The same, with the reservation naming the branch `branch`.
    fn taken_on(&self, holder: &str, worktree: Option<&Path>, scope: &str, branch: &str) -> String {
        let added = result(&self.card(None, &["add", "--title", "t", "--scope", scope]));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        let wt = worktree.map(|w| w.to_str().unwrap().to_string());
        let mut args = vec!["take", name.as_str()];
        if let Some(w) = wt.as_deref() {
            args.extend_from_slice(&["--worktree", w, "--branch", branch]);
        }
        result(&self.card(Some(holder), &args));
        name
    }

    fn reclaim_with(&self, session: &str, card: &str, generation: &str, reason: &str) -> Output {
        self.card(
            Some(session),
            &[
                "reclaim",
                card,
                "--generation",
                generation,
                "--by",
                "leader",
                "--reason",
                reason,
            ],
        )
    }

    fn reclaim(&self, session: &str, card: &str, generation: &str) -> Output {
        self.reclaim_with(session, card, generation, REASON)
    }

    fn rows(&self, sql: &str) -> Vec<Value> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l.conn().prepare(sql).unwrap();
        let n = st.column_count();
        st.query_map([], |r| {
            Ok(Value::Array(
                (0..n)
                    .map(|i| match r.get_ref(i).unwrap() {
                        rusqlite::types::ValueRef::Integer(v) => json!(v),
                        rusqlite::types::ValueRef::Text(t) => {
                            json!(String::from_utf8_lossy(t).to_string())
                        }
                        _ => Value::Null,
                    })
                    .collect(),
            ))
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
    }

    /// Every row of the tables a reclaim could touch, as text.
    fn dump(&self) -> Vec<String> {
        let mut out = Vec::new();
        for table in ["card", "reservation", "reservation_scope", "event"] {
            for r in self.rows(&format!("SELECT * FROM {table} ORDER BY 1, 2")) {
                out.push(format!("{table}:{r}"));
            }
        }
        out
    }

    /// The registry files with their bytes: what the proof must leave as it was.
    fn registry_bytes(&self) -> Vec<(String, Vec<u8>)> {
        let mut v = Vec::new();
        if let Ok(rd) = fs::read_dir(self.sessions_dir()) {
            for e in rd.flatten() {
                let p = e.path();
                if p.is_file() {
                    v.push((
                        e.file_name().to_string_lossy().into_owned(),
                        fs::read(&p).unwrap_or_default(),
                    ));
                }
            }
        }
        v.sort();
        v
    }

    fn refused_and_unchanged(&self, out: Output, needle: &str, before: &[String]) -> String {
        let (kind, message) = refusal(&out);
        assert_eq!(kind, "conflict", "{message}");
        assert!(message.contains(needle), "{needle}: {message}");
        assert_eq!(&self.dump(), before, "a refusal wrote something");
        message
    }
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn refusal(out: &Output) -> (String, String) {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

fn sleeper_in(dir: &Path) -> Child {
    Command::new("sleep")
        .arg("30")
        .current_dir(dir)
        .stdin(Stdio::null())
        .spawn()
        .expect("start sleep")
}

/// AC-1 (REQ-1, REQ-5): the holder's launcher has exited, the worktree is clean and nobody is
/// in it: the live lease is reclaimed, the new holder gets the next generation with the same
/// scope, worktree and branch, and the reason carries the proof.
#[test]
fn early_reclaim_succeeds() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/early/");
    w.register_gone(NAME, &sid(1));
    let lease_before = w.rows("SELECT lease_until FROM reservation WHERE generation = 1");
    let now = w.rows("SELECT unixepoch()")[0][0].as_i64().unwrap();
    assert!(
        lease_before[0][0].as_i64().unwrap() > now + 3000,
        "the lease is live: {lease_before:?}"
    );

    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
    assert_eq!(got["reservation"]["holder"], json!(sid(2)));
    assert_eq!(got["reservation"]["status"], json!("held"));
    assert_eq!(got["reservation"]["write_scope"], json!(["src/early/"]));
    assert_eq!(got["previous"]["status"], json!("reclaimed"));
    let reason = got["reason"].as_str().unwrap();
    assert!(reason.starts_with(REASON), "{reason}");
    assert!(
        reason.contains(
            "early reclaim: launcher exited (session agent-1); worktree clean; no process in it"
        ),
        "{reason}"
    );
}

/// A worktree that is gone is a proof too; the reason says so.
#[test]
fn early_reclaim_with_a_gone_worktree() {
    let w = world();
    let gone = w.host.path().join("not-here");
    let card = w.taken(&sid(1), Some(&gone), "src/gone/");
    w.register_gone(NAME, &sid(1));
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    let reason = got["reason"].as_str().unwrap();
    assert!(
        reason.contains("worktree gone; no process in it"),
        "{reason}"
    );
}

/// AC-2 (REQ-2): an alive answer and an unknown answer each leave the live lease refused, with
/// nothing written and the registry as it was.
#[test]
fn refused_without_the_liveness_proof() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/alive/");
    let before = w.dump();

    // Unknown: no registry entry names the holder session.
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "lease", &before);
    assert!(m.contains("liveness is unknown"), "{m}");

    // Unknown: an entry exists but it carries another session's id.
    w.register(NAME, Some(&sid(9)));
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "lease", &before);
    assert!(m.contains("liveness is unknown"), "{m}");

    // Unknown: an entry without any session id.
    w.register("agent-2", None);
    w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "lease", &before);

    // Alive: the entry names the holder session and its launcher holds the lock.
    w.register("agent-3", Some(&sid(1)));
    let held = moai_cli::session::lock_named_live(&w.sessions_dir(), "agent-3").unwrap();
    let bytes = w.registry_bytes();
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "lease", &before);
    assert!(m.contains("launcher is still running"), "{m}");
    assert_eq!(
        w.registry_bytes(),
        bytes,
        "the proof wrote into the registry"
    );
    drop(held);
}

/// AC-3 (REQ-3): a worktree with an uncommitted change, one with an unpushed commit and a
/// reservation that records no worktree each leave the live lease refused.
#[test]
fn refused_without_the_worktree_proof() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/dirty/");
    w.register_gone(NAME, &sid(1));
    let before = w.dump();

    fs::write(w.wt.join("new.txt"), "x\n").unwrap();
    w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "worktree dirty", &before);
    git_in(&w.wt, &["add", "new.txt"]);
    git_in(&w.wt, &["commit", "-q", "-m", "second"]);
    w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "1 unpushed", &before);
    git_in(&w.wt, &["push", "-q", "origin", "main"]);
    // Pushed and clean again: the same reclaim now goes through.
    result(&w.reclaim(&sid(2), &card, "1"));

    // No worktree recorded: nothing shows that no work is left.
    let other = w.taken(&sid(3), None, "src/nowt/");
    w.register_gone("agent-4", &sid(3));
    let before = w.dump();
    let m = w.refused_and_unchanged(w.reclaim(&sid(4), &other, "1"), "lease", &before);
    assert!(m.contains("records no worktree"), "{m}");
}

/// AC-4 (REQ-4): a process whose working folder is inside the worktree leaves the live lease
/// refused; a process in a folder that only starts with the same name does not.
#[test]
fn process_in_the_worktree() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/proc/");
    w.register_gone(NAME, &sid(1));
    let before = w.dump();

    let mut inside = sleeper_in(&w.wt);
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "lease", &before);
    assert!(m.contains("running process"), "{m}");
    let _ = inside.kill();
    let _ = inside.wait();

    let sibling = w.host.path().join("wt-2");
    fs::create_dir_all(&sibling).unwrap();
    let mut beside = sleeper_in(&sibling);
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
    let _ = beside.kill();
    let _ = beside.wait();
}

/// AC-10 (REQ-12): a worktree whose only untracked file is a listed tool cache (the statusline's
/// GitHub counts) counts as clean; any other untracked file, an unlisted name in the same
/// folder and a changed tracked file each keep it dirty.
#[test]
fn a_listed_tool_cache_is_not_work() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/cache/");
    w.register_gone(NAME, &sid(1));
    let before = w.dump();

    // The state folder comes from the resolver (leader rule: tests never spell it).
    let cache_dir = moai_cli::state_root::path(&w.wt, "state/github");
    fs::create_dir_all(&cache_dir).unwrap();
    fs::write(cache_dir.join("counts.json"), "{\"open_issues\":1}\n").unwrap();

    // The cache plus one real untracked file: the file is work, the count names only it.
    fs::write(w.wt.join("new.txt"), "x\n").unwrap();
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "worktree dirty", &before);
    assert!(m.contains("holds 1 uncommitted change(s)"), "{m}");
    fs::remove_file(w.wt.join("new.txt")).unwrap();

    // An untracked file beside the state folder, not under it, is work (REQ-4).
    let beside = moai_cli::state_root::path(&w.wt, "other.json");
    fs::write(&beside, "{}\n").unwrap();
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "worktree dirty", &before);
    assert!(m.contains("holds 1 uncommitted change(s)"), "{m}");
    fs::remove_file(&beside).unwrap();

    // A tracked file that changed is work even when a listed cache sits beside it.
    fs::write(w.wt.join("a.txt"), "changed\n").unwrap();
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "worktree dirty", &before);
    assert!(m.contains("holds 1 uncommitted change(s)"), "{m}");
    git_in(&w.wt, &["checkout", "--", "a.txt"]);

    // Only the listed cache is left: the reclaim goes through and the cache file stays.
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
    assert!(
        cache_dir.join("counts.json").exists(),
        "the proof deleted a file"
    );
}

/// SPEC REQ-3 (AC-2): every untracked file under the runtime state folder is tool state, not
/// work: the counts file, another file beside it, a file in a folder below, and the earlier
/// generation's state folder. The reclaim goes through and no file is touched.
#[test]
fn a_runtime_state_folder_is_not_work() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/cache/");
    w.register_gone(NAME, &sid(1));
    // The current state folder from the resolver (read before any file is made), and the
    // earlier generation's `.moai/state/` through its constant.
    let state = |sub: &str| moai_cli::state_root::rel(&w.wt, sub);
    let files = [
        state("state/github/counts.json"),
        state("state/github/other.json"),
        state("state/deep/er/inner.txt"),
        format!("{}/state/earlier.json", moai_cli::state_root::MOAI_DIR),
    ];
    for f in &files {
        let p = w.wt.join(f);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(&p, "{}\n").unwrap();
    }
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
    for f in &files {
        assert!(w.wt.join(f).exists(), "the proof deleted {f}");
    }
}

/// AC-5 (REQ-5, REQ-7): the event keeps its kind, carries the proof in its own field, and the
/// ledger tables and the schema version are what they were; an ordinary reclaim of an ended
/// lease carries no such field.
#[test]
fn event_names_the_proof() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/event/");
    w.register_gone(NAME, &sid(1));
    let schema = |w: &World| {
        (
            w.rows("SELECT type, name FROM sqlite_master ORDER BY 1, 2"),
            w.rows("PRAGMA user_version"),
        )
    };
    let before = schema(&w);
    result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(schema(&w), before, "the schema is untouched");

    let rows = w.rows("SELECT kind, payload FROM event WHERE kind LIKE '%reclaim%'");
    assert_eq!(rows.len(), 1, "one reclaim event: {rows:?}");
    assert_eq!(rows[0][0], json!("card.reclaim"));
    let payload: Value = serde_json::from_str(rows[0][1].as_str().unwrap()).unwrap();
    assert_eq!(payload["from_holder"], json!(sid(1)));
    assert!(
        payload["reason"]
            .as_str()
            .unwrap()
            .contains("early reclaim: launcher exited"),
        "{payload}"
    );
    assert_eq!(
        payload["early_proof"],
        json!({
            "liveness": "launcher-exited",
            "session_name": NAME,
            "live_file": true,
            "registry_lease_ended": true,
            "worktree": "clean",
            "processes_in_worktree": 0
        }),
        "{payload}"
    );

    // An ended lease needs no proof and the event has no such field.
    let other = w.taken(&sid(5), Some(&w.wt), "src/ended/");
    let mut l = ledger::open_write(w.home.path(), "demo").unwrap();
    l.with_write_tx(|tx, _| {
        tx.execute_batch(
            "UPDATE reservation SET lease_until = unixepoch() - 5 \
             WHERE status = 'held' AND holder = '00000005-0000-4000-8000-000000000005'",
        )?;
        Ok(())
    })
    .unwrap();
    drop(l);
    result(&w.reclaim(&sid(6), &other, "1"));
    let rows = w.rows(
        "SELECT payload FROM event WHERE kind = 'card.reclaim' ORDER BY event_id DESC LIMIT 1",
    );
    let payload: Value = serde_json::from_str(rows[0][0].as_str().unwrap()).unwrap();
    assert!(payload.get("early_proof").is_none(), "{payload}");
}

/// AC-7 (REQ-7): an early reclaim counts toward the limit of three, so a fourth is refused with
/// the limit, and the new lease is the four-hour default.
#[test]
fn early_reclaim_counts() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/limit/");
    for n in 1..=4u64 {
        w.register_gone(&format!("agent-{n}"), &sid(n));
    }
    for n in 1..=3u64 {
        let g = n.to_string();
        let got = result(&w.reclaim(&sid(n + 1), &card, &g));
        assert_eq!(got["reservation"]["generation"], json!(n + 1));
    }
    let lease = w.rows("SELECT lease_until - acquired_at FROM reservation WHERE generation = 4");
    assert_eq!(lease[0][0], json!(14_400), "the lease default is unchanged");
    let before = w.dump();
    let (kind, message) = refusal(&w.reclaim(&sid(5), &card, "4"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("reclaim limit"), "{message}");
    assert_eq!(w.dump(), before, "the fourth reclaim wrote something");
}

/// Its own live lease is no early reclaim: the holder is not "gone".
#[test]
fn own_live_lease_is_not_an_early_reclaim() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/own/");
    w.register_gone(NAME, &sid(1));
    let before = w.dump();
    w.refused_and_unchanged(w.reclaim(&sid(1), &card, "1"), "lease", &before);
}

/// AC-9 (REQ-11): a reason that fits alone but not with the proof is refused before any write;
/// one that fits exactly is taken.
#[test]
fn reason_too_long() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/long/");
    w.register_gone(NAME, &sid(1));
    let suffix =
        " [early reclaim: launcher exited (session agent-1); worktree clean; no process in it]";
    let room = 500 - suffix.chars().count();
    let before = w.dump();
    let too_long = "a".repeat(room + 1);
    let (kind, message) = refusal(&w.reclaim_with(&sid(2), &card, "1", &too_long));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("early-reclaim proof"), "{message}");
    assert_eq!(w.dump(), before, "a refusal wrote something");
    let fits = "a".repeat(room);
    let got = result(&w.reclaim_with(&sid(2), &card, "1", &fits));
    assert_eq!(got["reason"].as_str().unwrap().chars().count(), 500);
}

/// AC-6 (REQ-6): the ledger transaction of an early reclaim refuses a proof that was made for
/// another holder than the current one and writes nothing; with the right holder the live
/// lease is taken and the event keeps the proof.
#[test]
fn reclaim_early_holder_must_match() {
    use moai_cli::ledger::reservation::{self, EarlyProof, ReclaimError};
    let home = TempHome::new();
    ledger::register(home.path(), "demo").unwrap();
    let mut w = ledger::open_write(home.path(), "demo").unwrap();
    let card = w
        .with_write_tx(|tx, now| {
            tx.execute(
                "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                 VALUES ('card', 'ready', 'test', ?1, ?1)",
                rusqlite::params![now],
            )?;
            Ok(tx.last_insert_rowid())
        })
        .unwrap();
    reservation::reserve(
        &mut w,
        card,
        "holder-a",
        &["zone/".to_string()],
        3600,
        None,
        None,
    )
    .unwrap();
    let detail = json!({"liveness": "launcher-exited"});
    let count = |w: &mut ledger::WriteLedger| -> i64 {
        w.with_write_tx(|tx, _| Ok(tx.query_row("SELECT COUNT(*) FROM event", [], |r| r.get(0))?))
            .unwrap()
    };
    let events = count(&mut w);

    let wrong = EarlyProof {
        holder: "holder-b",
        detail: &detail,
    };
    let err = reservation::reclaim_early(&mut w, card, "heir", 1, "leader", "r", 3600, &wrong)
        .unwrap_err();
    assert!(matches!(err, ReclaimError::HolderChanged { .. }), "{err}");
    assert_eq!(err.code(), "holder_changed");
    assert_eq!(count(&mut w), events, "a refusal wrote an event");

    let right = EarlyProof {
        holder: "holder-a",
        detail: &detail,
    };
    let got =
        reservation::reclaim_early(&mut w, card, "heir", 1, "leader", "r", 3600, &right).unwrap();
    assert_eq!(got.reservation.generation, 2);
    assert_eq!(got.reservation.holder, "heir");
    assert_eq!(count(&mut w), events + 1);
    // Without a proof the live lease of the new holder is still refused.
    let plain = reservation::reclaim(&mut w, card, "other", 2, "leader", "r", 3600).unwrap_err();
    assert!(matches!(plain, ReclaimError::LeaseLive { .. }), "{plain}");
    assert_eq!(count(&mut w), events + 1, "a refusal wrote an event");
}

/// REQ-1 (audit of the early reclaim, B1): a session registered with `session-register` alone
/// has no `.live` file, and the lookup answers `launcher-exited` for it; that is no proof. A
/// session that is running, with a clean and pushed worktree, keeps its reservation. A `.live`
/// file nobody locks is still not enough while the registry lease runs; both together are the
/// dead state, and only that is reclaimed.
#[test]
fn a_session_without_a_launcher_record_is_not_gone() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/live/");
    let before = w.dump();

    // The probe of the audit: a running lease, no `.live` file, a clean pushed worktree.
    w.register(NAME, Some(&sid(1)));
    let bytes = w.registry_bytes();
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "lease", &before);
    assert!(m.contains("no launcher lock file"), "{m}");
    assert_eq!(
        w.registry_bytes(),
        bytes,
        "the proof wrote into the registry"
    );

    // The lock file exists and nobody holds it, but the registry lease has not ended.
    fs::write(w.sessions_dir().join(format!("{NAME}.live")), b"").unwrap();
    let bytes = w.registry_bytes();
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "lease", &before);
    assert!(m.contains("registry lease"), "{m}");
    assert!(m.contains("has not ended"), "{m}");
    assert_eq!(
        w.registry_bytes(),
        bytes,
        "the proof wrote into the registry"
    );

    // The lease ends too: the dead state, reclaimed.
    w.set_registry_lease(NAME, -1);
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
}

/// REQ-3 (audit of the early reclaim, N1): work the takeover gate leaves out keeps the live
/// lease refused: a stash entry of the worktree's branch and an ignored entry that is not on
/// the regenerable list. The regenerable ignored entries, a stash entry of another branch, an
/// unrelated local branch with unpushed commits that is checked out nowhere and another
/// worktree's own unpushed branch are no work (the reservation's own branch is the next test).
#[test]
fn work_the_takeover_gate_misses_keeps_the_early_reclaim_refused() {
    let w = world();
    fs::write(w.wt.join(".gitignore"), "scratch/\n.moai/logs/\ntarget/\n").unwrap();
    git_in(&w.wt, &["add", ".gitignore"]);
    git_in(&w.wt, &["commit", "-q", "-m", "ignore list"]);
    git_in(&w.wt, &["push", "-q", "origin", "main"]);
    let card = w.taken(&sid(1), Some(&w.wt), "src/work/");
    w.register_gone(NAME, &sid(1));
    let before = w.dump();

    // Regenerable ignored entries sit there the whole time and block nothing.
    for dir in ["target", ".moai/logs"] {
        fs::create_dir_all(w.wt.join(dir)).unwrap();
        fs::write(w.wt.join(dir).join("out.txt"), "x\n").unwrap();
    }

    // A stash entry of this worktree's branch.
    fs::write(w.wt.join("a.txt"), "draft\n").unwrap();
    git_in(&w.wt, &["stash", "push", "-q", "-m", "draft"]);
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "stash", &before);
    assert!(m.contains("branch main"), "{m}");
    git_in(&w.wt, &["stash", "drop", "-q"]);

    // An unrelated local branch with an unpushed commit, checked out nowhere: not this
    // reservation's work. A stash entry of that other branch is not either.
    git_in(&w.wt, &["checkout", "-q", "-b", "other"]);
    fs::write(w.wt.join("o.txt"), "o\n").unwrap();
    git_in(&w.wt, &["add", "o.txt"]);
    git_in(&w.wt, &["commit", "-q", "-m", "on other"]);
    fs::write(w.wt.join("o.txt"), "o draft\n").unwrap();
    git_in(&w.wt, &["stash", "push", "-q", "-m", "draft of other"]);
    git_in(&w.wt, &["checkout", "-q", "main"]);

    // An ignored file that is not on the regenerable list.
    fs::create_dir_all(w.wt.join("scratch")).unwrap();
    fs::write(w.wt.join("scratch/draft.md"), "x\n").unwrap();
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "ignored entry", &before);
    assert!(m.contains("scratch/"), "{m}");
    fs::remove_file(w.wt.join("scratch/draft.md")).unwrap();
    fs::remove_dir(w.wt.join("scratch")).unwrap();

    // Another worktree's own branch with an unpushed commit is that worktree's work.
    let side = w.host.path().join("wt-side");
    git_in(
        &w.wt,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "side",
            side.to_str().unwrap(),
        ],
    );
    fs::write(side.join("s.txt"), "s\n").unwrap();
    git_in(&side, &["add", "s.txt"]);
    git_in(&side, &["commit", "-q", "-m", "on side"]);

    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
}

/// N4 of the audit: the proof's own worktree check is what refuses a dirty worktree, not the
/// takeover gate that runs before it. Asked directly, without that gate, a clean worktree is
/// proven and the same worktree with an untracked file is not.
#[test]
fn the_early_proof_checks_the_worktree_by_itself() {
    use moai_cli::card::early_reclaim::{assess, Early};
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/own/");
    w.register_gone(NAME, &sid(1));
    let id: i64 = card.trim_start_matches("C-").parse().unwrap();

    match assess(&w.project, w.home.path(), id, &sid(2)) {
        Early::Proven(proof) => assert_eq!(proof.detail["worktree"], json!("clean")),
        other => panic!("a clean worktree is proven: {other:?}"),
    }
    fs::write(w.wt.join("draft.txt"), "x\n").unwrap();
    match assess(&w.project, w.home.path(), id, &sid(2)) {
        Early::Unproven(why) => assert!(why.contains("worktree dirty"), "{why}"),
        other => panic!("a dirty worktree is not proven: {other:?}"),
    }
}

// ---- the local integration branch (REQ-13) ----

/// A project profile with the given work flow and, when named, its integration branch.
fn flow_profile(flow: &str, integration: Option<&str>) -> String {
    let branch = integration.map_or(String::new(), |b| {
        format!(", \"integration_branch\": \"{b}\"")
    });
    format!(
        "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \
         \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \
         \"work_flow\": {{\"flow\": \"{flow}\"{branch}}}}}"
    )
}

/// One new commit in `dir` that adds the file `name`.
fn commit_file(dir: &Path, name: &str) {
    fs::write(dir.join(name), format!("{name}\n")).unwrap();
    git_in(dir, &["add", name]);
    git_in(dir, &["commit", "-q", "-m", name]);
}

impl World {
    /// Writes `profile` as the project profile of the worktree and commits and pushes it, so
    /// the worktree stays clean and nothing is ahead of the remote.
    fn commit_profile(&self, profile: &str) {
        let dir = self.wt.join(".moai-cli");
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("project.yaml"), profile).unwrap();
        git_in(&self.wt, &["add", ".moai-cli/project.yaml"]);
        git_in(&self.wt, &["commit", "-q", "-m", "profile"]);
        git_in(&self.wt, &["push", "-q", "origin", "main"]);
    }

    /// `n` commits on a new branch `lane`, which the worktree has checked out afterwards.
    fn lane_commits(&self, n: u32) {
        git_in(&self.wt, &["checkout", "-q", "-b", "lane"]);
        for i in 1..=n {
            commit_file(&self.wt, &format!("lane-{i}.txt"));
        }
    }

    /// How many commits of HEAD are on no remote-tracking ref.
    fn ahead_of_remotes(&self) -> String {
        common::git::git(
            &self.wt,
            &["rev-list", "--count", "HEAD", "--not", "--remotes"],
        )
    }
}

/// AC-13 (REQ-13): the field case. A clean worktree with four commits ahead of the remote, all
/// four already in the local integration branch the m-flow profile names, has no work left:
/// the takeover gate and the early proof both let the live lease be reclaimed.
#[test]
fn commits_inside_the_integration_branch_do_not_stop_an_early_reclaim() {
    use moai_cli::card::early_reclaim::{assess, Early};
    let w = world();
    w.commit_profile(&flow_profile("m-flow", Some("integ")));
    let card = w.taken(&sid(1), Some(&w.wt), "src/landed/");
    w.register_gone(NAME, &sid(1));
    w.lane_commits(4);
    git_in(&w.wt, &["branch", "integ", "HEAD"]);
    assert_eq!(
        w.ahead_of_remotes(),
        "4",
        "four commits ahead of the remote"
    );
    let id: i64 = card.trim_start_matches("C-").parse().unwrap();

    match assess(&w.project, w.home.path(), id, &sid(2)) {
        Early::Proven(proof) => assert_eq!(proof.detail["worktree"], json!("clean")),
        other => panic!("commits inside the integration branch are no work: {other:?}"),
    }
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
}

/// AC-13 (REQ-13), N and N+1: four contained commits pass, a fifth that is not in the
/// integration branch refuses and the count names exactly that one; once it lands there too the
/// same reclaim goes through.
#[test]
fn one_commit_outside_the_integration_branch_still_stops_an_early_reclaim() {
    let w = world();
    w.commit_profile(&flow_profile("m-flow", Some("integ")));
    let card = w.taken(&sid(1), Some(&w.wt), "src/outside/");
    w.register_gone(NAME, &sid(1));
    w.lane_commits(4);
    git_in(&w.wt, &["branch", "integ", "HEAD"]);
    commit_file(&w.wt, "lane-5.txt");
    assert_eq!(w.ahead_of_remotes(), "5");
    let before = w.dump();

    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "worktree dirty", &before);
    assert!(
        m.contains("holds 0 uncommitted change(s) and 1 unpushed commit(s)"),
        "{m}"
    );

    git_in(&w.wt, &["branch", "-f", "integ", "HEAD"]);
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
}

/// AC-13 (REQ-13): without a usable integration branch every commit that is on no
/// remote-tracking ref stays unpushed work: no profile, another work flow, m-flow without a
/// branch, a branch the clone does not have and a profile that cannot be read.
#[test]
fn without_a_usable_integration_branch_an_early_reclaim_counts_every_unpushed_commit() {
    let cases: [(&str, Option<String>); 6] = [
        ("no profile", None),
        ("another work flow", Some(flow_profile("github-flow", None))),
        // The branch is named and exists locally, but the flow is not m-flow: only the flow
        // check keeps this branch out of the count.
        (
            "another work flow that names an integration branch",
            Some(flow_profile("github-flow", Some("integ"))),
        ),
        (
            "m-flow without a branch",
            Some(flow_profile("m-flow", None)),
        ),
        (
            "m-flow with a branch that is not local",
            Some(flow_profile("m-flow", Some("absent"))),
        ),
        ("an unreadable profile", Some("{not json".to_string())),
    ];
    for (label, profile) in cases {
        eprintln!("case: {label}");
        let w = world();
        if let Some(p) = &profile {
            w.commit_profile(p);
        }
        let card = w.taken(&sid(1), Some(&w.wt), "src/unusable/");
        w.register_gone(NAME, &sid(1));
        w.lane_commits(4);
        // A local branch with those commits exists, but the profile does not name it.
        git_in(&w.wt, &["branch", "integ", "HEAD"]);
        let before = w.dump();
        let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "worktree dirty", &before);
        assert!(
            m.contains("holds 0 uncommitted change(s) and 4 unpushed commit(s)"),
            "{label}: {m}"
        );
    }
}

/// AC-12, AC-13 (REQ-3, REQ-13): the branch the reservation names, which the worktree does not
/// have checked out, is checked with the same rule as HEAD: unpushed commits on it that are not
/// in the integration branch keep the lease refused and the refusal names the branch; once they
/// are inside the integration branch the reclaim goes through.
#[test]
fn the_reserved_branch_ahead_only_inside_the_integration_branch_is_not_work() {
    let w = world();
    w.commit_profile(&flow_profile("m-flow", Some("integ")));
    let card = w.taken_on(&sid(1), Some(&w.wt), "src/branches/", "side");
    w.register_gone(NAME, &sid(1));
    // `integ` holds two commits that no remote has; `side` is `integ` plus one more commit.
    git_in(&w.wt, &["checkout", "-q", "-b", "integ"]);
    commit_file(&w.wt, "integ-1.txt");
    commit_file(&w.wt, "integ-2.txt");
    git_in(&w.wt, &["branch", "side", "integ"]);
    git_in(&w.wt, &["checkout", "-q", "side"]);
    commit_file(&w.wt, "side-1.txt");
    git_in(&w.wt, &["checkout", "-q", "main"]);
    assert_eq!(w.ahead_of_remotes(), "0", "the worktree itself is pushed");
    let before = w.dump();

    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "branch side", &before);
    assert!(m.contains("1 unpushed commit(s)"), "{m}");

    // The commit lands in the integration branch: the reserved branch is contained in it.
    git_in(&w.wt, &["branch", "-f", "integ", "side"]);
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
}

/// AC-12 (REQ-3): a clone has many local branches that belong to nobody here (backups, old
/// work). None of them is this reservation's work, so they never block an early reclaim of a
/// clean, dead-holder worktree, with an m-flow profile or without one.
#[test]
fn an_unrelated_local_branch_with_unpushed_commits_does_not_stop_an_early_reclaim() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/unrelated/");
    w.register_gone(NAME, &sid(1));
    for name in ["backup-1", "old-2", "wip-3"] {
        git_in(&w.wt, &["checkout", "-q", "-b", name]);
        commit_file(&w.wt, &format!("{name}.txt"));
        git_in(&w.wt, &["checkout", "-q", "main"]);
    }
    assert_eq!(w.ahead_of_remotes(), "0", "the worktree itself is pushed");
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
}

// A lane worktree has no state-root marker (the marker is ignored by git), so its own profile
// is the earlier `.moai-cli/project.yaml`, which has no work flow; the m-flow profile with the
// integration branch is the project folder's. REQ-13 reads the project folder first.

impl World {
    /// Gives the project folder (not the worktree) the m-flow profile with `integration` as its
    /// integration branch, at the state root a marked project reads: the marker file and
    /// `.moai/project.yaml`. Written last, after the card commands that may rewrite a profile.
    fn mflow_project(&self, integration: &str) {
        let moai = self.project.join(".moai");
        fs::create_dir_all(&moai).unwrap();
        fs::write(
            moai.join("moai.json"),
            moai_cli::state_root::marker_bytes("test", &[]),
        )
        .unwrap();
        fs::write(
            moai.join("project.yaml"),
            flow_profile("m-flow", Some(integration)),
        )
        .unwrap();
    }
}

/// The worktree's own profile: valid, and without a work flow.
fn plain_profile() -> String {
    "{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \
     \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]}"
        .to_string()
}

/// AC-13 (REQ-13): the shape of a real lane. The m-flow profile with the integration branch is
/// the project folder's, the lane worktree's own profile has no work flow, and four commits
/// ahead of the remote are all inside the local integration branch: the takeover gate and the
/// early proof both let the live lease be reclaimed.
#[test]
fn the_project_folders_integration_branch_counts_when_the_worktree_profile_has_no_work_flow() {
    use moai_cli::card::early_reclaim::{assess, Early};
    let w = world();
    w.commit_profile(&plain_profile());
    let card = w.taken(&sid(1), Some(&w.wt), "src/projectflow/");
    w.register_gone(NAME, &sid(1));
    w.mflow_project("integ");
    w.lane_commits(4);
    git_in(&w.wt, &["branch", "integ", "HEAD"]);
    assert_eq!(
        w.ahead_of_remotes(),
        "4",
        "four commits ahead of the remote"
    );
    let id: i64 = card.trim_start_matches("C-").parse().unwrap();

    match assess(&w.project, w.home.path(), id, &sid(2)) {
        Early::Proven(proof) => assert_eq!(proof.detail["worktree"], json!("clean")),
        other => panic!("commits inside the project's integration branch are no work: {other:?}"),
    }
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
}

/// AC-13 (REQ-13), N and N+1 for the same shape: four contained commits pass, one more that is
/// outside the integration branch refuses and the count names exactly that one.
#[test]
fn one_commit_outside_the_project_folders_integration_branch_still_stops_an_early_reclaim() {
    let w = world();
    w.commit_profile(&plain_profile());
    let card = w.taken(&sid(1), Some(&w.wt), "src/projectoutside/");
    w.register_gone(NAME, &sid(1));
    w.mflow_project("integ");
    w.lane_commits(4);
    git_in(&w.wt, &["branch", "integ", "HEAD"]);
    commit_file(&w.wt, "lane-5.txt");
    let before = w.dump();

    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "worktree dirty", &before);
    assert!(
        m.contains("holds 0 uncommitted change(s) and 1 unpushed commit(s)"),
        "{m}"
    );

    git_in(&w.wt, &["branch", "-f", "integ", "HEAD"]);
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
}
