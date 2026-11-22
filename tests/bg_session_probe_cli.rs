//! The background-session probe (SPEC-20261122-bg-session-research): a harness that measures what
//! `claude --bg`, `claude agents --json`, `claude rm` and the signals do, in a temporary home, a
//! temporary Claude config folder and a temporary repository, never in the real ones.
//!
//! Nothing here ships: the harness lives in this one test target. Its tests run against a stub
//! `claude` shell script, so no real Claude Code, no model and no network is involved. The real
//! measurement is the ignored test `measure_real_claude`, run by hand with
//! `MOAI_BG_PROBE_CLAUDE=<path> MOAI_BG_PROBE_OUT=<file> cargo test --test bg_session_probe_cli
//! measure_real_claude -- --ignored`; it needs no credentials and asks no model anything (a session
//! without credentials stops at its first request), and it writes the rows as JSON to the file.

mod common;

use common::CleanEnv;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

/// The only variable names a probed process sees. Nothing of the caller's environment is passed.
const CHILD_ENV_NAMES: &[&str] = &[
    "HOME",
    "USERPROFILE",
    "CLAUDE_CONFIG_DIR",
    "TMPDIR",
    "PATH",
    "LANG",
    "GIT_CONFIG_GLOBAL",
    "GIT_CONFIG_SYSTEM",
    "DISABLE_AUTOUPDATER",
];

/// The model-needing rows of the SPEC and why they are not run (the user's decision of 2026-11-22:
/// model-free measurements first, the rest after the report is read).
const MODEL_ROWS: &[(&str, &str)] = &[
    (
        "placement_by_edit",
        "needs a model reply: the move into a worktree happens before a session's first edit",
    ),
    (
        "git_behaviour_no_remote",
        "needs a model reply: commit, push and pull request come from a session that edits",
    ),
    (
        "git_behaviour_with_remote",
        "needs a model reply: commit, push and pull request come from a session that edits",
    ),
    (
        "instruction_place_task_claude_md_memory",
        "needs a model reply: which instruction place keeps git to the user is read from what a session does",
    ),
    (
        "idle_stop_real_turn",
        "needs a model reply and about one hour of wall time: a session must finish a real turn first",
    ),
];

/// What a signal or idle measurement needs from a live session process.
const NO_LIVE_PROCESS: &str =
    "no live session process: a session without credentials stops at its first request";

fn child_env(root: &Path) -> BTreeMap<String, String> {
    let home = root.join("home");
    let mut env = BTreeMap::new();
    env.insert("HOME".into(), home.display().to_string());
    env.insert("USERPROFILE".into(), home.display().to_string());
    env.insert(
        "CLAUDE_CONFIG_DIR".into(),
        root.join("config").display().to_string(),
    );
    env.insert("TMPDIR".into(), root.join("tmp").display().to_string());
    env.insert(
        "PATH".into(),
        format!(
            "{}:/usr/bin:/bin:/usr/sbin:/sbin",
            root.join("bin").display()
        ),
    );
    env.insert("LANG".into(), "en_US.UTF-8".into());
    env.insert("GIT_CONFIG_GLOBAL".into(), "/dev/null".into());
    env.insert("GIT_CONFIG_SYSTEM".into(), "/dev/null".into());
    env.insert("DISABLE_AUTOUPDATER".into(), "1".into());
    env
}

/// One command that was run.
#[derive(Debug, Clone)]
struct Row {
    id: String,
    command: Vec<String>,
    status: Option<i32>,
    timed_out: bool,
    stdout: String,
    stderr: String,
}

struct Sandbox {
    dir: tempfile::TempDir,
    claude: PathBuf,
    runs: std::sync::atomic::AtomicUsize,
}

impl Sandbox {
    fn new(claude: &Path) -> Self {
        let dir = tempfile::Builder::new()
            .prefix("moai-bgprobe-")
            .tempdir()
            .expect("temporary root");
        let sb = Sandbox {
            dir,
            claude: claude.to_path_buf(),
            runs: std::sync::atomic::AtomicUsize::new(0),
        };
        for d in ["home", "config", "tmp", "bin", "repo"] {
            fs::create_dir_all(sb.root().join(d)).unwrap();
        }
        let repo = sb.repo();
        sb.git(&repo, &["init", "-q", "-b", "main"]);
        fs::write(repo.join("a.txt"), "a\n").unwrap();
        sb.git(&repo, &["add", "a.txt"]);
        sb.git(&repo, &["commit", "-q", "-m", "init"]);
        sb
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    fn repo(&self) -> PathBuf {
        self.root().join("repo")
    }

    fn config(&self) -> PathBuf {
        self.root().join("config")
    }

    fn envelope(&self, program: impl AsRef<std::ffi::OsStr>) -> Command {
        let mut cmd = Command::new(program);
        cmd.clean_env();
        for (k, v) in child_env(self.root()) {
            cmd.env(k, v);
        }
        cmd.env("GIT_AUTHOR_NAME", "probe")
            .env("GIT_AUTHOR_EMAIL", "probe@invalid")
            .env("GIT_COMMITTER_NAME", "probe")
            .env("GIT_COMMITTER_EMAIL", "probe@invalid")
            .stdin(Stdio::null());
        cmd
    }

    /// `git` in the scratch repository, with the sandbox environment; panics on failure.
    fn git(&self, dir: &Path, args: &[&str]) -> String {
        let out = self
            .envelope("git")
            .current_dir(dir)
            .args(args)
            .output()
            .expect("git");
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    /// The text with the temporary root named as `<root>`, so nothing real is recorded.
    fn redact(&self, text: &str) -> String {
        let root = self.root().display().to_string();
        let canonical = fs::canonicalize(self.root())
            .map(|p| p.display().to_string())
            .unwrap_or_else(|_| root.clone());
        text.replace(&canonical, "<root>").replace(&root, "<root>")
    }

    /// Runs `claude <args>` in `cwd` with the sandbox environment and a time limit.
    fn claude(&self, id: &str, args: &[&str], cwd: &Path, limit: Duration) -> Row {
        self.run(id, &self.claude, args, cwd, limit, &[])
    }

    fn run(
        &self,
        id: &str,
        program: &Path,
        args: &[&str],
        cwd: &Path,
        limit: Duration,
        extra: &[(&str, &str)],
    ) -> Row {
        let mut cmd = self.envelope(program);
        for (k, v) in extra {
            cmd.env(k, v);
        }
        // Output goes to files, not pipes: a program that leaves a supervisor or a session behind
        // hands it the descriptors, and a pipe would then never reach its end.
        let n = self.runs.fetch_add(1, std::sync::atomic::Ordering::SeqCst);
        let out_path = self.root().join("tmp").join(format!("run-{n}.out"));
        let err_path = self.root().join("tmp").join(format!("run-{n}.err"));
        let mut child = cmd
            .args(args)
            .current_dir(cwd)
            .stdout(fs::File::create(&out_path).unwrap())
            .stderr(fs::File::create(&err_path).unwrap())
            .spawn()
            .expect("start the probed program");
        let started = Instant::now();
        let mut timed_out = false;
        let status = loop {
            if let Some(s) = child.try_wait().unwrap() {
                break s.code();
            }
            if started.elapsed() > limit {
                timed_out = true;
                let _ = child.kill();
                let _ = child.wait();
                break None;
            }
            std::thread::sleep(Duration::from_millis(50));
        };
        let read =
            |p: &Path| String::from_utf8_lossy(&fs::read(p).unwrap_or_default()).into_owned();
        let (stdout, stderr) = (read(&out_path), read(&err_path));
        let name = program
            .file_name()
            .map(|n| n.to_string_lossy().into_owned())
            .unwrap_or_default();
        let mut command = vec![name];
        command.extend(args.iter().map(|a| self.redact(a)));
        Row {
            id: id.to_string(),
            command,
            status,
            timed_out,
            stdout: self.redact(&stdout),
            stderr: self.redact(&stderr),
        }
    }

    /// Processes whose command line names the temporary root, as `(pid, command)`.
    fn leftovers(&self) -> Vec<(u32, String)> {
        let out = Command::new("ps")
            .args(["-axo", "pid=,command="])
            .output()
            .expect("ps");
        let root = self.root().display().to_string();
        let canonical = fs::canonicalize(self.root())
            .map(|p| p.display().to_string())
            .unwrap_or_default();
        String::from_utf8_lossy(&out.stdout)
            .lines()
            .filter(|l| l.contains(&root) || (!canonical.is_empty() && l.contains(&canonical)))
            .filter(|l| !l.contains(" ps -axo"))
            .filter_map(|l| {
                let l = l.trim_start();
                let (pid, rest) = l.split_once(' ')?;
                Some((pid.parse().ok()?, self.redact(rest.trim())))
            })
            .collect()
    }

    /// Every file and folder under `dir`, relative to the temporary root, sorted.
    fn listing(&self, dir: &Path) -> Vec<String> {
        fn walk(base: &Path, dir: &Path, out: &mut Vec<String>) {
            let Ok(entries) = fs::read_dir(dir) else {
                return;
            };
            for e in entries.flatten() {
                let p = e.path();
                out.push(p.strip_prefix(base).unwrap().display().to_string());
                if p.is_dir() && !fs::symlink_metadata(&p).unwrap().file_type().is_symlink() {
                    walk(base, &p, out);
                }
            }
        }
        let mut out = Vec::new();
        walk(self.root(), dir, &mut out);
        out.sort();
        out
    }
}

/// The names directly inside the real `~/.claude`: a read-only listing that two moments are
/// compared by, to show the probe added or removed nothing there. Modification times are not
/// compared: other live sessions write their own history and cache files all the time.
fn real_config_listing() -> Vec<String> {
    let Some(home) = std::env::var_os("HOME") else {
        return Vec::new();
    };
    let dir = PathBuf::from(home).join(".claude");
    let Ok(entries) = fs::read_dir(&dir) else {
        return Vec::new();
    };
    let mut out: Vec<String> = entries
        .flatten()
        .map(|e| e.file_name().to_string_lossy().into_owned())
        .collect();
    out.sort();
    out
}

fn row_json(r: &Row) -> Value {
    json!({
        "id": r.id, "command": r.command, "status": r.status, "timed_out": r.timed_out,
        "stdout": r.stdout, "stderr": r.stderr,
    })
}

/// The report: the version and platform the rows were measured on, the rows, the rows that were
/// not run with their reasons, and what the harness could not stop.
struct Report {
    version: String,
    platform: String,
    rows: Vec<Value>,
    not_run: Vec<Value>,
    leftovers: Vec<Value>,
}

impl Report {
    fn new() -> Self {
        Report {
            version: String::new(),
            platform: format!("{}-{}", std::env::consts::OS, std::env::consts::ARCH),
            rows: Vec::new(),
            not_run: Vec::new(),
            leftovers: Vec::new(),
        }
    }

    fn row(&mut self, r: &Row) {
        self.rows.push(row_json(r));
    }

    fn fact(&mut self, id: &str, value: Value) {
        self.rows.push(json!({"id": id, "fact": value}));
    }

    fn not_run(&mut self, id: &str, reason: &str) {
        self.not_run.push(json!({"id": id, "reason": reason}));
    }

    fn json(&self) -> Value {
        json!({
            "claude_version": self.version, "platform": self.platform, "rows": self.rows,
            "not_run": self.not_run, "leftovers": self.leftovers,
        })
    }
}

const SHORT: Duration = Duration::from_secs(60);

/// Adds the project trust entry and the onboarding flag to the sandbox's own config file, so a
/// background session starts in the scratch repository. It writes the sandbox only.
fn trust_repo(sb: &Sandbox) {
    let path = sb.config().join(".claude.json");
    let mut value: Value = fs::read_to_string(&path)
        .ok()
        .and_then(|t| serde_json::from_str(&t).ok())
        .unwrap_or_else(|| json!({}));
    let repo = fs::canonicalize(sb.repo()).unwrap().display().to_string();
    value["hasCompletedOnboarding"] = json!(true);
    value["projects"] = json!({ repo: {"hasTrustDialogAccepted": true} });
    fs::write(&path, serde_json::to_vec_pretty(&value).unwrap()).unwrap();
}

/// The id of the first background session `agents --json --all` lists, if any.
fn first_session(json_text: &str) -> Option<(String, Option<u32>)> {
    let list: Value = serde_json::from_str(json_text.trim()).ok()?;
    let first = list
        .as_array()?
        .iter()
        .find(|s| s["kind"] == "background")?;
    Some((
        first["id"].as_str()?.to_string(),
        first["pid"].as_u64().map(|p| p as u32),
    ))
}

/// Whether a process exists.
fn alive(pid: u32) -> bool {
    Command::new("kill")
        .args(["-0", &pid.to_string()])
        .stderr(Stdio::null())
        .status()
        .map(|s| s.success())
        .unwrap_or(false)
}

fn signal(pid: u32, name: &str) {
    let _ = Command::new("kill")
        .args([&format!("-{name}"), &pid.to_string()])
        .status();
}

/// Measures what needs no model on the installed program. `wait` is the time the idle probe may
/// wait; a budget shorter than `idle_required` makes it a not-run row.
fn measure_model_free(sb: &Sandbox, idle_required: Duration, idle_budget: Duration) -> Report {
    let mut rep = Report::new();
    let repo = sb.repo();
    let real_before = real_config_listing();

    let v = sb.claude("version", &["--version"], &repo, SHORT);
    rep.version = v.stdout.trim().to_string();
    rep.row(&v);

    // The state with no session, and the daemon.
    for (id, args) in [
        ("agents_json_empty", vec!["agents", "--json"]),
        ("agents_json_all_empty", vec!["agents", "--json", "--all"]),
        (
            "agents_json_cwd_empty",
            vec!["agents", "--json", "--cwd", "."],
        ),
        ("daemon_status_empty", vec!["daemon", "status"]),
    ] {
        let r = sb.claude(id, &args, &repo, SHORT);
        rep.row(&r);
    }
    rep.fact(
        "config_listing_after_empty_reads",
        json!(sb.listing(&sb.config())),
    );

    // A background session in a repository that is not trusted yet.
    let untrusted = sb.claude(
        "bg_untrusted_repo",
        &["--bg", "probe: reply with the word ok"],
        &repo,
        SHORT,
    );
    rep.row(&untrusted);

    // The repository trusted (in the sandbox's own config), a worktree for the session, a file
    // that is ignored and listed in `.worktreeinclude`.
    fs::write(repo.join(".gitignore"), ".env.probe\n").unwrap();
    fs::write(repo.join(".worktreeinclude"), ".env.probe\n").unwrap();
    fs::write(repo.join(".env.probe"), "x\n").unwrap();
    sb.git(&repo, &["add", ".gitignore", ".worktreeinclude"]);
    sb.git(&repo, &["commit", "-q", "-m", "ignore and include"]);
    trust_repo(sb);
    let started = sb.claude(
        "bg_worktree_start",
        &[
            "--bg",
            "--worktree",
            "probe-wt",
            "probe: reply with the word ok",
        ],
        &repo,
        SHORT,
    );
    rep.row(&started);
    std::thread::sleep(Duration::from_millis(1500));
    let listed = sb.claude(
        "agents_json_all_after_start",
        &["agents", "--json", "--all"],
        &repo,
        SHORT,
    );
    rep.row(&listed);
    let session = first_session(&listed.stdout);

    let wt_list = sb.git(&repo, &["worktree", "list", "--porcelain"]);
    let wt_dir = repo.join(".claude/worktrees/probe-wt");
    rep.fact(
        "worktree_placement",
        json!({
            "worktree_list": sb.redact(&wt_list),
            "dir_exists": wt_dir.is_dir(),
            "branch_exists": !sb.git(&repo, &["branch", "--list", "worktree-probe-wt"]).is_empty(),
            "locked": wt_list.lines().any(|l| l.starts_with("locked")),
            "worktreeinclude_file_arrived": wt_dir.join(".env.probe").is_file(),
        }),
    );

    // Removal rules on that worktree: uncommitted change, then an unpushed commit, then a lock.
    if let Some((id, _)) = session.clone() {
        if wt_dir.is_dir() {
            fs::write(wt_dir.join("dirty.txt"), "dirty\n").unwrap();
            let r = sb.claude("rm_with_uncommitted_changes", &["rm", &id], &repo, SHORT);
            rep.row(&r);
            rep.fact(
                "rm_with_uncommitted_changes_dir_exists",
                json!(wt_dir.is_dir()),
            );
            sb.git(&wt_dir, &["add", "dirty.txt"]);
            sb.git(&wt_dir, &["commit", "-q", "-m", "unpushed"]);
            let r = sb.claude("rm_with_unpushed_commit", &["rm", &id], &repo, SHORT);
            rep.row(&r);
            rep.fact("rm_with_unpushed_commit_dir_exists", json!(wt_dir.is_dir()));
        } else {
            rep.not_run(
                "rm_rules",
                "the session made no worktree directory to remove",
            );
        }
    } else {
        rep.not_run("rm_rules", "no background session was listed");
    }

    // Idle stop first, while the session still lives: only with a live session and a wait that
    // covers the documented idle time.
    let live = session
        .as_ref()
        .and_then(|(_, pid)| *pid)
        .filter(|p| alive(*p));
    if idle_budget < idle_required {
        rep.not_run(
            "idle_stop",
            "the wait budget is shorter than the documented idle time, so the wait was not completed",
        );
    } else if let Some(pid) = live {
        let before = sb.claude(
            "agents_json_all_idle_before",
            &["agents", "--json", "--all"],
            &repo,
            SHORT,
        );
        rep.row(&before);
        std::thread::sleep(idle_required);
        let after = sb.claude(
            "agents_json_all_idle_after",
            &["agents", "--json", "--all"],
            &repo,
            SHORT,
        );
        rep.row(&after);
        rep.fact("idle_stop_process", json!({"pid_alive_after": alive(pid)}));
    } else {
        rep.not_run("idle_stop", NO_LIVE_PROCESS);
    }

    // Signals last: the terminate signal ends the session.
    match session
        .as_ref()
        .and_then(|(_, pid)| *pid)
        .filter(|p| alive(*p))
    {
        Some(pid) => {
            for (name, label) in [("INT", "signal_interrupt"), ("TERM", "signal_terminate")] {
                signal(pid, name);
                std::thread::sleep(Duration::from_millis(800));
                let state = sb.claude(
                    &format!("agents_json_all_after_{label}"),
                    &["agents", "--json", "--all"],
                    &repo,
                    SHORT,
                );
                rep.row(&state);
                rep.fact(label, json!({"pid_alive_after": alive(pid)}));
            }
        }
        None => rep.not_run("signals", NO_LIVE_PROCESS),
    }

    for (id, why) in MODEL_ROWS {
        rep.not_run(id, why);
    }

    // Stop what was started and report what could not be stopped.
    let stop = sb.claude("daemon_stop", &["daemon", "stop", "--any"], &repo, SHORT);
    rep.row(&stop);
    std::thread::sleep(Duration::from_millis(500));
    for (pid, command) in sb.leftovers() {
        rep.leftovers.push(json!({"pid": pid, "command": command}));
    }
    rep.fact(
        "real_config_folder_unchanged",
        json!(real_config_listing() == real_before),
    );
    rep
}

// ---- the stub claude the tests run against ----------------------------------------------------

/// A shell script that behaves like the parts of `claude` the probe touches: it records the names
/// of the environment it was given, answers `--version`, `agents`, `daemon`, `logs` and `rm`, and
/// `--bg` starts a long-lived `sh` whose command line names the config folder (so a leftover is
/// found by pid). It writes only under `CLAUDE_CONFIG_DIR` and the scratch repository.
const STUB: &str = r#"#!/bin/sh
env | sed 's/=.*//' | sort > "$CLAUDE_CONFIG_DIR/stub-env-names.txt"
cfg="$CLAUDE_CONFIG_DIR"
case "$1" in
  --version) echo "2.1.330 (Claude Code)"; exit 0 ;;
  agents)
    if [ -f "$cfg/jobs.json" ]; then cat "$cfg/jobs.json"; else echo "[]"; fi
    exit 0 ;;
  daemon)
    case "$2" in
      status) echo "not running"; exit 1 ;;
      stop)
        if [ ! -f "$cfg/stub-no-stop" ] && [ -f "$cfg/stub.pid" ]; then kill "$(cat "$cfg/stub.pid")" 2>/dev/null; fi
        echo "stopped"; exit 0 ;;
    esac ;;
  logs) echo "stub log"; exit 0 ;;
  rm)
    wt="$PWD/.claude/worktrees/probe-wt"
    if [ -d "$wt" ]; then
      if [ -n "$(git -C "$wt" status --porcelain)" ]; then echo "keeps the worktree: uncommitted changes"; exit 0; fi
      if [ -n "$(git -C "$wt" log main..HEAD --format=%h)" ]; then echo "refuses: unpushed commits in its worktree" >&2; exit 1; fi
    fi
    echo "removed"; exit 0 ;;
  --bg)
    if ! grep -q hasTrustDialogAccepted "$cfg/.claude.json" 2>/dev/null; then
      echo "Workspace not trusted. Run claude once and accept the trust prompt, then retry."; exit 1
    fi
    if [ "$2" = "--worktree" ]; then
      wt="$PWD/.claude/worktrees/$3"
      git worktree add -q -b "worktree-$3" "$wt" >/dev/null 2>&1
      git worktree lock "$wt" >/dev/null 2>&1
      if [ -f .worktreeinclude ]; then
        while read -r f; do [ -n "$f" ] && [ -f "$f" ] && cp "$f" "$wt/$f"; done < .worktreeinclude
      fi
    fi
    if [ -f "$cfg/stub-no-stop" ]; then
      sh -c "trap '' INT TERM; while :; do sleep 1; done # $cfg" >/dev/null 2>&1 </dev/null &
    else
      sh -c "sleep 300 # $cfg" >/dev/null 2>&1 </dev/null &
    fi
    echo $! > "$cfg/stub.pid"
    printf '[{"id":"abc12345","kind":"background","state":"working","status":"busy","pid":%s,"cwd":"%s","startedAt":1}]\n' "$!" "$PWD" > "$cfg/jobs.json"
    echo "backgrounded · abc12345"; exit 0 ;;
esac
exit 0
"#;

fn stub_claude(dir: &Path) -> PathBuf {
    let path = dir.join("claude-stub");
    common::write_executable(&path, STUB);
    path
}

fn stub_world() -> (tempfile::TempDir, Sandbox) {
    let stubs = tempfile::Builder::new()
        .prefix("moai-bgprobe-stub-")
        .tempdir()
        .unwrap();
    let claude = stub_claude(stubs.path());
    let sb = Sandbox::new(&claude);
    (stubs, sb)
}

fn kill_leftovers(sb: &Sandbox) {
    for (pid, _) in sb.leftovers() {
        signal(pid, "KILL");
    }
}

fn find<'a>(rep: &'a Value, id: &str) -> &'a Value {
    rep["rows"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["id"] == id)
        .unwrap_or_else(|| panic!("no row {id} in {rep}"))
}

// ---- isolation_ -------------------------------------------------------------------------------

#[test]
fn isolation_child_env_is_rebuilt_from_a_short_list() {
    let root = tempfile::tempdir().unwrap();
    let env = child_env(root.path());
    let names: Vec<&str> = env.keys().map(String::as_str).collect();
    let mut want: Vec<&str> = CHILD_ENV_NAMES.to_vec();
    want.sort_unstable();
    assert_eq!(names, want, "the child environment is not the short list");
    for (k, v) in &env {
        if matches!(
            k.as_str(),
            "HOME" | "USERPROFILE" | "CLAUDE_CONFIG_DIR" | "TMPDIR"
        ) {
            assert!(v.starts_with(&root.path().display().to_string()), "{k}={v}");
        }
        assert!(
            !k.contains("KEY")
                && !k.contains("TOKEN")
                && !k.contains("ANTHROPIC")
                && !k.contains("SESSION"),
            "{k} is a credential or session name"
        );
    }
}

#[test]
fn isolation_a_stub_run_leaves_the_real_config_folder_and_every_record_under_the_root() {
    let (_stubs, sb) = stub_world();
    let real_before = real_config_listing();
    // The stub refuses to stop its process, so the leftover must be reported by pid.
    fs::write(sb.config().join("stub-no-stop"), "").unwrap();
    let rep = measure_model_free(&sb, Duration::from_secs(1), Duration::from_secs(0));
    let json = rep.json();
    assert_eq!(
        real_config_listing(),
        real_before,
        "the real config folder changed"
    );
    assert_eq!(
        find(&json, "real_config_folder_unchanged")["fact"],
        json!(true)
    );
    let left = json["leftovers"].as_array().unwrap();
    assert!(
        !left.is_empty(),
        "a process the stub kept was not reported: {json}"
    );
    assert!(left[0]["pid"].as_u64().is_some());
    kill_leftovers(&sb);
    // The names the stub saw are the short list plus the shell's own.
    let seen = fs::read_to_string(sb.config().join("stub-env-names.txt")).unwrap();
    for name in seen.lines() {
        assert!(
            CHILD_ENV_NAMES.contains(&name)
                || [
                    "PWD",
                    "SHLVL",
                    "_",
                    "OLDPWD",
                    "DEVELOPER_DIR",
                    "GIT_AUTHOR_NAME",
                    "GIT_AUTHOR_EMAIL",
                    "GIT_COMMITTER_NAME",
                    "GIT_COMMITTER_EMAIL",
                ]
                .contains(&name),
            "the probed process saw {name}"
        );
    }
}

#[test]
fn isolation_a_stopped_process_is_not_a_leftover() {
    let (_stubs, sb) = stub_world();
    let rep = measure_model_free(&sb, Duration::from_secs(1), Duration::from_secs(0));
    assert!(
        rep.json()["leftovers"].as_array().unwrap().is_empty(),
        "{}",
        rep.json()
    );
    kill_leftovers(&sb);
}

// ---- report_ ----------------------------------------------------------------------------------

#[test]
fn report_rows_carry_command_status_output_and_version() {
    let (_stubs, sb) = stub_world();
    let rep = measure_model_free(&sb, Duration::from_secs(1), Duration::from_secs(0));
    let json = rep.json();
    assert_eq!(json["claude_version"], json!("2.1.330 (Claude Code)"));
    assert!(json["platform"].as_str().unwrap().contains('-'));
    let empty = find(&json, "agents_json_empty");
    assert_eq!(empty["command"], json!(["claude-stub", "agents", "--json"]));
    assert_eq!(empty["status"], json!(0));
    // The recorded output is the program's output unchanged, newline included.
    assert_eq!(empty["stdout"], json!("[]\n"));
    let daemon = find(&json, "daemon_status_empty");
    assert_eq!(daemon["status"], json!(1));
    kill_leftovers(&sb);
}

#[test]
fn report_names_every_not_run_row_with_its_reason() {
    let (_stubs, sb) = stub_world();
    let rep = measure_model_free(&sb, Duration::from_secs(1), Duration::from_secs(0));
    let json = rep.json();
    let ids: Vec<&str> = json["not_run"]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| r["id"].as_str().unwrap())
        .collect();
    for (id, _) in MODEL_ROWS {
        assert!(ids.contains(id), "{id} is not listed as not run: {ids:?}");
    }
    assert!(ids.contains(&"idle_stop"), "{ids:?}");
    for r in json["not_run"].as_array().unwrap() {
        assert!(!r["reason"].as_str().unwrap().is_empty(), "{r}");
        assert!(
            json["rows"]
                .as_array()
                .unwrap()
                .iter()
                .all(|row| row["id"] != r["id"]),
            "{} is both measured and not run",
            r["id"]
        );
    }
    kill_leftovers(&sb);
}

#[test]
fn report_holds_no_card_number_no_key_and_no_real_path() {
    let (_stubs, sb) = stub_world();
    let rep = measure_model_free(&sb, Duration::from_secs(1), Duration::from_secs(0));
    let text = serde_json::to_string(&rep.json()).unwrap();
    kill_leftovers(&sb);
    let real_home = std::env::var("HOME").unwrap_or_default();
    if !real_home.is_empty() {
        assert!(!text.contains(&real_home), "the real home is in the report");
    }
    let root = sb.root().display().to_string();
    assert!(
        !text.contains(&root),
        "the temporary root path is in the report"
    );
    assert!(
        text.contains("<root>"),
        "paths are not written relative to the root"
    );
    assert!(!text.contains("sk-"), "a key-shaped text is in the report");
    let bytes = text.as_bytes();
    for i in 0..bytes.len().saturating_sub(3) {
        let w = &bytes[i..i + 3];
        assert!(
            !(w[0] == b'C' && w[1] == b'-' && w[2].is_ascii_digit()),
            "a card number in the report near {}",
            &text[i.saturating_sub(10)..(i + 12).min(text.len())]
        );
    }
}

// ---- rows_ ------------------------------------------------------------------------------------

#[test]
fn rows_an_untrusted_repository_is_a_row_with_the_documented_refusal() {
    let (_stubs, sb) = stub_world();
    let rep = measure_model_free(&sb, Duration::from_secs(1), Duration::from_secs(0));
    let json = rep.json();
    let row = find(&json, "bg_untrusted_repo");
    assert_eq!(row["status"], json!(1));
    assert!(row["stdout"]
        .as_str()
        .unwrap()
        .contains("Workspace not trusted"));
    kill_leftovers(&sb);
}

#[test]
fn rows_placement_reads_the_worktree_the_branch_the_lock_and_the_included_file_from_git() {
    let (_stubs, sb) = stub_world();
    let rep = measure_model_free(&sb, Duration::from_secs(1), Duration::from_secs(0));
    let json = rep.json();
    let placement = &find(&json, "worktree_placement")["fact"];
    assert_eq!(placement["dir_exists"], json!(true), "{placement}");
    assert_eq!(placement["branch_exists"], json!(true), "{placement}");
    assert_eq!(placement["locked"], json!(true), "{placement}");
    assert_eq!(
        placement["worktreeinclude_file_arrived"],
        json!(true),
        "{placement}"
    );
    let list = placement["worktree_list"].as_str().unwrap();
    assert!(
        list.contains("<root>/repo/.claude/worktrees/probe-wt"),
        "{list}"
    );
    kill_leftovers(&sb);
}

#[test]
fn rows_removal_outcomes_are_recorded_for_changes_and_for_unpushed_commits() {
    let (_stubs, sb) = stub_world();
    let rep = measure_model_free(&sb, Duration::from_secs(1), Duration::from_secs(0));
    let json = rep.json();
    let dirty = find(&json, "rm_with_uncommitted_changes");
    assert_eq!(dirty["status"], json!(0));
    assert!(dirty["stdout"].as_str().unwrap().contains("uncommitted"));
    assert_eq!(
        find(&json, "rm_with_uncommitted_changes_dir_exists")["fact"],
        json!(true)
    );
    let unpushed = find(&json, "rm_with_unpushed_commit");
    assert_eq!(unpushed["status"], json!(1));
    assert!(unpushed["stderr"].as_str().unwrap().contains("unpushed"));
    kill_leftovers(&sb);
}

#[test]
fn rows_signals_are_sent_to_a_live_session_process_and_the_state_is_read_after_each() {
    let (_stubs, sb) = stub_world();
    let rep = measure_model_free(&sb, Duration::from_secs(1), Duration::from_secs(0));
    let json = rep.json();
    find(&json, "agents_json_all_after_signal_interrupt");
    find(&json, "agents_json_all_after_signal_terminate");
    let term = &find(&json, "signal_terminate")["fact"];
    assert_eq!(term["pid_alive_after"], json!(false), "{term}");
    kill_leftovers(&sb);
}

#[test]
fn rows_the_environment_of_a_probed_process_is_recorded_by_name_and_never_by_value() {
    let (_stubs, sb) = stub_world();
    // A stand-in for a variable the supervisor injected: its value must never reach the report.
    let secret = sb.run(
        "stub_with_secret",
        &sb.claude,
        &["--version"],
        &sb.repo(),
        SHORT,
        &[("ANTHROPIC_API_KEY", "sk-live-secret-value")],
    );
    let names = fs::read_to_string(sb.config().join("stub-env-names.txt")).unwrap();
    assert!(
        names.contains("ANTHROPIC_API_KEY"),
        "the stand-in variable was not passed: {names}"
    );
    let mut rep = Report::new();
    rep.fact("probed_env_names", json!(names.lines().collect::<Vec<_>>()));
    rep.row(&secret);
    let text = serde_json::to_string(&rep.json()).unwrap();
    assert!(
        !text.contains("sk-live-secret-value"),
        "a value is in the report"
    );
    assert!(
        text.contains("ANTHROPIC_API_KEY"),
        "the name is not recorded"
    );
}

// ---- idle_ ------------------------------------------------------------------------------------

#[test]
fn idle_is_not_run_when_the_wait_budget_is_shorter_than_the_documented_idle_time() {
    let (_stubs, sb) = stub_world();
    let rep = measure_model_free(&sb, Duration::from_secs(3600), Duration::from_secs(5));
    let json = rep.json();
    let row = json["not_run"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["id"] == "idle_stop")
        .expect("idle_stop is not listed as not run");
    assert!(
        row["reason"].as_str().unwrap().contains("not completed"),
        "{row}"
    );
    assert!(json["rows"]
        .as_array()
        .unwrap()
        .iter()
        .all(|r| r["id"] != "agents_json_all_idle_after"));
    kill_leftovers(&sb);
}

#[test]
fn idle_records_both_readings_when_the_wait_completes() {
    let (_stubs, sb) = stub_world();
    let rep = measure_model_free(&sb, Duration::from_secs(1), Duration::from_secs(5));
    let json = rep.json();
    find(&json, "agents_json_all_idle_before");
    find(&json, "agents_json_all_idle_after");
    assert!(
        json["not_run"]
            .as_array()
            .unwrap()
            .iter()
            .all(|r| r["id"] != "idle_stop"),
        "{json}"
    );
    kill_leftovers(&sb);
}

// ---- the real measurement ---------------------------------------------------------------------

/// The measurement of the installed Claude Code, by hand: `MOAI_BG_PROBE_CLAUDE` names the
/// program and `MOAI_BG_PROBE_OUT` the file the JSON report is written to. Ignored in a normal run.
#[test]
#[ignore]
fn measure_real_claude() {
    let claude = PathBuf::from(
        std::env::var("MOAI_BG_PROBE_CLAUDE").expect("MOAI_BG_PROBE_CLAUDE names the program"),
    );
    let out = std::env::var("MOAI_BG_PROBE_OUT").expect("MOAI_BG_PROBE_OUT names the report file");
    let sb = Sandbox::new(&claude);
    // The idle stop needs a live session and about an hour: not part of this run.
    let rep = measure_model_free(&sb, Duration::from_secs(3600), Duration::from_secs(0));
    kill_leftovers(&sb);
    fs::write(&out, serde_json::to_vec_pretty(&rep.json()).unwrap()).unwrap();
}
