//! SPEC-20261119-flaky-followups, slice A (card C-99, REQ-1 to REQ-3): the flaky rule's card
//! command carries a criterion and the card it makes can be read back with that criterion; the
//! repair rule (no sleep, retry or skip; a test whose root cause is not found is deleted with the
//! reason written) and the one record home (the flaky card) are in the RUN workflow's rule
//! reference, `references/internal/check-rules.md`. The rules came from the retired VERIFY route
//! and the mode workflows (SPEC-20261122-next-flow-skills-agents REQ-2) and moved out of
//! `run.md` with SPEC-20261125-run-sync-procedure-receipt; the `[flake]` Issue procedure for a
//! project without a card ledger and its two commands are retired there (decisions of
//! 2026-11-16 to 2026-11-25: the card ledger is the work ledger, GitHub is no ledger).
//!
//! The tests read the rule reference, compiled in with `include_str!`, and run `moai card add` and
//! `moai card show` against a temporary HOME; the real `~/.moai` is not read or written.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::Value;

const RULES: &str = include_str!("../skills/moai/references/internal/check-rules.md");

/// The command the flaky rule gives for the flaky card: the code span that opens with
/// `moai card add --kind task --title "flaky: <test>"`, without its backticks.
fn flaky_command() -> String {
    let at = RULES
        .find("`moai card add --kind task --title \"flaky: <test>\"")
        .expect("the flaky rule gives the flaky card command");
    let rest = &RULES[at + 1..];
    rest[..rest.find('`').expect("the command ends with a backtick")].to_string()
}

/// Split a command line at spaces, keeping a double-quoted part together without its quotes.
fn words(line: &str) -> Vec<String> {
    let (mut out, mut cur, mut quoted) = (Vec::new(), String::new(), false);
    for c in line.chars() {
        match c {
            '"' => quoted = !quoted,
            ' ' if !quoted => {
                if !cur.is_empty() {
                    out.push(std::mem::take(&mut cur));
                }
            }
            _ => cur.push(c),
        }
    }
    if !cur.is_empty() {
        out.push(cur);
    }
    out
}

/// The line of the file that holds `mark` (one paragraph is one line); `mark` is written once.
fn line_with<'a>(text: &'a str, mark: &str) -> &'a str {
    assert_eq!(
        text.matches(mark).count(),
        1,
        "{mark:?} must be written exactly once"
    );
    text.lines()
        .find(|l| l.contains(mark))
        .unwrap_or_else(|| panic!("no line holds {mark:?}"))
}

/// The sentence that starts at `mark` inside `line`, up to and with its first full stop followed
/// by a space or the end of the line.
fn sentence_from<'a>(line: &'a str, mark: &str) -> &'a str {
    let at = line.find(mark).unwrap();
    let rest = &line[at..];
    let end = rest.find(". ").map(|i| i + 1).unwrap_or(rest.len());
    &rest[..end]
}

/// SPEC-20261119-flaky-followups AC-1 (REQ-1).
#[test]
fn the_flaky_card_command_carries_an_ac_and_the_card_can_be_read_with_it() {
    let line = flaky_command();
    assert!(line.contains(" --ac "), "the command has no --ac: {line}");
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    let project = project.to_str().unwrap().to_string();
    // The sentence as written, with the test name filled in and the project named.
    let mut args: Vec<String> = words(&line.replace("<test>", "ci_gate_times_out"))
        .into_iter()
        .skip(1) // `moai`
        .collect();
    args.extend(["--project".to_string(), project.clone()]);
    let args: Vec<&str> = args.iter().map(String::as_str).collect();
    let out = home.command(&args).output().expect("run moai");
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let shown = home
        .command(&["card", "show", "C-1", "--project", &project])
        .output()
        .expect("run moai");
    let v: Value = serde_json::from_slice(&shown.stdout).unwrap();
    assert_eq!(v["result"]["title"], "flaky: ci_gate_times_out", "{v}");
    let criteria = v["result"]["criteria"].as_array().unwrap();
    assert_eq!(criteria.len(), 1, "{v}");
    assert_eq!(criteria[0]["status"], "open", "{v}");
    assert!(
        criteria[0]["text"].as_str().unwrap().contains("flake"),
        "{v}"
    );
}

/// SPEC-20261119-flaky-followups AC-2 (REQ-2). The repair rule is two sentences of one bullet
/// now: the repair itself, then the test whose root cause is not found.
#[test]
fn the_repair_rule_is_in_the_mode_agent_workflow() {
    let mark = "A repair of a flaky test never adds";
    let para = line_with(RULES, mark);
    let repair = sentence_from(para, mark);
    // Whole phrases, so a sentence that turns the rule round ("is never deleted") does not pass.
    for word in ["never adds a sleep, a retry or a skip", "root cause"] {
        assert!(
            repair.contains(word),
            "the repair sentence must say {word:?}: {repair}"
        );
    }
    let deleted = sentence_from(para, "A test whose root cause is not found");
    assert!(
        deleted.contains("is deleted, with the reason written in the card and the SPEC"),
        "the unfound-root-cause sentence must say the test is deleted with its reason: {deleted}"
    );
}

/// SPEC-20261119-flaky-followups AC-2 (REQ-3). The one record home is the flaky card; the
/// `[flake]` Issue, `moai flake-record` and the project-without-a-ledger path are retired (the
/// card ledger is the work ledger), so the pin reads the card's own sentences and that the Issue
/// path stays out.
#[test]
fn the_record_home_is_in_the_verify_workflow() {
    let para = line_with(
        RULES,
        "A failure that a re-run on the same SHA turned green",
    );
    // Whole phrases: one card per test, a note on the open one, the four facts it records, and
    // the merge comment when the card cannot be written.
    for word in [
        "does not stop the merge and is not forgotten",
        "Open one card per test:",
        "when one is open, add a `moai card note <card> --kind report` note instead",
        "Record the test, the failed run id, the passing run id and the SHA.",
        "If the card cannot be written, put those four facts in the merge comment and say the card is missing.",
    ] {
        assert!(
            para.contains(word),
            "the record-home bullet must say {word:?}: {para}"
        );
    }
    assert!(
        !RULES.contains("[flake]") && !RULES.contains("flake-record"),
        "check-rules.md names the retired Issue path"
    );
}
