//! `moai card adopt` (SPEC-20261123-lane-continues-card): after a /clear the same lane — the same
//! registered name, the same worktree, a new host session id — takes its own card over at the
//! next generation, without waiting for the lease and without the early-reclaim proofs.
//!
//! Every command runs against a temporary HOME and a temporary git repository with real linked
//! worktrees; the real home folder and the real session registry are never read or written. The
//! session is the id in the environment (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

use crate::common;

use common::git::git_in;
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const NAME: &str = "agent-1";

/// A well-formed host session id for number `n`.
fn sid(n: u64) -> String {
    format!("{n:08x}-0000-4000-8000-{n:012x}")
}

struct World {
    home: TempHome,
    host: tempfile::TempDir,
    project: PathBuf,
    /// The lane's worktree: a linked worktree of `project`, its work pushed.
    wt_a: PathBuf,
    /// Another worktree of the same clone.
    wt_b: PathBuf,
    path_env: std::ffi::OsString,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    git_in(&project, &["init", "-q", "-b", "main"]);
    fs::write(project.join("a.txt"), "a\n").unwrap();
    git_in(&project, &["add", "-A"]);
    git_in(&project, &["commit", "-q", "-m", "first"]);
    let host = tempfile::Builder::new()
        .prefix("moai-test-host-")
        .tempdir()
        .unwrap();
    let base = fs::canonicalize(host.path()).unwrap();
    let bare = base.join("remote.git");
    fs::create_dir_all(&bare).unwrap();
    git_in(&bare, &["init", "-q", "--bare", "-b", "main"]);
    git_in(
        &project,
        &["remote", "add", "origin", bare.to_str().unwrap()],
    );
    git_in(&project, &["push", "-q", "origin", "main"]);
    let mut worktrees = Vec::new();
    for name in ["lane-a", "lane-b"] {
        let path = base.join(name);
        git_in(
            &project,
            &["worktree", "add", "-q", "-b", name, path.to_str().unwrap()],
        );
        worktrees.push(fs::canonicalize(path).unwrap());
    }
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
        wt_b: worktrees.pop().unwrap(),
        wt_a: worktrees.pop().unwrap(),
        path_env,
    }
}

impl World {
    /// `moai <args>` in `cwd` as the host session `session` (none: no id).
    fn run(&self, cwd: &Path, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd = self.spawn_cmd(cwd, session, args);
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    fn spawn_cmd(&self, cwd: &Path, session: Option<&str>, args: &[&str]) -> Command {
        let mut cmd = self.home.command(args);
        cmd.env("PATH", &self.path_env).current_dir(cwd);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd
    }

    /// `moai card <args> --project <project>`, run in the project folder.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        self.card_in(&self.project, session, args)
    }

    fn card_in(&self, cwd: &Path, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", self.project.to_str().unwrap()]);
        self.run(cwd, session, &all)
    }

    fn adopt(&self, cwd: &Path, session: &str, card: &str, generation: &str) -> Output {
        self.card_in(
            cwd,
            Some(session),
            &["adopt", card, "--generation", generation],
        )
    }

    /// `moai session-register` for `name` from `cwd`, as `session`.
    fn register(&self, name: &str, session: &str, cwd: &Path, continue_: bool) -> Output {
        let project = self.project.to_str().unwrap();
        let mut args = vec!["session-register", project, "--name", name];
        if continue_ {
            args.push("--continue");
        }
        self.run(cwd, Some(session), &args)
    }

    /// A simulated /clear: the same lane name and worktree, the new host session `to`.
    fn clear(&self, name: &str, to: &str, cwd: &Path) -> Value {
        let out = self.register(name, to, cwd, true);
        let v = result(&out);
        assert_eq!(v["status"], json!("continued"), "{v}");
        v
    }

    fn sessions_dir(&self) -> PathBuf {
        moai_cli::clone_root::path(&self.project.join(".git")).join("sessions")
    }

    fn entry_path(&self, name: &str) -> PathBuf {
        self.sessions_dir().join(format!("{name}.json"))
    }

    fn entry(&self, name: &str) -> Value {
        serde_json::from_slice(&fs::read(self.entry_path(name)).unwrap()).unwrap()
    }

    fn write_entry(&self, name: &str, entry: &Value) {
        fs::create_dir_all(self.sessions_dir()).unwrap();
        fs::write(self.entry_path(name), serde_json::to_vec(entry).unwrap()).unwrap();
    }

    /// A hand-written entry of schema 1 with a live lease.
    fn forged_entry(&self, name: &str, session: &str, worktree: Option<&Path>, previous: &[&str]) {
        let now = self.rows("SELECT unixepoch()")[0][0].as_i64().unwrap();
        let mut e = json!({
            "schema_version": 1, "name": name, "instance": "3".repeat(32),
            "lease_until": now + 3600, "registered_at": now, "session": session,
        });
        if let Some(w) = worktree {
            e["worktree"] = json!(w.to_str().unwrap());
        }
        if !previous.is_empty() {
            e["previous_sessions"] = json!(previous);
        }
        self.write_entry(name, &e);
    }

    /// A ready card with one criterion taken by `holder`, with `worktree` and `branch`.
    fn taken(&self, holder: &str, worktree: &Path, branch: &str, scope: &str) -> String {
        let added = result(&self.card(
            None,
            &["add", "--title", "t", "--scope", scope, "--ac", "one"],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(
            Some(holder),
            &[
                "take",
                &name,
                "--worktree",
                worktree.to_str().unwrap(),
                "--branch",
                branch,
            ],
        ));
        name
    }

    /// The standard case: `sid(1)` holds a card in `wt_a` under `NAME`, then a /clear brings
    /// `sid(2)` into the same worktree under the same name.
    fn cleared_lane(&self) -> String {
        let card = self.taken(&sid(1), &self.wt_a, "lane-a", "src/lane/");
        result(&self.register(NAME, &sid(1), &self.wt_a, false));
        self.clear(NAME, &sid(2), &self.wt_a);
        card
    }

    fn note(&self, session: &str, card: &str, text: &str) {
        let mut cmd = self.spawn_cmd(
            &self.project,
            Some(session),
            &[
                "card",
                "note",
                card,
                "--kind",
                "comment",
                "--project",
                self.project.to_str().unwrap(),
            ],
        );
        let mut child = cmd
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(text.as_bytes())
            .unwrap();
        let out = child.wait_with_output().unwrap();
        result(&out);
    }

    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
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

    /// Every row of the tables an adoption could touch, as text.
    fn dump(&self) -> Vec<String> {
        self.dump_of(&[
            "card",
            "card_ac",
            "card_note",
            "card_link",
            "reservation",
            "reservation_scope",
            "event",
        ])
    }

    fn dump_of(&self, tables: &[&str]) -> Vec<String> {
        let mut out = Vec::new();
        for table in tables {
            for r in self.rows(&format!("SELECT * FROM {table} ORDER BY 1, 2")) {
                out.push(format!("{table}:{r}"));
            }
        }
        out
    }

    /// Notes, criteria and links: what an adoption must leave as it was.
    fn kept_tables(&self) -> Vec<String> {
        self.dump_of(&["card_ac", "card_note", "card_link"])
    }

    fn events_of(&self, kind: &str) -> Vec<Value> {
        self.rows(&format!(
            "SELECT actor, card_id, payload FROM event WHERE kind = '{kind}' ORDER BY event_id"
        ))
    }

    /// The registry's files with their bytes.
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

    /// A refusal of `code` for `out`: kind `conflict`, the code word in the message, and the
    /// ledger exactly as `before`.
    fn refused_with(&self, label: &str, out: Output, code: &str, before: &[String]) -> String {
        let (kind, message) = refusal(&out);
        assert_eq!(kind, "conflict", "{label}: {message}");
        assert!(message.contains(code), "{label}: {code}: {message}");
        assert_eq!(&self.dump(), before, "{label}: a refusal wrote something");
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

/// Every file below `dir` with its bytes, sorted: the running folder as a digest source.
fn tree_bytes(dir: &Path) -> Vec<(String, Vec<u8>)> {
    fn walk(root: &Path, at: &Path, out: &mut Vec<(String, Vec<u8>)>) {
        for e in fs::read_dir(at).unwrap().flatten() {
            let p = e.path();
            let rel = p.strip_prefix(root).unwrap().to_string_lossy().into_owned();
            if p.is_dir() && !p.is_symlink() {
                walk(root, &p, out);
            } else {
                out.push((rel, fs::read(&p).unwrap_or_default()));
            }
        }
    }
    let mut out = Vec::new();
    walk(dir, dir, &mut out);
    out.sort();
    out
}

fn card_number(card: &str) -> i64 {
    card.trim_start_matches("C-").parse().unwrap()
}

/// AC-3 (REQ-4, REQ-11): after the /clear of the same lane `card adopt` moves the reservation
/// to the new session at generation + 1 with the same scope, worktree and branch and a four hour
/// lease; notes, criteria and links are as before; the old session's `card check` at the old
/// generation refuses as stale.
#[test]
fn ac3_adopt_continues_the_card_at_the_next_generation() {
    let w = world();
    let card = w.taken(&sid(1), &w.wt_a, "lane-a", "src/lane/");
    w.note(&sid(1), &card, "a note written before the clear");
    result(&w.register(NAME, &sid(1), &w.wt_a, false));
    w.clear(NAME, &sid(2), &w.wt_a);
    let kept = w.kept_tables();
    let now = w.rows("SELECT unixepoch()")[0][0].as_i64().unwrap();

    let got = result(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(got["action"], json!("adopt"));
    let new = &got["reservation"];
    assert_eq!(new["generation"], json!(2));
    assert_eq!(new["holder"], json!(sid(2)));
    assert_eq!(new["status"], json!("held"));
    assert_eq!(new["write_scope"], json!(["src/lane/"]));
    assert_eq!(new["worktree"], json!(w.wt_a.to_str().unwrap()));
    assert_eq!(new["branch"], json!("lane-a"));
    let lease = new["lease_until"].as_i64().unwrap() - now;
    assert!(
        (14_400..=14_410).contains(&lease),
        "four hours, got {lease}"
    );
    assert_eq!(got["previous"]["generation"], json!(1));
    assert_eq!(got["previous"]["status"], json!("reclaimed"));
    assert_eq!(got["events"], json!(1));
    assert_eq!(got["writes"], json!(true));
    assert_eq!(w.kept_tables(), kept, "notes, criteria and links");
    assert_eq!(
        w.rows("SELECT state FROM card"),
        vec![json!(["active"])],
        "the card's state"
    );

    let old = w.card(
        Some(&sid(1)),
        &[
            "check",
            &card,
            "--generation",
            "1",
            "--path",
            "src/lane/x.rs",
        ],
    );
    let (kind, message) = refusal(&old);
    assert_eq!(kind, "conflict", "{message}");
    let new = w.card(
        Some(&sid(2)),
        &[
            "check",
            &card,
            "--generation",
            "2",
            "--path",
            "src/lane/x.rs",
        ],
    );
    result(&new);
}

/// AC-1 (REQ-1): the session an entry names now is never listed in its own `previous_sessions`,
/// also when an older session of the lane comes back after a /clear and a second /clear.
#[test]
fn ac1_a_session_that_comes_back_is_not_its_own_predecessor() {
    let w = world();
    result(&w.register(NAME, &sid(1), &w.wt_a, false));
    w.clear(NAME, &sid(2), &w.wt_a);
    w.clear(NAME, &sid(3), &w.wt_a);
    assert_eq!(
        w.entry(NAME)["previous_sessions"],
        json!([sid(2), sid(1)]),
        "the replaced sessions, newest first"
    );

    w.clear(NAME, &sid(1), &w.wt_a);
    let entry = w.entry(NAME);
    assert_eq!(entry["session"], json!(sid(1)));
    assert_eq!(
        entry["previous_sessions"],
        json!([sid(3), sid(2)]),
        "the returning session is not its own predecessor"
    );
}

/// AC-3: a card in `review` is adopted too and stays in `review`.
#[test]
fn ac3_a_card_in_review_is_adopted_and_stays_in_review() {
    let w = world();
    let card = w.taken(&sid(1), &w.wt_a, "lane-a", "src/lane/");
    result(&w.register(NAME, &sid(1), &w.wt_a, false));
    result(&w.card(
        Some(&sid(1)),
        &["review", &card, "--generation", "1", "--keep", "src/lane/"],
    ));
    w.clear(NAME, &sid(2), &w.wt_a);
    let got = result(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
    assert_eq!(got["reservation"]["status"], json!("held"));
    assert_eq!(w.rows("SELECT state FROM card"), vec![json!(["review"])]);
}

/// AC-4 (REQ-5): a session whose entry does not list the holder is refused with `not_same_lane`
/// and the ledger is as it was: another lane's entry, a fresh registration without
/// `--continue`, an entry whose lease has ended, and no entry at all.
#[test]
fn ac4_a_session_whose_entry_does_not_list_the_holder_is_refused() {
    // Another lane's entry.
    let w = world();
    let card = w.taken(&sid(1), &w.wt_a, "lane-a", "src/lane/");
    result(&w.register(NAME, &sid(1), &w.wt_a, false));
    result(&w.register("agent-2", &sid(2), &w.wt_a, false));
    let before = w.dump();
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    w.refused_with("another lane", out, "not_same_lane", &before);

    // A fresh registration of the same name after the holder's entry was freed (no `--continue`).
    let w = world();
    let card = w.taken(&sid(1), &w.wt_a, "lane-a", "src/lane/");
    let first = result(&w.register(NAME, &sid(1), &w.wt_a, false));
    let freed = w.run(
        &w.wt_a,
        Some(&sid(1)),
        &[
            "session-register",
            w.project.to_str().unwrap(),
            "--name",
            NAME,
            "--release",
            "--instance",
            first["instance"].as_str().unwrap(),
        ],
    );
    result(&freed);
    result(&w.register(NAME, &sid(2), &w.wt_a, false));
    let before = w.dump();
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    w.refused_with("a fresh registration", out, "not_same_lane", &before);

    // An entry whose lease has ended.
    let w = world();
    let card = w.cleared_lane();
    let mut entry = w.entry(NAME);
    let now = w.rows("SELECT unixepoch()")[0][0].as_i64().unwrap();
    entry["lease_until"] = json!(now - 60);
    w.write_entry(NAME, &entry);
    let before = w.dump();
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    w.refused_with("an ended lease", out, "not_same_lane", &before);

    // No entry at all.
    let w = world();
    let card = w.taken(&sid(1), &w.wt_a, "lane-a", "src/lane/");
    let before = w.dump();
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    w.refused_with("no entry", out, "not_same_lane", &before);

    // The holder itself has no entry listing itself.
    let w = world();
    let card = w.taken(&sid(1), &w.wt_a, "lane-a", "src/lane/");
    result(&w.register(NAME, &sid(1), &w.wt_a, false));
    let before = w.dump();
    let out = w.adopt(&w.wt_a, &sid(1), &card, "1");
    w.refused_with("the holder", out, "not_same_lane", &before);
}

/// AC-5 (REQ-6): an entry that still names the holder session id, under another lane name and
/// also with an ended lease, refuses with `holder_registered` naming the lane.
#[test]
fn ac5_an_entry_that_still_names_the_holder_is_refused() {
    let w = world();
    let card = w.cleared_lane();
    w.forged_entry("agent-3", &sid(1), Some(&w.wt_a), &[]);
    let before = w.dump();
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    let message = w.refused_with("a live entry", out, "holder_registered", &before);
    assert!(message.contains("agent-3"), "{message}");

    let mut entry = w.entry("agent-3");
    let now = w.rows("SELECT unixepoch()")[0][0].as_i64().unwrap();
    entry["lease_until"] = json!(now - 60);
    w.write_entry("agent-3", &entry);
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    w.refused_with("an ended entry", out, "holder_registered", &before);

    // Once the entry is gone the adoption goes through.
    fs::remove_file(w.entry_path("agent-3")).unwrap();
    let got = result(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(got["reservation"]["holder"], json!(sid(2)));
}

/// AC-6 (REQ-7): a reservation recorded for another worktree, a link that resolves to another
/// worktree, and an entry that records no worktree or another one each refuse with
/// `other_worktree`; a link that resolves to the running worktree is adopted.
#[test]
fn ac6_a_worktree_that_is_not_the_running_one_is_refused() {
    // The reservation records another worktree.
    let w = world();
    let card = w.taken(&sid(1), &w.wt_b, "lane-b", "src/lane/");
    result(&w.register(NAME, &sid(1), &w.wt_a, false));
    w.clear(NAME, &sid(2), &w.wt_a);
    let before = w.dump();
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    w.refused_with("another worktree", out, "other_worktree", &before);

    // The reservation records a symbolic link that resolves to another worktree.
    let w = world();
    let link = w.host.path().join("link-to-b");
    std::os::unix::fs::symlink(&w.wt_b, &link).unwrap();
    let card = w.taken(&sid(1), &link, "lane-a", "src/lane/");
    result(&w.register(NAME, &sid(1), &w.wt_a, false));
    w.clear(NAME, &sid(2), &w.wt_a);
    let before = w.dump();
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    w.refused_with("a link to another worktree", out, "other_worktree", &before);

    // The reservation records a symbolic link that resolves to the running worktree: the proof
    // compares canonical paths, so the adoption goes through. Without the canonical path of the
    // reserved worktree the link text would differ from the running worktree and be refused.
    let w = world();
    let link = w.host.path().join("link-to-a");
    std::os::unix::fs::symlink(&w.wt_a, &link).unwrap();
    let card = w.taken(&sid(1), &link, "lane-a", "src/lane/");
    result(&w.register(NAME, &sid(1), &w.wt_a, false));
    w.clear(NAME, &sid(2), &w.wt_a);
    let got = result(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(got["reservation"]["holder"], json!(sid(2)));
    assert_eq!(got["reservation"]["generation"], json!(2));
    assert_eq!(got["worktree"], json!(w.wt_a.to_str().unwrap()));
    assert_eq!(w.events_of("card.adopt").len(), 1);

    // The entry records no worktree, and then another worktree.
    let w = world();
    let card = w.cleared_lane();
    let mut entry = w.entry(NAME);
    entry.as_object_mut().unwrap().remove("worktree");
    w.write_entry(NAME, &entry);
    let before = w.dump();
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    w.refused_with("an entry with no worktree", out, "other_worktree", &before);
    entry["worktree"] = json!(w.wt_b.to_str().unwrap());
    w.write_entry(NAME, &entry);
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    w.refused_with(
        "an entry for another worktree",
        out,
        "other_worktree",
        &before,
    );

    // Running in another worktree than the reservation records.
    let out = w.adopt(&w.wt_b, &sid(2), &card, "1");
    w.refused_with(
        "a call from another worktree",
        out,
        "other_worktree",
        &before,
    );
}

/// AC-7 (REQ-8, REQ-11): a `done` card, a released reservation, a lower and a higher generation
/// each refuse with the answer `card reclaim` gives for the same state; a card that already
/// carries three `card.reclaim` events is adopted.
#[test]
fn ac7_a_card_in_another_state_is_refused_like_reclaim() {
    let reclaim = |w: &World, generation: &str| {
        let out = w.card(
            Some(&sid(2)),
            &[
                "reclaim",
                "C-1",
                "--generation",
                generation,
                "--by",
                "leader",
                "--reason",
                "the same state",
            ],
        );
        refusal(&out)
    };

    // A done card.
    let w = world();
    let card = w.cleared_lane();
    w.sql(&format!(
        "UPDATE card_ac SET status = 'met', evidence_ref = 'test' WHERE card_id = {n}; \
         INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
         VALUES ('card.landed:{n}:test', unixepoch(), 'test', {n}, 'card.landed', '{{}}'); \
         UPDATE card SET state = 'review' WHERE card_id = {n}; \
         UPDATE card SET state = 'done', closed_at = unixepoch() WHERE card_id = {n};",
        n = card_number(&card)
    ));
    let before = w.dump();
    let (kind, message) = refusal(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("is `done`"), "{message}");
    assert_eq!((kind, message), reclaim(&w, "1"), "done");
    assert_eq!(w.dump(), before);

    // A released reservation.
    let w = world();
    let card = w.cleared_lane();
    result(&w.card(Some(&sid(1)), &["release", &card, "--generation", "1"]));
    let before = w.dump();
    let adopted = refusal(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    assert!(adopted.1.contains("released"), "{}", adopted.1);
    assert_eq!(adopted, reclaim(&w, "1"), "released");
    assert_eq!(w.dump(), before);

    // A lower and a higher generation: the card is at generation 2 after one adoption.
    let w = world();
    let card = w.cleared_lane();
    result(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    let before = w.dump();
    let lower = refusal(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    assert!(lower.1.contains("is stale"), "{}", lower.1);
    assert_eq!(lower, reclaim(&w, "1"), "lower");
    let higher = refusal(&w.adopt(&w.wt_a, &sid(2), &card, "3"));
    assert!(higher.1.contains("was never issued"), "{}", higher.1);
    assert_eq!(higher, reclaim(&w, "3"), "higher");
    assert_eq!(w.dump(), before);

    // Three reclaims in the card's past do not stop an adoption.
    let w = world();
    let card = w.cleared_lane();
    let n = card_number(&card);
    let mut insert = String::new();
    for i in 0..3 {
        insert.push_str(&format!(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES ('card.reclaim:{n}:old{i}', unixepoch(), 'leader', {n}, 'card.reclaim', '{{}}');"
        ));
    }
    w.sql(&insert);
    let got = result(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(got["reservation"]["generation"], json!(2));
}

/// AC-9 (REQ-10): a refused and a successful call leave the registry files and the running
/// folder as they were, compared by their bytes; the refusal leaves every ledger row.
#[test]
fn ac9_the_proofs_read_and_write_nothing() {
    let w = world();
    let card = w.cleared_lane();
    fs::write(w.wt_a.join("work.txt"), "in progress\n").unwrap();

    // A refused call: the generation does not exist.
    let (registry, tree, ledger_rows) = (w.registry_bytes(), tree_bytes(&w.wt_a), w.dump());
    refusal(&w.adopt(&w.wt_a, &sid(2), &card, "9"));
    assert_eq!(w.registry_bytes(), registry, "refused: registry");
    assert_eq!(tree_bytes(&w.wt_a), tree, "refused: running folder");
    assert_eq!(w.dump(), ledger_rows, "refused: ledger");

    // A refused call at the proofs.
    w.forged_entry("agent-3", &sid(1), Some(&w.wt_a), &[]);
    let (registry, tree, ledger_rows) = (w.registry_bytes(), tree_bytes(&w.wt_a), w.dump());
    refusal(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(w.registry_bytes(), registry, "proof refusal: registry");
    assert_eq!(tree_bytes(&w.wt_a), tree, "proof refusal: running folder");
    assert_eq!(w.dump(), ledger_rows, "proof refusal: ledger");
    fs::remove_file(w.entry_path("agent-3")).unwrap();

    // A successful call writes only ledger rows.
    let (registry, tree) = (w.registry_bytes(), tree_bytes(&w.wt_a));
    result(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(w.registry_bytes(), registry, "success: registry");
    assert_eq!(tree_bytes(&w.wt_a), tree, "success: running folder");
}

/// AC-10 (REQ-12): an adoption with an uncommitted file, an unpushed commit, a stash entry, a
/// process inside the worktree and a held `.live` lock all present succeeds.
#[test]
fn ac10_the_worktree_and_the_launcher_lock_are_not_read() {
    let w = world();
    let card = w.cleared_lane();
    // An unpushed commit, a stash entry and an uncommitted file.
    fs::write(w.wt_a.join("a.txt"), "changed\n").unwrap();
    git_in(&w.wt_a, &["stash", "push", "-q", "-m", "adopt-test"]);
    fs::write(w.wt_a.join("b.txt"), "b\n").unwrap();
    git_in(&w.wt_a, &["add", "b.txt"]);
    git_in(&w.wt_a, &["commit", "-q", "-m", "unpushed"]);
    fs::write(w.wt_a.join("c.txt"), "uncommitted\n").unwrap();
    // A process whose working folder is the worktree.
    let mut child: Child = Command::new("sleep")
        .arg("30")
        .current_dir(&w.wt_a)
        .stdin(Stdio::null())
        .spawn()
        .expect("start sleep");
    // A launcher that holds the lane's `.live` lock.
    let live = fs::File::create(w.sessions_dir().join(format!("{NAME}.live"))).unwrap();
    live.lock().unwrap();

    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    let _ = child.kill();
    let _ = child.wait();
    let got = result(&out);
    assert_eq!(got["reservation"]["generation"], json!(2));
    assert_eq!(got["reservation"]["holder"], json!(sid(2)));
    drop(live);
}

/// AC-11 (REQ-4): the `card.adopt` event holds the old and the new holder, both generations, the
/// lane name, the worktree and the `previous_sessions` proof; one event per adoption and none
/// for a refusal; two adoptions at once write one event and one of them is refused as stale.
#[test]
fn ac11_the_adopt_event_names_everything() {
    let w = world();
    let card = w.cleared_lane();
    refusal(&w.adopt(&w.wt_a, &sid(2), &card, "7"));
    assert!(w.events_of("card.adopt").is_empty(), "none for a refusal");

    result(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    let events = w.events_of("card.adopt");
    assert_eq!(events.len(), 1);
    assert_eq!(events[0][0], json!(NAME), "the actor is the lane name");
    assert_eq!(events[0][1], json!(card_number(&card)));
    let p: Value = serde_json::from_str(events[0][2].as_str().unwrap()).unwrap();
    assert_eq!(p["from_holder"], json!(sid(1)));
    assert_eq!(p["holder"], json!(sid(2)));
    assert_eq!(p["from_generation"], json!(1));
    assert_eq!(p["generation"], json!(2));
    assert_eq!(p["lane"], json!(NAME));
    assert_eq!(p["worktree"], json!(w.wt_a.to_str().unwrap()));
    assert_eq!(p["previous_sessions"], json!([sid(1)]));
    assert!(w.events_of("card.reclaim").is_empty());

    // Two adoptions at once, from the same session and generation: one event, one refusal.
    let w = world();
    let card = w.cleared_lane();
    let args = [
        "card",
        "adopt",
        card.as_str(),
        "--generation",
        "1",
        "--project",
        w.project.to_str().unwrap(),
    ];
    let mut a = w.spawn_cmd(&w.wt_a, Some(&sid(2)), &args);
    let mut b = w.spawn_cmd(&w.wt_a, Some(&sid(2)), &args);
    let (ca, cb) = (
        a.stdin(Stdio::null())
            .stdout(Stdio::piped())
            .spawn()
            .unwrap(),
        b.stdin(Stdio::null())
            .stdout(Stdio::piped())
            .spawn()
            .unwrap(),
    );
    let (oa, ob) = (
        ca.wait_with_output().unwrap(),
        cb.wait_with_output().unwrap(),
    );
    let codes = [oa.status.code(), ob.status.code()];
    assert!(
        codes.contains(&Some(0)) && codes.contains(&Some(1)),
        "exactly one adoption wins: {codes:?}"
    );
    assert_eq!(w.events_of("card.adopt").len(), 1);
    assert_eq!(w.rows("SELECT generation FROM card"), vec![json!([2])]);

    // A third session of the lane continues after the second: the second can no longer adopt, the
    // third can.
    let w = world();
    let card = w.cleared_lane();
    w.clear(NAME, &sid(3), &w.wt_a);
    let before = w.dump();
    let out = w.adopt(&w.wt_a, &sid(2), &card, "1");
    w.refused_with("a replaced session", out, "not_same_lane", &before);
    result(&w.adopt(&w.wt_a, &sid(3), &card, "1"));
    assert_eq!(w.events_of("card.adopt").len(), 1);
}

/// AC-12 (REQ-11): the same card adopted four times in turn by four sessions of one lane keeps
/// the reclaim count at zero, and a later `card reclaim` of a dead holder still works up to its
/// own limit of three.
#[test]
fn ac12_adoptions_do_not_use_up_the_reclaim_allowance() {
    let w = world();
    let card = w.cleared_lane();
    result(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    for (generation, session) in [(2, 3), (3, 4), (4, 5)] {
        w.clear(NAME, &sid(session), &w.wt_a);
        let got = result(&w.adopt(&w.wt_a, &sid(session), &card, &generation.to_string()));
        assert_eq!(got["reservation"]["generation"], json!(generation + 1));
    }
    assert_eq!(w.events_of("card.adopt").len(), 4);
    assert!(w.events_of("card.reclaim").is_empty(), "the reclaim count");

    // Three reclaims of a dead holder still go through; the fourth is the limit.
    let n = card_number(&card);
    let end_lease = format!(
        "UPDATE reservation SET lease_until = unixepoch() - 5 WHERE card_id = {n} AND status = 'held'"
    );
    let mut generation = 5;
    for session in [10, 11, 12] {
        w.sql(&end_lease);
        let out = w.card(
            Some(&sid(session)),
            &[
                "reclaim",
                &card,
                "--generation",
                &generation.to_string(),
                "--by",
                "leader",
                "--reason",
                "the holder is gone",
            ],
        );
        let got = result(&out);
        generation += 1;
        assert_eq!(got["reservation"]["generation"], json!(generation));
    }
    w.sql(&end_lease);
    let out = w.card(
        Some(&sid(13)),
        &[
            "reclaim",
            &card,
            "--generation",
            &generation.to_string(),
            "--by",
            "leader",
            "--reason",
            "the holder is gone",
        ],
    );
    let (kind, message) = refusal(&out);
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("reclaim limit"), "{message}");
}
