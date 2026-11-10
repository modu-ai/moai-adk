//! Issue #2359: without `ANTHROPIC_AUTH_TOKEN` in the environment, `glm` takes the token from
//! `~/.moai/.env.glm` (the global store shared with the Go tool), then `~/.moai-cli/.env.glm`.
//! Both are read only. The value reaches the host process only and never appears in the
//! launcher's own output.
use serde_json::{json, Value};
use std::{fs, path::Path, process::Command};

const TOKEN: &str = "moai-test-file-token-not-real";
const ENV_TOKEN: &str = "moai-test-env-token-not-real";

fn home(files: &[(&str, &str)]) -> tempfile::TempDir {
    let home = tempfile::tempdir().unwrap();
    for (name, text) in files {
        let path = home.path().join(name);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    home
}

fn launcher(home: &Path, env_token: Option<&str>) -> Command {
    let mut cmd = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    cmd.env("HOME", home)
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env_remove("ANTHROPIC_AUTH_TOKEN");
    if let Some(t) = env_token {
        cmd.env("ANTHROPIC_AUTH_TOKEN", t);
    }
    cmd
}

/// `(exit code, stdout JSON, stdout + stderr)` of `glm -a --dry-run`.
fn dry(home: &Path, env_token: Option<&str>) -> (i32, Value, String) {
    let out = launcher(home, env_token)
        .args(["glm", "-a", "--dry-run"])
        .output()
        .unwrap();
    let stdout = String::from_utf8_lossy(&out.stdout).to_string();
    let text = format!("{stdout}{}", String::from_utf8_lossy(&out.stderr));
    (
        out.status.code().unwrap_or(-1),
        serde_json::from_str(&stdout).unwrap_or(Value::Null),
        text,
    )
}

#[test]
fn key_file_stands_in_for_the_variable_without_showing_the_value() {
    for (file, line) in [
        (
            ".moai-cli/.env.glm",
            format!("ANTHROPIC_AUTH_TOKEN={TOKEN}\n"),
        ),
        (".moai-cli/.env.glm", format!("GLM_API_KEY={TOKEN}\n")),
        (".moai/.env.glm", format!("# z.ai\nGLM_API_KEY={TOKEN}\n")),
        (
            ".moai/.env.glm",
            format!("export GLM_API_KEY=\"{TOKEN}\"\n"),
        ),
    ] {
        let h = home(&[(file, &line)]);
        let (code, v, text) = dry(h.path(), None);
        assert_eq!(code, 0, "{file} {line:?}: {text}");
        assert!(!text.contains(TOKEN), "token value echoed for {file}");
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
        assert!(v["notes"].to_string().contains("key file"), "{text}");
        // Read only: the key file is byte-identical and nothing was added to the home.
        assert_eq!(fs::read_to_string(h.path().join(file)).unwrap(), line);
        assert_eq!(fs::read_dir(h.path()).unwrap().count(), 1);
    }
}

#[test]
fn no_usable_key_anywhere_is_blocked() {
    for files in [
        vec![],
        vec![(".moai-cli/.env.glm", "GLM_API_KEY=\n")],
        vec![(".moai/.env.glm", "GLM_API_KEY=\"\"\n")],
        vec![(".moai/.env.glm", "OTHER_KEY=value\n")],
        vec![(".moai/.env.glm", "# GLM_API_KEY=commented\n")],
    ] {
        let h = home(&files);
        let (code, v, text) = dry(h.path(), None);
        assert_ne!(code, 0, "{files:?}");
        assert_eq!(v, Value::Null, "{files:?}");
        assert!(
            text.contains("blocked") && text.contains("ANTHROPIC_AUTH_TOKEN"),
            "{text}"
        );
        assert!(text.contains(".moai-cli/.env.glm"), "{text}");
    }
}

#[test]
fn an_unreadable_or_malformed_key_file_is_blocked_without_a_panic_or_an_echo() {
    // A directory where the file should be.
    let h = tempfile::tempdir().unwrap();
    fs::create_dir_all(h.path().join(".moai-cli/.env.glm")).unwrap();
    let (code, _, text) = dry(h.path(), None);
    assert_eq!(code, 1, "{text}");
    assert!(
        text.contains("blocked") && !text.contains("panicked"),
        "{text}"
    );
    // Bytes that are not UTF-8, and a value carrying a control character.
    for bytes in [
        b"GLM_API_KEY=\xff\xfebad\n".to_vec(),
        b"GLM_API_KEY=bad\x1bvalue-not-shown\n".to_vec(),
    ] {
        let h = tempfile::tempdir().unwrap();
        fs::create_dir(h.path().join(".moai")).unwrap();
        fs::write(h.path().join(".moai/.env.glm"), bytes).unwrap();
        let (code, _, text) = dry(h.path(), None);
        assert_eq!(code, 1, "{text}");
        assert!(
            text.contains("blocked") && !text.contains("panicked"),
            "{text}"
        );
        assert!(!text.contains("value-not-shown") && !text.contains("bad"));
    }
}

#[cfg(unix)]
#[test]
fn an_empty_global_file_falls_through_to_the_rust_home_one() {
    let h = home(&[
        (".moai/.env.glm", "GLM_API_KEY=\n"),
        (".moai-cli/.env.glm", "GLM_API_KEY=from-rust-home\n"),
    ]);
    assert_eq!(host_token(h.path(), None), "from-rust-home");
}

/// A FIFO at the key path has no writer: opening it would block for ever. The launcher must
/// skip it and end `blocked` at once.
#[cfg(unix)]
#[test]
fn a_fifo_at_the_key_path_does_not_hang_the_launcher() {
    let h = tempfile::tempdir().unwrap();
    fs::create_dir(h.path().join(".moai")).unwrap();
    let made = Command::new("mkfifo")
        .arg(h.path().join(".moai/.env.glm"))
        .status()
        .unwrap();
    assert!(made.success());
    let mut child = launcher(h.path(), None)
        .args(["glm", "-a", "--dry-run"])
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()
        .unwrap();
    let started = std::time::Instant::now();
    let status = loop {
        if let Some(status) = child.try_wait().unwrap() {
            break status;
        }
        if started.elapsed() > std::time::Duration::from_secs(20) {
            child.kill().unwrap();
            child.wait().unwrap();
            panic!("the launcher hung on a FIFO key file");
        }
        std::thread::sleep(std::time::Duration::from_millis(20));
    };
    assert_eq!(status.code(), Some(1));
}

/// The size limit is 65536 bytes: a file of exactly that size is read, one byte more is not.
#[test]
fn the_key_file_size_limit_is_65536_bytes() {
    let line = format!("GLM_API_KEY={TOKEN}\n");
    for (size, code) in [(65536usize, 0), (65537usize, 1)] {
        let mut text = line.clone();
        text.push('#');
        text.push_str(&"x".repeat(size - line.len() - 2));
        text.push('\n');
        assert_eq!(text.len(), size);
        let h = home(&[(".moai/.env.glm", &text)]);
        let (got, _, out) = dry(h.path(), None);
        assert_eq!(got, code, "size {size}: {out}");
        assert!(!out.contains(TOKEN));
    }
}

#[test]
fn the_variable_wins_and_the_file_is_not_reported() {
    let h = home(&[(".moai-cli/.env.glm", "GLM_API_KEY=file-value-unused\n")]);
    let (code, v, text) = dry(h.path(), Some(ENV_TOKEN));
    assert_eq!(code, 0, "{text}");
    assert!(!text.contains(ENV_TOKEN) && !text.contains("file-value-unused"));
    assert!(!v["notes"].to_string().contains("key file"));
}

#[cfg(unix)]
#[test]
fn the_global_file_is_read_before_the_rust_home_one() {
    let h = home(&[
        (".moai-cli/.env.glm", "GLM_API_KEY=from-rust-home\n"),
        (".moai/.env.glm", "GLM_API_KEY=from-global\n"),
    ]);
    assert_eq!(host_token(h.path(), None), "from-global");
}

/// Starts a stand-in `claude` that records the token it was given, and returns that record.
#[cfg(unix)]
fn host_token(home: &Path, env_token: Option<&str>) -> String {
    use std::os::unix::fs::PermissionsExt;
    let bin = tempfile::tempdir().unwrap();
    let seen = bin.path().join("seen");
    let fake = bin.path().join("claude");
    fs::write(
        &fake,
        format!(
            "#!/bin/sh\nprintf %s \"$ANTHROPIC_AUTH_TOKEN|$ANTHROPIC_BASE_URL|$API_TIMEOUT_MS|$ANTHROPIC_DEFAULT_OPUS_MODEL|$ANTHROPIC_DEFAULT_SONNET_MODEL|$ANTHROPIC_DEFAULT_HAIKU_MODEL|$ANTHROPIC_DEFAULT_FABLE_MODEL|$CLAUDE_CODE_AUTO_COMPACT_WINDOW|$CLAUDE_CODE_MAX_CONTEXT_TOKENS|$CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS|$Z_AI_API_KEY\" > '{}'\n",
            seen.display()
        ),
    )
    .unwrap();
    fs::set_permissions(&fake, fs::Permissions::from_mode(0o755)).unwrap();
    let out = launcher(home, env_token)
        .env("PATH", format!("{}:/usr/bin:/bin", bin.path().display()))
        .args(["glm", "-a"])
        .output()
        .unwrap();
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(out.status.success(), "{text}");
    let record = fs::read_to_string(&seen).unwrap();
    let (token, rest) = record.split_once('|').unwrap();
    // The connection set of the Go tool, pinned by literal value; the last field is the
    // same key under the name the z.ai MCP server reads.
    assert_eq!(
        rest,
        format!(
            "https://api.z.ai/api/anthropic|3000000|glm-5.3|glm-5.3|glm-5.3-flash|glm-5.3|1000000|1000000|1|{token}"
        )
    );
    assert!(!text.contains(token), "token value echoed by the launcher");
    token.to_string()
}

#[cfg(unix)]
#[test]
fn the_host_process_receives_the_token() {
    let h = home(&[(".moai/.env.glm", &format!("GLM_API_KEY={TOKEN}\n"))]);
    assert_eq!(host_token(h.path(), None), TOKEN);
    // With the variable set the child inherits it and the file is ignored.
    assert_eq!(host_token(h.path(), Some(ENV_TOKEN)), ENV_TOKEN);
}
