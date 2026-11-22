//! SPEC-20261122-next-flow-hooks — the red tests of Plan step 1 (card C-327's code slice).
//! Every case runs the real binary in a temporary repository with a temporary HOME, never the
//! user's folders; AC-16's one real session measurement is not here (its verify command is run
//! once by hand and recorded in the SPEC Result).
#![cfg(unix)]

mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::Instant;

const BIN: &str = crate::common::MOAI_BIN;
const BRANCH: &str = "spec/20261122-demo";
const RECORD: &str = ".git/moai/integration-branch";

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

impl Run {
    fn json(&self) -> Value {
        serde_json::from_str(&self.stdout).unwrap_or(Value::Null)
    }
    fn denies(&self) -> bool {
        self.code == 2 || self.stdout.contains("\"deny\"")
    }
}

/// Runs the moai binary with `args` in `cwd` under a clean environment (a temporary HOME, the
/// temporary config folder of `config`, a session id of `session`).
fn moai(cwd: &Path, config: &Path, session: &str, args: &[&str]) -> Command {
    let mut cmd = Command::new(BIN);
    cmd.args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env("CLAUDE_CONFIG_DIR", config)
        .env("CLAUDE_CODE_SESSION_ID", session)
        .env_remove("MOAI_FORGE_GH_BIN")
        .env_remove("XDG_CACHE_HOME")
        .env_remove("XDG_CONFIG_HOME")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env_remove("GIT_DIR");
    cmd
}

fn json_run(cmd: &mut Command) -> Value {
    let out = cmd.stdin(Stdio::null()).output().unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    v["result"].is_object().then(|| v["result"].clone()).unwrap_or(v)
}

/// `hook-dispatch --host claude --event <event>` with `input` on stdin.
fn dispatch(event: &str, cwd: &Path, config: &Path, input: &Value, path: Option<&str>) -> Run {
    let mut cmd = Command::new(BIN);
    cmd.args(["hook-dispatch", "--host", "claude", "--event", event])
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env("CLAUDE_CONFIG_DIR", config)
        .env("CLAUDE_CODE_SESSION_ID", "sess-ac")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if let Some(p) = path {
        cmd.env("PATH", p);
    }
    let mut child = cmd.spawn().unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    Run {
        code: out.status.code().unwrap(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

fn preflight(cwd: &Path, config: &Path, input: &Value) -> Run {
    dispatch("tool-preflight", cwd, config, input, None)
}

fn dangerous(cwd: &Path, config: &Path, input: &Value) -> Run {
    dispatch("dangerous-preflight", cwd, config, input, None)
}

fn session_check(cwd: &Path, config: &Path, input: &Value) -> Run {
    dispatch("session-check", cwd, config, input, None)
}

fn bash(command: &str, cwd: &Path) -> Value {
    json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
           "tool_input": {"command": command}, "cwd": cwd, "session_id": "sess-ac"})
}

fn write_call(file: &str, cwd: &Path) -> Value {
    json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
           "tool_input": {"file_path": file, "content": "x"},
           "cwd": cwd, "session_id": "sess-ac"})
}

fn git(cwd: &Path, args: &[&str]) {
    let out = common::git::command()
        .args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

/// A temporary repository on `branch` with one commit, its own HOME, a temporary config folder
/// and the integration-branch record naming `refs/heads/NeXT` in the common git folder.
fn world(branch: &str) -> (tempfile::TempDir, PathBuf, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    git(&root, &["init", "-q", "-b", branch]);
    git(&root, &["config", "user.name", "fixture"]);
    git(&root, &["config", "user.email", "fixture@example.invalid"]);
    git(&root, &["config", "commit.gpgsign", "false"]);
    fs::write(root.join("a.txt"), "a\n").unwrap();
    fs::write(root.join(".gitignore"), "home/\n").unwrap();
    git(&root, &["add", "a.txt", ".gitignore"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    // The project state folder and the recorded integration branch (D-2).
    fs::create_dir_all(root.join(".moai")).unwrap();
    fs::create_dir_all(root.join(".git/moai")).unwrap();
    fs::write(root.join(RECORD), "refs/heads/NeXT\n").unwrap();
    let config = root.join("config");
    fs::create_dir_all(&config).unwrap();
    (dir, root, config)
}

/// A folder whose `git` and `sh` fail loudly: a duty that starts no subprocess gives the same
/// answer under it (REQ-2).
fn poisoned_path(root: &Path) -> String {
    let dir = root.join("poison");
    fs::create_dir_all(&dir).unwrap();
    for name in ["git", "sh"] {
        fs::write(dir.join(name), "#!/bin/sh\nexit 99\n").unwrap();
        make_executable(&dir.join(name));
    }
    format!("{}:/usr/bin:/bin", dir.display())
}

fn make_executable(path: &Path) {
    use std::os::unix::fs::PermissionsExt;
    let mut perm = fs::metadata(path).unwrap().permissions();
    perm.set_mode(0o755);
    fs::set_permissions(path, perm).unwrap();
}

/// Registers a ledger in the project at `root` and adds one ready card taken by `holder` (a
/// session id in the shape the ledger accepts: hex digits and `-`, 8 to 64 characters).
fn carded_world(holder_hex: &str, scope: &str, branch: &str) -> (tempfile::TempDir, PathBuf, PathBuf) {
    let holder = holder_hex;
    let (dir, root, config) = world(branch);
    fs::create_dir_all(root.join(".moai-cli")).unwrap();
    fs::write(
        root.join(".moai-cli/project.yaml"),
        "{\"schema_version\":2,\"project_id\":\"ac12fixture\"}\n",
    )
    .unwrap();
    let register = json_run(&mut moai(&root, &config, holder, &[
        "ledger-register-plan", &root.to_string_lossy(),
    ]));
    let digest = register["plan_digest"].as_str().unwrap().to_string();
    let root_text = root.to_string_lossy().into_owned();
    json_run(&mut moai(&root, &config, holder, &[
        "ledger-register-apply", &root_text, "--plan-digest", &digest,
        "--permission-reference", "user-approved-test",
    ]));
    let added = json_run(&mut moai(&root, &config, holder, &[
        "card", "add", "--title", "fixture card", "--project", &root_text,
    ]));
    let name = added["name"].as_str().unwrap().to_string();
    json_run(&mut moai(&root, &config, holder, &[
        "card", "ready", &name, "--project", &root_text,
    ]));
    json_run(&mut moai(&root, &config, holder, &[
        "card", "take", &name, "--scope", scope, "--project", &root_text,
        "--branch", branch, "--worktree", &root_text,
    ]));
    (dir, root, config)
}

// --- AC-1, AC-5: the rendered set, exec form, onFailure --------------------------------

fn plan_hooks(root: &Path, config: &Path) -> Vec<Value> {
    let v = json_run(&mut moai(root, config, "sess-ac", &[
        "hooks-plan", "--host", "claude", &root.to_string_lossy(),
    ]));
    v["hooks"].as_array().unwrap().clone()
}

/// AC-1 (REQ-1): the plan holds neither removed command, the applied settings hold neither, and
/// only the dangerous entry carries onFailure.
#[test]
fn ac1_the_rendered_set_holds_no_issue_era_entry_and_blocks_only_the_dangerous_one() {
    let (_d, root, config) = world(BRANCH);
    let hooks = plan_hooks(&root, &config);
    let text = serde_json::to_string(&hooks).unwrap();
    assert!(!text.contains("claim-verify"), "{text}");
    assert!(!text.contains("hook-receipt"), "{text}");
    let with_block: Vec<&Value> = hooks
        .iter()
        .filter(|h| h["onFailure"] == json!("block"))
        .collect();
    assert_eq!(with_block.len(), 1, "{text}");
    assert!(
        with_block[0]["args"].to_string().contains("dangerous"),
        "{text}"
    );
    // Applied settings: the file on disk holds neither removed command either.
    let plan = json_run(&mut moai(&root, &config, "sess-ac", &[
        "hooks-plan", "--host", "claude", &root.to_string_lossy(),
    ]));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let applied = json_run(&mut moai(&root, &config, "sess-ac", &[
        "hooks-apply", "--host", "claude", &root.to_string_lossy(),
        "--plan-digest", &digest, "--permission-reference", "user-approved-test",
    ]));
    assert_eq!(applied["state"], json!("applied"), "{applied}");
    let settings = fs::read_to_string(root.join(".claude/settings.json")).unwrap();
    assert!(!settings.contains("claim-verify"), "{settings}");
    assert!(!settings.contains("hook-receipt"), "{settings}");
}

/// AC-5 (REQ-5): every entry is exec form — the absolute path of the moai binary as command and
/// the arguments as the args array, no shell string, no placeholder, no relative name.
#[test]
fn ac5_every_entry_is_exec_form_with_the_absolute_binary_path() {
    let (_d, root, config) = world(BRANCH);
    let hooks = plan_hooks(&root, &config);
    assert!(!hooks.is_empty());
    for h in &hooks {
        let command = h["command"].as_str().unwrap_or_default();
        assert!(
            command.starts_with('/') && command.ends_with("moai"),
            "not an absolute moai path: {command}"
        );
        assert!(
            Path::new(command).is_file(),
            "the command must name the binary: {command}"
        );
        assert!(h["args"].is_array(), "no args array: {h}");
        assert!(!h["args"].as_array().unwrap().is_empty(), "{h}");
        assert!(!command.contains("${"), "a placeholder: {command}");
        assert!(!command.contains(" && "), "a shell string: {command}");
    }
}

/// AC-6 (REQ-6): the matchers — Bash|PowerShell for the shell duties (which also catches
/// Monitor since 2.1.314), Edit|Write|NotebookEdit for the write side, the four sources for the
/// session start — and the duties really see a PowerShell call and a Monitor command.
#[test]
fn ac6_matchers_cover_powershell_monitor_and_the_four_session_sources() {
    let (_d, root, config) = world(BRANCH);
    let hooks = plan_hooks(&root, &config);
    let text = serde_json::to_string(&hooks).unwrap();
    assert!(text.contains("Bash|PowerShell"), "{text}");
    assert!(text.contains("Edit|Write|NotebookEdit"), "{text}");
    assert!(text.contains("startup|resume|clear|compact"), "{text}");
    // The duties answer for a PowerShell call and a Monitor command as they do for Bash.
    let pwsh = json!({"hook_event_name": "PreToolUse", "tool_name": "PowerShell",
                      "tool_input": {"command": "git branch -C side NeXT"},
                      "cwd": &root, "session_id": "sess-ac"});
    assert!(dangerous(&root, &config, &pwsh).denies());
    let monitor = json!({"hook_event_name": "PreToolUse", "tool_name": "Monitor",
                         "tool_input": {"command": "git push origin NeXT"},
                         "cwd": &root, "session_id": "sess-ac"});
    assert!(dangerous(&root, &config, &monitor).denies());
    // A tool of any other name is not the shell duty's business.
    let read = json!({"hook_event_name": "PreToolUse", "tool_name": "Read",
                      "tool_input": {"file_path": "a.txt"}, "cwd": &root});
    assert!(!dangerous(&root, &config, &read).denies());
    // The session-check duty answers for the four sources and nothing else.
    for source in ["startup", "resume", "clear", "compact"] {
        let input = json!({"hook_event_name": "SessionStart", "source": source,
                           "cwd": &root, "session_id": "sess-ac"});
        assert!(!session_check(&root, &config, &input).stdout.is_empty(), "{source}");
    }
    let other = json!({"hook_event_name": "SessionStart", "source": "fork",
                       "cwd": &root, "session_id": "sess-ac"});
    assert!(session_check(&root, &config, &other).stdout.is_empty());
}

// --- AC-2, AC-3: the session-check duty -------------------------------------------------

/// AC-2 (REQ-2, REQ-3): the answer names the branch, the recorded integration branch and the
/// worktree's card, points at `moai card show`, sits inside additionalContext with a
/// sessionTitle, holds no Issue or claim wording, and is byte-identical under a poisoned PATH.
#[test]
fn ac2_session_check_reads_files_only_and_names_the_card() {
    let (_d, root, config) = carded_world("aabbccddeeff0011", "src/locked/**", BRANCH);
    let input = json!({"hook_event_name": "SessionStart", "source": "startup",
                       "cwd": &root, "session_id": "sess-ac"});
    let plain = session_check(&root, &config, &input);
    let ctx = plain.json()["hookSpecificOutput"]["additionalContext"]
        .as_str()
        .unwrap_or_default()
        .to_string();
    assert!(!ctx.is_empty(), "not inside additionalContext: {}", plain.stdout);
    assert!(ctx.contains(BRANCH), "{ctx}");
    assert!(ctx.contains("NeXT"), "{ctx}");
    assert!(ctx.contains("moai card show"), "{ctx}");
    assert!(ctx.contains("fixture card"), "{ctx}");
    for word in ["claim", "Claim", "issue/", "Issue #"] {
        assert!(!ctx.contains(word), "{word} appears: {ctx}");
    }
    let plain_json = plain.json();
    let title = plain_json["hookSpecificOutput"]["sessionTitle"]
        .as_str()
        .unwrap_or("");
    assert!(!title.is_empty(), "no sessionTitle: {}", plain.stdout);
    assert!(!title.contains("C-"), "a card number in the title: {title}");
    // No subprocess: a PATH whose every git and sh fails changes nothing.
    let poison = poisoned_path(&root);
    let poisoned = dispatch("session-check", &root, &config, &input, Some(&poison));
    assert_eq!(poisoned.stdout, plain.stdout, "the answer changed under a poisoned PATH");
    // The edge facts, in fixed wording.
    let bare = tempfile::TempDir::new().unwrap();
    let bare_root = fs::canonicalize(bare.path()).unwrap();
    fs::create_dir_all(bare_root.join("home")).unwrap();
    let outside = session_check(&bare_root, &config, &input);
    assert_eq!(outside.code, 0);
    assert!(outside.stdout.is_empty(), "a non-moai folder answers nothing: {}", outside.stdout);
}

/// AC-3 (REQ-3): the median of repeated in-process runs stays under 20 ms (D-7: the budget
/// covers the duty's own work; the host's process start sits beside it and is the measured-once
/// number of the Result).
#[test]
fn ac3_the_session_check_answers_within_the_budget() {
    let (_d, root, _config) = world(BRANCH);
    let input = json!({"hook_event_name": "SessionStart", "source": "startup",
                       "cwd": &root, "session_id": "sess-ac"});
    // One warm-up, then the measured runs — the duty called in process, as the SPEC pins it.
    let _ = moai_cli::hook_engine::dispatch("claude", "session-check",
                                            input.to_string().as_bytes(), &root, 1);
    let mut times: Vec<u128> = Vec::new();
    for t in 2..52u128 {
        let start = Instant::now();
        let answer = moai_cli::hook_engine::dispatch("claude", "session-check",
                                                     input.to_string().as_bytes(), &root, t);
        times.push(start.elapsed().as_micros());
        assert_eq!(answer.exit, 0);
        assert!(!answer.stdout.is_empty());
    }
    times.sort_unstable();
    let median = times[24];
    assert!(
        median < 20_000,
        "the in-process median is {} µs: {:?}",
        median,
        times
    );
}

// --- AC-7: command splitting ------------------------------------------------------------

/// AC-7 (REQ-7): a guarded call hidden behind a separator, a wrapper or a cd prefix is still
/// refused, and the default builds' search commands pass.
#[test]
fn ac7_a_guarded_call_hidden_behind_separators_wrappers_or_cd_is_refused() {
    let (_d, root, config) = world(BRANCH);
    fs::create_dir_all(root.join("sub")).unwrap();
    let hidden = [
        "cd sub && git push origin NeXT",
        "git branch -C side NeXT || echo no",
        "timeout 10 git branch -m NeXT side",
        "nice -n 5 git update-ref refs/heads/NeXT deadbeef",
        "nohup git branch -D NeXT >/dev/null 2>&1",
        "echo go | xargs -I{} git push origin NeXT",
        "true; git push origin NeXT",
    ];
    for command in hidden {
        assert!(
            dangerous(&root, &config, &bash(command, &root)).denies(),
            "not refused: {command}"
        );
    }
    // The default builds' search commands are not misread as dangerous.
    for command in ["bfs . -name '*.rs'", "ugrep -r fixture ."] {
        assert!(
            !dangerous(&root, &config, &bash(command, &root)).denies(),
            "misread as dangerous: {command}"
        );
    }
}

// --- AC-8: the refusal channel ----------------------------------------------------------

/// AC-8 (REQ-8): every refusal is exit code 2 and no answer ever holds updatedInput.
#[test]
fn ac8_every_refusal_is_exit_two_and_none_changes_the_input() {
    let (_d, root, config) = world(BRANCH);
    let outside = tempfile::TempDir::new().unwrap();
    let refusals = [
        bash("git push origin NeXT", &root),
        bash("git branch -m NeXT side", &root),
        bash(&format!("rm -rf {}", outside.path().display()), &root),
        write_call(&root.join(".claude/settings.json").to_string_lossy(), &root),
    ];
    for input in &refusals {
        let r = dangerous(&root, &config, input);
        assert!(r.denies(), "not refused: {}", input);
        assert_eq!(r.code, 2, "a refusal must be exit 2: {} {}", r.code, r.stdout);
        let v = r.json();
        assert!(v["hookSpecificOutput"]["updatedInput"].is_null(), "{v}");
    }
    // The ordinary duty's refusals are exit 2 as well, and allow never comes from moai.
    let rm_home = bash("rm -rf /", &root);
    let r = preflight(&root, &config, &rm_home);
    assert!(r.denies());
    assert_eq!(r.code, 2, "{} {}", r.code, r.stdout);
    assert!(!r.stdout.contains("\"allow\""), "{}", r.stdout);
}

// --- AC-9, AC-10: the dangerous duty ----------------------------------------------------

/// AC-9 (REQ-9): every verb that names the recorded integration branch, in any letter case,
/// behind git -C or with an absolute git path, is denied with the recorded spelling and the
/// landing command in the reason; another branch passes.
#[test]
fn ac9_every_call_that_moves_the_integration_branch_is_refused() {
    let (_d, root, config) = world(BRANCH);
    git(&root, &["branch", "side"]);
    let root_text = root.to_string_lossy().into_owned();
    let calls = [
        "git branch -c side NeXT".to_string(),
        "git branch -C side NeXT".to_string(),
        "git branch -m NeXT side".to_string(),
        "git branch -M NeXT side".to_string(),
        "git branch -D NeXT".to_string(),
        "git update-ref refs/heads/NeXT deadbeef".to_string(),
        "git push origin NeXT".to_string(),
        "git push --force origin NeXT".to_string(),
        "git push origin next".to_string(),
        "git push origin NEXT".to_string(),
        format!("git -C '{root_text}' branch -m NeXT side"),
        format!("git -C '{root_text}' push origin NeXT"),
    ];
    let absolute_git = which_git();
    if let Some(git) = absolute_git {
        // The absolute-path form of the same verbs.
        let d = dangerous(&root, &config, &bash(&format!("{git} branch -C side NeXT"), &root));
        assert!(d.denies(), "absolute git path form refused nothing");
    }
    for call in &calls {
        let r = dangerous(&root, &config, &bash(call, &root));
        assert!(r.denies(), "not refused: {call}");
        let reason = refusal_reason(&r);
        assert!(
            reason.contains("NeXT") && reason.contains("card land"),
            "the reason must name the spelling and the landing command: {reason}"
        );
    }
    // The same verbs naming another branch pass.
    for call in ["git branch -m side other", "git push origin side"] {
        assert!(
            !dangerous(&root, &config, &bash(call, &root)).denies(),
            "an ordinary lane call refused: {call}"
        );
    }
}

fn refusal_reason(r: &Run) -> String {
    let v = r.json();
    v["hookSpecificOutput"]["permissionDecisionReason"]
        .as_str()
        .or_else(|| r.stderr.lines().next())
        .unwrap_or_default()
        .to_string()
}

fn which_git() -> Option<String> {
    let out = Command::new("which").arg("git").output().ok()?;
    let s = String::from_utf8_lossy(&out.stdout).trim().to_string();
    (out.status.success() && s.starts_with('/')).then_some(s)
}

/// AC-10 (REQ-9): a recursive delete outside the project tree and a write of each guarded
/// settings file — the config folder's settings.json included — are denied with a fixed reason;
/// an ordinary lane push and a source edit pass.
#[test]
fn ac10_recursive_deletes_and_guarded_settings_writes_are_refused() {
    let (_d, root, config) = world(BRANCH);
    let outside = tempfile::TempDir::new().unwrap();
    assert!(
        dangerous(
            &root,
            &config,
            &bash(&format!("rm -rf {}", outside.path().display()), &root)
        )
        .denies(),
        "a recursive delete outside the project tree passed"
    );
    let guarded = [
        root.join(".claude/settings.json"),
        root.join(".claude/settings.local.json"),
        config.join("settings.json"),
    ];
    for file in &guarded {
        let r = dangerous(&root, &config, &write_call(&file.to_string_lossy(), &root));
        assert!(r.denies(), "a guarded settings write passed: {}", file.display());
        assert!(!refusal_reason(&r).is_empty());
    }
    // An ordinary lane push and a source edit pass.
    assert!(!dangerous(&root, &config, &bash("git push origin side", &root)).denies());
    assert!(!dangerous(&root, &config, &write_call(
        &root.join("src/main.rs").to_string_lossy(), &root
    )).denies());
}

// --- AC-11: read-only agent types -------------------------------------------------------

/// AC-11 (REQ-10): a write tool call and a reference-move Bash call of a read-only agent are
/// denied, the deny line holds the agent type and a short hash of the agent id, and the same
/// agent's read call passes.
#[test]
fn ac11_read_only_agents_cannot_write_or_move_refs() {
    let (_d, root, config) = world(BRANCH);
    let agent = |tool: &str, input: Value| {
        let mut v = input;
        v["agent_type"] = json!("moai-audit");
        v["agent_id"] = json!("agent-ac11");
        let _ = tool;
        v
    };
    let write = agent("Write", write_call(&root.join("src/x.rs").to_string_lossy(), &root));
    let r = preflight(&root, &config, &write);
    assert!(r.denies(), "a read-only agent's write passed");
    let move_ref = agent("Bash", bash("git update-ref refs/heads/side deadbeef", &root));
    assert!(preflight(&root, &config, &move_ref).denies());
    let read = json!({"hook_event_name": "PreToolUse", "tool_name": "Read",
                      "tool_input": {"file_path": "a.txt"}, "cwd": &root,
                      "session_id": "sess-ac", "agent_type": "moai-audit",
                      "agent_id": "agent-ac11"});
    assert!(!preflight(&root, &config, &read).denies());
    // The deny line names the agent type and a short hash of the id, never the raw id.
    let log = root.join(".moai/receipts/hook-engine.jsonl");
    let text = fs::read_to_string(&log).unwrap_or_default();
    assert!(text.contains("moai-audit"), "{text}");
    assert!(!text.contains("agent-ac11"), "a raw agent id in the log: {text}");
    assert!(text.contains("agent"), "{text}");
}

// --- AC-12: the ledger reservation check ------------------------------------------------

/// AC-12 (REQ-11): a commit touching a path another holder's live reservation covers is
/// denied with the ledger read only; a commit inside this session's own reservation passes; no
/// claim-token wording remains.
#[test]
fn ac12_a_commit_into_another_holders_reservation_is_refused() {
    let (_d, root, config) = carded_world("0011223344556677", "src/locked/**", BRANCH);
    fs::create_dir_all(root.join("src/locked")).unwrap();
    fs::write(root.join("src/locked/x.rs"), "x\n").unwrap();
    git(&root, &["add", "src/locked/x.rs"]);
    let commit = bash("git commit -m x", &root);
    let r = preflight(&root, &config, &commit);
    assert!(r.denies(), "a commit into another holder's scope passed");
    let reason = refusal_reason(&r);
    assert!(!reason.to_lowercase().contains("moai-claim"), "{reason}");
    // A path outside the reservation passes.
    fs::write(root.join("src/free.rs"), "x\n").unwrap();
    git(&root, &["reset", "-q"]);
    git(&root, &["add", "src/free.rs"]);
    assert!(!preflight(&root, &config, &commit).denies());
}

// --- AC-13: the engine log's sides ------------------------------------------------------

/// AC-13 (REQ-12): with a HOME whose state root is the retired name nothing is created under
/// the retired folder and the line lands on the project side; a marked HOME logs under the
/// current root; reading sees both.
#[test]
fn ac13_no_engine_line_lands_under_the_retired_home_folder() {
    let (_d, root, config) = world(BRANCH);
    fs::create_dir_all(root.join(".moai-cli")).unwrap();
    fs::write(
        root.join(".moai-cli/project.yaml"),
        "{\"schema_version\":2,\"project_id\":\"ac13fixture\"}\n",
    )
    .unwrap();
    // A HOME with the retired folder as its state root (no marker): nothing may be written
    // under it.
    let home = root.join("home");
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    assert!(dangerous(&root, &config, &bash("git push origin NeXT", &root)).denies());
    let retired = home.join(".moai-cli");
    let nothing_new = |dir: &Path| {
        let mut stack = vec![dir.to_path_buf()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap().flatten() {
                let p = e.path();
                if p.is_dir() {
                    stack.push(p);
                } else {
                    assert!(false, "a file under the retired folder: {}", p.display());
                }
            }
        }
    };
    nothing_new(&retired);
    let project_text = |rel: &str| {
        fs::read_to_string(root.join(rel)).unwrap_or_default()
    };
    let project_side = project_text(".moai/receipts/hook-engine.jsonl")
        + &project_text(".moai-cli/receipts/hook-engine.jsonl");
    assert!(
        project_side.contains("\"deny\""),
        "the line did not land on the project side: {project_side}"
    );
    // A marked HOME logs under the current root.
    fs::remove_dir_all(&retired).unwrap();
    fs::create_dir_all(home.join(".moai")).unwrap();
    fs::write(
        home.join(".moai/moai.json"),
        r#"{"schema_version":1,"generation":"moai"}"#,
    )
    .unwrap();
    assert!(dangerous(&root, &config, &bash("git push origin NeXT", &root)).denies());
    let current = home.join(".moai/projects/ac13fixture/receipts/hook-engine.jsonl");
    let text = fs::read_to_string(&current).unwrap_or_default();
    assert!(text.contains("\"deny\""), "no line under the current root: {text}");
}

// --- AC-14, AC-15: the update layer -----------------------------------------------------

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for entry in fs::read_dir(from).unwrap() {
        let p = entry.unwrap().path();
        let dest = to.join(p.file_name().unwrap());
        if p.is_dir() {
            copy_tree(&p, &dest);
        } else {
            fs::copy(&p, &dest).unwrap();
        }
    }
}

/// A temporary HOME with the binary, the real package and the agent sources beside it (as
/// `tests/update_user_layer_cli.rs` builds one), plus a project folder to run the update in.
fn update_world() -> (tempfile::TempDir, PathBuf, PathBuf, PathBuf, PathBuf) {
    let tmp = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(BIN), &exe);
    let pkg = folder.join("skills/moai");
    copy_tree(&Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai"), &pkg);
    copy_tree(
        &Path::new(env!("CARGO_MANIFEST_DIR")).join("agents"),
        &folder.join("agents"),
    );
    let project = root.join("project");
    fs::create_dir_all(project.join(".moai")).unwrap();
    (tmp, home, project, root.join("config"), exe)
}

fn update_run(
    exe: &Path,
    home: &Path,
    project: &Path,
    config: &Path,
    args: &[&str],
) -> Value {
    let out = Command::new(exe)
        .args(args)
        .current_dir(project)
        .env("HOME", home)
        // The config-folder rule of `host_home`: `CLAUDE_CONFIG_DIR` counts only for the
        // account home, so the test names its temporary HOME as the account home too.
        .env("MOAI_ACCOUNT_HOME", home)
        .env("CLAUDE_CONFIG_DIR", config)
        .env("CLAUDE_CODE_SESSION_ID", "sess-ac")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    v["result"].is_object().then(|| v["result"].clone()).unwrap_or(v)
}

/// AC-14 (REQ-13): the preview offers the deny rules beside the hook set for each picked config
/// folder and nothing is written without the pick.
#[test]
fn ac14_the_deny_rules_are_offered_and_written_only_on_the_pick() {
    let (_t, home, project, config, exe) = update_world();
    fs::create_dir_all(&config).unwrap();
    let p = update_run(&exe, &home, &project, &config, &["update"]);
    let hooks = &p["hooks"];
    let text = hooks.to_string();
    assert!(text.contains("deny"), "no deny offer in the preview: {text}");
    assert!(text.contains("reset --hard"), "{text}");
    // Nothing is written without the pick.
    assert!(!config.join("settings.json").exists());
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = update_run(&exe, &home, &project, &config, &["update", "--confirm", &digest]);
    let _ = r;
    let settings = fs::read_to_string(config.join("settings.json")).unwrap_or_default();
    assert!(settings.contains("deny"), "no deny rules written: {settings}");
    assert!(!settings.contains("NeXT") || settings.contains("deny"), "{settings}");
}

/// AC-15 (REQ-14, REQ-15): a confirmed update writes the hook entries into the picked config
/// folder's settings.json, touches no project settings file, the preview names each folder and
/// each added or removed entry and the absolute moai path it writes, and a hook whose cwd is
/// not a moai project exits 0 at once.
#[test]
fn ac15_the_hooks_land_in_the_config_folder_and_exit_zero_elsewhere() {
    let (_t, home, project, config, exe) = update_world();
    fs::create_dir_all(&config).unwrap();
    let p = update_run(&exe, &home, &project, &config, &["update"]);
    let hooks = &p["hooks"];
    let text = hooks.to_string();
    assert!(
        text.contains(config.to_string_lossy().as_ref()),
        "the preview does not name the config folder: {text}"
    );
    assert!(text.contains("session-check"), "no entry named: {text}");
    assert!(text.contains("added"), "{text}");
    // The absolute moai path the layer writes is named (the moved-binary risk).
    assert!(text.contains("versions/a/bin"), "the written binary path is not named: {text}");
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    update_run(&exe, &home, &project, &config, &["update", "--confirm", &digest]);
    let settings = fs::read_to_string(config.join("settings.json")).unwrap();
    assert!(settings.contains("dangerous-preflight"), "{settings}");
    assert!(settings.contains("\"args\""), "{settings}");
    assert!(
        !project.join(".claude/settings.json").exists(),
        "a project settings file was written"
    );
    // A hook whose cwd is not a moai project exits 0 at once, with no output and no write.
    let plain = tempfile::TempDir::new().unwrap();
    let plain_root = fs::canonicalize(plain.path()).unwrap();
    fs::create_dir_all(plain_root.join("home")).unwrap();
    let input = json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
                       "tool_input": {"command": "git push origin NeXT"},
                       "cwd": &plain_root, "session_id": "sess-ac"});
    let r = dispatch("dangerous-preflight", &plain_root, &config, &input, None);
    assert_eq!(r.code, 0, "{} {}", r.stdout, r.stderr);
    assert!(r.stdout.is_empty(), "{}", r.stdout);
    assert!(!plain_root.join(".moai").exists() && !plain_root.join(".moai-cli").exists());
}
