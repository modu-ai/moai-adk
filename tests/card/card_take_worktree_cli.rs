//! `moai card take` and `moai card reclaim` record the worktree of the reservation
//! (SPEC-20261125-card-take-records-worktree): a take without `--worktree` records the git
//! worktree of the folder it runs in, a reclaim fills a missing worktree from the folder it runs
//! in and keeps a recorded one, and `card adopt` then finds the worktree it needs.
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
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

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
    /// A folder that is no git worktree.
    plain: PathBuf,
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
    let plain = base.join("plain");
    fs::create_dir_all(&plain).unwrap();
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
        plain: fs::canonicalize(plain).unwrap(),
        path_env,
    }
}

impl World {
    /// `moai <args>` in `cwd` as the host session `session` (none: no id).
    fn run(&self, cwd: &Path, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd: Command = self.home.command(args);
        cmd.env("PATH", &self.path_env).current_dir(cwd);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    /// `moai card <args> --project <project>`, run in `cwd`.
    fn card_in(&self, cwd: &Path, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", self.project.to_str().unwrap()]);
        self.run(cwd, session, &all)
    }

    /// A ready card with one criterion and the write scope `scope`.
    fn ready_card(&self, scope: &str) -> String {
        let added = result(&self.card_in(
            &self.project,
            None,
            &["add", "--title", "t", "--scope", scope, "--ac", "one"],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card_in(&self.project, None, &["ready", &name]));
        name
    }

    /// `card take` of `card` as `holder`, run in `cwd`, with `extra` arguments.
    fn take(&self, cwd: &Path, holder: &str, card: &str, extra: &[&str]) -> Value {
        let mut args = vec!["take", card];
        args.extend_from_slice(extra);
        result(&self.card_in(cwd, Some(holder), &args))
    }

    fn reclaim(&self, cwd: &Path, session: &str, card: &str, generation: &str) -> Output {
        self.card_in(
            cwd,
            Some(session),
            &[
                "reclaim",
                card,
                "--generation",
                generation,
                "--by",
                "leader",
                "--reason",
                "the holder session ended",
            ],
        )
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

    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
    }

    fn end_lease(&self) {
        self.sql("UPDATE reservation SET lease_until = unixepoch() - 5 WHERE status = 'held'");
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

    /// Every row of the tables a reservation change could touch, as text.
    fn dump(&self) -> Vec<String> {
        let mut out = Vec::new();
        for table in ["card", "reservation", "reservation_scope", "event"] {
            for r in self.rows(&format!("SELECT * FROM {table} ORDER BY 1, 2")) {
                out.push(format!("{table}:{r}"));
            }
        }
        out
    }

    /// The payload of the one `card.reclaim` event, as JSON.
    fn reclaim_payload(&self) -> Value {
        let rows = self.rows("SELECT payload FROM event WHERE kind = 'card.reclaim'");
        assert_eq!(rows.len(), 1, "one reclaim event: {rows:?}");
        serde_json::from_str(rows[0][0].as_str().unwrap()).unwrap()
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

fn text(p: &Path) -> &str {
    p.to_str().unwrap()
}

/// AC-1 (REQ-1): a take without `--worktree` records the git worktree of the folder it runs in;
/// from a folder that is no git worktree it records none and says so.
#[test]
fn ac1_a_take_without_worktree_records_the_running_worktree() {
    let w = world();
    let card = w.ready_card("src/one/");
    let got = w.take(&w.wt_a, &sid(1), &card, &["--branch", "lane-a"]);
    assert_eq!(
        got["reservation"]["worktree"],
        json!(text(&w.wt_a)),
        "the running worktree: {got}"
    );

    let other = w.ready_card("src/two/");
    let got = w.take(&w.plain, &sid(2), &other, &[]);
    assert_eq!(got["reservation"]["worktree"], Value::Null, "{got}");
    let warnings = got["warnings"].to_string();
    assert!(
        warnings.contains("records no worktree"),
        "the answer says it: {warnings}"
    );
    let _ = &w.host;
}

/// AC-2 (REQ-2): a take with `--worktree` records the path as given, whatever the folder.
#[test]
fn ac2_a_take_with_worktree_records_the_given_path() {
    let w = world();
    let card = w.ready_card("src/one/");
    let got = w.take(
        &w.wt_a,
        &sid(1),
        &card,
        &["--worktree", text(&w.wt_b), "--branch", "lane-b"],
    );
    assert_eq!(
        got["reservation"]["worktree"],
        json!(text(&w.wt_b)),
        "{got}"
    );
}

/// AC-3 (REQ-3, REQ-4, REQ-5): a reclaim fills a missing worktree from its own folder and says
/// so in the event; a recorded worktree is kept and the event has no such key.
#[test]
fn ac3_a_reclaim_fills_a_missing_worktree_and_keeps_a_recorded_one() {
    let w = world();
    let card = w.ready_card("src/one/");
    let got = w.take(&w.plain, &sid(1), &card, &[]);
    assert_eq!(got["reservation"]["worktree"], Value::Null, "{got}");
    w.end_lease();
    let got = result(&w.reclaim(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(
        got["reservation"]["worktree"],
        json!(text(&w.wt_a)),
        "filled from the running folder: {got}"
    );
    assert_eq!(w.reclaim_payload()["worktree_filled"], json!(true));

    let w = world();
    let card = w.ready_card("src/one/");
    w.take(
        &w.wt_b,
        &sid(1),
        &card,
        &["--worktree", text(&w.wt_b), "--branch", "lane-b"],
    );
    w.end_lease();
    let got = result(&w.reclaim(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(
        got["reservation"]["worktree"],
        json!(text(&w.wt_b)),
        "a recorded worktree is kept: {got}"
    );
    let payload = w.reclaim_payload();
    assert!(
        payload.get("worktree_filled").is_none(),
        "no filling, no key: {payload}"
    );
}

/// AC-4 (REQ-6): a lane that took its card without `--worktree` from its own worktree is
/// replaced by a /clear and `card adopt` takes the card over, in the same worktree.
#[test]
fn ac4_a_lane_that_took_without_worktree_adopts_after_a_clear() {
    let w = world();
    let card = w.ready_card("src/lane/");
    w.take(&w.wt_a, &sid(1), &card, &["--branch", "lane-a"]);
    result(&w.register(NAME, &sid(1), &w.wt_a, false));
    let cleared = result(&w.register(NAME, &sid(2), &w.wt_a, true));
    assert_eq!(cleared["status"], json!("continued"), "{cleared}");

    let got = result(&w.adopt(&w.wt_a, &sid(2), &card, "1"));
    let new = &got["reservation"];
    assert_eq!(new["generation"], json!(2), "{got}");
    assert_eq!(new["holder"], json!(sid(2)));
    assert_eq!(new["worktree"], json!(text(&w.wt_a)));
}

/// AC-5 (REQ-7): a live reservation without a worktree is not taken by a reclaim, and the
/// ledger stays as it was.
#[test]
fn ac5_a_live_reservation_without_a_worktree_is_still_refused() {
    let w = world();
    let card = w.ready_card("src/one/");
    let got = w.take(&w.plain, &sid(1), &card, &[]);
    assert_eq!(got["reservation"]["worktree"], Value::Null, "{got}");
    let before = w.dump();

    let (kind, message) = refusal(&w.reclaim(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("the lease runs until"), "{message}");
    assert_eq!(w.dump(), before, "a refusal wrote something");
}

/// The hook scripts the settings of a worktree point at, the way the Go generation wrote them,
/// and the ones the worktree holds.
fn hook_settings(worktree: &Path, pointed: &[&str], present: &[&str]) {
    let claude = worktree.join(".claude");
    fs::create_dir_all(claude.join("hooks/moai")).unwrap();
    let handlers: Vec<Value> = pointed
        .iter()
        .map(|s| {
            json!({"type": "command",
                   "command": format!("\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/moai/{s}")})
        })
        .collect();
    let settings = json!({"hooks": {"Stop": [{"hooks": handlers}]}});
    fs::write(claude.join("settings.json"), settings.to_string()).unwrap();
    hook_files(worktree, present);
}

/// Puts the hook scripts `present` into the hook folder of `folder`.
fn hook_files(folder: &Path, present: &[&str]) {
    let hooks = folder.join(".claude/hooks/moai");
    fs::create_dir_all(&hooks).unwrap();
    for name in present {
        fs::write(hooks.join(name), "#!/bin/sh\n").unwrap();
    }
}

/// The words of the step that is only said when starting in the worktree finds every script.
const START_IN_WORKTREE: &str = "start the lane session in the worktree folder";

/// The lines of an answer's `warnings` that name the hook settings, one per line.
fn hook_warnings(answer: &Value) -> String {
    let all: Vec<&str> = answer["warnings"]
        .as_array()
        .map(|a| a.iter().filter_map(Value::as_str).collect())
        .unwrap_or_default();
    all.into_iter()
        .filter(|s| s.contains("hook settings"))
        .collect::<Vec<_>>()
        .join("\n")
}

/// SPEC-20261125-card-worktree-hook-config AC-2 (a) (REQ-1, REQ-5): the measured case. The
/// primary checkout, where a lane session starts, lacks the scripts the worktree's settings point
/// at and the worktree has them: the take says so in one line and names the step that finds them.
#[test]
fn the_hook_settings_of_the_worktree_are_checked_by_a_take() {
    let w = world();
    let primary = fs::canonicalize(&w.project).unwrap();
    hook_settings(
        &w.wt_a,
        &["handle-stop.sh", "handle-pre-tool.sh", "ok.sh"],
        &["handle-stop.sh", "handle-pre-tool.sh", "ok.sh"],
    );
    hook_files(&primary, &["ok.sh"]);
    let card = w.ready_card("src/one/");
    let got = w.take(&w.wt_a, &sid(1), &card, &["--branch", "lane-a"]);
    assert_eq!(got["state"], json!("active"), "{got}");
    let line = hook_warnings(&got);
    assert_eq!(line.lines().count(), 1, "one line: {line}");
    assert!(
        line.contains("handle-pre-tool.sh, handle-stop.sh"),
        "{line}"
    );
    assert!(
        !line.contains("ok.sh"),
        "a script the start folder has is not named: {line}"
    );
    assert!(
        line.contains(text(&primary)),
        "names the start folder: {line}"
    );
    assert!(line.contains(text(&w.wt_a)), "names the worktree: {line}");
    assert!(line.contains(START_IN_WORKTREE), "{line}");
    // following the step finds every named script; nothing was written for it
    for name in ["handle-stop.sh", "handle-pre-tool.sh"] {
        assert!(w.wt_a.join(".claude/hooks/moai").join(name).is_file());
        assert!(!primary.join(".claude/hooks/moai").join(name).exists());
    }
}

/// AC-2 (b) (REQ-1, REQ-5): when neither folder has every pointed script, the take does not tell
/// anyone where to start the session; it names the read-only plan and leaves the settings to the
/// user.
#[test]
fn the_hook_settings_of_the_worktree_with_no_complete_start_folder_name_the_plan() {
    for (in_primary, in_worktree) in [(vec!["a.sh"], vec!["b.sh"]), (vec![], vec![])] {
        let w = world();
        let primary = fs::canonicalize(&w.project).unwrap();
        hook_settings(&w.wt_a, &["a.sh", "b.sh"], &in_worktree);
        hook_files(&primary, &in_primary);
        let card = w.ready_card("src/one/");
        let got = w.take(&w.wt_a, &sid(1), &card, &["--branch", "lane-a"]);
        assert_eq!(got["state"], json!("active"), "{got}");
        let line = hook_warnings(&got);
        assert_eq!(line.lines().count(), 1, "one line: {got}");
        assert!(line.contains("b.sh"), "{line}");
        assert!(!line.contains(START_IN_WORKTREE), "{line}");
        assert!(line.contains("no start folder has them all"), "{line}");
        assert!(
            line.contains(&format!("moai go-hooks-plan {}", text(&w.wt_a))),
            "{line}"
        );
    }
}

/// AC-2 (c), (d) (REQ-2, REQ-3): no line when the start folder has every pointed script, even if
/// the worktree lacks them, when there is no settings file, or when it cannot be read; the take
/// still succeeds.
#[test]
fn the_hook_settings_of_the_worktree_stay_quiet_when_the_start_folder_has_them() {
    let w = world();
    hook_settings(&w.wt_a, &["a.sh", "b.sh"], &[]);
    hook_files(&w.project, &["a.sh", "b.sh"]);
    let card = w.ready_card("src/one/");
    let got = w.take(&w.wt_a, &sid(1), &card, &["--branch", "lane-a"]);
    assert_eq!(got["state"], json!("active"), "{got}");
    assert_eq!(hook_warnings(&got), "", "{got}");

    // no settings file in the worktree: no line
    let second = w.ready_card("src/two/");
    let got = w.take(&w.wt_b, &sid(2), &second, &["--branch", "lane-b"]);
    assert_eq!(hook_warnings(&got), "", "{got}");

    // a folder that is no git worktree is judged as its own start folder
    let third = w.ready_card("src/three/");
    hook_settings(&w.plain, &["a.sh"], &["a.sh"]);
    let got = w.take(&w.plain, &sid(3), &third, &["--worktree", text(&w.plain)]);
    assert_eq!(hook_warnings(&got), "", "{got}");

    // settings that cannot be read: no line, the take succeeds
    fs::write(w.plain.join(".claude/settings.json"), "not json").unwrap();
    let fourth = w.ready_card("src/four/");
    let got = w.take(&w.plain, &sid(4), &fourth, &["--worktree", text(&w.plain)]);
    assert_eq!(got["state"], json!("active"), "{got}");
    assert_eq!(hook_warnings(&got), "", "{got}");
    let _ = &w.host;
}

/// SPEC-20261125-card-worktree-hook-config AC-2 (REQ-1, REQ-2, REQ-5): a reclaim into a worktree
/// judges the start folder the same way, and says nothing when it has every script.
#[test]
fn the_hook_settings_of_the_worktree_are_checked_by_a_reclaim() {
    let w = world();
    hook_settings(&w.wt_a, &["handle-stop.sh"], &["handle-stop.sh"]);
    let card = w.ready_card("src/one/");
    w.take(&w.plain, &sid(1), &card, &[]);
    w.end_lease();
    let got = result(&w.reclaim(&w.wt_a, &sid(2), &card, "1"));
    let line = hook_warnings(&got);
    assert_eq!(line.lines().count(), 1, "one line: {got}");
    assert!(line.contains("handle-stop.sh"), "{line}");
    assert!(line.contains(START_IN_WORKTREE), "{line}");

    let w = world();
    hook_settings(&w.wt_a, &["handle-stop.sh"], &[]);
    hook_files(&w.project, &["handle-stop.sh"]);
    let card = w.ready_card("src/one/");
    w.take(&w.plain, &sid(1), &card, &[]);
    w.end_lease();
    let got = result(&w.reclaim(&w.wt_a, &sid(2), &card, "1"));
    assert_eq!(hook_warnings(&got), "", "{got}");
    assert!(
        got.get("warnings").is_none(),
        "no warnings key without a warning: {got}"
    );
}
