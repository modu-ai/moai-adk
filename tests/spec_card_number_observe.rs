//! SPEC-20261123-card-number-observe: `moai spec check` counts card numbers (`C-` and digits) of
//! new and revised SPECs as an observe notice, and the sentence of rule 10 in `sdd.md` is pinned.
//! Texts are built in memory; two cases read the checked-in tree (`.moai/specs` and `sdd.md`)
//! without writing anything.

use moai_cli::spec_doc::{parse, Finding, Options, Severity, SpecDoc};
use std::fs;
use std::path::Path;

const CODE: &str = "card_number_observe";
const ROOT: &str = env!("CARGO_MANIFEST_DIR");

/// The sentence of rule 10 the pin test requires, word for word.
const RULE_10: &str = "Card numbers (`C-<n>`) never appear in a SPEC, branch name, commit message or pull request: cards are personal.";

/// A card number written at run time so this file holds no card number of its own.
fn card(n: u32) -> String {
    format!("C-{n}")
}

const REQS: &str = "## Requirements\n\n- REQ-1: When a user sends a valid password, the server \
                    shall accept the login.\n- REQ-2 (REQ-1): IF the password check fails, THEN \
                    the server shall refuse the login.\n\n";
const ACS: &str = "## Acceptance\n\n- AC-1 (REQ-1): A valid password logs in · verify: \
                   `cargo test --test login`\n\n";
const TAIL: &str = "## Scope\n\n- write: `src/login.rs`\n\n## Plan\n\n1. Fix the comparison.\n\n";

/// A minimal SPEC text with `intent` as its Intent and `extra` as the front matter after
/// `created` (an `approval:` block, or nothing).
fn text(id: &str, status: &str, created: Option<&str>, extra: &str, intent: &str) -> String {
    let created = created.map_or(String::new(), |c| format!("created: {c}\n"));
    format!(
        "---\nid: {id}\ntitle: \"Fix login\"\nstatus: {status}\nrevision: 1\n{created}{extra}---\n\n\
         # Fix login\n\n## Intent\n\n{intent}\n\n{REQS}{ACS}{TAIL}"
    )
}

const ID: &str = "SPEC-20261125-login-fix";

/// `t` with every card number's letter changed to `X`, read by a different method from the
/// implementation: a `C` or `c`, a `-`, a digit, and a letter or digit before it is not a card.
fn break_cards(t: &str) -> String {
    let cs: Vec<char> = t.chars().collect();
    let mut out = String::new();
    for (i, c) in cs.iter().enumerate() {
        let card_start = matches!(c, 'C' | 'c')
            && cs.get(i + 1) == Some(&'-')
            && cs.get(i + 2).is_some_and(|d| d.is_ascii_digit())
            && (i == 0 || !cs[i - 1].is_alphanumeric());
        out.push(if card_start { 'X' } else { *c });
    }
    out
}

fn parse_text(t: &str) -> SpecDoc {
    parse(t, &Options::default())
}

fn notices(d: &SpecDoc) -> Vec<&Finding> {
    d.findings.iter().filter(|f| f.code == CODE).collect()
}

/// An approved SPEC created on `created` and approved on `at`: the digest is read from the
/// draft of the same body so the approval is valid.
fn approved(created: &str, at: &str, intent: &str) -> String {
    let draft = parse_text(&text(ID, "draft", Some(created), "", intent));
    let digest = draft.digest.expect("a GEARS draft has a digest");
    let approval = format!(
        "approval:\n  by: user\n  at: {at}\n  basis: in-session selection\n  digest: {digest}\n"
    );
    text(ID, "approved", Some(created), &approval, intent)
}

fn intent_with_numbers() -> String {
    format!(
        "Source of the fix: card {} and card {}. AC-1 and SPEC-20261123-x and ABC-5 are not cards.",
        card(12),
        card(7)
    )
}

/// SPEC-20261123-card-number-observe AC-1
#[test]
fn a_new_spec_with_card_numbers_gets_one_observe_warning() {
    let intent = intent_with_numbers();
    let t = text(ID, "draft", Some("2026-11-25"), "", &intent);
    let d = parse_text(&t);
    let n = notices(&d);
    assert_eq!(
        n.len(),
        1,
        "one notice, not one per number: {:?}",
        d.findings
    );
    assert_eq!(n[0].severity, Severity::Warning);
    let first = t
        .lines()
        .position(|l| l.contains(&card(12)))
        .map(|i| i + 1)
        .unwrap();
    assert_eq!(n[0].line, Some(first), "the first line that holds a number");
    assert!(
        n[0].message.contains("2 card"),
        "the count: {}",
        n[0].message
    );
    assert!(
        !n[0].message.contains(&card(12)) && !n[0].message.contains("Source of the fix"),
        "the message never quotes the author's text: {}",
        n[0].message
    );
    // The notice adds no error: the same text with the numbers broken has the same errors.
    let broken = break_cards(&t);
    let plain = parse_text(&broken);
    assert!(notices(&plain).is_empty());
    assert_eq!(d.errors(), plain.errors());
    assert_eq!(d.warnings(), plain.warnings() + 1);
}

/// SPEC-20261123-card-number-observe AC-2
#[test]
fn a_revised_spec_is_judged_by_its_approval_date() {
    let intent = intent_with_numbers();
    let revised = parse_text(&approved("2026-11-17", "2026-11-25", &intent));
    assert_eq!(notices(&revised).len(), 1, "{:?}", revised.findings);
    assert_eq!(notices(&revised)[0].severity, Severity::Warning);
    let before = parse_text(&approved("2026-11-17", "2026-11-24", &intent));
    assert!(notices(&before).is_empty(), "{:?}", before.findings);
    // A SPEC without a card number gets no notice on either date.
    let clean = parse_text(&approved("2026-11-17", "2026-11-25", "No card here."));
    assert!(notices(&clean).is_empty());
}

/// SPEC-20261123-card-number-observe AC-3
#[test]
fn older_go_and_undated_specs_get_no_notice() {
    let intent = intent_with_numbers();
    // The day before the boundary, and the boundary day itself.
    let day_before = parse_text(&text(ID, "draft", Some("2026-11-24"), "", &intent));
    assert!(notices(&day_before).is_empty(), "{:?}", day_before.findings);
    let boundary = parse_text(&text(ID, "draft", Some("2026-11-25"), "", &intent));
    assert_eq!(notices(&boundary).len(), 1);
    // No date at all.
    let undated = parse_text(&text(ID, "draft", None, "", &intent));
    assert!(notices(&undated).is_empty(), "{:?}", undated.findings);
    // A Go-generation SPEC, even with a late date.
    let go = parse_text(&text(
        "SPEC-LOGIN-001",
        "draft",
        Some("2026-12-01"),
        "",
        &intent,
    ));
    assert!(notices(&go).is_empty(), "{:?}", go.findings);
}

/// The first number of a word start only: `AC-1`, `SPEC-2026…` and `ABC-5` are not cards, and
/// both cases of the letter count (the card reference reader takes both).
#[test]
fn only_a_card_number_at_a_word_start_counts() {
    let none = parse_text(&text(
        ID,
        "draft",
        Some("2026-11-25"),
        "",
        "AC-1 and SPEC-20261123-x and ABC-5 and C- and c-x and C-x.",
    ));
    assert!(notices(&none).is_empty(), "{:?}", none.findings);
    let lower = parse_text(&text(
        ID,
        "draft",
        Some("2026-11-25"),
        "",
        "Seen on card c-9.",
    ));
    assert_eq!(notices(&lower).len(), 1);
}

fn spec_texts() -> Vec<(String, String)> {
    let dir = Path::new(ROOT).join(".moai/specs");
    let mut out = Vec::new();
    for entry in fs::read_dir(&dir).expect("the specs folder") {
        let entry = entry.unwrap();
        let path = entry.path().join("spec.md");
        if let Ok(t) = fs::read_to_string(&path) {
            out.push((entry.file_name().to_string_lossy().to_string(), t));
        }
    }
    out.sort();
    out
}

/// SPEC-20261123-card-number-observe AC-4: the checked-in SPECs, read as they are.
#[test]
fn the_real_specs_before_the_date_get_no_notice() {
    let specs = spec_texts();
    assert!(
        specs.len() > 100,
        "the specs folder was read: {}",
        specs.len()
    );
    let mut judged = 0;
    for (folder, t) in &specs {
        let opts = Options {
            folder: Some(folder.clone()),
            closure: false,
        };
        let d = parse(t, &opts);
        // The notice is never an error, and never moves the error count of a real SPEC.
        for f in notices(&d) {
            assert_eq!(f.severity, Severity::Warning, "{folder}");
        }
        let broken = parse(&break_cards(t), &opts);
        // A number inside a requirement, a criterion or the scope changes the content digest of an
        // approved SPEC, which is the test's own edit and not the notice, so that code is left out.
        let codes = |s: &SpecDoc| {
            s.findings
                .iter()
                .filter(|f| f.severity == Severity::Error && f.code != "approval_digest")
                .map(|f| f.code)
                .collect::<Vec<_>>()
        };
        assert_eq!(codes(&d), codes(&broken), "{folder}: errors moved");
        let late = |date: &Option<String>| date.as_deref().is_some_and(|c| c >= "2026-11-25");
        let new_or_revised =
            late(&d.created) || late(&d.approval.as_ref().and_then(|a| a.at.clone()));
        if !new_or_revised {
            judged += 1;
            assert!(
                notices(&d).is_empty(),
                "{folder}: an older SPEC got a notice"
            );
        }
    }
    assert!(judged > 100, "older SPECs were judged: {judged}");
}

/// Whether `sdd` holds rule 10's sentence word for word.
fn rule_10_holds(sdd: &str) -> bool {
    sdd.lines()
        .any(|l| l.starts_with("10. **Shared places.** ") && l.contains(RULE_10))
}

/// SPEC-20261123-card-number-observe AC-5
#[test]
fn rule_10_sentence_is_pinned() {
    let sdd = fs::read_to_string(Path::new(ROOT).join("skills/moai/references/internal/sdd.md"))
        .expect("sdd.md");
    assert!(
        rule_10_holds(&sdd),
        "sdd.md rule 10 (Shared places) no longer holds the pinned sentence: {RULE_10}"
    );
    // A copy with one word changed, a copy without the rule, and an empty text are refused.
    assert!(!rule_10_holds(&sdd.replace("never appear", "may appear")));
    assert!(!rule_10_holds(
        &sdd.replace("cards are personal", "cards are shared")
    ));
    assert!(!rule_10_holds(
        &sdd.replace("10. **Shared places.** ", "10. **Other.** ")
    ));
    assert!(!rule_10_holds(""));
}
