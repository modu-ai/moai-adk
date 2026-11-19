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

/// Leader decision after the audit of slice B (note 3362, B1): a numbered `Yes` row together with
/// a numbered `No` row is a confirmation whatever stands before the numbers (a cursor glyph the
/// reader never listed, `»`) and with or without a question the reader knows: Enter would pick
/// Yes. Each of these screens read as `working` before the repair.
#[test]
fn lane_screen_state_numbered_yes_no_after_any_prefix_is_a_permission_prompt() {
    for name in [
        // A working footer, `◆` before the number, no question row at all.
        "synthetic-foreign-cursor-no-question-working.txt",
        // A working footer, a question sentence that is none of the known question words.
        "synthetic-foreign-cursor-far-question-working.txt",
        // Slice A's R-N1 screen: `»` before the number, under a working footer.
        "synthetic-unreadable-choice-working.txt",
    ] {
        let r = run(&fixture(name), &[]);
        assert_eq!(r.result()["state"], json!("permission_prompt"), "{name}");
        let evidence = r.result()["evidence"].as_str().unwrap();
        assert!(!evidence.is_empty(), "{name}");
        assert_eq!(r.result()["pane"]["state"], json!("permission_prompt"));
    }
    // Any prefix of marks is skipped: the nine of the audit and a few more. A prefix with a
    // letter or a digit is no mark, so `ok 1. Yes` is a sentence and not a choice.
    for prefix in [
        "◆ ", "» ", "• ", "- ", "* ", "→→ ", ">>> ", "✓ ", "· ", "│ ◆ ", "[•] ", "☞ ", "◆",
    ] {
        let screen =
            format!("row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\n{prefix}1. Yes\n   2. No\n");
        assert_eq!(state_of(&screen), "permission_prompt", "prefix {prefix:?}");
    }
    let screen = "row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\nok 1. Yes\nok 2. No\n";
    assert_eq!(state_of(screen), "working");
}

/// Leader decision of note 3256 (R-N1) narrowed by note 3362 (S-N2): a confirmation question in
/// the last 8 non-blank rows with a `yes` or `no` word on a row below it, and no choice rows the
/// reader can use (here `[1] Yes` and `[2] No`), is `unknown`. A question that ends the lane's
/// turn, with only the input line below it, is not (see the next test).
#[test]
fn lane_screen_state_a_question_above_unreadable_yes_and_no_rows_is_unknown() {
    let r = run(&fixture("synthetic-unreadable-choice-idle.txt"), &[]);
    assert_eq!(r.result()["state"], json!("unknown"));
    let evidence = r.result()["evidence"].as_str().unwrap();
    assert!(evidence.contains("Do you want"), "{evidence}");
    assert!(
        r.result()["reason"].as_str().unwrap().contains("question"),
        "{}",
        r.result()["reason"]
    );
    assert_eq!(r.result()["pane"], json!(null));
    // A bracket question marks itself: `[y/n]` is the choice, so it is unknown with no row below.
    let screen = "row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\nOverwrite the file [y/n]\n";
    assert_eq!(state_of(screen), "unknown");
    // The window is the last 8 non-blank rows. The question with `[1] Yes`, `[2] No` and five
    // more rows under it is the 8th row from the bottom (inside), with six more it is the 9th.
    let window = |below: usize| {
        let mut rows = vec![
            "row 1".to_string(),
            "row 2".to_string(),
            "row 3".to_string(),
            "row 4".to_string(),
            "esc to interrupt".to_string(),
            "Do you want me to look at this?".to_string(),
            "[1] Yes".to_string(),
            "[2] No".to_string(),
        ];
        rows.extend((0..below).map(|i| format!("more {i}")));
        state_of(&rows.join("\n"))
    };
    assert_eq!(
        window(5),
        "unknown",
        "the question is the 8th from the bottom"
    );
    assert_eq!(
        window(6),
        "working",
        "the question is the 9th from the bottom"
    );
}

/// Re-audit of slice B (note 3433, R-B1 and R-N1): a known question near the bottom is `unknown`
/// when a choice word (`yes no allow deny y n`) is on the question row itself or below it, or
/// when two or more numbered rows are below it, whatever the words of the choices are. Each of
/// these five screens read as `working` (a send went in) before the repair.
#[test]
fn lane_screen_state_a_question_with_its_choices_on_the_row_or_below_is_unknown() {
    for name in [
        // R-B1: the choice is on the question row.
        "synthetic-question-same-line-working.txt",
        "synthetic-fingerprint-question-working.txt",
        // R-N1: the choices do not use the words yes and no.
        "synthetic-question-allow-deny-working.txt",
        "synthetic-question-allow-once-always-deny-working.txt",
        "synthetic-question-numbered-y-n-working.txt",
    ] {
        let r = run(&fixture(name), &[]);
        assert_eq!(r.result()["state"], json!("unknown"), "{name}");
        let evidence = r.result()["evidence"].as_str().unwrap();
        assert!(
            evidence.contains("Do you want") || evidence.contains("Are you sure"),
            "{name}: {evidence}"
        );
        assert_eq!(r.result()["pane"], json!(null), "{name}");
    }
    // Two numbered rows under a known question are enough, whatever they say; one is not.
    let two = "row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\nWould you like me to push?\n  1. Proceed\n  2. Cancel\n";
    assert_eq!(state_of(two), "unknown");
    let one = "row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\nWould you like me to push?\n  1. maybe later\n";
    assert_eq!(state_of(one), "working");
    // Each choice word makes a row below the question a choice row, as a whole word only.
    for word in ["Allow", "deny", "y", "N", "yes", "No", "(y/N/a)"] {
        let screen = format!(
            "row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\nWould you like me to push?\n{word}\n"
        );
        assert_eq!(state_of(&screen), "unknown", "{word}");
    }
    for word in ["Allowance", "denied", "yard", "Nine", "maybe"] {
        let screen = format!(
            "row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\nWould you like me to push?\n{word}\n"
        );
        assert_eq!(state_of(&screen), "working", "{word}");
    }
    // The same on the question row itself.
    let same = "row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\nDo you want to proceed? (y/N/a)\n";
    assert_eq!(state_of(same), "unknown");
}

/// Leader decision S-N2 (note 3362): a lane that ended its turn with a question (`Would you like
/// me to …?`) and waits at its input line is the screen a send is meant for. Only a `yes` or a
/// `no` as a whole word on a row below the question makes it `unknown`.
#[test]
fn lane_screen_state_a_lane_that_ended_on_a_question_is_idle() {
    let idle = fixture("synthetic-question-then-input-idle.txt");
    assert_eq!(state_of(&idle), "idle");
    // The same under a working footer.
    let working = "row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\nWould you like me to push?\n";
    assert_eq!(state_of(working), "working");
    // Words that only contain yes or no are no answer row: `Yesterday`, `Nobody`, `Not now`.
    for below in ["Yesterday it ran", "Nobody", "Not now"] {
        let screen = format!(
            "row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\nWould you like me to push?\n{below}\n"
        );
        assert_eq!(state_of(&screen), "working", "{below}");
    }
    // A whole word below the question is unknown, in any case and with any mark next to it.
    for below in ["yes, go ahead", "No - wait", "(Yes)", "answer: NO."] {
        let screen = format!(
            "row 1\nrow 2\nrow 3\nrow 4\nesc to interrupt\nWould you like me to push?\n{below}\n"
        );
        assert_eq!(state_of(&screen), "unknown", "{below}");
    }
    // A yes above the question does not count: the rule looks below it.
    let screen = "row 1\nrow 2\nrow 3\nyes it did\nesc to interrupt\nWould you like me to push?\n";
    assert_eq!(state_of(screen), "working");
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
