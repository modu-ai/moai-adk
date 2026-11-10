//! Issue #2053: `moai-cli cc|glm|codex` without `-l|-a` composes a plain host session.
//! Since #2359 (user decision 2026-11-10) it starts with full access like every launched
//! session; `--permissions` keeps the prompts. Every case uses `--dry-run`; no host is started
//! (AC-4 stays not_run).
use serde_json::{json, Value};
use std::process::Command;

/// Runs the launcher with a clean advisor/GLM environment; `token` sets a dummy GLM token.
fn dry(args: &[&str], token: Option<&str>) -> (i32, Value, String) {
    let mut cmd = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    cmd.args(args)
        .arg("--dry-run")
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        // No home, so no key file (#2359) can stand in for the token.
        .env(
            "HOME",
            std::env::temp_dir().join("moai-cli-test-no-such-home"),
        );
    if let Some(t) = token {
        cmd.env("ANTHROPIC_AUTH_TOKEN", t);
    }
    let out = cmd.output().unwrap();
    let stdout = String::from_utf8_lossy(&out.stdout).to_string();
    let v = serde_json::from_str(&stdout).unwrap_or(Value::Null);
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    (
        out.status.code().unwrap_or(-1),
        v,
        format!("{stdout}{stderr}"),
    )
}
fn argv(v: &Value) -> Vec<String> {
    serde_json::from_value(v["argv"].clone()).unwrap()
}

#[test]
fn roleless_cc_is_a_plain_claude_session() {
    let (code, v, _) = dry(&["cc"], None);
    assert_eq!(code, 0);
    assert_eq!(v["role"], Value::Null, "REQ-5");
    assert_eq!(v["name"], Value::Null);
    let a = argv(&v);
    assert!(!a.iter().any(|x| x == "--name"), "AC-1: no session name");
    // User decision 2026-11-08 (HOLD-L1 = A, #2062): the launcher's Claude defaults; user
    // decision 2026-11-10 (#2359, replaces #2053's "no skip"): full access.
    const { assert!(moai_cli::session::ROLELESS_LAUNCHER_DEFAULTS) };
    const { assert!(moai_cli::session::FULL_ACCESS_DEFAULT) };
    assert_eq!(
        a,
        [
            "claude",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--advisor",
            "fable",
            "--dangerously-skip-permissions"
        ]
    );
    let notes = v["notes"].to_string();
    assert!(!notes.contains("HOLD"), "decided, no HOLD wording: {notes}");
    assert!(notes.contains("user decision 2026-11-08: A"), "{notes}");
    assert!(notes.contains("no role"));
    // Options still apply; an explicit name is passed through, a role is never guessed.
    let (_, v, _) = dry(
        &["cc", "--name", "mine", "--model", "sonnet", "--no-advisor"],
        None,
    );
    assert_eq!(v["role"], Value::Null);
    assert_eq!(
        argv(&v),
        [
            "claude",
            "--name",
            "mine",
            "--model",
            "sonnet",
            "--effort",
            "medium",
            "--dangerously-skip-permissions"
        ]
    );
    // `--permissions` keeps the prompts: the skip goes, the defaults stay.
    let (_, v, _) = dry(&["cc", "--permissions"], None);
    assert_eq!(
        argv(&v),
        [
            "claude",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--advisor",
            "fable"
        ]
    );
    // -l and -a together stay an error.
    let (code, _, _) = dry(&["cc", "-l", "-a"], None);
    assert_ne!(code, 0);
}

#[test]
fn roleless_codex_has_no_name_prompt() {
    let (code, v, _) = dry(&["codex"], None);
    assert_eq!(code, 0);
    assert_eq!(v["role"], Value::Null);
    assert_eq!(
        argv(&v),
        ["codex", "-s", "danger-full-access", "-a", "never"],
        "AC-2 (full access since #2359)"
    );
    // An extra prompt goes in as given, without a session-name line.
    let (_, v, _) = dry(&["codex", "review the diff"], None);
    assert_eq!(
        argv(&v),
        [
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "review the diff"
        ]
    );
    assert!(!v.to_string().contains("MoAI session name"));
    // `--permissions` keeps the sandboxed, ask-first pair.
    let (_, v, _) = dry(&["codex", "--permissions"], None);
    assert_eq!(
        argv(&v),
        ["codex", "-s", "workspace-write", "-a", "on-request"]
    );
}

#[test]
fn roleless_glm_keeps_the_token_check() {
    // Without the token: blocked, naming the variable only.
    let (code, _, text) = dry(&["glm"], None);
    assert_ne!(code, 0);
    assert!(text.contains("ANTHROPIC_AUTH_TOKEN"));
    // With a dummy token: names only, never the value; no name and no advisor on glm, full
    // access (#2359).
    let dummy = "moai-dummy-token-value";
    let (code, v, text) = dry(&["glm"], Some(dummy));
    assert_eq!(code, 0);
    assert!(!text.contains(dummy), "token value echoed");
    assert_eq!(v["role"], Value::Null);
    assert_eq!(
        v["env_keys"],
        json!([
            "ANTHROPIC_BASE_URL",
            "API_TIMEOUT_MS",
            "ANTHROPIC_DEFAULT_OPUS_MODEL",
            "ANTHROPIC_DEFAULT_SONNET_MODEL",
            "ANTHROPIC_DEFAULT_HAIKU_MODEL",
            "ANTHROPIC_DEFAULT_FABLE_MODEL",
            "CLAUDE_CODE_AUTO_COMPACT_WINDOW",
            "CLAUDE_CODE_MAX_CONTEXT_TOKENS",
            "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS",
            "ANTHROPIC_AUTH_TOKEN",
            "Z_AI_API_KEY"
        ])
    );
    assert_eq!(
        argv(&v),
        [
            "claude",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--dangerously-skip-permissions"
        ]
    );
}

/// REQ-4 / AC-3: the role paths keep their argv byte for byte; codex `-a` carries full access
/// since #2359 (user decision 2026-11-10).
#[test]
fn role_paths_are_unchanged() {
    let skip = "--dangerously-skip-permissions";
    let (_, v, _) = dry(&["cc", "-l"], None);
    assert_eq!(v["role"], "leader");
    assert_eq!(
        argv(&v),
        [
            "claude",
            "--name",
            "leader",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--advisor",
            "fable",
            skip
        ]
    );
    let (_, v, _) = dry(&["-a"], None);
    assert_eq!(
        argv(&v),
        [
            "claude",
            "--name",
            "agent-1",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--advisor",
            "fable",
            skip
        ]
    );
    let (_, v, _) = dry(&["codex", "-a", "--name", "agent-9"], None);
    assert_eq!(
        argv(&v),
        [
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "MoAI session name: agent-9"
        ]
    );
    let (_, v, _) = dry(&["glm", "-a"], Some("moai-dummy-token-value"));
    assert_eq!(
        argv(&v),
        ["claude", "--name", "agent-1", "--model", "opus", "--effort", "medium", skip]
    );
    // The legacy `session` subcommand is unchanged too.
    let (_, v, _) = dry(&["session", "--name", "s1"], None);
    assert_eq!(v["role"], "agent");
    assert_eq!(
        argv(&v),
        ["claude", "--name", "s1", "--model", "opus", "--effort", "medium"]
    );
}
