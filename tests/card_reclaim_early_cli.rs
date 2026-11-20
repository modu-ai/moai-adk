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
        assert!(out.status.success(), "{}", String::from_utf8_lossy(&out.stdout));
    }

    /// A ready card taken by `holder` with the worktree `worktree` (none: no worktree).
    fn taken(&self, holder: &str, worktree: Option<&Path>, scope: &str) -> String {
        let added = result(&self.card(None, &["add", "--title", "t", "--scope", scope]));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        let wt = worktree.map(|w| w.to_str().unwrap().to_string());
        let mut args = vec!["take", name.as_str()];
        if let Some(w) = wt.as_deref() {
            args.extend_from_slice(&["--worktree", w, "--branch", "main"]);
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
    w.register(NAME, Some(&sid(1)));
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
        reason.contains("early reclaim: launcher exited (session agent-1); worktree clean; no process in it"),
        "{reason}"
    );
}

/// A worktree that is gone is a proof too; the reason says so.
#[test]
fn early_reclaim_with_a_gone_worktree() {
    let w = world();
    let gone = w.host.path().join("not-here");
    let card = w.taken(&sid(1), Some(&gone), "src/gone/");
    w.register(NAME, Some(&sid(1)));
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    let reason = got["reason"].as_str().unwrap();
    assert!(reason.contains("worktree gone; no process in it"), "{reason}");
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
    assert_eq!(w.registry_bytes(), bytes, "the proof wrote into the registry");
    drop(held);
}

/// AC-3 (REQ-3): a worktree with an uncommitted change, one with an unpushed commit and a
/// reservation that records no worktree each leave the live lease refused.
#[test]
fn refused_without_the_worktree_proof() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/dirty/");
    w.register(NAME, Some(&sid(1)));
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
    w.register("agent-4", Some(&sid(3)));
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
    w.register(NAME, Some(&sid(1)));
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
    w.register(NAME, Some(&sid(1)));
    let before = w.dump();

    let cache_dir = w.wt.join(".moai-cli/state/github");
    fs::create_dir_all(&cache_dir).unwrap();
    fs::write(cache_dir.join("counts.json"), "{\"open_issues\":1}\n").unwrap();

    // The cache plus one real untracked file: the file is work, the count names only it.
    fs::write(w.wt.join("new.txt"), "x\n").unwrap();
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "worktree dirty", &before);
    assert!(m.contains("holds 1 uncommitted change(s)"), "{m}");
    fs::remove_file(w.wt.join("new.txt")).unwrap();

    // An unlisted name beside the cache is not a listed cache.
    fs::write(cache_dir.join("other.json"), "{}\n").unwrap();
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "worktree dirty", &before);
    assert!(m.contains("holds 1 uncommitted change(s)"), "{m}");
    fs::remove_file(cache_dir.join("other.json")).unwrap();

    // A tracked file that changed is work even when a listed cache sits beside it.
    fs::write(w.wt.join("a.txt"), "changed\n").unwrap();
    let m = w.refused_and_unchanged(w.reclaim(&sid(2), &card, "1"), "worktree dirty", &before);
    assert!(m.contains("holds 1 uncommitted change(s)"), "{m}");
    git_in(&w.wt, &["checkout", "--", "a.txt"]);

    // Only the listed cache is left: the reclaim goes through and the cache file stays.
    let got = result(&w.reclaim(&sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
    assert!(cache_dir.join("counts.json").exists(), "the proof deleted a file");
}

/// AC-5 (REQ-5, REQ-8): the event keeps its kind, carries the proof in its own field, and the
/// ledger tables and the schema version are what they were; an ordinary reclaim of an ended
/// lease carries no such field.
#[test]
fn event_names_the_proof() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/event/");
    w.register(NAME, Some(&sid(1)));
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
        payload["reason"].as_str().unwrap().contains("early reclaim: launcher exited"),
        "{payload}"
    );
    assert_eq!(
        payload["early_proof"],
        json!({
            "liveness": "launcher-exited",
            "session_name": NAME,
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
    let rows = w.rows("SELECT payload FROM event WHERE kind = 'card.reclaim' ORDER BY event_id DESC LIMIT 1");
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
        w.register(&format!("agent-{n}"), Some(&sid(n)));
    }
    for n in 1..=3u64 {
        let g = n.to_string();
        let got = result(&w.reclaim(&sid(n + 1), &card, &g));
        assert_eq!(got["reservation"]["generation"], json!(n + 1));
    }
    let lease = w.rows(
        "SELECT lease_until - acquired_at FROM reservation WHERE generation = 4",
    );
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
    w.register(NAME, Some(&sid(1)));
    let before = w.dump();
    w.refused_and_unchanged(w.reclaim(&sid(1), &card, "1"), "lease", &before);
}

/// AC-9 (REQ-11): a reason that fits alone but not with the proof is refused before any write;
/// one that fits exactly is taken.
#[test]
fn reason_too_long() {
    let w = world();
    let card = w.taken(&sid(1), Some(&w.wt), "src/long/");
    w.register(NAME, Some(&sid(1)));
    let suffix = " [early reclaim: launcher exited (session agent-1); worktree clean; no process in it]";
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
    reservation::reserve(&mut w, card, "holder-a", &["zone/".to_string()], 3600, None, None)
        .unwrap();
    let detail = json!({"liveness": "launcher-exited"});
    let count = |w: &mut ledger::WriteLedger| -> i64 {
        w.with_write_tx(|tx, _| {
            Ok(tx.query_row("SELECT COUNT(*) FROM event", [], |r| r.get(0))?)
        })
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
    let got = reservation::reclaim_early(&mut w, card, "heir", 1, "leader", "r", 3600, &right)
        .unwrap();
    assert_eq!(got.reservation.generation, 2);
    assert_eq!(got.reservation.holder, "heir");
    assert_eq!(count(&mut w), events + 1);
    // Without a proof the live lease of the new holder is still refused.
    let plain = reservation::reclaim(&mut w, card, "other", 2, "leader", "r", 3600).unwrap_err();
    assert!(matches!(plain, ReclaimError::LeaseLive { .. }), "{plain}");
    assert_eq!(count(&mut w), events + 1, "a refusal wrote an event");
}
