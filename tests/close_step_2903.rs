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
    for phrase in [
        "The lane closes its own card (Issue #2903, #2899 G7",
        "once the card's last pull request has merged",
        "first line is `CLOSE-TABLE #<n>`",
        "a verdict column (`상태`, `판정` or `결과`)",
        "`moai close-check <n> --repo <owner>/<repo>`",
        "On `close`, release the claim and close the Issue with `gh issue close <n> --comment` naming the close table by its link",
        "read the Issue's state back",
        "On `keep_open`, write the follow-up card number into each row the answer lists under `unmet`",
        "a lane opens no new Issue",
        "leave the closing to the leader",
        "an open pull request that still claims the Issue means the card is not done",
        "`no_table` and `uncertain` close nothing",
        "The leader only confirms a lane's closing (MODE-LEADER `lane-done` step)",
    ] {
        assert!(hand_off.contains(phrase), "hand-off step must say `{phrase}`");
    }
    // The old sentence, which left closing to nobody in particular, is gone.
    assert!(!text.contains("After the merge the Issue is closed and no longer a candidate"));
    assert!(text.contains(
        "- (4) An Issue-closing comparison table for a merged pull request, then `moai close-check` and, on `close`, the Issue's closing (`hand-off` step)."
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
