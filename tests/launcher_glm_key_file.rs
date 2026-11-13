//! Issue #2359, changed by #2393 (user decision, Issue #2393 comment 6778150969): `glm` takes
//! `ANTHROPIC_AUTH_TOKEN` from the environment, then from the moai-cli vault. The legacy key
//! files `~/.moai/.env.glm` and `~/.moai-cli/.env.glm` are never opened or read: with no
//! vault entry, the presence of one ends `blocked: legacy_key_only`, naming the file and the
//! way to move the key, and the file is left as it is. The value reaches the host process
//! only and never appears in the launcher's own output.
mod common;
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
    sync::OnceLock,
};

const TOKEN: &str = "moai-test-file-token-not-real";
const ENV_TOKEN: &str = "moai-test-env-token-not-real";

fn home(files: &[(&str, &[u8])]) -> tempfile::TempDir {
    let home = tempfile::tempdir().unwrap();
    for (name, bytes) in files {
        let path = home.path().join(name);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, bytes).unwrap();
    }
    home
}

/// A git project outside HOME: `glm -a` reads and takes its `agent-N` number there (#2359),
/// never in this repository's own counter. #2655: one fresh temporary folder per test
/// process — cargo-nextest runs every test as its own process, and a shared fixed folder made
/// their `git init` runs collide on `.git/config`'s lock.
fn project() -> &'static Path {
    static PROJECT: OnceLock<tempfile::TempDir> = OnceLock::new();
    PROJECT
        .get_or_init(|| {
            let p = tempfile::Builder::new()
                .prefix("launcher-glm-key-file-project-")
                .tempdir()
                .unwrap();
            let init = common::git::command()
                .args(["init", "-q"])
                .current_dir(p.path())
                .status()
                .unwrap();
            assert!(init.success());
            p
        })
        .path()
}

fn launcher(home: &Path, env_token: Option<&str>) -> Command {
    let mut cmd = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    cmd.current_dir(project())
        .env("HOME", home)
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

/// Every file under HOME with its bytes (a snapshot that shows nothing was created or changed).
fn snapshot(home: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![home.to_path_buf()];
    while let Some(p) = stack.pop() {
        let m = fs::symlink_metadata(&p).unwrap();
        if m.is_dir() {
            for e in fs::read_dir(&p).unwrap() {
                stack.push(e.unwrap().path());
            }
        } else {
            out.push((p.clone(), fs::read(&p).unwrap_or_default()));
        }
    }
    out.sort();
    out
}

/// A legacy file is named, never read: a valid key, an empty or malformed one, a value with a
/// control character, a file over 64 KiB — all end `legacy_key_only` with the same answer,
/// the value is not shown, and HOME is unchanged.
#[test]
fn a_legacy_key_file_is_named_and_never_read() {
    let mut big = format!("GLM_API_KEY={TOKEN}\n#").into_bytes();
    big.resize(65_537, b'x');
    for file in [".moai/.env.glm", ".moai-cli/.env.glm"] {
        for bytes in [
            format!("GLM_API_KEY={TOKEN}\n").into_bytes(),
            format!("export ANTHROPIC_AUTH_TOKEN=\"{TOKEN}\"\n").into_bytes(),
            b"GLM_API_KEY=\n".to_vec(),
            b"OTHER_KEY=value\n".to_vec(),
            b"GLM_API_KEY=\xff\xfebad\n".to_vec(),
            format!("GLM_API_KEY={TOKEN}\x1bx\n").into_bytes(),
            big.clone(),
        ] {
            let h = home(&[(file, &bytes)]);
            let before = snapshot(h.path());
            let (code, v, text) = dry(h.path(), None);
            assert_eq!(code, 1, "{file}: {text}");
            assert_eq!(v, Value::Null, "{file}");
            assert!(text.contains("blocked: legacy_key_only"), "{file}: {text}");
            assert!(text.contains(file), "{file}: {text}");
            assert!(text.contains("moai-cli key-set glm"), "{text}");
            assert!(!text.contains(TOKEN), "value shown for {file}");
            assert!(!text.contains("panicked"), "{text}");
            assert_eq!(snapshot(h.path()), before, "{file} changed");
        }
    }
}

/// Entries that cannot be read as files still count as present and are never opened: a
/// directory, a file with mode 000, and a FIFO with no writer (opening it would block).
#[cfg(unix)]
#[test]
fn an_unreadable_entry_at_the_legacy_path_is_named_without_opening_it() {
    use std::os::unix::fs::PermissionsExt;
    let h = tempfile::tempdir().unwrap();
    fs::create_dir_all(h.path().join(".moai-cli/.env.glm")).unwrap();
    let (code, _, text) = dry(h.path(), None);
    assert_eq!(code, 1, "{text}");
    assert!(text.contains("blocked: legacy_key_only"), "{text}");

    let h = home(&[(
        ".moai/.env.glm",
        format!("GLM_API_KEY={TOKEN}\n").as_bytes(),
    )]);
    let file = h.path().join(".moai/.env.glm");
    fs::set_permissions(&file, fs::Permissions::from_mode(0o000)).unwrap();
    let (code, _, text) = dry(h.path(), None);
    fs::set_permissions(&file, fs::Permissions::from_mode(0o600)).unwrap();
    assert_eq!(code, 1, "{text}");
    assert!(text.contains("blocked: legacy_key_only"), "{text}");

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
        // #2511: a hang guard (a FIFO read blocks forever), so the bound is generous.
        if started.elapsed() > std::time::Duration::from_secs(120) {
            child.kill().unwrap();
            child.wait().unwrap();
            panic!("the launcher hung on a FIFO key file");
        }
        std::thread::sleep(std::time::Duration::from_millis(20));
    };
    assert_eq!(status.code(), Some(1));
}

/// With both legacy files present, the Go tool's global one is named.
#[test]
fn the_global_file_is_named_first() {
    let h = home(&[
        (".moai-cli/.env.glm", b"GLM_API_KEY=a\n"),
        (".moai/.env.glm", b"GLM_API_KEY=b\n"),
    ]);
    let (_, _, text) = dry(h.path(), None);
    assert!(text.contains("~/.moai/.env.glm"), "{text}");
    assert!(!text.contains(".moai-cli/.env.glm"), "{text}");
}

/// No variable, no vault entry and no legacy file: `key_absent`, naming the vault command.
#[test]
fn no_key_anywhere_is_blocked() {
    let h = home(&[]);
    let (code, v, text) = dry(h.path(), None);
    assert_eq!(code, 1, "{text}");
    assert_eq!(v, Value::Null);
    assert!(text.contains("blocked: key_absent"), "{text}");
    assert!(text.contains("ANTHROPIC_AUTH_TOKEN") && text.contains("key-set glm"));
    assert!(!text.contains(".env.glm"), "{text}");
}

#[test]
fn the_variable_wins_and_the_legacy_file_is_not_touched() {
    let h = home(&[(".moai-cli/.env.glm", b"GLM_API_KEY=file-value-unused\n")]);
    let before = snapshot(h.path());
    let (code, v, text) = dry(h.path(), Some(ENV_TOKEN));
    assert_eq!(code, 0, "{text}");
    assert!(!text.contains(ENV_TOKEN) && !text.contains("file-value-unused"));
    assert_eq!(v["key_source"], "env:ANTHROPIC_AUTH_TOKEN");
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
    assert_eq!(snapshot(h.path()), before);
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
    if token.len() >= 8 {
        assert!(!text.contains(token), "token value echoed by the launcher");
    }
    token.to_string()
}

/// AC4 boundary for the variable (audit F1 on 8e6cb1e4a4): an empty or whitespace-only value is
/// no token, a one-character value is one. Literal values, not a shared constant.
#[test]
fn an_empty_or_blank_variable_is_no_token_and_one_character_is_one() {
    let bare = home(&[]);
    for value in ["", " ", "   ", "\t"] {
        let (code, v, text) = dry(bare.path(), Some(value));
        assert_eq!(code, 1, "env {value:?}: {text}");
        assert_eq!(v, Value::Null, "env {value:?}");
        assert!(
            text.contains("blocked") && text.contains("ANTHROPIC_AUTH_TOKEN"),
            "env {value:?}: {text}"
        );
    }
    let (code, _, text) = dry(bare.path(), Some("k"));
    assert_eq!(code, 0, "env \"k\": {text}");
}

/// A blank variable is no token, and a legacy file behind it is still only named.
#[test]
fn a_blank_variable_with_a_legacy_file_is_legacy_key_only() {
    for file in [".moai/.env.glm", ".moai-cli/.env.glm"] {
        let h = home(&[(file, b"GLM_API_KEY=from-file\n")]);
        for value in ["", "   "] {
            let (code, _, text) = dry(h.path(), Some(value));
            assert_eq!(code, 1, "{file} env {value:?}: {text}");
            assert!(text.contains("blocked: legacy_key_only"), "{text}");
        }
    }
}

#[cfg(unix)]
#[test]
fn the_host_process_receives_the_variable() {
    let h = home(&[(
        ".moai/.env.glm",
        format!("GLM_API_KEY={TOKEN}\n").as_bytes(),
    )]);
    assert_eq!(host_token(h.path(), Some(ENV_TOKEN)), ENV_TOKEN);
    assert_eq!(host_token(h.path(), Some("k")), "k");
}
