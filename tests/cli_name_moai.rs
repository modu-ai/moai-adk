//! Issue #2790 (public CLI name `moai`, steps R0 and R1 of the design on #2725, comment
//! 6863726387).
//!
//! R0: a Go-generation project hook runs `moai hook <event> [args]` (34 events in the Go
//! templates' `.claude/hooks/moai/*.sh`, `exec` so the exit code reaches the host). This build
//! has no `hook` command, and clap's usage error would exit 2 — "block this tool call" to the
//! host. Every event must exit 0 with an empty stdout and one stderr line.
//!
//! R1: the binaries `moai` and `moai-cli` run one `main` and print the same output, and
//! `--version` reads `moai-cli <ver> (rust)` from both.
//!
//! Every process runs with a throwaway HOME; the real HOME and PATH are never touched.

mod common;

use std::process::{Command, Output, Stdio};
use tempfile::TempDir;

const MOAI: &str = common::MOAI_BIN;
const MOAI_CLI: &str = common::MOAI_CLI_BIN;

/// The events the Go templates' hook scripts `exec moai hook <event>` with (moai-adk-go
/// `4fa2aa9db2`, `internal/template/templates/.claude/hooks/moai/*.sh`, 34 scripts).
const GO_HOOK_EVENTS: [&str; 34] = [
    "agent",
    "compact",
    "config-change",
    "cwd-changed",
    "elicitation",
    "elicitation-result",
    "file-changed",
    "harness-observe",
    "harness-observe-stop",
    "harness-observe-subagent-stop",
    "harness-observe-user-prompt-submit",
    "instructions-loaded",
    "notification",
    "permission-denied",
    "permission-request",
    "post-compact",
    "post-tool",
    "post-tool-failure",
    "security-commit",
    "security-scan",
    "security-turn",
    "session-end",
    "session-start",
    "session-start-compact",
    "stop",
    "stop-failure",
    "subagent-start",
    "subagent-stop",
    "task-completed",
    "task-created",
    "teammate-idle",
    "user-prompt-submit",
    "worktree-create",
    "worktree-remove",
];

/// Runs `bin args…` with `stdin` piped in, HOME a fresh temporary folder.
fn run(bin: &str, args: &[&str], stdin: &[u8]) -> Output {
    let home = TempDir::new().unwrap();
    let mut child = Command::new(bin)
        .args(args)
        .env("HOME", home.path())
        .current_dir(home.path())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, stdin);
    child.wait_with_output().unwrap()
}

fn text(b: &[u8]) -> String {
    String::from_utf8_lossy(b).into_owned()
}

/// AC-1: each of the 34 Go hook events, from both names, with the host's JSON on stdin (and
/// the trailing arguments the `agent` and `security-*` scripts pass): exit 0, empty stdout,
/// exactly one stderr line naming the event.
#[test]
fn every_go_hook_event_exits_zero_with_an_empty_stdout() {
    assert_eq!(GO_HOOK_EVENTS.len(), 34);
    let payload = br#"{"session_id":"s","hook_event_name":"PreToolUse","tool_name":"Bash"}"#;
    for bin in [MOAI, MOAI_CLI] {
        for event in GO_HOOK_EVENTS {
            let mut args = vec!["hook", event];
            if event == "agent" {
                args.push("moai-audit");
            }
            if event.starts_with("security-") {
                args.extend(["--quiet", "x"]);
            }
            let out = run(bin, &args, payload);
            let err = text(&out.stderr);
            assert_eq!(out.status.code(), Some(0), "{bin} {args:?}: {err}");
            assert!(
                out.stdout.is_empty(),
                "{bin} {args:?}: stdout {:?}",
                text(&out.stdout)
            );
            assert_eq!(err.lines().count(), 1, "{bin} {args:?}: {err}");
            assert!(
                err.contains(&format!("(hook {event})")),
                "{bin} {args:?}: {err}"
            );
        }
    }
}

/// Boundaries of R0: `hook` with no event, an argument that is not an event word and a large
/// stdin are still answered 0; only the exact first word `hook` is taken — `hooks` and a later
/// `hook` remain ordinary command lines, so clap's usage error (exit 2) still answers them.
#[test]
fn only_the_first_word_hook_is_taken_and_its_input_never_blocks() {
    for (args, event) in [
        (vec!["hook"], "?"),
        (vec!["hook", "Pre Tool; rm -rf /"], "?"),
        (vec!["hook", ""], "?"),
    ] {
        let out = run(MOAI, &args, b"{}");
        let err = text(&out.stderr);
        assert_eq!(out.status.code(), Some(0), "{args:?}: {err}");
        assert!(out.stdout.is_empty());
        assert!(err.contains(&format!("(hook {event})")), "{args:?}: {err}");
        assert!(!err.contains("rm -rf"), "the argument is not echoed: {err}");
    }
    // #2796 audit nit 1 (#2792): the event name is shown only as `[a-z-]{1,64}` — 64 lowercase
    // characters pass, 65 or an upper-case letter print `?`.
    let at_limit = "a".repeat(64);
    let over = "a".repeat(65);
    for (event, shown) in [
        (at_limit.as_str(), at_limit.as_str()),
        (over.as_str(), "?"),
        ("Pre-tool", "?"),
    ] {
        let out = run(MOAI, &["hook", event], b"{}");
        let err = text(&out.stderr);
        assert_eq!(out.status.code(), Some(0), "{event}: {err}");
        assert!(err.contains(&format!("(hook {shown})")), "{event}: {err}");
    }
    // A stdin larger than the pipe buffer is read and dropped, not left to block the writer.
    // The writer (the host) sees every byte accepted: no broken pipe.
    let big = vec![b'x'; 1024 * 1024];
    let home = TempDir::new().unwrap();
    let mut child = Command::new(MOAI)
        .args(["hook", "post-tool"])
        .env("HOME", home.path())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    let written = std::io::Write::write_all(&mut pipe, &big);
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    assert!(written.is_ok(), "the host's write failed: {written:?}");
    assert_eq!(out.status.code(), Some(0));

    for args in [vec!["hooks", "session-start"], vec!["status", "hook"]] {
        let out = run(MOAI, &args, b"");
        assert_eq!(out.status.code(), Some(2), "{args:?} is not a Go hook call");
    }
}

/// AC-2: both names print the same `--version` line, `moai-cli <ver> (rust)`, and the same help
/// and usage error byte for byte: since #3004 P1 usage and help name `moai` whichever name ran
/// (#2802 S1 had them name the binary that ran).
#[test]
fn both_names_print_the_same_output() {
    let version = run(MOAI, &["--version"], b"");
    assert_eq!(version.status.code(), Some(0));
    let line = text(&version.stdout);
    let first = line.lines().next().unwrap();
    assert!(first.starts_with("moai-cli "), "{first}");
    assert!(first.ends_with(" (rust)"), "{first}");
    // Card C-67 S1: the first line names the HeadVer build version, the second line which
    // kind of build it is; `-V` prints the first line alone.
    assert_eq!(
        first,
        format!("moai-cli {} (rust)", moai_cli::BUILD_VERSION),
        "{first} names the build version"
    );
    assert_eq!(
        line,
        format!("moai-cli {}\n", moai_cli::LONG_VERSION),
        "{line}"
    );
    assert_eq!(
        text(&run(MOAI, &["-V"], b"").stdout),
        format!("moai-cli {}\n", moai_cli::VERSION_SIGNATURE)
    );
    // `--version` is printed by both names as it is (D3 decision 2).
    assert_eq!(
        text(&run(MOAI, &["--version"], b"").stdout),
        text(&run(MOAI_CLI, &["--version"], b"").stdout)
    );
    for args in [
        vec!["--help"],
        vec!["claim-verify", "--help"],
        vec!["no-such-command"],
        vec!["pr-self-audit-check"],
    ] {
        let a = run(MOAI, &args, b"");
        let b = run(MOAI_CLI, &args, b"");
        assert_eq!(a.status.code(), b.status.code(), "{args:?}");
        // Since #3004 P1 the two names print the very same bytes, no name swapped.
        assert_eq!(text(&a.stdout), text(&b.stdout), "{args:?} stdout");
        assert_eq!(text(&a.stderr), text(&b.stderr), "{args:?} stderr");
    }
    // The usage line names `moai` whichever name ran (#3004 AC-1).
    let help = text(&run(MOAI, &["--help"], b"").stdout);
    assert!(help.contains("Usage: moai <command>"), "{help}");
    let help = text(&run(MOAI_CLI, &["--help"], b"").stdout);
    assert!(help.contains("Usage: moai <command>"), "{help}");
    assert!(!help.contains("moai-cli"), "{help}");
}

/// Impact table B2: a `moai` hard-linked beside `moai-cli` is the same installation — the
/// receipt recorded for `moai-cli` is `self` when `moai install-inspect` runs. A `moai` whose
/// bytes differ from its sibling is not (its own path, no receipt: `unrecorded`). The HOME is
/// a temporary folder.
#[test]
fn moai_beside_the_same_build_reads_the_moai_cli_receipt() {
    use moai_cli::install::{plan_receipt, record_receipt, InstallerOwner};
    let home = TempDir::new().unwrap();
    let home_path = std::fs::canonicalize(home.path()).unwrap();
    let (cli, moai) = common::install_both_names(&home_path.join("bin"));
    let plan = plan_receipt(&cli, &home_path, InstallerOwner::Direct).unwrap();
    record_receipt(&home_path, &plan).unwrap();
    // Planning from the hard link names the same installation: same id, same recorded path.
    let via_moai = plan_receipt(&moai, &home_path, InstallerOwner::Direct).unwrap();
    assert_eq!(via_moai.install_id, plan.install_id);
    assert_eq!(via_moai.receipt.binary_path, "~/bin/moai-cli");
    assert_eq!(plan.receipt.binary_path, "~/bin/moai-cli");

    let inspect = |exe: &std::path::Path| -> serde_json::Value {
        let out = Command::new(exe)
            .arg("install-inspect")
            .env("HOME", &home_path)
            .env("PATH", "")
            .current_dir(&home_path)
            .output()
            .unwrap();
        assert_eq!(out.status.code(), Some(0), "{}", text(&out.stderr));
        serde_json::from_slice::<serde_json::Value>(&out.stdout).unwrap()["result"].clone()
    };
    assert_eq!(inspect(&cli)["self"]["state"], "self");
    assert_eq!(
        inspect(&moai)["self"]["state"],
        "self",
        "the hard link is the same installation"
    );

    // A different build named `moai` beside it keeps its own identity.
    std::fs::remove_file(&moai).unwrap();
    common::copy_executable(std::path::Path::new(MOAI), &moai);
    assert_ne!(std::fs::read(&moai).unwrap(), std::fs::read(&cli).unwrap());
    assert_eq!(inspect(&moai)["self"]["state"], "unrecorded");
}

/// #2812 W3, audit 6868377245 B1, #3004 P4-2: the generated hook commands call `moai`. With a
/// Go-generation `moai` first on PATH, the hook binary is `wrong_generation` (never `resolved`),
/// the Go `moai` is asked for `--version` and nothing else, and `hooks-apply` refuses to write a
/// fresh hook set that would hand the claim guard to it (the Go `moai` has no `claim-verify` and
/// exits 1, which a host reads as a non-blocking error, so the guard would be off while
/// everything reported success). With the Rust `moai` first, the same apply writes. The fake Go
/// records every call. Temporary HOME, project and PATH.
#[test]
fn a_go_moai_first_on_path_is_the_wrong_generation_and_no_hook_is_written() {
    use std::fs;
    let tmp = TempDir::new().unwrap();
    let root = std::fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let p = root.join("p");
    let go = root.join("go-bin");
    let rust = root.join("rust-bin");
    for d in [&home, &p, &go, &rust] {
        fs::create_dir_all(d).unwrap();
    }
    fs::create_dir_all(p.join(".moai-cli")).unwrap();
    common::git::git_ok(&p, &["init", "-q", "."]);
    let go_log = root.join("go-calls");
    common::write_executable(
        &go.join("moai"),
        &format!(
            "#!/bin/sh\necho \"$*\" >> '{}'\nif [ \"$1\" = --version ]; then echo 'moai-adk v3.2.0-rc.23'; exit 0; fi\necho \"Error: unknown command \\\"$1\\\"\" >&2\nexit 1\n",
            go_log.display()
        ),
    );
    common::copy_executable(std::path::Path::new(MOAI_CLI), &rust.join("moai"));
    let go_first = format!("{}:{}:/usr/bin:/bin", go.display(), rust.display());
    let rust_first = format!("{}:{}:/usr/bin:/bin", rust.display(), go.display());

    let run = |path: &str, args: &[&str]| -> (Option<i32>, serde_json::Value) {
        let out = Command::new(MOAI)
            .args(args)
            .env("HOME", &home)
            .env("PATH", path)
            .current_dir(&p)
            .output()
            .unwrap();
        let v = serde_json::from_slice::<serde_json::Value>(&out.stdout)
            .unwrap_or_else(|_| panic!("{}", text(&out.stderr)));
        (out.status.code(), v)
    };
    let plan_with = |path: &str, host: &str| -> serde_json::Value {
        let (code, v) = run(path, &["hooks-plan", ".", "--host", host]);
        assert_eq!(code, Some(0), "{v}");
        v["result"].clone()
    };
    for host in moai_cli::hooks::HOSTS.iter().copied() {
        let plan = plan_with(&go_first, host);
        assert_eq!(plan["hook_binary"]["name"], "moai", "{host}: {plan}");
        assert_eq!(
            plan["hook_binary"]["state"], "wrong_generation",
            "{host}: {plan}"
        );
        assert_eq!(plan["hooks_active"], false, "{host}: {plan}");
        assert_eq!(plan["gate"]["allowed"], false, "{host}: {plan}");
        for h in plan["hooks"].as_array().unwrap() {
            let c = h["command"].as_str().unwrap();
            assert!(
                c.starts_with("moai ") && !c.contains("moai-cli"),
                "{host}: {c}"
            );
        }
        let (code, refused) = run(
            &go_first,
            &[
                "hooks-apply",
                ".",
                "--host",
                host,
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                "test-3004",
            ],
        );
        assert_ne!(code, Some(0), "{host}: {refused}");
        assert!(
            refused.to_string().contains("wrong_generation"),
            "{refused}"
        );
    }
    // Nothing was written, and the Go `moai` heard `--version` only.
    assert!(!p.join(".claude").exists(), ".claude was written");
    let calls = fs::read_to_string(&go_log).unwrap();
    assert!(
        calls.lines().count() >= 1 && calls.lines().all(|l| l == "--version"),
        "only --version may reach the Go moai: {calls:?}"
    );

    // The Rust `moai` first: the same plan is `resolved`, and the apply writes.
    let plan = plan_with(&rust_first, "claude");
    assert_eq!(plan["hook_binary"]["state"], "resolved", "{plan}");
    assert_eq!(plan["gate"]["allowed"], true, "{plan}");
    let (code, applied) = run(
        &rust_first,
        &[
            "hooks-apply",
            ".",
            "--host",
            "claude",
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "test-3004",
        ],
    );
    assert_eq!(code, Some(0), "{applied}");
    assert!(p.join(".claude/settings.json").is_file());

    // Run the generated claim guard the way a host does with the Rust `moai` first: its
    // command string through a shell with that PATH, the PreToolUse input on stdin. The Go
    // `moai` is never reached.
    fs::remove_file(&go_log).unwrap();
    let guard = plan["hooks"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["event"] == "PreToolUse")
        .map(|h| h["command"].as_str().unwrap().to_string())
        .unwrap();
    let input = format!(
        r#"{{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{{"command":"git commit -m x"}},"cwd":"{}"}}"#,
        p.display()
    );
    let mut child = Command::new("/bin/sh")
        .args(["-c", &guard])
        .env("HOME", &home)
        .env("PATH", &rust_first)
        .current_dir(&p)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.as_bytes());
    let out = child.wait_with_output().unwrap();
    assert!(
        !go_log.exists(),
        "the Go moai was called: {:?}",
        fs::read_to_string(&go_log)
    );
    assert!(
        !text(&out.stderr).contains("unknown command"),
        "{}",
        text(&out.stderr)
    );
    assert_ne!(out.status.code(), Some(1), "{}", text(&out.stderr));
}
