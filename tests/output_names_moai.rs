//! #3004 P2 (public CLI name `moai`): a sentence the program prints that tells a person to run
//! a command names it `moai <command>`, never `moai-cli <command>`.
//!
//! A representative set of outputs is read: a launcher error, a launcher block that names the
//! vault command, the `next_actions` of `project-detect`, a usage error (stderr and envelope)
//! and the `--help` of a few subcommands. What may stay is a folder or file name
//! (`.moai-cli/…`, `~/.moai-cli`, `moai-cli.json`, `path/moai-cli`) and the `--version`
//! signature, which other tools compare; `as_command` does not count those.
//!
//! Every process runs with a throwaway HOME and working folder; the real HOME is never touched.

mod common;

use std::process::{Command, Output};
use tempfile::TempDir;

/// The lines of `text` that name `moai-cli` as a program a person types: not part of a longer
/// name (`moai-cli.json`, `moai-cli-home-`, `moai_cli`), not a folder or path (`.moai-cli/`,
/// `~/.moai-cli`, `path/moai-cli`) and not the `--version` signature (`moai-cli <ver> (rust)`).
fn as_command(text: &str) -> Vec<String> {
    let mut hits = Vec::new();
    for line in text.lines() {
        for (i, _) in line.match_indices("moai-cli") {
            let before = line[..i].chars().next_back();
            let after = line[i + "moai-cli".len()..].chars().next();
            let named_thing = matches!(before, Some('.' | '/' | '~' | '-' | '_'))
                || before.is_some_and(char::is_alphanumeric)
                || matches!(after, Some('/' | '-' | '.' | '_'))
                || after.is_some_and(char::is_alphanumeric);
            let signature = line.starts_with("moai-cli ") && line.trim_end().ends_with("(rust)");
            if !named_thing && !signature {
                hits.push(line.trim().to_string());
            }
        }
    }
    hits
}

/// Runs `bin args…` in a fresh working folder with a throwaway HOME; stdout and stderr joined.
fn output_of(bin: &str, args: &[&str]) -> String {
    let dir = TempDir::new().unwrap();
    let home = dir.path().join("home");
    std::fs::create_dir(&home).unwrap();
    let work = dir.path().join("work");
    std::fs::create_dir(&work).unwrap();
    let out: Output = Command::new(bin)
        .args(args)
        .env("HOME", &home)
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .current_dir(&work)
        .output()
        .unwrap();
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// The recogniser itself: it flags a sentence that tells a person to run `moai-cli …` and lets
/// folder names, paths, the marker file and the version signature through.
#[test]
fn the_recogniser_flags_commands_and_passes_names_of_things() {
    for flagged in [
        "run `moai-cli --help` to see the commands",
        "to undo: moai-cli agents-merge-revert --receipt r <project>",
        "use `moai-cli cc|glm -l|-a`",
        "store it with `moai-cli key-set glm`",
    ] {
        assert_eq!(as_command(flagged).len(), 1, "{flagged}");
    }
    for passed in [
        "state in `.moai-cli/receipts/`",
        "the folder ~/.moai-cli/vault",
        "default `${XDG_CACHE_HOME:-~/.cache}/moai-cli/heavy-slots`",
        "the marker moai-cli.json",
        "`bin/moai-cli` is the alias file",
        "moai-cli 0.1.0 (rust)",
        "run `moai key-set glm`",
    ] {
        assert!(
            as_command(passed).is_empty(),
            "{passed}: {:?}",
            as_command(passed)
        );
    }
}

/// Both binary names print the same sentences, and none tells the person to run `moai-cli`.
fn each_name(args: &[&str]) -> Vec<(String, String)> {
    [common::MOAI_BIN, common::MOAI_CLI_BIN]
        .iter()
        .map(|bin| ((*bin).to_string(), output_of(bin, args)))
        .collect()
}

fn assert_no_command_moai_cli(what: &str, args: &[&str]) -> Vec<String> {
    let mut texts = Vec::new();
    for (bin, text) in each_name(args) {
        let hits = as_command(&text);
        assert!(
            hits.is_empty(),
            "{what} ({bin} {args:?}) names moai-cli:\n{text}"
        );
        texts.push(text);
    }
    texts
}

/// The legacy `session` launcher says how to start the other hosts.
#[test]
fn the_session_launcher_error_says_moai() {
    let args = ["session", "--host", "glm", "--name", "guard", "--dry-run"];
    for text in assert_no_command_moai_cli("session --host glm", &args) {
        assert!(text.contains("use `moai cc|glm -l|-a`"), "{text}");
    }
}

/// With no `glm` key anywhere, the launcher block names the vault command.
#[test]
fn the_glm_key_block_says_moai_key_set() {
    let args = ["glm", "-a", "--name", "guard", "--dry-run"];
    for text in assert_no_command_moai_cli("glm key_absent", &args) {
        assert!(text.contains("blocked: key_absent"), "{text}");
        assert!(text.contains("`moai key-set glm`"), "{text}");
    }
}

/// `project-detect` suggests commands in `next_actions`.
#[test]
fn project_detect_next_actions_say_moai() {
    let dir = TempDir::new().unwrap();
    let project = dir.path().to_string_lossy().into_owned();
    let args = ["project-detect", project.as_str()];
    for text in assert_no_command_moai_cli("project-detect", &args) {
        assert!(text.contains("moai index-build "), "{text}");
    }
}

/// With a profile on disk, `project-detect` hands it on to `project-plan`.
#[test]
fn project_detect_with_a_profile_suggests_moai_project_plan() {
    let dir = TempDir::new().unwrap();
    let state = dir.path().join(".moai-cli");
    std::fs::create_dir(&state).unwrap();
    std::fs::write(
        state.join("project.yaml"),
        r#"{"schema_version":2,"project_id":"demo","purpose":"p","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude"]}"#,
    )
    .unwrap();
    let project = dir.path().to_string_lossy().into_owned();
    let args = ["project-detect", project.as_str()];
    for text in assert_no_command_moai_cli("project-detect with a profile", &args) {
        assert!(text.contains("`moai project-plan "), "{text}");
        assert!(text.contains("then `moai project-apply`"), "{text}");
    }
}

/// `language-setup` hands its profile on to `project-plan`.
#[test]
fn language_setup_next_actions_say_moai() {
    for bin in [common::MOAI_BIN, common::MOAI_CLI_BIN] {
        let dir = TempDir::new().unwrap();
        let home = dir.path().join("home");
        std::fs::create_dir(&home).unwrap();
        let project = dir.path().join("work");
        std::fs::create_dir_all(project.join(".moai-cli")).unwrap();
        std::fs::write(
            project.join(".moai-cli/project.yaml"),
            r#"{"schema_version":2,"project_id":"demo","purpose":"p","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude"]}"#,
        )
        .unwrap();
        let mut child = Command::new(bin)
            .args(["language-setup", project.to_str().unwrap(), "--answers"])
            .env("HOME", &home)
            .current_dir(&project)
            .stdin(std::process::Stdio::piped())
            .stdout(std::process::Stdio::piped())
            .stderr(std::process::Stdio::piped())
            .spawn()
            .unwrap();
        std::io::Write::write_all(&mut child.stdin.take().unwrap(), b"{}").unwrap();
        let out = child.wait_with_output().unwrap();
        let text = format!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        assert!(as_command(&text).is_empty(), "{bin}: {text}");
        assert!(
            text.contains("`moai project-plan <project>`"),
            "{bin}: {text}"
        );
    }
}

/// A usage error: the envelope and stderr say nothing about `moai-cli`. (The `next:` line is
/// terminal-only; its sentence is pinned in `error_next`'s unit tests and the tty test.)
#[test]
fn a_usage_error_says_moai() {
    for args in [&[][..], &["no-such-command"][..], &["identity"][..]] {
        assert_no_command_moai_cli("usage error", args);
    }
}

/// The `--help` of a few subcommands that mention other commands or the launcher.
#[test]
fn subcommand_help_says_moai() {
    for word in [
        "session",
        "key-set",
        "project-plan",
        "project-detect",
        "agents-merge-plan",
    ] {
        assert_no_command_moai_cli(word, &[word, "--help"]);
    }
}
