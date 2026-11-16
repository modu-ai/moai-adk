//! The text of a card's summary (#3071 H4): what `CardSummary::render` puts on the host. A pure
//! function of the card, so the cases the command line cannot reach (a met or withdrawn
//! criterion, a text exactly at the cut) are checked here.

use moai_cli::ledger::summary::{summary_marker, CardSource, CardSummary, CriterionLine};

fn card(criteria: Vec<CriterionLine>) -> CardSummary {
    CardSummary {
        card_id: 7,
        title: "Add login".into(),
        kind: "feature".into(),
        state: "review".into(),
        source: Some(CardSource {
            host: "github".into(),
            project: "acme/widgets".into(),
            number: 12,
        }),
        criteria,
    }
}

fn line(seq: i64, text: &str, status: &str) -> CriterionLine {
    CriterionLine {
        seq,
        text: text.into(),
        status: status.into(),
        evidence_ref: None,
    }
}

#[test]
fn the_marker_names_the_card() {
    assert_eq!(summary_marker(7), "<!-- moai-card-summary C-7 -->");
    assert_ne!(summary_marker(7), summary_marker(8));
}

#[test]
fn each_criterion_shows_its_own_state() {
    let text = card(vec![
        line(1, "first", "met"),
        line(2, "second", "withdrawn"),
        line(3, "third", "open"),
    ])
    .render();
    assert!(
        text.starts_with("**MoAI card C-7** · feature · review\n\nAdd login\n"),
        "{text}"
    );
    assert!(text.contains("- AC1 [met] first\n"), "{text}");
    assert!(text.contains("- AC2 [withdrawn] second\n"), "{text}");
    assert!(text.contains("- AC3 [open] third\n"), "{text}");
}

#[test]
fn a_card_without_criteria_says_so() {
    let text = card(vec![]).render();
    assert!(
        text.contains("No acceptance criteria are recorded"),
        "{text}"
    );
    assert!(!text.contains("Acceptance criteria:"), "{text}");
}

#[test]
fn criterion_text_is_one_line_cut_at_300_characters() {
    let exactly = "x".repeat(300);
    let over = "y".repeat(301);
    let text = card(vec![
        line(1, &exactly, "open"),
        line(2, &over, "open"),
        line(3, "first line\n# a heading\n| a | table |", "open"),
    ])
    .render();
    assert!(
        text.contains(&format!("- AC1 [open] {exactly}\n")),
        "no ellipsis at 300"
    );
    assert!(
        text.contains(&format!("- AC2 [open] {}…\n", "y".repeat(300))),
        "an ellipsis after 300"
    );
    assert!(text.contains("- AC3 [open] first line\n"), "{text}");
    assert!(!text.contains("# a heading"), "{text}");
}
