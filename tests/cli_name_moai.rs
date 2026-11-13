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

/// AC-2: both names print the same `--version` line, `moai-cli <ver> (rust)`, the same help
/// and the same usage error, byte for byte.
#[test]
fn both_names_print_the_same_output() {
    let version = run(MOAI, &["--version"], b"");
    assert_eq!(version.status.code(), Some(0));
    let line = text(&version.stdout);
    let first = line.lines().next().unwrap();
    assert!(first.starts_with("moai-cli "), "{first}");
    assert!(first.ends_with(" (rust)"), "{first}");
    assert!(
        first.contains(env!("CARGO_PKG_VERSION")),
        "{first} names the package version"
    );
    for args in [
        vec!["--version"],
        vec!["--help"],
        vec!["claim-verify", "--help"],
        vec!["no-such-command"],
        vec!["pr-self-audit-check"],
    ] {
        let a = run(MOAI, &args, b"");
        let b = run(MOAI_CLI, &args, b"");
        assert_eq!(a.status.code(), b.status.code(), "{args:?}");
        assert_eq!(text(&a.stdout), text(&b.stdout), "{args:?} stdout");
        assert_eq!(text(&a.stderr), text(&b.stderr), "{args:?} stderr");
    }
    // The usage line names `moai` whichever binary ran (clap `bin_name`, #2812 W3).
    let help = text(&run(MOAI, &["--help"], b"").stdout);
    assert!(help.contains("Usage: moai "), "{help}");
    assert!(!help.contains("Usage: moai-cli"), "{help}");
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

/// Runs `moai args…` in `dir` with HOME `home`, returning the envelope's `result`.
fn result_in(dir: &std::path::Path, home: &std::path::Path, args: &[&str]) -> serde_json::Value {
    let out = Command::new(MOAI)
        .args(args)
        .env("HOME", home)
        .current_dir(dir)
        .output()
        .unwrap();
    assert_eq!(
        out.status.code(),
        Some(0),
        "{args:?}: {}",
        text(&out.stderr)
    );
    serde_json::from_slice::<serde_json::Value>(&out.stdout).unwrap()["result"].clone()
}

/// #2812 W3 AC-2: newly generated hook commands call `moai`, and a project whose hooks were
/// generated earlier with `moai-cli` sees them as an `update` (not a conflict) in the hooks
/// preview, for both hosts. The earlier generation is reproduced by rewriting the commands
/// MoAI wrote — in the host file and in MoAI's ownership receipt — back to `moai-cli`.
/// Everything runs in a temporary project with a temporary HOME.
#[test]
fn hooks_generated_with_moai_cli_preview_as_an_update_to_moai() {
    use std::fs;
    let tmp = TempDir::new().unwrap();
    let home = tmp.path().join("home");
    let p = tmp.path().join("p");
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(p.join(".moai-cli")).unwrap();
    common::git::git_ok(&p, &["init", "-q", "."]);

    for host in ["claude", "codex"] {
        let plan = result_in(&p, &home, &["hooks-plan", ".", "--host", host]);
        let commands: Vec<String> = plan["hooks"]
            .as_array()
            .unwrap()
            .iter()
            .map(|h| h["command"].as_str().unwrap().to_string())
            .collect();
        assert!(!commands.is_empty());
        for c in &commands {
            assert!(c.starts_with("moai "), "{host}: {c}");
        }
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        result_in(
            &p,
            &home,
            &[
                "hooks-apply",
                ".",
                "--host",
                host,
                "--plan-digest",
                &digest,
                "--permission-reference",
                "w3-test",
            ],
        );
        let again = result_in(&p, &home, &["hooks-plan", ".", "--host", host]);
        assert_eq!(again["action"], "unchanged", "{host}: {again}");
    }

    // Turn the generated commands back into the earlier `moai-cli` generation.
    let older = |s: String| {
        s.replace("\"moai hook-receipt", "\"moai-cli hook-receipt")
            .replace("\"moai claim-verify", "\"moai-cli claim-verify")
            .replace(
                "generated by moai hooks-apply",
                "generated by moai-cli hooks-apply",
            )
    };
    for f in [
        ".claude/settings.json",
        ".moai-cli/manifests/claude-settings.json",
        ".codex/hooks.json",
    ] {
        let path = p.join(f);
        let before = fs::read_to_string(&path).unwrap();
        let after = older(before.clone());
        assert_ne!(before, after, "{f} carried generated commands");
        fs::write(&path, after).unwrap();
    }
    // The Codex receipt owns the file by its sha256.
    let codex = fs::read(p.join(".codex/hooks.json")).unwrap();
    let receipt_path = p.join(".moai-cli/manifests/codex-hooks.json");
    let mut receipt: serde_json::Value =
        serde_json::from_slice(&fs::read(&receipt_path).unwrap()).unwrap();
    receipt["sha256"] = serde_json::json!(moai_cli::digest(&codex));
    let receipt_text = older(serde_json::to_string_pretty(&receipt).unwrap());
    fs::write(&receipt_path, receipt_text + "\n").unwrap();

    for host in ["claude", "codex"] {
        let plan = result_in(&p, &home, &["hooks-plan", ".", "--host", host]);
        assert_eq!(plan["action"], "update", "{host}: {plan}");
        assert_eq!(plan["conflicts"], serde_json::json!([]), "{host}: {plan}");
        let shown = plan.to_string();
        assert!(
            shown.contains("moai claim-verify . --hook"),
            "{host}: {shown}"
        );
    }
}
