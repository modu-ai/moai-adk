//! SPEC-20261123-lane-completion-handshake: the shipped skill text tells a work lane, after its
//! READY or REPAIRED note and the release of the card, to send the leader one cross-session message to the
//! registered session id, and tells the leader to read the ledger first and then clear and
//! dispatch. The tests read the two package files and pin each sentence; whitespace is collapsed
//! so a re-wrapped line does not break a pin.
use std::{fs, path::Path};

fn package(rel: &str) -> String {
    let text = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(rel),
    )
    .unwrap_or_else(|e| panic!("{rel}: {e}"));
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

fn next_flow() -> String {
    package("references/flows/next-flow.md")
}

fn work() -> String {
    package("workflows/work.md")
}

fn assert_has(text: &str, file: &str, needle: &str) {
    assert!(
        text.contains(needle),
        "{file} lacks the sentence:\n  {needle}"
    );
}

fn assert_lacks(text: &str, file: &str, needle: &str) {
    assert!(!text.contains(needle), "{file} still carries: {needle}");
}

/// REQ-1: the order, the hand-back before any message.
const ORDER: &str =
    "A lane writes its READY or REPAIRED note and gives the card back before it sends any message.";
/// REQ-2: the one message to the registered leader.
const SEND: &str = "When a registered leader holds a live lease, the lane then sends that leader one cross-session message that names the card and the head.";
/// REQ-3: the ledger-only fallback and no retry.
const FALLBACK: &str = "Without a registered leader, a live lease or a cross-session channel, or when the send is refused or held, the lane writes the ledger note only; the leader's ledger watch picks the note up.";
/// REQ-3, audit #373 N1: one message per finished card, never a retry.
const RETRY: &str = "One message per finished card, no retry loop.";
/// Audit #373 N4: the order of the message and the `LANE-DONE` line.
const LAST_LINE: &str = "`LANE-DONE` is the lane's last line, after the message when one is sent.";
/// Audit #373 N5 (SPEC R-2): the registered id is matched to a row of the host's list.
const MATCH: &str = "The lane matches that id to a row of the host's session list; with no matching row it sends nothing.";
/// REQ-4: the address.
const ADDRESS: &str = "The lane reads the leader's session id with `moai session-register --show --name leader` and addresses exactly that session; it never sends to a row picked by the name `leader`, because another project's session can carry that name.";
/// The message is a notice only (Transition rule, ledger first).
const NOTICE: &str = "The message is a notice, not proof: the ledger decides.";
/// REQ-7: what a lane never does.
const LANE_NEVER: &str = "A lane never sends `/clear` to itself, and never takes its next card on its own while a leader is registered.";
/// REQ-5: the leader reads the ledger first.
const LEADER: &str = "When the leader gets the message, it reads the ledger first, and sends `/clear` and then the next card's dispatch only when the ledger shows the card done or released and nothing is open.";
/// REQ-6: a notice the ledger does not back.
const LEADER_ELSE: &str = "A message the ledger does not back starts nothing: the leader sends no `/clear` and no dispatch, and waits for the ledger.";
/// REQ-8: the sending command is not implemented.
const GAP: &str = "No `moai` command sends this message yet: `moai host-wake` addresses a session by name, which the address rule forbids, so a lane uses the host's own message tool (SPEC-20261122-session-messages).";
/// REQ-9: the language of the leader's lines.
const LANG: &str = "The lane's message and the leader's dispatch, answer and line after `/clear` are written in the user's conversation language; only a prompt to a subagent is English, and commands, paths, card ids and protocol lines stay verbatim.";
/// Audit #373 B2: a blocked card moves the lane on only when no leader is registered.
const BLOCKED: &str = "A decision the lane cannot make goes on the card as a blocked note (`workflows/work.md`); with no registered leader the lane takes the next card, and with a registered leader it waits for that leader's dispatch.";

const NEXT_FLOW: &str = "references/flows/next-flow.md";
const WORK: &str = "workflows/work.md";

#[test]
fn ac1_the_lane_hands_back_then_sends_one_message_and_falls_back_to_the_ledger() {
    let nf = next_flow();
    for s in [ORDER, SEND, FALLBACK, RETRY, MATCH, LAST_LINE] {
        assert_has(&nf, NEXT_FLOW, s);
    }
    // The order is the text's order: the hand-back sentence comes before the message sentence.
    assert!(nf.find(ORDER).unwrap() < nf.find(SEND).unwrap());
    let w = work();
    for needle in [
        "READY or REPAIRED note",
        "one cross-session message",
        "ledger note only",
        "one message per finished card, never a retry",
        "`LANE-DONE` is your last line",
    ] {
        assert_has(&w, WORK, needle);
    }
    assert!(
        w.find("READY or REPAIRED note").unwrap() < w.find("one cross-session message").unwrap()
    );
    assert_lacks(&w, WORK, "HANDOFF");
    assert_lacks(&nf, NEXT_FLOW, "HANDOFF");
}

#[test]
fn ac2_the_address_is_the_registered_id_and_a_lane_never_clears_itself() {
    let nf = next_flow();
    for s in [ADDRESS, NOTICE, LANE_NEVER] {
        assert_has(&nf, NEXT_FLOW, s);
    }
    let w = work();
    for needle in [
        "moai session-register --show --name leader",
        "never a row picked by name",
        "never send `/clear` to yourself",
        "never take the next card on your own while a leader is registered",
        "match that id to a row of the host's session list",
        // SPEC-20261123-lane-next-card-rule REQ-6: the no-row sentence is pinned.
        "with no matching row, send nothing",
    ] {
        assert_has(&w, WORK, needle);
    }
}

#[test]
fn ac3_the_leader_reads_the_ledger_first_and_an_unbacked_notice_starts_nothing() {
    let nf = next_flow();
    for s in [LEADER, LEADER_ELSE] {
        assert_has(&nf, NEXT_FLOW, s);
    }
    // The leader's sentences stand in the Transition part, next to the clear conditions.
    let transition = nf.find("### Transition").expect("Transition heading");
    let final_flow = nf.find("### Final flow").expect("Final flow heading");
    for s in [LEADER, LEADER_ELSE] {
        let at = nf.find(s).unwrap();
        assert!(
            transition < at && at < final_flow,
            "{s} is outside Transition"
        );
    }
}

#[test]
fn ac4_the_missing_command_and_the_language_rule_are_written_down() {
    let nf = next_flow();
    for s in [GAP, LANG] {
        assert_has(&nf, NEXT_FLOW, s);
    }
    // Written as not implemented, never as working: no sentence says a command sends the notice.
    for working in [
        "moai lane-notify",
        "moai card notify",
        "the command sends the message",
    ] {
        assert_lacks(&nf, NEXT_FLOW, working);
    }
    // The skill texts keep their own forbidden words out (the scan's list, the ones this text
    // could slip into): the registered window is "the leader", never a role or a mode.
    for forbidden in ["resident leader", "leader mode"] {
        assert_lacks(&nf.to_lowercase(), NEXT_FLOW, forbidden);
        assert_lacks(&work().to_lowercase(), WORK, forbidden);
    }
}

#[test]
fn ac6_a_blocked_card_moves_the_lane_on_only_when_no_leader_is_registered() {
    let nf = next_flow();
    assert_has(&nf, NEXT_FLOW, BLOCKED);
    assert_has(
        &nf,
        NEXT_FLOW,
        "With no registered leader the worker reserves the next eligible card once,",
    );
    let w = work();
    assert_has(
        &w,
        WORK,
        "With no registered leader, take the next card; with a registered leader, leave the blocked card as its note and wait for that leader's dispatch.",
    );
    // The old unconditional wording is gone from both files.
    assert_lacks(&w, WORK, "never stops the loop");
    assert_lacks(
        &nf,
        NEXT_FLOW,
        "as a blocked note (`workflows/work.md`); the lane takes the next card.",
    );
}

/// SPEC-20261123-lane-next-card-rule REQ-1, REQ-2: `CLAUDE.md` says the same thing in both places.
const CLAUDE_BLOCKED: &str = "as a note on the card; with no registered leader take the next card, otherwise wait for that leader's dispatch.";
const CLAUDE_STEP9: &str = "nothing uncommitted is lost; with no registered leader take the next card, otherwise wait for that leader's dispatch.";
/// REQ-3: step 7 (repeat) of `work.md`.
const WORK_REPEAT: &str = "is not the start signal: with no registered leader, take the next card once, when the ledger shows the previous card done or released (`flows/next-flow`); with a registered leader, wait for that leader's dispatch.";
/// REQ-4: the general `/clear` line of the Transition part.
const CLEAR_LINE: &str = "The leader sends `/clear` to the exact existing window only after the ledger shows the lane's card done or released, and only when nothing is open:";
/// REQ-5: the two phrases all three texts share.
const SHARED: [&str; 2] = ["no registered leader", "for that leader's dispatch"];

fn claude_md() -> String {
    let text = fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("CLAUDE.md"))
        .unwrap_or_else(|e| panic!("CLAUDE.md: {e}"));
    assert!(
        text.lines().count() <= 120,
        "CLAUDE.md grew past 120 lines (SPEC REQ-7)"
    );
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

#[test]
fn ac7_every_next_card_sentence_waits_for_a_registered_leader() {
    let c = claude_md();
    assert_has(&c, "CLAUDE.md", CLAUDE_BLOCKED);
    assert_has(&c, "CLAUDE.md", CLAUDE_STEP9);
    let w = work();
    assert_has(&w, WORK, WORK_REPEAT);
    // The three texts say the same thing: the same two phrases stand in each of them.
    let nf = next_flow();
    for (text, file) in [(&c, "CLAUDE.md"), (&nf, NEXT_FLOW), (&w, WORK)] {
        for phrase in SHARED {
            assert_has(text, file, phrase);
        }
    }
    // The old unconditional wording is gone.
    assert_lacks(
        &c,
        "CLAUDE.md",
        "as a note on the card and take the next card.",
    );
    assert_lacks(&c, "CLAUDE.md", "is lost, then take the next card.");
}

#[test]
fn ac8_the_general_clear_line_needs_the_ledger_to_show_done_or_released() {
    assert_has(&next_flow(), NEXT_FLOW, CLEAR_LINE);
}
