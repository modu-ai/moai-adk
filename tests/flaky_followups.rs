//! SPEC-20261119-flaky-followups, slice A (card C-99, REQ-1 to REQ-3): the flaky rule's card
//! command carries a criterion and the card it makes can be read back with that criterion; the
//! repair rule (no sleep, retry or skip; a test whose root cause is not found is deleted with the
//! reason written) is in the MODE-AGENT text; the one record home (the flaky card, the `[flake]`
//! Issue only for a project without a card ledger) is in the VERIFY text.
//!
//! The tests read three workflow files, compiled in with `include_str!`, and run `moai card add`
//! and `moai card show` against a temporary HOME; the real `~/.moai` is not read or written.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::Value;

const MODE_LEADER: &str = include_str!("../skills/moai/workflows/mode-leader.md");
const MODE_AGENT: &str = include_str!("../skills/moai/workflows/mode-agent.md");
const VERIFY: &str = include_str!("../skills/moai/workflows/verify.md");

/// The command the merge-train step gives for the flaky card: the text between the backtick
/// after `titled "flaky: <test>" (` and the closing backtick.
fn flaky_command() -> String {
    let at = MODE_LEADER
        .find("(`moai card add --kind task --title \"flaky: <test>\"")
        .expect("the merge-train step gives the flaky card command");
    let rest = &MODE_LEADER[at + 2..];
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

/// SPEC-20261119-flaky-followups AC-2 (REQ-2).
#[test]
fn the_repair_rule_is_in_the_mode_agent_workflow() {
    let step = line_with(MODE_AGENT, "Flaky repair (card C-99): ");
    assert!(
        step.starts_with("7. **heavy-checks**"),
        "the sentence sits in the heavy-checks step"
    );
    let s = sentence_from(step, "Flaky repair (card C-99): ");
    // Whole phrases, so a sentence that turns the rule round ("is never deleted") does not pass.
    for word in [
        "never adds a sleep, a retry or a skip",
        "root cause",
        "is deleted, with the reason written in the card and the SPEC",
    ] {
        assert!(
            s.contains(word),
            "the repair sentence must say {word:?}: {s}"
        );
    }
}

/// SPEC-20261119-flaky-followups AC-2 (REQ-3).
#[test]
fn the_record_home_is_in_the_verify_workflow() {
    let para = line_with(VERIFY, "Record home (card C-99): ");
    assert!(
        para.starts_with("A test that fails on one run and passes when the same head is run again"),
        "the sentence sits in the flaky paragraph"
    );
    let s = sentence_from(para, "Record home (card C-99): ");
    // Whole phrases: the card is the home where a ledger exists, the Issue only where none does.
    for word in [
        "where the project has a card ledger, the record is the flaky card of the merge-train step (`flaky: <test>`)",
        "the `[flake]` Issue and the two commands below are for a project without a card ledger",
    ] {
        assert!(
            s.contains(word),
            "the record-home sentence must say {word:?}: {s}"
        );
    }
    // The Issue path and its two commands stay.
    assert!(para.contains("flake-record"), "the commands stay");
}
