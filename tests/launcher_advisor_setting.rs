//! Issue #2476 (user decision 2026-11-12, replacing the 2026-11-08 "advisor fable" default):
//! a Claude session gets no `--advisor` unless CLI `--advisor <model>` or the user-only
//! `~/.moai-cli/config/launcher.yaml` key `launcher.advisor` asks for one. Precedence is
//! `--no-advisor` / `--advisor` > `launcher.advisor` > none. Every home is a temporary
//! directory; real launches run stub `claude` / `codex` executables from a temporary PATH.
#![cfg(unix)]
mod common;
use serde_json::Value;
use std::os::unix::fs::PermissionsExt;
use std::{
    fs,
    path::Path,
    process::{Command, Output},
};

const HOSTS: [&str; 3] = ["cc", "glm", "codex"];
const ROLES: [&[&str]; 3] = [&["-l"], &["-a"], &[]];
/// Rejected `launcher.advisor` values and broken files. `SECRET` must never reach stderr.
const INVALID: [&str; 14] = [
    r#"{"launcher":{"advisor":"SECRET-unknown-advisor"}}"#,
    r#"{"launcher":{"advisor":null}}"#,
    r#"{"launcher":{"advisor":7}}"#,
    r#"{"launcher":{"advisor":true}}"#,
    r#"{"launcher":{"advisor":[]}}"#,
    r#"{"launcher":{"advisor":["fable"]}}"#,
    r#"{"launcher":{"advisor":{}}}"#,
    r#"{"launcher":{"advisor":""}}"#,
    r#"{"launcher":{"advisor":"Fable"}}"#,
    r#"{"launcher":{"advisor":"claude-SECRET with-space"}}"#,
    r#"{"launcher":{"advisor":"SECRET-claude-opus"}}"#,
    r#"{"launcher":{"advisor":"fable","role_access":"SECRET-bad-access"}}"#,
    "{ SECRET-malformed",
    "launcher:\n  advisor: SECRET-plain-yaml\n",
];

/// The note a launch carries when the advisor came from `launcher.advisor`.
const SOURCE_NOTE: &str = "--advisor comes from the user setting launcher.advisor";

struct Fixture {
    project: tempfile::TempDir,
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
}

fn fixture() -> Fixture {
    let project = tempfile::tempdir().unwrap();
    assert!(common::git::command()
        .args(["init", "-q"])
        .current_dir(project.path())
        .status()
        .unwrap()
        .success());
    let bin = tempfile::tempdir().unwrap();
    for host in ["claude", "codex"] {
        let path = bin.path().join(host);
        fs::write(
            &path,
            "#!/bin/sh\nprintf started > \"$MOAI_TEST_HOST_MARKER\"\nexit 0\n",
        )
        .unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    Fixture {
        project,
        bin,
        home: tempfile::tempdir().unwrap(),
    }
}

fn config(home: &Path, text: &str) {
    fs::create_dir_all(home.join(".moai-cli/config")).unwrap();
    fs::write(home.join(".moai-cli/config/launcher.yaml"), text).unwrap();
}

fn command(f: &Fixture, args: &[&str]) -> Command {
    let mut c = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    c.args(args)
        .current_dir(f.project.path())
        .env("PATH", format!("{}:/usr/bin:/bin", f.bin.path().display()))
        .env("HOME", f.home.path())
        .env("USERPROFILE", f.home.path())
        .env("MOAI_TEST_HOST_MARKER", f.home.path().join("host-started"))
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env_remove(moai_cli::session::AGENT_SEQ_TEST_HOLD_VAR);
    c
}

fn dry(f: &Fixture, args: &[&str]) -> Value {
    let out = command(f, args).arg("--dry-run").output().unwrap();
    assert!(
        out.status.success(),
        "{args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    serde_json::from_slice(&out.stdout).unwrap()
}

fn argv(v: &Value) -> Vec<String> {
    v["argv"]
        .as_array()
        .unwrap()
        .iter()
        .map(|x| x.as_str().unwrap().to_string())
        .collect()
}

/// A refused launch: exit 1, nothing on stdout, no echo of the rejected value.
fn rejected(out: &Output) -> String {
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert_eq!(out.status.code(), Some(1), "{stderr}");
    assert!(
        out.stdout.is_empty(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert!(!stderr.contains("SECRET"), "value echoed: {stderr}");
    stderr
}

fn host_started(f: &Fixture) -> bool {
    f.home.path().join("host-started").exists()
}
fn number_store(f: &Fixture) -> std::path::PathBuf {
    f.project.path().join(".git/.moai-cli/sessions/agent-seq")
}

/// Literal complete argv with no advisor: the decided default for every host and role
/// (role index 0 = `-l`, 1 = `-a`, 2 = none).
fn plain(host: &str, role: usize) -> Vec<&'static str> {
    match (host, role) {
        ("cc" | "glm", 0) => vec![
            "claude",
            "--name",
            "leader",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--dangerously-skip-permissions",
        ],
        ("cc" | "glm", 1) => vec![
            "claude",
            "--name",
            "agent-1",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--dangerously-skip-permissions",
        ],
        ("cc" | "glm", 2) => vec!["claude", "--model", "opus", "--effort", "medium"],
        ("codex", 0) => vec![
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "MoAI session name: leader",
        ],
        ("codex", 1) => vec![
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "MoAI session name: agent-1",
        ],
        ("codex", 2) => vec!["codex"],
        _ => unreachable!(),
    }
}

/// Literal complete `cc` argv with `--advisor <a>` in its documented place.
fn with_advisor(role: usize, advisor: &'static str) -> Vec<&'static str> {
    match role {
        0 => vec![
            "claude",
            "--name",
            "leader",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--advisor",
            advisor,
            "--dangerously-skip-permissions",
        ],
        1 => vec![
            "claude",
            "--name",
            "agent-1",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--advisor",
            advisor,
            "--dangerously-skip-permissions",
        ],
        2 => vec![
            "claude",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--advisor",
            advisor,
        ],
        _ => unreachable!(),
    }
}

fn args_for<'a>(host: &'a str, flags: &[&'a str], extra: &[&'a str]) -> Vec<&'a str> {
    let mut args = vec![host];
    args.extend_from_slice(flags);
    args.extend_from_slice(extra);
    args
}

/// AC-1: no setting file → no `--advisor` on cc/glm for `-l`, `-a` and no role; codex as before.
/// A file without the key, or `none`, gives byte-identical output.
#[test]
fn no_setting_means_no_advisor_on_every_host_and_role() {
    let f = fixture();
    let mut baseline = Vec::new();
    for host in HOSTS {
        for (role, flags) in ROLES.iter().enumerate() {
            let args = args_for(host, flags, &[]);
            let v = dry(&f, &args);
            assert_eq!(v["argv"], serde_json::json!(plain(host, role)), "{args:?}");
            assert!(
                !v["notes"].to_string().contains("usage credits"),
                "{args:?}: {}",
                v["notes"]
            );
            baseline.push(command(&f, &args).arg("--dry-run").output().unwrap());
        }
    }
    for text in [
        "{}",
        r#"{"launcher":{}}"#,
        r#"{"launcher":{"advisor":"none"}}"#,
    ] {
        config(f.home.path(), text);
        let mut i = 0;
        for host in HOSTS {
            for flags in ROLES {
                let out = command(&f, &args_for(host, flags, &[]))
                    .arg("--dry-run")
                    .output()
                    .unwrap();
                assert_eq!(out.stdout, baseline[i].stdout, "{text} {host} {flags:?}");
                assert_eq!(out.stderr, baseline[i].stderr);
                i += 1;
            }
        }
    }
}

/// AC-2: `--advisor` and `launcher.advisor` each attach the advisor; CLI beats the setting;
/// `--no-advisor` switches the setting off; glm drops it and codex never takes it.
#[test]
fn cli_and_setting_enable_the_advisor_with_cli_precedence() {
    let f = fixture();
    for (role, flags) in ROLES.iter().enumerate() {
        let v = dry(&f, &args_for("cc", flags, &["--advisor", "fable"]));
        assert_eq!(v["argv"], serde_json::json!(with_advisor(role, "fable")));
        assert!(
            v["notes"].to_string().contains("usage credits"),
            "{}",
            v["notes"]
        );
    }
    config(f.home.path(), r#"{"launcher":{"advisor":"fable"}}"#);
    for (role, flags) in ROLES.iter().enumerate() {
        let v = dry(&f, &args_for("cc", flags, &[]));
        assert_eq!(v["argv"], serde_json::json!(with_advisor(role, "fable")));
        let notes = v["notes"].to_string();
        assert!(notes.contains("usage credits"), "{notes}");
        assert!(notes.contains(SOURCE_NOTE), "{notes}");
        // CLI over setting.
        let v = dry(&f, &args_for("cc", flags, &["--advisor", "opus"]));
        assert_eq!(v["argv"], serde_json::json!(with_advisor(role, "opus")));
        assert!(!v["notes"].to_string().contains(SOURCE_NOTE));
        // `--no-advisor` over setting.
        let v = dry(&f, &args_for("cc", flags, &["--no-advisor"]));
        assert_eq!(v["argv"], serde_json::json!(plain("cc", role)));
        // glm: the advisor is Anthropic API only; codex: Claude hosts only.
        let v = dry(&f, &args_for("glm", flags, &[]));
        assert_eq!(v["argv"], serde_json::json!(plain("glm", role)));
        assert!(v["notes"].to_string().contains("--advisor omitted on glm"));
        assert_eq!(
            dry(&f, &args_for("codex", flags, &[]))["argv"],
            serde_json::json!(plain("codex", role))
        );
    }
    // A full id from the setting, and the legacy top-level `-l` / `-a cc` forms.
    config(
        f.home.path(),
        r#"{"launcher":{"advisor":"claude-fable-5-2"}}"#,
    );
    assert_eq!(
        dry(&f, &["-l"])["argv"],
        serde_json::json!(with_advisor(0, "claude-fable-5-2"))
    );
    assert_eq!(
        dry(&f, &["-a", "cc"])["argv"],
        serde_json::json!(with_advisor(1, "claude-fable-5-2"))
    );
    // Both access and advisor from the one user file.
    config(
        f.home.path(),
        r#"{"launcher":{"role_access":"prompt","advisor":"fable"}}"#,
    );
    assert_eq!(
        argv(&dry(&f, &["cc", "-l"])),
        [
            "claude",
            "--name",
            "leader",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--advisor",
            "fable"
        ]
    );
}

/// AC-2: `--advisor` with `--no-advisor` is refused before a host starts or a number is taken.
#[test]
fn advisor_and_no_advisor_together_are_refused() {
    let f = fixture();
    for setting in [None, Some(r#"{"launcher":{"advisor":"fable"}}"#)] {
        if let Some(text) = setting {
            config(f.home.path(), text);
        }
        for host in HOSTS {
            for flags in ROLES {
                for dry_run in [true, false] {
                    let mut c = command(
                        &f,
                        &args_for(host, flags, &["--advisor", "fable", "--no-advisor"]),
                    );
                    if dry_run {
                        c.arg("--dry-run");
                    }
                    let stderr = rejected(&c.output().unwrap());
                    assert!(stderr.contains("--no-advisor"), "{stderr}");
                    assert!(!host_started(&f), "{host} {flags:?}");
                    assert!(!number_store(&f).exists(), "{host} {flags:?}");
                }
            }
        }
    }
}

/// AC-3: an invalid setting is refused for every host and role, also when the CLI decides
/// (`--advisor`, `--no-advisor`); no host starts, no number is taken, nothing is echoed.
#[test]
fn invalid_setting_is_refused_without_host_or_number() {
    for bad in INVALID {
        let f = fixture();
        config(f.home.path(), bad);
        for host in HOSTS {
            for flags in ROLES {
                for cli in [&[][..], &["--advisor", "fable"][..], &["--no-advisor"][..]] {
                    let args = args_for(host, flags, cli);
                    rejected(&command(&f, &args).arg("--dry-run").output().unwrap());
                    rejected(&command(&f, &args).output().unwrap());
                    assert!(!host_started(&f), "{bad} {args:?}");
                    assert!(!number_store(&f).exists(), "{bad} {args:?}");
                }
            }
        }
    }
}

/// AC-3: a setting that names a lower-ranked advisor than the main model is refused at
/// compose time, still before a number is taken.
#[test]
fn lower_ranked_setting_is_refused_before_numbering() {
    let f = fixture();
    config(f.home.path(), r#"{"launcher":{"advisor":"sonnet"}}"#);
    let stderr = rejected(&command(&f, &["cc", "-a"]).output().unwrap());
    assert!(stderr.contains("rank"), "{stderr}");
    assert!(!host_started(&f));
    assert!(!number_store(&f).exists());
    // A lower main model makes the same setting valid.
    let v = dry(&f, &["cc", "-l", "--model", "sonnet"]);
    assert_eq!(
        argv(&v),
        [
            "claude",
            "--name",
            "leader",
            "--model",
            "sonnet",
            "--effort",
            "medium",
            "--advisor",
            "sonnet",
            "--dangerously-skip-permissions"
        ]
    );
}

/// The project file cannot turn the advisor on; only the user file can.
#[test]
fn project_setting_cannot_enable_the_advisor() {
    let f = fixture();
    fs::create_dir_all(f.project.path().join(".moai-cli")).unwrap();
    fs::write(
        f.project.path().join(".moai-cli/project.yaml"),
        r#"{"launcher":{"advisor":"fable"}}"#,
    )
    .unwrap();
    for host in HOSTS {
        for (role, flags) in ROLES.iter().enumerate() {
            assert_eq!(
                dry(&f, &args_for(host, flags, &[]))["argv"],
                serde_json::json!(plain(host, role))
            );
        }
    }
}

/// A real launch with the setting starts the stub host with the advisor argv and takes a number.
#[test]
fn real_launch_uses_the_setting() {
    let f = fixture();
    config(f.home.path(), r#"{"launcher":{"advisor":"fable"}}"#);
    let out = command(&f, &["cc", "-a"]).output().unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["argv"], serde_json::json!(with_advisor(1, "fable")));
    assert_eq!(v["executed"], serde_json::json!(true));
    assert!(host_started(&f));
    assert_eq!(fs::read_to_string(number_store(&f)).unwrap(), "1\n");
}
