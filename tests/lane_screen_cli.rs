//! SPEC-20261119-lane-screen-guard (card C-120): `moai lane-screen` reads a lane's rendered
//! screen into a state, the one row it was read from, and for a usage limit the reset time in
//! UTC. The screens are `tests/fixtures/lane_screens/synthetic-*.txt` (hand-written in the shape
//! of a lane screen; see the README there) and Orca-shaped JSON built here. Every case runs the
//! real binary in a temporary folder with a temporary HOME; nothing is sent to any terminal.
mod common;

use moai_cli::usage_limit::{record, Window};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
/// 2026-11-19T00:00:00Z.
const NOW: i64 = 1_795_046_400;
const NOW_TEXT: &str = "2026-11-19T00:00:00Z";

fn fixture(name: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures/lane_screens")
        .join(name);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
}

struct Run {
    code: i32,
    envelope: Value,
    stderr: String,
}

impl Run {
    fn result(&self) -> &Value {
        assert_eq!(self.code, 0, "{} / {}", self.envelope, self.stderr);
        assert_eq!(self.envelope["ok"], json!(true), "{}", self.envelope);
        assert_eq!(self.envelope["operation"], json!("lane.screen"));
        &self.envelope["result"]
    }
}

/// `moai lane-screen ARGS` in `cwd` with HOME `home`, `input` on standard input.
fn run_in(cwd: &Path, home: &Path, input: &str, args: &[&str]) -> Run {
    let mut child = Command::new(BIN)
        .arg("lane-screen")
        .args(args)
        .current_dir(cwd)
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.as_bytes());
    let out = child.wait_with_output().unwrap();
    Run {
        code: out.status.code().unwrap(),
        envelope: serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "an envelope on stdout; stdout {:?} stderr {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        }),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

/// A run in empty temporary folders with `--now` fixed.
fn run(input: &str, extra: &[&str]) -> Run {
    let dir = tempfile::tempdir().unwrap();
    let mut args = vec!["--now", NOW_TEXT];
    args.extend_from_slice(extra);
    run_in(dir.path(), dir.path(), input, &args)
}

fn state_of(input: &str) -> String {
    run(input, &[]).result()["state"]
        .as_str()
        .unwrap()
        .to_string()
}

/// An Orca `terminal read --screen --json` answer holding `rows`.
fn orca(rows: &[&str], source: &str) -> String {
    json!({
        "id": "fixture", "ok": true,
        "result": {"terminal": {
            "handle": "term_fixture", "status": "running", "tail": rows,
            "truncated": false, "limited": false, "returnedLineCount": rows.len(),
            "source": source,
        }}
    })
    .to_string()
}

/// A usage-limit screen whose message row is `message`.
fn limit_screen(message: &str) -> String {
    format!("row 1\nrow 2\nrow 3\n{message}\nrow 5\nrow 6\n")
}

fn reset_utc(message: &str, now: &str) -> Value {
    let dir = tempfile::tempdir().unwrap();
    let r = run_in(
        dir.path(),
        dir.path(),
        &limit_screen(message),
        &["--now", now],
    );
    assert_eq!(r.result()["state"], json!("usage_limit"), "{message}");
    r.result()["reset"].clone()
}

// ---- AC-1: state ----

#[test]
fn lane_screen_state_working_idle_and_shell() {
    let r = run(&fixture("synthetic-working.txt"), &[]);
    assert_eq!(r.result()["state"], json!("working"));
    assert!(r.result()["evidence"]
        .as_str()
        .unwrap()
        .contains("esc to interrupt"));
    let r = run(&fixture("synthetic-idle.txt"), &[]);
    assert_eq!(r.result()["state"], json!("idle"));
    assert!(!r.result()["evidence"].as_str().unwrap().is_empty());
    // A shell is never `idle`.
    let r = run(&fixture("synthetic-shell.txt"), &[]);
    assert_eq!(r.result()["state"], json!("unknown"));
    assert_eq!(r.result()["evidence"], json!(null));
}

#[test]
fn lane_screen_state_a_yes_no_dialog_is_a_permission_prompt() {
    let r = run(&fixture("synthetic-permission-prompt.txt"), &[]);
    assert_eq!(r.result()["state"], json!("permission_prompt"));
    assert!(r.result()["evidence"]
        .as_str()
        .unwrap()
        .contains("Do you want"));
    // The same dialog with a working marker above it is still a permission prompt.
    let r = run(
        &fixture("synthetic-permission-prompt-while-working.txt"),
        &[],
    );
    assert_eq!(r.result()["state"], json!("permission_prompt"));
}

/// Audit of slice A (note 3212, B1): a numbered Yes/No choice never reads as `working` or `idle`,
/// whatever draws its cursor or its frame; a Yes sent with Enter would run the command.
#[test]
fn lane_screen_state_a_numbered_yes_no_choice_is_never_working_or_idle() {
    for name in [
        // `>` as the cursor instead of `❯`.
        "synthetic-ascii-cursor-dialog.txt",
        // The dialog inside a vertical frame.
        "synthetic-bordered-dialog.txt",
        // No cursor mark and no `Do you want` question, above an idle input line.
        "synthetic-no-cursor-dialog.txt",
    ] {
        let r = run(&fixture(name), &[]);
        assert_eq!(r.result()["state"], json!("permission_prompt"), "{name}");
        let evidence = r.result()["evidence"].as_str().unwrap();
        assert!(!evidence.is_empty(), "{name}");
        assert_eq!(
            r.result()["pane"]["state"],
            json!("permission_prompt"),
            "{name}"
        );
    }
}

/// Leader decision after the re-audit of slice A (note 3256, R-N1): a question such as
/// `Do you want …` near the bottom of the screen whose choices cannot be read (here `»` and
/// `[1]` in front of the numbers) is not `working` or `idle`: when it is not known, it is unknown.
#[test]
fn lane_screen_state_a_question_near_the_bottom_with_unreadable_choices_is_unknown() {
    for name in [
        "synthetic-unreadable-choice-working.txt",
        "synthetic-unreadable-choice-idle.txt",
    ] {
        let r = run(&fixture(name), &[]);
        assert_eq!(r.result()["state"], json!("unknown"), "{name}");
        let evidence = r.result()["evidence"].as_str().unwrap();
        assert!(evidence.contains("Do you want"), "{name}: {evidence}");
        assert!(
            r.result()["reason"].as_str().unwrap().contains("question"),
            "{name}: {}",
            r.result()["reason"]
        );
        assert_eq!(r.result()["pane"], json!(null), "{name}");
    }
    // The same words above the bottom window (the last 8 non-blank rows) are the transcript, not
    // a dialog: the lane still reads as `working` or `idle`. The idle fixture has exactly 8
    // non-blank rows, so a question put on top of it is the 9th row from the bottom.
    for (name, state) in [
        ("synthetic-working.txt", "working"),
        ("synthetic-idle.txt", "idle"),
    ] {
        let screen = format!("Do you want me to look at this?\n{}", fixture(name));
        assert_eq!(
            state_of(&screen),
            state,
            "{name}: the question is above the window"
        );
    }
    // Boundary N / N-1: the same question as the 8th row from the bottom is inside the window.
    let idle = fixture("synthetic-idle.txt");
    let (first, rest) = idle.split_once('\n').unwrap();
    let inside = format!("{first}\nDo you want me to look at this?\n{rest}");
    assert_eq!(
        state_of(&inside),
        "unknown",
        "the question is the 8th row from the bottom"
    );
}

#[test]
fn lane_screen_state_numbered_rows_that_are_not_yes_and_no_do_not_make_a_prompt() {
    // `Yesterday` and `Not now` only start with the letters of Yes and No.
    let screen = "row 1\nrow 2\nrow 3\n  1. Yesterday\n  2. Not now\nrow 6\nesc to interrupt\n";
    assert_eq!(state_of(screen), "working");
    // A quoted line with `>` but no number is no choice either.
    let screen = "row 1\nrow 2\nrow 3\n> Yes, that is what I said\n> No\nrow 6\nesc to interrupt\n";
    assert_eq!(state_of(screen), "working");
}

#[test]
fn lane_screen_state_another_choice_dialog_is_input_pending() {
    let r = run(&fixture("synthetic-choice-dialog.txt"), &[]);
    assert_eq!(r.result()["state"], json!("input_pending"));
}

#[test]
fn lane_screen_state_a_usage_limit_wins_over_a_menu_and_an_idle_prompt() {
    for name in [
        "synthetic-usage-limit-glm.txt",
        "synthetic-usage-limit-claude.txt",
    ] {
        let r = run(&fixture(name), &[]);
        assert_eq!(r.result()["state"], json!("usage_limit"), "{name}");
        let evidence = r.result()["evidence"].as_str().unwrap();
        assert!(
            evidence.to_lowercase().contains("limit"),
            "{name}: {evidence}"
        );
    }
}

#[test]
fn lane_screen_state_an_orca_screen_read_is_read_like_the_text() {
    let text = fixture("synthetic-working.txt");
    let rows: Vec<&str> = text.lines().collect();
    let r = run(&orca(&rows, "screen"), &[]);
    assert_eq!(r.result()["state"], json!("working"));
    assert_eq!(r.result()["pane"]["source"], json!("orca:screen"));
}

#[test]
fn lane_screen_state_the_pane_object_is_the_lane_status_observation() {
    let r = run(&fixture("synthetic-working.txt"), &[]);
    let pane = r.result()["pane"].as_object().unwrap();
    let keys: Vec<&str> = pane.keys().map(String::as_str).collect();
    assert_eq!(keys, ["state", "source", "observed_at", "until"]);
    assert_eq!(pane["state"], json!("working"));
    assert_eq!(pane["observed_at"], json!(NOW_TEXT));
    assert_eq!(pane["until"], json!(null));
    let r = run(&fixture("synthetic-permission-prompt.txt"), &[]);
    assert_eq!(r.result()["pane"]["state"], json!("permission_prompt"));
    // `lane-status` has no name for these two: no pane object.
    for name in ["synthetic-idle.txt", "synthetic-shell.txt"] {
        let r = run(&fixture(name), &[]);
        assert_eq!(r.result()["pane"], json!(null), "{name}");
    }
}

#[test]
fn lane_screen_state_prints_one_row_and_never_the_whole_screen() {
    let r = run(&fixture("synthetic-permission-prompt.txt"), &[]);
    let text = r.envelope.to_string();
    assert!(text.contains("Do you want"));
    assert!(!text.contains("Drop the last commit"), "{text}");
    assert!(!text.contains("/Users/example/project"), "{text}");
}

// ---- AC-2: reset time ----

#[test]
fn lane_screen_reset_a_time_with_no_zone_is_unknown_with_its_text() {
    // Seen for real on 2026-11-19: a glm lane's limit screen shows no zone, and it was UTC+8.
    let r = run(&fixture("synthetic-usage-limit-glm.txt"), &[]);
    let reset = &r.result()["reset"];
    assert_eq!(reset["utc"], json!(null), "{reset}");
    assert_eq!(reset["unknown"], json!(true), "{reset}");
    assert!(reset["raw"]
        .as_str()
        .unwrap()
        .contains("Your limit will reset at 2026-11-19 17:30:55"));
    assert!(reset["why"].as_str().unwrap().contains("time zone"));
    assert_eq!(r.result()["pane"]["until"], json!(null));
}

#[test]
fn lane_screen_reset_utc_offsets_and_named_zones_are_converted() {
    let base = "Your limit will reset at 2026-11-19 17:30:55";
    let utc = |zone: &str| reset_utc(&format!("{base} {zone}"), NOW_TEXT)["utc"].clone();
    assert_eq!(utc("UTC"), json!("2026-11-19T17:30:55Z"));
    assert_eq!(utc("UTC+8"), json!("2026-11-19T09:30:55Z"));
    assert_eq!(utc("UTC+9"), json!("2026-11-19T08:30:55Z"));
    assert_eq!(utc("(Asia/Seoul)"), json!("2026-11-19T08:30:55Z"));
    assert_eq!(utc("(Asia/Shanghai)"), json!("2026-11-19T09:30:55Z"));
    let r = reset_utc(&format!("{base} UTC+8"), NOW_TEXT);
    assert_eq!(r["source"], json!("screen"));
    assert_eq!(r["unknown"], json!(null));
}

#[test]
fn lane_screen_reset_utc8_is_one_hour_after_utc9() {
    // The accident of 2026-11-19: the printed time was read as UTC (and as local time) when the
    // zone was UTC+8; the renew request went out an hour early.
    let at = |zone: &str| {
        let r = reset_utc(
            &format!("Your limit will reset at 2026-11-19 17:30:55 {zone}"),
            NOW_TEXT,
        );
        moai_cli::forge::policy::parse_utc(r["utc"].as_str().unwrap()).unwrap()
    };
    assert_eq!(at("UTC+8") - at("UTC+9"), 3600);
}

#[test]
fn lane_screen_reset_a_date_or_time_without_a_year_is_the_next_one_after_now() {
    let seoul = reset_utc(
        "5-hour limit reached ∙ resets Nov 20 at 7pm (Asia/Seoul)",
        NOW_TEXT,
    );
    assert_eq!(seoul["utc"], json!("2026-11-20T10:00:00Z"));
    let from_file = run(&fixture("synthetic-usage-limit-claude.txt"), &[]);
    assert_eq!(
        from_file.result()["reset"]["utc"],
        json!("2026-11-20T10:00:00Z")
    );
    assert_eq!(
        reset_utc("limit reached ∙ resets 7pm (UTC)", "2026-11-19T20:00:00Z")["utc"],
        json!("2026-11-20T19:00:00Z")
    );
    assert_eq!(
        reset_utc("limit reached ∙ resets 7pm (UTC)", "2026-11-19T10:00:00Z")["utc"],
        json!("2026-11-19T19:00:00Z")
    );
    assert_eq!(
        reset_utc(
            "limit reached ∙ resets Jan 2 at 1am (UTC)",
            "2026-12-30T00:00:00Z"
        )["utc"],
        json!("2027-01-02T01:00:00Z")
    );
}

#[test]
fn lane_screen_reset_an_unknown_zone_or_an_impossible_time_is_unknown_with_its_text() {
    for message in [
        "limit reached ∙ resets 7pm (America/New_York)",
        "limit reached ∙ resets 7pm (PST)",
        "limit reached ∙ resets 13pm (UTC)",
        "Your limit will reset at 2026-11-19 25:00:00 UTC",
        "Your limit will reset soon",
    ] {
        let reset = reset_utc(message, NOW_TEXT);
        assert_eq!(reset["utc"], json!(null), "{message}: {reset}");
        assert_eq!(reset["unknown"], json!(true), "{message}");
        assert!(
            reset["raw"].as_str().unwrap().contains("limit"),
            "{message}"
        );
    }
}

fn home_with(windows: &[(&str, f64, i64)]) -> tempfile::TempDir {
    let home = tempfile::tempdir().unwrap();
    let map: BTreeMap<String, Window> = windows
        .iter()
        .map(|(k, used, resets)| {
            (
                k.to_string(),
                Window {
                    used_percentage: *used,
                    resets_at: *resets,
                },
            )
        })
        .collect();
    record(home.path(), map, NOW).unwrap();
    home
}

#[test]
fn lane_screen_reset_the_lane_home_record_beats_the_screen_text() {
    let resets_at = NOW + 5 * 3600 + 7;
    let home = home_with(&[
        ("five_hour", 100.0, resets_at),
        ("seven_day", 40.0, NOW + 86_400 * 3),
    ]);
    let cwd = tempfile::tempdir().unwrap();
    let r = run_in(
        cwd.path(),
        cwd.path(),
        &fixture("synthetic-usage-limit-glm.txt"),
        &["--now", NOW_TEXT, "--home", home.path().to_str().unwrap()],
    );
    let reset = &r.result()["reset"];
    assert_eq!(reset["source"], json!("home"), "{reset}");
    assert_eq!(
        reset["utc"],
        json!(moai_cli::forge::policy::format_utc(resets_at))
    );
    assert_eq!(r.result()["pane"]["until"], reset["utc"]);
    // The screen text of the same screen alone stays unknown.
    let r = run(&fixture("synthetic-usage-limit-glm.txt"), &[]);
    assert_eq!(r.result()["reset"]["unknown"], json!(true));
}

#[test]
fn lane_screen_reset_a_window_that_is_not_at_its_limit_is_not_the_reset() {
    let home = home_with(&[("five_hour", 90.0, NOW + 3600)]);
    let cwd = tempfile::tempdir().unwrap();
    let r = run_in(
        cwd.path(),
        cwd.path(),
        &limit_screen("Your limit will reset at 2026-11-19 17:30:55 UTC"),
        &["--now", NOW_TEXT, "--home", home.path().to_str().unwrap()],
    );
    let reset = &r.result()["reset"];
    assert_eq!(reset["source"], json!("screen"), "{reset}");
    assert_eq!(reset["utc"], json!("2026-11-19T17:30:55Z"));
    // A home with no record at all changes nothing.
    let empty = tempfile::tempdir().unwrap();
    let r = run_in(
        cwd.path(),
        cwd.path(),
        &limit_screen("Your limit will reset at 2026-11-19 17:30:55 UTC"),
        &["--now", NOW_TEXT, "--home", empty.path().to_str().unwrap()],
    );
    assert_eq!(r.result()["reset"]["source"], json!("screen"));
}

#[test]
fn lane_screen_reset_only_a_usage_limit_has_a_reset() {
    for name in ["synthetic-working.txt", "synthetic-permission-prompt.txt"] {
        let r = run(&fixture(name), &[]);
        assert_eq!(r.result()["reset"], json!(null), "{name}");
    }
}

// ---- AC-3: unknown, unreachable, nothing written ----

#[test]
fn lane_screen_unknown_a_narrow_tail_is_unknown() {
    // Seen for real: a narrow window left a tail of three rows showing only `esc to interrupt`.
    let r = run(&fixture("synthetic-narrow-tail.txt"), &[]);
    assert_eq!(r.result()["state"], json!("unknown"));
    assert_eq!(r.result()["rows"], json!(3));
    assert!(r.result()["reason"].as_str().unwrap().contains("six"));
}

#[test]
fn lane_screen_unknown_five_rows_are_too_few_and_six_are_enough() {
    let screen = |filler: usize, tail: &[&str]| {
        let mut rows: Vec<String> = (0..filler).map(|i| format!("row {i}")).collect();
        rows.extend(tail.iter().map(|s| s.to_string()));
        rows.join("\n")
    };
    let working = ["✳ Frosting… (esc to interrupt)"];
    assert_eq!(state_of(&screen(4, &working)), "unknown");
    assert_eq!(state_of(&screen(5, &working)), "working");
    let idle = ["❯ ", "  ⏵⏵ bypass permissions on (shift+tab to cycle)"];
    assert_eq!(state_of(&screen(3, &idle)), "unknown");
    assert_eq!(state_of(&screen(4, &idle)), "idle");
}

#[test]
fn lane_screen_unknown_a_short_screen_still_refuses_on_its_own_marker() {
    let prompt = "Do you want to proceed?\n ❯ 1. Yes\n   3. No\n";
    assert_eq!(state_of(prompt), "permission_prompt");
    assert_eq!(
        state_of("Your limit will reset at 2026-11-19 17:30:55\n"),
        "usage_limit"
    );
}

#[test]
fn lane_screen_unknown_what_is_not_a_rendered_screen_is_unknown_never_idle() {
    let text = fixture("synthetic-idle.txt");
    let rows: Vec<&str> = text.lines().collect();
    for source in ["stream", "screen-unavailable", "something-else"] {
        let r = run(&orca(&rows, source), &[]);
        assert_eq!(r.result()["state"], json!("unknown"), "{source}");
        assert!(
            r.result()["reason"].as_str().unwrap().contains("screen"),
            "{source}"
        );
    }
    // No source at all: the host predates the field.
    let no_source =
        json!({"ok": true, "result": {"terminal": {"status": "running", "tail": rows}}});
    assert_eq!(state_of(&no_source.to_string()), "unknown");
    // No rows, blank rows, no input.
    assert_eq!(state_of(&orca(&[], "screen")), "unknown");
    assert_eq!(state_of(&orca(&["", "  ", ""], "screen")), "unknown");
    assert_eq!(state_of(""), "unknown");
    assert_eq!(state_of("\n\n   \n"), "unknown");
    let no_tail = json!({"ok": true, "result": {"terminal": {"source": "screen"}}});
    assert_eq!(state_of(&no_tail.to_string()), "unknown");
}

#[test]
fn lane_screen_unknown_a_gone_terminal_is_unreachable() {
    let exited = json!({"ok": true, "result": {"terminal": {"status": "exited", "tail": [], "source": "screen"}}});
    let r = run(&exited.to_string(), &[]);
    assert_eq!(r.result()["state"], json!("unreachable"));
    assert_eq!(r.result()["pane"]["state"], json!("unreachable"));
    let failed = json!({"ok": false, "error": {"message": "terminal not found"}});
    assert_eq!(state_of(&failed.to_string()), "unreachable");
}

#[test]
fn lane_screen_unknown_writes_nothing() {
    let cwd = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let before = common::cli::tree_state(cwd.path());
    let home_before = common::cli::tree_state(home.path());
    for name in [
        "synthetic-working.txt",
        "synthetic-usage-limit-glm.txt",
        "synthetic-permission-prompt.txt",
    ] {
        let r = run_in(
            cwd.path(),
            home.path(),
            &fixture(name),
            &["--now", NOW_TEXT, "--home", home.path().to_str().unwrap()],
        );
        assert_eq!(r.code, 0, "{name}: {}", r.envelope);
    }
    assert_eq!(common::cli::tree_state(cwd.path()), before);
    assert_eq!(common::cli::tree_state(home.path()), home_before);
}

#[test]
fn lane_screen_unknown_a_bad_now_or_an_oversize_input_is_refused() {
    let dir = tempfile::tempdir().unwrap();
    let r = run_in(dir.path(), dir.path(), "x", &["--now", "yesterday"]);
    assert_eq!(r.code, 1, "{}", r.envelope);
    assert_eq!(r.envelope["ok"], json!(false));
    let big = "x\n".repeat(40_000);
    let r = run_in(dir.path(), dir.path(), &big, &["--now", NOW_TEXT]);
    assert_eq!(r.code, 1, "{}", r.envelope);
    assert_eq!(r.envelope["ok"], json!(false));
    // Exactly at the limit is read.
    let at_limit = "x".repeat(65_536);
    let r = run_in(dir.path(), dir.path(), &at_limit, &["--now", NOW_TEXT]);
    assert_eq!(r.code, 0, "{}", r.envelope);
    // One byte over the limit is refused (audit note 3212, N2).
    let over = "x".repeat(65_537);
    let r = run_in(dir.path(), dir.path(), &over, &["--now", NOW_TEXT]);
    assert_eq!(r.code, 1, "{}", r.envelope);
    assert_eq!(r.envelope["ok"], json!(false));
}

/// Audit of slice A (note 3212, N1): a year far outside the calendar made the date arithmetic
/// overflow and a debug build panic (exit 101, no envelope).
#[test]
fn lane_screen_reset_a_year_outside_the_calendar_is_unknown_and_never_a_panic() {
    for message in [
        "limit reached ∙ resets 9999999999999-01-01T00:00:00Z",
        "limit reached ∙ resets 99999999999999999999-01-01T00:00:00Z",
        "limit reached ∙ resets 10000-01-01T00:00:00Z",
        "limit reached ∙ resets 1969-12-31T23:59:59Z",
    ] {
        let reset = reset_utc(message, NOW_TEXT);
        assert_eq!(reset["utc"], json!(null), "{message}");
        assert_eq!(reset["unknown"], json!(true), "{message}");
    }
    // The edges that are inside the calendar still read.
    assert_eq!(
        reset_utc("limit reached ∙ resets 9999-12-31T23:59:59Z", NOW_TEXT)["utc"],
        json!("9999-12-31T23:59:59Z")
    );
    assert_eq!(
        reset_utc("limit reached ∙ resets 1970-01-01T00:00:00Z", NOW_TEXT)["utc"],
        json!("1970-01-01T00:00:00Z")
    );
}

/// SPEC-20261119-lane-screen-scan (slice C): `moai lane-screen --scan` lists every terminal that
/// `orca terminal list` reports with its screen state. The Orca program is a fake shell script
/// (`MOAI_ORCA_BIN`) that records every call in `calls.txt`, so "sent nothing" is "no `send` call
/// and no `sent.txt`".
#[cfg(unix)]
mod scan_tests {
    use super::*;
    use std::path::PathBuf;

    /// The fake `orca`. `list.json` is what `terminal list` prints (`list-fail` makes it exit 5),
    /// `read-<handle>.json` what `terminal read` prints (`read-fail-<handle>` makes it exit 3).
    const FAKE_ORCA: &str = r#"#!/bin/sh
d="$FAKE_ORCA_DIR"
printf '%s\n' "$*" >> "$d/calls.txt"
case "$2" in
list)
  if [ -f "$d/list-fail" ]; then echo "orca is not running" >&2; exit 5; fi
  cat "$d/list.json"
  ;;
read)
  if [ -f "$d/read-fail-$4" ]; then echo "terminal not found" >&2; exit 3; fi
  cat "$d/read-$4.json"
  ;;
send)
  printf '%s\n' "$*" >> "$d/sent.txt"
  echo '{"ok":true}'
  ;;
*)
  exit 9
  ;;
esac
"#;

    struct Fake {
        dir: tempfile::TempDir,
        bin: PathBuf,
    }

    impl Fake {
        /// A fake whose `terminal list` prints `list` as it is.
        fn new(list: &str) -> Fake {
            let dir = tempfile::tempdir().unwrap();
            let bin = dir.path().join("orca");
            common::write_executable(&bin, FAKE_ORCA);
            fs::write(dir.path().join("list.json"), list).unwrap();
            Fake { dir, bin }
        }

        /// `handle` reads as the rendered screen of `text`.
        fn screen(&self, handle: &str, text: &str) {
            let rows: Vec<&str> = text.lines().collect();
            self.raw(handle, &orca(&rows, "screen"));
        }

        /// `handle` reads as exactly `content`.
        fn raw(&self, handle: &str, content: &str) {
            fs::write(self.dir.path().join(format!("read-{handle}.json")), content).unwrap();
        }

        fn flag(&self, name: &str) {
            fs::write(self.dir.path().join(name), "").unwrap();
        }

        /// Every call of the fake, one line each, in order.
        fn calls(&self) -> Vec<String> {
            fs::read_to_string(self.dir.path().join("calls.txt"))
                .map(|t| t.lines().map(String::from).collect())
                .unwrap_or_default()
        }

        fn sent(&self) -> bool {
            self.dir.path().join("sent.txt").exists()
        }

        /// `moai lane-screen --scan ARGS` with this fake as the Orca program.
        fn run(&self, cwd: &Path, home: &Path, extra: &[&str]) -> Run {
            run_scan(&self.bin, self.dir.path(), cwd, home, extra)
        }
    }

    fn run_scan(bin: &Path, fake_dir: &Path, cwd: &Path, home: &Path, extra: &[&str]) -> Run {
        let out = Command::new(BIN)
            .args(["lane-screen", "--scan", "--now", NOW_TEXT])
            .args(extra)
            .current_dir(cwd)
            .env("HOME", home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env("MOAI_ORCA_BIN", bin)
            .env("FAKE_ORCA_DIR", fake_dir)
            .stdin(Stdio::null())
            .output()
            .unwrap();
        Run {
            code: out.status.code().unwrap(),
            envelope: serde_json::from_slice(&out.stdout).unwrap_or(Value::Null),
            stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        }
    }

    /// An `orca terminal list --json` answer.
    fn list(terminals: &[Value], total: usize, truncated: bool) -> String {
        json!({
            "id": "fixture", "ok": true,
            "result": {
                "terminals": terminals, "hostScope": "local", "topologyRevisions": {},
                "totalCount": total, "truncated": truncated,
            }
        })
        .to_string()
    }

    fn term(handle: &str, title: &str, branch: &str) -> Value {
        json!({
            "handle": handle, "title": title, "branch": branch, "connected": true,
            "writable": true, "worktreePath": "/example/work", "preview": "",
        })
    }

    fn lines(result: &Value) -> Vec<String> {
        result["table"]
            .as_array()
            .unwrap()
            .iter()
            .map(|l| l.as_str().unwrap().to_string())
            .collect()
    }

    /// The failure message of a run that must have failed with an error envelope.
    fn failure(r: &Run) -> &str {
        assert_eq!(r.code, 1, "{} / {}", r.envelope, r.stderr);
        assert_eq!(r.envelope["ok"], json!(false), "{}", r.envelope);
        r.envelope["error"]["message"].as_str().unwrap()
    }

    #[test]
    fn lane_screen_scan_lists_every_terminal_and_reads_them_one_by_one() {
        let fake = Fake::new(&list(
            &[
                term("t_work", "work lane", "spec/a"),
                term("t_ask", "ask lane", "spec/b"),
                term("t_gone", "gone lane", "spec/c"),
            ],
            3,
            false,
        ));
        fake.screen("t_work", &fixture("synthetic-working.txt"));
        fake.screen("t_ask", &fixture("synthetic-permission-prompt.txt"));
        fake.flag("read-fail-t_gone");
        let cwd = tempfile::tempdir().unwrap();
        let home = tempfile::tempdir().unwrap();
        let before = (
            common::cli::tree_state(cwd.path()),
            common::cli::tree_state(home.path()),
        );
        let r = fake.run(cwd.path(), home.path(), &[]);
        let res = r.result();
        let rows = res["terminals"].as_array().unwrap();
        let handles: Vec<_> = rows.iter().map(|t| t["handle"].as_str().unwrap()).collect();
        assert_eq!(handles, ["t_work", "t_ask", "t_gone"]);
        let states: Vec<_> = rows.iter().map(|t| t["state"].as_str().unwrap()).collect();
        assert_eq!(states, ["working", "permission_prompt", "unreachable"]);
        assert_eq!(rows[0]["title"], json!("work lane"));
        assert_eq!(rows[1]["branch"], json!("spec/b"));
        assert!(
            rows[1]["evidence"].as_str().unwrap().contains("Yes"),
            "{}",
            rows[1]
        );
        assert!(
            rows[2]["reason"]
                .as_str()
                .unwrap()
                .contains("exited with code 3"),
            "{}",
            rows[2]
        );
        assert_eq!(res["count"], json!(3));
        assert_eq!(res["total"], json!(3));
        assert_eq!(res["truncated"], json!(false));
        assert_eq!(
            res["by_state"],
            json!({"working": 1, "permission_prompt": 1, "unreachable": 1})
        );
        let table = lines(res);
        assert_eq!(table.len(), 4, "{table:?}");
        for (line, (handle, state)) in table[1..]
            .iter()
            .zip([("t_work", "working"), ("t_ask", "permission_prompt"), ("t_gone", "unreachable")])
        {
            assert!(
                line.contains(handle) && line.contains(state),
                "{line} / {handle} {state}"
            );
        }
        assert!(!table.last().unwrap().contains("cut short"), "{table:?}");
        // One list, then one read per terminal in list order, and nothing else.
        assert_eq!(
            fake.calls(),
            [
                "terminal list --json",
                "terminal read --terminal t_work --screen --json",
                "terminal read --terminal t_ask --screen --json",
                "terminal read --terminal t_gone --screen --json",
            ]
        );
        assert!(!fake.sent());
        assert_eq!(
            (
                common::cli::tree_state(cwd.path()),
                common::cli::tree_state(home.path())
            ),
            before
        );
    }

    #[test]
    fn lane_screen_scan_a_handle_orca_could_read_as_an_option_is_not_read() {
        let fake = Fake::new(&list(
            &[
                term("-bad", "dash", "b"),
                json!({"title": "no handle"}),
                term("has space", "space", "b"),
                term("ok1", "fine", "b"),
            ],
            4,
            false,
        ));
        fake.screen("ok1", &fixture("synthetic-idle.txt"));
        let dir = tempfile::tempdir().unwrap();
        let r = fake.run(dir.path(), dir.path(), &[]);
        let res = r.result();
        let rows = res["terminals"].as_array().unwrap();
        let states: Vec<_> = rows.iter().map(|t| t["state"].as_str().unwrap()).collect();
        assert_eq!(states, ["unreachable", "unreachable", "unreachable", "idle"]);
        for row in &rows[..3] {
            assert!(
                row["reason"].as_str().unwrap().contains("handle"),
                "{row}"
            );
        }
        assert_eq!(rows[1]["handle"], Value::Null);
        // Only the good handle reached Orca.
        assert_eq!(
            fake.calls(),
            [
                "terminal list --json",
                "terminal read --terminal ok1 --screen --json",
            ]
        );
    }

    #[test]
    fn lane_screen_scan_an_answer_that_is_plain_text_is_not_a_screen() {
        let fake = Fake::new(&list(&[term("t_text", "text", "b")], 1, false));
        // The text of a working screen, but not an Orca answer: it must not read as `working`.
        fake.raw("t_text", &fixture("synthetic-working.txt"));
        let dir = tempfile::tempdir().unwrap();
        let r = fake.run(dir.path(), dir.path(), &[]);
        let row = &r.result()["terminals"][0];
        assert_eq!(row["state"], json!("unreachable"), "{row}");
        assert!(
            row["reason"].as_str().unwrap().contains("not an Orca answer"),
            "{row}"
        );
    }

    #[test]
    fn lane_screen_scan_title_and_branch_are_cleaned_and_cut() {
        let title = format!("evil\u{1b}[31m\ttitle{}", "x".repeat(200));
        let fake = Fake::new(&list(&[term("t_one", &title, "a\nb\u{7}")], 1, false));
        fake.screen("t_one", &fixture("synthetic-working.txt"));
        let dir = tempfile::tempdir().unwrap();
        let r = fake.run(dir.path(), dir.path(), &[]);
        let res = r.result();
        let row = &res["terminals"][0];
        let shown = row["title"].as_str().unwrap();
        assert!(shown.starts_with("evil title"), "{shown:?}");
        assert!(shown.chars().count() <= 81, "{}", shown.chars().count());
        assert!(!shown.chars().any(char::is_control), "{shown:?}");
        let branch = row["branch"].as_str().unwrap();
        assert!(!branch.chars().any(char::is_control), "{branch:?}");
        for line in lines(res) {
            assert!(!line.chars().any(char::is_control), "{line:?}");
        }
    }

    #[test]
    fn lane_screen_scan_reset_is_in_utc_next_to_the_state() {
        let fake = Fake::new(&list(
            &[
                term("t_claude", "claude", "b"),
                term("t_glm", "glm", "b"),
                term("t_work", "work", "b"),
            ],
            3,
            false,
        ));
        fake.screen("t_claude", &fixture("synthetic-usage-limit-claude.txt"));
        fake.screen("t_glm", &fixture("synthetic-usage-limit-glm.txt"));
        fake.screen("t_work", &fixture("synthetic-working.txt"));
        let dir = tempfile::tempdir().unwrap();
        let r = fake.run(dir.path(), dir.path(), &[]);
        let res = r.result();
        let rows = res["terminals"].as_array().unwrap();
        assert_eq!(rows[0]["state"], json!("usage_limit"));
        assert_eq!(rows[0]["reset"]["utc"], json!("2026-11-20T10:00:00Z"));
        assert_eq!(rows[0]["reset"]["source"], json!("screen"));
        let table = lines(res);
        assert!(
            table[1].contains("t_claude")
                && table[1].contains("usage_limit")
                && table[1].contains("2026-11-20T10:00:00Z"),
            "{}",
            table[1]
        );
        // A reset time with no zone is never turned into a UTC time.
        assert_eq!(rows[1]["state"], json!("usage_limit"));
        assert_eq!(rows[1]["reset"]["utc"], Value::Null);
        assert_eq!(rows[1]["reset"]["unknown"], json!(true));
        assert!(table[2].contains("unknown"), "{}", table[2]);
        assert!(!table[2].contains("T17:30:55Z"), "{}", table[2]);
        assert!(!table[2].contains("T09:30:55Z"), "{}", table[2]);
        // A lane that is not at a limit has no reset time.
        assert_eq!(rows[2]["reset"], Value::Null);
    }

    #[test]
    fn lane_screen_scan_truncated_list_says_how_many_of_how_many() {
        let fake = Fake::new(&list(
            &[term("t_a", "a", "b"), term("t_b", "b", "b")],
            5,
            true,
        ));
        fake.screen("t_a", &fixture("synthetic-working.txt"));
        fake.screen("t_b", &fixture("synthetic-idle.txt"));
        let dir = tempfile::tempdir().unwrap();
        let r = fake.run(dir.path(), dir.path(), &[]);
        let res = r.result();
        assert_eq!(res["truncated"], json!(true));
        assert_eq!(res["total"], json!(5));
        assert_eq!(res["count"], json!(2));
        let table = lines(res);
        let last = table.last().unwrap();
        assert!(
            last.contains("2 of 5") && last.contains("cut short"),
            "{table:?}"
        );
    }

    #[test]
    fn lane_screen_scan_fails_when_the_list_cannot_be_read_and_sends_and_writes_nothing() {
        type Setup = fn(&Fake);
        let cases: [(&str, Setup, &str); 4] = [
            ("list exits non-zero", |f| f.flag("list-fail"), "exited with code 5"),
            (
                "list is not JSON",
                |f| fs::write(f.dir.path().join("list.json"), "orca says hello").unwrap(),
                "list could not be read",
            ),
            (
                "list says ok false",
                |f| {
                    fs::write(
                        f.dir.path().join("list.json"),
                        r#"{"ok":false,"error":{"message":"no runtime"}}"#,
                    )
                    .unwrap()
                },
                "list could not be read",
            ),
            (
                "list has no terminals array",
                |f| fs::write(f.dir.path().join("list.json"), r#"{"ok":true,"result":{}}"#).unwrap(),
                "list could not be read",
            ),
        ];
        for (name, setup, detail) in cases {
            let fake = Fake::new("{}");
            setup(&fake);
            let cwd = tempfile::tempdir().unwrap();
            let home = tempfile::tempdir().unwrap();
            let r = fake.run(cwd.path(), home.path(), &[]);
            let message = failure(&r);
            assert!(message.contains("list could not be read"), "{name}: {message}");
            assert!(message.contains(detail), "{name}: {message}");
            // Only the list was asked for: no read, no send, no file.
            assert_eq!(fake.calls(), ["terminal list --json"], "{name}");
            assert!(!fake.sent(), "{name}");
            assert!(common::cli::tree_state(cwd.path()).is_empty(), "{name}");
            assert!(common::cli::tree_state(home.path()).is_empty(), "{name}");
        }
        // The Orca program is absent.
        let fake = Fake::new("{}");
        let cwd = tempfile::tempdir().unwrap();
        let missing = cwd.path().join("no-such-orca");
        let r = run_scan(&missing, fake.dir.path(), cwd.path(), cwd.path(), &[]);
        let message = failure(&r);
        assert!(message.contains("list could not be read"), "{message}");
        assert!(message.contains("cannot run"), "{message}");
        assert!(fake.calls().is_empty());
        assert!(!fake.sent());
    }

    #[test]
    fn lane_screen_scan_cannot_be_combined_with_home() {
        let fake = Fake::new(&list(&[term("t_a", "a", "b")], 1, false));
        let dir = tempfile::tempdir().unwrap();
        let r = fake.run(dir.path(), dir.path(), &["--home", dir.path().to_str().unwrap()]);
        assert_ne!(r.code, 0, "{}", r.envelope);
        assert!(r.stderr.contains("--home"), "{}", r.stderr);
        assert!(fake.calls().is_empty(), "Orca must not run: {:?}", fake.calls());
    }
}
