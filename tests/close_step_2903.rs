//! #2903 (#2899 O4) stage 2: the agent workflow has the lane close its own card — close table,
//! `moai close-check`, then close on `close` or keep open with follow-up cards on `keep_open` —
//! and the leader workflow only confirms it. The flags the workflow names are the command's own.

mod common;
use std::{fs, path::Path, process::Command};

fn read(rel: &str) -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)).unwrap()
}

#[test]
fn the_agent_workflow_closes_its_own_card_through_close_check() {
    let text = read("skills/moai/workflows/mode-agent.md");
    let hand_off = text
        .lines()
        .find(|l| l.starts_with("10. **hand-off**"))
        .expect("hand-off step");
    // Card C-53 S5: the lane closes its card in the ledger (decision L5) — criteria marked,
    // `card pr` after the merge, `card done`, whose refusal names the rule — instead of the
    // Issue's close table, `close-check` and `gh issue close`. A criterion the merge does not
    // meet is withdrawn naming the follow-up card (was: `keep_open` rows with follow-up cards).
    for phrase in [
        "The lane closes its own card: once",
        "once the card's last pull request has merged",
        "`card criterion <card> <n> --generation <g> --met --evidence <ref>`",
        "run `card pr` again (it records the merge) and `card done <card> --generation <g>`",
        "`done` refuses a card with no met criterion, an open criterion or no merged pull request, and its refusal names the rule",
        "withdrawn with a reason that names the follow-up card the leader adds (a lane adds no card)",
        "an open pull request of the card means it is not done",
        "The leader only confirms a lane's closing (MODE-LEADER `lane-done` step)",
    ] {
        assert!(hand_off.contains(phrase), "hand-off step must say `{phrase}`");
    }
    // The old sentence, which left closing to nobody in particular, is gone; so are the Issue
    // close table and `close-check`.
    assert!(!text.contains("After the merge the Issue is closed and no longer a candidate"));
    for gone in ["CLOSE-TABLE", "close-check", "gh issue close"] {
        assert!(!text.contains(gone), "mode-agent.md still names `{gone}`");
    }
    assert!(text.contains(
        "- (3) A card of yours whose pull request has merged: `card pr`, `card done` and the cleanup (`hand-off` step)."
    ));
    // The leader side (O1, #2910) defers to a lane that already closed.
    assert!(read("skills/moai/workflows/mode-leader.md").contains("unless the lane already did"));
}

/// The command, its verdict words and the flag the workflow names exist.
#[test]
fn the_workflow_names_only_what_close_check_has() {
    let out = Command::new(crate::common::MOAI_BIN)
        .args(["close-check", "--help"])
        .output()
        .unwrap();
    assert!(out.status.success());
    let help = String::from_utf8_lossy(&out.stdout);
    assert!(help.contains("--repo"), "{help}");
    for verdict in ["close", "keep_open", "no_table", "uncertain"] {
        assert!(help.contains(verdict), "{verdict} not in help: {help}");
    }
    for heading in ["상태", "판정", "결과"] {
        assert!(
            moai_cli::close_check::VERDICT_HEADINGS.contains(&heading),
            "{heading}"
        );
    }
}
