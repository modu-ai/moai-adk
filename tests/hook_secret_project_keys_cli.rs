//! SPEC-20261120-secret-project-keys: tool-preflight A4 refuses four more access-token prefixes in
//! written content — the project, service-account and admin keys of OpenAI and the keys of
//! Anthropic — when at least 40 letters, digits, hyphens or underscores follow the prefix at a word
//! start. Every case runs the real binary in a temporary folder with a temporary HOME (as
//! `tests/hook_secret_tokens_cli.rs`); a real host session is not run here. Every prefix is
//! assembled at run time from two halves, so no line of this file holds a whole prefix.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::path::Path;
use std::process::{Command, Stdio};
use std::time::Instant;

const BIN: &str = crate::common::MOAI_BIN;
/// REQ-3: the one fixed refusal text of the access-token rule.
const FIXED_REASON: &str =
    "MoAI tool-preflight: writing a private key or an access token into a file is refused.";

struct Run {
    code: i32,
    stdout: String,
}

fn dispatch(host: &str, cwd: &Path, tool: &str, tool_input: Value) -> Run {
    // A moai project: outside one, every duty exits 0 at once
    // (SPEC-20261122-next-flow-hooks REQ-15).
    std::fs::create_dir_all(cwd.join(".moai")).unwrap();
    let input = json!({"hook_event_name": "PreToolUse", "tool_name": tool,
                       "tool_input": tool_input, "cwd": cwd});
    let mut child = Command::new(BIN)
        .args(["hook-dispatch", "--host", host, "--event", "tool-preflight"])
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    Run {
        code: out.status.code().unwrap(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
    }
}

fn denied(r: &Run) -> bool {
    // A refusal is exit code 2 with the documented deny object beside it (REQ-8).
    matches!(r.code, 0 | 2) && r.stdout.contains("\"permissionDecision\":\"deny\"")
}

fn allowed(r: &Run) -> bool {
    r.code == 0 && r.stdout.is_empty()
}

fn patch_adding(path: &str, line: &str) -> Value {
    json!({"command": format!("*** Begin Patch\n*** Add File: {path}\n+{line}\n*** End Patch")})
}

/// The five ways written content reaches tool-preflight: four Claude tools and one Codex patch.
/// `content` must be one line (a patch line starts with `+`).
fn every_write(cwd: &Path, content: &str) -> Vec<(&'static str, Run)> {
    let path = cwd.join("out.txt");
    let path = path.to_str().unwrap();
    vec![
        (
            "Write",
            dispatch(
                "claude",
                cwd,
                "Write",
                json!({"file_path": path, "content": content}),
            ),
        ),
        (
            "Edit",
            dispatch(
                "claude",
                cwd,
                "Edit",
                json!({"file_path": path, "old_string": "x", "new_string": content}),
            ),
        ),
        (
            "MultiEdit",
            dispatch(
                "claude",
                cwd,
                "MultiEdit",
                json!({"file_path": path, "edits": [{"old_string": "x", "new_string": content}]}),
            ),
        ),
        (
            "NotebookEdit",
            dispatch(
                "claude",
                cwd,
                "NotebookEdit",
                json!({"notebook_path": path, "new_source": content}),
            ),
        ),
        (
            "apply_patch",
            dispatch("codex", cwd, "apply_patch", patch_adding(path, content)),
        ),
    ]
}

fn write_only(cwd: &Path, content: &str) -> Run {
    let path = cwd.join("out.txt");
    dispatch(
        "claude",
        cwd,
        "Write",
        json!({"file_path": path.to_str().unwrap(), "content": content}),
    )
}

/// The four prefixes of the SPEC table, each in two halves joined at run time.
const PREFIXES: [(&str, &str); 4] = [
    ("s", "k-proj-"),
    ("s", "k-svcacct-"),
    ("s", "k-admin-"),
    ("s", "k-ant-"),
];
/// REQ-1: the shortest run after the prefix that counts.
const MIN: usize = 40;

fn prefix((a, b): (&str, &str)) -> String {
    format!("{a}{b}")
}

fn value(p: (&str, &str), n: usize) -> String {
    format!("{}{}", prefix(p), "a".repeat(n))
}

/// A realistic one-line context: the value sits after a quote, so it is at a word start.
fn in_line(value: &str) -> String {
    format!("export KEY=\"{value}\"")
}

// ---- AC-1 (REQ-1, REQ-3) ----

/// AC-1: each prefix with exactly 40 accepted characters is refused through every write tool and
/// the Codex patch.
#[test]
fn refused_each_prefix_at_exactly_forty_through_every_write_tool() {
    let dir = tempfile::TempDir::new().unwrap();
    for p in PREFIXES {
        for (tool, run) in every_write(dir.path(), &in_line(&value(p, MIN))) {
            assert!(
                denied(&run),
                "{} through {tool}: code {} stdout {:?}",
                prefix(p),
                run.code,
                run.stdout
            );
        }
    }
}

/// AC-1: hyphens and underscores inside the run count, a realistic Anthropic and a realistic
/// OpenAI project key shape are refused, and the fixed text never repeats the value.
#[test]
fn refused_realistic_shapes_and_the_text_never_holds_the_value() {
    let dir = tempfile::TempDir::new().unwrap();
    let anthropic = format!("{}{}{}{}", "s", "k-ant-api03-", "Ab9_-".repeat(18), "AA");
    let openai = format!(
        "{}{}{}{}{}",
        "s",
        "k-proj-",
        "Qx7-".repeat(8),
        ["T3Blbk", "FJ"].concat(),
        "_n3-Zk".repeat(6)
    );
    for text in [&anthropic, &openai] {
        let run = write_only(dir.path(), &in_line(text));
        assert!(denied(&run), "{text}: {:?}", run.stdout);
        assert!(run.stdout.contains(FIXED_REASON), "{:?}", run.stdout);
        assert!(
            !run.stdout.contains(text.as_str()) && !run.stdout.contains(&text[..16]),
            "the refusal repeats the value: {:?}",
            run.stdout
        );
    }
}

// ---- AC-2 (REQ-2) ----

/// AC-2: 39 characters after the prefix are allowed through every write tool.
#[test]
fn boundary_thirty_nine_is_allowed_through_every_write_tool() {
    let dir = tempfile::TempDir::new().unwrap();
    for p in PREFIXES {
        for (tool, run) in every_write(dir.path(), &in_line(&value(p, MIN - 1))) {
            assert!(
                allowed(&run),
                "{} one short through {tool}: code {} stdout {:?}",
                prefix(p),
                run.code,
                run.stdout
            );
        }
    }
}

/// AC-2: a character outside the accepted set ends the run: a period or a space inside it leaves
/// both halves short.
#[test]
fn boundary_a_character_outside_the_set_ends_the_run() {
    let dir = tempfile::TempDir::new().unwrap();
    for p in PREFIXES {
        for end in ['.', ' ', '+', '/'] {
            let text = format!("{}{end}{}", value(p, MIN - 1), "a".repeat(10));
            assert!(
                allowed(&write_only(dir.path(), &in_line(&text))),
                "{} ended by {end:?}",
                prefix(p)
            );
        }
        // Hyphens and underscores count: 39 plain characters and one of each in the run.
        for inside in ['-', '_'] {
            let text = format!("{}{inside}", value(p, MIN - 1));
            assert!(
                denied(&write_only(dir.path(), &in_line(&text))),
                "{} with {inside:?} as the 40th character",
                prefix(p)
            );
        }
    }
}

/// AC-2: the prefix counts only at a word start (nothing, or a character that is no ASCII letter
/// or digit, before it) and only in its own letter case.
#[test]
fn boundary_word_start_and_letter_case() {
    let dir = tempfile::TempDir::new().unwrap();
    for p in PREFIXES {
        let v = value(p, MIN);
        for lead in ["q", "7", "Z"] {
            assert!(
                allowed(&write_only(dir.path(), &format!("{lead}{v}"))),
                "{} after {lead:?}",
                prefix(p)
            );
        }
        for lead in ["_", "-", "\"", " ", "=", "é", "한", "😀"] {
            assert!(
                denied(&write_only(dir.path(), &format!("{lead}{v}"))),
                "{} after {lead:?}",
                prefix(p)
            );
        }
        let upper = format!("{}{}", prefix(p).to_ascii_uppercase(), "a".repeat(MIN));
        assert!(
            allowed(&write_only(dir.path(), &in_line(&upper))),
            "{} upper case",
            prefix(p)
        );
    }
}

/// AC-2: a prefix named in prose, short hyphenated names and ordinary kebab-case words are
/// allowed; only the four prefixes count, not every name after the first-slice prefix.
#[test]
fn boundary_prose_short_names_and_other_names_are_allowed() {
    let dir = tempfile::TempDir::new().unwrap();
    for p in PREFIXES {
        let prose = format!("the prefix {} is documented in the guide", prefix(p));
        assert!(allowed(&write_only(dir.path(), &prose)), "{}", prefix(p));
        let name = format!("{}config-name-for-the-widget", prefix(p));
        assert!(allowed(&write_only(dir.path(), &name)), "{name}");
    }
    let other = format!("{}{}{}", "s", "k-other-", "a".repeat(60));
    assert!(allowed(&write_only(dir.path(), &in_line(&other))));
    let kebab = format!("a task-{}", "very-long-kebab-case-word-".repeat(4));
    assert!(allowed(&write_only(dir.path(), &kebab)));
}

// ---- AC-3 (REQ-4) ----

/// AC-3: the first-slice shape is neither widened nor narrowed: 32 letters and digits after the
/// short prefix are still refused, one hyphen in the 32 still ends that run.
#[test]
fn unchanged_first_slice_openai_shape_keeps_its_answers() {
    let dir = tempfile::TempDir::new().unwrap();
    let short = format!("{}{}{}", "s", "k-", "a".repeat(32));
    assert!(denied(&write_only(dir.path(), &in_line(&short))));
    let broken = format!("{}{}{}-a", "s", "k-", "a".repeat(31));
    assert!(allowed(&write_only(dir.path(), &in_line(&broken))));
}

// ---- AC-4 (REQ-5) ----

/// AC-4: 3 MiB of near misses for each prefix is judged within a generous bound, with the right
/// answer; the time is printed, not asserted tightly.
#[test]
fn linear_judging_time_on_three_mib_of_near_misses() {
    let dir = tempfile::TempDir::new().unwrap();
    let mib = 1024 * 1024;
    for p in PREFIXES {
        let unit = format!("{}.", value(p, MIN - 1));
        let body = unit.repeat(3 * mib / unit.len());
        let start = Instant::now();
        let run = write_only(dir.path(), &body);
        let took = start.elapsed();
        eprintln!("linear {}: {took:?}", prefix(p));
        assert!(allowed(&run), "{}: near misses are allowed", prefix(p));
        assert!(took.as_secs() < 20, "{}: {took:?}", prefix(p));
        // A genuine value at the very end is still found.
        let ending = format!("{body} {}", value(p, MIN));
        assert!(denied(&write_only(dir.path(), &ending)), "{}", prefix(p));
    }
}
