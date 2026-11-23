//! SPEC-20261124-moai-leader-route: the skill has a ninth route, LEADER, whose workflow makes the
//! window that types `/moai leader` the registered coordination window of the project: it reads
//! the registry entry first, registers on `free`, stops on a held entry, renews once per round,
//! reads the ledger, sends audits to another session, lands only after a `PASS` for the exact head
//! and releases the registration when it ends. The tests read the package files of this checkout
//! and pin each sentence whole; whitespace is collapsed so a re-wrapped line does not break a pin.
//! Nothing is run and no HOME is touched.
use std::{fs, path::Path};

fn raw(rel: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(rel),
    )
    .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn package(rel: &str) -> String {
    raw(rel).split_whitespace().collect::<Vec<_>>().join(" ")
}

const LEADER: &str = "workflows/leader.md";

fn leader() -> String {
    package(LEADER)
}

fn assert_has(text: &str, needle: &str) {
    assert!(
        text.contains(needle),
        "{LEADER} lacks the sentence:\n  {needle}"
    );
}

fn assert_before(text: &str, first: &str, second: &str) {
    let a = text
        .find(first)
        .unwrap_or_else(|| panic!("{LEADER} lacks: {first}"));
    let b = text
        .find(second)
        .unwrap_or_else(|| panic!("{LEADER} lacks: {second}"));
    assert!(a < b, "{LEADER}: `{first}` must come before `{second}`");
}

/// REQ-10: what the window is, and that it holds no standing role.
const INTRO: &str = "`/moai leader` makes the window that types it the registered coordination window of the project; the window holds no standing role, and when it ends the lanes go on from the ledger alone.";
/// REQ-3: the registry is read first.
const READ_FIRST: &str = "Read the entry with `moai session-register --show --name leader` first.";
/// REQ-3: the registration, on `free` only.
const REGISTER: &str = "When it answers `free`, register this session with `moai session-register . --name leader --mode leader` and keep the `instance` the answer returns.";
/// REQ-4: a held entry or a conflict stops the route; any other refusal is an error.
const STOP: &str = "When another session holds the entry, or the registration is refused with `conflict`, stop with one line that names the holder's lease end and register nothing; a refusal of any other kind is an error to report, not an answer to who leads.";
/// REQ-5: the renewal, once per round, before the ledger is read.
const RENEW: &str = "Once per round, before you read the ledger, renew the registration with `moai session-register . --name leader --instance <instance>`.";
/// REQ-6: the ledger is the input, a completion message a notice.
const INPUT: &str = "Read the ledger with `moai card inbox`, `moai card wait` and `moai card next`; a lane's completion message is a notice that you check against the ledger before you act, as the \"Completion message\" and \"Transition\" sections of `flows/next-flow` say.";
/// D-3: the dispatch channel.
const DISPATCH: &str = "Send the next card's dispatch with the host's message tool, or with `moai lane-send` when the lanes run in terminals that command can reach: it reads the screen first and sends only to a lane that is working or idle; no `moai` command sends a cross-session message yet.";
/// REQ-7: the audit goes to a session that wrote none of the change.
const AUDIT: &str = "Send the audit of a head to a session that wrote none of the change (the AUDIT route); the author of a change never audits it.";
/// REQ-7: landing after a `PASS` for the exact head, by the leader.
const LAND: &str = "Land that head with `moai card land <card> --head <40 hex> --by leader` only after a verdict `PASS` for that exact head; the command accepts only `--by leader` today, so this window is the caller.";
/// REQ-8: a decision note is sorted in the order of the skill entry.
const DECIDE: &str = "A `LANE-BLOCKED kind=decision` note is sorted in the order of the skill entry: a reversible choice of category none is decided here and recorded as a note on the card, and a choice that touches a HOLD category or a `guarded` setting goes to the user as a selection with the recommended option first and waits for the pick.";
/// REQ-9: the release at the end.
const RELEASE: &str = "When the route ends, release the registration with `moai session-register . --name leader --release --instance <instance>`.";
/// REQ-9: a closed window releases nothing.
const CLOSED: &str = "A window that is closed releases nothing: its lease ends by itself (four hours by default), and the lanes then write the ledger note only.";

/// The procedure's step keys, in order.
const STEPS: [&str; 6] = ["register", "round", "dispatch", "audit", "land", "end"];

#[test]
fn register_reads_the_entry_first_and_registers_on_free() {
    let t = leader();
    for s in [READ_FIRST, REGISTER] {
        assert_has(&t, s);
    }
    assert_before(&t, READ_FIRST, REGISTER);
}

#[test]
fn register_stops_on_a_held_entry_or_a_conflict() {
    assert_has(&leader(), STOP);
}

#[test]
fn round_renews_once_before_it_reads_the_ledger() {
    let t = leader();
    for s in [RENEW, INPUT] {
        assert_has(&t, s);
    }
    assert_before(&t, RENEW, INPUT);
}

#[test]
fn round_points_to_the_flow_reference_and_repeats_none_of_it() {
    let t = leader();
    assert_has(
        &t,
        "Read the flow reference `flows/next-flow` before the first round",
    );
    let flow = package("references/flows/next-flow.md");
    // Whole sentences of the flow reference's two sections stay there (parent REQ-11).
    for s in [
        "When the leader gets the message, it reads the ledger first",
        "A message the ledger does not back starts nothing",
        "A lane never sends `/clear` to itself",
        "The leader sends `/clear` to the exact existing window only after",
    ] {
        assert!(flow.contains(s), "next-flow lost the sentence: {s}");
        assert!(!t.contains(s), "{LEADER} repeats a flow sentence: {s}");
    }
}

#[test]
fn round_sends_the_dispatch_by_a_channel_that_exists() {
    assert_has(&leader(), DISPATCH);
}

#[test]
fn round_sends_the_audit_to_another_session_and_lands_only_after_a_pass() {
    let t = leader();
    for s in [AUDIT, LAND] {
        assert_has(&t, s);
    }
    assert_before(&t, AUDIT, LAND);
}

#[test]
fn round_sorts_a_decision_note_by_the_open_question_order() {
    assert_has(&leader(), DECIDE);
}

#[test]
fn round_steps_run_in_order() {
    let text = raw(LEADER);
    let start = text.find("## Procedure").expect("a Procedure section");
    let rest = &text[start + "## Procedure".len()..];
    let body = &rest[..rest.find("\n## ").unwrap_or(rest.len())];
    let mut keys: Vec<(usize, String)> = Vec::new();
    for line in body.lines() {
        let Some((num, tail)) = line.split_once(". **") else {
            continue;
        };
        let Ok(n) = num.trim().parse::<usize>() else {
            continue;
        };
        let key = tail.split("**").next().unwrap_or("").to_string();
        keys.push((n, key));
    }
    let numbers: Vec<usize> = keys.iter().map(|(n, _)| *n).collect();
    let names: Vec<&str> = keys.iter().map(|(_, k)| k.as_str()).collect();
    assert_eq!(names, STEPS, "the LEADER steps");
    assert_eq!(numbers, (1..=STEPS.len()).collect::<Vec<_>>());
}

#[test]
fn end_releases_the_registration() {
    assert_has(&leader(), RELEASE);
}

#[test]
fn end_a_closed_window_releases_nothing() {
    assert_has(&leader(), CLOSED);
}

#[test]
fn end_the_window_holds_no_standing_role() {
    assert_has(&leader(), INTRO);
}

#[test]
fn end_workflow_stays_within_its_byte_budget() {
    let n = raw(LEADER).len();
    assert!(n <= 6144, "{LEADER} is {n} bytes, over 6144");
}
