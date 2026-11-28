//! SPEC-20261126-hook-write-boundary: conformance of the three registered hook names
//! (`session-context`, `bash-boundary`, `write-boundary`) with the official hook contract, and
//! the guards that must still refuse when Claude Code hooks are off. Every case runs the real
//! binary in a temporary project and HOME; nothing here touches a real repository, hook or
//! ledger. A hook command is never executed: the entries only judge text.
#![cfg(unix)]
mod common;

use common::home::TempHome;
use common::CleanEnv;
use moai_cli::ledger::{self, WriteLedger};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const PROJECT: &str = "demo";
const BRANCH: &str = "spec/demo";
const RECEIPT: &str = "hook-entry.jsonl";
const NAMES: [&str; 3] = ["session-context", "bash-boundary", "write-boundary"];

// ---------------------------------------------------------------------- AC-10

struct Fx {
    _tmp: tempfile::TempDir,
    checkout: PathBuf,
    home: PathBuf,
}

impl Fx {
    /// A MoAI project that is a git checkout on `BRANCH`, a HOME holding a registered ledger with
    /// one card, its reservation of this checkout and a note.
    fn new() -> Fx {
        let tmp = tempfile::TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let checkout = root.join("checkout");
        let home = root.join("home");
        let state = checkout.join(moai_cli::state_root::MOAI_DIR);
        fs::create_dir_all(&state).unwrap();
        fs::create_dir_all(checkout.join(".git")).unwrap();
        fs::create_dir_all(checkout.join("src")).unwrap();
        fs::create_dir_all(&home).unwrap();
        fs::write(
            state.join("project.yaml"),
            format!("{{\"schema_version\": 2, \"project_id\": \"{PROJECT}\"}}\n"),
        )
        .unwrap();
        fs::write(
            checkout.join(".git/HEAD"),
            format!("ref: refs/heads/{BRANCH}\n"),
        )
        .unwrap();
        ledger::register(&home, PROJECT).unwrap();
        let fx = Fx {
            _tmp: tmp,
            checkout,
            home,
        };
        fx.seed_ledger();
        fx
    }

    fn seed_ledger(&self) {
        let mut w: WriteLedger = ledger::open_write(&self.home, PROJECT).unwrap();
        w.with_write_tx(|tx, now| {
            tx.execute(
                "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                 VALUES ('conformance', 'active', 'test', ?1, ?1)",
                rusqlite::params![now],
            )
            .unwrap();
            let card = tx.last_insert_rowid();
            tx.execute(
                "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
                 lease_until, updated_at, worktree, branch) \
                 VALUES (?1, 1, 'test', 'held', ?2, ?2 + 3600, ?2, ?3, ?4)",
                rusqlite::params![card, now, self.checkout.to_str().unwrap(), BRANCH],
            )
            .unwrap();
            tx.execute(
                "INSERT INTO card_note (card_id, kind, author, body, created_at) \
                 VALUES (?1, 'report', 'test', 'RESUME note', ?2)",
                rusqlite::params![card, now],
            )
            .unwrap();
            Ok(())
        })
        .unwrap();
    }

    fn hook(&self, name: &str, stdin: &[u8]) -> (i32, String, String) {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args(["hook", name])
            .current_dir(&self.checkout)
            .env("HOME", &self.home)
            .env_remove("CLAUDE_PROJECT_DIR")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        let mut child = cmd.spawn().unwrap();
        common::feed_stdin(&mut child, stdin);
        let out = child.wait_with_output().unwrap();
        (
            out.status.code().unwrap(),
            String::from_utf8_lossy(&out.stdout).into_owned(),
            String::from_utf8_lossy(&out.stderr).into_owned(),
        )
    }
}

/// Every file under `root` with its bytes, except the entry's own receipt file (a refused input
/// writes one line there by design).
fn snapshot(root: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(base: &Path, dir: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.file_name().is_some_and(|n| n == RECEIPT) {
                continue;
            }
            let rel = path
                .strip_prefix(base)
                .unwrap()
                .to_string_lossy()
                .into_owned();
            let meta = fs::symlink_metadata(&path).unwrap();
            if meta.file_type().is_symlink() {
                let target = fs::read_link(&path).unwrap();
                out.insert(
                    format!("{rel} -> link"),
                    target.to_string_lossy().as_bytes().to_vec(),
                );
            } else if meta.is_dir() {
                out.insert(format!("{rel}/"), Vec::new());
                walk(base, &path, out);
            } else {
                out.insert(rel, fs::read(&path).unwrap());
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(root, root, &mut out);
    out
}

/// Fails when a value anywhere in the output is `allow` or `ask`, when a permission decision is
/// not `deny`, or when the hook-specific object holds a key no contract of this product uses.
fn assert_no_allow_or_ask(stdout: &str, what: &str) {
    if stdout.trim().is_empty() {
        return;
    }
    let value: Value = serde_json::from_str(stdout.trim())
        .unwrap_or_else(|e| panic!("{what}: stdout is not JSON ({e}): {stdout}"));
    fn walk(v: &Value, what: &str) {
        match v {
            Value::String(s) => assert!(
                s != "allow" && s != "ask" && s != "defer",
                "{what}: a value is {s:?}"
            ),
            Value::Array(items) => items.iter().for_each(|i| walk(i, what)),
            Value::Object(map) => map.values().for_each(|i| walk(i, what)),
            _ => {}
        }
    }
    walk(&value, what);
    let object = value.as_object().expect("an object");
    assert_eq!(
        object.keys().collect::<Vec<_>>(),
        ["hookSpecificOutput"],
        "{what}"
    );
    let inner = object["hookSpecificOutput"].as_object().expect("an object");
    for key in inner.keys() {
        assert!(
            [
                "hookEventName",
                "permissionDecision",
                "permissionDecisionReason",
                "additionalContext"
            ]
            .contains(&key.as_str()),
            "{what}: key {key}"
        );
    }
    if let Some(decision) = inner.get("permissionDecision") {
        assert_eq!(decision, "deny", "{what}");
    }
}

/// The fixture set: allowed, denied, malformed, wrong event and wrong tool inputs for each name.
fn fixture_set(fx: &Fx) -> Vec<(&'static str, String, Vec<u8>)> {
    let cwd = fx.checkout.display().to_string();
    let call = |event: &str, tool: &str, input: Value| {
        json!({
            "hook_event_name": event, "session_id": "s1", "cwd": cwd,
            "tool_name": tool, "tool_input": input, "source": "startup",
        })
        .to_string()
        .into_bytes()
    };
    let mut set: Vec<(&'static str, String, Vec<u8>)> = Vec::new();
    for source in ["startup", "resume", "clear", "compact"] {
        let input = json!({"hook_event_name": "SessionStart", "session_id": "s1",
                           "cwd": cwd, "source": source});
        set.push((
            "session-context",
            format!("start {source}"),
            input.to_string().into_bytes(),
        ));
    }
    for command in [
        "ls -la",
        "git status",
        "rm -rf /opt/moai-conformance-outside",
        "git push --force origin main",
        "git commit --no-verify -m x",
        "rm -rf $UNKNOWN_VALUE",
    ] {
        set.push((
            "bash-boundary",
            format!("bash {command}"),
            call("PreToolUse", "Bash", json!({"command": command})),
        ));
    }
    let own = format!("{cwd}/src/a.rs");
    let secret = format!("{}{}", "ghp_", "a".repeat(36));
    for (tool, path, content) in [
        ("Write", own.as_str(), "hello"),
        ("Write", "/opt/moai-conformance/.ssh/k", "hello"),
        ("Edit", own.as_str(), secret.as_str()),
        ("NotebookEdit", own.as_str(), "x = 1"),
    ] {
        let input = match tool {
            "Edit" => json!({"file_path": path, "old_string": "a", "new_string": content}),
            "NotebookEdit" => json!({"notebook_path": path, "new_source": content}),
            _ => json!({"file_path": path, "content": content}),
        };
        set.push((
            "write-boundary",
            format!("{tool} {path}"),
            call("PreToolUse", tool, input),
        ));
    }
    for name in NAMES {
        let right = if name == "session-context" {
            "SessionStart"
        } else {
            "PreToolUse"
        };
        let wrong = if name == "session-context" {
            "PreToolUse"
        } else {
            "SessionStart"
        };
        set.push((
            name,
            "wrong event".into(),
            call(wrong, "Write", json!({"file_path": "/x", "content": "y"})),
        ));
        set.push((
            name,
            "wrong tool".into(),
            call(right, "Read", json!({"file_path": "/x"})),
        ));
        set.push((name, "not json".into(), b"not json".to_vec()));
        set.push((name, "an array".into(), b"[1,2]".to_vec()));
        set.push((name, "empty".into(), Vec::new()));
    }
    set
}

/// AC-10: no `allow` or `ask` in any output of the three names over the fixture set, and the
/// project's ledger, notes and reservations (every file but the entry's receipt) keep their bytes.
#[test]
fn ac10_no_name_allows_or_asks_and_none_changes_the_ledger() {
    let fx = Fx::new();
    let set = fixture_set(&fx);
    assert!(set.len() >= 25, "{} cases", set.len());
    let before = (snapshot(&fx.checkout), snapshot(&fx.home));
    let mut denied = 0;
    let mut context = 0;
    for (name, what, input) in &set {
        let (code, stdout, _stderr) = fx.hook(name, input);
        assert_eq!(code, 0, "{name} {what}");
        assert_no_allow_or_ask(&stdout, &format!("{name} {what}"));
        denied += stdout.contains("\"deny\"") as usize;
        context += stdout.contains("additionalContext") as usize;
    }
    assert!(denied >= 4, "the set must hold denials, found {denied}");
    assert!(context >= 1, "the set must hold a session context");
    assert_eq!(before, (snapshot(&fx.checkout), snapshot(&fx.home)));
}

// ---------------------------------------------------------------------- AC-11

/// AC-11: the plan and the written default settings of `moai hooks-plan` hold none of the three
/// events, and no settings file of the temporary project names a new hook entry after the run.
#[test]
fn ac11_the_default_hook_set_installs_no_worktree_or_stop_event() {
    let home = TempHome::new();
    let project = home.project(PROJECT);
    let root = project.to_str().unwrap();
    let run = |args: &[&str]| -> Value {
        let out = home
            .command(args)
            .current_dir(&project)
            .stdin(Stdio::null())
            .output()
            .unwrap();
        assert_eq!(
            out.status.code(),
            Some(0),
            "{}",
            String::from_utf8_lossy(&out.stdout)
        );
        serde_json::from_slice(&out.stdout).unwrap()
    };
    let plan = run(&["hooks-plan", root, "--host", "claude"]);
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_owned();
    let applied = run(&[
        "hooks-apply",
        root,
        "--host",
        "claude",
        "--plan-digest",
        &digest,
        "--permission-reference",
        "conformance-test",
    ]);
    assert_eq!(applied["result"]["state"], json!("applied"));
    let settings = fs::read_to_string(project.join(".claude/settings.json")).unwrap();
    let written: Value = serde_json::from_str(&settings).unwrap();
    let mut events: Vec<&str> = written["hooks"]
        .as_object()
        .unwrap()
        .keys()
        .map(String::as_str)
        .collect();
    events.sort_unstable();
    assert_eq!(events, ["PostToolUse", "PreToolUse", "SessionStart"]);
    // The planned hooks carry the same three events, and no text of the plan's hook list or of the
    // written settings names the two worktree events or one of the new entry names. (The plan's
    // engine capability prose may say "Stop"; only the hook list and the settings are the set.)
    let mut planned: Vec<&str> = plan["result"]["hooks"]
        .as_array()
        .unwrap()
        .iter()
        .map(|h| h["event"].as_str().unwrap())
        .collect();
    planned.sort_unstable();
    planned.dedup();
    assert_eq!(planned, ["PostToolUse", "PreToolUse", "SessionStart"]);
    for text in [plan["result"]["hooks"].to_string(), settings] {
        for word in [
            "WorktreeCreate",
            "WorktreeRemove",
            "write-boundary",
            "bash-boundary",
            "session-context",
        ] {
            assert!(!text.contains(word), "a plan or settings text names {word}");
        }
    }
}

// ---------------------------------------------------------------------- AC-12

const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";

fn git_executable() -> PathBuf {
    let path = std::env::var_os("PATH").expect("PATH");
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
        .expect("a git on PATH")
}

struct Guards {
    home: TempHome,
    repo: PathBuf,
    _dir: tempfile::TempDir,
}

impl Guards {
    fn path_var() -> String {
        std::env::join_paths([
            git_executable().parent().unwrap().to_path_buf(),
            PathBuf::from("/usr/bin"),
            PathBuf::from("/bin"),
        ])
        .unwrap()
        .to_string_lossy()
        .into_owned()
    }

    fn git_command(&self, args: &[&str]) -> Command {
        let mut cmd = Command::new(git_executable());
        cmd.arg("-C")
            .arg(&self.repo)
            .args([
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "-c",
                "gc.auto=0",
                "-c",
                "maintenance.auto=false",
            ])
            .args(args)
            .clean_env()
            .env("HOME", self.home.path())
            .env("PATH", Self::path_var())
            .env("LC_ALL", "C")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(Stdio::null());
        cmd
    }

    fn git(&self, args: &[&str]) -> String {
        let out = self.git_command(args).output().unwrap();
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_owned()
    }

    fn moai(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.current_dir(&self.repo)
            .env("PATH", Self::path_var())
            .env("MOAI_GIT_BIN", git_executable())
            .env("LC_ALL", "C")
            .stdin(Stdio::null());
        if let Some(id) = session {
            cmd.env("CLAUDE_CODE_SESSION_ID", id);
        }
        cmd.output().expect("run moai")
    }
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "no envelope ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

/// AC-12: with Claude Code hooks disabled (`disableAllHooks` true, no hook entry registered), the
/// guards that do not read a hook still refuse: the installed reference-transaction guard refuses
/// a hand move of the integration branch, a path outside the card's reserved scope is not
/// allowed, and a commit without the trailers fails `moai trailer-check`. The two classes that
/// only a hook refuses (a write into a protected path, secret content) are printed as the named
/// gap of the SPEC.
#[test]
fn ac12_the_guards_that_do_not_read_a_hook_refuse_with_hooks_disabled() {
    let home = TempHome::new();
    let dir = tempfile::Builder::new()
        .prefix("moai-test-conformance-")
        .tempdir()
        .unwrap();
    let repo = dir.path().join("r");
    fs::create_dir_all(&repo).unwrap();
    let g = Guards {
        home,
        repo: repo.clone(),
        _dir: dir,
    };
    // The throwaway project: a profile naming the integration branch, hooks off.
    g.git(&["init", "-q", "-b", "main"]);
    fs::write(repo.join("a.txt"), "a\n").unwrap();
    g.git(&["add", "a.txt"]);
    g.git(&["commit", "-q", "-m", "init"]);
    let a = g.git(&["rev-parse", "HEAD"]);
    g.git(&["branch", "NeXT", &a]);
    let tree = g.git(&["rev-parse", &format!("{a}^{{tree}}")]);
    let b = g.git(&["commit-tree", &tree, "-p", &a, "-m", "the lane head"]);
    g.git(&["checkout", "-q", "--detach"]);
    let profile = format!(
        "{{\"schema_version\": 1, \"project_id\": \"{PROJECT}\", \"purpose\": \"fixture\", \
         \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \
         \"work_flow\": {{\"flow\": \"next-flow\", \"integration_branch\": \"NeXT\"}}}}"
    );
    let cfg = moai_cli::state_root::path(&repo, "project.yaml");
    fs::create_dir_all(cfg.parent().unwrap()).unwrap();
    fs::write(&cfg, profile).unwrap();
    fs::create_dir_all(repo.join(".claude")).unwrap();
    fs::write(
        repo.join(".claude/settings.json"),
        "{\"disableAllHooks\": true}\n",
    )
    .unwrap();
    let bin = moai_cli::home_root::path(g.home.path()).join("bin");
    fs::create_dir_all(&bin).unwrap();
    std::os::unix::fs::symlink(common::MOAI_BIN, bin.join("moai")).unwrap();
    ledger::register(g.home.path(), PROJECT).unwrap();

    // 1. The reference-transaction guard, installed by the product's own installer.
    let plan = envelope(&g.moai(None, &["git-hooks-plan", repo.to_str().unwrap()]));
    let token = plan["result"]["confirm_token"]
        .as_str()
        .expect("a token")
        .to_owned();
    let applied = g.moai(
        None,
        &[
            "git-hooks-apply",
            repo.to_str().unwrap(),
            "--confirm-git-hooks",
            &token,
        ],
    );
    assert_eq!(
        applied.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&applied.stdout)
    );
    let moved = g
        .git_command(&["update-ref", "refs/heads/NeXT", &b])
        .output()
        .unwrap();
    assert!(!moved.status.success(), "the hand move was not refused");
    assert!(String::from_utf8_lossy(&moved.stderr).contains("moai ref guard"));
    assert_eq!(
        g.git(&["rev-parse", "refs/heads/NeXT"]),
        a,
        "the branch moved"
    );

    // 2. The scope check of a card: a path outside the reserved scope is not allowed.
    let project = repo.to_str().unwrap();
    let added = envelope(&g.moai(
        None,
        &[
            "card",
            "add",
            "--title",
            "scope",
            "--scope",
            "src/card/",
            "--priority",
            "p2",
            "--ac",
            "it works",
            "--project",
            project,
        ],
    ));
    let card = added["result"]["name"]
        .as_str()
        .expect("a card name")
        .to_owned();
    for step in [["ready", card.as_str()], ["take", card.as_str()]] {
        let mut args = vec!["card", step[0], step[1]];
        args.extend(["--project", project]);
        let out = g.moai(Some(S_A), &args);
        assert_eq!(
            out.status.code(),
            Some(0),
            "{}",
            String::from_utf8_lossy(&out.stdout)
        );
    }
    let check = |path: &str| {
        g.moai(
            Some(S_A),
            &[
                "card",
                "check",
                &card,
                "--generation",
                "1",
                "--path",
                path,
                "--project",
                project,
            ],
        )
    };
    let inside = check("src/card/work.rs");
    assert_eq!(inside.status.code(), Some(0));
    assert_eq!(envelope(&inside)["result"]["allowed"], json!(true));
    let outside = check("src/ledger/workflow.rs");
    assert_ne!(
        outside.status.code(),
        Some(0),
        "a path outside the scope was allowed"
    );

    // 3. The trailer check: a commit of a SPEC branch without the trailers fails.
    g.git(&["checkout", "-q", "-b", "spec/guard-20261126", &a]);
    g.git(&["commit", "-q", "--allow-empty", "-m", "no trailers"]);
    let range = format!("{a}..HEAD");
    let checked = g.moai(None, &["trailer-check", &range, "--dir", project]);
    assert_eq!(
        checked.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&checked.stdout)
    );
    assert_eq!(envelope(&checked)["result"]["ok"], json!(false));

    // The named gap: nothing refuses these two with hooks off (SPEC R-2).
    println!(
        "named gap with hooks disabled: a write into a protected path and secret content in a \
         file are refused only by a hook; no hook-independent guard covers them yet"
    );
}
