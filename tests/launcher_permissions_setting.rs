//! #2244 (I7, REQ-I7-01): when the project's `.claude/settings.json` (in the directory the
//! launcher runs from) has a `permissions` key, a role session omits
//! `--dangerously-skip-permissions` and says so on stderr. Every case is `--dry-run` in a
//! temporary project; no host is started and no real settings file is read or written.
use serde_json::Value;
use std::{fs, path::Path, process::Command};

const SKIP: &str = "--dangerously-skip-permissions";

fn dry(dir: &Path, args: &[&str]) -> (i32, Vec<String>, String, String) {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .arg("--dry-run")
        .current_dir(dir)
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    let argv: Vec<String> = serde_json::from_value(v["argv"].clone()).unwrap_or_default();
    (
        out.status.code().unwrap_or(-1),
        argv,
        v["notes"].to_string(),
        String::from_utf8_lossy(&out.stderr).to_string(),
    )
}

fn project(settings: Option<&str>) -> tempfile::TempDir {
    let p = tempfile::tempdir().unwrap();
    if let Some(text) = settings {
        fs::create_dir(p.path().join(".claude")).unwrap();
        fs::write(p.path().join(".claude/settings.json"), text).unwrap();
    }
    p
}

#[test]
fn permissions_key_present_omits_the_skip_flag_and_says_so_on_stderr() {
    for (settings, args) in [
        (
            r#"{"permissions":{"allow":["Bash(cargo test *)"]}}"#,
            &["cc", "-l"][..],
        ),
        (r#"{"model":"opus","permissions":{}}"#, &["cc", "-a"][..]),
    ] {
        let p = project(Some(settings));
        let (code, argv, notes, stderr) = dry(p.path(), args);
        assert_eq!(code, 0, "{stderr}");
        assert!(!argv.iter().any(|a| a == SKIP), "{settings}: {argv:?}");
        assert!(
            stderr.contains("permissions key") && stderr.contains(SKIP),
            "stderr names the reason: {stderr}"
        );
        assert!(notes.contains("permissions key"), "{notes}");
    }
}

#[test]
fn no_permissions_key_keeps_the_interim_default() {
    for settings in [None, Some(r#"{"model":"opus"}"#)] {
        let p = project(settings);
        let (code, argv, _, stderr) = dry(p.path(), &["cc", "-l"]);
        assert_eq!(code, 0, "{stderr}");
        assert!(argv.iter().any(|a| a == SKIP), "{settings:?}: {argv:?}");
        assert!(!stderr.contains("permissions key"), "{stderr}");
    }
}

#[test]
fn unreadable_settings_keep_the_default_and_say_why() {
    let p = project(Some("{ not json"));
    let (code, argv, notes, stderr) = dry(p.path(), &["cc", "-l"]);
    assert_eq!(code, 0, "{stderr}");
    assert!(argv.iter().any(|a| a == SKIP), "{argv:?}");
    assert!(notes.contains("unreadable"), "{notes}");
}

#[test]
fn explicit_permissions_flag_roleless_and_codex_are_unchanged() {
    let p = project(Some(r#"{"permissions":{}}"#));
    // --permissions already drops the skip; no notice is needed.
    let (_, argv, _, stderr) = dry(p.path(), &["cc", "-l", "--permissions"]);
    assert!(!argv.iter().any(|a| a == SKIP));
    assert!(!stderr.contains("permissions key"), "{stderr}");
    // A roleless session never had the skip.
    let (_, argv, _, stderr) = dry(p.path(), &["cc"]);
    assert!(!argv.iter().any(|a| a == SKIP));
    assert!(!stderr.contains("permissions key"), "{stderr}");
    // Codex argv never carries it.
    let (_, argv, _, stderr) = dry(p.path(), &["codex", "-a"]);
    assert!(!argv.iter().any(|a| a == SKIP));
    assert!(!stderr.contains("permissions key"), "{stderr}");
}
