//! SPEC-20261120-secret-token-shapes (slice S1 of the Rust security hooks): tool-preflight A4
//! refuses six more access-token shapes in written content — an OpenAI-style key, a GitHub OAuth
//! token, a GitLab token, a Slack token, an AWS access key id (long-term and temporary) and a
//! Google OAuth token. Every case runs the real binary in a temporary folder with a temporary
//! HOME (as `tests/spec_before_code_hook.rs`); a real host session is not run here. Every value
//! is assembled at run time from two halves, so no line of this file holds a whole token.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::path::Path;
use std::process::{Command, Stdio};
use std::time::Instant;

const BIN: &str = crate::common::MOAI_BIN;
/// REQ-9: the one fixed refusal text; it names private keys and access tokens.
const FIXED_REASON: &str =
    "MoAI tool-preflight: writing a private key or an access token into a file is refused.";

struct Run {
    code: i32,
    stdout: String,
}

fn dispatch(host: &str, cwd: &Path, tool: &str, tool_input: Value) -> Run {
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
    r.code == 0 && r.stdout.contains("\"permissionDecision\":\"deny\"")
}

fn allowed(r: &Run) -> bool {
    r.code == 0 && r.stdout.is_empty()
}

/// The four ways written content reaches tool-preflight: the four Claude write tools.
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

/// One token shape of SPEC-20261120-secret-token-shapes REQ-1..REQ-6.
#[derive(Clone, Copy)]
struct Shape {
    name: &'static str,
    /// The prefix in two halves, joined at run time.
    head: (&'static str, &'static str),
    /// The shortest value that counts: this many characters after the prefix.
    min: usize,
    /// A character every position after the prefix accepts.
    fill: char,
    /// The prefix counts only at a word start.
    word_start: bool,
}

const SHAPES: [Shape; 7] = [
    Shape {
        name: "openai-style",
        head: ("s", "k-"),
        min: 32,
        fill: 'a',
        word_start: true,
    },
    Shape {
        name: "github-oauth",
        head: ("gh", "o_"),
        min: 36,
        fill: 'a',
        word_start: false,
    },
    Shape {
        name: "gitlab",
        head: ("glp", "at-"),
        min: 20,
        fill: 'a',
        word_start: false,
    },
    Shape {
        name: "slack",
        head: ("xo", "xb-"),
        min: 10,
        fill: 'a',
        word_start: true,
    },
    Shape {
        name: "aws-long-term",
        head: ("AK", "IA"),
        min: 16,
        fill: 'A',
        word_start: true,
    },
    Shape {
        name: "aws-temporary",
        head: ("AS", "IA"),
        min: 16,
        fill: 'A',
        word_start: true,
    },
    Shape {
        name: "google-oauth",
        head: ("ya", "29."),
        min: 20,
        fill: 'a',
        word_start: true,
    },
];

impl Shape {
    fn prefix(&self) -> String {
        format!("{}{}", self.head.0, self.head.1)
    }

    fn value(&self, n: usize) -> String {
        format!("{}{}", self.prefix(), self.fill.to_string().repeat(n))
    }

    fn at_min(&self) -> String {
        self.value(self.min)
    }

    fn one_short(&self) -> String {
        self.value(self.min - 1)
    }
}

/// A realistic one-line context: the value sits after a quote, so it is at a word start.
fn in_line(value: &str) -> String {
    format!("export KEY=\"{value}\"")
}

// ---- AC-1 (REQ-1..REQ-6) ----

/// AC-1: each shape at exactly its minimum is refused through every write tool.
#[test]
fn refused_every_shape_at_its_minimum_through_every_write_tool() {
    let dir = tempfile::TempDir::new().unwrap();
    for shape in SHAPES {
        for (tool, run) in every_write(dir.path(), &in_line(&shape.at_min())) {
            assert!(
                denied(&run),
                "{} through {tool}: code {} stdout {:?}",
                shape.name,
                run.code,
                run.stdout
            );
        }
    }
}

/// AC-1 (REQ-4): the Slack token takes each of its five letters.
#[test]
fn refused_slack_token_with_each_of_its_five_letters() {
    let dir = tempfile::TempDir::new().unwrap();
    for letter in ['b', 'a', 'p', 'r', 's'] {
        let value = format!("{}{}{letter}-{}", "xo", "x", "a".repeat(10));
        let run = write_only(dir.path(), &in_line(&value));
        assert!(denied(&run), "letter {letter}: {:?}", run.stdout);
    }
    let other = format!("{}{}{}-{}", "xo", "x", "z", "a".repeat(10));
    assert!(allowed(&write_only(dir.path(), &in_line(&other))));
}

// ---- AC-2 (REQ-7) ----

/// AC-2: one character short of the minimum is allowed through every write tool.
#[test]
fn boundary_one_character_short_is_allowed_through_every_write_tool() {
    let dir = tempfile::TempDir::new().unwrap();
    for shape in SHAPES {
        for (tool, run) in every_write(dir.path(), &in_line(&shape.one_short())) {
            assert!(
                allowed(&run),
                "{} one short through {tool}: code {} stdout {:?}",
                shape.name,
                run.code,
                run.stdout
            );
        }
    }
}

/// AC-2: a prefix named in prose, and ordinary hyphenated words, are allowed.
#[test]
fn boundary_prefix_in_prose_and_ordinary_words_are_allowed() {
    let dir = tempfile::TempDir::new().unwrap();
    for shape in SHAPES {
        let prose = format!("the prefix {} is documented in the guide", shape.prefix());
        assert!(
            allowed(&write_only(dir.path(), &prose)),
            "{} in prose",
            shape.name
        );
    }
    for text in [
        "a task-based plan with a risk-free rollout",
        "the desk-top task-based flow, risk-free",
        "ask-me-anything sessions and mask-wearing rules",
    ] {
        assert!(allowed(&write_only(dir.path(), text)), "{text}");
    }
}

/// AC-2 (REQ-7): the word-start rule looks at ASCII letters and digits only, so a value that
/// follows a non-ASCII letter or an emoji is still refused, for every shape.
#[test]
fn boundary_non_ascii_character_before_the_prefix_is_a_word_start() {
    let dir = tempfile::TempDir::new().unwrap();
    for shape in SHAPES {
        for lead in ["é", "한", "😀"] {
            let run = write_only(dir.path(), &format!("{lead}{}", shape.at_min()));
            assert!(
                denied(&run),
                "{} after {lead}: {:?}",
                shape.name,
                run.stdout
            );
        }
    }
}

/// AC-2: a word-start shape does not count inside a longer word; the two shapes without that
/// rule still do.
#[test]
fn boundary_inner_prefix_counts_only_for_the_shapes_without_a_word_start_rule() {
    let dir = tempfile::TempDir::new().unwrap();
    for shape in SHAPES {
        let run = write_only(dir.path(), &format!("q{}", shape.at_min()));
        if shape.word_start {
            assert!(
                allowed(&run),
                "{} inside a word: {:?}",
                shape.name,
                run.stdout
            );
        } else {
            assert!(
                denied(&run),
                "{} inside a word: {:?}",
                shape.name,
                run.stdout
            );
        }
    }
}

/// AC-2: the shapes are case-sensitive; the other-case prefix is allowed.
#[test]
fn boundary_other_case_prefix_is_allowed() {
    let dir = tempfile::TempDir::new().unwrap();
    for shape in SHAPES {
        let prefix = shape.prefix();
        let swapped = if prefix.chars().any(|c| c.is_ascii_lowercase()) {
            prefix.to_ascii_uppercase()
        } else {
            prefix.to_ascii_lowercase()
        };
        let value = format!("{swapped}{}", shape.fill.to_string().repeat(shape.min));
        assert!(
            allowed(&write_only(dir.path(), &in_line(&value))),
            "{} other case",
            shape.name
        );
    }
}

/// AC-2: only the characters a shape accepts count toward its minimum.
#[test]
fn boundary_characters_a_shape_does_not_accept_end_the_run() {
    let dir = tempfile::TempDir::new().unwrap();
    let openai = &SHAPES[0];
    let broken = format!("{}-a", openai.value(openai.min - 1));
    assert!(
        allowed(&write_only(dir.path(), &in_line(&broken))),
        "hyphen ends an OpenAI-style run"
    );
    let oauth = &SHAPES[1];
    let broken = format!("{}_a", oauth.value(oauth.min - 1));
    assert!(
        allowed(&write_only(dir.path(), &in_line(&broken))),
        "underscore ends an OAuth run"
    );
    let aws = &SHAPES[4];
    let broken = format!("{}a", aws.value(aws.min - 1));
    assert!(
        allowed(&write_only(dir.path(), &in_line(&broken))),
        "lower case ends an AWS run"
    );
    // The shapes that accept hyphens or underscores count them.
    let gitlab = format!("{}{}", SHAPES[2].prefix(), "a-a_".repeat(5));
    assert!(
        denied(&write_only(dir.path(), &in_line(&gitlab))),
        "20 mixed GitLab characters"
    );
    let slack = format!("{}{}", SHAPES[3].prefix(), "a-a-a-a-a-");
    assert!(
        denied(&write_only(dir.path(), &in_line(&slack))),
        "10 Slack characters with hyphens"
    );
    let google = format!("{}{}", SHAPES[6].prefix(), "a_a-".repeat(5));
    assert!(
        denied(&write_only(dir.path(), &in_line(&google))),
        "20 mixed Google characters"
    );
    // Every one of the 13 prefixes, with a character its shape does not accept in place of the
    // last required one, leaves the run one short: allowed. The same text with an accepted
    // character there is refused, so the one changed character is what decides (first-slice audit:
    // widening the accepted set of Slack or of the GitHub classic prefix must fail a test).
    let cases: Vec<(String, usize, char, Vec<char>)> = [
        (["gh", "p_"].concat(), 36, 'a', vec!['_', '-', '.']),
        (["github", "_pat_"].concat(), 22, 'a', vec!['-', '.']),
        (["gh", "o_"].concat(), 36, 'a', vec!['_', '-', '.']),
        (["glp", "at-"].concat(), 20, 'a', vec!['.']),
        (["s", "k-"].concat(), 32, 'a', vec!['_', '-', '.']),
        (["AK", "IA"].concat(), 16, 'A', vec!['-', '_', 'a']),
        (["AS", "IA"].concat(), 16, 'A', vec!['-', '_', 'a']),
        (["ya", "29."].concat(), 20, 'a', vec!['.']),
    ]
    .into_iter()
    .chain(
        ["b", "a", "p", "r", "s"]
            .map(|letter| (["xo", "x", letter, "-"].concat(), 10, 'a', vec!['_', '.'])),
    )
    .collect();
    assert_eq!(cases.len(), 13, "one case per prefix of the table");
    for (prefix, min, fill, refused) in cases {
        let accepted = format!("{prefix}{}", fill.to_string().repeat(min));
        assert!(
            denied(&write_only(dir.path(), &in_line(&accepted))),
            "{prefix}: {min} accepted characters"
        );
        for bad in refused {
            let broken = format!("{prefix}{}{bad}", fill.to_string().repeat(min - 1));
            assert!(
                allowed(&write_only(dir.path(), &in_line(&broken))),
                "{prefix}: {bad:?} ends the run one short"
            );
        }
    }
}

// ---- AC-3 (REQ-8, REQ-10) ----

/// AC-3: certificates and public keys stay allowed and the existing refusals answer as before.
#[test]
fn unchanged_existing_answers() {
    let dir = tempfile::TempDir::new().unwrap();
    let cwd = dir.path();
    let block = |kind: &str| format!("{}{} {kind}{}", "-----", "BEGIN", "-----");
    assert!(
        allowed(&write_only(cwd, &block("CERTIFICATE"))),
        "certificate"
    );
    assert!(
        allowed(&write_only(cwd, &block("PUBLIC KEY"))),
        "public key"
    );
    let private = format!("{}{}", "RSA PRIVATE", " KEY");
    assert!(
        denied(&write_only(cwd, &block(&private))),
        "private key block"
    );
    let ghp = |n: usize| format!("gh{}{}", "p_", "a".repeat(n));
    assert!(
        denied(&write_only(cwd, &in_line(&ghp(36)))),
        "36-character GitHub token"
    );
    assert!(
        allowed(&write_only(cwd, &in_line(&ghp(35)))),
        "35-character GitHub token"
    );
    assert!(
        allowed(&write_only(cwd, &format!("use the gh{} prefix", "p_"))),
        "prose"
    );
    let pat = |n: usize| format!("{}{}{}", "github", "_pat_", "a".repeat(n));
    assert!(
        denied(&write_only(cwd, &in_line(&pat(22)))),
        "22-character fine-grained token"
    );
    assert!(
        allowed(&write_only(cwd, &in_line(&pat(21)))),
        "21-character fine-grained token"
    );
}

// ---- AC-4 (REQ-9) ----

/// AC-4: every refusal is the documented deny object with the fixed reason, and the output holds
/// none of the matched value.
#[test]
fn reason_is_the_fixed_text_and_never_echoes_the_value() {
    let dir = tempfile::TempDir::new().unwrap();
    let mut values: Vec<(String, String)> = SHAPES
        .iter()
        .map(|s| (s.name.to_string(), s.at_min()))
        .collect();
    values.push((
        "github-token".into(),
        format!("gh{}{}", "p_", "a".repeat(36)),
    ));
    for (name, value) in values {
        let line = in_line(&value);
        for (tool, run) in every_write(dir.path(), &line) {
            assert!(denied(&run), "{name} through {tool}: {:?}", run.stdout);
            let v: Value = serde_json::from_str(run.stdout.trim()).unwrap();
            let out = &v["hookSpecificOutput"];
            assert_eq!(out["hookEventName"], json!("PreToolUse"), "{name} {tool}");
            assert_eq!(out["permissionDecision"], json!("deny"), "{name} {tool}");
            assert_eq!(
                out["permissionDecisionReason"],
                json!(FIXED_REASON),
                "{name} {tool}"
            );
            assert!(
                !run.stdout.contains(&value),
                "{name} {tool} echoes the value"
            );
            assert!(
                !run.stdout.contains("out.txt"),
                "{name} {tool} echoes the path"
            );
        }
    }
}

// ---- AC-5 (REQ-11) ----

/// AC-5: a long content of near-misses, and a long run behind a real prefix, are judged in a
/// time that does not grow with the run (the time is printed, only a generous bound is asserted).
#[test]
fn linear_long_content_is_judged_in_proportion_to_its_length() {
    let dir = tempfile::TempDir::new().unwrap();
    let cwd = dir.path();
    const BOUND_SECS: f64 = 30.0;
    for shape in SHAPES {
        // Near-misses, one per line, about 3 MiB in all: allowed.
        let unit = format!("{}\n", shape.one_short());
        let content = unit.repeat(3 * 1024 * 1024 / unit.len());
        let start = Instant::now();
        let run = write_only(cwd, &content);
        let secs = start.elapsed().as_secs_f64();
        eprintln!(
            "linear {} near-misses {} KiB: {secs:.3} s",
            shape.name,
            content.len() / 1024
        );
        assert!(
            allowed(&run),
            "{} near-misses: {:?}",
            shape.name,
            run.stdout
        );
        assert!(secs < BOUND_SECS, "{} near-misses took {secs}", shape.name);

        // One real prefix followed by a 3 MiB run of accepted characters: refused.
        let content = shape.value(3 * 1024 * 1024);
        let start = Instant::now();
        let run = write_only(cwd, &content);
        let secs = start.elapsed().as_secs_f64();
        eprintln!(
            "linear {} long run {} KiB: {secs:.3} s",
            shape.name,
            content.len() / 1024
        );
        assert!(denied(&run), "{} long run: {:?}", shape.name, run.stdout);
        assert!(secs < BOUND_SECS, "{} long run took {secs}", shape.name);
    }
    let plain = "a".repeat(1024 * 1024);
    let start = Instant::now();
    let run = write_only(cwd, &plain);
    eprintln!("linear plain 1 MiB: {:.3} s", start.elapsed().as_secs_f64());
    assert!(allowed(&run));
}
