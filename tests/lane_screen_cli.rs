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
  if [ -f "$d/list-fail-msg" ]; then cat "$d/list-fail-msg" >&2; exit 5; fi
  if [ -f "$d/list-fail" ]; then echo "orca is not running" >&2; exit 5; fi
  cat "$d/list.json"
  ;;
read)
  if [ -f "$d/read-fail-msg-$4" ]; then cat "$d/read-fail-msg-$4" >&2; exit 3; fi
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

        /// A file of the fake holding `content` (the message of a failing call, for example).
        fn file(&self, name: &str, content: &str) {
            fs::write(self.dir.path().join(name), content).unwrap();
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
            rows[1]["evidence"]
                .as_str()
                .unwrap()
                .contains("Do you want to proceed?"),
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
        for (line, (handle, state)) in table[1..].iter().zip([
            ("t_work", "working"),
            ("t_ask", "permission_prompt"),
            ("t_gone", "unreachable"),
        ]) {
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
        assert_eq!(
            states,
            ["unreachable", "unreachable", "unreachable", "idle"]
        );
        for row in &rows[..3] {
            assert!(row["reason"].as_str().unwrap().contains("handle"), "{row}");
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
            row["reason"]
                .as_str()
                .unwrap()
                .contains("not an Orca answer"),
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

    /// Re-audit R-N3 (note 3256): a Yes/No menu wins over a usage limit in the classifier, so the
    /// lane reads `permission_prompt` with no reset time and drops out of a list of limits. The
    /// scan keeps the reset time of a limit message that is on a `permission_prompt` screen too.
    #[test]
    fn lane_screen_scan_reset_is_kept_when_a_yes_no_menu_hides_the_usage_limit() {
        let fake = Fake::new(&list(
            &[term("t_both", "both", "b"), term("t_menu", "menu", "b")],
            2,
            false,
        ));
        let both = format!(
            "  5-hour limit reached ∙ resets Nov 20 at 7pm (Asia/Seoul)\n{}",
            fixture("synthetic-permission-prompt.txt")
        );
        fake.screen("t_both", &both);
        fake.screen("t_menu", &fixture("synthetic-permission-prompt.txt"));
        let dir = tempfile::tempdir().unwrap();
        let r = fake.run(dir.path(), dir.path(), &[]);
        let res = r.result();
        let rows = res["terminals"].as_array().unwrap();
        assert_eq!(rows[0]["state"], json!("permission_prompt"), "{}", rows[0]);
        assert_eq!(rows[0]["reset"]["utc"], json!("2026-11-20T10:00:00Z"));
        assert!(
            rows[0]["reason"].as_str().unwrap().contains("usage-limit"),
            "{}",
            rows[0]
        );
        assert!(
            lines(res)[1].contains("2026-11-20T10:00:00Z"),
            "{:?}",
            lines(res)
        );
        // A menu with no limit message has no reset time.
        assert_eq!(rows[1]["state"], json!("permission_prompt"));
        assert_eq!(rows[1]["reset"], Value::Null);
        assert!(!rows[1]["reason"].as_str().unwrap().contains("usage-limit"));
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
            (
                "list exits non-zero",
                |f| f.flag("list-fail"),
                "exited with code 5",
            ),
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
                |f| {
                    fs::write(f.dir.path().join("list.json"), r#"{"ok":true,"result":{}}"#).unwrap()
                },
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
            assert!(
                message.contains("list could not be read"),
                "{name}: {message}"
            );
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

    /// Boundary N / N+1: at most 200 terminals are read; one more is "cut short" and says so.
    #[test]
    fn lane_screen_scan_reads_at_most_200_terminals() {
        for (n, cut) in [(200usize, false), (201, true)] {
            let terminals: Vec<Value> = (0..n)
                .map(|i| term(&format!("t{i}"), "lane", "spec/x"))
                .collect();
            let fake = Fake::new(&list(&terminals, n, false));
            let dir = tempfile::tempdir().unwrap();
            let r = fake.run(dir.path(), dir.path(), &[]);
            let res = r.result();
            assert_eq!(res["count"], json!(n.min(200)), "n={n}");
            assert_eq!(res["total"], json!(n), "n={n}");
            assert_eq!(res["truncated"], json!(cut), "n={n}");
            assert_eq!(fake.calls().len(), 1 + n.min(200), "n={n}");
            let table = lines(res);
            let last = table.last().unwrap();
            assert_eq!(last.contains("cut short"), cut, "n={n}: {last}");
            if cut {
                assert!(last.contains("200 of 201"), "{last}");
            }
        }
    }

    #[test]
    fn lane_screen_scan_cannot_be_combined_with_home() {
        let fake = Fake::new(&list(&[term("t_a", "a", "b")], 1, false));
        let dir = tempfile::tempdir().unwrap();
        let r = fake.run(
            dir.path(),
            dir.path(),
            &["--home", dir.path().to_str().unwrap()],
        );
        assert_ne!(r.code, 0, "{}", r.envelope);
        assert!(r.stderr.contains("--home"), "{}", r.stderr);
        assert!(
            fake.calls().is_empty(),
            "Orca must not run: {:?}",
            fake.calls()
        );
    }

    /// The three synthetic secret shapes of re-audit C-B1 (note 3497), each built from pieces at
    /// run time so no shape stands in this file: (GitHub token prefix, key-like `sk` shape,
    /// `password=` pair) with the value that must not come out of the answer.
    fn secrets() -> [(String, &'static str); 3] {
        let gh_value = "A1b2C3d4E5f6G7h8I9j0";
        let sk_value = "abcdefghij0123456789";
        let pw_value = "hunter2hunter2";
        [
            (["gh", "p_"].concat() + gh_value, gh_value),
            (["s", "k-"].concat() + sk_value, sk_value),
            (["pass", "word="].concat() + pw_value, pw_value),
        ]
    }

    /// Re-audit C-B1: `--scan` takes text from terminals the caller has never seen (a title a
    /// shell sets from the command line it runs, the screen, the error of a failed read, the
    /// error of a failed list); a secret shape in any of the five places never reaches the answer.
    #[test]
    fn lane_screen_scan_secret_shapes_from_other_terminals_never_reach_the_answer() {
        for (secret, value) in secrets() {
            let fake = Fake::new(&list(
                &[
                    term(
                        "t_title",
                        &format!("deploy {secret}"),
                        &format!("fix/{secret}"),
                    ),
                    term("t_ask", "ask", "b"),
                    term("t_limit", "limit", "b"),
                    term("t_gone", "gone", "b"),
                ],
                4,
                false,
            ));
            fake.screen("t_title", &fixture("synthetic-working.txt"));
            let ask = format!(
                "{}Do you want to proceed? (yes/no) {secret}\n",
                fixture("synthetic-working.txt")
            );
            fake.screen("t_ask", &ask);
            let limit = fixture("synthetic-usage-limit-glm.txt")
                .replace("17:30:55", &format!("17:30:55 {secret}"));
            fake.screen("t_limit", &limit);
            fake.file("read-fail-msg-t_gone", &format!("denied {secret}\n"));
            let dir = tempfile::tempdir().unwrap();
            let r = fake.run(dir.path(), dir.path(), &[]);
            let res = r.result();
            let everything = format!("{} {}", r.envelope, r.stderr);
            assert!(!everything.contains(value), "{secret}: {everything}");
            let rows = res["terminals"].as_array().unwrap();
            for (i, key) in [(0, "title"), (0, "branch")] {
                assert!(
                    rows[i][key].as_str().unwrap().contains("[redacted]"),
                    "{key}: {}",
                    rows[i]
                );
            }
            assert_eq!(rows[1]["state"], json!("unknown"), "{}", rows[1]);
            assert!(rows[1]["evidence"].as_str().unwrap().contains("[redacted]"));
            assert_eq!(rows[2]["state"], json!("usage_limit"), "{}", rows[2]);
            assert!(rows[2]["reset"]["raw"]
                .as_str()
                .unwrap()
                .contains("[redacted]"));
            assert!(rows[2]["evidence"].as_str().unwrap().contains("[redacted]"));
            assert_eq!(rows[3]["state"], json!("unreachable"), "{}", rows[3]);
            assert!(rows[3]["reason"].as_str().unwrap().contains("[redacted]"));
            for line in lines(res) {
                assert!(!line.contains(value), "{line}");
            }
        }
        // The error of a list that cannot be read.
        for (secret, value) in secrets() {
            let fake = Fake::new("{}");
            fake.file("list-fail-msg", &format!("no runtime {secret}\n"));
            let dir = tempfile::tempdir().unwrap();
            let r = fake.run(dir.path(), dir.path(), &[]);
            let message = failure(&r);
            assert!(message.contains("exited with code 5"), "{message}");
            assert!(message.contains("[redacted]"), "{message}");
            assert!(!message.contains(value), "{message}");
            assert!(!r.stderr.contains(value), "{}", r.stderr);
        }
    }

    /// Re-audit C-N1 and C-N5 (mutant C5): the bidirectional and zero-width characters of text a
    /// terminal chose are removed from the same five places as the control characters.
    #[test]
    fn lane_screen_scan_bidi_and_zero_width_characters_are_removed_from_every_shown_text() {
        let marks = "\u{202e}a\u{200b}b\u{2066}c\u{feff}d\u{2028}e";
        let fake = Fake::new(&list(
            &[
                term("t_title", &format!("lane {marks}"), &format!("br{marks}")),
                term("t_ask", "ask", "b"),
                term("t_limit", "limit", "b"),
                term("t_gone", "gone", "b"),
            ],
            4,
            false,
        ));
        fake.screen("t_title", &fixture("synthetic-working.txt"));
        let ask = format!(
            "{}Do you want to proceed? (yes/no) {marks}\n",
            fixture("synthetic-working.txt")
        );
        fake.screen("t_ask", &ask);
        let limit = fixture("synthetic-usage-limit-glm.txt")
            .replace("17:30:55", &format!("17:30:55 {marks}"));
        fake.screen("t_limit", &limit);
        fake.file("read-fail-msg-t_gone", &format!("denied {marks}\n"));
        let dir = tempfile::tempdir().unwrap();
        let r = fake.run(dir.path(), dir.path(), &[]);
        let res = r.result();
        let rows = res["terminals"].as_array().unwrap();
        assert_eq!(rows[0]["title"], json!("lane abcde"), "{}", rows[0]);
        assert_eq!(rows[0]["branch"], json!("brabcde"), "{}", rows[0]);
        assert!(rows[1]["evidence"].as_str().unwrap().ends_with("abcde"));
        assert!(rows[2]["reset"]["raw"].as_str().unwrap().ends_with("abcde"));
        assert!(rows[3]["reason"].as_str().unwrap().ends_with("abcde"));
        let all = res.to_string();
        for c in ['\u{202e}', '\u{200b}', '\u{2066}', '\u{feff}', '\u{2028}'] {
            assert!(!all.contains(c), "{c:?} left in {all}");
        }
        // The error of a list that cannot be read.
        let fake = Fake::new("{}");
        fake.file("list-fail-msg", &format!("no runtime {marks}\n"));
        let r = fake.run(dir.path(), dir.path(), &[]);
        let message = failure(&r);
        assert!(message.ends_with("abcde"), "{message}");
    }

    /// Re-audit C-N5 (mutants C2 and C3): the list is cut short when Orca says it has more
    /// terminals than it listed, and also when Orca says `truncated` alone; neither is needed
    /// for the other.
    #[test]
    fn lane_screen_scan_each_of_the_two_cut_short_signals_is_enough() {
        let two = [term("t_a", "a", "b"), term("t_b", "b", "b")];
        for (total, truncated, cut) in [(5, false, true), (2, true, true), (2, false, false)] {
            let fake = Fake::new(&list(&two, total, truncated));
            fake.screen("t_a", &fixture("synthetic-working.txt"));
            fake.screen("t_b", &fixture("synthetic-idle.txt"));
            let dir = tempfile::tempdir().unwrap();
            let r = fake.run(dir.path(), dir.path(), &[]);
            let res = r.result();
            assert_eq!(
                res["truncated"],
                json!(cut),
                "total={total} truncated={truncated}"
            );
            assert_eq!(
                res["total"],
                json!(total),
                "total={total} truncated={truncated}"
            );
            let last = lines(res).last().unwrap().clone();
            assert_eq!(last.contains("cut short"), cut, "{last}");
        }
    }

    /// Re-audit C-N5 (mutant C4): `by_state` counts every terminal of a state, not only the first.
    #[test]
    fn lane_screen_scan_by_state_counts_every_terminal_of_a_state() {
        let fake = Fake::new(&list(
            &[
                term("t_a", "a", "b"),
                term("t_b", "b", "b"),
                term("t_c", "c", "b"),
                term("t_d", "d", "b"),
            ],
            4,
            false,
        ));
        fake.screen("t_a", &fixture("synthetic-working.txt"));
        fake.screen("t_b", &fixture("synthetic-working.txt"));
        fake.screen("t_c", &fixture("synthetic-working.txt"));
        fake.screen("t_d", &fixture("synthetic-idle.txt"));
        let dir = tempfile::tempdir().unwrap();
        let r = fake.run(dir.path(), dir.path(), &[]);
        let by = &r.result()["by_state"];
        assert_eq!(by["working"], json!(3), "{by}");
        assert_eq!(by["idle"], json!(1), "{by}");
    }

    /// Re-audit C-N2 (slice B code): an `orca` that has exited but whose child still holds its
    /// output open must not hold the scan: the call ends within its limit, that one terminal is
    /// `unreachable`, and the others are read. Before the repair the call waited for the child.
    #[test]
    fn lane_screen_scan_an_orca_child_holding_the_output_does_not_hold_the_scan() {
        let dir = tempfile::tempdir().unwrap();
        let bin = dir.path().join("orca");
        common::write_executable(
            &bin,
            r#"#!/bin/sh
d="$FAKE_ORCA_DIR"
case "$2" in
list) cat "$d/list.json" ;;
read)
  if [ "$4" = t_hold ]; then sleep 10 & fi
  cat "$d/read-$4.json"
  ;;
*) exit 9 ;;
esac
"#,
        );
        fs::write(
            dir.path().join("list.json"),
            list(
                &[term("t_hold", "hold", "b"), term("t_ok", "ok", "b")],
                2,
                false,
            ),
        )
        .unwrap();
        let screen = fixture("synthetic-working.txt");
        let rows: Vec<&str> = screen.lines().collect();
        fs::write(dir.path().join("read-t_hold.json"), orca(&rows, "screen")).unwrap();
        fs::write(dir.path().join("read-t_ok.json"), orca(&rows, "screen")).unwrap();
        let started = std::time::Instant::now();
        let out = Command::new(BIN)
            .args(["lane-screen", "--scan", "--now", NOW_TEXT])
            .current_dir(dir.path())
            .env("HOME", dir.path())
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env("MOAI_ORCA_BIN", &bin)
            .env("MOAI_ORCA_TIMEOUT_SECS", "1")
            .env("FAKE_ORCA_DIR", dir.path())
            .stdin(Stdio::null())
            .output()
            .unwrap();
        let elapsed = started.elapsed();
        assert!(elapsed.as_secs() < 6, "the scan waited {elapsed:?}");
        let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
        let rows = envelope["result"]["terminals"].as_array().unwrap();
        assert_eq!(rows[0]["state"], json!("unreachable"), "{}", rows[0]);
        assert!(
            rows[0]["reason"]
                .as_str()
                .unwrap()
                .contains("within 1 seconds"),
            "{}",
            rows[0]
        );
        assert_eq!(rows[1]["state"], json!("working"), "{}", rows[1]);
    }
}
