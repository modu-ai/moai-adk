//! SPEC-20261118-launcher-agent-entry (card C-47 slice; user decision 2026-11-18, card note
//! 2092): `-l|-a` only name a session and set its access, and the role is entered inside the
//! session with `/moai --mode-agent` / `/moai --mode-leader` (Codex: `$moai --mode-agent` /
//! `$moai --mode-leader`). The help and the `--dry-run` notes say so, and `moai codex -a` starts
//! its first prompt with `$moai --mode-agent`. Every case uses `--dry-run` in a temporary git
//! project with a HOME that does not exist; no host is started.
mod common;
use common::CleanEnv;
use serde_json::Value;
use std::process::Command;

const NAME_ONLY: &str = "only names the session and sets its access";

/// The launcher with `--dry-run` in a fresh git project; `token` sets a dummy GLM token.
fn dry(args: &[&str], token: Option<&str>) -> Value {
    let project = tempfile::tempdir().unwrap();
    let init = common::git::command()
        .args(["init", "-q"])
        .current_dir(project.path())
        .status()
        .unwrap();
    assert!(init.success());
    let mut cmd = Command::new(common::MOAI_BIN);
    cmd.args(args)
        .arg("--dry-run")
        .current_dir(project.path())
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        .env(
            "HOME",
            std::env::temp_dir().join("moai-cli-test-no-such-home"),
        );
    if let Some(t) = token {
        cmd.env("ANTHROPIC_AUTH_TOKEN", t);
    }
    let out = cmd.output().unwrap();
    assert!(out.status.success(), "{args:?}: {out:?}");
    serde_json::from_slice(&out.stdout).unwrap()
}

fn argv(v: &Value) -> Vec<String> {
    serde_json::from_value(v["argv"].clone()).unwrap()
}

fn notes(v: &Value) -> Vec<String> {
    serde_json::from_value(v["notes"].clone()).unwrap()
}

/// `moai <args> --help` in English, with an empty temporary HOME, its whitespace runs (the help
/// wraps its lines at the terminal width) collapsed to one space.
fn help(args: &[&str]) -> String {
    let home = tempfile::tempdir().unwrap();
    let out = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("HOME", home.path())
        .env("LANG", "C")
        .args(args)
        .arg("--help")
        .output()
        .unwrap();
    assert!(out.status.success(), "{args:?}: {out:?}");
    let text = String::from_utf8(out.stdout).unwrap();
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

/// AC-1 (REQ-1): `codex -a` starts its first prompt with the mode line, then the name line, then
/// the extra prompt; `--name` and the top-level form keep it.
#[test]
fn codex_agent_prompt_starts_with_the_mode_line() {
    let v = dry(&["codex", "-a"], None);
    assert_eq!(
        argv(&v),
        [
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "$moai --mode-agent\nMoAI session name: agent-1"
        ]
    );
    let v = dry(&["codex", "-a", "fix the login"], None);
    assert_eq!(
        argv(&v).last().unwrap(),
        "$moai --mode-agent\nMoAI session name: agent-1\nfix the login"
    );
    let v = dry(&["codex", "-a", "--name", "x"], None);
    assert_eq!(
        argv(&v).last().unwrap(),
        "$moai --mode-agent\nMoAI session name: x"
    );
    let v = dry(&["-a", "codex"], None);
    assert_eq!(
        argv(&v).last().unwrap(),
        "$moai --mode-agent\nMoAI session name: agent-1"
    );
    // The prompts selection keeps the line too.
    let v = dry(&["codex", "-a", "--permissions"], None);
    assert_eq!(
        argv(&v),
        [
            "codex",
            "-s",
            "workspace-write",
            "-a",
            "on-request",
            "$moai --mode-agent\nMoAI session name: agent-1"
        ]
    );
}

/// AC-2 (REQ-2): each role session's notes name the mode command for its host and say that the
/// flag only names the session and sets its access; a roleless session's notes name none.
#[test]
fn role_notes_say_how_the_role_is_entered() {
    for (args, token, command) in [
        (&["cc", "-l"][..], None, "`/moai --mode-leader`"),
        (&["cc", "-a"][..], None, "`/moai --mode-agent`"),
        (&["glm", "-l"][..], Some("t"), "`/moai --mode-leader`"),
        (&["glm", "-a"][..], Some("t"), "`/moai --mode-agent`"),
        (&["codex", "-l"][..], None, "`$moai --mode-leader`"),
        (&["codex", "-a"][..], None, "`$moai --mode-agent`"),
        (&["-l"][..], None, "`/moai --mode-leader`"),
        (&["-a", "codex"][..], None, "`$moai --mode-agent`"),
    ] {
        let notes = notes(&dry(args, token));
        let entry: Vec<&String> = notes.iter().filter(|n| n.contains("mode-")).collect();
        assert_eq!(entry.len(), 1, "{args:?}: {notes:?}");
        assert!(entry[0].contains(command), "{args:?}: {}", entry[0]);
        assert!(entry[0].contains(NAME_ONLY), "{args:?}: {}", entry[0]);
        assert!(entry[0].contains("inside the session"), "{}", entry[0]);
    }
    for (args, token) in [
        (&["cc"][..], None),
        (&["glm"][..], Some("t")),
        (&["codex"][..], None),
    ] {
        let v = dry(args, token);
        assert!(!v["notes"].to_string().contains("mode-"), "{args:?}: {v}");
    }
}

/// AC-3 (REQ-3): the long help of the per-host `-l` and `-a` and the launcher long about say how
/// the role is entered, for each host with its own spelling.
#[test]
fn the_help_says_how_the_role_is_entered() {
    for (host, prefix) in [("cc", "/moai"), ("glm", "/moai"), ("codex", "$moai")] {
        let text = help(&[host]);
        for mode in ["leader", "agent"] {
            let command = format!("`{prefix} --mode-{mode}`");
            assert!(
                text.contains(&command),
                "{host} --help lacks {command}:\n{text}"
            );
        }
        assert!(
            text.contains("only name the session and set its access"),
            "{host}:\n{text}"
        );
        assert!(
            text.contains("Names the session `leader`"),
            "{host} -l:\n{text}"
        );
        assert!(text.contains("inside the session"), "{host}:\n{text}");
    }
    // Codex says where the agent line goes.
    assert!(help(&["codex"]).contains("first prompt"));
}

/// AC-7 (REQ-7): the top-level `-l` and `-a` one-line help of `moai help --all` names the entry
/// command.
#[test]
fn the_top_level_flags_name_the_entry_command() {
    let home = tempfile::tempdir().unwrap();
    let out = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("HOME", home.path())
        .env("LANG", "C")
        .env("MOAI_LANG", "en")
        .args(["help", "--all"])
        .output()
        .unwrap();
    assert!(out.status.success(), "{out:?}");
    let text = String::from_utf8(out.stdout).unwrap();
    let block = |flag: &str| -> String {
        let start = text.find(flag).unwrap_or_else(|| panic!("{flag}:\n{text}"));
        let rest = &text[start..];
        let end = rest[2..].find("\n  -").map_or(rest.len(), |i| i + 2);
        rest[..end].split_whitespace().collect::<Vec<_>>().join(" ")
    };
    let leader = block("-l, --leader");
    assert!(leader.contains("`/moai --mode-leader`"), "{leader}");
    let agent = block("-a, --agent");
    assert!(agent.contains("`/moai --mode-agent`"), "{agent}");
    assert!(agent.contains("`$moai --mode-agent`"), "{agent}");
}

/// AC-8 (REQ-8): the root screen's `cc -l` example names the command that enters the role, on
/// one line at 60 and 80 columns (the launcher line itself has no room left, SPEC D-4).
#[test]
fn the_root_example_names_the_entry_command() {
    for columns in ["60", "80", ""] {
        let home = tempfile::tempdir().unwrap();
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.clean_env()
            .env("HOME", home.path())
            .env("LANG", "C")
            .arg("--help");
        if !columns.is_empty() {
            cmd.env("COLUMNS", columns);
        }
        let out = cmd.output().unwrap();
        assert!(out.status.success(), "{out:?}");
        let text = String::from_utf8(out.stdout).unwrap();
        let line = text
            .lines()
            .find(|l| l.trim_start().starts_with("moai cc -l "))
            .unwrap_or_else(|| panic!("COLUMNS={columns}: no cc -l example:\n{text}"));
        assert!(
            line.ends_with("leader, then /moai --mode-leader"),
            "COLUMNS={columns}: {line:?}"
        );
    }
}

/// AC-4 (REQ-4): the other launches compose the argv they composed before.
#[test]
fn other_launches_keep_their_argv() {
    let claude = |name: &str| {
        vec![
            "claude".to_string(),
            "--name".into(),
            name.into(),
            "--model".into(),
            "opus".into(),
            "--effort".into(),
            "medium".into(),
            "--dangerously-skip-permissions".into(),
        ]
    };
    assert_eq!(argv(&dry(&["cc", "-l"], None)), claude("leader"));
    assert_eq!(argv(&dry(&["cc", "-a"], None)), claude("agent-1"));
    assert_eq!(argv(&dry(&["glm", "-l"], Some("t"))), claude("leader"));
    assert_eq!(argv(&dry(&["glm", "-a"], Some("t"))), claude("agent-1"));
    assert_eq!(
        argv(&dry(&["codex", "-l"], None)),
        [
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "MoAI session name: leader"
        ]
    );
    assert_eq!(
        argv(&dry(&["codex", "-l", "plan the week"], None))
            .last()
            .unwrap(),
        "MoAI session name: leader\nplan the week"
    );
    assert_eq!(argv(&dry(&["codex"], None)), ["codex"]);
    assert_eq!(argv(&dry(&["codex", "hi"], None)), ["codex", "hi"]);
    assert_eq!(
        argv(&dry(&["cc"], None)),
        ["claude", "--model", "opus", "--effort", "medium"]
    );
}

/// AC-5 (REQ-5): every mode command the launcher help and notes name is one of the six
/// spellings of #2286 (the four of Claude Code and Codex, and OMP's `/skill:moai` form from
/// SPEC-20261118-omp-launcher).
#[test]
fn only_the_two_spellings_appear() {
    let mut texts: Vec<String> = ["cc", "glm", "codex"].iter().map(|h| help(&[h])).collect();
    texts.push(help(&[]));
    for (args, token) in [
        (&["cc", "-l"][..], None),
        (&["cc", "-a"][..], None),
        (&["glm", "-a"][..], Some("t")),
        (&["codex", "-l"][..], None),
        (&["codex", "-a"][..], None),
    ] {
        let v = dry(args, token);
        texts.push(v["notes"].to_string());
        texts.push(v["argv"].to_string());
    }
    let allowed = [
        "/moai --mode-agent",
        "/moai --mode-leader",
        "$moai --mode-agent",
        "$moai --mode-leader",
        "/skill:moai --mode-agent",
        "/skill:moai --mode-leader",
    ];
    let mut seen = 0;
    for text in &texts {
        for (at, _) in text.match_indices("mode-") {
            // The occurrence is the `mode-` of one allowed spelling written out in full.
            let ok = allowed.iter().any(|a| {
                let off = a.find("mode-").unwrap();
                at >= off && text[at - off..].starts_with(a)
            });
            let near = &text[at.saturating_sub(12)..(at + 14).min(text.len())];
            assert!(ok, "another spelling near: {near}");
            seen += 1;
        }
        for bad in [
            "--mode agent",
            "--mode leader",
            "--agent-mode",
            "--leader-mode",
        ] {
            assert!(!text.contains(bad), "{bad}");
        }
    }
    assert!(seen >= 10, "{seen}");
}
