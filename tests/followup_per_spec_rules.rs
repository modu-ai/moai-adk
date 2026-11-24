//! SPEC-20261124-followup-per-spec: the rule texts that used to open a card per finding now add to
//! an existing card. The leader's refill (`docs/leader-rules.md`), the lane rules, the recurrence
//! rule and the failed-remote-run sentence of `CLAUDE.md` and of the flow reference, and the
//! provenance row. The tests read files only; no host, no HOME and no command runs here.

use std::fs;

fn read(path: &str) -> String {
    fs::read_to_string(format!("{}/{path}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{path}: {e}"))
}

/// The first line of `text` that starts with `prefix`.
fn line_starting(text: &str, prefix: &str) -> String {
    text.lines()
        .find(|l| l.starts_with(prefix))
        .unwrap_or_else(|| panic!("no line starts with `{prefix}`"))
        .to_string()
}

fn assert_has(what: &str, text: &str, phrases: &[&str]) {
    for phrase in phrases {
        assert!(text.contains(phrase), "{what} lacks: {phrase}\n{text}");
    }
}

/// REQ-9, REQ-12: the leader turns a non-blocking finding into a criterion of the SPEC's open
/// follow-up card and opens one only when none is open; a lane adds no card for a finding.
#[test]
fn leader_and_lane_rules_name_the_follow_up_card() {
    let refill = line_starting(&read("docs/leader-rules.md"), "4. Refill");
    assert_has(
        "the Refill step",
        &refill,
        &[
            "an audit's non-blocking finding",
            "criterion of the SPEC's open follow-up card",
            "`moai card criterion-add <card> --ac <text>`",
            "`moai card add --spec <SPEC-ID> --follow-up`",
            "only when the SPEC has none open",
            "defect candidates with the RED test first",
        ],
    );
    let lane = read("docs/lane-rules.md");
    let item = line_starting(&lane, "- A non-blocking finding");
    assert_has(
        "the lane rule",
        &item,
        &[
            "its file and line",
            "a lane adds no card for it",
            "criterion of the SPEC's open follow-up card",
        ],
    );
}

/// REQ-10, REQ-11: a recurrence adds a note to the open card of its cause, and a failed remote
/// run becomes a criterion of the open follow-up card of the SPEC its head names; the pinned
/// words stay.
#[test]
fn recurrence_and_remote_failure_rules_use_existing_cards() {
    let claude = read("CLAUDE.md");
    let recurring = line_starting(&claude, "- Recurring problems change the source");
    assert_has(
        "the recurrence rule",
        &recurring,
        &[
            "the leader opens a card that round which changes moai-adk itself",
            "A recurrence whose cause already has an open card adds a note to that card and opens no new one.",
        ],
    );
    let landing = line_starting(&claude, "7. Landing.");
    let flow = read("skills/moai/references/flows/next-flow.md");
    let remote = line_starting(&flow, "- Remote checks never block a landing");
    let sentence = "Remote checks never block a landing; a failed remote run becomes a criterion \
                    of the open follow-up card of the SPEC the failed head names \
                    (`moai card criterion-add`).";
    assert_has("Flow 7 of CLAUDE.md", &landing, &[sentence]);
    assert_has("the flow reference", &remote, &[sentence]);
    assert!(
        !landing.contains("repair card") && !remote.contains("repair card"),
        "a failed run no longer opens a repair card"
    );
    let provenance = read("docs/claude-md-rule-provenance.md");
    let row = line_starting(
        &provenance,
        "| Standing principles: recurring problems change the source",
    );
    assert_has("the provenance row", &row, &["SPEC-20261124-followup-per-spec"]);
}
