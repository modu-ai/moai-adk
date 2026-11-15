//! What the launcher (`moai cc|glm|codex -l|-a`) prints around a session (#3004 P9).
//!
//! After the host exits the launcher used to print one JSON line (`executed:true`), which a
//! person at a terminal should never see. The JSON envelope stays for a pipe, a file and
//! `--json`; on a terminal the launcher prints plain text instead, the way `moai update` does
//! (#2894): nothing after a normal exit, a line or two on stderr after a failed one, and a short
//! readable summary for `--dry-run`. Notes that carry a warning are printed once, before the
//! host starts. Nothing here ever prints a key or token value: the envelope itself holds
//! variable names only.
use crate::session::{
    NOTE_ADVISOR_NOT_PASSED_CODEX, NOTE_ADVISOR_OMITTED_GLM, NOTE_EFFORT_NOT_PASSED_CODEX,
};
use serde_json::Value;

/// Whether the launcher prints the JSON envelope: always with `--json`, and whenever stdout is
/// not a terminal, so every pipe and file reader sees the answer it saw before.
pub fn wants_envelope(json_flag: bool, stdout_is_tty: bool) -> bool {
    json_flag || !stdout_is_tty
}

/// The terminal rule for this process: plain text, no envelope.
pub fn wants_human(json_flag: bool) -> bool {
    use std::io::IsTerminal;
    !wants_envelope(json_flag, std::io::stdout().is_terminal())
}

/// Lines printed on stderr before the host starts: the warnings among the composed notes and
/// the agent-number notes (`number_notes`). Informational notes stay out.
pub fn startup_lines(composed: &Value, number_notes: &[String]) -> Vec<String> {
    // Decided in #3004 P9: the notes that say an option was dropped are limits the user must
    // see; the other notes (role defaults, model mapping, permission mode, key source) only
    // describe what the launcher chose and stay in the envelope.
    const WARNINGS: [&str; 3] = [
        NOTE_ADVISOR_OMITTED_GLM,
        NOTE_EFFORT_NOT_PASSED_CODEX,
        NOTE_ADVISOR_NOT_PASSED_CODEX,
    ];
    let mut lines: Vec<String> = number_notes.iter().map(|n| format!("moai: {n}")).collect();
    let notes = composed["notes"].as_array().into_iter().flatten();
    lines.extend(
        notes
            .filter_map(Value::as_str)
            .filter(|n| WARNINGS.contains(n))
            .map(|n| format!("moai: {n}")),
    );
    lines
}

/// The launcher word for a host, for the line that tells the user what to run next.
fn host_word(host: &str) -> &str {
    match host {
        "claude" => "cc",
        other => other,
    }
}

/// The text printed on stderr after the host ended, or `None` after a normal exit.
pub fn after_text(v: &Value) -> Option<String> {
    let host = v["host"].as_str().unwrap_or("the host");
    let what = match v["exit_code"].as_i64() {
        Some(0) => return None,
        Some(code) => format!("moai: {host} exited with code {code}."),
        None => format!("moai: {host} was stopped by a signal (no exit code)."),
    };
    let role = match v["role"].as_str() {
        Some("leader") => " -l",
        Some("agent") => " -a",
        _ => "",
    };
    Some(format!(
        "{what}\nRun `moai {}{role} --dry-run` to see the command it was started with.",
        host_word(v["host"].as_str().unwrap_or("cc"))
    ))
}

const WIDTH: usize = 80;

/// One shell word: left as it is when it needs no quoting, else in single quotes.
fn shell_word(w: &str) -> String {
    let plain = !w.is_empty()
        && w.chars()
            .all(|c| c.is_ascii_alphanumeric() || "_-./=:,@%+".contains(c));
    if plain {
        w.to_string()
    } else {
        format!("'{}'", w.replace('\'', "'\\''"))
    }
}

/// Greedy word wrap to `WIDTH` columns. `first` starts the first line and `rest` every later
/// one; `eol` ends every line but the last (a ` \` keeps a command line readable by the shell).
fn wrap(words: &[String], first: &str, rest: &str, eol: &str) -> String {
    let mut out = String::new();
    let mut line = first.to_string();
    let mut fresh = true;
    for w in words {
        let width = |s: &str| s.chars().count();
        let need = width(&line) + width(w) + width(eol) + usize::from(!fresh);
        if !fresh && need > WIDTH {
            out.push_str(&line);
            out.push_str(eol);
            out.push('\n');
            line = rest.to_string();
            fresh = true;
        }
        if !fresh {
            line.push(' ');
        }
        line.push_str(w);
        fresh = false;
    }
    out.push_str(&line);
    out.push('\n');
    out
}

/// The readable summary of a launch that was not executed (`--dry-run`), for stdout: the host,
/// the role and name, the command line, the environment variable names (never values), and the
/// notes.
pub fn preview_text(v: &Value) -> String {
    let text = |key: &str| v[key].as_str().unwrap_or("(none)").to_string();
    let words = |key: &str| -> Vec<String> {
        v[key]
            .as_array()
            .into_iter()
            .flatten()
            .filter_map(Value::as_str)
            .map(str::to_string)
            .collect()
    };
    let mut out = String::from("Dry run: nothing was started.\n");
    for (label, value) in [
        ("host", text("host")),
        ("role", text("role")),
        ("name", text("name")),
    ] {
        out.push_str(&format!("  {label:<8} {value}\n"));
    }
    let command: Vec<String> = words("argv").iter().map(|w| shell_word(w)).collect();
    out.push_str(&wrap(&command, "  command  ", "             ", " \\"));
    let names = words("env_keys");
    if !names.is_empty() {
        let last = names.len() - 1;
        let listed: Vec<String> = names
            .iter()
            .enumerate()
            .map(|(i, n)| {
                if i == last {
                    n.clone()
                } else {
                    format!("{n},")
                }
            })
            .collect();
        out.push_str("  env      (names only, no values)\n");
        out.push_str(&wrap(&listed, "           ", "           ", ""));
    }
    if let Some(source) = v["key_source"].as_str() {
        out.push_str(&format!("  key      read from {source} (never shown)\n"));
    }
    let notes = words("notes");
    if !notes.is_empty() {
        out.push_str("Notes:\n");
    }
    for note in &notes {
        let parts: Vec<String> = note.split_whitespace().map(str::to_string).collect();
        out.push_str(&wrap(&parts, "  - ", "    ", ""));
    }
    out
}

/// Prints the result of a launch: the envelope, or the plain text for a terminal.
pub fn report(v: &Value, human: bool) {
    if !human {
        println!("{v}");
        return;
    }
    if v["executed"] == Value::Bool(true) {
        if let Some(text) = after_text(v) {
            eprintln!("{text}");
        }
    } else {
        print!("{}", preview_text(v));
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::session::{
        AGENT_ONE_WARNING, NOTE_ADVISOR_NOT_PASSED_CODEX, NOTE_ADVISOR_OMITTED_GLM,
        NOTE_EFFORT_NOT_PASSED_CODEX,
    };
    use serde_json::json;

    /// The envelope of a session that ran, as `launch` returns it.
    fn ran(host: &str, role: Value, exit_code: Value) -> Value {
        json!({"host": host, "role": role, "name": "agent-21",
            "argv": ["claude", "--name", "agent-21", "--model", "opus"],
            "env_keys": [], "notes": ["agent number 21 taken: next after live agent-20"],
            "host_loaded": false, "executed": true, "exit_code": exit_code})
    }

    // #3004 P9: the decision, in all four directions. Linux CI runs this one; the pseudo
    // terminal tests are macOS only.
    #[test]
    fn the_envelope_goes_to_a_pipe_or_json_and_never_to_a_bare_terminal() {
        assert!(wants_envelope(false, false), "a pipe or a file");
        assert!(wants_envelope(true, false), "--json on a pipe");
        assert!(wants_envelope(true, true), "--json on a terminal");
        assert!(!wants_envelope(false, true), "a terminal without --json");
    }

    #[test]
    fn a_normal_exit_prints_nothing() {
        assert_eq!(after_text(&ran("claude", json!("agent"), json!(0))), None);
    }

    #[test]
    fn a_failed_exit_says_what_failed_and_what_to_run_next() {
        for (host, role, word, flag) in [
            ("claude", json!("agent"), "cc", " -a"),
            ("claude", json!("leader"), "cc", " -l"),
            ("glm", json!("agent"), "glm", " -a"),
            ("codex", Value::Null, "codex", ""),
        ] {
            let text = after_text(&ran(host, role, json!(3))).expect("a failure prints");
            assert!(
                text.contains(&format!("{host} exited with code 3")),
                "{text}"
            );
            assert!(
                text.contains(&format!("`moai {word}{flag} --dry-run`")),
                "{text}"
            );
            assert!(text.lines().count() <= 2, "{text}");
            assert!(!text.contains('{'), "{text}");
        }
    }

    #[test]
    fn a_host_stopped_by_a_signal_has_no_code_but_still_prints() {
        let text = after_text(&ran("claude", json!("agent"), Value::Null)).unwrap();
        assert!(text.contains("claude was stopped by a signal"), "{text}");
        assert!(text.contains("moai cc -a --dry-run"), "{text}");
    }

    fn dry_run() -> Value {
        json!({"host": "glm", "role": "agent", "name": "agent-1",
            "argv": ["claude", "--name", "agent-1", "--model", "opus", "do it; rm 'x'"],
            "env_keys": ["ANTHROPIC_BASE_URL", "ANTHROPIC_AUTH_TOKEN"],
            "notes": ["first note", "warning: a claim of an earlier agent-1"],
            "host_loaded": false, "executed": false, "key_source": "vault:shared:glm"})
    }

    #[test]
    fn the_preview_is_readable_text_with_names_only() {
        let text = preview_text(&dry_run());
        assert!(!text.contains('{') && !text.contains('"'), "{text}");
        for want in [
            "nothing was started",
            "glm",
            "agent-1",
            "claude --name agent-1 --model opus 'do it; rm '\\''x'\\'''",
            "ANTHROPIC_BASE_URL, ANTHROPIC_AUTH_TOKEN",
            "vault:shared:glm",
            "first note",
            "warning: a claim of an earlier agent-1",
        ] {
            assert!(text.contains(want), "{want}\n{text}");
        }
        assert!(text.lines().all(|l| l.chars().count() <= 80), "{text}");
    }

    #[test]
    fn a_long_command_wraps_with_a_continuation_the_shell_still_reads() {
        let mut v = dry_run();
        v["argv"] = json!([
            "claude",
            "--name",
            "agent-1",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--dangerously-skip-permissions",
            "--advisor",
            "fable",
            "a-prompt-word"
        ]);
        let text = preview_text(&v);
        assert!(text.lines().any(|l| l.ends_with(" \\")), "{text}");
        assert!(text.lines().all(|l| l.chars().count() <= 80), "{text}");
    }

    #[test]
    fn only_warnings_and_the_number_are_printed_before_the_host_starts() {
        let composed = json!({"notes": [
            "no role: a plain host session; no role or session name is added (#2053)",
            "model aliases map to z.ai models for this session: opus to glm-5.3",
            NOTE_ADVISOR_OMITTED_GLM,
            NOTE_EFFORT_NOT_PASSED_CODEX,
            NOTE_ADVISOR_NOT_PASSED_CODEX,
            "codex sessions have no cross-session channel in RC1 (no ListAgents/SendMessage)",
        ]});
        let number = vec![
            "agent number 21 taken: next after live agent-20 (store /x)".to_string(),
            AGENT_ONE_WARNING.to_string(),
        ];
        let lines = startup_lines(&composed, &number);
        assert_eq!(
            lines,
            [
                "moai: agent number 21 taken: next after live agent-20 (store /x)".to_string(),
                format!("moai: {AGENT_ONE_WARNING}"),
                format!("moai: {NOTE_ADVISOR_OMITTED_GLM}"),
                format!("moai: {NOTE_EFFORT_NOT_PASSED_CODEX}"),
                format!("moai: {NOTE_ADVISOR_NOT_PASSED_CODEX}"),
            ]
        );
        assert!(startup_lines(&json!({"notes": []}), &[]).is_empty());
    }
}
