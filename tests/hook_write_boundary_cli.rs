//! SPEC-20261126-hook-write-boundary: `moai hook write-boundary` (H3). Every case runs the real
//! binary against a temporary world: a primary checkout, linked worktrees (the session's own, a
//! sibling and one nested in the primary), links into other checkouts and into a plain folder, a
//! scratch folder, a plain folder and a HOME with the state and Claude folders. No file of a case
//! is ever written: the hook only judges the path and the text. Secret-shaped values are
//! assembled at run time so no line of this file holds one whole.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::Instant;

const CLASSES: [&str; 3] = ["protected-path", "other-checkout", "secret-content"];
const TOOLS: [&str; 3] = ["Write", "Edit", "NotebookEdit"];

// ---------------------------------------------------------------- the world

#[derive(Clone, Copy)]
enum S {
    /// The session's own linked worktree.
    Own,
    /// A session in a subfolder of its own worktree.
    Sub,
    /// A session in the primary checkout.
    Pri,
    /// A session whose `cwd` lies in no checkout (the project folder is set by the variable).
    Nowhere,
}

struct World {
    _tmp: tempfile::TempDir,
    root: PathBuf,
    home: PathBuf,
    primary: PathBuf,
    own: PathBuf,
    other: PathBuf,
    inner: PathBuf,
    scratch: PathBuf,
    plain: PathBuf,
}

fn git(dir: &Path, args: &[&str]) {
    let out = Command::new("git")
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args([
            "-c",
            "user.name=Fixture",
            "-c",
            "user.email=fixture@example.invalid",
        ])
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .output()
        .expect("spawn git");
    assert!(
        out.status.success(),
        "git {args:?} failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

impl World {
    fn new() -> World {
        let tmp = tempfile::Builder::new().prefix("hwb-").tempdir().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let primary = root.join("primary");
        let home = root.join("home");
        let scratch = root.join("scratch");
        let plain = root.join("plain");
        for dir in [&primary, &home, &scratch, &plain] {
            fs::create_dir_all(dir).unwrap();
        }
        let home_state = home.join(moai_cli::state_root::MOAI_DIR);
        fs::create_dir_all(home_state.join("reports/demo")).unwrap();
        fs::create_dir_all(home.join(".claude/projects/p/memory")).unwrap();
        git(&primary, &["init", "-q", "-b", "NeXT"]);
        fs::create_dir_all(moai_cli::state_root::path(&primary, "")).unwrap();
        fs::write(
            moai_cli::state_root::path(&primary, "project.yaml"),
            "{\"schema_version\":2,\"project_id\":\"demo\",\"work_flow\":{\"flow\":\"m-flow\",\"integration_branch\":\"NeXT\"}}\n",
        )
        .unwrap();
        for dir in ["src", "docs"] {
            fs::create_dir_all(primary.join(dir)).unwrap();
        }
        fs::write(primary.join("src/a.txt"), "a\n").unwrap();
        fs::write(primary.join("docs/d.md"), "d\n").unwrap();
        git(&primary, &["add", "-A"]);
        git(&primary, &["commit", "-q", "-m", "fixture"]);
        let own = root.join("wt-own");
        let other = root.join("wt-other");
        let inner = primary.join(".claude/worktrees/inner");
        for (branch, dir) in [
            ("lane/own", &own),
            ("lane/other", &other),
            ("lane/inner", &inner),
        ] {
            git(
                &primary,
                &["worktree", "add", "-q", "-b", branch, dir.to_str().unwrap()],
            );
        }
        let link = |target: &Path, name: &str| {
            std::os::unix::fs::symlink(target, own.join(name)).unwrap();
        };
        link(&primary.join("src"), "lnk-primary");
        link(&other.join("src"), "lnk-other");
        link(&plain, "lnk-plain");
        link(&home.join(".ssh"), "lnk-ssh");
        World {
            _tmp: tmp,
            root,
            home,
            primary,
            own,
            other,
            inner,
            scratch,
            plain,
        }
    }

    fn fill(&self, text: &str) -> String {
        let pairs: [(&str, String); 9] = [
            (
                "{HOMESTATE}",
                self.home
                    .join(moai_cli::state_root::MOAI_DIR)
                    .display()
                    .to_string(),
            ),
            ("{OWN}", self.own.display().to_string()),
            ("{PRI}", self.primary.display().to_string()),
            ("{OTH}", self.other.display().to_string()),
            ("{INNER}", self.inner.display().to_string()),
            ("{SCR}", self.scratch.display().to_string()),
            ("{HOME}", self.home.display().to_string()),
            ("{PLAIN}", self.plain.display().to_string()),
            ("{ROOT}", self.root.display().to_string()),
        ];
        let mut out = text.to_owned();
        for (key, value) in pairs {
            out = out.replace(key, &value);
        }
        out
    }

    /// The folder the hook process starts in, the `cwd` of the hook input and the project folder
    /// variable for a session.
    fn session(&self, s: S) -> (PathBuf, PathBuf, Option<PathBuf>) {
        match s {
            S::Own => (self.own.clone(), self.own.clone(), None),
            S::Sub => (self.own.clone(), self.own.join("src"), None),
            S::Pri => (self.primary.clone(), self.primary.clone(), None),
            S::Nowhere => (self.own.clone(), self.plain.clone(), Some(self.own.clone())),
        }
    }

    fn spawn(&self, proc_cwd: &Path, stdin: &[u8], project_dir: Option<&Path>) -> Run {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args(["hook", "write-boundary"])
            .current_dir(proc_cwd)
            .env("HOME", &self.home)
            .env_remove("CLAUDE_PROJECT_DIR")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        if let Some(dir) = project_dir {
            cmd.env("CLAUDE_PROJECT_DIR", dir);
        }
        let mut child = cmd.spawn().unwrap();
        common::feed_stdin(&mut child, stdin);
        let out = child.wait_with_output().unwrap();
        Run {
            code: out.status.code().unwrap(),
            stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
            stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        }
    }

    fn input(&self, tool: &str, path: &str, content: &str, cwd: &Path, mode: &str) -> Value {
        json!({
            "session_id": "fixture",
            "cwd": cwd,
            "scratchpad_dir": self.scratch,
            "permission_mode": mode,
            "hook_event_name": "PreToolUse",
            "tool_name": tool,
            "tool_input": tool_input(tool, &self.fill(path), content),
        })
    }

    fn run(&self, s: S, tool: &str, path: &str, content: &str, mode: &str) -> Run {
        let (proc_cwd, cwd, project) = self.session(s);
        let input = self.input(tool, path, content, &cwd, mode);
        self.spawn(&proc_cwd, input.to_string().as_bytes(), project.as_deref())
    }

    fn verdict(&self, s: S, tool: &str, path: &str, content: &str) -> Option<String> {
        let run = self.run(s, tool, path, content, "default");
        parse_verdict(&run, &format!("{tool} {path}"))
    }

    fn raw(&self, s: S, input: &Value) -> Run {
        let (proc_cwd, _, project) = self.session(s);
        self.spawn(&proc_cwd, input.to_string().as_bytes(), project.as_deref())
    }
}

/// The `tool_input` object of each tool for one path and one new content.
fn tool_input(tool: &str, path: &str, content: &str) -> Value {
    match tool {
        "Write" => json!({ "file_path": path, "content": content }),
        "Edit" => json!({
            "file_path": path,
            "old_string": "a",
            "new_string": content,
            "replace_all": false,
        }),
        _ => json!({
            "notebook_path": path,
            "new_source": content,
            "cell_id": "c1",
            "edit_mode": "replace",
        }),
    }
}

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

/// `None` is no decision; `Some(class)` is a deny with that class word. Any other output fails.
fn parse_verdict(run: &Run, what: &str) -> Option<String> {
    assert_eq!(run.code, 0, "exit code for {what:?}: {}", run.stderr);
    assert!(run.stderr.is_empty(), "stderr for {what:?}: {}", run.stderr);
    if run.stdout.trim().is_empty() {
        return None;
    }
    let value: Value = serde_json::from_str(run.stdout.trim())
        .unwrap_or_else(|e| panic!("stdout for {what:?} is not JSON ({e}): {}", run.stdout));
    let object = value.as_object().expect("an object");
    assert_eq!(object.len(), 1, "one top key for {what:?}");
    let inner = object["hookSpecificOutput"]
        .as_object()
        .expect("hookSpecificOutput");
    let mut keys: Vec<&str> = inner.keys().map(String::as_str).collect();
    keys.sort_unstable();
    assert_eq!(
        keys,
        [
            "hookEventName",
            "permissionDecision",
            "permissionDecisionReason"
        ],
        "keys for {what:?}"
    );
    assert_eq!(inner["hookEventName"], "PreToolUse");
    assert_eq!(
        inner["permissionDecision"], "deny",
        "only deny is ever printed"
    );
    let reason = inner["permissionDecisionReason"]
        .as_str()
        .expect("a reason");
    let rest = reason
        .strip_prefix("moai-boundary/")
        .unwrap_or_else(|| panic!("reason prefix for {what:?}: {reason}"));
    let class = rest.split(':').next().unwrap();
    assert!(CLASSES.contains(&class), "class {class:?} for {what:?}");
    assert!(reason.len() < 400, "reason length for {what:?}");
    Some(class.to_owned())
}

/// A token of the given prefix and `n` accepted characters, assembled at run time.
fn tok(prefix: &str, ch: char, n: usize) -> String {
    format!("{prefix}{}", ch.to_string().repeat(n))
}

/// (prefix, the character that every shape accepts, minimum length after the prefix).
const SHAPES: [(&str, char, usize); 17] = [
    ("ghp_", 'a', 36),
    ("github_pat_", 'a', 22),
    ("gho_", 'a', 36),
    ("glpat-", 'a', 20),
    ("sk-", 'a', 32),
    ("sk-proj-", 'a', 40),
    ("sk-svcacct-", 'a', 40),
    ("sk-admin-", 'a', 40),
    ("sk-ant-", 'a', 40),
    ("xoxb-", 'a', 10),
    ("xoxa-", 'a', 10),
    ("xoxp-", 'a', 10),
    ("xoxr-", 'a', 10),
    ("xoxs-", 'a', 10),
    ("AKIA", 'A', 16),
    ("ASIA", 'A', 16),
    ("ya29.", 'a', 20),
];

fn key_block(kind: &str) -> String {
    format!("{}{} {kind}{}{}", "-----", "BEGIN", "-----", "\nMIIB\n")
}

fn private_key_block() -> String {
    format!(
        "{}{} RSA {}{}\nMIIB\n",
        "-----", "BEGIN", "PRIVATE", " KEY-----"
    )
}

const NEUTRAL: &str = "{OWN}/src/new.txt";

/// Path cases for a session in the own worktree: the template and the class a path alone gives.
fn path_cases() -> Vec<(&'static str, Option<&'static str>)> {
    vec![
        // allowed
        ("{OWN}/src/a.txt", None),
        ("{OWN}/new/deep/x.rs", None),
        ("{OWN}/docs/../src/a.txt", None),
        ("{SCR}/n.txt", None),
        ("{HOMESTATE}/reports/demo/r.md", None),
        ("{HOME}/.claude/projects/p/memory/m.md", None),
        ("{PLAIN}/x", None),
        ("{OWN}/.gitignore", None),
        ("{OWN}/.github/workflows/x.yml", None),
        ("{OWN}/docs/id_rsa.pub", None),
        ("{OWN}/docs/my.ssh.md", None),
        // another checkout
        ("{PRI}/src/a.txt", Some("other-checkout")),
        ("{PRI}/new/folder/x", Some("other-checkout")),
        ("{OTH}/src/a.txt", Some("other-checkout")),
        ("{OTH}/new/deep/z", Some("other-checkout")),
        ("{INNER}/src/a.txt", Some("other-checkout")),
        (
            "{PRI}/.claude/worktrees/inner/src/a.txt",
            Some("other-checkout"),
        ),
        ("{OWN}/../primary/src/a.txt", Some("other-checkout")),
        ("{OWN}/lnk-primary/a.txt", Some("other-checkout")),
        ("{OWN}/lnk-other/x", Some("other-checkout")),
        // protected
        ("{OWN}/.git", Some("protected-path")),
        ("{OWN}/x/.git/config", Some("protected-path")),
        ("{PRI}/.git/config", Some("protected-path")),
        ("{HOME}/.ssh/id_x", Some("protected-path")),
        ("{HOME}/.aws/credentials", Some("protected-path")),
        ("{HOME}/.kube/config", Some("protected-path")),
        ("{HOME}/.gnupg/k", Some("protected-path")),
        ("{HOME}/.config/gcloud/a", Some("protected-path")),
        ("{PLAIN}/id_rsa", Some("protected-path")),
        ("{PLAIN}/sub/id_ed25519", Some("protected-path")),
        ("{OWN}/lnk-ssh/new_key", Some("protected-path")),
        ("{OWN}/new/.ssh/x", Some("protected-path")),
    ]
}

// ---------------------------------------------------------------------- AC-1

/// AC-1: the path table and the content table, each for every tool, with the class and the
/// decision compared. At least 8 allowed and 8 denied cases of each check and tool.
#[test]
fn ac1_the_case_tables_decide_for_every_tool() {
    let w = World::new();
    let cases = path_cases();
    for class in [None, Some("other-checkout"), Some("protected-path")] {
        let n = cases.iter().filter(|(_, c)| *c == class).count();
        assert!(n >= 8, "{class:?}: {n} path cases");
    }
    for tool in TOOLS {
        for (path, expected) in &cases {
            let got = w.verdict(S::Own, tool, path, "hello\n");
            assert_eq!(got.as_deref(), *expected, "{tool} {path}");
        }
    }
    let mut denied = 0;
    let mut allowed = 0;
    for tool in TOOLS {
        for (prefix, ch, min) in SHAPES {
            let content = format!("export KEY={}\n", tok(prefix, ch, min));
            let got = w.verdict(S::Own, tool, NEUTRAL, &content);
            assert_eq!(got.as_deref(), Some("secret-content"), "{tool} {prefix}");
            denied += 1;
        }
        let block = private_key_block();
        assert_eq!(
            w.verdict(S::Own, tool, NEUTRAL, &block).as_deref(),
            Some("secret-content"),
            "{tool} key block"
        );
        for content in allowed_contents() {
            assert_eq!(w.verdict(S::Own, tool, NEUTRAL, &content), None, "{tool}");
            allowed += 1;
        }
    }
    assert!(denied >= 8 * 3 && allowed >= 8 * 3, "{denied} {allowed}");
}

/// Content that names a prefix, is too short, is a public block or an ordinary assignment.
fn allowed_contents() -> Vec<String> {
    let mut all = vec![
        String::new(),
        "plain text\n".to_owned(),
        "use a ghp_ token, never paste it\n".to_owned(),
        "the sk- prefix is documented here\n".to_owned(),
        "let password=prompt(); let token=read(); let api_key=load();\n".to_owned(),
        key_block("CERTIFICATE"),
        key_block("PUBLIC KEY"),
        format!("{}\n", tok("task-", 'a', 40)),
        "a risk-free desk-top task-based plan\n".to_owned(),
        "ssh-rsa AAAAB3NzaC1yc2E user@host\n".to_owned(),
        "x".repeat(200_000),
    ];
    for (prefix, ch, min) in SHAPES {
        all.push(format!("short {}\n", tok(prefix, ch, min - 1)));
    }
    all
}

// ---------------------------------------------------------------------- AC-2

/// AC-2: outside a MoAI project the hook prints nothing on either stream for every case, and in
/// a MoAI project a denial is one JSON object with exactly the three fields and an allowed case
/// prints nothing.
#[test]
fn ac2_outside_a_project_silent_and_inside_one_exactly_one_object() {
    let w = World::new();
    for (path, _) in path_cases() {
        for tool in TOOLS {
            let input = w.input(tool, path, "hello", &w.plain, "default");
            let run = w.spawn(&w.plain, input.to_string().as_bytes(), None);
            assert_eq!((run.code, run.stdout.as_str()), (0, ""), "{tool} {path}");
            assert_eq!(run.stderr, "", "{tool} {path}");
        }
    }
    let denied = w.run(S::Own, "Write", "{PRI}/src/a.txt", "x", "default");
    let value: Value = serde_json::from_str(denied.stdout.trim()).unwrap();
    assert_eq!(value["hookSpecificOutput"]["permissionDecision"], "deny");
    assert_eq!(denied.stdout.trim().lines().count(), 1);
    for word in ["\"ask\"", "\"allow\"", "updatedInput", "additionalContext"] {
        assert!(!denied.stdout.contains(word), "{word}");
    }
    let allowed = w.run(S::Own, "Write", "{OWN}/src/a.txt", "x", "default");
    assert_eq!((allowed.code, allowed.stdout.as_str()), (0, ""));
    assert_eq!(allowed.stderr, "");
}

// ---------------------------------------------------------------------- AC-3

/// AC-3: another tool name, no tool name, a missing, numeric, null, array or empty path, a tool
/// input that is no object, `NotebookEdit` with only `file_path` and `Write` with only
/// `notebook_path` each exit 0 with nothing printed, though the path would be denied if read.
#[test]
fn ac3_input_it_does_not_judge_prints_nothing() {
    let w = World::new();
    let bad = "{PLAIN}/.ssh/id_x";
    let path = w.fill(bad);
    let mut inputs: Vec<(String, Value)> = Vec::new();
    for tool in ["Bash", "Read", "MultiEdit", "write", "WriteX", ""] {
        inputs.push((
            format!("tool {tool:?}"),
            json!({"cwd": w.own, "hook_event_name": "PreToolUse", "tool_name": tool,
                   "tool_input": {"file_path": path, "content": "x"}}),
        ));
    }
    inputs.push((
        "no tool name".to_owned(),
        json!({"cwd": w.own, "hook_event_name": "PreToolUse",
               "tool_input": {"file_path": path, "content": "x"}}),
    ));
    inputs.push((
        "tool name not a string".to_owned(),
        json!({"cwd": w.own, "hook_event_name": "PreToolUse", "tool_name": 7,
               "tool_input": {"file_path": path, "content": "x"}}),
    ));
    for value in [json!(7), json!(null), json!(["x"]), json!(""), json!({})] {
        inputs.push((
            format!("path {value}"),
            json!({"cwd": w.own, "hook_event_name": "PreToolUse", "tool_name": "Write",
                   "tool_input": {"file_path": value, "content": "x"}}),
        ));
    }
    for tool_input in [json!("x"), json!(null), json!([1]), json!(5)] {
        inputs.push((
            format!("tool_input {tool_input}"),
            json!({"cwd": w.own, "hook_event_name": "PreToolUse", "tool_name": "Write",
                   "tool_input": tool_input}),
        ));
    }
    inputs.push((
        "no tool_input".to_owned(),
        json!({"cwd": w.own, "hook_event_name": "PreToolUse", "tool_name": "Write"}),
    ));
    inputs.push((
        "NotebookEdit by file_path".to_owned(),
        json!({"cwd": w.own, "hook_event_name": "PreToolUse", "tool_name": "NotebookEdit",
               "tool_input": {"file_path": path, "new_source": "x"}}),
    ));
    inputs.push((
        "Write by notebook_path".to_owned(),
        json!({"cwd": w.own, "hook_event_name": "PreToolUse", "tool_name": "Write",
               "tool_input": {"notebook_path": path, "content": "x"}}),
    ));
    for (what, input) in &inputs {
        let run = w.raw(S::Own, input);
        assert_eq!((run.code, run.stdout.as_str()), (0, ""), "{what}");
        assert_eq!(run.stderr, "", "{what}");
    }
    // The same call with the right key is judged: the table is not silent for every shape.
    assert_eq!(
        w.verdict(S::Own, "NotebookEdit", bad, "x").as_deref(),
        Some("protected-path")
    );
    // A token in `old_string` alone is not new content; a missing content is judged by the path.
    let old_only = json!({"cwd": w.own, "hook_event_name": "PreToolUse", "tool_name": "Edit",
        "tool_input": {"file_path": w.fill(NEUTRAL), "new_string": "ok",
                       "old_string": tok("ghp_", 'a', 36)}});
    assert_eq!(parse_verdict(&w.raw(S::Own, &old_only), "old_string"), None);
    let no_content = json!({"cwd": w.own, "hook_event_name": "PreToolUse", "tool_name": "Write",
        "tool_input": {"file_path": w.fill("{PRI}/src/a.txt")}});
    assert_eq!(
        parse_verdict(&w.raw(S::Own, &no_content), "no content").as_deref(),
        Some("other-checkout")
    );
}

// ---------------------------------------------------------------------- AC-4

/// AC-4: what counts as the session's own checkout and what lies outside every checkout.
#[test]
fn ac4_own_checkout_others_and_places_in_no_checkout() {
    let w = World::new();
    let table: [(S, &str, Option<&str>); 22] = [
        (S::Own, "{OWN}/src/a.txt", None),
        (S::Own, "{OWN}/missing/dir/file.rs", None),
        (S::Own, "{SCR}/n", None),
        (S::Own, "{HOMESTATE}/reports/demo/r.md", None),
        (S::Own, "{HOME}/.claude/projects/p/memory/m.md", None),
        (S::Own, "{PRI}/src/a.txt", Some("other-checkout")),
        (S::Own, "{OTH}/src/a.txt", Some("other-checkout")),
        (S::Own, "{INNER}/src/a.txt", Some("other-checkout")),
        (
            S::Own,
            "{OWN}/docs/../../primary/src/a.txt",
            Some("other-checkout"),
        ),
        (S::Own, "{OWN}/lnk-primary/a.txt", Some("other-checkout")),
        (S::Sub, "{OWN}/src/b.txt", None),
        (S::Sub, "a.txt", None),
        (S::Sub, "../docs/x.md", None),
        (S::Sub, "../../wt-other/src/a.txt", Some("other-checkout")),
        (S::Sub, "../../primary/src/a.txt", Some("other-checkout")),
        (S::Pri, "{PRI}/src/a.txt", None),
        (S::Pri, "{PRI}/new/x", None),
        (S::Pri, "{OWN}/src/a.txt", Some("other-checkout")),
        (S::Pri, "{INNER}/src/a.txt", Some("other-checkout")),
        (
            S::Pri,
            "{PRI}/.claude/worktrees/inner/src/a.txt",
            Some("other-checkout"),
        ),
        (S::Nowhere, "{PRI}/src/a.txt", None),
        (S::Nowhere, "{OTH}/src/a.txt", None),
    ];
    for (s, path, expected) in table {
        for tool in TOOLS {
            let got = w.verdict(s, tool, path, "hello\n");
            assert_eq!(got.as_deref(), expected, "{tool} {path}");
        }
    }
}

// ---------------------------------------------------------------------- AC-5

/// AC-5: every protected place in four spellings, and look-alikes allowed.
#[test]
fn ac5_every_protected_place_in_each_spelling() {
    let w = World::new();
    let places = [
        ".git/config",
        ".ssh/x",
        ".aws/x",
        ".azure/x",
        ".kube/x",
        ".gnupg/x",
        ".config/gcloud/x",
        "id_rsa",
        "id_ed25519",
        "id_ecdsa",
        "id_dsa",
    ];
    for place in places {
        for spelling in [
            "{PLAIN}/PLACE",
            "{PLAIN}/a/../PLACE",
            "{OWN}/lnk-plain/PLACE",
            "{PLAIN}/m/n/PLACE",
        ] {
            let path = spelling.replace("PLACE", place);
            for tool in TOOLS {
                let got = w.verdict(S::Own, tool, &path, "hello\n");
                assert_eq!(got.as_deref(), Some("protected-path"), "{tool} {path}");
            }
        }
    }
    for like in [
        ".gitignore",
        ".github/x",
        ".gitmodules",
        "my.ssh.md",
        "id_rsa.pub",
        "id_rsa_backup",
        "ssh/x",
        "aws/x",
        "gcloud/x",
        "x/.config/other/y",
    ] {
        let path = format!("{{PLAIN}}/{like}");
        for tool in TOOLS {
            assert_eq!(
                w.verdict(S::Own, tool, &path, "hello\n"),
                None,
                "{tool} {path}"
            );
        }
    }
}

/// AC-5 (case and unreadable shapes): a Mac ignores letter case in names, so the protected
/// places match without regard to ASCII case; a path whose links cannot be followed is refused
/// and a broken `.git` file never hides the checkout around it.
#[test]
fn ac5_case_spellings_and_unreadable_shapes_fail_closed() {
    let w = World::new();
    for place in [
        ".GIT/config",
        ".Git/hooks/x",
        ".Ssh/x",
        ".AWS/x",
        ".KUBE/x",
        ".config/GCloud/x",
        ".CONFIG/gcloud/x",
        "ID_RSA",
        "Id_Ed25519",
    ] {
        for tool in TOOLS {
            let path = format!("{{PLAIN}}/{place}");
            assert_eq!(
                w.verdict(S::Own, tool, &path, "hello\n").as_deref(),
                Some("protected-path"),
                "{tool} {path}"
            );
        }
    }

    // On a file system that ignores case another spelling of a checkout folder is the same
    // checkout; on one that does not, that spelling names nothing and is allowed.
    let respelled = w.root.join("WT-OTHER");
    let ignores_case = respelled.exists();
    let want = if ignores_case {
        Some("other-checkout")
    } else {
        None
    };
    for tool in TOOLS {
        let path = format!("{}/src/a.txt", respelled.display());
        assert_eq!(
            w.verdict(S::Own, tool, &path, "hello\n").as_deref(),
            want,
            "{tool} {path}"
        );
    }

    // More links than the resolver follows: refused, never skipped.
    let mut previous = w.plain.clone();
    for n in 0..40 {
        let link = w.plain.join(format!("chain-{n}"));
        std::os::unix::fs::symlink(&previous, &link).unwrap();
        previous = link;
    }
    std::os::unix::fs::symlink(w.plain.join("loop-b"), w.plain.join("loop-a")).unwrap();
    std::os::unix::fs::symlink(w.plain.join("loop-a"), w.plain.join("loop-b")).unwrap();
    for path in ["{PLAIN}/chain-39/x.txt", "{PLAIN}/loop-a/x.txt"] {
        for tool in TOOLS {
            assert_eq!(
                w.verdict(S::Own, tool, path, "hello\n").as_deref(),
                Some("protected-path"),
                "{tool} {path}"
            );
        }
    }

    // A `.git` file that names nothing readable: the search goes on with the parent folder.
    let broken_in_primary = w.primary.join("sub");
    let broken_alone = w.plain.join("broken");
    for dir in [&broken_in_primary, &broken_alone] {
        fs::create_dir_all(dir).unwrap();
        fs::write(dir.join(".git"), "garbage without a gitdir line\n").unwrap();
    }
    for tool in TOOLS {
        let inside = format!("{}/x.txt", broken_in_primary.display());
        assert_eq!(
            w.verdict(S::Own, tool, &inside, "hello\n").as_deref(),
            Some("other-checkout"),
            "{tool} {inside}"
        );
        let alone = format!("{}/x.txt", broken_alone.display());
        assert_eq!(
            w.verdict(S::Own, tool, &alone, "hello\n"),
            None,
            "{tool} {alone}"
        );
    }
}

// ---------------------------------------------------------------------- AC-6

/// AC-6: each shape at its minimum length is denied in the three content keys, one character
/// shorter and prose naming a prefix are allowed, and the word-start rule of the OpenAI-style
/// prefix holds.
#[test]
fn ac6_each_token_shape_at_its_minimum_and_not_below() {
    let w = World::new();
    for tool in TOOLS {
        for (prefix, ch, min) in SHAPES {
            let at_min = format!("see {}.", tok(prefix, ch, min));
            let longer = format!("{} and more", tok(prefix, ch, min + 7));
            let below = format!("see {}.", tok(prefix, ch, min - 1));
            for (text, expected) in [(at_min, true), (longer, true), (below, false)] {
                let got = w.verdict(S::Own, tool, NEUTRAL, &text);
                assert_eq!(got.is_some(), expected, "{tool} {prefix} {}", text.len());
            }
        }
        // The word-start rule: an ordinary word before the prefix is not a key.
        assert_eq!(
            w.verdict(S::Own, tool, NEUTRAL, &tok("task-", 'a', 40)),
            None
        );
        assert!(w
            .verdict(S::Own, tool, NEUTRAL, &format!("({})", tok("sk-", 'a', 32)))
            .is_some());
        // A run that stops at a character the shape does not accept does not count.
        let broken = format!("{}-{}", tok("sk-", 'a', 20), "a".repeat(20));
        assert_eq!(w.verdict(S::Own, tool, NEUTRAL, &broken), None, "{tool}");
        // A key block needs its two halves.
        assert!(w
            .verdict(S::Own, tool, NEUTRAL, &private_key_block())
            .is_some());
        assert_eq!(
            w.verdict(S::Own, tool, NEUTRAL, &key_block("CERTIFICATE")),
            None
        );
    }
}

// ---------------------------------------------------------------------- AC-7

/// AC-7: every reason starts with `moai-boundary/` and a class word, is under 400 bytes and
/// carries no byte of the path or the content.
#[test]
fn ac7_reasons_carry_the_class_and_no_input_byte() {
    let w = World::new();
    let secret = format!("CONTENTMARKER {}", tok("ghp_", 'a', 36));
    let cases = [
        (
            "{PLAIN}/PATHMARKER/.ssh/k",
            "CONTENTMARKER",
            "protected-path",
        ),
        ("{PRI}/PATHMARKER/x", "CONTENTMARKER", "other-checkout"),
        ("{OWN}/PATHMARKER/x", secret.as_str(), "secret-content"),
    ];
    let mut seen = Vec::new();
    for (path, content, class) in cases {
        for tool in TOOLS {
            let run = w.run(S::Own, tool, path, content, "default");
            assert_eq!(parse_verdict(&run, path).as_deref(), Some(class));
            assert!(!run.stdout.contains("PATHMARKER"), "{}", run.stdout);
            assert!(!run.stdout.contains("CONTENTMARKER"), "{}", run.stdout);
            assert!(!run.stdout.contains(&w.root.display().to_string()));
            assert!(!run.stdout.contains(&tok("ghp_", 'a', 10)));
            let value: Value = serde_json::from_str(run.stdout.trim()).unwrap();
            let reason = value["hookSpecificOutput"]["permissionDecisionReason"]
                .as_str()
                .unwrap();
            assert!(
                reason.starts_with(&format!("moai-boundary/{class}:")),
                "{reason}"
            );
            seen.push(reason.to_owned());
        }
    }
    seen.sort();
    seen.dedup();
    assert_eq!(seen.len(), 3, "one reason text per class");
}

// ---------------------------------------------------------------------- AC-9

/// Every file of a folder with its bytes, for the "nothing is written" check.
fn snapshot(root: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(base: &Path, dir: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
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

/// AC-9: the module names no process, network, environment or write word, uses only three read
/// calls of `fs`, the answer is the same for every permission mode, and no case changes a file
/// of the world.
#[test]
fn ac9_the_handler_is_read_only_and_mode_independent() {
    let source = include_str!("../src/hook_write_boundary.rs");
    let code: String = source
        .lines()
        .filter(|l| !l.trim_start().starts_with("//"))
        .collect::<Vec<_>>()
        .join("\n");
    let code = code
        .split("#[cfg(test)]")
        .next()
        .unwrap_or_default()
        .to_owned();
    for word in [
        "std::process",
        "Command",
        "TcpStream",
        "UdpSocket",
        "reqwest",
        "ureq",
        "hyper",
        "std::env",
        "env::var",
        "permission_mode",
        "OpenOptions",
        "File::create",
        "create_dir",
        "remove_",
        ".write(",
        "write_all",
        "set_permissions",
    ] {
        assert!(
            !code.contains(word),
            "src/hook_write_boundary.rs names `{word}`"
        );
    }
    let mut calls: Vec<&str> = code
        .match_indices("fs::")
        .map(|(at, _)| {
            let rest = &code[at + 4..];
            &rest[..rest.find('(').unwrap()]
        })
        .collect();
    calls.sort_unstable();
    calls.dedup();
    assert_eq!(
        calls,
        [
            "canonicalize",
            "read_link",
            "read_to_string",
            "symlink_metadata"
        ]
    );

    let w = World::new();
    let before = snapshot(&w.root);
    for (path, content) in [
        ("{PRI}/src/a.txt", "x"),
        ("{OWN}/src/a.txt", "x"),
        ("{PLAIN}/.ssh/k", "x"),
        ("{OWN}/src/n.txt", "export K=ghp_ short"),
    ] {
        let mut first: Option<(String, String)> = None;
        for mode in [
            "default",
            "acceptEdits",
            "plan",
            "dontAsk",
            "auto",
            "bypassPermissions",
        ] {
            let run = w.run(S::Own, "Write", path, content, mode);
            let seen = (run.stdout.clone(), run.stderr.clone());
            assert_eq!(run.code, 0);
            match &first {
                None => first = Some(seen),
                Some(f) => assert_eq!(*f, seen, "{path} in mode {mode}"),
            }
        }
    }
    for tool in TOOLS {
        for (path, _) in path_cases() {
            let _ = w.verdict(S::Own, tool, path, "hello\n");
        }
    }
    assert_eq!(before, snapshot(&w.root), "a case changed a file");
}

// --------------------------------------------------------------------- AC-13

fn percentile(sorted: &[f64], p: f64) -> f64 {
    let rank = ((p / 100.0) * sorted.len() as f64).ceil() as usize;
    sorted[rank.clamp(1, sorted.len()) - 1]
}

/// AC-13: 200 runs of the real binary for an allowed and a denied call; the p50 and p90 in
/// milliseconds are printed with the build profile and the machine load, and the test fails only
/// when the p90 is over 10 times the target (40 ms).
#[test]
fn ac13_latency_is_measured_and_reported() {
    let w = World::new();
    let profile = if cfg!(debug_assertions) {
        "debug"
    } else {
        "release"
    };
    let load = Command::new("uptime")
        .output()
        .map(|o| String::from_utf8_lossy(&o.stdout).trim().to_owned())
        .unwrap_or_default();
    for (label, path, expect_denied) in [
        ("allowed", "{OWN}/src/a.txt", false),
        ("denied", "{PRI}/src/a.txt", true),
    ] {
        let (proc_cwd, cwd, project) = w.session(S::Own);
        let input = w.input("Write", path, "hello\n", &cwd, "default");
        let bytes = input.to_string().into_bytes();
        let mut ms = Vec::new();
        for _ in 0..200 {
            let start = Instant::now();
            let run = w.spawn(&proc_cwd, &bytes, project.as_deref());
            ms.push(start.elapsed().as_secs_f64() * 1000.0);
            assert_eq!(run.stdout.trim().is_empty(), !expect_denied, "{label}");
        }
        ms.sort_by(f64::total_cmp);
        let (p50, p90) = (percentile(&ms, 50.0), percentile(&ms, 90.0));
        println!("write-boundary {label}: p50 {p50:.1} ms, p90 {p90:.1} ms, profile {profile}, load {load}");
        assert!(
            p90 <= 400.0,
            "{label}: p90 {p90:.1} ms is over 10 times the target"
        );
    }
}
