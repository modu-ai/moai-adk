//! SPEC-20261124-followup-per-spec, as `SPEC-20261125-completion-flow` changed it: the rule texts
//! that used to add a criterion to a follow-up card now add one line to `docs/later.md`. The
//! leader's refill (`docs/leader-rules.md`), the lane rules, the recurrence rule and the
//! failed-remote-run sentence of `CLAUDE.md` and of the flow reference, and the provenance rows.
//! The tests read files only; no host, no HOME and no command runs here.

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

/// The leader turns a line of the later list into a card only while the work cap allows it; a
/// lane adds neither a criterion nor a card for a finding, only a line once the card has landed.
#[test]
fn leader_and_lane_rules_name_the_later_list() {
    let refill = line_starting(&read("docs/leader-rules.md"), "4. Refill");
    assert_has(
        "the Refill step",
        &refill,
        &[
            "a line of `docs/later.md` turned into a card",
            "only while the work cap allows it",
        ],
    );
    assert!(
        !refill.contains("criterion-add") && !refill.contains("--follow-up"),
        "the Refill step still adds criteria to a follow-up card"
    );
    let lane = read("docs/lane-rules.md");
    let item = line_starting(&lane, "- A non-blocking finding");
    assert_has(
        "the lane rule",
        &item,
        &[
            "its file and line",
            "one line of `docs/later.md`",
            "nobody adds a criterion or a card for it",
        ],
    );
    assert!(
        read("docs/later.md").contains("One line per item"),
        "docs/later.md does not say its one-line form"
    );
}

/// A recurrence adds one line to the later list, and a failed remote run becomes one line of it
/// too; neither opens a card or adds a criterion.
#[test]
fn recurrence_and_remote_failure_rules_use_the_later_list() {
    let claude = read("CLAUDE.md");
    let recurring = line_starting(&claude, "- Recurring problems change the source");
    assert_has(
        "the recurrence rule",
        &recurring,
        &["adds one line to `docs/later.md` that names the cause and the change at the source"],
    );
    let landing = line_starting(&claude, "7. Landing.");
    let flow = read("skills/moai/references/flows/next-flow.md");
    let remote = line_starting(&flow, "- Remote checks never block a landing");
    let sentence = "Remote checks never block a landing; a failed remote run becomes one line \
                    of `docs/later.md`.";
    assert_has("Flow 7 of CLAUDE.md", &landing, &[sentence]);
    assert_has("the flow reference", &remote, &[sentence]);
    assert!(
        !landing.contains("criterion-add") && !remote.contains("criterion-add"),
        "a failed run no longer adds a criterion"
    );
    let provenance = read("docs/claude-md-rule-provenance.md");
    // The earlier row keeps its origin; the newer row names the SPEC that changed the rule.
    let rows: Vec<&str> = provenance
        .lines()
        .filter(|l| l.starts_with("| Standing principles: recurring problems change the source"))
        .collect();
    assert_eq!(rows.len(), 2, "{rows:#?}");
    assert_has(
        "the earlier provenance row",
        rows[0],
        &["SPEC-20261124-followup-per-spec"],
    );
    assert_has(
        "the newer provenance row",
        rows[1],
        &["SPEC-20261125-completion-flow", "`docs/later.md`"],
    );
}
